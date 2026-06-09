# tskflwctl architecture

A local-first planning CLI over markdown+frontmatter. Design rationale lives in
`planning/research/2026-06-06-*` and `planning/epics/17-pm-go-cli.md`; this is
the one-screen orientation for contributors.

## The rule: CLI/TUI are primary adapters over a shared core; the filesystem is a secondary adapter

```
   primary adapters            core                    secondary adapter
  ┌──────────────┐      ┌──────────────────┐         ┌──────────────────┐
  │ cli (cobra)  │ ───▶ │ core.Service      │ ──port▶ │ store.FS         │
  │ tui (later)  │      │  + domain (pure)  │         │ markdown+yaml    │
  └──────────────┘      └──────────────────┘         └──────────────────┘
```

- **`internal/domain`** — pure entities + invariants (`Task`, `Status`). No fs,
  no cobra. The `Status` value *is* the directory name.
- **`internal/core`** — use cases (`Service`) + the ports it needs
  (`TaskStore`, defined here at the consumer). Pure; unit-testable without fs.
- **`internal/store`** — the secondary adapter: tasks as
  `<root>/tasks/<status>/<slug>.md`. Splits frontmatter with a zero-dep byte
  scanner; parses YAML with `go.yaml.in/yaml/v3`. `var _ core.TaskStore = (*FS)(nil)`.
- **`internal/cli`** — the primary adapter: the cobra tree. The future
  `internal/tui` is a *second* primary adapter calling the **same** core, so it
  never duplicates logic.
- **`internal/config`** — discovers the planning root (walk up for tasks/;
  terminates at a `.git`/root boundary).
- **`cmd/tskflwctl`** — thin; the sole composition root.

## Non-negotiable patterns
- **DI via one `*cli.App`**, populated in root `PersistentPreRunE` (the lazy
  shell — deps depend on flags). **No package globals**, **no `cmd.Context()`
  for DI**.
- **All output through injected `io.Writer`** (never `fmt.Println`) → commands
  are testable in-process (see `internal/cli/task_test.go`).
- **Render is separate from logic**: commands call the service, then
  `render.TasksHuman`/`TasksJSON`. `--json` is a global flag; JSON carries a
  semver `schema_version` and never emits ANSI.
- **The core never touches the fs or cobra.**

## Testing
Three layers: pure domain/core units, store round-trips against `t.TempDir()`,
and in-process CLI tests that execute `NewRootCmd` with a captured buffer.
`just go-test` + `just go-lint` (golangci-lint).

## Status (2026-06-08)
Substantially functional — the full create→update→move→lint loop runs without
the Python prototype:
- `init`, `completion` (command/flag/slug, status-aware), `lint` (+`--fix`/`--dry-run`)
- `task new|list|show|set|move|start|promote|demote|complete|defer|deprecate`
- `epic new|list|show`, `audit list|show|close|reopen|defer`

Throughout: explicit noun-verb, semantic exit codes (10–13), atomic +
surgical-`yaml.v3` writes, `--json` everywhere (`schema_version`), resilient
reads with actionable frontmatter errors, agent safety annotations.

Remaining (see `planning/`): `adr`/`project` groups, audit finding-level
commands, reporting views (`stats`/`index`/`tags`), `track`, `schema --type
cli`, global `--dry-run`, advisory `flock`, structured JSON error envelope,
interactive `init` wizard. Out of scope by a long shot: MCP / semantic engine /
pgvector.
