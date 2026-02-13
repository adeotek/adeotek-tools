package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

func TestBackupService_CreateBackup_SQLite(t *testing.T) {
	// Create a temporary SQLite database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	err := os.WriteFile(dbPath, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create connection parameters
	connectionParams := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: dbPath,
	}

	// Create backup service
	backupService := NewBackupService(false, false)

	// Create backup
	backupPath, err := backupService.CreateBackup(connectionParams)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file does not exist: %s", backupPath)
	}

	// Verify backup content matches original
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != "test data" {
		t.Errorf("Backup content does not match. Got: %s", string(backupContent))
	}

	// Cleanup
	backupDir := backupService.getBackupDirectory()
	os.RemoveAll(backupDir)
}

func TestBackupService_CreateBackup_DryRun(t *testing.T) {
	// Create a temporary SQLite database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	err := os.WriteFile(dbPath, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create connection parameters
	connectionParams := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: dbPath,
	}

	// Create backup service in dry-run mode
	backupService := NewBackupService(true, false)

	// Create backup (dry-run)
	backupPath, err := backupService.CreateBackup(connectionParams)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify backup file does NOT exist (dry-run)
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Errorf("Backup file should not exist in dry-run mode: %s", backupPath)
	}
}

func TestBackupService_GetLastBackupPath(t *testing.T) {
	// Create a temporary SQLite database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	err := os.WriteFile(dbPath, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create connection parameters
	connectionParams := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: dbPath,
	}

	// Create backup service
	backupService := NewBackupService(false, false)

	// Create first backup
	backup1, err := backupService.CreateBackup(connectionParams)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Wait a bit to ensure different timestamp
	time.Sleep(1100 * time.Millisecond)

	// Create second backup
	backup2, err := backupService.CreateBackup(connectionParams)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Get last backup path
	lastBackupPath, err := backupService.GetLastBackupPath(connectionParams)
	if err != nil {
		t.Fatalf("GetLastBackupPath failed: %v", err)
	}

	// Verify it returns the second (most recent) backup
	if lastBackupPath != backup2 {
		t.Errorf("GetLastBackupPath returned wrong backup.\nExpected: %s\nGot: %s", backup2, lastBackupPath)
	}

	// Verify backup1 was indeed created earlier
	info1, _ := os.Stat(backup1)
	info2, _ := os.Stat(backup2)
	if info2.ModTime().Before(info1.ModTime()) {
		t.Error("Second backup should have later modification time")
	}

	// Cleanup
	backupDir := backupService.getBackupDirectory()
	os.RemoveAll(backupDir)
}

func TestBackupService_GetLastBackupPath_NoBackup(t *testing.T) {
	// Create connection parameters
	connectionParams := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: "nonexistent.db",
	}

	// Create backup service
	backupService := NewBackupService(false, false)

	// Get last backup path (should return empty string)
	lastBackupPath, err := backupService.GetLastBackupPath(connectionParams)
	if err != nil {
		t.Fatalf("GetLastBackupPath failed: %v", err)
	}

	if lastBackupPath != "" {
		t.Errorf("GetLastBackupPath should return empty string when no backup exists. Got: %s", lastBackupPath)
	}
}

func TestBackupService_RestoreLastBackup_SQLite(t *testing.T) {
	// Create a temporary SQLite database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	originalData := "original data"
	err := os.WriteFile(dbPath, []byte(originalData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create connection parameters
	connectionParams := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: dbPath,
	}

	// Create backup service
	backupService := NewBackupService(false, false)

	// Create backup
	backupPath, err := backupService.CreateBackup(connectionParams)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Modify backup file to have different content
	backupData := "backup data"
	err = os.WriteFile(backupPath, []byte(backupData), 0644)
	if err != nil {
		t.Fatalf("Failed to modify backup file: %v", err)
	}

	// Modify original database
	err = os.WriteFile(dbPath, []byte("modified data"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify original database: %v", err)
	}

	// Restore backup
	err = backupService.RestoreLastBackup(connectionParams)
	if err != nil {
		t.Fatalf("RestoreLastBackup failed: %v", err)
	}

	// Verify database content matches backup
	restoredContent, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read restored database: %v", err)
	}

	if string(restoredContent) != backupData {
		t.Errorf("Restored content does not match backup.\nExpected: %s\nGot: %s", backupData, string(restoredContent))
	}

	// Cleanup
	backupDir := backupService.getBackupDirectory()
	os.RemoveAll(backupDir)
}

func TestBackupService_RestoreLastBackup_NoBackup(t *testing.T) {
	// Create connection parameters
	connectionParams := &models.ConnectionParameters{
		Provider:     "sqlite",
		DatabaseName: "nonexistent.db",
	}

	// Create backup service
	backupService := NewBackupService(false, false)

	// Try to restore (should fail with no backup)
	err := backupService.RestoreLastBackup(connectionParams)
	if err == nil {
		t.Error("RestoreLastBackup should fail when no backup exists")
	}
}

func TestBackupService_GetDatabaseName(t *testing.T) {
	backupService := NewBackupService(false, false)

	tests := []struct {
		name     string
		dbName   string
		expected string
	}{
		{"Normal name", "testdb", "testdb"},
		{"With slash", "test/db", "test_db"},
		{"With backslash", "test\\db", "test_db"},
		{"With colon", "test:db", "test_db"},
		{"Empty name", "", "unknown"},
		{"With multiple invalid chars", "test:db/path\\file", "test_db_path_file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionParams := &models.ConnectionParameters{
				DatabaseName: tt.dbName,
			}
			result := backupService.getDatabaseName(connectionParams)
			if result != tt.expected {
				t.Errorf("getDatabaseName(%s) = %s; want %s", tt.dbName, result, tt.expected)
			}
		})
	}
}
