# Makefile for git-migration

BINARY=git-migration
VERSION=0.1.0
MAIN_PACKAGE=./cmd/git-migration
BUILD_DIR=bin
PLATFORMS=linux windows darwin
ARCH=amd64

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -o $(BINARY) $(MAIN_PACKAGE)

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

.PHONY: clean
clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf $(BUILD_DIR)

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: install
install: build
	cp $(BINARY) $(GOPATH)/bin/
