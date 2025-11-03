package client

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/server/security"
	"github.com/ioustamora/zfscloud/shared/config"
	"github.com/ioustamora/zfscloud/shared/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client represents the ZFSCloud backup client
type Client struct {
	conn          *grpc.ClientConn
	backupService proto.BackupServiceClient
	keyPair       *crypto.KeyPair
	clientID      string
	serverAddress string
	sessionActive bool
	lastHeartbeat time.Time
	tlsConfig     *config.ClientTLSSettings
}

// NewClient creates a new ZFSCloud client
func NewClient(serverAddress, clientID string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *Client {
	return &Client{
		serverAddress: serverAddress,
		clientID:      clientID,
		keyPair: &crypto.KeyPair{
			PublicKey:  publicKey,
			PrivateKey: privateKey,
		},
	}
}

// SetTLSConfig sets TLS configuration for the client
func (c *Client) SetTLSConfig(tlsConfig *config.ClientTLSSettings) {
	c.tlsConfig = tlsConfig
}

// Connect establishes a connection to the server
func (c *Client) Connect() error {
	var opts []grpc.DialOption

	// Setup TLS if enabled
	if c.tlsConfig != nil && c.tlsConfig.Enabled {
		creds, err := security.LoadClientTLSCredentials(
			c.tlsConfig.CAFile,
			c.tlsConfig.CertFile,
			c.tlsConfig.KeyFile,
			c.tlsConfig.ServerName,
			c.tlsConfig.InsecureSkipVerify,
		)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(c.serverAddress, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.backupService = proto.NewBackupServiceClient(conn)
	return nil
}

// Close closes the connection to the server
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns true if the client has an active connection to the server
func (c *Client) IsConnected() bool {
	return c.conn != nil
}

// Register registers the client with the server
func (c *Client) Register(metadata *proto.ClientMetadata, encryptedPassphrase []byte) error {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := crypto.GetTimestamp()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  c.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(c.clientID),
	}

	signature, err := c.keyPair.Sign(sigData)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	req := &proto.RegisterRequest{
		ClientId:            c.clientID,
		PublicKey:           c.keyPair.PublicKey,
		Metadata:            metadata,
		EncryptedPassphrase: encryptedPassphrase,
		Signature:           signature,
		Timestamp:           timestamp,
		Nonce:               nonce,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.backupService.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration rejected: %s", resp.Message)
	}

	fmt.Printf("Registration successful: %s\n", resp.Message)
	fmt.Printf("Registration ID: %s\n", resp.RegistrationId)
	fmt.Printf("Status: %s\n", resp.Status.String())

	return nil
}

// GetServerPublicKey retrieves the server's RSA public key for encrypting passphrases
func (c *Client) GetServerPublicKey() ([]byte, error) {
	if c.backupService == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.ServerPublicKeyRequest{}

	resp, err := c.backupService.GetServerPublicKey(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get server public key: %w", err)
	}

	if len(resp.PublicKey) == 0 {
		return nil, fmt.Errorf("server returned empty public key")
	}

	return resp.PublicKey, nil
}

// StorePassphrase stores the passphrase securely in OS keychain
func (c *Client) StorePassphrase(passphrase []byte) error {
	return storePassphraseInKeychain(c.clientID, passphrase)
}

// LoadPassphrase loads the passphrase from OS keychain
func (c *Client) LoadPassphrase() ([]byte, error) {
	return loadPassphraseFromKeychain(c.clientID)
}

// DeletePassphrase removes the stored passphrase from OS keychain
func (c *Client) DeletePassphrase() error {
	return deletePassphraseFromKeychain(c.clientID)
}

// HasStoredPassphrase checks if a passphrase is stored in OS keychain
func (c *Client) HasStoredPassphrase() bool {
	return hasStoredPassphrase(c.clientID)
}

// Heartbeat sends a heartbeat to keep the session alive
func (c *Client) Heartbeat() error {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := crypto.GetTimestamp()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  c.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(c.clientID),
	}

	signature, err := c.keyPair.Sign(sigData)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	req := &proto.HeartbeatRequest{
		ClientId:  c.clientID,
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
	}

	// Add auth metadata (base64-encode binary values)
	md := metadata.Pairs(
		"client-id", c.clientID,
		"signature", base64.StdEncoding.EncodeToString(signature),
		"timestamp", fmt.Sprintf("%d", timestamp),
		"nonce", nonce,
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.backupService.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("heartbeat rejected")
	}

	c.lastHeartbeat = time.Now()
	c.sessionActive = true
	return nil
}

// StartHeartbeatLoop starts a background goroutine that sends periodic heartbeats
func (c *Client) StartHeartbeatLoop(interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.Heartbeat(); err != nil {
					fmt.Printf("Heartbeat error: %v\n", err)
				}
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// ListFiles lists files in the backup
func (c *Client) ListFiles(path string, recursive bool) ([]*proto.FileInfo, error) {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := crypto.GetTimestamp()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  c.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(path),
	}

	signature, err := c.keyPair.Sign(sigData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	req := &proto.ListFilesRequest{
		ClientId:  c.clientID,
		Path:      path,
		Recursive: recursive,
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.backupService.ListFiles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return resp.Files, nil
}

// DownloadFile downloads a file from the backup and saves it to destPath
func (c *Client) DownloadFile(filePath, destPath string) error {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := crypto.GetTimestamp()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  c.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(filePath),
	}

	signature, err := c.keyPair.Sign(sigData)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	req := &proto.DownloadFileRequest{
		ClientId:  c.clientID,
		FilePath:  filePath,
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stream, err := c.backupService.DownloadFile(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Create output file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Receive chunks
	var totalBytes int64
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %w", err)
		}

		switch data := resp.Data.(type) {
		case *proto.DownloadFileResponse_Metadata:
			// First message contains metadata
			log.Printf("Downloading file: %s (size: %d bytes)", data.Metadata.Path, data.Metadata.Size)
		case *proto.DownloadFileResponse_ChunkData:
			// Write chunk to file
			n, err := outFile.Write(data.ChunkData)
			if err != nil {
				return fmt.Errorf("failed to write chunk: %w", err)
			}
			totalBytes += int64(n)
		}
	}

	log.Printf("Download complete: %d bytes written to %s", totalBytes, destPath)
	return nil
}

// ListSnapshots lists available snapshots
func (c *Client) ListSnapshots() ([]*proto.SnapshotInfo, error) {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := crypto.GetTimestamp()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  c.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(c.clientID),
	}

	signature, err := c.keyPair.Sign(sigData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	req := &proto.ListSnapshotsRequest{
		ClientId:  c.clientID,
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.backupService.ListSnapshots(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	return resp.Snapshots, nil
}

// GetClientID returns the client ID
func (c *Client) GetClientID() string {
	return c.clientID
}

// GetPublicKey returns the public key as base64
func (c *Client) GetPublicKey() string {
	return crypto.PublicKeyToString(c.keyPair.PublicKey)
}
