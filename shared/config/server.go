package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig represents the server configuration
type ServerConfig struct {
	Server    ServerSettings   `yaml:"server"`
	ZFS       ZFSSettings      `yaml:"zfs"`
	Quotas    QuotaSettings    `yaml:"quotas"`
	Snapshots SnapshotSettings `yaml:"snapshots"`
	Security  SecuritySettings `yaml:"security"`
	Database  DatabaseSettings `yaml:"database"`
}

type ServerSettings struct {
	ListenAddress string      `yaml:"listen_address"`
	TLSCert       string      `yaml:"tls_cert"`
	TLSKey        string      `yaml:"tls_key"`
	RSAKeyPath    string      `yaml:"rsa_key_path"`
	TLS           TLSSettings `yaml:"tls"`
}

type TLSSettings struct {
	Enabled      bool   `yaml:"enabled"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	CAFile       string `yaml:"ca_file"`
	ClientAuth   string `yaml:"client_auth"` // none, request, require
	CertDir      string `yaml:"cert_dir"`
	AutoGenerate bool   `yaml:"auto_generate"`
}

type ZFSSettings struct {
	Pool          string `yaml:"pool"`
	Compression   string `yaml:"compression"`
	Deduplication bool   `yaml:"deduplication"`
}

type QuotaSettings struct {
	DefaultQuotaGB int64 `yaml:"default_quota_gb"`
	MaxQuotaGB     int64 `yaml:"max_quota_gb"`
}

type SnapshotSettings struct {
	Schedule SnapshotSchedule `yaml:"schedule"`
}

type SnapshotSchedule struct {
	Hourly  int `yaml:"hourly"`
	Daily   int `yaml:"daily"`
	Weekly  int `yaml:"weekly"`
	Monthly int `yaml:"monthly"`
}

type SecuritySettings struct {
	RequireAdminApproval            bool   `yaml:"require_admin_approval"`
	SnapshotRestoreRequiresApproval bool   `yaml:"snapshot_restore_requires_approval"`
	SessionTimeoutMinutes           int    `yaml:"session_timeout_minutes"`
	AdminToken                      string `yaml:"admin_token"`
}

type DatabaseSettings struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"` // sqlite, postgres, etc.
}

// LoadServerConfig loads server configuration from a YAML file
func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.Server.ListenAddress == "" {
		config.Server.ListenAddress = "0.0.0.0:8443"
	}
	if config.ZFS.Compression == "" {
		config.ZFS.Compression = "zstd"
	}
	if config.Quotas.DefaultQuotaGB == 0 {
		config.Quotas.DefaultQuotaGB = 100
	}
	if config.Quotas.MaxQuotaGB == 0 {
		config.Quotas.MaxQuotaGB = 1024
	}
	if config.Security.SessionTimeoutMinutes == 0 {
		config.Security.SessionTimeoutMinutes = 60
	}
	if config.Database.Type == "" {
		config.Database.Type = "sqlite"
	}
	if config.Database.Path == "" {
		config.Database.Path = "./data/zfscloud.db"
	}

	return &config, nil
}

// SaveServerConfig saves server configuration to a YAML file
func SaveServerConfig(config *ServerConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
