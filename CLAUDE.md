# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository description

This repository contains a collection of tools written in GO, .NET, and TypeScript/JavaScript.

## Tools

### `git-repos-backup`

A tool to backup Git repositories from GitHub or Gitea to a local directory. It supports both public and private repositories, and can be configured to run periodically using a cron job.
The `git-repos-backup` tool is written in Go and provides a simple command-line interface to manage backups of Git repositories.

#### `git-repos-backup` Build/Lint/Test Commands
- Build: `make build`
- Build all platforms: `make build-all`
- Install: `make install`
- Format code: `make fmt`
- Lint code: `make lint`
- Run all tests: `make test` or `go test -v ./...`
- Run single test: `go test -v ./path/to/package -run TestName`
- Run integration tests: `make integration-test` or `RUN_INTEGRATION_TESTS=1 go test -v ./tests`

#### `git-repos-backup` Code Style Guidelines
- Imports: Standard library first, third-party next, grouped with blank lines
- Formatting: Go standard (go fmt), 2-space indentation
- Types: Exported types have comments, use structs for configs and models
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: Check immediately, descriptive messages, wrap errors with context
- Testing: Standard Go testing package, mock dependencies, separate integration tests
- Documentation: Package and exported function comments follow Go conventions
- Architecture: Follow Go conventions with cmd/, internal/, pkg/ directories

### `sql-migration`

A tool to manage and apply SQL database migrations (SQL scripts). It supports PostgreSQL and SQLite databases and provides a simple command-line interface for managing migrations.
The `sql-migration` tool is written in .NET and allows users to apply to databases migrations (in the form of a collection of SQL scripts).

#### `sql-migration` Code Style Guidelines
- .NET version: .NET 9.0 or later
- Formatting: standard .NET formatting (dotnet format)
- Error handling: prefere using return values instead of exceptions for expected errors, use exceptions for unexpected errors only when absolutely necessary
- Testing: Standard .NET testing package using xUnit, and NSubstitute for mocking
- Documentation: Package and exported function comments follow .NET conventions
- Build: the project should be built with AOT (Ahead of Time) compilation enabled, for both Linux x64 and Windows x64 platforms
- Architecture: Follow .NET conventions with src/, tests/, and tools/ directories

### `infrastructure-management-portal`

A modern, enterprise-ready web application for managing infrastructure data including servers, SSL certificates, applications, and services. Built with Next.js and Supabase.
The application features a dynamic schema management system allowing administrators to create custom data models via UI, role-based access control, audit logging, and comprehensive security with Row Level Security (RLS).

#### `infrastructure-management-portal` Build/Run Commands
- Install dependencies: `npm install`
- Development server: `npm run dev` (requires Supabase services running)
- Build for production: `npm run build`
- Start production: `npm start`
- Lint: `npm run lint`
- Type check: `npx tsc --noEmit`
- Docker Compose (all services): `docker-compose up -d`
- Stop services: `docker-compose down`

#### `infrastructure-management-portal` Code Style Guidelines
- TypeScript strict mode enabled
- React Server Components by default, Client Components when needed ('use client')
- Formatting: Prettier with 2-space indentation
- Imports: Group by external, internal (@/), components, types
- Naming: PascalCase for components, camelCase for functions/variables
- Error handling: Try-catch with user-friendly error messages
- Testing: Jest + React Testing Library (when added)
- Components: Functional components with TypeScript interfaces for props
- State management: React hooks, Server Actions for mutations
- Styling: Tailwind CSS utility classes
- Database: Supabase client with Row Level Security (RLS)
- API routes: Next.js App Router convention in src/app/api/
- Architecture: Next.js App Router with src/app/, src/components/, src/lib/
