---
date: 2026-06-06
topic: Go CLI foundation — layout, architecture, and patterns for the pm port
purpose: Define a rock-solid, idiomatic-Go foundation (code architecture, layout, patterns, testing) to build the pm CLI on, reusing the taskflow repo. Companion to the CLI-architecture research.
status: in-progress
related_tasks:
  - port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md
  - rethink-pm-command-hierarchy-pm-noun-verb-research-cli-best-practices.md
---

# Go CLI foundation: layout, architecture, patterns

**Goal.** Reuse the `../taskflow` repo, but stand the Go pm CLI on a
foundation that's rock-solid and idiomatic *before* porting commands.
Companion to `research/2026-06-06-pm-cli-architecture-and-go-port.md` (that
doc = hierarchy/framework/scope; this doc = *how the code is structured*).

> **Naming (decided 2026-06-06):** the Go binary is **`tskflwctl`**
> (kubectl-style), built in the reused **taskflow** repo/module
> (`github.com/andy-esch/taskflow`), entrypoint `cmd/tskflwctl/`. Below,
> **`pm`** always refers to the *Python prototype* (`desirelines-planning/bin/pm`)
> that `tskflwctl` ports from.

## Guiding principle: align with the org's pragmatic hexagonal

Don't invent a fresh style — match what the desirelines Go services already
do (and what `research/hexagonal-architecture-go-best-practices.md`
validated): **ports & adapters, but pragmatic** — implicit interfaces kept
near their consumer, minimal abstraction, composition root in `main.go`,
compile-time interface assertions (`var _ Iface = (*Impl)(nil)`),
`ports/portstest`-style test doubles. The services use
`cmd/ · adapters/<tech>/ · ports/ · internal/<domain>/ · config/ · pkg/`.

> "Keep interfaces close to where they're used; avoid over-abstraction;
> don't blindly copy Java/C#." — the org's own Go research.

## The core architectural decision

**CLI and the (later) TUI are both *primary adapters* over one shared core;
the markdown filesystem is a *secondary adapter* (a repository).**

```
        primary adapters                 core                 secondary adapter
   ┌───────────────────────┐      ┌──────────────┐        ┌────────────────────┐
   │  cli (cobra)          │ ───▶ │   Service    │ ──────▶│  fsstore           │
   │  tui (bubbletea, later)│      │  (use cases) │  port  │  (markdown+frontm.) │
   └───────────────────────┘      │  + domain    │        └────────────────────┘
                                   └──────────────┘
```

This is the single thing taskflow got wrong: its `ui` command wired
Bubble Tea straight to nothing (no core), so a CLI and a TUI would
duplicate logic. With a real core, `tskflwctl task list` and the TUI's task list
call the **same** `Service.ListTasks(...)` (`tskflwctl task list` and the
TUI list share one code path).

**Personal Project Goals (Modular Fun):**
Since this is a personal project for fun, the architecture must allow the
**CLI and TUI to evolve at their own pace**. By modularizing them as
independent primary adapters over a shared base library (the `core`), we
can sink as much time as desired into TUI polish without blocking CLI
utility or correctness.

Consequences:
- The **core has no cobra and no direct fs** — pure domain + use cases,
  unit-testable in isolation. It depends on a `TaskStore` *interface*
  (defined in the core, the consumer) that `fsstore` implements.
- Output/rendering is **not** in the core — commands get typed results and
  hand them to a renderer (human vs `--json`).

## Proposed layout (reuse taskflow repo, strip to the CLI fence)

```text
taskflow/                      # reuse ../taskflow (module github.com/andy-esch/taskflow)
├── cmd/
│   └── tskflwctl/main.go      # THIN: build deps (composition root), run root cmd, single os.Exit
├── internal/
│   ├── domain/                # entities + invariants, pure: Task, Epic, Audit, Status
│   │   ├── task.go            #   typed Status enum, validation, "status matches dir" rule
│   │   ├── epic.go
│   │   └── audit.go
│   ├── core/                  # use cases (the Service/App API the adapters call)
│   │   ├── service.go         #   List/Show/New/Set/Lint/Promote/Complete/... return typed results
│   │   └── store.go           #   TaskStore/EpicStore interfaces (defined here, near use)
│   ├── store/                 # secondary adapter: markdown+frontmatter on disk
│   │   ├── fsstore.go         #   implements core.TaskStore; var _ core.TaskStore = (*FS)(nil)
│   │   └── frontmatter.go     #   parse/serialize via yaml.v3; preserve body + key order
│   ├── cli/                   # primary adapter: cobra tree
│   │   ├── root.go            #   NewRootCmd(svc, cfg, out, errOut) — NO globals
│   │   ├── task.go / audit.go / epic.go / adr.go
│   │   └── render/            #   human (table/lipgloss) vs json; ANSI never in --json
│   ├── tui/                   # primary adapter (LATER): bubbletea over the same core
│   └── config/                # viper: flags>env>file>default; split-repo paths
├── testdata/                  # golden files + fixture task trees
├── go.mod · Justfile · .golangci.yml
```

Drop from taskflow: `services/` (the Python brain), `contracts/`+buf
(unless proto adopted — leaning no), `internal/index` JSON struct (defer to
the perf phase), and the direct `ui→tui` wiring (rebuild over the core).

## Foundation patterns (the "rock solid" part)

1. **DI via one deps struct, populated lazily; no global command state.**
   taskflow uses the common `var rootCmd` + `init()` global pattern — it's
   untestable. Instead bundle deps once: `cli.NewRootCmd(app *cli.App)`,
   where `App{ Svc *core.Service; Cfg *config.Config; Out, ErrOut io.Writer }`,
   and every subcommand constructor takes `*App` (one struct, not a 4-tuple).
   - **Lazy App Shell (resolves the flag-ordering chicken-and-egg):** deps
     depend on flags (`--chdir`/`--config`) that cobra parses *after* the
     tree is built. So pass an **empty `*App`** to the factories and
     **populate it in the root `PersistentPreRunE`** (runs post-parse,
     pre-`RunE`). `main.go` builds the tree + the empty App; `PersistentPreRunE`
     loads config, opens the store, constructs the service.
   - **Not** `cmd.Context()` for DI — context-as-DI is a Go anti-pattern
     (untyped `Value` lookups); reserve context for cancellation.
2. **All output through injected `io.Writer`** (`cmd.OutOrStdout()` /
   the injected `out`), never `fmt.Println`. This is what makes command
   tests possible (capture a `bytes.Buffer`).
3. **Render separated from logic.** Commands: parse flags → call `svc` →
   get a typed struct → `render.Human(w, x)` or `render.JSON(w, x)`.
   `--json` is a persistent root flag.
4. **Errors return; `main` exits once with semantic codes.** `RunE` returns
   wrapped errors (`%w`); domain sentinels (`ErrNotFound`,
   `ErrAmbiguousMatch`, `ErrInvalidTransition`, `ErrValidation`,
   `ErrLockConflict`). `main.go` maps each sentinel to a **semantic exit
   code** (10 not-found · 11 validation · 12 invalid-transition · 13
   ambiguous · 14 lock-conflict; 0 incl. idempotent no-op). In `--json`,
   emit a structured envelope on stderr (`{error_code, message,
   candidates?}`). Fixes pm's exit-0-with-error quirk. (See the spec's
   Agent-interaction contract.)
5. **Typed domain, not stringly-typed.** `Status` is a typed enum with
   `Parse`/`Valid`; the "status must equal directory" invariant lives in
   the domain/store, not scattered in commands.
   - *Re the "relocation breaks links" red flag:* cross-references are
     **slug `[[…]]` links** (location- and extension-independent) and
     **filenames are stable** across status moves (pm already moves files on
     every lifecycle change without renaming), so moving a task between
     status dirs does **not** break links. The resolver maps slug→file by
     scanning, not by path. (Slug uniqueness is already enforced.) So this
     red flag is a non-issue *for our link style* — but it does raise a real
     layout choice; see open Q.
6. **Frontmatter: surgical, preserving writes (re-revised 2026-06-06 — the
   review is right; I was too hasty last round).** The decisive point isn't
   comments, it's **unknown/custom fields**: pm keeps fields it doesn't know
   (`lint` only *advises* on them), so a static-struct round-trip would
   **silently delete** them. So:
   - **Split** the `---`-fenced frontmatter from the body with a tiny
     **zero-dep byte scanner** — **not** `go.abhg.dev/frontmatter`, which
     drags in the full Goldmark AST for what is a ~15-line scan. Body
     preserved verbatim.
   - **Read** into a typed struct for validation/logic.
   - **Write surgically** via stdlib **`yaml.v3` `yaml.Node`**: update only
     the keys being changed, re-encode. Preserves unknown fields, comments
     (yaml.v3 Node carries Head/Line comments), and key order. No `goccy`
     needed.
   - **Normalize** (canonical reorder, flag unknowns) **only** on explicit
     `lint --fix` — never silently on every write.
   - **Encapsulate** the Node-twiddling in `internal/store/frontmatter` so
     it never leaks into the store/core. *Supersedes both the original
     "goccy AST" and last round's "canonical struct" calls.*
7. **Planning repo + tracked repos (settled 2026-06-06):** one **planning
   repo** holds the *whole* `tasks/epics/projects/adrs/audits` tree for one
   product — standalone *or* a subdir of a code repo. **No cross-product
   registry** (one planning repo per product). It **tracks 1..N code repos**
   it plans for (`tracked_repos`) — e.g. desirelines-planning →
   [desirelines, desirelines-deploy]; metadata only, tasks cite code by
   path. The tool anchors by walking up for `.tskflwctl.toml` (git-style);
   no `source`/`context` noun. `init` writes config + scaffolds. Don't
   conflate the planning repo with the in-repo **project** initiative.
   **Walk-up termination (red flag):** discovery climbs for `.tskflwctl.toml`
   but stops at a `.git` dir / `$HOME` / filesystem root (and skips network
   mounts) — never an infinite climb. If none found: error "not a taskflow
   repo — run `tskflwctl init`".
8. **Pragmatic Concurrency:** Design the `Store` and `Service` to be
   thread-safe. Benchmarks confirm `filepath.WalkDir` or `fastwalk` can
   process ~1,000 files in <50ms, making a JSON cache unnecessary for
   local-only data.
   *(External reviewer comment: Benchmark data shows cache validation overhead (statting N files) for <1k entries is often higher than a fresh scan in Go, making the index a 'complexity debt' until N > 5k or network is involved.)*
   **Concurrent writers — advisory lock in v1 (revised 2026-06-06).** Atomic
   writes (#12) prevent torn reads, but **not lost-updates** when two
   mutations hit the same file. The user runs **parallel agent loops**, so
   that's a real pattern, not a single-user edge case → take an **advisory
   `flock` on `.tskflwctl.lock`** (repo root) around mutating commands, and
   surface `ERR_LOCK_CONFLICT` (exit 14). `flock` auto-releases on process
   exit, so the stale-lock worry is minimal — which flips my earlier
   "defer it" call. (`gofrs/flock`.)
9. **Future remote: local is the source of truth; sync is explicit + async
   (revised 2026-06-06 — answers the "sync network in the read path" red
   flag).** Do **not** put remote (GitHub Issues/Comments) calls behind a
   synchronous store read — that would wreck the sub-10ms read path and
   invite offline/sync conflicts. When/if remote lands (out of scope now):
   local markdown stays authoritative; a dedicated **`tskflwctl sync`**
   command reconciles to/from remote out-of-band. A read-time `HybridStore`
   composite is acceptable *only* if the remote half is a local mirror the
   `sync` command refreshes (never a live network call on read).
   **Prereqs for remote (red flags, defer):** (a) an immutable `id:` per
   task — the **slug is the stable *local* key** (filenames don't change on
   status moves), but a remote adapter would read a move as delete+create
   without a path-independent id; backfilling ids is a cheap one-time
   migration, so defer rather than burden every task now. (b) Cache
   invalidation must use **content checksums / git blob hashes, never
   `mtime`** (git checkout/pull bumps mtime → spurious full re-scans).
10. **Compile-time interface assertions** at each adapter (org pattern).
11. **Small interfaces at the consumer** (`TaskStore` in `core`, not a
   sprawling `ports/` package) — the org's explicit guidance.
12. **Atomic, durable writes (red flag).** Never `os.Create` over a live
   task file — a Ctrl-C/crash mid-write truncates it. Write to a temp file
   **in the same dir** → `f.Sync()` → close → `os.Rename` over the target
   (atomic on one filesystem; a status move is already a rename). Readers
   see either the old or new file, never a torn one. (Better than pm's
   non-atomic `write_text`.)

## Testing strategy (three layers + golden files)

- **Domain/core:** pure table-driven unit tests, no fs (the payoff of the
  core boundary).
- **Store:** round-trip against `t.TempDir()` with fixture trees in
  `testdata/`; assert parse/serialize and lifecycle moves.
- **CLI:** execute `NewRootCmd(...)` in-process with args + a
  `bytes.Buffer`, assert against **golden files** (`-update` flag). Faster
  and more precise than Python's subprocess tests.
- **Spec source:** translate `tests/test_pm.py` (62 cases) into these
  layers, group-by-group, red→green — the executable parity contract.

## Build / lint / distribution

- **golangci-lint** config mirroring the impl repo (they gate gosec,
  wrapcheck, nolintlint — I hit all three this session). Same bar = same
  muscle memory.
- **Justfile**: `build` (`go build -o bin/tskflwctl ./cmd/tskflwctl`),
  `test`, `lint` — mirror the impl repo's `go-test`/`go-lint` style.
- **Version** via `-ldflags -X` (taskflow doesn't yet).
- Single static binary `tskflwctl`; `./bin/tskflwctl` resolves to it.

## Borrow the good: taskflow asset disposition

| taskflow asset | Disposition |
|---|---|
| cobra+viper+bubbletea+lipgloss deps | **Keep** (settled stack) |
| Pattern-C bones (`cmd/`+`internal/`) | **Keep**, extend with domain/core/store/cli split |
| `cmd/taskflow/main.go` (thin, ctx) | **Keep** shape; rename → `cmd/tskflwctl`, add composition root |
| `internal/cli/root.go` (global `rootCmd`) | **Refactor** → `NewRootCmd(deps)`, no globals |
| `internal/tui/model.go` (stub) | **Keep** as the seed; rebuild over the core, later |
| `internal/index` JSON struct | **Defer** to the perf phase |
| `contracts/` proto + buf | **Drop** for now (Q in companion doc: structs vs proto) |
| `services/` Python brain | **Drop** (out of scope by a long shot) |
| `planning/research/*.md` | **Keep** as reference (this doc cites it) |

## Open / resolved questions

1. ✅ **Module/binary identity (resolved 2026-06-06):** reuse the taskflow
   repo, module `github.com/andy-esch/taskflow`; binary **`tskflwctl`**
   (kubectl-style), entrypoint `cmd/tskflwctl/`. No module rename needed.
2. **`domain` vs `core` split** vs a single `core` package — start merged
   and split only if it grows (avoid-over-abstraction). *(Lean: one `core`
   until size forces a split.)*
3. ✅ **Frontmatter (re-revised 2026-06-06):** **surgical, preserving
   writes** — zero-dep byte-scanner split, body verbatim, read into a struct
   for logic, **write via stdlib `yaml.v3` `yaml.Node` updating only changed
   keys** (preserves unknown fields + comments + order). Normalize only on
   `lint --fix`. No `go.abhg.dev/frontmatter`, no `goccy`. See pattern #6.
4. ✅ **Config (settled, pattern #7):** `.tskflwctl.toml` = `taskflow_root`
   + `tracked_repos`, written by `init`, discovered by walking up
   (git-style). No global registry, no `source` noun. (Only the exact
   filename is cosmetic.)
5. **Layout: status-as-directory vs flat `tasks/` + frontmatter status?**
   (Red-flag #1.) Current/Python = status-dirs (the `move` does a git mv;
   "status == dir" is a free consistency check; `ls tasks/next-up/` is
   glanceable). Flat-dir = no moves at all (status is pure frontmatter; the
   Go tool reads frontmatter anyway). *Lean: keep status-dirs* (consistency
   with the existing tree + Python pm + audits/routines; move-churn is a
   non-issue since pm already does it) — but it's a conscious data-model
   call. Links don't force it either way (slug-based; see pattern #5).

## Sources

- `research/hexagonal-architecture-go-best-practices.md` (org's validated
  Go architecture research) + the live `../desirelines` Go services
  (`packages/{dispatcher,apigateway}`: `cmd/ adapters/ ports/ internal/
  config/ pkg/`, `ports/portstest`, compile-time assertions).
- `../taskflow/` (skeleton + `planning/research/`).
- `research/2026-06-06-pm-cli-architecture-and-go-port.md` (companion).
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
  · [Three Dots Labs — Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/)
- Frontmatter: stdlib `gopkg.in/yaml.v3` `yaml.Node` for surgical writes +
  a hand-rolled byte-scanner to split the `---` block (no goldmark/goccy).
