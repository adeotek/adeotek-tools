package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These are integration tests that test the actual behavior of the program
// They require a test Gitea server to be running, or they can be skipped

func TestIntegrationClone(t *testing.T) {
	// Skip this test unless we explicitly want to run integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Create temporary directories for testing
	tempConfigDir, err := os.MkdirTemp("", "config-dir")
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}
	defer os.RemoveAll(tempConfigDir)

	tempTargetDir, err := os.MkdirTemp("", "target-dir")
	if err != nil {
		t.Fatalf("Failed to create temp target directory: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	// Create test config file
	// Note: This requires a real Gitea server for integration testing
	configPath := filepath.Join(tempConfigDir, "config.yaml")
	configContent := `
gitea_url: "https://gitea.example.com"  # Replace with a real test server
api_token: "test-token"                 # Replace with a real token
target_dir: "` + tempTargetDir + `"
username: "test-user"                   # Replace with a real username
password: "test-pass"                   # Replace with a real password
use_basic_auth: true
override_exising_local_repos: true
clone_as_mirror: false
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Get the path to the git-multi-repo-clone binary
	// For this test to work, you need to build the binary first:
	// go build -o git-multi-repo-clone
	binPath := filepath.Join("..", "git-multi-repo-clone")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		// If binary doesn't exist, build it
		buildCmd := exec.Command("go", "build", "-o", binPath, "..")
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build binary: %v\nOutput: %s", err, buildOutput)
		}
	}

	// Run the program with our test config
	cmd := exec.Command(binPath, "-config", configPath)
	output, err := cmd.CombinedOutput()

	// Check for expected output
	if err != nil {
		t.Errorf("Program execution failed: %v\nOutput: %s", err, output)
	}

	// Verify that repositories were cloned
	// This depends on your test server having repositories
	files, err := os.ReadDir(tempTargetDir)
	if err != nil {
		t.Fatalf("Failed to read target directory: %v", err)
	}

	// Just check if any repositories were cloned
	repoFound := false
	for _, file := range files {
		if file.IsDir() && isGitRepo(filepath.Join(tempTargetDir, file.Name())) {
			repoFound = true
			break
		}
	}

	if !repoFound {
		t.Errorf("No repositories were cloned. Output: %s", output)
	}
}

// isGitRepo checks if a directory is a git repository
func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		return true
	}

	// For mirror repositories, the directory itself is a .git directory
	if strings.HasSuffix(path, ".git") {
		return true
	}

	return false
}