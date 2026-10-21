package git

import (
	"strings"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
)

func TestRedactUrl(t *testing.T) {
	tests := []struct {
		name   string
		rawUrl string
		secret string
	}{
		{
			name:   "token auth",
			rawUrl: "https://ghp_secrettoken@example.com/owner/repo.git",
			secret: "ghp_secrettoken",
		},
		{
			name:   "basic auth",
			rawUrl: "https://user:s3cretpass@example.com/owner/repo.git",
			secret: "s3cretpass",
		},
		{
			name:   "no credentials",
			rawUrl: "https://example.com/owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactUrl(tt.rawUrl)
			if tt.secret != "" && strings.Contains(got, tt.secret) {
				t.Errorf("RedactUrl() = %q, must not contain %q", got, tt.secret)
			}
			if !strings.Contains(got, "example.com/owner/repo.git") {
				t.Errorf("RedactUrl() = %q, lost the host/path", got)
			}
		})
	}
}

func TestGetRepoUrl_Credentials(t *testing.T) {
	tests := []struct {
		name     string
		provider *config.ProviderConfig
		rawUrl   string
		want     string
	}{
		{
			name:     "token auth",
			provider: &config.ProviderConfig{AccessToken: "token123"},
			rawUrl:   "https://example.com/owner/repo.git",
			want:     "https://token123@example.com/owner/repo.git",
		},
		{
			name:     "basic auth with special characters",
			provider: &config.ProviderConfig{UseBasicAuth: true, Username: "user@corp", Password: "p@ss:word"},
			rawUrl:   "https://example.com/owner/repo.git",
			want:     "https://user%40corp:p%40ss%3Aword@example.com/owner/repo.git",
		},
		{
			name:     "basic auth without username falls back to URL",
			provider: &config.ProviderConfig{UseBasicAuth: true, Password: "pass", AccessToken: "tok"},
			rawUrl:   "https://example.com/owner/repo.git",
			want:     "https://tok@example.com/owner/repo.git",
		},
		{
			name:     "SSH scheme is left alone",
			provider: &config.ProviderConfig{AccessToken: "token123"},
			rawUrl:   "ssh://git@example.com/owner/repo.git",
			want:     "ssh://git@example.com/owner/repo.git",
		},
		{
			name:     "scp-like URL is left alone",
			provider: &config.ProviderConfig{AccessToken: "token123"},
			rawUrl:   "git@example.com:owner/repo.git",
			want:     "git@example.com:owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRepoUrl(tt.provider, tt.rawUrl)
			if err != nil {
				t.Fatalf("GetRepoUrl() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GetRepoUrl() = %q, want %q", got, tt.want)
			}
		})
	}
}
