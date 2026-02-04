# Research: TaskFlow Monorepo Structure

**Status**: Proposal
**Created**: 2026-01-03
**Goal**: Define the optimal directory structure for a polyglot (Go/Python/SQL) project with shared contracts.

## The Constraints
1.  **Languages**: Go (CLI), Python (API/AI), SQL (Migrations).
2.  **Shared Contracts**: Need a single source of truth (likely Protobuf) that generates code for both Go and Python.
3.  **Build System**: Needs to be simple initially (Make) but scalable (Pants/Bazel friendly).
4.  **Distribution**: The Go CLI is distributed as a binary; the Python API is a Docker container.

## Pattern A: Language-First (The "Google" Style)
Organize by language at the top level.
```
repo/
├── go/
│   ├── cmd/taskflow/
│   ├── internal/
│   └── go.mod
├── python/
│   ├── api/
│   ├── ai/
│   └── pyproject.toml
├── proto/
└── migrations/
```
- **Pros**: Clear tooling boundaries. Easy for language-specific linters.
- **Cons**: Separates related domain logic. "TaskFlow" concepts are split across root folders.

## Pattern B: Service-First (The Microservices Style)
Organize by deployable unit.
```
repo/
├── cli/                 # Go project
│   ├── main.go
│   └── go.mod
├── server/              # Python project
│   ├── app.py
│   └── pyproject.toml
├── contracts/           # Shared
├── migrations/          # Shared DB
└── tools/               # Build scripts
```
- **Pros**: "What runs where" is obvious. CLI and Server are distinct artifacts.
- **Cons**: If we add a second Go service later, `cli/` might be a bad name.

## Pattern C: The "Hybrid Standard" (Recommended)
A blend of the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) and modern polyglot practices.

```text
taskflow/
├── bin/                    # Local build artifacts (gitignored)
├── cmd/                    # Go main applications
│   └── taskflow/           # The CLI entrypoint
├── internal/               # Private Go code (CLI logic, TUI)
├── services/               # Backend Services
│   └── api/                # The Python "Brain"
│       ├── src/
│       ├── tests/
│       ├── Dockerfile
│       └── uv.lock         # Modern python packaging
├── contracts/              # The Interface Layer
│   ├── proto/              # .proto definitions
│   └── gen/                # Generated code (committed or gitignored?)
│       ├── go/
│       └── python/
├── migrations/             # Flyway SQL scripts
│   ├── V1__init.sql
│   └── V2__vectors.sql
├── dev/                    # Local dev environment
│   ├── docker-compose.yml
│   └── .env.example
├── docs/                   # Documentation
└── Makefile                # The universal task runner
```

### Why Pattern C?
1.  **Go Best Practices**: `cmd/` and `internal/` are what Go developers expect.
2.  **Service Isolation**: `services/api` creates a clear sandbox for the Python environment.
3.  **Contracts as First-Class Citizens**: Explicit `contracts/` folder highlights the importance of the shared schema.
4.  **Dev/Ops Separation**: `dev/` holds the docker-compose glue, keeping the root clean.

## Build Tooling Strategy

### The "Universal Makefile"
We use `make` as the entry point for humans, wrapping language-specific tools.

```makefile
# Makefile
.PHONY: all dev proto

all: proto build-cli build-api

proto:
    # Run protoc via docker to ensure consistent version
    ./tools/codegen.sh

build-cli:
    go build -o bin/taskflow ./cmd/taskflow

build-api:
    cd services/api && docker build -t taskflow-api .

dev:
    docker-compose -f dev/docker-compose.yml up -d
    go run ./cmd/taskflow
```

### Dependency Management
- **Go**: `go.mod` in root (or if `services/api` grows, maybe separate? No, usually keep one go.mod for the repo if it's mostly Go, but here it's mixed).
    - *Correction*: Since Python is in `services/api`, the root doesn't strictly need to be the Go module root, but it's convenient for `internal/`.
    - **Decision**: `go.mod` at root. `services/api` has its own `pyproject.toml`.

- **Python**: Use `uv` (faster `pip` replacement) in `services/api`.

## Open Questions
1.  **Generated Code**: Commit it (`gen/`) or build on CI?
    - *Recommendation*: **Commit it**. Makes `go get` / `go build` work without needing `protoc` installed. Easier for onboarding.
2.  **Database Access**: Does Go CLI need DB access?
    - *Decision*: **No**. Per previous research, Go reads JSON (fast path) or calls API (smart path). Only Python and Migration tool touch Postgres. This simplifies the Go build massively (no CGO for sqlite/postgres drivers needed potentially).

## Recommendation
Adopt **Pattern C**. It respects Go conventions while treating the Python API as a distinct, encapsulated service.
