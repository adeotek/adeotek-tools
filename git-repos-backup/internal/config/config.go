// Package config provides configuration functionality for the git-repos-backup tool
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ProviderType defines the type of git provider
type ProviderType string

const (
	// ProviderGitea is for Gitea servers
	ProviderGitea ProviderType = "gitea"
	// ProviderGitHub is for GitHub
	ProviderGitHub ProviderType = "github"
)

// ProviderConfig contains configuration for a git provider
type ProviderConfig struct {
	Type              ProviderType `yaml:"type"`
	ServerURL         string       `yaml:"server_url"`
	AccessToken       string       `yaml:"access_token"`
	Username          string       `yaml:"username"`
	Password          string       `yaml:"password"`
	UseBasicAuth      bool         `yaml:"use_basic_auth"`
	SkipSslValidation bool         `yaml:"skip_ssl_validation"`
	Include           []string     `yaml:"include,omitempty"`
	Exclude           []string     `yaml:"exclude,omitempty"`
	TargetDir         string       `yaml:"target_dir"`
}

// Config contains application configuration loaded from YAML
type Config struct {
	Providers []ProviderConfig `yaml:"providers"`
}

// Load loads configuration from the specified YAML file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// CreateFromArgs creates a config from command line arguments
func CreateFromArgs(
	providerType string, 
	serverURL string, 
	accessToken string, 
	username string,
	password string,
	useBasicAuth bool,
	skipSSLValidation bool,
	include []string,
	exclude []string,
	targetDir string,
) *Config {
	provider := ProviderConfig{
		Type:              ProviderType(providerType),
		ServerURL:         serverURL,
		AccessToken:       accessToken,
		Username:          username,
		Password:          password,
		UseBasicAuth:      useBasicAuth,
		SkipSslValidation: skipSSLValidation,
		Include:           include,
		Exclude:           exclude,
		TargetDir:         targetDir,
	}

	return &Config{
		Providers: []ProviderConfig{provider},
	}
}
