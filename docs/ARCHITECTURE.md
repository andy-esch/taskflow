# tskflwctl architecture

A local-first planning CLI over markdown+frontmatter. Design rationale lives in
[`planning/research/`](../planning/research/) (`tskflwctl research list` to browse)
and [`planning/epics/17-pm-go-cli.md`](../planning/epics/17-pm-go-cli.md); this is
the one-screen orientation for contributors.

## The rule: CLI/TUI are primary adapters over a shared core; the filesystem is a secondary adapter

```
   primary adapters            core                    secondary adapter
  ┌──────────────┐      ┌──────────────────┐         ┌──────────────────┐
  │ cli (cobra)  │ ───▶ │ core.Service      │ ──port▶ │ store.FS         │
  │ tui (bubble) │      │  + domain (pure)  │         │ markdown+yaml    │
  └──────────────┘      └──────────────────┘         └──────────────────┘
```

## Current package map and dependency direction

Reviewed against the production import graph on 2026-08-22 (`go list` over
`./internal/...`). The diagram above is the rule of thumb; these are the actual
packages that implement it:

| Role | Packages | Current inward dependencies |
| --- | --- | --- |
| Foundations | `internal/id`, `internal/tomledit`, `internal/editor`, `internal/listfilter` | Standard library or their focused third-party primitive only |
| Domain | `internal/domain` | `internal/id` |
| Application | `internal/core` | `internal/domain`, `internal/id` |
| Neutral machine contract | `internal/wire` | `internal/core`, `internal/domain` |
| Planning/config secondary adapters | `internal/store`, `internal/config`, `internal/userconfig`, `internal/spacehealth`, `internal/configstore`, `internal/spacestore` | Consumer-owned `core` ports and/or the narrower domain/config adapters they compose |
| Primary adapters | `internal/cli`, `internal/tui`, `internal/configui` | `core`/`domain` values, `wire`, and presentation utilities; only the CLI composition root constructs secondary adapters |
| Presentation utilities | `internal/theme`, `internal/design`, `internal/progressbar`, `internal/themepreview`, `internal/cli/render`, `internal/cli/prompt` | Semantic values and UI libraries, never planning/config persistence |
| Test/tool support | `internal/testutil`, `internal/tools/*` | Concrete dependencies appropriate to a test fixture or one-off executable; not runtime layers |

The important direct adapter graph is concrete rather than symmetrical:

```text
domain      -> id
core        -> domain, id
wire        -> core, domain
store       -> core, domain, id
config      -> domain, id, tomledit
userconfig  -> tomledit
spacehealth -> config, domain, userconfig
configstore -> core, config, userconfig
spacestore  -> config, core, domain, spacehealth, store, userconfig
configui    -> core + presentation utilities
tui         -> core, domain, configui + presentation/process utilities
cli/render  -> core, domain, wire + presentation utilities
cli         -> composition of the application, primary adapters, and secondary adapters
```

### Enforced fitness rules

`.golangci.yml` makes the stable part of that direction executable for production
files:

- `internal/domain` may import only `internal/id` from this repository.
- `internal/core` may import only `internal/domain` and `internal/id` from this
  repository. Ports therefore stay with their application consumer; an adapter
  implements them without becoming a core dependency.
- `internal/wire` may import only `internal/core` and `internal/domain`, keeping the
  machine contract neutral rather than tied to Cobra, Bubble Tea, or filesystem TOML.
- `internal/tui`, `internal/configui`, `internal/cli/render`, and
  `internal/cli/prompt` cannot import the planning/config secondary adapters or sibling
  primary adapters. The one deliberate primary-adapter composition is the full TUI
  embedding the focused `configui` model.
- `internal/config` cannot import home-scoped `internal/userconfig`; cwd discovery
  remains independent of a user's registry and preferences.

Tests are excluded from the production adapter rule because UI integration tests
deliberately construct `store.FS` instances under `t.TempDir()`. That is test
composition, not runtime dependency direction.

### Composition-root exception and current direct adapter imports

`internal/cli/*.go` is deliberately not constrained like its `render` and `prompt`
subpackages. It is today's composition root: `NewRootCmd` constructs
`configstore.FS`/`spacestore.FS`/`workspacestore.FS`, repo resolution constructs `store.FS`, and the `ui`
commands launch `tui`/`configui`. Forbidding those imports would move wiring without
improving the boundary.

The 2026-08-21 audit of direct primary-to-secondary edges classified the remaining
ones as follows:

| Edge | Classification | Disposition |
| --- | --- | --- |
| `cli -> configstore`, `cli -> spacestore`, `cli -> workspacestore`, `cli -> store` construction | Composition root | Intentional; secondary adapters are injected into consumer-owned core ports. |
| `cli -> store` through `Fixer`, `Linter`, `Layout`, and completion | Narrow fs/text adapter capability | Intentional today; these are not planning use cases. The broader reusable workspace decision is tracked separately. |
| `cli -> config` for discovery/init/maintenance | Adapter orchestration | Atlas workspace opening now uses `core.WorkspaceService`; `ui` additionally reads `config.ErrNoConfig` to tell an ordinary discovery miss from a broken config. The deferred [`reusable-workspace-discovery-seam`](../planning/tasks/6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md) retains only init/doctor/fix work that still lacks another consumer. |
| `cli -> userconfig` for initial presentation-preference loading | Adapter orchestration | Intentional today; registry catalog, selection, mutations, completion, and diagnosis all use `core.SpaceRegistryService`. |
| `tui -> configui` | Focused primary-adapter composition | Intentional: the full TUI embeds the same configuration editor launched by `config edit`. |
| `tui -> editor` / `os/exec` | Narrow process/terminal capability | Intentional; planning data still flows only through `core.Service`. |

No new architecture task is needed for these edges: the two material seams already
have explicit trigger-scoped work, and the remaining edges are composition or narrow
adapter capabilities rather than leaked persistence.

- **`internal/domain`** — entities + invariants (`Task`, `Status`). No fs, no
  cobra logic (the one pragmatic concession: `Task`/`Thread`/`Epic`/`Audit`/`Research` carry a
  `Path` the store stamps, so callers can locate the source file). Frontmatter **is** the
  authoritative status/bucket (ADR-0003 §4): tasks, audits, and research are stored
  **flat and id-led** — `tasks/<id>-<slug>.md`, `audits/<id>-<slug>.md`,
  `research/<id>-<slug>.md`, `threads/<id>-<slug>.md` — with no status/bucket directory to mirror or drift against
  (epics keep `NN-<slug>`; research has no status at all, so nothing to mirror). A file whose
  frontmatter status is missing or unrecognized is **listed but flagged** by `lint`
  (`StatusFellBack`; shown with a `⚠` in `task list`/`show`), never moved or dropped;
  a non-id-led `.md` in a scanned dir is a loud `FileProblem` ("move it to `meta/`")
  — except `README.md`, silently carved. `meta/` is the sanctioned home for
  non-entity files.
  Per-entity metadata — the top-level dir, authoring fields, conventions, and
  body scaffold for `task`/`thread`/`epic`/`audit`/`research` — lives in **one registry** (`entity.go`'s
  `Descriptor`); `SchemaKinds`/`AuthoringFields`/`Conventions`/`BodyTemplate` read
  that table instead of parallel `switch kind` blocks, so a kind's schema/scaffold
  surface is a registry entry, not a per-layer edit. Honest remaining fan-out for a
  new entity, after the descriptor + the generic
  seams: a typed `domain` struct + a `parse*`, thin `*Store` port methods (the scan
  is generic `scanDir[T]`, resolution generic `resolveID`), its `core.Service` use
  cases, a cli command, and render + TUI *display* delegates (a Human/JSON formatter
  — column layout is the generic `Column[T]`/`WriteTablePlain` — plus an `entityTab`
  entry + row delegate). That residual is the cost of a **typed** domain whose five
  entities have genuinely different shapes (tasks: status/tier/priority; epics:
  rollups; audits: findings/buckets; research: dated, taggable, no lifecycle; Threads:
  membership plus a derived DAG view); the generics remove the *mechanics*, not the
  per-entity shape. Threads are intentionally different: persisted
  initiative metadata plus a sorted task-ID membership set, with all DAG state derived
  at read time. What IS collapsed: the metadata fan-out into the descriptor
  (M1), and TUI *lifecycle* (the `a` menu + `:` verbs) into each entity's transition
  table (M10), so an entity opts into close/move actions by declaring transitions,
  not by editing the reducer. A further data-driven persistence/render collapse
  isn't pursued — for five heterogeneous entities it trades clarity for machinery.
- **`internal/core`** — use cases (`Service`) + the ports it needs, defined here
  at the consumer. `Store` (composed of
  `TaskStore`/`EpicStore`/`AuditStore`/`ResearchStore`) is the *use-case* port the
  `Service` depends on; the three fs/text operations that aren't use cases live in
  narrow sibling ports — `Fixer` (frontmatter repair), `Linter` (link integrity),
  and `Layout` (watch paths) — so a second `Store` and the test fakes don't carry
  them. `SpaceRegistryService` owns the repo-independent catalog, grouping, explicit
  selection, label policy, and mutation receipts through `SpaceRegistryStore`.
  `SpaceOverviewService` composes that catalog with `SpaceOverviewStore`, whose only job
  is opening the narrow read-only `SummaryStore` needed for a dashboard scan, then selects
  one healthy entry point per identity. Roles and states are typed core vocabulary and
  entry health is derived from state, so an adapter cannot claim that a missing checkout
  is healthy.
  `WorkspaceService` is the separate local-tree opening boundary used when a primary
  adapter needs a complete entity service and watcher layout for an explicit start path.
  Its `WorkspaceStore` port returns neutral capabilities; registry labels are carried as
  presentation context and never influence discovery. `WorkspaceSource` carries
  `TaskGraphSource` and `ThreadStore` independently from the complete entity `Store`, so a
  split adapter can supply Thread projections without making both reads methods on one concrete
  value. `WorkspaceService` intentionally still requires the complete `Store` and `Layout` needed
  by the local TUI; a read-only primary adapter composes `Service` directly from the narrow ports.
  Complete adapters remain source-compatible because `Service` defaults both read capabilities
  from their aggregate store.
  `TaskGraph` is an immutable read projection over one repository scan. It owns graph
  health (`healthy`/`degraded`/`broken`), SCC-based cycle attribution, derived lifecycle
  role and gate state, sound completion, topology, downstream impact, and separately
  named causal-blocker and action-frontier projections. It also resolves ordinary task
  references inside that immutable snapshot, so dependency planners never pre-resolve a
  slug through persistence and carry a stale choice into the guarded callback. Service
  dependency use cases emit taskflow-owned, adapter-neutral receipts for edge set
  operations and the repository-wide legacy migration; typed failures retain any sound
  durable prefix for explicit recovery. The analyzer uses only taskflow types and owned
  deterministic algorithms; a graph package cannot leak into domain, persistence, or
  wire contracts. Exactly resolved legacy edges participate in diagnostic traversal and
  derived gates before migration, so a degraded snapshot cannot issue a false all-clear;
  present-but-empty legacy keys still keep health degraded until migration removes them.
  Eligibility is read from the queried task's explicit derived state, never inferred from
  an empty blocker list. Both queued (`next-up`) and candidate (`ready-to-start`) work are
  eligible when their gate is clear in a healthy snapshot; readiness remains a handoff signal,
  not a second authorization gate. Task listings, board, blockers, downstream queries, and Thread
  projections all consume the narrow `TaskGraphSource`; `task list --unblocked` fails closed unless
  its snapshot is healthy.
  `ThreadView` is a pure projection over that same immutable graph: it distinguishes
  members from direct external gates, reports nominal `done/total` and sound
  `drained/total`, separates repository graph health from Thread-local projection
  health, and emits a frontier of clear-gated `next-up` and `ready-to-start` members only
  when the combined evidence is healthy. Completed inconsistency carries stable reason codes;
  list projections hoist global graph problems rather than repeating them per Thread. Guarded
  Thread creation resolves initial members and validates the global task/Thread identity
  namespace inside the repository critical section; it can create only `unstarted`. Existing
  Thread membership and lifecycle plans are likewise pure validations over one immutable
  task/Thread snapshot. Task lifecycle impact compares every Thread projection before and after
  the task transition, including indirect member and external-gate effects, without making
  Thread documents own task state. Thread and dependency read use cases consume the narrow,
  independently injectable `TaskGraphSource`; Thread reads combine it with `ThreadStore`
  rather than relying on a concrete aggregate adapter. Joined reads fetch Thread data first and
  tasks second; paired adapters must ensure the later task snapshot is no older than the Thread
  snapshot, or coordinate a compatible snapshot themselves. These point-in-time diagnostics never
  authorize mutation. `ThreadGraphProjection` extends this boundary with stable-ID-ordered raw
  member and immediate-external-gate nodes, every prerequisite-to-dependent edge induced by that
  bounded node set, member-only explanatory waves, the complete `ThreadView`, and an explicit
  topology-completeness verdict. Waves contract ordering paths through included gates without
  treating those gates as Thread-owned work. CLI, TUI, and future web adapters consume that
  projection; `thread plan` presents its waves and marked gates, while `thread graph` exports it as
  Mermaid or DOT. Both commands return the same renderer-neutral projection under `--json`.
  `internal/graphfmt` owns pure Mermaid/DOT escaping and formatting without UI dependencies, and no
  Cobra, Bubble Tea, HTTP, filesystem, or graph-library type enters core or wire contracts. The
  source returns a neutral `TaskGraphRead`: unreadable records carry optional stable identity
  directly, while filesystem adapters translate `FileProblem` paths once at their boundary. The
  analyzer never parses an opaque source location for identity.
  Per-space failures remain data in the projection; the CLI renders the complete sweep
  before applying its partial-failure exit policy. Pure; unit-testable without fs.
- **`internal/store`** — the secondary adapter: tasks as
  `<root>/tasks/<id>-<slug>.md` (flat, id-led). Splits frontmatter with a zero-dep byte
  scanner; parses YAML with `go.yaml.in/yaml/v3`. One `*FS` satisfies the entity
  `Store`, the narrow `Fixer`/`Linter`/`Layout` ports, and the guarded graph and
  lifecycle mutation capabilities. Its `ReadTaskGraph` adapter translates file diagnostics into
  neutral record identity in the same task scan. The Service gets the use-case ports; CLI lint and
  the TUI watcher get their narrower capabilities wired directly. It owns the *layout*
  knowledge — `WatchPaths()`
  hands the TUI watcher its dir set so the path convention isn't reconstructed
  outside the store. Task dependency fields are graph-owned: generic create/set/edit
  paths cannot introduce a semantic delta, and text-level lint repair skips a would-be
  dependency normalization instead of manufacturing unchecked edges.
  `ThreadStore` is a separate narrow read port. `ThreadCreationMutationStore` owns guarded
  unstarted creation, while `ThreadMutationStore` owns atomic existing-document membership
  and lifecycle changes. Its lock-free update materializer is surgical: it changes only
  membership/status/timestamps and preserves unknown fields, key order, body, and comment content;
  the shared YAML editor may normalize inline-comment spacing.
  Thread files own metadata and membership only; task files remain the sole source of dependency
  edges. Task creation and Thread creation check one cross-kind stable-ID namespace under the
  same canonical-root guard. The legacy
  `projects/` scaffold is no longer created; only an empty directory or a lone regular
  `.gitkeep` is eligible for automatic retirement, and other content is preserved.
  `TaskGraphMutationStore` is the control-inverted write capability: `FS` takes the
  repository guard, loads the canonical strict snapshot, invokes a pure core planner over
  taskflow-owned values, asks core to validate the complete plan and every recovery prefix,
  applies surgical task dependency writes through lock-free internal helpers, and releases
  the guard. The planner callback is exclusive at canonical-planning-root scope: any Store
  call through the same or another `FS` during that brief phase fails with `ErrConflict`
  instead of escaping the snapshot or self-deadlocking. This intentionally includes an
  unrelated concurrent caller; a future long-lived adapter may retry after the mutation.
  Multi-file plans are deterministic and resumable. Planner-provided write order is
  semantic recovery data (stable-ID sorting is not generally prefix-safe); each replacement
  is atomic, every supplied prefix is validated as non-broken before the first write, and an
  error result identifies the durable applied prefix. A private byte-version on each scanned
  task supports a whole-snapshot CAS before application plus a per-target CAS immediately
  before each replacement; planner-facing task projections never expose those tokens.
  Dependency writes stamp `updated_at` from the caller-injected clock only when graph-owned
  fields actually change. A dry run holds the same exclusive guard for an authoritative
  preview but, because it writes nothing, makes no CAS durability claim about later apply.
  `TaskLifecycleMutationStore` is the lifecycle sibling over that same canonical-root
  guard. It authorizes a typed transition against one strict graph snapshot, materializes
  status and timestamps, verifies the whole snapshot plus the target version, and writes
  before releasing the guard. `task start`, generic move, defer, the TUI, and guarded
  create-and-start all enter through this capability. Dependency-gate and
  acceptance-criteria overrides are distinct core values even though the CLI spells both
  contextually as `--force`; generic set/edit paths cannot change status, and ordinary
  creation accepts only non-start states. Lifecycle results distinguish a durable commit
  from pre-commit failure. If repository-guard cleanup fails after the atomic replacement,
  the Service returns a typed failure carrying the committed receipt, never auto-retries it
  (including when cleanup wraps `ErrConflict`), and adapters tell the caller to inspect the
  current task state. The TUI reloads on that outcome instead of leaving a stale view.
  Task lifecycle authorization also scans and validates every Thread under the guard. Its
  receipt compares all before/after Thread projections, and its whole-snapshot CAS covers the
  Thread set so a concurrent membership edit cannot make the attribution stale.
  `ThreadMutationStore` follows the same transaction shape for membership and lifecycle:
  strict graph/Thread snapshot, pure plan, surgical materialization, whole task+Thread CAS,
  immediate target CAS, and one atomic Thread replacement. Cooperating dependency, task
  lifecycle, and Thread writers therefore serialize and re-authorize from fresh state. Raw
  editors do not join the advisory lock; the two CAS checks detect their changes where possible
  but cannot make a cross-process guarantee across the final verify-to-rename window.
  `ThreadApplyMutationStore` is the compound sibling for a generated bulk-link plan. The CLI
  injects configuration rediscovery into `FS`, allowing apply to verify both the physical planning
  root and durable repository ID while the canonical-root guard is held. One strict snapshot feeds
  the pure additive planner; lock-free dependency materialization runs first, the no-clobber Thread
  create runs last, and each atomic task replacement advances a structured operation receipt. An
  interrupted prefix is therefore graph-valid and resumable from the same durable plan without
  nesting `TaskGraphMutationStore` or `ThreadCreationMutationStore`. A raw edit before the first
  write fails the whole-source CAS; a target edit after a durable prefix fails the immediate CAS and
  leaves that prefix visible in the receipt. Apply re-reads identity, the graph, and Threads after
  every changed dependency prefix—even when the planned Thread already exists—so an out-of-band
  edit that undoes a repaired edge cannot produce a false complete receipt.
  The authoring manifest's `member: false` is a compose-time graph-context declaration, not a
  persisted Thread role. Such a task must be transitively upstream of a member and may be needed to
  express a multi-hop dependency chain. Runtime Thread projections reserve **external gate** for a
  direct prerequisite outside the membership set; deeper context remains discoverable through
  causal blocker queries without entering Thread rollups or the default boundary projection.
  The shared graph-plan validator recognizes canonical edge-only supersets: every physical prefix
  is then an edge-subset of the validated final graph, so it skips redundant per-prefix graph
  rebuilds. Removal and legacy-migration plans retain full prefix validation.
  Concurrency is **version-CAS** (epic 24): every write, just
  before committing, re-resolves the file by its **id** and re-hashes it
  against the content read at the start of the op (`verifyUnchanged` in `cas.go` — a
  strong whole-file SHA-256 computed on read, **never stored**), so a concurrent
  in-place edit is `ErrConflict` (exit 14). The repository guard combines a
  canonical-root in-process mutex with root-directory `flock` on Unix. The supported
  release matrix is macOS and Linux; Windows and other non-Unix source builds reject
  repository mutation explicitly until they have a runtime-tested, shared-repository lock
  identity. This serializes verify→write for cooperating writers so that CAS is
  *atomic* — without it two writers both pass their verify before either renames and the
  later silently clobbers the earlier. Raw editors do not honor the advisory lock; the
  immediate content check narrows but cannot eliminate their verify→rename race.
  The token is **internal**: scriptable mutations auto-retry pre-commit conflicts in
  `core.Service` (bounded + jittered, so agents don't reimplement the loop), the human `edit` surfaces the conflict
  (no retry, and the lock is held only for the write, never the editor session), and creates map
  the empty precondition onto `createFileAtomic`'s `O_EXCL`. Exposing it over HTTP
  (`If-Match`) is the web adapter's job (epic 19), not the FS store's.
- **`internal/cli`** — a primary adapter: the cobra tree.
- **`internal/tui`** — the *second* primary adapter (shipped): a Bubble Tea
  browser calling the **same** `core.Service`, never the store/fs. Its Config/About
  overlay embeds `internal/configui` and calls `core.ConfigurationService`, so it
  also shares the configuration application seam with Cobra. See the TUI section below.
- **`internal/theme`** — presentation-framework-free semantic tokens (status/bucket/priority
  → glyph + a `Color` enum), imported by **both** `cli/render` (→ SGR: the theme's
  truecolor hue, or a 16-color slot fallback) and `tui` (→ lipgloss), so
  "in-progress is a yellow ●" is decided in one place.
- **`internal/design`** — the concrete-color layer one level below `theme`: a
  `Palette` (the semantic enum → truecolor + a 16-color ANSI slot, plus chrome
  tokens — accent/borders/match/gradient) and a named-`Theme` registry (the neon
  default). Imports `theme` (keys its palette by the enum); `theme` must **never**
  import `design`. Consumed by `cli/render`, `tui`, `cli/prompt`, and
  `progressbar` — the one place a *hue* is chosen, as `theme` is for a *glyph*.
- **`internal/themepreview`** — the canonical ordered projection of a design palette
  into semantic swatches and a sample gradient. CLI theme preview and the live config
  editor own different layouts but review the same tokens and renderer.
- **`internal/config`** — discovers the planning root (walk up for tasks/;
  terminates at a `.git`/root boundary). **Repo-scoped**, and deliberately does not
  import `userconfig`. It also owns cross-repo *identity and placement*: `init`
  scaffolds a tree (at the root or a `taskflow_root` subdir) and mints a **durable id**
  into the repo's committed config; a `planning_repo` pointer that records
  `planning_repo_id` **verifies it after resolving**, so a mismatch — or a target with
  no id — is `ErrConflict` rather than a silent bind to whatever sat at that relative
  path. Which side of the repo boundary a config path points at decides what it resolves
  against: `taskflow_root` is in-tree and anchors to the config file's own dir, while
  `planning_repo`/`tracked_repos` point outward and anchor to the **canonical checkout**,
  so a committed relative path means the same thing from every git worktree
  (`worktree.go`).
  `init` only bootstraps a new topology. Its typed `InitResult` reports whether this call
  created the config plus the checkout directory and durable planning id; the CLI may then
  ask `SpaceRegistryService.RegisterInitialized` to perform a separate, best-effort home
  registration without importing `userconfig` into repo-scoped config. Safe upgrades to
  existing configs live in
  the explicit, idempotent `config.Migrate` path: direct `id` and pointer
  `planning_repo_id` backfills are surgical, atomic, and dry-run from the same result
  model used by apply output. Migration, typed preferences, and tracked-repo linkbacks
  share a config-directory Unix `flock` and re-read inside it, preventing cooperating
  writers from atomically replacing one another's unrelated edits.
- **`internal/userconfig`** — the *user*-scoped (home) tier: terminal preferences
  that belong to a person rather than a repo (`[theme]`, `[pager]`) plus the separate
  advisory `spaces.toml` registry, read from
  `$TSKFLW_CONFIG_HOME` / `$XDG_CONFIG_HOME/tskflwctl` / `~/.config/tskflwctl`
  (**not** `os.UserConfigDir()`, which is `~/Library/Application Support` on darwin).
  Precedence is flag > env > repo config > **user config** > default, merged
  **field-by-field** — a nil `Pager.Enabled` means "unset here, defer down", which is
  why it is a `*bool` at both tiers. Loaded in `setStyle` (not `resolve`) because
  `init`/`doctor`/`completion` skip discovery yet still need a theme. A missing file
  is normal; a malformed one **warns and degrades** rather than failing — the
  deliberate opposite of the repo config, where a bad marker is fatal because guessing
  there would fork the data. Preference and registry mutations are atomic, surgical TOML
  edits and take one directory-level Unix `flock`, so concurrent cooperating writers cannot lose changes;
  comments and unknown keys survive. `config` never imports this package, so home-scope data
  cannot influence planning-root discovery — a layering rule, kept honest by a
  `depguard` rule in `.golangci.yml` rather than by memory. (The compiler alone would
  not catch it: nothing structurally prevents the import, so `just lint` is what makes
  the claim true.)
- **`internal/configstore`** — the composite filesystem secondary adapter for the
  consumer-owned `core.ConfigurationStore` port. It combines repo discovery/migration,
  user preferences, and link diagnosis into neutral core DTOs. `ConfigurationService`
  composes registry diagnosis from `SpaceRegistryService`, keeping the repo-scoped port
  independent of the home registry; neither Cobra nor Bubble Tea imports those storage
  packages to implement a config use case.
- **`internal/configui`** — one focused Bubble Tea primary adapter, runnable by
  `config edit` or embedded in `tskflwctl ui`. It exposes only typed presentation
  preferences, defaults writes to user scope, requires an explicit repository scope,
  and leaves IDs/topology/registry state read-only. Its asynchronous commands call
  `core.ConfigurationService`; it never edits TOML.
- **`internal/spacehealth`** — the read-only projection over `userconfig.Space` plus
  repo-scoped `config.Discover`: one typed diagnosis (`ok`, `empty`, `missing`,
  `not-a-repo`, `unreadable`, `mismatch`), derived role (`direct`/`pointer`/`unknown`),
  and remedy per registered checkout. It is a storage/discovery helper of `spacestore`,
  not an application contract; primary adapters consume the resulting core projection.
  It never repairs, removes, groups, or adds relationship metadata to registry data.
- **`internal/spacestore`** — the composite filesystem secondary adapter for
  `core.SpaceRegistryStore` and the narrower `core.SpaceOverviewStore` tree-opening port.
  It is the only layer that translates `userconfig`/`spacehealth` records into neutral
  core entries, performs planning-checkout preparation, classifies registry persistence
  errors, and opens a Markdown summary reader. `core.SpaceRegistryService` owns catalog
  grouping, explicit selection, label policy, ordinary mutation receipts, and the narrow
  already-initialized-checkout registration used by `init`; CLI, completion,
  doctor, `status --all`, and future primary adapters consume that same boundary.
- **`internal/workspacestore`** — the filesystem secondary adapter for
  `core.WorkspaceStore`. It translates one explicit local start directory through
  repo-scoped `config.Discover`, constructs the concrete Markdown store once, and exposes
  it as separate entity and watcher-layout capabilities. The TUI receives only the
  resulting `core.WorkspaceService`; it does not import discovery or filesystem packages.
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
- **Ports live with their consumer.** `core.ConfigurationStore` carries neutral
  configuration entities and is implemented by `configstore.FS`; CLI, focused TUI,
  full TUI, and a future served adapter all call `ConfigurationService` rather than
  importing filesystem configuration packages.
- **A word that means the same thing in two places is spelled once.**
  `domain/resolution.go` holds the vocabulary an audit finding and an acceptance
  criterion share — `deferred`, `wontfix`, `tracked` — and each entity declares its own
  full set from that pool plus its own additions. Modelled as an OVERLAP, not a subset:
  `met` is not a finding status and `superseded` is not a criterion state, and claiming
  otherwise is exactly what lets the two drift. `theme.CriterionState` delegates the
  shared words to `theme.FindingStatus` for the same reason, so one word renders as one
  mark. Drift tests fail if either side diverges. This exists because it already went
  wrong: `landed` was legal in code, absent from the docs, and contradicted by an
  assertion (finding M3 of `2026-08-17-finding-status-surface`).
- **A state the tool cannot write is a state nobody can be held to.** Every
  closed-vocabulary field has a validated, atomic write verb — `task ac` for criterion
  state, `audit finding` for finding status and its resolution note — because the
  alternative is hand-edited markdown, which is how a vocabulary drifts from its own
  documentation. Reads stay tolerant so `lint` can REPORT malformed data already on
  disk; writes refuse to create it.

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

- **`model.go` / `view.go`** — the root `Model`, `Update` reducer, rendering, layout,
  focus (list ⇄ detail), window size, and shell-level key routing.
- **`atlas.go` / `session.go`** — the cross-space atlas and workspace-session boundary.
  The atlas renders one card per durable planning identity from
  `core.SpaceOverviewService`, exposes every diagnosed entry point, and opens a healthy
  selection through `core.WorkspaceService`. Which screen `ui` lands on is decided by the
  CLI, not by how many spaces are registered: standing in a repo names the space you
  meant, so that repo opens and the atlas stays one keystroke away; launched outside one
  there is no such statement, so `ui` forgives `config.ErrNoConfig`, seeds a registered
  space as the runtime workspace, and passes `WithAtlasLanding`. The seeded space keeps
  every surface live but was not chosen, so the top-rail chip reads `[atlas]` until the
  user leaves the atlas by any route. A successful open atomically swaps the
  entity service, config start, and watcher layout; a failed or stale open leaves the
  current workspace intact. `spaceSession` caches tabs/cursors/filters/dashboard/nav per
  checkout, while an outer workspace-generation stamp drops every asynchronous result
  launched by an older session. Atlas chrome has its own home-scoped theme (global
  flag/environment/home precedence, never the launch repo's override), shows each
  entry-point path, and keeps name/activity/registration ordering as explicit TUI state.
- **`entity.go`** — the **entity registry**: tasks/epics/audits/research as `*entityTab`s,
  each owning its own `list.Model`, cursor, loaders, list-scoped state (status
  view, sort, filter restore), and its **lifecycle table** (the transitions it
  offers + an `applyMove`). Read/browse is keybinding-free; lifecycle is declared
  here per entity (tasks by status via `Move`, audits by bucket via `MoveAudit`,
  epics and research none), so adding Projects/ADRs later is a new registry entry — including
  any `a`-menu / `:`-verb actions — not a reducer edit.
- **`dashboard.go`** — the per-space **dashboard** (`tskflwctl ui` opens here whenever
  it was launched inside a planning repo): a
  read-only composite of widgets over a single `core.Summary`, the in-app
  counterpart of `status`. Deliberately **not** an `entityTab` — it has no list,
  filter, sort, or lifecycle, so it carries its own small `dashboard` model plus a
  `Model.onDash` flag rather than joining `m.tabs` (a `-1` `entityDashboard`
  sentinel gives it `?`-help/title context without a tab slot). It's a *launch*
  surface: each navigable row jumps to an item/view on a real tab via `dashJump`,
  never mutating. Rule of thumb for new screens: a browsable list ⇒ a new
  `entityTab`; a read-only orientation screen ⇒ the dashboard pattern.
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
  `s`/`S` cycle), the `:` command bar, the `m` lifecycle action menu, and the modal
  registry. The action menu and `:` verbs are **registry-driven**: both read the
  active tab's transition table + `applyMove`, so tasks move by status and audits by
  bucket (close/reopen/defer, in-TUI now) through one entity-agnostic path —
  `movedMsg.to` is a plain string the closure interprets. Overlays (help, action,
  follow, edit) satisfy a small `modal` interface and live in an ordered stack the
  reducer loops; ForceQuit is handled once ahead of the loop, so a new overlay is one
  entry, not a new `handleKey` guard block + `bodyView` case.
- **`edit.go`** — the `e` inline field editor: the human face of `task set`. A form
  modal listing the typed editable fields (description / priority / tags / effort /
  tier) with their meanings (from the entity descriptor), the active field's widget
  inline (enum cursor, single-line input, or a wrapped `textarea` for description).
  Apply writes through `core.Service.SetFields` as a `tea.Cmd`; a core validation
  error stays on the field (shown inline) for an in-place fix, success returns to the
  picker. Task-only; status stays in the action menu. No new validation path —
  core re-validates, the same as `task set`.
- **`nav.go`** — S6 cross-link navigation: `f` follows structured references
  (a task's epic; an epic's tasks via a picker modal), `ctrl+o` pops the
  back-stack; hidden targets escalate the tasks view to `:all` rather than fail.
- **`watch.go`** — active-space-only `fsnotify` live reload: a self-perpetuating listener `Cmd`
  feeds `fsEventMsg`; a generation-guarded `tea.Tick` debounce (200ms) coalesces
  save-storms into one reload of every loaded tab, cursor preserved by id. The
  watched dir set comes from the active `core.Workspace`'s `Layout` port
  (`WatchPaths()`), not from a root the TUI reconstructs — layout knowledge stays in
  the store. A successful atlas switch closes the old watcher and starts only the new
  session's watcher; cached sessions do not retain file descriptors.
- **`help.go`** — the `?` keybinding overlay (`helpSections` is the runtime
  source of truth for keys) composited over the body with `ansi.Cut`.
- **`style.go` / `keys.go`** — the per-Model `styles` bundle (the active palette
  plus every chrome lipgloss style and the color helpers, built by `newStyles`
  from `design`/`theme`; the root Model holds a `*styles` the delegates share,
  repopulated once in `Run` after background detection) and the `key.Binding` map.

**Layout discipline is load-bearing** (a clipped-top-border class of bug):
subtract the border frame before sizing children, guard `View` before the first
`WindowSizeMsg`, truncate (never wrap) anything fed to a `Join`, and clamp the
composed view to the terminal. `TestModel_ViewFitsTerminal` locks the invariant
(View height == terminal height; no line wider than the terminal). The full
checklist is in `planning/research/6faxn1800y6n-tui-design-decisions.md`.

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

## Status (2026-08-25)

The layered shape now has two substantial primary consumers and several reusable
application seams; it is no longer architecture held in reserve for a hypothetical UI:

- The CLI exposes full task lifecycle and body mutation, epic/audit/research authoring,
  schema and wire discovery, configuration lifecycle, planning-space registry verbs,
  global `--space` selection, and the cross-space `status --all` overview.
- The Bubble Tea app provides its dashboard plus task/epic/audit/research tabs,
  a cross-space atlas with reversible entry-point navigation, filtering/sorting/find,
  structured navigation, task and audit lifecycle actions, task field editing,
  active-space filesystem reloads, and the embedded configuration editor. Reads and
  writes remain asynchronous commands over core services; production TUI code does not
  open planning/config stores.
- The configuration and cross-space features have dedicated core services and composite
  filesystem adapters. `SpaceRegistryService` owns catalog/selection/mutation policy and
  `SpaceOverviewService` owns cross-space summaries; Cobra is one primary adapter over
  both rather than the owner of either use case.
- Writes use atomic replacement/exclusive create, a repo-wide lock, content-version CAS,
  bounded retry for agent mutations, and parse-before-accept editor loops. Machine output
  is a versioned `internal/wire` contract with generated JSON Schema and golden coverage.
- The planning model's closed vocabularies now have write verbs rather than only linters.
  An acceptance criterion carries a state and a reason (`task ac`), an audit finding
  carries a status and a resolution paragraph (`audit finding`), and the words the two
  share are declared once in `domain/resolution.go`. Two invariants are enforced at write
  time rather than reported after: `task complete` refuses a task with an unexplained
  unmet criterion, mirroring `audit close` refusing while findings are open.

The atlas decision has now activated the narrow reusable workspace-opening boundary:
`core.WorkspaceService` and `internal/workspacestore` can open an explicit local entry
point for a primary adapter without exposing discovery or concrete persistence. The
broader `doctor`/fix promotion remains deferred because the atlas does not consume it.
Thread `context.Context` when an HTTP request path exists; reshape findings pagination
when a web findings view is scoped. Until those triggers occur, package movement or
speculative interfaces are out of scope.
