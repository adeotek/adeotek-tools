package app

import (
	"fmt"
	"log"
	"runtime"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/services"
	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Back up one or more PostgreSQL databases using a YAML config file",
		Long: `Back up one or more remote PostgreSQL databases using a YAML configuration file.
Supports optional gzip compression and S3/S3-compatible storage upload.
Uses pure Go PostgreSQL dump (no pg_dump binary required).`,
		RunE:         runBackup,
		SilenceUsage: true, // Don't show usage on execution errors, only on argument errors
	}

	cmd.Flags().StringP("config", "c", "", "[required] Path to the YAML configuration file")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolP("dry-run", "d", false, "Run in dry-run mode, simulating execution without making changes")
	cmd.MarkFlagRequired("config")

	return cmd
}

func runBackup(cmd *cobra.Command, args []string) error {
	fmt.Printf("sql-toolbox version %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

	configPath, _ := cmd.Flags().GetString("config")
	verbose, _ := cmd.Flags().GetBool("verbose")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if verbose {
		log.Printf("Loading backup configuration from: %s", configPath)
	}

	// Load unified config and extract backup section
	unifiedConfig, err := models.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if unifiedConfig.Backup == nil {
		return fmt.Errorf("no 'backup' section found in config file")
	}

	config := unifiedConfig.Backup

	// Expand wildcard database entries (with SSH tunnel support)
	if err := services.ExpandWildcardsWithTunnels(config, verbose); err != nil {
		return fmt.Errorf("failed to expand wildcard databases: %w", err)
	}

	// Validate the backup config (validation happens in LoadBackupConfig, but check count)
	if len(config.Databases) == 0 {
		return fmt.Errorf("no databases to backup after wildcard expansion")
	}

	// Link defaults and S3 config to each database target
	for i := range config.Databases {
		config.Databases[i].SetDefaults(&config.Defaults, &config.S3)
	}

	if verbose {
		log.Printf("Configuration loaded: %d database(s) configured", len(config.Databases))
		if config.S3.Enabled {
			log.Printf("S3 upload enabled: bucket=%s, region=%s", config.S3.Bucket, config.S3.Region)
		}
	}

	orchestrator := services.NewBackupOnlyOrchestrator(config, verbose, dryRun)
	summary, err := orchestrator.Run()
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Print summary
	fmt.Println()
	fmt.Println("=== Backup Summary ===")
	fmt.Printf("Total: %d, Succeeded: %d, Failed: %d\n",
		summary.Total, summary.Succeeded, summary.Failed)
	for _, result := range summary.Results {
		if result.Error != nil {
			fmt.Printf("  [FAIL] %s: %v\n", result.DatabaseName, result.Error)
		} else {
			fmt.Printf("  [OK]   %s -> %s", result.DatabaseName, result.FilePath)
			if result.UploadedToS3 {
				fmt.Printf(" (uploaded to S3: %s)", result.S3Key)
			}
			fmt.Println()
		}
	}

	if summary.Failed > 0 {
		return fmt.Errorf("%d database backup(s) failed", summary.Failed)
	}

	fmt.Println("Backup DONE!")
	return nil
}
