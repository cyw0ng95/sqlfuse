# Makefile for SQLsmith-Go
# Optimized for fast parallel builds and efficient caching

.PHONY: all build clean test install-deps help vendor dev-build release-build
.DEFAULT_GOAL := help

# Build configuration
GO := go
GOFLAGS := 
OUTPUT_DIR := output
CACHE_DIR := .cache

# Build optimization flags
# -trimpath: removes absolute file paths for better caching across environments
# -ldflags="-s -w": strips debug info and symbol table to reduce binary size
COMMON_BUILD_FLAGS := -trimpath
DEV_LDFLAGS := 
RELEASE_LDFLAGS := -ldflags="-s -w"

# Executables to build
TURSO_EXEC := $(OUTPUT_DIR)/turso_embedded_executor
GO_SQLITE3_EXEC := $(OUTPUT_DIR)/go_sqlite3_embedded_executor
SERVER_EXEC := $(OUTPUT_DIR)/server

# Source directories
TURSO_SRC := ./cmd/executors/turso_embedded
GO_SQLITE3_SRC := ./cmd/executors/go_sqlite3_embedded
SERVER_SRC := ./cmd/server

# Parallel build support - automatically use available CPU cores
MAKEFLAGS += -j$(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

all: dev-build ## Build all executables (development mode with optimizations)

dev-build: $(TURSO_EXEC) $(GO_SQLITE3_EXEC) $(SERVER_EXEC) ## Fast development build with caching optimizations
	@echo "Development build complete!"

release-build: clean ## Production build with maximum optimization (strips debug info)
	@echo "Building release binaries..."
	@$(MAKE) COMMON_BUILD_FLAGS="$(COMMON_BUILD_FLAGS) $(RELEASE_LDFLAGS)" dev-build
	@echo "Release build complete!"

$(TURSO_EXEC): | $(OUTPUT_DIR)
	@echo "Building turso_embedded_executor..."
	@cd $(TURSO_SRC) && $(GO) build $(COMMON_BUILD_FLAGS) -o ../../../$@ .

$(GO_SQLITE3_EXEC): | $(OUTPUT_DIR)
	@echo "Building go_sqlite3_embedded_executor..."
	@cd $(GO_SQLITE3_SRC) && $(GO) build $(COMMON_BUILD_FLAGS) -o ../../../$@ .

$(SERVER_EXEC): | $(OUTPUT_DIR)
	@echo "Building server..."
	@cd $(SERVER_SRC) && $(GO) build $(COMMON_BUILD_FLAGS) -o ../../$@ .

$(OUTPUT_DIR):
	@mkdir -p $(OUTPUT_DIR)

$(CACHE_DIR):
	@mkdir -p $(CACHE_DIR)

##@ Dependencies

install-deps: ## Download and verify Go module dependencies
	@echo "Downloading dependencies..."
	@$(GO) work sync
	@$(GO) mod download -x
	@echo "Dependencies installed!"

vendor: ## Vendor dependencies for faster CI builds
	@echo "Vendoring dependencies for all modules..."
	@cd internal && $(GO) mod vendor
	@cd $(TURSO_SRC) && $(GO) mod vendor
	@cd $(GO_SQLITE3_SRC) && $(GO) mod vendor
	@cd $(SERVER_SRC) && $(GO) mod vendor
	@echo "Vendoring complete!"

##@ Testing

test: | $(CACHE_DIR) ## Run tests with coverage
	@echo "Running tests..."
	@mkdir -p $(CACHE_DIR)
	@cd internal && $(GO) test -coverprofile=../$(CACHE_DIR)/coverage.out ./...
	@$(GO) tool cover -html=$(CACHE_DIR)/coverage.out -o $(CACHE_DIR)/coverage.html
	@echo "Tests complete! Coverage report: $(CACHE_DIR)/coverage.html"

test-verbose: | $(CACHE_DIR) ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@mkdir -p $(CACHE_DIR)
	@cd internal && $(GO) test -v -coverprofile=../$(CACHE_DIR)/coverage.out ./...
	@$(GO) tool cover -html=$(CACHE_DIR)/coverage.out -o $(CACHE_DIR)/coverage.html
	@echo "Tests complete! Coverage report: $(CACHE_DIR)/coverage.html"

##@ Development

build-watch: ## Rebuild on file changes (requires entr)
	@echo "Watching for changes... (Press Ctrl+C to stop)"
	@find . -name '*.go' | entr -c make dev-build

fmt: ## Format all Go code
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Formatting complete!"

lint: ## Run linters (requires golangci-lint)
	@echo "Running linters..."
	@golangci-lint run ./...
	@echo "Linting complete!"

##@ Maintenance

clean: ## Remove build artifacts and caches
	@echo "Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f cmd/executors/turso_embedded/turso_embedded
	@rm -f cmd/executors/go_sqlite3_embedded/go_sqlite3_embedded
	@rm -f cmd/server/server
	@echo "Clean complete!"

clean-all: clean ## Remove all artifacts including Go build cache
	@echo "Cleaning all caches..."
	@$(GO) clean -cache -modcache -testcache
	@rm -rf $(CACHE_DIR)
	@rm -rf vendor
	@echo "All caches cleaned!"

##@ Build Performance

build-benchmark: ## Benchmark clean build time
	@echo "Benchmarking clean build..."
	@$(MAKE) clean > /dev/null 2>&1
	@time $(MAKE) dev-build

build-profile: ## Profile build performance
	@echo "Profiling build with -x flag..."
	@$(MAKE) clean > /dev/null 2>&1
	@cd $(TURSO_SRC) && $(GO) build -x $(COMMON_BUILD_FLAGS) -o ../../../$(TURSO_EXEC) . 2>&1 | head -50

cache-info: ## Display Go cache information
	@echo "Go Build Cache: $$($(GO) env GOCACHE)"
	@echo "Go Module Cache: $$($(GO) env GOMODCACHE)"
	@echo "Cache size: $$(du -sh $$($(GO) env GOCACHE) 2>/dev/null | cut -f1 || echo 'N/A')"
