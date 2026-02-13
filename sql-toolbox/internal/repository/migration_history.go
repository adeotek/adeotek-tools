package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/database"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// MigrationHistoryRepository handles migration history operations
type MigrationHistoryRepository struct {
	dbFactory *database.Factory
}

// NewMigrationHistoryRepository creates a new migration history repository
func NewMigrationHistoryRepository(dbFactory *database.Factory) *MigrationHistoryRepository {
	return &MigrationHistoryRepository{
		dbFactory: dbFactory,
	}
}

const tableName = "__migrations_history"

// IsHistoryTableCreated checks if the migration history table exists
func (r *MigrationHistoryRepository) IsHistoryTableCreated() (bool, error) {
	db, err := r.dbFactory.CreateConnection()
	if err != nil {
		return false, fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	var query string
	provider := r.dbFactory.GetProvider()

	switch provider {
	case models.PostgreSQL:
		query = fmt.Sprintf("SELECT count(*) FROM information_schema.tables WHERE table_name = '%s'", tableName)
	case models.SQLite:
		query = fmt.Sprintf("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='%s'", tableName)
	default:
		return false, fmt.Errorf("unsupported database provider: %s", provider.String())
	}

	var count int
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if history table exists: %w", err)
	}

	return count > 0, nil
}

// CreateHistoryTable creates the migration history table
func (r *MigrationHistoryRepository) CreateHistoryTable() error {
	db, err := r.dbFactory.CreateConnection()
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	var query string
	provider := r.dbFactory.GetProvider()

	switch provider {
	case models.PostgreSQL:
		query = fmt.Sprintf(`CREATE TABLE "%s" (
			"ScriptFile" varchar(255) NOT NULL,
			"ScriptHash" varchar(150) NOT NULL,
			"ExecutedAt" timestamptz DEFAULT now() NOT NULL,
			CONSTRAINT "%s_pk" PRIMARY KEY ("ScriptFile")
		)`, tableName, tableName)
	case models.SQLite:
		query = fmt.Sprintf(`CREATE TABLE "%s" (
			"ScriptFile" TEXT NOT NULL PRIMARY KEY,
			"ScriptHash" TEXT NOT NULL,
			"ExecutedAt" DATETIME NOT NULL
		)`, tableName)
	default:
		return fmt.Errorf("unsupported database provider: %s", provider.String())
	}

	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create history table: %w", err)
	}

	return nil
}

// GetExecutedScripts returns all executed scripts from the history table
func (r *MigrationHistoryRepository) GetExecutedScripts() ([]models.ScriptExecutionHistory, error) {
	db, err := r.dbFactory.CreateConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`SELECT "ScriptFile", "ScriptHash", "ExecutedAt" FROM "%s"`, tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query executed scripts: %w", err)
	}
	defer rows.Close()

	var scripts []models.ScriptExecutionHistory
	for rows.Next() {
		var script models.ScriptExecutionHistory
		err := rows.Scan(&script.ScriptFile, &script.ScriptHash, &script.ExecutedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan script execution history: %w", err)
		}
		scripts = append(scripts, script)
	}

	return scripts, nil
}

// UpsertExecutedScript inserts or updates a script execution record
func (r *MigrationHistoryRepository) UpsertExecutedScript(script *models.ScriptExecutionHistory) error {
	db, err := r.dbFactory.CreateConnection()
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	// Check if script already exists
	selectQuery := fmt.Sprintf(`SELECT "ScriptFile" FROM "%s" WHERE "ScriptFile" = $1`, tableName)
	var existingScript string
	err = db.QueryRow(selectQuery, script.ScriptFile).Scan(&existingScript)

	if err == sql.ErrNoRows {
		// Insert new record
		insertQuery := fmt.Sprintf(`INSERT INTO "%s" ("ScriptFile", "ScriptHash", "ExecutedAt") VALUES ($1, $2, $3)`, tableName)
		_, err = db.Exec(insertQuery, script.ScriptFile, script.ScriptHash, script.ExecutedAt)
		if err != nil {
			return fmt.Errorf("failed to insert script execution history: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing script: %w", err)
	} else {
		// Update existing record
		script.ExecutedAt = time.Now().UTC()
		updateQuery := fmt.Sprintf(`UPDATE "%s" SET "ScriptHash" = $1, "ExecutedAt" = $2 WHERE "ScriptFile" = $3`, tableName)
		_, err = db.Exec(updateQuery, script.ScriptHash, script.ExecutedAt, script.ScriptFile)
		if err != nil {
			return fmt.Errorf("failed to update script execution history: %w", err)
		}
	}

	return nil
}
