//go:build darwin
// +build darwin

package client

import (
    "fmt"

    "github.com/zalando/go-keyring"
)

const keyringService = "zfscloud"

// A keyring failure must not save a raw passphrase to a local file.
// Legacy fallback files remain untouched for explicit safe migration.
func storePassphraseInKeychain(clientID string, passphrase []byte) error {
    if err := keyring.Set(keyringService, clientID, string(passphrase)); err != nil {
        return fmt.Errorf("macOS keychain unavailable; refusing plaintext fallback: %w", err)
    }
    return nil
}

func loadPassphraseFromKeychain(clientID string) ([]byte, error) {
    value, err := keyring.Get(keyringService, clientID)
    if err != nil {
        return nil, fmt.Errorf("macOS keychain unavailable or passphrase missing; legacy plaintext fallback disabled: %w", err)
    }
    return []byte(value), nil
}

func deletePassphraseFromKeychain(clientID string) error {
    if err := keyring.Delete(keyringService, clientID); err != nil {
        return fmt.Errorf("macOS keychain deletion failed: %w", err)
    }
    return nil
}

func hasStoredPassphrase(clientID string) bool {
    _, err := keyring.Get(keyringService, clientID)
    return err == nil
}
