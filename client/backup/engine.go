package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/shared/config"
)

// BackupState represents the state of the backup engine
type BackupState string

const (
	StateIdle     BackupState = "idle"
	StateScanning BackupState = "scanning"
	StateBacking  BackupState = "backing_up"
	StateError    BackupState = "error"
	StatePaused   BackupState = "paused"
)

// BackupStats holds statistics about the backup
type BackupStats struct {
	FilesScanned   int64
	FilesUploaded  int64
	FilesFailed    int64
	BytesUploaded  int64
	LastBackupTime time.Time
	CurrentState   BackupState
	LastError      error
}

// FileChangeQueue represents a queue of file changes to backup
type FileChangeQueue struct {
	items map[string]*FileEvent
	mu    sync.RWMutex
}

// NewFileChangeQueue creates a new file change queue
func NewFileChangeQueue() *FileChangeQueue {
	return &FileChangeQueue{
		items: make(map[string]*FileEvent),
	}
}

// Add adds or updates a file change in the queue
func (q *FileChangeQueue) Add(event FileEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Use path as key to deduplicate multiple changes to same file
	q.items[event.Path] = &event
}

// Remove removes a file from the queue
func (q *FileChangeQueue) Remove(path string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.items, path)
}

// GetAll returns all queued changes and clears the queue
func (q *FileChangeQueue) GetAll() []FileEvent {
	q.mu.Lock()
	defer q.mu.Unlock()

	events := make([]FileEvent, 0, len(q.items))
	for _, event := range q.items {
		events = append(events, *event)
	}

	// Clear the queue
	q.items = make(map[string]*FileEvent)
	return events
}

// Size returns the number of items in the queue
func (q *FileChangeQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Engine is the main backup engine
type Engine struct {
	client      *client.Client
	config      *config.ClientConfig
	watcher     *Watcher
	compressor  *Compressor
	uploader    *client.Uploader
	changeQueue *FileChangeQueue
	stats       BackupStats
	statsMu     sync.RWMutex

	// Control channels
	stopChan  chan struct{}
	pauseChan chan bool
	running   bool
	paused    bool
	mu        sync.RWMutex

	// Configuration
	backupInterval  time.Duration
	debounceTime    time.Duration
	excludePatterns []string
	compressionType CompressionType
}

// NewEngine creates a new backup engine
func NewEngine(c *client.Client, cfg *config.ClientConfig) (*Engine, error) {
	// Create watcher
	watcher, err := NewWatcher(true) // Recursive watching
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	// Set up exclude filter
	excludePatterns := cfg.Backup.ExcludePatterns
	watcher.SetExcludeFilter(func(path string) bool {
		return shouldExclude(path, excludePatterns)
	})

	// Create compressor
	compressionType := CompressionZstd // Default to zstd
	compressor, err := NewCompressor(compressionType)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressor: %w", err)
	}

	// Create uploader
	uploader := client.NewUploader(c)

	engine := &Engine{
		client:          c,
		config:          cfg,
		watcher:         watcher,
		compressor:      compressor,
		uploader:        uploader,
		changeQueue:     NewFileChangeQueue(),
		stopChan:        make(chan struct{}),
		pauseChan:       make(chan bool),
		backupInterval:  time.Duration(cfg.Backup.Schedule.CheckIntervalMinutes) * time.Minute,
		debounceTime:    5 * time.Second, // Wait 5 seconds after last change before backing up
		excludePatterns: excludePatterns,
		compressionType: compressionType,
	}

	engine.stats.CurrentState = StateIdle
	return engine, nil
}

// Start starts the backup engine
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("engine already running")
	}
	e.running = true
	e.mu.Unlock()

	log.Println("Starting backup engine...")

	// Add watched directories
	for _, folder := range e.config.Backup.Folders {
		log.Printf("Adding watch for: %s", folder)
		if err := e.watcher.AddDirectory(folder); err != nil {
			log.Printf("Warning: failed to watch %s: %v", folder, err)
		}
	}

	// Start watcher
	if err := e.watcher.Start(); err != nil {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Start worker goroutines
	go e.fileEventProcessor(ctx)
	go e.backupWorker(ctx)

	log.Println("Backup engine started successfully")
	return nil
}

// Stop stops the backup engine
func (e *Engine) Stop() error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = false
	e.mu.Unlock()

	log.Println("Stopping backup engine...")

	close(e.stopChan)

	if err := e.watcher.Stop(); err != nil {
		log.Printf("Warning: error stopping watcher: %v", err)
	}

	if err := e.compressor.Close(); err != nil {
		log.Printf("Warning: error closing compressor: %v", err)
	}

	log.Println("Backup engine stopped")
	return nil
}

// Pause pauses the backup engine
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.paused {
		e.paused = true
		e.pauseChan <- true
		e.updateState(StatePaused)
		log.Println("Backup engine paused")
	}
}

// Resume resumes the backup engine
func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.paused {
		e.paused = false
		e.pauseChan <- false
		e.updateState(StateIdle)
		log.Println("Backup engine resumed")
	}
}

// fileEventProcessor processes file system events
func (e *Engine) fileEventProcessor(ctx context.Context) {
	log.Println("File event processor started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return

		case event := <-e.watcher.Events():
			// Ignore delete and rename events for now
			if event.Operation == "delete" || event.Operation == "rename" {
				continue
			}

			// Check if file should be backed up
			if shouldExclude(event.Path, e.excludePatterns) {
				continue
			}

			// Check if it's a directory
			info, err := os.Stat(event.Path)
			if err != nil {
				continue
			}

			if info.IsDir() {
				continue
			}

			log.Printf("File changed: %s (%s)", event.Path, event.Operation)
			e.changeQueue.Add(event)

		case err := <-e.watcher.Errors():
			log.Printf("Watcher error: %v", err)
			e.setError(err)
		}
	}
}

// backupWorker performs the actual backups
func (e *Engine) backupWorker(ctx context.Context) {
	log.Println("Backup worker started")

	ticker := time.NewTicker(e.backupInterval)
	defer ticker.Stop()

	debounceTimer := time.NewTimer(e.debounceTime)
	debounceTimer.Stop()

	var debouncePending bool

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return

		case paused := <-e.pauseChan:
			if paused {
				debounceTimer.Stop()
				debouncePending = false
			}

		case <-ticker.C:
			// Periodic backup check
			if !e.isPaused() && e.changeQueue.Size() > 0 {
				e.performBackup(ctx)
			}

		case <-debounceTimer.C:
			// Debounce timer expired, perform backup
			debouncePending = false
			if !e.isPaused() {
				e.performBackup(ctx)
			}
		}

		// Check if we need to start debounce timer
		if !debouncePending && e.changeQueue.Size() > 0 && !e.isPaused() {
			debounceTimer.Reset(e.debounceTime)
			debouncePending = true
		}
	}
}

// performBackup performs the actual backup of queued files
func (e *Engine) performBackup(ctx context.Context) {
	e.updateState(StateBacking)

	events := e.changeQueue.GetAll()
	if len(events) == 0 {
		e.updateState(StateIdle)
		return
	}

	log.Printf("Starting backup of %d files...", len(events))
	startTime := time.Now()

	uploaded := 0
	failed := 0
	var totalBytes int64

	for _, event := range events {
		if err := e.backupFile(ctx, event.Path); err != nil {
			log.Printf("Failed to backup %s: %v", event.Path, err)
			failed++
			e.incrementFailed()
		} else {
			info, err := os.Stat(event.Path)
			if err == nil {
				totalBytes += info.Size()
			}
			uploaded++
			e.incrementUploaded()
		}
	}

	duration := time.Since(startTime)
	log.Printf("Backup completed: %d uploaded, %d failed in %s", uploaded, failed, duration)

	e.statsMu.Lock()
	e.stats.LastBackupTime = time.Now()
	e.stats.BytesUploaded += totalBytes
	e.statsMu.Unlock()

	e.updateState(StateIdle)
}

// backupFile backs up a single file
func (e *Engine) backupFile(ctx context.Context, localPath string) error {
	// Calculate relative path from watched folder
	var remotePath string
	for _, watchedFolder := range e.config.Backup.Folders {
		if rel, err := filepath.Rel(watchedFolder, localPath); err == nil && !filepath.IsAbs(rel) {
			remotePath = filepath.ToSlash(rel)
			break
		}
	}

	if remotePath == "" {
		return fmt.Errorf("file not in any watched folder")
	}

	log.Printf("Backing up: %s -> %s", localPath, remotePath)
	return e.uploader.UploadFile(localPath, remotePath)
}

// GetStats returns current backup statistics
func (e *Engine) GetStats() BackupStats {
	e.statsMu.RLock()
	defer e.statsMu.RUnlock()
	return e.stats
}

// updateState updates the current state
func (e *Engine) updateState(state BackupState) {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()
	e.stats.CurrentState = state
}

// setError sets the last error
func (e *Engine) setError(err error) {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()
	e.stats.LastError = err
	e.stats.CurrentState = StateError
}

// incrementUploaded increments the uploaded counter
func (e *Engine) incrementUploaded() {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()
	e.stats.FilesUploaded++
}

// incrementFailed increments the failed counter
func (e *Engine) incrementFailed() {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()
	e.stats.FilesFailed++
}

// isPaused returns whether the engine is paused
func (e *Engine) isPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

// shouldExclude checks if a path should be excluded based on patterns
func shouldExclude(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple wildcard matching
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}

		// Check if path contains the pattern
		if filepath.Base(path) == pattern {
			return true
		}
	}
	return false
}

// PerformInitialScan performs an initial scan of all watched folders
func (e *Engine) PerformInitialScan(ctx context.Context) error {
	e.updateState(StateScanning)
	log.Println("Starting initial scan...")

	var filesFound int64

	for _, folder := range e.config.Backup.Folders {
		log.Printf("Scanning folder: %s", folder)

		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check exclusions
			if shouldExclude(path, e.excludePatterns) {
				return nil
			}

			filesFound++

			// Add to queue for backup
			e.changeQueue.Add(FileEvent{
				Path:      path,
				Operation: "create",
				Timestamp: time.Now(),
			})

			return nil
		})

		if err != nil {
			log.Printf("Error scanning %s: %v", folder, err)
		}
	}

	e.statsMu.Lock()
	e.stats.FilesScanned = filesFound
	e.statsMu.Unlock()

	log.Printf("Initial scan complete: found %d files", filesFound)
	e.updateState(StateIdle)

	return nil
}
