//go:build linux
// +build linux

package client

import (
	"fmt"
)

// storePassphraseInKeychain stores the passphrase securely using Linux Secret Service
// TODO: Implement using libsecret or gnome-keyring
func storePassphraseInKeychain(clientID string, passphrase []byte) error {
	// For now, fall back to file storage with warning
	fmt.Println("⚠️  WARNING: Linux keychain integration not yet implemented")
	fmt.Println("⚠️  Passphrase will be stored in encrypted file instead")

	// Use file-based storage as fallback
	// In production, this should use libsecret or gnome-keyring
	return storeEncryptedDataToFile(clientID, passphrase)
}

// loadPassphraseFromKeychain loads the passphrase from Linux Secret Service
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
