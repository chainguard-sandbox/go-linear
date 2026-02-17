# Makefile for go-linear
# Self-documenting with auto-generated help
#
# Usage: make <target>
# Run 'make' or 'make help' to see available commands

.PHONY: help build build-cli

# Project configuration
BINARY_NAME := go-linear
MODULE := github.com/chainguard-sandbox/go-linear
BINDIR := bin
GOFILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./upstream/*" -not -path "./internal/graphql/generated.go")

# Version information from git
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TREESTATE := $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
SOURCE_DATE_EPOCH := $(shell git log -1 --pretty=%ct 2>/dev/null || echo "0")

# Detect OS for date command compatibility
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    BUILD_DATE := $(shell date -u -r $(SOURCE_DATE_EPOCH) "+%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u "+%Y-%m-%dT%H:%M:%SZ")
else
    BUILD_DATE := $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "+%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u "+%Y-%m-%dT%H:%M:%SZ")
endif

# Version package for embedding version info
VERSION_PKG := $(MODULE)/pkg/linear

# Linker flags for embedding version information
LDFLAGS := -buildid= \
	-X $(VERSION_PKG).Version=$(GIT_VERSION) \
	-X $(VERSION_PKG).GitCommit=$(GIT_HASH) \
	-X $(VERSION_PKG).GitTreeState=$(GIT_TREESTATE) \
	-X $(VERSION_PKG).BuildDate=$(BUILD_DATE)

# Go build flags
GOFLAGS := -trimpath

# Default target - show help
help:  ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

#
# Build targets
#

build: build-cli  ## Build the CLI (includes MCP server)

#
# CLI targets (includes MCP via 'mcp' subcommand)
#

build-cli:  ## Build go-linear (CLI + MCP server in one binary)
	@echo "Building go-linear..."
	@mkdir -p $(BINDIR)
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINDIR)/go-linear ./cmd/linear
	@echo "✓ Built: $(BINDIR)/go-linear"
	@echo "  CLI mode: ./bin/go-linear issue list"
	@echo "  MCP mode: ./bin/go-linear mcp start"

install: build-cli  ## Install go-linear to $GOPATH/bin
	@echo "Installing go-linear..."
	@cp $(BINDIR)/go-linear $(GOPATH)/bin/go-linear
	@echo "✓ Installed go-linear to $(GOPATH)/bin/"

#
# Code generation
#

generate:  ## Run code generation (genqlient)
	@echo "Running code generation..."
	@go generate ./...
	@echo "✓ Code generation complete"

#
# Testing targets
#

test:  ## Run tests with race detection (mock tests only, no API key needed)
	@echo "Running mock tests..."
	go test -race -cover ./...

test-read:  ## Run live read-only tests (requires LINEAR_API_KEY with Read permission)
	@echo "Running live read-only tests..."
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "ERROR: LINEAR_API_KEY not set"; \
		echo "Run: export LINEAR_API_KEY=your-key"; \
		exit 1; \
	fi
	go test -tags=read -race -cover -v ./...

test-write:  ## Run live mutation tests (requires LINEAR_API_KEY with Write permission)
	@echo "Running live mutation tests..."
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "ERROR: LINEAR_API_KEY not set"; \
		echo "Run: export LINEAR_API_KEY=your-key"; \
		exit 1; \
	fi
	@echo "⚠️  WARNING: This will CREATE/UPDATE/DELETE data on test-server"
	go test -tags=write -race -cover -v ./...

test-verbose:  ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	go test -race -cover -v ./...

test-coverage:  ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✓ Coverage report saved to coverage.out"
	@echo "View with: go tool cover -html=coverage.out"

test-all:  ## Run all tests (mock + read + write, requires LINEAR_API_KEY with Write permission)
	@echo "Running all tests..."
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "ERROR: LINEAR_API_KEY not set for live tests"; \
		exit 1; \
	fi
	@$(MAKE) test
	@$(MAKE) test-read
	@$(MAKE) test-write

benchmark:  ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

#
# Profiling targets
#

profile-cpu:  ## Run benchmarks with CPU profiling
	@echo "Running benchmarks with CPU profiling..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof ./...
	@echo "✓ CPU profile saved to cpu.prof"
	@echo "Analyze with: go tool pprof cpu.prof"

profile-mem:  ## Run benchmarks with memory profiling
	@echo "Running benchmarks with memory profiling..."
	go test -bench=. -benchmem -memprofile=mem.prof ./...
	@echo "✓ Memory profile saved to mem.prof"
	@echo "Analyze with: go tool pprof -alloc_space mem.prof"

profile-all:  ## Run benchmarks with CPU and memory profiling
	@echo "Running benchmarks with full profiling..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "✓ CPU profile saved to cpu.prof"
	@echo "✓ Memory profile saved to mem.prof"
	@echo "\nAnalyze with:"
	@echo "  go tool pprof cpu.prof"
	@echo "  go tool pprof -alloc_space mem.prof"
	@echo "  go tool pprof -inuse_space mem.prof"

profile-example:  ## Run profiling example (generates cpu.prof, mem.prof, trace.out)
	@echo "Running profiling example..."
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "ERROR: LINEAR_API_KEY not set"; \
		exit 1; \
	fi
	@cd examples/profiling && go run main.go
	@echo "\n✓ Profiles generated in examples/profiling/"
	@echo "\nAnalyze with:"
	@echo "  cd examples/profiling"
	@echo "  go tool pprof cpu.prof"
	@echo "  go tool pprof -alloc_space mem.prof"
	@echo "  go tool trace trace.out"

profile-server:  ## Run profiling server (live profiling at :6060)
	@echo "Starting profiling server..."
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "ERROR: LINEAR_API_KEY not set"; \
		exit 1; \
	fi
	@echo "Server will run at http://localhost:6060/debug/pprof/"
	@PROFILE_MODE=server cd examples/profiling && go run main.go

benchmark-baseline:  ## Save benchmark baseline (for comparison)
	@echo "Saving benchmark baseline..."
	@go test -bench=. -benchmem -count=5 ./... | tee benchmark-baseline.txt
	@echo "✓ Baseline saved to benchmark-baseline.txt"

benchmark-compare:  ## Compare current benchmarks with baseline
	@echo "Running benchmarks and comparing with baseline..."
	@if [ ! -f benchmark-baseline.txt ]; then \
		echo "ERROR: No baseline found. Run 'make benchmark-baseline' first"; \
		exit 1; \
	fi
	@go test -bench=. -benchmem -count=5 ./... | tee benchmark-current.txt
	@echo "\n✓ Comparison:"
	@benchstat benchmark-baseline.txt benchmark-current.txt || echo "Install benchstat: go install golang.org/x/perf/cmd/benchstat@latest"

profile-clean:  ## Remove profile files
	@echo "Cleaning profile files..."
	@rm -f cpu.prof mem.prof trace.out block.prof mutex.prof
	@rm -f benchmark-baseline.txt benchmark-current.txt
	@rm -f examples/profiling/*.prof examples/profiling/*.out
	@echo "✓ Profile files cleaned"

#
# Code quality targets
#

fmt:  ## Format Go code
	@echo "Formatting code..."
	@gofmt -w $(GOFILES)
	@goimports -w -local $(MODULE) $(GOFILES)
	@echo "✓ Code formatted"

checkfmt:  ## Check code formatting
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -l $(GOFILES))" || (echo "Error: code is not formatted. Run 'make fmt'" && gofmt -l $(GOFILES) && exit 1)
	@echo "✓ Code formatting is correct"

lint:  ## Run linters
	@echo "Running linters..."
	@golangci-lint run
	@echo "✓ Linting passed"

vet:  ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet passed"

vulncheck:  ## Check for known vulnerabilities
	@echo "Checking for vulnerabilities..."
	@govulncheck ./...
	@echo "✓ No vulnerabilities found"

gosec:  ## Run gosec security scanner
	@echo "Running gosec..."
	@gosec ./...
	@echo "✓ Gosec passed"

nilaway:  ## Run nilaway nil safety checker
	@echo "Running nilaway..."
	@# Check all packages except pkg/linear (has example tests we skip)
	@go list ./pkg/... ./cmd/... ./internal/... | grep -v "pkg/linear$$" | xargs nilaway
	@# Check pkg/linear non-example files only (examples are documentation)
	@nilaway $$(find ./pkg/linear -maxdepth 1 -name '*.go' ! -name 'examples*_test.go')
	@echo "✓ Nilaway passed"

trivy:  ## Scan with trivy (if available)
	@echo "Scanning with trivy..."
	@trivy fs --severity HIGH,CRITICAL .

zizmor:  ## Run zizmor GitHub Actions security scanner
	@echo "Running zizmor..."
	@zizmor .github
	@echo "✓ Zizmor passed"

check: checkfmt vet lint test check-tidy  ## Run all checks (fmt, vet, lint, test, tidy) - use before commit
	@echo "✓ All checks passed!"

check-tidy:  ## Verify go.mod is tidy
	@echo "Verifying go.mod is tidy..."
	@go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "ERROR: go.mod is not tidy. Run 'go mod tidy'" && exit 1)

check-full: check vulncheck  ## Run all checks including vulncheck (slower)
	@echo "✓ All checks including vulncheck passed!"

#
# Dependency management
#

tidy:  ## Tidy go.mod and go.sum
	@echo "Tidying go.mod and go.sum..."
	@go mod tidy
	@echo "✓ go.mod and go.sum are tidy"

verify:  ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "✓ Dependencies verified"

download:  ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "✓ Dependencies downloaded"

#
# Schema management
#

schema:  ## Download Linear GraphQL schema
	@echo "Downloading Linear GraphQL schema..."
	@mkdir -p .
	@curl -sSL https://raw.githubusercontent.com/linear/linear/master/packages/sdk/src/schema.graphql -o schema.graphql
	@echo "✓ Schema downloaded to schema.graphql"

#
# Upstream sync (maintainers only)
#

check-sync-mode:  ## Check if sync mode is enabled
	@if [ ! -e "upstream/.git" ]; then \
		echo "❌ Not in sync mode. Upstream submodule not initialized."; \
		echo "To enable sync mode: git submodule update --init upstream"; \
		exit 1; \
	fi
	@echo "✅ Sync mode enabled"

sync-schema: check-sync-mode  ## Sync GraphQL schema from upstream Linear SDK (sync mode only)
	@./scripts/sync-schema.sh

sync-upstream: check-sync-mode  ## Sync everything from upstream SDK (sync mode only)
	@echo "🔄 Syncing with upstream Linear SDK..."
	@./scripts/sync-schema.sh
	@echo "🔨 Regenerating code..."
	@$(MAKE) generate
	@echo "🧪 Running tests..."
	@$(MAKE) test
	@echo "✅ Sync complete! Review changes and commit."

#
# Cleanup targets
#

clean:  ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINDIR)
	@rm -f coverage.out
	@rm -rf dist/
	@go clean -cache -testcache
	@echo "✓ Cleaned build artifacts"

#
# Release targets
#

#
# Setup targets
#

setup-golangci-lint:  ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo "✓ golangci-lint installed"

setup-goimports:  ## Install goimports
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✓ goimports installed"

setup-genqlient:  ## Install genqlient
	@echo "Installing genqlient..."
	@go install github.com/Khan/genqlient@latest
	@echo "✓ genqlient installed"

setup-govulncheck:  ## Install govulncheck
	@echo "Installing govulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✓ govulncheck installed"

setup-nilaway:  ## Install nilaway
	@echo "Installing nilaway..."
	@go install go.uber.org/nilaway/cmd/nilaway@latest
	@echo "✓ nilaway installed"

setup-trivy:  ## Install trivy (macOS)
	@echo "Installing trivy..."
	@brew install aquasecurity/trivy/trivy || echo "⚠ Install trivy manually: https://aquasecurity.github.io/trivy/"

setup-zizmor:  ## Install zizmor
	@echo "Installing zizmor..."
	@go install github.com/woodruffw/zizmor/cmd/zizmor@latest
	@echo "✓ zizmor installed"

setup: setup-golangci-lint setup-goimports setup-genqlient setup-govulncheck setup-nilaway  ## Install all development tools
	@echo "✓ All tools installed"
	@echo "Optional: Run 'make setup-trivy' and 'make setup-zizmor' for additional security tools"

dev: setup  ## Complete developer onboarding (setup tools + deps + verify)
	@echo "Installing dependencies..."
	@go mod download
	@go mod verify
	@echo "Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Get Linear API key: https://linear.app/settings/account/security"
	@echo "  2. Set environment variable: export LINEAR_API_KEY=your-key"
	@echo "  3. Run 'make help' to see available commands"

#
# Run targets
#

run:  ## Run the application (use ARGS="..." to pass arguments)
	@go run ./cmd/linear $(ARGS)

#
# Information targets
#

version:  ## Show version information
	@echo "Version:    $(GIT_VERSION)"
	@echo "Commit:     $(GIT_HASH)"
	@echo "Tree State: $(GIT_TREESTATE)"
	@echo "Build Date: $(BUILD_DATE)"

goversion:  ## Show Go version
	@go version

#
# All-in-one targets
#

all: clean build test lint  ## Clean, build, test, and lint

.DEFAULT_GOAL := help

modernize:  ## Run modernize analyzer for modern Go patterns (modern Go)
	@echo "Running modernize..."
	@go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest \
		./cmd/... ./pkg/... ./internal/... \
		2>&1 | { grep -v "internal/graphql\|examples/" || true; } | \
		{ grep ":" && echo "⚠️  Modernize suggestions found (review above)" || echo "✓ Modernize passed"; }
