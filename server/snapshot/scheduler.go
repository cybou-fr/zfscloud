package snapshot

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/server/zfs"
	"github.com/ioustamora/zfscloud/shared/config"
)

// SnapshotType represents the type of snapshot
type SnapshotType string

const (
	SnapshotHourly  SnapshotType = "hourly"
	SnapshotDaily   SnapshotType = "daily"
	SnapshotWeekly  SnapshotType = "weekly"
	SnapshotMonthly SnapshotType = "monthly"
)

// Scheduler manages automatic snapshot creation and cleanup
type Scheduler struct {
	database   *db.Database
	zfsManager *zfs.Manager
	config     *config.ServerConfig

	// Control channels
	stopChan chan struct{}
	running  bool
	mu       sync.RWMutex
}

// NewScheduler creates a new snapshot scheduler
func NewScheduler(database *db.Database, zfsManager *zfs.Manager, cfg *config.ServerConfig) *Scheduler {
	return &Scheduler{
		database:   database,
		zfsManager: zfsManager,
		config:     cfg,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the snapshot scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	s.running = true
	s.mu.Unlock()

	log.Println("Starting snapshot scheduler...")

	// Start scheduler goroutines
	go s.runHourlyScheduler(ctx)
	go s.runDailyScheduler(ctx)
	go s.runWeeklyScheduler(ctx)
	go s.runMonthlyScheduler(ctx)

	log.Println("Snapshot scheduler started successfully")
	return nil
}

// Stop stops the snapshot scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	log.Println("Stopping snapshot scheduler...")
	close(s.stopChan)
	log.Println("Snapshot scheduler stopped")
	return nil
}

// runHourlyScheduler runs hourly snapshot creation
func (s *Scheduler) runHourlyScheduler(ctx context.Context) {
	// Run at the top of every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run immediately on start
	s.createSnapshotsForAllUsers(SnapshotHourly)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.createSnapshotsForAllUsers(SnapshotHourly)
			s.enforceRetentionPolicy(SnapshotHourly)
		}
	}
}

// runDailyScheduler runs daily snapshot creation
func (s *Scheduler) runDailyScheduler(ctx context.Context) {
	// Run at midnight every day
	ticker := s.createDailyTicker()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.createSnapshotsForAllUsers(SnapshotDaily)
			s.enforceRetentionPolicy(SnapshotDaily)
			ticker = s.createDailyTicker() // Reset ticker for next day
		}
	}
}

// runWeeklyScheduler runs weekly snapshot creation
func (s *Scheduler) runWeeklyScheduler(ctx context.Context) {
	// Run on Sunday at midnight
	ticker := s.createWeeklyTicker()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.createSnapshotsForAllUsers(SnapshotWeekly)
			s.enforceRetentionPolicy(SnapshotWeekly)
			ticker = s.createWeeklyTicker() // Reset ticker for next week
		}
	}
}

// runMonthlyScheduler runs monthly snapshot creation
func (s *Scheduler) runMonthlyScheduler(ctx context.Context) {
	// Run on the 1st of each month at midnight
	ticker := s.createMonthlyTicker()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.createSnapshotsForAllUsers(SnapshotMonthly)
			s.enforceRetentionPolicy(SnapshotMonthly)
			ticker = s.createMonthlyTicker() // Reset ticker for next month
		}
	}
}

// createDailyTicker creates a ticker that fires at midnight
func (s *Scheduler) createDailyTicker() *time.Ticker {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	duration := nextMidnight.Sub(now)
	return time.NewTicker(duration)
}

// createWeeklyTicker creates a ticker that fires on Sunday at midnight
func (s *Scheduler) createWeeklyTicker() *time.Ticker {
	now := time.Now()
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 {
		daysUntilSunday = 7
	}
	nextSunday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, 0, 0, 0, 0, now.Location())
	duration := nextSunday.Sub(now)
	return time.NewTicker(duration)
}

// createMonthlyTicker creates a ticker that fires on the 1st of each month
func (s *Scheduler) createMonthlyTicker() *time.Ticker {
	now := time.Now()
	var nextFirst time.Time
	if now.Day() == 1 && now.Hour() == 0 && now.Minute() == 0 {
		// If it's already the 1st at midnight, schedule for next month
		nextFirst = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	} else {
		nextFirst = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	}
	duration := nextFirst.Sub(now)
	return time.NewTicker(duration)
}

// createSnapshotsForAllUsers creates snapshots for all active users
func (s *Scheduler) createSnapshotsForAllUsers(snapshotType SnapshotType) {
	log.Printf("Creating %s snapshots for all users...", snapshotType)

	// Get all active users
	users, err := s.database.ListUsers(false) // false = only active users
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return
	}

	successCount := 0
	failCount := 0

	for _, user := range users {
		if err := s.CreateSnapshot(user.UserID, snapshotType); err != nil {
			log.Printf("Error creating %s snapshot for user %s: %v", snapshotType, user.UserID, err)
			failCount++
		} else {
			successCount++
		}
	}

	log.Printf("Completed %s snapshots: %d succeeded, %d failed", snapshotType, successCount, failCount)
}

// CreateSnapshot creates a snapshot for a specific user
func (s *Scheduler) CreateSnapshot(userID string, snapshotType SnapshotType) error {
	// Get user info
	user, err := s.database.GetUserByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Generate snapshot name: {type}-{timestamp}
	timestamp := time.Now().UTC().Format("20060102-150405")
	snapshotName := fmt.Sprintf("%s-%s", snapshotType, timestamp)
	fullSnapshotName := fmt.Sprintf("%s@%s", user.ZFSDataset, snapshotName)

	// Create ZFS snapshot
	if err := s.zfsManager.CreateSnapshot(user.ZFSDataset, snapshotName); err != nil {
		return fmt.Errorf("failed to create ZFS snapshot: %w", err)
	}

	// Get snapshot size
	usedSpace, err := s.zfsManager.GetUsedSpace(fullSnapshotName)
	if err != nil {
		log.Printf("Warning: failed to get snapshot size: %v", err)
		usedSpace = 0
	}

	// Record in database
	snapshot := &db.Snapshot{
		UserID:          userID,
		ZFSSnapshotName: fullSnapshotName,
		SizeBytes:       usedSpace,
	}

	if err := s.database.CreateSnapshot(snapshot); err != nil {
		// Try to clean up ZFS snapshot if database insert fails
		if delErr := s.zfsManager.DeleteSnapshot(fullSnapshotName); delErr != nil {
			log.Printf("Error deleting snapshot after database record failed for user %s: %v", userID, delErr)
		}
		return fmt.Errorf("failed to record snapshot in database: %w", err)
	}

	log.Printf("Created %s snapshot for user %s: %s (%d bytes)", snapshotType, userID, fullSnapshotName, usedSpace)

	// Log action
	s.database.LogAction(userID, "", "snapshot_created",
		fmt.Sprintf("Type: %s, Name: %s, Size: %d", snapshotType, fullSnapshotName, usedSpace), "")

	return nil
}

// enforceRetentionPolicy deletes old snapshots based on retention policy
func (s *Scheduler) enforceRetentionPolicy(snapshotType SnapshotType) {
	log.Printf("Enforcing retention policy for %s snapshots...", snapshotType)

	// Get retention count for this snapshot type
	retentionCount := s.getRetentionCount(snapshotType)
	if retentionCount <= 0 {
		log.Printf("Retention policy disabled for %s snapshots", snapshotType)
		return
	}

	// Get all users
	users, err := s.database.ListUsers(true) // true = include revoked
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return
	}

	totalDeleted := 0

	for _, user := range users {
		deleted, err := s.cleanupUserSnapshots(user.UserID, snapshotType, retentionCount)
		if err != nil {
			log.Printf("Error cleaning up %s snapshots for user %s: %v", snapshotType, user.UserID, err)
		} else {
			totalDeleted += deleted
		}
	}

	if totalDeleted > 0 {
		log.Printf("Deleted %d old %s snapshots", totalDeleted, snapshotType)
	}
}

// cleanupUserSnapshots removes old snapshots for a user based on retention policy
func (s *Scheduler) cleanupUserSnapshots(userID string, snapshotType SnapshotType, retentionCount int) (int, error) {
	// Get all snapshots for this user
	allSnapshots, err := s.database.ListSnapshotsByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Filter snapshots by type
	var typeSnapshots []*db.Snapshot
	typePrefix := fmt.Sprintf("@%s-", snapshotType)
	for _, snap := range allSnapshots {
		if strings.Contains(snap.ZFSSnapshotName, typePrefix) {
			typeSnapshots = append(typeSnapshots, snap)
		}
	}

	// If we have fewer or equal to retention count, nothing to delete
	if len(typeSnapshots) <= retentionCount {
		return 0, nil
	}

	// Sort by creation time (oldest first)
	sort.Slice(typeSnapshots, func(i, j int) bool {
		return typeSnapshots[i].CreatedAt.Before(typeSnapshots[j].CreatedAt)
	})

	// Delete oldest snapshots beyond retention count
	deleteCount := len(typeSnapshots) - retentionCount
	deletedCount := 0

	for i := 0; i < deleteCount; i++ {
		snap := typeSnapshots[i]

		// Delete from ZFS
		if err := s.zfsManager.DeleteSnapshot(snap.ZFSSnapshotName); err != nil {
			log.Printf("Warning: failed to delete ZFS snapshot %s: %v", snap.ZFSSnapshotName, err)
			// Continue anyway to clean up database
		}

		// Delete from database
		if err := s.database.DeleteSnapshot(snap.SnapshotID); err != nil {
			log.Printf("Warning: failed to delete snapshot record %s: %v", snap.SnapshotID, err)
		} else {
			deletedCount++
			log.Printf("Deleted old %s snapshot: %s", snapshotType, snap.ZFSSnapshotName)
		}

		// Log action
		s.database.LogAction(userID, "", "snapshot_deleted",
			fmt.Sprintf("Type: %s, Name: %s (retention policy)", snapshotType, snap.ZFSSnapshotName), "")
	}

	return deletedCount, nil
}

// getRetentionCount returns the retention count for a snapshot type
func (s *Scheduler) getRetentionCount(snapshotType SnapshotType) int {
	switch snapshotType {
	case SnapshotHourly:
		return s.config.Snapshots.Schedule.Hourly
	case SnapshotDaily:
		return s.config.Snapshots.Schedule.Daily
	case SnapshotWeekly:
		return s.config.Snapshots.Schedule.Weekly
	case SnapshotMonthly:
		return s.config.Snapshots.Schedule.Monthly
	default:
		return 0
	}
}

// GetStatus returns the current status of the scheduler
func (s *Scheduler) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"running": s.running,
		"retention_policy": map[string]int{
			"hourly":  s.config.Snapshots.Schedule.Hourly,
			"daily":   s.config.Snapshots.Schedule.Daily,
			"weekly":  s.config.Snapshots.Schedule.Weekly,
			"monthly": s.config.Snapshots.Schedule.Monthly,
		},
	}
}
