package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/cli"
)

func main() {
	fmt.Println("=== ZFS Cloud Upload Test ===")

	// Load config
	configPath := "test-client-config-50052.yaml"
	cfg, err := cli.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("✓ Config loaded: server=%s\n", cfg.ServerAddress)

	// Load or generate keys
	keyPair, err := cli.LoadOrGenerateKeys("test-client-keys.json")
	if err != nil {
		log.Fatalf("Failed to load keys: %v", err)
	}

	clientID := "test-170503"

	// Create client
	c := client.NewClient(cfg.ServerAddress, clientID, keyPair.PrivateKey, keyPair.PublicKey)

	// Connect
	fmt.Printf("Connecting to %s...\n", cfg.ServerAddress)
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()
	fmt.Println("✓ Connected!")

	// Test heartbeat
	fmt.Println("Testing heartbeat...")
	if err := c.Heartbeat(); err != nil {
		log.Fatalf("Heartbeat failed: %v", err)
	}
	fmt.Println("✓ Heartbeat successful!")

	// List files (should be empty)
	fmt.Println("\nListing remote files...")
	files, err := c.ListFiles("", false)
	if err != nil {
		log.Fatalf("List files failed: %v", err)
	}
	fmt.Printf("✓ Remote files before upload: %d\n", len(files))

	// Create uploader
	uploader := client.NewUploader(c)

	// Upload directory
	uploadDir := "/home/user/Téléchargements"
	fmt.Printf("\nScanning %s...\n", uploadDir)

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	var filesToUpload []string
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(uploadDir, entry.Name())
			filesToUpload = append(filesToUpload, filePath)
		}
	}

	fmt.Printf("Found %d files to upload\n\n", len(filesToUpload))

	// Upload each file
	successCount := 0
	failCount := 0

	for i, filePath := range filesToUpload {
		fileName := filepath.Base(filePath)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("[%d/%d] ✗ %s: stat failed: %v\n", i+1, len(filesToUpload), fileName, err)
			failCount++
			continue
		}

		fmt.Printf("[%d/%d] Uploading %s (%.2f MB)...\n",
			i+1, len(filesToUpload), fileName, float64(fileInfo.Size())/1024/1024)

		// Upload with remote path as just the filename
		err = uploader.UploadFile(filePath, fileName)

		if err != nil {
			fmt.Printf("✗ Upload failed: %v\n\n", err)
			failCount++
		} else {
			fmt.Printf("✓ Upload successful!\n\n")
			successCount++
		}
	}

	fmt.Printf("\n=== Upload Summary ===\n")
	fmt.Printf("Total files: %d\n", len(filesToUpload))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)

	// List files again
	fmt.Println("\nListing remote files after upload...")
	files, err = c.ListFiles("", false)
	if err != nil {
		log.Fatalf("List files failed: %v", err)
	}
	fmt.Printf("✓ Remote files: %d\n", len(files))
	var totalSize int64
	for _, f := range files {
		fmt.Printf("  - %s (%.2f MB)\n", f.Path, float64(f.Size)/1024/1024)
		totalSize += f.Size
	}
	fmt.Printf("Total size: %.2f MB\n", float64(totalSize)/1024/1024)

	fmt.Println("\n=== TEST COMPLETE ===")
}
