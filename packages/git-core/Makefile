# Makefile for building BrowserGit WASM module

# Build configuration
TINYGO := tinygo
GO := go
OUTPUT_DIR := ../browser-git/dist
WASM_FILE := $(OUTPUT_DIR)/git-core.wasm
WASM_EXEC := $(OUTPUT_DIR)/wasm_exec.js
SOURCE := main.go

# TinyGo build flags for maximum optimization
TINYGO_FLAGS := -target=wasm \
	-opt=z \
	-no-debug \
	-panic=trap \
	-scheduler=none \
	-gc=leaking

# Additional optimization with wasm-opt
WASM_OPT := wasm-opt
WASM_OPT_FLAGS := -Oz --enable-bulk-memory --enable-sign-ext

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

.PHONY: all build build-dev build-optimized clean test lint fmt help watch install-deps analyze-size

# Default target
all: build

# Production build
build: install-deps
	@echo "$(COLOR_BLUE)Building WASM module (production)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	@$(TINYGO) build $(TINYGO_FLAGS) -o $(WASM_FILE) $(SOURCE)
	@echo "$(COLOR_GREEN)✓ WASM built successfully: $(WASM_FILE)$(COLOR_RESET)"
	@$(MAKE) analyze-size
	@cp $$($(TINYGO) env TINYGOROOT)/targets/wasm_exec.js $(WASM_EXEC)
	@echo "$(COLOR_GREEN)✓ wasm_exec.js copied to $(WASM_EXEC)$(COLOR_RESET)"

# Optimized build with wasm-opt
build-optimized: install-deps
	@echo "$(COLOR_BLUE)Building WASM module (production with wasm-opt)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	@$(TINYGO) build $(TINYGO_FLAGS) -o $(WASM_FILE).tmp $(SOURCE)
	@if command -v $(WASM_OPT) > /dev/null 2>&1; then \
		echo "$(COLOR_BLUE)Running wasm-opt for additional optimization...$(COLOR_RESET)"; \
		$(WASM_OPT) $(WASM_OPT_FLAGS) $(WASM_FILE).tmp -o $(WASM_FILE); \
		rm $(WASM_FILE).tmp; \
		echo "$(COLOR_GREEN)✓ WASM optimized with wasm-opt$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ wasm-opt not found, using unoptimized WASM$(COLOR_RESET)"; \
		mv $(WASM_FILE).tmp $(WASM_FILE); \
	fi
	@echo "$(COLOR_GREEN)✓ WASM built successfully: $(WASM_FILE)$(COLOR_RESET)"
	@$(MAKE) analyze-size
	@cp $$($(TINYGO) env TINYGOROOT)/targets/wasm_exec.js $(WASM_EXEC)
	@echo "$(COLOR_GREEN)✓ wasm_exec.js copied to $(WASM_EXEC)$(COLOR_RESET)"

# Analyze WASM size
analyze-size:
	@if [ -f $(WASM_FILE) ]; then \
		SIZE=$$(stat -f%z $(WASM_FILE) 2>/dev/null || stat -c%s $(WASM_FILE) 2>/dev/null); \
		SIZE_KB=$$(echo "scale=2; $$SIZE / 1024" | bc); \
		SIZE_MB=$$(echo "scale=2; $$SIZE / 1024 / 1024" | bc); \
		echo "$(COLOR_BLUE)WASM Size Analysis:$(COLOR_RESET)"; \
		echo "  Raw:     $$SIZE_KB KB ($$SIZE_MB MB)"; \
		if command -v gzip > /dev/null 2>&1; then \
			GZIP_SIZE=$$(gzip -c $(WASM_FILE) | wc -c); \
			GZIP_KB=$$(echo "scale=2; $$GZIP_SIZE / 1024" | bc); \
			GZIP_MB=$$(echo "scale=2; $$GZIP_SIZE / 1024 / 1024" | bc); \
			RATIO=$$(echo "scale=1; $$SIZE / $$GZIP_SIZE" | bc); \
			echo "  Gzipped: $$GZIP_KB KB ($$GZIP_MB MB)"; \
			echo "  Ratio:   $$RATIO:1"; \
			if [ $$(echo "$$GZIP_MB > 2" | bc) -eq 1 ]; then \
				echo "$(COLOR_YELLOW)⚠ Warning: Gzipped size exceeds 2MB target$(COLOR_RESET)"; \
			else \
				echo "$(COLOR_GREEN)✓ Size meets 2MB target$(COLOR_RESET)"; \
			fi \
		fi \
	else \
		echo "$(COLOR_YELLOW)⚠ WASM file not found$(COLOR_RESET)"; \
	fi

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
	@echo "  $(COLOR_GREEN)build$(COLOR_RESET)             Build production WASM module (optimized)"
	@echo "  $(COLOR_GREEN)build-dev$(COLOR_RESET)         Build development WASM module (with debug info)"
	@echo "  $(COLOR_GREEN)build-optimized$(COLOR_RESET)   Build with aggressive optimization (wasm-opt)"
	@echo "  $(COLOR_GREEN)analyze-size$(COLOR_RESET)      Analyze WASM bundle size (raw and gzipped)"
	@echo "  $(COLOR_GREEN)clean$(COLOR_RESET)             Remove build artifacts"
	@echo "  $(COLOR_GREEN)test$(COLOR_RESET)              Run Go unit tests"
	@echo "  $(COLOR_GREEN)test-coverage$(COLOR_RESET)     Run tests with coverage report"
	@echo "  $(COLOR_GREEN)lint$(COLOR_RESET)              Lint Go code with go vet"
	@echo "  $(COLOR_GREEN)fmt$(COLOR_RESET)               Format Go code with go fmt"
	@echo "  $(COLOR_GREEN)watch$(COLOR_RESET)             Watch for changes and rebuild (requires fswatch)"
	@echo "  $(COLOR_GREEN)verify-tinygo$(COLOR_RESET)     Verify TinyGo installation"
	@echo "  $(COLOR_GREEN)info$(COLOR_RESET)              Show build information"
	@echo "  $(COLOR_GREEN)install-deps$(COLOR_RESET)      Install Go dependencies"
	@echo "  $(COLOR_GREEN)help$(COLOR_RESET)              Show this help message"
	@echo ""
	@echo "$(COLOR_BOLD)Examples:$(COLOR_RESET)"
	@echo "  make build            # Build production WASM"
	@echo "  make build-optimized  # Build with maximum optimization"
	@echo "  make build-dev        # Build development WASM"
	@echo "  make analyze-size     # Check bundle size"
	@echo "  make watch            # Watch and rebuild on changes"
	@echo "  make test             # Run tests"
