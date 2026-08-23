# Mayfly Optimization Algorithm - Task Runner

# Default recipe to display available commands
default:
    @just --list

# Build the project
build:
    go build -v ./...

# Run tests with coverage (fast - without race detection)
test:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests without coverage (quickest)
test-quick:
    go test -v -short ./...

# Run tests with race detection (slower, skips long-running benchmark suite)
test-race:
    go test -v -race -short -timeout 5m ./...

# Run all tests including long-running benchmark suite (no race detection)
test-full:
    go test -v -timeout 10m ./...

# Build, vet, and test the root module and every nested example module.
test-modules:
    #!/usr/bin/env bash
    set -euo pipefail
    while IFS= read -r module_file; do
        module_dir="$(dirname "$module_file")"
        echo "==> $module_dir"
        (
            cd "$module_dir"
            go build ./...
            go vet ./...
            go test ./...
        )
    done < <(find . -name go.mod -not -path './.git/*' -print | sort)

# Run integration tests (Gherkin/Cucumber)
test-integration:
    go test -v -run TestFeatures

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Run the examples
run:
    cd examples && go run main.go

# Build the WebAssembly demo into ./dist
build-wasm-demo:
    ./scripts/build-wasm-demo.sh

# Build and serve the WebAssembly demo locally
run-wasm-demo: build-wasm-demo
    @echo "Serving the demo at http://localhost:8090"
    python3 -m http.server -d dist 8090

# Build the demo for js/wasm without emitting a binary (a fast compile check)
check-wasm-demo:
    cd examples/wasm-demo && GOOS=js GOARCH=wasm go build -o /dev/null . && go build -o /dev/null ./...

# Install the formatters and linters used by `just fmt` / `just lint`
setup-deps:
    #!/usr/bin/env bash
    set -euo pipefail
    export PATH=$HOME/go/bin:$PATH
    echo "Installing development dependencies..."

    # treefmt (formatter multiplexer)
    command -v treefmt >/dev/null 2>&1 || { echo "Installing treefmt..."; curl -fsSL https://github.com/numtide/treefmt/releases/download/v2.5.0/treefmt_2.5.0_linux_amd64.tar.gz | sudo tar -C /usr/local/bin -xz treefmt; }

    # golangci-lint v2 (linter + formatter runner)
    command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2; }

    # Go formatters
    command -v gofumpt >/dev/null 2>&1 || { echo "Installing gofumpt..."; go install mvdan.cc/gofumpt@latest; }
    command -v gci >/dev/null 2>&1 || { echo "Installing gci..."; go install github.com/daixiang0/gci@latest; }

    # Shell formatter
    command -v shfmt >/dev/null 2>&1 || { echo "Installing shfmt..."; go install mvdan.cc/sh/v3/cmd/shfmt@latest; }

    # Markdown/JSON/YAML formatter
    command -v prettier >/dev/null 2>&1 || { echo "Installing prettier..."; npm install -g prettier || echo "Prettier installation failed - npm not found."; }

    # TOML formatter
    command -v taplo >/dev/null 2>&1 || { echo "Installing taplo..."; cargo install taplo-cli --locked || echo "Taplo installation failed - cargo not found."; }

# Format all files with treefmt
fmt:
    #!/usr/bin/env bash
    export PATH=$HOME/go/bin:$PATH
    treefmt --allow-missing-formatter

# Alias for `just fmt`
treefmt: fmt

# Run linter
lint:
    #!/usr/bin/env bash
    export PATH=$HOME/go/bin:$PATH
    golangci-lint run --config ./.golangci.toml --timeout 5m ./...

# Run linter (with fix)
lint-fix:
    #!/usr/bin/env bash
    export PATH=$HOME/go/bin:$PATH
    golangci-lint fmt --config ./.golangci.toml
    golangci-lint run --config ./.golangci.toml --timeout 5m --fix ./...

# Tidy up dependencies
tidy:
    go mod tidy

# Verify dependencies
verify:
    go mod verify

# Clean build artifacts
clean:
    go clean
    rm -f coverage.out coverage.html
    rm -f *.test *.prof
    rm -rf dist/

# Generate documentation
docs:
    godoc -http=:6060

# Fail if any file is not formatted
check-formatted:
    #!/usr/bin/env bash
    export PATH=$HOME/go/bin:$PATH
    treefmt --allow-missing-formatter --fail-on-change

# Fail if go.mod/go.sum are not tidy
check-tidy:
    go mod tidy -diff

# Run all checks (format, lint, test)
check: check-formatted check-tidy lint test

# Run all checks with race detection
check-race: check-formatted check-tidy lint test-race

# Full CI pipeline
ci: verify check test-modules

# Full CI pipeline with race detection
ci-race: verify check-race test-modules

# Profile CPU performance
profile-cpu:
    go test -run '^$' -bench '^BenchmarkOptimizeBaseline$' -benchtime=5s -cpuprofile=cpu.pprof .
    @echo "CPU profile written to cpu.pprof; inspect it with: go tool pprof -top cpu.pprof"

# Profile memory usage
profile-mem:
    go test -run '^$' -bench '^BenchmarkOptimizeBaseline$' -benchtime=5s -memprofile=memory.pprof .
    @echo "Memory profile written to memory.pprof; inspect it with: go tool pprof -top -alloc_space memory.pprof"

# Run optimization with different algorithms for comparison
compare:
    #!/usr/bin/env bash
    echo "Running algorithm comparison..."
    cd examples
    echo "=== Standard Run ==="
    go run main.go
    echo ""
    echo "=== Performance comparison complete ==="

# Initialize development environment
init:
    go mod download
    @echo "Development environment ready!"
    @echo "Run 'just run' to test the examples"

# Create a new benchmark function template
new-benchmark name:
    #!/usr/bin/env bash
    echo "// {{name}} is a benchmark function." >> functions.go
    echo "// Global minimum is at f(?, ..., ?) = ?" >> functions.go
    echo "func {{name}}(x []float64) float64 {" >> functions.go
    echo "    // TODO: Implement {{name}} function" >> functions.go
    echo "    return 0.0" >> functions.go
    echo "}" >> functions.go
    echo "" >> functions.go
    echo "Added {{name}} function template to functions.go"

# Run specific optimization function
optimize func="Sphere" size="30" iter="1000":
    #!/usr/bin/env bash
    cd examples
    cat > temp_optimize.go << EOF
    package main
    import (
        "fmt"
        "github.com/cwbudde/mayfly"
    )
    func main() {
        config := mayfly.NewDefaultConfig()
        config.ObjectiveFunc = mayfly.{{func}}
        config.ProblemSize = {{size}}
        config.MaxIterations = {{iter}}
        config.LowerBound = -10
        config.UpperBound = 10
        
        result, err := mayfly.Optimize(config)
        if err != nil {
            panic(err)
        }
        
        fmt.Printf("Function: {{func}}\n")
        fmt.Printf("Best Cost: %.10f\n", result.GlobalBest.Cost)
        fmt.Printf("Evaluations: %d\n", result.FuncEvalCount)
    }
    EOF
    go run temp_optimize.go
    rm temp_optimize.go

# Install development tools (see also: just setup-deps)
install-tools: setup-deps
    go install golang.org/x/tools/cmd/godoc@latest

# Check for security vulnerabilities
security:
    go list -json -deps ./... | nancy sleuth

# Validate a prospective release without creating a tag
release-check version:
    #!/usr/bin/env bash
    set -euo pipefail
    release_version="{{version}}"
    release_version="${release_version#version=}"
    if [[ ! "$release_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
        echo "Invalid semantic version: $release_version" >&2
        exit 1
    fi
    grep -Fq "## [$release_version]" CHANGELOG.md
    test -s LICENSE
    test -s README.md
    test "$(go list -m)" = "github.com/cwbudde/mayfly"
    just verify
    just check-formatted
    just check-tidy
    just lint
    go vet ./...
    go test -timeout 20m ./...
    just test-modules

# Validate and create an annotated release tag locally
release version:
    #!/usr/bin/env bash
    set -euo pipefail
    release_version="{{version}}"
    release_version="${release_version#version=}"
    release_tag="v$release_version"
    just release-check "$release_version"
    if [[ -n "$(git status --porcelain)" ]]; then
        echo "Release requires a clean worktree" >&2
        exit 1
    fi
    if git rev-parse --verify --quiet "refs/tags/$release_tag" >/dev/null; then
        echo "Tag already exists: $release_tag" >&2
        exit 1
    fi
    git tag -a "$release_tag" -m "Release $release_tag"
    echo "Ready to push: git push origin main $release_tag"
