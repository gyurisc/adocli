# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

`adocli` is a Go CLI tool (`ado`) for Azure DevOps. It uses Cobra for commands, Viper for config, and go-keyring for secure PAT storage. No external Azure DevOps SDK — raw `net/http` with a hand-rolled API client.

## Build & Development Commands

```bash
make build      # Build binary to bin/ado (injects version via ldflags)
make install    # Build + copy to $GOPATH/bin/ado
make test       # go test ./... -v
make lint       # golangci-lint run
make fmt        # gofmt -s -w .
make vet        # go vet ./...
make all        # fmt vet lint test build
```

Prerequisites: Go 1.22+, golangci-lint (for linting).

## Architecture

**Three layers:**
- `cmd/` — Cobra command definitions (presentation layer). Each file has `init()` that wires subcommands onto `rootCmd`.
- `internal/api/` — HTTP client for Azure DevOps REST API v7.1. Two request paths: `do()` for org-level URLs, `doRaw()` for pre-built URLs (project-scoped).
- `internal/config/` — JSON config at `~/.config/ado/config.json` (organization, project, output_format).

**Key patterns:**
- Commands use `RunE` (not `Run`) with named handler functions, not anonymous.
- `newAPIClient()` helper in `cmd/workitem.go` constructs the API client from Viper config + keyring PAT. Used by all API-calling commands.
- `resolveProject(cmd)` checks `--project` flag first, then Viper config fallback.
- Output format (`table`/`json`/`plain`) controlled by `--json`/`--plain` global flags with Viper fallback. Commands switch on `OutputFormat()`.
- Work item mutations use Azure DevOps JSON Patch format (`application/json-patch+json`) with `PatchField{Op, Path, Value}`.
- Auth: PAT in OS keyring (service `"adocli"`, user `"pat"`), sent as HTTP Basic with empty username.

## Command Tree

```
ado
├── auth login|logout|status
├── config set|get|list
├── workitem (alias: wi) list|show|create|update
├── pr (alias: pullrequest) list|show|create|approve|reject
└── version
```

## Conventions

- Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`
- Run `make fmt` before committing
- All exported functions need doc comments
- Errors use `fmt.Errorf("context: %w", err)` wrapping, bubble up through `RunE`
- No tests exist yet — `_test.go` files need to be created
