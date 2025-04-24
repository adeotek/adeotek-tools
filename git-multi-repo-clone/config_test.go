package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "git-multi-repo-clone-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
gitea_url: "https://test-gitea.example.com"
api_token: "test-token"
target_dir: "/tmp/test-target"
username: "test-user"
password: "test-pass"
use_basic_auth: true
override_exising_local_repos: true
clone_as_mirror: true
include:
  - "repo1"
  - "repo2"
exclude:
  - "repo3"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading the config
	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	// Verify loaded values
	if config.GiteaURL != "https://test-gitea.example.com" {
		t.Errorf("Expected GiteaURL to be 'https://test-gitea.example.com', got '%s'", config.GiteaURL)
	}
	if config.APIToken != "test-token" {
		t.Errorf("Expected APIToken to be 'test-token', got '%s'", config.APIToken)
	}
	if config.TargetDir != "/tmp/test-target" {
		t.Errorf("Expected TargetDir to be '/tmp/test-target', got '%s'", config.TargetDir)
	}
	if config.Username != "test-user" {
		t.Errorf("Expected Username to be 'test-user', got '%s'", config.Username)
	}
	if config.Password != "test-pass" {
		t.Errorf("Expected Password to be 'test-pass', got '%s'", config.Password)
	}
	if !config.UseBasicAuth {
		t.Error("Expected UseBasicAuth to be true")
	}
	if !config.OverrideExistingLocalRepos {
		t.Error("Expected OverrideExistingLocalRepos to be true")
	}
	if !config.CloneAsMirror {
		t.Error("Expected CloneAsMirror to be true")
	}
	if len(config.Include) != 2 || config.Include[0] != "repo1" || config.Include[1] != "repo2" {
		t.Errorf("Include list not loaded correctly, got %v", config.Include)
	}
	if len(config.Exclude) != 1 || config.Exclude[0] != "repo3" {
		t.Errorf("Exclude list not loaded correctly, got %v", config.Exclude)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := loadConfig("nonexistent-file.yaml")
	if err == nil {
		t.Error("Expected an error when loading non-existent file, got nil")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create a temporary invalid config file
	tempDir, err := os.MkdirTemp("", "git-multi-repo-clone-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "invalid-config.yaml")
	invalidContent := `
gitea_url: "https://test-gitea.example.com"
api_token: 123 : invalid : yaml
`
	err = os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading the invalid config
	_, err = loadConfig(configPath)
	if err == nil {
		t.Error("Expected an error when loading invalid YAML, got nil")
	}
}