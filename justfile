# Mayfly Optimization Algorithm - Task Runner

# Default recipe to display available commands
default:
    @just --list

# Build the project
build:
    go build -v ./...

# Run tests with coverage
test:
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests without coverage
test-quick:
    go test -v ./...

# Run integration tests (Gherkin/Cucumber)
test-integration:
    go test -v -run TestFeatures

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Run the examples
run:
    cd examples && go run main.go

# Format code with go fmt
fmt:
    go fmt ./...

# Format all files with treefmt
treefmt:
    #!/usr/bin/env bash
    export PATH=$HOME/go/bin:$PATH
    treefmt

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

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

# Generate documentation
docs:
    godoc -http=:6060

# Run all checks (format, lint, test)
check: fmt treefmt lint test

# Full CI pipeline
ci: tidy verify check

# Profile CPU performance
profile-cpu:
    cd examples && go run -cpuprofile=cpu.prof main.go
    go tool pprof cpu.prof

# Profile memory usage
profile-mem:
    cd examples && go run -memprofile=mem.prof main.go
    go tool pprof mem.prof

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

# Install development tools
install-tools:
    go install golang.org/x/tools/cmd/godoc@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check for security vulnerabilities
security:
    go list -json -deps ./... | nancy sleuth

# Release preparation
release version:
    #!/usr/bin/env bash
    echo "Preparing release {{version}}"
    just ci
    git tag -a v{{version}} -m "Release version {{version}}"
    echo "Ready to push: git push origin v{{version}}"