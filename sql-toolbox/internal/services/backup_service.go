package services

import (
	"compress/gzip"
	"database/sql"
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
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const backupDirectoryName = ".db-backups"

// BackupService handles database backup and restore operations
type BackupService struct {
	isDryRun     bool
	verbose      bool
	backupMethod string
}

// NewBackupService creates a new backup service instance
func NewBackupService(isDryRun, verbose bool, backupMethod string) *BackupService {
	if backupMethod == "" {
		backupMethod = BackupMethodPgDump
	}
	return &BackupService{
		isDryRun:     isDryRun,
		verbose:      verbose,
		backupMethod: backupMethod,
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

	restorePath := lastBackupPath
	if strings.HasSuffix(restorePath, ".gz") {
		decompressed, cleanup, err := decompressBackup(restorePath)
		if err != nil {
			return err
		}
		defer cleanup()
		restorePath = decompressed
	}

	provider := connectionParams.GetDbProvider()
	switch provider {
	case models.PostgreSQL:
		return bs.restorePostgresBackup(connectionParams, restorePath)
	case models.SQLite:
		return bs.restoreSQLiteBackup(connectionParams, restorePath)
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
	var extensions []string
	if provider == models.PostgreSQL {
		// Support both .sql (pure Go) and .dump (pg_dump custom format),
		// including gzip-compressed backups produced by the backup subcommand.
		extensions = []string{".sql.gz", ".dump.gz", ".sql", ".dump"}
	} else {
		extensions = []string{".db"}
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return "", fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Only consider backups belonging to this database; otherwise a restore of
	// database A could silently apply the most recent backup of database B.
	dbPrefix := bs.getDatabaseName(connectionParams) + "_backup_"
	var backupFiles []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), dbPrefix) {
			continue
		}
		for _, ext := range extensions {
			if strings.HasSuffix(entry.Name(), ext) {
				backupFiles = append(backupFiles, filepath.Join(backupDir, entry.Name()))
				break
			}
		}
	}

	if len(backupFiles) == 0 {
		return "", nil
	}

	// Sort by modification time (most recent first); skip files we cannot stat
	type backupFileInfo struct {
		path    string
		modTime time.Time
	}
	infos := make([]backupFileInfo, 0, len(backupFiles))
	for _, path := range backupFiles {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		infos = append(infos, backupFileInfo{path: path, modTime: info.ModTime()})
	}
	if len(infos) == 0 {
		return "", nil
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.After(infos[j].modTime)
	})

	return infos[0].path, nil
}

func (bs *BackupService) createPostgresBackup(connectionParams *models.ConnectionParameters, backupDir, backupFileName string) (string, error) {
	// Use .dump extension for pg_dump custom format, .sql for pure Go plain format
	extension := ".sql"
	if bs.backupMethod == BackupMethodPgDump {
		extension = ".dump"
	}
	backupPath := filepath.Join(backupDir, backupFileName+extension)
	log.Printf("Creating PostgreSQL backup: %s (method: %s)", backupPath, bs.backupMethod)

	if bs.isDryRun {
		log.Printf("Dry run: would create PostgreSQL backup at %s (method: %s)", backupPath, bs.backupMethod)
		return backupPath, nil
	}

	options := PgDumpOptions{
		NoOwner: true,
		Clean:   true,
	}

	// For "go" method, we need a database connection
	var db *sql.DB
	if bs.backupMethod == BackupMethodGo {
		connStr, err := connectionParams.GetConnectionString()
		if err != nil {
			return "", fmt.Errorf("failed to build connection string: %w", err)
		}
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return "", fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			return "", fmt.Errorf("failed to ping database: %w", err)
		}
	}

	dumper, err := NewDatabaseDumper(bs.backupMethod, connectionParams, db, options)
	if err != nil {
		return "", fmt.Errorf("failed to create database dumper: %w", err)
	}

	file, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	if err := bs.dumpToFile(file, backupPath, dumper); err != nil {
		return "", err
	}

	log.Printf("PostgreSQL backup created successfully: %s (method: %s)", backupPath, bs.backupMethod)
	return backupPath, nil
}

// dumpToFile runs the dump into file and removes the partially written backup on failure,
// so a failed dump never leaves a truncated .sql/.dump file behind that could later be
// picked as "the last backup".
func (bs *BackupService) dumpToFile(file *os.File, backupPath string, dumper DatabaseDumper) error {
	dumpSucceeded := false
	defer func() {
		if !dumpSucceeded {
			file.Close()
			_ = os.Remove(backupPath)
		}
	}()

	if err := dumper.Dump(file); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close backup file: %w", err)
	}
	dumpSucceeded = true
	return nil
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

	if isSQLiteDatabase(sourcePath) {
		// VACUUM INTO produces a consistent snapshot, including committed data
		// still living in the WAL file.
		if err := vacuumSQLiteInto(sourcePath, backupPath); err != nil {
			return "", fmt.Errorf("failed to back up SQLite database: %w", err)
		}
		if err := os.Chmod(backupPath, 0600); err != nil {
			return "", fmt.Errorf("failed to set backup file permissions: %w", err)
		}
		log.Printf("SQLite backup created successfully: %s", backupPath)
		return backupPath, nil
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	// Remove a partially copied backup on failure (e.g. disk full mid-copy).
	copySucceeded := false
	defer func() {
		if !copySucceeded {
			destFile.Close()
			_ = os.Remove(backupPath)
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("failed to copy database file: %w", err)
	}
	if err := destFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close backup file: %w", err)
	}
	copySucceeded = true

	log.Printf("SQLite backup created successfully: %s", backupPath)
	return backupPath, nil
}

func (bs *BackupService) restorePostgresBackup(connectionParams *models.ConnectionParameters, backupPath string) error {
	log.Printf("Restoring PostgreSQL backup from: %s", backupPath)

	// Build connection arguments; a raw connection string takes precedence and must be
	// normalized (lib/pq format) because pg_restore/psql accept conninfo strings
	var connArgs []string
	password := connectionParams.Password
	if connectionParams.RawConnectionString != "" {
		connStr, err := connectionParams.GetConnectionString()
		if err != nil {
			return fmt.Errorf("failed to build connection string: %w", err)
		}
		// Keep the password out of the process arguments.
		connStr, extractedPassword := models.RemovePasswordFromConnInfo(connStr)
		if extractedPassword != "" {
			password = extractedPassword
		}
		connArgs = []string{fmt.Sprintf("--dbname=%s", connStr)}
	} else {
		host := connectionParams.Host
		port := connectionParams.Port
		if activeTunnel := connectionParams.GetActiveTunnel(); activeTunnel != nil {
			type localPortGetter interface {
				GetLocalPort() int
			}
			if t, ok := activeTunnel.(localPortGetter); ok {
				host = "127.0.0.1"
				port = t.GetLocalPort()
			}
		}
		connArgs = []string{
			fmt.Sprintf("--host=%s", host),
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--username=%s", connectionParams.User),
			fmt.Sprintf("--dbname=%s", connectionParams.DatabaseName),
		}
	}

	var cmd *exec.Cmd
	var cmdName string

	// Use pg_restore for custom format (.dump), psql for plain format (.sql).
	// Plain SQL restores must stop on the first error and run atomically; psql
	// otherwise continues and can exit 0 after failed statements.
	if strings.HasSuffix(backupPath, ".dump") {
		cmdName = "pg_restore"
		args := append(connArgs, "--clean", "--if-exists", "--exit-on-error", backupPath)
		cmd = exec.Command(cmdName, args...)
	} else {
		cmdName = "psql"
		args := append(connArgs, "--single-transaction", "-v", "ON_ERROR_STOP=1",
			fmt.Sprintf("--file=%s", backupPath))
		cmd = exec.Command(cmdName, args...)
	}

	if password != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s restore failed: %w\nOutput: %s", cmdName, err, string(output))
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

	// Copy to a temporary file in the target directory and rename it into place, so a
	// failed restore never leaves a half-written (and therefore corrupt) database file.
	tmpFile, err := os.CreateTemp(filepath.Dir(targetPath), ".restore-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary restore file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := io.Copy(tmpFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy backup file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to flush restore file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close restore file: %w", err)
	}
	// Preserve the target's permissions, or use private permissions for new files.
	mode := os.FileMode(0600)
	if info, err := os.Stat(targetPath); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return fmt.Errorf("failed to set restore file permissions: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("failed to replace database file: %w", err)
	}
	// Stale WAL/SHM files from the replaced database would corrupt or roll back
	// the restored snapshot, so remove them.
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.Remove(targetPath + suffix); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove stale SQLite sidecar %s: %w", suffix, err)
		}
	}

	log.Printf("SQLite backup restored successfully")
	return nil
}

// decompressBackup writes a gunzipped copy of path to a temporary file and
// returns its name plus a cleanup function.
func decompressBackup(path string) (string, func(), error) {
	cleanup := func() {}
	source, err := os.Open(path)
	if err != nil {
		return "", cleanup, fmt.Errorf("failed to open compressed backup: %w", err)
	}
	defer source.Close()

	reader, err := gzip.NewReader(source)
	if err != nil {
		return "", cleanup, fmt.Errorf("failed to read compressed backup: %w", err)
	}
	defer reader.Close()

	// Keep the inner extension so restore can still choose pg_restore vs psql.
	inner := strings.TrimSuffix(path, ".gz")
	tmp, err := os.CreateTemp("", "sql-toolbox-restore-*"+filepath.Ext(inner))
	if err != nil {
		return "", cleanup, fmt.Errorf("failed to create temporary restore file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup = func() { _ = os.Remove(tmpPath) }

	if _, err := io.Copy(tmp, reader); err != nil {
		tmp.Close()
		cleanup()
		return "", func() {}, fmt.Errorf("failed to decompress backup: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("failed to close temporary restore file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("failed to set temporary restore permissions: %w", err)
	}
	return tmpPath, cleanup, nil
}

// isSQLiteDatabase reports whether path starts with the SQLite file header.
func isSQLiteDatabase(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 16)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return string(header) == "SQLite format 3\x00"
}

// vacuumSQLiteInto writes a consistent snapshot of sourcePath to destPath.
func vacuumSQLiteInto(sourcePath, destPath string) (err error) {
	db, err := sql.Open("sqlite3", sourcePath)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	defer func() {
		if err != nil {
			_ = os.Remove(destPath)
		}
	}()

	quotedDest := "'" + strings.ReplaceAll(destPath, "'", "''") + "'"
	_, err = db.Exec("VACUUM INTO " + quotedDest)
	return err
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
	return sanitizeFileName(dbName)
}

// sanitizeFileName replaces characters that are unsafe in backup file names.
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}
