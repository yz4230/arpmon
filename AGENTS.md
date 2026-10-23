# Repository Guidelines

## Project Structure & Module Organization

`main.go` starts the Cobra CLI. Add commands under `cmd/`; each subcommand has its own package, such as `cmd/servecmd` or `cmd/migratedbcmd`. Application code lives in `internal/`: ARP scanning in `arpscan`, PostgreSQL setup and migrations in `db`, domain types in `domain`, HTTP handlers and templates in `gateway`, persistence in `repository`, and shared test helpers in `testutil`. Keep HTML templates in `internal/gateway/templates`. Environment defaults are documented in `.env.example`; local `.env` files must remain uncommitted.

## Build, Test, and Development Commands

- `mise install` installs the pinned development tools, including Air.
- `cp .env.example .env` creates local database and server configuration.
- `mise run dev-db:start` starts the PostgreSQL 18.6 development container; `mise run dev-db:stop` stops it.
- `go run . --help` lists available CLI commands; `go run . serve` runs the HTTP service.
- `air` runs the application with live reload during development.
- `go build ./...` compiles every package.
- `go test ./...` runs all tests. Database and repository tests require the development PostgreSQL instance and `.env` settings.
- `go vet ./...` performs standard Go static checks.

## Coding Style & Naming Conventions

Use `gofmt` for all Go files; do not hand-align code. Follow standard Go naming: exported identifiers use `PascalCase`, internal identifiers use `camelCase`, and package names are short lowercase words. Keep command wiring in `cmd` and reusable behavior in the appropriate `internal` package. Wrap errors with useful operation context and preserve the original error using `%w`.

## Testing Guidelines

Use Go's `testing` package and place tests beside their implementation as `*_test.go`. Name tests `TestBehavior` or `TestTypeBehavior`; use table-driven tests when several cases share setup. Reuse `internal/testutil` for temporary databases and assertions. Run `go test ./...` before submitting changes. No coverage threshold is currently enforced; add focused regression tests for changed behavior.

## Commit & Pull Request Guidelines

History follows Conventional Commit-style subjects such as `feat(probe): add probe persistence`, `feat(db): add embedded database migrations`, and `build: add environment configuration template`. Use an imperative, concise subject with an optional scope.

Pull requests should explain the behavior change, note database or configuration impacts, link the relevant issue, and list verification commands. Include screenshots when changing gateway templates or rendered pages. Keep unrelated refactors in separate commits or pull requests.
