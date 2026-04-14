# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

**Gorge** is a hydrological data harvesting service. It polls 40+ river gauge data sources (government APIs, HTML pages, CSVs) on a cron schedule and stores time-series measurements in a database. It exposes a REST API and a CLI client.

## Commands

```bash
make tools       # Install required dev tools (modd, golangci-lint, tz lookup)
make build       # Build gorge-server and gorge-cli binaries to /build/
make test        # Run all tests (except postgres-dependent tests)
make lint        # Run golangci-lint
make run         # Dev mode with modd live reload (rebuilds + retests on file change)
make typescript  # Regenerate TypeScript definitions from Go structs
```

Run a single test:
```bash
go test -v ./core -run TestNewLatestFilter
go test -v ./scripts/quebec -run TestQuebec_Harvest_JSON
```

Postgres-dependent tests require Docker and use build tag `nodocker`:
```bash
go test -tags nodocker ./...
```

## Architecture

### Core Concepts

- **Script** — a data source adapter. Each implements `ListGauges()` and `Harvest()` in [core/script.go](core/script.go). The `Harvest` method streams `Measurement` values over a channel.
- **Job** — a scheduled harvest: ties a Script + gauge filter + cron expression together.
- **ScriptRegistry** — maps script names to factory functions. All scripts self-register in `init()`.

### Package Map

| Package                | Role                                                                                                    |
| ---------------------- | ------------------------------------------------------------------------------------------------------- |
| [core/](core/)         | Domain types (`Gauge`, `Measurement`, `Job`, `Status`), Script interface, HTTP client, filter logic     |
| [scripts/](scripts/)   | 40+ data source implementations (one folder per country/agency)                                         |
| [server/](server/)     | REST API server (chi router), wired via `go.uber.org/fx` DI                                             |
| [cli/](cli/)           | Cobra CLI that makes HTTP calls to a running server                                                     |
| [schedule/](schedule/) | Job scheduler using `robfig/cron`; executes harvests and persists results                               |
| [storage/](storage/)   | `DatabaseManager` (PostgreSQL/SQLite) + `CacheManager` (Redis/in-memory) interfaces and implementations |
| [config/](config/)     | CLI flags + env vars merged into a single config struct                                                 |
| [version/](version/)   | Version string injected via ldflags at build time                                                       |

### Storage Layer

Two separate interfaces:
- **DatabaseManager** — durable time-series storage for measurements and job definitions (postgres or sqlite). Schema migrations live in [storage/migrations/](storage/migrations/).
- **CacheManager** — fast-access store for latest measurements and job statuses (redis or in-memory).

### Adding a New Data Source Script

Each script lives in its own folder under [scripts/](scripts/) and must:
1. Define a descriptor with `DefaultOptions()` and a `Factory` function matching the `ScriptFactory` signature.
2. Register itself via `ScriptRegistry.Register()` in `init()`.
3. Implement `ListGauges(ctx, options)` and `Harvest(ctx, options, gaugeIDs, measurements chan<- Measurement, mode HarvestMode)`.

See [scripts/testscripts/](scripts/testscripts/) for minimal reference implementations used in tests.

### HTTP Client

The shared HTTP client in [core/http.go](core/http.go) rotates fake user-agents, handles cookies, supports proxy, and has retry logic — use it (via `core.NewHTTPClient`) rather than `http.DefaultClient` in scripts.

## Key Configuration

Config is built from CLI flags and environment variables. Key env vars:
- `DB_URL` — postgres connection string (omit to use SQLite)
- `CACHE_URL` — redis connection string (omit to use in-memory)
- `HTTP_PROXY` — proxy for all script HTTP requests

Development credentials for external APIs are in [.env.development](.env.development).

## Linter

Configuration is in [.golangci.yml](.golangci.yml). The codebase uses `logrus` for structured logging and `testify` for assertions in tests.
