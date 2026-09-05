---
schema: 1
id: 6g6wn6791emf
bucket: closed
area: atlas-partial-refresh-recovery-implementation-claude
date: "2026-09-04"
updated_at: "2026-09-04"
---
# Audit: Atlas partial-refresh recovery — claude — 2026-09-04

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

**Verdict: not ready.**

The core reconciliation is well designed and, in isolation, correct: the state table over every
prior/current combination is sound, classification is genuinely adapter-neutral (a pathless
`remote://` adapter behaves identically), `RetryContended` reopens exactly the contended roots, a
partial refresh cannot reorder, drop, or resurrect groups, and real held planner windows behave
correctly whether they span the delayed retry or close before it. Nine of the twelve required
mutations are killed by named tests.

It is `not ready` for one reason: **`applyOverview` is reused for the partial-retry path while still
resetting three pieces of state that only a complete re-read may reset** (H1). Two symptoms are
demonstrated on production code with ordinary user actions, the more serious being that a successful
bounded retry landing while the user is off the Atlas erases the watcher's "planning data changed"
signal, so the next Atlas entry skips its re-read entirely and shows pre-change data with no stale
marker — the exact false freshness this feature exists to prevent. Four further findings (M1, M2,
L1, L2) follow.

### Isolation attestation

| item | value |
| --- | --- |
| Sandbox path | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.Exrekh` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.Exrekh/.git` |
| Sandbox baseline commit | `83946ac chore: capture review sandbox baseline` (on `fd9dfda`) |
| Captured source-audit blob | `33a6dc18b20b69280eb7863b70229d1c9855cd41` |
| Guarded transfer | **succeeded** — source blob re-verified as `33a6dc18b20b69280eb7863b70229d1c9855cd41` immediately before the write, then `cp -p` to a `mktemp` sibling and atomic `mv`; `cmp -s` confirmed the source audit is byte-identical to the sandbox copy. This report's own transfer-outcome line was added in a second guarded pass against blob `a946497d3c60f417f3ba35da58cf9b61af5d7781`. Only this audit file was copied back; no source, generated file, Git metadata, test artifact, or other planning file was. |

Isolation was verified, not assumed, before any implementation file was read:

- `git rev-parse --absolute-git-dir` resolves **inside** the sandbox and **not** under
  `/Users/andyeschbacher/git/andy-esch/taskflow`; both directions were asserted against
  `pwd -P`-normalised paths (my first check tripped on a doubled slash from `$TMPDIR` and was
  corrected rather than waved through).
- `.git` is a real directory, not a gitfile or symlink, so this is not a `git worktree`.
- No `objects/info/alternates` file exists (the clone used `--no-hardlinks`), so no object is shared
  with the source.
- `git worktree list` reports only the sandbox; `core.worktree` is unset.
- The working tree was overlaid with `rsync -a --delete --exclude=.git` from the source, so the
  captured state includes the handoff's uncommitted and untracked files; `git add -A` plus the
  checkpoint commit made that the restoration baseline. `git diff --stat 3c74392..HEAD` restricted to
  `internal/`, `cmd/`, `docs/`, `README.md` is empty relative to the handoff, confirming the review
  ran against the handoff snapshot exactly.
- Every build, test, generator, fixture, mutation, and edit in this review ran in the sandbox. The
  source checkout was used only to read the brief, hash the audit, and rsync the initial copy — no
  commit, branch switch, stage, restore, clean, stash, reset, or write-capable project command.
- The checkpoint commit is the only commit created.

### Reviewed state, runtime, and validation results

| item | value |
| --- | --- |
| Branch | `feat/atlas-partial-refresh-recovery` |
| Base | `fd9dfdad4244` (merge of #188) |
| Implementation | the uncommitted working-tree diff captured in the sandbox baseline |
| Delta reviewed | 14 files, +642/−70 (excluding the two review audits) |
| Go | `go1.26.6 darwin/arm64` |

All commands below ran in the sandbox.

| check | command | result |
| --- | --- | --- |
| full suite, **uncached**, race | `go clean -testcache && go test -race ./...` | **PASS** — exit 0, 24 `ok` packages, 0 `FAIL`; 15.8 s wall / 28.0 s user |
| focused repeat, race | `go test -race -count=10 -run '<8 new/affected tests>' ./internal/core/ ./internal/tui/ ./internal/wire/ ./internal/cli/` | **PASS** — 8.1 s wall, no flakes across 10 runs |
| static analysis | `golangci-lint run ./...` | **0 issues** — 5.7 s wall |
| vet | `go vet ./...` | **clean** (exit 0) |
| generated CLI docs | `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | **no drift** |
| module tidiness | `go mod tidy -diff` | **clean** |
| whitespace | `git diff --check` | **clean** (exit 0) |
| planning lint | `tskflwctl -C planning lint` | `✔ all planning entities and dependency links pass lint`, exit 0 |
| audit lint | `tskflwctl -C planning audit lint 6g6wn6791emf` | `✔ all audit findings pass lint`, exit 0 |

Cached/uncached: the full race run was explicitly uncached (`go clean -testcache` immediately
before; exit code and package tally captured to a file). Mutation-ledger rows used the ordinary
cache — each mutation invalidates the packages it touches, and every row's verdict was read from
that run's own `--- FAIL` lines. The focused repeat used `-count=10`, which bypasses the cache.

Constraints recorded rather than substituted: no dependency was installed, so JSON envelopes were
checked with Go's `encoding/json` under `DisallowUnknownFields` plus explicit key enumeration rather
than an external JSON-Schema validator; the repository's own `TestJSONSchema_ValidatesRealOutput`
supplies real schema validation in-suite and was run as part of the full suite.

### Data flow and state machine, re-derived from production code

```
Overview()  ── per group ──▶ summarizeSpaceGroup ──▶ SpaceSummary{Summary XOR Failure}
                                                       Failure.Class = domain.Classify(err)
                                                       (ClassConflict ⇒ Contended())
        │
        ▼  atlas.setOverview (only when a.loaded)
RetainContendedSpaceSummaries(previous, current)      ← caller-side reconciliation
        │   key = planning:<PlanningID> | entry:<Entries[0].ID> | summary:<ID>
        │   retain iff current.Contended() && previous.Summary != nil
        ▼
applyOverview(overview)  ─┬─▶ a.overview / a.spaces / setWork / sortSpaces / restoreCursor
                          ├─▶ a.loaded = true
                          ├─▶ a.stale   = overviewHasContention(overview)   ◀── H1
                          ├─▶ a.loadErr = nil                               ◀── H1
                          └─▶ a.openErr = ""                                ◀── H1
        │
        ▼  contention ⇒ retrying = true; deferAtlasRetry(gen) after fsDebounce (200ms)
atlasRetryMsg{gen}  → guard: gen == loadGen && retrying && svc != nil  → RetryContended(overview)
atlasRetriedMsg{gen}→ guard: gen == loadGen && retrying                → ApplySpaceOverviewRefresh
                                                                          → applyOverview  ◀── H1
outer guard: Model.Update drops any sessionMsg whose gen != m.sessionGen (model.go:281)
```

`RetryContended` re-reads only groups where `Contended()` holds, rebuilding each `SpaceGroup` from
the *captured* `PlanningID` and a copied `Entries` slice — it never re-reads the registry.
`ApplySpaceOverviewRefresh` positions by the same key and calls `retainContendedSpaceSummary` with
(current-state, retry-result), so a repeated conflict keeps the same coherent summary while success
or a durable failure replaces it.

### State-transition table (expected vs. observed)

Verified exhaustively by probe over all 4 prior × 4 current states, with three invariants asserted on
every result: `Stale ⇒ Summary != nil`, `Stale ⇒ Contended()`, `Summary != nil && Failure != nil ⇒ Stale`,
plus work-row/space staleness agreement. **All 16 combinations matched expectation and no invariant
was violated.**

| prior | current | Summary | Failure | Stale | work rows | matches expectation |
| --- | --- | --- | --- | --- | --- | --- |
| none (first load) | success | new | — | false | 1 | ✔ |
| none (first load) | conflict | **nil** | conflict | false | 0 | ✔ invents nothing |
| none (first load) | durable | nil | unknown | false | 0 | ✔ |
| none (first load) | not-found | nil | not-found | false | 0 | ✔ not contention |
| coherent | success | new | — | false | 1 | ✔ advances |
| coherent | conflict | **old** | conflict | **true** | 1 (stale) | ✔ retains |
| coherent | durable | nil | unknown | false | 0 | ✔ replaces |
| coherent | not-found | nil | not-found | false | 0 | ✔ replaces |
| stale (already retained) | success | new | — | false | 1 | ✔ |
| stale | conflict | old | conflict | true | 1 (stale) | ✔ chains |
| stale | durable | nil | unknown | false | 0 | ✔ drops |
| stale | not-found | nil | not-found | false | 0 | ✔ drops |
| failed (no summary) | conflict | **nil** | conflict | **false** | 0 | ✔ nothing to retain |
| failed | success/durable/not-found | as fresh | — | false | — | ✔ |

Retry subset (`ApplySpaceOverviewRefresh`), with an untouched healthy space alongside:

| retry outcome | contended space | healthy space | group order/count |
| --- | --- | --- | --- |
| success | new summary, `Stale=false`, `Failure=nil` | untouched | preserved |
| repeated conflict | same coherent summary, `Stale=true` | untouched | preserved |
| durable failure | summary dropped, `Failure` set, `Stale=false` | untouched | preserved |
| names a removed group | — | untouched | **not appended, not resurrected** |

Registry-shape cases: reordering the fresh overview matches by identity, not position (verified). A
legacy group with no planning id keys on `Entries[0].ID`; reordering its entries **loses retention
and fails safe** — it never misassociates one group's summary with another. Newer full load and
newer workspace session are covered under supersession below.

### Consumer and composition inventory

**P** = production consumer · **T** = test double · **G** = generated artifact.

| symbol | production consumers | classification |
| --- | --- | --- |
| `SpaceOverview` | `wire.ToStatusAllEnvelope`; `render.StatusAllHuman`; `render.StatusAllJSON`; `cli.statusAllProblemsError`; `atlas.overview` field; `atlas.setOverview`/`applyOverview`; `overviewHasContention`; `retryContendedAtlas`; `atlasLoadedMsg` | **P** 9 |
| `SpaceSummary` | `atlasSpace.summary`; `statsFor`; `atlasActivity`; `atlasSpaceName`; `atlasSpaceKey`; `isCurrentSpace`; `spaceGlyph`; `spaceRows`; `entryBand`; render + wire mappers | **P** 10 |
| `SpaceSummary.Failure` (**new**) | `render.StatusAllHuman` (else-branch only); `cli.statusAllProblemsError`; `wire.ToStatusAllEnvelope`; `atlas.spaceRows`; `atlas.entryBand`; `Contended()` | **P** 6 |
| `SpaceSummary.Stale` (**new**) | `atlas.spaceRows` (column + width); `atlas.spaceGlyph`; `atlas.entryBand`; `atlas.staleSpaceCount`; `spaceOverviewFromSummaries` (work provenance) | **P** 5 — **no wire field, no CLI human/exit consumer** |
| `SpaceLoadFailure` (**new**) | constructed at 3 sites in `summarizeSpaceGroup`; read by the 6 `Failure` consumers | **P** 3 writers / 6 readers |
| `SpaceInProgress.Stale` (**new**) | `atlas.workRows` (grouped heading + ungrouped space cell) | **P** 1 (2 code paths) |
| `SpaceOverviewRefresh` (**new**) | `RetryContended` (producer); `atlasRetriedMsg.refresh`; `ApplySpaceOverviewRefresh` | **P** 3 |
| `Overview` | `loadAtlas` (TUI); `cli/status.go` `--all` | **P** 2 |
| `RetryContended` (**new**) | `retryContendedAtlas` only | **P** 1 |
| `RetainContendedSpaceSummaries` (**new**) | `atlas.setOverview` only | **P** 1 |
| `ApplySpaceOverviewRefresh` (**new**) | `handleAtlasRetried` only | **P** 1 |
| `atlasRetryMsg` / `atlasRetriedMsg` (**new**) | `model.update` dispatch; `handleAtlasRetry`/`handleAtlasRetried`; both ride the outer `sessionMsg` wrapper | **P** 2 each |
| `atlas.stale` | **written by** `markAtlasStale` ("data changed") **and** `applyOverview` ("contended"); **read by** `enterAtlas` | **P** 2 writers / 1 reader — the collision in **H1** |
| `atlas.retrying` (**new**) | set in `handleAtlasLoaded`; cleared in `handleAtlasLoaded`/`handleAtlasRetried`/`enterAtlas`/`markAtlasStale`; read by both retry guards and `statusLines` | **P** 4 writers / 3 readers |
| `SpaceOverviewStore` implementations | `spacestore.FS` (**P**, sole production impl) · `core.fakeSpaceOverviewStore` (**T**) · `tui.atlasTestAdapter` (**T**, filesystem-backed `*store.FS`) · my `pathlessOverviewStore` (**T**, pathless) | **P** 1 |
| CLI render/exit policy | `StatusAllHuman` renders `Summary` **or** `Failure` (else-branch); `statusAllProblemsError` counts a space as failed only when `Summary == nil` | **P** 2 — both treat the pair as mutually exclusive (**L1**) |
| wire mapping | `ToStatusAllEnvelope` emits `summary` **and** `error` independently; no `stale` field; `SchemaVersion` unchanged at 1.60 | **P** 1 — schema correctly not widened |

### Acceptance-criteria traceability

| criterion | verdict | evidence |
| --- | --- | --- |
| Real planner-window test, ≥2 spaces, contended space retains while the other stays current | **met** | `TestAtlasRetainsContendedSpaceAndRetriesOnlyThatSpaceAfterQuietPeriod` uses a real `MutateTaskGraph` window. My independent probes reproduced it with a window spanning the retry, a window closing before it, and both spaces held at once |
| Per-space failures structured through core, rendered without string parsing or filesystem policy in the TUI | **met** | `SpaceLoadFailure{Class,Message}`; TUI switches on `Contended()`, never on message text. A pathless `remote://` adapter with wrapped `ErrConflict` classified and retried identically; no path leaked into any message |
| Transient work retried with a documented bound; superseded generations/sessions cannot land stale results | **met in production; half-tested** | Bound holds (probe: 5 full loads → 5 retries, never more; a still-contended retry issues no follow-up). Both guards work. But the *result*-side generation guard has no test — **M2** |
| Repeated contention and non-conflict failures visible without a retry loop; later manual/watcher refresh recovers | **met** | Repeated conflict → `press r to retry`, no scheduled work; durable → summary dropped. I ran the emitted remedy: pressing `r` after the window closed did recover the space |
| Single-space dashboard/entity/Thread retry behaviour unchanged | **met** | `read_retry.go`, `dashboard.go`, `entity.go`, `detail.go` are untouched; `model.go` gains only two message cases |

### Planning and documentation truthfulness

Accurate: the ADR-0006 amendment, the `ARCHITECTURE.md` core paragraph, and the closeout's
description of the mechanism all match the code. The closeout's validation list (full race suite,
lint, generated docs, planning/audit lint, diff check) is true — I re-ran every item.

Two claims are stronger than what shipped:

- The closeout says validation *covers* "superseded Atlas and workspace generations". It covers the
  superseded **request**; the superseded **result** has no test, and removing its guard leaves the
  suite green while producing false freshness (**M2**).
- README and `ARCHITECTURE.md` both say the retained summary is "visibly stale" / "visibly marked
  stale". That holds in the spaces table (a `◐` glyph at every width plus a `stale` column ≥60
  columns) and in the grouped work view (`beta · stale`), but **not** in the ungrouped work view,
  where staleness is carried by colour alone (**M1**).

### Hostile-evidence ledger

**A. Real guarded planner windows (not a fake error map).** `store.FS.MutateTaskGraph` held open on
a real tree, so concurrent reads fail with the genuine wrapped `domain.ErrConflict`.

| scenario | result |
| --- | --- |
| window spans the delayed retry | beta retained + `Stale`, alpha advanced, retry contended again, **no follow-up command**, `press r to retry` shown |
| the emitted remedy actually run | after releasing the window, `r` fully recovered the space (`Stale=false`, `Failure=nil`) |
| window closes before the retry | full recovery; view carries no stale marker; opens = `[alpha, beta, beta]` (one full pass + one targeted reopen) |
| both spaces held simultaneously | both retained; retry reopened exactly both; releasing only alpha recovered alpha while beta kept its evidence |

**B. Supersession, both halves × three vectors.** Newer manual refresh (`r`), watcher-triggered
refresh, and workspace-session replacement, each against an obsolete **delayed request** and an
already-started **retry result**: in all six the obsolete message was dropped, no command ran, and
the visible overview, cursor, retry flag, and banner were unchanged. I additionally hand-built the
gen-N-result-lands-after-gen-N+1 interleaving; with the guard present, nothing moved.

**C. Retry bound in production state.** One manual refresh followed by four watcher refreshes under
sustained contention: 5 full loads, 5 automatic retries, never more than one per full load; no retry
result ever issued a follow-up command; the `retrying` flag never survived its own retry. The
documented bound ("a later manual/watcher refresh may start a new bounded attempt") is what the code
does. Worth noting for capacity, not correctness: each watcher event under contention now costs one
full registry pass plus one reopen per contended group.

**D. Portability.** A `pathlessOverviewStore` with `remote://a` / `remote://b` roots and no
filesystem concepts: wrapped `ErrConflict` from `OpenPlanningStore` **and** from inside `summarize`
both classified as contention; retention, targeted retry (`opened == [remote://b]`), and recovery all
worked; an unknown error classified durable and dropped retained data; "no healthy entry point" is
`ClassNotFound` and correctly does **not** schedule a retry. No message leaked a local path. The only
filesystem policy left is in `spacestore.FS`, the primary adapter, which is correct.

**E. Presentation.** Spaces table at 40/60/80/120 columns: no line exceeded the width; the `◐` glyph
is present at every width; the `stale` column appears from 60 up; the footer states the retained
count and switches between "retrying after the quiet period" and "press r to retry". Grouped work
view labels the heading `beta · stale`. Ungrouped work view: colour only (**M1**). Durable failure
renders the durable message in red with no stale wording and no retry hint. Cursor and entry
selection survived both the contended full load and the partial retry.

**F. CLI/wire compatibility matrix.** Four inputs through `StatusAllHuman`, `StatusAllJSON` (decoded
with `DisallowUnknownFields`), and `statusAllProblemsError`:

| input | human | JSON | exit error |
| --- | --- | --- | --- |
| ordinary success | summary | `summary`, no `error` | no |
| ordinary failure | failure message | `error`, no `summary` | yes |
| contended, nothing retained | failure message | `error`, no `summary` | yes |
| **reconciled: retained summary + conflict** | **summary only — failure never shown** | **`summary` *and* `error`** | **no** |

`schema_version` stays `1.60` and the space object gains no key — the schema was correctly not
widened. The disagreement in the last row is **L1**.

### Restored-mutation ledger

Every row applied to the sandbox checkpoint, run against the **full** `go test ./...`, then restored
with `git checkout -- .`; the sandbox was verified clean (`git status --short` empty) after each.

| # | mutation | outcome | killed by |
| --- | --- | --- | --- |
| M1 | classify every per-space failure as contention | **killed** | `TestSpaceOverviewRepeatedContentionIsBoundedAndDurableFailureReplacesStaleData`, `TestAtlasFirstLoadAndDurableFailuresDoNotInventRetainedData`, `TestAtlasWorkViewOmitsSpacesWhoseSummaryFailed` |
| M2a | retain prior data on a durable failure too | **killed** | `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups`, `…RepeatedContentionIsBounded…`, `TestAtlasRetainsContendedSpaceAndRetriesOnlyThatSpaceAfterQuietPeriod`, +2 |
| M2b | reconcile on first load too (drop the `if a.loaded` guard) | **survived — equivalent mutant** | none, and correctly so: `!a.loaded ⇒ a.overview` is the zero value, so `RetainContendedSpaceSummaries(zero, current)` provably returns `current` unchanged. The guard is defensive, not load-bearing; first-load behaviour is genuinely pinned by `TestAtlasFirstLoadAndDurableFailuresDoNotInventRetainedData` |
| M3 | retry every registered group, not only contended ones | **killed** | `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups`, `TestAtlasRetainsContendedSpace…`, `TestAtlasRepeatedContentionStopsAfterOneRetry…` |
| M4 | apply the retry subset as a complete overview | **killed** | `TestAtlasRepeatedContentionStopsAfterOneRetryAndKeepsStaleEvidence` |
| M5 | let a repeated conflict schedule another automatic retry | **killed** | `TestAtlasRepeatedContentionStopsAfterOneRetryAndKeepsStaleEvidence` |
| M6a | remove the load-generation guard on the retry **request** | **killed** | `TestAtlasDropsSupersededAndOldSessionRetryRequests` |
| M6b | remove the load-generation guard on the retry **result** | **SURVIVED** | nothing → **M2**. Harm demonstrated separately: an obsolete gen-N result flipped beta from `stale/contended` to `fresh/healthy`, cleared `retrying`, and cancelled gen N+1's own scheduled retry |
| M7 | remove the outer workspace-session guard | **killed** | `TestAtlasDropsSupersededAndOldSessionRetryRequests`, `TestAtlasDropsStaleWorkspaceResultsAndOldSessionMessages` |
| M8 | rebuild combined work rows without stale provenance | **killed** | `TestSpaceOverviewRetainsStructuredContentionAndRetriesOnlyFailedGroups` |
| M9a | wire mapper discards a retained summary when a failure exists | **killed** | `TestStatusAllEnvelope_PreservesRetainedSummaryAndFailure` |
| M9b | wire mapper discards the failure message when a summary exists | **killed** | `TestStatusAllEnvelope_PreservesRetainedSummaryAndFailure` |
| M10 | *(coordinated, bonus)* drift the TUI's duplicated `atlasSpaceKey` away from core's `spaceSummaryKey` | **SURVIVED** | nothing → **L2** |
| M11 | *(bonus)* share the previous `*Summary` pointer instead of copying it | **SURVIVED** | nothing → **L2** |

Nine of twelve required mutations are killed by named tests. M2b is an equivalent mutant (proved, not
assumed). M6b, M10, and M11 are genuine coverage gaps.

### Demonstrated defects, source-supported risks, and unverified concerns

**Demonstrated defects** (reproduced against production code in the sandbox): H1, M1, M2, L1, L2 below.

**Source-supported risks** (real in the source, not reachable through a production path today):

1. *The reconciled shape has no wire staleness marker.* `SpaceSummaryJSON` gained no `stale` field —
   correctly, since the schema must not widen — but the branch simultaneously declared the
   summary-plus-error shape supported and pinned it with a test. A machine consumer handed that shape
   sees `summary` and `error` and would reasonably read the summary as current. Only `status --all`
   maps this type today, and it never produces the shape. Recorded under **L1**.
2. *`atlas.stale` is now a two-owner flag.* Beyond the demonstrated H1 symptom, any future reader of
   `stale` inherits the ambiguity between "planning data changed" and "a space is contended".
3. *The retry runs while the Atlas is off-screen.* `handleAtlasRetry` does not check `m.onAtlas`, so a
   bounded retry reads contended trees after the user has left. Harmless in itself (read-only,
   bounded, and it leaves the Atlas fresher) — but it is the mechanism that makes H1 reachable.
4. *Watcher amplification.* Under sustained contention each fsnotify quiet period now costs one full
   registry pass **plus** one reopen per contended group. Bounded per event and consistent with the
   documented contract; noted for capacity, not correctness.

**Unverified concerns** (stated, not resolved):

1. *Real terminal rendering.* All presentation evidence came from `atlas.view`/`workView` with ANSI
   stripped, which faithfully models a no-color profile but is not a live TTY. Interaction between the
   new status line and scrolling/focus at small heights is unverified beyond the rendered string.
2. *Multi-process contention.* The planner guard I exercised (`repositoryGuardFor`) is process-local.
   Contention from a genuinely separate `tskflwctl` process goes through the platform file lock, which
   I did not drive; classification there depends on that path also wrapping `domain.ErrConflict`.
3. *Wall-clock behaviour of `tea.Tick`.* Retry timing was driven by invoking the returned command
   directly, so the 200 ms `fsDebounce` was never actually awaited; ordering, not latency, is what I
   verified.

### Settled concerns (challenged, then closed with evidence)

- **Can a partial refresh erase, reorder, or resurrect unrelated groups?** No. `ApplySpaceOverviewRefresh`
  positions by key and only overwrites matched indices; a refresh naming a removed group is a no-op
  (probed). Group order and count were preserved in every retry outcome.
- **Can identity matching misassociate two groups?** No. Reordering the fresh overview matches
  correctly by identity. A legacy group whose entries are reordered simply fails to match and loses
  retention — it never adopts another group's summary. Catalog grouping keys distinct roots into
  distinct groups, so the `entry:` fallback cannot collide in a well-formed registry.
- **Impossible states?** None reachable. All 16 prior×current combinations satisfied
  `Stale ⇒ Summary != nil`, `Stale ⇒ Contended()`, and `Summary+Failure ⇒ Stale`, and every work row's
  staleness matched its space.
- **Does a first-load conflict invent data?** No — `Summary` stays nil, `Stale` false, and the retry is
  still scheduled. Proved both by the existing test and by the equivalent-mutant analysis of M2b.
- **Is the retry bound real, or just absent from one test?** Real, proved in production state across
  five refresh cycles under sustained contention.
- **Is classification message-based or filesystem-shaped?** Neither. It is `domain.Classify` over
  wrapped sentinels; a pathless remote adapter classifies identically and no local path appears in any
  message.
- **Nil-service paths.** `(*SpaceOverviewService)(nil).RetryContended` panics — but so does the
  pre-existing `.Overview()`, on the same nil receiver. Not a new class, and no production path
  constructs a nil service (`enterAtlas` and `handleAtlasRetry` both check `m.spaceOverviewSvc == nil`).
- **Does the change disturb single-space retry machinery?** No: `read_retry.go`, `dashboard.go`,
  `entity.go`, and `detail.go` are untouched, and `model.go` gains only two dispatch cases.
- **Was the schema silently widened?** No. `schema_version` stays 1.60, the space object gains no key,
  and envelopes decode under `DisallowUnknownFields`.

## Findings

### High

#### H1. The partial-retry path reuses applyOverview, which resets state only a complete re-read may reset · **Status:** fixed

`applyOverview` (`internal/tui/atlas.go:442`) ends by writing three fields:

```go
a.loaded = true
a.stale = overviewHasContention(overview)
a.loadErr = nil
a.openErr = ""
```

Before this change that function ran **only** from `setOverview`, i.e. only after a complete
`Overview()` that re-read every registered tree, where resetting all three is sound. This change adds
a second caller — `handleAtlasRetried` (`atlas.go:220`) — which applies a **partial** refresh that
re-read only the contended groups. Resetting whole-projection state from a subset read is unsound,
and two symptoms are demonstrated on production code.

**Symptom 1 — a successful retry erases the watcher's "data changed" signal, and the next Atlas entry
shows pre-change data.** `atlas.stale` has two writers with two different meanings: `markAtlasStale`
(`atlas.go:343`) means *planning data changed underneath, re-read on next entry*, and `applyOverview`
now means *some space is currently contended*. `enterAtlas` (`atlas.go:322`) reads it and skips the
reload when it is false. Reproduced end to end:

1. Refresh; beta is contended (real held planner window) → `retrying = true`, tick scheduled, `stale = true`.
2. The user leaves the Atlas.
3. A task file is really created in alpha; the watcher calls `markAtlasStale`, which sets `stale = true`
   and returns **nil** because the Atlas is off-screen — relying on `enterAtlas` to re-read later.
4. Beta's window closes; the delayed retry runs (it is not gated on `m.onAtlas`) and succeeds.
   `applyOverview` overwrites `stale = false`. **Alpha was never re-read.**
5. The user re-enters: `enterAtlas(false)` sees `loaded && !stale` → **no reload at all**
   (`summaryOpens` is empty). Alpha shows 1 in-progress row; the task added in step 3 is invisible,
   and nothing is marked stale.

Probe output:

```
after exit:                onAtlas=false stale=true retrying=true
after watcher event:       stale=true (watcher signal) reload-issued=false
after successful retry:    stale=false retrying=false
on re-entry:               reload issued=false (opens=[])
alpha in-progress rows shown: 1   (2 = includes the task added while away)
```

The control case is safe and confirms the scope: with the user **on** the Atlas, `markAtlasStale`
bumps `loadGen` and reloads, so the older retry is superseded and the new task appears (2 rows). The
defect needs the user to be off the Atlas when the watcher fires.

This is the false freshness the feature exists to prevent, and it defeats the invariant
`markAtlasStale`'s own comment states ("The next entry re-reads instead… otherwise the atlas keeps
showing the working set as it was at program start"). It self-corrects on the next watcher event
received while on-screen, or on `r` — but the user has no signal that anything is wrong.

**Symptom 2 — the automatic retry silently discards an unacknowledged open error.** `a.openErr` is set
when entering a space fails (`atlas.go:230`, and synchronously at `atlas.go:373-384` for an
unavailable opener, a space with no entry point, or an unhealthy entry). Those synchronous cases land
well inside the 200 ms `fsDebounce` window. Reproduced: with a retry pending, an open error on screen
is gone after the retry lands, with no user action:

```
before retry, open error visible: true
after  retry, open error visible: false   (atlas.openErr="")
```

The user is told why the space would not open, and ~200 ms later the explanation vanishes.

(`a.loadErr` is not affected: every load error path sets `retrying = false`, so a retry cannot be in
flight alongside one. That asymmetry is itself evidence that the three fields are not interchangeable.)

**Suggested direction** (owner's call): keep whole-projection state out of the shared applier — have
`handleAtlasRetried` preserve what the subset cannot know, e.g. capture `wasStale := m.atlas.stale`
and `openErr` before applying and restore `stale = wasStale || overviewHasContention(...)`; or split
`applyOverview` into a rows/cursor projector plus an explicit full-refresh state reset that only
`setOverview` calls.

**Resolution:** Separated whole-overview dirty state from per-space contention.
Only a complete setOverview clears dirty/load/open state; partial retry
projection preserves watcher dirtiness and open errors. Added a regression
proving re-entry still performs the deferred full refresh.

### Medium

#### M1. The ungrouped work view distinguishes stale rows by ANSI colour alone · **Status:** fixed

`workRows` (`internal/tui/atlas.go:1182-1192`) renders the per-row space cell as:

```go
spaceLabel := padRight(truncate("["+row.SpaceID+"]", spaceW), spaceW)
if row.Stale {
    spaceLabel = st.fg(theme.ColorYellow, spaceLabel)
} else {
    spaceLabel = st.dim(spaceLabel)
}
```

The two branches produce identical **text**; only the colour differs. With ANSI stripped — a no-color
profile, a monochrome terminal, or a reader who cannot separate yellow from dim — a stale row and a
fresh row are byte-identical. Probed at 120 columns with one fresh and one retained-stale space, in
both ungrouped orderings (`started`, `priority`):

```
› [alpha] fresh-task  01-e
  [beta]  stale-task  01-e
```

(The task slugs above are fixture names; the rows carry no staleness text of their own.) The grouped
ordering does it correctly — `beta · stale` in the heading — which shows both the intent and how small
the fix is. The spaces table is also fine: a `◐` glyph at every width plus a `stale` column from 60
columns up.

This contradicts an explicit contract of the change ("remain usable at narrow widths/no color") and
the README/`ARCHITECTURE.md` claim that retained data is "visibly stale". The footer does say
`⚠ 1 space summary retained from the last coherent read`, so the user learns *something* is stale —
but in the ungrouped work view, which is the default `started` ordering, they cannot tell **which**
rows, and those rows are the ones they navigate into and act on.

**Resolution:** Ungrouped retained work now carries an explicit stale text
prefix before the task label, independent of ANSI color and preserved even if
the space column is dropped. Added started/priority order coverage.

#### M2. The load-generation guard on the retry result is untested, and its removal produces false freshness · **Status:** fixed

`handleAtlasRetried` (`atlas.go:216`) guards on `msg.gen != m.atlas.loadGen`. Deleting that clause
leaves the **entire** `go test ./...` green. `TestAtlasDropsSupersededAndOldSessionRetryRequests`
covers only the request half (`atlasRetryMsg`), despite the plural in its name; the result half
(`atlasRetriedMsg`) has no generation-supersession coverage. The brief's own evidence floor asks for
"both halves of the asynchronous operation", and the task closeout claims validation "covers …
superseded Atlas and workspace generations".

The guard is load-bearing. I built the interleaving by hand — a gen-N retry result landing after a
gen-N+1 full load has already replaced the overview and scheduled its own retry — and ran it against
production code and against the mutant:

```
guard present:  after the obsolete gen-N result landed: stale=true  contended=true  retrying=true
                gen N+1's own retry still runs: true
guard removed:  after the obsolete gen-N result landed: stale=false contended=false retrying=false
                gen N+1's own retry still runs: false
```

Without the guard, an obsolete result (a) presents a space as freshly healthy when the current read
says it is contended, (b) clears the newer generation's retry flag, and (c) cancels the newer
generation's scheduled retry, so the Atlas stops retrying altogether. Production is correct; nothing
holds it there. Extending the existing supersession test with an `atlasRetriedMsg` case whose `gen` is
stale would pin it.

**Resolution:** Extended supersession coverage to the completed atlasRetriedMsg
path. Obsolete load-generation results cannot replace current data or clear the
current retry.

### Low

#### L1. Human status --all and its exit policy discard the failure evidence the wire mapper was changed to preserve · **Status:** fixed

The change makes `ToStatusAllEnvelope` emit `summary` **and** `error` independently and adds
`TestStatusAllEnvelope_PreservesRetainedSummaryAndFailure` to pin it; the type comment now advertises
the shape ("Long-lived adapters may deliberately retain a stale Summary beside Error; the wire mapping
preserves both rather than inventing its own refresh policy"). The other two consumers of that same
type were not adapted, and both treat `Summary` and `Failure` as mutually exclusive:

- `render.StatusAllHuman` (`internal/cli/render/status.go:140-147`) renders the failure only in the
  `else` branch of `if space.Summary != nil`.
- `cli.statusAllProblemsError` (`internal/cli/status.go:91-99`) `continue`s on a non-nil summary before
  it ever reaches the failure check.

Fed the reconciled shape, the three surfaces tell three different stories:

| surface | reconciled summary + conflict + stale |
| --- | --- |
| human | summary only — the failure and staleness are never mentioned |
| JSON | `summary` **and** `error: "planner window is active"` |
| exit policy | counted as success |

Additionally the wire carries no staleness marker (correctly, since the schema must not widen), so a
machine consumer given this shape reads a stale summary as current.

None of this is reachable today: `status --all` calls `Overview()`, which never yields the pair. The
finding is that the branch declared and tested a shape that two of its three consumers silently
mishandle — a trap for the long-lived adapter the comment invites.

**Resolution:** Made human status render retained summary and failure
independently, and removed the exit aggregator's mutually exclusive branch so a
paired failure remains non-zero. Added human and exit-policy regression tests.

#### L2. The new identity rule is a verbatim duplicate and the "last coherent snapshot" copy is shallow; neither is pinned · **Status:** fixed

Two surviving bonus mutations, one theme: the safety properties the new core helpers claim are not
held by anything.

*Duplicated identity rule.* `core.spaceSummaryKey` (`space_overview.go:194`) is byte-for-byte identical
to the pre-existing `tui.atlasSpaceKey` (`atlas.go:652`). Retention/reconciliation now keys on one
copy; cursor restoration keys on the other. Adding an `ID`-first branch to the TUI copy — so the two
disagree — leaves the whole suite green. Nothing detects the drift, and when it happens retention and
cursor restoration would silently disagree about which space is which (blast radius: a cursor that
stops restoring when a group's selected entry changes).

*Shallow snapshot copy.* `retainContendedSpaceSummary` (`space_overview.go:170-171`) does
`retained := *previous.Summary; current.Summary = &retained`, which reads as isolating the last
coherent snapshot. It is a struct copy only: `Counts`, `InProgress`, `Epics`, `OpenAudits`, and
`Problems` remain shared with the previous generation. Probed:

```
retained summary is a distinct *Summary value (struct copy)
ALIASED: mutating the previous generation's task slice changed the retained snapshot
```

Replacing the copy with a plain pointer share also leaves the suite green, confirming the copy is
defensive-only and currently unobservable. That is fine today because nothing mutates a `core.Summary`
after construction — but the defence is advertised by the code shape and does not actually hold, and
the retention chain means one snapshot can persist across many generations.

**Resolution:** Centralized reconciliation/cursor identity in
core.SpaceSummary.ReconciliationKey and replaced the shallow retained Summary
copy with an owned copy of every reachable mutable slice. Added identity
precedence and alias-mutation tests.

## Candidate tasks

No deferral is warranted. H1 is a small change at one call site in code this branch already touches;
M1 is the same one-line treatment the grouped work view already applies; M2 is an added case in an
existing test; L1 and L2 are contained. Findings are left open for triage with
`tskflwctl audit finding`.

---

_Transfer: the guarded one-file copy back to the shared checkout succeeded; the sandbox is left in place at `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.Exrekh` until the implementation owner confirms receipt._
