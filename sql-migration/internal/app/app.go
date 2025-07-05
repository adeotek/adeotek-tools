// Package app provides the core functionality for the sql-migration application
package app

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/adeotek/adeotek-tools/sql-migration/internal/models"
	"github.com/adeotek/adeotek-tools/sql-migration/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information
const (
	Version = "0.3.0"
	EnvPrefix = "CLI_SQL_MIGRATION"
)

// Run executes the main application logic
func Run() {
	var rootCmd = &cobra.Command{
		Use:   "sql-migration",
		Short: "Executes SQL migration scripts against a database",
		Long:  "A CLI tool for executing SQL migration scripts against PostgreSQL and SQLite databases with migration history tracking.",
		Run:   runMigration,
	}

	// Add flags
	rootCmd.Flags().StringP("target-path", "t", "", "[required] Target path (path to the SQL scripts directory)")
	rootCmd.Flags().StringP("provider", "r", "postgresql", "Database provider (postgresql/sqlite)")
	rootCmd.Flags().StringP("connection-string", "c", "", "Database connection string")
	rootCmd.Flags().StringP("host", "o", "", "Database host")
	rootCmd.Flags().IntP("port", "p", 0, "Database port")
	rootCmd.Flags().StringP("database", "b", "", "Database name")
	rootCmd.Flags().StringP("user", "u", "", "Database user")
	rootCmd.Flags().StringP("password", "s", "", "Database password")
	rootCmd.Flags().BoolP("dry-run", "d", false, "Run in dry-run mode, simulating execution without making changes")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().Bool("version", false, "Show version information and exit")

		// Set up environment variable support
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind flags to viper
	viper.BindPFlag("target-path", rootCmd.Flags().Lookup("target-path"))
	viper.BindPFlag("provider", rootCmd.Flags().Lookup("provider"))
	viper.BindPFlag("connection-string", rootCmd.Flags().Lookup("connection-string"))
	viper.BindPFlag("host", rootCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("database", rootCmd.Flags().Lookup("database"))
	viper.BindPFlag("user", rootCmd.Flags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.Flags().Lookup("password"))
	viper.BindPFlag("dry-run", rootCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("version", rootCmd.Flags().Lookup("version"))

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func hasVersionFlag() bool {
	for _, arg := range os.Args {
		if arg == "--version" {
			return true
		}
	}
	return false
}

func runMigration(cmd *cobra.Command, args []string) {
	// Show version
	fmt.Printf("sql-migration version %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

	// Check if only version was requested
	if viper.GetBool("version") || hasVersionFlag() {
		return
	}

	// Get configuration values
	targetPath := viper.GetString("target-path")
	provider := viper.GetString("provider")
	connectionString := viper.GetString("connection-string")
	host := viper.GetString("host")
	port := viper.GetInt("port")
	database := viper.GetString("database")
	user := viper.GetString("user")
	password := viper.GetString("password")
	isDryRun := viper.GetBool("dry-run")
	verbose := viper.GetBool("verbose")

	// Validate target path
	if targetPath == "" {
		log.Fatal("--target-path is required")
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		log.Fatalf("target-path '%s' does not exist", targetPath)
	}

	// Check if directory is empty
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		log.Fatalf("failed to read target directory: %v", err)
	}
	if len(entries) == 0 {
		log.Fatalf("target-path directory '%s' is empty", targetPath)
	}

	// Create connection parameters
	connectionParams, err := models.ParseConnectionParameters(
		provider, connectionString, host, strconv.Itoa(port), database, user, password)
	if err != nil {
		log.Fatalf("failed to parse connection parameters: %v", err)
	}

	// Validate connection parameters
	if valid, errors := connectionParams.IsValid(); !valid {
		log.Fatalf("invalid connection parameters: %v", errors)
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
	migrationService := services.NewMigrationService(isDryRun, verbose)

	if isDryRun {
		fmt.Println("Executing SQL migration command in [DryRun] mode...")
	} else {
		fmt.Println("Executing SQL migration command...")
	}

	err = migrationService.Run(targetPath, connectionParams)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration DONE!")
}
