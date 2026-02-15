# Makefile for sql-toolbox

# Variables
BINARY_NAME=sql-toolbox
BUILD_DIR=build
MAIN_PATH=./cmd/sql-toolbox
MODULE_NAME=github.com/adeotek/adeotek-tools/sql-toolbox

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-w -s"

.PHONY: all build clean test deps tidy run install

# Default target
all: clean deps build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all: build-linux build-windows build-darwin

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) ./...

# Tidy up go.mod and go.sum
tidy:
	@echo "Tidying up modules..."
	$(GOMOD) tidy

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PATH)

# Show help
help:
	@echo "Available targets:"
	@echo "  all         - Clean, download deps, and build"
	@echo "  build       - Build the application"
	@echo "  build-all   - Build for all platforms"
	@echo "  build-linux - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-darwin - Build for macOS"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Download dependencies"
	@echo "  tidy        - Tidy up go.mod and go.sum"
	@echo "  run         - Run the application"
	@echo "  install     - Install the application"
	@echo "  help        - Show this help message"
