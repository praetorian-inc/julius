.DELETE_ON_ERROR:

.PHONY: all test clean help install validate
.PHONY: build/linux build/darwin build/windows build/all
.PHONY: test/unit test/coverage test/verbose
.PHONY: lint lint/fix

.DEFAULT_GOAL := help

# Variables with environment override support
GO ?= go
BUILD_DIR ?= ./dist
INSTALL_DIR ?= /usr/local/bin

# Version information from git
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
GO_LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)

GO_FLAGS ?= -v -trimpath

# Auto-discover Go sources
GO_SOURCES := $(shell find . -name '*.go' -not -path './vendor/*' -type f)
PROBE_SOURCES := $(shell find ./probes -name '*.yaml' -type f)

# Test configuration
TEST_FLAGS ?= -race -coverprofile=coverage.out
TEST_TIMEOUT ?= 5m

#
# Entry Point Targets
#

help: ## Display available targets
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: build test ## Build binary and run tests

#
# Build Targets
#

$(BUILD_DIR):
	mkdir -p $@

build: $(BUILD_DIR)/julius ## Build julius binary for current platform

$(BUILD_DIR)/julius: $(GO_SOURCES) $(PROBE_SOURCES) | $(BUILD_DIR)
	$(GO) build $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@ ./cmd/julius

build/all: build/linux build/darwin build/windows ## Build for all platforms

build/linux: $(BUILD_DIR)/julius-linux-amd64 ## Build for Linux amd64

$(BUILD_DIR)/julius-linux-amd64: $(GO_SOURCES) $(PROBE_SOURCES) | $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@ ./cmd/julius

build/darwin: $(BUILD_DIR)/julius-darwin-amd64 $(BUILD_DIR)/julius-darwin-arm64 ## Build for macOS (amd64 and arm64)

$(BUILD_DIR)/julius-darwin-amd64: $(GO_SOURCES) $(PROBE_SOURCES) | $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@ ./cmd/julius

$(BUILD_DIR)/julius-darwin-arm64: $(GO_SOURCES) $(PROBE_SOURCES) | $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@ ./cmd/julius

build/windows: $(BUILD_DIR)/julius-windows-amd64.exe ## Build for Windows amd64

$(BUILD_DIR)/julius-windows-amd64.exe: $(GO_SOURCES) $(PROBE_SOURCES) | $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@ ./cmd/julius

#
# Test Targets
#

test: test/unit ## Run all tests (alias for test/unit)

test/unit: ## Run unit tests
	$(GO) test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./...

test/coverage: ## Run tests with coverage report
	$(GO) test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test/verbose: ## Run tests with verbose output
	$(GO) test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) -v ./...

#
# Lint Targets
#

lint: ## Run golangci-lint
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

lint/fix: ## Run golangci-lint with auto-fix
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --fix ./...

#
# Validation Targets
#

validate: ## Validate probe YAML files
	$(BUILD_DIR)/julius validate --probes ./probes

#
# Installation Targets
#

install: $(BUILD_DIR)/julius ## Install julius to INSTALL_DIR (default: /usr/local/bin)
	install -m 755 $< $(INSTALL_DIR)/julius
	@echo "Installed julius to $(INSTALL_DIR)/julius"

#
# Cleanup Targets
#

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
