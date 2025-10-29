package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ClientConfig represents the client configuration
type ClientConfig struct {
	Client    ClientSettings    `yaml:"client"`
	Backup    BackupSettings    `yaml:"backup"`
	Bandwidth BandwidthSettings `yaml:"bandwidth"`
}

type ClientSettings struct {
	ServerAddress         string `yaml:"server_address"`
	ServerCertFingerprint string `yaml:"server_cert_fingerprint"`
	ClientID              string `yaml:"client_id"`
	KeyPath               string `yaml:"key_path"` // Path to store private key
}

type BackupSettings struct {
	Folders         []string         `yaml:"folders"`
	ExcludePatterns []string         `yaml:"exclude_patterns"`
	Schedule        ScheduleSettings `yaml:"schedule"`
}

type ScheduleSettings struct {
	CheckIntervalMinutes int  `yaml:"check_interval_minutes"`
	BackupOnChange       bool `yaml:"backup_on_change"`
}

type BandwidthSettings struct {
	LimitMbps         int  `yaml:"limit_mbps"`
	ThrottleOnMetered bool `yaml:"throttle_on_metered"`
}

// LoadClientConfig loads client configuration from a YAML file
func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ClientConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.Backup.Schedule.CheckIntervalMinutes == 0 {
		config.Backup.Schedule.CheckIntervalMinutes = 5
	}
	if config.Client.KeyPath == "" {
		config.Client.KeyPath = "./data/client.key"
	}

	return &config, nil
}

// SaveClientConfig saves client configuration to a YAML file
func SaveClientConfig(config *ClientConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
