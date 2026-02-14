package services

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/tunnel"
	_ "github.com/lib/pq"
)

// DatabaseBackupResult holds the result of a single database backup
type DatabaseBackupResult struct {
	DatabaseName string
	FilePath     string
	UploadedToS3 bool
	S3Key        string
	Error        error
}

// BackupSummary holds the overall results of the backup-only operation
type BackupSummary struct {
	Total     int
	Succeeded int
	Failed    int
	Results   []DatabaseBackupResult
}

// BackupOnlyOrchestrator coordinates the multi-database backup workflow
type BackupOnlyOrchestrator struct {
	config  *models.BackupConfig
	verbose bool
	dryRun  bool

	// Overridable for testing
	s3UploaderFactory func(models.S3Config) (S3Uploader, error)
	dbConnector       func(connStr string) (*sql.DB, error)
}

// NewBackupOnlyOrchestrator creates a new orchestrator
func NewBackupOnlyOrchestrator(config *models.BackupConfig, verbose, dryRun bool) *BackupOnlyOrchestrator {
	return &BackupOnlyOrchestrator{
		config:  config,
		verbose: verbose,
		dryRun:  dryRun,
		s3UploaderFactory: func(cfg models.S3Config) (S3Uploader, error) {
			return NewS3UploadService(cfg)
		},
		dbConnector: func(connStr string) (*sql.DB, error) {
			return sql.Open("postgres", connStr)
		},
	}
}

// Run executes the backup for all configured databases
func (o *BackupOnlyOrchestrator) Run() (*BackupSummary, error) {
	summary := &BackupSummary{
		Total: len(o.config.Databases),
	}

	// Set up S3 uploader if needed
	var s3Uploader S3Uploader
	if o.config.S3.Enabled {
		var err error
		s3Uploader, err = o.s3UploaderFactory(o.config.S3)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 uploader: %w", err)
		}
	}

	for _, dbTarget := range o.config.Databases {
		result := o.backupDatabase(dbTarget, s3Uploader)
		summary.Results = append(summary.Results, result)
		if result.Error != nil {
			summary.Failed++
			log.Printf("[FAIL] %s: %v", dbTarget.Name, result.Error)
		} else {
			summary.Succeeded++
			if o.verbose {
				log.Printf("[OK] %s -> %s", dbTarget.Name, result.FilePath)
			}
		}
	}

	return summary, nil
}

func (o *BackupOnlyOrchestrator) backupDatabase(dbTarget models.DatabaseTarget, s3Uploader S3Uploader) DatabaseBackupResult {
	result := DatabaseBackupResult{
		DatabaseName: dbTarget.Name,
	}

	if o.verbose {
		log.Printf("Starting backup for database: %s", dbTarget.Name)
	}

	// Establish SSH tunnel if configured
	var sshTunnel *tunnel.SSHTunnel
	sshConfig := dbTarget.GetEffectiveSSHTunnel()
	if sshConfig != nil && sshConfig.Enabled && !o.dryRun {
		if o.verbose {
			log.Printf("Establishing SSH tunnel for %s to %s:%d...",
				dbTarget.Name, sshConfig.Host, sshConfig.GetEffectivePort())
		}

		var err error
		sshTunnel, err = tunnel.EstablishIfNeeded(sshConfig, dbTarget.Host, dbTarget.Port)
		if err != nil {
			result.Error = fmt.Errorf("failed to establish SSH tunnel: %w", err)
			return result
		}

		if sshTunnel != nil {
			defer sshTunnel.Close()
			if o.verbose {
				log.Printf("SSH tunnel established for %s: %s -> %s:%d",
					dbTarget.Name, sshTunnel.GetLocalAddress(), dbTarget.Host, dbTarget.Port)
			}
		}
	}

	// Create output directory
	outputDir := dbTarget.GetEffectiveOutputDir()
	if !o.dryRun {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			result.Error = fmt.Errorf("failed to create output directory: %w", err)
			return result
		}
	}

	// Build file path
	timestamp := time.Now().UTC().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_backup_%s.sql", dbTarget.Name, timestamp)
	compress := dbTarget.GetEffectiveCompress()
	if compress {
		fileName += ".gz"
	}
	filePath := filepath.Join(outputDir, fileName)
	result.FilePath = filePath

	if o.dryRun {
		log.Printf("[DRY RUN] Would backup %s to %s", dbTarget.Name, filePath)
		if dbTarget.GetEffectiveUploadToS3() && o.config.S3.Enabled {
			s3Prefix := dbTarget.GetEffectiveS3Prefix()
			s3Key := BuildS3Key(s3Prefix, dbTarget.Name, filePath)
			log.Printf("[DRY RUN] Would upload to S3: %s", s3Key)
			result.UploadedToS3 = true
			result.S3Key = s3Key
		}
		return result
	}

	// Create output file with optional compression
	writer, closer, err := createOutputWriter(filePath, compress)
	if err != nil {
		result.Error = fmt.Errorf("failed to create output file: %w", err)
		return result
	}
	defer closer()

	// Run dump using the configured backup method
	method := dbTarget.GetEffectiveBackupMethod()
	options := PgDumpOptions{
		Schemas:       dbTarget.GetEffectiveSchemas(),
		ExcludeTables: dbTarget.ExcludeTables,
		NoOwner:       dbTarget.GetEffectiveNoOwner(),
		Clean:         dbTarget.GetEffectiveClean(),
	}

	if method == BackupMethodPgDump {
		// Use external pg_dump binary - no DB connection needed
		connParams := &models.ConnectionParameters{
			Provider:     "postgresql",
			Host:         dbTarget.Host,
			Port:         dbTarget.Port,
			DatabaseName: dbTarget.Database,
			User:         dbTarget.User,
			Password:     dbTarget.Password,
		}
		if dbTarget.ConnectionString != "" {
			connParams.RawConnectionString = dbTarget.ConnectionString
		}

		// Set active tunnel if established
		if sshTunnel != nil {
			connParams.SetActiveTunnel(sshTunnel)
		}

		dumper, dErr := NewDatabaseDumper(BackupMethodPgDump, connParams, nil, options)
		if dErr != nil {
			result.Error = fmt.Errorf("failed to create dumper: %w", dErr)
			return result
		}

		if dErr = dumper.Dump(writer); dErr != nil {
			result.Error = fmt.Errorf("dump failed: %w", dErr)
			return result
		}
	} else {
		// Use pure Go dump - needs DB connection
		connStr := dbTarget.ToConnectionString()

		// If SSH tunnel is active, modify connection string to use tunnel's local address
		if sshTunnel != nil {
			connStr = fmt.Sprintf("host=127.0.0.1 port=%d dbname=%s user=%s password=%s sslmode=disable",
				sshTunnel.GetLocalPort(), dbTarget.Database, dbTarget.User, dbTarget.Password)
		}

		db, dErr := o.dbConnector(connStr)
		if dErr != nil {
			result.Error = fmt.Errorf("failed to connect to database: %w", dErr)
			return result
		}
		defer db.Close()

		if dErr = db.Ping(); dErr != nil {
			result.Error = fmt.Errorf("failed to ping database: %w", dErr)
			return result
		}

		dumper, dErr := NewDatabaseDumper(BackupMethodGo, nil, db, options)
		if dErr != nil {
			result.Error = fmt.Errorf("failed to create dumper: %w", dErr)
			return result
		}

		if dErr = dumper.Dump(writer); dErr != nil {
			result.Error = fmt.Errorf("dump failed: %w", dErr)
			return result
		}
	}

	if o.verbose {
		log.Printf("Dump completed: %s (method: %s)", filePath, method)
	}

	// Upload to S3 if configured
	if dbTarget.GetEffectiveUploadToS3() && o.config.S3.Enabled && s3Uploader != nil {
		s3Prefix := dbTarget.GetEffectiveS3Prefix()
		s3Key := BuildS3Key(s3Prefix, dbTarget.Name, filePath)
		if o.verbose {
			log.Printf("Uploading to S3: %s", s3Key)
		}
		if err := s3Uploader.Upload(filePath, s3Key); err != nil {
			result.Error = fmt.Errorf("S3 upload failed: %w", err)
			return result
		}
		result.UploadedToS3 = true
		result.S3Key = s3Key

		// Delete local file after upload if configured
		if dbTarget.GetEffectiveDeleteLocalAfterUpload() {
			if o.verbose {
				log.Printf("Deleting local file after S3 upload: %s", filePath)
			}
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: failed to delete local file after upload: %v", err)
			}
		}
	}

	return result
}

// createOutputWriter creates a writer for the dump output, optionally wrapping with gzip
func createOutputWriter(path string, compress bool) (io.Writer, func() error, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}

	if !compress {
		return file, file.Close, nil
	}

	gzWriter := gzip.NewWriter(file)
	closer := func() error {
		if err := gzWriter.Close(); err != nil {
			file.Close()
			return err
		}
		return file.Close()
	}
	return gzWriter, closer, nil
}
