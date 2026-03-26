# Repository Guidelines

## Project Structure & Module Organization
`cmd/bookmux` contains the CLI entrypoint. Core logic lives under `internal/`: `audio` handles probing and M4B assembly, `cli` owns flags/completions/interactive mode, `ffmpeg` wraps dependency checks and execution, `input` discovers and sorts source files, and `model` defines shared types. Build output goes to `.bin/` for local work and `dist/` for release artifacts. Keep new code inside `internal/<area>` unless it is part of the public CLI binary.

## Build, Test, and Development Commands
Use the existing `Makefile` targets:

- `make build` builds the CLI to `.bin/bookmux`.
- `make test` runs all Go tests with `go test ./...`.
- `make lint` runs `golangci-lint` using the repo config in `.golangci.yml`.

For quick manual checks, run `.bin/bookmux --help` after building, or test a merge flow with `--input`, `--output`, `--title`, and `--author`.

## Coding Style & Naming Conventions
Follow standard Go formatting: tabs for indentation, `gofmt`-formatted files, and lower-case package names. Exported identifiers use `CamelCase`; internal helpers use `camelCase`. Keep packages focused and small; prefer names that match the current layout (`input`, `audio`, `ffmpeg`) over generic utility buckets. Linting is strict on vet/staticcheck/revive-style issues, security checks, and test assertions, so run `make lint` before opening a PR.

## Testing Guidelines
Write table-driven tests in `*_test.go` files next to the code they cover, following the existing pattern in `internal/ffmpeg/exec_test.go`. Run `make test` locally before pushing. There is no published coverage gate in this repo, but new parsing, validation, and ffmpeg-related branches should include unit tests where practical.

## Commit & Pull Request Guidelines
Recent history uses short, imperative commit subjects. Keep commit messages concise, focused, and scoped to one change. For pull requests, include:

- a brief summary of user-visible behavior,
- linked issues when relevant,
- terminal output or screenshots for CLI/TUI changes,
- confirmation that `make test` and `make lint` passed.

## Release & Configuration Notes
Releases are published through GitHub Actions in `.github/workflows/release.yml` using GoReleaser on version tags like `v1.2.3`. Do not commit secrets or machine-specific paths; ffmpeg packaging and signing are handled in CI.
