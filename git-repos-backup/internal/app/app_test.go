package app

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	// Just ensure it doesn't panic
	PrintUsage()
}

func TestVersionDisplay(t *testing.T) {
	// Save original arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temp dir for outputs
	tmpDir, err := os.MkdirTemp("", "git-repos-backup-app-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock arguments for version
	os.Args = []string{"git-repos-backup", "-version"}

	// Reset the flag package state to handle the new args
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the app
	Run()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	out, _ := ioutil.ReadAll(r)
	output := string(out)

	// Check if output contains version
	if output == "" {
		t.Error("Expected version output, got empty string")
	}
}

func TestHelpDisplay(t *testing.T) {
	// Save original arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temp dir for outputs
	tmpDir, err := os.MkdirTemp("", "git-repos-backup-app-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock arguments for help
	os.Args = []string{"git-repos-backup", "-help"}

	// Reset the flag package state to handle the new args
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the app
	Run()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	out, _ := ioutil.ReadAll(r)
	output := string(out)

	// Check if output contains help info
	if output == "" {
		t.Error("Expected help output, got empty string")
	}
}

func TestConfigLoad(t *testing.T) {
	// Save original arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temp dir for fake config
	tmpDir, err := os.MkdirTemp("", "git-repos-backup-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal valid config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
providers:
  - type: gitea
    server_url: https://gitea.example.com
    access_token: fake_token
    target_dir: ` + tmpDir + `
`
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Mock arguments to use our config
	os.Args = []string{"git-repos-backup", "-config", configPath}

	// Reset the flag package state to handle the new args
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Run the app with the custom config
	// This will fail as we don't mock the API calls
	// but we just want to test that it attempts to load the config
	Run()

	// No assertions needed - we're just making sure it doesn't panic
	// when loading a valid config
}
