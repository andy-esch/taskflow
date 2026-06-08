---
status: in-progress
epic: 17-pm-go-cli
description: Build the Go pm CLI in the decided noun-verb hierarchy; reach parity with the Python prototype using tests/test_pm.py as the acceptance spec.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli]
created: 2026-06-06
started_at: 2026-06-07
updated_at: 2026-06-08
---

# Port pm to Go CLI (parity with Python prototype + test suite as spec)

## Objective

Reimplement `bin/pm` as a Go CLI in the decided noun-verb hierarchy,
reaching behavioral parity with the Python prototype. The Python tool
(`bin/pm`) and `tests/test_pm.py` are the **executable spec** — the Go
CLI must satisfy the same behaviors (frontmatter parse/serialize, lint
rules, schema, lifecycle moves, audit commands, validation).

## Blocked on (do not start until)

- [[phase-0.5-formal-tskflwctl-command-hierarchy-purpose-spec]] — the
  formal command spec is the build-to contract (Python tests are
  incomplete); port to *it*, not just to parity.
- [[go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness]]
  — the rock-solid foundation (layout, core/store/cli boundary, DI, test +
  lint harness) must land first; commands port *onto* it.
- [[rethink-pm-command-hierarchy-pm-noun-verb-research-cli-best-practices]]
  — hierarchy + Go framework + port strategy must be decided first.
- The Python `pm audit` prototype ([[bucket-audits-into-openclosed-and-ship-pm-audit-cli]])
  has soaked through a few weeks of real audits, so audit conventions are
  stable before they're frozen in Go.

## Prior art & scope fence

Bootstrap from the **taskflow** spike (`../taskflow`): cobra + viper +
Bubble Tea, Pattern-C layout, JSON fast-index — strip to the fence.
**Scope:** CLI parity NOW; **TUI preview (Bubble Tea) eventually**;
MCP/RAG/brain **out by a long shot**. (Open: is this task taskflow revived,
or a fresh build borrowing its skeleton? — see research doc Q2.)

## Likely implementation plan (refine after the research lands)

- [ ] Stand up the Go module + cobra/viper; wire the noun-verb skeleton
      from the research doc (explicit `<noun> <verb>`; no aliases). Borrow
      taskflow's `cmd/`+`internal/cli` skeleton; drop its `services/`/brain.
- [ ] `tskflwctl init` (interactive + `--yes/--path`): scaffold the
      single-repo taskflow tree + config file.
- [ ] Port the data layer: frontmatter parse/serialize (real YAML — no
      fallback-parser split), task/epic discovery, status-dir invariants.
- [ ] Port commands group-by-group, each gated by translating the
      matching `tests/test_pm.py` cases into Go table tests.
- [ ] Port `lint`/`schema`/`set` validation (incl. the §1–§4 rules from
      [[tighten-pm-cli-ergonomics]]) and the `audit` group.
- [ ] Decide distribution (how `./bin/tskflwctl` is built/installed;
      build/install recipe) and update AI_README/README/routers.

## Acceptance criteria

- [ ] `tskflwctl {task|audit|epic|adr} {command}` works in the decided
      hierarchy — fully explicit, no flat aliases, lifecycle via `task move`.
- [ ] Behavioral parity: every behavior asserted by `tests/test_pm.py` has
      an equivalent passing Go test.
- [ ] `--json` output matches the Python tool's shape (agent-consumed).
- [ ] Docs (AI_README command table, README, CLAUDE/GEMINI routers) point
      at the Go CLI; Python `bin/pm` retired or clearly marked legacy.

## Out of scope

- Changing task/epic/audit **file formats** — same markdown+frontmatter.
- Re-deciding the hierarchy (that's the research task's job).
- A storage-layer change (SQLite etc.) — see
  `research/sqlite-vs-markdown-for-pm-system.md`; separate decision.

## Progress Log

**2026-06-07 — task group #1: read + lifecycle.** Built on the foundation:
- **`task show`** (human + `--json` with body), **`task list`** now defaults
  to active statuses (`--all`/`--status` to widen) — pm-parity.
- **Transition engine:** internal `move` over the store, exposed as the
  explicit verbs `start/promote/demote/complete/defer/deprecate` + a generic
  `task move <t>... <status>`. Idempotent (re-applying current status = no-op);
  per-target date stamping; batch slugs with a per-item report + first-error
  exit.
- **Atomic writes** (`writeFileAtomic`: temp+fsync+rename) + **surgical
  frontmatter** (`yaml.v3` Node: updates only changed keys, **preserves
  unknown fields/order**, emits valid YAML). Move = update-in-place atomic →
  rename into the target status dir.
- **Semantic exit codes** (`cli.ExitCode`: 10 not-found · 11 validation ·
  12 invalid-transition · 13 ambiguous) wired into `main`; sentinels in
  `internal/domain/errors.go`.
- Verified on real data: `promote` then `demote` round-trip preserved all
  other frontmatter; **the Python pm still reads the Go-written file**
  (interop confirmed). `go test ./...` + `golangci-lint` green.

**2026-06-07 — task group #2: `set` + `init`.**
- **`task set <slug>`** — typed flags (`--description/--priority/--epic/
  --tier/--autonomy/--tags/--effort`) + repeatable `--set key=value` for
  custom fields, in **one atomic surgical write** (only changed keys touched;
  unknown fields + body preserved; lists encoded as valid YAML). Write-time
  validation in `domain.ValidateField` (priority enum, tier/autonomy 1-5,
  description length/single-line, `status` refused → use a verb) → `ERR_VALIDATION`
  (exit 11), nothing written on failure.
- **`tskflwctl init [--path]`** — idempotently scaffolds the planning tree
  (`tasks/* epics/ projects/ audits/{open,closed,deferred}/`) + `.tskflwctl.toml`.
  Non-interactive by design (no TTY-hang); skips repo discovery via its own
  no-op `PersistentPreRunE`.
- Unit tests added at every layer: `domain` validation, store `SetFields`
  (preservation + valid-YAML list), `config.Init` (idempotent + Discover),
  CLI `set`/`init` (incl. exit-11). `go test`/`golangci-lint` green; demoed
  init→set→show on a throwaway repo.

**2026-06-07 — group #3: `epic` group + `lint`.**
- **`epic list`** (rollup: total/done/% joined on tasks' `epic:` field) +
  **`epic show <id>`** (epic + its tasks + body), human + `--json`. The port
  grew from `TaskStore` to a combined `core.Store` (TaskStore + EpicStore);
  `var _ core.Store = (*FS)(nil)`.
- **`lint`** — validates active task frontmatter (required fields, `tier`/
  `autonomy` 1-5, `priority` enum, `description` length/single-line +
  required for next-up/in-progress, **unknown-epic** check). `domain.LintTask`
  is pure + table-tested. Exits `ERR_VALIDATION` (11) on any issues.
  Verified on real data: **agrees with the Python pm (all 3 active pass)**.
- Tests added: `domain.LintTask` (clean/dirty/bad-constraints/unknown-epic),
  store `ListEpics`/`GetEpic`, a **pure-core rollup test with a fake Store**,
  and CLI `epic list/show` + `lint` clean/dirty-exit-11.

**2026-06-07 — robustness: actionable frontmatter errors + resilient reads.**
- Frontmatter decode failures now produce **field-level, fix-oriented**
  messages instead of leaking yaml.v3 internals: comma-string list fields →
  `field "tags" must be a YAML list … fix: tags: [a, b, c]`; an unquoted
  colon in a value → `field "description" (line N) … wrap the value in
  quotes`. (`internal/store/diagnose.go`.)
- **Resilient listing:** `ListTasks`/`ListEpics` now return per-file
  `domain.FileProblem`s instead of dying on the first bad file. `list`/`epic
  list` show the good data + report bad files to stderr (exit 11); `lint`
  folds them into its report. The store port grew a problems return value;
  threaded through core → cli.
- Tests: `diagnose` (comma-list + unquoted-colon), resilient `ListTasks`
  (skips bad, returns problem), CLI list reports-bad-but-shows-good.
- Demoed on the real (messy) `desirelines-planning`: lints every task,
  pinpoints each broken file with a fix.

**2026-06-08 — `lint --fix` (auto-repair the data the diagnostics flag).**
- `lint --fix` + `--dry-run`: a **text-level** frontmatter normalizer (works
  on files that don't even parse) — quotes scalar values containing an
  unquoted `": "`, and normalizes list fields (`tags`/`related_tasks`/… given
  a bare/comma value) into YAML flow lists. Conservative: only touches
  top-level `key: value` lines; leaves already-quoted/flow/block values and
  the body verbatim. Atomic write; idempotent. `store.FixFrontmatter(dryRun)`
  walks all task + epic files; threaded core → cli.
- Tests: fixer (colon-quote, comma→list, no-op, idempotent), `FixFrontmatter`
  dry-run-vs-write (file becomes readable after), CLI `lint --fix`
  dry-run→fix→list. Demoed `--dry-run` on real `desirelines-planning`:
  previews the exact files+changes without touching them.
- This makes the `strict-yaml-frontmatter` migration a single command:
  `tskflwctl lint --fix`.

**2026-06-08 — `audit` group (read + bucket lifecycle, first cut).**
- `audit list [--all|--closed|--deferred]` (default open) with **finding-count
  rollup** parsed from the body (`#### CODE.` headers + `**Status:** open`);
  `audit show <slug>`; `audit close|reopen|defer <slug>...` = moves between
  `audits/{open,closed,deferred}/` (idempotent). Store grew an `AuditStore`
  port + `auditsDir`; `domain.AuditBucket`.
- Demoed on real `desirelines-planning`: lists the 7 open audits with correct
  counts (e.g. `2026-06-06-schemas-scripts 5/5 open`), 21 across buckets.
- Tests: store finding-count + move + not-found; CLI list-default-open / --all
  / close-moves-bucket.
- **Deferred (audit finding-level):** `status`/`fixed`/`landed`/`followup`/
  `sync` (parse + edit finding Status lines + candidate list), `new`/`noop`
  (routine generation), `findings`/`stats`, and Closeout-append on `close`.

**Deferred (new features, not parity — 2026-06-08):** `adr` + `project`
groups. Pick up after the core port is solid.

**2026-06-08 — critical-review pass (two adversarial subagents) + tightening.**
Both reviewers independently judged the architecture sound ("holding up well
as it scales", "no broad correctness problem"). Applied the must-fix +
high-value-cheap batch (lint + all tests green; `Move` re-smoke-tested):
- **Data-loss guard (C1):** `documentMapping` no longer silently overwrites
  frontmatter that's valid YAML but *not* a mapping (a bare scalar/sequence) —
  `Move`/`SetFields` now error instead of discarding it. Test added.
- **Crash-safe `Move` (M2):** was write-in-place-then-rename, which on a crash
  could leave a file in the old dir whose frontmatter claims the new status
  (breaking status==dir, non-self-healing). Now writes atomically into the
  *target* dir then removes the old file last — a crash leaves at worst a
  recoverable duplicate, never a mismatch.
- **JSON contract (M3):** `task/epic/audit list --json` now carry an
  `unreadable` array (like `lint --json`) so an agent never silently loses
  unreadable files; human output still prints problems to stderr.
- **Dedup (H1):** the task and audit transition loops collapsed into one
  generic `runMoves` helper (`cli/moves.go`) — kills a live copy-paste before
  `adr`/`project` add a third movable noun.
- **Debt (H2):** `EpicShowJSON`'s triplicated anonymous struct → named
  `epicMetaJSON`.
- **Comment preservation (T1):** `setMapNode` now carries comments from the
  replaced value node onto the new one (an inline `# note` on an updated field
  was being dropped); added a test that pins comment + key-order preservation,
  verified it fails without the fix. Plus a `--set key=value` round-trip test
  (T2) covering the riskiest untouched write path.

**2026-06-08 — removed vestigial proto layer.** The spike's protobuf
"contract" (`contracts/` — `task.proto` + generated Go/Python/JSON-schema, root
`buf.yaml`/`buf.gen.yaml`) and its only consumer (`internal/index/`, the dead
JSON fast-index) were leftovers, imported by nothing in the live binary — the
real Task contract is the hand-written `internal/domain` struct over real YAML.
Deleted both, dropped the now-unused `google.golang.org/protobuf` dep
(`go mod tidy`), and rewrote the stale spike-era `README.md` (which advertised
`contracts/` as the protobuf source of truth + non-existent `just proto`/
`dev-up` recipes). `go build/vet/test ./...` + lint green. Still-vestigial spike
dirs left in place pending a call: `services/` (Python "brain") and
`internal/tui/` (Bubble Tea sketch — to be rebuilt over the current `core`).

**Documented, deliberately deferred (review findings, not yet actioned):**
- *Polish (low):* audit finding regexes over-match — `#### CODE.` inside a
  fenced code block is counted, and `**Status:** open-ish` matches via `\b`
  (`store/auditstore.go`); `domain.ErrNotFound`'s message says "task" even when
  used for epics/audits; `render.MoveResult.Status`'s JSON key is `"status"`
  even for audit buckets. Batch these in a follow-up cleanup.
- *Architecture — left explicit on purpose (both reviewers concur):* the
  per-noun `list` command shape stays duplicated (collapsing behind a generic
  `runList[T]` obscures more than it saves); `render` importing `core` for two
  view-models (`EpicSummary`/`LintResult`) is fine at this size — revisit only
  if a third appears; `core.Store` stays one combined interface (it's honest
  about the Service's surface) — if `fakeStore` grows painful, fix it with an
  embedded no-op stub, *not* by splitting the port. One flagged smell to watch:
  `FixFrontmatter` is the one `Store` method that isn't a noun-CRUD op —
  candidate to split into a `Fixer` port later.

**Remaining for parity:** audit finding-level commands (above); `track`;
`schema`(+`--type cli`); global `--dry-run`; advisory `flock`; structured JSON
error envelope; interactive `init` wizard. Minor: dates written quoted vs pm's
unquoted — valid + pm-readable.

## Related

- Epic [[17-pm-go-cli]].
