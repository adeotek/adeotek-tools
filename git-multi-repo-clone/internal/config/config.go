// Package config provides configuration functionality for the git-multi-repo-clone tool
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config contains application configuration loaded from YAML
type Config struct {
	GiteaURL                   string   `yaml:"gitea_url"`
	APIToken                   string   `yaml:"api_token"`
	TargetDir                  string   `yaml:"target_dir"`
	Username                   string   `yaml:"username"`
	Password                   string   `yaml:"password"`
	UseBasicAuth               bool     `yaml:"use_basic_auth"`
	SkipSslValidation          bool     `yaml:"skip_ssl_validation"`
	OverrideExistingLocalRepos bool     `yaml:"override_exising_local_repos,omitempty"`
	CloneAsMirror              bool     `yaml:"clone_as_mirror,omitempty"`
	Include                    []string `yaml:"include,omitempty"`
	Exclude                    []string `yaml:"exclude,omitempty"`
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
