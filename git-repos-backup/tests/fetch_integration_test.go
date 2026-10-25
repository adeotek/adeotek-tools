package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/git"
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/repository"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestFetchRepositoryRealGit exercises the backup flow against a real local git
// repository whose default branch is not "master" and that has extra branches
// and tags. The resulting bare backup must be complete and cloneable.
func TestFetchRepositoryRealGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	baseDir := t.TempDir()
	originDir := filepath.Join(baseDir, "origin")
	if err := os.MkdirAll(originDir, 0755); err != nil {
		t.Fatal(err)
	}

	runGit(t, originDir, "init", "--bare", "--quiet", "-b", "main", ".")
	workDir := filepath.Join(baseDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "init", "--quiet", "-b", "main", ".")
	runGit(t, workDir, "remote", "add", "origin", originDir)

	if err := os.WriteFile(filepath.Join(workDir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "add", "README.md")
	runGit(t, workDir, "commit", "--quiet", "-m", "initial")
	runGit(t, workDir, "tag", "v1.0.0")
	runGit(t, workDir, "push", "--quiet", "origin", "main", "--tags")

	runGit(t, workDir, "checkout", "--quiet", "-b", "feature")
	if err := os.WriteFile(filepath.Join(workDir, "feature.txt"), []byte("feature\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "add", "feature.txt")
	runGit(t, workDir, "commit", "--quiet", "-m", "feature")
	runGit(t, workDir, "push", "--quiet", "origin", "feature")
	runGit(t, workDir, "checkout", "--quiet", "main")

	provider := &config.ProviderConfig{
		Type:      config.ProviderGitea,
		TargetDir: filepath.Join(baseDir, "backup"),
	}
	repo := repository.Repository{
		Id:       1,
		Login:    "owner",
		Name:     "project",
		FullName: "owner/project",
		URL:      originDir,
	}

	if err := git.FetchRepository(provider, repo, false); err != nil {
		t.Fatalf("FetchRepository() error = %v", err)
	}

	backupDir := filepath.Join(provider.TargetDir, "owner", "project")

	// All branches must be present.
	branches := runGit(t, backupDir, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	for _, want := range []string{"main", "feature"} {
		if !strings.Contains(branches, want) {
			t.Errorf("backup is missing branch %q, has: %s", want, branches)
		}
	}

	// Tags must be present.
	if tags := runGit(t, backupDir, "tag", "--list"); !strings.Contains(tags, "v1.0.0") {
		t.Errorf("backup is missing tag v1.0.0, has: %q", tags)
	}

	// HEAD must resolve to a commit (otherwise the backup cannot be cloned).
	head := runGit(t, backupDir, "rev-parse", "HEAD")

	// The backup must be cloneable and check out the default branch.
	cloneDir := filepath.Join(baseDir, "clone")
	runGit(t, baseDir, "clone", "--quiet", backupDir, cloneDir)
	if branch := runGit(t, cloneDir, "rev-parse", "--abbrev-ref", "HEAD"); branch != "main" {
		t.Errorf("clone checked out %q, want main", branch)
	}
	if cloneHead := runGit(t, cloneDir, "rev-parse", "HEAD"); cloneHead != head {
		t.Errorf("clone HEAD = %s, want %s", cloneHead, head)
	}
}

// TestFetchRepositoryRepeated verifies that a second run updates the backup.
func TestFetchRepositoryRepeated(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	baseDir := t.TempDir()
	originDir := filepath.Join(baseDir, "origin")
	workDir := filepath.Join(baseDir, "work")
	for _, dir := range []string{originDir, workDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	runGit(t, originDir, "init", "--bare", "--quiet", "-b", "main", ".")
	runGit(t, workDir, "init", "--quiet", "-b", "main", ".")
	runGit(t, workDir, "remote", "add", "origin", originDir)
	if err := os.WriteFile(filepath.Join(workDir, "a.txt"), []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "add", "a.txt")
	runGit(t, workDir, "commit", "--quiet", "-m", "one")
	runGit(t, workDir, "push", "--quiet", "origin", "main")

	provider := &config.ProviderConfig{TargetDir: filepath.Join(baseDir, "backup")}
	repo := repository.Repository{
		Id:       1,
		Login:    "owner",
		Name:     "project",
		FullName: "owner/project",
		URL:      originDir,
	}

	if err := git.FetchRepository(provider, repo, false); err != nil {
		t.Fatalf("first FetchRepository() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(workDir, "b.txt"), []byte("b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, workDir, "add", "b.txt")
	runGit(t, workDir, "commit", "--quiet", "-m", "two")
	runGit(t, workDir, "push", "--quiet", "origin", "main")

	if err := git.FetchRepository(provider, repo, false); err != nil {
		t.Fatalf("second FetchRepository() error = %v", err)
	}

	backupDir := filepath.Join(provider.TargetDir, "owner", "project")
	originHead := runGit(t, originDir, "rev-parse", "main")
	backupHead := runGit(t, backupDir, "rev-parse", "main")
	if backupHead != originHead {
		t.Errorf("backup main = %s, want origin main %s", backupHead, originHead)
	}
}
