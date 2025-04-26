// Package app provides the core functionality for the git-multi-repo-clone application
package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/adeotek/git-multi-repo-clone/internal/config"
	"github.com/adeotek/git-multi-repo-clone/internal/git"
	"github.com/adeotek/git-multi-repo-clone/internal/repository"
	"github.com/adeotek/git-multi-repo-clone/pkg/filter"
)

// Version information
const (
	Version = "0.1.0"
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
	fmt.Printf("git-multi-repo-clone version %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

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

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(cfg.TargetDir, 0755); err != nil {
		log.Fatalf("Failed to create target directory: %v", err)
	}

	// Get list of repositories
	repos, err := repository.GetRepositories(cfg, *verbose)
	if err != nil {
		log.Fatalf("Failed to get repositories: %v", err)
	} else if *verbose {
		fmt.Printf("----> %d repos found\n", len(repos))
	}

	// Filter repositories based on include/exclude lists
	repos = filter.FilterRepositories(repos, cfg, *verbose)

	if *verbose {
		fmt.Printf("----> %d repos filtered\n", len(repos))
	}

	// Clone each repository
	for _, repo := range repos {
		if *verbose {
			fmt.Printf("----> Processing repo: %s\n", repo.FullName)
		}

		if err := git.CloneRepository(cfg, repo, *verbose); err != nil {
			log.Printf("Failed to clone repository %s: %v", repo.Name, err)
		}
	}
}

// PrintUsage displays the command-line usage information
func PrintUsage() {
	fmt.Println("Git Multi-Repo Clone - Clone multiple Git repositories from a Gitea server")
	fmt.Println("\nUsage:")
	fmt.Println("  git-multi-repo-clone [flags]")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	fmt.Println("\nConfiguration file (YAML):")
	fmt.Println("  gitea_url: URL of the Gitea server")
	fmt.Println("  api_token: API token for authentication (if use_basic_auth is false)")
	fmt.Println("  username: Username for basic authentication (if use_basic_auth is true)")
	fmt.Println("  password: Password for basic authentication (if use_basic_auth is true)")
	fmt.Println("  use_basic_auth: Whether to use basic authentication (default: false)")
	fmt.Println("  target_dir: Directory to clone repositories into")
	fmt.Println("  override_exising_local_repos: Whether to update existing repositories (default: false)")
	fmt.Println("  clone_as_mirror: Whether to clone as mirror (default: false)")
	fmt.Println("  include: List of repository names to include (optional)")
	fmt.Println("  exclude: List of repository names to exclude (optional, ignored if include is specified)")
}
