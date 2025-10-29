package backup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	Operation string // "create", "modify", "delete", "rename"
	Timestamp time.Time
}

// Watcher watches filesystem changes
type Watcher struct {
	fsWatcher     *fsnotify.Watcher
	watchedDirs   map[string]bool
	eventChannel  chan FileEvent
	errorChannel  chan error
	stopChannel   chan struct{}
	mu            sync.RWMutex
	running       bool
	recursive     bool
	excludeFilter func(string) bool
}

// NewWatcher creates a new filesystem watcher
func NewWatcher(recursive bool) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		fsWatcher:    fsWatcher,
		watchedDirs:  make(map[string]bool),
		eventChannel: make(chan FileEvent, 100),
		errorChannel: make(chan error, 10),
		stopChannel:  make(chan struct{}),
		recursive:    recursive,
	}, nil
}

// SetExcludeFilter sets a function to filter excluded paths
func (w *Watcher) SetExcludeFilter(filter func(string) bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.excludeFilter = filter
}

// AddDirectory adds a directory to watch
func (w *Watcher) AddDirectory(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Add the directory itself
	if err := w.fsWatcher.Add(absPath); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", absPath, err)
	}
	w.watchedDirs[absPath] = true

	// If recursive, walk and add all subdirectories
	if w.recursive {
		if err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				// Check if excluded
				if w.excludeFilter != nil && w.excludeFilter(path) {
					return filepath.SkipDir
				}

				if path != absPath {
					if err := w.fsWatcher.Add(path); err != nil {
						log.Printf("Warning: failed to watch %s: %v", path, err)
					} else {
						w.watchedDirs[path] = true
					}
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to walk directory tree: %w", err)
		}
	}

	log.Printf("Watching %d directories", len(w.watchedDirs))
	return nil
}

// RemoveDirectory removes a directory from watching
func (w *Watcher) RemoveDirectory(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := w.fsWatcher.Remove(absPath); err != nil {
		return fmt.Errorf("failed to remove watch on %s: %w", absPath, err)
	}

	delete(w.watchedDirs, absPath)
	return nil
}

// Start starts watching for file changes
func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watcher already running")
	}
	w.running = true
	w.mu.Unlock()

	go w.watchLoop()
	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = false
	w.mu.Unlock()

	close(w.stopChannel)
	return w.fsWatcher.Close()
}

// Events returns the channel for file events
func (w *Watcher) Events() <-chan FileEvent {
	return w.eventChannel
}

// Errors returns the channel for errors
func (w *Watcher) Errors() <-chan error {
	return w.errorChannel
}

// watchLoop is the main event processing loop
func (w *Watcher) watchLoop() {
	defer close(w.eventChannel)
	defer close(w.errorChannel)

	for {
		select {
		case <-w.stopChannel:
			log.Println("Watcher stopped")
			return

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			// Check if excluded
			if w.excludeFilter != nil && w.excludeFilter(event.Name) {
				continue
			}

			// Convert fsnotify event to our FileEvent
			fileEvent := w.convertEvent(event)
			if fileEvent != nil {
				w.eventChannel <- *fileEvent

				// If a new directory was created and we're in recursive mode, watch it
				if w.recursive && event.Op&fsnotify.Create != 0 {
					w.handleNewDirectory(event.Name)
				}
			}

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			w.errorChannel <- err
		}
	}
}

// convertEvent converts fsnotify.Event to FileEvent
func (w *Watcher) convertEvent(event fsnotify.Event) *FileEvent {
	var operation string

	switch {
	case event.Op&fsnotify.Create != 0:
		operation = "create"
	case event.Op&fsnotify.Write != 0:
		operation = "modify"
	case event.Op&fsnotify.Remove != 0:
		operation = "delete"
	case event.Op&fsnotify.Rename != 0:
		operation = "rename"
	case event.Op&fsnotify.Chmod != 0:
		// Ignore chmod events
		return nil
	default:
		return nil
	}

	return &FileEvent{
		Path:      event.Name,
		Operation: operation,
		Timestamp: time.Now(),
	}
}

// handleNewDirectory watches a newly created directory
func (w *Watcher) handleNewDirectory(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, watched := w.watchedDirs[path]; !watched {
		if err := w.fsWatcher.Add(path); err != nil {
			log.Printf("Warning: failed to watch new directory %s: %v", path, err)
		} else {
			w.watchedDirs[path] = true
			log.Printf("Started watching new directory: %s", path)
		}
	}
}

// IsWatching returns true if a path is being watched
func (w *Watcher) IsWatching(path string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	return w.watchedDirs[absPath]
}

// GetWatchedDirectories returns a list of all watched directories
func (w *Watcher) GetWatchedDirectories() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	dirs := make([]string, 0, len(w.watchedDirs))
	for dir := range w.watchedDirs {
		dirs = append(dirs, dir)
	}
	return dirs
}
