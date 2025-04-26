# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Testing Commands
- Build/run: `go run main.go` (in the git-multi-repo-clone directory)
- Build executable: `go build -o git-multi-repo-clone` (in the git-multi-repo-clone directory)
- Install dependencies: `go mod download`
- Format code: `go fmt ./...`
- Lint code: `golint ./...`
- Run tests: `go test ./...`

## Code Style Guidelines
- Follow standard Go naming conventions (CamelCase for exported, camelCase for internal)
- Use meaningful variable names that reflect purpose
- Group imports: standard library first, then third-party packages
- Add clear error messages with context (use `fmt.Errorf("context: %w", err)`)
- Document all exported functions, types, and packages with comments
- Handle errors explicitly, avoid naked returns
- Use early returns for error conditions to reduce nesting
- Use constants for magic values
- Keep functions small and focused on a single responsibility