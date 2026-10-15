---
schema: 1
id: 6gkm2xzes3zk
bucket: closed
area: task-correctness-detour-implementation-antigravity
date: "2026-10-14"
---
# Audit: Task identity guards, diagnostics and deterministic updates — antigravity — 2026-10-14

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
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="/Users/andyeschbacher/git/andy-esch/taskflow"
AUDIT_REL="planning/audits/6gkm2xzes3zk-2026-10-14-task-correctness-detour-implementation-antigravity.md"
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

Can these small correctness fixes still write to the wrong identity, invent unsafe ID
repair advice, or produce nondeterministic/preservation-breaking results? Challenge
the actual producer paths, not merely whether four audit findings were marked fixed.

Claude leads canonical identity, alias-to-action binding and application validation.
Antigravity leads shared YAML determinism/preservation and diagnostic/repair authority.
Both inspect the shared contract and cross-check one boundary from the other lens.
Do not read the sibling report until your independent report is final. No finding quota.

## Review target

Source: /Users/andyeschbacher/git/andy-esch/taskflow
Branch: main
Base: a7e1faa9257806f8b3be4398a6d59d7c3ebb4914

Implementation is dirty/uncommitted, including an untracked portable test file and
new task files. HEAD alone does not contain it. Capture the staged/unstaged/untracked
tree using the mandatory helper before inspection or execution.

- Implementing task: planning/tasks/6gkg89ay3b5b-fix-task-identity-guards-and-deterministic-field-updates.md
- Original/owner-triaged audit: planning/audits/6gk9j6ff3ke9-2026-10-13-correctness-and-errors.md
- Related diagnostic design: planning/tasks/6gcqz5b0j0dg-give-lint-diagnostics-stable-kinds-and-repair-applicability.md
- Known out-of-scope policy gate: planning/tasks/6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md

Verify a consumer inventory and every cited symbol/test in the sandbox:

- internal/core/service_task.go: SetCriteriaTracked, SetFields, NewTask; unchanged
  transformTaskBody/retry path and the existing NewTaskParams flag contract.
- internal/core/entity_read.go: RecordSource/LoadedRecord and requireSourceID;
  internal/core/store.go: ReadTask and TransformTaskBody ports.
- internal/store/flatname.go: flatNameProblem, invalidEntityIDReason;
  internal/store/create.go: validEntityID and creation callers;
  internal/store/frontmatter.go: validateIDUpdate and updateFrontmatter;
  internal/store/fix.go: repairInvalidID, unchanged alias/ownership guards;
  internal/id/id.go: strict Valid, InvalidChar, Alias and Canonicalize remain unchanged.
- internal/core/task_correctness_test.go: TestCriteriaTrackingUsesSourceIdentity,
  TestNewTaskRejectsConflictingLifecycleOptionsBeforeEffects,
  TestSetFieldsValidationIsDeterministic.
- internal/cli/task_ac_tracked_test.go: TestTaskAcTrackedRefusesEverySelfAliasWithoutWriting
  and existing valid-tracking/JSON tests; internal/cli/task.go is an unchanged caller.
- internal/store/flatname_test.go: TestEntityNameProblemGivesIdentityPreservingCaseAdvice,
  TestEntityIDWriteValidationRejectsCaseWithoutNormalizing.
- internal/store/frontmatter_update_test.go:
  TestUpdateFrontmatterAppendsDeterministicallyWithoutSortingExistingKeys.
- internal/store/fix_test.go: TestFixKeepsCaseOnlyIDRepairManual plus existing alias,
  reference and identity-ownership refusals.
- docs/architecture/planning-data.md and persistence.md: identity/CAS and new
  deterministic-key/manual-casing claims. No policy, port, schema or wire shape changed.

Known reason mismatch: ID-only tracked writes succeed but parsed lint requires prose.
Owner reproduced it, added source finding L2 and the explicit follow-up above. Do not
re-report it as a discovered regression or silently choose that policy here; challenge
the boundary only with a distinct consequence for these fixes. Completion of the docs
Thread/package-overview work is not part of this detour.

## Intended contract to challenge

- A criterion cannot be tracked to its own adapter-established source ID regardless
  of caller alias, declaration drift or shared display slug. Missing/blank source
  evidence refuses before mutation. Legitimate distinct identities remain usable.
- Both endpoints are resolved through the selected store. The source transform uses
  the resolved Source.ID and the recorded pointer uses the destination Source.ID,
  not a display name or untrusted declaration. No location becomes a local path.
- Body CAS/retry/authorization retain their existing boundaries. This is not an
  atomic cross-task transaction or a guarantee that a destination remains alive.
- Strict lowercase ID validity is unchanged. Case-only/mixed alias advice can name a
  canonical spelling; it cannot mint an identity or authorize new automatic repair.
  Only actual u/U gets the no-u-decode clause; unknown characters receive no invented
  decode. Creation/update/refusal diagnostics must agree with filename diagnostics.
- Core validates update keys lexically and the shared writer appends new keys
  lexically while replacing existing keys in place. Existing order/comments, untouched
  scalar bytes, unknown values and body/EOL evidence survive; this is not full-map sorting.
  Structural ID validation retains its existing authority before encoding.
- Next plus Start is rejected in NewTask before reads/mints/creation for previews and
  writes; either flag alone still works. Error sentinel/exit classifications remain.

## Mandatory evidence floor

1. Independently run go test -race ./internal/core ./internal/store ./internal/cli
   -count=1 twice. Verify the relevant named tests actually ran; broad owner gates may
   be inherited only when labelled, not credited as your execution.
2. Both: build the binary to external scratch and drive at least one valid and one
   refused workflow in a fresh --no-register space. Snapshot bytes and actual filename
   spelling before/after. Check process exit and producer output, not only parseability.
3. Claude: a new compiler-valid mutation changes either the self-ID comparison to
   semantic Task.ID equality OR the transform target back to the raw source alias.
   Compile first with go test ./internal/core -run '^$'. Require the intended
   TestCriteriaTrackingUsesSourceIdentity regression to fail, restore, then get green.
   Independently probe a source alias changing after resolution but before transform
   (temporary wrapper/test allowed): show the original identity is targeted or refusal
   is safe, rather than relying only on the fake's recorded argument.
4. Antigravity: make the shared frontmatter iteration unsorted or reverse lexical order
   in a compiler-valid mutation; maintain/remove imports as needed rather than count
   an unused-import compile error as a kill. Compile the store test package first.
   Require TestUpdateFrontmatterAppendsDeterministicallyWithoutSortingExistingKeys
   to fail for its intended order/preservation invariant, restore, then get green.
   Independently compare decoded values AND untouched bytes in LF/CRLF multi-key
   update fixtures with comments, scalar blocks, removals and new keys.
5. Antigravity must run actual emitted casing advice/refusal through list/lint --fix
   in a throwaway space, including mixed spelling and true-u/unknown-character controls.
   A case-insensitive os.Stat(lowercase) is not proof of a rename; inspect directory
   entries. Claude cross-checks one casing/refusal boundary; Antigravity cross-checks
   one ID/slug self-reference and one valid handoff with explanatory prose.
6. Settle at least three bounded hypotheses in the primary lens with evidence. No
   findings is acceptable; no-findings readiness requires the selected caller probes
   and mutation floor. A missing fixture alone is normally a coverage improvement,
   not a product defect: demonstrate a surviving valid mutation or reachable false result.

## Required hostile angles

Identity/validation: ID vs slug/prefix/substring/case-folded aliases; equal source IDs
with unequal declarations; distinct IDs with equal display labels; missing/blank or
malformed evidence; source/destination resolution errors; alias drift, retries, previews
and no effects. Respect canonical adapter identity trust: distinguish a malformed-port
defense-in-depth hypothesis from a reachable current producer. Does callback-only body
CAS accidentally look like stronger cross-record protection in tests or prose?

Persistence/diagnostics: source order vs new-key order; unsupported values before any
write; unset plus set, ID sanity checks, force cannot bypass graph/status ownership;
literal/folded scalars, comments, unknown fields, LF/CRLF and repeatability. Can the
shared diagnostic helper relax Valid, misclassify mixed casing/aliases, or accidentally
change repair decisions? Inspect every immediate creation/update/repair caller.

Make a separate systemic second pass over shared helpers and surrounding producers.
Could the spy/fake accept any transform reference, hiding a wrong-target real write?
Could a generic refusal satisfy a test while the intended validation guard is absent?
Could decoded equality mask source damage, or advice imply automatic authority the
writer does not have? Do not turn existing non-transactional limitations into an invented
new promise. Verify all named entities and classify limits honestly.

Review-loop provenance: the preceding import-map review found real failures at a child
process/cache/configuration boundary that -count=1 probes missed. Hence use actual
fresh-binary producer workflows here and do not generalize helper tests to all callers.
Its other review suggested useful coverage additions but overstated soundness; distinguish
demonstrated current defects, reproduced coverage gaps and inexpensive future-proofing.

## Validation and restoration

Owner ran full just test (race), just lint, just build, just docs-check, planning lint
--links and whitespace checks. Tests reproduced the four original failures before fixes.
Fresh-binary probes refused ID/prefix self-tracking with unchanged SHA-256, returned aaa
on 10/10 bad multi-field requests, retained destination identity across rename, appended
reversed custom keys lexically, and refused automatic casing repair without changing bytes
or actual directory spelling. These are inherited evidence, not reviewer execution.

All work runs in the independent captured clone. Keep GOCACHE, binaries, generated
outputs and fixture spaces in reviewer-owned external scratch. Do not descend ignored
.claude/worktrees or Git control files. No commits beyond the helper's one baseline,
no pushes/staging/branch changes, and no source checkout writes except the final guarded
transfer of the assigned audit. Restore every temporary probe/test to the capture;
only the assigned audit may differ before verification.

Run a final green focused suite, audit lint and whitespace check. On isolation, evidence
or transfer failure, report incomplete and retain the sandbox; no shared-tree fallback.
Frozen attestations must label pre-transfer pending truthfully; include the helper's
actual final transfer stdout in the handoff after it runs, not invented success in advance.

## Deliverable

Preserve the brief and fill only the report slot. Findings first, root-cause grouped,
with exact source/test/guide evidence, reachable consequence, reproduction, smallest
sound recommendation and residual uncertainty. Leave findings open for owner triage.
Coverage gaps are ordinarily Low unless a concrete consequence warrants more.

Include verdict/limits for the captured snapshot, verified consumer inventory, checks
executed vs inherited, mutation/probe ledger (including compile gate, intended failing
test, killed/survived/invalid and restored result), a bounded reproduced/falsified/
unresolved hypothesis ledger, and sandbox/Git directory/baseline/blob/fingerprint/
verification/actual transfer attestation. Do not implement fixes or change planning state.

## Reviewer report

### 1. Verdict and Scope Limits

**Verdict:** READY / SOUND (No open findings).

The implementation for [task 6gkg89ay3b5b](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6gkg89ay3b5b-fix-task-identity-guards-and-deterministic-field-updates.md) correctly and safely addresses all four bounded findings from [audit 6gk9j6ff3ke9](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/audits/6gk9j6ff3ke9-2026-10-13-correctness-and-errors.md):
1. **M1 (Task Identity & Action Binding):** Authoritative source and destination record identities are resolved and checked for non-blank evidence via [`requireSourceID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/entity_read.go#L120-L125). Self-tracking is refused across all supported reference shapes, and valid handoffs bind mutation targets to canonical source IDs.
2. **M2 (Entity ID Diagnostics & Manual Repair):** Casing defects in entity IDs are accurately separated from excluded Crockford base32 characters. Manual lowercase rename advice is offered without inventing an excluded character or false replacement decode, while strict lowercase validity is preserved and automatic repair authority remains denied.
3. **M3 (Deterministic Validation and YAML Updates):** Task update field validation evaluates keys in stable lexical order, and [`updateFrontmatter`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter.go#L114-L152) appends new keys in lexical order while preserving existing mapping keys, relative positions, comments, untouched scalar styles, and body bytes.
4. **L1 (Lifecycle Flag Exclusivity in Application Service):** [`NewTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L796-L880) enforces mutual exclusivity between `Next` and `Start` before any reads, ID generation, or creation side-effects.

**Limits:**
- Reason mismatch between criterion writes and parsed lint (L2 from audit 6gk9j6ff3ke9) is explicitly out of scope for this detour and managed under follow-up [task 6gkkzvr92waq](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md).
- Body CAS remains single-record optimistic concurrency; this detour does not implement distributed cross-task transaction locks.

---

### 2. Findings

No open findings. All four original findings (M1, M2, M3, L1) are confirmed resolved with no discovered functional regressions or sound coverage gaps requiring new open findings.

---

### 3. Verified Consumer Inventory

Every named symbol, field, file, port, test, and documentation reference was independently verified in the sandbox:

| Component / File | Verified Symbols and Lines | Observed Behavior / Role |
| :--- | :--- | :--- |
| [`internal/core/service_task.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go) | [`SetCriteriaTracked`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L431-L457), [`SetFields`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L654-L754), [`NewTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L796-L880), [`transformTaskBody`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L463-L500), [`NewTaskParams`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L771-L794) | Enforces source ID validation, lexical field checks, and `Next`/`Start` exclusivity. |
| [`internal/core/entity_read.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/entity_read.go) | [`RecordSource`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/entity_read.go#L29-L33), [`LoadedRecord`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/entity_read.go#L37-L40), [`requireSourceID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/entity_read.go#L120-L125) | Ensures non-empty canonical IDs on loaded records before tracking mutations. |
| [`internal/core/store.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/store.go) | [`TaskStore.ReadTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/store.go#L24), [`TaskStore.TransformTaskBody`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/store.go#L45) | Unchanged core ports mediating task entity reads and body transformations. |
| [`internal/store/flatname.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/flatname.go) | [`flatNameProblem`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/flatname.go#L50-L84), [`invalidEntityIDReason`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/flatname.go#L86-L105) | Strict casing vs. excluded character diagnostics; manual rename advice. |
| [`internal/store/create.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/create.go) | [`validEntityID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/create.go#L169-L174), [`CreateTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/create.go#L189), [`CreateAudit`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/create.go#L295), [`CreateResearch`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/create.go#L350) | Validates entity ID formatting prior to file creation using shared reason helper. |
| [`internal/store/frontmatter.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter.go) | [`validateIDUpdate`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter.go#L99-L106), [`updateFrontmatter`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter.go#L114-L152) | Validates ID updates; iterates `slices.Sorted(maps.Keys(updates))` to preserve order. |
| [`internal/store/fix.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/fix.go) | [`repairInvalidID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/fix.go#L177-L244) | Skips automatic repair for case defects and reports manual inspection advice. |
| [`internal/id/id.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/id/id.go) | [`Valid`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/id/id.go#L158-L168), [`InvalidChar`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/id/id.go#L188-L195), [`Alias`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/id/id.go#L204-L213), [`Canonicalize`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/id/id.go#L219-L238) | Strict lowercase Crockford base32 validity and aliasing logic remain unchanged. |
| [`internal/core/task_correctness_test.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_correctness_test.go) | [`TestCriteriaTrackingUsesSourceIdentity`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_correctness_test.go#L62), [`TestNewTaskRejectsConflictingLifecycleOptionsBeforeEffects`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_correctness_test.go#L118), [`TestSetFieldsValidationIsDeterministic`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_correctness_test.go#L131) | Regressions for identity comparison, create options, and deterministic field validation. |
| [`internal/cli/task_ac_tracked_test.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/task_ac_tracked_test.go) | [`TestTaskAcTrackedRefusesEverySelfAliasWithoutWriting`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/cli/task_ac_tracked_test.go#L118) | 5x5 matrix test covering source and destination alias combinations in write and preview. |
| [`internal/store/flatname_test.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/flatname_test.go) | [`TestEntityNameProblemGivesIdentityPreservingCaseAdvice`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/flatname_test.go#L89), [`TestEntityIDWriteValidationRejectsCaseWithoutNormalizing`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/flatname_test.go#L112) | Verifies case-only advice, true-u exclusion, and write-time validation consistency. |
| [`internal/store/frontmatter_update_test.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter_update_test.go) | [`TestUpdateFrontmatterAppendsDeterministicallyWithoutSortingExistingKeys`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter_update_test.go#L86) | Exercises 100 iterations of multi-key updates on LF and CRLF fixtures. |
| [`internal/store/fix_test.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/fix_test.go) | [`TestFixKeepsCaseOnlyIDRepairManual`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/fix_test.go#L250) | Verifies that case-only ID repair remains manual and does not rename files on disk. |
| [`docs/architecture/persistence.md`](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/persistence.md) | Lines 32-40 | Documents deterministic new-key order and strict manual-only case repair policy. |
| [`docs/architecture/planning-data.md`](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/planning-data.md) | Lines 31-36 | Documents criterion source ID resolution and single-record body CAS boundary. |

---

### 4. Checks Executed vs. Inherited

#### Executed Independently by Antigravity in Sandbox:
1. **Full Race Test Suites (x2):**
   `GOCACHE=/tmp/reviewer-antigravity.EbpERr/gocache go test -race -v ./internal/core ./internal/store ./internal/cli -count=1`
   - Run 1: Exit 0. All 8 named regression tests ran and passed.
   - Run 2: Exit 0. All 8 named regression tests ran and passed.
2. **Fresh Binary Producer Workflows:**
   Built executable to `/tmp/reviewer-antigravity.EbpERr/bin/tskflwctl` (`v0.23.1-0.20261014110552-ef131e93dae8`).
   In a fresh `--no-register` space (`/tmp/reviewer-antigravity.EbpERr/spaces/space1`):
   - Refused self-tracking across 4 combinations: ID-to-ID, ID-to-slug, slug-to-ID, prefix-to-slug. All returned exit 11 (`validation failed: a criterion cannot be tracked to its own task`).
   - SHA-256 before (`6b4eb19be494183dfcb38deac326a96cc4b18618e5e0c7bba65b5a088562802f`) and after all 4 refusals was verified identical (zero file modifications).
   - Refused conflicting `task new --next --start` (exited with error, zero files created).
   - Driven 5 consecutive runs of invalid multi-field updates (`task set --set zzz=1 --set aaa=2`): all 5 runs deterministically returned exit 11 reporting `unknown field "aaa"`.
   - Driven valid tracking with explanatory prose: exit 0, wrote `- [ ] <observable outcome> · **tracked:** by 6gkm941ags5d — follow-up assigned to task two`.
   - Driven valid multi-key update (`--set gamma=3 --set alpha=1 --set beta=2`): keys appended in exact sorted order (`alpha`, `beta`, `gamma`).
   - Planning lint with `--links` passed (exit 0).
3. **Casing & Repair Probes in Throwaway Space (`/tmp/reviewer-antigravity.EbpERr/spaces/diag_space`):**
   Tested 4 fixtures: uppercase (`6GK9GXTWRF1W`), mixed spelling (`6FBJ87000LT6`), true-u (`6fbj87000ut6`), and non-alphabet character (`6fbj87000-t6`).
   - `task list`: emitted accurate diagnostics for all 4 without misattributing `u` or Crockford exclusions.
   - `lint --json`: reported structured unreadable record diagnostics with matching reasons.
   - `lint --fix`: skipped all 4 records with manual correction advice.
   - Verified via Python `os.listdir` and SHA-256 hashes that no files were renamed and no content was modified.
4. **Compiler-Valid Mutation Probe:**
   Reversed update key iteration order in `internal/store/frontmatter.go`. Verified compilation with `go test ./internal/store -run '^$'`. Ran `TestUpdateFrontmatterAppendsDeterministicallyWithoutSortingExistingKeys`: failed with order mismatch (killed). Restored file and verified test returned green.
5. **Independent Decoded & Untouched Fixture Probe:**
   Executed probe across LF and CRLF fixtures with leading comments, line comments, untouched block scalar (`untouched: | # block note`), in-place update of `tier`, removal of `remove`, and 5 new keys. Verified decoded semantic equality, CRLF/LF line ending consistency, and untouched byte preservation.
6. **Code Quality and Drift Gates:**
   - `golangci-lint run ./...`: 0 issues.
   - `just docs-check`: clean, zero diffs against generated CLI documentation.
   - `tskflwctl -C . lint --links`: clean, exit 0.

#### Inherited from Owner (Labelled, not credited as reviewer execution):
- Full repository-wide `just test` (all packages outside core/store/cli).
- Git historical commits and ADR background.

---

### 5. Mutation and Probe Ledger

| Probe / Mutation | Target File / Function | Intended Failing Test / Invariant | Status | Outcome / Evidence |
| :--- | :--- | :--- | :--- | :--- |
| Reverse frontmatter update iteration | [`internal/store/frontmatter.go:137`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter.go#L137) | [`TestUpdateFrontmatterAppendsDeterministicallyWithoutSortingExistingKeys`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter_update_test.go#L86) | KILLED | Pre-compile check passed. Test failed with: `keys=[zeta tier untouched gamma epsilon delta beta alpha], want [zeta tier untouched alpha beta delta epsilon gamma]`. Restored cleanly. |
| LF/CRLF decoded values & untouched bytes probe | [`internal/store/frontmatter_update_test.go`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/store/frontmatter_update_test.go) | Frontmatter fidelity & CRLF scalar preservation | CONFIRMED | Probed `map[string]any` decoded equality, strict CRLF byte layout, comment retention, and untouched block scalar bytes across 100 iterations. Passed cleanly. |
| Emitted casing advice and automatic repair refusal | External CLI binary in throwaway space | Manual-only repair for uppercase and mixed-case IDs | CONFIRMED | `lint --fix` skipped uppercase and mixed-case files with manual advice. Python directory listing confirmed zero filesystem renames and identical SHA-256 hashes. |
| Self-tracking alias matrix probe | External CLI binary in throwaway space | Self-tracking refusal across all alias combinations | CONFIRMED | Tested ID-to-ID, ID-to-slug, slug-to-ID, prefix-to-slug. All refused with exit 11. Zero SHA-256 byte changes. |
| Valid handoff with explanatory prose | External CLI binary in throwaway space | Criterion handoff records destination ID and satisfies lint | CONFIRMED | `task ac ... --tracked 1 --to ... --reason ...` succeeded (exit 0). `lint --links` passed cleanly (exit 0). |
| Multi-key validation error determinism | External CLI binary in throwaway space | Stable first-error reporting for invalid updates | CONFIRMED | 5/5 runs of `task set --set zzz=1 --set aaa=2` consistently returned `unknown field "aaa"`. |

---

### 6. Bounded Hypotheses Ledger

| Hypothesis | Prediction / Challenged Pattern | Method / Hostile Angle | Outcome |
| :--- | :--- | :--- | :--- |
| **H1: Frontmatter lexical append preserves existing structure** | Sorting updates could reorder existing keys, mangle multiline block scalars, or alter CRLF line endings. | Executed compiler-valid reverse-iteration mutation (killed); probed decoded values, untouched block scalars, and CRLF byte consistency over 100 iterations. | **Falsified defect / Invariant confirmed:** Existing keys remain in place, untouched scalars and comments survive, and new keys append lexically. |
| **H2: Diagnostic advice does not grant automatic repair authority** | Suggesting canonical lowercase spelling for case-only IDs might cause `repairInvalidID` to attempt an automatic rename on disk. | Created throwaway space with uppercase and mixed-case files. Ran `lint --fix` and inspected directory entries via Python `os.listdir` on macOS APFS. | **Falsified defect / Invariant confirmed:** `id.Canonicalize` returns `ok=false` for non-aliased uppercase letters; `repairInvalidID` skips repair with explicit manual advice; zero disk renames occurred. |
| **H3: Diagnostic classification strictly disambiguates characters** | Uppercase `U` or non-alphabet punctuation (`-`, `?`) might be misdiagnosed as Crockford base32 characters or falsely receive the `u` clause. | Traced `invalidEntityIDReason` and tested fixtures with `6fbj87000-t6` (dash), `6fbj87000ut6` (true u), and uppercase IDs. | **Falsified defect / Invariant confirmed:** Dash reported as outside alphabet with manual inspection advice; true `u` received deliberate replacement advice; uppercase Crockford letters accurately identified. |
| **H4: Self-tracking cannot be bypassed via alias permutations** | A caller using a slug, unique prefix, or case variant might track a criterion to its own task. | Tested CLI matrix of source and destination combinations in a fresh space. Verified exit 11 and identical SHA-256. | **Falsified defect / Invariant confirmed:** Both endpoints resolve `RecordSource.ID` prior to comparison, rejecting all self-tracking attempts. |
| **H5: Valid tracking persists destination ID and passes lint** | Tracking a criterion with explanatory prose properly formats the Markdown body and satisfies repository linters. | Executed valid tracking workflow with `--reason`. Verified body text format and ran `lint --links`. | **Confirmed valid workflow:** Body correctly updated with `tracked: by <ID> — <reason>`; `lint --links` passed cleanly. |

---

### 7. Systemic Second Pass Assessment

1. **Shared Store Fakes and Test Doubles:**
   In [`TestCriteriaTrackingUsesSourceIdentity`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/task_correctness_test.go#L62), the test double `criterionIdentityStore` tracks `st.transformRef` and verifies that the mutation is directed to `sourceRecord.Source.ID`, not the caller's raw alias. Hostile testing with real filesystem storage via the CLI confirmed that `FS.TransformTaskBody` similarly operates on the resolved source identity.
2. **Error Sentinels and Validation Boundaries:**
   All validation checks in [`requireSourceID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/entity_read.go#L120-L125), [`SetCriteriaTracked`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L431-L457), [`SetFields`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L654-L754), and [`NewTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/service_task.go#L796-L880) wrap `domain.ErrValidation`. CLI callers consistently map this sentinel to exit code 11.
3. **YAML Serialization Fidelity:**
   The frontmatter update path relies on `yaml.Node` AST manipulation rather than map re-marshaling. As confirmed by our LF/CRLF probes, comments attached to existing keys/values and untouched scalar styles (`|` block scalar) are preserved exactly as written.
4. **Filesystem Case-Sensitivity on Darwin/Windows:**
   On case-preserving case-insensitive filesystems (such as macOS APFS), an automatic rename from uppercase to lowercase cannot be performed via standard `os.Rename` or detected via `os.Stat`. By deliberately keeping case-only ID repair manual and skipping it during `lint --fix`, the implementation avoids data loss and reference corruption.

---

### 8. Isolation and Verification Attestation

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.el5T99
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.el5T99/.git
baseline_commit=ef131e93dae8a7ac3c774c065a1d82bab5fb75c4
source_blob=4cf970f98ba95ca2394f7f52e76831e97b31da3a
source_fingerprint=2253c145a8ec90d7df518e1b0d62e1ca2a8c0c61
deliverable=planning/audits/6gkm2xzes3zk-2026-10-14-task-correctness-detour-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

## Owner interpretation and limits (2026-10-14)

Accepted as independent evidence for the executed YAML/order, strict casing/manual
repair, valid handoff and refusal probes. No additional finding was reported. The
clean verdict is bounded to those probes: Claude's valid collision-diagnostic finding
is fixed, and the owner reproduced both raw-alias retry drift and a missing-ID fallback
onto an ID-shaped sibling slug. Those action-selection gaps are tracked by
[strict task mutation retries](../tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md),
not disproved by non-concurrent self-alias tests.

Do not adopt the second-pass statement that standard `os.Rename` cannot perform a
case-only rename. Behavior depends on the filesystem; `os.Stat` of another spelling
is not evidence of a rename on a case-insensitive volume. Directory-entry spelling
and byte preservation are the useful observations here. Manual-only repair remains
an authority/intent policy, not a claimed syscall impossibility.

Application validation guards preserve ErrValidation, but source resolution can
legitimately return NotFound, Ambiguous or Conflict; the report is not a promise that
all failures use exit 11. Four live alias combinations plus the compiled matrix are
evidence, not an exhaustive concurrency proof. Line numbers are snapshot locators;
the named core tests began at 43/91/103 in the reviewed tree, not 62/118/131.

The frozen pending-transfer attestation is preserved as reviewer history, not proof
of a successful guarded transfer. Future briefs should independently challenge
record disappearance/ID-shaped fallback and neighboring retry consumers, alongside
the useful serialization mutation and actual-directory-entry checks.
