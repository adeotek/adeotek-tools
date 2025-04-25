# Git Multi-Repo Clone

A cross-platform CLI tool to automatically clone/mirror all repositories from a Gitea server.

## Features

- Connect to a Gitea server using API token or basic authentication
- List all available repositories
- Clone repositories as mirrors to a specified target directory
- Update repositories if they already exist
- Filter repositories using include/exclude lists
- Cross-platform support (Linux, Windows, macOS)

## Installation

### From Source

#### Prerequisites
- Go 1.21 or later

#### Build

```bash
# Clone the repository
git clone https://github.com/adeotek/tools.git adeotek-tools
cd adeotek-tools/git-multi-repo-clone

# Build for your current platform
make build

# Or build for all supported platforms (Linux, Windows, macOS)
make build-all

# Alternatively, you can build using Go directly
go build -o git-multi-repo-clone
```

The binaries will be available in the `bin` directory after running `make build-all`:
- Linux: `bin/git-multi-repo-clone_[version]_linux_amd64`
- Windows: `bin/git-multi-repo-clone_[version]_windows_amd64.exe`
- macOS: `bin/git-multi-repo-clone_[version]_darwin_amd64`

### Using Go

```bash
??? go install github.com/adeotek/git-multi-repo-clone@latest
```

## Setup

1. Copy the example configuration file:
   ```bash
   cp config.yaml.example config.yaml
   ```

2. Edit the `config.yaml` file with your Gitea server information:
   ```yaml
   gitea_url: "https://gitea.example.com"
   api_token: "your_api_token"
   username: "your_username"
   password: "your_password"
   use_basic_auth: false
   target_dir: "/path/to/clone/target"
   ```

3. Choose authentication method:
   - For token authentication: provide an `api_token` and set `use_basic_auth` to `false`
   - For basic authentication: provide `username` and `password` and set `use_basic_auth` to `true`

4. Optionally configure repository filtering:
   - To only clone specific repositories, use the `include` list
   - To exclude specific repositories, use the `exclude` list
   - If both are specified, only the `include` list is used
   
   Example:
   ```yaml
   # Only clone these repositories
   include:
     - "important-repo"
     - "another-repo"
   
   # Exclude these repositories (only used if include is not specified)
   exclude:
     - "skip-this-repo"
     - "also-skip-this"
   ```

## Usage

```bash
# Using default config.yaml in current directory
git-multi-repo-clone

# Specify a custom config file
git-multi-repo-clone -config /path/to/config.yaml

# Show version information
git-multi-repo-clone -version

# Show help
git-multi-repo-clone -help
```

### Command Line Flags

- `-config`: Path to configuration file (default: "config.yaml")
- `-version`: Show version information and exit
- `-help`: Show help message and exit

This will:
1. Connect to your Gitea server
2. Retrieve the list of repositories
3. Clone each repository to the target directory
4. Update any repositories that already exist (if configured)

## Project Structure

```
.
├── cmd/                     # Command-line interfaces
│   └── git-multi-repo-clone/ # Main CLI package
│       └── main.go          # CLI entry point
├── internal/                # Private application and library code
│   ├── app/                 # Core application logic
│   ├── config/              # Configuration handling
│   ├── git/                 # Git operations
│   └── repository/          # Repository API handling
├── pkg/                     # Public library code
│   └── filter/              # Repository filtering
├── tests/                   # Integration tests
├── bin/                     # Output binaries
├── Makefile                 # Project build instructions
├── README.md                # Project documentation
├── config.yaml.example      # Example configuration
├── main.go                  # Convenience wrapper for building from root
└── go.mod                   # Go module definition
```

## Development

### Code Structure

The application is structured with a clear separation of concerns:
- `internal/app` contains the core application logic, extracted to avoid code duplication
- `cmd/git-multi-repo-clone/main.go` is the standard entry point following Go conventions
- `main.go` in the root directory is a convenience wrapper allowing builds from the project root
- Both entry points call the same `app.Run()` function, ensuring consistent behavior

### Dependencies

- Go 1.21 or later
- `gopkg.in/yaml.v3` package

### Install Dependencies

```bash
go mod download
```

### Run Unit Tests

```bash
make test
```

### Run Integration Tests

```bash
make integration-test
```

### Format Code

```bash
make fmt
```

### Lint Code

```bash
make lint
```

## Cross-Compilation

The project supports cross-compilation for Linux, Windows, and macOS. To build for all platforms:

```bash
make build-all
```

This will generate binaries in the `bin` directory with appropriate file extensions (.exe for Windows).
