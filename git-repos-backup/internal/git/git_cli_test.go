package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
)

// Mock for exec.Command
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// Test helper process that mocks command execution
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Get the command and args that were passed to exec.Command
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	// Mock git command responses
	switch args[0] {
	case "git":
		// For simplicity, all git commands succeed in our tests
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

func TestGetRepoUrl(t *testing.T) {
	tests := []struct {
		name      string
		rawUrl    string
		want      string
		wantError bool
	}{
		{
			name:   "HTTP URL stays clean",
			rawUrl: "http://example.com/repo.git",
			want:   "http://example.com/repo.git",
		},
		{
			name:   "HTTPS URL stays clean",
			rawUrl: "https://example.com/repo.git",
			want:   "https://example.com/repo.git",
		},
		{
			name:   "HTTPS URL with token is stripped",
			rawUrl: "https://token123@example.com/repo.git",
			want:   "https://example.com/repo.git",
		},
		{
			name:   "HTTPS URL with basic auth is stripped",
			rawUrl: "https://user:pass@example.com/repo.git",
			want:   "https://example.com/repo.git",
		},
		{
			name:   "escaped basic auth is stripped",
			rawUrl: "https://user%40corp:p%40ss%3Aword@example.com/repo.git",
			want:   "https://example.com/repo.git",
		},
		{
			name:   "SSH URL keeps its user",
			rawUrl: "ssh://git@example.com/owner/repo.git",
			want:   "ssh://git@example.com/owner/repo.git",
		},
		{
			name:   "scp-like URL keeps its user",
			rawUrl: "git@github.com:user/repo.git",
			want:   "git@github.com:user/repo.git",
		},
		{
			name:   "absolute local path",
			rawUrl: "/srv/git/repo.git",
			want:   "/srv/git/repo.git",
		},
		{
			name:   "relative local path",
			rawUrl: "../git/repo.git",
			want:   "../git/repo.git",
		},
		{
			name:      "Invalid URL",
			rawUrl:    "git",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRepoUrl(tt.rawUrl)

			// Check error
			if (err != nil) != tt.wantError {
				t.Errorf("GetRepoUrl() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Check result
			if got != tt.want {
				t.Errorf("GetRepoUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRepoPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test creating path for a repo
	targetDir := tmpDir
	userName := "testuser"
	repoName := "testrepo"

	repoPath, err := GetRepoPath(targetDir, userName, repoName, false)
	if err != nil {
		t.Fatalf("GetRepoPath() error = %v", err)
	}

	expectedPath := filepath.Join(targetDir, userName, repoName)
	if repoPath != expectedPath {
		t.Errorf("GetRepoPath() = %v, want %v", repoPath, expectedPath)
	}

	// Check if directories were created
	userDir := filepath.Join(targetDir, userName)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		t.Errorf("User directory was not created: %s", userDir)
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Errorf("Repo directory was not created: %s", repoPath)
	}

	// Test with verbose mode
	verbosePath, err := GetRepoPath(targetDir, "verboseuser", "verboserepo", true)
	if err != nil {
		t.Fatalf("GetRepoPath() with verbose error = %v", err)
	}

	expectedVerbosePath := filepath.Join(targetDir, "verboseuser", "verboserepo")
	if verbosePath != expectedVerbosePath {
		t.Errorf("GetRepoPath() with verbose = %v, want %v", verbosePath, expectedVerbosePath)
	}
}

func TestRepoExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test non-existent repo
	nonExistentPath := filepath.Join(tmpDir, "nonexistent")
	if RepoExists(nonExistentPath, false) {
		t.Errorf("RepoExists() for non-existent repo returned true, want false")
	}

	// Test repo with no HEAD file
	emptyRepoPath := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyRepoPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if RepoExists(emptyRepoPath, false) {
		t.Errorf("RepoExists() for repo without HEAD returned true, want false")
	}

	// Test repo with empty HEAD file
	repoWithEmptyHEAD := filepath.Join(tmpDir, "emptyhead")
	if err := os.MkdirAll(repoWithEmptyHEAD, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	headPath := filepath.Join(repoWithEmptyHEAD, "HEAD")
	if err := os.WriteFile(headPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty HEAD file: %v", err)
	}
	if RepoExists(repoWithEmptyHEAD, false) {
		t.Errorf("RepoExists() for repo with empty HEAD returned true, want false")
	}

	// Test valid repo
	validRepoPath := filepath.Join(tmpDir, "valid")
	if err := os.MkdirAll(validRepoPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	validHeadPath := filepath.Join(validRepoPath, "HEAD")
	if err := os.WriteFile(validHeadPath, []byte("ref: refs/heads/main"), 0644); err != nil {
		t.Fatalf("Failed to create HEAD file: %v", err)
	}
	if !RepoExists(validRepoPath, false) {
		t.Errorf("RepoExists() for valid repo returned false, want true")
	}

	// Test with verbose mode
	RepoExists(validRepoPath, true) // Just ensure it doesn't panic
}

func TestGetGitCommand(t *testing.T) {
	provider := &config.ProviderConfig{
		SkipSslValidation: false,
	}

	// Test basic command
	cmd := GetGitCommand(provider, "status")
	if !strings.Contains(cmd.String(), "git status") {
		t.Errorf("GetGitCommand() = %v, should contain 'git status'", cmd.String())
	}

	// Test with SSL validation disabled
	providerWithSkipSSL := &config.ProviderConfig{
		SkipSslValidation: true,
	}
	cmdWithSkipSSL := GetGitCommand(providerWithSkipSSL, "status")
	if !strings.Contains(cmdWithSkipSSL.String(), "http.sslVerify=false") {
		t.Errorf("GetGitCommand() with SkipSslValidation = %v, should contain 'http.sslVerify=false'", cmdWithSkipSSL.String())
	}

	// Test with multiple arguments
	cmdWithMultipleArgs := GetGitCommand(provider, "fetch", "--prune", "origin")
	if !strings.Contains(cmdWithMultipleArgs.String(), "git fetch --prune origin") {
		t.Errorf("GetGitCommand() = %v, should contain 'git fetch --prune origin'", cmdWithMultipleArgs.String())
	}
}

func TestGetGitCommandForRemote(t *testing.T) {
	tests := []struct {
		name             string
		provider         *config.ProviderConfig
		repoUrl          string
		wantUsername     string
		wantPassword     string
		wantNoCredential bool
	}{
		{
			name:         "token auth over HTTP",
			provider:     &config.ProviderConfig{Type: config.ProviderGitHub, AccessToken: "token123"},
			repoUrl:      "http://github.com/owner/repo.git",
			wantUsername: "token123",
		},
		{
			name:         "token auth over HTTPS",
			provider:     &config.ProviderConfig{Type: config.ProviderForgejo, AccessToken: "token123"},
			repoUrl:      "https://github.com/owner/repo.git",
			wantUsername: "token123",
		},
		{
			name: "basic auth",
			provider: &config.ProviderConfig{
				Type:         config.ProviderGitea,
				UseBasicAuth: true,
				Username:     "user@corp",
				Password:     "p@ss:word",
			},
			repoUrl:      "https://gitea.example.com/owner/repo.git",
			wantUsername: "user@corp",
			wantPassword: "p@ss:word",
		},
		{
			name:             "no configured credentials",
			provider:         &config.ProviderConfig{},
			repoUrl:          "https://example.com/owner/repo.git",
			wantNoCredential: true,
		},
		{
			name:             "SSH remote is left alone",
			provider:         &config.ProviderConfig{AccessToken: "token123"},
			repoUrl:          "ssh://git@example.com/owner/repo.git",
			wantNoCredential: true,
		},
		{
			name:             "scp-like remote is left alone",
			provider:         &config.ProviderConfig{AccessToken: "token123"},
			repoUrl:          "git@example.com:owner/repo.git",
			wantNoCredential: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := GetGitCommandForRemote(tt.provider, tt.repoUrl, "fetch", "origin")
			joined := strings.Join(cmd.Args, " ")

			if tt.wantNoCredential {
				if strings.Contains(joined, "credential.helper") || hasEnv(cmd, credentialEnvUsername) {
					t.Fatalf("command unexpectedly contains credentials: args=%v env=%v", cmd.Args, cmd.Env)
				}
			} else {
				if !strings.Contains(joined, "credential.helper="+credentialHelper) {
					t.Errorf("command does not use the credential helper: %v", cmd.Args)
				}
				if got := envValue(cmd, credentialEnvUsername); got != tt.wantUsername {
					t.Errorf("username env = %q, want %q", got, tt.wantUsername)
				}
				if got := envValue(cmd, credentialEnvPassword); got != tt.wantPassword {
					t.Errorf("password env = %q, want %q", got, tt.wantPassword)
				}
			}

			if !strings.HasSuffix(joined, "fetch origin") {
				t.Errorf("git arguments were not preserved: %v", cmd.Args)
			}
		})
	}
}

func TestCredentialHelperAppliesToAllProviders(t *testing.T) {
	for _, providerType := range config.SupportedProviders {
		t.Run(string(providerType)+"/token", func(t *testing.T) {
			provider := &config.ProviderConfig{
				Type:        providerType,
				AccessToken: "token123",
			}
			assertCredentialHelper(t, provider, "token123", "")
		})

		t.Run(string(providerType)+"/basic", func(t *testing.T) {
			provider := &config.ProviderConfig{
				Type:         providerType,
				UseBasicAuth: true,
				Username:     "user",
				Password:     "pass",
			}
			assertCredentialHelper(t, provider, "user", "pass")
		})
	}
}

func assertCredentialHelper(t *testing.T, provider *config.ProviderConfig, wantUsername, wantPassword string) {
	t.Helper()

	repoUrl := "https://git.example.com/owner/repo.git"
	cmd := GetGitCommandForRemote(provider, repoUrl, "fetch", repoUrl)
	joined := strings.Join(cmd.Args, " ")

	if !strings.Contains(joined, "credential.helper="+credentialHelper) {
		t.Fatalf("command does not use the credential helper: %v", cmd.Args)
	}
	if got := envValue(cmd, credentialEnvUsername); got != wantUsername {
		t.Errorf("username env = %q, want %q", got, wantUsername)
	}
	if got := envValue(cmd, credentialEnvPassword); got != wantPassword {
		t.Errorf("password env = %q, want %q", got, wantPassword)
	}
	if !strings.Contains(joined, repoUrl) {
		t.Errorf("clean repository URL was not passed to git: %v", cmd.Args)
	}
}

func TestRunGitFetchUsesCredentialHelper(t *testing.T) {
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()

	var executed *exec.Cmd
	ExecCommand = func(command string, args ...string) *exec.Cmd {
		executed = fakeExecCommand(command, args...)
		return executed
	}

	provider := &config.ProviderConfig{
		Type:        config.ProviderGitHub,
		AccessToken: "token123",
	}
	repoUrl := "https://github.com/owner/repo.git"

	if err := RunGitFetch(provider, t.TempDir(), repoUrl, "owner/repo", false); err != nil {
		t.Fatalf("RunGitFetch() error = %v", err)
	}
	if executed == nil {
		t.Fatal("RunGitFetch() did not execute git")
	}

	joined := strings.Join(executed.Args, " ")
	if !strings.Contains(joined, "credential.helper="+credentialHelper) {
		t.Errorf("git fetch did not receive the credential helper: %v", executed.Args)
	}
	if strings.Contains(joined, "token123@") {
		t.Errorf("git fetch URL contains embedded credentials: %v", executed.Args)
	}
	if !strings.Contains(joined, repoUrl) {
		t.Errorf("git fetch did not receive the clean repository URL: %v", executed.Args)
	}
	if got := envValue(executed, credentialEnvUsername); got != "token123" {
		t.Errorf("username env = %q, want %q", got, "token123")
	}
}

func envValue(cmd *exec.Cmd, key string) string {
	prefix := key + "="
	for _, entry := range cmd.Env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func hasEnv(cmd *exec.Cmd, key string) bool {
	prefix := key + "="
	for _, entry := range cmd.Env {
		if strings.HasPrefix(entry, prefix) {
			return true
		}
	}
	return false
}

func TestFetchRepository(t *testing.T) {
	// Save the original ExecCommand and restore it after the test
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()

	// Replace exec.Command with our fake
	ExecCommand = fakeExecCommand

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test provider config
	provider := &config.ProviderConfig{
		Type:        config.ProviderGitea,
		ServerURL:   "https://gitea.example.com",
		AccessToken: "faketoken",
		TargetDir:   tmpDir,
	}

	// Create test repository
	repo := repository.Repository{
		Id:       1,
		Login:    "testuser",
		Name:     "testrepo",
		FullName: "testuser/testrepo",
		URL:      "https://gitea.example.com/testuser/testrepo.git",
	}

	// Create directories and files needed for the test
	repoDir := filepath.Join(tmpDir, repo.Login, repo.Name)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	headPath := filepath.Join(repoDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main"), 0644); err != nil {
		t.Fatalf("Failed to create HEAD file: %v", err)
	}

	// Test the function (it should not panic with our mocks)
	err := FetchRepository(provider, repo, false)
	if err != nil {
		// Since we're mocking, we expect our command to fail but not panic
		// The important thing is that the function runs through its logic
	}

	// Also test with verbose mode
	err = FetchRepository(provider, repo, true)
	if err != nil {
		// Since we're mocking, we expect our command to fail but not panic
	}
}

func TestFetchRepositoryStripsURLCredentialsForAllProviders(t *testing.T) {
	for _, providerType := range config.SupportedProviders {
		t.Run(string(providerType), func(t *testing.T) {
			oldExecCommand := ExecCommand
			defer func() { ExecCommand = oldExecCommand }()

			var commands []*exec.Cmd
			ExecCommand = func(command string, args ...string) *exec.Cmd {
				cmd := fakeExecCommand(command, args...)
				commands = append(commands, cmd)
				return cmd
			}

			provider := &config.ProviderConfig{
				Type:        providerType,
				AccessToken: "provider-token",
				TargetDir:   t.TempDir(),
			}
			repo := repository.Repository{
				Login:    "owner",
				Name:     "repo",
				FullName: "owner/repo",
				URL:      "https://url-user:url-password@example.com/owner/repo.git",
			}

			if err := FetchRepository(provider, repo, false); err != nil {
				t.Fatalf("FetchRepository() error = %v", err)
			}

			fetchSeen := false
			for _, cmd := range commands {
				joined := strings.Join(cmd.Args, " ")
				for _, secret := range []string{"url-user", "url-password", "provider-token@"} {
					if strings.Contains(joined, secret) {
						t.Errorf("git command contains embedded credentials %q: %v", secret, cmd.Args)
					}
				}
				if !hasArg(cmd.Args, "fetch") {
					continue
				}
				fetchSeen = true
				if !strings.Contains(joined, "credential.helper="+credentialHelper) {
					t.Errorf("fetch did not use the credential helper: %v", cmd.Args)
				}
				if !strings.Contains(joined, "https://example.com/owner/repo.git") {
					t.Errorf("fetch did not use the clean URL: %v", cmd.Args)
				}
			}
			if !fetchSeen {
				t.Fatal("FetchRepository() did not run git fetch")
			}
		})
	}
}

func hasArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
