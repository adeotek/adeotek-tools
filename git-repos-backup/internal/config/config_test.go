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

func TestCreateFromArgs(t *testing.T) {
	// Test creating config from arguments
	testCases := []struct {
		name              string
		providerType      string
		serverURL         string
		accessToken       string
		username          string
		password          string
		useBasicAuth      bool
		skipSSLValidation bool
		include           []string
		exclude           []string
		targetDir         string
	}{
		{
			name:         "GitHub with token",
			providerType: "github",
			accessToken:  "github_token",
			targetDir:    "/path/to/backup",
		},
		{
			name:         "Gitea with token and server URL",
			providerType: "gitea",
			serverURL:    "https://gitea.example.com",
			accessToken:  "gitea_token",
			targetDir:    "/path/to/backup",
		},
		{
			name:              "GitHub with basic auth and include",
			providerType:      "github",
			username:          "user",
			password:          "pass",
			useBasicAuth:      true,
			skipSSLValidation: true,
			include:           []string{"owner/repo1", "owner/repo2"},
			targetDir:         "/path/to/backup",
		},
		{
			name:         "GitHub with exclude",
			providerType: "github",
			accessToken:  "token",
			exclude:      []string{"owner/repo3"},
			targetDir:    "/path/to/backup",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := CreateFromArgs(
				tc.providerType,
				tc.serverURL,
				tc.accessToken,
				tc.username,
				tc.password,
				tc.useBasicAuth,
				tc.skipSSLValidation,
				tc.include,
				tc.exclude,
				tc.targetDir,
			)

			// Verify config
			if len(cfg.Providers) != 1 {
				t.Errorf("Expected 1 provider, got %d", len(cfg.Providers))
				return
			}

			provider := cfg.Providers[0]

			// Check provider type
			if string(provider.Type) != tc.providerType {
				t.Errorf("Expected type %s, got %s", tc.providerType, provider.Type)
			}

			// Check server URL
			if provider.ServerURL != tc.serverURL {
				t.Errorf("Expected server_url %s, got %s", tc.serverURL, provider.ServerURL)
			}

			// Check authentication
			if provider.AccessToken != tc.accessToken {
				t.Errorf("Expected access_token %s, got %s", tc.accessToken, provider.AccessToken)
			}
			if provider.Username != tc.username {
				t.Errorf("Expected username %s, got %s", tc.username, provider.Username)
			}
			if provider.Password != tc.password {
				t.Errorf("Expected password %s, got %s", tc.password, provider.Password)
			}
			if provider.UseBasicAuth != tc.useBasicAuth {
				t.Errorf("Expected use_basic_auth %v, got %v", tc.useBasicAuth, provider.UseBasicAuth)
			}
			if provider.SkipSslValidation != tc.skipSSLValidation {
				t.Errorf("Expected skip_ssl_validation %v, got %v", tc.skipSSLValidation, provider.SkipSslValidation)
			}

			// Check include/exclude
			if len(provider.Include) != len(tc.include) {
				t.Errorf("Expected %d include entries, got %d", len(tc.include), len(provider.Include))
			} else {
				for i, inc := range tc.include {
					if provider.Include[i] != inc {
						t.Errorf("Expected include[%d] %s, got %s", i, inc, provider.Include[i])
					}
				}
			}

			if len(provider.Exclude) != len(tc.exclude) {
				t.Errorf("Expected %d exclude entries, got %d", len(tc.exclude), len(provider.Exclude))
			} else {
				for i, exc := range tc.exclude {
					if provider.Exclude[i] != exc {
						t.Errorf("Expected exclude[%d] %s, got %s", i, exc, provider.Exclude[i])
					}
				}
			}

			// Check target dir
			if provider.TargetDir != tc.targetDir {
				t.Errorf("Expected target_dir %s, got %s", tc.targetDir, provider.TargetDir)
			}
		})
	}
}
