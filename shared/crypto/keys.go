package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// KeyPair represents an Ed25519 key pair
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair creates a new Ed25519 key pair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// SignatureData represents the data to be signed
type SignatureData struct {
	ClientID  string
	Timestamp int64
	Nonce     string
	Payload   []byte
}

// Sign creates a signature for the given data
func (kp *KeyPair) Sign(data *SignatureData) ([]byte, error) {
	if kp.PrivateKey == nil {
		return nil, fmt.Errorf("private key is required for signing")
	}

	message := kp.buildSignatureMessage(data)
	signature := ed25519.Sign(kp.PrivateKey, message)
	return signature, nil
}

// Verify checks if a signature is valid
func Verify(publicKey ed25519.PublicKey, data *SignatureData, signature []byte) bool {
	kp := &KeyPair{PublicKey: publicKey}
	message := kp.buildSignatureMessage(data)
	return ed25519.Verify(publicKey, message, signature)
}

// VerifySignature checks if a signature is valid
func VerifySignature(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// buildSignatureMessage constructs the message to be signed
func (kp *KeyPair) buildSignatureMessage(data *SignatureData) []byte {
	// Format: ClientID|Timestamp|Nonce|SHA256(Payload)
	hash := sha256.Sum256(data.Payload)
	message := fmt.Sprintf("%s|%d|%s|%x",
		data.ClientID,
		data.Timestamp,
		data.Nonce,
		hash,
	)
	return []byte(message)
}

// GenerateNonce creates a random nonce for replay protection
func GenerateNonce() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(nonce), nil
}

// GenerateClientID creates a new unique client identifier
func GenerateClientID() string {
	return uuid.New().String()
}

// GetTimestamp returns the current Unix timestamp
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// ValidateTimestamp checks if a timestamp is within acceptable range (5 minutes)
func ValidateTimestamp(timestamp int64) bool {
	now := time.Now().Unix()
	diff := now - timestamp

	// Allow 5 minutes clock skew in either direction
	return diff >= -300 && diff <= 300
}

// PublicKeyToString converts a public key to base64 string
func PublicKeyToString(key ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(key)
}

// StringToPublicKey converts a base64 string to public key
func StringToPublicKey(s string) (ed25519.PublicKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	if len(bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(bytes))
	}

	return ed25519.PublicKey(bytes), nil
}

// PrivateKeyToString converts a private key to base64 string
func PrivateKeyToString(key ed25519.PrivateKey) string {
	return base64.StdEncoding.EncodeToString(key)
}

// StringToPrivateKey converts a base64 string to private key
func StringToPrivateKey(s string) (ed25519.PrivateKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: %d", len(bytes))
	}

	return ed25519.PrivateKey(bytes), nil
}
