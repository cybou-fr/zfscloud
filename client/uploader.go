package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/shared/crypto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// DefaultChunkSize is the default size for file chunks (1MB)
	DefaultChunkSize = 1024 * 1024
)

// Uploader handles file uploads to the server
type Uploader struct {
	client    *Client
	chunkSize int
}

// NewUploader creates a new uploader
func NewUploader(client *Client) *Uploader {
	return &Uploader{
		client:    client,
		chunkSize: DefaultChunkSize,
	}
}

// SetChunkSize sets the chunk size for uploads
func (u *Uploader) SetChunkSize(size int) {
	u.chunkSize = size
}

// UploadFile uploads a single file to the server
func (u *Uploader) UploadFile(localPath, remotePath string) error {
	// Open file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate file hash
	fileHash, err := u.calculateFileHash(localPath)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Calculate number of chunks
	totalChunks := (fileInfo.Size() + int64(u.chunkSize) - 1) / int64(u.chunkSize)

	fmt.Printf("Uploading %s (%d bytes, %d chunks)\n", filepath.Base(localPath), fileInfo.Size(), totalChunks)

	// Create upload stream
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stream, err := u.client.backupService.UploadChunk(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %w", err)
	}

	// Send metadata for the file
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := time.Now().Unix()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  u.client.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(remotePath),
	}

	signature, err := u.client.keyPair.Sign(sigData)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	metadata := &proto.ChunkMetadata{
		ClientId:     u.client.clientID,
		FilePath:     remotePath,
		FileHash:     fileHash,
		FileSize:     fileInfo.Size(),
		TotalChunks:  totalChunks,
		ChunkIndex:   0,
		Compressed:   false,
		ModifiedTime: timestamppb.New(fileInfo.ModTime()),
		Signature:    signature,
		Timestamp:    timestamp,
		Nonce:        nonce,
	}

	if err := stream.Send(&proto.UploadChunkRequest{
		Data: &proto.UploadChunkRequest_Metadata{
			Metadata: metadata,
		},
	}); err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	// Send file chunks
	buffer := make([]byte, u.chunkSize)
	chunkIndex := int64(0)
	totalBytes := int64(0)

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Send chunk
		if err := stream.Send(&proto.UploadChunkRequest{
			Data: &proto.UploadChunkRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}); err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", chunkIndex, err)
		}

		totalBytes += int64(n)
		chunkIndex++

		// Progress indicator
		progress := float64(totalBytes) / float64(fileInfo.Size()) * 100
		fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", progress, totalBytes, fileInfo.Size())
	}

	fmt.Println() // New line after progress

	// Close stream and get response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("upload rejected: %s", resp.Message)
	}

	fmt.Printf("Upload successful: %s (%d bytes received)\n", resp.Message, resp.BytesReceived)
	return nil
}

// UploadDirectory uploads all files in a directory recursively
func (u *Uploader) UploadDirectory(localPath, remotePath string) error {
	var uploadedFiles []string
	var totalBytes int64

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Construct remote path
		remoteFilePath := filepath.Join(remotePath, relPath)
		remoteFilePath = filepath.ToSlash(remoteFilePath) // Convert to forward slashes

		// Upload file
		fmt.Printf("\nUploading: %s -> %s\n", relPath, remoteFilePath)
		if err := u.UploadFile(path, remoteFilePath); err != nil {
			return fmt.Errorf("failed to upload %s: %w", relPath, err)
		}

		uploadedFiles = append(uploadedFiles, remoteFilePath)
		totalBytes += info.Size()

		return nil
	})

	if err != nil {
		return fmt.Errorf("directory upload failed: %w", err)
	}

	fmt.Printf("\nDirectory upload complete: %d files, %d total bytes\n", len(uploadedFiles), totalBytes)

	// Send completion notification
	return u.CompleteBackup(uploadedFiles)
}

// CompleteBackup notifies the server that a backup is complete
func (u *Uploader) CompleteBackup(filePaths []string) error {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	timestamp := time.Now().Unix()

	// Create signature
	sigData := &crypto.SignatureData{
		ClientID:  u.client.clientID,
		Timestamp: timestamp,
		Nonce:     nonce,
		Payload:   []byte(u.client.clientID),
	}

	signature, err := u.client.keyPair.Sign(sigData)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	req := &proto.CompleteBackupRequest{
		ClientId:  u.client.clientID,
		BackupId:  fmt.Sprintf("backup-%d", time.Now().Unix()),
		FilePaths: filePaths,
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := u.client.backupService.CompleteBackup(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to complete backup: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("backup completion rejected: %s", resp.Message)
	}

	fmt.Printf("Backup completed: %s\n", resp.Message)
	return nil
}

// calculateFileHash calculates the SHA-256 hash of a file
func (u *Uploader) calculateFileHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

// FormatHash formats a hash as a hex string
func FormatHash(hash []byte) string {
	return hex.EncodeToString(hash)
}
