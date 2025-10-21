# Repository Guidelines

## Project Structure & Modules
- Root package: `github.com/cwbudde/mayfly` (Go library).
- Source and tests live at repo root: `*.go`, `*_test.go`.
- Docs in `docs/`; runnable examples in `examples/`; Gherkin features in `features/`.
- Tooling configs: `.golangci.yml`, `treefmt.toml`, `justfile`.

## Build, Test, and Dev Commands
- `just build`: Compile all packages (`go build ./...`).
- `just test`: Run unit tests with coverage and write `coverage.{out,html}`.
- `just test-race` / `just test-quick`: Race-checked or short tests.
- `just test-integration`: Run Godog-backed feature tests (`features/*.feature`).
- `just bench`: Execute benchmarks with memory stats.
- `just run`: Run the example app in `examples/`.
- `just fmt` / `just treefmt` / `just lint`: Format and lint (golangci-lint).
- `just check`: Format, lint, and test; use in PRs/CI. 

## Coding Style & Naming
- Use Go defaults: `go fmt` (tabs, standard imports). Keep files idiomatic.
- Naming: Exported identifiers `CamelCase`; internal `lowerCamel`; packages short, lowercase.
- Linting: `golangci-lint` per `.golangci.yml`. Prefer small, cohesive files and pure funcs.

## Testing Guidelines
- Framework: standard `go test`; BDD via `godog` (run with `just test-integration`).
- Test files: `*_test.go`; names `TestXxx`, `BenchmarkXxx`, `ExampleXxx`.
- Coverage: ensure `just test` stays green; check `coverage.html` locally.
- Benchmarks: place in `*_test.go` using `BenchmarkXxx(b *testing.B)`.

## Commit & PR Guidelines
- Commits: conventional style like `feat: ...`, `fix: ...`, `chore: ...` (see `git log`).
- PRs: include goal, key changes, before/after impact, and links to issues.
- Checks: run `just check` (or `check-race`) and attach relevant results/screenshots.
- Tests: add/adjust unit + integration tests for behavior changes.

## Security & Maintenance
- Dependencies: `just tidy` then `just verify`. 
- Scan: `just security` (uses Nancy) before release.
- Release prep: `just ci` then tag with `just release version=<semver>`.
