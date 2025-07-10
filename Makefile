# GoRDP Makefile
# A comprehensive build system for the GoRDP project

# Variables
BINARY_NAME=gordp
GUI_BINARY_NAME=gordp-gui
QT_BINARY_NAME=gordp-qt-gui
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME)_windows.exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOLINT=golangci-lint
GODOC=$(GOCMD) doc
GOCOVER=$(GOCMD) test -cover
GOBENCH=$(GOCMD) test -bench=.
GORACE=$(GOCMD) test -race

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Directories
BUILD_DIR=build
DIST_DIR=dist
DOCS_DIR=docs
EXAMPLES_DIR=examples
TEST_DIR=test
GUI_BUILD_DIR=$(BUILD_DIR)/gui
QT_BUILD_DIR=$(BUILD_DIR)/qt-gui

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "GoRDP - Go Remote Desktop Protocol Client"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build          # Build the main binary"
	@echo "  make build-gui      # Build the Go GUI"
	@echo "  make build-qt       # Build the Qt GUI"
	@echo "  make test           # Run all tests"
	@echo "  make lint           # Run linter"
	@echo "  make clean          # Clean build artifacts"
	@echo "  make install        # Install dependencies"

# Build targets
.PHONY: build
build: ## Build the main binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: build-gui
build-gui: ## Build the Go GUI application
	@echo "Building $(GUI_BINARY_NAME)..."
	@mkdir -p $(GUI_BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(GUI_BUILD_DIR)/$(GUI_BINARY_NAME) ./gui

.PHONY: build-qt
build-qt: ## Build the Qt C++ GUI application
	@echo "Building $(QT_BINARY_NAME)..."
	@mkdir -p $(QT_BUILD_DIR)
	@if [ -d "qt-gui" ]; then \
		cd qt-gui && ./build.sh; \
		cp qt-gui/build/bin/gordp-gui $(QT_BUILD_DIR)/$(QT_BINARY_NAME); \
	else \
		echo "Qt GUI source not found. Run 'make build-qt-source' first."; \
		exit 1; \
	fi

.PHONY: build-qt-source
build-qt-source: ## Build Qt GUI from source (requires Qt6)
	@echo "Building Qt GUI from source..."
	@if [ ! -d "qt-gui" ]; then \
		echo "Qt GUI source directory not found."; \
		exit 1; \
	fi
	@cd qt-gui && ./build.sh

.PHONY: build-all
build-all: build build-gui ## Build core and Go GUI for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) .
	@echo "Building GUI for all platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(GUI_BUILD_DIR)/$(GUI_BINARY_NAME)-linux-amd64 ./gui
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(GUI_BUILD_DIR)/$(GUI_BINARY_NAME)-windows-amd64.exe ./gui
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(GUI_BUILD_DIR)/$(GUI_BINARY_NAME)-darwin-amd64 ./gui

.PHONY: build-examples
build-examples: ## Build example applications
	@echo "Building examples..."
	@mkdir -p $(BUILD_DIR)/examples
	$(GOBUILD) -o $(BUILD_DIR)/examples/interactive_client $(EXAMPLES_DIR)/interactive_example/interactive_client.go
	$(GOBUILD) -o $(BUILD_DIR)/examples/config_client $(EXAMPLES_DIR)/config_example/config_client.go
	$(GOBUILD) -o $(BUILD_DIR)/examples/management_console $(EXAMPLES_DIR)/management_example/management_console.go
	$(GOBUILD) -o $(BUILD_DIR)/examples/webrtc_gateway $(EXAMPLES_DIR)/webrtc_example/webrtc_gateway.go

# Test targets
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-gui
test-gui: ## Run GUI-specific tests
	@echo "Running GUI tests..."
	$(GOTEST) -v ./gui/...

.PHONY: test-qt
test-qt: ## Run Qt GUI tests
	@echo "Running Qt GUI tests..."
	@if [ -d "qt-gui/tests" ]; then \
		cd qt-gui/tests && ./run_tests.sh; \
	else \
		echo "Qt GUI tests not found."; \
	fi

.PHONY: test-short
test-short: ## Run tests with short flag
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GORACE) ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOCOVER) ./...

.PHONY: test-coverage-html
test-coverage-html: ## Run tests with coverage and generate HTML report
	@echo "Running tests with coverage HTML report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-benchmark
test-benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOBENCH) ./...

.PHONY: test-benchmark-memory
test-benchmark-memory: ## Run benchmarks with memory profiling
	@echo "Running benchmarks with memory profiling..."
	$(GOBENCH) -benchmem ./...

# Linting and formatting
.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(GOLINT) run ./...; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

.PHONY: fmt-check
fmt-check: ## Check code formatting
	@echo "Checking code formatting..."
	@if [ "$$(gofmt -l . | wc -l)" -gt 0 ]; then \
		echo "Code is not formatted. Run 'make fmt' to format."; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "Code is properly formatted."; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

# Dependencies
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOGET) -v -t -d ./...

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: deps-clean
deps-clean: ## Clean module cache
	@echo "Cleaning module cache..."
	$(GOMOD) clean -cache

# Qt dependencies
.PHONY: deps-qt
deps-qt: ## Install Qt6 dependencies
	@echo "Installing Qt6 dependencies..."
	@if command -v apt >/dev/null 2>&1; then \
		sudo apt update && sudo apt install -y qt6-base-dev qt6-tools-dev qt6-websockets-dev qt6-charts-dev qt6-declarative-dev cmake build-essential pkg-config; \
	elif command -v yum >/dev/null 2>&1; then \
		sudo yum install -y qt6-qtbase-devel qt6-qttools-devel qt6-qtwebsockets-devel qt6-qtcharts-devel qt6-qtdeclarative-devel cmake gcc-c++ pkgconfig; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y qt6-qtbase-devel qt6-qttools-devel qt6-qtwebsockets-devel qt6-qtcharts-devel qt6-qtdeclarative-devel cmake gcc-c++ pkgconfig; \
	elif command -v brew >/dev/null 2>&1; then \
		brew install qt6 cmake pkg-config; \
	else \
		echo "Package manager not detected. Please install Qt6 manually."; \
		exit 1; \
	fi

# Documentation
.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)
	$(GODOC) -all ./... > $(DOCS_DIR)/api.md 2>/dev/null || echo "Documentation generation completed"

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "Serving documentation at http://localhost:6060"
	$(GODOC) -http=:6060

# Clean targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

.PHONY: clean-qt
clean-qt: ## Clean Qt build artifacts
	@echo "Cleaning Qt build artifacts..."
	@if [ -d "qt-gui/build" ]; then \
		rm -rf qt-gui/build; \
	fi
	@if [ -d "$(QT_BUILD_DIR)" ]; then \
		rm -rf $(QT_BUILD_DIR); \
	fi

.PHONY: clean-all
clean-all: clean clean-qt ## Clean everything including dependencies
	@echo "Cleaning everything..."
	rm -rf go.sum
	$(GOMOD) clean -cache

# Install targets
.PHONY: install
install: build ## Install the main binary
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

.PHONY: install-gui
install-gui: build-gui ## Install the Go GUI
	@echo "Installing $(GUI_BINARY_NAME)..."
	@mkdir -p $(GOPATH)/bin
	cp $(GUI_BUILD_DIR)/$(GUI_BINARY_NAME) $(GOPATH)/bin/

.PHONY: install-qt
install-qt: build-qt ## Install the Qt GUI
	@echo "Installing $(QT_BINARY_NAME)..."
	@if [ -f "$(QT_BUILD_DIR)/$(QT_BINARY_NAME)" ]; then \
		sudo mkdir -p /usr/local/bin; \
		sudo cp $(QT_BUILD_DIR)/$(QT_BINARY_NAME) /usr/local/bin/gordp-gui; \
		sudo chmod +x /usr/local/bin/gordp-gui; \
		echo "✓ Qt GUI installed as gordp-gui"; \
	else \
		echo "Qt GUI binary not found. Run 'make build-qt' first."; \
		exit 1; \
	fi

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Run targets
.PHONY: run
run: build ## Build and run the main application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-gui
run-gui: build-gui ## Build and run the Go GUI
	@echo "Running $(GUI_BINARY_NAME)..."
	./$(GUI_BUILD_DIR)/$(GUI_BINARY_NAME)

.PHONY: run-qt
run-qt: build-qt ## Build and run the Qt GUI
	@echo "Running $(QT_BINARY_NAME)..."
	./$(QT_BUILD_DIR)/$(QT_BINARY_NAME)

.PHONY: run-example
run-example: build-examples ## Build and run the interactive example
	@echo "Running interactive example..."
	@echo "Usage: ./$(BUILD_DIR)/examples/interactive_client <host:port> <username> <password>"
	@echo "Example: ./$(BUILD_DIR)/examples/interactive_client 192.168.1.100:3389 administrator password"

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t gordp:latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -it --rm gordp:latest

.PHONY: docker-clean
docker-clean: ## Clean Docker images
	@echo "Cleaning Docker images..."
	docker rmi gordp:latest 2>/dev/null || true

# Release targets
.PHONY: release
release: clean build-all ## Create release builds
	@echo "Creating release builds..."
	@mkdir -p $(DIST_DIR)
	cp $(BUILD_DIR)/* $(DIST_DIR)/
	@echo "Release builds created in $(DIST_DIR)"

.PHONY: release-zip
release-zip: release ## Create release zip files
	@echo "Creating release zip files..."
	cd $(DIST_DIR) && zip -r gordp-linux-amd64.zip $(BINARY_UNIX)
	cd $(DIST_DIR) && zip -r gordp-windows-amd64.zip $(BINARY_WINDOWS)
	cd $(DIST_DIR) && zip -r gordp-darwin-amd64.zip $(BINARY_DARWIN)
	@echo "Release zip files created in $(DIST_DIR)"

# Security targets
.PHONY: security-scan
security-scan: ## Run security scan
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

.PHONY: security-audit
security-audit: ## Run security audit
	@echo "Running security audit..."
	$(GOMOD) verify
	$(GOMOD) download -json | jq -r '.Path + "@" + .Version' | xargs -I {} go list -m -json {} | jq -r 'select(.Security != null) | .Path + ": " + .Security.Vulnerable'

# Performance targets
.PHONY: profile
profile: ## Generate CPU profile
	@echo "Generating CPU profile..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	$(GOCMD) tool pprof cpu.prof

.PHONY: profile-memory
profile-memory: ## Generate memory profile
	@echo "Generating memory profile..."
	$(GOTEST) -memprofile=mem.prof -bench=. ./...
	$(GOCMD) tool pprof mem.prof

# Development workflow
.PHONY: dev-setup
dev-setup: install-tools deps ## Setup development environment
	@echo "Development environment setup complete"

.PHONY: dev-setup-full
dev-setup-full: install-tools deps deps-qt ## Setup full development environment with Qt
	@echo "Full development environment setup complete"

.PHONY: pre-commit
pre-commit: fmt lint vet test ## Run pre-commit checks
	@echo "Pre-commit checks completed"

.PHONY: ci
ci: deps fmt-check lint vet test-coverage ## Run CI checks
	@echo "CI checks completed"

# Utility targets
.PHONY: version
version: ## Show version information
	@echo "GoRDP Version: $(shell git describe --tags --always --dirty)"
	@echo "Build Time: $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"
	@echo "Go Version: $(shell go version)"
	@echo "Git Commit: $(shell git rev-parse HEAD)"

.PHONY: info
info: ## Show project information
	@echo "GoRDP - Go Remote Desktop Protocol Client"
	@echo "Repository: https://github.com/kdsmith18542/gordp"
	@echo "Go Module: $(shell go list -m)"
	@echo "Dependencies:"
	@go list -m all | head -10

.PHONY: check
check: ## Check if all required tools are installed
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Git is required but not installed. Aborting." >&2; exit 1; }
	@echo "All required tools are installed."

.PHONY: check-qt
check-qt: ## Check if Qt6 is installed
	@echo "Checking Qt6 installation..."
	@if command -v qmake6 >/dev/null 2>&1 || command -v qmake >/dev/null 2>&1; then \
		echo "✓ Qt6 found"; \
	else \
		echo "✗ Qt6 not found. Run 'make deps-qt' to install."; \
		exit 1; \
	fi

# Default targets for common workflows
.PHONY: all
all: check deps fmt lint vet test build build-gui ## Run all checks and build

.PHONY: all-qt
all-qt: check check-qt deps deps-qt fmt lint vet test build build-gui build-qt ## Run all checks and build with Qt

.PHONY: quick
quick: test build build-gui ## Quick test and build

.PHONY: full
full: clean deps-update fmt lint vet test-race test-coverage build-all ## Full development cycle

# Print variables (for debugging)
.PHONY: print-vars
print-vars: ## Print Makefile variables
	@echo "BINARY_NAME: $(BINARY_NAME)"
	@echo "GUI_BINARY_NAME: $(GUI_BINARY_NAME)"
	@echo "QT_BINARY_NAME: $(QT_BINARY_NAME)"
	@echo "BUILD_DIR: $(BUILD_DIR)"
	@echo "DIST_DIR: $(DIST_DIR)"
	@echo "GOCMD: $(GOCMD)"
	@echo "LDFLAGS: $(LDFLAGS)" 