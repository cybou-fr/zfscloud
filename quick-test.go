package main

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	"github.com/ioustamora/zfscloud/client"
	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/shared/config"
	"github.com/ioustamora/zfscloud/shared/crypto"
)

func main() {
	// Load config
	cfg, err := config.LoadClientConfig("test-client-config-50052.yaml")
	if err != nil {
		log.Fatalf("Config load failed: %v", err)
	}

	fmt.Printf("✓ Config loaded: server=%s\n", cfg.Client.ServerAddress)

	// Generate keys
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("Key gen failed: %v", err)
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerAddress, "test-"+time.Now().Format("150405"), priv, pub)

	// Connect
	fmt.Printf("Connecting to %s...\n", cfg.Client.ServerAddress)
	if err := c.Connect(); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer c.Close()

	fmt.Println("✓ Connected!")

	// Try GetServerPublicKey
	fmt.Println("Getting server public key...")
	pubKey, err := c.GetServerPublicKey()
	if err != nil {
		log.Fatalf("GetServerPublicKey failed: %v", err)
	}

	fmt.Printf("✓ Got server public key (%d bytes)\n", len(pubKey))

	// Try register
	fmt.Println("Registering...")
	passphrase, _ := crypto.GeneratePassphrase()
	encPass, err := crypto.EncryptPassphrase(passphrase, pubKey)
	if err != nil {
		log.Fatalf("Encrypt failed: %v", err)
	}

	metadata := &proto.ClientMetadata{
		Hostname:      "test-host",
		Username:      "testuser",
		Os:            "linux",
		ClientVersion: "test",
	}

	if err := c.Register(metadata, encPass); err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	fmt.Println("✓ Registration successful!")
	fmt.Println("\n=== ALL TESTS PASSED ===")
}
