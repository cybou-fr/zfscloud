package main

import (
	"log"
	"os"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/client/backup"
	"github.com/ioustamora/zfscloud/client/gui"
	"github.com/ioustamora/zfscloud/shared/config"
	"github.com/ioustamora/zfscloud/shared/crypto"
)

func main() {
	// Load configuration
	cfg, err := config.LoadClientConfig("client-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// For GUI version, we'll generate a key pair if needed
	// In production, this should be loaded/saved to cfg.Client.KeyPath
	var keyPair *crypto.KeyPair
	if _, err := os.Stat(cfg.Client.KeyPath); os.IsNotExist(err) {
		log.Println("Generating new key pair...")
		keyPair, err = crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}
		// TODO: Save key pair to file
	} else {
		log.Println("Loading existing key pair...")
		// TODO: Load key pair from file
		keyPair, err = crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate key pair: %v", err)
		}
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
	app := gui.NewApp(c, engine)

	if err := app.Run(); err != nil {
		log.Fatalf("GUI error: %v", err)
	}
}
