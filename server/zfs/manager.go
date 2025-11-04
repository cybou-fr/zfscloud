package zfs

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager handles ZFS operations
type Manager struct {
	pool string
}

// NewManager creates a new ZFS manager
func NewManager(pool string) *Manager {
	return &Manager{
		pool: pool,
	}
}

// CreateDataset creates a new encrypted ZFS dataset for a user with a passphrase
// The passphrase is provided by the user/client and used to encrypt the dataset
func (m *Manager) CreateDataset(datasetName string, passphrase []byte, encryption bool, compression string) error {
	args := []string{"create"}
	if encryption {
		args = append(args, "-o", "encryption=aes-256-gcm", "-o", "keyformat=passphrase", "-o", "keylocation=prompt")
	}
	if compression != "" {
		args = append(args, "-o", fmt.Sprintf("compression=%s", compression))
	}
	args = append(args, datasetName)

	cmd := exec.Command("zfs", args...)

	// Provide passphrase via stdin if encryption is enabled
	if encryption {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdin pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start zfs create: %w", err)
		}

		// Write passphrase twice (ZFS requires confirmation)
		if _, err := stdin.Write(append(passphrase, '\n')); err != nil {
			return fmt.Errorf("failed to write passphrase: %w", err)
		}
		if _, err := stdin.Write(append(passphrase, '\n')); err != nil {
			return fmt.Errorf("failed to write passphrase confirmation: %w", err)
		}
		stdin.Close()

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}
	} else {
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create dataset: %w, output: %s", err, string(output))
		}
	}

	return nil
}

// DestroyDataset destroys a ZFS dataset
func (m *Manager) DestroyDataset(datasetName string, recursive bool) error {
	args := []string{"destroy"}
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, datasetName)

	cmd := exec.Command("zfs", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to destroy dataset: %w, output: %s", err, string(output))
	}
	return nil
}

// LoadKey loads the encryption key for a dataset
func (m *Manager) LoadKey(datasetName string, passphrase []byte) error {
	cmd := exec.Command("zfs", "load-key", datasetName)

	// Write passphrase to stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start load-key command: %w", err)
	}

	if _, err := stdin.Write(passphrase); err != nil {
		return fmt.Errorf("failed to write passphrase: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}

	return nil
}

// UnloadKey unloads the encryption key for a dataset
func (m *Manager) UnloadKey(datasetName string) error {
	cmd := exec.Command("zfs", "unload-key", datasetName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to unload key: %w, output: %s", err, string(output))
	}
	return nil
}

// CreateSnapshot creates a new snapshot for a dataset
func (m *Manager) CreateSnapshot(datasetName, snapshotName string) error {
	fullName := fmt.Sprintf("%s@%s", datasetName, snapshotName)
	cmd := exec.Command("zfs", "snapshot", fullName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create snapshot: %w, output: %s", err, string(output))
	}
	return nil
}

// ListSnapshots lists all snapshots for a dataset
func (m *Manager) ListSnapshots(datasetName string) ([]string, error) {
	cmd := exec.Command("zfs", "list", "-t", "snapshot", "-H", "-o", "name", "-r", datasetName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w, output: %s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var snapshots []string
	for _, line := range lines {
		if line != "" {
			snapshots = append(snapshots, line)
		}
	}

	return snapshots, nil
}

// DeleteSnapshot deletes a snapshot
func (m *Manager) DeleteSnapshot(snapshotName string) error {
	cmd := exec.Command("zfs", "destroy", snapshotName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w, output: %s", err, string(output))
	}
	return nil
}

// SetQuota sets the quota for a dataset
func (m *Manager) SetQuota(datasetName string, quotaBytes int64) error {
	quota := fmt.Sprintf("%d", quotaBytes)
	cmd := exec.Command("zfs", "set", fmt.Sprintf("quota=%s", quota), datasetName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set quota: %w, output: %s", err, string(output))
	}
	return nil
}

// GetUsedSpace returns the used space for a dataset in bytes
func (m *Manager) GetUsedSpace(datasetName string) (int64, error) {
	cmd := exec.Command("zfs", "get", "-H", "-o", "value", "-p", "used", datasetName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get used space: %w, output: %s", err, string(output))
	}

	var used int64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &used); err != nil {
		return 0, fmt.Errorf("failed to parse used space: %w", err)
	}

	return used, nil
}

// GetMountpoint returns the mountpoint for a dataset
func (m *Manager) GetMountpoint(datasetName string) (string, error) {
	cmd := exec.Command("zfs", "get", "-H", "-o", "value", "mountpoint", datasetName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get mountpoint: %w, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// DatasetExists checks if a dataset exists
func (m *Manager) DatasetExists(datasetName string) (bool, error) {
	cmd := exec.Command("zfs", "list", "-H", "-o", "name", datasetName)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check dataset existence: %w", err)
	}
	return true, nil
}

// GetPoolStatus returns the status of the ZFS pool
func (m *Manager) GetPoolStatus() (*PoolStatus, error) {
	cmd := exec.Command("zpool", "list", "-H", "-o", "name,size,alloc,free,health", m.pool)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get pool status: %w, output: %s", err, string(output))
	}

	var name, size, alloc, free, health string
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%s\t%s\t%s\t%s\t%s", &name, &size, &alloc, &free, &health); err != nil {
		return nil, fmt.Errorf("failed to parse pool status: %w", err)
	}

	// Parse sizes (they come in human-readable format like "10G", "500M")
	totalBytes, err := parseSize(size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total size: %w", err)
	}

	usedBytes, err := parseSize(alloc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse used size: %w", err)
	}

	availableBytes, err := parseSize(free)
	if err != nil {
		return nil, fmt.Errorf("failed to parse free size: %w", err)
	}

	// Convert health to numeric value
	var healthValue float64
	switch strings.ToLower(health) {
	case "online":
		healthValue = 1.0
	case "degraded":
		healthValue = 0.7
	case "faulted":
		healthValue = 0.0
	default:
		healthValue = 0.5
	}

	return &PoolStatus{
		Name:           name,
		TotalBytes:     totalBytes,
		UsedBytes:      usedBytes,
		AvailableBytes: availableBytes,
		Health:         healthValue,
	}, nil
}

// PoolStatus represents the status of a ZFS pool
type PoolStatus struct {
	Name           string
	TotalBytes     int64
	UsedBytes      int64
	AvailableBytes int64
	Health         float64
}

// parseSize parses human-readable size strings like "10G", "500M", "1T"
func parseSize(sizeStr string) (int64, error) {
	var size int64
	var unit string

	if _, err := fmt.Sscanf(sizeStr, "%d%s", &size, &unit); err != nil {
		return 0, err
	}

	switch strings.ToUpper(unit) {
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	case "G":
		return size * 1024 * 1024 * 1024, nil
	case "T":
		return size * 1024 * 1024 * 1024 * 1024, nil
	default:
		return size, nil // Assume bytes if no unit
	}
}
