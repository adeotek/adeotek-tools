// Package app provides the core functionality for the git-repos-backup application
package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/git"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
	"github.com/adeotek/adeotek-tools/git-repos-backup/pkg/filter"
)

// Version information
const (
	Version = "0.1.1"
)

// Run executes the main application logic
func Run() {
	// Define command-line flags
	configPath := flag.String("config", "", "Path to configuration file (default: config.yaml)")
	providerType := flag.String("provider", "", "Provider type (gitea or github)")
	serverURL := flag.String("server-url", "", "URL of the Git server (required for Gitea, optional for GitHub)")
	accessToken := flag.String("token", "", "API token for authentication")
	username := flag.String("username", "", "Username for basic authentication")
	password := flag.String("password", "", "Password for basic authentication")
	useBasicAuth := flag.Bool("use-basic-auth", false, "Whether to use basic authentication")
	skipSSLValidation := flag.Bool("skip-ssl", false, "Whether to skip SSL validation")
	includeRepos := flag.String("include", "", "Comma-separated list of repository full names to include")
	excludeRepos := flag.String("exclude", "", "Comma-separated list of repository full names to exclude")
	targetDir := flag.String("target-dir", "", "Directory to clone repositories into")
	showVersion := flag.Bool("version", false, "Show version information and exit")
	verbose := flag.Bool("verbose", false, "Show all messages")
	showHelp := flag.Bool("help", false, "Show help message and exit")

	// Parse command-line flags
	flag.Parse()

	// Show version
	fmt.Printf("git-repos-backup version %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

	if *showVersion {
		return
	}

	// Show help if requested
	if *showHelp {
		PrintUsage()
		return
	}

	var cfg *config.Config
	var err error

	// Check if using config file or command-line arguments
	if *configPath != "" {
		// Load configuration from file
		if *verbose {
			fmt.Printf("----> Loading configuration from file: %s\n", *configPath)
		}
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else if *providerType != "" {
		// Check if required parameters are provided
		if *targetDir == "" {
			log.Fatalf("Target directory is required when not using a config file")
		}

		// Parse include/exclude repositories
		var include, exclude []string
		if *includeRepos != "" {
			include = splitCommaSeparatedList(*includeRepos)
		}
		if *excludeRepos != "" && *includeRepos == "" {
			exclude = splitCommaSeparatedList(*excludeRepos)
		}

		// Create configuration from arguments
		if *verbose {
			fmt.Printf("----> Creating configuration from command-line arguments\n")
		}
		cfg = config.CreateFromArgs(
			*providerType,
			*serverURL,
			*accessToken,
			*username,
			*password,
			*useBasicAuth,
			*skipSSLValidation,
			include,
			exclude,
			*targetDir,
		)
	} else {
		// Default to config.yaml in current directory if exists
		defaultConfig := "config.yaml"
		if _, err := os.Stat(defaultConfig); err == nil {
			if *verbose {
				fmt.Printf("----> Loading configuration from default file: %s\n", defaultConfig)
			}
			cfg, err = config.Load(defaultConfig)
			if err != nil {
				log.Fatalf("Failed to load default config: %v", err)
			}
		} else {
			log.Fatalf("No configuration provided. Either specify a config file with -config or provide required command-line arguments")
		}
	}

	// Process each provider
	for i, provider := range cfg.Providers {
		providerName := string(provider.Type)
		if *verbose {
			fmt.Printf("----> Processing provider %d: %s\n", i+1, providerName)
		}

		// Create target directory if it doesn't exist
		if err := os.MkdirAll(provider.TargetDir, 0755); err != nil {
			log.Fatalf("Failed to create target directory: %v", err)
		}

		// Get list of repositories
		repos, err := repository.GetRepositories(&provider, *verbose)
		if err != nil {
			log.Printf("Failed to get repositories from %s: %v", providerName, err)
			continue
		} else if *verbose {
			fmt.Printf("----> %d repos found from %s\n", len(repos), providerName)
		}

		// Filter repositories based on include/exclude lists
		repos = filter.FilterRepositories(repos, &provider, *verbose)

		if *verbose {
			fmt.Printf("----> %d repos filtered from %s\n", len(repos), providerName)
		}

		// Clone each repository
		for _, repo := range repos {
			if *verbose {
				fmt.Printf("----> Processing repo: %s\n", repo.FullName)
			}

			if err := git.FetchRepository(&provider, repo, *verbose); err != nil {
				log.Printf("Failed to fetch repository %s: %v", repo.Name, err)
			}
		}
	}
}

// splitCommaSeparatedList splits a comma-separated string into a slice of strings
func splitCommaSeparatedList(list string) []string {
	if list == "" {
		return nil
	}
	parts := strings.Split(list, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// PrintUsage displays the command-line usage information
func PrintUsage() {
	fmt.Println("Git Repos Backup - Backup multiple Git repositories from Gitea and GitHub")
	fmt.Println("\nUsage:")
	fmt.Println("  git-repos-backup [flags]")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	fmt.Println("\nConfiguration Examples:")
	fmt.Println("\n1. Using config file:")
	fmt.Println("   git-repos-backup -config /path/to/config.yaml [-verbose]")
	fmt.Println("\n2. Using command-line arguments:")
	fmt.Println("   git-repos-backup -provider github -token your_github_token -target-dir /path/to/backups [-verbose]")
	fmt.Println("\nConfiguration file (YAML):")
	fmt.Println("  providers:")
	fmt.Println("    - type: gitea|github")
	fmt.Println("      server_url: URL of the Git server (for GitHub Enterprise)")
	fmt.Println("      access_token: API token for authentication (if use_basic_auth is false)")
	fmt.Println("      username: Username for basic authentication (if use_basic_auth is true)")
	fmt.Println("      password: Password for basic authentication (if use_basic_auth is true)")
	fmt.Println("      use_basic_auth: Whether to use basic authentication (default: false)")
	fmt.Println("      skip_ssl_validation: Whether to skip SSL validation (default: false)")
	fmt.Println("      include: List of repository full names to include (optional)")
	fmt.Println("      exclude: List of repository full names to exclude (optional, ignored if include is specified)")
	fmt.Println("      target_dir: Directory to clone repositories into")
}
