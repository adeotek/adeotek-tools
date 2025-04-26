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
func CloneRepository(config *config.Config, repo repository.Repository, verbose bool) error {

	repoDir, err := GetRepoPath(config.TargetDir, repo.Login, repo.Name, verbose)
	if err != nil {
		log.Fatalf("Failed to create repo directory: %v", err)
	}

	// Prepare clone URL with authentication if needed
	repoUrl, err := GetRepoUrl(config, repo.URL)
	if err != nil {
		log.Fatalf("Failed to create clone URL: %v", err)
	}

	// Init repository if it doesn't exist
	if !RepoExists(repoDir, verbose) {
		err := RunGitInit(repoDir, verbose)
		if err != nil {
			log.Fatalf("Failed to init repository: %v", err)
		}
	}

	// Fetch repository
	return RunGitFetch(config, repoDir, repoUrl, repo.FullName, verbose)
}

func RunGitFetch(config *config.Config, repoDir string, repoUrl string, repoName string, verbose bool) error {
	log.Printf("Fetching repository: %s", repoName)
	cmd := GetGitCommand(config, "-C", repoDir, "fetch", "--force", "--prune", "--tags", repoUrl, "refs/heads/*:refs/heads/*")
	if verbose {
		fmt.Printf("----> %s \n", cmd.String())
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunGitInit(repoDir string, verbose bool) error {
	log.Printf("Initializing repository in path: %s", repoDir)
	cmd := ExecCommand("git", "-C", repoDir, "init", "--bare", "--quiet")
	if verbose {
		fmt.Printf("----> %s \n", cmd.String())
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetGitCommand(config *config.Config, elems ...string) *exec.Cmd {
	cmd := ExecCommand("git")
	if config.SkipSslValidation {
		cmd.Args = append(cmd.Args, "-c", "http.sslVerify=false")
	}
	if len(elems) > 0 {
		cmd.Args = append(cmd.Args, elems...)
	}
	return cmd
}

func GetRepoUrl(config *config.Config, rawUrl string) (string, error) {
	if len(rawUrl) < 10 {
		return "", fmt.Errorf("invalid URL: %s", rawUrl)
	}

	var protocol string
	var cloneURL string

	if rawUrl[0:7] == "http://" {
		protocol = "http://"
		cloneURL = rawUrl[7:]
	} else if rawUrl[0:8] == "https://" {
		protocol = "https://"
		cloneURL = rawUrl[8:]
	} else {
		// SSH or other protocols
		return rawUrl, nil
	}

	if config.UseBasicAuth {
		// Insert username and password for basic authentication
		cloneURL = fmt.Sprintf("%s%s:%s@%s",
			protocol,
			config.Username,
			config.Password,
			cloneURL)
	} else if config.AccessToken != "" {
		// Use token authentication
		cloneURL = fmt.Sprintf("%s%s@%s",
			protocol,
			config.AccessToken,
			cloneURL)
	} else {
		// SSH or other protocols
		cloneURL = rawUrl
	}

	return cloneURL, nil
}

func GetRepoPath(targetDir string, userName string, repoName string, verbose bool) (string, error) {
	userDir := filepath.Join(targetDir, userName)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		if err := os.MkdirAll(userDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create user directory %s: %w", userDir, err)
		}
		if verbose {
			log.Printf("----> Created directory: %s", userDir)
		}
	}

	repoDir := filepath.Join(userDir, repoName)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create user directory %s: %w", repoDir, err)
		}
		if verbose {
			log.Printf("----> Created directory: %s", repoDir)
		}
	}

	return repoDir, nil
}

func RepoExists(repoDir string, verbose bool) bool {
	// Check if file HEAD exists in repodir and is not empty
	headFile := filepath.Join(repoDir, "HEAD")
	if verbose {
		log.Printf("----> Checking if repository exists at %s", repoDir)
	}

	info, err := os.Stat(headFile)
	if err != nil {
		if verbose {
			log.Printf("----> HEAD file not found in %s", repoDir)
		}
		return false
	}

	// Check if the file is not empty
	if info.Size() == 0 {
		if verbose {
			log.Printf("----> HEAD file is empty in %s", repoDir)
		}
		return false
	}

	if verbose {
		log.Printf("----> Repository exists at %s", repoDir)
	}
	return true
}
