package database

import (
	"database/sql"
	"fmt"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Factory provides database connection creation
type Factory struct {
	connectionParams *models.ConnectionParameters
}

// NewFactory creates a new database factory
func NewFactory(connectionParams *models.ConnectionParameters) *Factory {
	return &Factory{
		connectionParams: connectionParams,
	}
}

// CreateConnection creates a new database connection
func (f *Factory) CreateConnection() (*sql.DB, error) {
	provider := f.connectionParams.GetDbProvider()
	connectionString, err := f.connectionParams.GetConnectionString()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	switch provider {
	case models.PostgreSQL:
		return sql.Open("postgres", connectionString)
	case models.SQLite:
		return sql.Open("sqlite3", connectionString)
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", provider.String())
	}
}

// GetProvider returns the database provider
func (f *Factory) GetProvider() models.DatabaseProvider {
	return f.connectionParams.GetDbProvider()
}
