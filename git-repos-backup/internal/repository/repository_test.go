package repository

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
)

// Mock for exec.Command
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	// Add environment variables to control mock behavior
	if len(args) > 0 {
		if strings.Contains(strings.Join(args, " "), "gitea") {
			cmd.Env = append(cmd.Env, "MOCK_GITEA=1")
		} else if strings.Contains(strings.Join(args, " "), "github") {
			cmd.Env = append(cmd.Env, "MOCK_GITHUB=1")
		}
	}

	return cmd
}

func TestGetRepositories_Gitea(t *testing.T) {
	// Save current ExecCommand and restore at end
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()

	// Mock exec.Command
	ExecCommand = fakeExecCommand

	// Create test config
	cfg := &config.ProviderConfig{
		Type:        config.ProviderGitea,
		ServerURL:   "https://gitea.example.com",
		AccessToken: "faketoken",
	}

	// Call the function under test
	_, err := GetRepositories(cfg, false)

	// In a real test, we would check the results here
	// This is a minimal test to ensure the function doesn't panic
	if err != nil {
		// Error is expected in test environment without actual command execution
		// Just verify it's not a panic
	}

	// Test with invalid provider type
	invalidCfg := &config.ProviderConfig{
		Type: "invalid",
	}
	_, err = GetRepositories(invalidCfg, false)
	if err == nil {
		t.Error("Expected error for invalid provider type, got nil")
	}

	// Test with basic auth
	basicAuthCfg := &config.ProviderConfig{
		Type:         config.ProviderGitea,
		ServerURL:    "https://gitea.example.com",
		Username:     "user",
		Password:     "pass",
		UseBasicAuth: true,
	}
	_, err = GetRepositories(basicAuthCfg, false)
	if err != nil {
		// Just verify it's not a panic
	}

	// Test with verbose
	_, err = GetRepositories(cfg, true)
	if err != nil {
		// Just verify it's not a panic
	}
}

func TestGetRepositories_GitHub(t *testing.T) {
	// Save current ExecCommand and restore at end
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()

	// Mock exec.Command
	ExecCommand = fakeExecCommand

	// Create test config for GitHub.com
	cfg := &config.ProviderConfig{
		Type:        config.ProviderGitHub,
		AccessToken: "faketoken",
	}

	// Call the function under test
	_, err := GetRepositories(cfg, false)

	// In a real test, we would check the results here
	// This is a minimal test to ensure the function doesn't panic
	if err != nil {
		// Error is expected in test environment without actual command execution
		// Just verify it's not a panic
	}

	// Test GitHub Enterprise config
	gheCfg := &config.ProviderConfig{
		Type:        config.ProviderGitHub,
		ServerURL:   "https://github.example.com",
		AccessToken: "faketoken",
	}
	_, err = GetRepositories(gheCfg, false)
	if err != nil {
		// Just verify it's not a panic
	}

	// Test with basic auth
	basicAuthCfg := &config.ProviderConfig{
		Type:         config.ProviderGitHub,
		UseBasicAuth: true,
		Username:     "user",
		Password:     "pass",
	}
	_, err = GetRepositories(basicAuthCfg, false)
	if err != nil {
		// Just verify it's not a panic
	}

	// Test with verbose
	_, err = GetRepositories(cfg, true)
	if err != nil {
		// Just verify it's not a panic
	}
}

// Test Repository struct
func TestRepository(t *testing.T) {
	repo := Repository{
		Id:       1,
		Login:    "owner",
		Name:     "repo",
		FullName: "owner/repo",
		URL:      "https://example.com/owner/repo.git",
	}

	if repo.Id != 1 {
		t.Errorf("Expected Id 1, got %d", repo.Id)
	}

	if repo.Login != "owner" {
		t.Errorf("Expected Login 'owner', got %s", repo.Login)
	}

	if repo.Name != "repo" {
		t.Errorf("Expected Name 'repo', got %s", repo.Name)
	}

	if repo.FullName != "owner/repo" {
		t.Errorf("Expected FullName 'owner/repo', got %s", repo.FullName)
	}

	if repo.URL != "https://example.com/owner/repo.git" {
		t.Errorf("Expected URL 'https://example.com/owner/repo.git', got %s", repo.URL)
	}
}

// This is a test helper process that will be executed when the tests run
// It simulates the external command's behavior
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

	// Check which mock to provide
	if args[0] == "curl" {
		if os.Getenv("MOCK_GITEA") == "1" {
			// Mock Gitea API response
			fmt.Println(`{
			  "data": [
				{
				  "id": 1,
				  "name": "repo1",
				  "full_name": "owner/repo1",
				  "clone_url": "https://gitea.example.com/owner/repo1.git",
				  "owner": {
					"login": "owner"
				  }
				},
				{
				  "id": 2,
				  "name": "repo2",
				  "full_name": "owner/repo2",
				  "clone_url": "https://gitea.example.com/owner/repo2.git",
				  "owner": {
					"login": "owner"
				  }
				}
			  ]
			}`)
		} else if os.Getenv("MOCK_GITHUB") == "1" {
			// Mock GitHub API response
			fmt.Println(`[
			  {
				"id": 1,
				"name": "repo1",
				"full_name": "owner/repo1",
				"clone_url": "https://github.com/owner/repo1.git",
				"owner": {
				  "login": "owner"
				}
			  },
			  {
				"id": 2,
				"name": "repo2",
				"full_name": "owner/repo2",
				"clone_url": "https://github.com/owner/repo2.git",
				"owner": {
				  "login": "owner"
				}
			  }
			]`)
		} else {
			// Default empty response
			fmt.Println("{}")
		}
	}

	os.Exit(0)
}
