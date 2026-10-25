// Package config provides configuration functionality for the git-repos-backup tool
package config

import (
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

// ProviderType defines the type of git provider
type ProviderType string

const (
	// ProviderGitea is for Gitea servers
	ProviderGitea ProviderType = "gitea"
	// ProviderForgejo is for Forgejo servers. Forgejo is a Gitea fork and
	// keeps the same API, so it is handled by the Gitea implementation.
	ProviderForgejo ProviderType = "forgejo"
	// ProviderGitHub is for GitHub
	ProviderGitHub ProviderType = "github"
)

// SupportedProviders lists the provider types accepted in the configuration.
var SupportedProviders = []ProviderType{ProviderGitea, ProviderForgejo, ProviderGitHub}

// IsSupportedProvider reports whether providerType is a known provider.
func IsSupportedProvider(providerType ProviderType) bool {
	for _, supported := range SupportedProviders {
		if providerType == supported {
			return true
		}
	}
	return false
}

// SupportedProviderNames returns the supported provider types as a
// comma-separated string for use in messages.
func SupportedProviderNames() string {
	names := make([]string, 0, len(SupportedProviders))
	for _, provider := range SupportedProviders {
		names = append(names, string(provider))
	}
	return strings.Join(names, ", ")
}

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
