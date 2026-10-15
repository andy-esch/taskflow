---
schema: 1
id: 6gkm2xz5s9qx
bucket: closed
area: task-correctness-detour-implementation-claude
date: "2026-10-14"
updated_at: "2026-10-14"
---
# Audit: Task identity guards, diagnostics and deterministic updates — claude — 2026-10-14

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
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="/Users/andyeschbacher/git/andy-esch/taskflow"
AUDIT_REL="planning/audits/6gkm2xz5s9qx-2026-10-14-task-correctness-detour-implementation-claude.md"
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

Reviewer: Claude (Opus 5.5), lead for canonical identity, alias-to-action binding and application
validation. All inspection, builds, tests, fixtures, mutations and edits ran in the independent sandbox
named in the isolation attestation. GOCACHE, both binaries, the exported base tree and every fixture space
live in reviewer-owned scratch outside both trees. I did not read the sibling Antigravity report.

### Verdict and limits

**Ready as scoped for the four fixes, with one Low and one Medium finding left open.** The Medium is
pre-existing and lies outside this detour's criteria.

For the captured snapshot (baseline `2c815765`), the identity, validation and determinism contracts hold
under hostile evidence:

- Both endpoints of a criterion handoff are resolved through the selected store.
- Self-tracking is refused by `RecordSource.ID` for every unique alias shape, in write and preview, with
  bytes and the actual filename spelling unchanged.
- The source transform stays bound to the resolved ID when the alias drifts. That holds after resolution and
  also inside the CAS-conflict retry. I checked it with the real FS store, not just the fake's recorded
  argument.
- Next plus Start is refused before any effect.
- Multi-field validation reports a stable first error.
- Case-only ID advice is accurate, and repair stays manual.

The findings:

- **M1, systemic and pre-existing:** the sibling body and field verbs (`task ac --check/--defer/...`,
  `task append`, `task set`) still pass the caller's raw alias into `retryOnConflict`. One concurrent
  `task rename` inside the CAS window makes the retry re-resolve the alias through the fuzzy tiers. The
  write then lands, with exit 0, on a different, pre-existing task. At base `a7e1faa9` the tracked verb did
  the same; this detour fixed only that verb.
- **L1, introduced:** two distinct files claiming one canonical ID used to fail with exit 13, naming both
  files. They are now refused as self-tracking, with exit 11. This is fail-closed, and the base preview,
  which wrongly succeeded, is now fixed. But the actionable collision diagnosis is lost.

Limits:

- M1 is shown through the store's production code path and its existing concurrency seam
  (`testHookBeforeBodyWrite` / `testHookBeforeSetFieldsWrite`). It was not reproduced as a two-process timing
  race.
- Lifecycle moves (`task start`/`done`) were not probed for the same retry class.
- I ran no full `just test`, `just lint` or docs gates; those are inherited and labelled below.
- Antigravity leads YAML preservation and diagnostic authority. My casing check covers one boundary only.

## Findings

#### M1. Sibling task mutations retry with the raw caller alias, so one concurrent rename redirects the write to another task · **Status:** tracked by 6gkmexz4vh23

**File:** internal/core/service_task.go:480 | **Component:** core/retry,cli/task
**Effort:** S · **Urgency:** eventually

**Not introduced by this detour.** It shares the retry primitive that the source audit's M1 fix relies on, and
the detour fixed it for the tracked verb only.

**Mechanism:**

- `transformTaskBody` (`service_task.go:473-484`), `AppendBody` (`:507-516`) and `SetFields` (`:723-725`)
  wrap the store call in `retryOnConflict` (`retry.go:62-72`). Every attempt passes the **raw** caller ref.
- The CLI passes the typed argument straight through: `task.go:540` (check, uncheck, defer, wontfix, na),
  `:489` (add/remove/replace), `:676` (set) and `:789` (append).
- Each attempt re-runs `s.resolve(slug)` (`body.go:261`, `fsstore.go:274`), so after a conflict the full
  exact-ID, exact-slug, prefix and substring ladder runs again (`resolve.go:214-272`).
- The CAS guard deliberately re-resolves by canonical ID, "never the caller's raw query" (`cas.go:127-132`).
  It reports a move or rename as a retryable conflict that "can legitimately succeed" (`cas.go:144-148`).
- The retry above that guard then re-resolves the raw query anyway. That undoes the guard's own reasoning,
  and the `retry.go:56-58` claim that these mutations are "concurrency-safe" no longer holds.

**Reachable consequence:** take a repository with `alpha` and a pre-existing `alpha-two`.

1. A cooperating writer runs `task rename alpha "Omega"` between another command's read and its lock.
2. That command's CAS conflicts.
3. The retry resolves `alpha` through the prefix tier to `alpha-two` alone.
4. `alpha-two` is mutated and the command exits 0. The original task, now `omega`, is untouched.

No concurrent create is needed. The hook fires before the write lock (`body.go:63`, `fsstore.go:316`), so it
models a cooperating `task rename` that completes before the stale writer locks.

**Reproduction:** a temporary `internal/store` test, deleted before verification. It wires
`core.MustNewService(NewFS(...), WithRetry(1, ...))`, and a one-shot hook renames
`<id>-alpha.md` to `<id>-omega.md`. Results on the reviewed tree:

| Verb | Error | Original (`omega`) | Sibling (`alpha-two`) |
| --- | --- | --- | --- |
| `SetCriteriaState` (defer; `--check` uses the same call) | nil | unchanged | mutated |
| `SetAcceptanceCriterion` | nil | unchanged | mutated |
| `AppendBody` | nil | unchanged | mutated |
| `SetFields` (priority) | nil | unchanged | mutated |
| `SetCriteriaTracked` | nil | mutated | unchanged |

With the same probe, the exported base `a7e1faa9` mutated `alpha-two` for all five verbs, including tracked.
A rename-plus-new-alias-holder variant gave the same result for check, append and set.

The TUI is not exposed. It already passes canonical `Source.ID` keys (`tui/commands.go:81`, `:102`;
`atlas.go:424`; `tui/edit.go:358`). `ReplaceBody` has no retry (`service_task.go:500-502`).

**Recommendation:**

- Track this as a separate, scoped task rather than widening this detour.
- Bind each retried mutation to the identity its first attempt resolved. In core, that means resolving once
  through the selected store and retrying with `Source.ID`, the pattern `SetCriteriaTracked` now uses.
  Equivalently, a retry could refuse when the re-resolved ID differs.
- Add one shared real-store drift regression covering every retried verb.
- Correct the `retry.go:56-58` safety claim.

**Residual uncertainty:**

- Likelihood depends on concurrent renames. Exposure in practice is unmeasured.
- Lifecycle `move` uses a different snapshot path and was not probed.

**Resolution:** Owner reproduced all five legacy wrong-target paths with real FS
    pre-lock rename seams; tracked is the passing rename control. Follow-up
    6gkmexz4vh23 also covers the independently reproduced missing-ID fallback
    onto an ID-shaped sibling slug: binding a query to Source.ID alone is
    insufficient without strict selected-identity action lookup. The retry
    safety comment is corrected here; production retry migration remains out of
    scope. No existing open Thread owns this bounded epic-21 work.

#### L1. A duplicate canonical ID across two files is now refused as self-tracking instead of as an ambiguity · **Status:** fixed

**File:** internal/core/service_task.go:452 | **Component:** core/task-criteria
**Effort:** XS · **Urgency:** eventually

The new self-guard compares `destinationRecord.Source.ID == sourceRecord.Source.ID`. If two distinct files
claim one ID, for example after copying a task file as a template, the slug refs `alpha` and `copy` both
resolve uniquely, and the guard reports "a criterion cannot be tracked to its own task".

Reproduction with both fresh binaries, in unregistered spaces `w5-base` and `w5-review`. The setup is
`cp <id>-alpha.md <id>-copy.md`, then `task ac alpha --tracked 1 --to copy --reason x`:

| Build | Write | Preview |
| --- | --- | --- |
| Base | exit 13: `id "…" is claimed by 2 files: alpha (…), copy (…): ambiguous match` | exit 0: "would set criterion 1 tracked alpha" (a false preview) |
| Reviewed | exit 11, self-tracking text | exit 11, self-tracking text |

Bytes were unchanged throughout.

So the reviewed build is safer for previews, but the write lost its exit-13 classification and the
file-naming remedy. An agent that keys off the exit code would be told it chose the wrong destination,
when it should be told to repair the colliding ID. `task info <id>` in the same space returns exit 13 naming
both files, so the port can already surface the real cause.

**Recommendation:** when the source IDs are equal, first resolve that canonical ID through the selected
store, for example `s.store.ReadTask(sourceRecord.Source.ID)`. Propagate any `ErrAmbiguous` before
refusing self-tracking. Add a two-file shared-ID regression. This does not use `Location` as identity.

**Residual uncertainty:** whether equal canonical IDs should *always* mean "self" is a policy reading. I
classify this as diagnostic accuracy, not a wrong write.

**Resolution:** Equal endpoint source IDs are resolved canonically before
    self-track refusal, preserving adapter ambiguity and other resolution
    sentinels. Portable error tests and real CLI two-file duplicate-ID
    write/preview tests failed before the fix and pass afterward; collisions
    name both records with exit 13 and leave both files unchanged. Ordinary self
    aliases remain exit 11. No Location or declaration fallback was introduced.

## Reviewer evidence

### Verified consumer inventory

All paths and lines were verified in the sandbox at the baseline. "Unchanged" means
`git diff a7e1faa9 HEAD --stat` printed nothing for that file.

- **Core use cases, `service_task.go`:**
  - `SetCriteriaTracked` `:433-457`: source `ReadTask` `:438`, `requireSourceID` `:441` and `:448`, the ID
    comparison `:452`, and the transform by `sourceRecord.Source.ID` `:456`.
  - `setCriteriaState` `:459`; `transformTaskBody` `:473-484`, unchanged, with retry `:480`.
  - `SetFields` `:657`, with the sorted loop at `:664`.
  - `NewTaskParams` `:777`, where Next/Start `:786-787` document exclusivity; `NewTask` guard `:796-799`.
  - The only caller is `internal/cli/task.go:538` (unchanged).
- **Core types and ports:**
  - `entity_read.go`: `RecordSource` `:29`, `LoadedRecord` `:37`, `requireSourceID` `:120`
    (TrimSpace check; unchanged).
  - `store.go`: `ReadTask` `:24`, `TransformTaskBody` `:45` (unchanged).
  - `retry.go:62` (unchanged).
- **Store identity:**
  - `taskSource` derives the ID from `splitFlatName` (`store/entity_read.go:14-16`), which requires
    `id.Valid` (`flatname.go:37`).
  - `resolveID` gives exact-ID precedence (`resolve.go:214-238`); `resolveExactID` is at `:288`.
  - `FS.TransformTaskBody` is at `body.go:249` (unchanged).
- **Store diagnostics and writers:**
  - `flatname.go`: `flatNameProblem` `:62`, `invalidEntityIDReason` `:86`.
  - `create.go`: `validEntityID` `:169`, with callers `create.go:189,295,350`, `threadcreation.go:130`,
    `lifecyclemutation.go:241` and `edit.go:51`.
  - `frontmatter.go`: `validateIDUpdate` `:96`, which delegates to `validEntityID`; `updateFrontmatter`
    `:112`, sorted at `:137`.
  - `fix.go`: `repairInvalidID` `:183`, with `Canonicalize` `:192`, the u/U clause `:194`, the new manual
    refusal `:197` and the unchanged `referencesTo` guard `:202`.
- **ID primitives:** `internal/id/id.go` `Valid` `:158`, `InvalidChar` `:188`, `Alias` `:204` and
  `Canonicalize` `:219` are unchanged.
- **Tests, all present and passing in both runs:**
  - `task_correctness_test.go`: `:43`, `:91`, `:103`.
  - `task_ac_tracked_test.go`: `:118` (50 subtests); the existing `:34`, `:54`, `:67`, `:88`, `:106`,
    `:144` and `:159` (JSON).
  - `flatname_test.go`: `:89`, `:112`.
  - `frontmatter_update_test.go`: `:86`.
  - `fix_test.go`: `:250`, plus the existing alias repair `:143`, ownership `:171`, reference `:204` and u
    `:233` refusals.
  - `threadcreation_test.go:390`, the cross-kind collision test.
  - The existing Next/Start regressions `TestService_NewTask_Next` and the `CreateAndStart` lifecycle tests.
- **Guides:**
  - `planning-data.md:31-34` matches the evidence: per-record CAS, not cross-task.
  - `persistence.md:32-39` matches W2 and W3.
- **Shape:** no port, schema or wire change. `internal/wire`, `internal/tui`, `cmd`, `internal/core/store.go`
  and `internal/cli/task.go` are unchanged.

### Checks executed and inherited

Executed by me in the sandbox:

- `go test -race ./internal/core ./internal/store ./internal/cli -count=1` (`-json`) twice. Both runs were
  green with 2708 passing, 0 failing and 0 skipped.
  - In each run, all 8 new named tests, the 13 cited existing regressions and the 4 Next/Start lifecycle
    regressions passed.
  - The parametrised new tests ran all their subtests: 12, 50 and 4.
- A third, final run after restoration was green, 2708/0, and no probe test was present.
- Fresh binaries built to scratch: the reviewed tree (`review-2c815765`) and an exported base `a7e1faa9`
  (`git archive`).
- Fresh-binary producer workflows W1–W5 in new `--no-register` spaces, with `TSKFLW_NO_REGISTER=1` and XDG
  paths in scratch. Before/after `ls -1` and SHA-256 snapshots were taken.
- At the end: `git diff --check`, `audit lint 6gkm2xz5s9qx` and planning `lint` with the reviewed binary.

Inherited from the owner and not counted as mine:

- Full `just test`, `just lint`, `just build`, `just docs-check`, and planning `lint --links`.
- The owner's red-before-green reproductions and probes.

### Producer workflows (fresh binary)

- **W1 identity:**
  - `task ac … --tracked 1 --to …` gave exit 11, "its own task", for every self pair in write and preview:
    `alpha`↔ID, ID↔`alpha`, `alpha`→8-char prefix, uppercase-ID↔`ALPHA`, ID-substring↔`alpha`,
    `ALPHA`→substring, and prefix→uppercase ID. That is 10 unique shapes, 20 invocations.
  - Every refusal left all names and SHA-256 hashes unchanged.
  - Genuinely ambiguous refs (`alp`, `lpha`, 6-char ID prefixes shared by same-millisecond mints) gave
    exit 13, naming both candidates.
  - A valid handoff from the uppercase-ID source alias to `beta`: the preview left bytes unchanged, and the
    write exited 0 and changed only the source hash. The JSON `task.id` was the source and the body recorded
    `by <beta-id> — beta owns the artifact`.
  - After `task rename beta …`, `ac --list --json` kept `tracked_by` and the ID still resolved.
    `lint --links` exited 0.
  - Distinct IDs with an equal slug: a second `alpha`. The bare `alpha` gave exit 13. ID→ID tracking exited
    0 and recorded the second ID.
- **W2 application validation:**
  - `task new --next --start` gave exit 1 from Cobra (unchanged adapter guard) in write and preview, and no
    file was created.
  - `--next` alone produced `next-up` and `--start` alone produced `in-progress`.
  - 10/10 identical first errors, all exit 11:
    - `unknown field "aaa"` for `zzz/mmm/aaa`;
    - `autonomy_level must be an integer` for two bad typed fields;
    - `depends_on is graph-owned` for `--force` status plus `depends_on`.
  - Bytes were unchanged.
  - Forced `zeta/beta/alpha` plus `--priority low`: priority was replaced in place, and `alpha, beta,
    updated_at, zeta` were appended. Three repeats gave an identical hash.
- **W3 casing cross-check:**
  - An uppercase filename ID gave the list/lint/JSON advice `must be lowercase (canonical spelling …) —
    rename to "<lower>-…md"`.
  - `lint --fix` and its preview reported "not repaired automatically"; the real run exited 11.
  - Bytes and the actual `ls -1` spelling were unchanged.
  - Following the advice with a two-step rename made that record readable and lint-clean.
  - `--dry-run lint --fix` exiting 0 is deliberate (`cli/lint.go:91-98`) and not attributable to this change.
- **W4 malformed source** (`tier: [1, 2]`), base vs reviewed: `--tracked` and `--defer` both exited 11 with
  JSON `validation`, and bytes were unchanged. The reviewed message adds `resolve task "src": <path>:`
  context.
- **W5:** see L1.

### Mutation and probe ledger

| Probe | Exact change | Compile gate | Intended test | Result |
| --- | --- | --- | --- | --- |
| A | `:452` → `destinationRecord.Value.Task.ID == sourceRecord.Value.Task.ID` | `go test ./internal/core -run '^$'` ok | `TestCriteriaTrackingUsesSourceIdentity` | **Killed**: `same_source_different_declarations` write and preview reached the transform. CLI tracked tests survived, as expected, because real fixtures do not drift. Restored. |
| B | `:456` transform target → raw `slug` | ok | same | **Killed**: `distinct_sources_same_display_slug` gave `ref="source-alias"`. CLI tests survived. Restored. |
| B-real | B, plus the real-FS drift probes P1 and P2 | ok | P1, P2 | **Killed**: the impostor holding the old alias was mutated and the original was untouched. Restored. |
| P1 | No mutation. A wrapper `ReadTask` renames the source and creates a new `alpha` after core resolves | — | — | The original ID was targeted and the impostor's bytes were unchanged. |
| P2 | No mutation. `testHookBeforeBodyWrite` does the same drift inside the CAS window | — | — | Conflict, then retry by ID: the original was targeted and the impostor unchanged. |
| C | Next+Start guard moved after `ReadEpics` | ok | `TestNewTaskRejectsConflictingLifecycleOptionsBeforeEffects` | **Killed** (`reads=1`). Restored. |
| D | Core `SetFields` loop → `for field, val := range updates`, with `maps` kept referenced | ok | `TestSetFieldsValidationIsDeterministic` | **Killed**: all four subtests. Restored. |
| E | Both `requireSourceID` checks removed (coordinated) | ok | `TestCriteriaTrackingUsesSourceIdentity` | **Killed**: 8/8 missing- and blank-evidence subtests. No generic refusal masks them. Restored. |
| P3/P4 | Sibling verbs under rename drift (see M1) | — | — | **Reproduced**; also at base. |

Every mutation was applied in the sandbox from a scratch backup and restored with `cp`. After each one,
`git diff --quiet` passed on the restored file. The probe test was saved to scratch
(`zz_review_drift_probe_test.go.saved`, sha256 `e70be251…`) and deleted from the sandbox. It also ran
against the exported base tree, also in scratch.

### Hypothesis ledger

| # | Hypothesis | Result | Evidence |
| --- | --- | --- | --- |
| 1 | Some alias shape (ID, prefix, substring, case-folded, slug) bypasses the self-guard | Falsified | W1 20 invocations; CLI `:118` 50 subtests; mutations A and E |
| 2 | The transform still targets a raw alias or declaration under drift | Falsified | P1/P2 real store; B and B-real killed |
| 3 | Equal source IDs with unequal declarations, or distinct IDs with an equal slug, are misjudged | Falsified | Core cases; W1 second `alpha`; A killed |
| 4 | Missing or blank evidence reaches effects | Falsified | E killed 8/8; transforms=0 asserted |
| 5 | Whitespace-padded `Source.ID` evades the untrimmed `==` (`requireSourceID` trims) | Not reachable from the current producer | FS IDs come only from `id.Valid` stems (`flatname.go:37`); defense-in-depth only; not filed |
| 6 | Source-first resolution changes malformed-source classification | Falsified | W4: exit 11 and `validation` on base and reviewed |
| 7 | Duplicate canonical ID across files is handled accurately | **Reproduced false** | L1 |
| 8 | The shared retry keeps sibling mutations on the first-resolved record | **Reproduced false** (pre-existing) | M1; base probe |
| 9 | Next+Start reaches reads, mints or creates; single flags regress | Falsified | C killed; W2 |
| 10 | The multi-field first error is nondeterministic or writes | Falsified | D killed; W2 30/30 stable; bytes unchanged |
| 11 | The store `SetFields` graph-owned loop (`fsstore.go:268-272`) is still unsorted | Unreachable | Its only caller is core `:724`, which rejects graph-owned fields first; CLI and TUI go through core |
| 12 | The fake accepts any ref and hides a wrong-target real write | Falsified | Asserts `transformRef == sourceID`; exact-ID precedence `resolve.go:224-238`; P1/P2 |
| 13 | Case advice is wrong or grants automatic authority | Falsified for case-only | W3; mixed, u and unknown-character controls are Antigravity's lead |
| 14 | Docs overstate cross-record CAS | Falsified | `planning-data.md:33-34` explicitly disclaims it |
| 15 | Lifecycle `move` retries share the M1 class | Unresolved | Not probed; `retry.go:56` lists move |
| 16 | Historic self-tracked criteria are detected | Out of scope | Lint checks only the reason (`domain/body.go:747`), not `TrackedBy` targets; pre-existing; not filed |
| 17 | The known L2 reason mismatch has a distinct consequence here | None found | Refusals are independent of `--reason`; valid W1 writes used prose |

### Second pass (systemic)

- **Shared retry primitive.** It is the one demonstrated systemic issue (M1). Core's retry re-resolves the
  caller's text even though the store's CAS deliberately avoids that. The tracked fix is correct because it
  feeds the canonical ID into that primitive. Its siblings do not.
- **Test helpers.** `criterionIdentityStore` records the ref, and the test asserts it. Real-store
  equivalence follows from exact-ID precedence and was confirmed by P1/P2, so the helper is not masking a
  defect class. A real-store drift regression would still be cheap future-proofing; that is part of M1's
  recommendation, not a separate gap.
- **Fail-closed appearance.** The CLI self-alias test asserts exit 11 plus the specific message, so a
  generic refusal cannot satisfy it. Core refusal cases assert `transforms == 0`, and mutation E shows each
  guard is load-bearing.
- **Shared diagnostic helper.** `invalidEntityIDReason` checks `id.Valid` first and only formats text. Every
  write caller of `validEntityID`, including `validateIDUpdate`, inherits wording, not relaxed validity.
  Automatic repair is still gated on `Canonicalize(bad)` (`fix.go:192`) plus the unchanged reference and
  ownership guards.

### Isolation attestation

| Item | Value |
| --- | --- |
| Source checkout (read-only) | `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `main`, base `a7e1faa9257806f8b3be4398a6d59d7c3ebb4914` (dirty/untracked state captured) |
| Sandbox | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9EHSOP` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.9EHSOP/.git` (independent `--no-hardlinks` clone; no alternates; not a worktree or symlink) |
| Baseline commit | `2c81576561155f8ddb1ee7dae9f92f4793a1abe5` ("chore: capture isolated review baseline"; the only sandbox commit) |
| Captured source blob | `fc8801e81cf12dcb30476da6fc7c85d492bd4dd7` |
| Source fingerprint | `2253c145a8ec90d7df518e1b0d62e1ca2a8c0c61` |
| Deliverable | `planning/audits/6gkm2xz5s9qx-2026-10-14-task-correctness-detour-implementation-claude.md` |
| Scratch (GOCACHE, binaries, base export, spaces) | `/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/46c0d1a5-c9d5-499d-b76a-7c8a37975b68/scratchpad/review` |
| Restoration | Every mutation was restored and the probe test deleted. Before verification only this audit differs. |
| Final checks | The probe file is absent. The final `go test -race ./internal/core ./internal/store ./internal/cli -count=1` was green (2708/0). `git diff --check` was clean. The reviewed binary's `audit lint 6gkm2xz5s9qx` passed with 2 open findings, and planning `lint` exited 0. |
| Verify | Helper `verify` exited 0 with `baseline_commit=2c815765…`, `source_blob=fc8801e8…`, `source_fingerprint=2253c145…`, `deliverable_changed=true` and `transfer=pending`. It was rerun after this row was written, immediately before transfer; that result is in the handoff. |
| Transfer | Pending when this report was frozen. The helper's actual transfer output is in the handoff message, not asserted here. |

## Owner triage (2026-10-14)

Both findings hold up. L1 is fixed with a canonical resolution check on equal endpoint
IDs, retaining ambiguity before self-track validation. New core and real-CLI regressions
were red before the correction and green afterward; write/preview collisions now
return exit 13 with both candidates and no changed bytes. Ordinary self aliases remain
exit 11. The tracked rename/CAS interleaving is retained as a permanent real-store test.

M1 is independently reproduced and tracked by
[strict task mutation retries](../tasks/6gkmexz4vh23-bind-task-mutation-retries-to-strict-source-identities.md).
Five legacy paths silently changed the sibling after a rename; tracked was the passing
control. A second owner probe found a deeper limitation: delete the source, leave a
pre-existing sibling whose slug equals its ID, and even tracked's ID query falls back
to that sibling on retry. Thus the recommendation to resolve once needs strict
selected-identity lookup as well. The follow-up owns both layers; the current guide,
retry comment and task criterion explicitly retain that residual work.

The frozen reviewer verdict/evidence above refers to its captured snapshot, not the
post-triage implementation or a universal absence of defects. Both reports retain
`transfer=pending` from capture; successful transfer stdout was not available in this
conversation, so owner validation does not treat that frozen text as proof of delivery.
