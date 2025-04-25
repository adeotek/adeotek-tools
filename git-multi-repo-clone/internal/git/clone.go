// Package git handles Git operations
package git

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/adeotek/git-multi-repo-clone/internal/config"
	"github.com/adeotek/git-multi-repo-clone/internal/repository"
)

// ExecCommand is a variable that holds the exec.Command function.
// It can be replaced in tests to mock command execution.
var ExecCommand = exec.Command

// CloneRepository clones a repository to the target directory
func CloneRepository(config *config.Config, repo repository.Repository) error {
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
					return fmt.Errorf("failed to delete existing repository: %w", err)
				}
			} else {
				// For standard repos, update with fetch and pull
				log.Printf("Updating existing repository: %s", repo.Name)
				
				// Fetch with prune
				fetchCmd := ExecCommand("git", "-C", repoDir, "fetch", "--prune")
				fetchCmd.Stdout = os.Stdout
				fetchCmd.Stderr = os.Stderr
				if err := fetchCmd.Run(); err != nil {
					return fmt.Errorf("failed to fetch updates: %w", err)
				}
				
				// Pull updates
				pullCmd := ExecCommand("git", "-C", repoDir, "pull")
				pullCmd.Stdout = os.Stdout
				pullCmd.Stderr = os.Stderr
				if err := pullCmd.Run(); err != nil {
					return fmt.Errorf("failed to pull updates: %w", err)
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
		cmd = ExecCommand("git", "clone", "--mirror", cloneURL, repoDir)
		log.Printf("Cloning repository as mirror: %s", repo.Name)
	} else {
		// Standard clone
		cmd = ExecCommand("git", "clone", cloneURL, repoDir)
		log.Printf("Cloning repository: %s", repo.Name)
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}