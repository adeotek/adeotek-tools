// Package main provides the entry point for the git-repos-backup application
package main

import (
	"os"

	"github.com/adeotek/adeotek-tools/git-repos-backup/internal/app"
)

func main() {
	os.Exit(app.Run())
}
