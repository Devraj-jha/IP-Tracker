.PHONY: build clean install test all linux darwin windows help

# Build variables
APP_NAME := ip-tracker
VERSION := $(shell grep -o 'appVersion = ".*"' main.go | cut -d'"' -f2)
BUILD_DIR := build
LDFLAGS := -ldflags "-s -w -X main.appVersion=$(VERSION)"

# Default target
all: build

help: ## Show this help
	@echo ""
	@echo "IP Tracker Build System"
	@echo "======================"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## Build for current platform
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)"

install: ## Install to $GOPATH/bin
	@echo "Installing $(APP_NAME)..."
	go install $(LDFLAGS) .
	@echo "Installed: $$(which $(APP_NAME) 2>/dev/null || echo '$$GOPATH/bin/$(APP_NAME)')"

linux: ## Build for Linux (amd64)
	@echo "Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

darwin: ## Build for macOS (arm64 + amd64)
	@echo "Building for macOS arm64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-macos-arm64 .
	@echo "Building for macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-macos-amd64 .
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)-macos-*"

windows: ## Build for Windows (amd64)
	@echo "Building for Windows amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe"

all-platforms: linux darwin windows ## Build for all platforms
	@echo ""
	@echo "All builds complete!"
	@ls -la $(BUILD_DIR)/

clean: ## Remove build artifacts
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."

test: ## Run tests
	go test -v ./...

fmt: ## Format code
	gofmt -s -w .

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Install golangci-lint: brew install golangci-lint"; exit 1; }
	golangci-lint run
