# tskflwctl architecture

A local-first planning CLI over markdown+frontmatter. Design rationale lives in
`planning/research/2026-06-06-*` and `planning/epics/17-pm-go-cli.md`; this is
the one-screen orientation for contributors.

## The rule: CLI/TUI are primary adapters over a shared core; the filesystem is a secondary adapter

```
   primary adapters            core                    secondary adapter
  ┌──────────────┐      ┌──────────────────┐         ┌──────────────────┐
  │ cli (cobra)  │ ───▶ │ core.Service      │ ──port▶ │ store.FS         │
  │ tui (bubble) │      │  + domain (pure)  │         │ markdown+yaml    │
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
- **`internal/cli`** — a primary adapter: the cobra tree.
- **`internal/tui`** — the *second* primary adapter (shipped): a Bubble Tea
  browser calling the **same** `core.Service`, never the store/fs. See the TUI
  section below.
- **`internal/theme`** — dependency-free semantic tokens (status/bucket/priority
  → glyph + color), imported by **both** `cli/render` (→ ANSI) and `tui`
  (→ lipgloss), so "in-progress is a yellow ●" is decided in one place.
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
**multi-adapter system**: the CLI and a **Bubble Tea TUI both ship as primary
adapters over the same `core`**. That single fact answers most of the critique —
the layering exists so the TUI reuses the use-cases without duplicating logic,
not for hypothetical future flexibility. The specifics:

- **Cross-package exported types aren't "leakage."** `domain.FileProblem`,
  `core.EpicSummary`, `core.NewTaskParams`, `render.MoveResult` are the *contract
  between layers*. Everything lives under `internal/`, so "exported" means
  "visible to sibling packages in this binary," never to the outside world —
  exactly what a layered design needs.
- **`core.Store` earns its keep today, not speculatively.** The core's unit
  tests run against an in-memory `fakeStore` (`core/service_epic_test.go`), so
  rollup/validation logic is tested with no filesystem. That's a real second
  implementation now, plus the shipped TUI is a second primary adapter over the
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

## The TUI (`internal/tui`)
A Bubble Tea (Elm-architecture) browser, launched by `tskflwctl ui`. It is the
**second primary adapter**: every read goes through `core.Service` as a `tea.Cmd`
returning a custom `tea.Msg` — **never I/O in `Update`/`View`**, never the store.
Files split by concern:

- **`model.go`** — the root `Model` + the `Update` reducer and `View`. Owns the
  tab set, focus (list ⇄ detail), window size, and key routing.
- **`entity.go`** — the **entity registry**: tasks/epics/audits as `*entityTab`s,
  each owning its own `list.Model`, cursor, loaders, and list-scoped state
  (status view, sort, filter restore). Adding Projects/ADRs later is a new
  registry entry — no new keybindings or layout.
- **`commands.go` / `messages.go`** — the async load `tea.Cmd`s and the `tea.Msg`
  types they return (list loads, lazy detail loads, reload, errors).
- **`detail.go` / `find.go`** — the right pane (a `viewport`) + vim-like `/` `n`
  `N` find-in-body (matches tracked by wrapped-line index; highlight rebuilt from
  stripped text so it can't split an escape).
- **`item.go`** — per-entity `list.ItemDelegate`s (the glyph rows) and the
  `sortFields`/`FilterValue` each row exposes.
- **`sort.go` / `statusview.go` / `command.go`** — interactive sort, the unified
  status-view table (`:` words + `s`/`S` cycle derive from it), and the `:`
  command bar.
- **`watch.go`** — `fsnotify` live reload: a self-perpetuating listener `Cmd`
  feeds `fsEventMsg`; a generation-guarded `tea.Tick` debounce (200ms) coalesces
  save-storms into one reload of every loaded tab, cursor preserved by id.
- **`help.go`** — the `?` keybinding overlay (`helpSections` is the runtime
  source of truth for keys) composited over the body with `ansi.Cut`.
- **`style.go` / `keys.go`** — lipgloss styles (delegating to `theme`) and the
  `key.Binding` map.

**Layout discipline is load-bearing** (a clipped-top-border class of bug):
subtract the border frame before sizing children, guard `View` before the first
`WindowSizeMsg`, truncate (never wrap) anything fed to a `Join`, and clamp the
composed view to the terminal. `TestModel_ViewFitsTerminal` locks the invariant
(View height == terminal height; no line wider than the terminal). The full
checklist is in `planning/research/2026-06-10-tui-design-decisions.md`.

## Testing
Three layers for the CLI/core: pure domain/core units (incl. a `fakeStore` for
the core), store round-trips against `t.TempDir()`, and in-process CLI tests that
execute `NewRootCmd` with a captured buffer. The hand-rolled byte parsers in `store`
have fuzz targets (`store/fuzz_test.go`). The TUI is tested by **message
injection** (build the model, send `tea.Msg`s to `Update`, assert on state /
`View()` substrings) plus a few `x/teatest` full-program tests and the layout
invariant; fs-event behavior uses synthetic messages, not real `fsnotify` timing.
`just test` + `just lint`.

## Status (2026-06-11)
Substantially functional — the full create→update→move→lint loop runs without
the Python prototype:
- `init`, `completion` (command/flag/slug, status-aware), `lint` (+`--fix`/`--dry-run`)
- `task new|list|show|set|move|start|promote|demote|complete|defer|deprecate`
- `epic new|list|show`, `audit list|show|close|reopen|defer`
- `ui` — the Bubble Tea browser (epic 18): two-pane read-only browse of
  tasks/epics/audits, `:` jump, `/` filter, sort, status views, detail find, `?`
  help, and `fsnotify` live reload (S0–S3 shipped; mutations/glamour/cross-link
  are the remaining sprints).

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
