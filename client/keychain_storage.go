package client

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// getKeychainDir returns the directory for storing encrypted passphrases
func getKeychainDir() (string, error) {
	// Use user's AppData\Local\zfscloud directory
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		// Fallback to user profile
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			return "", fmt.Errorf("cannot determine user profile directory")
		}
		appData = filepath.Join(userProfile, "AppData", "Local")
	}

	keychainDir := filepath.Join(appData, "zfscloud", "keys")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(keychainDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create keychain directory: %w", err)
	}

	return keychainDir, nil
}

// storeEncryptedDataToFile stores encrypted passphrase data to a file
func storeEncryptedDataToFile(clientID string, encryptedData []byte) error {
	keychainDir, err := getKeychainDir()
	if err != nil {
		return err
	}

	// Use base64-encoded client ID as filename (safe for filesystem)
	filename := base64.URLEncoding.EncodeToString([]byte(clientID)) + ".key"
	filepath := filepath.Join(keychainDir, filename)

	// Write encrypted data
	if err := os.WriteFile(filepath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write keychain file: %w", err)
	}

	return nil
}

// loadEncryptedDataFromFile loads encrypted passphrase data from a file
func loadEncryptedDataFromFile(clientID string) ([]byte, error) {
	keychainDir, err := getKeychainDir()
	if err != nil {
		return nil, err
	}

	// Use base64-encoded client ID as filename
	filename := base64.URLEncoding.EncodeToString([]byte(clientID)) + ".key"
	filepath := filepath.Join(keychainDir, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("passphrase not found in keychain")
	}

	// Read encrypted data
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keychain file: %w", err)
	}

	return data, nil
}

// deleteEncryptedDataFile removes the encrypted passphrase file
func deleteEncryptedDataFile(clientID string) error {
	keychainDir, err := getKeychainDir()
	if err != nil {
		return err
	}

	// Use base64-encoded client ID as filename
	filename := base64.URLEncoding.EncodeToString([]byte(clientID)) + ".key"
	filepath := filepath.Join(keychainDir, filename)

	// Remove file
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete keychain file: %w", err)
	}

	return nil
}

// hasStoredPassphrase checks if a passphrase is stored for the client ID
func hasStoredPassphrase(clientID string) bool {
	_, err := loadEncryptedDataFromFile(clientID)
	return err == nil
}
