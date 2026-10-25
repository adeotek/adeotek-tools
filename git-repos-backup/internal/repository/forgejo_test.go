package repository

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
)

// recordingExecCommand records the curl arguments and returns a fixed response,
// so auth headers and endpoints can be asserted without a network.
func recordingExecCommand(t *testing.T, body string) func(string, ...string) *exec.Cmd {
	t.Helper()
	return func(command string, args ...string) *exec.Cmd {
		t.Helper()
		t.Logf("curl args: %v", args)
		dir := t.TempDir()
		// Construct a fake command that echoes the canned body.
		cmd := exec.Command("sh", "-c", "cat $FAKE_BODY")
		cmd.Env = append(os.Environ(), "FAKE_BODY="+writeTempFile(t, dir, body))
		return cmd
	}
}

func writeTempFile(t *testing.T, dir, content string) string {
	t.Helper()
	path := dir + "/body.json"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

const forgjoSinglePage = `{"data":[{"id":1,"name":"alpha","full_name":"backuptest/alpha",
"clone_url":"https://forgejo.example.com/backuptest/alpha.git","owner":{"login":"backuptest"}}]}`

// TestGetRepositories_ForgejoIsSupported verifies that a "forgejo" provider is
// accepted and queried through the Gitea-compatible API.
func TestGetRepositories_ForgejoIsSupported(t *testing.T) {
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()

	var seen [][]string
	ExecCommand = func(command string, args ...string) *exec.Cmd {
		seen = append(seen, args)
		dir := t.TempDir()
		cmd := exec.Command("sh", "-c", "cat $FAKE_BODY")
		cmd.Env = append(os.Environ(), "FAKE_BODY="+writeTempFile(t, dir, forgjoSinglePage))
		return cmd
	}

	for _, providerType := range []config.ProviderType{config.ProviderGitea, config.ProviderForgejo} {
		t.Run(string(providerType), func(t *testing.T) {
			seen = nil
			cfg := &config.ProviderConfig{
				Type:        providerType,
				ServerURL:   "https://forgejo.example.com",
				AccessToken: "tok123",
			}

			repos, err := GetRepositories(cfg, false)
			if err != nil {
				t.Fatalf("GetRepositories(%s) error = %v", providerType, err)
			}
			if len(repos) != 1 || repos[0].FullName != "backuptest/alpha" {
				t.Fatalf("GetRepositories(%s) = %+v, want one repo backuptest/alpha", providerType, repos)
			}
			if repos[0].Login != "backuptest" {
				t.Errorf("Login = %q, want backuptest", repos[0].Login)
			}

			if len(seen) == 0 {
				t.Fatal("no curl invocation recorded")
			}
			joined := strings.Join(seen[0], " ")
			if !strings.Contains(joined, "/api/v1/repos/search") {
				t.Errorf("request %q does not use the Gitea-compatible search endpoint", joined)
			}
			if !strings.Contains(joined, "Authorization: token tok123") {
				t.Errorf("request %q does not send the Forgejo/Gitea token header", joined)
			}
		})
	}
}

func TestGetRepositories_UnknownProvider(t *testing.T) {
	cfg := &config.ProviderConfig{Type: "gitlab", ServerURL: "https://x", TargetDir: "/tmp"}
	_, err := GetRepositories(cfg, false)
	if err == nil {
		t.Fatal("GetRepositories() error = nil, want an error for an unknown provider")
	}
	if !strings.Contains(err.Error(), "gitlab") {
		t.Errorf("error %q should name the unsupported provider", err)
	}
}

func TestRunCurl_GitHubUsesBearer(t *testing.T) {
	oldExecCommand := ExecCommand
	defer func() { ExecCommand = oldExecCommand }()

	var args []string
	ExecCommand = func(command string, a ...string) *exec.Cmd {
		args = a
		dir := t.TempDir()
		cmd := exec.Command("sh", "-c", "cat $FAKE_BODY")
		cmd.Env = append(os.Environ(), "FAKE_BODY="+writeTempFile(t, dir, "[]"))
		return cmd
	}

	_, err := GetRepositories(&config.ProviderConfig{
		Type:        config.ProviderGitHub,
		AccessToken: "ghtoken",
	}, false)
	if err != nil {
		t.Fatalf("GetRepositories() error = %v", err)
	}

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "Authorization: Bearer ghtoken") {
		t.Errorf("GitHub request %q must use the Bearer scheme", joined)
	}
	if !strings.Contains(joined, "https://api.github.com/user/repos") {
		t.Errorf("GitHub request %q must target the default API", joined)
	}
}
