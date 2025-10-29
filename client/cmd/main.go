package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ioustamora/zfscloud/shared/config"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	version    = "0.1.0"
)

func main() {
	flag.Parse()

	log.Printf("ZFSCloud Client v%s starting...", version)

	// Load configuration
	cfg, err := config.LoadClientConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded from %s", *configPath)
	log.Printf("Server address: %s", cfg.Client.ServerAddress)
	log.Printf("Backup folders: %v", cfg.Backup.Folders)

	// TODO: Initialize client
	// TODO: Load or generate key pair
	// TODO: Connect to server
	// TODO: Start UI (Gio)
	// TODO: Start backup watcher

	log.Println("Client initialized successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")

	// TODO: Cleanup and shutdown

	log.Println("Client stopped")
}
