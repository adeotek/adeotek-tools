package models

import (
	"errors"
	"fmt"
	"net/url"
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
	Provider            string
	RawConnectionString string
	Host                string
	Port                int
	DatabaseName        string
	User                string
	Password            string
	SSHTunnel           *SSHTunnelConfig
	activeTunnel        interface{} // Holds *tunnel.SSHTunnel when active
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

// normalizeConnectionString converts various connection string formats to lib/pq format
func normalizeConnectionString(connStr string) (string, error) {
	connStr = strings.TrimSpace(connStr)

	// Check if it's a PostgreSQL URL format (postgresql://... or postgres://...)
	if strings.HasPrefix(connStr, "postgresql://") || strings.HasPrefix(connStr, "postgres://") {
		return convertPostgreSQLURL(connStr)
	}

	// Check if it's .NET format (contains semicolons)
	if strings.Contains(connStr, ";") {
		return convertDotNetFormat(connStr)
	}

	// Assume it's already in lib/pq format (space-separated key=value pairs)
	return connStr, nil
}

// convertPostgreSQLURL converts PostgreSQL URL format to lib/pq format
// Example: postgresql://user:password@host:5432/dbname?sslmode=disable
func convertPostgreSQLURL(connStr string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", fmt.Errorf("invalid PostgreSQL URL: %w", err)
	}

	var parts []string

	// Host
	if u.Hostname() != "" {
		parts = append(parts, fmt.Sprintf("host=%s", u.Hostname()))
	}

	// Port
	if u.Port() != "" {
		parts = append(parts, fmt.Sprintf("port=%s", u.Port()))
	} else {
		parts = append(parts, "port=5432") // Default PostgreSQL port
	}

	// Database name (from path)
	dbname := strings.TrimPrefix(u.Path, "/")
	if dbname != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", dbname))
	}

	// User
	if u.User != nil && u.User.Username() != "" {
		parts = append(parts, fmt.Sprintf("user=%s", u.User.Username()))
	}

	// Password
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			parts = append(parts, fmt.Sprintf("password=%s", password))
		}
	}

	// Query parameters (e.g., sslmode)
	for key, values := range u.Query() {
		if len(values) > 0 {
			parts = append(parts, fmt.Sprintf("%s=%s", key, values[0]))
		}
	}

	// Add sslmode=disable if not specified
	if !strings.Contains(connStr, "sslmode") {
		parts = append(parts, "sslmode=disable")
	}

	return strings.Join(parts, " "), nil
}

// convertDotNetFormat converts .NET connection string format to lib/pq format
// Example: Server=host;Port=5432;Database=mydb;Username=user;Password=pass;SslMode=Disable;
func convertDotNetFormat(connStr string) (string, error) {
	// Parse .NET format (semicolon-separated key=value pairs)
	dotNetParams := make(map[string]string)
	pairs := strings.Split(connStr, ";")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		dotNetParams[strings.ToLower(key)] = value
	}

	// Map .NET keys to lib/pq keys
	var parts []string

	if val, ok := dotNetParams["server"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("host=%s", val))
	} else if val, ok := dotNetParams["host"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("host=%s", val))
	}

	if val, ok := dotNetParams["port"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("port=%s", val))
	}

	if val, ok := dotNetParams["database"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", val))
	}

	if val, ok := dotNetParams["username"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("user=%s", val))
	} else if val, ok := dotNetParams["user id"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("user=%s", val))
	} else if val, ok := dotNetParams["user"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("user=%s", val))
	}

	if val, ok := dotNetParams["password"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("password=%s", val))
	}

	// Handle SSL mode
	if val, ok := dotNetParams["sslmode"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", strings.ToLower(val)))
	} else {
		parts = append(parts, "sslmode=disable")
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("no valid connection parameters found in .NET format string")
	}

	return strings.Join(parts, " "), nil
}

// GetConnectionString builds the connection string based on the provider
func (cp *ConnectionParameters) GetConnectionString() (string, error) {
	if cp.RawConnectionString != "" {
		// Normalize connection string to lib/pq format
		normalized, err := normalizeConnectionString(cp.RawConnectionString)
		if err != nil {
			return "", fmt.Errorf("failed to normalize connection string: %w", err)
		}
		return normalized, nil
	}

	provider := cp.GetDbProvider()
	switch provider {
	case PostgreSQL:
		host := cp.Host
		port := cp.Port

		// Use tunnel's local address if active
		if cp.activeTunnel != nil {
			// Type assertion to access GetLocalPort() method
			type localPortGetter interface {
				GetLocalPort() int
			}
			if tunnel, ok := cp.activeTunnel.(localPortGetter); ok {
				host = "127.0.0.1"
				port = tunnel.GetLocalPort()
			}
		}

		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			host, port, cp.DatabaseName, cp.User, cp.Password), nil
	case SQLite:
		return cp.DatabaseName, nil
	default:
		return "", errors.New("unknown database provider")
	}
}

// SetActiveTunnel stores the active SSH tunnel reference
func (cp *ConnectionParameters) SetActiveTunnel(tunnel interface{}) {
	cp.activeTunnel = tunnel
}

// GetActiveTunnel retrieves the active SSH tunnel reference
func (cp *ConnectionParameters) GetActiveTunnel() interface{} {
	return cp.activeTunnel
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
