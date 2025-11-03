package main

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
	clientgui "github.com/ioustamora/zfscloud/client/gui"
	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/shared/config"
	"github.com/ioustamora/zfscloud/shared/crypto"
)

var (
	configPath  = flag.String("config", "client-config.yaml", "Path to configuration file")
	keyPath     = flag.String("keys", "client-keys.dat", "Path to key file")
	interactive = flag.Bool("interactive", false, "Run in interactive CLI mode")
	gui         = flag.Bool("gui", true, "Run in GUI mode")
	version     = "0.2.0"
)

func main() {
	flag.Parse()

	log.Printf("ZFSCloud Client v%s\n", version)

	// Load configuration
	cfg, err := config.LoadClientConfig(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load configuration: %v", err)
		log.Println("Using default configuration...")
		cfg = &config.ClientConfig{
			Client: config.ClientSettings{
				ServerAddress: "localhost:50051",
			},
		}
	}

	// Load or generate keys (also returns client ID)
	privateKey, publicKey, clientID, err := loadOrGenerateKeys(*keyPath)
	if err != nil {
		log.Fatalf("Failed to load/generate keys: %v", err)
	}

	// Use loaded/generated client ID if not in config
	if cfg.Client.ClientID == "" {
		cfg.Client.ClientID = clientID
	}

	log.Printf("Client ID: %s", cfg.Client.ClientID)
	log.Printf("Public Key: %s", base64.StdEncoding.EncodeToString(publicKey))

	// Create client
	c := client.NewClient(cfg.Client.ServerAddress, cfg.Client.ClientID, privateKey, publicKey)

	// Connect to server
	log.Printf("Connecting to server at %s...", cfg.Client.ServerAddress)
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer c.Close()

	log.Println("Connected successfully!")

	// Choose mode
	if *interactive {
		runInteractiveMode(c, cfg)
	} else if *gui {
		runGUIMode(c, cfg, *configPath, privateKey, publicKey)
	} else {
		log.Println("No mode specified. Use --interactive or --gui")
	}
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

		// Decode base64 private key
		keyData, err := base64.StdEncoding.DecodeString(lines[1])
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to decode key data: %w", err)
		}

		if len(keyData) != ed25519.PrivateKeySize {
			return nil, nil, "", fmt.Errorf("invalid key size: %d", len(keyData))
		}

		privateKey := ed25519.PrivateKey(keyData)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		log.Printf("Loaded existing keys from %s", keyPath)
		return privateKey, publicKey, clientID, nil
	}

	// Generate new keys
	log.Println("Generating new Ed25519 key pair...")
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate keys: %w", err)
	}

	// Generate client ID
	clientID := fmt.Sprintf("client-%d", time.Now().Unix())

	// Save keys: client ID on first line, private key on second line
	keyData := fmt.Sprintf("%s\n%s", clientID, base64.StdEncoding.EncodeToString(keyPair.PrivateKey))
	if err := os.WriteFile(keyPath, []byte(keyData), 0600); err != nil {
		return nil, nil, "", fmt.Errorf("failed to save keys: %w", err)
	}

	log.Printf("Saved new keys to %s", keyPath)
	return keyPair.PrivateKey, keyPair.PublicKey, clientID, nil
}

func runInteractiveMode(c *client.Client, cfg *config.ClientConfig) {
	scanner := bufio.NewScanner(os.Stdin)

	// Create backup engine (but don't start it yet)
	var engine *backup.Engine
	var engineErr error
	ctx := context.Background()

	fmt.Println("\n=== ZFSCloud Interactive Client ===")
	fmt.Println("Type 'help' for available commands")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]

		switch command {
		case "help":
			printHelp()

		case "register":
			handleRegister(c, cfg)

		case "heartbeat":
			handleHeartbeat(c)

		case "upload":
			if len(parts) < 2 {
				fmt.Println("Usage: upload <local-path> [remote-path]")
				continue
			}
			localPath := parts[1]
			remotePath := filepath.Base(localPath)
			if len(parts) >= 3 {
				remotePath = parts[2]
			}
			handleUpload(c, localPath, remotePath)

		case "upload-dir":
			if len(parts) < 2 {
				fmt.Println("Usage: upload-dir <local-directory> [remote-directory]")
				continue
			}
			localPath := parts[1]
			remotePath := filepath.Base(localPath)
			if len(parts) >= 3 {
				remotePath = parts[2]
			}
			handleUploadDir(c, localPath, remotePath)

		case "list":
			path := ""
			if len(parts) >= 2 {
				path = parts[1]
			}
			handleList(c, path)

		case "snapshots":
			handleListSnapshots(c)

		case "status":
			fmt.Printf("Client ID: %s\n", c.GetClientID())
			fmt.Printf("Public Key: %s\n", c.GetPublicKey())
			if engine != nil {
				stats := engine.GetStats()
				fmt.Printf("\nBackup Engine:\n")
				fmt.Printf("  State: %s\n", stats.CurrentState)
				fmt.Printf("  Files Scanned: %d\n", stats.FilesScanned)
				fmt.Printf("  Files Uploaded: %d\n", stats.FilesUploaded)
				fmt.Printf("  Files Failed: %d\n", stats.FilesFailed)
				fmt.Printf("  Bytes Uploaded: %d\n", stats.BytesUploaded)
				if !stats.LastBackupTime.IsZero() {
					fmt.Printf("  Last Backup: %s\n", stats.LastBackupTime.Format(time.RFC3339))
				}
				if stats.LastError != nil {
					fmt.Printf("  Last Error: %v\n", stats.LastError)
				}
			} else {
				fmt.Println("\nBackup Engine: Not initialized")
			}

		case "auto-start":
			if engine != nil {
				fmt.Println("Backup engine already running")
				continue
			}
			if len(cfg.Backup.Folders) == 0 {
				fmt.Println("Error: No backup folders configured. Edit your config file.")
				continue
			}
			fmt.Println("Starting automatic backup engine...")
			engine, engineErr = backup.NewEngine(c, cfg)
			if engineErr != nil {
				fmt.Printf("Failed to create backup engine: %v\n", engineErr)
				engine = nil
				continue
			}
			if err := engine.Start(ctx); err != nil {
				fmt.Printf("Failed to start backup engine: %v\n", err)
				engine = nil
				continue
			}
			fmt.Println("✓ Backup engine started!")
			fmt.Printf("  Watching %d folder(s)\n", len(cfg.Backup.Folders))
			for _, folder := range cfg.Backup.Folders {
				fmt.Printf("    - %s\n", folder)
			}

		case "auto-stop":
			if engine == nil {
				fmt.Println("Backup engine not running")
				continue
			}
			fmt.Println("Stopping backup engine...")
			if err := engine.Stop(); err != nil {
				fmt.Printf("Warning: error stopping engine: %v\n", err)
			}
			engine = nil
			fmt.Println("✓ Backup engine stopped")

		case "auto-pause":
			if engine == nil {
				fmt.Println("Backup engine not running")
				continue
			}
			engine.Pause()
			fmt.Println("✓ Backup engine paused")

		case "auto-resume":
			if engine == nil {
				fmt.Println("Backup engine not running")
				continue
			}
			engine.Resume()
			fmt.Println("✓ Backup engine resumed")

		case "auto-scan":
			if engine == nil {
				fmt.Println("Backup engine not running. Use 'auto-start' first.")
				continue
			}
			fmt.Println("Performing initial scan...")
			if err := engine.PerformInitialScan(ctx); err != nil {
				fmt.Printf("Scan failed: %v\n", err)
				continue
			}
			stats := engine.GetStats()
			fmt.Printf("✓ Scan complete: found %d files\n", stats.FilesScanned)

		case "exit", "quit":
			if engine != nil {
				fmt.Println("Stopping backup engine...")
				engine.Stop()
			}
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", command)
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("  General:")
	fmt.Println("    help                                 - Show this help message")
	fmt.Println("    register                             - Register with the server")
	fmt.Println("    heartbeat                            - Send a heartbeat")
	fmt.Println("    status                               - Show client and backup status")
	fmt.Println("    exit/quit                            - Exit the client")
	fmt.Println()
	fmt.Println("  Manual Backup:")
	fmt.Println("    upload <local> [remote]              - Upload a single file")
	fmt.Println("    upload-dir <local-dir> [remote-dir]  - Upload a directory")
	fmt.Println()
	fmt.Println("  Automatic Backup:")
	fmt.Println("    auto-start                           - Start automatic backup engine")
	fmt.Println("    auto-stop                            - Stop automatic backup engine")
	fmt.Println("    auto-pause                           - Pause automatic backups")
	fmt.Println("    auto-resume                          - Resume automatic backups")
	fmt.Println("    auto-scan                            - Perform initial full scan")
	fmt.Println()
	fmt.Println("  Restore:")
	fmt.Println("    list [path]                          - List files in backup")
	fmt.Println("    snapshots                            - List available snapshots")
}

func handleRegister(c *client.Client, cfg *config.ClientConfig) {
	fmt.Println("Registering with server...")

	// Get metadata
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	metadata := &proto.ClientMetadata{
		Hostname:      hostname,
		Username:      username,
		Os:            fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")),
		ClientVersion: version,
	}

	// Generate passphrase for ZFS encryption
	passphrase, err := crypto.GeneratePassphrase()
	if err != nil {
		fmt.Printf("Error generating passphrase: %v\n", err)
		return
	}

	// Get server's RSA public key
	fmt.Println("Fetching server's public key...")
	serverPubKey, err := c.GetServerPublicKey()
	if err != nil {
		fmt.Printf("Error fetching server public key: %v\n", err)
		fmt.Println("Cannot securely encrypt passphrase. Registration aborted.")
		return
	}

	// Encrypt passphrase with server's public key
	fmt.Println("Encrypting passphrase...")
	encryptedPassphrase, err := crypto.EncryptPassphrase(passphrase, serverPubKey)
	if err != nil {
		fmt.Printf("Error encrypting passphrase: %v\n", err)
		return
	}

	// Register with server first
	if err := c.Register(metadata, encryptedPassphrase); err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	fmt.Println("Registration successful! Wait for admin approval.")

	// Store passphrase securely in OS keychain
	fmt.Println("\nStoring passphrase in OS keychain...")
	if err := c.StorePassphrase(passphrase); err != nil {
		fmt.Printf("⚠️  WARNING: Failed to store passphrase in keychain: %v\n", err)
		fmt.Println("⚠️  Please save this passphrase securely:")
		fmt.Printf("⚠️  Passphrase: %s\n", base64.StdEncoding.EncodeToString(passphrase))
		fmt.Println("⚠️  You will need it to reconnect to the server")
	} else {
		fmt.Println("✅ Passphrase stored securely in OS keychain")
		fmt.Println("💡 The passphrase will be automatically loaded when you connect")
	}
}

func handleHeartbeat(c *client.Client) {
	if err := c.Heartbeat(); err != nil {
		fmt.Printf("Heartbeat failed: %v\n", err)
		return
	}
	fmt.Println("Heartbeat sent successfully!")
}

func handleUpload(c *client.Client, localPath, remotePath string) {
	uploader := client.NewUploader(c)

	fmt.Printf("Uploading %s to %s...\n", localPath, remotePath)
	if err := uploader.UploadFile(localPath, remotePath); err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		return
	}
}

func handleUploadDir(c *client.Client, localPath, remotePath string) {
	uploader := client.NewUploader(c)

	fmt.Printf("Uploading directory %s to %s...\n", localPath, remotePath)
	if err := uploader.UploadDirectory(localPath, remotePath); err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		return
	}
}

func handleList(c *client.Client, path string) {
	fmt.Printf("Listing files in: %s\n", path)
	files, err := c.ListFiles(path, true)
	if err != nil {
		fmt.Printf("Failed to list files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files found.")
		return
	}

	fmt.Printf("\nFound %d files:\n", len(files))
	fmt.Println("------------------------------------------------------------")
	for _, file := range files {
		fmt.Printf("%-50s %10d bytes\n", file.Path, file.Size)
	}
	fmt.Println("------------------------------------------------------------")
}

func handleListSnapshots(c *client.Client) {
	fmt.Println("Listing snapshots...")
	snapshots, err := c.ListSnapshots()
	if err != nil {
		fmt.Printf("Failed to list snapshots: %v\n", err)
		return
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found.")
		return
	}

	fmt.Printf("\nFound %d snapshots:\n", len(snapshots))
	fmt.Println("------------------------------------------------------------")
	for _, snap := range snapshots {
		fmt.Printf("%-30s %s (%d bytes)\n", snap.SnapshotName, snap.CreatedAt.AsTime().Format(time.RFC3339), snap.SizeBytes)
	}
	fmt.Println("------------------------------------------------------------")
}

func runGUIMode(c *client.Client, cfg *config.ClientConfig, configPath string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) {
	log.Println("Starting GUI mode...")

	// Create and run GUI application (engine will be created on demand)
	app := clientgui.NewApp(c, nil, cfg, configPath)
	if err := app.Run(); err != nil {
		log.Fatalf("GUI error: %v", err)
	}
}
