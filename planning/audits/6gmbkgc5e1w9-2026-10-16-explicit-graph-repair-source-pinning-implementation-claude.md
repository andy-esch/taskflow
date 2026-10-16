---
schema: 1
id: 6gmbkgc5e1w9
bucket: closed
area: explicit-graph-repair-source-pinning-implementation-claude
date: "2026-10-16"
---
# Audit: Explicit graph repair source pinning implementation — claude — 2026-10-16

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Your handoff prompt prints exact `SOURCE_ROOT` and `AUDIT_REL`
values. Substitute them below; never derive them from your current directory, which may be a
different checkout. Then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="<SOURCE_ROOT from your handoff prompt>"
AUDIT_REL="<AUDIT_REL from your handoff prompt>"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)" && cd "${SANDBOX:?create failed: stop and report the blocker}"
```

Every later block begins from `${SANDBOX:?…}`, so it cannot run in the shared checkout after a
failed `create`. Run the helper from `$SOURCE_ROOT`, not from the sandbox: you may mutate the
sandbox's copy while probing, and only the read-only source copy can be trusted to verify your work.

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. The sandbox's push URLs are disabled; never re-enable them or
push to the source path. If creation fails, report the blocker; never fall back to the shared
checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run
the helper's fail-closed verification:

```sh
cd "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX" &&
  git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent or nested `.git`, a
removed or weakened push guard, source-deliverable drift, and empty reports; it copies back only
the assigned audit through a same-directory atomic rename. A successful transfer records a
helper-owned receipt and prints it (`transfer=succeeded` with the transferred blob). A refusal exits
2 and records nothing. Exit 3 means the report was already delivered when a later step failed: run
`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"` and put its
output, or its refusal, in your handoff. Put the printed receipt in your handoff message. Do not
copy anything else manually. Leave the workspace in place and report its path until the
implementation owner confirms delivery with that same `receipt` command. On refusal, preserve it and
report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

## Review brief

Challenge whether explicit graph-repair retries preserve the selected source
without freezing stale authorization or accidentally binding auto-cleanup.
This is a wrong-target write fix, not a new graph engine or repair policy.

Primary assignments, read from this audit's reviewer assignment:

- **Claude:** portable Service retry state, binding lifetime, empty-result conflicts,
  current defect/occurrence authorization and durable retry stops. Cross-check one
  real-store duplicate-ID case.
- **Antigravity:** actual filesystem source targeting, raw rename/deletion and source
  preservation, plus executable CLI intent. Cross-check the portable empty-result
  regression. Do not substitute a broad green test run for these probes.

Review independently; do not read the sibling report until your own report is final.

## Review target

Read-only source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Base: `552a9b863a88` on `fix/pin-explicit-graph-repair-sources`.
The implementation is **uncommitted**, including two untracked regression files:
capture staged, unstaged, deleted and untracked state, not HEAD alone.

Task: [Pin explicit graph-repair selections](../tasks/6gm9ny98zyf2-pin-explicit-graph-repair-source-selections-across-retries.md).
Thread: [Finish selected identity safety](../threads/6gm4h71hm1yt-finish-selected-identity-safety.md).

Verify a consumer inventory through:

- `internal/cli/task_dependency_repair.go`: selector/manifest translation and Service
  call; help and its generated page now describe within-call pinning.
- `internal/core/dependency_repair.go`: `Service.RepairTaskGraph`,
  `resolveRepairRetryEdits`, `PlanTaskGraphRepair`, `normalizeRepairEditSource`,
  `authorizeRequestedRepair`, validation and semantic receipts.
- `internal/core/store.go`: `TaskGraphRepairStore` / planner contract.
- `internal/store/graphrepair.go`: guard, authoritative snapshots, materialization,
  snapshot/per-step CAS, prefix receipts. Persistence production code is unchanged.
- `internal/core/dependency_repair_retry_test.go` and
  `internal/store/graphrepair_target_test.go`: new portable and real-store tests;
  `graphrepair_test.go` now checks conflict-after-prefix with a nonzero retry budget.
- Owning guides: `docs/architecture/planning-data.md` and `persistence.md`;
  compatibility classification: ADR-0008. No wire shape/revision change is proposed.

The new concurrency audit was separately triaged. Lint ID relocation M1/M2 are
tracked by `6gmbfxy1d561`; watcher L1 by `6gmbfxy96tpw`. They are not implemented
here. Challenge this separation only with a concrete in-scope consequence. The
earlier source-binding audit is evidence, not an independent approval of this fix.

## Intended contract to challenge

1. Initial explicit ID/slug/path selection occurs under the repair guard. The first
   successful plan binds **source ID plus exact LocalPath**, even if subsequent
   validation returns an empty conflict result. Pre-planner contention binds nothing.
2. A retry needs exactly one readable occurrence at that ID/path. Missing, moved,
   unreadable or changed-ID selections fail closed; duplicate IDs at other paths
   do not redirect or invalidate a uniquely selected occurrence.
3. This is logical address identity, **not inode continuity**. Byte replacement at
   identical ID/path is current content, not independently detectable identity
   replacement. The user explicitly approved refusing rename following rather than
   introducing a cross-rename occurrence capability. Do not claim a stronger guarantee.
4. Every retry diagnoses and authorizes fresh defects/values/occurrences, strict
   improvement, preservation and impacts. Already-removed declarations may converge;
   missing selected sources may not. New non-defective dependencies cannot be erased.
5. Opaque `Location` never becomes path authority. Only explicitly supplied location
   context remains a constraint; refreshed diagnostic locations are not frozen.
6. Auto selections rediagnose fresh sources, including in mixed calls; they do not
   enter the explicit binding set. Preview, non-conflict errors, exhausted retries,
   committed outcomes and applied prefixes are not automatically replayed.

## Mandatory evidence floor

Verify every named symbol/test before relying on it. Report actual commands and
observations; a nonexistent test or compiler failure is not evidence.

- Both: run focused race tests for the changed core/store paths. Demonstrate at
  least one valid hostile repair and one refused repair through production code,
  checking semantic receipts and exact untouched bytes, not merely graph health.
- Claude mutation: remove the successful-callback `selectedSources = sources`
  assignment without breaking compilation. Require
  `TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent` to fail;
  restore and prove green. Independently check the real slug-swap regression.
- Antigravity mutation: remove the LocalPath conjunct from the retained-source
  matcher without breaking compilation. Require a specifically named case of
  `TestRepairRetriesRefuseMissingReplacedAndAmbiguousOccurrences` or the real
  duplicate-ID test to fail for the intended reason; restore and prove green.
- Antigravity production probe: in an unregistered throwaway planning tree, invoke
  an explicit CLI drop or YAML repair plan and show that only the selected defective
  declaration changes. Separately use real-store hooks to exercise a pre-verify
  rename/deletion, not a fabricated receipt. Trace selector -> Service -> store.
- Both: validate the error's inspect/rerun advice on a moved-source fixture by
  inspecting current diagnostics and making a new selection. Same-call refusal
  must not imply that the task can never be repaired after a legitimate rename.

These are bounded evidence floors, not finding quotas. A no-findings verdict needs
the assigned mutation and hostile probes. Label the portable committed-without-
applied-sources fixture as defensive port coverage unless you find a reachable
filesystem producer; do not invent one.

## Required hostile angles

Settle these hypotheses as reproduced, falsified by a named probe, or unresolved:

- An empty conflict result or intervening pre-planner conflict erases a previous
  explicit selection, allowing a reused slug/path to target another record.
- Duplicate IDs cause representative-task lookup to substitute an occurrence, or
  overly broad refusal prevents safe exact-path repair while damage remains residual.
- Retrying a drop/dedupe/empty-key selection removes newly valid or newly added
  intent, caches old occurrences, loses unrelated body/frontmatter changes, or claims
  the old prospective graph as durable state.
- Auto-only and mixed calls retain stale auto targets or expand explicit intent;
  missing sources are incorrectly treated as already repaired.

Do a distinct systemic pass after the checklist: inspect helper assumptions that
fixtures might hide (path-derived identity vs declarations, mutable label/context,
ID reuse at the same address, source-complete vs representative snapshots). Try
to falsify at least one shared assumption through a real caller. Distinguish the
approved same-address limitation from a new bypass; proposed stronger identity
schemes belong in the existing integrity design work, not invented guarantees.

## Validation and restoration

Owner evidence before handoff: `just build`, full `just test` (race), `just lint`
(0 issues), focused core/store repair race tests with `-count=3`, planning lint
with links. Generated docs were regenerated; `just docs-check` reports the one
intentional uncommitted repair-help diff. A fresh temporary docgen output is
compared to `docs/cli` for reproducibility. Independently run focused evidence;
you may inherit the full-suite baseline if time is better spent on hostile probes.

Use scratch fixtures and caches outside the sandbox tree, and build the sandbox's
own CLI. No registry pollution or real planning fixtures. Use an absolute scratch
GOCACHE if needed. Follow the mandatory isolated-workspace protocol injected above:
restore all probe changes to its captured baseline, keep only this audit modified,
verify with the **source** helper, then transfer exactly this audit. Only the helper's
sandbox baseline commit is permitted; no other commits or pushes. A permission,
isolation or evidence failure means incomplete review, not an invented clean result.

## Deliverable

Keep this brief intact and replace the report slot below. Lead with verdict/limits,
verified inventory and findings consolidated by root cause. For each defect include
producer trace, reachable consequence, reproduction, expected/actual evidence and a
bounded remedy. Distinguish product failures, coverage gaps and design preferences.
Use canonical finding creation/status verbs; leave findings open for owner triage.

Include a concise mutation ledger: exact edit, intended test/case, compiler-valid
killed/survived/invalid result, restoration and restored green run. Include a bounded
surviving-hypothesis ledger with uncertainty, not a speculative finding dump.

Include the helper verification attestation in the final report. Return its
confirmable transfer receipt in the handoff message, since the report cannot attest
its own later delivery. Retain your sandbox for owner receipt confirmation.

## Reviewer report

Reviewer: Claude (Opus 5.5), 2026-10-16. Independent; the sibling Antigravity report
was not read.

### Verdict and limits

**Ready for owner triage. No product findings** within the approved contract: a
logical source ID plus exact LocalPath address, not inode continuity. The required
mutation was killed by the intended test and by the real slug-swap regression. A
valid hostile repair and a refused hostile repair both behaved correctly through
production code, with exact-byte and receipt checks. The rerun advice works on a
moved-source fixture through Service and CLI. No finding is opened. Observations
that are design preferences, pre-existing behavior or defensive-only coverage are
in the surviving-hypothesis ledger below.

Limits:

- Lens: Claude assignment only (Service retry state, binding lifetime, empty-result
  conflicts, reauthorization, durable stops) plus one real-store duplicate-ID
  cross-check.
- Antigravity's raw rename/deletion and source-preservation probe set was not
  repeated, except where it overlapped (the moved-source refusal and duplicate
  cross-check).
- The guarantee verified is same-call logical address only. Byte replacement at the
  identical path is accepted as current content; probe P5 shows this.

### Verified inventory

All names below were checked in the sandbox. Line numbers refer to the sandbox
baseline `381204f0`.

- `internal/core/dependency_repair.go`:
  - `Service.RepairTaskGraph` (:627) copies `request.Edits` (:632).
  - It binds `selectedSources = sources` only inside a successful planner callback
    when the set is still nil (:642, :645).
  - It stops on preview, a non-conflict error, `result.Committed`, a non-empty
    `AppliedSources` or an exhausted budget (:650).
  - `resolveRepairRetryEdits` (:664) uses `normalizeRepairEditSource` only before
    binding (:671). After binding, it matches `TaskID && LocalPath` (:685).
  - Refusals are `ErrNotFound` for a missing occurrence (:690) and `ErrAmbiguous`
    for a non-unique one (:693). Only the caller-supplied `Location` is checked
    (:695). The edit is then re-sourced from the fresh record, so the label and
    context are not frozen.
  - `PlanTaskGraphRepair` (:332), `ValidateTaskGraphRepairPlan` (:454),
    `authorizeRequestedRepair` (:810), `repairEditTargetPresent` (:831),
    `normalizeRepairEditSource` (:874) and `taskGraphRepairReceipt` (:704) recompute
    diagnosis, authorization, strict improvement, preservation and receipts on every
    attempt.
- `internal/core/store.go`: `TaskGraphRepairPlanner` (:140). The
  `TaskGraphRepairStore` comment documents callback-retained selections (:145).
- `internal/store/graphrepair.go`: this production file is unchanged.
  `git diff --stat HEAD~1 HEAD -- internal/store/graphrepair.go internal/wire internal/cli/render`
  is empty.
  - The guard is taken at :33, and the planner runs under it on the authoritative
    graph at :60. The verify hook is at :89.
  - Snapshot CAS is at :99 (conflict at :101). Per-step CAS is at :115 and :124.
  - Materialization is at :188. `AppliedSources` and `Committed` are set together at
    :146 and :148.
- FS identity: `taskSource` (`internal/store/entity_read.go:14-16`) derives
  `Source.ID` from the basename, and `parseTask` takes the slug from the filename.
  On the filesystem, ID and slug are therefore path-derived. The duplicate-ID
  representative is the lowest Location (`internal/core/dependency_graph.go:455-467`,
  :507).
- `internal/cli/task_dependency_repair.go`:
  - Help `Long` (:52) and the Service call (:88).
  - Selector parsing chooses path when the selector has a separator, ID when it
    satisfies `id.Valid`, and slug otherwise (:139-146).
  - Manifest `location` is a constraint only when paired with a different `path`
    (:218-222).
  - The generated page `docs/cli/tskflwctl_task_depend_repair.md` matches a fresh
    docgen run.
- Tests: all exist and pass.
  - Portable, in `dependency_repair_retry_test.go`:
    - `TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent` (:66)
    - `TestRepairRetriesRefuseMissingReplacedAndAmbiguousOccurrences` (:116)
    - `TestRepairRetryReauthorizesDefectsAndCallerLocation` (:152)
    - `TestRepairRetriesRediagnoseAutoWithoutExpandingExplicitBindings` (:183)
    - `TestRepairRetryBoundariesAndDurableFailureReceipt` (:219)
    - `TestRepairOwnsCallerEditSlice` (:242)
    - `TestRepairRetryReauthorizesDedupeAndEmptyFieldIntent` (:254)
  - Real-store, in `graphrepair_target_test.go`:
    - `TestRepairTaskGraphRetryNeverReselectsReusedSlug` (:16)
    - `TestRepairTaskGraphRetryRefusesMovedMissingAndUnreadableSources` (:61)
    - `TestRepairTaskGraphRetryKeepsDuplicateIDOccurrenceAndConcurrentContent` (:121)
    - `TestRepairTaskGraphRetryReauthorizesCurrentDeclaration` (:174)
    - `TestRepairTaskGraphAutoRetryCanFollowNewSourceSet` (:217)
  - Changed: `TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix`
    (`graphrepair_test.go:262`) now uses an `ErrConflict`-wrapped stop and
    `WithRetry(3, fatal)`.
- Docs: the `planning-data.md` and `persistence.md` link targets exist. ADR-0008
  exists. There is no wire or revision change: the `internal/wire` and render diffs
  are empty.
- Defensive-only, with no reachable filesystem producer:
  - The portable `committed without applied sources` and `empty` result fixtures
    (graphrepair.go sets both flags together and keeps `result.Plan` after a
    successful plan).
  - `path reused with other ID` and `indistinguishable occurrences`: a path-derived
    ID means the same path implies the same ID, and two files cannot share a path.

### Findings

None. The probes and mutations below reproduced no product failure, coverage gap
that requires work, or untruthful contract claim in the assigned scope.

### Hostile angles settled

| Hypothesis | Result | Evidence |
| --- | --- | --- |
| An empty conflict result or a pre-planner conflict erases the selection, so a reused slug or path retargets | Falsified | Binding is assigned only once (:642-645). Mutation 1 (below) was killed by the intended test at `:103` and by the real slug swap at `graphrepair_target_test.go:47`. The intended test's steps mix pre-planner, empty-result and pre-planner conflicts. |
| A duplicate ID makes representative lookup substitute an occurrence | Falsified | Probe P1: the selected file is not the representative (a shadow sorts first). The retry wrote only the selected path, and the shadow stayed byte-identical. |
| An over-broad refusal blocks a safe exact-path repair | Falsified | P1 and the real duplicate-ID test both commit. Mutation 2 shows that ID-only matching would make this refusal `ErrAmbiguous`, and it is killed. |
| A retry removes newly valid or newly added intent, caches old occurrences, or loses body/frontmatter | Falsified | P1 exact bytes (concurrent `depends_on` value, comment field and body kept). Existing tests cover a now-valid dependency (`ErrValidation`), dedupe of fresh occurrences and an empty field gaining intent. |
| A retry claims the old prospective graph as durable state | Falsified | Refusals (P2 and the store refusal tests) return `Committed=false`, `Changed=false` and no applied files. On a pre-verify conflict, the store resets the analysis to the unchanged graph. |
| Auto-only or mixed calls retain stale auto targets or expand explicit intent | Falsified | Portable auto/mixed test. Real-store P4: the explicit operation stays pinned to its original path, while auto follows a renamed source and a new source, with nothing written to the old auto path. |
| A missing source is treated as already repaired | Falsified | P2 and the deleted/renamed/unreadable store tests return `ErrNotFound`. Only declaration absence at a retained source converges (`already removed declaration`). |

### Systemic pass

- **Path-derived identity vs declarations.** Probe P5 tried to falsify "identity" through
  a real caller. It replaces the selected file at the same path with different
  content whose frontmatter `id:` claims another existing task.
  - Result: the retry proceeds at that address. It commits only that file, leaves a
    residual `task-id-drift`, and leaves the other task's file byte-identical.
  - This matches the documented limit: replacing bytes at the same ID/path is
    current content, and a declaration is not identity. It is not a new bypass.
- **Mutable label and context.** The intended portable test asserts that the retry
  plan carries the fresh slug and fresh opaque Location (:107). Mutation 3 shows
  that only the caller-supplied Location acts as a constraint.
- **Source-complete vs representative snapshots.** `SourceRecords` fails closed on an
  incomplete graph, and P1 shows the shadow kept in residual evidence.
- **Shared test seams.** No test that uses a graph-repair hook calls `t.Parallel()`.
  The fake store counts only successful planner calls, so the `len(st.plans)`
  assertions are meaningful.
- **Other callers.** The only caller of `Service.RepairTaskGraph` is the CLI, and the
  only implementation of `TaskGraphRepairStore` is `*FS`. No caller bypasses the
  binding.

### Mutation ledger

Every probe compiled (`go build ./...`), was restored with
`git checkout -- internal/core/dependency_repair.go`, and the focused core/store
race tests were green after restoration. Numbers are review labels, not finding
codes.

| # | Edit (core/dependency_repair.go) | Intended test | Result |
| --- | --- | --- | --- |
| 1 | :645 `selectedSources = sources` -> `_ = sources` | `TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent` | Killed at `:103` ("explicit selection drifted"). Also killed: refusal subtests `renamed/ID`, `path_reused_with_other_ID/path` and `duplicate_ID_elsewhere…/ID`. Real store: `TestRepairTaskGraphRetryNeverReselectsReusedSlug` (`:47`, "retry repaired the reused slug's different source", `AppliedFiles=[…/h46gx48dzz67-alpha.md]`, `err=nil`), `…RefusesMoved…/ID/renamed` and `…KeepsDuplicateID…/duplicate-initially=false`. |
| 2 | :685 drop the `&& record.Source.LocalPath == selected[i].LocalPath` conjunct (duplicate-ID cross-check) | Real `TestRepairTaskGraphRetryKeepsDuplicateIDOccurrenceAndConcurrentContent` | Killed: both subtests at `:153` (`ErrAmbiguous`, empty receipt), probe P1, portable `renamed/{ID,slug,path}` and `duplicate_ID_elsewhere…/{ID,slug,path}`, and real slug-swap plus `…/{ID,slug,path}/renamed`. |
| 3 | :695 prefix the caller-Location check with `false &&` | `TestRepairRetryReauthorizesDefectsAndCallerLocation/explicit_stale_context` | Killed at `:174`. |
| 4 | :650 remove `result.Committed \|\|` | `…BoundariesAndDurableFailureReceipt/committed_without_applied_sources` | Killed by the portable fixture only. Real-store tests stay green, as expected, because no filesystem producer exists. |
| 5 | :650 remove both `result.Committed \|\|` and `len(result.AppliedSources) > 0 \|\|` (coordinated) | `TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix` | Killed at `graphrepair_test.go:282` ("durable prefix was retried"). |

### Production probes

The scratch `store` package test was removed before verify. A copy is kept
outside the sandbox at `scratchpad/zz_review_probe_test.go.saved`. CLI probes used
the sandbox-built binary on two throwaway trees, with `TSKFLW_CONFIG_HOME` pointed
at an empty scratch directory. That directory was still empty afterwards, so no
registry was written.

- **P1: valid hostile repair.**
  - Setup: an ID-only selection binds `<id>-zzsel.md`. Before verify, the hook adds
    a same-ID shadow `<id>-aa-shadow.md`, which becomes the representative. It also
    edits the selected file to `[invalid-token, concurrent-intent]` and adds a body
    line.
  - Result: one retry. `Committed=true`, `AppliedFiles=[selected]`, and `Removed`
    and `Addressed` contain exactly `selected:depends_on="invalid-token"#0`.
  - Residual: two `duplicate-task-id` problems, the shadow's `invalid-token` and the
    selected file's `concurrent-intent`. `FinalHealth` is `broken`.
  - Bytes: the selected file equals the concurrent bytes with `invalid-token, `
    removed and `updated_at: "2026-10-16"` added. The shadow is byte-identical.
- **P2: refused hostile repair.**
  - Setup: an auto+explicit call (slug selector). The explicit file is renamed before
    verify.
  - Result: `ErrNotFound` with the message "selected graph-repair source <id> at
    <path> moved, disappeared or changed identity; inspect the graph and rerun
    repair". `Committed=false`, `Changed=false`, no applied files.
  - Bytes: the auto-fixable file and the moved file are byte-identical.
- **P3: advice.** After P2, `InspectTaskGraphRepair` lists the defect at the moved
  path. A new `RepairTaskGraph` using that path commits only the moved file, and the
  auto file stays untouched.
- **P4: real-store mixed call.** The explicit edit (`automatic=false`) applies to its
  original path. Auto (`automatic=true`) repairs the renamed auto source and a newly
  added self-edge source. The old auto path is not recreated, the explicit
  concurrent body is kept, and `FinalHealth` is `healthy`.
- **P5: same-address declared-ID drift.** The retry proceeds, commits only the
  selected path, and leaves residual `task-id-drift`. The file of the task whose ID
  was claimed stays byte-identical. This is the approved limitation.
- **C1: CLI rerun advice.**
  - A bare `task depend repair` printed a copyable absolute-path `--drop`.
  - After `mv alpha -> alpha-renamed`, the stale slug and stale relative-path
    selectors both returned exit 10 ("does not identify a readable task record")
    with no write (`cmp` confirmed).
  - A fresh inspect showed the renamed path. Running the printed command committed;
    `diff` shows only the removal of `invalid-token, ` and the new `updated_at`, and
    the other task is byte-identical.
- **C2: optional manifest branches with non-default values.**
  - `path` plus `location: db://stale/context` plus `occurrence: 1` returned exit 14
    ("conflict: readable source location changed") with the file byte-identical.
  - `path` plus a matching absolute `location` plus `occurrence: 1` removed exactly
    self-dependency `#1` and left `#0` residual (`diff` confirmed).

### Surviving-hypothesis ledger

These are bounded; none is a finding.

1. **A permanently stale caller `location` is retried as a conflict.** On the
   filesystem, Location always equals the path, so a mismatched manifest `location`
   can never succeed. It still consumes the retry budget and exits 14
   (contention-style). It writes nothing. This behavior predates the change (the
   first-attempt `normalizeRepairEditSource` already returned `ErrConflict`); the
   change only applies it on retry too. It is a classification preference, not a
   wrong-target write. Uncertainty is low.
2. **Wording of the retry refusal for an unreadable source.** When the selected
   source becomes unreadable mid-call, the retry `ErrNotFound` text says "moved,
   disappeared or changed identity" and does not mention "unreadable". The advice
   is still correct: inspect shows the unreadable problem. Wording preference only.
3. **Representative-based impacts.** When the selected occurrence is a
   non-representative duplicate, receipt `Impacts` is empty even after a write (P1).
   Impacts are representative-ID projections, and that ID's graph state really is
   unchanged. This predates the change and is out of scope.
4. **Evidence routing for mixed calls.** No committed real-store regression covers
   a mixed auto/explicit retry. The portable test covers the shared Service path,
   and P4 showed correct real-store behavior. `persistence.md` asserts mixed-call
   rediagnosis but links only the store test files; the portable test is linked from
   `planning-data.md`. Documentation-routing nit.
5. **Cross-call replay after a durable prefix.** After a durable prefix,
   `TaskGraphRepairFailure` advises retrying "the same repair intent", and a new call
   re-resolves slug selectors afresh. This is outside the within-call contract,
   predates the change, and the cause text already says to inspect first.

### Validation run

All commands ran in the sandbox with Go 1.27.2 on darwin/arm64:

- `go test -race -count=3 ./internal/core/ -run TestRepair`: ok.
- `go test -race -count=3 ./internal/store/ -run 'TestRepairTaskGraph|TestMutateTaskGraphRepair'`: ok.
- `go test -race ./...` (the `just test` recipe): exit 0, no failing packages.
- `golangci-lint run ./...` (the `just lint` recipe, v2.14.0): `0 issues.`
- `go run ./internal/tools/docgen -out <scratch>` followed by `diff -r` against
  `docs/cli`: identical.
- `tskflwctl -C <sandbox> lint --links` with the sandbox-built binary: "all planning
  entities and dependency links pass lint".
- Final `git status --short` before verify: only this audit was modified.

### Isolation attestation

Output of the source helper
`scripts/isolated-review-workspace.sh verify --sandbox <sandbox>`, run after all
probes were restored:

```text
sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/0f1a264c-374a-4d81-aed9-af4b4b1bf70f/scratchpad/isolated-review.NVHD8y
git_dir=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/0f1a264c-374a-4d81-aed9-af4b4b1bf70f/scratchpad/isolated-review.NVHD8y/.git
baseline_commit=381204f05e874a393073f0ea486c4fe6c755c73e
source_blob=d0dd39f027365c68d6c1ee0d433952ee10989bcc
source_fingerprint=6c03cad4496b957a41b0b0aa830f528d46978756
deliverable=planning/audits/6gmbkgc5e1w9-2026-10-16-explicit-graph-repair-source-pinning-implementation-claude.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

The transfer receipt is returned in the handoff message, not in this report.

## Owner disposition (2026-10-16)

Delivery confirmed with the source helper's `receipt` command before any owner
edits: baseline `381204f05e874a393073f0ea486c4fe6c755c73e`, transferred blob
`5d343d2add8e1b18d984088212eba01a584d54a6`, `delivery=confirmed`.

Accept the no-product-findings verdict within the approved ID/path contract.
The mutation ledger distinguishes compiler-valid failures, coordinated durable-stop
mutation, and defensive-only port fixtures; same-address replacement is explicitly
not inode continuity. No hypothesis is promoted to a defect without a demonstrated
in-scope consequence.

Three useful small observations are incorporated: retained a real-store mixed
auto/explicit retry test (both successful rediagnosis and no writes when the
explicit source moves), made the missing-occurrence message mention unreadability,
and linked the persistence guide directly to portable retry evidence. The report's
pre-existing location-error classification and representative-state impact behavior
are not changed. Cross-call selection remains a new intent boundary, not a promise
of automatic replay. No new task is warranted for those bounded non-defects.
