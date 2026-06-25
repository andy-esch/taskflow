---
status: planning
description: Explore a web UI sister to the TUI (tskflwctl serve first), module separation into core + bundled apps, and GitHub discoverability
priority: low
tags: [web, architecture]
created: "2026-06-13"
---

# Web companion apps over a shared core

**Goal.** Explore a web UI sister to the TUI (tskflwctl serve first), module separation into core + bundled apps, and GitHub discoverability

## Why this is its own epic

The web companion is a **third primary adapter** over the same `core.Service` the
CLI and TUI use — not a fork of the data. It earns its own epic because it adds a
runtime shape the terminal adapters don't have (a long-running `serve` process,
HTTP/JSON transport, auth) plus a module-separation question (core + bundled
apps). Its read side has a concrete first slice already in view — the materialized
projection / board (below).

## Read model / projection (convergence)

The generated **board** scoped in
`planning/research/2026-06-24-task-storage-model-files-logs-or-versioned-db.md`
is this epic's read side in embryo: **one `core` projection** of planning state
(`core.Summary()` is today's seed), rendered by every adapter — a committed
`BOARD.md`, the TUI board, a CLI `board`, and here as a JSON/HTTP read endpoint.
So `serve`'s read path is "expose the projection," not new logic; writes funnel
through `core.Service` with the version-aware OCC (see
`planning/research/2026-06-24-remote-planning-repos-backends-and-sync.md`).
Building the board (epic 23 / storage spike) is a down payment on this epic.

## Out of scope

- <explicitly excluded>
