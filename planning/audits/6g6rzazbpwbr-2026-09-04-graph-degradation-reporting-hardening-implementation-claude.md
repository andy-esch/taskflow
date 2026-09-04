---
schema: 1
id: 6g6rzazbpwbr
bucket: closed
area: graph-degradation-reporting-hardening-implementation-claude
date: "2026-09-04"
updated_at: "2026-09-04"
---

# Audit: graph degradation reporting hardening — claude — 2026-09-04

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> Required second pass: after completing the checklist, review again as a devil's advocate for systemic failure modes. Challenge shared abstractions, adapters that only happen to be filesystem-backed, test helpers that conceal divergent snapshots, and error paths that appear noisy but still report a false clean state. Prefer one demonstrated systemic issue over several speculative findings, and settle every challenged pattern with hostile evidence.
>
> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only source. Before inspecting implementation, running tests or generators, or making mutation probes, create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement whose `.git` metadata points back to the shared checkout. At completion, copy back only this assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Create an isolated clone whose working tree includes the exact current
source contents, including staged, unstaged, untracked, and deleted files:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g6rzazbpwbr-2026-09-04-graph-degradation-reporting-hardening-implementation-claude.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review-claude.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
test -d "$SANDBOX/.git"
cd "$SANDBOX"

git add -A
git -c user.name='Taskflow Review Sandbox' \
  -c user.email='review-sandbox@invalid' \
  -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null \
  commit --allow-empty --no-verify -m 'chore: capture review sandbox baseline'
```

The checkpoint is the only commit you may create. Confirm `git rev-parse --git-dir` resolves inside
`$SANDBOX`. Perform all inspection, builds, tests,
formatting, generation, fixtures, mutations, and report editing there. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`; inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then transfer
only the audit, guarded against concurrent source edits:

```sh
test "$(git -C "$SOURCE_ROOT" hash-object "$SOURCE_AUDIT")" = "$SOURCE_AUDIT_BLOB" || {
  printf 'source audit changed; do not overwrite it; preserve sandbox at %s\n' "$SANDBOX" >&2
  exit 1
}
TRANSFER="$(mktemp "${SOURCE_AUDIT}.review-transfer.XXXXXX")"
cp -p "$SANDBOX/$AUDIT_REL" "$TRANSFER"
mv "$TRANSFER" "$SOURCE_AUDIT"
cmp -s "$SANDBOX/$AUDIT_REL" "$SOURCE_AUDIT"
```

Do not copy anything else back. Leave the sandbox in place and report its path until the
implementation owner confirms receipt. If the hash guard fails, report the conflict and sandbox
path instead of resolving it in the shared checkout.

## Review brief

Perform an independent adversarial implementation review of
[report graph degradation in status and lint](../tasks/6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
on branch `feat/graph-degradation-reporting-hardening`, based on `main` at commit `92f9893`.
The implementation is in commits `0e8a25c`, `2ed59b6`, `1ab44d0`, and `198d585`.
Review the complete `git diff 92f9893...HEAD`; use this assigned audit as the review brief. Judge the work
against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), and the task acceptance criteria.

Assume the implementation can be systemically wrong despite green tests. It moves repository-wide
task-graph health into the ordinary summary path and then publishes it through CLI text, JSON, the
TUI dashboard, and the cross-space Atlas. It also relies on existing lint ownership for legacy
dependency diagnostics. Re-derive those contracts from production code; comments, golden churn,
and filesystem-backed test doubles are not proof.

Do not edit implementation or other planning files. The newly tracked
[`preserve-portable-load-diagnostics-in-board-and-status`](../tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
follow-up acknowledges that Board and Summary currently collapse neutral load problems to
`domain.FileProblem`. Do not duplicate that as a finding unless you demonstrate a broader defect,
incorrect scope/sequencing, or present data loss beyond the recorded debt.

## Intended contract to challenge

- `core.Service.Summary` loads task records exactly once through its injected `TaskGraphSource`.
  Counts, in-progress tasks, unreadable-task problems, graph health, and graph detail all describe
  that one read; the legacy store is not a second task authority.
- `SummaryStore` owns only non-task summary reads. `PlanningSummarySource` explicitly composes the
  non-task summary port and `TaskGraphSource`, so cross-space/local composition cannot silently omit
  graph health. Alternate and future adapters can satisfy the capability without filesystem paths.
- Healthy, degraded, and broken retain the same core meanings used by mutation guards, graph reads,
  Thread projections, and Board. Non-healthy detail is actionable and truthful; healthy output does
  not acquire a spurious warning.
- Human `status`, `status --all`, dashboard, and Atlas make non-healthy graphs conspicuous without
  turning a read-only status command into a failing process or independently deriving graph state.
- Plain `status --json` exposes optional top-level `graph`; `status --all --json` exposes optional
  `spaces[].summary.graph`. Non-healthy values carry health and detail through the shared wire type.
  The schema bump to 1.60 and all generated/golden artifacts accurately describe compatibility.
- Ordinary lint reports graph-invalid states through one established owner per defect. Missing and
  ambiguous legacy references are visible once through grouped legacy diagnostics, remain advisory
  under the accepted policy, and cannot produce a false clean message. Strict graph mutations still
  refuse every degraded/broken graph independently of lint exit policy.
- Existing status counts, audit/epic problems, cross-space failure retention, TUI attention counts,
  other JSON commands, and non-filesystem portability do not regress.
- Planning/docs state only what shipped. The portable diagnostic follow-up is appropriately scoped,
  sequenced after its lint prerequisite, and dogfooded in the production Threads graph.

## Mandatory evidence floor

A `ready` verdict is not credible unless the report contains all of the following:

1. A consumer and composition inventory for `Summary`, `SummaryStore`, `PlanningSummarySource`,
   `TaskGraphSource`, `Service.Summary`, `summarize`, `OpenPlanningStore`, `SummaryJSON`,
   `GraphHealthJSON`, `RenderSummary`, dashboard summary rendering, Atlas attention/detail, and every
   schema encoder/decoder or version declaration affected by 1.60. Classify each production use;
   grep counts alone are not an inventory.
2. End-to-end throwaway-space probes for healthy, degraded, and broken graphs, including unreadable
   task input, missing canonical dependency, ambiguous legacy slug reference, duplicate/missing task
   identity, invalid status, and a cycle. Compare `status`, `status --json`, `status --all`, TUI, and
   `lint`; also attempt an ordinary dependency mutation against each invalid state.
3. A split-authority adapter probe where `TaskGraphSource` intentionally returns a different task
   set from any legacy `ListTasks` capability. Prove counts, in-progress rows, problems, health, and
   detail all use the graph-source snapshot. Repeat with a pathless in-memory or remote-shaped
   adapter and with errors from each composed capability.
4. Multi-space probes containing healthy, degraded, broken, unavailable, and stale-last-good spaces.
   Check Atlas attention, detail text, sorting/selection stability, refresh recovery, and whether one
   space's failure can hide or relabel another space's graph verdict.
5. JSON/schema compatibility evidence: exact plain versus `--all` field placement, healthy omission,
   non-healthy inclusion, stable health vocabulary, detail omission rules, schema 1.60 declaration,
   JSON Schema output, and representative unrelated command envelopes. Check both pretty and compact
   output and any decode/round-trip consumers found in the inventory.
6. Lint output/exit evidence for every `TaskGraphProblemCode`, with special attention to missing and
   ambiguous legacy references, unreadable input, duplicate messages, false `✓ no issues found`, and
   advisory versus blocking policy. Trace each problem to exactly one presentation owner.
7. At least these temporary, restored mutation probes, naming the test that kills each mutation:
   - make `Summary` count tasks from the store while health comes from `TaskGraphSource`;
   - remove either half of the `PlanningSummarySource` composite capability;
   - suppress degraded or broken output in one of CLI, dashboard, or Atlas;
   - put `graph` at the wrong JSON nesting level or emit an empty healthy object;
   - stop bumping the schema or leave one representative golden envelope on 1.59;
   - print a legacy dependency defect twice or let it end with a false clean message; and
   - weaken the normal mutation guard because lint treats resolved legacy debt as advisory.
   A surviving mutation is a coverage finding even if production currently looks correct.
8. Repeated focused tests under `-race`, an uncached full `go test -race ./...`, static analysis,
   generated-doc drift, module tidiness, planning/audit lint, and `git diff --check`, with exact
   commands, Go version, durations, and cached/uncached distinction. If resource limits prevent an
   item, record that rather than silently substituting weaker evidence.

## Required adversarial angles

1. **Snapshot coherence and split authority.** Look for a second task scan, a hidden `ListTasks`
   dependency, or task data reconstructed from graph problems. Challenge mutation between task,
   epic, and audit reads and distinguish unavoidable cross-entity skew from a false claim that task
   graph fields share a snapshot.
2. **Port strictness and portability.** Inspect compile-time interface satisfaction and every
   `OpenPlanningStore` implementation/test double. Challenge nil or partially implemented graph
   sources, filesystem assumptions, pathless diagnostics, remote latency/errors, and whether the
   composite interface is too broad or accidentally narrower than the production need.
3. **Health semantic drift.** Compare Summary with Board, mutation guards, graph queries, and Thread
   projections for every problem code. Look for zero-value health, nil reads, detail chosen from an
   unstable order, degraded reported as broken (or vice versa), and health/detail contradictions.
4. **Presentation honesty.** Confirm all human surfaces are noisy enough without double counting a
   graph plus its unreadable file, hiding the warning below routine content, or claiming a repair
   action unsupported by an adapter. Check no-color, narrow TUI, empty repository, and recovery.
5. **Wire compatibility.** Treat 1.60 as a public contract. Challenge `omitempty`, top-level versus
   nested placement, shared object reuse, generated comments, consumers that switch on exact schema,
   and the broad unrelated-golden churn. Decide whether the versioning policy was followed rather
   than assuming a bump makes the change safe.
6. **Lint ownership and policy.** Re-derive why each graph code is omitted from or added to domain
   lint. Attack the grouped legacy path with multiple owners/tasks, mixed missing and ambiguous refs,
   resolved legacy refs, unreadable documents, and simultaneous cycle/identity faults. Distinguish
   advisory exit behavior from missing visibility.
7. **Error containment and refresh.** Force graph load, non-task summary, per-space open, and render
   failures. Look for stale graph warnings surviving recovery, last-good data being erased too soon,
   partial success presented as healthy, or one source error masking more actionable evidence.
8. **Regression blast radius.** Exercise ordinary status counts, epics/audits, dashboard focus,
   Atlas attention, all registered spaces, other JSON envelopes, and direct core callers. Search for
   mocks whose easy filesystem conformance masks a production capability break.
9. **Planning truthfulness.** Compare task closeout, ADR amendments, README/CLI docs, architecture
   claims, follow-up scope, dependency direction, and Thread placement with actual code and CLI
   behavior. Flag materially stronger claims, not stylistic omissions.
10. **Systemic second pass.** After the checklist, deliberately seek a shared helper, zero value,
    interface composition, projection/presentation split, golden-update mechanism, or diagnostic
    ownership convention that could hide an entire class of defects. Demonstrate it or record the
    evidence that settles it.

## Validation and restoration

Run proportionate validation and hostile scratch-space probes inside the reviewer sandbox. Restore
every probe and generated artifact to the sandbox checkpoint. Do not install dependencies, push,
edit implementation permanently, create follow-up tasks, change finding statuses, close this audit,
or edit the other reviewer's audit. At finish, sandbox `git status --short` must show only this
assigned audit before its guarded one-file transfer back to the source checkout.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/HEAD/worktree state, runtime, and exact validation results;
- a compact data-flow and port-boundary re-derivation;
- findings grouped by severity, each with stable code and `**Status:** open` in the heading;
- acceptance-criteria traceability, hostile-evidence, and restored-mutation ledgers;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but the evidence and mutation ledgers are still required.
Do not pre-resolve findings; the implementation owner will triage them with
`tskflwctl audit finding`.

## Reviewer report

**Verdict: ready with tracked follow-ups.**

The shipped behaviour is correct on every repository state I could construct. `Summary` really does
read tasks once through `TaskGraphSource`; the cross-space composite port is real and
compile-enforced; the four human surfaces and both JSON shapes agree with `board` and with the
mutation guards; schema 1.60 is declared and generated consistently; and no probe produced a false
clean, a false healthy, a spurious warning on a healthy repository, or a cross-space verdict that
hid or relabelled another space. Every finding below is in the **test net or the prose**, not in the
runtime: one regression test cannot observe the duplication it was added to prevent (M1), lint
prescribes a remedy the tool always refuses (L1), the new wire object is never exercised by the
in-suite schema validator (L2), and the composite port's documented invariant has no pin of its own
(L3).

### Reviewer sandbox

The brief's mandatory sandbox was created and used for all builds, tests, generators, probes,
mutations, and report editing:

- Sandbox: `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review-claude.raYgSa`
- `git rev-parse --absolute-git-dir` →
  `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review-claude.raYgSa/.git`
  (resolves inside the sandbox; a `--no-hardlinks` clone plus `rsync -a --delete --exclude=.git`,
  not a worktree or symlink)
- Baseline checkpoint: `7a1452e chore: capture review sandbox baseline` on top of `3c74392`
- `git diff --stat 3c74392 HEAD` in the sandbox touches **only** the two review briefs and
  `scripts/README.md` / `scripts/prepare-adversarial-review-audits.sh` — the source root's
  uncommitted state at capture time. **Zero drift under `internal/`, `cmd/`, `docs/`, `README.md`**,
  so the implementation under review is byte-identical to `3c74392`.
- Captured source audit blob for the transfer guard: `62fbae95dc956fd7453af38e4469f62055b3cfcc`.
  That guard **failed** at transfer time: the implementation owner edited and then committed the
  brief's own sandbox instructions (`4df372e`, blob `87b93874084e2079351313576698c1efab6b907b`).
  Per the brief I did not overwrite the source. This report was re-assembled in the sandbox on top
  of the **current** brief, preserved byte-for-byte, with the reviewer-report placeholder — which
  no one else had touched — as the only removed line.

**Disclosure.** The sandbox requirement was added to this brief while the review was already in
progress; the earlier half of the evidence-gathering ran in the shared checkout. Two consequences,
stated plainly rather than papered over:

1. While working there I observed `internal/core/service.go` and `internal/cli/lint_test.go` being
   edited by another process mid-session (mtimes 09:16; `lint_test.go` carried a
   `t.Logf("ACTUAL LINT OUTPUT:…")` probe and `service.go` a mutation identical in shape to my M6a).
   My `git checkout -- internal/core/service.go` restore step **discarded that in-flight edit**. I
   stopped reverting files I had not dirtied as soon as I noticed. This is exactly the collision the
   new sandbox rule exists to prevent.
2. Because of that contention, **every load-bearing result was re-derived from scratch in the
   sandbox** — the complete mutation ledger below, the full uncached race suite, static analysis,
   generator drift, and tidiness. Nothing in this report rests on a measurement taken in the shared
   checkout. Read-only CLI probes (`status`, `lint`, `task depend …` against throwaway trees under
   the scratch directory) used a binary built from the identical source.

The source root was never committed to, staged, branched, reset, cleaned, or stashed by me, and no
write-capable project command was run against it.

### Reviewed state, runtime, and validation results

| item | value |
| --- | --- |
| Branch | `feat/graph-degradation-reporting-hardening` |
| Reviewed HEAD | `3c74392` (branch tip advanced to `4df372e` during the review; that commit touches only the two review briefs and `scripts/`, **zero implementation change**, so every result below holds at `4df372e` unchanged) |
| Base | `92f9893` (merge of #186) |
| Commits reviewed | `0e8a25c`, `2ed59b6`, `1ab44d0`, `198d585` (+ `3c74392`, the review briefs) |
| Worktree | source root clean at session start; sandbox clean at its baseline and after every probe |
| Go | `go1.26.6 darwin/arm64` |

All commands below were run in the sandbox.

| check | command | result |
| --- | --- | --- |
| full suite, **uncached**, race | `go clean -testcache && go test -race ./...` | **PASS** — exit 0, 24 `ok` packages, 0 `FAIL`; 14.2 s wall / 15.5 s user |
| focused repeat, race | `go test -race -count=20 -run '<12 graph/status/lint/golden tests>' ./internal/core/ ./internal/cli/ ./internal/cli/render/ ./internal/tui/ ./internal/wire/` | **PASS** — 10.4 s wall |
| static analysis | `golangci-lint run ./...` (v2 config) | **0 issues** — 2.0 s wall |
| vet | `go vet ./...` | **clean** (exit 0) |
| generated CLI docs | `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | **no drift** |
| module tidiness | `go mod tidy -diff` | **clean** |
| whitespace | `git diff --check` | **clean** (exit 0) |
| planning lint | `tskflwctl -C planning lint` | `✔ all planning entities and dependency links pass lint`, exit 0 |
| audit lint | `tskflwctl -C planning audit lint 6g6rzazbpwbr` | `✔ all audit findings pass lint`, exit 0 |

Cached/uncached: the full race run was explicitly uncached (`go clean -testcache` immediately
before, exit code and package tally captured to a file). Mutation-ledger runs used the ordinary
cache; every mutation invalidates the packages it touches, and each row's result was read from that
run's own `--- FAIL` lines.

**Not run, recorded rather than substituted.** Dependency installation was forbidden and the Python
`jsonschema` package is unavailable, so live-envelope conformance against `schema --json-schema` was
proved with a hand-written structural validator (declared-property closure under
`additionalProperties:false`, `required` presence, recursive `$ref` descent) rather than a
general-purpose validator. That is weaker on type and format constraints, and I say so instead of
claiming full validation. The repository's own `TestJSONSchema_ValidatesRealOutput` does run a real
validator in-suite — but never over a non-healthy summary, which is finding **L2**.

### Data-flow and port-boundary re-derivation

```
                       ┌──────────────── ONE task read ─────────────────┐
store.FS.ListTasks ──▶ FS.ReadTaskGraph ──▶ TaskGraphRead{Tasks, Problems}
                                                     │
        ┌────────────────────────────────────────────┼────────────────────────────────┐
        ▼                                            ▼                                ▼
   read.Tasks                       taskGraphFileProblems(read.Problems)     NewTaskGraphRead(read)
   ├─ Counts                            └─ Summary.Problems (p1)             ├─ .Health() → GraphHealth
   ├─ InProgress                                                             └─ taskGraphHealthDetail()
   ├─ RevisitDue                                                                 → GraphDetail
   └─ rollupEpics(epics, tasks)

SummaryStore  (epics + audits ONLY) ──▶ Epics / OpenAudits / ReadyToClose / Findings / BadEpicStatus
```

- `service.go:286` — `Service.Summary()` → `summarize(s.store, s.taskGraphs, s.now())`.
  `s.taskGraphs` is fixed once in `NewService` (`service.go:189-193`): the aggregate `Store` when it
  satisfies `TaskGraphSource`, else the `taskStoreGraphSource` shim. `WithTaskGraphSource`
  (`service.go:49`) overrides it and is nil-guarded through `isNilCapability`.
- `service.go:289-296` — `summarize` calls `loadTaskGraphRecords(taskGraphs)` **once** and derives
  `tasks`, `p1`, and `graph` from that single value. `store.ListTasks` is not merely unused on this
  path — `SummaryStore` (`store.go:281`) no longer declares it, so a relapse is a compile error, not
  a convention.
- `space_overview.go:14-17,114` — `PlanningSummarySource = SummaryStore + TaskGraphSource`, and
  `summarizeSpaceGroup` passes the one opened store into **both** parameters of `summarize`. The
  composite is the return type of `SpaceOverviewStore.OpenPlanningStore` (`space_overview.go:23`),
  so an adapter lacking the graph capability cannot compile rather than silently omitting health.
- `spacestore/fs.go:81` — the only production `OpenPlanningStore`; returns `*store.FS`, which carries
  `var _ core.TaskGraphSource = (*FS)(nil)` (`store/fsstore.go:68`).
- Single-owner presentation: `render.taskGraphWarning` (`render/render.go:390`) is the one human
  sentence, shared by `BoardHuman` and `SummaryHuman` (`render/status.go:81`);
  `wire.toGraphHealthJSON` (`wire/envelopes.go:93`) is the one JSON projection, shared by
  `ToBoardEnvelope` and `ToSummaryJSON`. `renderCompactSpaceSummary` (`render/status.go:210`) and
  `tui/dashboard.go:187` re-compose the same wording inline; `tui/atlas.go:876` folds the verdict
  into the `⚠N` count and shows no detail.
- Wire placement follows from struct composition, not from two encoders: `SummaryEnvelope` **embeds**
  `SummaryJSON` (top-level `graph` for `status --json`), and `SpaceSummaryJSON.Summary` is a
  `*SummaryJSON` (`spaces[].summary.graph` for `--all`). One `GraphHealthJSON` type serves board and
  both summary shapes.
- `Service.Lint()` (`service.go:414-433`) deliberately does **not** use `s.taskGraphs`: it needs
  bodies, so it reads `ListTasksWithBodies()` and builds the same strict projection via
  `NewTaskGraph`. That is a documented second task read, not a hidden one, and both reads land on the
  same `*store.FS` in every production composition root (`cli/root.go:422`,
  `workspacestore/fs.go:28-31`). Settled; the portability caveat is under *source-supported risks*.

### Consumer and composition inventory

Classification: **P** production · **T** test double · **G** generated artifact.

| symbol | uses | classification |
| --- | --- | --- |
| `core.Summary` (struct) | `wire.ToSummaryJSON`/`ToSummaryEnvelope`; `render.SummaryHuman`/`SummaryJSON`/`renderCompactSpaceSummary`; `tui.dashboard.setSummary`; `tui.statsFor`; `core.SpaceSummary.Summary` | **P** — 8 consumers; the 2 new fields are read by 5 of them, the rest are epic/audit-only paths |
| `Service.Summary()` | `cli/status.go:117`; `tui/commands.go:70` | **P** — exactly 2 call sites (9 further test callers) |
| `summarize()` | `Service.Summary` (`service.go:286`); `summarizeSpaceGroup` (`space_overview.go:114`) | **P** — 2 call sites, both now passing an explicit graph source |
| `SummaryStore` | `summarize` parameter; embedded in `PlanningSummarySource` | **P** — 2 uses; no production type names it directly any more. `ListTasks` removal is compile-enforced (proved by mutation M1) |
| `PlanningSummarySource` | `SpaceOverviewStore.OpenPlanningStore` return type; `spacestore.FS.OpenPlanningStore` | **P** 2 · **T** 2 (`core.fakeSpaceOverviewStore`, `tui.atlasTestAdapter`) — both doubles hand back `*store.FS` or an FS-shaped fake, i.e. they conform trivially (see L3) |
| `OpenPlanningStore` | `spacestore.FS` (**P**, sole production impl); called once at `space_overview.go:109`; `core.fakeSpaceOverviewStore` (**T**), `tui.atlasTestAdapter` (**T**) | **P** — 1 impl, 1 call site |
| `TaskGraphSource` | `Service.taskGraphs`; `WithTaskGraphSource`; `WorkspaceSource.TaskGraphs`; `loadTaskGraphRecords`; `LoadTaskGraph`; `PlanningSummarySource`; `store.FS` compile assertion. Read use cases: `Board` (`board.go:43`), **`Summary` (new)**, `TaskList` (`service_task.go:46`), Thread reads (`service_thread.go:218,368`), Thread apply (`service_thread_apply.go:23`), `dependency_operations.go:435` | **P** — 7 read use cases; `Summary` is the newly joined one |
| `taskStoreGraphSource` | `NewService` fallback for a `Store` without `ReadTaskGraph` | **P** — 1; keeps non-conforming aggregate stores working, so the port widening is not a breaking change for embedders |
| `wire.SummaryJSON` | `ToSummaryJSON`; embedded in `SummaryEnvelope`; pointer field of `SpaceSummaryJSON` | **P** — 3; this composition is *why* one field yields both the top-level and the nested placement |
| `wire.GraphHealthJSON` | `toGraphHealthJSON` (sole constructor); `BoardEnvelope.Graph`; `SummaryJSON.Graph` | **P** — 3; shared object, single construction point |
| `toGraphHealthJSON` | `ToBoardEnvelope`; `ToSummaryJSON` | **P** — 2; the healthy-nil / `omitempty` rule exists in exactly one place, so board and status cannot diverge on omission |
| `render.taskGraphWarning` | `BoardHuman`; `SummaryHuman` | **P** — 2 |
| `render.SummaryHuman` / `render.SummaryJSON` | `cli/status.go` | **P** — 2 |
| dashboard summary rendering | `tui/dashboard.go:187`, inside `setSummary`'s "needs attention" block | **P** — 1; also clears `allClear` |
| Atlas attention/detail | `tui/atlas.go:864` (`statsFor` → `attention`) via `graphAttention` (`atlas.go:876`) | **P** — 1. Atlas carries **no detail text** by design (`atlas.go:846-851`: one `⚠N` column, "the space's own overview is where you dig in") |
| `wire.SchemaVersion` | one `const` (`wire.go:257`); 50 production `SchemaVersion:` assignments; **0** consumers compare against a literal version (`grep -E 'SchemaVersion\s*==\|"1\.[0-9]+"'` over non-test production code returns only the declaration) | **P** — single declaration; the bump reaches every envelope mechanically |
| schema encoders / generated artifacts | `wire.JSONSchema()` → `schema --json-schema`; `wire/schema_comments.json` (**G**, 1 new entry); 30 golden envelopes + `schema_jsonschema.golden` (**G**) | **G** — verified to differ from base only by the version string plus two `graph` property blocks |

### Acceptance-criteria traceability (task `6g697mp8s4tx`)

| criterion | verdict | evidence |
| --- | --- | --- |
| `status` names a degraded or broken graph, its cause and remedy, matching `board`'s wording | **met** | One shared `render.taskGraphWarning`; on the unreadable-file tree, `status` and `board` emit byte-identical verdict text and identical JSON `detail`. Confirmed across all 11 probe trees |
| `lint` cannot report a clean repository while the graph guard would refuse a mutation on it; a test asserts the two agree | **met in substance; thinly pinned** | All 11 non-healthy trees either exited 11 or exited 0 with the advisory issue printed; **none** printed `✔ all planning entities and dependency links pass lint`. The asserting test covers one state (resolved legacy). Structural caveat under *source-supported risks* |
| resolvable-legacy keeps advisory severity and exit zero | **met** | `TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` unchanged in intent and passing; live probe on a fully-populated tree: `1 advisory finding(s)`, exit 0, `"severity":"advisory"` in `lint --json` |
| a missing or ambiguous legacy field is reported by `lint` once — not twice, and not silently | **behaviour met; the pin does not hold it** | Live probes show exactly one line per defect. **M1**: the test added for this criterion passes while the defect prints twice |
| `--json` carries the graph verdict on every surface that renders it | **met** | `status --json` top-level `graph`; `status --all --json` `spaces[].summary.graph`; `board --json` `graph`; healthy omits in all three; all validated against live `schema --json-schema` |

### Hostile-evidence ledger

**A. End-to-end probes — 13 throwaway trees × (`status`, `status --json`, `lint`, ordinary
dependency mutation).** Every `TaskGraphProblemCode` is covered; each row's lint owner is named, and
in every row exactly one owner fires.

| tree | `status` health | exit | `lint` owner (exactly one) | lint exit | `task depend add --on` |
| --- | --- | --- | --- | --- | --- |
| healthy | *(omitted)* | 0 | — | 11 *(unrelated field nags)* | accepted |
| empty repo | *(omitted)* | 0 | `✔ … pass lint` | 0 | n/a |
| resolved legacy (`blocked_by`) | degraded | 0 | grouped legacy diagnostic, **advisory** | 0 on a fully-populated tree | refused: "planned dependency state is degraded" |
| empty legacy field (`blocked_by: []`) | degraded | 0 | grouped legacy, "field is present but empty", advisory | 0 | refused |
| missing canonical dep | broken | 0 | `depends_on: … depends on missing task …` | 11 | refused |
| legacy ref missing | broken | 0 | grouped legacy, blocking | 11 | refused |
| legacy ref ambiguous | broken | 0 | grouped legacy, blocking | 11 | refused |
| unreadable task file | broken | **11** | `FileProblem` (stderr) | 11 | refused |
| non-id-led file in `tasks/` | broken | **11** | `FileProblem`, "move it to meta/" | 11 | refused |
| duplicate task id | broken | 0 | `id: duplicate stable task id …` (both files) | 11 | refused |
| missing task id | broken | 0 | `id: missing stable id — lint --fix assigns one` | 11 | refused |
| id/filename drift | broken | 0 | `id: frontmatter id … disagrees with the filename id …` | 11 | refused |
| invalid status | broken | 0 | `status: frontmatter status missing or unrecognized …` | 11 | refused |
| cycle | broken | 0 | `depends_on: dependency cycle: …` on each member | 11 | refused |
| mixed (ambiguous + unreadable + cycle) | broken, `(3 additional problem(s))` | **11** | all three owners fire, each once | 11 | refused |

Reading of this table: graph health never turns a read-only `status` into a failure — the only
non-zero `status` exits come from the pre-existing unreadable-file contract, which is unchanged.
The mutation guard refuses **every** degraded and broken state independently of lint's exit policy.

**B. Split-authority and portability probes** (temporary `internal/core` test, since removed).

- A `composite{metaOnly, pathlessSource}` where the aggregate half would report a contradictory task
  set: `ReadTaskGraph` called **exactly once**; counts, in-progress rows, `Problems`, health and
  detail all came from the graph snapshot. The repository's own
  `TestService_SummaryUsesInjectedTaskGraphSnapshot` and `TestSpaceOverviewUsesPlanningStoresGraphSnapshot`
  encode the same contradiction fixture, and mutation M1 proves they bite.
- **Pathless / remote-shaped adapter**: a `TaskGraphLoadProblem{TaskID, TaskSlug, Message}` with no
  `Path` produced `health=broken` with a correct detail and **no fabricated path** in
  `Summary.Problems` (`Path` stayed `""`). The identity loss is exactly the recorded debt in
  `6g6jqqcdehne`; I found no data loss beyond it and do not re-file it.
- **Error from each composed capability**: a `ReadTaskGraph` error aborts `summarize` before any
  non-task read (`epicCalls == 0`) and propagates verbatim; `ListEpics` and `ListAuditsWithFindings`
  errors each propagate verbatim. No partial success is presented as healthy.
- **Health/detail agreement**: for healthy / degraded / broken fixtures, `Summary.GraphHealth`
  equalled `NewTaskGraphRead(read).Health()`; `healthy ⟺ ValidateTaskGraphMutationPlan` succeeds;
  non-healthy always carried a non-empty detail; healthy never acquired one.
- **Detail stability**: all 6 permutations of a 3-task multi-fault input produced an identical
  `GraphDetail`. `sortGraphProblems` (`dependency_graph.go:1293`) sorts on
  `TaskID∙Code∙Field∙RelatedTaskID∙Path∙Message`, so "first cause" is deterministic (lowest task id,
  not highest severity — see *settled concerns*).
- **Nil capabilities**: `NewService(nil).Summary()` returns
  `task graph reads are unavailable from this store` instead of panicking (a strict improvement:
  the old path would have called `ListTasks` on a nil `SummaryStore`).

**C. Multi-space probes.** Registry with 6 spaces: healthy, degraded, broken-missing-dep,
broken-ambiguous, broken-unreadable, and a `vanishing` space whose checkout was deleted.

- Each space's verdict is independent: `spaces[].summary.graph` was `null / degraded / broken /
  broken / broken / null` respectively, and human `status --all` printed the right `! task graph …`
  line under each.
- The unavailable space carried `"error": "no healthy entry point"` with a `null` graph and did not
  suppress, delay, or relabel any other space's verdict.
- Forcing a **read failure** (`chmod 000` on one space's `tasks/`) turned exactly that space into
  `error: … permission denied` with a `null` graph; the other five kept their verdicts unchanged,
  including the two broken ones. One space's failure cannot hide or relabel another's.
- Recovery: restoring permissions restored that space's summary. `SpaceOverviewService.Overview()`
  holds no cache, so there is no stale-verdict class in core; the atlas's last-good retention
  (`atlas.setOverview` clears `loadErr`/`stale` on every successful load, `atlas.go:384-421`) applies
  to the whole overview, and `Overview()` only fails on registry decode — so a per-space failure can
  never leave a stale *graph* warning standing.
- Selection/sorting stability: `setOverview` re-keys the cursor on the previously selected space and
  entry (`atlas.go:386-410`) before `sortSpaces`/`restoreCursor`; adding graph attention changes only
  the `⚠N` column, not the row identity used for restoration.

**D. JSON / schema compatibility.**

- **Placement**: `status --json` → top-level `graph`; `status --all --json` →
  `spaces[].summary.graph`; `board --json` → top-level `graph`. Exactly as documented.
- **Omission**: healthy and empty repositories emit **no** `graph` key at all (not `null`, not `{}`)
  in every one of the three shapes. `detail` is `omitempty` and `health` is required.
- **Vocabulary**: only `degraded` and `broken` are ever emitted — `toGraphHealthJSON` returns `nil`
  for both `""` (zero value) and `GraphHealthy`, so a zero-value `Summary` cannot leak a
  `{"health":""}` object.
- **Structural conformance**: live envelopes from 6 trees plus `--all` validated against the live
  `schema --json-schema` — no undeclared keys under `additionalProperties:false`, no missing
  `required`. `graph` is optional in `SummaryEnvelope`, `SummaryJSON` and `BoardEnvelope`;
  `GraphHealthJSON` requires only `health`.
- **Compactness**: every `--json` envelope is a single line; `EncodeJSON` is unchanged. Pretty output
  is not offered by the CLI, so "pretty vs compact" reduces to the schema-generation path, which is
  the indented `schema --json-schema` document — checked separately and drift-free.
- **Version policy**: one `const SchemaVersion` feeds 50 envelopes; **no** consumer switches on an
  exact version, so the broad golden churn is a mechanical consequence of a single additive,
  `omitempty` field — the correct outcome for an additive change under this repo's policy, not
  churn masking a breaking change. Every one of the 30 unrelated golden diffs is version-string-only;
  the sole content additions are the two `graph` property blocks in `schema_jsonschema.golden`.

**E. Presentation honesty.**

- `--color=never` renders the verdict as plain text on every surface; no colour-only signal.
- TUI dashboard at widths 40/60/80/120: the verdict row appears at every width, no line exceeds the
  terminal width, and `all clear` is correctly suppressed. At 40 columns the row truncates to
  `⚠ task graph broken: unreadable task …` — the remedy is lost, but that is the standard row
  truncation applied to every "needs attention" line, not a graph-specific defect.
- Empty repository and zero-value `core.Summary{}`: `✔ all clear`, no `task graph` text — the zero
  `GraphHealth` does not read as non-healthy.
- Placement: on `status` the warning sits directly above the unreadable-file line and below the
  counts/epics/audits body, matching `board`'s existing position.
- Recovery is real, not just cosmetic: running the advertised `task depend migrate` on a degraded
  tree cleared `blocked_by`, wrote `depends_on`, and `status` afterwards showed no warning while
  `lint` printed `✔ all planning entities and dependency links pass lint` with exit 0.

**F. Restored-mutation ledger.** All rows run in the sandbox against the baseline checkpoint; the
sandbox was verified clean (`git status --short` empty) after every restore.

| # | mutation | outcome | killed by |
| --- | --- | --- | --- |
| M1 | `Summary` counts tasks from `Store.ListTasks` while health comes from `TaskGraphSource` (re-adding `ListTasks` to `SummaryStore`) | **killed** | `TestService_SummaryUsesInjectedTaskGraphSnapshot`, `TestSpaceOverviewUsesPlanningStoresGraphSnapshot` |
| M2a | drop the `TaskGraphSource` half of `PlanningSummarySource` (usage unchanged) | **killed at compile time** | `space_overview.go:113:43: … PlanningSummarySource does not implement TaskGraphSource (missing method ReadTaskGraph)` |
| M2b | drop that half **and** rediscover it with `planningStore.(TaskGraphSource)` | **SURVIVED** — builds, full suite green | nothing → finding **L3** |
| M3a | suppress the graph warning in human `status` | **killed** | `TestStatusReportsGraphDegradationWithoutTurningReadIntoFailure` (both subtests), `TestSummaryReportsNonHealthyGraphInHumanAndJSON` |
| M3b | suppress the graph row in the TUI dashboard | **killed** | `TestDashboardNeedsAttentionReportsGraphDegradation` |
| M3c | suppress the graph contribution to Atlas attention | **killed** | `TestAtlasAttentionFoldsOnlyWhatWantsAPerson` |
| M3d | suppress the verdict in the `status --all` compact space summary | **killed** | `TestStatusAll_GroupsEntryPointsAndIsolatesBrokenOnes`, `TestStatusAllHumanReportsSpaceGraphHealth` |
| M4a | move the summary graph key on the wire (`graph` → `summary_graph`) | **killed — by the generated schema only** | `TestGolden_MachineContract/schema_jsonschema`. No behavioural test noticed, because they all decode into the typed `wire.SummaryEnvelope`, whose tag moved with it → finding **L2** |
| M4b | emit an empty `graph` object for a healthy repository | **killed** | `TestGolden_MachineContract/status_json`, `/board_json`, `TestGolden_StatusAllJSON`, `TestSummaryReportsNonHealthyGraphInHumanAndJSON` |
| M5a | revert the schema bump to 1.59 | **killed** | `TestGolden_MachineContract` (14+ subtests across every envelope) |
| M5b | leave one representative golden (`task_list_json`) on 1.59 | **killed** | `TestGolden_MachineContract/task_list_json` |
| M6a | print the legacy missing/ambiguous defect twice (remove the two codes from the `dependencyLintIssues` skip list, adding a second owner) | **SURVIVED** — builds, **entire** `go test ./...` green | nothing → finding **M1** |
| M6b | drop the grouped legacy diagnostic entirely (the false-clean route) | **killed** | `TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce`, `TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero`, `TestLintReportsLegacyAndCanonicalDependencyDefects`, `TestLintUnsafeLegacyDependencyRemainsValidationError` |
| M7 | weaken the ordinary mutation guard (`MutationReady` accepts degraded) because lint calls resolved legacy debt advisory | **killed** | `TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` (the assertions commit `1ab44d0` added), `TestTaskGraphLegacyResolutionHealthAndDirection` |

Restoration: every mutation was reverted with `git checkout -- .` and the sandbox re-verified clean;
two temporary probe test files (`internal/core/zz_probe_review_test.go`,
`internal/tui/zz_probe_review_test.go`) were deleted after use. No generated artifact was left
regenerated — `docs/cli` was regenerated once for the drift check and produced no diff.

### Demonstrated defects

The four findings below. All are coverage or documentation defects; none changes what the tool does
today.

### Source-supported risks (not filed as findings)

1. **The "lint can never report clean on a guard-refusing repository" property is a coincidence of
   five independent owners, not a structural invariant.** `runLint` prints
   `✔ all planning entities and dependency links pass lint` when `len(results)==0 && len(problems)==0`
   (`cli/lint.go:65`) — it never consults `graph.Health()`. Non-healthy therefore implies non-clean
   only because each skipped code happens to have a separate owner (`ProblemUnreadable`→FileProblem,
   `ProblemMissingTaskID`→`domain.MissingIDIssue`, `ProblemTaskIDDrift`→`domain.IDDriftIssue`,
   `ProblemInvalidStatus`→`domain.FrontmatterStatusIssues`, legacy→grouped diagnostic) *and* because
   `dependencyLintIssues` silently `continue`s on `problem.Path == ""` (`service.go:583`) which no FS
   record can produce. I probed all 11 FS-reachable states and the property holds in every one. In a
   constructed pathless snapshot it does **not**: a broken graph with two faults yielded one lint
   issue keyed under the empty path and dropped the other entirely. `Service.Lint` is FS-bound today
   (`ListTasksWithBodies`), so this is unreachable — but it is the class of latent defect the
   portability push makes reachable later, and it is adjacent to, not covered by, `6g6jqqcdehne`
   (which scopes Board/Summary diagnostics, not lint's drop-on-empty-path).
2. **`renderCompactSpaceSummary` duplicates the warning wording instead of calling
   `taskGraphWarning`** (`render/status.go:210` vs `render/render.go:390`). The strings agree today;
   nothing enforces it. The `--all` surface is the one place `status` and `board` could drift.
3. **A split `WorkspaceSource` would make `status` and `lint` disagree.** `Summary` reads
   `s.taskGraphs`; `Lint` reads `s.store`. `workspacestore.FS` puts the same `*store.FS` in every
   field, so they agree in production — but `WorkspaceSource`'s own doc comment invites split
   adapters, and nothing pins the two task corpora to one another.

### Unverified concerns (stated, not resolved)

1. **Remote latency and cancellation.** `TaskGraphSource` has no context parameter, so a slow remote
   adapter blocks `status`/the TUI dashboard indefinitely. I could not test a real remote adapter —
   none exists — and this predates the change; I raise it only because this change adds `Summary` to
   the list of read paths that would block.
2. **Terminal-level rendering.** Dashboard behaviour was exercised through `d.view(...)` with stripped
   ANSI, not a real TTY; I did not drive the interactive TUI, so focus/scroll interaction with the new
   "needs attention" row is unverified beyond the rendered string.

### Settled concerns (challenged, then closed with evidence)

- **"Is there a second task scan?"** No. `SummaryStore` no longer declares `ListTasks`, and M1 shows
  the tests catch a reintroduction. `store.FS.ReadTaskGraph` delegates to a single `ListTasks` scan.
  Cross-entity skew between the task read and the epic/audit reads is real but unavoidable and
  unclaimed — the doc comment scopes the shared-snapshot claim to "Counts/InProgress", which is
  exactly what holds.
- **Does the Atlas double-count one corrupt file?** Yes — attention = 2 for a single unreadable task
  (`len(s.Problems)` + `graphAttention`). This is deliberate and pinned:
  `TestAtlasAttentionFoldsOnlyWhatWantsAPerson` asserts `6 = 2 acute + 1 closable + 1 revisit + 1
  unreadable + 1 graph verdict`, and `atlasStats`' comment declares the number a deliberately coarse
  magnitude. `status` shows the same file twice too (verdict line + `! 1 unreadable file(s)`), which
  matches `board`'s pre-existing behaviour exactly. Not a defect.
- **Does a nil or internally-nil graph capability panic?** A struct embedding a nil
  `TaskGraphSource` panics — but `isNilCapability` is a value check, and `WorkspaceSource`'s doc
  comment already states "Every non-nil interface must wrap an operational implementation rather than
  delegating through an internally nil value." No production composition root does this
  (`workspacestore/fs.go:28-31` and `cli/root.go:422` both pass one concrete `*store.FS`), and
  directly-passed nil interfaces and typed-nil pointers *are* handled. Straw man; closed.
- **Is "first cause" chosen from an unstable order?** No. `sortGraphProblems` is a total order over
  six fields and `resolveLegacyDiagnostics` walks the stably-sorted record list; 6 input permutations
  gave an identical detail. The selection is by lowest task id rather than by severity — so a
  mixed-fault repository can headline an ambiguous legacy ref while an unreadable file waits at
  position 4 — but the count suffix (`(3 additional problem(s))`) and the separate unreadable line
  keep it truthful, and the remedy text is generic and correct.
- **Is degraded ever reported as broken, or vice versa?** No. `newTaskGraph` sets broken when
  `len(problems) > 0`, degraded when only `len(legacy) > 0`. Missing/ambiguous/unsafe legacy refs add
  a `GraphProblem`, so they correctly read broken; only exactly-resolved (or present-but-empty)
  legacy fields read degraded. Confirmed on live trees for both.
- **Do the FS-backed test doubles mask a capability break?** `atlasTestAdapter` and
  `fakeSpaceOverviewStore` do conform trivially — but the composite's teeth are the compiler, and
  M2a shows the compiler bites. What the doubles cannot pin is the *shape* of the interface itself,
  which is finding L3.
- **Did the versioning policy get followed, or does the bump just paper over a change?** Followed.
  The field is additive, `omitempty`, absent for healthy, and no consumer switches on an exact
  version. The 30-file golden churn is version-string-only.
- **Planning truthfulness.** The closeout's central claim — "the suspected lint gap was not present;
  grouped `LegacyDependencyDiagnostic` rendering already owns missing and ambiguous references" — is
  correct and I verified it directly. The ADR-0006 amendment ("Graph health alone remains
  informational on these read surfaces; `lint` is the validation gate, while unreadable files retain
  the existing partial-result/non-zero dashboard contract") matches observed behaviour exactly,
  including the unreadable-file carve-out. `ARCHITECTURE.md`'s composite claim matches the code. The
  follow-up `6g6jqqcdehne` is correctly scoped (identity through Board/Summary/SpaceOverview, not a
  competing type), correctly sequenced (`depends_on: [6g5vm4efjcdv, 6g697mp8s4tx]` — after both the
  lint-diagnostic prerequisite and this task), and the reverse edge on
  `6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs` (`depends_on` now includes
  `6g697mp8s4tx`) is right, since a repair path should follow the report that motivates it. The
  production planning tree lints clean and `status --json` reports no `graph` object, so the change
  is dogfooded. The one materially over-strong sentence is in README, and it is over-strong in the
  *safe* direction — it scopes `task depend migrate` to "the advisory legacy case", which is more
  accurate than the tool's own lint output (finding L1).

## Findings

Grouped by severity. No finding is pre-resolved; each is `open` for the implementation owner to
triage with `tskflwctl audit finding`.

### Medium

#### M1. The lint ownership regression test cannot observe the duplication it was added to prevent · **Status:** fixed

`TestLintReportsMissingAndAmbiguousLegacyReferencesExactlyOnce`
(`internal/cli/lint_test.go:141`, added by commit `1ab44d0` "test(lint): pin legacy graph diagnostic
ownership") is the pin for the acceptance criterion "a legacy field that is missing or ambiguous is
reported by `lint`, once — not twice". It cannot detect a second owner, because it counts
occurrences of the **grouped renderer's own phrasing** rather than of the defect:

```go
for _, want := range []string{`"gone" has no exact task ID or slug match`, `"same" is ambiguous across`} {
    if got := strings.Count(out, want); got != 1 { … }
}
if strings.Count(out, "legacy dependency field") != 1 || strings.Contains(out, "pass lint") { … }
```

The competing owner — the raw graph-problem loop in `dependencyLintIssues`
(`internal/core/service.go:579`) — renders the same defect as
`legacy blocked_by reference "gone" on task 6g0000000002 has no exact task ID or slug match`. The
substring `"gone" has no exact task ID or slug match` never appears in it (the graph message
interposes `on task <id>`), nor does the phrase `legacy dependency field`. All three assertions
therefore stay at exactly 1 while the defect is printed twice.

Demonstrated (mutation **M6a**, run and restored in the sandbox): removing `ProblemLegacyMissing,
ProblemLegacyAmbiguous` from the `dependencyLintIssues` skip list — precisely the ownership
regression the commit exists to prevent — leaves the **entire** `go test ./...` green, including
this test, while `lint` emits:

```
  blocked_by: legacy blocked_by reference "nowhere" on task 6g0000000002 has no exact task ID or slug match
  blocked_by: legacy dependency field: "nowhere" has no exact task ID or slug match; run `tskflwctl task depend migrate`
```

and, for the ambiguous case:

```
  blocked_by: legacy blocked_by reference "same" on task 6g0000000002 is ambiguous across task IDs 6g0000000001, 6g0000000003
  blocked_by: legacy dependency field: "same" is ambiguous across 6g0000000001, 6g0000000003; run `tskflwctl task depend migrate`
```

Production behaviour is correct today — the skip list is intact and every live probe printed each
defect once. What is wrong is the guard rail: the one thing standing between this criterion and a
silent regression does not stand. A count keyed on the *field-and-task* pair rather than on the
renderer's wording would hold — e.g. asserting that the number of `blocked_by:` issue lines for the
dependent task is 1, or asserting the count of the task id plus the reference value together.

Severity medium rather than high because nothing user-visible is broken now; it is a coverage defect
on a criterion this branch reports as met.

**Resolution:** Strengthened the regression at the core result boundary: it now
requires exactly one blocked_by issue for the dependent task. Temporarily
restoring the competing raw graph-problem owner produces three issues and fails
this test, proving the intended mutation is killed.

### Low

#### L1. Lint tells the user to run `task depend migrate` for missing and ambiguous references, which that command always refuses · **Status:** fixed

The grouped legacy diagnostic appends one fixed remedy to every resolution class
(`internal/core/service.go:617`):

```go
Message: fmt.Sprintf("legacy dependency field: %s; run `tskflwctl task depend migrate`", message),
```

For a *resolved* reference that is right, and I verified the round trip end to end (migrate rewrote
`blocked_by` to `depends_on`, after which `status` was quiet and `lint` printed
`✔ all planning entities and dependency links pass lint`, exit 0).

For a **missing** or **ambiguous** reference it is a dead end. Those classes add a `GraphProblem`, so
the repository is `broken`, and `ValidateTaskGraphMutationSource` refuses before migration can start:

```
$ tskflwctl -C <tree> lint
  blocked_by: legacy dependency field: "nowhere" has no exact task ID or slug match; run `tskflwctl task depend migrate`

$ tskflwctl -C <tree> task depend migrate --dry-run
error: validation failed: repository task graph is broken: legacy blocked_by reference "nowhere" on
task 6g0000000002 has no exact task ID or slug match …; repair the graph-owned frontmatter directly,
then run `tskflwctl lint`
```

Same for the ambiguous tree. The correct remedy is the one `status`/`board` already print for the
identical repository — "repair the graph-owned frontmatter directly, then run `tskflwctl lint`" — so
this branch's own new surface contradicts the older one, and the README paragraph added in `198d585`
gets it right too ("Run `task depend migrate` for the advisory legacy case").

The line itself predates this branch and is not in the diff. I raise it because this task's stated
scope is "reconciling lint's advisory/exit-zero treatment of legacy dependency fields", because the
branch adds a regression test that pins this exact output, and because the branch simultaneously
documents the correct narrower remedy in prose — leaving the tool as the least accurate of the three
voices. Selecting the remedy per resolution class (migrate for `LegacyResolved`, direct repair for
`LegacyMissing`/`LegacyAmbiguous`/`LegacyUnsafe`) is a contained change in the same loop that already
switches on `ref.Resolution`.

**Resolution:** Added LegacyDependencyDiagnostic.MigrationReady and made lint
choose remedies from that semantic result. Resolved and present-empty fields
recommend guarded migration; missing, ambiguous, and structurally unsafe fields
now require direct frontmatter repair followed by lint. The graph-query renderer
uses the same classification.

#### L2. The new `graph` wire object is never exercised by the in-suite JSON Schema validator · **Status:** fixed

`TestJSONSchema_ValidatesRealOutput` (`internal/wire/envelopes_test.go:23`) is the repository's
round-trip proof that the emitted schema validates real `--json` output, and it already carries a
real validator. Its `SummaryEnvelope` and `StatusAllEnvelope` cases both build a `core.Summary` with
no `GraphHealth` (`envelopes_test.go:173-195`), so the new optional object is never validated there —
schema 1.60's addition is the one part of 1.60 the validator does not see.

Every behavioural test decodes into the typed `wire.SummaryEnvelope`, so a JSON **key** change moves
on both sides and stays invisible to them. Mutation **M4a** confirms this: renaming
`json:"graph,omitempty"` to `json:"summary_graph,omitempty"` was caught **only** by
`TestGolden_MachineContract/schema_jsonschema` — a generated-artifact golden, which happens to pin
the name as a side effect. `TestStatusReportsGraphDegradationWithoutTurningReadIntoFailure`,
`TestSummaryReportsNonHealthyGraphInHumanAndJSON` and
`TestStatusAll_GroupsEntryPointsAndIsolatesBrokenOnes` all stayed green.

So the wire name is pinned, but incidentally, by an artifact whose job is documentation rather than
contract enforcement. I validated live envelopes against `schema --json-schema` by hand — non-healthy
`status --json` (top-level), non-healthy `spaces[].summary.graph`, healthy omission, and
`GraphHealthJSON`'s `required: [health]` — and found no violation; that manual check should be a test.
Adding a degraded `core.Summary` to the existing `SummaryEnvelope`/`StatusAllEnvelope` cases costs two
lines and closes both gaps at once.

**Resolution:** Populated the real JSON Schema validator's SummaryEnvelope and
StatusAllEnvelope cases with degraded and broken graph objects, so both optional
placements now validate non-default health and detail values against schema
1.60.

#### L3. The explicit `PlanningSummarySource` composite has no pin of its own · **Status:** fixed

`space_overview.go:10-17` states the invariant plainly — "TaskGraphSource is explicit rather than
discovered with a type assertion: every adapter must provide the same strict task snapshot contract"
— and it is the load-bearing claim of the cross-space half of this change, repeated in
`docs/ARCHITECTURE.md` and in the task closeout.

Nothing pins it. The compiler catches the *naive* removal only as a side effect of the call shape:
mutation **M2a** (drop `TaskGraphSource` from the interface, leave
`summarize(planningStore, planningStore, asOf)` alone) fails to build with
`space_overview.go:113:43: … PlanningSummarySource does not implement TaskGraphSource`. But mutation
**M2b** — drop the embedded interface *and* rediscover the capability with
`graphs, _ := planningStore.(TaskGraphSource)` — builds cleanly and leaves the **entire** test suite
green, reverting the invariant to exactly the optional runtime discovery the comment forbids. The two
test doubles (`core.fakeSpaceOverviewStore`, `tui.atlasTestAdapter`) cannot help: both hand back an
`*store.FS` or an FS-shaped fake, so they satisfy any shape the interface takes.

The blast radius of the mutated form is bounded — a non-conforming adapter would surface as a
per-space `LoadError` rather than a silent healthy verdict — which is why this is low, not medium. A
one-line compile-time assertion in the package, e.g.

```go
var _ TaskGraphSource = PlanningSummarySource(nil)
```

would pin the documented invariant directly rather than relying on an unrelated call site keeping it
honest.

**Resolution:** Added explicit compile-time assertions that
PlanningSummarySource satisfies both SummaryStore and TaskGraphSource.
Coordinated removal plus runtime rediscovery now fails at the invariant pin
rather than surviving behind the immediate call shape.

## Candidate tasks

Deferral is not warranted for any of these: M1, L2 and L3 are each a small, self-contained test edit
in code this branch already touches, and L1 is a per-class remedy selection inside a loop that
already switches on `ref.Resolution`. All four are cheaper to fix now than to carry. No task
commands suggested; the implementation owner should triage with `tskflwctl audit finding`.
