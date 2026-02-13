package models

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the unified YAML configuration for sql-toolbox
type Config struct {
	Migration *MigrationConfig `yaml:"migration"`
	Backup    *BackupConfig    `yaml:"backup"`
}

// MigrationConfig holds migration-specific configuration loaded from YAML
type MigrationConfig struct {
	TargetPath       string `yaml:"target_path"`
	Provider         string `yaml:"provider"`
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	Database         string `yaml:"database"`
	User             string `yaml:"user"`
	Password         string `yaml:"password"`
	ConnectionString string `yaml:"connection_string"`
	Backup           bool   `yaml:"backup"`
	BackupOnly       bool   `yaml:"backup_only"`
	Restore          bool   `yaml:"restore"`
}

// LoadConfig parses a unified YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
