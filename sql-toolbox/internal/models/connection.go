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
	SSLMode             string
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
		parts = append(parts, fmt.Sprintf("host=%s", quoteConnValue(u.Hostname())))
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
		parts = append(parts, fmt.Sprintf("dbname=%s", quoteConnValue(dbname)))
	}

	// User
	if u.User != nil && u.User.Username() != "" {
		parts = append(parts, fmt.Sprintf("user=%s", quoteConnValue(u.User.Username())))
	}

	// Password
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			parts = append(parts, fmt.Sprintf("password=%s", quoteConnValue(password)))
		}
	}

	// Query parameters (e.g., sslmode)
	query := u.Query()
	for key, values := range query {
		if len(values) > 0 {
			parts = append(parts, fmt.Sprintf("%s=%s", key, quoteConnValue(values[0])))
		}
	}

	// Add sslmode=disable if not specified
	if _, ok := query["sslmode"]; !ok {
		parts = append(parts, "sslmode=disable")
	}

	return strings.Join(parts, " "), nil
}

// convertDotNetFormat converts .NET connection string format to lib/pq format
// Example: Server=host;Port=5432;Database=mydb;Username=user;Password=pass;SslMode=Disable;
func convertDotNetFormat(connStr string) (string, error) {
	// Parse .NET format (semicolon-separated key=value pairs)
	dotNetParams := make(map[string]string)
	pairs := splitDotNetPairs(connStr)

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
		// .NET connection strings may quote values containing delimiters
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		dotNetParams[strings.ToLower(key)] = value
	}

	// Map .NET keys to lib/pq keys
	var parts []string

	if val, ok := dotNetParams["server"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("host=%s", quoteConnValue(val)))
	} else if val, ok := dotNetParams["host"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("host=%s", quoteConnValue(val)))
	}

	if val, ok := dotNetParams["port"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("port=%s", quoteConnValue(val)))
	}

	if val, ok := dotNetParams["database"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", quoteConnValue(val)))
	}

	if val, ok := dotNetParams["username"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("user=%s", quoteConnValue(val)))
	} else if val, ok := dotNetParams["user id"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("user=%s", quoteConnValue(val)))
	} else if val, ok := dotNetParams["user"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("user=%s", quoteConnValue(val)))
	}

	if val, ok := dotNetParams["password"]; ok && val != "" {
		parts = append(parts, fmt.Sprintf("password=%s", quoteConnValue(val)))
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

// quoteConnValue quotes a connection string value using lib/pq's key=value syntax.
// Values containing whitespace, quotes or backslashes are wrapped in single quotes
// with backslash escaping so passwords with special characters keep working.
func quoteConnValue(value string) string {
	if value == "" {
		return "''"
	}
	if !strings.ContainsAny(value, " '\\") {
		return value
	}
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "'" + escaped + "'"
}

type connInfoPair struct {
	key   string
	value string
}

// parseConnInfo parses a lib/pq key=value connection string.
func parseConnInfo(connStr string) ([]connInfoPair, error) {
	var pairs []connInfoPair
	i := 0
	for {
		for i < len(connStr) && (connStr[i] == ' ' || connStr[i] == '\t' || connStr[i] == '\n') {
			i++
		}
		if i >= len(connStr) {
			return pairs, nil
		}

		keyStart := i
		for i < len(connStr) && connStr[i] != '=' && connStr[i] != ' ' {
			i++
		}
		if i >= len(connStr) || connStr[i] != '=' {
			return nil, fmt.Errorf("invalid connection parameter near %q", connStr[keyStart:])
		}
		key := connStr[keyStart:i]
		i++ // skip '='

		var value strings.Builder
		if i < len(connStr) && connStr[i] == '\'' {
			i++
			closed := false
			for i < len(connStr) {
				switch connStr[i] {
				case '\\':
					if i+1 < len(connStr) {
						value.WriteByte(connStr[i+1])
						i += 2
						continue
					}
					value.WriteByte(connStr[i])
					i++
				case '\'':
					i++
					closed = true
				default:
					value.WriteByte(connStr[i])
					i++
				}
				if closed {
					break
				}
			}
			if !closed {
				return nil, fmt.Errorf("unterminated quoted value for %q", key)
			}
		} else {
			for i < len(connStr) && connStr[i] != ' ' && connStr[i] != '\t' && connStr[i] != '\n' {
				if connStr[i] == '\\' && i+1 < len(connStr) {
					value.WriteByte(connStr[i+1])
					i += 2
					continue
				}
				value.WriteByte(connStr[i])
				i++
			}
		}
		pairs = append(pairs, connInfoPair{key: key, value: value.String()})
	}
}

func buildConnInfo(pairs []connInfoPair) string {
	parts := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		parts = append(parts, pair.key+"="+quoteConnValue(pair.value))
	}
	return strings.Join(parts, " ")
}

// RemovePasswordFromConnInfo returns a connection string without its password and
// the extracted password (empty when absent), so callers can pass the secret via
// PGPASSWORD instead of exposing it in process arguments.
func RemovePasswordFromConnInfo(connInfo string) (string, string) {
	pairs, err := parseConnInfo(connInfo)
	if err != nil {
		return connInfo, ""
	}
	password := ""
	clean := make([]connInfoPair, 0, len(pairs))
	for _, pair := range pairs {
		if strings.EqualFold(pair.key, "password") {
			password = pair.value
			continue
		}
		clean = append(clean, pair)
	}
	return buildConnInfo(clean), password
}

// RewriteConnInfoHostPort returns a copy of a lib/pq connection string pointing at
// host:port. It is used to route raw connection strings through an SSH tunnel.
func RewriteConnInfoHostPort(connInfo, host string, port int) string {
	pairs, err := parseConnInfo(connInfo)
	if err != nil {
		return connInfo
	}
	portValue := strconv.Itoa(port)
	hostSet, portSet := false, false
	for i := range pairs {
		switch strings.ToLower(pairs[i].key) {
		case "host":
			pairs[i].value = host
			hostSet = true
		case "port":
			pairs[i].value = portValue
			portSet = true
		}
	}
	if !hostSet {
		pairs = append(pairs, connInfoPair{key: "host", value: host})
	}
	if !portSet {
		pairs = append(pairs, connInfoPair{key: "port", value: portValue})
	}
	return buildConnInfo(pairs)
}

// splitDotNetPairs splits a .NET connection string on semicolons that are not
// inside single- or double-quoted values.
func splitDotNetPairs(connStr string) []string {
	var pairs []string
	var current strings.Builder
	var quote byte
	for i := 0; i < len(connStr); i++ {
		c := connStr[i]
		if quote != 0 {
			current.WriteByte(c)
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
			current.WriteByte(c)
		case ';':
			pairs = append(pairs, current.String())
			current.Reset()
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		pairs = append(pairs, current.String())
	}
	return pairs
}

// GetConnectionString builds the connection string based on the provider
func (cp *ConnectionParameters) GetConnectionString() (string, error) {
	host, port := cp.Host, cp.Port
	// Use tunnel's local address if active
	if localPort, ok := activeTunnelLocalPort(cp.activeTunnel); ok {
		host = "127.0.0.1"
		port = localPort
	}

	if cp.RawConnectionString != "" {
		// Normalize connection string to lib/pq format
		normalized, err := normalizeConnectionString(cp.RawConnectionString)
		if err != nil {
			return "", fmt.Errorf("failed to normalize connection string: %w", err)
		}
		// A raw connection string would otherwise bypass the SSH tunnel
		if cp.activeTunnel != nil {
			normalized = RewriteConnInfoHostPort(normalized, host, port)
		}
		return normalized, nil
	}

	provider := cp.GetDbProvider()
	switch provider {
	case PostgreSQL:
		sslMode := cp.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
			quoteConnValue(host), port, quoteConnValue(cp.DatabaseName),
			quoteConnValue(cp.User), quoteConnValue(cp.Password), sslMode), nil
	case SQLite:
		return cp.DatabaseName, nil
	default:
		return "", errors.New("unknown database provider")
	}
}

// activeTunnelLocalPort returns the tunnel's local port when the active tunnel exposes one.
func activeTunnelLocalPort(activeTunnel interface{}) (int, bool) {
	if activeTunnel == nil {
		return 0, false
	}
	type localPortGetter interface {
		GetLocalPort() int
	}
	tunnel, ok := activeTunnel.(localPortGetter)
	if !ok {
		return 0, false
	}
	port := tunnel.GetLocalPort()
	if port <= 0 {
		return 0, false
	}
	return port, true
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
		Host:                host,
		DatabaseName:        database,
		User:                user,
		Password:            password,
	}

	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", portStr)
		}
		cp.Port = port
	} else if ParseDatabaseProvider(provider) == PostgreSQL {
		cp.Port = 5432
	}

	return cp, nil
}
