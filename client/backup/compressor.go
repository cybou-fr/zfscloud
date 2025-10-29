package backup

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// CompressionType represents the type of compression algorithm
type CompressionType string

const (
	// CompressionNone indicates no compression
	CompressionNone CompressionType = "none"
	// CompressionGzip indicates gzip compression
	CompressionGzip CompressionType = "gzip"
	// CompressionZstd indicates zstd compression
	CompressionZstd CompressionType = "zstd"
)

// Compressor handles data compression
type Compressor struct {
	compressionType CompressionType
	level           int // Compression level
	zstdEncoder     *zstd.Encoder
	zstdDecoder     *zstd.Decoder
}

// NewCompressor creates a new compressor
func NewCompressor(compressionType CompressionType) (*Compressor, error) {
	c := &Compressor{
		compressionType: compressionType,
		level:           3, // Default compression level
	}

	// Initialize zstd encoder/decoder if needed
	if compressionType == CompressionZstd {
		encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
		}
		c.zstdEncoder = encoder

		decoder, err := zstd.NewReader(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		c.zstdDecoder = decoder
	}

	return c, nil
}

// SetLevel sets the compression level
func (c *Compressor) SetLevel(level int) {
	c.level = level
}

// Compress compresses data
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	switch c.compressionType {
	case CompressionNone:
		return data, nil

	case CompressionGzip:
		return c.compressGzip(data)

	case CompressionZstd:
		return c.compressZstd(data)

	default:
		return nil, fmt.Errorf("unsupported compression type: %s", c.compressionType)
	}
}

// Decompress decompresses data
func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	switch c.compressionType {
	case CompressionNone:
		return data, nil

	case CompressionGzip:
		return c.decompressGzip(data)

	case CompressionZstd:
		return c.decompressZstd(data)

	default:
		return nil, fmt.Errorf("unsupported compression type: %s", c.compressionType)
	}
}

// compressGzip compresses data using gzip
func (c *Compressor) compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// decompressGzip decompresses gzip data
func (c *Compressor) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return buf.Bytes(), nil
}

// compressZstd compresses data using zstd
func (c *Compressor) compressZstd(data []byte) ([]byte, error) {
	if c.zstdEncoder == nil {
		return nil, fmt.Errorf("zstd encoder not initialized")
	}

	return c.zstdEncoder.EncodeAll(data, make([]byte, 0, len(data))), nil
}

// decompressZstd decompresses zstd data
func (c *Compressor) decompressZstd(data []byte) ([]byte, error) {
	if c.zstdDecoder == nil {
		return nil, fmt.Errorf("zstd decoder not initialized")
	}

	return c.zstdDecoder.DecodeAll(data, nil)
}

// Close closes the compressor and releases resources
func (c *Compressor) Close() error {
	if c.zstdEncoder != nil {
		c.zstdEncoder.Close()
	}
	if c.zstdDecoder != nil {
		c.zstdDecoder.Close()
	}
	return nil
}

// GetCompressionRatio calculates the compression ratio
func GetCompressionRatio(originalSize, compressedSize int64) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(compressedSize) / float64(originalSize) * 100
}

// ShouldCompress determines if a file should be compressed based on its extension
func ShouldCompress(filename string) bool {
	// Skip compression for already compressed formats
	skipExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".mp3":  true,
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".zip":  true,
		".rar":  true,
		".7z":   true,
		".gz":   true,
		".bz2":  true,
		".xz":   true,
	}

	ext := filepath.Ext(strings.ToLower(filename))
	return !skipExtensions[ext]
}
