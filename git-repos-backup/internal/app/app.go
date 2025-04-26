// Package app provides the core functionality for the git-repos-backup application
package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

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
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
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

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

// PrintUsage displays the command-line usage information
func PrintUsage() {
	fmt.Println("Git Repos Backup - Backup multiple Git repositories from Gitea and GitHub")
	fmt.Println("\nUsage:")
	fmt.Println("  git-repos-backup [flags]")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
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
