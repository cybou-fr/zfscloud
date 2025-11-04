package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
	"github.com/ioustamora/zfscloud/client/gui"
	"github.com/ioustamora/zfscloud/shared/config"
	"github.com/ioustamora/zfscloud/shared/crypto"
)

func saveKeyPair(keyPair *crypto.KeyPair, filePath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert keys to strings
	pubKeyStr := crypto.PublicKeyToString(keyPair.PublicKey)
	privKeyStr := crypto.PrivateKeyToString(keyPair.PrivateKey)

	// Write to file (private key first, then public key)
	data := fmt.Sprintf("%s\n%s\n", privKeyStr, pubKeyStr)
	if err := os.WriteFile(filePath, []byte(data), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

func loadKeyPair(filePath string) (*crypto.KeyPair, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse lines
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		return nil, fmt.Errorf("invalid key file format")
	}

	// Load private key
	privKey, err := crypto.StringToPrivateKey(lines[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Load public key
	pubKey, err := crypto.StringToPublicKey(lines[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &crypto.KeyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadClientConfig("client-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// For GUI version, we'll load or generate a key pair
	var keyPair *crypto.KeyPair

	if _, statErr := os.Stat(cfg.Client.KeyPath); os.IsNotExist(statErr) {
		log.Println("Generating new key pair...")
		var err error
		keyPair, err = crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}

		// Save the key pair
		if err := saveKeyPair(keyPair, cfg.Client.KeyPath); err != nil {
			log.Printf("Warning: Failed to save key pair: %v", err)
		} else {
			log.Println("Key pair saved successfully")
		}
	} else {
		log.Println("Loading existing key pair...")
		var err error
		keyPair, err = loadKeyPair(cfg.Client.KeyPath)
		if err != nil {
			log.Fatalf("Failed to load key pair: %v", err)
		}
		log.Println("Key pair loaded successfully")
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerAddress, cfg.Client.ClientID, keyPair.PrivateKey, keyPair.PublicKey)

	// Set TLS config if enabled
	if cfg.Client.TLS.Enabled {
		c.SetTLSConfig(&cfg.Client.TLS)
	}

	// Create backup engine (will be started from GUI)
	engine, err := backup.NewEngine(c, cfg)
	if err != nil {
		log.Fatalf("Failed to create backup engine: %v", err)
	}

	// Create and run GUI
	log.Println("Starting GUI...")
	app := gui.NewApp(c, engine, cfg, "client-config.yaml")

	if err := app.Run(); err != nil {
		log.Fatalf("GUI error: %v", err)
	}
}
