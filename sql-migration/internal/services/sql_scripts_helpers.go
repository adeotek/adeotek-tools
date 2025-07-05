package services

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
func (s *SqlScriptsHelpers) ScanForSqlFiles(directory string) ([]string, error) {
	if directory == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory '%s' does not exist", directory)
	}

  scripts := make(map[string][]string)
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
      // Extract first folder as prefix
      prefix := strings.SplitN(relativePath, "/", 2)[0]
      if _, exists := scripts[prefix]; !exists {
        scripts[prefix] = []string{}
      }
      scripts[prefix] = append(scripts[prefix], relativePath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory '%s': %w", directory, err)
	}

  return sortScripts(scripts), nil
}

// sortScripts sorts scripts by prefix (folder)
func sortScripts(scripts map[string][]string) []string {
	var sorted = []string{}

  // Get all keys and sort them
  keys := make([]string, 0, len(scripts))
  for key := range scripts {
    keys = append(keys, key)
  }
  sort.Strings(keys)

  // Iterate over sorted keys and append scripts
  for _, key := range keys {
    if scripts[key] != nil {
      sort.Strings(scripts[key])
      sorted = append(sorted, scripts[key]...)
    }
  }

	return sorted
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
	statements := s.splitSqlStatements(scriptContent)

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

// splitSqlStatements splits a SQL script into individual statements
func (s *SqlScriptsHelpers) splitSqlStatements(script string) []string {
	// Simple statement splitter - splits on semicolon followed by newline
	// This is a basic implementation and might need enhancement for complex scripts
	statements := strings.Split(script, ";")
	var result []string

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}
