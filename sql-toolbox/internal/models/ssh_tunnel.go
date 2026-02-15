package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSHTunnelConfig defines SSH tunnel configuration for database connections
type SSHTunnelConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	AuthMethod string `yaml:"auth_method"` // "key", "password", or "agent"
	KeyFile    string `yaml:"key_file"`    // Path to private key (for "key" method)
	Password   string `yaml:"password"`    // SSH password (for "password" method)
	LocalPort  int    `yaml:"local_port"`  // Local port (0 = auto-assign)
}

// Validate checks if the SSH tunnel configuration is valid
func (c *SSHTunnelConfig) Validate() error {
	if !c.Enabled {
		return nil // No validation needed if disabled
	}

	if c.Host == "" {
		return fmt.Errorf("SSH tunnel host is required")
	}

	if c.User == "" {
		return fmt.Errorf("SSH tunnel user is required")
	}

	// Validate auth_method
	validAuthMethods := map[string]bool{"key": true, "password": true, "agent": true}
	if !validAuthMethods[c.AuthMethod] {
		return fmt.Errorf("invalid auth_method '%s': must be 'key', 'password', or 'agent'", c.AuthMethod)
	}

	// Validate auth_method-specific requirements
	switch c.AuthMethod {
	case "key":
		if c.KeyFile == "" {
			return fmt.Errorf("key_file is required when auth_method is 'key'")
		}
		// Expand home directory
		keyPath := c.KeyFile
		if strings.HasPrefix(keyPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to expand home directory in key_file: %w", err)
			}
			keyPath = filepath.Join(home, keyPath[2:])
		}
		// Check if key file exists
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key file not found: %s", c.KeyFile)
		}

	case "password":
		if c.Password == "" {
			return fmt.Errorf("password is required when auth_method is 'password'")
		}

	case "agent":
		// Check if SSH_AUTH_SOCK is set
		if os.Getenv("SSH_AUTH_SOCK") == "" {
			return fmt.Errorf("SSH_AUTH_SOCK environment variable not set (required for agent auth)")
		}
	}

	return nil
}

// GetEffectivePort returns the SSH port, defaulting to 22 if not set
func (c *SSHTunnelConfig) GetEffectivePort() int {
	if c.Port <= 0 {
		return 22
	}
	return c.Port
}

// ExpandKeyFilePath expands ~ in key file path to absolute path
func (c *SSHTunnelConfig) ExpandKeyFilePath() (string, error) {
	if c.KeyFile == "" {
		return "", nil
	}

	keyPath := c.KeyFile
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand home directory: %w", err)
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	return keyPath, nil
}
