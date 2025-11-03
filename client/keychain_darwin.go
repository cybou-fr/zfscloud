//go:build darwin
// +build darwin

package client

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "zfscloud"
)

// storePassphraseInKeychain stores the passphrase securely using macOS Keychain
func storePassphraseInKeychain(clientID string, passphrase []byte) error {
	err := keyring.Set(keyringService, clientID, string(passphrase))
	if err != nil {
		fmt.Printf("⚠️  WARNING: Could not store passphrase in macOS keychain: %v\n", err)
		fmt.Println("⚠️  Passphrase will be stored in encrypted file instead")
		return storeEncryptedDataToFile(clientID, passphrase)
	}
	return nil
}

// loadPassphraseFromKeychain loads the passphrase from macOS Keychain
func loadPassphraseFromKeychain(clientID string) ([]byte, error) {
	secret, err := keyring.Get(keyringService, clientID)
	if err != nil {
		fmt.Printf("⚠️  WARNING: Could not load passphrase from macOS keychain: %v\n", err)
		fmt.Println("⚠️  Attempting to load from encrypted file instead")
		return loadEncryptedDataFromFile(clientID)
	}
	return []byte(secret), nil
}

// deletePassphraseFromKeychain removes the stored passphrase
func deletePassphraseFromKeychain(clientID string) error {
	err := keyring.Delete(keyringService, clientID)
	if err != nil {
		fmt.Printf("⚠️  WARNING: Could not delete passphrase from macOS keychain: %v\n", err)
		fmt.Println("⚠️  Attempting to delete from encrypted file instead")
		return deleteEncryptedDataFile(clientID)
	}
	return nil
}

// hasStoredPassphrase checks if a passphrase is stored for the client ID
func hasStoredPassphrase(clientID string) bool {
	_, err := keyring.Get(keyringService, clientID)
	if err == nil {
		return true
	}
	// If not in keyring, check file as fallback
	_, err = loadEncryptedDataFromFile(clientID)
	return err == nil
}
