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

// Page sizes used when listing repositories. Gitea caps the page size at
// MAX_RESPONSE_ITEMS (50 by default) while GitHub allows up to 100.
const (
	giteaPageSize  = 50
	githubPageSize = 100
	maxPages       = 1000
)

// apiRepository is the common shape of a repository in the Gitea and GitHub APIs.
type apiRepository struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// GetRepositories retrieves repositories from a Git provider
func GetRepositories(provider *config.ProviderConfig, verbose bool) ([]Repository, error) {
	switch provider.Type {
	case config.ProviderGitea, config.ProviderForgejo:
		// Forgejo is a Gitea fork and exposes the same API.
		return getGiteaCompatibleRepositories(provider, verbose)
	case config.ProviderGitHub:
		return getGitHubRepositories(provider, verbose)
	default:
		return nil, fmt.Errorf("unsupported provider type %q (supported: gitea, forgejo, github)", provider.Type)
	}
}

// getGiteaCompatibleRepositories retrieves all repositories from a Gitea or
// Forgejo server. Forgejo is a Gitea fork and exposes the same API.
func getGiteaCompatibleRepositories(provider *config.ProviderConfig, verbose bool) ([]Repository, error) {
	if provider.ServerURL == "" {
		return nil, fmt.Errorf("server_url is required for the %s provider", provider.Type)
	}
	baseURL := strings.TrimSuffix(provider.ServerURL, "/") + "/api/v1/repos/search"

	repos := make([]Repository, 0)
	seen := make(map[int]bool)

	for page := 1; page <= maxPages; page++ {
		apiURL := fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page, giteaPageSize)

		output, err := runCurl(provider, apiURL, "application/json", "", verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories from %s: %w", provider.Type, err)
		}

		var response struct {
			Data []apiRepository `json:"data"`
		}
		if err := json.Unmarshal(output, &response); err != nil {
			return nil, fmt.Errorf("failed to parse %s API response: %w", provider.Type, err)
		}

		added := appendRepositories(&repos, seen, response.Data)
		if len(response.Data) == 0 || len(response.Data) < giteaPageSize || added == 0 {
			break
		}
	}

	return repos, nil
}

// getGitHubRepositories retrieves all repositories from GitHub.
func getGitHubRepositories(provider *config.ProviderConfig, verbose bool) ([]Repository, error) {
	// Construct API URL (GitHub API v3)
	apiURL := "https://api.github.com/user/repos"
	if provider.ServerURL != "" && !strings.Contains(provider.ServerURL, "github.com") {
		// For GitHub Enterprise
		apiURL = strings.TrimSuffix(provider.ServerURL, "/") + "/api/v3/user/repos"
	}

	repos := make([]Repository, 0)
	seen := make(map[int]bool)

	for page := 1; page <= maxPages; page++ {
		// affiliation is left at its default (owner, collaborator and
		// organization_member) so that every accessible repository is listed.
		pageURL := fmt.Sprintf("%s?per_page=%d&page=%d", apiURL, githubPageSize, page)

		output, err := runCurl(provider, pageURL, "application/vnd.github+json", "2022-11-28", verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories from GitHub: %w", err)
		}

		var response []apiRepository
		if err := json.Unmarshal(output, &response); err != nil {
			return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
		}

		added := appendRepositories(&repos, seen, response)
		if len(response) == 0 || len(response) < githubPageSize || added == 0 {
			break
		}
	}

	return repos, nil
}

// appendRepositories adds the repositories that have not been seen yet and
// returns how many were added.
func appendRepositories(repos *[]Repository, seen map[int]bool, page []apiRepository) int {
	added := 0
	for _, r := range page {
		if seen[r.Id] {
			continue
		}
		seen[r.Id] = true
		*repos = append(*repos, Repository{
			Id:       r.Id,
			Login:    r.Owner.Login,
			Name:     r.Name,
			FullName: r.FullName,
			URL:      r.CloneURL,
		})
		added++
	}
	return added
}

// runCurl performs an authenticated GET request against the provider API and
// returns the response body. curl is invoked with --fail so that HTTP error
// responses produce an error instead of being parsed as JSON.
func runCurl(provider *config.ProviderConfig, apiURL string, accept string, apiVersion string, verbose bool) ([]byte, error) {
	args := []string{"--silent", "--show-error", "--fail", "--location"}

	if provider.SkipSslValidation {
		args = append(args, "--insecure")
	}

	args = append(args, "-X", "GET", apiURL)
	args = append(args, "-H", "accept: "+accept)
	if apiVersion != "" {
		args = append(args, "-H", "X-GitHub-Api-Version: "+apiVersion)
	}

	// Add authentication
	authHeader := ""
	if provider.UseBasicAuth {
		args = append(args, "-u", fmt.Sprintf("%s:%s", provider.Username, provider.Password))
	} else if provider.AccessToken != "" {
		// Gitea and Forgejo expect "Authorization: token <token>"; GitHub uses Bearer.
		scheme := "token"
		if provider.Type == config.ProviderGitHub {
			scheme = "Bearer"
		}
		authHeader = "Authorization: " + scheme + " " + provider.AccessToken
		args = append(args, "-H", authHeader)
	}

	if verbose {
		fmt.Printf("----> %s\n", redactedCurlCommand(args, authHeader))
	}

	cmd := ExecCommand("curl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// redactedCurlCommand formats a curl command line for logging without exposing
// the authentication credentials.
func redactedCurlCommand(args []string, authHeader string) string {
	redacted := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case authHeader != "" && arg == authHeader:
			redacted = append(redacted, "Authorization: ***")
		case arg == "-u":
			redacted = append(redacted, arg, "***")
			i++ // skip the credentials argument
		default:
			redacted = append(redacted, arg)
		}
	}
	return "curl " + strings.Join(redacted, " ")
}
