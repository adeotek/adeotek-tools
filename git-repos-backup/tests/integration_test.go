package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/git"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
	"github.com/adeotek/adeotek-tools/git-repos-backup/pkg/filter"
)

// TestIntegration runs basic integration tests
// These tests are only run when RUN_INTEGRATION_TESTS=1 is set
func TestIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "git-repos-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test repository filtering
	testFilterIntegration(t, tmpDir)
}

func testFilterIntegration(t *testing.T, baseDir string) {
	// Create test repositories
	testRepos := []repository.Repository{
		{
			Id:       1,
			Login:    "owner1",
			Name:     "repo1",
			FullName: "owner1/repo1",
			URL:      "https://example.com/owner1/repo1.git",
		},
		{
			Id:       2,
			Login:    "owner1",
			Name:     "repo2",
			FullName: "owner1/repo2",
			URL:      "https://example.com/owner1/repo2.git",
		},
		{
			Id:       3,
			Login:    "owner2",
			Name:     "repo3",
			FullName: "owner2/repo3",
			URL:      "https://example.com/owner2/repo3.git",
		},
	}

	// Test include filter
	providerWithInclude := &config.ProviderConfig{
		Type:      config.ProviderGitea,
		Include:   []string{"owner1/repo1"},
		TargetDir: filepath.Join(baseDir, "include-test"),
	}

	filtered := filter.FilterRepositories(testRepos, providerWithInclude, false)
	if len(filtered) != 1 || filtered[0].FullName != "owner1/repo1" {
		t.Errorf("Include filter failed: expected 1 repo (owner1/repo1), got %d", len(filtered))
	}

	// Test exclude filter
	providerWithExclude := &config.ProviderConfig{
		Type:      config.ProviderGitea,
		Exclude:   []string{"owner1/repo1"},
		TargetDir: filepath.Join(baseDir, "exclude-test"),
	}

	filtered = filter.FilterRepositories(testRepos, providerWithExclude, false)
	if len(filtered) != 2 {
		t.Errorf("Exclude filter failed: expected 2 repos, got %d", len(filtered))
	}
	for _, repo := range filtered {
		if repo.FullName == "owner1/repo1" {
			t.Errorf("Exclude filter failed: found excluded repo owner1/repo1")
		}
	}
}

// TestBasicWorkflow tests the basic workflow without making real API calls
func TestBasicWorkflow(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-repos-backup-workflow")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake repo to verify directory creation
	repoLogin := "testowner"
	repoName := "testrepo"
	repo := repository.Repository{
		Id:       1,
		Login:    repoLogin,
		Name:     repoName,
		FullName: repoLogin + "/" + repoName,
		URL:      "https://example.com/" + repoLogin + "/" + repoName + ".git",
	}

	// Create a provider config
	provider := &config.ProviderConfig{
		Type:      config.ProviderGitea,
		ServerURL: "https://gitea.example.com",
		TargetDir: tmpDir,
	}

	// Test GetRepoPath (creates directory structure)
	repoPath, err := git.GetRepoPath(provider.TargetDir, repo.Login, repo.Name, false)
	if err != nil {
		t.Fatalf("GetRepoPath failed: %v", err)
	}

	// Verify directories were created
	expectedPath := filepath.Join(tmpDir, repoLogin, repoName)
	if repoPath != expectedPath {
		t.Errorf("Expected repo path %s, got %s", expectedPath, repoPath)
	}

	// Verify owner directory exists
	ownerPath := filepath.Join(tmpDir, repoLogin)
	if _, err := os.Stat(ownerPath); os.IsNotExist(err) {
		t.Errorf("Owner directory not created: %s", ownerPath)
	}

	// Verify repo directory exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Errorf("Repo directory not created: %s", repoPath)
	}

	// Test RepoExists (initially false)
	if git.RepoExists(repoPath, false) {
		t.Errorf("RepoExists should return false for new directory")
	}

	// Create a fake HEAD file to simulate a Git repository
	headPath := filepath.Join(repoPath, "HEAD")
	if err := ioutil.WriteFile(headPath, []byte("ref: refs/heads/main"), 0644); err != nil {
		t.Fatalf("Failed to create test HEAD file: %v", err)
	}

	// Test RepoExists (now true)
	if !git.RepoExists(repoPath, false) {
		t.Errorf("RepoExists should return true after creating HEAD file")
	}
}
