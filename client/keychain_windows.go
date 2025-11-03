//go:build windows
// +build windows

package client

import (
	"crypto/sha256"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	crypt32               = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectData    = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData  = crypt32.NewProc("CryptUnprotectData")
	procLocalFree           = syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

// storePassphraseInKeychain stores the passphrase securely using Windows DPAPI
func storePassphraseInKeychain(clientID string, passphrase []byte) error {
	// Use client ID as part of the entropy for additional security
	entropy := sha256.Sum256([]byte(clientID))

	// Prepare input data
	inBlob := dataBlob{
		cbData: uint32(len(passphrase)),
		pbData: &passphrase[0],
	}

	// Prepare entropy blob
	entropyBlob := dataBlob{
		cbData: uint32(len(entropy)),
		pbData: &entropy[0],
	}

	var outBlob dataBlob

	// Call CryptProtectData
	// dwFlags = 0x01 (CRYPTPROTECT_UI_FORBIDDEN) - no UI prompts
	ret, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(&inBlob)),
		0, // Description (null)
		uintptr(unsafe.Pointer(&entropyBlob)),
		0,    // Reserved
		0,    // PromptStruct (null)
		0x01, // Flags (CRYPTPROTECT_UI_FORBIDDEN)
		uintptr(unsafe.Pointer(&outBlob)),
	)

	if ret == 0 {
		return fmt.Errorf("CryptProtectData failed: %v", err)
	}

	// Copy encrypted data
	encryptedData := make([]byte, outBlob.cbData)
	copy(encryptedData, (*[1 << 30]byte)(unsafe.Pointer(outBlob.pbData))[:outBlob.cbData:outBlob.cbData])

	// Free the output buffer
	if _, _, err := procLocalFree.Call(uintptr(unsafe.Pointer(outBlob.pbData))); err != nil && err.Error() != "The operation completed successfully." {
		fmt.Printf("WARNING: Failed to free memory: %v\n", err)
	}

	// Store in user's AppData
	return storeEncryptedDataToFile(clientID, encryptedData)
}

// loadPassphraseFromKeychain loads the passphrase from Windows DPAPI storage
func loadPassphraseFromKeychain(clientID string) ([]byte, error) {
	// Load encrypted data from file
	encryptedData, err := loadEncryptedDataFromFile(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to load encrypted data: %w", err)
	}

	// Use client ID as entropy
	entropy := sha256.Sum256([]byte(clientID))

	// Prepare input blob
	inBlob := dataBlob{
		cbData: uint32(len(encryptedData)),
		pbData: &encryptedData[0],
	}

	// Prepare entropy blob
	entropyBlob := dataBlob{
		cbData: uint32(len(entropy)),
		pbData: &entropy[0],
	}

	var outBlob dataBlob

	// Call CryptUnprotectData
	ret, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&inBlob)),
		0, // Description
		uintptr(unsafe.Pointer(&entropyBlob)),
		0,    // Reserved
		0,    // PromptStruct (null)
		0x01, // Flags (CRYPTPROTECT_UI_FORBIDDEN)
		uintptr(unsafe.Pointer(&outBlob)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("CryptUnprotectData failed: %v", err)
	}

	// Copy decrypted data
	passphrase := make([]byte, outBlob.cbData)
	copy(passphrase, (*[1 << 30]byte)(unsafe.Pointer(outBlob.pbData))[:outBlob.cbData:outBlob.cbData])

	// Free the output buffer
	if _, _, err := procLocalFree.Call(uintptr(unsafe.Pointer(outBlob.pbData))); err != nil && err.Error() != "The operation completed successfully." {
		fmt.Printf("WARNING: Failed to free memory: %v\n", err)
	}

	return passphrase, nil
}

// deletePassphraseFromKeychain removes the stored passphrase
func deletePassphraseFromKeychain(clientID string) error {
	return deleteEncryptedDataFile(clientID)
}

