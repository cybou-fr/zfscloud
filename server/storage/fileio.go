package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ioustamora/zfscloud/proto"
)

// ChunkWriter handles writing chunks to the filesystem
type ChunkWriter struct {
	basePath     string
	currentFile  *os.File
	currentHash  []byte
	bytesWritten int64
}

// NewChunkWriter creates a new chunk writer for a dataset
func NewChunkWriter(datasetPath string) *ChunkWriter {
	return &ChunkWriter{
		basePath: datasetPath,
	}
}

// StartFile begins writing a new file
func (w *ChunkWriter) StartFile(metadata *proto.ChunkMetadata) error {
	// Close any previously open file
	if w.currentFile != nil {
		w.currentFile.Close()
	}

	// Sanitize the file path to prevent directory traversal
	cleanPath := filepath.Clean(metadata.FilePath)
	if strings.HasPrefix(cleanPath, "..") {
		return fmt.Errorf("invalid file path: %s", metadata.FilePath)
	}

	// Create full path
	fullPath := filepath.Join(w.basePath, cleanPath)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Open file for writing
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	w.currentFile = file
	w.currentHash = metadata.FileHash
	w.bytesWritten = 0

	return nil
}

// WriteChunk writes a chunk of data to the current file
func (w *ChunkWriter) WriteChunk(data []byte) error {
	if w.currentFile == nil {
		return fmt.Errorf("no file open for writing")
	}

	n, err := w.currentFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	w.bytesWritten += int64(n)
	return nil
}

// FinishFile completes the current file and verifies its hash
func (w *ChunkWriter) FinishFile() error {
	if w.currentFile == nil {
		return fmt.Errorf("no file open")
	}

	defer func() {
		w.currentFile.Close()
		w.currentFile = nil
	}()

	// Seek to beginning for hash verification
	if _, err := w.currentFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Calculate hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, w.currentFile); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	calculatedHash := hex.EncodeToString(hasher.Sum(nil))
	expectedHash := hex.EncodeToString(w.currentHash)

	if calculatedHash != expectedHash {
		// Delete the corrupted file
		os.Remove(w.currentFile.Name())
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, calculatedHash)
	}

	return nil
}

// Close closes any open file
func (w *ChunkWriter) Close() error {
	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}

// BytesWritten returns the number of bytes written to the current file
func (w *ChunkWriter) BytesWritten() int64 {
	return w.bytesWritten
}

// FileReader handles reading files from the filesystem
type FileReader struct {
	basePath string
}

// NewFileReader creates a new file reader for a dataset
func NewFileReader(datasetPath string) *FileReader {
	return &FileReader{
		basePath: datasetPath,
	}
}

// ListFiles recursively lists all files in the dataset
func (r *FileReader) ListFiles(prefix string) ([]*proto.FileInfo, error) {
	var files []*proto.FileInfo

	searchPath := filepath.Join(r.basePath, prefix)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(r.basePath, path)
		if err != nil {
			return err
		}

		// Calculate hash
		hash, err := calculateFileHash(path)
		if err != nil {
			return err
		}

		files = append(files, &proto.FileInfo{
			Path:        filepath.ToSlash(relPath),
			Size:        info.Size(),
			Hash:        []byte(hash),
			IsDirectory: false,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// OpenFile opens a file for reading
func (r *FileReader) OpenFile(filePath string) (*os.File, error) {
	// Sanitize the file path
	cleanPath := filepath.Clean(filePath)
	if strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: %s", filePath)
	}

	fullPath := filepath.Join(r.basePath, cleanPath)

	// Check if file exists and is not a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", filePath)
	}

	return os.Open(fullPath)
}

// GetFileInfo gets information about a file
func (r *FileReader) GetFileInfo(filePath string) (*proto.FileInfo, error) {
	cleanPath := filepath.Clean(filePath)
	if strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: %s", filePath)
	}

	fullPath := filepath.Join(r.basePath, cleanPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	hash, err := calculateFileHash(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	return &proto.FileInfo{
		Path:        filepath.ToSlash(cleanPath),
		Size:        info.Size(),
		Hash:        []byte(hash),
		IsDirectory: info.IsDir(),
	}, nil
}

// calculateFileHash calculates the SHA-256 hash of a file
func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GetDatasetUsedSpace calculates total space used in a dataset
func GetDatasetUsedSpace(datasetPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(datasetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate used space: %w", err)
	}

	return totalSize, nil
}
