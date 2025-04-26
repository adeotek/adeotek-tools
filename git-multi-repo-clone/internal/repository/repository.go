// Package repository handles interaction with the Gitea API
package repository

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/adeotek/git-multi-repo-clone/internal/config"
)

// Repository represents a Git repository from the Gitea API
type Repository struct {
	Id       int    `json:"id"`
	Login    string `json:"owner.login"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"clone_url"`
}

// ExecCommand is a variable that holds the exec.Command function.
// It can be replaced in tests to mock command execution.
var ExecCommand = exec.Command

// GetRepositories retrieves a list of repositories from the Gitea server
func GetRepositories(config *config.Config, verbose bool) ([]Repository, error) {
	// Construct API URL
	apiURL := fmt.Sprintf("%s/api/v1/repos/search", config.ServerURL)

	// Prepare curl command
	cmd := ExecCommand("curl", "-s")

	if config.SkipSslValidation {
		cmd.Args = append(cmd.Args, "--insecure")
	}

	cmd.Args = append(cmd.Args, "-X", "GET") // Add method
	cmd.Args = append(cmd.Args, apiURL)      // Add API URL
	cmd.Args = append(cmd.Args, "-H", "accept: application/json")

	// Add authentication
	if config.UseBasicAuth {
		cmd.Args = append(cmd.Args, "-u", fmt.Sprintf("%s:%s", config.Username, config.Password))
	} else if config.AccessToken != "" {
		cmd.Args = append(cmd.Args, "-H", fmt.Sprintf("Authorization: token %s", config.AccessToken))
	}

	if verbose {
		fmt.Printf("----> %s\n", cmd.String())
	}

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Parse response
	var response struct {
		Data []Repository `json:"data"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return response.Data, nil
}
