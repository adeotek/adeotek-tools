package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adeotek/adeotek-tools/sql-migration/internal/database"
	"github.com/adeotek/adeotek-tools/sql-migration/internal/models"
	"github.com/adeotek/adeotek-tools/sql-migration/internal/repository"
)

// MigrationService handles the execution of database migrations
type MigrationService struct {
	scriptsHelpers *SqlScriptsHelpers
	backupService  *BackupService
	isDryRun       bool
	verbose        bool
}

// NewMigrationService creates a new migration service instance
func NewMigrationService(isDryRun, verbose bool, backupService *BackupService) *MigrationService {
	return &MigrationService{
		scriptsHelpers: NewSqlScriptsHelpers(),
		backupService:  backupService,
		isDryRun:       isDryRun,
		verbose:        verbose,
	}
}

// Run executes the migration process
func (ms *MigrationService) Run(scriptsPath string, connectionParams *models.ConnectionParameters) error {
	if ms.verbose {
		log.Printf("Starting migration process...")
		log.Printf("Scripts path: %s", scriptsPath)
		log.Printf("Dry run mode: %t", ms.isDryRun)
	}

	// Validate connection parameters
	if valid, errors := connectionParams.IsValid(); !valid {
		return fmt.Errorf("invalid connection parameters: %v", errors)
	}

	// Create database factory and migration history repository
	dbFactory := database.NewFactory(connectionParams)
	migrationRepo := repository.NewMigrationHistoryRepository(dbFactory)

	// Check if history table exists, create if not
	var executedScripts []models.ScriptExecutionHistory
	historyExists, err := migrationRepo.IsHistoryTableCreated()
	if err != nil {
		return fmt.Errorf("failed to check history table: %w", err)
	}

	if !historyExists {
		if !ms.isDryRun {
			if ms.verbose {
				log.Println("History table not found. Creating...")
			}
			err = migrationRepo.CreateHistoryTable()
			if err != nil {
				return fmt.Errorf("failed to create history table: %w", err)
			}
			if ms.verbose {
				log.Println("History table created.")
			}
		}
    executedScripts = []models.ScriptExecutionHistory{}
	} else {
		// Get executed scripts from history table (both dry-run and normal modes)
		executedScripts, err = migrationRepo.GetExecutedScripts()
		if err != nil {
			return fmt.Errorf("failed to get executed scripts: %w", err)
		}
	}

	// Get the full path to the scripts directory
	targetDir, err := filepath.Abs(scriptsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Scan for SQL files
	scriptFiles, err := ms.scriptsHelpers.ScanForSqlFiles(targetDir)
	if err != nil {
		return fmt.Errorf("failed to scan for SQL files: %w", err)
	}

	if ms.verbose {
		log.Printf("Found %d script files in directory %s", len(scriptFiles), targetDir)
	}

	// Check if there are unapplied scripts
	hasUnappliedScripts := false
	for _, scriptName := range scriptFiles {
		scriptFile := filepath.Join(targetDir, scriptName)
		hash, err := ms.scriptsHelpers.CalculateHash(scriptFile)
		if err != nil {
			continue
		}

		var executedScript *models.ScriptExecutionHistory
		for _, script := range executedScripts {
			if script.ScriptFile == scriptName {
				executedScript = &script
				break
			}
		}

		if executedScript == nil || executedScript.ScriptHash != hash {
			hasUnappliedScripts = true
			break
		}
	}

	// Create backup if there are unapplied scripts and backup service is provided
	if hasUnappliedScripts && ms.backupService != nil {
		log.Println("Creating database backup before applying migrations...")
		backupPath, err := ms.backupService.CreateBackup(connectionParams)
		if err != nil {
			return fmt.Errorf("failed to create database backup. Migrations will not be applied: %w", err)
		}
		log.Printf("Backup created successfully: %s", backupPath)
	}

	// Process each script
	successCount := 0
	errorsCount := 0
	skipCount := 0

	for _, scriptName := range scriptFiles {
		scriptFile := filepath.Join(targetDir, scriptName)

		// Calculate hash
		hash, err := ms.scriptsHelpers.CalculateHash(scriptFile)
		if err != nil {
			log.Printf("Error calculating hash for script %s: %v", scriptName, err)
			errorsCount++
			continue
		}

		// Check if script has been executed
		var executedScript *models.ScriptExecutionHistory
		for _, script := range executedScripts {
			if script.ScriptFile == scriptName {
				executedScript = &script
				break
			}
		}

		if executedScript != nil && executedScript.ScriptHash == hash {
			skipCount++
			if ms.isDryRun {
				log.Printf("Dry run: would skip script %s [%s] (already executed)", scriptName, hash)
			} else {
				log.Printf("Skipping script %s [%s]", scriptName, hash)
			}
			continue
		}

		// Execute script
		if ms.isDryRun {
			log.Printf("Dry run: would execute script %s [%s]", scriptName, hash)
			successCount++
			continue
		}

		// Read script content
		scriptContent, err := os.ReadFile(scriptFile)
		if err != nil {
			log.Printf("Error reading script %s: %v", scriptName, err)
			errorsCount++
			continue
		}

		// Execute the script
		err = ms.scriptsHelpers.ExecuteScript(string(scriptContent), connectionParams)
		if err != nil {
			log.Printf("Error executing script %s [%s]: %v", scriptName, hash, err)
			errorsCount++
			continue
		}

    log.Printf("Script %s executed successfully [%s]", scriptName, hash)

		// Update execution history
		executionHistory := &models.ScriptExecutionHistory{
			ScriptFile: scriptName,
			ScriptHash: hash,
			ExecutedAt: time.Now().UTC(),
		}

		err = migrationRepo.UpsertExecutedScript(executionHistory)
		if err != nil {
			log.Printf("Error updating execution history for script %s: %v", scriptName, err)
			// Don't increment error count as the script executed successfully
		}

		successCount++
	}

	// Print summary
	log.Printf("Total scripts: %d | Success: %d | Skipped: %d | Errors: %d",
		len(scriptFiles), successCount, skipCount, errorsCount)

	if errorsCount > 0 {
		return fmt.Errorf("migration completed with %d errors", errorsCount)
	}

	return nil
}
