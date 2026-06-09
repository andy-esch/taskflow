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
  no cobra. The directory **is** the authoritative status (the read path always
  uses the folder); the frontmatter value is kept as `Task.Declared` only to
  detect drift. A *recognized* status that disagrees with its folder is
  "misfiled" — flagged by `lint` (and `lint --fix` realigns it), shown with a
  `⚠` in `task list`/`show`. A foreign/legacy status word is tolerated.
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

## Why these boundaries (and why not collapse them)
Reviews periodically suggest folding the packages together ("Go favors fewer
packages / concrete types"). That advice evaluates this **as a CLI**, but it is a
**multi-adapter system**: the CLI ships now and a **Bubble Tea TUI is planned as
a second primary adapter over the same `core`**. That single fact answers most
of the critique — the layering exists so the TUI reuses the use-cases without
duplicating logic, not for hypothetical future flexibility. The specifics:

- **Cross-package exported types aren't "leakage."** `domain.FileProblem`,
  `core.EpicSummary`, `core.NewTaskParams`, `render.MoveResult` are the *contract
  between layers*. Everything lives under `internal/`, so "exported" means
  "visible to sibling packages in this binary," never to the outside world —
  exactly what a layered design needs.
- **`core.Store` earns its keep today, not speculatively.** The core's unit
  tests run against an in-memory `fakeStore` (`core/service_epic_test.go`), so
  rollup/validation logic is tested with no filesystem. That's a real second
  implementation now, plus the planned TUI is a second primary adapter over the
  same core. (One known wart: `FixFrontmatter` sits awkwardly on the port — a
  candidate to split into a `Fixer` later.)
- **Frontmatter logic is already cohesive.** `frontmatter.go` (parse + surgical
  write), `fix.go` (text repair), `diagnose.go` (error diagnosis) are all one
  package (`store`), split into files by concern — idiomatic Go. `domain/
  validate.go` is *semantic field rules* (tier 1–5, priority enum), a domain
  concern, deliberately not coupled to the storage format.
- **`cli/render` is the one genuinely revisitable call.** It's cli-only (the TUI
  renders via Bubble Tea views, not these text/JSON formatters) and imports
  `core` for two view-models (a mild `cli→render→core` diamond). Keeping it a
  package buys isolation + the `render.` namespace; folding it into `cli` as
  `render.go` would also be fine. Left split for now; not dogma — collapse it if
  the boundary ever causes friction. (Note this is the *opposite* of dropping the
  core seam: render is presentation that the TUI replaces; `core` is logic the
  TUI reuses.)

## Testing
Three layers: pure domain/core units (incl. a `fakeStore` for the core), store
round-trips against `t.TempDir()`, and in-process CLI tests that execute
`NewRootCmd` with a captured buffer. The hand-rolled byte parsers in `store`
have fuzz targets (`store/fuzz_test.go`). `just test` + `just lint`.

## Status (2026-06-08)
Substantially functional — the full create→update→move→lint loop runs without
the Python prototype:
- `init`, `completion` (command/flag/slug, status-aware), `lint` (+`--fix`/`--dry-run`)
- `task new|list|show|set|move|start|promote|demote|complete|defer|deprecate`
- `epic new|list|show`, `audit list|show|close|reopen|defer`

Throughout: explicit noun-verb, semantic exit codes (`10` not-found · `11`
validation · `12` invalid-transition · `13` ambiguous · `14` conflict), atomic
writes (`writeFileAtomic` overwrite, `createFileAtomic` exclusive) + surgical
`yaml.v3` edits, `--json` everywhere (`schema_version`), resilient reads with
actionable frontmatter errors, agent safety annotations.

Remaining (see `planning/`): `adr`/`project` groups, audit finding-level
commands, reporting views (`stats`/`index`/`tags`), `track`, `schema --type
cli`, global `--dry-run`, advisory `flock`, structured JSON error envelope,
interactive `init` wizard. Out of scope by a long shot: MCP / semantic engine /
pgvector.
