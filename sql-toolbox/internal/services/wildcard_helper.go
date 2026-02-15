package services

import (
	"fmt"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/tunnel"
)

// ExpandWildcardDatabasesWithTunnel expands a wildcard database target, establishing SSH tunnel if needed
func ExpandWildcardDatabasesWithTunnel(template models.DatabaseTarget, verbose bool) ([]models.DatabaseTarget, error) {
	if template.Database != "*" {
		return []models.DatabaseTarget{template}, nil // Not a wildcard, return as-is
	}

	// Establish SSH tunnel if configured
	var sshTunnel *tunnel.SSHTunnel
	sshConfig := template.GetEffectiveSSHTunnel()
	if sshConfig != nil && sshConfig.Enabled {
		var err error
		sshTunnel, err = tunnel.EstablishIfNeeded(sshConfig, template.Host, template.Port)
		if err != nil {
			return nil, fmt.Errorf("failed to establish SSH tunnel for wildcard expansion: %w", err)
		}
		if sshTunnel != nil {
			defer sshTunnel.Close()
		}
	}

	// Build connection string for querying databases
	var connStr string
	if sshTunnel != nil {
		// Use SSH tunnel's local address
		connStr = fmt.Sprintf("host=127.0.0.1 port=%d dbname=postgres user=%s password=%s sslmode=disable",
			sshTunnel.GetLocalPort(), template.User, template.Password)
	} else {
		// Use direct connection (connStr will be built in QueryDatabases)
		connStr = ""
	}

	// Delegate to the models function with the connection string
	return models.ExpandWildcardDatabasesWithConnStr(template, verbose, connStr)
}

// ExpandWildcardsWithTunnels processes all DatabaseTarget entries in a BackupConfig and expands
// any with database: "*", establishing SSH tunnels as needed
func ExpandWildcardsWithTunnels(config *models.BackupConfig, verbose bool) error {
	var expandedDatabases []models.DatabaseTarget

	for _, dbTarget := range config.Databases {
		expanded, err := ExpandWildcardDatabasesWithTunnel(dbTarget, verbose)
		if err != nil {
			return fmt.Errorf("failed to expand wildcard for '%s': %w", dbTarget.Name, err)
		}
		expandedDatabases = append(expandedDatabases, expanded...)
	}

	config.Databases = expandedDatabases

	// Re-link defaults after expansion
	for i := range config.Databases {
		config.Databases[i].SetDefaults(&config.Defaults, &config.S3)
	}

	return nil
}
