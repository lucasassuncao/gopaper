# Contributing

## Requirements

- Go 1.26+ (see `go.mod`)
- Make (all common dev tasks are Makefile targets that fetch their own tools via `go run`, so nothing else needs to be installed globally)

## Getting set up

```pwsh
git clone https://github.com/lucasassuncao/gopaper.git
cd gopaper
make deps    # go mod download && go mod tidy
make build   # binary in bin/ for your current platform
```

A `.devcontainer` is included if you prefer to develop inside a container (VS Code: "Reopen in Container").

## Common tasks

| Command | Does |
|---|---|
| `make build` | Build a binary for the current platform only |
| `make build-all` | Build binaries for every supported platform |
| `make run` | `go run .` |
| `make fmt` | `go fmt ./...` |
| `make lint` | golangci-lint |
| `make security` | gosec static analysis |
| `make test` | Run the test suite (testdox output) |
| `make test-watch` | Re-run tests on file changes |
| `make test-coverage` | Tests with HTML + Cobertura coverage reports |
| `make docs` | Regenerate `internal/*/README.md` from doc comments (gomarkdoc) |
| `make all` | fmt, docs, lint, security, test-coverage — run this before opening a PR |
| `make clean` | Remove build artifacts and caches |

Run `make help` for the full list with descriptions.

## Before opening a PR

```pwsh
make all
```

This formats the code, regenerates the gomarkdoc `README.md` files under `internal/`, lints, runs `gosec`, and runs the full test suite with coverage. All four should be clean — a lint or test failure here is the same one CI will report.

If you touched `gopaper.yaml`'s schema (`internal/models/gopaper.go`), also run:

```pwsh
gopaper validate -c gopaper.yaml
```

against a real config, and check whether `docs/CONFIGURATION.md` and `docs/FILTERS.md` still match. `gopaper show-docs` renders the same schema reference live, so it's a quick way to sanity-check that field descriptions and defaults line up with what you changed.

## Project layout

- `internal/cmd` — CLI commands (cobra) and their wiring
- `internal/config` — viper integration, path resolution, tilde expansion
- `internal/models` — config schema (`Config`, `Categories`, `Filter`, ...) and their `yedit` metadata
- `internal/filters` — category file-filter compilation and matching
- `internal/history` — wallpaper history persistence (`prev`/`next`)
- `internal/helper` — wallpaper selection and OS wallpaper API calls
- `internal/updater` — self-update against GitHub releases

## Code style

- Match the existing style in the file you're editing; don't reformat unrelated code.
- Keep changes scoped to what the PR is about — a bug fix shouldn't carry drive-by refactors.
- Add or update tests for behavior you change, not just for new code.
- `golangci-lint` (config in `.golangci.yaml`) and `gofmt` must be clean; `make lint` runs both.

## Commit messages

Short, imperative subject line; a `type:` prefix matching the existing history (`feat:`, `fix:`, `refactor:`, `chore:`) is preferred. Explain *why* in the body when it isn't obvious from the diff.

## Releasing

Maintainer-only: `make tag VERSION=vX.Y.Z` creates and pushes an annotated tag (requires a clean working tree), which triggers `.github/workflows/release.yml` to build and publish via goreleaser.
