package snapshot

import (
	"fmt"
	"time"

	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/server/zfs"
)

// CreateManualSnapshot creates a manual snapshot for a user
func CreateManualSnapshot(database *db.Database, zfsManager *zfs.Manager, userID string) (string, int64, error) {
	user, err := database.GetUserByUserID(userID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get user: %w", err)
	}

	if user.ZFSDataset == "" {
		return "", 0, fmt.Errorf("user does not have a ZFS dataset")
	}

	snapshotName := fmt.Sprintf("manual-%s", time.Now().UTC().Format("2006-01-02-150405"))

	if err := zfsManager.CreateSnapshot(user.ZFSDataset, snapshotName); err != nil {
		return "", 0, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Get snapshot size (this is an estimate)
	sizeBytes, err := zfsManager.GetUsedSpace(fmt.Sprintf("%s@%s", user.ZFSDataset, snapshotName))
	if err != nil {
		// Don't fail the whole operation if we can't get the size
		fmt.Printf("WARNING: Failed to get snapshot size: %v\n", err)
	}

	// Log the snapshot creation
	if err := database.LogAction(userID, "", "snapshot_created_manual", fmt.Sprintf("Manual snapshot '%s' created", snapshotName), ""); err != nil {
		// If logging fails, we should probably delete the snapshot to avoid inconsistency
		if delErr := zfsManager.DeleteSnapshot(fmt.Sprintf("%s@%s", user.ZFSDataset, snapshotName)); delErr != nil {
			return "", 0, fmt.Errorf("failed to log action and failed to delete snapshot: %w, original error: %v", delErr, err)
		}
		return "", 0, fmt.Errorf("failed to log action: %w", err)
	}

	return snapshotName, sizeBytes, nil
}

// DeleteManualSnapshot deletes a manual snapshot for a user
func DeleteManualSnapshot(userID, snapshotName string, zfsManager *zfs.Manager, database *db.Database) error {
	user, err := database.GetUserByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.ZFSDataset == "" {
		return fmt.Errorf("user does not have a ZFS dataset")
	}

	fullSnapshotName := fmt.Sprintf("%s@%s", user.ZFSDataset, snapshotName)

	if err := zfsManager.DeleteSnapshot(fullSnapshotName); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	// Log the snapshot deletion
	if err := database.LogAction(userID, "", "snapshot_deleted_manual", fmt.Sprintf("Manual snapshot '%s' deleted", snapshotName), ""); err != nil {
		return fmt.Errorf("failed to log action: %w", err)
	}

	return nil
}
