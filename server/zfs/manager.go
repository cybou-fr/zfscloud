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

// CreateDataset creates a new encrypted ZFS dataset for a user
func (m *Manager) CreateDataset(userID string, compression string) (string, error) {
	datasetName := fmt.Sprintf("%s/%s", m.pool, userID)

	args := []string{
		"create",
		"-o", "encryption=aes-256-gcm",
		"-o", "keyformat=passphrase",
		"-o", "keylocation=prompt",
		"-o", fmt.Sprintf("compression=%s", compression),
		datasetName,
	}

	cmd := exec.Command("zfs", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create dataset: %w, output: %s", err, string(output))
	}

	return datasetName, nil
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
