package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Mock commands for testing
func mockGitSuccessCommand(command string, args ...string) *exec.Cmd {
	// Create a fake successful command
	cmd := exec.Command("echo", "Success")
	return cmd
}

func TestCloneRepositoryNewClone(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-clone")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup test data
	config := &Config{
		TargetDir:                 tempDir,
		UseBasicAuth:              false,
		OverrideExistingLocalRepos: false,
		CloneAsMirror:             false,
	}
	repo := Repository{
		Name: "test-repo",
		URL:  "https://example.com/test-repo.git",
	}

	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockGitSuccessCommand

	// Run the function
	err = cloneRepository(config, repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCloneRepositoryAsMirror(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-clone-mirror")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup test data
	config := &Config{
		TargetDir:                 tempDir,
		UseBasicAuth:              false,
		OverrideExistingLocalRepos: false,
		CloneAsMirror:             true,
	}
	repo := Repository{
		Name: "test-repo",
		URL:  "https://example.com/test-repo.git",
	}

	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockGitSuccessCommand

	// Run the function
	err = cloneRepository(config, repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCloneRepositoryWithBasicAuth(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-basic-auth")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup test data
	config := &Config{
		TargetDir:                 tempDir,
		UseBasicAuth:              true,
		Username:                  "testuser",
		Password:                  "testpass",
		OverrideExistingLocalRepos: false,
		CloneAsMirror:             false,
	}
	repo := Repository{
		Name: "test-repo",
		URL:  "https://example.com/test-repo.git",
	}

	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockGitSuccessCommand

	// Run the function
	err = cloneRepository(config, repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCloneExistingRepositoryWithOverride(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-override")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a fake repository directory
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create test repo directory: %v", err)
	}

	// Setup test data for mirror mode
	config := &Config{
		TargetDir:                  tempDir,
		UseBasicAuth:               false,
		OverrideExistingLocalRepos: true,
		CloneAsMirror:              true,
	}
	repo := Repository{
		Name: "test-repo",
		URL:  "https://example.com/test-repo.git",
	}

	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockGitSuccessCommand

	// Run the function
	err = cloneRepository(config, repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify directory was deleted and recreated (which we simulate by checking if it exists)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Error("Repository directory should exist")
	}
}

func TestCloneExistingRepositoryStandardMode(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-standard")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a fake repository directory
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create test repo directory: %v", err)
	}

	// Setup test data for standard mode with override
	config := &Config{
		TargetDir:                  tempDir,
		UseBasicAuth:               false,
		OverrideExistingLocalRepos: true,
		CloneAsMirror:              false,
	}
	repo := Repository{
		Name: "test-repo",
		URL:  "https://example.com/test-repo.git",
	}

	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockGitSuccessCommand

	// Run the function
	err = cloneRepository(config, repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify directory still exists (not deleted since it's standard mode)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Error("Repository directory should still exist")
	}
}