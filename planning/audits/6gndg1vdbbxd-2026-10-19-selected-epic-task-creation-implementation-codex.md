---
schema: 1
id: 6gndg1vdbbxd
bucket: closed
area: selected-epic-task-creation-implementation-codex
date: "2026-10-19"
updated_at: "2026-10-20"
---
# Audit: Selected-epic task creation implementation — codex — 2026-10-19

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

Challenge the selected-epic task-creation foundation before the TUI form consumes it.
This is not a TUI implementation or a request to build a universal creation API.

Assigned primary lenses:
- Codex: adapter-neutral API/composition, selected identity, shared preparation and compatibility.
- Antigravity: guarded parent validation, authorization after waits, no-clobber and committed/finalization recovery.
Each reviewer independently cross-checks one important invariant from the other lens.
Do not read the sibling report before finalizing your own. No finding quota.

## Review target

Source checkout: /Users/andyeschbacher/git/andy-esch/taskflow.
Base: 5f62e684238209d9a8f111003623cbcc563f9552.
The implementation is UNCOMMITTED on main: capture tracked edits, untracked new
Go files, docs and planning. Reviewing HEAD alone misses this feature.

Implementing task: planning/tasks/6gmy4c0x6e31-bind-task-creation-to-the-selected-epic-identity.md.
Design: planning/research/6gmrkq00348a-tui-authoring-interaction-checkpoint.md.
Owning contracts: docs/architecture/planning-data.md and persistence.md.

Verify these actual paths/symbols in your sandbox:
- internal/core/selected_task_creation.go: SelectedEpicTaskCreationStore,
  WithSelectedEpicTaskCreationStore, NewTaskForSelectedEpic.
- internal/core/service.go: optional aggregate discovery and validateSourceSets.
- internal/core/service_task.go: NewTask and shared prepareTaskCreation.
- internal/core/store.go: EpicMutationTarget and ordinary TaskStore.CreateTask.
- internal/store/create.go: CreateTaskForSelectedEpic, createTask, createEntityFile.
- internal/store/entity_target.go: resolveEpicMutationPath; resolve.go: resolveExactID.
- internal/store/lock.go: checkedWriteLock; epicstore.go: parseEpic.
- internal/core/selected_task_creation_test.go, source_set_test.go;
  internal/store/selected_task_creation_test.go, mutation_authorization_test.go.

Build a repository-wide consumer inventory: callers, composition roots, source-ID/reference semantics, template
consumers and receipts with rg. No production TUI/CLI caller has been switched to
the new API; test-invoked public Service/adapter paths are the reachable feature.
Do not claim a future form or unimplemented selected graph API already exists.

## Intended contract to challenge

1. This optional sibling port does not widen aggregate Store or require local paths.
   Split capabilities need stable, nonzero matching SourceSetID witnesses; explicit
   override wins and missing/typed-nil capability refuses without legacy fallback.
2. Target is an adapter-established exact epic source ID, not a prefix, declaration,
   Location or path. The full epic stem remains identity locally. Epic input is blank
   or exactly the selected ID; implicit substitution is forbidden.
3. Core shares ordinary creation defaults, input validation and scaffolds. Selected
   creation supports ready/next, explicitly rejects Start, and makes one attempt.
   Ordinary CLI matching/create-and-start and wire output remain unchanged.
4. Local parent resolution/parsing occurs inside the guard used for task creation.
   Reauthorize after waiting, before parent reads/create. Read-only/denied previews
   cannot create roots/files or read a malformed parent first.
5. Preserve task/Thread collision checks, no-clobber and exact durable receipts.
   A conflict or cleanup error cannot remint/replay or erase committed evidence.
6. Limits are explicit: dry-run is a non-reserving point-in-time check; raw writers
   retain a final race; later parent removal is not blocked. Same-ID delete/recreate
   is not detectable. No new permanent FK, persisted child IDs or schema is promised.

Criterion authoring, TUI form/session behavior, selected registry removal and other
graph/entity write APIs are separate tickets. Ordinary create's pre-existing
authorization timing is not silently changed by this selected-only entry point;
an adjacent defect still needs a concrete producer and scope classification.

## Mandatory evidence floor

Both: independently run the focused baseline and at least one valid-but-hostile
and one refused fixture through the public selected Service or adapter entry point.
Check durable bytes/modes/record identities, not just error text. Give exact commands
and results; baseline owner gates alone are not a clean review.

Codex must:
- Exercise standalone pathless composition with nil aggregate Store, a mismatched
  witness and an absent/typed-nil capability; confirm no reference-path fallback.
- Compare ordinary vs selected preparation using a non-default case (Next with
  description, explicit metadata or body/template), and inspect the Start refusal.
- Temporarily remove only the new selected-creation source-set validation row.
  Require the intended selected-epic case of
  TestNewServiceRejectsEveryMismatchedSplitCapability to fail; restore green.
- Cross-check one real-store stale-parent or finalization probe.

Antigravity must:
- Run the two deterministic queued-write tests with -race -count=20:
  TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard and
  TestSelectedEpicTaskCreationReauthorizesAfterWaiting. Explain which guard is held
  and why the authorizer signal proves the create is waiting; challenge that oracle.
- Temporarily replace resolveEpicMutationPath with ordinary ResolveEpicPath in
  the new adapter callback. Require
  TestSelectedEpicTaskCreationRefusesFallbackAndUnreadableParents (the prefix
  fixture) to fail, then restore green.
- Temporarily remove ONLY the in-guard reauthorization, preserving outer checks.
  Require TestSelectedEpicTaskCreationReauthorizesAfterWaiting to fail; restore green.
- Exercise a collision and a committed unlock error through the real Service;
  confirm one mint/attempt and the destination/Committed evidence. Cross-check
  capability/source-set binding rather than assuming the registry analogy proves it.

Mutations must compile. A compile error is invalid, not a killed mutant; name the
intended regression that fails and capture its restored pass. Keep a short ledger
(edit, invariant, test, observed result). Do not count unrelated failures as proof.

## Required hostile angles

Separate systemic pass after the local probes:
- Can shared prepareTaskCreation or createTask change legacy behavior, accidentally
  join the wrong parent/namespace, or make tests share the same mistaken oracle?
- Can a delayed permit, epic rename, missing/ambiguous parent, unreadable metadata,
  source declaration, symlink/type drift or direct low-level call bypass strict
  parent ownership? Distinguish actual reachable bugs from the documented raw-writer
  boundary and same-ID replacement limit.
- Does a no-op/preview/error path bypass policy or fabricate a path/identity? Does
  finalization failure survive both core and adapter boundaries without unsafe retry?
- Would another primary/secondary adapter need TUI, filesystem or broad Store types?
  Challenge source-set binding and unsupported capability behavior without demanding
  speculative framework abstractions or a change to existing identity promises.

A no-findings outcome must state which hypotheses were falsified versus unresolved.

## Validation and restoration

Use the injected isolated-review protocol BEFORE inspection or commands. Source is
read-only. No shared worktree/index, independent sibling report access, source Git
writes, pushes or manual report copy-back. The helper's sandbox baseline commit is
the only permitted commit; verify then guarded transfer the assigned audit only.

Use GOCACHE and GOLANGCI_LINT_CACHE in a reviewer-owned temporary directory OUTSIDE
the sandbox corpus. Do not write probe fixtures into the user's real planning space.
Restore temporary production/test mutations and remove only your scratch artifacts
before verification. On capture/permission/evidence failure, keep the sandbox and
report incomplete; never fall back to the live checkout.

Owner baseline: fresh just build; full just test (-race); just lint; planning/link
lint and diff checks passed. Reviewers may inherit those full-suite results but
must independently run focused hostile evidence:
GOCACHE=<review-cache> go test -race ./internal/core ./internal/store -run 'TestSelectedEpic|TestNewService|TestEveryFSMutation' -count=1
For repeat probes restrict -run to the two waiting tests and use -count=20.
Run broader affected tests if your finding/probe changes shared behavior.
A resource/permission failure is not equivalent to passing.

## Deliverable

Keep this brief; replace only the Reviewer report slot. Findings first, with the
injected audit finding grammar, all findings OPEN. For each: root cause, reachable
producer, consequence, reproduction, evidence, recommended direction and remaining
uncertainty. Distinguish product defects, coverage gaps and preferences; consolidate
shared roots rather than inflate severity or count.

Include verified boundary inventory, focused commands/results, mutation ledger and
a bounded surviving-hypothesis list. Preserve documentation limitations truthfully.
Report the final isolation verification attestation (transfer still pending there).
Your handoff must carry the transfer receipt; the owner can confirm it against the
retained sandbox. Do not claim ready if material probes or isolation did not finish.

## Reviewer report

## Findings

#### L1. Shared preparation parity can hide loss of explicit metadata · **Status:** fixed

**File:** internal/core/selected_task_creation_test.go:41 | **Component:** core
**Effort:** S · **Urgency:** eventually

Classification: demonstrated coverage gap in the shared-preparation regression, not an observed defect in the restored implementation.

Root cause: `internal/core/selected_task_creation_test.go:41-53` obtains `want` by calling ordinary `NewTask`, which uses the same `prepareTaskCreation` as selected creation (`internal/core/service_task.go:938,963-1031`). Although the fixture supplies explicit metadata at test line 37, it never checks those fields against independent expected values. A shared loss of caller intent therefore makes both sides equally wrong.

Reachable producer and consequence: both public Service methods accept `Priority: "high"`. A regression defaulting every prepared task to medium would silently discard that input for ordinary CLI and selected-capability consumers. This is a future-regression risk; baseline code correctly stores `p.Priority` at `internal/core/service_task.go:1012`.

Reproduction executed in this sandbox: change only `Priority: p.Priority` in `prepareTaskCreation` to `Priority: "medium"`, leaving validation and both entry points intact. The mutant compiled. These commands both exited 0:

```sh
go test -race ./internal/core -run '^TestSelectedEpicTaskCreationSharesDefaultsAndScaffolding$' -count=1 -v
go test -race ./internal/core ./internal/store ./internal/cli -skip '^TestReviewer' -count=1
```

The second command ran the full three affected package suites, excluding only the temporary reviewer probes. It passed core (1.621s), store (12.934s), and CLI (10.248s). A temporary independent public-entry-point probe, `TestReviewerSelectedIndependentPreparation`, required `Priority == "high"` alongside explicit Next/description/tier/autonomy/effort/date/template values and failed at runtime with actual medium versus expected high. Its ordinary case failed first; the selected parity case also used the mutated shared helper. No compile error or unrelated failure was counted as mutation evidence.

Restoration: restored `internal/core/service_task.go` byte-for-byte from the sandbox baseline; reran selected tests, the independent probes, and the full affected core/store/CLI plus wire/architecture suites successfully. No production fix or permanent test was added by the reviewer.

Remaining uncertainty: this demonstrates one shared-field blind spot, not that every possible shared regression escapes the whole repository suite. The complete repository suite was inherited from the owner, not rerun under this mutant. Do not infer a current data-loss defect or a need to split the production helper.

**Recommendation:** Keep the parity comparison and add independent expected-field and exact named-template assertions for non-default requests through both public creation entry points.

**Resolution:** Added TestTaskCreationPreservesIndependentExpectedValues
    alongside parity: both public entry points, write/preview, literal default
    and non-default task expectations, exact body override and named-template
    output. A Go overlay replacing shared Priority input with medium compiles
    and fails the explicit-metadata cases in both entry points; the unmodified
    implementation passes.

### Review result and scope

Completed the Codex evidence floor and a separate systemic pass against the captured uncommitted implementation. One open low-severity coverage finding above; no baseline product defect was demonstrated. This is review evidence, not a task-completion or release-readiness disposition. No finding was settled, no implementation was changed, and no sibling report was read.

The initial checkpoint contains the source's tracked and untracked implementation, not just base commit `5f62e684238209d9a8f111003623cbcc563f9552`. All implementation reads, builds, tests, mutations and report preparation were run from `/private/tmp/isolated-review.Oh18JF`; persistence probes used disposable t.TempDir roots, never the user's planning space. Final evidence runs additionally set TMPDIR to the sandbox's `.reviewer-runtime` directory so their fixture files stayed within the sandbox corpus. Temporary probe sources named below were reviewer-created, executed, and removed; they are not shipped regressions. The fresh CLI generated L1 in a disposable copy of this audit; its generated finding block was inserted unchanged into this report slot so the original brief and frontmatter remain byte-for-byte intact. The first disposable-copy CLI invocation refused because the fixture lacked `tasks/`; adding that required directory made generation succeed. No candidate row or follow-up task was created.

### Verified boundary and consumer inventory

Inventory commands used in the sandbox:

```sh
rg -n 'NewTaskForSelectedEpic|CreateTaskForSelectedEpic|SelectedEpicTaskCreationStore|prepareTaskCreation' --glob '*.go' .
rg -n '\.NewTask\(|\.CreateTask\(|NewService\(|WithTemplateSource|TemplateSource|type TaskCreationReceipt|type CreateFinalizationError' cmd internal --glob '*.go' --glob '!**/*_test.go'
rg -n 'NewBuiltinTemplateService|WithTemplateSource|\.templateBody|renderTemplate' internal --glob '*.go' --glob '!**/*_test.go'
git diff 5f62e684238209d9a8f111003623cbcc563f9552 --name-only -- internal/cli internal/wire cmd
```

The last command produced no paths. The first found the new public Service/FS methods, their composition/assertion, and tests; it found no production CLI/TUI consumer of selected creation.

| Boundary | Verified implementation/evidence |
| --- | --- |
| Optional API | `internal/core/selected_task_creation.go:23,29,46`: narrow semantic port, explicit option, public use case. No local paths, Bubble Tea, Cobra or filesystem types. `internal/core/store.go:435` aggregate Store is unchanged; ordinary `TaskStore.CreateTask` remains at line 63. The FS compile-time sibling-port assertion is `internal/store/fsstore.go:99`. |
| Discovery and binding | `internal/core/service.go:369,388,424`: aggregate discovery precedes explicit options; selected creation participates in validation. `internal/core/source_set.go:41-65` rejects missing, zero, unstable and mismatched witnesses. Witnesses are composition evidence, not permissions or persisted IDs. Nil options are ignored under existing option conventions; a missing selected capability refuses without calling legacy creation. |
| Composition roots | `internal/appwiring/wiring.go:58-74` obtains the FS through `workspacestore.NewPlanningStore`, then calls `core.NewService(fs)`. `internal/core/workspace.go:107-115` composes the workspace Service from supplied capabilities. `internal/workspacestore/planning.go:17-47` owns local identity-reader composition. This change adds no selected graph API, form, or session implementation. |
| Selected versus ordinary identity | `internal/core/selected_task_creation.go:47-67` requires a nonblank target, blank/exact Epic input, rejects Start, and forwards one prepared request. `internal/store/entity_target.go:52-64` uses exact candidate identity; `internal/store/resolve.go:298-315` implements `resolveExactID`. `internal/store/epicstore.go:273-288` parses bytes and derives identity from the complete filename stem, ignoring an `id:` declaration. Ordinary `NewTask` still uses `epicExists`/`canonicalEpic` (`internal/core/service_task.go:928-937`, `internal/core/service_epic.go:66-88`). |
| Shared preparation/templates | `internal/core/service_task.go:954-1031` owns defaults, validation, slug, ID/date, metadata and body rendering. `internal/core/template.go:11-30,76-78` supplies built-in templates through a port; `internal/core/service.go:187` allows explicit replacement. Template list/show and other entity creates use this same template source, as the rg inventory shows. Repo-local template layering remains future-looking commentary, not a capability established by this review. |
| Local commit and guard | `internal/store/create.go:188-210,245-262` reauthorizes and resolves/parses the parent in the create callback, before collision checks/write. `createEntityFile` at lines 70-115 obtains the guard for real writes and performs a non-reserving, unlocked preview. `internal/store/lock.go:51-66,114-137` uses a canonical-root process mutex followed by the platform lock. `internal/store/lock_unix.go:21-39` flocks the planning-root directory itself; there is no new lockfile path. |
| Namespace/no-clobber/receipts | `internal/store/create.go:251-262,297-306` retains task/Thread identity checks. `internal/store/atomic.go:96-125` uses exclusive creation. `internal/core/creation_receipt.go:9,18,35` defines cleanup error, optional destination evidence and kind-specific task receipt. Parent checks/finalization errors are returned unchanged by the selected Service call. |
| Existing consumers/wire | `internal/cli/task.go:167-180` still calls ordinary NewTask, handles lifecycle/committed errors and presents the receipt. `internal/wire/envelopes.go:711-737` retains created/recovery envelopes. New selected creation is exercised through public APIs by tests, not a shipped CLI/TUI form. No changed optional wire branch or repair advice exists in this patch, so those conditional evidence requirements are not applicable. |

Owning guides were compared with code: `docs/architecture/planning-data.md:69-84` and `docs/architecture/persistence.md:80-98`. They state whole-stem selection safety, one attempt, non-reserving previews, raw-writer/later-removal limits and retained receipts. The implementation task's checked criteria and design checkpoint are requirements/history, not additional executable evidence. No promise of enduring foreign keys, persisted child IDs or a new schema was inferred.

### Executed hostile evidence

All Go commands below used these reviewer-owned caches outside the sandbox corpus:

```sh
export GOCACHE=/private/tmp/selected-epic-codex-cache/go
export GOLANGCI_LINT_CACHE=/private/tmp/selected-epic-codex-cache/lint
export TMPDIR=/private/tmp/isolated-review.Oh18JF/.reviewer-runtime
go test -race ./internal/core ./internal/store -run 'TestSelectedEpic|TestNewService|TestEveryFSMutation' -count=1
```

Focused baseline passed: core 1.200s, store 1.292s. This includes the shipped Start/input refusals (`internal/core/selected_task_creation_test.go:76-106`), override/absence cases (lines 136-153), and the mutation-policy entry inventory (`internal/store/mutation_authorization_test.go:27-30,156-189`). Denied requests and previews did not make a missing planning root; malformed-parent read-only cases returned policy errors before parsing.

```sh
go test -race ./internal/core -run '^TestReviewerSelected' -count=1 -v
go test -race ./internal/store -run '^TestReviewerSelected' -count=1 -v
```

These were temporary independent probes, not committed fixtures:

- `TestReviewerSelectedIndependentPreparation` passed ordinary and nil-aggregate/pathless selected calls in write and preview modes against an explicit domain.Task literal: Next, description, high priority, tier 2, autonomy 4, one-day effort, two tags, fixed ID/date and safe title-derived slug. A named injected template had exact expected rendered bytes, literal `100%`, and an untouched unknown placeholder. Empty local receipt fields remained empty for the pathless capability. The expected task/body did not call shared preparation or derive their values from the other public entry point.
- `TestReviewerSelectedAbsentCapabilities` passed with nil aggregate and absent/typed-nil selected capability, returning validation and a zero receipt instead of panicking/falling back. `TestReviewerSelectedStandaloneWitnesses` passed with nil aggregate plus split reads and selected creation: mismatched, zero, missing and unstable selected witnesses all refused before use. Shipped discovery/explicit-override tests independently passed.
- `TestReviewerSelectedDurableHostile` passed real Service creation and preview under `202601010000-parent`, alongside a full-stem sibling and an impostor `id:` declaration in the selected parent. The created task retained the selected full stem and non-default Next metadata; its body suffix and serialized fields were checked from raw bytes, then `ReadTask` confirmed record source ID and epic reference. Parent/sibling bytes and 0640 modes were unchanged; the created task was a regular 0644 file at the exact planned/committed path. Preview made no tasks directory and reported no committed destination.
- `TestReviewerSelectedRefusedTypesAndCollision` passed ten write/preview cases: prefix, symlink parent, directory parent, malformed metadata, and an unreadable Thread owning the candidate task ID. Refusals returned the intended sentinel and zero receipt, with exact bytes, modes, links and directory entries unchanged; each validly prepared request minted once and slept zero times. Its filesystem oracle used independent Lstat/ReadFile/Readlink traversal rather than the shared test snapshot helper.
- `TestReviewerSelectedCommittedUnlock` passed through the real Service/adapter: an injected unlock conflict returned `CreateFinalizationError`, Committed=true, equal exact planned/committed destination, durable body bytes, 0644 regular-file mode, readable source identity and correct epic. ID mint count=1, release count=1, retry sleep count=0. The shipped collision/finalization test also passed.

```sh
go test -race ./internal/store -run '^TestSelectedEpicTaskCreation(ChecksParentAfterWaitingForGuard|ReauthorizesAfterWaiting)$' -count=20
```

Passed (store 1.189s): 20 executions of each queued-write regression. The tests hold `processRepositoryLock(root)` while creation reaches the second authorizer call in checkedWriteLock. The channel signal marks the pre-acquisition authorization point, not proof that the goroutine is already parked in Mutex.Lock; the held mutex nevertheless prevents its guarded parent read/commit until the test changes the parent or permit and releases it. Thus the oracle establishes the required ordering without relying on sleep. Parent rename refused without selecting the lookalike sibling; a revoked permit refused before effects. This cross-check covers both stale-parent and finalization invariants from the other review lens.

### Mutation ledger

| Temporary edit | Invariant/test | Compiler-valid observed result and restoration |
| --- | --- | --- |
| Remove only `{"selected-epic task creation", s.selectedTaskCreations}` from `validateSourceSets` | `go test -race ./internal/core -run '^TestNewServiceRejectsEveryMismatchedSplitCapability$/^selected-epic_creation$' -count=1 -v` | Failed at `source_set_test.go:93`: nonnil Service and nil error for a foreign selected capability. This was the intended test failure, not a compiler failure. Restored service.go from the sandbox baseline; identical command passed (1.122s). |
| In shared preparation only, replace `Priority: p.Priority` with `Priority: "medium"` | New parity test, full affected package suites, independent expected-value probe | New parity test passed; full core/store/CLI suites passed with only reviewer probes excluded; independent preparation probe failed on medium versus high. L1 records this surviving mutant and its scope. Restored service_task.go from the sandbox baseline; independent/selected tests and all affected suites passed. |

No nearby caller accidentally rescued either challenged invariant, so coordinated multi-site mutation was unnecessary for these bounded probes. No compile failure was counted as a killed mutant.

### Separate systemic pass and surviving hypotheses

After the checklist/probes, reviewed shared preparation, identity resolution, guard order, no-effect paths and committed errors again. Shared-oracle confidence failed under the demonstrated L1 mutant. The independent expected-value probes falsified actual baseline metadata/scaffold loss; shared production preparation need not be duplicated to fix its test coverage.

The strict-parent bypass hypotheses were falsified for full-stem rename/impostor/prefix, declaration substitution, malformed metadata, already-present symlink/directory targets, mismatched task Epic, missing capability and revoked queued authorization by the evidence above and shipped tests. No selected operation borrowed reference/path or graph/lifecycle authority. The finalization probe falsified loss/replay of committed evidence across both Service and adapter.

Bounded limits retained:

- Raw writers can still change type/bytes after the adapter's directory scan/read or final check; previews do not lock or reserve. Same-ID delete/recreate and later parent removal remain outside the contract. These were not misclassified as cooperating-writer defects.
- Exact full-stem mutation identity does not replace legacy numeric epic-reference grouping. `internal/domain/epic.go:13-22` and `internal/core/service_epic.go:173-183` still group by the leading number; `internal/domain/lint.go:49-85` explicitly diagnoses duplicate-number epics. The hostile lookalike fixture exercises selection in that invalid duplicate-number state, not a promise to redesign rollups or persisted references.
- Stable source-set witnesses detect bad composition at construction; they do not prove an arbitrary adapter is honest forever or authorize another workspace. A future form must retain its originating Service, but no form or selected graph write capability was reviewed as implemented here.
- Ordinary create's existing authorization timing and shared atomic-helper failure modes were not changed or exhaustively fault-injected. No new concrete baseline defect was established there. The current probe floor is not evidence for every OS/filesystem failure, raw race or hypothetical remote adapter.

### Final validation and restoration

```sh
go test -race ./internal/core ./internal/store -run '^TestReviewerSelected|^TestSelectedEpic|^TestNewServiceRejectsEveryMismatchedSplitCapability$' -count=1
go test -race ./internal/core ./internal/store ./internal/cli ./internal/wire ./internal/tools/architecturecheck -count=1
GOFLAGS=-buildvcs=false just build
```

Restored targeted tests passed (core 1.222s, store 1.277s). Full affected suites passed: core 1.524s, store 13.212s, CLI 10.192s, wire 2.539s, architecture/navigation 7.755s. The later standalone-witness addition to the temporary reviewer probe also passed. The original just build exited 0 but emitted an attempted Go module stat-cache write permission warning; the explicit build above exited 0 without that warning and produced the fresh CLI used for finding generation. GOCACHE and linter cache remained in the external reviewer-owned directory.

Full `just test` and `just lint` were not independently rerun; the brief's owner results are inherited. No permanent code/command metadata changed, so no CLI docs generation was needed. Final planning/link/audit lint, brief-preservation and diff checks are recorded below after report assembly. All temporary production edits are restored; temporary tests, fixtures and build output are removed before helper verification. Only the assigned audit is eligible for transfer.

Final corpus-local fixture repeat: when TMPDIR was first moved under the sandbox, the full CLI suite's outside-repository tests correctly discovered the ancestor planning checkout and failed their outside-repository assumptions. This was an unrelated fixture-placement failure, not a killed mutant. A disposable `.reviewer-runtime/.git` file with no metadata pointer supplied the discovery climb boundary implemented in `internal/config/config.go:114-160`. Repeating the priority mutant with that boundary passed all three affected suites (core 1.524s, store 12.618s, CLI 9.713s), then failed the independent priority assertion as intended; restoration passed core 1.168s and store 1.224s. The source-set mutant was also repeated with corpus-local temp roots: intended runtime failure, then restored pass (core 1.154s). The boundary file and all fixture artifacts are removed before verification.

Final report checks: `./bin/tskflwctl -C . audit lint 6gndg1vdbbxd --json` returned empty unreadable/issues arrays; `audit findings 6gndg1vdbbxd --json` returned exactly L1 with status open. `./bin/tskflwctl -C . lint --links` exited 0 with all planning entities and dependency links passing. Comparing the complete prefix before `## Reviewer report` against the sandbox baseline confirmed byte-identical brief/frontmatter; `git diff --check` passed. After sandbox creation, source checkout access was limited to the prescribed helper protocol.

The final corpus-local baseline and restored-suite repeat also passed: focused baseline core 1.169s/store 1.244s; queued-write tests at race count=20 store 1.184s; full affected core 1.439s, store 12.486s, CLI 9.538s, wire 2.484s, architecture/navigation 7.675s. These repeats used the declared external caches and sandbox-local TMPDIR with its disposable discovery boundary.

### Isolation verification attestation

Output from the read-only source helper after all probes and scratch artifacts were removed. Only this audit differs from the sandbox baseline. This report is final; delivery is attested separately by the helper-owned transfer receipt.

```text
sandbox_path=/private/tmp/isolated-review.Oh18JF
git_dir=/private/tmp/isolated-review.Oh18JF/.git
baseline_commit=1c79baeef305c6fc0a3a775a237ef3af108ba23d
source_blob=78ac8f43ce5bb1dcd13053786520a24e24f71d56
source_fingerprint=475eaedfcbe42c81cce980a2c73b9e850c2c8a5e
deliverable=planning/audits/6gndg1vdbbxd-2026-10-19-selected-epic-task-creation-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```
