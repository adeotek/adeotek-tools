# Git Repos Backup

A cross-platform CLI tool to backup multiple Git repositories from Gitea and/or GitHub servers.

> **Note:** This tool is part of the [AdeoTEK Tools](https://github.com/adeotek/adeotek-tools) mono-repo.

## Features

- Support for multiple Git providers (Gitea and GitHub)
- Mirror-based backup of repositories (bare repositories)
- Filtering repositories via include/exclude lists
- Authentication via tokens or basic auth
- SSL verification skip option for self-signed certificates
- Separate target directories for each provider

## Docker

You can also run git-repos-backup using Docker:

### Using the pre-built image

```bash
# Pull the image
docker pull adeotek/git-repos-backup:latest

# Run with a configuration file (mount it as a volume)
docker run -v $(pwd)/config.yml:/app/config.yml -v /path/to/backups:/backups adeotek/git-repos-backup:latest

# Run with command-line arguments
docker run -v /path/to/backups:/backups adeotek/git-repos-backup:latest -provider github -token your_github_token -target-dir /backups
```

### Building the Docker image locally

```bash
# Navigate to the project directory
cd git-repos-backup

# Build the Docker image
docker build -t git-repos-backup .

# Run the container
docker run -v $(pwd)/config.yml:/app/config.yml -v /path/to/backups:/backups git-repos-backup
```

### Docker environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GB_BACKUP_INTERVAL` | Seconds between backup runs (for scheduled backups) | `86400` (24 hours) |
| `GB_VERBOSE` | Enable verbose output | `false` |
| `GB_CONFIG` | Path to config file (if using config file mode) | `/app/config.yaml` |
| `GB_PROVIDER` | Provider type (gitea or github) | - |
| `GB_TARGET_DIR` | Directory to clone repositories into | - |
| `GB_SERVER_URL` | URL of the Git server (required for Gitea, optional for GitHub) | - |
| `GB_TOKEN` | API token for authentication | - |
| `GB_USERNAME` | Username for basic authentication | - |
| `GB_PASSWORD` | Password for basic authentication | - |
| `GB_USE_BASIC_AUTH` | Whether to use basic authentication | `false` |
| `GB_SKIP_SSL_VALIDATION` | Whether to skip SSL validation | `false` |
| `GB_INCLUDE_REPOS` | Comma-separated list of repository full names to include | - |
| `GB_EXCLUDE_REPOS` | Comma-separated list of repository full names to exclude | - |

The Docker image can operate in two modes:

1. **Config file mode (default)**: Loads configuration from a mounted config file
   ```bash
   docker run -v $(pwd)/config.yml:/app/config.yml -v /path/to/backups:/backups adeotek/git-repos-backup:latest
   ```

2. **Command-line mode**: Uses environment variables directly
   ```bash
   docker run -v /path/to/backups:/backups \
     -e GB_PROVIDER=github \
     -e GB_TOKEN=your_github_token \
     -e GB_TARGET_DIR=/backups \
     adeotek/git-repos-backup:latest
   ```

## Project Structure

```
git-repos-backup/
├── cmd/                    # Command-line interface
│   └── git-repos-backup/   # Main entry point
├── internal/               # Internal packages (not exported)
│   ├── app/                # Application logic
│   ├── config/             # Configuration handling
│   ├── git/                # Git operations
│   └── repository/         # Git provider API interactions
├── pkg/                    # Public packages (can be imported)
│   └── filter/             # Repository filtering functionality
└── tests/                  # Integration tests
```

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/adeotek/adeotek-tools.git
cd adeotek-tools/git-repos-backup

# Build the binary
go build -o git-repos-backup

# Install to a location in your PATH (optional)
cp git-repos-backup /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/adeotek/adeotek-tools/git-repos-backup/cmd/git-repos-backup@latest
```

## Build

The project includes a Makefile with several useful targets:

```bash
# Build for current platform
make build

# Build for all platforms (Linux, Windows, macOS)
make build-all

# Run tests
make test

# Run integration tests
make integration-test

# Clean build artifacts
make clean

# Format and lint code
make fmt lint
```

## Usage

You can use git-repos-backup in two different ways:

### 1. Using a configuration file

1. Create a configuration file based on the example:
   ```bash
   cp config.yaml.example config.yaml
   # Edit the config.yaml file with your settings
   ```

2. Run the command:
   ```bash
   ./git-repos-backup -config /path/to/config.yaml
   ```

### 2. Using command-line arguments

You can also run the tool directly with command-line arguments:

```bash
./git-repos-backup -provider github -token your_github_token -target-dir /path/to/backups
```

### Command-line Options

```
git-repos-backup [flags]

Flags:
  -config string
        Path to configuration file (if not specified, defaults to config.yaml in current directory if it exists)
  -provider string
        Provider type (gitea or github)
  -server-url string
        URL of the Git server (required for Gitea, optional for GitHub)
  -token string
        API token for authentication
  -username string
        Username for basic authentication
  -password string
        Password for basic authentication
  -use-basic-auth
        Whether to use basic authentication
  -skip-ssl
        Whether to skip SSL validation
  -include string
        Comma-separated list of repository full names to include
  -exclude string
        Comma-separated list of repository full names to exclude
  -target-dir string
        Directory to clone repositories into
  -help
        Show help message and exit
  -verbose
        Show all messages
  -version
        Show version information and exit
```

#### Examples

1. Using config file:
   ```bash
   ./git-repos-backup -config /path/to/config.yaml -verbose
   ```

2. Using command-line arguments for GitHub:
   ```bash
   ./git-repos-backup -provider github -token your_github_token -target-dir /path/to/github/backups -verbose
   ```

3. Using command-line arguments for Gitea:
   ```bash
   ./git-repos-backup -provider gitea -server-url https://gitea.example.com -token your_gitea_token -target-dir /path/to/gitea/backups -verbose
   ```

4. With repository filtering:
   ```bash
   ./git-repos-backup -provider github -token your_github_token -target-dir /path/to/backups -include "owner/repo1,owner/repo2" -verbose
   ```

## Configuration

The configuration file uses YAML format:

```yaml
# Git Repos Backup Configuration

# Providers configuration
providers:
  # Gitea provider
  - type: gitea
    server_url: https://gitea.example.com
    # Authentication (use either token or username/password)
    access_token: your_gitea_access_token
    # username: your_username
    # password: your_password
    # use_basic_auth: false
    # Set to true to skip the SSL validation (e.g., when Gitea is using a self-signed certificate)
    # skip_ssl_validation: true
    # Optional repository filtering
    # include:
    #   - owner/repo1
    #   - owner/repo2
    # exclude:
    #   - owner/repo3
    # Target directory for repositories backup
    target_dir: /path/to/gitea/backups
    
  # GitHub provider
  - type: github
    # For GitHub Enterprise, specify the server URL
    # server_url: https://github.example.com
    # GitHub authentication (use personal access token)
    access_token: your_github_access_token
    # Optional repository filtering
    # include:
    #   - owner/repo1
    #   - owner/repo2
    # exclude:
    #   - owner/repo3
    # Target directory for repositories backup
    target_dir: /path/to/github/backups
```

### Provider Configuration

Each provider configuration requires:
- `type`: Provider type (either `gitea` or `github`)
- `server_url`: URL of the Git server (required for Gitea, optional for GitHub - only needed for GitHub Enterprise)
- `target_dir`: Directory where repositories will be backed up

Authentication options:
- `access_token`: API token for authentication (recommended)
- `username` and `password`: For basic authentication
- `use_basic_auth`: Set to `true` to use basic authentication instead of token

Additional options:
- `skip_ssl_validation`: Set to `true` to skip SSL certificate validation (useful for self-signed certificates)
- `include`: List of repository full names to include (optional)
- `exclude`: List of repository full names to exclude (optional, ignored if include is specified)

## Repository Structure

Repositories are backed up following this structure:
```
target_dir/
├── owner1/                # Repository owner's login
│   ├── repo1/             # Repository name (bare repository)
│   └── repo2/
└── owner2/
    └── repo3/
```

## Development

### Prerequisites

- Go 1.21 or later
- Git

### Setup Development Environment

1. Clone the repository:
```bash
git clone https://github.com/adeotek/git-repos-backup.git
cd git-repos-backup
```

2. Install dependencies:
```bash
go mod download
```

3. Run tests:
```bash
go test ./...
```

### Testing

The project includes both unit tests and integration tests:

- **Unit tests**: Test individual components in isolation
  ```bash
  go test ./...
  ```

- **Integration tests**: Test components working together
  ```bash
  RUN_INTEGRATION_TESTS=1 go test ./tests
  ```

- **Test with coverage**:
  ```bash
  go test -cover ./...
  ```

### Code Quality

Before submitting contributions, ensure the code is formatted and linted:

```bash
# Format the code
go fmt ./...

# Lint the code
go vet ./...
```

## License

MIT License
