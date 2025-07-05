package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSqlScriptsHelpers_ScanForSqlFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directories
	tablesDir := filepath.Join(tempDir, "tables")
	viewsDir := filepath.Join(tempDir, "views")
	dataDir := filepath.Join(tempDir, "data")

	err := os.MkdirAll(tablesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create tables directory: %v", err)
	}

	err = os.MkdirAll(viewsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create views directory: %v", err)
	}

	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	// Create test SQL files
	testFiles := []string{
		"tables/002_create_posts.sql",
		"tables/001_create_users.sql",
		"views/001_user_posts.sql",
		"data/001_seed_data.sql",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.WriteFile(fullPath, []byte("SELECT 1;"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test the scanner
	helper := NewSqlScriptsHelpers()
	result, err := helper.ScanForSqlFiles(tempDir)
	if err != nil {
		t.Fatalf("ScanForSqlFiles() error = %v", err)
	}

	// Expected order: tables (sorted), views (sorted), data (sorted)
	expected := []string{
		"tables/001_create_users.sql",
		"tables/002_create_posts.sql",
		"views/001_user_posts.sql",
		"data/001_seed_data.sql",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(result))
	}

	for i, expectedFile := range expected {
		if i >= len(result) {
			t.Errorf("Missing expected file: %s", expectedFile)
			continue
		}
		if result[i] != expectedFile {
			t.Errorf("Expected file %s at index %d, got %s", expectedFile, i, result[i])
		}
	}
}

func TestSqlScriptsHelpers_CalculateHash(t *testing.T) {
	// Create a temporary file for testing
	tempFile := filepath.Join(t.TempDir(), "test.sql")
	content := "SELECT * FROM users;"

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	helper := NewSqlScriptsHelpers()
	hash, err := helper.CalculateHash(tempFile)
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Hash should be consistent
	hash2, err := helper.CalculateHash(tempFile)
	if err != nil {
		t.Fatalf("CalculateHash() second call error = %v", err)
	}

	if hash != hash2 {
		t.Errorf("Hash should be consistent: %s != %s", hash, hash2)
	}
}
