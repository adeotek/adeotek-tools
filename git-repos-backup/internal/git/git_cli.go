// Package git handles Git operations
package git

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
)

// ExecCommand is a variable that holds the exec.Command function.
// It can be replaced in tests to mock command execution.
var ExecCommand = exec.Command

// FetchRepository fetches all branches and tags of a repository into a bare
// repository located under the provider's target directory.
func FetchRepository(provider *config.ProviderConfig, repo repository.Repository, verbose bool) error {
	repoDir, err := GetRepoPath(provider.TargetDir, repo.Login, repo.Name, verbose)
	if err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Prepare clone URL with authentication if needed
	repoUrl, err := GetRepoUrl(provider, repo.URL)
	if err != nil {
		return fmt.Errorf("failed to create clone URL: %w", err)
	}

	// Init repository if it doesn't exist
	if !RepoExists(repoDir, verbose) {
		if err := RunGitInit(repoDir, verbose); err != nil {
			return fmt.Errorf("failed to init repository: %w", err)
		}
	}

	// Fetch repository
	if err := RunGitFetch(provider, repoDir, repoUrl, repo.FullName, verbose); err != nil {
		return err
	}

	// Make sure HEAD points to an existing branch so that the backup can be cloned
	if err := EnsureHead(provider, repoDir, repoUrl, verbose); err != nil {
		return fmt.Errorf("failed to update HEAD of %s: %w", repo.FullName, err)
	}

	return nil
}

// RunGitFetch fetches all branches and tags from repoUrl into the bare repository at repoDir.
func RunGitFetch(provider *config.ProviderConfig, repoDir string, repoUrl string, repoName string, verbose bool) error {
	log.Printf("Fetching repository: %s", repoName)
	cmd := GetGitCommand(provider, "-C", repoDir, "fetch", "--force", "--prune",
		repoUrl, "+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*")
	if verbose {
		fmt.Printf("----> git -C %s fetch --force --prune %s +refs/heads/*:refs/heads/* +refs/tags/*:refs/tags/*\n",
			repoDir, RedactUrl(repoUrl))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGitInit initializes a bare repository at repoDir.
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

// EnsureHead points the repository HEAD at the remote's default branch when the
// local HEAD does not resolve to a commit. A bare repository created with
// "git init --bare" points at refs/heads/master by default, which leaves the
// backup unusable ("remote HEAD refers to nonexistent ref") when the remote's
// default branch is named differently.
func EnsureHead(provider *config.ProviderConfig, repoDir string, repoUrl string, verbose bool) error {
	check := GetGitCommand(provider, "-C", repoDir, "rev-parse", "--verify", "--quiet", "HEAD")
	check.Stdout = nil
	check.Stderr = nil
	if err := check.Run(); err == nil {
		// HEAD already resolves to a commit, nothing to do
		return nil
	}

	ref, err := remoteHeadRef(provider, repoUrl, verbose)
	if err != nil {
		return err
	}
	if ref == "" {
		if verbose {
			log.Printf("----> Could not determine remote HEAD for %s, keeping local HEAD", repoDir)
		}
		return nil
	}

	if verbose {
		fmt.Printf("----> git -C %s symbolic-ref HEAD %s\n", repoDir, ref)
	}
	cmd := GetGitCommand(provider, "-C", repoDir, "symbolic-ref", "HEAD", ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// remoteHeadRef returns the branch ref the remote's HEAD points to, or an empty
// string when it cannot be determined.
func remoteHeadRef(provider *config.ProviderConfig, repoUrl string, verbose bool) (string, error) {
	cmd := GetGitCommand(provider, "ls-remote", "--symref", repoUrl, "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to resolve remote HEAD: %w", err)
	}

	// The first line of the output has the form:
	//   ref: refs/heads/main	HEAD
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "ref:" && fields[2] == "HEAD" {
			return fields[1], nil
		}
	}
	return "", nil
}

// GetGitCommand builds a git command, optionally disabling SSL verification.
func GetGitCommand(provider *config.ProviderConfig, elems ...string) *exec.Cmd {
	cmd := ExecCommand("git")
	if provider.SkipSslValidation {
		cmd.Args = append(cmd.Args, "-c", "http.sslVerify=false")
	}
	if len(elems) > 0 {
		cmd.Args = append(cmd.Args, elems...)
	}
	return cmd
}

// GetRepoUrl returns the URL to fetch the repository from, embedding the
// configured credentials when HTTP(S) is used.
func GetRepoUrl(provider *config.ProviderConfig, rawUrl string) (string, error) {
	// SCP-style syntax (git@host:owner/repo.git) is not a valid URL and cannot
	// carry an embedded password, so leave it untouched.
	if !strings.Contains(rawUrl, "://") {
		if isScpLikeUrl(rawUrl) {
			return rawUrl, nil
		}
		// Local paths are valid git remotes and need no credentials.
		if filepath.IsAbs(rawUrl) || strings.HasPrefix(rawUrl, "./") || strings.HasPrefix(rawUrl, "../") {
			return rawUrl, nil
		}
		return "", fmt.Errorf("invalid URL: %s", rawUrl)
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", rawUrl, err)
	}

	// Leave SSH and other non-HTTP(S) protocols untouched
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return rawUrl, nil
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("invalid URL: %s", rawUrl)
	}

	if provider.UseBasicAuth && provider.Username != "" {
		parsed.User = url.UserPassword(provider.Username, provider.Password)
	} else if provider.AccessToken != "" {
		parsed.User = url.User(provider.AccessToken)
	}

	return parsed.String(), nil
}

// isScpLikeUrl reports whether rawUrl uses the scp-like "user@host:path" form.
func isScpLikeUrl(rawUrl string) bool {
	at := strings.Index(rawUrl, "@")
	if at <= 0 {
		return false
	}
	colon := strings.Index(rawUrl[at+1:], ":")
	return colon > 0 && !strings.HasPrefix(rawUrl[at+1:], "[")
}

// RedactUrl removes credentials from a URL so it can be safely logged.
func RedactUrl(rawUrl string) string {
	parsed, err := url.Parse(rawUrl)
	if err != nil || parsed.User == nil {
		return rawUrl
	}
	username := parsed.User.Username()
	if _, hasPassword := parsed.User.Password(); hasPassword {
		// Basic auth: the username is not secret, but the password is.
		parsed.User = nil
		return strings.Replace(parsed.String(), parsed.Scheme+"://", parsed.Scheme+"://"+username+":***@", 1)
	}
	// Token auth: the token is carried in the username field.
	parsed.User = nil
	return strings.Replace(parsed.String(), parsed.Scheme+"://", parsed.Scheme+"://***@", 1)
}

// GetRepoPath creates and returns the directory used to store a repository backup.
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
			return "", fmt.Errorf("failed to create repository directory %s: %w", repoDir, err)
		}
		if verbose {
			log.Printf("----> Created directory: %s", repoDir)
		}
	}

	return repoDir, nil
}

// RepoExists reports whether repoDir already contains an initialized repository.
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
