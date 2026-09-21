# taskflow — Claude Code guide

This repo is **`tskflwctl`**, a local-first planning CLI (Go) over
markdown+frontmatter task/Thread/epic/audit/research files. It **self-hosts its own planning**
under `planning/`. Two hats: the Go implementation, and the planning that tracks
its own work.

## Build / test / lint

- `just build` → `bin/tskflwctl` · `just test` → `go test ./...` · `just lint` →
  `golangci-lint run ./...`. Get all three green before calling code done.
- Standard Go layout: `main` is `./cmd/tskflwctl`, so `go build .` at the root
  does nothing useful — use `just build` / `go build ./...`.

## Architecture (read before changing code)

`docs/ARCHITECTURE.md` is the one-screen orientation: `cli` and `tui` are
**primary adapters** over `core`; the markdown filesystem is the
**secondary adapter** (`store`). Non-negotiables: DI via one `*cli.App`
populated in `PersistentPreRunE` (no globals), all output through injected
`io.Writer`, `--json` everywhere with a `schema_version`, the core never touches
fs/cobra, and **`status`/`bucket` is authoritative in frontmatter** (ADR-0003
§4 — tasks/audits/research/Threads are stored **flat and id-led**, `tasks/<id>-<slug>.md` ·
`audits/<id>-<slug>.md` · `research/<id>-<slug>.md` · `threads/<id>-<slug>.md`;
there is no status/bucket
directory (`threads/<id>-<slug>.md` stores metadata and task-ID membership only), epics
stay `NN-<slug>`, and research has no status at all). The TUI never touches the store —
it reads through `core.Service` as `tea.Cmd`s (no I/O in `Update`/`View`).
`schema_version` is [ADR-0008](planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)'s
monotonic machine-contract revision, not SemVer;
every revision from 1.68 declares `ADDITIVE` or `NOT ADDITIVE` beside the constant.

## Planning workflow — use `tskflwctl`, not `pm`

We dogfood: drive this repo's planning with the tool itself.

- **Create:** `./bin/tskflwctl task new "Title" --epic <id> [--next]` ·
  `epic new "Title" --description "..."` · `audit new <area> [--date]` ·
  `research new "Title" [--created <date>]` (the id is minted from `created`) ·
  `thread new "Title" --description "..." --goal "..." [--task <ref>]` (always unstarted).
- **Lifecycle:** `task start|next|ready|complete|defer|deprecate <slug>...` (verbs
  name the destination status; `next`/`ready` replace the old promote/demote, which
  still work as hidden aliases). `defer` takes `--until <YYYY-MM-DD>` (snooze).
  **`complete` REFUSES** a task with an acceptance criterion that is unmet and gives no
  reason — tick it, give it a state (below), or pass `--force`. Same rule as `audit
  close` refusing while findings are open.
- **Read/edit:** `task list|show|set|edit|append|ac`, `epic list|show`,
  `audit new|list|show|findings|finding|lint|close|reopen|defer`,
  `research new|list|show|path|set|edit|append` (no lifecycle verbs — research has no
  status), `thread new|list|show|path|frontier` (membership/lifecycle mutations have not
  landed yet). Two faces of mutation: **agent**
  (field-level `task set`; body via `task append` / `task set --body|--body-file`,
  all scriptable + atomic) vs **human** (`task edit` — $EDITOR on the whole file,
  re-validated on save).
- **Acceptance criteria — never hand-edit them; `task ac` owns them.** A criterion is
  `- [x]` (met) or `- [ ]` (not met), and a not-met one may say WHY with a trailing
  suffix: `- [ ] Ship the migration · **deferred:** waiting on the schema ADR`. States
  are `deferred | wontfix | tracked | n/a`, each REQUIRING a reason.
  `task ac <slug>` lists them numbered; `--check/--uncheck <n>` flips one;
  `--defer|--wontfix|--tracked|--na <n> --reason <why>` sets a state; and
  `--add <text>` / `--remove <n>` / `--replace <n> --text <new>` change which criteria
  exist. A criterion carrying a state has been DECIDED and does not block `complete`;
  only a silently unticked box does.
- **Finding creation, status, and its managed candidate row are tool-owned.** Create a
  finding with `audit finding new <audit> <title> --band H|M|L` and optional
  `--file`, `--component`, `--effort`, `--urgency`, `--body|--body-file`,
  `--recommendation`, and `--candidate`; the tool allocates the code and renders the
  Markdown. Update it with `audit finding <audit> <code> [--status <v>] [--pr N]
  [--note <text>] [--candidate <one-line>]` writes the `**Status:**`, the
  `**Resolution:**` paragraph, and that finding's optional `candidate-tasks:v1` row in one
  validated, atomic edit. An empty `--candidate` removes the row. Unversioned Candidate
  tasks sections are legacy prose: readable, but deliberately never guessed at or rewritten.
  Statuses: `open | in-progress | fixed | tracked | deferred | superseded | wontfix`.
  **`tracked` means handed to a task and REQUIRES the destination** (`tracked by
  <task-id>`) — it counts toward the audit's done band, because the audit's interest
  concludes when a finding is transferred. An audit's headline percent is the **settled**
  share (everything with a terminal disposition), so 100% is exactly when it is ready to
  close.
- **Triage (agents, cheapest first):** lead with the terse path — `epic show
  <id>` for an epic's task roster, and `task list -o table -c
  slug,status,description` for a compact, byte-stable table. `--json` is compact
  (not pretty-printed) and also takes `-c` to project just the fields you need
  (`task list --json -c slug,status,description`) — the cheap machine path. Reach
  for *full* `--json` (no `-c`) only when you need every frontmatter field
  (`tags`, `tier`, `priority`, `autonomy_level`, timestamps…). Note a `--json -c`
  projection is a string-valued column **view** (like `-o table`/`csv`); only
  full `--json` validates against `schema --json-schema`.
- **Self-describe (agents):** `schema` (contract: statuses, field registry,
  complete active/reserved process-exit taxonomy) ·
  `schema task|thread|epic|audit|research` (authoring guidance) ·
  `schema --json-schema` (Draft 2020-12 schema for the `--json` envelopes). Runs
  anywhere, no planning repo needed.
- **Hygiene:** `tskflwctl lint` (`--fix` to auto-repair ordinary frontmatter; Thread
  membership and graph defects remain deliberate repairs). Keep `planning/`
  lint-clean.
- Tasks live **flat** in `planning/tasks/` as `<id>-<slug>.md`; `status:` is
  authoritative in frontmatter (no mirror directory) — change status with the
  lifecycle verbs (never a hand-edit), which edit frontmatter **in place** (no file
  move). `lint` **flags** a missing/unrecognized status rather than relocating
  anything; a non-id-led `.md` under `tasks/`/`audits/`/`research/`/`threads/` is a `FileProblem` —
  non-entity files belong in `meta/`. Every active task needs a one-line
  `description`.
- **`pm` (Python) is gone** — it was the prototype `tskflwctl` was ported from;
  it and its tests now live only in git history. The Go suite is the spec.

## Git

Inspection (`status`/`diff`/`log`) is fine; never run state-changing git
(`add`/`commit`/`branch`/…) unless asked. `tskflwctl` deliberately does **not**
touch git — it writes files; the user stages/commits.

## Code conventions

Match the surrounding code (naming, comment density, idiom). Errors wrap the
domain sentinels (`ErrNotFound` / `ErrValidation` / `ErrAmbiguous` /
`ErrConflict`) so the CLI maps them to exit codes 10, 11, 13, and 14. The full
process taxonomy is discoverable through `schema --json`: 0 success, 1 generic
error, 130 interactive abort, and retired code 12 explicitly reserved alongside
the classified failures. New
file writes go through the atomic helpers in `store/atomic.go`
(`writeFileAtomic` to overwrite, `createFileAtomic` for exclusive create).
Frontmatter is edited **surgically** — preserve unknown fields, comments, and
key order.
