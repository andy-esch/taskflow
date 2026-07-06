---
status: completed
epic: 20-cli-ux-and-ergonomics
description: JSON error envelope, --body-file/stdin for long bodies, body editing through the tool, JSON create envelope gaps - from dogfooding by an agent
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, ux, dx]
created: "2026-06-12"
started_at: "2026-06-20"
updated_at: "2026-06-20"
completed_at: "2026-06-20"
id: 6fbj87000anh
---
# Agent-facing CLI ergonomics batch

> ⚠️ **Externally proposed — filed 2026-06-12** from an agent dogfooding
> session (filing ~15 review tasks through `tskflwctl`). These are the
> friction points that actually bit, in rough impact order. Complements
> [global-dry-run-for-mutating-commands](6fakbec03zrw-global-dry-run-for-mutating-commands.md) and
> [json-and-output-contract-fidelity](6fbj8700122m-json-and-output-contract-fidelity.md).

## Objective

1. **Errors are never machine-readable, even under `--json`.** Failures
   print plain text to stderr (`cmd/tskflwctl/main.go`); an agent driving
   `--json` must parse prose to learn *why* something failed. Emit a JSON
   error envelope (`{"schema_version", "error": {"code": "not-found",
   "message", …}}`) on stdout or stderr when `--json` is set — the exit-code
   sentinels already define the `code` vocabulary. This is the single
   biggest agent-facing win.
2. **`--body` needs a file/stdin variant.** Multi-paragraph bodies inline in
   a shell argument force heredoc-in-command-substitution gymnastics where
   one quoting slip corrupts the task. Add `--body-file <path>` (and `-` for
   stdin) to `task new` — cheap and standard.
3. **No body editing through the tool.** `task set` is frontmatter-only, so
   appending review notes to an existing task means hand-editing the file —
   exactly what the tool's atomic-write discipline exists to avoid. Consider
   `task set --body-file` (replace) or a `task append` for the common
   "add a section" case.
4. **JSON create envelope gaps.** `task new --json` returning the resolved
   `id` is excellent (title→slug is unguessable: "Esc/q must pop…" →
   `…escq…`) — but the envelope omits the resulting `status`, and the
   human-mode "→ next:" hint has no JSON counterpart. Add `status` (and
   consider `next` as a machine hint). Human and JSON modes also disagree on
   path form (relative vs absolute) — pick one.
5. **Document the `--set` coercion contract in help.** `task set --help`
   says "arbitrary key=value" with no hint that tier/tags/epic are typed,
   validated, and coerced while other keys pass through raw. One paragraph
   in the long help (and the unknown-key warning tracked in
   [task-set-follow-ups-sentinels-unknown-keys-canonical-field-table](6fbj87001d81-task-set-follow-ups-sentinels-unknown-keys-canonical-field-table.md))
   makes the contract legible to agents.
6. **`task new --start` (create-and-start shortcut).** *Added 2026-06-17 from
   the same dogfooding feedback that filed
   [audit-finding-level-operations-query-write-lint-sync](6fd5r5c03v5y-audit-finding-level-operations-query-write-lint-sync.md).* `task new` has
   `--next` (→ next-up) but no way to land straight in in-progress, so an
   agent that knows it's working the task immediately pays a `task new` +
   `task start` round-trip. Add `--start` (mirrors `--next`; mutually
   exclusive with it) that scaffolds directly into `in-progress`. One
   `domain.Status` switch in `NewTask` (`service.go` already branches `Next`).
   Minor caveat: mildly tensions with the draft readiness-gate idea
   ([task-readiness-state-draft-vs-finalized-in-frontmatter](6fbj87001m03-task-readiness-state-draft-vs-finalized-in-frontmatter.md)) — a `--start`
   that skips a future "draft can't be started" gate — but that gate is
   undecided and this is the autonomous-agent path, so it's additive for now.

Observed and *already good* (keep): exit codes 10/11 fire exactly as
documented; `task show --json` includes the body; multi-slug transitions;
`-C/--chdir`; `--json` on every read path; the description length error is
precise ("160 > 150").

## Acceptance criteria

- [x] `--json` failures emit a parseable envelope with a stable error code.
- [x] `task new --body-file -` works from stdin (and from a path); quoting
      torture gone. Mutually exclusive with `--body`.
- [x] A body can be replaced/appended through the tool, atomically.
- [x] Create envelope carries `status` (task status / epic status / audit
      bucket); `path` is now relative to the planning root in both human and
      JSON modes (was absolute in JSON). schema_version 1.4 → 1.5.
- [x] `task new --start` lands the task in in-progress; mutually exclusive with
      `--next`; flag-conflict errors cleanly.
- [x] Suite + lint green; README "agent use" section updated.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md)
- Touches `cmd/tskflwctl/main.go`, `internal/cli/`, `internal/cli/render/`,
  `README.md`.
## Progress (2026-06-12)

Item 1 (the headline) shipped per decision D9: `--json` failures emit
`{"schema_version","error":{"code","message"}}` on **stderr** with stdout
empty; codes reuse the exit-code vocabulary (`cli.WriteError`, wired in main;
pinned by the binary smoke test). Item 5 partially done: `task set --set` help
now states the typed/validated/--force contract. Remaining: `--body-file`/
stdin for `task new`, body editing through the tool, and the create-envelope
`status` field + path-form consistency.

## Progress (2026-06-17)

Shipped two clean items: `task new --body-file <path|->` (reads body from a file
or stdin via `resolveBody`, mutually exclusive with `--body`) and `task new
--start` (scaffolds straight into in-progress; `NewTaskParams.Start`; mutually
exclusive with `--next`). README create examples + 4 tests. Remaining: body
replace/append (the item-3 fork below — needs the `set --body-file` vs `task
append` decision) and the create-envelope `status` field + path-form
consistency. Left ready-to-start as a partial batch.

- **2026-06-17b**: Shipped the create-envelope item too — `task/epic/audit new
  --json` now carries `created.status` (additive) and `created.path` relative to
  the planning root in both modes (`render.CreatedJSON` + the three call sites;
  schema_version → 1.5; envelope tests assert it). Only the body replace/append
  fork remains (it needs the `set --body-file` vs `task append` decision).

## Progress (2026-06-20) — final item, batch complete

Shipped the body replace/append fork (item 3), resolving the design fork **both
ways** as a coherent pair so the agent face mirrors the human `task edit`:

- **`task append <slug>`** — append a section (`--body`/`--body-file`/stdin via
  `-`), separated by one blank line, single trailing newline.
- **`task set --body`/`--body-file`** — replace the whole body; its own atomic
  write, **mutually exclusive with field flags** (a combined call is exit-11, so
  field-surgery and body-surgery never share one ambiguous write).

Both go through one store primitive — `FS.EditBody(slug, text, appendMode, now,
dryRun)` (port + `Service.ReplaceBody`/`AppendBody`) — which swaps the body via
the existing surgical `yaml.Node` path (`replaceBodyStamped`): **unknown
frontmatter keys, comments, and key order survive**, `updated_at` is stamped, and
it carries the same **parse-before-write + compare-and-swap** discipline as
`SetFields`/`Move` (a concurrent relocation → `ErrConflict`, never a resurrected
slug). `--dry-run` previews without writing; `--json` returns the task envelope.
Tests: store engine (replace/append/preserve-unknown/dry-run/broken-frontmatter)
+ cli (append, stdin, empty→11, set-body replace, body+fields→11, dry-run).
Docs regenerated (`task append`); README update + lifecycle examples added.

Item 5 (the `--set` coercion help paragraph) shipped earlier; the full batch is
now done.

## Note (2026-06-12)

A draft task ([task-edit-opens-editor-on-the-body](6fbj87000pys-task-edit-opens-editor-on-the-body.md)) proposes the *human*
face of the body-editing gap this batch owns (items 2–3: --body-file, body
replace/append). Planning needs to pick one owner before either starts —
this batch's framing takes precedence; the draft defers here.
