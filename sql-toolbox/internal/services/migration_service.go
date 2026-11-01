package services

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/database"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/repository"
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

	// Prevent two migration runs from applying the same scripts concurrently.
	unlock, lockErr := acquireMigrationLock(dbFactory, connectionParams, scriptsPath)
	if lockErr != nil {
		return lockErr
	}
	defer unlock()

	// In dry-run mode against a SQLite database that does not exist yet, skip history
	// lookups entirely: opening the connection would create the database file, which a
	// dry run must never do.
	skipHistoryLookup := ms.isDryRun && connectionParams.GetDbProvider() == models.SQLite &&
		!fileExists(connectionParams.DatabaseName)

	// Check if history table exists, create if not
	var executedScripts []models.ScriptExecutionHistory
	if skipHistoryLookup {
		if ms.verbose {
			log.Println("Dry run: database does not exist yet, treating all scripts as unapplied.")
		}
		executedScripts = []models.ScriptExecutionHistory{}
	} else {
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
		if connectionParams.GetDbProvider() == models.SQLite && !fileExists(connectionParams.DatabaseName) {
			log.Println("Skipping pre-migration backup: SQLite database does not exist yet")
		} else {
			log.Println("Creating database backup before applying migrations...")
			backupPath, err := ms.backupService.CreateBackup(connectionParams)
			if err != nil {
				return fmt.Errorf("failed to create database backup. Migrations will not be applied: %w", err)
			}
			log.Printf("Backup created successfully: %s", backupPath)
		}
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

		// Execute the script and record its history in one transaction, so a
		// failure cannot leave partially applied changes behind.
		db, dbErr := dbFactory.CreateConnection()
		if dbErr != nil {
			log.Printf("Error connecting to database for script %s: %v", scriptName, dbErr)
			errorsCount++
			continue
		}
		tx, txErr := db.Begin()
		if txErr != nil {
			db.Close()
			log.Printf("Error starting transaction for script %s: %v", scriptName, txErr)
			errorsCount++
			continue
		}

		executionHistory := &models.ScriptExecutionHistory{
			ScriptFile: scriptName,
			ScriptHash: hash,
			ExecutedAt: time.Now().UTC(),
		}
		scriptErr := ms.scriptsHelpers.ExecuteScriptTx(tx, string(scriptContent))
		if scriptErr == nil {
			scriptErr = migrationRepo.UpsertExecutedScriptTx(tx, executionHistory)
		}
		if scriptErr != nil {
			_ = tx.Rollback()
			db.Close()
			log.Printf("Error executing script %s [%s]: %v", scriptName, hash, scriptErr)
			errorsCount++
			continue
		}
		if commitErr := tx.Commit(); commitErr != nil {
			db.Close()
			log.Printf("Error committing script %s [%s]: %v", scriptName, hash, commitErr)
			errorsCount++
			continue
		}
		db.Close()

		log.Printf("Script %s executed successfully [%s]", scriptName, hash)
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

// acquireMigrationLock takes a PostgreSQL advisory lock or a SQLite lock file
// for the duration of the migration.
func acquireMigrationLock(dbFactory *database.Factory, connectionParams *models.ConnectionParameters, scriptsPath string) (func(), error) {
	noop := func() {}

	absPath, err := filepath.Abs(scriptsPath)
	if err != nil {
		absPath = scriptsPath
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(connectionParams.DatabaseName + "|" + absPath))
	lockKey := int64(hasher.Sum64())

	switch connectionParams.GetDbProvider() {
	case models.PostgreSQL:
		db, err := dbFactory.CreateConnection()
		if err != nil {
			return noop, fmt.Errorf("failed to connect for migration lock: %w", err)
		}
		var acquired bool
		if err := db.QueryRow("SELECT pg_try_advisory_lock($1)", lockKey).Scan(&acquired); err != nil {
			db.Close()
			return noop, fmt.Errorf("failed to acquire migration lock: %w", err)
		}
		if !acquired {
			db.Close()
			return noop, fmt.Errorf("another migration is already running for this database and scripts directory")
		}
		return func() {
			_, _ = db.Exec("SELECT pg_advisory_unlock($1)", lockKey)
			_ = db.Close()
		}, nil
	case models.SQLite:
		lockPath := filepath.Join(os.TempDir(), fmt.Sprintf("sql-toolbox-migration-%d.lock", lockKey))
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil && os.IsExist(err) {
			// Recover from a lock left behind by a crashed process.
			if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > 24*time.Hour {
				if removeErr := os.Remove(lockPath); removeErr == nil {
					lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				}
			}
		}
		if err != nil {
			if os.IsExist(err) {
				return noop, fmt.Errorf("another migration is already running for this database and scripts directory")
			}
			return noop, fmt.Errorf("failed to create migration lock file: %w", err)
		}
		_, _ = fmt.Fprintf(lockFile, "%d\n", os.Getpid())
		_ = lockFile.Close()
		return func() { _ = os.Remove(lockPath) }, nil
	default:
		return noop, nil
	}
}

// fileExists reports whether path exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
