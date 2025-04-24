package main

import (
	"os"
	"testing"
)

// TestMain is used to set up any test prerequisites that apply to all tests
func TestMain(m *testing.M) {
	// Replace the exec.Command with our mock for testing
	// In each test that needs it, we'll set a specific mock implementation
	
	// Run the tests
	exitCode := m.Run()
	
	// Exit with the same code
	os.Exit(exitCode)
}