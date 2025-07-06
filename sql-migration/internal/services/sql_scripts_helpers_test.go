package services

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSqlScriptsHelpers_SplitSqlStatements_ComplexScript(t *testing.T) {
	helper := NewSqlScriptsHelpers()

	// Test with PostgreSQL stored procedure
	complexScript := `CREATE OR REPLACE PROCEDURE public.sp_test()
LANGUAGE plpgsql
AS $$
DECLARE
    counter_companies INT := 0;
    counter_legal_entities INT := 0;
    counter_contacts INT := 0;
BEGIN
  SELECT COUNT(*) INTO counter_companies FROM public.companies;
  SELECT COUNT(*) INTO counter_legal_entities FROM public.legal_entities;
  SELECT COUNT(*) INTO counter_contacts FROM public.contacts;

  RAISE NOTICE 'Total Companies: %, Total Legal Entities: %, Total Contacts: %',
    counter_companies, counter_legal_entities, counter_contacts;
END;
$$;

-- Grant execute permission to necessary roles (adjust as needed)
GRANT EXECUTE ON PROCEDURE public.sp_test() TO PUBLIC;`

	statements := helper.SplitSqlStatements(complexScript)

	// Should result in 2 statements: CREATE PROCEDURE and GRANT
	expectedCount := 2
	if len(statements) != expectedCount {
		t.Errorf("Expected %d statements, got %d", expectedCount, len(statements))
		for i, stmt := range statements {
			t.Logf("Statement %d: %s", i+1, stmt)
		}
	}

	// First statement should be the CREATE PROCEDURE
	if len(statements) > 0 {
		if !strings.Contains(statements[0], "CREATE OR REPLACE PROCEDURE") {
			t.Errorf("First statement should contain CREATE OR REPLACE PROCEDURE")
		}
		if !strings.Contains(statements[0], "$$") {
			t.Errorf("First statement should contain $$ delimiters")
		}
	}

	// Second statement should be the GRANT
	if len(statements) > 1 {
		if !strings.Contains(statements[1], "GRANT EXECUTE") {
			t.Errorf("Second statement should contain GRANT EXECUTE")
		}
	}
}

func TestSqlScriptsHelpers_SplitSqlStatements_SimpleStatements(t *testing.T) {
	helper := NewSqlScriptsHelpers()

	// Test with simple statements
	simpleScript := `CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100));
INSERT INTO users (id, name) VALUES (1, 'John Doe');
INSERT INTO users (id, name) VALUES (2, 'Jane Smith');`

	statements := helper.SplitSqlStatements(simpleScript)

	expectedCount := 3
	if len(statements) != expectedCount {
		t.Errorf("Expected %d statements, got %d", expectedCount, len(statements))
		for i, stmt := range statements {
			t.Logf("Statement %d: %s", i+1, stmt)
		}
	}
}

func TestSqlScriptsHelpers_SplitSqlStatements_WithComments(t *testing.T) {
	helper := NewSqlScriptsHelpers()

	// Test with comments
	scriptWithComments := `-- Create users table
CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100));

-- Insert test data
INSERT INTO users (id, name) VALUES (1, 'John Doe');

-- Another comment
INSERT INTO users (id, name) VALUES (2, 'Jane Smith');`

	statements := helper.SplitSqlStatements(scriptWithComments)

	expectedCount := 3
	if len(statements) != expectedCount {
		t.Errorf("Expected %d statements, got %d", expectedCount, len(statements))
		for i, stmt := range statements {
			t.Logf("Statement %d: %s", i+1, stmt)
		}
	}

	// Statements should not contain standalone comments
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if strings.HasPrefix(trimmed, "--") && !strings.Contains(trimmed, "CREATE") && !strings.Contains(trimmed, "INSERT") {
			t.Errorf("Statement should not be a standalone comment: %s", stmt)
		}
	}
}
