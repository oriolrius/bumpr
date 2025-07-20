BINARY_NAME := bumpr
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
BUILD_DIR := dist

.PHONY: build clean test lint build-all install dev

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for Linux and Windows
build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Install locally
install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Development build with race detection
dev:
	go build -race $(LDFLAGS) -o $(BINARY_NAME) .

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"