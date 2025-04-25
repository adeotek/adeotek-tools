package repository

import (
	"os/exec"
	"testing"

	"github.com/adeotek/git-multi-repo-clone/internal/config"
)

// Create a mock for the exec.Command used in GetRepositories
func mockCurlSuccess() func(string, ...string) *exec.Cmd {
	// Create a simple function that returns a mock command
	// We'll use a shell script to output a fixed JSON response
	return func(command string, args ...string) *exec.Cmd {
		jsonResponse := `{"data":[{"name":"repo1","clone_url":"https://example.com/repo1.git"},{"name":"repo2","clone_url":"https://example.com/repo2.git"}]}`
		return exec.Command("bash", "-c", "echo -n '"+jsonResponse+"'")
	}
}

func TestGetRepositories(t *testing.T) {
	// Save the original exec.Command
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()

	// Set our mock
	ExecCommand = mockCurlSuccess()

	// Run the function with a test config
	config := &config.Config{
		GiteaURL: "https://gitea.example.com",
		APIToken: "test-token",
	}

	repos, err := GetRepositories(config)
	if err != nil {
		t.Fatalf("GetRepositories returned error: %v", err)
	}

	// Verify the results
	if len(repos) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(repos))
	}

	if repos[0].Name != "repo1" || repos[0].URL != "https://example.com/repo1.git" {
		t.Errorf("First repository doesn't match expected values: %+v", repos[0])
	}

	if repos[1].Name != "repo2" || repos[1].URL != "https://example.com/repo2.git" {
		t.Errorf("Second repository doesn't match expected values: %+v", repos[1])
	}
}

func TestGetRepositoriesWithBasicAuth(t *testing.T) {
	// Save the original exec.Command
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()

	// Set our mock
	ExecCommand = mockCurlSuccess()

	// Run the function with a test config using basic auth
	config := &config.Config{
		GiteaURL:     "https://gitea.example.com",
		UseBasicAuth: true,
		Username:     "testuser",
		Password:     "testpass",
	}

	repos, err := GetRepositories(config)
	if err != nil {
		t.Fatalf("GetRepositories returned error: %v", err)
	}

	// We just verify no error occurred, since we can't easily check
	// the command args with our simple mock
	if len(repos) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(repos))
	}
}

func TestGetRepositoriesWithToken(t *testing.T) {
	// Save the original exec.Command
	origExecCommand := ExecCommand
	defer func() { ExecCommand = origExecCommand }()

	// Set our mock
	ExecCommand = mockCurlSuccess()

	// Run the function with a test config using token auth
	config := &config.Config{
		GiteaURL:     "https://gitea.example.com",
		UseBasicAuth: false,
		APIToken:     "test-token",
	}

	repos, err := GetRepositories(config)
	if err != nil {
		t.Fatalf("GetRepositories returned error: %v", err)
	}

	// We just verify no error occurred, since we can't easily check
	// the command args with our simple mock
	if len(repos) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(repos))
	}
}
