# Git Multi-Repo Clone

A Go application to automatically clone/mirror all repositories from a Gitea server.

## Features

- Connect to a Gitea server using API token or basic authentication
- List all available repositories
- Clone repositories as mirrors to a specified target directory
- Update repositories if they already exist
- Filter repositories using include/exclude lists

## Setup

1. Copy the example configuration file:
   ```
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

```
go run main.go
```

This will:
1. Connect to your Gitea server
2. Retrieve the list of repositories
3. Clone each repository as a mirror to the target directory
4. Update any repositories that already exist

## Dependencies

- Go 1.15 or later
- `gopkg.in/yaml.v3` package

Install dependencies:
```
go mod init git-multi-repo-clone
go get gopkg.in/yaml.v3
```