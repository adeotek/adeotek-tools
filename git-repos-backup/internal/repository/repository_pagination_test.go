package repository

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
)

// pagedExecCommand runs the test binary as a fake curl that serves paginated
// results, so pagination and error handling can be verified without a network.
func pagedExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestPagedHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_PAGED_HELPER=1")
	return cmd
}

func TestPagedHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_PAGED_HELPER") != "1" {
		return
	}

	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) < 1 || args[0] != "curl" {
		os.Exit(1)
	}

	full := strings.Join(args, " ")

	// Simulate an API error response: curl --fail exits non-zero on HTTP >= 400.
	if strings.Contains(full, "fail-page") {
		fmt.Fprintln(os.Stderr, "curl: (22) The requested URL returned error: 500")
		os.Exit(22)
	}

	page := 1
	for _, arg := range args {
		if strings.Contains(arg, "page=2") {
			page = 2
		}
	}

	count := giteaPageSize // full first page
	if page == 2 {
		count = 3 // partial second page
	}

	repos := make([]string, 0, count)
	for i := 0; i < count; i++ {
		id := (page-1)*giteaPageSize + i + 1
		repos = append(repos, fmt.Sprintf(`{
			"id": %d,
			"name": "repo%d",
			"full_name": "owner/repo%d",
			"clone_url": "https://gitea.example.com/owner/repo%d.git",
			"owner": {"login": "owner"}
		}`, id, id, id, id))
	}
	fmt.Printf(`{"data": [%s]}`, strings.Join(repos, ","))
	os.Exit(0)
}

func TestGetGiteaRepositories_Pagination(t *testing.T) {
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()
	ExecCommand = pagedExecCommand

	cfg := &config.ProviderConfig{
		Type:        config.ProviderGitea,
		ServerURL:   "https://gitea.example.com",
		AccessToken: "faketoken",
	}

	repos, err := GetRepositories(cfg, false)
	if err != nil {
		t.Fatalf("GetRepositories() error = %v", err)
	}

	want := giteaPageSize + 3
	if len(repos) != want {
		t.Fatalf("GetRepositories() returned %d repos, want %d (pagination not followed?)", len(repos), want)
	}

	// Ids must be unique and ordered across pages.
	for i, repo := range repos {
		if repo.Id != i+1 {
			t.Errorf("repos[%d].Id = %d, want %d", i, repo.Id, i+1)
		}
	}
}

func TestGetRepositories_APIErrorIsReturned(t *testing.T) {
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()
	ExecCommand = pagedExecCommand

	cfg := &config.ProviderConfig{
		Type:      config.ProviderGitea,
		ServerURL: "https://gitea.example.com/fail-page",
	}

	if _, err := GetRepositories(cfg, false); err == nil {
		t.Fatal("GetRepositories() error = nil, want an error for a failed API request")
	}

	githubCfg := &config.ProviderConfig{
		Type:        config.ProviderGitHub,
		ServerURL:   "https://github.example.com/fail-page",
		AccessToken: "token",
	}
	if _, err := GetRepositories(githubCfg, false); err == nil {
		t.Fatal("GetRepositories() error = nil, want an error for a failed GitHub API request")
	}
}

func TestGetGiteaRepositories_RequiresServerURL(t *testing.T) {
	cfg := &config.ProviderConfig{Type: config.ProviderGitea}
	if _, err := GetRepositories(cfg, false); err == nil {
		t.Fatal("GetRepositories() error = nil, want an error when server_url is missing")
	}
}

func TestRedactedCurlCommand(t *testing.T) {
	args := []string{"--silent", "-H", "Authorization: token secret", "https://example.com"}
	got := redactedCurlCommand(args, "Authorization: token secret")
	if strings.Contains(got, "secret") {
		t.Errorf("redactedCurlCommand() = %q, must not contain the token", got)
	}

	basicArgs := []string{"-u", "user:password", "https://example.com"}
	got = redactedCurlCommand(basicArgs, "")
	if strings.Contains(got, "password") {
		t.Errorf("redactedCurlCommand() = %q, must not contain the password", got)
	}
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("redactedCurlCommand() = %q, drops the URL", got)
	}
}
