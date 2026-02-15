// Package app provides the core functionality for the sql-toolbox application
package app

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information
const (
	Version   = "0.8.3"
	EnvPrefix = "SQLTB"
)

// Run executes the main application logic
func Run() {
	var rootCmd = &cobra.Command{
		Use:           "sql-toolbox",
		Short:         "A SQL database toolbox for migrations and backups",
		Long:          "A CLI tool for executing SQL migration scripts and performing database backups for PostgreSQL and SQLite databases.",
		Version:       fmt.Sprintf("%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH),
		SilenceErrors: true, // We handle errors ourselves below
	}

	// Add subcommands
	rootCmd.AddCommand(newMigrationCmd())
	rootCmd.AddCommand(newBackupCmd())

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
