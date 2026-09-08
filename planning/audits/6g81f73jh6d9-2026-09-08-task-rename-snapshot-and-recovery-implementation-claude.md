---
schema: 1
id: 6g81f73jh6d9
bucket: closed
area: task-rename-snapshot-and-recovery-implementation-claude
date: "2026-09-08"
updated_at: "2026-09-08"
---
# Audit: Task rename snapshot and recovery implementation — claude — 2026-09-08

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Adversarially review the implementation of the task `guard-renametask-against-stale-cascade-plans-and-concurrent-identity-duplication`. Do not merely confirm that tests pass. Try to demonstrate a stale plan, lost update, duplicate stable identity, clobbered target, unsafe retry recommendation, inaccurate durable-prefix receipt, adapter leak, or recovery state from which the documented advice cannot converge. Treat the implementation as suspect until each invariant survives a concrete hostile probe.

## Review target

Review the complete uncommitted implementation snapshot copied into your mandatory sandbox, especially:

- `planning/tasks/6g7wxs43g7nh-guard-renametask-against-stale-cascade-plans-and-concurrent-identity-duplication.md`
- the 2026-09-08 amendment in `planning/adrs/0003-stable-key-id-addressed-storage.md`
- the `RenameTask` architecture contract in `docs/ARCHITECTURE.md`
- `internal/store/rename.go` and `internal/store/rename_test.go`
- `internal/core/task_rename.go`, `internal/core/service_task.go`, and the affected Store port
- `internal/cli/task.go`, `internal/cli/exit.go`, and `internal/cli/task_rename_test.go`
- `internal/wire/task_rename.go`, the error-envelope/schema changes, schema comments, and CLI goldens

Build a consumer inventory for the changed Store and Service signatures and for `TaskRenameReceipt` / `TaskRenameFailure`. Verify every production caller, test double, CLI adapter, prospective TUI-facing boundary, wire mapper, schema registration, and error-classification path. Separate existing implemented consumers from future capabilities mentioned only in planning.

## Intended contract to challenge

A real rename captures the selected source version before waiting, then takes the canonical planning-root repository guard. Once guarded, it rechecks target availability and the original source, plans the link cascade from fresh current bytes, content-CAS checks each cascade document immediately before atomic replacement, content-CAS checks the source before destination materialization, creates a moved destination without clobbering, content-CAS checks the old source again immediately before removal, and releases the guard. Two cooperating concurrent renames cannot both commit; cooperating guarded edits are incorporated or rejected without loss.

The physical order is cascade documents, destination content, then old-source deletion. Failure before destination creation may leave a convergent link-rewrite prefix and advises rerunning the same rename by stable id after resolving the cause. Failure after destination creation but before source removal deliberately retains both files and requires inspection rather than a blind retry. Failure while releasing the guard after completion reports the rename as already complete. Store and core preserve planned/applied documents and links plus destination/source milestones; CLI JSON failures expose the adapter-neutral receipt. Dry-run is a non-durable preview and claims no reservation. Raw editors remain outside the advisory lock, but CAS/O_EXCL checks must not make stronger claims than they can prove.

## Mandatory evidence floor

Run focused tests with `-race` and repeat the coordinated concurrency cases enough to expose nondeterminism. Run the full validation that the implementation claims. Inspect the repository guard, path resolution, CAS, atomic create/replace, error classification, and JSON-schema machinery rather than assuming their names imply the required behavior.

For every new concurrency or recovery regression test, temporarily kill the exact production guard it claims to pin and prove that the named test fails for the intended reason. At minimum challenge: capture-before-lock source validation; the post-lock target recheck; per-cascade CAS; exclusive destination creation; pre-destination source validation; final pre-removal source CAS; and committed/unlock failure wrapping. Restore the sandbox baseline after every mutation. A mutation killed only by an unrelated compile failure, broad fixture failure, or another invariant is not evidence.

Exercise the new optional `error.task_rename` wire branch with non-default values and validate it against the generated JSON Schema. Verify the schema minor bump propagated to all emitted envelopes and that non-rename failures do not acquire a misleading rename receipt. Demonstrate the human and machine error behavior for at least one retryable durable prefix and the destination-written/source-retained state; do not accept prose inspection alone when a structured assertion is feasible.

## Required hostile angles

1. Coordinate two FS instances and, if practical, two processes renaming the same source to both different and identical targets. Challenge stable-id uniqueness, loser attribution, guard identity, and any stale target/source window.
2. Coordinate guarded task body/field changes to the source and to inbound-link documents at each meaningful point around snapshot, lock acquisition, plan, CAS, write, destination create, and source removal. A successful cooperating write must not disappear.
3. Simulate raw creation, modification, deletion, and relocation of the source, destination, and early/late cascade files. Identify precisely which windows are detected and which cannot be protected; flag documentation that overstates the boundary.
4. Force failures before the first write, after one or several cascade writes, during destination creation, immediately after destination write, during source CAS/removal, and during unlock. Compare physical disk state with every receipt field, error class, exit code, remedy, and safe next action.
5. Actually execute or faithfully simulate each recommended recovery. Prove a retryable prefix converges without double-rewriting or loss. Prove the inspection-only state does not recommend an operation that resolution ambiguity makes impossible or that could delete a newer raw edit.
6. Challenge the decision to write links before the destination exists. Consider interruption, process death, malformed destination files, links in the rename source itself, same-slug/title-only renames, duplicate or ambiguous identities, and a second competing rename after a partial prefix.
7. Audit deterministic planning and accounting: walk order, planned/applied document counts, link counts, source/destination milestones, exact-target collision behavior, self/reference/fenced links, unknown Markdown documents, modes/permissions, and same-path behavior.
8. Examine architectural fit. Determine whether core receipt/failure types are presentation-neutral and sufficient for CLI, TUI, and a future web/backend adapter, while filesystem-specific locking/CAS stays in the adapter. Flag both leaked filesystem assumptions and insufficient portable recovery information.
9. Look for systemic duplication with task lifecycle, graph mutation/repair, Thread apply, and entity creation. Do not demand an out-of-scope universal transaction framework without demonstrated value, but identify a concrete shared primitive if this implementation repeats a hazardous pattern inconsistently.
10. Perform a second pass focused on paths happy-path tests obscure: deferred named returns, joined unlock errors, an uncommitted conflict plus unlock failure, partial write counts when the source has self-links, raw edits after CAS but before replace/remove, crashes during exclusive create, symlinks or non-regular Markdown entries, and malformed/ambiguous task trees.

Reject speculative findings. For every finding, provide a minimal reproduction or exact code path, impact, why current tests miss it, and the smallest sound correction. If an issue is genuinely adjacent or out of scope, still record it and say why it belongs in a follow-up rather than silently omitting it.

## Validation and restoration

Perform every mutation and generated-artifact check only inside the isolated sandbox. Before delivery, restore all probes and generated drift to the sandbox baseline so the assigned audit is the only changed file. Run `git diff --check`, focused race tests, full `go test -race ./...`, planning lint, Go lint, module-tidy check, CLI docs drift check, schema-comment freshness, schema validation, and golden checks where the environment supports them. Report skipped checks and the exact blocker; do not call them passing.

## Deliverable

Update only your assigned audit. Keep this brief intact. Add findings using the repository grammar, with every new finding left `open` for implementation-owner triage. Include severity, component, effort, urgency, exact evidence, a concrete recommendation, the mutation/probe performed, and affected acceptance criteria. If no finding survives both passes, write a substantive no-findings report listing the invariants challenged, exact mutations killed, commands run, residual risks, and why those risks are acceptable or already scoped.

## Reviewer report

Both passes are complete. The guarded ordering, the CAS boundaries, the recovery contract, and the
new wire branch behave as the brief describes: every ordering claim I challenged with a hostile probe
held, both documented recoveries execute and converge end-to-end, and the schema bump is uniform.
Three findings survive, all left `open`. The primary one is a review-effectiveness defect rather than
a behavioral defect: three of the seven guards the evidence floor names survive deletion with
`go test -race ./...` fully green, so the suite's apparent coverage of the fail-closed boundary is
partly illusory.

### Isolation attestation

Helper output, verbatim (`scripts/isolated-review-workspace.sh`):

```
sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/7fc333dd-76dc-454d-a643-f26cffc927ed/scratchpad/isolated-review.C7GbaL
git_dir=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/7fc333dd-76dc-454d-a643-f26cffc927ed/scratchpad/isolated-review.C7GbaL/.git
baseline_commit=a36158e8791276da7cea2c7422f7e2e9933a70d3
deliverable=planning/audits/6g81f73jh6d9-2026-09-08-task-rename-snapshot-and-recovery-implementation-claude.md
source_blob=277cd956843f680f6ffae0722e6eedd9e967f031
source_fingerprint=54434e430af528823cffe718e78d126dd9b468c5
```

`--no-hardlinks` clone, in-tree `.git` (no alternates file, exactly one worktree, no `core.worktree`),
baseline `a36158e` over source `HEAD 3b77a27`. Every build, test, mutation, generator run, scratch
fixture, and edit below happened inside that workspace; the only operations performed in the handoff
checkout were reading this brief and the initial copy. Transfer result is recorded at the end of this
report.

One side effect crossed the sandbox boundary and was reverted: `tskflwctl init` in a scratch fixture
registered an entry point in the *user-global* `~/.config/tskflwctl/spaces.toml`. It was removed with
`tskflwctl space forget zz-probe-repo` (confirmed gone from `space list`); no repository state was
touched. Flagging it because the helper cannot isolate user-global config.

### Contract and consumer inventory

Every name below was verified in the sandbox at the cited path and line; nothing here is taken from a
planning document.

**Changed Store port** — `core.TaskStore.RenameTask` (`internal/core/store.go:52`), now returning
`TaskRenameMutationResult`. Implementations: `store.FS.RenameTask` (`internal/store/rename.go:31`) and
the test double `nopStore` (`internal/core/service_epic_test.go:43`, updated). A search for the port's
other methods (`TransformTaskBody`) confirms `conflictStore` (`internal/core/occ_retry_test.go:47`)
implements a *different, narrower* interface and needed no change. No other implementation exists.

**Changed Service signature** — `core.Service.RenameTask` (`internal/core/service_task.go:678`),
returning `TaskRenameReceipt`. Sole caller: `cli.newTaskRenameCmd`'s `RunE`
(`internal/cli/task.go:709`).

**`TaskRenameReceipt` / `TaskRenameFailure`** (`internal/core/task_rename.go:30,49`). Producers:
`taskRenameReceipt` (`:71`) and `service_task.go:682`. Consumers: `cli/task.go:711-713` (classifies
into `taskRenameCommandFailure`, `internal/cli/task.go:20`), `cli/exit.go:99-103` (`errors.As` →
`wire.ToTaskRenameRecoveryJSON`), and `internal/wire/task_rename.go:26`. Wire registration:
`ErrorItem.TaskRename` (`internal/wire/envelopes.go:1046`, `omitempty`), schema comment
(`internal/wire/schema_comments.json:208`), `SchemaVersion = "1.62"` (`internal/wire/wire.go:259`).

**Not shipped — planning only.** There is **no** TUI rename path: `grep -rn "Rename" internal/tui/`
returns one unrelated comment about fsnotify write/rename/chmod coalescing (`internal/tui/watch.go:17`).
The "prospective TUI-facing boundary" and the "future web/backend adapter" in ADR-0003's amendment are
prospective consumers only; the CLI is the sole adapter today. Likewise `render` gained no rename
receipt renderer — the success path still reuses `render.TaskMutationJSON`.

### Mutation and probe results

Every mutation was applied inside the sandbox and restored with `git checkout --` before the next one;
`git status --porcelain` was empty between probes. "Full suite" means `go test -race ./...`.

| # | Guard (site) | Mutation | Killed by | Verdict |
| --- | --- | --- | --- | --- |
| M1 | capture-before-lock source validation (`rename.go:74`) | call removed | **nothing — full suite exit 0** | **unpinned** |
| M2a | post-lock target recheck (`rename.go:68`) | call removed | `TestRenameTask_RechecksTargetAfterRepositoryGuard` | pinned |
| M2b | + planner target check (`rename.go:215`) | both removed (coordinated) | `…RechecksTargetAfterRepositoryGuard`, `TestRenameTask_TargetCollisionRefused` | pinned |
| M3 | per-cascade CAS (`rename.go:89`) | call removed | `TestRenameTask_CASCatchesRawCascadeDocumentEdit` | pinned |
| M4 | exclusive destination create (`rename.go:118`) | `createFileAtomic` → `writeFileAtomic` | **nothing — full suite exit 0** | **unpinned** |
| M5 | pre-destination source validation (`rename.go:111`) | call removed | **nothing — full suite exit 0** | **unpinned** |
| M6 | final pre-removal source CAS (`rename.go:143`) | call removed | `TestRenameTask_SourceRemovalCASCatchesRawEdit` | pinned |
| M7a | committed-failure wrapping (`service_task.go:681`) | branch removed | 4 tests incl. `…PartialCascadeReceiptIsResumable` | pinned |
| M7b | unlock-error join (`rename.go:61`) | `errors.Join` result discarded, compile-clean | `TestRenameTask_CompleteUnlockFailureIsNotRetryable` | pinned |

M2a and M2b are recorded as coordinated mutations: removing the post-lock recheck alone leaves
`prepareTaskRename`'s own `ensureTaskRenameTargetAvailable` in place, so both call sites were killed
together to reach the O_EXCL backstop. In every target-collision case the killer was
`verifyUnchanged`'s `ErrAmbiguous` (the colliding file necessarily shares the stable id), which is
exactly what the post-lock recheck's comment says it exists to pre-empt — the guard is pinned, and it
is pinned for the documented reason. M7b's first attempt failed to compile (unused `errors` import)
and was redone compile-clean, per the evidence floor.

Probes that produced the evidence for the findings and the falsifications, all deterministic:

- **PROBE A** — cooperating `SetFields` on the *rename source* lands while the rename waits at
  `testHookBeforeTaskRenameLock`.
- **PROBE B** — raw editor rewrites the source between the cascade prefix and destination
  materialization (via `testHookBeforeTaskRenameWrite(source.path)`).
- **PROBE C** — dangling symlink occupying the destination path.
- **PROBE D/D2/E/F** — real `tskflwctl` binary against a real scaffolded planning repository: human
  and `--json` failure output, both documented recoveries executed, and file modes measured.
- **PROBE G** — `Remedy` on a successful rename and on a dry run.
- **PROBE H** — uncommitted conflict combined with an unlock failure.
- **PROBE I** — planned/applied accounting when the source carries self-links.

### Findings

#### M1. Three of the seven named rename guards survive deletion with the full suite green · **Status:** fixed

**File:** internal/store/rename.go:74 | **Component:** store/rename
**Effort:** M · **Urgency:** soon
**Severity:** medium · **Affected acceptance criteria:** "Partial multi-file failure semantics and
operator recovery diagnostics are explicit and tested" and "Race-enabled focused tests and the full
validation suite pass".

The new suite pins the target recheck, the per-cascade CAS, the final pre-removal CAS, and both halves
of committed-failure reporting. It does **not** pin three guards the implementation and ADR-0003's
2026-09-08 amendment both advertise. Each was removed independently; in each case
`go test -race ./...` exited **0** with no failures:

1. **Capture-before-lock source validation** — `verifyUnchanged(...)` at `internal/store/rename.go:74`,
   the line whose comment says "The pre-lock source version is the intent boundary".
2. **Pre-destination source validation** — `verifyUnchanged(...)` at `internal/store/rename.go:111`,
   "Recheck the source immediately before materializing the destination".
3. **Exclusive destination creation** — `createFileAtomic` at `internal/store/rename.go:118`, the
   ADR's "The destination is created with a no-clobber precondition".

None is dead code; each was shown load-bearing by a probe that changes observable outcome:

**(1) Failure scenario, reproduced.** A cooperating guarded `SetFields("6fjangd7kva1", {"priority":
"low"})` lands on the rename source while the rename waits for the repository guard (PROBE A).
Baseline: `Committed:false AppliedDocuments:0`, a clean conflict before any write. With the guard
removed: `Committed:true AppliedDocuments:1 AppliedLinks:1` — the doomed rename first wrote a durable,
pointless cascade prefix into task B, then failed anyway at the *next* CAS. The path re-resolve alone
cannot catch this (the source path is unchanged) and the plan walk simply re-reads the newer bytes, so
this guard is the only thing standing between a stale intent and a needless durable prefix.

**(2) Failure scenario, reproduced.** A raw editor rewrites the source after the cascade prefix commits
and before the destination is materialized (PROBE B). Baseline: conflict, destination absent, source
retained, raw edit intact. With the guard removed: the destination **is** created from the stale
planned bytes (`# New title` over the pre-edit body — the raw editor's content absent), leaving
`DestinationWritten:true SourceRemoved:false` — the two-owner inspection state, entered with a
destination that silently dropped the newer edit. Only the *final* pre-removal CAS then prevents the
source deletion, so the raw bytes survive on disk but the tool has manufactured the ambiguity the
whole design exists to avoid.

**(3) Failure scenario, reproduced.** A dangling symlink occupies the destination path (PROBE C).
`os.Stat` follows it and reports `ENOENT`, so `ensureTaskRenameTargetAvailable`
(`internal/store/rename.go:274-286`) calls the target free; `markdownDoc`
(`internal/store/resolve.go:23`) requires a *regular* file, so it never becomes a resolution candidate
and never trips `verifyUnchanged`'s ambiguity check. O_EXCL is the only remaining guard. Baseline:
`conflict: target filename already exists`, symlink and source both intact. With `writeFileAtomic`
substituted: `err=<nil>, Complete:true, SourceRemoved:true` — the symlink is replaced and the rename
reports full success on a path no check ever authorized.

**Why current tests miss it.** For (1) and (2) the seam exists but is unused: `rename_test.go` calls
`testHookBeforeTaskRenameLock` only to sequence two renames or to plant a target file, and
`testHookBeforeTaskRenameWrite` is filtered to `path == bPath`
(`internal/store/rename_test.go:350-357`), so the hook's `source.path` invocation at
`internal/store/rename.go:107-109` is never exercised. For (3) the seam is *missing*: there is no hook
between the pre-destination CAS (`:111`) and the create (`:118`), which is precisely why no ordinary
race test can reach the O_EXCL branch — every reachable target collision is intercepted earlier by
`ensureTaskRenameTargetAvailable` or by `verifyUnchanged`'s `ErrAmbiguous`. That is the systemic half
of this finding: the hook set covers the boundaries that were easy to instrument, not the boundaries
that carry the residual risk, so the suite reads as complete while three fail-closed claims are
unverified.

**Smallest sound correction.** Three tests, no production change:
(a) a cooperating `SetFields` on the *source* during the pre-lock wait, asserting `Committed == false`
(PROBE A is the ready-made body);
(b) a raw source rewrite gated on `path == source.path` inside `testHookBeforeTaskRenameWrite`,
asserting the destination is absent (PROBE B);
(c) a dangling-symlink destination asserting `ErrConflict` and that the symlink survives (PROBE C) —
this one needs no new hook and closes the O_EXCL gap without adding a seam.

**Resolution:** Added deterministic regressions for source changes while
waiting, raw source edits before destination creation, post-CAS target creation,
and dangling-symlink targets; mutation checks prove each exact guard is
load-bearing.

#### M2. `task rename` silently widens the renamed task's file mode from 0600 to 0644 · **Status:** fixed

**File:** internal/store/rename.go:118 | **Component:** store/rename
**Effort:** XS · **Urgency:** soon
**Severity:** medium · **Affected acceptance criteria:** none directly; this is hostile angle 7
("modes/permissions"), and it contradicts an invariant the repository states in prose.

`writeFileAtomic` documents and enforces a permission-preservation rule
(`internal/store/atomic.go:41-50`): *"a user (or synced/encrypted setup) that chmod'd a task to 0600
must not have it silently widened to 0644 on the next edit."* The rename's destination does not go
through that path — `createFileAtomic(plan.newPath, plan.renamedContent, 0o644)`
(`internal/store/rename.go:118`) passes a literal and never consults the source's mode — so a rename is
the one mutation that breaks the rule.

**Failure scenario (reproduced end-to-end with the real binary, PROBE E).** In a scaffolded planning
repo, `chmod 600` on both a task and a document that links to it:

```
-rw-------  6g81j5p3a5he-new-title.md
-rw-------  6g81j5p6vtka-linker.md
tskflwctl task set 6g81j5p3a5he --priority high   → -rw-------   (preserved)
tskflwctl task rename 6g81j5p3a5he "Third Title"
-rw-r--r--  6g81j5p3a5he-third-title.md           ← widened
-rw-------  6g81j5p6vtka-linker.md                 (cascade replace preserved it)
```

The inconsistency is visible within the single command: the cascade document keeps 0600 because it goes
through `writeFileAtomic`, while the renamed task loses it. Impact is a silent, permanent
loosening of a deliberately restricted file — small blast radius (local markdown) but it defeats a
protection the tool explicitly promises, and the user gets no signal.

**Why current tests miss it.** No rename test asserts a file mode; `renameRepo`
(`internal/store/rename_test.go:16-35`) writes every fixture 0o644, so source and destination modes
agree by construction and the reset is invisible.

**Smallest sound correction.** Stat the source in the real-write branch and pass its `Mode().Perm()`
to `createFileAtomic` instead of the `0o644` literal (the source is already open to CAS, so no extra
read is required). Note `os.OpenFile`'s perm is masked by umask, so an exact-preservation fix needs the
same post-create `os.Chmod` that `stageTemp` (`internal/store/atomic.go:33`) already performs; passing
the source mode alone is a strict improvement either way, since umask can only narrow. One test
asserting the destination mode equals the source mode.

**Resolution:** Moved destinations now use the source effective permission bits
and an exact-mode exclusive create; a 0600 regression test proves the rename
does not widen access.

#### L1. A successful rename hands adapters failure-shaped recovery prose · **Status:** fixed

**File:** internal/core/task_rename.go:80 | **Component:** core/task-rename
**Effort:** XS · **Urgency:** eventually
**Severity:** low · **Affected acceptance criteria:** none directly; hostile angle 8
(presentation-neutral core types sufficient for CLI, TUI, and a future web adapter).

`taskRenameReceipt`'s first `switch` arm is `case result.Complete`, and `Complete` is true on **every**
successful rename. So `core.Service.RenameTask` returns, on a clean success:

```
PROBE-G success   err=<nil> complete=true remedy="inspect the renamed task by stable id before retrying; the planned rename is already complete"
PROBE-G dry-run   err=<nil> complete=false remedy=""
```

An adapter that renders `receipt.Remedy` whenever it is non-empty — the natural reading of an
adapter-neutral recovery field, and the reading the prospective TUI/web adapters in ADR-0003 would take
— prints "before retrying" recovery advice after a rename that succeeded. The sibling receipts do not
do this: `taskLifecycleRemedy` (`internal/core/service_task.go:400-412`) returns `""` unless something
actually needs repair, and `threadMutationReceipt` (`internal/core/service_thread.go:190-205`) sets
only success-appropriate prose. The CLI is insulated today because it discards the receipt on success
(`internal/cli/task.go:718`), which is why nothing fails.

Same arm, smaller: `taskRenameResult` (`internal/store/rename.go:198-206`) hardcodes `Changed: true`,
so `changed` is a constant in every emitted envelope and cannot report a genuine no-op re-title.

**Smallest sound correction.** Reserve the "already complete" remedy for the case it was written for —
`result.Complete && err != nil`, i.e. the completed-then-unlock-failed state — by giving
`taskRenameReceipt` the operation's error (or by setting `Remedy` only along the failure path in
`Service.RenameTask`). `TestRenameTask_CompleteUnlockFailureIsNotRetryable` already pins the behavior
that must survive; add an assertion that a successful rename's `Remedy` is empty.

**Resolution:** Recovery remedies are now derived only on committed failures;
successful receipts carry no failure prose, and exact title/slug no-ops report
no planned or committed write.

#### L2. `task rename --json` success discards the receipt the human path prints · **Status:** tracked by 6g81npee5kv2

**File:** internal/cli/task.go:718 | **Component:** cli/task
**Effort:** S · **Urgency:** eventually
**Severity:** low · **Affected acceptance criteria:** none. Not a blocker for this task.

Recorded here because it is adjacent rather than silently omitted: it is a pre-existing asymmetry in
shape, but this change is the one that built an adapter-neutral receipt carrying exactly the missing
numbers, so it is newly cheap to fix and squarely inside hostile angle 8.

The human path prints `✔ renamed to third-title (1 inbound link(s) repointed)`; `--json` emits a plain
`TaskMutationEnvelope`:

```
tskflwctl --json task rename 6g81j5p6vtka "Linker Renamed"
{"schema_version":"1.62","dry_run":false,"task":{…},"workspace":{…}}
```

No `from_slug`, no `planned_documents`, no `applied_links`. An agent driving `--json` therefore gets
*less* information on success than a human does, and less than it gets on failure — against the rule
stated at `internal/cli/exit.go:63-66` ("an agent driving --json must never have to parse prose to
learn why a command failed"), whose spirit the success path now inverts.

**Smallest sound correction.** A follow-up task: emit a `TaskRenameJSON` success envelope (or extend
`TaskMutationEnvelope` with the rename receipt) behind a schema minor bump, reusing
`wire.ToTaskRenameRecoveryJSON`'s field set. Out of scope for the guard-hardening task under review.

**Resolution:** A dedicated follow-up will design and version a rename-specific
success and dry-run JSON envelope without widening this concurrency-hardening
patch.

### Hostile hypotheses falsified

Each of these was a specific attempt to break the contract; each failed, and the mechanism that
defeated it is named.

- **Two concurrent renames to different targets both commit.** Falsified. `TestRenameTask_Concurrent…`
  passes at `-count=20` under `-race`; the loser is rejected by `verifyUnchanged` (`:74`) because
  `resolvePath(id)` now returns the winner's path. Exactly one stable-id owner remains.
- **Two concurrent renames to the *same* target both commit.** Falsified. The loser is rejected by the
  post-lock `ensureTaskRenameTargetAvailable` (`:68`) with the precise
  `conflict: target filename already exists`, ahead of the ambiguity check — which is what that
  ordering is for.
- **A cooperating guarded write to an inbound-link document disappears.** Falsified. The plan is
  compiled *inside* the guard (`:78`), so a `SetFields` that lands during the wait is incorporated:
  `TestRenameTask_ReplansAfterCooperatingCascadeDocumentWrite` reads back `priority: low` *and* the
  repointed link. Cooperating writers cannot interleave inside the guard at all.
- **An uncommitted conflict plus an unlock failure loses one of the two causes, or acquires a
  misleading receipt.** Falsified (PROBE H). The post-lock target recheck fails while
  `testHookRepositoryUnlockError` fires; the deferred `errors.Join` (`:61`) produces both lines,
  `domain.Classify` still returns `ClassConflict` (exit 14), `errors.As(..., *TaskRenameFailure)` is
  false, and no `task_rename` receipt is emitted. Correct on every axis.
- **The retryable prefix double-rewrites or loses a link.** Falsified, executed rather than argued.
  Real binary: rename fails with `committed 1 of 2 … applied_links 1`; I removed the blocking symlink
  and reran by stable id; result is exactly one occurrence of the new link, `# New Title` in the
  destination, one stable-id owner, `lint` clean, exit 0.
- **The inspection-only state recommends something resolution ambiguity makes impossible.** Falsified.
  In the destination-written/source-retained state, `lint` reports `duplicate stable task id … no
  source is uniquely authoritative`, and `task show`/`task rename` by id return `ErrAmbiguous` (exit
  13) *naming both files*. The receipt's remedy is a filesystem inspection plus one deletion, not a
  tool operation — so it stays executable. I ran it: removing the retained old source restored a clean
  `lint` and a resolving `task show`.
- **The inspection-only state could recommend deleting a newer raw edit.** Falsified. The remedy is
  conditional ("if the destination is correct"), and the pre-removal CAS (`:143`) is what produced the
  state in the raw-edit case, so the newer bytes are the ones retained.
- **Writing links before the destination exists strands the tree with no signal.** Falsified.
  Reproduced the interrupted prefix (cascade repointed, destination absent):
  `lint --links` reports `body link to missing file: 6g81j5p3a5he-fifth-title.md`. The alternative
  ordering would be worse — destination-first would enter the two-owner state on *every* rename.
- **Self-links in the rename source corrupt the accounting.** Falsified (PROBE I). Source with two
  self-links plus one inbound link: `plannedLinks=3 appliedLinks=3 plannedDocs=2 appliedDocs=2`; the
  destination's self-links are repointed and no `…-old.md` reference survives.
- **The `error.task_rename` branch does not validate at non-default values.** Falsified. Compiled
  `schema --json-schema` with `santhosh-tekuri/jsonschema/v6` against `#/$defs/ErrorEnvelope` and
  validated (a) the two *real* captured CLI envelopes, (b) a synthetic payload with every boolean at
  its non-default value (`dry_run`, `complete`, `destination_written`, `source_removed` all true,
  non-zero counts, full workspace), and (c) a bare non-rename failure. All four VALID; the schema
  marks all 15 properties `required` with `additionalProperties: false`.
- **The minor bump missed an envelope, or a non-rename failure acquires a rename receipt.** Falsified.
  `"schema_version":"1.62"` is the only value present across all 33 goldens; a real
  `task show does-not-exist --json` emits `{"code":"not-found", …}` with no `task_rename` key.
- **The `errors.Join` unlock path hides the durable outcome.** Falsified. Completed-then-unlock-failed
  yields `Committed:true Complete:true` and the "already complete" remedy, with the old path gone
  (M7a/M7b both killed by the named test).
- **A leaked test hook masks defects.** Falsified, and I record it as a *non*-finding.
  `TestRenameTask_PartialCascadeReceiptIsResumable` sets `testHookAfterTaskRenameWrite`
  (`internal/store/rename_test.go:381`) and clears it by assignment at `:407` rather than by `defer`,
  unlike all six sibling tests. I forced an early `t.Fatalf` in that test to make the hook leak: no
  later test changed behavior, because the closure compares against a `t.TempDir()` absolute path that
  no subsequent test can produce. It is a style inconsistency, not a defect, and I am not raising it as
  a finding.

### Residual risks

Accepted, with the reason each is acceptable or already scoped.

- **The window between the pre-destination CAS (`:111`) and `createFileAtomic` (`:118`) is
  unobservable.** A raw editor that creates a *regular* id-led file there produces the two-owner state,
  after which a rerun by stable id returns `ErrAmbiguous` rather than converging — so the emitted
  "rerun by stable id" remedy would not apply. Acceptable: the window is microseconds, it requires a
  raw write to an exact id-led path the operator does not yet know, and exit 13 names both colliding
  files, so nobody is stranded. Worth a sentence in ADR-0003 rather than a guard.
- **Dry run takes no guard and performs no CAS**, so a concurrent rename can make its preview stale.
  Documented as a non-durable preview and confirmed to claim no reservation
  (`DryRun:true Committed:false AppliedDocuments:0 Remedy:""`). Correct as designed.
- **A process death during the cascade prefix leaves danglers with no automatic signal.** Detectable
  only via the opt-in `lint --links`, verified above. Reasonable given that `--links` is opt-in by
  design ("a tree can carry pre-existing danglers").
- **Every rename walks the whole planning tree reading every `.md`.** Correct and deterministic
  (`filepath.WalkDir` is lexical), but O(tree) per rename inside the repository guard. Not a defect at
  this repository's size; worth remembering if the guard ever becomes contended.
- **Six sibling `*Receipt`/`*Failure` pairs in `core` and seven `*CommandFailure` structs in `cli` now
  repeat the same shape** (`internal/cli/{task,moves,thread,thread_apply,task_dependency,task_dependency_repair}.go`).
  I looked for the inconsistency hostile angle 9 asks about and found the rename implementation
  *conforms* to the established pattern rather than diverging from it — L1 is the one place it
  diverges, and it is a one-line fix. I am deliberately not proposing a shared transaction primitive:
  there is no demonstrated defect caused by the duplication, and the brief rightly rules that out
  absent one.

### Commands

All run inside the sandbox, at baseline `a36158e` unless a mutation is named.

```
go build ./...                                            OK
go test -race ./...                                       PASS (24 packages)
go test -race -count=20 -run TestRenameTask ./internal/store/   PASS
golangci-lint run ./...                                   0 issues
go mod tidy -diff                                         clean
just docs-check                                           clean (docgen + git diff --exit-code docs/cli)
go run ./internal/tools/schemacomments                    260 comments, no drift
go run ./cmd/tskflwctl lint                               all planning entities pass
go test ./internal/cli/...                                PASS (CLI goldens)
git diff --check                                          clean
```

Schema validation was run through a throwaway program in the sandbox compiling
`schema --json-schema` with `santhosh-tekuri/jsonschema/v6` (the same library
`internal/wire/envelopes_test.go:9` uses) against `#/$defs/ErrorEnvelope`; it and the scratch planning
fixture were deleted before delivery. Nothing was skipped: every check named in the brief's
"Validation and restoration" section ran and passed in this environment.

All probes and generated drift were restored to the baseline commit; `git status --porcelain` is empty
apart from this audit.
