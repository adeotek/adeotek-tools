package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

const backupDirectoryName = ".db-backups"

// BackupService handles database backup and restore operations
type BackupService struct {
	isDryRun bool
	verbose  bool
}

// NewBackupService creates a new backup service instance
func NewBackupService(isDryRun, verbose bool) *BackupService {
	return &BackupService{
		isDryRun: isDryRun,
		verbose:  verbose,
	}
}

// CreateBackup creates a backup of the database
func (bs *BackupService) CreateBackup(connectionParams *models.ConnectionParameters) (string, error) {
	backupDir := bs.getBackupDirectory()
	if !bs.isDryRun {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create backup directory: %w", err)
		}
		if bs.verbose {
			log.Printf("Backup directory: %s", backupDir)
		}
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	dbName := bs.getDatabaseName(connectionParams)
	backupFileName := fmt.Sprintf("%s_backup_%s", dbName, timestamp)

	provider := connectionParams.GetDbProvider()
	switch provider {
	case models.PostgreSQL:
		return bs.createPostgresBackup(connectionParams, backupDir, backupFileName)
	case models.SQLite:
		return bs.createSQLiteBackup(connectionParams, backupDir, backupFileName)
	default:
		return "", fmt.Errorf("backup not supported for provider: %s", provider)
	}
}

// RestoreLastBackup restores the most recent backup
func (bs *BackupService) RestoreLastBackup(connectionParams *models.ConnectionParameters) error {
	lastBackupPath, err := bs.GetLastBackupPath(connectionParams)
	if err != nil {
		return err
	}
	if lastBackupPath == "" {
		return fmt.Errorf("no backup found to restore")
	}

	log.Printf("Restoring backup from: %s", lastBackupPath)

	if bs.isDryRun {
		log.Printf("Dry run: would restore backup from %s", lastBackupPath)
		return nil
	}

	provider := connectionParams.GetDbProvider()
	switch provider {
	case models.PostgreSQL:
		return bs.restorePostgresBackup(connectionParams, lastBackupPath)
	case models.SQLite:
		return bs.restoreSQLiteBackup(connectionParams, lastBackupPath)
	default:
		return fmt.Errorf("restore not supported for provider: %s", provider)
	}
}

// GetLastBackupPath returns the path to the most recent backup file
func (bs *BackupService) GetLastBackupPath(connectionParams *models.ConnectionParameters) (string, error) {
	backupDir := bs.getBackupDirectory()
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return "", nil
	}

	provider := connectionParams.GetDbProvider()
	var extension string
	if provider == models.PostgreSQL {
		extension = ".sql"
	} else {
		extension = ".db"
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return "", fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backupFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), extension) {
			backupFiles = append(backupFiles, filepath.Join(backupDir, entry.Name()))
		}
	}

	if len(backupFiles) == 0 {
		return "", nil
	}

	// Sort by modification time (most recent first)
	sort.Slice(backupFiles, func(i, j int) bool {
		infoI, _ := os.Stat(backupFiles[i])
		infoJ, _ := os.Stat(backupFiles[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return backupFiles[0], nil
}

func (bs *BackupService) createPostgresBackup(connectionParams *models.ConnectionParameters, backupDir, backupFileName string) (string, error) {
	backupPath := filepath.Join(backupDir, backupFileName+".sql")
	log.Printf("Creating PostgreSQL backup: %s", backupPath)

	if bs.isDryRun {
		log.Printf("Dry run: would create PostgreSQL backup at %s", backupPath)
		return backupPath, nil
	}

	args := []string{
		fmt.Sprintf("--host=%s", connectionParams.Host),
		fmt.Sprintf("--port=%d", connectionParams.Port),
		fmt.Sprintf("--username=%s", connectionParams.User),
		fmt.Sprintf("--dbname=%s", connectionParams.DatabaseName),
		fmt.Sprintf("--file=%s", backupPath),
		"--format=plain",
		"--no-owner",
		"--no-acl",
		"--clean",
		"--if-exists",
	}

	cmd := exec.Command("pg_dump", args...)
	if connectionParams.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", connectionParams.Password))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump failed: %w\nOutput: %s", err, string(output))
	}

	log.Printf("PostgreSQL backup created successfully: %s", backupPath)
	return backupPath, nil
}

func (bs *BackupService) createSQLiteBackup(connectionParams *models.ConnectionParameters, backupDir, backupFileName string) (string, error) {
	sourcePath := connectionParams.DatabaseName
	backupPath := filepath.Join(backupDir, backupFileName+".db")
	log.Printf("Creating SQLite backup: %s", backupPath)

	if bs.isDryRun {
		log.Printf("Dry run: would create SQLite backup at %s", backupPath)
		return backupPath, nil
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("SQLite database file not found: %s", sourcePath)
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("failed to copy database file: %w", err)
	}

	log.Printf("SQLite backup created successfully: %s", backupPath)
	return backupPath, nil
}

func (bs *BackupService) restorePostgresBackup(connectionParams *models.ConnectionParameters, backupPath string) error {
	log.Printf("Restoring PostgreSQL backup from: %s", backupPath)

	args := []string{
		fmt.Sprintf("--host=%s", connectionParams.Host),
		fmt.Sprintf("--port=%d", connectionParams.Port),
		fmt.Sprintf("--username=%s", connectionParams.User),
		fmt.Sprintf("--dbname=%s", connectionParams.DatabaseName),
		fmt.Sprintf("--file=%s", backupPath),
	}

	cmd := exec.Command("psql", args...)
	if connectionParams.Password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", connectionParams.Password))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("psql restore failed: %w\nOutput: %s", err, string(output))
	}

	log.Printf("PostgreSQL backup restored successfully")
	return nil
}

func (bs *BackupService) restoreSQLiteBackup(connectionParams *models.ConnectionParameters, backupPath string) error {
	targetPath := connectionParams.DatabaseName
	log.Printf("Restoring SQLite backup from %s to %s", backupPath, targetPath)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target database file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy backup file: %w", err)
	}

	log.Printf("SQLite backup restored successfully")
	return nil
}

func (bs *BackupService) getBackupDirectory() string {
	currentDir, _ := os.Getwd()
	return filepath.Join(currentDir, backupDirectoryName)
}

func (bs *BackupService) getDatabaseName(connectionParams *models.ConnectionParameters) string {
	dbName := connectionParams.DatabaseName
	if dbName == "" {
		dbName = "unknown"
	}
	// Remove invalid file name characters
	dbName = strings.ReplaceAll(dbName, "/", "_")
	dbName = strings.ReplaceAll(dbName, "\\", "_")
	dbName = strings.ReplaceAll(dbName, ":", "_")
	dbName = strings.ReplaceAll(dbName, "*", "_")
	dbName = strings.ReplaceAll(dbName, "?", "_")
	dbName = strings.ReplaceAll(dbName, "\"", "_")
	dbName = strings.ReplaceAll(dbName, "<", "_")
	dbName = strings.ReplaceAll(dbName, ">", "_")
	dbName = strings.ReplaceAll(dbName, "|", "_")
	return dbName
}
