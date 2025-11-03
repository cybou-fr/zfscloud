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
	"github.com/ioustamora/zfscloud/server/security"
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

	var grpcOptions []grpc_server.ServerOption
	grpcOptions = append(grpcOptions,
		grpc_server.UnaryInterceptor(authInterceptor.Unary()),
		grpc_server.StreamInterceptor(authInterceptor.Stream()),
	)

	// Setup TLS if enabled
	if cfg.Server.TLS.Enabled {
		log.Println("TLS enabled, loading credentials...")

		// Auto-generate certificates if requested and they don't exist
		if cfg.Server.TLS.AutoGenerate {
			certMgr := security.NewCertificateManager(cfg.Server.TLS.CertDir)
			if !certMgr.CertExists("ca-cert.pem") {
				log.Println("Generating CA certificate...")
				if err := certMgr.GenerateCA("ZFSCloud CA", 3650); err != nil {
					log.Fatalf("Failed to generate CA: %v", err)
				}
				log.Println("CA certificate generated successfully")
			}
			if !certMgr.CertExists("server-cert.pem") {
				log.Println("Generating server certificate...")
				dnsNames := []string{"localhost", "zfscloud-server"}
				if err := certMgr.GenerateServerCert("zfscloud-server", dnsNames, 365); err != nil {
					log.Fatalf("Failed to generate server certificate: %v", err)
				}
				log.Println("Server certificate generated successfully")
			}
		}

		clientAuth := security.GetClientAuthType(cfg.Server.TLS.ClientAuth)
		creds, err := security.LoadServerTLSCredentials(
			cfg.Server.TLS.CertFile,
			cfg.Server.TLS.KeyFile,
			cfg.Server.TLS.CAFile,
			clientAuth,
		)
		if err != nil {
			log.Fatalf("Failed to load TLS credentials: %v", err)
		}

		grpcOptions = append(grpcOptions, grpc_server.Creds(creds))
		log.Printf("TLS configured with client auth mode: %s", cfg.Server.TLS.ClientAuth)
	} else {
		log.Println("TLS disabled - running in insecure mode")
	}

	grpcServer := grpc_server.NewServer(grpcOptions...)

	// Load or generate RSA keys for passphrase encryption
	log.Println("Loading RSA keys...")
	var rsaPrivateKey, rsaPublicKeyPEM []byte

	rsaKeyPath := cfg.Server.RSAKeyPath
	if rsaKeyPath == "" {
		rsaKeyPath = "server-rsa-key.pem"
	}

	// Check if RSA key exists
	if _, err := os.Stat(rsaKeyPath); os.IsNotExist(err) {
		log.Printf("RSA key not found at %s, generating new key...", rsaKeyPath)
		// Generate new RSA key pair
		privateKey, publicKey, err := security.GenerateRSAKeyPair(4096)
		if err != nil {
			log.Fatalf("Failed to generate RSA key pair: %v", err)
		}

		// Save private key to file
		if err := os.WriteFile(rsaKeyPath, privateKey, 0600); err != nil {
			log.Fatalf("Failed to save RSA private key: %v", err)
		}
		log.Printf("RSA key pair generated and saved to %s", rsaKeyPath)

		rsaPrivateKey = privateKey
		rsaPublicKeyPEM = publicKey
	} else {
		// Load existing key
		rsaPrivateKey, err = os.ReadFile(rsaKeyPath)
		if err != nil {
			log.Fatalf("Failed to read RSA private key: %v", err)
		}
		log.Printf("RSA private key loaded from %s", rsaKeyPath)

		// Extract public key from private key
		rsaPublicKeyPEM, err = security.ExtractPublicKeyFromPrivateKey(rsaPrivateKey)
		if err != nil {
			log.Fatalf("Failed to extract public key from private key: %v", err)
		}
	}

	log.Println("RSA keys loaded successfully")

	// Register services
	backupService := grpc.NewBackupServiceServer(database, zfsManager, cfg, rsaPrivateKey, rsaPublicKeyPEM)
	proto.RegisterBackupServiceServer(grpcServer, backupService)

	adminService := grpc.NewAdminServiceServer(database, zfsManager, cfg, rsaPrivateKey)
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
