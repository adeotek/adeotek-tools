package app

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration",
		Short: "Execute SQL migration scripts against a database",
		Long:  "Execute SQL migration scripts against PostgreSQL and SQLite databases with migration history tracking.",
		RunE:  runMigration,
	}

	cmd.Flags().StringP("target-path", "t", "", "[required] Target path (path to the SQL scripts directory)")
	cmd.Flags().StringP("provider", "r", "postgresql", "Database provider (postgresql/sqlite)")
	cmd.Flags().StringP("connection-string", "c", "", "Database connection string")
	cmd.Flags().StringP("host", "o", "", "Database host")
	cmd.Flags().IntP("port", "p", 0, "Database port")
	cmd.Flags().StringP("database", "b", "", "Database name")
	cmd.Flags().StringP("user", "u", "", "Database user")
	cmd.Flags().StringP("password", "s", "", "Database password")
	cmd.Flags().BoolP("dry-run", "d", false, "Run in dry-run mode, simulating execution without making changes")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().Bool("backup", false, "Backup the database before applying migrations (only if there are unapplied scripts)")
	cmd.Flags().Bool("backup-only", false, "Create a database backup without running migrations")
	cmd.Flags().Bool("restore", false, "Restore the last database backup and skip running migrations")
	cmd.Flags().StringP("config", "f", "", "Path to a YAML configuration file (optional, CLI flags override config values)")

	// Bind flags to viper
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.BindPFlag("target-path", cmd.Flags().Lookup("target-path"))
	v.BindPFlag("provider", cmd.Flags().Lookup("provider"))
	v.BindPFlag("connection-string", cmd.Flags().Lookup("connection-string"))
	v.BindPFlag("host", cmd.Flags().Lookup("host"))
	v.BindPFlag("port", cmd.Flags().Lookup("port"))
	v.BindPFlag("database", cmd.Flags().Lookup("database"))
	v.BindPFlag("user", cmd.Flags().Lookup("user"))
	v.BindPFlag("password", cmd.Flags().Lookup("password"))
	v.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))
	v.BindPFlag("verbose", cmd.Flags().Lookup("verbose"))
	v.BindPFlag("backup", cmd.Flags().Lookup("backup"))
	v.BindPFlag("backup-only", cmd.Flags().Lookup("backup-only"))
	v.BindPFlag("restore", cmd.Flags().Lookup("restore"))
	v.BindPFlag("config", cmd.Flags().Lookup("config"))

	// Store viper instance in command context
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(setViper(cmd.Context(), v))
		return nil
	}

	return cmd
}

func runMigration(cmd *cobra.Command, args []string) error {
	fmt.Printf("sql-toolbox version %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

	v := getViper(cmd.Context())

	// Load config file if provided
	configPath := v.GetString("config")
	var migConfig *models.MigrationConfig
	if configPath != "" {
		config, err := models.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		migConfig = config.Migration
	}

	// Get configuration values (CLI flags > env vars > config file)
	targetPath := resolveString(v, "target-path", migConfig, func(c *models.MigrationConfig) string { return c.TargetPath })
	provider := resolveString(v, "provider", migConfig, func(c *models.MigrationConfig) string { return c.Provider })
	connectionString := resolveString(v, "connection-string", migConfig, func(c *models.MigrationConfig) string { return c.ConnectionString })
	host := resolveString(v, "host", migConfig, func(c *models.MigrationConfig) string { return c.Host })
	port := resolveInt(v, "port", migConfig, func(c *models.MigrationConfig) int { return c.Port })
	database := resolveString(v, "database", migConfig, func(c *models.MigrationConfig) string { return c.Database })
	user := resolveString(v, "user", migConfig, func(c *models.MigrationConfig) string { return c.User })
	password := resolveString(v, "password", migConfig, func(c *models.MigrationConfig) string { return c.Password })
	isDryRun := v.GetBool("dry-run")
	verbose := v.GetBool("verbose")
	isBackup := resolveBool(v, "backup", migConfig, func(c *models.MigrationConfig) bool { return c.Backup })
	isBackupOnly := resolveBool(v, "backup-only", migConfig, func(c *models.MigrationConfig) bool { return c.BackupOnly })
	isRestore := resolveBool(v, "restore", migConfig, func(c *models.MigrationConfig) bool { return c.Restore })

	// Default provider if empty
	if provider == "" {
		provider = "postgresql"
	}

	// Create connection parameters first (needed for both restore and migration)
	connectionParams, err := models.ParseConnectionParameters(
		provider, connectionString, host, strconv.Itoa(port), database, user, password)
	if err != nil {
		return fmt.Errorf("failed to parse connection parameters: %v", err)
	}

	// Validate connection parameters
	if valid, errors := connectionParams.IsValid(); !valid {
		return fmt.Errorf("invalid connection parameters: %v", errors)
	}

	// Create backup service
	backupService := services.NewBackupService(isDryRun, verbose, "")

	// Handle restore flag
	if isRestore {
		fmt.Println("Restoring last database backup...")
		err = backupService.RestoreLastBackup(connectionParams)
		if err != nil {
			return fmt.Errorf("restore failed: %v", err)
		}
		fmt.Println("Restore DONE!")
		return nil
	}

	// Handle backup-only flag
	if isBackupOnly {
		fmt.Println("Creating database backup...")
		backupPath, err := backupService.CreateBackup(connectionParams)
		if err != nil {
			return fmt.Errorf("backup failed: %v", err)
		}
		fmt.Printf("Backup created successfully: %s\n", backupPath)
		fmt.Println("Backup DONE!")
		return nil
	}

	// Validate target path (only required for migrations)
	if targetPath == "" {
		return fmt.Errorf("--target-path is required for migration operations")
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("target-path '%s' does not exist", targetPath)
	}

	// Check if directory is empty
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %v", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("target-path directory '%s' is empty", targetPath)
	}

	// Show configuration in verbose mode
	if verbose {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Target Path: %s\n", targetPath)
		fmt.Printf("  Provider: %s\n", provider)
		fmt.Printf("  Database: %s\n", database)
		if host != "" {
			fmt.Printf("  Host: %s\n", host)
		}
		if port != 0 {
			fmt.Printf("  Port: %d\n", port)
		}
		if user != "" {
			fmt.Printf("  User: %s\n", user)
		}
		fmt.Printf("  Dry Run: %t\n", isDryRun)
		fmt.Printf("  Verbose: %t\n", verbose)
		fmt.Println()
	}

	// Create and run migration service
	var backupServiceForMigration *services.BackupService
	if isBackup {
		backupServiceForMigration = backupService
	}
	migrationService := services.NewMigrationService(isDryRun, verbose, backupServiceForMigration)

	if isDryRun {
		fmt.Println("Executing SQL migration command in [DryRun] mode...")
	} else {
		fmt.Println("Executing SQL migration command...")
	}

	err = migrationService.Run(targetPath, connectionParams)
	if err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	fmt.Println("Migration DONE!")
	return nil
}

// resolveString returns the viper value if set via CLI/env, otherwise falls back to config file value
func resolveString(v *viper.Viper, key string, cfg *models.MigrationConfig, getter func(*models.MigrationConfig) string) string {
	val := v.GetString(key)
	if val != "" {
		return val
	}
	if cfg != nil {
		return getter(cfg)
	}
	return ""
}

// resolveInt returns the viper value if set via CLI/env, otherwise falls back to config file value
func resolveInt(v *viper.Viper, key string, cfg *models.MigrationConfig, getter func(*models.MigrationConfig) int) int {
	val := v.GetInt(key)
	if val != 0 {
		return val
	}
	if cfg != nil {
		return getter(cfg)
	}
	return 0
}

// resolveBool returns the viper value if set via CLI/env, otherwise falls back to config file value
func resolveBool(v *viper.Viper, key string, cfg *models.MigrationConfig, getter func(*models.MigrationConfig) bool) bool {
	if v.GetBool(key) {
		return true
	}
	if cfg != nil {
		return getter(cfg)
	}
	return false
}
