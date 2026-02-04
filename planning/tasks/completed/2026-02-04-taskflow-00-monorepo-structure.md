---
status: completed
epic: taskflow-v1-core
effort: 2-3 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [architecture, monorepo, tooling]
completed_at: 2026-02-04
---

# TaskFlow 00: Monorepo Structure Design

**Goal**: Define the directory structure and build tooling for the `taskflow` monorepo.

## Context
TaskFlow is a hybrid project with:
- **Go CLI** (`cmd/taskflow`)
- **Python API** (`services/api`)
- **Postgres DB** (`migrations/`)
- **Shared Contracts** (`proto/` or `schemas/`)

We need a clean structure that supports independent development of services while sharing contracts.

## Structure Decision: Pattern C (Hybrid Standard)
**Status**: Confirmed
We are adopting the "Hybrid Standard" layout. This balances Go conventions with Python service isolation.

```text
taskflow/
├── bin/                    # Local build artifacts (gitignored)
├── cmd/
│   └── taskflow/           # Go CLI entrypoint (Main package)
├── internal/               # Private Go packages (TUI, Logic)
├── services/
│   └── semantic-engine/    # Python API Service (Renamed from 'brain')
│       ├── src/
│       ├── tests/
│       ├── Dockerfile
│       └── uv.lock         # Python dependency lock
├── contracts/              # Shared Interface Definitions
│   ├── proto/
│   └── gen/                # Generated code (committed)
├── migrations/             # Flyway SQL scripts
├── dev/                    # Docker-compose & local env
├── Makefile                # Unified task runner
└── go.mod                  # Root Go module
```

## Decisions
1.  **Single Go Module**: `go.mod` at root. Covers `cmd/` and `internal/`.
2.  **Single Python Package**: The API lives in `services/semantic-engine`.
3.  **No Workspaces**: Avoid complexity until we have >1 artifact per language.

## Deliverables
- [ ] Create directory tree.
- [ ] Initialize `go.mod` at root.
- [ ] Initialize `services/api` with `uv init`.
- [ ] Create `Makefile` with `build`, `test`, `dev` targets.
