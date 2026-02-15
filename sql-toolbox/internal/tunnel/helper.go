package tunnel

import (
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// EstablishIfNeeded creates and establishes an SSH tunnel if the config is enabled
// Returns nil if config is nil or disabled
func EstablishIfNeeded(config *models.SSHTunnelConfig, dbHost string, dbPort int) (*SSHTunnel, error) {
	if config == nil || !config.Enabled {
		return nil, nil
	}

	return NewSSHTunnel(config, dbHost, dbPort)
}
