---
schema: 1
id: 6gg38rnpe94c
bucket: open
area: final-semantic-entity-source-boundary-checkpoint-codex
date: "2026-10-03"
---
# Audit: Final semantic-entity source-boundary checkpoint — codex — 2026-10-03

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Adversarially review the final semantic-entity source-boundary checkpoint for task
`6gcwcf88z57p`, `split-local-path-capabilities-from-semantic-entity-reads`, in draft PR #274.
Task, Thread, Epic, Audit, and Research now contain no `Path`, `FilenameID`, or
`SourceVersion`; Thread's `CanonicalID()` fallback is removed too. Decide whether
this task can be completed after triage, not merely whether the removed fields compile.

Do not implement fixes. Preserve this brief, update only your assigned audit,
and leave findings open. Distinguish regressions, pre-existing problems exposed by
this migration, and optional improvements. Earlier checkpoint findings have been
triaged; reopen them only with new failing evidence.

## Review target

Branch: `refactor/split-local-entity-source-capabilities`.
Implementation freeze: `506b7990c222a577c751ad87685a90a437d1fb04`.
Primary delta: `2d4d9f1..506b799` (Task path removal, Thread source-boundary
migration, and Thread callback isolation). The sandbox baseline additionally captures
the final planning update and these briefs; it must contain no later implementation.

Relevant implementation commits are `a31bd42`, `7afaecf`, and `506b799`.
The earlier local-capability and Task-identity checkpoint audits are
`planning/audits/6gfrtba5jky2-*`, `6gfrtbae3bp4-*`, `6gfyn1n6wyn6-*`, and
`6gfyn1nhawqq-*`. Read their owner dispositions as scope context, not as proof
that this final delta is correct.

Build a producer/consumer inventory for source ID, declared ID, diagnostic location,
path presentation hint, executable local handle, and source revision. Trace normal
and bulk-apply Thread scans, selected reads, semantic planner snapshots, source
validators, materialization, snapshot CAS, repair/lifecycle impact receipts, ordinary
CLI/wire projections, and TUI selection/local actions. Check production call sites,
not just the names or assertions of the new tests.

## Intended contract to challenge

1. Domain IDs are declarations. `RecordSource.ID` is the adapter's resolution
   identity; it remains independent when a declaration is missing or wrong.
   Ordinary reads/lint retain malformed records and attribute defects honestly.
2. `Location` is diagnostic context. `LocationIsPath` affects presentation only.
   Ordinary local actions request the optional path port; guarded local writers
   require `VersionedRecord.LocalPath` and the original `SourceVersion`.
   Neither path-shaped text nor a path hint creates write authority.
3. `ValidateThreadCreationSource`/`ValidateThreadMutationSource` validate the
   complete source-aware read before semantic planner values are constructed.
   Missing/duplicate source IDs, declaration drift, invalid documents, cross-kind
   collisions, missing members, and unreadable Threads block ordinary guarded work.
   Pure apply-snapshot validation does not substitute for this adapter source gate.
4. Thread source materialization and complete-snapshot CAS fail closed on changed
   handles, renamed sources, missing revisions, and changed readable/unreadable
   bytes. Receipts obtain local paths from operation metadata, never domain values.
5. Broken-graph repair can proceed past malformed Thread evidence. Readable Thread
   impact projections must retain their independent source identity/location,
   including declaration drift, and must not become healthy merely because the
   underlying task graph was repaired. Repair does not edit Thread documents.
6. Callback-owned Thread slices, nested tags/tasks, and apply-body maps cannot
   rewrite the owner's authorization snapshot. A cancelled Thread cannot be
   revived by a callback pretending its starting status was unstarted.
7. Valid-file CLI/wire behavior and persisted Markdown contracts remain compatible.
   Pathless adapters do not manufacture filenames, source IDs, or revision tokens.
   Bare Task/Thread compatibility projections do not impersonate a source-aware read.

## Mandatory evidence floor

- Inventory the actual producers/consumers with exact path/line evidence. Verify
  each removed-field test migration retained the hostile state it originally tested;
  equal-ID or pathless happy paths do not prove drift/duplicate behavior.
- Exercise matching IDs, missing declarations, different valid declarations,
  duplicate source IDs at distinct paths, and distinct sources sharing a declaration.
  Check Thread list/show, lint, source-ID path recovery, and at least one guarded
  operation. Include an unrelated malformed Thread and a pathless loaded record.
- Demonstrate the source gate at creation, membership/lifecycle, task lifecycle,
  compose, initial bulk apply, and final apply convergence. Prove failed preflight
  invokes no planner/write where applicable; do not assume one validator call
  implies every caller is safe. Check no-op and dry-run paths separately.
- Verify repair receipts with a readable drifting Thread in both dry-run and
  committed repair: task graph improves, Thread remains broken, canonical impact
  ID and defect location survive, and Thread bytes remain unchanged.
- Test a path-shaped location with the path hint enabled but no local handle;
  then test an opaque location with a separately supplied valid local handle.
  Include missing/stale revisions and readable/unreadable representation changes.
- Perform at least three restored, compiling mutation probes. Mandatory targets:
  (a) guarded local-path authority or handle CAS; (b) source identity retained in
  repair impacts; (c) callback snapshot isolation. Name the exact changed guard,
  focused regression, and observed behavioral failure. A compile error is not a kill.
  A green mutant must be investigated, not counted as success because another
  layer was assumed to protect it.
- Compare representative valid human/JSON outputs with the base, inspect generated
  schema comments, and verify revisions/local handles never leak through public
  projections. Report any actual compatibility change rather than silently
  approving an updated fixture.

## Required hostile angles

- Could a semantic-only constructor or the new pure snapshot validator be reached
  in production where source identity is required? Find the route and test it.
- Does `ThreadRead.LoadedThreads()` preserve every readable occurrence and own
  mutable slices while excluding guarded metadata? Can repair impacts collapse
  duplicate sources, lose drift, or use the declared ID as an actionable fallback?
- Does materialization compare against the original read revision or accidentally
  mint a fresh token after a race? Can an idempotent result report a fabricated
  snapshot because a callback or helper changed the validation inputs?
- Could a renamed/deleted record, changed body, or contradictory source/location
  pass one path while failing another? Separate guarded-write safety from the
  documented best-effort external-editor pathname race.
- Challenge shared test helpers: `graphFixturePath`, `semanticThreadRead`, and
  Thread fakes can only represent certain source states. Introduce independent
  evidence rather than relying on a helper that couples declaration and source ID.
- Second pass: assume a future database adapter with an opaque key, no filename,
  and no local path. Find remaining assumptions that would force it to invent
  local evidence or weaken its complete-snapshot guard.

Codex emphasis: trace architecture and compatibility routes end to end, including
all consumers of `SemanticThreads()`, `LoadedThreads()`, bare `ProjectThread`, and
repair/lifecycle Thread impacts. Test any fallback or inconsistent authority gate.

Antigravity emphasis: disprove these three concrete claims before broadening scope:
(1) a path hint cannot replace a local write handle; (2) graph repair cannot hide
readable Thread drift; (3) no Thread planner can mutate its own authorization inputs.
Use `internal/store/thread_source_boundary_test.go`,
`internal/core/thread_source_boundary_test.go`, and
`internal/store/thread_planner_snapshot_test.go` as entry points, not as proof.
Remove the relevant protection in your sandbox and require a behavioral test failure;
for callback isolation, challenge creation and apply nested values/body maps too,
not only the mutation scalar status. Then inspect call sites those tests do not reach.

## Validation and restoration

Use only the mandatory independent sandbox. Build a fresh sandbox binary for CLI
probes; do not use an installed binary that predates this branch. Run focused tests,
`go test ./...`, `go test -race ./internal/core ./internal/store ./internal/tui ./internal/wire`,
`just lint`, fresh-binary planning lint, schema-comment regeneration, and
`git diff --check` where feasible. Report checks not run and why.

Restore every probe to the sandbox baseline before the next probe and before transfer.
Transfer only your assigned audit through the helper. Do not push, stage, or create
additional commits, and never run mutations in the shared source checkout.

## Deliverable

Add a severity-ranked verdict, producer/consumer inventory, hostile fixture and
restored-mutation tables, commands/results, compatibility observations, and sandbox
attestation. Add actionable findings under `## Findings` with exact repository
grammar and open statuses. Include confidence limits and any untested critical seam.
Explicitly answer: can the implementing task now be completed, and what must change
first? A partial review must say partial rather than declare readiness.

## Reviewer report

Pending external review. No findings have been entered by the implementation owner.
