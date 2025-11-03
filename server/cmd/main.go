package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/server/grpc"
	"github.com/ioustamora/zfscloud/server/snapshot"
	"github.com/ioustamora/zfscloud/server/zfs"
	"github.com/ioustamora/zfscloud/shared/config"
	grpc_server "google.golang.org/grpc"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	version    = "0.1.0"
)

func main() {
	flag.Parse()

	log.Printf("ZFSCloud Server v%s starting...", version)

	// Load configuration
	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded from %s", *configPath)
	log.Printf("Listen address: %s", cfg.Server.ListenAddress)
	log.Printf("ZFS pool: %s", cfg.ZFS.Pool)

	// Initialize database
	log.Println("Initializing database...")
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	log.Println("Database initialized successfully")

	// Start cleanup goroutines
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := database.CleanupExpiredSessions(); err != nil {
				log.Printf("Error cleaning up sessions: %v", err)
			}
			if err := database.CleanupOldNonces(); err != nil {
				log.Printf("Error cleaning up nonces: %v", err)
			}
		}
	}()

	// Initialize ZFS manager
	log.Println("Initializing ZFS manager...")
	zfsManager := zfs.NewManager(cfg.ZFS.Pool)
	log.Println("ZFS manager initialized successfully")

	// Initialize snapshot scheduler
	log.Println("Initializing snapshot scheduler...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := snapshot.NewScheduler(database, zfsManager, cfg)
	if err := scheduler.Start(ctx); err != nil {
		log.Fatalf("Failed to start snapshot scheduler: %v", err)
	}
	log.Println("Snapshot scheduler initialized successfully")

	// Initialize gRPC server with authentication
	log.Println("Initializing gRPC server...")
	authInterceptor := grpc.NewAuthInterceptor(database)
	grpcServer := grpc_server.NewServer(
		grpc_server.UnaryInterceptor(authInterceptor.Unary()),
		grpc_server.StreamInterceptor(authInterceptor.Stream()),
	)

	// Register services
	backupService := grpc.NewBackupServiceServer(database, zfsManager, cfg)
	proto.RegisterBackupServiceServer(grpcServer, backupService)

	adminService := grpc.NewAdminServiceServer(database, zfsManager, cfg)
	proto.RegisterAdminServiceServer(grpcServer, adminService)

	log.Println("gRPC services registered successfully")

	// Start listening
	listener, err := net.Listen("tcp", cfg.Server.ListenAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Start gRPC server in background
	go func() {
		log.Printf("Starting gRPC server on %s", cfg.Server.ListenAddress)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	log.Println("Server initialized successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")

	// Stop snapshot scheduler
	log.Println("Stopping snapshot scheduler...")
	scheduler.Stop()

	// Stop gRPC server
	log.Println("Stopping gRPC server...")
	grpcServer.GracefulStop()

	// Cancel context
	cancel()

	// Close database
	log.Println("Closing database...")
	database.Close()

	log.Println("Server stopped")
}
