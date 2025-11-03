//go:build darwin
// +build darwin

package client

import (
	"fmt"
)

// storePassphraseInKeychain stores the passphrase securely using macOS Keychain
// TODO: Implement using macOS Keychain Access API
func storePassphraseInKeychain(clientID string, passphrase []byte) error {
	// For now, fall back to file storage with warning
	fmt.Println("⚠️  WARNING: macOS Keychain integration not yet implemented")
	fmt.Println("⚠️  Passphrase will be stored in encrypted file instead")

	// Use file-based storage as fallback
	// In production, this should use macOS Keychain Access
	return storeEncryptedDataToFile(clientID, passphrase)
}

// loadPassphraseFromKeychain loads the passphrase from macOS Keychain
func loadPassphraseFromKeychain(clientID string) ([]byte, error) {
	// For now, fall back to file storage
	return loadEncryptedDataFromFile(clientID)
}

// deletePassphraseFromKeychain removes the stored passphrase
func deletePassphraseFromKeychain(clientID string) error {
	return deleteEncryptedDataFile(clientID)
}

// hasStoredPassphrase checks if a passphrase is stored for the client ID
func hasStoredPassphrase(clientID string) bool {
	_, err := loadEncryptedDataFromFile(clientID)
	return err == nil
}
