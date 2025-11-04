package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/proto"
)

func main() {
	fmt.Println("=== ZFS Cloud Upload Test ===")

	// Load config
	cfg, err := client.LoadConfig("test-client-config-50052.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("✓ Config loaded: server=%s\n", cfg.ServerAddress)

	// Create client
	c := client.NewClient(cfg)

	// Connect
	fmt.Printf("Connecting to %s...\n", cfg.ServerAddress)
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()
	fmt.Println("✓ Connected!")

	// Test heartbeat
	fmt.Println("Testing heartbeat...")
	stats, err := c.Heartbeat()
	if err != nil {
		log.Fatalf("Heartbeat failed: %v", err)
	}
	fmt.Printf("✓ Heartbeat successful! Files: %d, Size: %d bytes\n", stats.FileCount, stats.TotalSize)

	// List files (should be empty)
	fmt.Println("\nListing remote files...")
	files, err := c.ListFiles("", false)
	if err != nil {
		log.Fatalf("List files failed: %v", err)
	}
	fmt.Printf("✓ Remote files: %d\n", len(files))
	for _, f := range files {
		fmt.Printf("  - %s (%d bytes)\n", f.Path, f.Size)
	}

	// Upload files from /home/user/Téléchargements
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

	fmt.Printf("Found %d files to upload\n", len(filesToUpload))

	// Upload each file
	successCount := 0
	failCount := 0

	for i, filePath := range filesToUpload {
		fileName := filepath.Base(filePath)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("  [%d/%d] ✗ %s: stat failed: %v\n", i+1, len(filesToUpload), fileName, err)
			failCount++
			continue
		}

		fmt.Printf("  [%d/%d] Uploading %s (%.2f MB)...\n",
			i+1, len(filesToUpload), fileName, float64(fileInfo.Size())/1024/1024)

		startTime := time.Now()

		// Open file
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("    ✗ Failed to open: %v\n", err)
			failCount++
			continue
		}

		// Upload
		err = uploadFile(c, file, fileName, fileInfo.Size())
		file.Close()

		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("    ✗ Upload failed: %v\n", err)
			failCount++
		} else {
			speed := float64(fileInfo.Size()) / duration.Seconds() / 1024 / 1024
			fmt.Printf("    ✓ Uploaded in %.2fs (%.2f MB/s)\n", duration.Seconds(), speed)
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

	// Final heartbeat
	fmt.Println("\nFinal heartbeat...")
	stats, err = c.Heartbeat()
	if err != nil {
		log.Fatalf("Heartbeat failed: %v", err)
	}
	fmt.Printf("✓ Files: %d, Size: %.2f MB\n", stats.FileCount, float64(stats.TotalSize)/1024/1024)

	fmt.Println("\n=== TEST COMPLETE ===")
}

func uploadFile(c *client.Client, file *os.File, filename string, size int64) error {
	// Create upload stream
	stream, err := c.GetBackupService().UploadFile(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Send metadata
	metadata := &proto.UploadFileRequest{
		Data: &proto.UploadFileRequest_Metadata{
			Metadata: &proto.FileMetadata{
				Path:         filename,
				Size:         size,
				ModifiedTime: time.Now().Unix(),
				IsDirectory:  false,
			},
		},
	}

	if err := stream.Send(metadata); err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	// Send file chunks
	buffer := make([]byte, 64*1024) // 64KB chunks
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		chunk := &proto.UploadFileRequest{
			Data: &proto.UploadFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("upload failed: %s", resp.Message)
	}

	return nil
}
