package testuploadclient
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/shared/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("server", "127.0.0.1:50052", "Server address")
	uploadDir  = flag.String("dir", "", "Directory to upload")
	uploadFile = flag.String("file", "", "Single file to upload")
	register   = flag.Bool("register", false, "Register new user")
	userID     = flag.String("user", "testuser", "User ID")
)

func main() {
	flag.Parse()

	log.Printf("ZFSCloud Test Client - Connecting to %s", *serverAddr)

	// Connect to server
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, *serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	log.Printf("✓ Connected to server at %s", *serverAddr)

	// Create minimal config
	cfg := &config.ClientConfig{
		Server: config.ServerConfig{
			Address: *serverAddr,
		},
	}

	// Create client
	c, err := client.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	log.Printf("✓ Client initialized")

	// Handle registration
	if *register {
		log.Printf("Registering user: %s", *userID)
		err = c.Register(*userID, *userID+"@example.com")
		if err != nil {
			log.Fatalf("Registration failed: %v", err)
		}
		log.Printf("✓ Registration request sent (waiting for admin approval)")
		return
	}

	// Handle file upload
	if *uploadFile != "" {
		log.Printf("Uploading file: %s", *uploadFile)
		
		uploader := client.NewUploader(c)
		remotePath := filepath.Base(*uploadFile)
		
		err = uploader.UploadFile(*uploadFile, remotePath)
		if err != nil {
			log.Fatalf("Upload failed: %v", err)
		}
		
		log.Printf("✓ File uploaded successfully: %s", remotePath)
		return
	}

	// Handle directory upload
	if *uploadDir != "" {
		log.Printf("Uploading directory: %s", *uploadDir)
		
		uploader := client.NewUploader(c)
		
		// Walk directory and upload files
		var files []string
		err = filepath.Walk(*uploadDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Size() < 100*1024*1024 { // Skip files > 100MB
				relPath, _ := filepath.Rel(*uploadDir, path)
				log.Printf("  Uploading: %s (%d bytes)", relPath, info.Size())
				
				err = uploader.UploadFile(path, relPath)
				if err != nil {
					log.Printf("  ✗ Failed: %v", err)
					return nil // Continue with other files
				}
				
				log.Printf("  ✓ Uploaded: %s", relPath)
				files = append(files, relPath)
			}
			return nil
		})
		
		if err != nil {
			log.Fatalf("Directory walk failed: %v", err)
		}
		
		log.Printf("✓ Uploaded %d files from directory", len(files))
		return
	}

	// Default: just test connectivity
	fmt.Println("\nNo action specified. Available options:")
	fmt.Printf("  -register -user USERNAME    Register new user\n")
	fmt.Printf("  -file PATH                  Upload single file\n")
	fmt.Printf("  -dir PATH                   Upload directory\n")
	fmt.Printf("  -server ADDRESS             Server address (default: 127.0.0.1:50052)\n")
}
