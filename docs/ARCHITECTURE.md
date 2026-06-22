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

- **`internal/domain`** — entities + invariants (`Task`, `Status`). No fs, no
  cobra logic (the one pragmatic concession: `Task`/`Epic`/`Audit` carry a `Path`
  the store stamps, so callers can locate the source file). The directory **is**
  the authoritative status (the read path always
  uses the folder); the frontmatter value is kept as `Task.Declared` only to
  detect drift. A *recognized* status that disagrees with its folder is
  "misfiled" — flagged by `lint` (and `lint --fix` realigns it), shown with a
  `⚠` in `task list`/`show`. A foreign/legacy status word is tolerated.
  Per-entity metadata — the top-level dir, authoring fields, conventions, and
  body scaffold for `task`/`epic`/`audit` — lives in **one registry** (`entity.go`'s
  `Descriptor`); `SchemaKinds`/`AuthoringFields`/`Conventions`/`BodyTemplate` read
  that table instead of parallel `switch kind` blocks, so a kind's schema/scaffold
  surface is a registry entry, not a per-layer edit. Honest remaining fan-out for a
  new entity (e.g. the scaffolded `projects/`): a store scan (`scanDir` + a per-kind
  parse) plus per-entity render/TUI *display* delegates (the row + JSON formatters)
  — deliberately still per-entity. TUI *lifecycle* (the `a` menu + `:` verbs) is no
  longer in that list: it's registry-driven off each entity's transition table
  (M10), so an entity opts into close/move actions by declaring transitions, not by
  editing the reducer.
- **`internal/core`** — use cases (`Service`) + the ports it needs, defined here
  at the consumer. `Store` (composed of `TaskStore`/`EpicStore`/`AuditStore`) is
  the *use-case* port the `Service` depends on; the two fs/text operations that
  aren't use cases live in narrow sibling ports — `Fixer` (frontmatter repair)
  and `Layout` (watch-path layout) — so a second `Store` and the test fakes don't
  carry them. Pure; unit-testable without fs.
- **`internal/store`** — the secondary adapter: tasks as
  `<root>/tasks/<status>/<slug>.md`. Splits frontmatter with a zero-dep byte
  scanner; parses YAML with `go.yaml.in/yaml/v3`. One `*FS` satisfies all three
  ports (`var _ core.Store/Fixer/Layout = (*FS)(nil)`): the Service gets the
  use-case `Store`, the CLI's `lint --fix` and the TUI watcher get the narrow
  `Fixer`/`Layout` wired directly. It owns the *layout* knowledge — `WatchPaths()`
  hands the TUI watcher its dir set so the path convention isn't reconstructed
  outside the store.
- **`internal/cli`** — a primary adapter: the cobra tree.
- **`internal/tui`** — the *second* primary adapter (shipped): a Bubble Tea
  browser calling the **same** `core.Service`, never the store/fs. See the TUI
  section below.
- **`internal/theme`** — dependency-free semantic tokens (status/bucket/priority
  → glyph + color), imported by **both** `cli/render` (→ ANSI) and `tui`
  (→ lipgloss), so "in-progress is a yellow ●" is decided in one place.
- **`internal/config`** — discovers the planning root (walk up for tasks/;
  terminates at a `.git`/root boundary).
- **`cmd/tskflwctl`** — thin entrypoint; the command tree and DI wiring live in
  `internal/cli` (`root.go`), which it calls.

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
  same core. The port stays *use-case-only*: `FixFrontmatter` and `WatchPaths`
  (fs/text operations, not use cases) were split off into the narrow `Fixer` and
  `Layout` ports the adapters wire to the FS directly, so the `Store` the fakes
  implement carries no presentation-adjacent baggage.
- **Frontmatter logic is already cohesive.** `frontmatter.go` (parse + surgical
  write), `fix.go` (text repair), `diagnose.go` (error diagnosis) are all one
  package (`store`), split into files by concern — idiomatic Go. `domain/
  validate.go` is *semantic field rules* (tier 1–5, priority enum), a domain
  concern, deliberately not coupled to the storage format.
- **`cli/render` is the one genuinely revisitable call.** It's cli-only (the TUI
  renders via Bubble Tea views, not these text/JSON formatters) and imports `core`
  for its read-side view-models — today **five** (`Summary`, `StatusCount`,
  `EpicSummary`, `AuditFinding`, `LintResult`), and growing roughly one per entity
  as stats/index/tags land, so this is a real `cli→render→core` diamond, not the
  "two types" an earlier draft claimed. It stays justified because render is the
  *isolation seam the TUI doesn't touch*: these are core *results*, not store
  internals, and render is where presentation is allowed to know them. The
  trend-reversal if the count ever bites is the pattern `taskJSON`/`auditJSON`
  already use — map core results into render-owned DTOs at the call site rather
  than importing more core types. Keeping it a package buys isolation + the
  `render.` namespace; folding it into `cli` would also be fine. Not dogma —
  collapse it if the boundary ever causes friction. (Note this is the *opposite*
  of dropping the core seam: render is presentation that the TUI replaces; `core`
  is logic the TUI reuses.)

## The TUI (`internal/tui`)
A Bubble Tea (Elm-architecture) browser, launched by `tskflwctl ui`. It is the
**second primary adapter**: every read goes through `core.Service` as a `tea.Cmd`
returning a custom `tea.Msg` — **never I/O in `Update`/`View`**, never the store.
Files split by concern:

- **`model.go`** — the root `Model` + the `Update` reducer and `View`. Owns the
  tab set, focus (list ⇄ detail), window size, and key routing.
- **`entity.go`** — the **entity registry**: tasks/epics/audits as `*entityTab`s,
  each owning its own `list.Model`, cursor, loaders, list-scoped state (status
  view, sort, filter restore), and its **lifecycle table** (the transitions it
  offers + an `applyMove`). Read/browse is keybinding-free; lifecycle is declared
  here per entity (tasks by status via `Move`, audits by bucket via `MoveAudit`,
  epics none), so adding Projects/ADRs later is a new registry entry — including
  any `a`-menu / `:`-verb actions — not a reducer edit.
- **`commands.go` / `messages.go`** — the async load `tea.Cmd`s and the `tea.Msg`
  types they return (list loads, lazy detail loads, reload, errors).
- **`detail.go` / `find.go` / `glamour.go`** — the right pane (a `viewport`): the
  field block + a markdown body rendered two ways (raw / `glamour`, both cached so
  `R` toggles for free) + vim-like `/` `n` `N` find-in-body over *occurrences*
  (ANSI-aware highlight that preserves the line's other colors; unicode-fold-safe).
- **`item.go`** — per-entity `list.ItemDelegate`s (the glyph rows) and the
  `sortFields`/`FilterValue` each row exposes.
- **`sort.go` / `statusview.go` / `command.go` / `action.go` / `overlay.go`** —
  interactive sort (per-entity columns), the unified status-view table (`:` words +
  `s`/`S` cycle), the `:` command bar, the `a` lifecycle action menu, and the modal
  registry. The action menu and `:` verbs are **registry-driven**: both read the
  active tab's transition table + `applyMove`, so tasks move by status and audits by
  bucket (close/reopen/defer, in-TUI now) through one entity-agnostic path —
  `movedMsg.to` is a plain string the closure interprets. Overlays (help, action,
  follow) satisfy a small `modal` interface and live in an ordered stack the reducer
  loops; ForceQuit is handled once ahead of the loop, so a new overlay is one entry,
  not a new `handleKey` guard block + `bodyView` case.
- **`nav.go`** — S6 cross-link navigation: `f` follows structured references
  (a task's epic; an epic's tasks via a picker modal), `ctrl+o` pops the
  back-stack; hidden targets escalate the tasks view to `:all` rather than fail.
- **`watch.go`** — `fsnotify` live reload: a self-perpetuating listener `Cmd`
  feeds `fsEventMsg`; a generation-guarded `tea.Tick` debounce (200ms) coalesces
  save-storms into one reload of every loaded tab, cursor preserved by id. The
  watched dir set comes from the `core.Layout` port (`WatchPaths()`, the FS
  injected by the CLI), not from a root the TUI reconstructs — layout knowledge
  stays in the store.
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
The CLI also has **golden snapshots** of the byte-stable machine contract (the
`--json` envelopes, `csv`, and `schema --json-schema`) under
`internal/cli/testdata/golden/`, run in-process against the committed
`testdata/planning/` fixture; regenerate them with `go test ./internal/cli
-update` (the `-update` flag is cli-package-scoped, so target that package, not
`./...`). The single subprocess smoke layer (real binary, exit codes, lifecycle)
lives in `cmd/tskflwctl/main_test.go`. `just test` + `just lint`.

## Status (2026-06-11)
Substantially functional — the full create→update→move→lint loop runs without
the Python prototype:
- `init`, `completion` (command/flag/slug, status-aware), `lint` (+`--fix`/`--dry-run`)
- `task new|list|show|set|edit|append|move|start|promote|demote|complete|defer|deprecate`
- `epic new|list|show`, `audit list|show|findings|lint|close|reopen|defer`
- `ui` — the Bubble Tea browser (epic 18): two-pane read-only browse of
  tasks/epics/audits, `:` jump, `/` filter, sort, status views, detail find, `?`
  help, `fsnotify` live reload, lifecycle mutations (`a` menu + `:` verbs), and
  glamour markdown with an `R` raw/pretty toggle (S0–S5 shipped; cross-link is the
  remaining sprint).

Throughout: explicit noun-verb, semantic exit codes (`10` not-found · `11`
validation · `13` ambiguous · `14` conflict), atomic
writes (`writeFileAtomic` overwrite, `createFileAtomic` exclusive) + surgical
`yaml.v3` edits, `--json` everywhere (`schema_version`), resilient reads with
actionable frontmatter errors, agent safety annotations.

Remaining (see `planning/`): `adr`/`project` groups, the audit finding-*write*
surface (`audit finding --status`/`sync`; the read surface — `audit findings`
query + `audit lint` — shipped), reporting views (`stats`/`index`/`tags`),
`track`, `schema --type cli`, advisory `flock`, interactive `init` wizard. Out of
scope by a long shot: MCP / semantic engine / pgvector.
