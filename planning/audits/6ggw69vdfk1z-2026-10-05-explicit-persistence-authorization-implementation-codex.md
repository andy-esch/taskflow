---
schema: 1
id: 6ggw69vdfk1z
bucket: closed
area: explicit-persistence-authorization-implementation-codex
date: "2026-10-05"
---
# Audit: Explicit persistence authorization implementation — codex — 2026-10-05

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

Challenge explicit persistence authorization, including constructor refusal, late opening, and
invocation composition failure. Find concrete data-safety or compatibility defects, not just API
preferences. Do not implement fixes or settle findings. This is a composition safeguard, not a
security boundary against malicious Go code. Required arguments and a named unrestricted mode
were explicitly user-approved before implementation; do not reopen those settled decisions.

## Review target

- Task: `planning/tasks/6gg7e59cyxxh-require-an-explicit-mutation-authorization-policy-at-persistence-composition.md`.
- Thread: `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.
- Source finding: M2 in `planning/audits/6gefs7wbbcyt-2026-09-28-arch-hexagonal-boundaries.md`.
- Branch `feat/explicit-persistence-authorization`, based on merged main `e6bce5b`. The target is
  the captured **uncommitted overlay**, including new files and fixture migration, not HEAD alone.
  PR #282 planning bookkeeping is intentionally bundled; it is not new implementation scope.
- Production: `internal/core/mutation_policy.go`, `internal/store/fsstore.go`,
  `internal/{configstore,spacestore,workspacestore}/fs.go`, `internal/workspacestore/planning.go`,
  `internal/appwiring/wiring.go`, `internal/cli/ports/runtime.go`, and CLI pre-run/composition routes.
- Produce a repository-wide **consumer inventory**: persistence constructors/options and callers;
  store mutation entries and nested helpers; configuration/registry writes; ordinary/workspace/Atlas
  opening; operational CLI hooks and optional discovery fallbacks; metadata and completion;
  test fixtures and their policy choices. Verify symbols and actual call paths rather than trusting
  the task's inventory or claiming that an internal constructor is a public Go SDK promise.

## Intended contract to challenge

1. `MutationPolicy` is an opaque value. Zero values and `GuardedMutations(nil)` are invalid; required
   constructor arguments remove implicit compatibility defaults. Invalid construction returns no
   usable adapter and does not invoke authorizers, discovery, or filesystem effects.
2. `ReadOnlyMutations` permits reads but denies every mutation API, including no-op/preview requests,
   before filesystem effects, editors, transformations, and planner callbacks. A direct zero-value
   adapter denies mutations too. Low-level topology scaffolding in CLI init remains its separately
   guarded, documented exception; this slice does not relocate config/userconfig primitives.
3. `GuardedMutations` consults its callback at every mutation boundary, never caching constructor
   authorization. A successful operation may cross several guarded boundaries; exactly one call
   per whole successful use case is **not** the contract. Guard errors retain their original cause.
4. Ordinary and direct/pointer workspace opening preserve the same policy, source-set identity,
   watcher corpus, and fresh marker-anchored identity reader. Read/open operations do not call the
   mutation authorizer. Atlas summary stores are explicitly read-only regardless of registry mode,
   including a consumer widening the returned interface to the concrete store.
5. `UnrestrictedMutations` is a deliberate trusted-fixture/tooling choice, not a runtime fallback.
   Production CLI/TUI wiring remains guarded. Locks, CAS, planner-window conflicts, eligibility,
   recovery, and domain validation still apply in unrestricted mode.
6. Failed `Bindings.Compose` discards **all** partial services and preserves refusal on operational
   CLI paths, without cwd/space/builtin fallbacks hiding it. Help, version, schema, and completion
   scripts work without persistence; entity completion emits its quiet empty protocol result.
   Missing bindings retain their established explicit-unavailable semantics.
7. Only internal Go constructors change. No machine-schema revision, persisted representation,
   public CLI flag/command change, or generated-doc drift is intended. `testutil.Must` unwraps valid
   fixture construction but does not select authorization or mask errors in negative tests.

Non-goals: generic permissions frameworks, tamper-proof adapters, new locking/CAS semantics,
another backend, a public SDK compatibility layer, or the separate final boundary regression task.
Evidence-backed out-of-scope problems should be proposed as bounded followups, not silently added
to this refactor's closure criteria.

## Mandatory evidence floor

- Verify implementation and test names; run focused/full tests, race tests, lint, build, planning
  lint, and generated-output comparisons in your independent sandbox. Cite exact results.
- Exercise constructor zero/nil policies, readable read-only stores, denied dry-run and committing
  writes, populated records, no-op requests, and mutation callbacks that would visibly run.
  Inspect bytes/directories afterward, including an isolated home registry/config.
- Exercise ordinary and direct/pointer workspace opening, plus Atlas summary capability widening.
  Change a guarded decision **after** opening: denied → allowed → denied. Demonstrate that neither
  construction nor opening caches permission. Preserve cause/classification, not only error prose.
- Execute at least three compiler-valid mutation probes at distinct boundaries, with the exact
  regression claimed to kill each probe. A compiler error is not a killed probe. Include a
  coordinated probe where duplicate guard checks would otherwise accidentally preserve behavior.
  Restore each probe to the sandbox baseline before continuing; rerun the unmodified tests.
- Test CLI failure routes with both a populated cwd and a partial injected services bundle. Cover
  schema/help/version/completion versus operational hooks, especially template/theme best-effort
  discovery, `status --all`, space/config, and UI startup. No leaked success payload on refusal.
- Provide a claim/evidence matrix, strongest/weakest test discussion, and an explicit verdict:
  ready, ready with tracked followups, or blocked. A green suite is not hostile evidence by itself.

## Required hostile angles

### Shared policy and mutation completeness

Independently inventory the mutation APIs rather than accepting the 26-row test table. Are there
public paths, helper callbacks, empty plans, previews, or zero-value adapters that evade the policy?
Can the policy value be accidentally invalidated after constructor validation? Is that a reachable
composition regression or arbitrary deliberate misuse? Distinguish the two. Challenge classification
and guard timing without replacing repository concurrency/domain checks with authorization.

### Late opening and privilege narrowing

Probe the shared observed-planning helper, both workspace entry forms, and the summary opener.
Suggested compiler-valid mutations: make a late opener choose unrestricted policy; have the Atlas
opener inherit the registry policy; invoke the guard during construction. Identify the exact named
tests that fail and why. Do not claim propagation from a fake WorkspaceStore that returns a stronger
capability than production. Verify real discovery/root/watcher/identity behavior remains pinned.

### Composition refusal and deceptive fallbacks

Supply an error with a non-empty bundle. Challenge publication and every pre-run branch. Removing
only one error check may leave another protecting the command; coordinate the changes if necessary.
In particular, remove the template/theme explicit refusal and see whether optional discovery turns
the failure into builtin success. Exercise hidden completion separately from completion scripts.
Do metadata-only bindings still work, and can failure affect later independent invocation trees?

### Systemic second pass — especially Antigravity

Choose two trust assumptions and actively disprove them: (a) writable fixture migration made tests
green by bypassing a real production guard; (b) a constructor test proves only an error, not absence
of I/O; (c) summary interface narrowing hides an unrestricted concrete store; (d) duplicate checks
mask incomplete propagation or a stale allow decision. For each report input, exact path, observed
result, and conclusion. One demonstrated systemic issue beats several speculative redesigns.
No readiness verdict based solely on restating owner claims or finding no local syntax problems.

## Validation and restoration

Use only the mandatory independent clone above; build its fresh binary. Never run probes,
generators, planning mutations, or Git writes in the source checkout. The captured sandbox baseline
is the restoration point; preserve only your assigned audit at transfer. Do not commit again,
push, open a PR, or change tasks/Thread states/source finding dispositions.

Suggested commands (verify names first):

```sh
go test ./internal/core -run '^TestMutationPolicy' -count=1
go test ./internal/appwiring -run 'TestPersistenceConstructors|TestInvalidLocalComposition|TestLocalCompositionPropagatesAuthorization' -count=1
go test ./internal/store -run 'TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects' -count=1
go test ./internal/configstore ./internal/spacestore -run 'TestConfigurationPolicies|TestRegistryPolicies|TestSummaryPlanningStores' -count=1
go test ./internal/workspacestore -run 'TestWorkspaceOpeningCarriesPolicy|TestZeroWorkspaceAdapter|TestNewPlanningStore' -count=1
go test ./internal/cli -run 'TestFailedComposition|TestMetadataOnlyCommandTree|TestLocalCompositionKeepsAuthorization' -count=1
go test ./...
go test -race ./...
golangci-lint run ./...
just build
./bin/tskflwctl -C . lint --json --no-input
```

Generate CLI docs and schema comments into private scratch paths and compare against checked-in
artifacts; do not update goldens to hide drift. For CLI dogfood create a throwaway directory with
`init --path "$SCRATCH" --taskflow-root planning --no-register`, and isolate `TSKFLW_CONFIG_HOME`.
Use explicit `--path` for init, not global `-C` as a bootstrap destination selector.

## Deliverable

Update only your assigned audit, retaining this brief. Add severity-coded **open** findings with
exact reproduction, implementation path/line, impact, bounded recommendation, and the regression
that should pin it. Do not settle findings. Keep falsified hypotheses and evidence limitations
separate. Include consumer inventory, commands/results, restored mutation table, claim/evidence
matrix, isolation attestation, and verdict. Do not duplicate one root cause or invent a quota.

## Reviewer report

**Verdict: ready with tracked followups.** Both review passes are complete against the captured uncommitted overlay. No new in-scope data-safety or compatibility defect was demonstrated. The executed markerless identity repair reproduces the existing, separately tracked task [6ggfd81jg0qg](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md); it does not become a new authorization-refactor prerequisite. No implementation fix, task/Thread state change, or finding disposition was made.

### Scope and consumer inventory

The independent baseline includes the handoff's staged, unstaged, untracked, and deleted state. Review covered the implementation delta from `e6bce5b`, constructor/fixture migration, production command composition, nested write helpers, and late openers. PR #282 bookkeeping is not treated as implementation evidence. All paths and line anchors below were verified in this baseline. The constructors live under `internal/`; no public Go SDK compatibility promise is inferred.

| Consumer or boundary | Verified implementation and conclusion |
| --- | --- |
| Policy value | `internal/core/mutation_policy.go:13`, `:46`, `:61`: private mode/callback fields; zero and nil guard invalid; validation does not execute the callback; authorization invokes it afresh and returns its original error. Both policy refusal sentinels wrap `domain.ErrValidation` (`:27`). No implicit unrestricted branch or exported policy setter exists. |
| Planning constructor/options | `internal/store/fsstore.go:101` validates before allocating the store/source set or executing options. `WithSourceSetID` (`:63`) and `WithPlanningIdentityReader` (`:69`) assign fields; neither discovers or executes the reader. `NewSourceSetID` uses an atomic counter (`internal/core/source_set.go:19`), not filesystem I/O. The old `WithMutationAuthorization` name has no remaining Go match. |
| Other constructors | `internal/configstore/fs.go:23`, `internal/spacestore/fs.go:24`, `internal/workspacestore/fs.go:17`, `internal/workspacestore/planning.go:17`: validate first and return nil on invalid policy. No constructor authorizes or performs discovery. The observed-planning helper delegates to the separately validating `store.NewFS`. |
| Production selection | `cmd/tskflwctl/main.go:23` supplies `appwiring.LocalBindings()` to `cli.NewRootCmd`. `internal/appwiring/wiring.go:26` selects `GuardedMutations(authorize)` once and passes that value to configuration, registry, workspace and lazy ordinary planning composition. Repository-wide production call search found no call to `UnrestrictedMutations` outside its factory declaration. Results: `repository-production-policy-inventory.txt`. |
| Configuration writes | `internal/configstore/fs.go:99` `MigrateConfiguration` and `:145` `SetPreference` authorize before config/userconfig operations, including preview. Load/diagnose paths are reads. The low-level config/userconfig primitives remain the intended existing implementation dependencies. |
| Registry writes | `internal/spacestore/fs.go:72` `AddSpace` and `:85` `ForgetSpace` authorize before home-registry mutation. `PrepareSpace` (`:52`) and `ListSpaceEntries` (`:40`) are discovery/read paths, not writers. |
| Ordinary opening | `internal/appwiring/wiring.go:55` validates before discovery, then `:63` constructs from that one observation. A read/open does not execute authorization. The published service/layout use that concrete store. |
| Workspace opening | `internal/workspacestore/fs.go:26` validates before real `config.Discover`, then passes its retained policy at `:34`. The workspace capabilities come from the same store. `internal/workspacestore/planning.go:27` captures marker directory, or legacy root, as the later identity-read start; it retains the initially observed root. |
| Atlas summary | `internal/spacestore/fs.go:99`, `:105`: validates the registry adapter, then explicitly constructs a read-only planning store. A concrete `*store.FS` assertion does not widen its permission, even when the registry's own guard allows mutation. |
| Nested filesystem writes | `internal/store/create.go:71` guards before mkdir/prepare; `:98` takes the checked write lock. `writeNewFileUnlocked` (`:126`) is private and reached under that guard. `internal/store/body.go:24` `writeBody` and `internal/store/edit.go` `editFile` receive the guarded lock function. `internal/store/lock.go:103`, `:115` route writes through policy before lock acquisition. On this Unix platform the advisory lock is `flock` on the root directory (`internal/store/lock_unix.go:21`), not an invented lockfile. |
| Identity/concurrency | `internal/store/threadapply.go:206` re-reads identity, rejects a changed root with conflict and missing durable ID with validation. Authorization has not replaced this check, CAS, planner-window restrictions, eligibility, or domain validation. Existing unrestricted fixtures still exercise those checks in the full suite. |
| CLI partial composition | `internal/cli/ports/runtime.go:41` returns `(Services, error)`. `internal/cli/root.go:284` retains the error and publishes the bundle only when it is nil. `styleOnlyPreRun:162`, `repoPreRun:360`, `resolveFrom:457` preserve refusals; missing services/openers retain explicit-unavailable validation. |
| Optional discovery routes | Explicit checks precede best-effort fallback in `internal/cli/template.go:28`, `theme.go:38`, `status.go:41`, `ui.go:41`. Removing template/theme checks demonstrably returns builtins despite the other resolution check. The UI ambient-miss predicate can similarly misclassify a composed `ErrNoConfig` without its explicit check. |
| Metadata/completion | `internal/cli/root.go:151` metadata hook and `:352` completion-script hook allow persistence-independent commands. Hidden entity completion returns before operational resolution; `internal/cli/completion.go:104` supplies a deferred service. Failure yields the exact quiet `:4\n` protocol. Space completion also receives no failed service. |
| Invocation and UI/init | `internal/cli/command_safety.go:52` consults the current command tree's classified safety. `internal/appwiring/launch.go:9`, `:22` receive guarded services rather than constructing an unrestricted adapter. `internal/cli/init.go:54` is the documented separately guarded topology-scaffolding exception (`docs/ARCHITECTURE.md:144`). |
| Fixtures/helper | `internal/testutil/repo.go:26` `Must[T]` only panics on an error and returns the value; it selects no policy. Negative constructor tests inspect errors directly. Baseline test-call census: 307 unrestricted calls in 61 files, 98 read-only calls in 35 files, 15 guarded calls in 9 files; lists are retained in `fixture-policy-inventory.json`. Writable migration is explicit, not a runtime fallback. |
| Machine/CLI compatibility | `internal/wire/wire.go:343` remains schema `1.81`. `git diff e6bce5b --stat -- internal/wire docs/cli internal/cli/testdata/golden` is empty. Private doc/schema-comment generation matches checked-in artifacts. No new optional wire branch exists in this slice, so the non-default optional-branch validator requirement has no new branch to exercise. |

The independent mutation inventory searched actual exported receiver methods, package functions and filesystem write sites, then reconciled them with the owner table. It found these **26** mutation APIs; it did not derive completeness from the table's row count:

| Implementation | Public entries | Authorization route |
| --- | --- | --- |
| `internal/store/create.go:185`, `:291`, `:346`, `:441` | `CreateTask`, `CreateAudit`, `CreateResearch`, `CreateEpic` | Shared `createEntityFile` before prepare/mkdir, then checked lock. Domain input validation may precede it without callbacks/effects. |
| `internal/store/auditstore.go:108` | `MoveAudit` | Entry guard, then guarded lock for write. |
| `internal/store/body.go:91`, `:145`, `:189`, `:249` | `AppendAuditBody`, `EditBody`, `TransformAuditBody`, `TransformTaskBody` | Entry guard before transformation/no-op; guarded body writer. |
| `internal/store/edit.go:145`, `:286` | `EditTask`, `EditAudit` | Entry guard before editor; guarded editor writer and CAS. |
| `internal/store/fix.go:25`; `fsstore.go:257`; `rename.go:31` | `FixFrontmatter`, `SetFields`, `RenameTask` | Entry guard including empty/preview requests; guarded writer. |
| `internal/store/epicstore.go:104`, `:174`, `:235` | `MoveEpic`, `SetEpicFields`, `EditEpic` | Entry guard; guarded writer/editor. |
| `internal/store/researchstore.go:146`, `:212`, `:243` | `SetResearchFields`, `EditResearch`, `AppendResearchBody` | Entry guard; guarded writer/editor. |
| `internal/store/graphmutation.go:26`; `graphrepair.go:21`; `lifecyclemutation.go:20` | `MutateTaskGraph`, `MutateTaskGraphRepair`, `MutateTaskLifecycle` | Checked lock before authoritative snapshot/planner, even preview/empty plans. |
| `internal/store/threadapply.go:18`; `threadmutation.go:19`; `threadcreation.go:17` | `MutateThreadApply`, `MutateThread`, `MutateThreadCreation` | Checked lock before planner; creation additionally guards before root mkdir. |

### Claim/evidence matrix — first pass

All evidence files named without a directory below live in the retained sandbox's `.git/isolated-review-workspace/evidence/`. Reviewer probes were temporary executable tests, archived there as `probe-store.go`, `probe-appwiring.go`, `probe-cli.go`, and `probe-workspacestore.go`; none is presented as a shipped regression or transferred implementation.

| Challenged claim | Hostile input and observed result |
| --- | --- |
| Invalid construction cannot publish or trigger work | Owner `TestPersistenceConstructorsRequireValidPoliciesWithoutIOOrAuthorization` tests zero/nil policies across all five constructors. Reviewer `TestAuditInvalidConstructorDoesNotRunOptions` adds an observable panicking option: both invalid modes return nil/error before executing it. Discovery/authorizer witnesses stay at zero. Removing the concrete constructor check publishes an invalid adapter and executes the option; coordinated helper/store removal defeats both validations. |
| Reads work while every mutation refuses before callbacks | `TestAuditPopulatedFSMutationsRefuseBeforeCallbacksAndNoops` creates readable task/audit/epic/research records, asserts successful reads, then invokes all 26 methods against read-only, guarded-refusal, and a populated adapter with zero policy. **156 denial cases** preserve cause and complete file/directory snapshots; editor/transform/planner counters remain zero. Empty fields, same-state moves and no-op callbacks are included. Preview/commit are exercised where the API has that parameter; editor APIs have no preview argument and are repeated in the two request groups. Owner tests separately cover actual direct zero-value `FS{}` and absent roots. |
| Fresh decisions survive real late opening | `TestAuditRealCompositionPermissionTransitionsAndSummaryNarrowing` uses real config discovery, ordinary planning and real workspace services for direct and pointer entry forms. Composition/opening authorizer calls are zero. Seven use cases × preview/commit × denied→allowed→denied run after opening: ordinary/workspace writes, repository/user preferences, idempotent migration, registry add and absent forget. Denials preserve wrapped conflict and every repo/home byte/directory. Calls advance 22→38→60 per entry form. Allowed absent-forget retains its legitimate not-found result. |
| Guard timing is per boundary, not once per use case | Direct denied store calls invoke the guard once and stop before callbacks. Core service operations may retry a denied conflict before writing, and successful operations can traverse several checks. The transition fixture therefore asserts a fresh increase, not exactly one call for a whole use case. No guard result is cached. |
| Summary privilege is narrowed in the concrete adapter | Owner `TestSummaryPlanningStoresCannotInheritRegistryMutationPrivilege` and the reviewer transition test assert actual `*store.FS`, then invoke a no-op mutation with an allowing registry guard. Preview/commit return `ErrReadOnlyPersistence` without consulting that guard. The inherited-policy mutant fails the owner regression. |
| Corpus identity/watchers stay pinned | Real ordinary/workspace stores have nonzero source-set witnesses and exactly the five watched entity directories at the initially observed root. Legal direct-marker and pointer repointing to another real corpus causes both apply previews and commits to return conflict; neither corpus changes and watchers remain pinned. Owner initial-observation, alias-mutation and reader-error tests additionally exercise the captured marker start and forbid cached identity fallback. |
| Partial services and fallbacks cannot hide refusal | `TestAuditFailedBundleCannotPublishRealServicesOrFallback` returns **all real composed services plus an error**, with a populated cwd and isolated home. Two causes (wrapped conflict and wrapped `config.ErrNoConfig`) × 21 command scenarios: 28 operational refusals preserve `errors.Is`/exit class and emit no success stdout; eight metadata/script successes; six hidden completions emit exactly `:4\n`. All service fields remain nil; opener, browser and configuration-editor hooks stay unused; repo/home trees remain equal and init creates no target. A later independently composed read invocation succeeds. |
| Runtime guarded wiring still permits intended work | Fresh `bin/tskflwctl` dogfood executes 21 commands using explicit `init --path … --taskflow-root planning --no-register` and isolated `TSKFLW_CONFIG_HOME`: epic/task creation; task read/preview/commit/read-back; config show/migration preview and idempotent commit; registry add/list/forget previews/commits; `status --all`; registered-space selection; repo-less schema/version/help/bash completion. Every exit is zero. Reads/previews/metadata preserve the tree; committed task/registry changes modify it as expected (`dogfood.json`). |
| Emitted repair commands actually run in the recommending state | `TestAuditLegacyIdentityRepairRunsAgainstRecommendedState` removes a direct marker after composing a valid plan. Apply refuses before document writes and recommends `config migrate`. The fresh binary executes that command in the same markerless tree: exit **11**, JSON code `validation`, requiring `init` first. Explicit `init --path <same-root> --no-register` creates only the marker. A newly composed plan commits; the old plan still conflicts with the regenerated ID. This is the already-tracked followup, not a new policy regression (`legacy-repair.log`). |

### Executed mutation probes and restoration

**32 compiler-valid probes; 29 semantic failures; three deliberately masked single-removal controls.** A surviving control is not counted as killed. Its coordinated counterpart demonstrates which duplicated boundary protects the invariant. `mutations.py` records exact replacements, package/run arguments and baseline restoration. `mutations.json` records each command, exit and failure evidence. Each row has `<probe>.log` and `<probe>-restored.log`; all **32** immediate reruns on restored code exit zero. Compiler errors are not used as evidence.

The test aliases below refer to verified, shipped names except R1/R2/R4, which are reviewer-only files archived privately:

| Alias | Exact test and source |
| --- | --- |
| T1 | `TestMutationPolicyModesFailClosed`, `internal/core/mutation_policy_test.go:11` |
| T2 | `TestMutationPolicyValidatesWithoutCallingAndNeverCachesAuthorization`, same file `:42` |
| T3 | `TestPersistenceConstructorsRequireValidPoliciesWithoutIOOrAuthorization`, `internal/appwiring/persistence_policy_test.go:17` |
| T4 | `TestInvalidLocalCompositionDoesNotPublishPartialServicesOrDiscover`, same file `:58` |
| T5 | `TestLocalCompositionPropagatesAuthorizationToEveryPersistenceFamily`, `internal/appwiring/wiring_test.go:95` |
| T6 | `TestWorkspaceOpeningCarriesPolicyAcrossDirectAndPointerEntryPoints`, `internal/workspacestore/fs_test.go:144` |
| T7 | `TestSummaryPlanningStoresCannotInheritRegistryMutationPrivilege`, `internal/spacestore/mutation_policy_test.go:49` |
| T8 | `TestZeroWorkspaceAdapterRefusesBeforeDiscovery`, `internal/workspacestore/fs_test.go:214` |
| T9 | `TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads`, `internal/workspacestore/planning_test.go:18` |
| T10 | `TestNewPlanningStoreIdentityReaderFailuresNeverUseCachedIdentity`, same file `:124` |
| T11 | `TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata`, `internal/cli/composition_test.go:64` |
| T12 | `TestLocalCompositionKeepsAuthorizationInvocationScoped`, same file `:253` |
| T13 | `TestConfigurationPoliciesDenyBeforeEffectsIncludingPreview`, `internal/configstore/mutation_policy_test.go:14` |
| T14 | `TestRegistryPoliciesDenyBeforeEffectsIncludingPreview`, `internal/spacestore/mutation_policy_test.go:16` |
| T15 | `TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects`, `internal/store/mutation_authorization_test.go:15` |
| R1 / R2 / R4 | `TestAuditPopulatedFSMutationsRefuseBeforeCallbacksAndNoops` / `TestAuditInvalidConstructorDoesNotRunOptions` in `probe-store.go`; `TestAuditFailedBundleCannotPublishRealServicesOrFallback` in `probe-cli.go` |

| Probe name | Executed mutation | Exact regression outcome |
| --- | --- | --- |
| `zero-policy-validates` | Policy `Validate` accepts zero | T1 `/zero` fails: authorization returns nil. |
| `read-only-allows` | Read-only authorization returns nil | T1 fails read-only refusal. |
| `guard-does-not-run` | Guarded authorization returns nil without callback | T2 fails: calls=0. |
| `eager-constructor-authorization` | Validation executes the callback | T2/T3 fail: eager call and valid construction refused. |
| `planning-constructor-publishes-invalid` | `NewFS` ignores validation error | T3/R2 fail: nonnil adapter; option runs. |
| `configuration-constructor-publishes-invalid` | Config constructor ignores validation error | T3 configuration case fails. |
| `registry-constructor-publishes-invalid` | Registry constructor ignores validation error | T3 registry case fails. |
| `workspace-constructor-publishes-invalid` | Workspace constructor ignores validation error | T3 workspace case fails. |
| `observed-helper-check-alone-masked` | Observed helper ignores validation error | T3 **passes**; concrete constructor still refuses. |
| `coordinated-observed-and-store-construction` | Helper and concrete constructor both ignore it | T3 planning and observed-planning cases fail publication. |
| `invalid-opening-discovers-first` | Ordinary opener skips early validation | T4 fails observable discovery before refusal. |
| `late-shared-planning-unrestricted` | Shared late helper selects unrestricted | T5/T6 fail real late denied/read-only writes. |
| `late-workspace-unrestricted` | Workspace caller substitutes unrestricted | T6 fails direct and pointer cases. |
| `ordinary-opening-unrestricted` | Ordinary caller substitutes unrestricted | T5 fails: mutation allowed, guard calls=0. |
| `atlas-inherits-registry-privilege` | Summary uses registry policy | T7 fails actual concrete-store mutation refusal. |
| `zero-workspace-discovers` | Zero adapter skips opener validation | T8 fails: discovery error replaces invalid-policy refusal. |
| `identity-reader-root-anchor` | Identity reader starts at resolved root | T9 fails retained-marker-start assertion. |
| `identity-reader-uses-cache` | Fresh discovery result/error replaced with cached config | T10 fails nil/empty/error/partial-plus-error cases. |
| `failed-compose-publishes-bundle` | Root publishes bundle despite error | T11 fails immediate service-field assertion. |
| `repo-refusal-alone-masked` | Remove repo pre-run refusal only | T11 **passes**; `resolveFrom` still refuses. |
| `coordinated-repo-and-resolution-refusal` | Remove both refusals | T11 fails: explicit-unavailable error replaces original composition cause. This is not claimed as leaked success. |
| `coordinated-publication-and-both-refusals` | Publish real bundle and remove both refusals | R4 fails: task show emits a real success payload; failed bundle is usable. |
| `template-builtin-hides-refusal` | Remove template explicit refusal | T11 fails: builtin list succeeds despite failed composition. |
| `theme-builtin-hides-refusal` | Remove theme explicit refusal | T11 fails: builtin themes succeed. |
| `ui-ambient-miss-hides-refusal` | Remove ambient UI explicit refusal | R4 `/ui/1` fails: wrapped missing-config cause becomes terminal-validation error. No browser launch is claimed. |
| `writer-check-alone-masked-on-create` | Remove shared checked-lock authorization | T15 `/guarded/CreateTask` **passes** because initial create guard remains. |
| `coordinated-create-and-writer-bypass` | Remove both initial create and checked-lock guards | Same exact T15 subset fails preview/write: nil error instead of denial. |
| `transform-noop-runs-before-authorization` | Remove task-transform entry guard | R1 fails populated no-op: callback runs once and returns nil rather than refusal. |
| `runtime-composition-unrestricted` | Production composition selects unrestricted | T5/T12 fail; read invocation can mutate and callback is skipped. |
| `configuration-guard-bypass` | Config adapter's authorizer returns nil | T13 fails original refusal before migration discovery. |
| `registry-guard-bypass` | Registry adapter's authorizer returns nil | T14 fails: add succeeds instead of original denial. |
| `all-fs-guards-bypass` | Shared filesystem authorizer returns nil | T15 fails across mutation families and policy modes. |

### Systemic second pass and test strength

| Trust assumption actively challenged | Input, observation and conclusion |
| --- | --- |
| Writable fixture migration could hide production authorization loss | Repository-wide non-test caller search shows the sole runtime selector is guarded (`internal/appwiring/wiring.go:27`), and fixtures explicitly choose their mode. Substituting unrestricted at that runtime selector makes T5/T12 fail, including a retained read invocation after another invocation mutates. The independent real-service transition fixture also refuses after opening, allows only after the callback changes, and refuses again. The shared factory/helper has not silently made production permissive. |
| Constructor error tests could miss I/O or callbacks | T3's absent-path assertions alone do not prove no reads. Inspection of each actual constructor/option proves its pre-validation path is pure; observable discovery/authorization counters and R2's panicking option test challenge callback timing. Eager validation and concrete publication mutants fail. No syscall trace was collected; the no-read conclusion is bounded to the inspected paths, not inferred merely from unchanged files. |
| A narrow summary interface could conceal unrestricted concrete capabilities | T7 and the reviewer fixture use the actual concrete production store and exercise no-op preview/write with registry permission allowed. Both refuse without consulting registry privilege; the inherited-policy mutant fails. This is permission narrowing in the store, not only interface narrowing. |
| Duplicate checks could mask missing propagation or stale permission | The three passing single removals were followed by coordinated constructor, CLI resolution/publication, and creation/writer removals. All coordinated cases fail semantically. The transition fixture changes decisions on already-open stores. The tests protect the end-to-end invariant; isolated removal survival is explicitly recorded rather than misreported as a kill. |

Strongest owner tests are T6's real direct/pointer permission changes, T7's concrete capability widening, and T10's reader failures including valid-looking partial data with an error. `TestPlanningOpenRetainsInitialCorpusWhenMarkerChangesDuringDiscovery` (`internal/appwiring/thread_apply_test.go:21`) and T9 pin initial observation and marker anchoring; they are implemented tests, not planned work.

Weakest evidence by itself is T15's absent-root table: missing references and no callback counters can hide populated no-ops or premature editors. R1 supplies readable records, callback counts, and whole-tree evidence. T3's stat checks alone cannot certify absence of reads. T11's original partial bundle contains only an opener and asserts publication before command execution; R4 supplies all real services and checks operational consequences, error classification, quiet completion, home/repo bytes, and subsequent invocation independence. These limitations do not establish an untested production defect after the supplemental probes and coordinated mutations.

The reviewer harness was corrected before its final passing run: whole-service guard counts allow multiple boundaries/retries, and the repoint uses a legal child corpus. One initial text replacement targeted the wrong CLI check and aborted before its claimed regression ran; baseline restoration ran, the replacement was corrected to the exact repo/resolution branches, and the final named logs are the reported outcomes. None of these harness corrections changed production code.

Falsified hypotheses are kept separate from findings: an arbitrary caller-defined `FSOption` could deliberately overwrite an adapter, but that is malicious Go misuse outside the stated composition contract; ordinary option/caller paths do not mutate private policy fields. Whole-use-case guard calls greater than one are expected, including pre-write conflict retries; they are not stale authorization. A marker repoint outside its allowed root is validation, so the hostile repoint fixture instead uses a legal, distinguishable alternate corpus to reach the fresh-root conflict check.

### Validation commands and results

All commands ran inside the independent sandbox with `GOCACHE=/tmp/taskflow-review-go-cache`; lint additionally used `GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache`. Platform/toolchain: **Darwin arm64, Go 1.27.1**.

| Executed command | Result/evidence |
| --- | --- |
| Focused policy, constructor, opener and CLI regressions (exact command below) | All seven packages pass (`baseline-focused.log`). |
| `go test ./... -count=1` | Pass before probes (`baseline-full.log`) and after all source/probe restoration (`restored-full.log`). |
| `go test -race ./... -count=1` | Pass (`baseline-race.log`). |
| `go test ./internal/store ./internal/appwiring ./internal/cli -run '^TestAudit' -count=1 -v` | Independent populated/transition/CLI probes pass (`independent-probes.log`); strengthened CLI tree snapshot rerun passes (`independent-cli.log`). |
| `go test ./internal/workspacestore -run '^TestAuditLegacyIdentityRepairRunsAgainstRecommendedState$' -count=1 -v` | Pass with the executed, asserted known migration refusal and successful init/recomposition (`legacy-repair.log`). |
| `go test -race ./internal/store ./internal/appwiring ./internal/cli ./internal/workspacestore -run '^TestAudit' -count=1` | All four independent-probe packages pass (`independent-race-final.log`). |
| `golangci-lint run ./...` | Exit 0, **0 issues** (`baseline-lint.log`). |
| `just build` | Exit 0; fresh `bin/tskflwctl` built and executed in dogfood/repair fixtures (`baseline-build.log`). Go printed a denied external module-stat-cache write; the build and binary execution still succeeded. |
| `go run ./internal/tools/docgen -out .git/isolated-review-workspace/evidence/generated-cli` then `diff -qr docs/cli .git/isolated-review-workspace/evidence/generated-cli` | Exit 0; no doc differences. |
| `go run ./internal/tools/schemacomments -out .git/isolated-review-workspace/evidence/schema_comments.json` then `cmp internal/wire/schema_comments.json .git/isolated-review-workspace/evidence/schema_comments.json` | Exit 0; all 267 comments match (`generated.log`). No golden/update generator ran. |
| `TSKFLW_CONFIG_HOME=<private-evidence-home> ./bin/tskflwctl -C . lint --json --no-input` | Exit 0: `unreadable: []`, `issues: []` (`planning-lint.json`); repeated after report (`final-planning-lint.json`). |
| `python3 .git/isolated-review-workspace/evidence/mutations.py` | All 32 compiler-valid outcomes and 32 restored reruns satisfy their asserted result (`mutations.json`). |
| Fresh-binary fixture commands | All 21 positive CLI commands, exact argv/output and tree results in `dogfood.json`. Markerless repair commands/results in `legacy-repair.log`. |

Exact focused command (the full/race runs also include the remaining verified owner regressions):

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/core ./internal/appwiring ./internal/store ./internal/configstore ./internal/spacestore ./internal/workspacestore ./internal/cli -run 'TestMutationPolicy|TestPersistenceConstructors|TestInvalidLocalComposition|TestLocalCompositionPropagatesAuthorization|TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects|TestConfigurationPolicies|TestRegistryPolicies|TestSummaryPlanningStores|TestWorkspaceOpeningCarriesPolicy|TestZeroWorkspaceAdapter|TestNewPlanningStore|TestFailedComposition|TestMetadataOnlyCommandTree|TestLocalCompositionKeepsAuthorization' -count=1
```

### Findings, followups and evidence limits

No new severity-coded finding is opened: the demonstrated current limitation is already tracked by [6ggfd81jg0qg](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md), originating in L2 of the workspace parity audit. `git diff e6bce5b -- internal/store/threadapply.go internal/config/migrate.go` is empty. The task's unchecked criteria are requirements, not shipped repair evidence; this report does not settle or change them. The existing Thread identity guard remains correct and rejects the stale plan after repair.

This review did not run Linux/Windows, every filesystem, terminal-driven TUI interaction, or all possible non-cooperating race schedules. UI/editor launch refusal was checked through the actual CLI command tree and observable injected launch hooks, not an interactive session. Constructors were inspected and instrumented for callbacks/effects, without syscall tracing. Whole-tree comparisons do not establish crash atomicity or new transaction guarantees. The scope remains explicit internal composition authorization, with existing locks/CAS/domain behavior; it is not tamper-proofing against malicious Go callers. Reviewer source files and generators are restored/removed; private evidence and the fresh ignored binary remain in the retained sandbox only.

### Isolation and guarded transfer attestation

The mandatory helper created an independent `--no-hardlinks` clone and committed only its captured baseline. Source access was limited to reading the assignment, helper creation and final helper transfer. All review work occurred inside that clone; no subsequent commit, staging, shared-checkout project command, implementation transfer, or task/finding state change was made.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YOesVU
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YOesVU/.git
baseline_commit=222e1250549b36469e68b8b7e2c7f3c86ede2610
source_blob=4eb2572a3a958cb3aa5b8e28dad67eb849c84f3d
source_fingerprint=134b7b74b5b53ae68c2ef3ca218bc5ac2e5e0ea0
deliverable=planning/audits/6ggw69vdfk1z-2026-10-05-explicit-persistence-authorization-implementation-codex.md
deliverable_changed=true
transfer=succeeded
```

`verify` and `transfer` enforce independent Git metadata, the baseline, no staging/extra changes and an unchanged source deliverable. The complete report diff was inspected. Only this assigned audit was transferred; evidence stays private in the retained workspace. Final helper output is retained as `final-transfer.log`. Keep the sandbox until the implementation owner confirms receipt.

## Owner reconciliation (2026-10-05)

Accepted the evidence-backed readiness verdict: no new in-scope defect or design decision is
required. The 32 compiler-valid probes distinguish semantic failures from intentionally masked
controls; the report does not claim temporary reviewer tests are shipped regressions.

Persisted the useful test-strength recommendations in the implementation:

- `TestPopulatedFSMutationsRefuseBeforeCallbacksAndNoops` exercises all 26 mutation entries
  against readable records and a healthy graph, including no-op requests, preview/write,
  callback counters, and whole-tree bytes/entries/modes across denied/read-only/zero policies.
- `TestInvalidPolicyDoesNotExecuteConstructorOptions` observes option execution directly for
  zero and nil-guard policy; absent-path checks alone are not evidence of no reads.
- `TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata` now returns a full real
  local services bundle plus each of three failure causes. It retains the real missing-planning
  predicate and checks operational refusal/classification, quiet completion, metadata, unused
  launch/open hooks, unchanged repo/home trees, and a subsequent healthy invocation.

The executed markerless recovery limitation remains tracked by
[6ggfd81jg0qg](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md)
in the CLI-contract Thread; its evidence was refreshed without duplicating the task or making
it an authorization closeout blocker. Full normal/race tests, lint, build, generated-output
comparisons, and planning/audit lint pass after reconciliation. Receipt confirmed; no production
authorization changes were needed. This closes the review, not a claim of merge or release.
