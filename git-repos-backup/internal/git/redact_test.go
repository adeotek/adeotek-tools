package git

import (
	"strings"
	"testing"
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

func TestGetRepoUrl_StripsCredentials(t *testing.T) {
	tests := []struct {
		name   string
		rawUrl string
		want   string
	}{
		{
			name:   "token auth",
			rawUrl: "https://token123@example.com/owner/repo.git",
			want:   "https://example.com/owner/repo.git",
		},
		{
			name:   "basic auth with special characters",
			rawUrl: "https://user%40corp:p%40ss%3Aword@example.com/owner/repo.git",
			want:   "https://example.com/owner/repo.git",
		},
		{
			name:   "no credentials unchanged",
			rawUrl: "https://example.com/owner/repo.git",
			want:   "https://example.com/owner/repo.git",
		},
		{
			name:   "SSH user is not a credential",
			rawUrl: "ssh://git@example.com/owner/repo.git",
			want:   "ssh://git@example.com/owner/repo.git",
		},
		{
			name:   "scp-like URL unchanged",
			rawUrl: "git@example.com:owner/repo.git",
			want:   "git@example.com:owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRepoUrl(tt.rawUrl)
			if err != nil {
				t.Fatalf("GetRepoUrl() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GetRepoUrl() = %q, want %q", got, tt.want)
			}
		})
	}
}
