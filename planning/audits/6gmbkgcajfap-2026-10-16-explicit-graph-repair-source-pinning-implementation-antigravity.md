---
schema: 1
id: 6gmbkgcajfap
bucket: closed
area: explicit-graph-repair-source-pinning-implementation-antigravity
date: "2026-10-16"
---
# Audit: Explicit graph repair source pinning implementation — antigravity — 2026-10-16

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

### Verdict and review limits

- **Verdict:** Ready (0 product findings: 0 high, 0 medium, 0 low).
- **Scope and limits:** Independent review conducted entirely within isolated clone `/private/tmp/isolated-review.52y6HY` at sandbox baseline `94f99222` (over `552a9b863a88` on `fix/pin-explicit-graph-repair-sources`). Primary lens: filesystem source targeting, raw rename/deletion and source preservation across retries, executable CLI `--drop` and `--plan` behavior, inspect/rerun recovery after legitimate rename, and cross-check of the portable empty-result retry regression.
- **Approved boundary limits respected:** Source pinning binds the logical `(TaskID, LocalPath)` occurrence address established during the first successful guarded planner callback (`internal/core/dependency_repair.go:636-648`, `664-702`), not an OS inode or cross-rename capability. Byte replacement at the identical `(TaskID, LocalPath)` address is evaluated as current content at that address and reauthorized from scratch, as explicitly approved for task `6gm9ny98zyf2`. Separately tracked concurrency audit items (`6gmbfxy1d561` for lint ID relocation M1/M2 and `6gmbfxy96tpw` for TUI watcher session binding L1) do not intersect `RepairTaskGraph` or `FS.MutateTaskGraphRepair`.

### Verified consumer and implementation inventory

- **CLI primary adapter (`internal/cli/task_dependency_repair.go`):**
  - `newTaskDependencyRepairCmd` (`internal/cli/task_dependency_repair.go:45-104`): updates `Long` help (`L52`) and generated reference `docs/cli/tskflwctl_task_depend_repair.md:7` to state that within one call, explicit selections stay bound to their first source ID and path while moved or missing sources require inspection and a new call. Calls `app.Svc.InspectTaskGraphRepair()` (`L58`) when no mutation flag is given and `app.Svc.RepairTaskGraph(request, app.DryRun)` (`L88`) when `--auto`, `--drop`, `--dedupe`, or `--plan` is supplied.
  - `parseGraphRepairSelector` (`internal/cli/task_dependency_repair.go:120-148`) and `normalizeGraphRepairPath` (`L106-L111`): reject URI-shaped `://` opaque locations (`L125-L127`), classify path-separator or absolute inputs as `Source.LocalPath`, valid Crockford IDs as `Source.TaskID`, and other tokens as `Source.TaskSlug`, then join relative `LocalPath` values under `app.Cfg.PlanningRoot`.
  - `readGraphRepairManifest` (`internal/cli/task_dependency_repair.go:179-241`): enforces `schema: 1`, validates action/field/occurrence, rejects `://` in `path`/legacy `location`, and sets `source.Location = operation.Location` only as a stale-context check when `operation.Path != ""` (`L220-L224`).
- **Core use case and planner (`internal/core/dependency_repair.go`, `internal/core/dependency_source.go`, `internal/core/service_task.go`, `internal/core/store.go`):**
  - `Service.RepairTaskGraph` (`internal/core/dependency_repair.go:627-658`): clones `request.Edits` once (`L632`) so caller slice mutation cannot alter in-flight retry state; maintains `var selectedSources []TaskGraphSourceRef` (`L633`) across `MutateTaskGraphRepair` attempts; calls `resolveRepairRetryEdits(graph, request.Edits, selectedSources)` (`L637`) followed by `PlanTaskGraphRepair` (`L641`); binds `selectedSources = sources` inside the callback when `err == nil && selectedSources == nil` (`L642-L646`) so a post-planner adapter `ErrConflict` that returns an empty `TaskGraphRepairMutationResult` cannot erase the selection; and stops retrying when `dryRun || !errors.Is(err, domain.ErrConflict) || result.Committed || len(result.AppliedSources) > 0 || attempt >= s.maxRetries` (`L650`).
  - `resolveRepairRetryEdits` (`internal/core/dependency_repair.go:664-702`): validates edit shape (`L668`); delegates initial resolution (`selected == nil`) to `normalizeRepairEditSource` (`L671-L678`); on retry (`selected != nil`), scans `graph.SourceRecords()` (`L679-L688`) for `record.Source.TaskID == selected[i].TaskID && record.Source.LocalPath == selected[i].LocalPath`. Zero matches fail closed with `domain.ErrNotFound` (`L689-L691`), multiple matches fail closed with `domain.ErrAmbiguous` (`L692-L694`), and caller-supplied `edit.Source.Location` mismatches fail closed with `domain.ErrConflict` (`L695-L697`), while refreshed `TaskSlug` and `Location` from `matches[0]` replace `edit.Source` (`L698-L699`).
  - `PlanTaskGraphRepair` (`internal/core/dependency_repair.go:332-420`), `normalizeRepairEditSource` (`L874-L906`), `authorizeRequestedRepair` (`L810-L829`), and `repairEditTargetPresent` (`L831-L872`): removed the dead `sourcePresent` boolean return from `normalizeRepairEditSource` (`case 0` always returned `domain.ErrNotFound`). Already-removed declarations on a still-present source converge (`alreadySatisfied == true` at `L825-L827`), whereas a present non-defective declaration fails closed with `domain.ErrValidation` (`L828`). Auto selections (`request.Auto`) are computed from `diagnosis.Defects` (`L340-L349`) and `prospectiveDiagnosis.Defects` (`L395-L403`) on each call and never enter `request.Edits` or `selectedSources`.
  - `TaskGraph.SourceRecords` (`internal/core/dependency_source.go:90-95`) and `validatedTaskGraphRead` (`internal/core/service_task.go:256-277`): require `g.sourceComplete` (`internal/core/dependency_source.go:236-244`) and reject records lacking a valid `Record.Source.ID` before constructing `TaskGraph.sourceRefs`.
  - `TaskGraphRepairStore` port comment (`internal/core/store.go:142-150`): documents callback-level `selectedSources` retention across empty conflict results.
- **Filesystem store secondary adapter (`internal/store/graphrepair.go`, `internal/store/fsstore.go`, `internal/store/entity_read.go`, `internal/store/lock.go`, `internal/store/lock_unix.go`):**
  - `FS.MutateTaskGraphRepair` (`internal/store/graphrepair.go:21-170`): acquires the repository guard via `s.checkedWriteLock()` (`internal/store/lock.go:115-137`), combining in-process mutex `processRepositoryLock(s.root)` (`internal/store/lock.go:63-67`) and advisory `syscall.Flock(..., syscall.LOCK_EX)` on `s.root` (`internal/store/lock_unix.go:21-40`). Loads `ReadThreads()` (`L49`) and `core.LoadTaskGraph(s)` (`L53`), runs `callTaskGraphRepairPlanner` (`L60`, guarded against re-entry by `enterRepositoryPlanner` at `internal/store/lock.go:69-83`), validates via `core.ValidateTaskGraphRepairPlan` (`L64`), and materializes per-file edits via `s.materializeTaskGraphRepair` (`L68`).
  - Pre-write receipt reset (`internal/store/graphrepair.go:82-83`): resets `result.Analysis` to the empty-plan analysis on `graph` and `result.FinalGraph = graph` before `testHookBeforeGraphRepairVerify` (`L88-L90`) and the pre-write snapshot/per-file CAS checks (`L91-L127`), ensuring a pre-write conflict never reports prospective state as durable.
  - `FS.materializeTaskGraphRepair` (`internal/store/graphrepair.go:188-234`): enforces `s.tasksDir` containment via `filepath.Rel(s.tasksDir, path)` (`L195-L198`), applies surgical YAML edits via `updateDependencySourceEdits` (`L203`), reparses with `parseTask(updated, path)` (`L210`), and verifies reload projection using `taskSource(path)` (`internal/store/entity_read.go:14-17`) against `analysis.Prospective` (`L216-L228`).
- **Regression suites and documentation (`internal/core/dependency_repair_retry_test.go`, `internal/store/graphrepair_target_test.go`, `internal/store/graphrepair_test.go`, `docs/`):**
  - Portable core retry suite (`internal/core/dependency_repair_retry_test.go:1-299`) and real-store target suite (`internal/store/graphrepair_target_test.go:1-237`), plus nonzero retry budget on durable prefix conflict in `TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix` (`internal/store/graphrepair_test.go:277,282`).
  - Architecture guides updated in `docs/architecture/planning-data.md:77-86` and `docs/architecture/persistence.md:149-157`. Wire schema version remains `1.88` (`internal/wire/wire.go:20`), consistent with ADR-0008.
  - **Defensive port coverage classification:** The `committed without applied sources` case in `TestRepairRetryBoundariesAndDurableFailureReceipt` (`internal/core/dependency_repair_retry_test.go:228`) exercising `result.Committed` in `internal/core/dependency_repair.go:650` is **defensive port coverage**. In the filesystem producer `FS.MutateTaskGraphRepair` (`internal/store/graphrepair.go:146-148`), `result.Committed = true` is assigned only immediately after `result.AppliedSources = append(result.AppliedSources, write.source)`, so `store.FS` has no reachable path that sets `Committed: true` with an empty `AppliedSources` slice.

### Findings

No product defects, coverage gaps, or policy drift were found.

### Mandatory evidence and hostile probes

1. **Focused race test execution:**
   - Built `./bin/tskflwctl` inside `$SANDBOX` with isolated `GOCACHE=/private/tmp/ag-review-cache-52y6HY/gocache`.
   - Ran `go test -race -count=3 ./internal/core -run 'Test.*Repair.*'` (`ok`, 1.305s), `go test -race -count=3 ./internal/store -run 'Test.*GraphRepair.*|Test.*RepairTaskGraph.*'` (`ok`, 2.150s), `go test -race -count=1 ./internal/cli -run 'Test.*Repair.*|Test.*Depend.*'` (`ok`, 1.662s), and `./bin/tskflwctl -C . lint --links` (`✔ all planning entities and dependency links pass lint`).
2. **Executable CLI production probe and selector -> Service -> store trace:**
   - In an unregistered throwaway tree `/private/tmp/ag-review-cache-52y6HY/throwaway/planning/tasks` (with isolated `HOME` and `XDG_CONFIG_HOME`), seeded `ekpd8vnydg2h-alpha.md` (`depends_on: [invalid-token, aaaaaaaaaaaa]`, `custom_key: keep-alpha # comment stays`, body prose) and `h46gx48dzz67-bravo.md` (`depends_on: [invalid-token]`, `custom_key: keep-bravo # bravo comment`, body prose).
   - **Trace:**
     1. `--drop 'alpha:depends_on=invalid-token#0' --json` entered `newTaskDependencyRepairCmd` (`internal/cli/task_dependency_repair.go:65-71`) -> `parseGraphRepairSelector` (`L120-L148`) produced `TaskGraphSourceEdit{Action: "drop", Source: {TaskSlug: "alpha"}, Field: "depends_on", Value: "invalid-token", Occurrence: 0}` -> `app.Svc.RepairTaskGraph` (`internal/core/dependency_repair.go:627`).
     2. Inside `FS.MutateTaskGraphRepair` (`internal/store/graphrepair.go:21`), under `checkedWriteLock()`, `resolveRepairRetryEdits` (`internal/core/dependency_repair.go:672`) called `normalizeRepairEditSource` (`L874`), matching `ekpd8vnydg2h` at `.../planning/tasks/ekpd8vnydg2h-alpha.md` and binding `selectedSources = [{TaskID: "ekpd8vnydg2h", TaskSlug: "alpha", Location: ".../ekpd8vnydg2h-alpha.md", LocalPath: ".../ekpd8vnydg2h-alpha.md"}]`.
     3. `FS.materializeTaskGraphRepair` (`internal/store/graphrepair.go:188`) rewrote only `ekpd8vnydg2h-alpha.md` (`depends_on: [aaaaaaaaaaaa]`, preserving `custom_key: keep-alpha # comment stays` and body prose), while `cmp -s` confirmed `h46gx48dzz67-bravo.md` remained byte-for-byte identical.
     4. Invoking `--plan /private/tmp/ag-review-cache-52y6HY/throwaway/plan.yaml --json` with relative `path: tasks/ekpd8vnydg2h-alpha.md` (`readGraphRepairManifest` at `internal/cli/task_dependency_repair.go:179` -> `normalizeGraphRepairPath` at `L106`) removed `aaaaaaaaaaaa` and deleted the now-empty canonical `depends_on` key while preserving `custom_key: keep-alpha # comment stays` and leaving `h46gx48dzz67-bravo.md` byte-for-byte untouched.
3. **Real-store hostile repair, refusal, and inspect/rerun recovery verification:**
   - Exercised pre-verify slug swap and rename refusal through `testHookBeforeGraphRepairVerify` (`internal/store/graphrepair.go:88`, `internal/store/graphrepair_target_test.go:16-119`) and verified with a temporary real-store probe (`TestAntigravityHostileProbeInspectAndRerunPlusFrontmatterIDDrift`, restored after run) that:
     - When `alpha` is renamed to `<id>-renamed.md` and `bravo` is renamed to `<id>-alpha.md` at `testHookBeforeGraphRepairVerify`, `svc.RepairTaskGraph` refuses the retry with `domain.ErrNotFound` (`selected graph-repair source ... moved, disappeared or changed identity; inspect the graph and rerun repair`), `receipt.Committed == false`, `len(receipt.AppliedFiles) == 0`, and both files retain their exact post-rename bytes.
     - Following the error's advice on that same fixture via `svc.InspectTaskGraphRepair()` surfaces the moved defect at `LocalPath: renamed`, and passing that inspected `Target` to a new `svc.RepairTaskGraph` call commits the repair on `renamed` (`Committed: true`, `AppliedFiles: [renamed]`) while leaving the `bravo` (`-alpha.md`) impostor bytes untouched. Also verified via CLI (`mv h46gx48dzz67-bravo.md h46gx48dzz67-bravo-renamed.md` -> `tskflwctl task depend repair --json` -> `tskflwctl task depend repair --drop 'bravo-renamed:depends_on=invalid-token#0' --json` -> `final_health: "healthy"`).

### Mutation ledger

| Probe | Exact edit | Target test / case | Result | Restoration & verification |
| --- | --- | --- | --- | --- |
| M-AG-1 (Assigned LocalPath conjunct removal) | In `internal/core/dependency_repair.go:685`, changed `if record.Source.TaskID == selected[i].TaskID && record.Source.LocalPath == selected[i].LocalPath` to `if record.Source.TaskID == selected[i].TaskID` | `TestRepairRetriesRefuseMissingReplacedAndAmbiguousOccurrences` (`internal/core/dependency_repair_retry_test.go:116`) and `TestRepairTaskGraphRetryKeepsDuplicateIDOccurrenceAndConcurrentContent` / `TestRepairTaskGraphRetryRefusesMovedMissingAndUnreadableSources` (`internal/store/graphrepair_target_test.go:61,121`) | **Killed (compiler-valid):** In `internal/core`, subcases `renamed/{ID,slug,path}` and `duplicate_ID_elsewhere_cannot_replace_deleted_occurrence/{ID,slug,path}` failed (`err=<nil>, want not found`). In `internal/store`, `TestRepairTaskGraphRetryKeepsDuplicateIDOccurrenceAndConcurrentContent/duplicate-initially={false,true}` failed with `ambiguous match: selected graph-repair source ... is no longer a unique occurrence`, and `TestRepairTaskGraphRetryRefusesMovedMissingAndUnreadableSources/{ID,slug,path}/renamed` failed (`Committed:true`, `err=<nil>`). | Restored `internal/core/dependency_repair.go:685`; re-ran both test commands with `-race -v -count=1` (`PASS`). |
| M-AG-2 (Cross-check empty-result callback binding) | In `internal/core/dependency_repair.go:645`, replaced `selectedSources = sources` inside the `MutateTaskGraphRepair` planner callback with `_ = sources` | `TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent` (`internal/core/dependency_repair_retry_test.go:66`) and `TestRepairTaskGraphRetryNeverReselectsReusedSlug` (`internal/store/graphrepair_target_test.go:16`) | **Killed (compiler-valid):** `TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent` failed at `dependency_repair_retry_test.go:103` (`explicit selection drifted` onto `impostor` `z3cxfnh17vq6`), and `TestRepairTaskGraphRetryNeverReselectsReusedSlug` failed at `graphrepair_target_test.go:47` (`retry repaired the reused slug's different source: applied=[.../h46gx48dzz67-alpha.md] err=<nil>`). | Restored `internal/core/dependency_repair.go:645`; re-ran both tests with `-race -v -count=1` (`PASS`). |

### Required hostile angles and surviving-hypothesis ledger

| Hypothesis / hostile angle | Status | Evidence |
| --- | --- | --- |
| H1: An empty conflict result or intervening pre-planner conflict erases a previous explicit selection, allowing a reused slug/path to target another record. | **Falsified** | `selectedSources` is bound inside the `MutateTaskGraphRepair` planner callback (`internal/core/dependency_repair.go:642-646`) as soon as `PlanTaskGraphRepair` succeeds, before the adapter can return an empty `TaskGraphRepairMutationResult` on a later validation/CAS conflict, and is guarded by `selectedSources == nil` so later pre-planner conflicts (`step.graph == nil`) cannot overwrite it. Verified by `TestRepairRetriesRetainSourcesAcrossEmptyResultsAndReplanFreshContent` (`internal/core/dependency_repair_retry_test.go:66-114`) and mutation probe `M-AG-2`. |
| H2: Duplicate IDs cause representative-task lookup to substitute an occurrence, or overly broad refusal prevents safe exact-path repair while damage remains residual. | **Falsified** | Both `normalizeRepairEditSource` (`internal/core/dependency_repair.go:878-905`) and `resolveRepairRetryEdits` (`L679-L700`) iterate `graph.SourceRecords()` (`internal/core/dependency_source.go:90-106`), which preserves all physical occurrences in `g.sourceTasks`/`g.sourceRefs` rather than consulting the `g.tasks` or `g.representative` maps. Selecting an occurrence by exact `LocalPath` when duplicate IDs exist initially (`duplicate-initially=true`) or when a duplicate shadow is created concurrently (`duplicate-initially=false`) repairs only the selected path while leaving `ProblemDuplicateTaskID` residual (`TestRepairTaskGraphRetryKeepsDuplicateIDOccurrenceAndConcurrentContent`, `internal/store/graphrepair_target_test.go:121-172`, and mutation probe `M-AG-1`). |
| H3: Retrying a drop/dedupe/empty-key selection removes newly valid or newly added intent, caches old occurrences, loses unrelated body/frontmatter changes, or claims the old prospective graph as durable state. | **Falsified** | Every retry calls `PlanTaskGraphRepair` (`internal/core/dependency_repair.go:641`) on the newly loaded `TaskGraph` and re-runs `FS.materializeTaskGraphRepair` (`internal/store/graphrepair.go:68,188-234`) against fresh file bytes read under the repository lock. Pre-write conflicts reset `result.Analysis` and `result.FinalGraph = graph` (`internal/store/graphrepair.go:82-83`). Verified by `TestRepairRetryReauthorizesDefectsAndCallerLocation` (`internal/core/dependency_repair_retry_test.go:152-181`), `TestRepairRetryReauthorizesDedupeAndEmptyFieldIntent` (`L254-L298`), and `TestRepairTaskGraphRetryReauthorizesCurrentDeclaration` (`internal/store/graphrepair_target_test.go:174-215`). |
| H4: Auto-only and mixed calls retain stale auto targets or expand explicit intent; missing sources are incorrectly treated as already repaired. | **Falsified** | `resolveRepairRetryEdits` (`internal/core/dependency_repair.go:664-702`) operates only on `request.Edits`; `request.Auto` defects are synthesized inside `PlanTaskGraphRepair` (`L340-L349`, `L395-L403`) per attempt and never enter `selectedSources`. Missing sources fail closed with `domain.ErrNotFound` in both `normalizeRepairEditSource` (`L893`) and `resolveRepairRetryEdits` (`L690`); the former unreachable `!sourcePresent` branch in `PlanTaskGraphRepair` was removed. Verified by `TestRepairRetriesRediagnoseAutoWithoutExpandingExplicitBindings` (`internal/core/dependency_repair_retry_test.go:183-217`) and `TestRepairTaskGraphAutoRetryCanFollowNewSourceSet` (`internal/store/graphrepair_target_test.go:217-236`). |
| H5 (Systemic pass): Portable test helper `localGuardedRecordAt` (`internal/core/dependency_graph_test.go:64-71`) sets `Record.Source.ID = task.ID` (from the struct field), whereas `FS.ReadTaskGraph` (`internal/store/fsstore.go:168`, `internal/store/entity_read.go:14-17`) derives `Record.Source.ID` and `Task.Slug` from the filename stem `<id>-<slug>.md`. Could a concurrent in-place edit to frontmatter `id:` at the same filename bypass or break retry source pinning in `store.FS`? | **Falsified (with bounded adapter distinction noted)** | Executed a real-store probe (`same filename concurrent frontmatter ID drift preserves adapter identity`) that mutated frontmatter `id:` to a different valid ID at `testHookBeforeGraphRepairVerify` without renaming the file. Because `taskSource(path)` derives `Record.Source.ID` from the filename stem and `FS.materializeTaskGraphRepair` (`internal/store/graphrepair.go:216-228`) reloads with `Source: taskSource(path)`, the retry matched `(TaskID, LocalPath)`, repaired the defective dependency, preserved the concurrent frontmatter `id:` edit, and reported `ProblemTaskIDDrift` (`FinalHealth: GraphBroken`). Conversely, on `store.FS`, `TestRepairRetriesRefuseMissingReplacedAndAmbiguousOccurrences/path_reused_with_other_ID` (`internal/core/dependency_repair_retry_test.go:131`) is a portable-adapter test case because a flat task path `<id>-<slug>.md` cannot change its filename-derived `Record.Source.ID` without changing `LocalPath`; same-filename byte replacement on `store.FS` is therefore bounded by the user-approved logical `(TaskID, LocalPath)` identity model rather than inode continuity. |

### Isolation and verification attestation

```text
sandbox_path=/private/tmp/isolated-review.52y6HY
git_dir=/private/tmp/isolated-review.52y6HY/.git
baseline_commit=94f99222b648d7dc43fd8b12676baf344df00cc4
source_blob=5af27bdfce8ddef96b3945324cbd88aac8d1bad1
source_fingerprint=6c03cad4496b957a41b0b0aa830f528d46978756
deliverable=planning/audits/6gmbkgcajfap-2026-10-16-explicit-graph-repair-source-pinning-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

## Owner disposition (2026-10-16)

Delivery confirmed with the source helper's `receipt` command before any owner
edits: baseline `94f99222b648d7dc43fd8b12676baf344df00cc4`, transferred blob
`3170c4bee1c3721e3ef16f230d4e50ac727c59c6`, `delivery=confirmed`.

Accept the no-product-findings verdict. The report supplies production CLI/store
traces, two discriminating compiler-valid mutations and a same-filename ID-drift
probe, rather than using a green baseline as approval. Its distinction between
portable-only cases and the filename-derived filesystem producer is appropriate.
One terminology clarification: `validatedTaskGraphRead` checks for a nonempty
adapter source ID; full identity syntax/graph health is diagnosed separately. It
does not certify arbitrary frontmatter identity as valid.

The approved logical ID/path limit remains explicit. Owner follow-through from
the complementary review adds a permanent real-store mixed-call regression,
unreadable-source advice and a direct portable-test documentation link. No product
finding or additional untracked follow-up remains from this review.
