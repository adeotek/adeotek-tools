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

	// Expected order: alphabetical by directory (data, tables, views), then alphabetical by filename within each directory
	expected := []string{
		"data/001_seed_data.sql",
		"tables/001_create_users.sql",
		"tables/002_create_posts.sql",
		"views/001_user_posts.sql",
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

func TestSqlScriptsHelpers_ScanForSqlFiles_Comprehensive(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test directories including stored_procedures
	directories := []string{"tables", "views", "stored_procedures", "data", "other"}
	for _, dir := range directories {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create %s directory: %v", dir, err)
		}
	}

	// Create test SQL files with various naming patterns
	testFiles := []string{
		"tables/001_create_users.sql",
		"tables/002_create_posts.sql",
		"views/001_user_posts.sql",
		"views/002_post_stats.sql",
		"stored_procedures/001_get_user.sql",
		"data/001_seed_users.sql",
		"data/002_seed_posts.sql",
		"other/001_misc.sql",
		"other/002_cleanup.sql",
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

	// Expected order: alphabetical by directory, then alphabetical by filename within each directory
	// Directories in alphabetical order: data, other, stored_procedures, tables, views
	expected := []string{
		"data/001_seed_users.sql",
		"data/002_seed_posts.sql",
		"other/001_misc.sql",
		"other/002_cleanup.sql",
		"stored_procedures/001_get_user.sql",
		"tables/001_create_users.sql",
		"tables/002_create_posts.sql",
		"views/001_user_posts.sql",
		"views/002_post_stats.sql",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(result))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", result)
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

func TestSqlScriptsHelpers_ScanForSqlFiles_ErrorCases(t *testing.T) {
	helper := NewSqlScriptsHelpers()

	// Test with empty directory path
	_, err := helper.ScanForSqlFiles("")
	if err == nil {
		t.Error("Expected error for empty directory path")
	}

	// Test with non-existent directory
	_, err = helper.ScanForSqlFiles("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Test with empty directory
	emptyDir := t.TempDir()
	result, err := helper.ScanForSqlFiles(emptyDir)
	if err != nil {
		t.Fatalf("ScanForSqlFiles() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(result))
	}
}
