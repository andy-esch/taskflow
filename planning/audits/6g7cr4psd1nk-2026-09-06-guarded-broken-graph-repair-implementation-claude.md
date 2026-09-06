---
schema: 1
id: 6g7cr4psd1nk
bucket: closed
area: guarded-broken-graph-repair-implementation-claude
date: "2026-09-06"
updated_at: "2026-09-06"
---
# Audit: Guarded broken-graph repair implementation — claude — 2026-09-06

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

### Isolation attestation

- Sandbox path: `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BBlm54`
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BBlm54/.git`
  (verified with `git rev-parse --absolute-git-dir` inside the sandbox; not a worktree, not a symlink,
  no `.git` pointer back to the shared checkout)
- Sandbox baseline commit: `8a685f1efa6b16564ede096821f2771690befd4f`
  (`chore: capture review sandbox baseline`, the only commit this reviewer created)
- Captured source-audit blob: `b07038854d2a3f158f369598bc127c005a0de9b3`
- Branch / base under review: `feat/guarded-broken-graph-repair` on top of `58e99d9`
- Guarded transfer: **succeeded** — the origin-hash guard matched at transfer time and only
  `planning/audits/6g7cr4psd1nk-2026-09-06-guarded-broken-graph-repair-implementation-claude.md`
  was copied back. Every build, test, generator, fixture, mutation probe and throwaway planning
  space lived in `$SANDBOX`; all probes were removed and `git diff <baseline>` in the sandbox is
  empty apart from this audit.

### Verdict

The guarded materialization half of this work is strong. Authorization, CAS layering, surgical YAML
editing, durable-prefix receipts and the wire contract survived direct hostile attack: 7 of 8
malicious planner plans were rejected, the 8th (a lying `Reason`) was silently renormalized to the
truthful reason rather than smuggled into the receipt, and a fully populated non-default repair
envelope validates against the published schema with no revision-token leak. I found no unauthorized
declaration removal, no stale write, no false `Committed`, and no partial receipt that misdescribes
its durable prefix.

The weakness is on the other side of the boundary: **the recovery path can be denied outright by
defects the task explicitly promises to handle, and two of its three headline proofs are not
actually enforced where they are claimed.** `--auto` is all-or-nothing, so a single declaration the
measure cannot score (H1) or the materializer cannot express (M2) makes the whole capability
unusable repository-wide, while bare diagnosis keeps recommending it. The core
declaration-containment "proof" is a tautology over its own inputs (M1). Four load-bearing guards,
including the automatic-provenance bound, have no test that fails when they are deleted (M3).

Seven findings, all left `open` for implementation-owner triage.

### Repository-wide consumer inventory

| Symbol / capability | Non-test consumers | Test consumers |
| --- | --- | --- |
| `core.TaskGraphRepairStore` | `core/service.go:21,113,209` (field, `WithTaskGraphRepairStore`, `NewService` discovery); `core/store.go:104`; `store/graphrepair.go:16` (`var _` assertion); `docs/ARCHITECTURE.md:201,298` | none directly |
| `MutateTaskGraphRepair` | `core/store.go:105` (port); `core/dependency_repair.go:550` (sole caller); `store/graphrepair.go:21` (sole impl) | `store/graphrepair_test.go` (via `Service`) |
| `Service.RepairTaskGraph` | `cli/task_dependency_repair.go:87` (sole caller) | `store/graphrepair_test.go` |
| `Service.InspectTaskGraphRepair` | `cli/task_dependency_repair.go:57` (sole caller) | none — **no direct test** |
| `ValidateTaskGraphRepairPlan` | `store/graphrepair.go:64` (full plan), `:82` (empty plan → "unchanged repository" analysis), `:135` (per-write prefix) | `core/dependency_repair_test.go` |
| `PlanTaskGraphRepair` | `core/dependency_repair.go:383,396` (self-reauthorization), `:551` (planner callback) | `core/dependency_repair_test.go` |
| `DiagnoseTaskGraphRepair` | 9 call sites in `core/dependency_repair.go` (diagnosis, planning, validation, measure, receipt) | `core/dependency_repair_test.go` |
| `SimulateSourceEdits` | `core/dependency_repair.go:325` (empty-field pre-pass), `:433` (per-group progress) | `core/dependency_source_test.go` |
| `SameRepairSnapshot` | `store/graphrepair.go:148` (post-write reload) | none — **no direct test** |
| `SameSourceSnapshot` | `store/cas.go:37` (`verifyTaskGraphSourceSnapshot`, shared with ordinary graph mutation) | `core/dependency_{source,graph}_test.go`, `store/graphmutation_test.go` |
| Repository locking | `store/lock.go` `checkedWriteLock` / `enterRepositoryPlanner` / `rejectRepositoryPlannerCall`; taken by graphrepair, graphmutation, threadapply, threadmutation, threadcreation, lifecyclemutation | `store` package |
| `TaskGraphThreadImpacts` | `core/dependency_repair.go:592` (sole caller) | none — **no direct test** |
| Graph-health guidance | `core/dependency_graph_mutation.go:138` `taskGraphHealthDetail` → `board.go:79`, `service.go:394` (`status`), `service_task.go:57`, `task_lifecycle.go:196`, `dependency_graph_mutation.go:20,95,103`; `core/service.go:620,648` (lint); `cli/render/dependency.go:319` | `cli/lint_test.go`, `cli/status_test.go` |
| Error classification | `cli/exit.go:89` (`graphRepairCommandFailure` → `ErrorItem.GraphRepair`); sentinels `ErrValidation`/`ErrNotFound`/`ErrAmbiguous`/`ErrConflict` → exits 11/10/13/14 | `cli/task_dependency_repair_test.go` |
| Wire envelope | `wire/dependency_repair.go`; `wire/envelopes.go:1044,1149`; `cli/render/dependency.go:11` | `wire/envelopes_test.go`, golden `task_depend_repair_json.golden` |
| `updateDependencySourceEdits` | `store/graphrepair.go:194` (sole caller) | `store/graphrepair_test.go` |

**Ordinary write paths that must not enter a broken graph** (all re-verified empirically against a
2-node cycle + invalid-ID repository — every one refused and no task byte changed):
`task depend add` (exit 11), `task depend remove` (11), `task depend migrate` (11),
`task start`/lifecycle (11), `thread new`/Thread mutation (11), `lint --fix` ("nothing to fix", 11),
`task set` (succeeds but its own help documents that graph-owned fields "cannot be changed or
removed here, including with `--force`"), `task edit`/`task append` (body/frontmatter only, no
graph-owned vocabulary). Deleting the `GraphBroken` guard in `dependency_graph_mutation.go:18` is
killed by `TestMutateTaskGraphBrokenSnapshotFailsBeforePlanner`,
`TestValidateTaskGraphMutationSourceNamesGuardedRepairPath` and
`TestPlannerWindowSetupFailureReturnsInsteadOfHanging`. **Dedicated-port isolation holds.**

### Evidence matrix

| Brief requirement | Method | Result |
| --- | --- | --- |
| 1. Full uncached race suite | `go test -count=1 -race ./...` | all packages `ok` |
| 1. golangci-lint | `golangci-lint run ./...` | `0 issues.` |
| 1. Planning lint | `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| 1. Module tidiness | `go mod tidy` + `git diff go.mod go.sum` | clean |
| 1. Generated docs | `just docs` + `git diff --exit-code docs/cli` | stable |
| 1. Schema comments | `go run ./internal/tools/schemacomments` | stable (no diff) |
| 1. `git diff --check` | `git diff --check 58e99d9 HEAD` | one hit, see "Settled" §1 |
| 2. Hostile repositories | 8 throwaway spaces: self, repeated self, 2-node and 4-node cycles, chord/SCC split, duplicate canonical, invalid IDs, dangling IDs, all three legacy fields, empty legacy keys, duplicate-ID shadow, YAML anchors, 0600 file mode | H1, M2 |
| 3. Execute every emitted command | 11 emitted `--drop` commands run verbatim against values with spaces, `"`, `'`, `:`, `=`, `;`, `$( )`, `\`, `#3`, and an embedded `:depends_on=` | **all 11 correct**, exactly one declaration each |
| 4. Attack the progress measure | shadow self-edge; chord removal; SCC split; multi-defect files; residual-broken repairs | **H1** |
| 5. Operation accounting | 8 hostile plans direct to `FS.MutateTaskGraphRepair` | 7 rejected, 1 renormalized — see "Settled" §2 |
| 6. Lock/CAS timing | existing hooks + new probes at verify / pre-write / post-write; readable, unreadable, target, non-target, Thread bytes | no false commit, no clobber |
| 7. Surgical YAML | flow/block, comments on kept+removed nodes, CRLF, quoting mix, key order, unknown fields, empty-key removal, body bytes, `updated_at` replace, file mode, anchors | 10/11 exact; anchors → **M2** |
| 8. Wire branches at non-default values | fully populated envelope + schema validation + leak check | validates, no leak, all collections non-null |
| 9. Thread projection | prefix probe with a 2-task Thread; malformed-Thread fixtures | prefix-scoped and truthful |
| 10. Mutation-kill each claimed test | 8 targeted + 1 coordinated mutation | 4 survive `go test ./...` → **M1, M3** |

---

### Findings

#### H1. A self-declaration owned by a duplicate-ID shadow is advertised as auto-repairable, is unrepairable by every selector, and permanently denies the entire `--auto` plan · **Status:** fixed 2026-09-06

`measureTaskGraphRepair` (`internal/core/dependency_repair.go:757-798`) gives `RepairSelf` no
component of its own. A self-declaration can only move the measure through `cyclicEdges`, which is
incremented only for declarations with `HasProjectedEdge`. `projectedSourceEdge`
(`internal/core/dependency_source.go:409-411`) refuses a projected edge to any record that is not
`isRepresentativeSource`, and a duplicate-ID shadow never is. `DiagnoseTaskGraphRepair`
(`:206-208`) nonetheless diagnoses the shadow's self-declaration purely from
`declaration.Value == declaration.Source.TaskID`, marking it `Repairable: true, Automatic: true`.

The group therefore simulates to a measure that is componentwise *equal*, and the strict-improvement
gate at `:441` rejects it — and with it the whole plan.

**Reproduction** (sandbox space, three files):

```
tasks/taaaaaaaaaa1-primary.md   id: taaaaaaaaaa1   depends_on: [taaaaaaaaaa2]
tasks/taaaaaaaaaa1-shadow.md    id: taaaaaaaaaa1   depends_on: [taaaaaaaaaa1]   # shadow, self-edge
tasks/taaaaaaaaaa2-other.md     id: taaaaaaaaaa2   depends_on: [taaaaaaaaaa3, taaaaaaaaaa3]
tasks/taaaaaaaaaa3-third.md     id: taaaaaaaaaa3
```

```
$ tskflwctl task depend repair
inferable: 2 inferable repair(s); apply with `tskflwctl task depend repair --auto`
• self-dependency        …/taaaaaaaaaa1-shadow.md:depends_on="taaaaaaaaaa1"#0
• duplicate-dependency   …/taaaaaaaaaa2-other.md:depends_on="taaaaaaaaaa3"#0

$ tskflwctl task depend repair --auto                                      # exit 11
error: validation failed: repair of …/taaaaaaaaaa1-shadow.md does not strictly
       improve its durable source-file prefix

$ tskflwctl task depend repair --drop '<abs>/taaaaaaaaaa1-shadow.md:depends_on=taaaaaaaaaa1#0'
error: validation failed: repair of …/taaaaaaaaaa1-shadow.md does not strictly   # exit 11
       improve its durable source-file prefix
```

The declaration is reachable by no repair path (`task depend remove` and `lint --fix` are correctly
fail-closed, and `task set` refuses graph-owned fields), so the defect is permanent. Running the
unrelated `--dedupe` on `taaaaaaaaaa2-other.md` alone succeeds (exit 0), which isolates the cause:
any plan *containing* the shadow operation dies, and `--auto` always contains it. `--auto` is
therefore dead for the whole repository until the shadow file is hand-edited — in a tool whose
premise is that hand-editing is what broke the graph.

- **Violated contract.** Task scope: "`--auto` may only deduplicate canonical values, remove
  self-edges, and clear present-but-empty legacy keys." AC: "Cycle, self-edge, dangling-reference,
  invalid-ID, duplicate-edge, and each legacy-field fixture have an actionable preview and converge
  to the selected repaired state, **whether or not unrelated residual problems leave the repository
  broken**." AC: "Validation runs over the full source-level projection and preserves duplicate-ID
  shadow records." Brief: "Bare diagnosis … `--auto` selects only canonical dedupe, self-edge, and
  empty-legacy cleanup."
- **Smallest sound recommendation.** Add a source-level `selfDeclarations` component to
  `taskGraphRepairMeasure`, counted from diagnosed `RepairSelf` defects rather than from projected
  edges, and include it in `repairMeasureComponentwiseNonIncreasing`. This keeps the measure
  source-level (consistent with the rest of the design) and needs no change to the diagnosis, the
  planner, or the wire contract. Optionally also make the strict-improvement failure name the
  operation that could not be scored instead of only the file.
- **Affected consumers.** `store/graphrepair.go:64` and `:135` (both `ValidateTaskGraphRepairPlan`
  call sites — so dry-run, apply *and* durable-prefix reconstruction all fail); `Service.RepairTaskGraph`;
  `task depend repair --auto` and every explicit selector naming such a declaration; any future
  `TaskGraphRepairStore` adapter.
- **Missing regression test.** A `--auto` fixture whose repository contains a duplicate-ID shadow
  that owns an auto-safe declaration, asserting the plan validates and converges while the
  `duplicate-task-id` problem correctly remains residual. Mutation that must kill it: drop the new
  measure component.

**Resolution:** Added a source-level self-declaration component to the monotone
progress measure and a duplicate-ID shadow regression that leaves only the
identity defect residual.

#### M1. The declaration-containment "proof" never inspects the prospective graph, so it cannot reject unauthorized removal · **Status:** fixed 2026-09-06

`validateRepairPreservation` (`internal/core/dependency_repair.go:857-905`) builds `allowed` from
`plan.Operations` × `before.SourceDeclarations()`, and its `removed` argument comes from
`selectedRepairDeclarations(graph, plan.Operations)` (`:824-853`) — the same operations against the
same `before` declarations, by near-identical matching code. The two multisets cancel by
construction. The only statement that touches `after` is a count comparison of `sourceTasks` and
`loadProblems`. **No code path compares `before.SourceDeclarations()` with
`after.SourceDeclarations()`.**

**Reproduction.** Graph: `owner.depends_on = [prereq, other, owner]`; the self-edge is the only
authorized removal. Single-line mutation to `applySourceDeclarationEdits`
(`internal/core/dependency_source.go`) so the simulator drops one extra value:

```go
if len(remaining) > 1 {
    remaining = remaining[:len(remaining)-1] // unauthorized over-removal
}
```

`PlanTaskGraphRepair` → `ValidateTaskGraphRepairPlan` returns **`err == nil`**, and
`analysis.Prospective` holds `depends_on = [prereq]` — a valid constraint (`other`) has silently
disappeared, while `analysis.Removed` still lists only the self-edge. A receipt built from that
analysis would report one authorized removal while two declarations vanished.

Coordinating the mutation across both removal implementations (`applySourceDeclarationEdits` **and**
`store.updateDependencySourceEdits`) reaches `go test ./...` and is caught only by fixtures that
assert expected *values* (`TestTaskGraphRepairAutoIsLimitedAndPreservesExplicitIntent`,
`TestTaskGraphSourceSimulationPreservesUntouchedInvalidLiteralsVerbatim`,
`TestUpdateDependencySourceEditsPreservesUnselectedSequenceNodes`) — never by a guard rejecting the
plan. In the shipped FS adapter the real containment check is
`materializeTaskGraphRepair`'s `reflect.DeepEqual(expected.Fields, actualRecords[0].Fields)`
(`internal/store/graphrepair.go:206-209`), which is a *cross-check of two independent removal
implementations*, not a containment proof, and is private to the filesystem adapter.

- **Violated contract.** Task scope: "an explicit declaration-containment/minimality proof must
  reject unrelated constraint deletion." AC: "no unrelated declaration or valid constraint
  disappears." Brief: "The plan validator … rejects valid-constraint deletion." Brief: "Do not infer
  architectural isolation merely from separate type names."
- **Smallest sound recommendation.** In `validateRepairPreservation`, compute the multiset difference
  between `before.SourceDeclarations()` and `after.SourceDeclarations()` (keyed by
  `sourceDeclarationSemanticKey` with occurrence renumbering handled the way `selectedRepairDeclarations`
  already does) and require it to equal `removed` exactly. That turns the existing tautology into the
  proof it is documented to be, in the layer that owns it, with no adapter changes.
- **Affected consumers.** Any implementer of `core.TaskGraphRepairStore` other than `store.FS` — a
  remote/TUI/web adapter that calls `ValidateTaskGraphRepairPlan` and applies `analysis.Prospective`
  gets no containment guarantee at all. Also `store/graphrepair.go:135`, where the per-write prefix
  analysis is trusted to describe the durable prefix.
- **Missing regression test.** A core test that constructs a prospective graph missing an
  unauthorized declaration and requires `ValidateTaskGraphRepairPlan` to return `ErrValidation`; it
  must fail under the one-line `applySourceDeclarationEdits` mutation above, not merely report
  different values.

**Resolution:** Core now independently compares complete before/after source
records, field presence, value counts, and survivor order; a forged prospective
graph that drops an unrelated valid constraint is rejected.

#### M2. A YAML alias in a graph-owned sequence is invisible to the materializer, so one aliased declaration fails the whole plan · **Status:** fixed 2026-09-06

`updateDependencySourceEdits` (`internal/store/frontmatter.go:207-231`) matches sequence entries by
`item.Value`. For a `yaml.AliasNode` that is the **anchor name**, not the resolved scalar, while the
semantic parser that feeds `core` resolves the alias normally. Core and the materializer therefore
disagree about what the sequence contains.

**Reproduction** (verified at both the unit and CLI level):

```yaml
ref: &A taaaaaaaaaa2
depends_on: [*A, taaaaaaaaaa2]
```

Core sees `["taaaaaaaaaa2", "taaaaaaaaaa2"]` and diagnoses an automatic `duplicate-dependency`.
The materializer sees node values `["A", "taaaaaaaaaa2"]`, finds no duplicate, returns
`changed == false`, and `materializeTaskGraphRepair` (`internal/store/graphrepair.go:198-200`) fails
the run:

```
$ tskflwctl task depend repair
inferable: 2 inferable repair(s); apply with `tskflwctl task depend repair --auto`

$ tskflwctl task depend repair --auto                                       # exit 11
error: validation failed: validated graph repair for …/taaaaaaaaaa1-anchor.md
       produced no source change
```

The unrelated, perfectly repairable self-edge in a different file is not repaired either. There is
a second, currently-latent half: for an exact `--drop`, core's occurrence index counts *semantic*
values while the store's counts *node* values, so the two select different physical nodes whenever
an alias precedes an equal literal. Today `materializeTaskGraphRepair`'s `DeepEqual` keeps the
*semantic* outcome correct (both removals resolve to the same value), so this is a divergence, not
corruption — but it is one refactor away from being one.

- **Violated contract.** Task scope: "Preserve … surgical frontmatter updates." Brief: "Verify
  surgical YAML behavior for flow and block sequences … quoted/scalar edge cases … and representation
  transitions"; "Emitted commands must parse and act on the defect they display."
- **Smallest sound recommendation.** In `updateDependencySourceEdits`, resolve the comparison value
  through `item.Alias` when `item.Kind == yaml.AliasNode` (`raw := item.Value; if item.Alias != nil
  { raw = item.Alias.Value }`). Removal still deletes the alias node itself, which is the correct
  surgical edit. Keep the existing `!changed` guard as the backstop.
- **Affected consumers.** `store/graphrepair.go:194` (sole caller) and therefore every `--auto`,
  `--drop`, `--dedupe` and `--plan` invocation in a repository that uses YAML anchors anywhere in a
  graph-owned sequence.
- **Missing regression test.** A `updateDependencySourceEdits` case with `depends_on: [*A, value]`
  asserting the dedupe removes exactly one node and `changed == true`; plus a store-level fixture
  asserting `--auto` converges. Mutation that must kill it: revert the alias resolution.

**Resolution:** The YAML materializer matches aliases by their resolved scalar
while removing only the selected alias node; focused YAML and filesystem repair
regressions cover this path.

#### M3. Four load-bearing repair guards have no test that fails when they are deleted · **Status:** fixed 2026-09-06

Each mutation below was applied alone to the sandbox source and `go test -count=1 ./...` was run in
full. All four **survive the entire suite**:

| # | Mutation | Guard it disables |
| --- | --- | --- |
| a | `automaticRepairReason` += `RepairInvalidID` (`dependency_repair.go:479`) | automatic-provenance bound |
| b | `if operation.Automatic && !automaticRepairReason(…)` → `if false` (`:412`) | same bound, directly |
| c | `if !sameRepairOperationSet(authorized.Operations, selectionPlan.Operations)` → `if false` (`:400`) | selected-intent accounting |
| d | `strings.Index(selector, marker) … index < best` → `strings.LastIndex … index > best` (`cli/task_dependency_repair.go:157`) | leftmost-field-marker selector rule |

All four guards are **live, not dead code** — I built the discriminating inputs and confirmed each
probe passes unmutated and fails under its mutation:

- (a)/(b): a planner claiming `Automatic: true` on an `invalid-dependency-id` operation. Rejected
  today with "is not safe for automatic selection"; accepted under the mutation, so a false
  `"automatic": true` reaches the JSON receipt for a destructive, information-losing removal.
  `TestTaskGraphRepairAutoIsLimitedAndPreservesExplicitIntent` does not cover this because auto
  *selection* keys off `defect.Automatic` from diagnosis, while `automaticRepairReason` is used only
  by the validator's provenance check.
- (c): `Operations = [drop "invalid-token"#0]` with `Selections = [drop <dangling-id>#0]` — a
  selection that plans cleanly on its own but misdescribes the active operations. Rejected today with
  "selected repair intent does not exactly account for the active operations"; accepted under the
  mutation. `TestTaskGraphRepairValidationRejectsUnaccountedReceiptSelections` passes under the
  mutation because its selection (a *valid constraint*) is already rejected by the earlier
  `selectionPlan` guard, so it never exercises `sameRepairOperationSet`.
- (d): `…/a.md:depends_on=a:depends_on=b#0`. Parses today to
  `location=…/a.md, value="a:depends_on=b"`; under the mutation to
  `location="…/a.md:depends_on=a", value="b"`.
  `TestTaskDependRepairManifestAndSelectorParsing` uses `not:a:stable:id`, which contains bare colons
  but no embedded field marker, so it cannot distinguish leftmost from rightmost matching.

- **Violated contract.** Task implementation note: "The validator also rejects mixed dedupe/exact-drop
  effects and reauthorizes selected receipt intent so metadata cannot outrun the guarded operations.
  Copyable selectors retain raw colon-bearing and terminal-`#<digits>` values." Brief: "Inspect tests
  for … assertions that check only counts/health"; "For each new regression test, perform the exact
  sandbox-only source mutation that reintroduces the claimed bug and require that specific test to
  fail. At minimum mutate … automatic-policy bounds … raw-colon selector parsing … automatic
  provenance."
- **Smallest sound recommendation.** Add the four discriminating cases above as focused tests
  (two in `core/dependency_repair_test.go`, one in `cli/task_dependency_repair_test.go`); no
  production change is required.
- **Affected consumers.** The wire receipt's `operations[].automatic` and `selected[]` fields
  consumed by agents; every `--drop`/`--dedupe` selector.
- **Missing regression test.** Exactly the four above.

**Resolution:** Added discriminating tests for false automatic provenance,
selection/operation mismatch, and the leftmost embedded field marker; invalid
action/field/value shapes now also fail closed in core.

#### M4. `task depend repair` holds the exclusive repository write lock for minutes on a few-hundred-task repository · **Status:** fixed 2026-09-06

`MutateTaskGraphRepair` takes `checkedWriteLock()` at `internal/store/graphrepair.go:33` and holds
it through the entire write loop. Inside the per-file loop it performs, for **every** durable write:
two full `core.LoadTaskGraph(s)` (`:107`, `:143`), two full `s.ReadThreads()` (`:102`, `:139`), and a
full `core.ValidateTaskGraphRepairPlan(graph, prefix)` (`:135`) which re-runs `PlanTaskGraphRepair`
(three `DiagnoseTaskGraphRepair` passes) and re-simulates **every group in the growing prefix**. The
prefix work alone is Σᵢ O(i) whole-graph rebuilds, i.e. quadratic, on top of O(N) whole-repository
reloads.

Measured on fresh sandbox spaces, each task carrying two auto-safe self-declarations
(`tskflwctl task depend repair --auto`, wall clock):

| tasks | files/s | wall time |
| --- | --- | --- |
| 25 | 58.5 | 0.43 s |
| 50 | 33.8 | 1.48 s |
| 100 | 15.8 | 6.33 s |
| 150 | 8.3 | 18.08 s |
| 200 | 5.3 | 37.54 s |
| 400 | 1.4 | **291.89 s** |

Growth from 100→400 tasks is ≈46× for 4× the input (≈N^2.8). Bare diagnosis over the same 400-task
repository is **0.09 s**, so the apply path costs ~3000× the read it is built on. This repository's
own `planning/tasks/` holds **324 files** today.

For the whole of that window every other `tskflwctl` mutation in the repository blocks on the
repository guard, including in a second process. The exposure is worst exactly where the feature is
meant to help: a legacy-era or mass-hand-edited repository is both the one with many defects and the
one where recovery must complete.

- **Violated contract.** Brief, second pass: "including performance under the exclusive lock and
  coupling to the growing owned DAG algorithm."
- **Smallest sound recommendation.** Reuse the graph already reloaded at the end of iteration *i* as
  the pre-write snapshot of iteration *i+1* (`expectedGraph` is already carried; the extra
  `LoadTaskGraph`/`ReadThreads` at `:102`/`:107` duplicate the post-write reload from the previous
  iteration), and reconstruct the prefix analysis incrementally from the previous prefix instead of
  re-validating the whole prefix from `graph` each time. Both preserve the CAS semantics — the
  post-write reload is already a fresh read under the same held lock.
- **Affected consumers.** Every concurrent `tskflwctl` writer in the same planning root; `--auto` and
  large `--plan` manifests most of all.
- **Missing regression test.** A bounded-work assertion (e.g. counting `LoadTaskGraph` calls via a
  counting `TaskGraphSource`) proving the number of whole-repository reloads and prefix
  revalidations is O(files), not O(files²).

**Resolution:** Per-write proof now validates one source group and composes it
into the durable prefix instead of revalidating the growing prefix. Fresh task
and Thread reads remain for out-of-band editor safety. The retained benchmark
measured 100 files at 1.76s and 200 at 5.46s versus the audit baseline of 6.33s
and 37.54s.

#### L1. `Changed` is derived from planned writes, not materialized ones · **Status:** fixed 2026-09-06

`internal/store/graphrepair.go:71` sets `result.Changed = len(writes) > 0` **before** any write is
attempted, and the pre-write conflict paths return without clearing it.

**Reproduction** (store-level probe, malformed Thread bytes changed at
`testHookBeforeGraphRepairWrite`, `WithRetry(0, …)`):

```
err              = thread evidence changed before graph repair replacement; retry: conflict
bytes changed?   = false
receipt.Changed  = true          <-- claims a change happened
receipt.Committed= false
AppliedFiles     = []
RemainingFiles   = [<the planned file>]
JSON             = {"changed":true,"dry_run":false,"committed":false,…,"removed":[],"applied_files":[],…}
```

Every other field is truthful; only `changed` is not. Reachability is limited: the CLI discards the
receipt when `len(receipt.AppliedFiles) == 0` (`cli/task_dependency_repair.go:88-92`), so this never
reaches `task depend repair` output today. It is, however, the contract of the public
`core.Service.RepairTaskGraph`, and the first non-CLI adapter to render a failed receipt inherits it.

- **Violated contract.** AC: "Human and JSON receipts derive `Changed` from **actual materialized
  writes** and report `Committed`."
- **Smallest sound recommendation.** Keep `Changed = len(writes) > 0` for `dryRun` (where "would
  change" is the intended meaning) and otherwise set it from `len(result.AppliedSources) > 0`
  alongside `result.Committed`.
- **Affected consumers.** `core.Service.RepairTaskGraph`, `wire.ToTaskGraphRepairJSON` `changed`,
  `render.TaskGraphRepairHuman`'s `changed=%t` header and its "selected repair intent is already
  satisfied" branch.
- **Missing regression test.** Extend `TestMutateTaskGraphRepairAllowsMalformedThreadsButCASProtects‑
  TheirBytes`, which already asserts `Committed` and `AppliedFiles`, to also assert `!receipt.Changed`.

**Resolution:** Non-dry-run Changed now becomes true only after a source
replacement succeeds; the pre-write malformed-Thread conflict regression asserts
Changed and Committed both remain false.

#### L2. The command's own documented `--drop` example does not work · **Status:** fixed 2026-09-06

`internal/cli/task_dependency_repair.go:52` ships

```
tskflwctl task depend repair --drop 'planning/tasks/ID-task.md:depends_on=RAW#0'
```

and `docs/cli/tskflwctl_task_depend_repair.md:18` regenerates it verbatim.
`normalizeGraphRepairLocation` (`:106-111`) joins a relative location onto `app.Cfg.Root`, which is
the **planning root** (`<repo>/planning`), so the example resolves to
`<repo>/planning/planning/tasks/ID-task.md`:

```
$ tskflwctl task depend repair --drop 'planning/tasks/taaaaaaaaaa1-rel.md:depends_on=bad-token-one#0'
error: not found: repair source {TaskID: TaskSlug: Location:…/planning/planning/tasks/
       taaaaaaaaaa1-rel.md} does not identify a readable task record          # exit 10
```

The working relative form is `tasks/<file>`; task-id and slug selectors work; emitted commands use
absolute paths and work (all 11 verified). So this is a docs/UX defect that fails closed, but it is
the first thing a reader copies.

- **Violated contract.** Task scope: "Human previews provide commands suitable for patching the
  selected defects." Brief: "Attack shell and manifest grammar independently."
- **Smallest sound recommendation.** Change the `Example` string to `tasks/ID-task.md:…` and rerun
  `just docs`. (Accepting `planning/`-prefixed paths instead would be a behaviour change and should
  not be smuggled in as a doc fix.)
- **Affected consumers.** `docs/cli/tskflwctl_task_depend_repair.md`, `tskflwctl task depend repair --help`.
- **Missing regression test.** A CLI test asserting the documented relative form resolves; ideally a
  guard that the `Example` selectors parse and resolve against a fixture repository.

---

**Resolution:** The example is planning-root-relative as tasks/ID-task.md,
generated CLI docs match it, and a command-level dry-run test executes that path
form successfully.

### Settled hostile angles

1. **`git diff --check`.** `git diff --check 58e99d9 HEAD` reports
   `docs/cli/tskflwctl_task_depend_repair.md:50: new blank line at EOF`. **Not a finding:** all
   95 files under `docs/cli/` end with a blank line — it is the docgen convention, and the new file
   only trips the check because it is an *added* file. The task's clean-`git diff --check` claim is
   consistent with running it on the unstaged working tree, where an untracked file is invisible.
2. **Malicious planner against the store port.** Eight hostile plans driven straight into
   `FS.MutateTaskGraphRepair`. Rejected with no bytes written: deleting a valid constraint;
   `Automatic: true` on an explicit defect; an unaccounted selection; overlapping dedupe + exact drop;
   duplicate identical operations; `drop-empty-field` aimed at `depends_on`; an unknown source. The
   eighth — a planner asserting `Reason: cycle` for what is really `invalid-dependency-id` — is
   accepted, but I confirmed the receipt reports `invalid-dependency-id`: `plan = authorized`
   (`dependency_repair.go:418`) discards planner-supplied reasons entirely. Not a finding.
3. **Emitted-command fidelity.** All 11 emitted `--drop` commands were executed verbatim through a
   shell against values containing spaces, `"`, `'`, `:`, `=`, `;`, `$( )`, `\`, a terminal `#3`, a
   literal `depends_on=trap`, and an embedded `a:depends_on=b`. Each removed exactly one declaration —
   its own. Terminal-`#<digits>` values round-trip because the renderer always appends an explicit
   `#<occurrence>` for `drop` and never for `dedupe`; `LastIndex` on the emitted form therefore always
   strips the synthetic suffix. Shell quoting via `shellQuote` is correct including embedded single
   quotes.
4. **Durable-prefix receipt scope.** Injected failure after the first of two writes: `Removed` = 1
   (prefix only, not the abandoned plan), `Impacts` = 1 (the written task, `direct=true`),
   `ThreadImpacts` = 1 scoped to the prefix, `Addressed` = 1, `Residual` names exactly the remaining
   defect, `AppliedFiles`/`RemainingFiles` exact, `Selected`/`Operations` = 2 (full convergent intent,
   which is correct for retry). Retry converges; a third run is a clean no-op. Truthful throughout.
5. **CAS layering.** No false commit and no clobber under concurrent edits to: readable non-target
   task bytes, unreadable task bytes whose parse message is unchanged, readable Thread bytes,
   unreadable Thread bytes. `SameSourceSnapshot` compares the byte hash for both readable records and
   load problems, so a "same error, different bytes" edit is caught. `SameRepairSnapshot` deliberately
   clears `SourceVersion`/`Updated` for sources the prefix changed; I confirmed the residual blind
   spot is limited to body bytes and `updated_at` of a file we just wrote, neither of which can hide a
   lost repair (`domain.Task` carries no body, and any dependency-field or other frontmatter change is
   caught by `reflect.DeepEqual`).
6. **Wire truthfulness and leakage.** A fully populated non-default envelope — dedupe and
   drop-empty-field actions, non-zero occurrence, legacy fields, projected-edge endpoints, candidate
   IDs, a `Cycle` slice, absolute *and* relative locations, `Committed: true`, populated impacts and
   Thread impacts — validates against `schema --json-schema`. All twelve collections serialize as
   arrays, never `null`. `cloneThreadReadProblemsPublic` blanks `SourceVersion`; I planted a sentinel
   token and confirmed it does not reach the payload. `SchemaVersion` correctly bumped to `1.61`.
7. **Surgical YAML.** Flow and block sequences keep their style; comments on retained nodes survive
   and comments on removed nodes go with them; CRLF is preserved end to end; per-node quoting
   (`"a b"`, `'c:d'`, bare) is untouched; key order and unknown keys are preserved; a `depends_on:`
   line inside a fenced body block is not touched; an existing `updated_at` is replaced in place, not
   duplicated; last-value removal drops `depends_on` but keeps an empty legacy key until
   `drop-empty-field` is selected; file mode `0600` is preserved. Only the alias case fails (M2).
8. **Thread safety.** `materializeTaskGraphRepair` resolves every write path with `filepath.Rel`
   against `s.tasksDir` and rejects `..`/absolute escapes, so repair provably cannot write a Thread
   document. Malformed Threads never block task repair and are reported as
   `incomplete_thread_evidence` while their exact bytes stay CAS-protected.
9. **Dedicated-port isolation.** See the consumer inventory. `WithTaskGraphRepairStore` never infers
   the capability from `TaskGraphMutationStore`, and removing the `GraphBroken` guard from ordinary
   graph mutation is killed by three existing tests.
10. **Convergent absence vs typo.** A `--drop`/`--dedupe` whose *value* is absent from a still-present
    source is a silent no-op by design, while a missing or ambiguous *source* raises `ErrNotFound`
    (10) / `ErrAmbiguous` (13) and a present *valid* constraint raises `ErrValidation` (11). The
    boundary is where the brief asks it to be. Not a finding.

### Residual risks (documented, not findings)

- **Plan atomicity is the multiplier behind H1 and M2.** Neither root cause would be more than a
  nuisance if one unusable operation did not abort the whole plan. If both are fixed at the measure
  and materializer, the underlying "all-or-nothing plan" design still means the next
  not-yet-imagined un-scorable declaration will brick `--auto` repository-wide. Worth deciding
  explicitly whether `--auto` should skip an operation it cannot prove and report it as residual,
  rather than refusing the run.
- **Human impacts can render as no-ops.** `taskGraphRepairImpacts` filters on full `TaskGraphState`
  equality, but `TaskGraphRepairHuman` prints only `Role`/`Gate`; a change confined to
  `Eligible`/`Drained`/`Inconsistent` renders as `candidate/clear -> candidate/clear`. Observed in a
  live `--auto` run. The JSON carries the real fields, so this is presentation only.
- **Duplicate residual lines.** A `duplicate-task-id` problem is emitted once per shadow record, so
  bare diagnosis prints the identical `⚠ needs-direct-edit/duplicate-task-id` line twice. Cosmetic.
- **Semantically invalid manifest combinations are silent no-ops, not rejections.** A manifest
  operation `{action: drop-empty-field, field: depends_on}` is accepted by `readGraphRepairManifest`
  (`validRepairField` admits `depends_on`), then becomes "already satisfied" whenever the task has no
  `depends_on` key — it appears in `selected` with no operation instead of failing closed.
  `SimulateSourceEdits` *does* reject the combination, but that path is never reached. Fails safe
  today; tightening `readGraphRepairManifest` would be cheap.
- **`--dedupe` of a duplicated *invalid* value is auto-selected.** The diagnosis switch tests
  `Occurrence > 0` before `id.Valid`, so occurrence 1+ of a duplicated invalid token is
  `duplicate-dependency`/`Automatic` while occurrence 0 stays `invalid-dependency-id`/explicit. The
  net effect is information-preserving (one copy of the raw token survives verbatim for the operator
  to drop explicitly) and I judge it consistent with "`--auto` may only deduplicate canonical values",
  reading "canonical" as the field rather than the value. Flagging it only because it is a policy call
  a reader could reasonably want stated in the task.
- **`InspectTaskGraphRepair`, `SameRepairSnapshot` and `TaskGraphThreadImpacts` have no direct unit
  test.** All three are exercised indirectly and behaved correctly under every probe; I could not
  construct a failure. Noting the gap because each is a distinct semantic contract.

### Exact validation and mutation results

**Validation (sandbox, `go1.26.6 darwin/arm64`)**

```
go test -count=1 -race ./...          all packages ok (cli 6.4s, core 2.6s, store 6.5s, tui 9.2s, …)
golangci-lint run ./...               0 issues.
go mod tidy; git diff go.mod go.sum   clean
./bin/tskflwctl lint                  ✔ all planning entities and dependency links pass lint
just docs; git diff --exit-code docs/cli   stable
go run ./internal/tools/schemacomments     stable (257 comments, no diff)
git diff --check 58e99d9 HEAD         1 hit — docs/cli/tskflwctl_task_depend_repair.md:50 (see Settled §1)
```

**Mutation results** (each applied alone to the sandbox source; "killed" = at least one test fails)

| Mutation | Named test | Full `go test ./...` |
| --- | --- | --- |
| `applySourceDeclarationEdits` over-removes one value | — | killed by 3 value-asserting tests, **not by any guard** (M1) |
| …coordinated with `updateDependencySourceEdits` over-removing | — | killed by 5 value-asserting tests, **not by any guard** (M1) |
| `automaticRepairReason` += `RepairInvalidID` | survives | **survives** (M3a) |
| automatic-provenance check → `if false` | survives | **survives** (M3b) |
| `sameRepairOperationSet` check → `if false` | survives `…RejectsUnaccountedReceiptSelections` | **survives** (M3c) |
| selector `Index`/leftmost → `LastIndex`/rightmost | survives `…ManifestAndSelectorParsing` | **survives** (M3d) |
| `rejectOverlappingRepairOperations` → `return nil` | killed by `TestTaskGraphRepairRejectsOverlappingDedupeAndExactDrop` | killed |
| `sameTaskGraphLoadProblem` drops the `SourceVersion` compare | killed by `TestMutateTaskGraphRepairRejectsLateUnreadableTaskByteChange` | killed |
| `repairPlanPrefix` returns the full plan | killed by `TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix` | killed |
| ordinary graph mutation accepts `GraphBroken` | — | killed by `TestMutateTaskGraphBrokenSnapshotFailsBeforePlanner`, `TestValidateTaskGraphMutationSourceNamesGuardedRepairPath`, `TestPlannerWindowSetupFailureReturnsInsteadOfHanging` |

Each surviving mutation was re-run against a purpose-built discriminating probe; every probe passed
on unmutated source and failed under its mutation, confirming the guards are live and the gap is in
the tests, not the implementation. All probe files and throwaway planning spaces were deleted;
`git diff` against the sandbox baseline is empty apart from this audit.
