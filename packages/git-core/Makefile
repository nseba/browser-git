# Makefile for building BrowserGit WASM module

# Build configuration
TINYGO := tinygo
GO := go
OUTPUT_DIR := ../browser-git/dist
WASM_FILE := $(OUTPUT_DIR)/git-core.wasm
WASM_EXEC := $(OUTPUT_DIR)/wasm_exec.js
SOURCE := main.go

# TinyGo build flags
TINYGO_FLAGS := -target=wasm \
	-opt=2 \
	-no-debug \
	-panic=trap \
	-scheduler=none \
	-gc=leaking

# Development build flags (with debug info)
TINYGO_DEV_FLAGS := -target=wasm \
	-opt=1 \
	-panic=trap \
	-scheduler=none \
	-gc=leaking

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m

.PHONY: all build build-dev clean test lint fmt help watch install-deps

# Default target
all: build

# Production build
build: install-deps
	@echo "$(COLOR_BLUE)Building WASM module (production)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	@$(TINYGO) build $(TINYGO_FLAGS) -o $(WASM_FILE) $(SOURCE)
	@echo "$(COLOR_GREEN)✓ WASM built successfully: $(WASM_FILE)$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Size: $$(du -h $(WASM_FILE) | cut -f1)$(COLOR_RESET)"
	@cp $$($(TINYGO) env TINYGOROOT)/targets/wasm_exec.js $(WASM_EXEC)
	@echo "$(COLOR_GREEN)✓ wasm_exec.js copied to $(WASM_EXEC)$(COLOR_RESET)"

# Development build (with source maps and less optimization)
build-dev: install-deps
	@echo "$(COLOR_BLUE)Building WASM module (development)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	@$(TINYGO) build $(TINYGO_DEV_FLAGS) -o $(WASM_FILE) $(SOURCE)
	@echo "$(COLOR_GREEN)✓ WASM built successfully: $(WASM_FILE)$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Size: $$(du -h $(WASM_FILE) | cut -f1)$(COLOR_RESET)"
	@cp $$($(TINYGO) env TINYGOROOT)/targets/wasm_exec.js $(WASM_EXEC)
	@echo "$(COLOR_GREEN)✓ wasm_exec.js copied to $(WASM_EXEC)$(COLOR_RESET)"

# Install dependencies
install-deps:
	@echo "$(COLOR_BLUE)Installing Go dependencies...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ Dependencies installed$(COLOR_RESET)"

# Clean build artifacts
clean:
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(OUTPUT_DIR)/git-core.wasm
	@rm -rf $(OUTPUT_DIR)/wasm_exec.js
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

# Run Go tests
test:
	@echo "$(COLOR_BLUE)Running Go tests...$(COLOR_RESET)"
	@$(GO) test -v ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

# Run Go tests with coverage
test-coverage:
	@echo "$(COLOR_BLUE)Running Go tests with coverage...$(COLOR_RESET)"
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

# Lint Go code
lint:
	@echo "$(COLOR_BLUE)Linting Go code...$(COLOR_RESET)"
	@$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ Lint complete$(COLOR_RESET)"

# Format Go code
fmt:
	@echo "$(COLOR_BLUE)Formatting Go code...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Format complete$(COLOR_RESET)"

# Watch for changes and rebuild (requires fswatch or inotifywait)
watch:
	@echo "$(COLOR_BLUE)Watching for changes...$(COLOR_RESET)"
	@if command -v fswatch > /dev/null; then \
		fswatch -o . | while read; do make build-dev; done; \
	elif command -v inotifywait > /dev/null; then \
		while inotifywait -r -e modify,create,delete .; do make build-dev; done; \
	else \
		echo "$(COLOR_YELLOW)⚠ Install fswatch (macOS) or inotify-tools (Linux) for watch mode$(COLOR_RESET)"; \
		exit 1; \
	fi

# Verify TinyGo installation
verify-tinygo:
	@echo "$(COLOR_BLUE)Verifying TinyGo installation...$(COLOR_RESET)"
	@which $(TINYGO) > /dev/null || (echo "$(COLOR_YELLOW)⚠ TinyGo not found. Install from https://tinygo.org$(COLOR_RESET)" && exit 1)
	@echo "$(COLOR_GREEN)✓ TinyGo found: $$($(TINYGO) version)$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)TINYGOROOT: $$($(TINYGO) env TINYGOROOT)$(COLOR_RESET)"

# Show build info
info:
	@echo "$(COLOR_BOLD)Build Information$(COLOR_RESET)"
	@echo "  Go version:     $$($(GO) version)"
	@echo "  TinyGo version: $$($(TINYGO) version 2>/dev/null || echo 'not installed')"
	@echo "  Output file:    $(WASM_FILE)"
	@echo "  Build flags:    $(TINYGO_FLAGS)"

# Help target
help:
	@echo "$(COLOR_BOLD)BrowserGit WASM Build System$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)build$(COLOR_RESET)           Build production WASM module (optimized)"
	@echo "  $(COLOR_GREEN)build-dev$(COLOR_RESET)       Build development WASM module (with debug info)"
	@echo "  $(COLOR_GREEN)clean$(COLOR_RESET)           Remove build artifacts"
	@echo "  $(COLOR_GREEN)test$(COLOR_RESET)            Run Go unit tests"
	@echo "  $(COLOR_GREEN)test-coverage$(COLOR_RESET)   Run tests with coverage report"
	@echo "  $(COLOR_GREEN)lint$(COLOR_RESET)            Lint Go code with go vet"
	@echo "  $(COLOR_GREEN)fmt$(COLOR_RESET)             Format Go code with go fmt"
	@echo "  $(COLOR_GREEN)watch$(COLOR_RESET)           Watch for changes and rebuild (requires fswatch)"
	@echo "  $(COLOR_GREEN)verify-tinygo$(COLOR_RESET)   Verify TinyGo installation"
	@echo "  $(COLOR_GREEN)info$(COLOR_RESET)            Show build information"
	@echo "  $(COLOR_GREEN)install-deps$(COLOR_RESET)    Install Go dependencies"
	@echo "  $(COLOR_GREEN)help$(COLOR_RESET)            Show this help message"
	@echo ""
	@echo "$(COLOR_BOLD)Examples:$(COLOR_RESET)"
	@echo "  make build         # Build production WASM"
	@echo "  make build-dev     # Build development WASM"
	@echo "  make watch         # Watch and rebuild on changes"
	@echo "  make test          # Run tests"
