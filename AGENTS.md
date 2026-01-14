# Repository Guidelines

## Project Structure & Module Organization
- `cmd/`: Cobra CLI commands and subcommands.
- `internal/`: Core implementation (config, gitops, sops, tofu, plugins, util).
- `tests/`: BDD suite (Gherkin features + step definitions).
- `testdata/`: Fixtures used by unit/BDD tests.
- `schema/`: Generated JSON schema outputs.
- `docs/`: User and developer documentation.
- `main.go`: CLI entry point.

## Build, Test, and Development Commands
Use Mise tasks from `.mise.toml`:
- `mise install`: install toolchain versions.
- `mise run build`: build `bin/openCenter` with version metadata.
- `mise run build-all`: cross-compile binaries.
- `mise run test`: run unit tests (`go test ./internal/...`).
- `mise run godog`: run BDD tests (`tests/features`).
- `mise run fmt`: format all Go code via `gofmt -w .`.
- `mise run lint`: run `golangci-lint` (requires `bin/golangci-lint`).

## Coding Style & Naming Conventions
- Go formatting is enforced with `gofmt`; run `mise run fmt` before committing.
- Standard Go naming: `CamelCase` for exported identifiers, `mixedCase` for locals.
- Test files use `*_test.go`; test functions use `TestXxx` naming.
- Gherkin features live under `tests/features/*.feature`.

## Testing Guidelines
- Unit tests: `go test ./internal/...` (via `mise run test`).
- BDD tests: `mise run godog` (non-@wip scenarios).
- Optional tags: `mise run godog-wip` runs only `@wip` scenarios.
- Keep fixtures in `testdata/` and reuse existing patterns.

## Commit & Pull Request Guidelines
- Commit messages in history are mostly Conventional Commit style (`feat:`, `fix:`), with occasional imperative summaries (e.g., `Add ...`). Prefer Conventional Commits for new work.
- PRs should include a concise description, test commands run, and any relevant docs updates.

## Security & Configuration Notes
- SOPS is used for secrets; avoid committing plaintext secrets. See `.sops.yaml`.
- YAML formatting is checked with `.yamllint`; consider `yamllint` for config changes.
