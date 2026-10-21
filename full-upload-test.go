//go:build ignore

// Standalone developer tool; run explicitly: go run full-upload-test.go

package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/shared/config"
	"github.com/ioustamora/zfscloud/shared/crypto"
)

func main() {
	fmt.Println("=== ZFS Cloud Full Upload Test ===\n")

	// Load configuration
	cfg, err := config.LoadClientConfig("test-client-config-50052.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Printf("✓ Config loaded: server=%s\n", cfg.Client.ServerAddress)

	// Load keys
	keyPath := "test-client-keys.json"
	privateKey, publicKey, clientID, err := loadOrGenerateKeys(keyPath)
	if err != nil {
		log.Fatalf("Failed to load keys: %v", err)
	}

	fmt.Printf("✓ Client ID: %s\n", clientID)
	fmt.Printf("✓ Public Key: %s...\n", base64.StdEncoding.EncodeToString(publicKey)[:32])

	// Create client
	c := client.NewClient(cfg.Client.ServerAddress, clientID, privateKey, publicKey)

	// Connect
	fmt.Printf("\nConnecting to %s...\n", cfg.Client.ServerAddress)
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()
	fmt.Println("✓ Connected!")

	// Check registration status - try heartbeat first
	fmt.Println("\nTesting heartbeat...")
	err = c.Heartbeat()
	if err != nil && strings.Contains(err.Error(), "not found") {
		fmt.Println("⚠ User not registered, registering now...")

		// Get server public key
		serverPubKey, err := c.GetServerPublicKey()
		if err != nil {
			log.Fatalf("Failed to get server public key: %v", err)
		}

		// Generate and encrypt passphrase
		passphrase := fmt.Sprintf("test-passphrase-%d", os.Getpid())
		encryptedPass, err := crypto.EncryptPassphrase([]byte(passphrase), serverPubKey)
		if err != nil {
			log.Fatalf("Failed to encrypt passphrase: %v", err)
		}

		// Register
		metadata := &proto.ClientMetadata{
			Hostname:      "test-host",
			Username:      "testuser",
			Os:            "linux",
			ClientVersion: "test",
		}

		if err := c.Register(metadata, encryptedPass); err != nil {
			log.Fatalf("Registration failed: %v", err)
		}

		fmt.Println("✓ Registration submitted, waiting for approval...")
		fmt.Println("\nTo approve, run:")
		fmt.Printf("  sqlite3 /tmp/zfscloud-test/zfscloud.db \"UPDATE registrations SET status='approved', processed_at=CURRENT_TIMESTAMP WHERE client_id='%s';\"\n", clientID)
		fmt.Println("\nThen run this test again.")
		return
	} else if err != nil {
		log.Fatalf("Heartbeat failed: %v", err)
	}
	fmt.Println("✓ Heartbeat successful!")

	// List files before upload
	fmt.Println("\nListing remote files before upload...")
	files, err := c.ListFiles("", false)
	if err != nil {
		log.Fatalf("List files failed: %v", err)
	}
	fmt.Printf("✓ Remote files: %d\n", len(files))
	if len(files) > 0 {
		for _, f := range files {
			fmt.Printf("  - %s (%.2f MB)\n", f.Path, float64(f.Size)/1024/1024)
		}
	}

	// Create uploader
	uploader := client.NewUploader(c)

	// Scan directory
	uploadDir := "/home/user/Téléchargements"
	fmt.Printf("\nScanning %s...\n", uploadDir)

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	var filesToUpload []string
	var totalSizeBytes int64

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(uploadDir, entry.Name())
			info, err := os.Stat(filePath)
			if err == nil {
				filesToUpload = append(filesToUpload, filePath)
				totalSizeBytes += info.Size()
			}
		}
	}

	fmt.Printf("Found %d files (%.2f MB total)\n\n", len(filesToUpload), float64(totalSizeBytes)/1024/1024)

	// Upload each file
	successCount := 0
	failCount := 0
	uploadedBytes := int64(0)

	for i, filePath := range filesToUpload {
		fileName := filepath.Base(filePath)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("[%d/%d] ✗ %s: stat failed: %v\n", i+1, len(filesToUpload), fileName, err)
			failCount++
			continue
		}

		fmt.Printf("[%d/%d] %s (%.2f MB)\n",
			i+1, len(filesToUpload), fileName, float64(fileInfo.Size())/1024/1024)

		// Upload with remote path as just the filename
		err = uploader.UploadFile(filePath, fileName)

		if err != nil {
			fmt.Printf("      ✗ Failed: %v\n\n", err)
			failCount++
		} else {
			successCount++
			uploadedBytes += fileInfo.Size()
		}
	}

	// Summary
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("UPLOAD SUMMARY\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("Total files:     %d\n", len(filesToUpload))
	fmt.Printf("Successful:      %d (%.2f MB)\n", successCount, float64(uploadedBytes)/1024/1024)
	fmt.Printf("Failed:          %d\n", failCount)
	fmt.Printf("Success rate:    %.1f%%\n", float64(successCount)/float64(len(filesToUpload))*100)

	// List files after upload
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("REMOTE FILES AFTER UPLOAD\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	files, err = c.ListFiles("", false)
	if err != nil {
		log.Fatalf("List files failed: %v", err)
	}

	fmt.Printf("Total files on server: %d\n\n", len(files))

	var remoteSize int64
	for _, f := range files {
		fmt.Printf("  %-40s %10.2f MB\n", f.Path, float64(f.Size)/1024/1024)
		remoteSize += f.Size
	}

	fmt.Printf("\nTotal remote size: %.2f MB\n", float64(remoteSize)/1024/1024)

	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Println("TEST COMPLETE")
	fmt.Printf(strings.Repeat("=", 60) + "\n")
}

func loadOrGenerateKeys(keyPath string) (ed25519.PrivateKey, ed25519.PublicKey, string, error) {
	// Check if key file exists
	if _, err := os.Stat(keyPath); err == nil {
		// Load existing keys
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to read key file: %w", err)
		}

		// Split into lines: first line is client ID, second is private key
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) != 2 {
			return nil, nil, "", fmt.Errorf("invalid key file format")
		}

		clientID := lines[0]
		privateKeyBytes, err := base64.StdEncoding.DecodeString(lines[1])
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to decode private key: %w", err)
		}

		privateKey := ed25519.PrivateKey(privateKeyBytes)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		return privateKey, publicKey, clientID, nil
	}

	// Generate new keys
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate keys: %w", err)
	}

	// Generate client ID
	clientID := fmt.Sprintf("test-%d", os.Getpid())

	// Save keys
	keyData := fmt.Sprintf("%s\n%s\n", clientID, base64.StdEncoding.EncodeToString(privateKey))
	if err := os.WriteFile(keyPath, []byte(keyData), 0600); err != nil {
		return nil, nil, "", fmt.Errorf("failed to save keys: %w", err)
	}

	return privateKey, publicKey, clientID, nil
}
