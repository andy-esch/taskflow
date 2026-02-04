---
status: completed
epic: taskflow-v1-core
effort: 2-4 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [cli, python, architecture]
completed_at: 2026-02-04
---

# TaskFlow 01: CLI Skeleton (Go)

**Goal**: Initialize the Go CLI project structure for `taskflow`.

## Requirements
- Create a proper Go project structure (`cmd/taskflow`, `internal/`).
- Initialize `go.mod`.
- Implement the entry point using **Cobra**.
- Integrate **Bubble Tea** for a "Hello World" TUI.
- Setup configuration handling (Viper).

## Deliverables
- Functional `taskflow --help` command (compiled binary).
- Basic project layout in the new repo.
