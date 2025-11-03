package snapshot

import (
	"fmt"
	"time"

	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/server/zfs"
)

// CreateManualSnapshot creates a manual snapshot for a specific user
// This is a convenience function that doesn't require a running scheduler
func CreateManualSnapshot(database *db.Database, zfsManager *zfs.Manager, userID string) (string, int64, error) {
	// Get user info
	user, err := database.GetUserByUserID(userID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate snapshot name: manual-{timestamp}
	timestamp := time.Now().UTC().Format("20060102-150405")
	snapshotName := fmt.Sprintf("manual-%s", timestamp)
	fullSnapshotName := fmt.Sprintf("%s@%s", user.ZFSDataset, snapshotName)

	// Create ZFS snapshot
	if err := zfsManager.CreateSnapshot(user.ZFSDataset, snapshotName); err != nil {
		return "", 0, fmt.Errorf("failed to create ZFS snapshot: %w", err)
	}

	// Get snapshot size
	usedSpace, err := zfsManager.GetUsedSpace(fullSnapshotName)
	if err != nil {
		// Continue even if we can't get size
		usedSpace = 0
	}

	// Record in database
	snapshot := &db.Snapshot{
		UserID:          userID,
		ZFSSnapshotName: fullSnapshotName,
		SizeBytes:       usedSpace,
	}

	if err := database.CreateSnapshot(snapshot); err != nil {
		// Try to clean up ZFS snapshot if database insert fails
		zfsManager.DeleteSnapshot(fullSnapshotName)
		return "", 0, fmt.Errorf("failed to record snapshot in database: %w", err)
	}

	// Log action
	database.LogAction(userID, "", "snapshot_created_manual",
		fmt.Sprintf("Manual snapshot: %s, Size: %d", fullSnapshotName, usedSpace), "")

	return fullSnapshotName, usedSpace, nil
}
