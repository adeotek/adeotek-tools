package services

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/database"
	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// SqlScriptsHelpers provides utilities for SQL script operations
type SqlScriptsHelpers struct{}

// NewSqlScriptsHelpers creates a new SqlScriptsHelpers instance
func NewSqlScriptsHelpers() *SqlScriptsHelpers {
	return &SqlScriptsHelpers{}
}

// ScanForSqlFiles scans a directory for SQL files and returns them in breadth-first order
// Execution order:
// 1. Files in the target directory (ignoring subdirectories), ordered alphabetically
// 2. For each subdirectory in alphabetical order:
//   - Files directly under it, ordered alphabetically
//   - Then recursively process its subdirectories in the same manner
func (s *SqlScriptsHelpers) ScanForSqlFiles(directory string) ([]string, error) {
	if directory == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory '%s' does not exist", directory)
	}

	var scripts []string
	err := s.scanDirectoryBreadthFirst(directory, "", &scripts)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory '%s': %w", directory, err)
	}

	return scripts, nil
}

// scanDirectoryBreadthFirst performs a breadth-first scan of a directory
// relativeDir is the relative path from the base directory (empty for root)
func (s *SqlScriptsHelpers) scanDirectoryBreadthFirst(baseDir, relativeDir string, scripts *[]string) error {
	currentDir := filepath.Join(baseDir, relativeDir)

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	// Separate files and directories
	var files []string
	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			files = append(files, entry.Name())
		}
	}

	// Sort files and directories alphabetically
	sort.Strings(files)
	sort.Strings(dirs)

	// First, add all SQL files in the current directory
	for _, file := range files {
		var relativePath string
		if relativeDir == "" {
			relativePath = file
		} else {
			relativePath = filepath.Join(relativeDir, file)
		}
		// Convert Windows path separators to Unix-style for consistency
		relativePath = strings.ReplaceAll(relativePath, "\\", "/")
		*scripts = append(*scripts, relativePath)
	}

	// Then, recursively process each subdirectory in alphabetical order
	for _, dir := range dirs {
		var subRelativeDir string
		if relativeDir == "" {
			subRelativeDir = dir
		} else {
			subRelativeDir = filepath.Join(relativeDir, dir)
		}
		err := s.scanDirectoryBreadthFirst(baseDir, subRelativeDir, scripts)
		if err != nil {
			return err
		}
	}

	return nil
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

// ExecuteScript executes a SQL script against the database in a transaction
func (s *SqlScriptsHelpers) ExecuteScript(scriptContent string, connectionParams *models.ConnectionParameters) error {
	dbFactory := database.NewFactory(connectionParams)
	db, err := dbFactory.CreateConnection()
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	if err := s.ExecuteScriptTx(tx, scriptContent); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit script transaction: %w", err)
	}
	return nil
}

// ExecuteScriptTx executes all statements of a SQL script inside an existing transaction
func (s *SqlScriptsHelpers) ExecuteScriptTx(tx *sql.Tx, scriptContent string) error {
	// Split script into individual statements for better error handling
	statements := s.SplitSqlStatements(scriptContent)

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	return nil
}

// SplitSqlStatements splits a SQL script into individual statements.
// Semicolons inside single/double quotes, line/block comments and dollar-quoted
// bodies do not terminate a statement.
func (s *SqlScriptsHelpers) SplitSqlStatements(script string) []string {
	var result []string
	var current strings.Builder

	var dollarTag string
	inSingle, inDouble, inLineComment, inBlockComment := false, false, false, false

	for i := 0; i < len(script); {
		c := script[i]
		next := byte(0)
		if i+1 < len(script) {
			next = script[i+1]
		}

		switch {
		case inLineComment:
			current.WriteByte(c)
			if c == '\n' {
				inLineComment = false
			}
			i++
		case inBlockComment:
			current.WriteByte(c)
			if c == '*' && next == '/' {
				current.WriteByte(next)
				inBlockComment = false
				i += 2
			} else {
				i++
			}
		case dollarTag != "":
			if strings.HasPrefix(script[i:], dollarTag) {
				current.WriteString(dollarTag)
				i += len(dollarTag)
				dollarTag = ""
			} else {
				current.WriteByte(c)
				i++
			}
		case inSingle:
			current.WriteByte(c)
			if c == '\'' {
				if next == '\'' {
					current.WriteByte(next)
					i += 2
					continue
				}
				inSingle = false
			}
			i++
		case inDouble:
			current.WriteByte(c)
			if c == '"' {
				if next == '"' {
					current.WriteByte(next)
					i += 2
					continue
				}
				inDouble = false
			}
			i++
		default:
			switch {
			case c == '\'':
				inSingle = true
				current.WriteByte(c)
				i++
			case c == '"':
				inDouble = true
				current.WriteByte(c)
				i++
			case c == '-' && next == '-':
				inLineComment = true
				current.WriteString("--")
				i += 2
			case c == '/' && next == '*':
				inBlockComment = true
				current.WriteString("/*")
				i += 2
			case c == '$':
				if tagLen := dollarTagLength(script[i:]); tagLen > 0 {
					dollarTag = script[i : i+tagLen]
					current.WriteString(dollarTag)
					i += tagLen
				} else {
					current.WriteByte(c)
					i++
				}
			case c == ';':
				current.WriteByte(c)
				if stmt := strings.TrimSpace(current.String()); hasExecutableSQL(stmt) {
					result = append(result, stmt)
				}
				current.Reset()
				i++
			default:
				current.WriteByte(c)
				i++
			}
		}
	}

	if stmt := strings.TrimSpace(current.String()); hasExecutableSQL(stmt) {
		result = append(result, stmt)
	}

	return result
}

// dollarTagLength returns the length of a dollar-quote opener at the start of s,
// or 0 when s does not start with one (e.g. "$1" in an identifier).
func dollarTagLength(s string) int {
	if len(s) < 2 || s[0] != '$' {
		return 0
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == '$' {
			return i + 1
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return 0
		}
	}
	return 0
}

// hasExecutableSQL reports whether a statement contains anything besides
// whitespace, semicolons and leading comments.
func hasExecutableSQL(statement string) bool {
	s := strings.TrimSpace(statement)
	for {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "--"):
			if idx := strings.IndexByte(s, '\n'); idx >= 0 {
				s = s[idx+1:]
				continue
			}
			return false
		case strings.HasPrefix(s, "/*"):
			if idx := strings.Index(s, "*/"); idx >= 0 {
				s = s[idx+2:]
				continue
			}
			return false
		}
		break
	}
	s = strings.TrimSpace(s)
	return s != "" && s != ";"
}
