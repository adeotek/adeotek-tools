package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Valid config file
	validConfig := `
providers:
  - type: gitea
    server_url: https://gitea.example.com
    access_token: fake_token
    target_dir: /path/to/backup
  - type: github
    access_token: github_token
    target_dir: /github/backup
    include:
      - owner/repo1
      - owner/repo2
`

	if err := os.WriteFile(configFile, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading valid config
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if len(cfg.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(cfg.Providers))
	}

	// Check Gitea provider
	if cfg.Providers[0].Type != ProviderGitea {
		t.Errorf("Expected type %s, got %s", ProviderGitea, cfg.Providers[0].Type)
	}
	if cfg.Providers[0].ServerURL != "https://gitea.example.com" {
		t.Errorf("Expected server_url %s, got %s", "https://gitea.example.com", cfg.Providers[0].ServerURL)
	}
	if cfg.Providers[0].AccessToken != "fake_token" {
		t.Errorf("Expected access_token %s, got %s", "fake_token", cfg.Providers[0].AccessToken)
	}
	if cfg.Providers[0].TargetDir != "/path/to/backup" {
		t.Errorf("Expected target_dir %s, got %s", "/path/to/backup", cfg.Providers[0].TargetDir)
	}

	// Check GitHub provider
	if cfg.Providers[1].Type != ProviderGitHub {
		t.Errorf("Expected type %s, got %s", ProviderGitHub, cfg.Providers[1].Type)
	}
	if cfg.Providers[1].AccessToken != "github_token" {
		t.Errorf("Expected access_token %s, got %s", "github_token", cfg.Providers[1].AccessToken)
	}
	if len(cfg.Providers[1].Include) != 2 {
		t.Errorf("Expected 2 include entries, got %d", len(cfg.Providers[1].Include))
	}
	if cfg.Providers[1].Include[0] != "owner/repo1" {
		t.Errorf("Expected include[0] %s, got %s", "owner/repo1", cfg.Providers[1].Include[0])
	}

	// Test loading invalid config
	_, err = Load("nonexistent-file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	// Test invalid YAML
	invalidConfig := `
providers:
  - type: gitea
    server_url: https://gitea.example.com
  invalid yaml
`
	if err := os.WriteFile(configFile, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err = Load(configFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
