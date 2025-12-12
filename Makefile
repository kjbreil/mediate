# Makefile for Mediate project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint
BINARY_NAME=mediate
BINARY_DIR=bin

# Build directory
BUILD_DIR=./$(BINARY_DIR)

# App variables
VERSION?=$(shell git describe --tags --always --dirty --match='v*' 2> /dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2> /dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Make targets
.PHONY: all build clean test lint deps tidy help darwin linux windows docker

all: clean deps build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mediate

# Cross-platform builds
darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mediate
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/mediate

linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mediate
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mediate

windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mediate

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN) ./...

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run linter
lint:
	@echo "Running linter..."
	$(GOLINT) run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download

# Tidy go.mod and go.sum
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Create default config
config:
	@echo "Creating default config..."
	@mkdir -p ~/.config/mediate
	./$(BINARY_DIR)/$(BINARY_NAME) --config=~/.config/mediate/config.yaml --create-config

# Docker
docker:
	@echo "Building Docker image..."
	docker build -t mediate:$(VERSION) .

# Version bumping - creates and pushes git tags to trigger CI
.PHONY: patch minor major

patch:
	@latest=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	major=$$(echo $$latest | sed 's/v//' | cut -d. -f1); \
	minor=$$(echo $$latest | sed 's/v//' | cut -d. -f2); \
	patch=$$(echo $$latest | sed 's/v//' | cut -d. -f3); \
	new="v$$major.$$minor.$$((patch + 1))"; \
	echo "Bumping $$latest -> $$new"; \
	git tag -a $$new -m "Release $$new" && git push origin $$new

minor:
	@latest=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	major=$$(echo $$latest | sed 's/v//' | cut -d. -f1); \
	minor=$$(echo $$latest | sed 's/v//' | cut -d. -f2); \
	new="v$$major.$$((minor + 1)).0"; \
	echo "Bumping $$latest -> $$new"; \
	git tag -a $$new -m "Release $$new" && git push origin $$new

major:
	@latest=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	major=$$(echo $$latest | sed 's/v//' | cut -d. -f1); \
	new="v$$((major + 1)).0.0"; \
	echo "Bumping $$latest -> $$new"; \
	git tag -a $$new -m "Release $$new" && git push origin $$new

# Help target
help:
	@echo "Mediate Makefile Help"
	@echo "---------------------"
	@echo "make              : Build the application after cleaning and getting dependencies"
	@echo "make build        : Build the application"
	@echo "make clean        : Clean build artifacts"
	@echo "make test         : Run tests"
	@echo "make lint         : Run linter"
	@echo "make deps         : Install dependencies"
	@echo "make tidy         : Tidy go.mod and go.sum"
	@echo "make darwin       : Build for macOS (amd64 and arm64)"
	@echo "make linux        : Build for Linux (amd64 and arm64)"
	@echo "make windows      : Build for Windows (amd64)"
	@echo "make docker       : Build Docker image"
	@echo "make config       : Create default config in ~/.config/mediate"
	@echo "make patch        : Bump patch version and push tag (v1.0.0 -> v1.0.1)"
	@echo "make minor        : Bump minor version and push tag (v1.0.0 -> v1.1.0)"
	@echo "make major        : Bump major version and push tag (v1.0.0 -> v2.0.0)"
	@echo "make help         : Show this help"

# Default target
.DEFAULT_GOAL := build
