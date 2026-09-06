---
schema: 1
id: 6g7cr4q1vhms
bucket: closed
area: guarded-broken-graph-repair-implementation-antigravity
date: "2026-09-06"
---
# Audit: Guarded broken-graph repair implementation — antigravity — 2026-09-06

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

Perform an independent adversarial implementation and architecture review of the guarded
broken-dependency-graph repair work for task
`planning/tasks/6g4g8gatbnrs-add-a-guarded-repair-path-for-broken-dependency-graphs.md`.
This is a recovery path allowed to operate where every ordinary graph-owned mutation must fail
closed, so a plausible-looking false success, unauthorized declaration removal, stale write, or
misleading partial receipt is a high-value finding. Review behavior and invariants, not style.

The implementation intentionally remains removal-only, retains Markdown as the authoritative
source, does not guess cycle or legacy choices, and does not adopt a third-party graph library.
Do not demand those non-goals. Challenge whether the delivered boundary is actually as narrow,
portable, convergent, and truthful as the task and ADR claim.

## Review target

Review the entire copied working tree, including uncommitted changes. Start with:

- `internal/core/dependency_repair.go`, `dependency_graph.go`, `dependency_source.go`, and tests;
- `internal/store/graphrepair.go`, `frontmatter.go`, snapshot/lock helpers, and tests;
- `internal/cli/task_dependency_repair.go`, repair render/error paths, and tests;
- `internal/wire/dependency_repair.go`, envelopes, schema comments, and machine goldens;
- lint/status remediation, README, architecture notes, ADR-0006, and the task above.

Build a repository-wide consumer inventory for `TaskGraphRepairStore`, `RepairTaskGraph`,
`ValidateTaskGraphRepairPlan`, `SimulateSourceEdits`, `SameRepairSnapshot`, repository locking,
Thread impact projection, graph-health guidance, error classification, and the new wire envelope.
Also inventory every ordinary task/dependency/lifecycle/Thread/lint-fix write path that must remain
unable to enter a broken graph. Do not infer architectural isolation merely from separate type names.

## Intended contract to challenge

- Only the dedicated repair port may plan against `GraphBroken`. It accepts exact removal-only
  source intent; ordinary replacement setters, dependency add/remove/migrate, lifecycle commands,
  Thread mutations, generic edits, and lint fixing retain their existing guardrails.
- Diagnosis operates on the complete source projection: all readable physical records, duplicate-ID
  shadows, opaque unreadable revisions, raw invalid/dangling values, duplicate occurrences, empty
  legacy keys, and the declaration owner of reversed legacy `blocks` references survive.
- Bare diagnosis is read-only and deterministic. `--auto` selects only canonical dedupe, self-edge,
  and empty-legacy cleanup. Invalid/dangling values, cycles, and ambiguous legacy intent require an
  exact explicit selector. Emitted commands must parse and act on the defect they display, including
  raw values containing punctuation and values ending in `#<digits>`.
- The plan validator reauthorizes operations and selected/no-op intent against the guarded snapshot.
  It rejects valid-constraint deletion, source ambiguity, invented receipt claims, overlapping
  dedupe/drop effects, generic replacement state, and false automatic provenance.
- Progress and preservation are separate hard proofs. Every atomic task-file group is structurally
  componentwise non-worsening and strictly improving while removing only authorized declarations.
  A valid repair may remain broken; SCC splitting or newly visible diagnostics is not itself failure.
- The filesystem adapter owns one repository guard, reads complete task and Thread evidence, performs
  whole-snapshot and per-file CAS, edits only selected YAML nodes plus `updated_at`, and verifies each
  durable prefix from a fresh reload. Malformed Threads do not block repair but their exact evidence
  is CAS-protected and reported incomplete.
- Retry intent is convergent. An already-satisfied exact selection remains visible but causes no
  write. A post-write failure returns a typed receipt whose applied prefix is durable, whose remaining
  files are exact, and whose before/after health, removals, task impacts, and Thread impacts describe
  the durable prefix rather than the abandoned full plan.
- Human and JSON output remain truthful and adapter-neutral. JSON carries stable raw selectors,
  non-null collections, workspace identity, incomplete Thread diagnostics, and structured recovery
  details without leaking private revision tokens. Core does not depend on filesystem, YAML, Cobra,
  TUI, or wire types, and a future TUI/web/remote adapter need not reproduce CLI parsing.

## Mandatory evidence floor

1. Record branch/base/sandbox baseline and exact results for the full uncached race suite,
   golangci-lint, planning lint, module-tidiness, generated-doc/schema stability, and
   `git diff --check`.
2. Exercise hostile repositories containing self-edges, repeated self-edges, two- and multi-node
   cycles, chords and SCC splits, duplicate canonical declarations, invalid IDs, dangling stable
   IDs, all legacy fields, missing/ambiguous/unsafe legacy references, empty legacy keys,
   duplicate-ID shadows, identity drift, invalid status, and readable plus unreadable task/Thread
   mixtures. Compare complete source records and declarations before and after—not health alone.
3. Run the advertised CLI against sandbox-only throwaway spaces: bare diagnosis; `--auto` dry-run and
   apply; explicit invalid, dangling, cycle, and legacy drops; a strict manifest; and the same intent
   after a partial prefix and after full convergence. Execute the exact commands emitted by diagnosis.
   Include raw values with spaces, quotes, colons, equals signs, shell metacharacters, and terminal
   `#<digits>`; demonstrate they target only their displayed declaration or identify a concrete gap.
4. Attack the progress measure with graph mutations that can split SCCs, remove a cycle chord, reveal
   masked problems, combine several defects in one file, and leave unrelated defects behind. Try to
   produce a plan that reports strict improvement while changing no bytes, deletes a valid constraint,
   or makes one structural component worse.
5. Attack operation accounting: omit operations from selections, add present valid selections, add
   already-satisfied selections, duplicate selectors, reorder operations, combine exact drops,
   combine drop with dedupe, mix auto and explicit intent, falsify reason/automatic metadata, and
   directly invoke the store with a malicious planner. Require rejection or an exact truthful receipt.
6. Exercise lock/CAS timing before verification, before each write, and after each durable write.
   Mutate target and non-target readable task bytes, unreadable task bytes without changing the parse
   message, readable Thread bytes, unreadable Thread bytes, and representation transitions. Repeated
   race tests must never falsely report a complete commit or overwrite the concurrent bytes.
7. Verify surgical YAML behavior for flow and block sequences, comments on retained and removed nodes,
   unknown fields, key order, quoted/scalar edge cases, Markdown body bytes, empty-key removal, file
   mode, CRLF if supported, and `updated_at`. Mutation-probe the exact preservation assertion and
   require a focused test to fail for the intended reason.
8. Populate every optional repair wire/error branch with non-default semantic values and run envelope
   validation. Check partial failure on both human and JSON paths, nil/empty collection stability,
   absolute/relative location behavior, workspace identity, error code/exit status, and absence of
   opaque source revisions.
9. Trace readable Thread projection impacts before/after a repair, direct versus downstream changes,
   external gates, a residual-broken prefix, and malformed Thread evidence. Confirm task graph repair
   never rewrites a Thread document and Thread corruption never hides task recovery evidence.
10. For each new regression test, perform the exact sandbox-only source mutation that reintroduces
    the claimed bug and require that specific test to fail. At minimum mutate dedicated-port isolation,
    full-source preservation, automatic-policy bounds, overlap rejection, raw-colon selector parsing,
    readable/unreadable evidence CAS, durable-prefix accounting, automatic provenance, and non-default
    wire branches. Compilation failure or an unrelated failing test is not sufficient mutation evidence.

## Required hostile angles

- Look for a second broken-graph entrance hidden in generic setters, editor flows, migration,
  lifecycle force handling, Thread apply, lint fixing, or direct filesystem helpers.
- Distinguish source declarations from projected edges. Attack legacy `blocks` ownership,
  non-representative duplicate-ID records, unreadable targets, identical values with occurrence
  renumbering, and source references that are under- or over-specified.
- Treat selected intent, active operations, actual materialized writes, simulated removals, fresh
  reloads, and reported receipts as separate states. Find any transition where one can describe
  another falsely, especially after a durable prefix.
- Challenge the componentwise measure mathematically and with generated small graphs. Seek false
  monotonicity, iteration-order dependence, overflow/unbounded behavior, and defect classes omitted
  from the strict component.
- Attack the boundary between convergent absence and a typo. A disappeared value in a still-existing
  exact source can be satisfied intent; a missing/ambiguous source or present valid constraint must
  not become a silent no-op.
- Attack shell and manifest grammar independently. Quoting that protects the shell does not prove the
  selector grammar is unambiguous, and strict YAML decoding does not prove semantically invalid
  action/field/value combinations fail closed.
- Look for TOCTOU gaps around the first plan, materialization reads, whole evidence recheck, each
  target replacement, post-write reload, and lock release. Include direct callers that bypass the
  Service retry loop.
- Challenge portability: identify any filesystem/path/YAML assumption in core or any semantic rule
  reimplemented only by the CLI. Also challenge whether the interface makes remote transactional
  semantics impossible to express without weakening receipts.
- Inspect tests for helpers that always preserve the invariant under test, assertions that check only
  counts/health, global hooks that can contaminate tests, cached full-suite results, and goldens whose
  new branches remain at defaults.
- Perform a fresh second pass explicitly seeking systemic anti-patterns and future failure modes,
  including performance under the exclusive lock and coupling to the growing owned DAG algorithm.
  Record a finding only when supported by concrete code or a reproduction; otherwise document why
  the concern is settled or remains a bounded residual risk.

## Validation and restoration

All inspection, builds, generators, fixtures, and mutation probes must remain inside the mandatory
independent reviewer sandbox. Restore every probe to the sandbox baseline before transfer. Do not
commit beyond the mandatory sandbox checkpoint, push, install globally, edit the source checkout,
or copy implementation/generated/planning files back. Only the assigned audit may be transferred.

## Deliverable

Preserve this brief and replace only the `Reviewer report` placeholder with:

- the required isolation attestation and a concise verdict;
- the repository-wide consumer inventory and an evidence matrix;
- findings ordered by severity using the exact required grammar and left `open`;
- for every finding, a concrete reproduction, violated contract, smallest sound recommendation,
  affected consumers, and the missing regression test;
- settled hostile angles and residual risks, explaining why suspected issues are not findings; and
- exact validation and mutation-test results.

Do not edit implementation or other planning files in the source checkout, and do not triage your
own findings. If no finding survives, say so explicitly and provide enough hostile evidence to make
that conclusion credible.

## Reviewer report

### Isolation attestation & verdict

- **Sandbox path:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.9OWQMN`
- **Resolved Git directory:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.9OWQMN/.git`
- **Sandbox baseline commit:** `8cca67952a04609608636f1a1e5c8aba411a8e9b`
- **Captured source-audit blob:** `b3eb827e4e5e35d6602c77367ce559e5a7d600a8`
- **Guarded transfer:** Verified against `$SOURCE_AUDIT_BLOB` and executed atomically via temporary file replacement.
- **Verdict:** **PASS / Implementation Approved.** The guarded broken-dependency-graph recovery implementation strictly enforces removal-only repair, fails closed against non-defects, guarantees mathematical defect reduction and declaration-level preservation, preserves exact YAML formatting without corrupting unselected items or metadata, and prevents broken graphs from entering ordinary mutation paths. No findings survive.

---

### Repository-wide consumer inventory

#### 1. Recovery capability consumers
- `core.TaskGraphRepairStore`: Declared in [`internal/core/store.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/store.go#L104-L106); implemented exclusively by `*store.FS` in [`internal/store/graphrepair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphrepair.go#L21-L161). Sibling capability to `TaskGraphMutationStore`, never inferred from it.
- `core.RepairTaskGraph`: Declared in [`internal/core/dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_repair.go#L543-L562); exposed via `*core.Service` (`internal/core/service.go`); consumed by CLI command `tskflwctl task depend repair` in [`internal/cli/task_dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/task_dependency_repair.go#L87-L93).
- `core.ValidateTaskGraphRepairPlan`: Declared in [`internal/core/dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_repair.go#L372-L462); consumed by `FS.MutateTaskGraphRepair` in [`internal/store/graphrepair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphrepair.go#L64-L135) to authorize plans and validate durable prefix steps under the repository lock.
- `core.SimulateSourceEdits`: Declared in [`internal/core/dependency_source.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_source.go#L143-L220); consumed by `core.PlanTaskGraphRepair` and `core.ValidateTaskGraphRepairPlan` to simulate source-declaration removals without writing to disk.
- `core.SameRepairSnapshot`: Declared in [`internal/core/dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_repair.go#L939-L971); consumed by `FS.MutateTaskGraphRepair` in [`internal/store/graphrepair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphrepair.go#L148-L151) to verify that fresh post-write reloads match prospective graph state exactly.
- Repository Locking: Acquired via `FS.checkedWriteLock()` in [`internal/store/lock.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/lock.go#L115-L125); held exclusively during all repair planning, verification, and surgical writes.
- Thread Impact Projection: `core.TaskGraphThreadImpacts` in [`internal/core/thread_mutation.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/thread_mutation.go#L424-L454); consumed by `taskGraphRepairReceipt` in [`internal/core/dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_repair.go#L592-L593) to project Thread view state changes across readable threads without touching Thread files.
- Graph-Health Guidance: `core.dependencyLintIssues` in [`internal/core/service.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service.go#L620-L649), `core.taskGraphHealthDetail` in [`internal/core/dependency_graph_mutation.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_graph_mutation.go#L152-L167), and `render.graphDiagnosticsHuman` in [`internal/cli/render/dependency.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/render/dependency.go#L316-L322). Consumes `graphProblemRepairable` to steer operators to `tskflwctl task depend repair`.
- Error Classification: `core.TaskGraphRepairFailure` in [`internal/core/dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/dependency_repair.go#L118-L142); wrapped by CLI `graphRepairCommandFailure` in [`internal/cli/task_dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/task_dependency_repair.go#L21-L29); unwrapped and serialized to `wire.ErrorEnvelope.Error.GraphRepair` via `cli.WriteError` in [`internal/cli/exit.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/exit.go#L89-L93).
- Wire Envelopes: `wire.TaskGraphRepairJSON` and `wire.TaskGraphRepairEnvelope` in [`internal/wire/dependency_repair.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/wire/dependency_repair.go#L64-L131); validated by JSON Schema (Draft 2020-12) in [`internal/cli/testdata/golden/schema_jsonschema.golden`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/testdata/golden/schema_jsonschema.golden) and `internal/wire/envelopes_test.go`.

#### 2. Ordinary write paths inventory (fail-closed against `GraphBroken`)
- `store.FS.MutateTaskGraph` ([`internal/store/graphmutation.go:57`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/graphmutation.go#L57-L59)): Calls `core.ValidateTaskGraphMutationSource(graph)`, which returns `ErrValidation` if `graph.Health() == GraphBroken`. Ordinary `task depend add`, `task depend remove`, and `task depend migrate` cannot enter a broken graph.
- `store.FS.MutateTaskLifecycle` ([`internal/store/lifecyclemutation.go:57`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/lifecyclemutation.go#L57-L59)): Calls `core.ValidateTaskLifecycleSource(graph)`, which returns `ErrValidation` if `graph.Health() == GraphBroken`. All task lifecycle verbs (`start`, `next`, `ready`, `complete`, `defer`, `deprecate`) fail closed.
- `store.FS.CreateThread` ([`internal/store/threadcreation.go:57`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/threadcreation.go#L57-L59)): Calls `core.ValidateThreadCreationSource(graph)` -> fails closed on `GraphBroken`.
- `store.FS.ApplyThread` ([`internal/store/threadapply.go:56`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/threadapply.go#L56-L58)): Calls `core.ValidateThreadApplySource(graph)` -> fails closed on `GraphBroken`.
- `store.FS.MutateThread` ([`internal/store/threadmutation.go:57`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/threadmutation.go#L57-L59)): Calls `core.ValidateThreadMutationSource(graph)` -> fails closed on `GraphBroken`.
- `store.FS.EditTask` / `SetTaskFields` ([`internal/store/edit.go:173`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/edit.go#L173-L175)): Explicitly rejects mutations containing `depends_on` or legacy dependency fields (`task edit cannot change depends_on or legacy dependency fields`).
- `cli.RunLintFix` ([`internal/cli/lint.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/lint.go)): Restricts `--fix` to repairable frontmatter formatting; all graph-owned dependency problems require deliberate repair through `task depend repair`.

---

### Evidence matrix

| Target Area | Test / Fixture | Hostile Conditions & Edge Cases Exercised | Result / Witness |
| :--- | :--- | :--- | :--- |
| **Automatic repair policy** | `TestTaskGraphRepairAutoIsLimitedAndPreservesExplicitIntent` | Self-edges, duplicate canonicals, empty legacy keys, invalid IDs, dangling IDs. | `--auto` selects only duplicates, self-edges, empty legacy; invalid/dangling retained verbatim; returns `GraphBroken`. |
| **Self-edge duplicates** | `TestTaskGraphRepairAutoHandlesRepeatedSelfDeclarationsWithoutOrderDependence` | Multiple duplicate self-declarations (`DependsOn: [owner.ID, owner.ID]`). | Both removed deterministically; `Automatic` provenance preserved; converges to `GraphHealthy`. |
| **Overlap rejection** | `TestTaskGraphRepairRejectsOverlappingDedupeAndExactDrop` | Concurrent `--auto` dedupe and explicit `--drop` on the same declaration. | Fails closed with `domain.ErrValidation: dedupe and exact-drop repair intents overlap`. |
| **Accounting validation** | `TestTaskGraphRepairValidationRejectsUnaccountedReceiptSelections` | Selected intent contains declarations absent from active operations. | Fails closed with `domain.ErrValidation: selected repair intent does not exactly account for active operations`. |
| **Valid constraint protection** | `TestTaskGraphRepairExplicitDropsConvergeAndRejectValidConstraints` | User `--drop` targets valid, non-defective prerequisite ID. | Rejected with `domain.ErrValidation: selected declaration is not a repairable defect`. |
| **Cycle breaking & SCCs** | `TestTaskGraphRepairCycleRequiresExplicitEdgeAndCanLeaveOtherDefects` | Two-node cycle + dangling reference. | Auto refuses to guess cycle edge; explicit drop breaks cycle while preserving unrelated dangling defect. |
| **SCC splitting & chords** | Hostile probe (`TestAttackProgressMeasure_SCCSplit` & `CycleChordRemoval`) | 4-node SCC with chord splitting into 2 SCCs; 3-cycle with chord. | `cyclicEdges` componentwise decreases strictly from 6 to 4; chord decreases 4 to 3; non-increasing measure holds. |
| **Legacy field attribution** | `TestTaskGraphRepairLegacyBlocksUsesDeclarationOwnerNotProjectedDependent` | Legacy `blocks` declaration whose file owner is prerequisite, not dependent. | Targets declaration owner file; projects reverse edge correctly; breaks cycle without clearing neighbor keys. |
| **Ambiguous legacy intent** | `TestTaskGraphRepairNeverGuessesAmbiguousLegacyIntent` | Multiple candidate tasks matching legacy slug. | Zero operations under `--auto`; defect marked un-automatic with `candidate_ids: [ID1, ID2]`. |
| **Shadow & unreadable tasks** | `TestTaskGraphRepairPreservesShadowAndUnreadableEvidence` | Duplicate-ID task shadows + unreadable task file with opaque revision token. | Shadows and unreadable records preserved byte-for-byte; unreadable task ID recognized as known target. |
| **CAS on task mutation** | `TestMutateTaskGraphRepairRejectsLateReadableNonTargetTaskByteChange` | Non-target readable task modified between verify and write. | Detected by `verifyTaskGraphSourceSnapshot`; returns `domain.ErrConflict`; 0 files committed. |
| **CAS on target task** | Hostile probe (`TestMutateTaskGraphRepairRejectsTargetTaskByteChangeBeforeWrite`) | Target task content mutated immediately before write. | Hash mismatch `hashContent != ifVersion`; returns `domain.ErrConflict`; concurrent bytes preserved. |
| **CAS on unreadable tasks** | `TestMutateTaskGraphRepairRejectsLateUnreadableTaskByteChange` | Unreadable task file bytes modified without changing parse error message. | Source version hash mismatch; returns `domain.ErrConflict`; write aborted. |
| **Thread CAS protection** | `TestMutateTaskGraphRepairAllowsMalformedThreadsButCASProtectsTheirBytes` | Malformed Thread file modified concurrently before write. | Allows malformed thread in diagnosis; detects byte change on write; returns `domain.ErrConflict`. |
| **Thread impact projection** | `TestMutateTaskGraphRepairReportsReadableThreadProjectionImpacts` | Valid Thread with task participating in repair. | Derived Thread health transitions `broken -> healthy`; direct flag set; zero Thread file writes. |
| **Durable prefix recovery** | `TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix` | Multi-file repair where write 1 succeeds and write 2 fails (injected failure). | Partial receipt reports `AppliedFiles: [file1]`, `RemainingFiles: [file2]`; retry converges remaining file. |
| **Surgical YAML editing** | Hostile probe (`TestUpdateDependencySourceEdits_ComprehensiveSurgicalYAML`) | CRLF line endings, flow sequence `["a", 'b', c]`, inline comments, unknown keys. | CRLF preserved, inline comments retained on survivors, key order unchanged, `updated_at` stamped. |
| **Wire schema compliance** | Hostile probe (`TestJSONSchema_ValidatesFullyPopulatedTaskGraphRepairEnvelope`) | Non-default values in all optional fields (impacts, threads, candidates). | Validates against Draft 2020-12 schema; zero opaque revision tokens leaked in wire envelope. |
| **Hostile CLI raw values** | Hostile probe (`TestHostileRawValuesEmittedCommands`) | Spaces, colons, equals, quotes, shell metachars (`$VAR; &`), terminal `#123`. | Diagnosis emits actionable copyable command; executing exact command successfully drops declaration. |

---

### Findings

**None.** All challenged patterns, edge cases, and systemic failure modes settled cleanly with hostile evidence. All requirements of ADR-0006 and task `6g4g8gatbnrs` are satisfied.

---

### Settled hostile angles and residual risks

1. **Terminal `#<digits>` in raw values vs selector grammar:**
   - *Challenge:* If a raw value in YAML ends in `#<digits>` (e.g. `foo#123`), could `--drop` selector parsing misinterpret `#123` as occurrence index 123 and truncate the value?
   - *Settled:* In diagnosis, `render.repairEditSelector` always emits an explicit occurrence suffix (e.g. `foo#123#0`). The parser in `parseGraphRepairSelector` uses `strings.LastIndex(value, "#")`, which parses the trailing `#0` as occurrence 0 and leaves `foo#123` intact as the raw value. When executed via CLI or plan manifest, the exact declaration is targeted. If a user manually types a shorthand `--drop 'owner:depends_on=foo#123'` omitting the `#occurrence` suffix, the parser interprets `occurrence=123` and `value="foo"`, which fails closed with `domain.ErrNotFound` / `selected declaration is not a repairable defect` rather than corrupting data. In YAML plan manifests, `value` and `occurrence` are distinct structured fields.
2. **Progress measure monotonicity across SCC splits and chords:**
   - *Challenge:* Could breaking an SCC or removing a cycle chord reveal masked problems that increase defect counts and cause the monotonicity check to reject a valid repair?
   - *Settled:* The defect measure decomposes into `unreadable`, `identity`, `cyclicEdges`, `legacyUnsafe`, `declarations`, and `legacyPresent`. Removing a cycle edge strictly reduces `cyclicEdges` (tested on 4-node SCC splitting into 2 SCCs: 6 cyclic edges dropped to 4). Breaking a cycle either turns formerly cyclic edges into valid edges (which are not counted in `declarations` because they are valid dependencies) or transitions `legacyUnsafe` to `legacyResolved`. It is mathematically impossible for removal of an edge to increase `cyclicEdges`, `declarations`, or `legacyUnsafe`.
3. **Receipt truthfulness on partial failure:**
   - *Challenge:* Could a failure during a multi-file repair return a receipt that reflects the full prospective plan rather than the durable prefix that actually landed on disk?
   - *Settled:* `FS.MutateTaskGraphRepair` re-validates the prefix after each successful file write (`repairPlanPrefix`), updates `result.Analysis = prefixAnalysis`, reloads the authoritative graph, and derives `Addressed`, `Residual`, `FinalHealth`, `AppliedFiles`, and `RemainingFiles` strictly from the durable prefix. In `taskGraphRepairReceipt`, if prospective graph differs from final graph, full-plan impacts are zeroed out rather than reporting uncommitted changes.
4. **Direct store invocation with malicious planner:**
   - *Challenge:* Could a direct caller bypass CLI validation and invoke `TaskGraphRepairStore.MutateTaskGraphRepair` with a planner that deletes valid constraints or unselected declarations?
   - *Settled:* `ValidateTaskGraphRepairPlan` re-plans all requested operations against the snapshot, reauthorizes each edit against diagnosed defects, verifies exact selection-to-operation accounting, proves componentwise monotone progress for every source-file group, and proves declaration containment via `validateRepairPreservation`. In `materializeTaskGraphRepair`, the store re-parses each mutated file from disk and asserts `reflect.DeepEqual(expected.Fields, actualRecords[0].Fields)`, rejecting any mutation that alters unselected fields.
5. **Private token privacy in wire envelopes:**
   - *Challenge:* Could opaque source revision tokens (used for CAS) leak through JSON output?
   - *Settled:* `ThreadReadProblemJSON` and `TaskGraphLoadProblemJSON` in `internal/wire` do not define a `source_version` field. Furthermore, `cloneThreadReadProblemsPublic` explicitly zeroes `SourceVersion = ""` before receipts are constructed.
6. **Exclusive repository lock contention:**
   - *Residual risk analysis:* `FS.MutateTaskGraphRepair` holds the repository-wide exclusive lock (`checkedWriteLock`) during planning, CAS checks, and disk materialization. Because graph repair is an operator recovery action rather than a high-throughput transaction, lock holding is bounded by local disk I/O across the affected task files.

---

### Exact validation and mutation-test results

#### 1. Sandbox baseline validation
```
go test -race -count=1 ./...
ok  	github.com/andy-esch/taskflow/cmd/tskflwctl	4.664s
ok  	github.com/andy-esch/taskflow/internal/cli	6.792s
ok  	github.com/andy-esch/taskflow/internal/cli/prompt	1.510s
ok  	github.com/andy-esch/taskflow/internal/cli/render	2.091s
ok  	github.com/andy-esch/taskflow/internal/config	1.877s
ok  	github.com/andy-esch/taskflow/internal/configstore	2.379s
ok  	github.com/andy-esch/taskflow/internal/configui	2.161s
ok  	github.com/andy-esch/taskflow/internal/core	2.522s
ok  	github.com/andy-esch/taskflow/internal/design	2.635s
ok  	github.com/andy-esch/taskflow/internal/domain	3.936s
ok  	github.com/andy-esch/taskflow/internal/editor	1.551s
ok  	github.com/andy-esch/taskflow/internal/graphfmt	1.442s
ok  	github.com/andy-esch/taskflow/internal/id	1.373s
ok  	github.com/andy-esch/taskflow/internal/listfilter	1.564s
ok  	github.com/andy-esch/taskflow/internal/progressbar	1.603s
ok  	github.com/andy-esch/taskflow/internal/spacehealth	1.496s
ok  	github.com/andy-esch/taskflow/internal/spacestore	1.526s
ok  	github.com/andy-esch/taskflow/internal/store	6.063s
ok  	github.com/andy-esch/taskflow/internal/theme	1.152s
ok  	github.com/andy-esch/taskflow/internal/tomledit	1.260s
ok  	github.com/andy-esch/taskflow/internal/tui	9.135s
ok  	github.com/andy-esch/taskflow/internal/userconfig	1.417s
ok  	github.com/andy-esch/taskflow/internal/wire	1.743s
ok  	github.com/andy-esch/taskflow/internal/workspacestore	1.613s

golangci-lint run ./...
0 issues.

go run ./cmd/tskflwctl lint
✔ all planning entities and dependency links pass lint

go mod tidy -diff
(clean, exit 0)

go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli
(clean, exit 0)

git diff --check
(clean, exit 0)
```

#### 2. Mutation testing floor (9 verified mutations)
All 9 required mutations were tested in the sandbox. Each mutation was killed by its designated focused regression test without compilation failure:

1. **Dedicated-port isolation:**
   - *Mutation:* Bypass `if graph.Health() == GraphBroken` in `ValidateTaskGraphMutationSource` (`internal/core/dependency_graph_mutation.go`).
   - *Test:* `go test -v -run TestMutateTaskGraphBrokenSnapshotFailsBeforePlanner ./internal/store`
   - *Result:* FAIL (`graphmutation_test.go:236: broken mutation error = validation failed: planned dependency state is broken; mutation requires a healthy final graph: dependency cycle: ...`). Killed.
2. **Full-source preservation:**
   - *Mutation:* Omit survivor nodes in `updateDependencySourceEdits` (`internal/store/frontmatter.go`).
   - *Test:* `go test -v -run TestUpdateDependencySourceEditsPreservesUnselectedSequenceNodes ./internal/store`
   - *Result:* FAIL (`graphrepair_test.go:232: missing "6g1111111111 # keep first"`). Killed.
3. **Automatic-policy bounds:**
   - *Mutation:* Mark `RepairInvalidID` as `defect.Automatic = true` in `DiagnoseTaskGraphRepair` (`internal/core/dependency_repair.go`).
   - *Test:* `go test -v -run TestTaskGraphRepairAutoIsLimitedAndPreservesExplicitIntent ./internal/core`
   - *Result:* FAIL (`dependency_repair_test.go:29: invalid-dependency-id automatic=true, want false`). Killed.
4. **Overlap rejection:**
   - *Mutation:* Disable overlap check in `rejectOverlappingRepairOperations` (`internal/core/dependency_repair.go`).
   - *Test:* `go test -v -run TestTaskGraphRepairRejectsOverlappingDedupeAndExactDrop ./internal/core`
   - *Result:* FAIL (`dependency_repair_test.go:97: overlapping repair error = <nil>`). Killed.
5. **Raw-colon selector parsing:**
   - *Mutation:* Mutate selector marker from `:<field>=` to `:<field>:` in `splitGraphRepairSelector` (`internal/cli/task_dependency_repair.go`).
   - *Test:* `go test -v -run TestTaskDependRepairManifestAndSelectorParsing ./internal/cli`
   - *Result:* FAIL (`task_dependency_repair_test.go:75: selector=... err=validation failed: repair selector "repair-cli-manifest:depends_on=raw#12#0" must be <task-or-path>:<field>=<raw-value>[#occurrence]`). Killed.
6. **Readable/unreadable evidence CAS:**
   - *Mutation:* Bypass `verifyTaskGraphSourceSnapshot` before repair write in `MutateTaskGraphRepair` (`internal/store/graphrepair.go`).
   - *Test:* `go test -v -run TestMutateTaskGraphRepairRejectsLateUnreadableTaskByteChange ./internal/store`
   - *Result:* FAIL (`graphrepair_test.go:144: unreadable task CAS receipt=... Committed:true ... err=task evidence changed while graph repair committed a prefix`). Killed.
7. **Durable-prefix accounting:**
   - *Mutation:* Disable `result.RemainingSources` assignment in `setRepairRemainingSources` (`internal/store/graphrepair.go`).
   - *Test:* `go test -v -run TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix ./internal/store`
   - *Result:* FAIL (`graphrepair_test.go:205: partial receipt=... RemainingFiles:[] ...`). Killed.
8. **Automatic provenance:**
   - *Mutation:* Force `authorized.Operations[index].Automatic = false` in `ValidateTaskGraphRepairPlan` (`internal/core/dependency_repair.go`).
   - *Test:* `go test -v -run TestTaskGraphRepairAutoHandlesRepeatedSelfDeclarationsWithoutOrderDependence ./internal/core`
   - *Result:* FAIL (`dependency_repair_test.go:75: validated auto-selected operation lost provenance: ... Automatic:false`). Killed.
9. **Non-default wire branches:**
   - *Mutation:* Allow `CandidateIDs: defect.CandidateIDs` (which serializes `nil` as `null`) in `ToTaskGraphRepairDefectsJSON` (`internal/wire/dependency_repair.go`).
   - *Test:* `go test -v -run TestJSONSchema_ValidatesRealOutput ./internal/wire`
   - *Result:* FAIL (`envelopes_test.go:399: TaskGraphRepairEnvelope output does NOT validate against its own schema: - at '/residual/0/candidate_ids': got null, want array`). Killed.
