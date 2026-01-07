package services

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/adeotek/adeotek-tools/sql-migration/internal/database"
	"github.com/adeotek/adeotek-tools/sql-migration/internal/models"
)

// SqlScriptsHelpers provides utilities for SQL script operations
type SqlScriptsHelpers struct{}

// NewSqlScriptsHelpers creates a new SqlScriptsHelpers instance
func NewSqlScriptsHelpers() *SqlScriptsHelpers {
	return &SqlScriptsHelpers{}
}

// ScanForSqlFiles scans a directory for SQL files and returns them in a specific order
// Supports multiple directory levels (e.g., tables/level1/level2/level3/script.sql)
// All subdirectories at any level are processed in alphabetical order
func (s *SqlScriptsHelpers) ScanForSqlFiles(directory string) ([]string, error) {
	if directory == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory '%s' does not exist", directory)
	}

	var scripts []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sql") {
			relativePath, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}
			// Convert Windows path separators to Unix-style for consistency
			relativePath = strings.ReplaceAll(relativePath, "\\", "/")
			scripts = append(scripts, relativePath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory '%s': %w", directory, err)
	}

	// Sort scripts by their full path, which ensures alphabetical order at all directory levels
	sort.Strings(scripts)

	return scripts, nil
}

// CalculateHash calculates the SHA256 hash of a file
func (s *SqlScriptsHelpers) CalculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash for file '%s': %w", filePath, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ExecuteScript executes a SQL script against the database
func (s *SqlScriptsHelpers) ExecuteScript(scriptContent string, connectionParams *models.ConnectionParameters) error {
	dbFactory := database.NewFactory(connectionParams)
	db, err := dbFactory.CreateConnection()
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	// Split script into individual statements for better error handling
	statements := s.SplitSqlStatements(scriptContent)

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		_, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	return nil
}

// SplitSqlStatements splits a SQL script into individual statements
// This handles complex scripts including stored procedures, functions, and other constructs
func (s *SqlScriptsHelpers) SplitSqlStatements(script string) []string {
	var result []string
	var current strings.Builder

	lines := strings.Split(script, "\n")
	inBlock := false
	blockDelimiter := ""

	// Regex pattern for PostgreSQL dollar-quoted strings: $tag$ where tag can be empty or contain letters, digits, underscores
	dollarQuotePattern := regexp.MustCompile(`\$([A-Za-z0-9_]*)\$`)

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments when not in a block
		if !inBlock && (trimmedLine == "" || strings.HasPrefix(trimmedLine, "--")) {
			continue
		}

		// Check for block delimiters ($$, $tag$, etc.)
		if !inBlock {
			// Look for PostgreSQL dollar-quoted strings or PL/pgSQL blocks
			matches := dollarQuotePattern.FindAllString(trimmedLine, -1)
			if len(matches) > 0 {
				// Use the first valid delimiter found
				blockDelimiter = matches[0]
				inBlock = true
			}
		}

		// Add the line to current statement
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)

		// Check if we're ending a block
		if inBlock && blockDelimiter != "" && strings.Contains(trimmedLine, blockDelimiter) {
			// Count occurrences to handle multiple delimiters in same line
			openCount := strings.Count(current.String(), blockDelimiter)
			if openCount%2 == 0 { // Even number means block is closed
				inBlock = false
				blockDelimiter = ""
			}
		}

		// If not in a block and line ends with semicolon, it's end of statement
		if !inBlock && strings.HasSuffix(trimmedLine, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && stmt != ";" {
				result = append(result, stmt)
			}
			current.Reset()
		}
	}

	// Add any remaining content as final statement
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" && stmt != ";" {
			result = append(result, stmt)
		}
	}

	return result
}
