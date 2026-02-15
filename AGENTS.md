# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Repository Overview

This repository contains a collection of command-line tools written in Go, following a monorepo structure. Each tool is self-contained with its own dependencies and build system.

## Repository Structure

```
adeotek-tools/
├── git-repos-backup/    # Git repository backup tool
│   ├── cmd/             # CLI entry point
│   ├── internal/        # Private application code
│   ├── pkg/             # Public/reusable packages
│   └── tests/           # Integration tests
├── sql-toolbox/         # SQL migration and backup tool
│   ├── cmd/             # CLI entry point
│   ├── internal/        # Private application code
│   └── integration-test.sh
└── scripts/             # Repository-wide scripts
```

## Build, Test & Lint Commands

### General Commands

Both tools follow similar Makefile patterns. Navigate to the tool directory first:

```bash
cd git-repos-backup/   # or cd sql-toolbox/
```

### Common Make Targets

```bash
# Build for current platform
make build

# Build for all platforms (Linux, Windows, macOS)
make build-all

# Run all tests
make test
# OR directly with go
go test -v ./...

# Run a single test
go test -v ./path/to/package -run TestName
# Example: go test -v ./internal/config -run TestLoad

# Run tests with coverage
go test -cover ./...

# Format code
make fmt
# OR directly
go fmt ./...

# Lint code (vet)
make lint
# OR directly
go vet ./...

# Clean build artifacts
make clean
```

### Tool-Specific Commands

#### git-repos-backup

```bash
# Install locally
make install

# Run integration tests
make integration-test
# OR
RUN_INTEGRATION_TESTS=1 go test -v ./tests

# Run the application
make run
```

#### sql-toolbox

```bash
# Install locally
make install
# OR
go install ./cmd/sql-toolbox

# Download dependencies
make deps

# Tidy up go.mod and go.sum
make tidy

# Run integration tests (requires build first)
bash integration-test.sh
```

## Code Style Guidelines

### Import Organization

**CRITICAL**: Follow strict import grouping with blank lines between groups:

```go
import (
	// Standard library packages first
	"fmt"
	"os"
	"path/filepath"

	// Third-party packages next
	"gopkg.in/yaml.v3"
	"github.com/lib/pq"

	// Internal packages last
	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/config"
	"github.com/adeotek/adeotek-tools/git-repos-backup/pkg/filter"
)
```

### Formatting

- Use Go standard formatting (`go fmt`)
- **2-space indentation** (defined in .editorconfig)
- UTF-8 encoding with LF line endings
- Trim trailing whitespace
- Insert final newline

### Types and Structs

- All exported types MUST have package comments
- Use structs for configurations and models
- Add struct tags for serialization (yaml, json)
- Example:

```go
// ProviderConfig contains configuration for a git provider
type ProviderConfig struct {
	Type              ProviderType `yaml:"type"`
	ServerURL         string       `yaml:"server_url"`
	AccessToken       string       `yaml:"access_token"`
}
```

### Naming Conventions

- **Exported identifiers**: PascalCase (e.g., `ProviderConfig`, `GetRepositories`)
- **Unexported identifiers**: camelCase (e.g., `providerType`, `getGiteaRepositories`)
- **Constants**: PascalCase with const blocks (e.g., `ProviderGitea`, `Version`)
- **Interfaces**: Noun or adjective ending in -er (e.g., `Repository`, `Commander`)

### Error Handling

**ALWAYS follow these patterns:**

1. **Check errors immediately** after function calls
2. **Use descriptive error messages** with context
3. **Wrap errors** using `fmt.Errorf` with `%w` verb for error chains

```go
// Good error handling
data, err := os.ReadFile(filename)
if err != nil {
	return nil, fmt.Errorf("failed to read config file: %w", err)
}

// Bad - missing context
data, err := os.ReadFile(filename)
if err != nil {
	return nil, err
}
```

### Function and Method Comments

All exported functions and methods MUST have comments:

```go
// Load loads configuration from the specified YAML file
func Load(filename string) (*Config, error) {
	// implementation
}

// NewMigrationService creates a new migration service instance
func NewMigrationService(isDryRun, verbose bool) *MigrationService {
	// implementation
}
```

### Package Comments

Every package MUST have a package comment:

```go
// Package config provides configuration functionality for the git-repos-backup tool
package config
```

## Testing Guidelines

### Test File Organization

- Test files are named `*_test.go`
- Place test files in the same package as the code being tested
- Integration tests go in separate directories (`tests/` or `integration-test.sh`)

### Test Structure

Use table-driven tests for multiple scenarios:

```go
func TestFilterRepositories(t *testing.T) {
	// Setup test data
	testRepos := []repository.Repository{
		{Id: 1, Name: "repo1", FullName: "owner1/repo1"},
		{Id: 2, Name: "repo2", FullName: "owner1/repo2"},
	}

	// Test cases
	tests := []struct {
		name        string
		include     []string
		exclude     []string
		expectedIDs []int
	}{
		{
			name:        "No filters",
			include:     []string{},
			exclude:     []string{},
			expectedIDs: []int{1, 2},
		},
		// ... more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test implementation
		})
	}
}
```

### Mocking Dependencies

- Create mock implementations for external dependencies
- Use dependency injection for testability
- Example: Replace `exec.Command` with a variable that can be mocked

### Integration Tests

- Separate integration tests from unit tests
- Use environment variables or flags to conditionally run them
- Example: `RUN_INTEGRATION_TESTS=1 go test -v ./tests`

## Architecture Guidelines

### Directory Structure

Follow Go project layout conventions:

- **`cmd/`**: Main applications (entry points)
- **`internal/`**: Private application code (cannot be imported by other projects)
  - `app/`: Application logic and orchestration
  - `config/`: Configuration handling
  - `models/`: Data models and structures
  - `repository/`: Data access layer
  - `services/`: Business logic services
  - `database/`: Database connection and operations
- **`pkg/`**: Public libraries (can be imported by other projects)
- **`tests/`**: Integration and end-to-end tests

### Dependency Direction

- `cmd/` depends on `internal/` and `pkg/`
- `internal/app/` orchestrates other `internal/` packages
- `internal/` packages should minimize dependencies on each other
- `pkg/` should have minimal external dependencies

## Version Tagging (Monorepo)

This is a monorepo, so version tags MUST be prefixed with the tool name:

```bash
# Format: toolname/vMAJOR.MINOR.PATCH
git tag git-repos-backup/v0.1.3
git tag sql-toolbox/v0.7.1

# Push tags
git push origin git-repos-backup/v0.1.3
git push origin sql-toolbox/v0.7.1
```

## Common Patterns

### Configuration Loading

Support both YAML config files and command-line arguments:

```go
// From file
cfg, err := config.Load(configPath)

// From CLI args
cfg := config.CreateFromArgs(arg1, arg2, ...)
```

### Verbose Logging

Always support a `verbose` flag for debugging:

```go
if verbose {
	fmt.Printf("----> Processing repository: %s\n", repo.FullName)
}
```

### Dry-Run Mode

For operations that modify state, support dry-run mode:

```go
if !isDryRun {
	// Perform actual operation
	err := performOperation()
} else {
	fmt.Printf("Dry run: would perform operation\n")
}
```

## Do's and Don'ts

### Do's ✅

- Use descriptive variable and function names
- Write tests for all public functions
- Document all exported types and functions
- Follow the existing code structure and patterns
- Use standard library whenever possible
- Check errors immediately and provide context
- Keep functions focused and single-purpose

### Don'ts ❌

- Don't ignore errors or use `_` to discard them
- Don't use panic for regular error handling
- Don't mix import groups (maintain proper spacing)
- Don't create exported functions without documentation
- Don't hardcode values that should be configurable
- Don't write deeply nested code (prefer early returns)

## Before Committing

Always run these commands before committing:

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Run tests
go test ./...

# Verify build succeeds
make build
```
