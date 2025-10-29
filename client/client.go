package client

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/shared/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// Connect establishes a connection to the server
func (c *Client) Connect() error {
	// TODO: Add TLS credentials
	conn, err := grpc.NewClient(c.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
