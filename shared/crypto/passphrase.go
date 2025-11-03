package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// PassphraseLength is the length of generated passphrases in bytes
	PassphraseLength = 32 // 256 bits

	// Argon2 parameters for key derivation
	Argon2Time    = 1
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4
	Argon2KeyLen  = 32
)

// GeneratePassphrase creates a strong random passphrase for ZFS encryption
func GeneratePassphrase() ([]byte, error) {
	passphrase := make([]byte, PassphraseLength)
	if _, err := io.ReadFull(rand.Reader, passphrase); err != nil {
		return nil, fmt.Errorf("failed to generate passphrase: %w", err)
	}
	return passphrase, nil
}

// DeriveKey derives a key from a passphrase using Argon2
func DeriveKey(passphrase, salt []byte) []byte {
	return argon2.IDKey(
		passphrase,
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLen,
	)
}

// EncryptPassphrase encrypts a passphrase with an RSA public key
func EncryptPassphrase(passphrase []byte, publicKeyPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	encrypted, err := rsa.EncryptOAEP(
		nil,
		rand.Reader,
		rsaPublicKey,
		passphrase,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt passphrase: %w", err)
	}

	return encrypted, nil
}

// DecryptPassphrase decrypts a passphrase with a PEM-encoded RSA private key
func DecryptPassphrase(encryptedPassphrase []byte, privateKeyPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return DecryptWithPrivateKey(encryptedPassphrase, rsaPrivateKey)
}

// DecryptWithPrivateKey decrypts a passphrase with an RSA private key
func DecryptWithPrivateKey(encryptedPassphrase []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(
		nil,
		rand.Reader,
		privateKey,
		encryptedPassphrase,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt passphrase: %w", err)
	}

	return decrypted, nil
}

// GenerateRSAKeyPair generates a new RSA key pair for server
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		bits = 2048
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return privateKey, nil
}

// RSAPrivateKeyToPEM converts an RSA private key to PEM format
func RSAPrivateKeyToPEM(key *rsa.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

// RSAPublicKeyToPEM converts an RSA public key to PEM format
func RSAPublicKeyToPEM(key *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}
