package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// LoadServerTLSCredentials loads server TLS credentials for gRPC
func LoadServerTLSCredentials(certFile, keyFile, caFile string, clientAuth tls.ClientAuthType) (credentials.TransportCredentials, error) {
	// Load server certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server key pair: %w", err)
	}

	// Create TLS config
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   clientAuth,
		MinVersion:   tls.VersionTLS13,
	}

	// Load CA certificate for client authentication if required
	if clientAuth != tls.NoClientCert && caFile != "" {
		certPool := x509.NewCertPool()
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to add CA certificate to pool")
		}

		config.ClientCAs = certPool
	}

	return credentials.NewTLS(config), nil
}

// LoadClientTLSCredentials loads client TLS credentials for gRPC
func LoadClientTLSCredentials(caFile, certFile, keyFile, serverName string, insecureSkipVerify bool) (credentials.TransportCredentials, error) {
	// Load CA certificate
	certPool := x509.NewCertPool()
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	// Create TLS config
	config := &tls.Config{
		RootCAs:            certPool,
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify,
		MinVersion:         tls.VersionTLS13,
	}

	// Load client certificate if provided (for mTLS)
	if certFile != "" && keyFile != "" {
		clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair: %w", err)
		}
		config.Certificates = []tls.Certificate{clientCert}
	}

	return credentials.NewTLS(config), nil
}

// GetClientAuthType converts string to tls.ClientAuthType
func GetClientAuthType(authType string) tls.ClientAuthType {
	switch authType {
	case "none":
		return tls.NoClientCert
	case "request":
		return tls.RequestClientCert
	case "require":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}
