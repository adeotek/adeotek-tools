package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GiteaURL                   string   `yaml:"gitea_url"`
	APIToken                   string   `yaml:"api_token"`
	TargetDir                  string   `yaml:"target_dir"`
	Username                   string   `yaml:"username"`
	Password                   string   `yaml:"password"`
	UseBasicAuth               bool     `yaml:"use_basic_auth"`
	OverrideExistingLocalRepos bool     `yaml:"override_exising_local_repos,omitempty"`
	CloneAsMirror              bool     `yaml:"clone_as_mirror,omitempty"`
	Include                    []string `yaml:"include,omitempty"`
	Exclude                    []string `yaml:"exclude,omitempty"`
}

func main() {
	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(config.TargetDir, 0755); err != nil {
		log.Fatalf("Failed to create target directory: %v", err)
	}

	// Get list of repositories
	repos, err := getRepositories(config)
	if err != nil {
		log.Fatalf("Failed to get repositories: %v", err)
	}

	// Filter repositories based on include/exclude lists
	repos = filterRepositories(repos, config)

	// Clone each repository
	for _, repo := range repos {
		if err := cloneRepository(config, repo); err != nil {
			log.Printf("Failed to clone repository %s: %v", repo.Name, err)
		}
	}
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

type Repository struct {
	Name string `json:"name"`
	URL  string `json:"clone_url"`
}

func getRepositories(config *Config) ([]Repository, error) {
	// Construct API URL
	apiURL := fmt.Sprintf("%s/api/v1/repos/search", config.GiteaURL)

	// Prepare curl command
	cmd := exec.Command("curl", "-s")

	// Add authentication
	if config.UseBasicAuth {
		cmd.Args = append(cmd.Args, "-u", fmt.Sprintf("%s:%s", config.Username, config.Password))
	} else if config.APIToken != "" {
		cmd.Args = append(cmd.Args, "-H", fmt.Sprintf("Authorization: token %s", config.APIToken))
	}

	// Add API URL
	cmd.Args = append(cmd.Args, apiURL)

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %v", err)
	}

	// Parse response
	var response struct {
		Data []Repository `json:"data"`
	}
	if err := yaml.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %v", err)
	}

	return response.Data, nil
}

// filterRepositories applies include and exclude filters from config
func filterRepositories(repos []Repository, config *Config) []Repository {
	// If neither include nor exclude is specified, return all repos
	if len(config.Include) == 0 && len(config.Exclude) == 0 {
		return repos
	}

	// Create a map for faster lookup
	includeMap := make(map[string]bool)
	for _, name := range config.Include {
		includeMap[name] = true
	}

	excludeMap := make(map[string]bool)
	for _, name := range config.Exclude {
		excludeMap[name] = true
	}

	var filtered []Repository
	for _, repo := range repos {
		// If include list is specified, only include repos in that list
		if len(config.Include) > 0 {
			if includeMap[repo.Name] {
				filtered = append(filtered, repo)
			}
		} else {
			// If exclude list is specified, exclude repos in that list
			if !excludeMap[repo.Name] {
				filtered = append(filtered, repo)
			}
		}
	}

	return filtered
}

func cloneRepository(config *Config, repo Repository) error {
	repoDir := filepath.Join(config.TargetDir, repo.Name)

	// Prepare clone URL with authentication if needed
	cloneURL := repo.URL
	if config.UseBasicAuth {
		// Insert username and password into URL
		cloneURL = fmt.Sprintf("https://%s:%s@%s",
			config.Username,
			config.Password,
			cloneURL[8:]) // Remove "https://" prefix
	}

	// Check if repository already exists
	if _, err := os.Stat(repoDir); err == nil {
		// Repository exists
		if config.OverrideExistingLocalRepos {
			if config.CloneAsMirror {
				// For mirror repos, delete and re-clone
				log.Printf("Deleting existing mirror repository: %s", repo.Name)
				if err := os.RemoveAll(repoDir); err != nil {
					return fmt.Errorf("failed to delete existing repository: %v", err)
				}
			} else {
				// For standard repos, update with fetch and pull
				log.Printf("Updating existing repository: %s", repo.Name)
				
				// Fetch with prune
				fetchCmd := exec.Command("git", "-C", repoDir, "fetch", "--prune")
				fetchCmd.Stdout = os.Stdout
				fetchCmd.Stderr = os.Stderr
				if err := fetchCmd.Run(); err != nil {
					return fmt.Errorf("failed to fetch updates: %v", err)
				}
				
				// Pull updates
				pullCmd := exec.Command("git", "-C", repoDir, "pull")
				pullCmd.Stdout = os.Stdout
				pullCmd.Stderr = os.Stderr
				if err := pullCmd.Run(); err != nil {
					return fmt.Errorf("failed to pull updates: %v", err)
				}
				
				return nil
			}
		} else {
			// Skip if not overriding
			log.Printf("Skipping existing repository: %s", repo.Name)
			return nil
		}
	}

	// Clone repository
	var cmd *exec.Cmd
	if config.CloneAsMirror {
		// Clone as mirror
		cmd = exec.Command("git", "clone", "--mirror", cloneURL, repoDir)
		log.Printf("Cloning repository as mirror: %s", repo.Name)
	} else {
		// Standard clone
		cmd = exec.Command("git", "clone", cloneURL, repoDir)
		log.Printf("Cloning repository: %s", repo.Name)
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
