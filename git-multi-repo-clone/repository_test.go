package main

import (
	"encoding/json"
	"os/exec"
	"testing"
)

// Create a mock for the exec.Command used in getRepositories
func mockCurlSuccess() func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		// Create a sample API response
		apiResponse := struct {
			Data []Repository `json:"data"`
		}{
			Data: []Repository{
				{Name: "repo1", URL: "https://example.com/repo1.git"},
				{Name: "repo2", URL: "https://example.com/repo2.git"},
			},
		}

		// Marshal the response to JSON
		jsonResponse, _ := json.Marshal(apiResponse)

		// Return a command that outputs this JSON
		cmd := exec.Command("echo", string(jsonResponse))
		return cmd
	}
}

func TestGetRepositories(t *testing.T) {
	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockCurlSuccess()

	// Run the function with a test config
	config := &Config{
		GiteaURL: "https://gitea.example.com",
		APIToken: "test-token",
	}

	repos, err := getRepositories(config)
	if err != nil {
		t.Fatalf("getRepositories returned error: %v", err)
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
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockCurlSuccess()

	// Run the function with a test config using basic auth
	config := &Config{
		GiteaURL:     "https://gitea.example.com",
		UseBasicAuth: true,
		Username:     "testuser",
		Password:     "testpass",
	}

	repos, err := getRepositories(config)
	if err != nil {
		t.Fatalf("getRepositories returned error: %v", err)
	}

	// We just verify no error occurred, since we can't easily check
	// the command args with our simple mock
	if len(repos) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(repos))
	}
}

func TestGetRepositoriesWithToken(t *testing.T) {
	// Save the original exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Set our mock
	execCommand = mockCurlSuccess()

	// Run the function with a test config using token auth
	config := &Config{
		GiteaURL:     "https://gitea.example.com",
		UseBasicAuth: false,
		APIToken:     "test-token",
	}

	repos, err := getRepositories(config)
	if err != nil {
		t.Fatalf("getRepositories returned error: %v", err)
	}

	// We just verify no error occurred, since we can't easily check
	// the command args with our simple mock
	if len(repos) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(repos))
	}
}