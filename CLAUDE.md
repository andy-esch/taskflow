# taskflow — Claude Code guide

This repo is **`tskflwctl`**, a local-first planning CLI (Go) over
markdown+frontmatter task/epic/audit files. It **self-hosts its own planning**
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
fs/cobra, and **status/bucket == directory**. The TUI never touches the store —
it reads through `core.Service` as `tea.Cmd`s (no I/O in `Update`/`View`).

## Planning workflow — use `tskflwctl`, not `pm`

We dogfood: drive this repo's planning with the tool itself.

- **Create:** `./bin/tskflwctl task new "Title" --epic <id> [--next]` ·
  `epic new "Title" --description "..."` · `audit new <area> [--date]`.
- **Lifecycle:** `task start|promote|demote|complete|defer|deprecate <slug>...`.
- **Read/edit:** `task list|show|set|edit|append`, `epic list|show`,
  `audit new|list|show|findings|lint|close|reopen|defer`. Two faces of mutation: **agent**
  (field-level `task set`; body via `task append` / `task set --body|--body-file`,
  all scriptable + atomic) vs **human** (`task edit` — $EDITOR on the whole file,
  re-validated on save).
- **Self-describe (agents):** `schema` (contract: statuses, field registry,
  exit codes) · `schema task|epic|audit` (authoring guidance) ·
  `schema --json-schema` (Draft 2020-12 schema for the `--json` envelopes). Runs
  anywhere, no planning repo needed.
- **Hygiene:** `tskflwctl lint` (`--fix` to auto-repair). Keep `planning/`
  lint-clean.
- Tasks live in `planning/tasks/<status>/`; a task's `status:` **is** its
  directory. Every active task needs a one-line `description`.
- **`pm` (Python) is gone** — it was the prototype `tskflwctl` was ported from;
  it and its tests now live only in git history. The Go suite is the spec.

## Git

Inspection (`status`/`diff`/`log`) is fine; never run state-changing git
(`add`/`commit`/`branch`/…) unless asked. `tskflwctl` deliberately does **not**
touch git — it writes files; the user stages/commits.

## Code conventions

Match the surrounding code (naming, comment density, idiom). Errors wrap the
domain sentinels (`ErrNotFound` / `ErrValidation` / `ErrAmbiguous` /
`ErrConflict`) so the CLI maps them to exit codes (10, 11, 13, 14; 12 is retired
but reserved). New
file writes go through the atomic helpers in `store/atomic.go`
(`writeFileAtomic` to overwrite, `createFileAtomic` for exclusive create).
Frontmatter is edited **surgically** — preserve unknown fields, comments, and
key order.
