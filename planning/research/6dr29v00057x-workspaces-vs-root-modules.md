---
schema: 1
id: 6dr29v00057x
created: "2026-01-03"
description: Whether to use Go workspaces or a single root module; recommends one root go.mod
tags: []
updated_at: "2026-08-18"
---

# Research: Workspaces vs. Root Modules for TaskFlow

**Status**: Proposal / Education
**Created**: 2026-01-03
**Context**: The user comes from a web app background (likely JS monorepos with `packages/`) and is evaluating if Go/Python "Workspaces" are a better fit than the proposed "Root Module" pattern.

## The Mental Model Shift
In web/JS monorepos (`npm workspaces`, `turborepo`), you often have:
```
packages/ui
packages/utils
apps/web
apps/admin
```
Everything is a "package" and they depend on each other source-to-source.

**In TaskFlow (Go CLI + Python Service):**
Crucially, **Go code cannot import Python code** (and vice versa). They are totally separate universes that talk only via:
1.  **HTTP/JSON** (Runtime)
2.  **Protobuf/Generated Code** (Compile time)

## Option A: The "Root Module" Pattern (Current Recommendation)
```
repo/
├── go.mod              # One Go universe
├── cmd/taskflow/       # Main CLI
├── internal/           # Shared Go logic
├── services/
│   └── api/            # One Python universe
│       └── pyproject.toml
```
- **Pros**: Simplest tooling. `go build` just works. No "workspace" config files.
- **Cons**: "Root" feels cluttered to some. Harder to add a *second* Go service later without refactoring.

## Option B: The "Packages" Pattern (Workspaces)
This mimics the web app structure you are familiar with.

```
repo/
├── go.work             # Go Workspace file
├── packages/           # Shared libraries
│   ├── go-utils/       # Go lib
│   │   └── go.mod
│   └── py-common/      # Python lib
│       └── pyproject.toml
├── apps/
│   ├── cli/            # Go CLI
│   │   └── go.mod      # Depends on ../../packages/go-utils
│   └── api/            # Python Service
│       └── pyproject.toml # Depends on ../../packages/py-common
└── uv.lock             # Root lock for Python workspace?
```

### Analysis for TaskFlow

#### Go Workspaces (`go.work`)
- **Use Case**: When you have multiple *independent* Go modules that need to be released separately versioned, OR when developing 2+ microservices that share a local library.
- **For TaskFlow**: You have **1** Go artifact (the CLI). Splitting `internal/` into a separate `packages/go-utils` module adds overhead (managing multiple `go.mod` files) with zero benefit because nothing else consumes that utility code.

#### Python Workspaces (`uv` workspaces)
- **Use Case**: Sharing code between a Flask API and a separate Celery Worker, or a shared data-science library.
- **For TaskFlow**: You have **1** Python service. Creating a workspace structure now is premature optimization.

## How Cross-Language Dependencies Work
Since Go can't import Python, the "Glue" is the **Contracts**.

```
repo/
├── contracts/
│   └── proto/task.proto  # "The Truth"
└── gen/
    ├── go/task.pb.go     # Go representation
    └── py/task_pb2.py    # Python representation
```

- **The CLI** imports `gen/go`.
- **The API** imports `gen/py`.
- They never touch each other's source code.

## Recommendation: Stick to "Service Isolation" (Modified Pattern C)

If you prefer the `packages/` mental model, we can group things under `apps/` or `services/` to keep the root clean, but **avoid enabling workspace features** until you actually have >1 package per language.

**Compromise Layout (Clean Root):**
```
repo/
├── apps/
│   ├── cli/            # Go module root
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── go.mod
│   └── brain/          # Python service root
│       ├── src/
│       └── pyproject.toml
├── contracts/          # Shared Proto
└── Makefile
```
- **Pros**: Root is very clean. CLI and Brain are peers.
- **Cons**: You have to `cd apps/cli` to run go commands (or use Makefile).

**Verdict**: The **Compromise Layout** matches your web-app intuition (folders = apps) without the complexity of actual workspaces/multi-module config.
