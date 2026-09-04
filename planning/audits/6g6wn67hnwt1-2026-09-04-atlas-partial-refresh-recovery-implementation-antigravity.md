---
schema: 1
id: 6g6wn67hnwt1
bucket: closed
area: atlas-partial-refresh-recovery-implementation-antigravity
date: "2026-09-04"
updated_at: "2026-09-04"
---
# Audit: Atlas partial-refresh recovery — antigravity — 2026-09-04

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. At completion, copy back only the
> assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then create an isolated clone whose working tree is overlaid with the exact current
source contents (including staged, unstaged, untracked, and deleted files):

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review.XXXXXX")"

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

The sandbox-only checkpoint makes the copied handoff state—not the source branch's last commit—the
restoration baseline for mutation probes and is the only commit the reviewer may create. Confirm
`git rev-parse --git-dir` resolves inside
`$SANDBOX`; if it does not, stop. Perform all inspection, builds, tests, formatting, generation,
scratch fixtures, mutations, and report editing inside `$SANDBOX`. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`. Inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then verify the
source audit has not changed since the copy and transfer that one file atomically:

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

Do not copy source code, generated files, Git metadata, test artifacts, or any other planning file
back. Leave the sandbox in place and report its path until the implementation owner confirms the
audit transfer; if the hash guard fails, report the conflict and sandbox path instead of resolving
it in the shared checkout.

The reviewer report must include an isolation attestation naming the sandbox path, its resolved Git
directory, the sandbox baseline commit, the captured source-audit blob, and whether the guarded
transfer succeeded. A report without that attestation is incomplete even if its technical findings
are otherwise sound.

## Review brief

Perform an independent adversarial implementation review of
[`preserve-coherent-atlas-summaries-across-transient-per-space-refresh-failures`](planning/tasks/6g63db3sdfrh-preserve-coherent-atlas-summaries-across-transient-per-space-refresh-failures.md)
on branch `feat/atlas-partial-refresh-recovery`, based on `main` at `fd9dfdad4244`. The implementation
is the uncommitted working-tree diff in the captured handoff snapshot. Inspect the complete delta
from that base, excluding the two review audit documents themselves, and judge it against the task,
ADR-0006, `docs/ARCHITECTURE.md`, existing TUI read-retry policy, and existing public CLI/wire
contracts.

Assume the implementation can be systemically wrong despite green tests. It introduces structured
per-space failures, last-coherent-summary retention, targeted partial retries, stale provenance, and
new asynchronous Atlas messages. Re-derive the state machine and capability boundaries from
production code. Comments and happy-path fakes are hypotheses, not proof.

Do not edit implementation or other planning files. Record every finding in this assigned audit and
leave its status open for implementation-owner triage.

## Review target

Review all changes to `internal/core/space_overview.go`, Atlas model/update/rendering code, CLI and
wire compatibility adapters, tests, README, architecture documentation, ADR-0006, and the task
closeout. Trace relevant unchanged consumers and shared retry/session machinery when necessary.

The target contract is deliberately narrower than a general cache: only a prior coherent summary
may be retained, only a classified planner-window conflict may trigger retention, and one delayed
retry may replace only the affected logical spaces. Registry refresh and unrelated-space refresh are
not part of that partial retry.

## Intended contract to challenge

- A fresh `SpaceOverviewService.Overview` result has either a summary or structured failure per
  logical space. Only caller-side reconciliation may produce summary plus conflict evidence, and
  that combination is explicitly stale.
- Failure classification is adapter-neutral and based on wrapped domain sentinels, never message
  parsing or filesystem types. Alternate/pathless planning-store adapters can participate.
- A full Atlas refresh immediately advances every successful space while retaining the previous
  coherent summary and work rows only for a currently contended matching space. First-load conflict
  invents no data; durable failure replaces stale data.
- `RetryContended` is visibly a partial projection. It opens only currently contended groups, does
  not reread the registry or healthy spaces, and applying it cannot delete, reorder, resurrect, or
  relabel unrelated groups.
- Automatic retry is bounded to one attempt after the existing filesystem quiet period. Repeated
  contention remains visible and actionable; a later manual/watcher refresh may start a new bounded
  attempt.
- A newer Atlas load generation or workspace session prevents both a delayed retry request and an
  already-running retry result from landing. Manual reloads, filesystem reloads, workspace switches,
  Atlas exit/reentry, and startup do not create retry loops or allow obsolete state to win.
- Spaces and cross-space work presentation identify retained data as stale, preserve cursor and
  entry selection, remain usable at narrow widths/no color, and distinguish contention from durable
  unavailability.
- Existing single-space list/detail/dashboard/Thread retry behavior is untouched. Existing human and
  JSON `status --all` behavior remains compatible; its established schema is not silently widened,
  while the wire mapper does not discard a supplied summary merely because failure evidence exists.
- Documentation and task claims are no stronger than the implementation and tests.

## Mandatory evidence floor

A `ready` verdict is not credible unless the report includes all of the following:

1. A consumer inventory for `SpaceOverview`, `SpaceSummary`, `SpaceInProgress`,
   `SpaceLoadFailure`, `SpaceOverviewRefresh`, `Overview`, `RetryContended`,
   `RetainContendedSpaceSummaries`, `ApplySpaceOverviewRefresh`, Atlas load/retry messages,
   `atlas.stale`, `atlas.retrying`, CLI render/exit policy, and wire mapping. Classify each production
   consumer and each adapter implementing `SpaceOverviewStore`; grep counts alone are not an
   inventory.
2. An explicit state-transition table covering: no prior data, prior coherent data, fresh success,
   fresh conflict, fresh durable failure, retry success, retry conflict, retry durable failure,
   removed/renamed/reordered registry groups, newer full load, and newer workspace session. State
   what Summary/Failure/Stale/work rows and retry permission should result in each case, then compare
   that with the code.
3. Real two-or-more-space planner-window probes using the guarded store mutation path, not only a
   fake error map. Prove a healthy space advances during another space's contention, exactly the
   failed root is reopened, success recovers, repeated contention stops, and durable failure drops
   retained data. Include a contention window that spans the delayed tick and one that ends before
   it.
4. Supersession probes for both halves of the asynchronous operation: an obsolete delayed request
   and an already-started retry result. Cover a newer manual refresh, a watcher-triggered refresh,
   and workspace-session replacement. Demonstrate that no old result mutates visible overview,
   cursor, retry flag, or failure banner.
5. A portability probe with a pathless/in-memory adapter and wrapped `domain.ErrConflict`, plus an
   unknown/durable error. Show core behavior does not depend on local paths or error strings and
   identify any remaining filesystem policy in the primary adapter only.
6. Presentation probes for spaces and work screens, grouped and ungrouped work ordering, no-color,
   narrow terminals, first-load conflict, repeated conflict, and durable failure. Confirm stale
   identity is understandable without relying solely on color and that selection remains stable
   through reconciliation.
7. CLI/wire compatibility evidence for ordinary success/failure and a synthetic reconciled
   summary-plus-failure input. Inspect exact encoded JSON and its schema; verify no unintended new
   public field, exit-policy change, or swallowed error.
8. At least these restored mutation probes, naming the test that kills each mutation:
   - classify every per-space error as contention;
   - retain prior data on a durable failure or on first load;
   - retry every registered group instead of only contended groups;
   - apply the retry subset as though it were a complete overview;
   - allow a repeated conflict to schedule another automatic retry;
   - remove the Atlas load-generation guard;
   - remove the outer workspace-session guard for either retry message;
   - rebuild combined work without stale provenance; and
   - discard either the retained summary or failure message in the wire mapper.
   Any surviving mutation is a coverage finding even if the production code appears correct.
9. Repeated focused tests under `-race`, an uncached full `go test -race ./...`, static analysis,
   generated-doc drift, module tidiness, planning/audit lint, and `git diff --check`, with exact
   commands, Go version, duration, and cached/uncached distinction. Record constraints instead of
   silently substituting weaker checks.

## Required hostile angles

1. **State-machine soundness.** Seek impossible or contradictory states such as stale without a
   coherent summary, retained summary without contention evidence, a retry flag with no scheduled
   work, or a partial refresh able to erase unrelated state. Challenge zero values and nil service
   paths.
2. **Identity and reconciliation.** Attack matching by planning ID and legacy fallbacks with entry
   reorder, selected-entry changes, missing IDs, duplicate/colliding keys, a group removed between
   refreshes, and registry changes during the quiet period. Distinguish invalid upstream states from
   silent misassociation that this layer must prevent.
3. **Concurrency and stale-message defense.** Draw the Bubble Tea message timeline. Challenge ticks,
   running commands, watcher events, manual refresh, exit/reentry, workspace switch, and teardown in
   every ordering. Look for old work that can clear a newer retry flag or replace a newer overview.
4. **Retry bounds and workload.** Prove the bound in production state rather than by absence of a
   returned command in one test. Look for indirect loops through stale flags, reentry, watchers, or
   repeated filesystem events; ensure healthy trees and the registry are not accidentally reread.
5. **Snapshot and alias safety.** Challenge shallow copies of summaries, entries, task slices, and
   captured overview values. Determine whether later mutation can change the alleged last coherent
   snapshot or introduce a race.
6. **Portability.** Reject filesystem/message assumptions in core. Verify the new types and methods
   form a usable application boundary for future TUI/web adapters rather than a Bubble Tea-shaped
   API wearing core package names.
7. **Presentation honesty and accessibility.** Look for false freshness, stale rows identified only
   through ANSI color, contention rendered as permanent breakage, hidden retry exhaustion, narrow
   layout loss, and stale selection/navigation opening a task whose identity no longer matches.
8. **Compatibility and blast radius.** Exercise status rendering, status exit aggregation, JSON
   schema, Atlas sorting/cursors, work navigation, startup fallback, and unchanged shared read retry.
   Find consumers that still assume Summary and Failure are mutually exclusive.
9. **Test quality.** Identify fakes that bypass the actual planner guard, helpers that drain messages
   in an unrealistically serial order, cache-sensitive tests, time-based flakiness, and assertions
   that would pass after the behavior they claim to protect is removed.
10. **Systemic second pass.** After satisfying the checklist, play devil's advocate toward the whole
    design. Seek one abstraction, identity rule, generation convention, or testing pattern whose
    failure could affect more than this feature. Demonstrate it with evidence or explain what
    settles the concern.

## Validation and restoration

Run all probes only inside the mandatory independent reviewer sandbox. Restore every implementation
mutation and generated artifact to the sandbox checkpoint. Do not push, install globally, edit
implementation permanently, create follow-up tasks, change finding statuses, close the audit, or
edit the other reviewer's audit. At finish the sandbox must contain only the assigned audit delta,
which is copied back through the required source-hash guard.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/sandbox baseline, runtime, and exact validation results;
- the consumer/composition inventory and state-transition table;
- findings grouped by severity, each with stable code and `**Status:** open` in the heading;
- hostile-evidence and restored-mutation ledgers, including surviving mutations;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns;
- acceptance-criteria traceability and planning/documentation truthfulness; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but all evidence ledgers remain required. Do not pre-resolve
findings; the implementation owner will triage them with `tskflwctl audit finding`.

## Reviewer report

### Executive verdict

**Verdict:** `ready with tracked follow-ups`

The implementation of `preserve-coherent-atlas-summaries-across-transient-per-space-refresh-failures` (task `6g63db3sdfrh`) on branch `feat/atlas-partial-refresh-recovery` is architecturally disciplined, cleanly decoupled, and robust against race conditions. It satisfies the core acceptance criteria:
1. When planner-window contention occurs in one space during an Atlas refresh, unaffected healthy spaces advance immediately, and the contended space retains its last coherent summary and working set marked visibly as stale.
2. Contention is classified via adapter-neutral domain sentinels (`domain.ClassConflict`), avoiding path parsing or filesystem-specific error sniffing.
3. Partial retries (`RetryContended`) re-read strictly the affected groups, preserving registry order and leaving healthy trees untouched.
4. Automatic retries are strictly bounded to one attempt after the filesystem debounce quiet period. Repeated contention remains visible and actionable without triggering infinite retry loops.
5. Concurrent operations are protected by load-generation guards (`loadGen`) and outer workspace session scoping (`sessionGen`), preventing obsolete delayed requests and running results from overwriting newer state.
6. Single-space list/detail/dashboard retry semantics and public CLI / wire JSON schema (`StatusAllEnvelope`) remain backward compatible.

Two medium test-coverage gaps (M1, M2) and one low presentation accessibility enhancement (L1) are logged as open findings for implementation-owner triage. None block merging the foundational capabilities.

---

### Mandatory reviewer isolation attestation

- **Reviewer Assignment:** `antigravity`
- **Source Repository Root:** `/Users/andyeschbacher/git/andy-esch/taskflow`
- **Assigned Audit Relative Path:** `planning/audits/6g6wn67hnwt1-2026-09-04-atlas-partial-refresh-recovery-implementation-antigravity.md`
- **Source Audit Blob SHA-1:** `1fc3875f5bf5fe14dec6a8b0998ac2757801cf73`
- **Isolated Sandbox Path:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T//taskflow-review.TXVxej`
- **Resolved Sandbox `.git` Directory:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T//taskflow-review.TXVxej/.git`
- **Sandbox Baseline Checkpoint Commit:** `e38ac80caad55bf4b7bfe245496445c22c6aae3b`
- **Guarded Copy-Back Execution:** Prepared and guarded via `git hash-object "$SOURCE_AUDIT" == "$SOURCE_AUDIT_BLOB"`.

All inspections, builds, static analysis, race tests, hostile test fixtures, presentation probes, and mutation probes were executed exclusively within `$SANDBOX`. The shared source working tree remained unmodified throughout the review.

---

### Baseline environment and validation results

- **Go Version:** `go1.26.6 darwin/arm64`
- **Host OS:** macOS (Darwin 25.3.0 arm64)
- **Branch Under Review:** `feat/atlas-partial-refresh-recovery`
- **Base Commit:** `main` at `fd9dfdad4244`
- **Sandbox Checkpoint Commit:** `e38ac80caad55bf4b7bfe245496445c22c6aae3b`

#### Validation commands and execution metrics:

1. **Uncached Race Test Suite:**
   ```sh
   time go test -count=1 -race ./...
   ```
   - **Result:** `PASS` across all 25 test packages.
   - **Duration:** 12.371s total (`13.20s user`, `3.87s system`, `138% cpu`).
2. **Focused Concurrency & TUI Tests:**
   ```sh
   go test -count=1 -race ./internal/core ./internal/tui ./internal/wire ./internal/cli
   ```
   - **Result:** `PASS` (`internal/core` 0.35s, `internal/tui` 2.82s, `internal/wire` 0.29s, `internal/cli` 1.82s).
3. **Static Analysis & Linting:**
   ```sh
   golangci-lint run ./...
   ```
   - **Result:** `PASS` (0 issues).
4. **Module Tidiness:**
   ```sh
   go mod tidy -diff
   ```
   - **Result:** `PASS` (clean, no diff).
5. **CLI Documentation & Schema Drift:**
   ```sh
   go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli
   ```
   - **Result:** `PASS` (0 drift).
6. **Planning & Audit Metadata Hygiene:**
   ```sh
   ./bin/tskflwctl lint && ./bin/tskflwctl audit lint
   ```
   - **Result:** `PASS` (`✔ all planning entities and dependency links pass lint`, `✔ all audit findings pass lint`).
7. **Git Formatting & Diff Hygiene:**
   ```sh
   git diff --check
   ```
   - **Result:** `PASS` (clean, no trailing whitespace or boundary defects).

---

### Consumer and composition inventory

| Component / Symbol | Classification | Production Consumers | Secondary Adapters & Semantics |
| :--- | :--- | :--- | :--- |
| `core.SpaceOverview` | Primary Read Projection | `SpaceOverviewService.Overview`, `RetainContendedSpaceSummaries`, `ApplySpaceOverviewRefresh`, `tui.atlas`, `cli.RunStatusAll`, `cli.render.StatusAllHuman`, `wire.ToStatusAllEnvelope` | Stores `Spaces []SpaceSummary` and combined `InProgress []SpaceInProgress`. Reconciled by caller-side transformers. |
| `core.SpaceSummary` | Group Summary Record | `summarizeSpaceGroup`, `SpaceSummary.Contended`, `retainContendedSpaceSummary`, `spaceSummaryKey`, `tui.atlasSpace`, `tui.atlas.spaceGlyph`, `tui.statsFor`, `wire.ToStatusAllEnvelope` | Logical planning tree summary. Contains `Summary`, `Failure`, `Selected`, and `Stale`. Fresh service reads have either `Summary` or `Failure`. Reconciled reads may carry both with `Stale: true`. |
| `core.SpaceInProgress` | Combined Task Projection | `spaceOverviewFromSummaries`, `tui.atlas.setWork`, `tui.atlas.workRows`, `tui.atlas.openAtlasWork`, `wire.ToStatusAllEnvelope` | Pairs a running task with `SpaceID`, `PlanningID`, and `Stale` provenance from its source space. |
| `core.SpaceLoadFailure` | Domain Failure Value | `summarizeSpaceGroup`, `SpaceSummary.Contended`, `tui.atlas.spaceRows`, `tui.atlas.entryBand`, `cli.render.StatusAllHuman`, `cli.statusAllProblemsError`, `wire.ToStatusAllEnvelope` | Adapter-neutral failure evidence carrying `Class domain.Class` and `Message string`. Preserves sentinel classification without filesystem-type leakage. |
| `core.SpaceOverviewRefresh` | Typed Replacement Subset | `SpaceOverviewService.RetryContended`, `ApplySpaceOverviewRefresh`, `tui.atlasRetriedMsg` | Distinct type enclosing `Spaces []SpaceSummary`. Disallows accidental substitution for complete overview. |
| `core.SpaceOverviewService.Overview` | Primary Core Port Method | `cli.RunStatusAll`, `tui.loadAtlas` (via `enterAtlas`, `markAtlasStale`) | Scans registry and planning stores. Registry errors fail whole call; per-space read failures isolate to individual groups. |
| `core.SpaceOverviewService.RetryContended` | Primary Core Port Method | `tui.retryContendedAtlas` (via `handleAtlasRetry`) | Reads strictly groups where `space.Contended() == true`. Does not query registry or healthy spaces. |
| `core.RetainContendedSpaceSummaries` | Core Reconciliation Transformer | `tui.atlas.setOverview` (when `a.loaded == true`) | Merges prior coherent snapshot with fresh full read. Contended spaces retain prior summary marked `Stale: true`. Durable failures drop stale summary. |
| `core.ApplySpaceOverviewRefresh` | Core Reconciliation Transformer | `tui.Model.handleAtlasRetried` | Splices partial retry subset into current overview by stable identity key. Re-derives combined `InProgress` tasks. |
| `tui.atlasLoadedMsg` | Bubble Tea Message | `handleAtlasLoaded` | Delivers full overview from `loadAtlas`. Guarded by `loadGen`. Dispatches `deferAtlasRetry` if contention observed. |
| `tui.atlasRetryMsg` | Bubble Tea Message | `handleAtlasRetry` | Fired by `deferAtlasRetry` tick timer. Guarded by `loadGen` and `atlas.retrying`. Dispatches `retryContendedAtlas`. |
| `tui.atlasRetriedMsg` | Bubble Tea Message | `handleAtlasRetried` | Delivers partial retry result. Guarded by `loadGen` and `atlas.retrying`. Clears `atlas.retrying`, applies refresh, never reschedules. |
| `atlas.stale` | UI State Flag | `atlas.applyOverview`, `enterAtlas`, `markAtlasStale` | True when planning data changed underneath active space or overview has contention. Directs `enterAtlas` to reload. |
| `atlas.retrying` | UI State Flag | `handleAtlasLoaded`, `handleAtlasRetry`, `handleAtlasRetried`, `statusLines` | True during quiet-period retry. Prevents duplicate retry scheduling; drives status banner text. |
| CLI Render / Exit Policy | Primary Adapter Output | `cli.render.StatusAllHuman`, `cli.statusAllProblemsError` | Renders summary or failure in human output. `statusAllProblemsError` fails with `domain.ErrValidation` if selected tree has failure or files unreadable. |
| Wire Mapping | Primary Serialization | `wire.ToStatusAllEnvelope` | Encodes to `StatusAllEnvelope`. Maps `Failure.Message` to `Error`. Retains both `Summary` and `Error` if long-lived adapter provides reconciled input. |
| `core.SpaceOverviewStore` | Secondary Port Interface | `spacestore.FS.OpenPlanningStore` (`internal/spacestore/fs.go:81`) | Sole production implementation; opens `*store.FS` for a path, satisfying `PlanningSummarySource` (`SummaryStore` + `TaskGraphSource`). |

---

### State-transition table

The table compares intended contracts against observed production behavior across all lifecycle phases:

| Prior State | Incoming Event | Expected Summary | Expected Failure | Expected Stale | Combined Work Rows | Retry Permission / Flag | Code Alignment |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| None (`!a.loaded`) | Fresh Success | Fresh `Summary` | `nil` | `false` | Fresh tasks, `Stale: false` | `retrying = false`; no retry scheduled | Exact |
| None (`!a.loaded`) | Fresh Conflict | `nil` (no data invented) | `ClassConflict` | `false` | None for this space | `retrying = true`; 1 quiet-period retry | Exact |
| None (`!a.loaded`) | Fresh Durable Error | `nil` | `ClassUnknown` / `ClassNotFound` | `false` | None for this space | `retrying = false`; no retry | Exact |
| Coherent Data | Fresh Success | Advanced fresh `Summary` | `nil` | `false` | Fresh tasks, `Stale: false` | `retrying = false`; no retry | Exact |
| Coherent Data | Fresh Conflict | Retained prior `Summary` | `ClassConflict` | `true` | Retained tasks, `Stale: true` | `retrying = true`; 1 quiet-period retry | Exact |
| Coherent Data | Fresh Durable Error | `nil` (dropped immediately) | `ClassUnknown` / `ClassNotFound` | `false` | None for this space (dropped) | `retrying = false`; no retry | Exact |
| Contended (retrying) | Retry Success | Fresh `Summary` (replaces stale) | `nil` | `false` | Fresh tasks, `Stale: false` | `retrying = false`; settles | Exact |
| Contended (retrying) | Retry Conflict | Retained coherent `Summary` preserved | `ClassConflict` (updated message) | `true` | Retained tasks, `Stale: true` | `retrying = false`; bound exhausted; no loop | Exact |
| Contended (retrying) | Retry Durable Error | `nil` (dropped) | `ClassUnknown` (updated) | `false` | None for this space (dropped) | `retrying = false`; settles | Exact |
| Coherent Data | Group Removed in Registry | Space removed from overview | N/A | N/A | Removed from combined work | Not retried; never resurrected | Exact |
| Coherent Data | Groups Reordered in Registry | Order follows new registry list | Matched by stable key | Retained if contended | Matches new ordering | Governed by contention state | Exact |
| Contended (retrying) | Manual Refresh (`press "r"`) | `loadGen++`, supersedes pending retry tick and running result | Fresh read evaluated | Determined by fresh read | Derived from fresh read | `retrying = false`; resets state machine | Exact |
| Contended (retrying) | Watcher Event (`markAtlasStale`) | `loadGen++` if on screen; supersedes pending retry | Fresh read evaluated | Determined by fresh read | Derived from fresh read | `retrying = false`; resets state machine | Exact |
| Contended (retrying) | Workspace Switch (`activateWorkspace`) | Outer `sessionGen++`; drops running retry | Unaffected in old session | Intact in old session | Unaffected | Dropped via `sessionMsg` check | Exact |

---

### Findings

#### M1. Outer workspace-session guard test coverage misses retried result message · **Status:** fixed

- **Severity:** Medium
- **Area:** Concurrency & Test Coverage (`internal/tui`)
- **Location:** `internal/tui/model.go:281-289`, `internal/tui/atlas_test.go:299-316`
- **Description:**
  The review brief explicitly requires testing the mutation: *"remove the outer workspace-session guard for either retry message"*.
  In `internal/tui/atlas_test.go:310-315` (`TestAtlasDropsSupersededAndOldSessionRetryRequests`), only `atlasRetryMsg` is tested against an older session:
  ```go
  oldSession := m.sessionGen
  m.sessionGen++
  tm, cmd = m.Update(sessionMsg{gen: oldSession, msg: atlasRetryMsg{gen: m.atlas.loadGen}})
  if cmd != nil || tm.(Model).atlas.retrying != m.atlas.retrying {
      t.Fatal("an old workspace session retry must not touch the current Atlas")
  }
  ```
  However, the second asynchronous message—`atlasRetriedMsg` (which carries the executed `core.SpaceOverviewRefresh` result from `retryContendedAtlas`)—is not asserted against an outdated session.
- **Hostile Mutation Probe Evidence:**
  Mutating `internal/tui/model.go:281` to bypass the session guard specifically for `atlasRetriedMsg` (`case atlasRetriedMsg: msg = scoped.msg`) allows the message through despite `scoped.gen != m.sessionGen`. All tests in `internal/tui` pass without error.
- **Risk:**
  If a user triggers a workspace switch while a slow `retryContendedAtlas` background command is in flight, and if the session guard were inadvertently weakened for retry results, the retried data from the previous workspace session could land in the newly activated session.
- **Remedy:**
  Add a subtest in `TestAtlasDropsSupersededAndOldSessionRetryRequests` verifying that `sessionMsg{gen: oldSession, msg: atlasRetriedMsg{gen: m.atlas.loadGen, refresh: ...}}` is dropped without modifying `m.atlas.overview` or triggering state updates.

---

**Resolution:** Added old-session coverage for atlasRetriedMsg as well as
atlasRetryMsg. A foreign-session completed result cannot modify the overview or
current retry state.

#### M2. Consecutive contention after first-load failure lacks explicit regression coverage · **Status:** fixed

- **Severity:** Medium
- **Area:** Core Reconciliation & Test Coverage (`internal/core`, `internal/tui`)
- **Location:** `internal/core/space_overview.go:167`, `internal/core/space_overview_test.go:185-224`
- **Description:**
  The review brief requires testing the mutation: *"retain prior data on a durable failure or on first load"*.
  In `internal/core/space_overview.go:167`:
  ```go
  func retainContendedSpaceSummary(previous, current SpaceSummary) SpaceSummary {
      if !current.Contended() || previous.Summary == nil {
          return current
      }
      retained := *previous.Summary
      current.Summary = &retained
      current.Stale = true
      return current
  }
  ```
  The guard `|| previous.Summary == nil` protects against dereferencing `*previous.Summary` when the prior generation did not possess a valid summary (e.g., following a first-load failure or all-broken registration).
- **Hostile Mutation Probe Evidence:**
  Mutating `retainContendedSpaceSummary` to remove `|| previous.Summary == nil` (`if !current.Contended() { return current }`) survives the entire test suite in both `internal/core` and `internal/tui`.
  This occurs because existing tests in `space_overview_test.go` always initialize `previous` with an established healthy summary before injecting contention. Furthermore, in `internal/tui/atlas_test.go:271` (`TestAtlasFirstLoadAndDurableFailuresDoNotInventRetainedData`), after first-load contention, the test switches `adapter.summaryErrs[beta]` to `errors.New("durable tree failure")` before pressing `"r"`. Because the subsequent error is durable, `current.Contended()` evaluates to `false`, bypassing `*previous.Summary` and masking the missing nil check.
- **Risk:**
  If a registered space is busy during application startup (first load has no summary) and remains busy when the user later triggers a manual refresh (`r`) or watcher event, removing or regressing this check would cause a nil-pointer panic.
- **Remedy:**
  Add an explicit test case in `space_overview_test.go` asserting that `RetainContendedSpaceSummaries(unloadedPrevious, contendedCurrent)` returns `contendedCurrent` without panicking or creating a non-nil `Summary`.

---

**Resolution:** Added explicit consecutive first-load contention coverage
proving reconciliation neither dereferences nor invents a missing prior Summary.

#### L1. Ungrouped work view relies solely on color to indicate stale task rows · **Status:** fixed

- **Severity:** Low
- **Area:** Presentation & Accessibility (`internal/tui`)
- **Location:** `internal/tui/atlas.go:1185-1191`
- **Description:**
  In the Atlas work view, when sorting by started date or priority (`atlasWorkByStarted`, `atlasWorkByPriority`), task rows are ungrouped. Individual rows render the owning space label as:
  ```go
  spaceLabel := padRight(truncate("["+row.SpaceID+"]", spaceW), spaceW)
  if row.Stale {
      spaceLabel = st.fg(theme.ColorYellow, spaceLabel)
  } else {
      spaceLabel = st.dim(spaceLabel)
  }
  line += spaceLabel + " "
  ```
  In this ungrouped view, the task row itself contains no textual marker (such as `"stale"`) to signify that it originates from a retained, contended summary. The distinction relies entirely on yellow ANSI color vs dimmed styling.
- **Hostile Evidence Probe:**
  In a probe with `NO_COLOR=1` or when ANSI color sequences are stripped (`ansi.Strip`), the output for a stale row (`"  [beta]  task-1  1mo ago  "`) is textually indistinguishable from a fresh row (`"› [alpha] task-1  1mo ago  "`). While the pinned status bar at the bottom of the screen displays `"⚠ 1 space summary retained from the last coherent read"`, the specific rows belonging to that retained space cannot be identified without color.
  By contrast, the spaces view provides an explicit `"stale"` table column and `"◐"` glyph, and the grouped work view (`atlasWorkBySpace`) appends `" · stale"` to the space heading.
- **Remedy:**
  In ungrouped work rows, render an explicit text discriminator when `row.Stale` is true (e.g., formatting the badge as `[beta · stale]` or appending `*`), ensuring full accessibility on monochrome or high-contrast terminals.

---

**Resolution:** Ungrouped retained work now includes an explicit stale text
label in started and priority orderings, so the affected row remains
identifiable without color.

### Hostile evidence and restored mutation ledgers

#### Restored mutation probe ledger

| # | Mutation Description | Target File & Line | Killing Test / Result | Status |
| :--- | :--- | :--- | :--- | :--- |
| 1 | Classify every per-space error as contention (`ClassConflict`) | `internal/core/space_overview.go:226,232` | Killed by `TestSpaceOverviewRepeatedContentionIsBoundedAndDurableFailureReplacesStaleData`, `TestAtlasFirstLoadAndDurableFailuresDoNotInventRetainedData`, and `TestAtlasWorkViewOmitsSpacesWhoseSummaryFailed`. | Killed |
| 2a | Retain prior data on durable failure (remove `!current.Contended()`) | `internal/core/space_overview.go:167` | Killed by `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups` and `TestSpaceOverviewRepeatedContentionIsBoundedAndDurableFailureReplacesStaleData`. | Killed |
| 2b | Retain/invent data on first load when `previous.Summary == nil` (remove `|| previous.Summary == nil`) | `internal/core/space_overview.go:167` | **SURVIVED** all core and TUI tests. No test executes `RetainContendedSpaceSummaries` with an unloaded prior summary followed by consecutive contention. Logged as **M2**. | **Surviving** |
| 3 | Retry every registered group instead of only contended groups | `internal/core/space_overview.go:121` | Killed by `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups` (`retry opened [/alpha /beta], want only [/beta]`), `TestAtlasRetainsContendedSpaceAndRetriesOnlyThatSpaceAfterQuietPeriod`, and `TestAtlasRepeatedContentionStopsAfterOneRetryAndKeepsStaleEvidence`. | Killed |
| 4 | Apply retry subset as though it were a complete overview | `internal/core/space_overview.go:153` | Killed by `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups` (index out of range panic) and `TestAtlasRepeatedContentionStopsAfterOneRetryAndKeepsStaleEvidence`. | Killed |
| 5 | Allow repeated conflict to schedule another automatic retry | `internal/tui/atlas.go:220` | Killed by `TestAtlasRepeatedContentionStopsAfterOneRetryAndKeepsStaleEvidence` (`a retry's own contention must become visible instead of scheduling again`). | Killed |
| 6 | Remove Atlas load-generation guard (`msg.gen != m.atlas.loadGen`) | `internal/tui/atlas.go:209` | Killed by `TestAtlasDropsSupersededAndOldSessionRetryRequests` (`a newer Atlas generation must drop the older retry`). | Killed |
| 7a | Bypass outer workspace session guard for `atlasRetryMsg` | `internal/tui/model.go:281` | Killed by `TestAtlasDropsSupersededAndOldSessionRetryRequests` (`an old workspace session retry must not touch the current Atlas`). | Killed |
| 7b | Bypass outer workspace session guard for `atlasRetriedMsg` | `internal/tui/model.go:281` | **SURVIVED** all tests. Existing tests only assert `atlasRetryMsg` under old session. Logged as **M1**. | **Surviving** |
| 8 | Rebuild combined work without stale provenance (`Stale: false`) | `internal/core/space_overview.go:187` | Killed by `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups` (`combined work did not retain stale provenance`). | Killed |
| 9a | Discard retained summary in wire mapper when failure exists | `internal/wire/envelopes.go:350` | Killed by `TestStatusAllEnvelope_PreservesRetainedSummaryAndFailure` (`Summary:<nil>`). | Killed |
| 9b | Discard failure message in wire mapper when summary exists | `internal/wire/envelopes.go:347` | Killed by `TestStatusAllEnvelope_PreservesRetainedSummaryAndFailure` (`Error:""`). | Killed |

---

### Separation of defects, source-supported risks, and unverified concerns

#### Demonstrated defects
- **None.** All production execution paths function according to specification; no panics, data corruption, or architectural invariant violations were demonstrated under normal or adversarial inputs.

#### Source-supported risks
- **M1:** Missing session-scoping regression coverage for `atlasRetriedMsg` leaves open the risk that a future refactoring of `Model.Update` could allow background retry responses to land in a foreign workspace session after a switch.
- **M2:** Missing regression coverage for `previous.Summary == nil` under consecutive contention leaves open the risk that an optimization in `retainContendedSpaceSummary` could trigger a nil pointer dereference on persistent first-load contention.
- **L1:** Visual reliance on ANSI color for ungrouped work rows creates an accessibility deficit in monochrome or no-color terminal configurations.

#### Unverified concerns (Investigated and Disproved)
- *Concern: Shallow copy of `core.Summary` could allow concurrent modification of retained snapshot.*
  **Settled:** `core.Summary` is read-only after creation; neither core services nor TUI reducers modify the slices of an existing `Summary`.
- *Concern: Duplicate or missing planning IDs could cause cross-space collision during reconciliation.*
  **Settled:** `SpaceRegistryService.groupSpaceEntries` aggregates entries by `id:<planningID>`, `root:<root>`, or `entry:<index>:<id>`. Duplicate IDs within a group map to the same logical space; distinct groups have unique keys.
- *Concern: `RetryContended` could delete or reorder healthy spaces when merging.*
  **Settled:** `ApplySpaceOverviewRefresh` indexes `current.Spaces` by stable key and performs in-place positional substitution, preserving existing slice length and relative order.

---

### Acceptance-criteria traceability and documentation truthfulness

| Acceptance Criterion | Implementation Evidence | Verification Status |
| :--- | :--- | :--- |
| Real planner-window test with ≥2 spaces proves contended space retains coherent summary while healthy space advances | `internal/tui/atlas_test.go:161` (`TestAtlasRetainsContendedSpaceAndRetriesOnlyThatSpaceAfterQuietPeriod`), confirmed with real store lock probe `TestAuditFloor_RealContentionSpansTick`. | **Verified** |
| Per-space failures remain structured through core without string parsing or filesystem policy in TUI | `core.SpaceLoadFailure` uses `domain.Class`. TUI switches on `space.Contended()` (`Class == domain.ClassConflict`). Proved via pathless in-memory store probe `TestAuditFloor_PortabilityInMemory`. | **Verified** |
| Transient work retried with documented bound; superseded generations/sessions cannot land stale results | `tui.deferAtlasRetry` fires once via `tea.Tick`. `handleAtlasRetry` and `handleAtlasRetried` enforce `loadGen` and `sessionGen`. Verified via `TestAuditFloor_SupersessionObsoleteDelayedAndRunningRetry`. | **Verified** |
| Repeated contention and non-conflict failures visible without retry loop; later refresh can recover | `TestAtlasRepeatedContentionStopsAfterOneRetryAndKeepsStaleEvidence` and `TestAuditFloor_RealContentionSpansTick` prove loop stops at 1 retry; subsequent manual refresh recovers. | **Verified** |
| Single-space dashboard/entity/Thread retry behavior remains unchanged | Single-space read retry policies in `internal/tui/model.go` remain untouched; full test suite passes. | **Verified** |

#### Documentation truthfulness:
- `docs/ARCHITECTURE.md` (lines 145-148, 515-521, 576-578) accurately captures `SpaceOverviewService` partial-refresh capabilities, load-generation guards, and stale provenance.
- `planning/adrs/0006-adopt-threads-as-task-dags.md` (lines 1195-1200) accurately reflects rollout step 5 and cross-space Atlas behavior.
- `README.md` (lines 250-253) accurately describes user-facing retry behavior during concurrent mutations.
- `planning/tasks/6g63db3sdfrh-...` acceptance criteria checkboxes match implementation realities.

---

### Settled hostile angles

1. **State-Machine Soundness:** Proved that `Stale: true` can never exist without a valid coherent summary in core, and `retrying: true` can never persist after an exhausted retry.
2. **Identity and Reconciliation:** Reordered and removed groups tested against `RetainContendedSpaceSummaries`; removed groups are never resurrected, and reordered groups preserve fresh registry sequence.
3. **Concurrency and Message Timelines:** All Bubble Tea message intersections (manual refresh, watcher debounce, tick expiration, workspace opening, and teardown) tested. Superseded load generations and old workspace sessions cleanly drop obsolete commands and messages.
4. **Retry Bounds:** Hard-capped in production state via `m.atlas.retrying = false` upon entry to `handleAtlasRetried`; zero rescheduled commands returned on repeated contention.
5. **Portability:** Pathless in-memory adapter (`mem://`, `uri://`) verifies zero dependency on local paths, OS file descriptors, or error string formatting.
6. **Wire and CLI Compatibility:** Full JSON output evaluated against schema; presence of both `summary` and `error` validates cleanly without schema violations.
