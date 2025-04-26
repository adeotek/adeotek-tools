// Package repository handles interaction with Git provider APIs
package repository

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
)

// Repository represents a Git repository
type Repository struct {
	Id       int    `json:"id"`
	Login    string // Owner login
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string // Clone URL
}

// ExecCommand is a variable that holds the exec.Command function.
// It can be replaced in tests to mock command execution.
var ExecCommand = exec.Command

// GetRepositories retrieves repositories from a Git provider
func GetRepositories(provider *config.ProviderConfig, verbose bool) ([]Repository, error) {
	switch provider.Type {
	case config.ProviderGitea:
		return getGiteaRepositories(provider, verbose)
	case config.ProviderGitHub:
		return getGitHubRepositories(provider, verbose)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}

// getGiteaRepositories retrieves repositories from a Gitea server
func getGiteaRepositories(provider *config.ProviderConfig, verbose bool) ([]Repository, error) {
	// Construct API URL
	apiURL := fmt.Sprintf("%s/api/v1/repos/search", provider.ServerURL)

	// Prepare curl command
	cmd := ExecCommand("curl", "-s")

	if provider.SkipSslValidation {
		cmd.Args = append(cmd.Args, "--insecure")
	}

	cmd.Args = append(cmd.Args, "-X", "GET") // Add method
	cmd.Args = append(cmd.Args, apiURL)      // Add API URL
	cmd.Args = append(cmd.Args, "-H", "accept: application/json")

	// Add authentication
	if provider.UseBasicAuth {
		cmd.Args = append(cmd.Args, "-u", fmt.Sprintf("%s:%s", provider.Username, provider.Password))
	} else if provider.AccessToken != "" {
		cmd.Args = append(cmd.Args, "-H", fmt.Sprintf("Authorization: token %s", provider.AccessToken))
	}

	if verbose {
		fmt.Printf("----> %s\n", cmd.String())
	}

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from Gitea: %w", err)
	}

	// Parse response
	var response struct {
		Data []struct {
			Id       int    `json:"id"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			CloneURL string `json:"clone_url"`
			Owner    struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Gitea API response: %w", err)
	}

	// Convert to common Repository structure
	repos := make([]Repository, 0, len(response.Data))
	for _, r := range response.Data {
		repos = append(repos, Repository{
			Id:       r.Id,
			Login:    r.Owner.Login,
			Name:     r.Name,
			FullName: r.FullName,
			URL:      r.CloneURL,
		})
	}

	return repos, nil
}

// getGitHubRepositories retrieves repositories from GitHub
func getGitHubRepositories(provider *config.ProviderConfig, verbose bool) ([]Repository, error) {
	// Construct API URL (GitHub API v3)
	apiURL := "https://api.github.com/user/repos"
	if provider.ServerURL != "" && !strings.Contains(provider.ServerURL, "github.com") {
		// For GitHub Enterprise
		apiURL = fmt.Sprintf("%s/api/v3/user/repos", provider.ServerURL)
	}

	// Prepare curl command
	cmd := ExecCommand("curl", "-s")

	if provider.SkipSslValidation {
		cmd.Args = append(cmd.Args, "--insecure")
	}

	cmd.Args = append(cmd.Args, "-X", "GET") // Add method
	cmd.Args = append(cmd.Args, apiURL)      // Add API URL
	cmd.Args = append(cmd.Args, "-H", "accept: application/vnd.github+json")
	cmd.Args = append(cmd.Args, "-H", "X-GitHub-Api-Version: 2022-11-28")

	// Add authentication
	if provider.UseBasicAuth {
		cmd.Args = append(cmd.Args, "-u", fmt.Sprintf("%s:%s", provider.Username, provider.Password))
	} else if provider.AccessToken != "" {
		cmd.Args = append(cmd.Args, "-H", fmt.Sprintf("Authorization: Bearer %s", provider.AccessToken))
	}

	if verbose {
		fmt.Printf("----> %s\n", cmd.String())
	}

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from GitHub: %w", err)
	}

	// Parse response
	var response []struct {
		Id       int    `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	// Convert to common Repository structure
	repos := make([]Repository, 0, len(response))
	for _, r := range response {
		repos = append(repos, Repository{
			Id:       r.Id,
			Login:    r.Owner.Login,
			Name:     r.Name,
			FullName: r.FullName,
			URL:      r.CloneURL,
		})
	}

	return repos, nil
}
