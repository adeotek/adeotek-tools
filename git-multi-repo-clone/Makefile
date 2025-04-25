# Makefile for git-multi-repo-clone

# Binary name
BINARY=git-multi-repo-clone
VERSION=0.1.0

# Main package path
MAIN_PACKAGE=./cmd/git-multi-repo-clone

# Build directory
BUILD_DIR=bin

# Platforms to build for
PLATFORMS=linux windows darwin

# Architecture to build for
ARCH=amd64

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	go build -o $(BINARY) $(MAIN_PACKAGE)

# Build for all platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	for platform in $(PLATFORMS); do \
		echo "Building for $${platform}/$(ARCH)..."; \
		if [ "$${platform}" = "windows" ]; then \
			GOOS=$${platform} GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY)_$(VERSION)_$${platform}_$(ARCH).exe $(MAIN_PACKAGE); \
		else \
			GOOS=$${platform} GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY)_$(VERSION)_$${platform}_$(ARCH) $(MAIN_PACKAGE); \
		fi; \
	done

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf $(BUILD_DIR)

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run integration tests
.PHONY: integration-test
integration-test:
	RUN_INTEGRATION_TESTS=1 go test -v ./tests

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	golint ./...

# Install locally
.PHONY: install
install: build
	cp $(BINARY) $(GOPATH)/bin/

# Create a config file from example if none exists
.PHONY: config
config:
	test -f config.yaml || cp config.yaml.example config.yaml

# Run the application
.PHONY: run
run: build config
	./$(BINARY)