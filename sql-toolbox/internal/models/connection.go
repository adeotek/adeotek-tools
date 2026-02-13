package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// DatabaseProvider represents the supported database providers
type DatabaseProvider int

const (
	Unknown DatabaseProvider = iota
	PostgreSQL
	SQLite
)

// String returns the string representation of the database provider
func (dp DatabaseProvider) String() string {
	switch dp {
	case PostgreSQL:
		return "PostgreSQL"
	case SQLite:
		return "SQLite"
	default:
		return "Unknown"
	}
}

// ParseDatabaseProvider parses a string into a DatabaseProvider
func ParseDatabaseProvider(provider string) DatabaseProvider {
	switch strings.ToLower(provider) {
	case "postgresql", "postgres":
		return PostgreSQL
	case "sqlite":
		return SQLite
	default:
		return Unknown
	}
}

// ConnectionParameters holds the database connection parameters
type ConnectionParameters struct {
	Provider           string
	RawConnectionString string
	Host               string
	Port               int
	DatabaseName       string
	User               string
	Password           string
}

// GetDbProvider returns the parsed database provider
func (cp *ConnectionParameters) GetDbProvider() DatabaseProvider {
	return ParseDatabaseProvider(cp.Provider)
}

// IsValid validates the connection parameters
func (cp *ConnectionParameters) IsValid() (bool, []string) {
	var errors []string
	provider := cp.GetDbProvider()

	if provider == Unknown {
		errors = append(errors, "Invalid or missing Database provider.")
	}

	if cp.RawConnectionString != "" {
		return len(errors) == 0, errors
	}

	if cp.Host == "" && cp.Port == 0 && cp.DatabaseName == "" {
		errors = append(errors, "Either the Database connection string or the individual parameters must be provided.")
		return len(errors) == 0, errors
	}

	if cp.DatabaseName == "" {
		errors = append(errors, "Database name must be provided.")
	}

	if provider == PostgreSQL && cp.Host == "" {
		errors = append(errors, "Database host must be provided for PostgreSQL databases.")
	}

	if provider == PostgreSQL && cp.Port == 0 {
		errors = append(errors, "Database port must be provided for PostgreSQL databases.")
	}

	if provider == PostgreSQL && (cp.Port <= 0 || cp.Port > 65535) {
		errors = append(errors, "Database port must be a valid integer between 1 and 65535.")
	}

	return len(errors) == 0, errors
}

// GetConnectionString builds the connection string based on the provider
func (cp *ConnectionParameters) GetConnectionString() (string, error) {
	if cp.RawConnectionString != "" {
		return cp.RawConnectionString, nil
	}

	provider := cp.GetDbProvider()
	switch provider {
	case PostgreSQL:
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			cp.Host, cp.Port, cp.DatabaseName, cp.User, cp.Password), nil
	case SQLite:
		return cp.DatabaseName, nil
	default:
		return "", errors.New("unknown database provider")
	}
}

// ParseConnectionParameters creates ConnectionParameters from command line arguments
func ParseConnectionParameters(provider, connectionString, host, portStr, database, user, password string) (*ConnectionParameters, error) {
	cp := &ConnectionParameters{
		Provider:            provider,
		RawConnectionString: connectionString,
		Host:               host,
		DatabaseName:       database,
		User:               user,
		Password:           password,
	}

	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", portStr)
		}
		cp.Port = port
	}

	return cp, nil
}
