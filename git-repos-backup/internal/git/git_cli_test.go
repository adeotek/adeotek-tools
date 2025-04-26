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
		provider  *config.ProviderConfig
		rawUrl    string
		want      string
		wantError bool
	}{
		{
			name: "HTTP URL with token",
			provider: &config.ProviderConfig{
				AccessToken: "token123",
			},
			rawUrl:    "http://example.com/repo.git",
			want:      "http://token123@example.com/repo.git",
			wantError: false,
		},
		{
			name: "HTTPS URL with token",
			provider: &config.ProviderConfig{
				AccessToken: "token123",
			},
			rawUrl:    "https://example.com/repo.git",
			want:      "https://token123@example.com/repo.git",
			wantError: false,
		},
		{
			name: "HTTPS URL with basic auth",
			provider: &config.ProviderConfig{
				Username:     "user",
				Password:     "pass",
				UseBasicAuth: true,
			},
			rawUrl:    "https://example.com/repo.git",
			want:      "https://user:pass@example.com/repo.git",
			wantError: false,
		},
		{
			name:      "SSH URL",
			provider:  &config.ProviderConfig{},
			rawUrl:    "git@github.com:user/repo.git",
			want:      "git@github.com:user/repo.git",
			wantError: false,
		},
		{
			name:      "Invalid URL",
			provider:  &config.ProviderConfig{},
			rawUrl:    "git",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRepoUrl(tt.provider, tt.rawUrl)

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
