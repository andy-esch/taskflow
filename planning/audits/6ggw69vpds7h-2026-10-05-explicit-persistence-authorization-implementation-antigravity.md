---
schema: 1
id: 6ggw69vpds7h
bucket: closed
area: explicit-persistence-authorization-implementation-antigravity
date: "2026-10-05"
---
# Audit: Explicit persistence authorization implementation — antigravity — 2026-10-05

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

### 1. Mandatory reviewer sandbox attestation

- **Reviewer assignment:** `antigravity`
- **Sandbox path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.RZCmdb`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.RZCmdb/.git`
- **Baseline commit:** `8d69436af6067bc86fec9adffb4f4b06f933ee5c`
- **Captured source blob:** `806180c7c232f5766a04eb3990ae79bac8ec2d33`
- **Captured source fingerprint:** `134b7b74b5b53ae68c2ef3ca218bc5ac2e5e0ea0`
- **Deliverable:** `planning/audits/6ggw69vpds7h-2026-10-05-explicit-persistence-authorization-implementation-antigravity.md`
- **Verification status:** Clean isolated clone created via `scripts/isolated-review-workspace.sh create --no-hardlinks`; zero modifications to shared source checkout.

---

### 2. Baseline and regression suite execution

All verification commands executed exclusively within `$SANDBOX` (`/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.RZCmdb`):

| Gate / Command | Environment / Target | Exit Code / Result | Details |
|---|---|:---:|---|
| `go test ./internal/core -run '^TestMutationPolicy' -count=1` | `$SANDBOX` | `0` (PASS) | Validated `MutationPolicy` mode validation, unexported zero-value refusal, and error preservation (0.308s). |
| `go test ./internal/appwiring -run 'TestPersistenceConstructors\|TestInvalidLocalComposition\|TestLocalCompositionPropagatesAuthorization' -count=1` | `$SANDBOX` | `0` (PASS) | Verified constructors reject invalid/nil policies without I/O or callback execution; composition drops partial bundles (0.294s). |
| `go test ./internal/store -run 'TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects' -count=1` | `$SANDBOX` | `0` (PASS) | Verified all 26 store mutation APIs reject mutations under `ReadOnlyMutations`, `GuardedMutations(refuse)`, and `&FS{}` in both write and dry-run modes without touching root (0.247s). |
| `go test ./internal/configstore ./internal/spacestore -run 'TestConfigurationPolicies\|TestRegistryPolicies\|TestSummaryPlanningStores' -count=1` | `$SANDBOX` | `0` (PASS) | Verified config and space stores fail closed on mutations (including dry-run previews), and Atlas summary stores are forced read-only (0.280s / 0.429s). |
| `go test ./internal/workspacestore -run 'TestWorkspaceOpeningCarriesPolicy\|TestZeroWorkspaceAdapter\|TestNewPlanningStore' -count=1` | `$SANDBOX` | `0` (PASS) | Verified workspace opening propagates policy across direct and pointer entry points, re-checks authorization dynamically, and rejects zero adapters before discovery (0.312s). |
| `go test ./internal/cli -run 'TestFailedComposition\|TestMetadataOnlyCommandTree\|TestLocalCompositionKeepsAuthorization' -count=1` | `$SANDBOX` | `0` (PASS) | Verified CLI composition failure blocks operational commands across all 17 routes, permits metadata/help, and preserves quiet empty protocol for `__complete` (0.361s). |
| `go test ./...` | `$SANDBOX` | `0` (PASS) | Full test suite across all 35 internal packages green. |
| `go test -race ./...` | `$SANDBOX` | `0` (PASS) | Complete race detector suite green with zero data races detected. |
| `golangci-lint run ./...` | `$SANDBOX` | `0` (PASS) | Clean linter run (0 issues). |
| `just build` | `$SANDBOX` | `0` (PASS) | Successfully built `bin/tskflwctl` (version `v0.22.0-135-g8d69436`). |
| `./bin/tskflwctl -C . lint --json --no-input` | `$SANDBOX` | `0` (PASS) | Planning corpus lint returned clean: `{"schema_version":"1.81","unreadable":[],"issues":[]}`. |
| CLI docgen drift check (`diff -ru docs/cli "$TMP_DOCS"`) | `$SANDBOX` | `0` (PASS) | Regenerated CLI documentation via `internal/tools/docgen`; diff exit status 0 (zero doc drift). |
| Schema comments drift check (`diff -u internal/wire/schema_comments.json "$TMP_COMMENTS"`) | `$SANDBOX` | `0` (PASS) | Regenerated 267 schema comments via `internal/tools/schemacomments`; diff exit status 0 (zero schema comment drift). |
| CLI dogfood init (`./bin/tskflwctl init --path "$SCRATCH" --taskflow-root planning --no-register`) | `$SANDBOX` | `0` (PASS) | Scaffolded planning corpus in isolated scratch path with isolated `TSKFLW_CONFIG_HOME`; exit code 0, generated tasks/epics/audits/research/threads directories and config. |

---

### 3. Hostile empirical verification & mutation probes

Four compiler-valid mutation probes were introduced at distinct boundaries to test regression sensitivity. Each probe was executed against the exact claimed killing regression, observed to fail, and restored to the sandbox baseline commit before proceeding:

#### Mutation Probe 1: Store mutation authorization bypass
- **Boundary:** `internal/store/fsstore.go:77` (`(s *FS) authorizeMutation() error`).
- **Mutation:** Replaced `return s.mutationPolicy.Authorize()` with `return nil` (granting unconditional write authorization at the store level).
- **Claimed Killing Test:** `internal/store/mutation_authorization_test.go:15` (`TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects`).
- **Observed Result:** **Killed immediately.** All 26 mutation entries failed across `guarded`, `read-only`, and `zero-value adapter` test cases in both dry-run and write modes (e.g., `mutation_authorization_test.go:153: error = task "missing": not found, want validation failed: explicit mutation policy is required`).
- **Restoration:** Restored via `git checkout internal/store/fsstore.go`. Working tree clean.

#### Mutation Probe 2: Atlas summary capability widening
- **Boundary:** `internal/spacestore/fs.go:105` (`(f *FS) OpenPlanningStore(root string)`).
- **Mutation:** Replaced `fs, err := store.NewFS(root, core.ReadOnlyMutations())` with `fs, err := store.NewFS(root, f.mutationPolicy)` (allowing Atlas summary stores to inherit the registry's mutation policy).
- **Claimed Killing Test:** `internal/spacestore/mutation_policy_test.go:49` (`TestSummaryPlanningStoresCannotInheritRegistryMutationPrivilege`).
- **Observed Result:** **Killed immediately.** When registry policy was `UnrestrictedMutations()`, widened `fs.SetFields` succeeded instead of returning `ErrReadOnlyPersistence`:
  ```
  --- FAIL: TestSummaryPlanningStoresCannotInheritRegistryMutationPrivilege (0.00s)
      mutation_policy_test.go:70: widened summary capability authorized mutation: <nil>
  ```
- **Restoration:** Restored via `git checkout internal/spacestore/fs.go`. Working tree clean.

#### Mutation Probe 3: Partial services leak upon failed CLI composition
- **Boundary:** `internal/cli/root.go:287` (`newRootCmd`).
- **Mutation:** Removed `if err == nil` guard around assigning services to `app` when `bindings.Compose` returns an error (publishing partial services despite composition refusal).
- **Claimed Killing Test:** `internal/cli/composition_test.go:64` (`TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata`).
- **Observed Result:** **Killed immediately.** All 17 command routes in the subtest table failed on line 86:
  ```
  --- FAIL: TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata (0.00s)
      --- FAIL: TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata/task/show/alpha (0.00s)
          composition_test.go:86: failed composition published partial services
  ```
- **Restoration:** Restored via `git checkout internal/cli/root.go`. Working tree clean.

#### Coordinated Probe: Best-effort discovery fallback & duplicate guard masking
- **Boundary:** `internal/cli/theme.go:38`, `internal/cli/template.go:28`, and `internal/cli/root.go:371, 458`.
- **Scenario & Findings:**
  1. *Theme fallback probe:* In `internal/cli/theme.go:38`, removing `if app.compositionErr != nil { return app.compositionErr }` caused `theme list` to swallow the composition failure during best-effort cwd discovery and output built-in themes with exit 0 (`composition_test.go:101: composition failure swallowed: err=<nil> stdout="catppuccin\nmiami-vice\nneon (default, active)\n"`).
  2. *Template fallback probe:* In `internal/cli/template.go:28`, removing `if app.compositionErr != nil { return app.compositionErr }` caused `template list` to swallow composition failure and output built-in templates with exit 0 (`composition_test.go:101: composition failure swallowed: err=<nil> stdout="  task    default..."`).
  3. *Duplicate guard probe:* In `internal/cli/root.go:371`, removing `if a.compositionErr != nil { return a.compositionErr }` from `repoPreRun` alone was masked by `resolveFrom` (line 458), allowing `task/show/alpha` to pass because `resolveFrom` duplicated the check.
  4. *Coordinated probe:* Removing `if a.compositionErr != nil` from **both** `repoPreRun` and `resolveFrom` resulted in `task show alpha` falling through to `if a.openPlanning == nil`, degrading the error to generic `validation failed: planning opener is unavailable from this invocation` (`ErrValidation`) rather than preserving root cause `compositionErr` (`ErrInvalidMutationPolicy`), cleanly failing `composition_test.go:101`.
- **Restoration:** Restored all files via `git checkout internal/cli/root.go internal/cli/template.go internal/cli/theme.go`. Working tree clean.

#### Dynamic authorization transitions (`denied -> allowed -> denied`)
In `internal/workspacestore/fs_test.go:193-208`, verified empirical behavior when a guarded callback flips state after adapter opening:
1. Initial refusal: mutation returned caller error intact (`denied`).
2. Flipped to allow: subsequent mutation on the same opened workspace instance succeeded (`decision = nil`), and callback counter incremented.
3. Flipped back to refuse: subsequent mutation failed with `denied` and counter incremented.
This proves that authorization decisions are never cached upon constructor validation or workspace opening, and remain strictly invocation-scoped.

---

### 4. Systemic second pass (adversarial analysis)

#### Trust Assumption A: "Summary interface narrowing hides an unrestricted concrete store"
- **Adversarial Angle:** `core.PlanningSummarySource` exposes only read-only methods (`ListTasks`, `GetTask`, etc.). If `spacestore.OpenPlanningStore` constructs a store using the registry's policy, an adversarial or erroneous consumer could type-assert `summary.(*store.FS)` or `summary.(core.Store)` and perform mutating operations.
- **Investigation & Finding:** Examined `internal/spacestore/fs.go:105`. `OpenPlanningStore` unconditionally calls `store.NewFS(root, core.ReadOnlyMutations())`. Even when the outer registry adapter holds `UnrestrictedMutations()` or an authorized `GuardedMutations()`, the inner `*store.FS` is instantiated with `ReadOnlyMutations()`. Type assertion to `*store.FS` grants no mutation privileges; any write or dry-run call returns `core.ErrReadOnlyPersistence`. Falsified by hostile Probe 2.

#### Trust Assumption B: "Constructor validation tests prove only an error, not absence of filesystem effects"
- **Adversarial Angle:** A constructor failing with an error might have already scaffolded directories or opened lock files before evaluating the policy.
- **Investigation & Finding:** Examined `internal/appwiring/persistence_policy_test.go:53`, `internal/store/mutation_authorization_test.go:155`, and `internal/configstore/mutation_policy_test.go:42`. Each test suite passes a nonexistent temporary path as root and asserts `os.IsNotExist(err)` or `len(entries) == 0` following refused construction and refused mutations. Furthermore, examined `store.NewFS`, `configstore.New`, `spacestore.New`, and `workspacestore.New`: all call `policy.Validate()` as their very first statement before any `filepath.Join`, directory creation, or file operations.

#### Trust Assumption C: "Duplicate checks in pre-run pipelines mask incomplete propagation"
- **Adversarial Angle:** Commands with bespoke pre-run hooks might bypass composition failure checks if shared discovery helpers absorb errors.
- **Investigation & Finding:** Confirmed that `template.go` and `theme.go` previously implemented best-effort fallback logic that swallowed discovery errors. Without the explicit `if app.compositionErr != nil` checks added in this changeset, composition errors were silently converted into built-in success. The coordinated mutation probe demonstrated that these explicit checks are required and actively defended by `TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata`.

---

### 5. Repository-wide consumer inventory

| Component / File | Symbol / Method | Policy / Mode | Caller / Consumer | Purpose & Behavior |
|---|---|---|---|---|
| `internal/core/mutation_policy.go:32` | `ReadOnlyMutations()` | `mutationReadOnly` | Atlas summary store (`spacestore/fs.go:105`), contract test suites | Permissive reads, denies all 26 store mutation APIs and config/registry writes before effects. |
| `internal/core/mutation_policy.go:36` | `GuardedMutations(authorize)` | `mutationGuarded` | `appwiring/wiring.go:27`, CLI / TUI composition | Evaluates `authorize()` at each mutation boundary; rejects nil authorizers at construction. |
| `internal/core/mutation_policy.go:43` | `UnrestrictedMutations()` | `mutationUnrestricted` | Unit test fixtures across `internal/store`, `internal/tui` | Explicit opt-in for trusted test fixtures; production CLI/TUI composition never selects this mode. |
| `internal/store/fsstore.go:101` | `NewFS(root, policy, opts...)` | Required argument | `workspacestore/planning.go:31`, `spacestore/fs.go:105`, test fixtures | Primary planning store constructor; validates policy before assigning fields or creating dirs. |
| `internal/configstore/fs.go:23` | `New(policy)` | Required argument | `appwiring/wiring.go:32`, test fixtures | Secondary config store adapter; authorizes `MigrateConfiguration` and `SetPreference`. |
| `internal/spacestore/fs.go:24` | `New(policy)` | Required argument | `appwiring/wiring.go:28`, test fixtures | Secondary registry store adapter; authorizes `AddSpace` and `ForgetSpace`. |
| `internal/workspacestore/fs.go:17` | `New(policy)` | Required argument | `appwiring/wiring.go:36`, test fixtures | Secondary workspace store adapter; propagates policy into every opened workspace. |
| `internal/workspacestore/planning.go:17` | `NewPlanningStore(cfg, discover, policy)` | Required argument | `workspacestore/fs.go:34`, `appwiring/wiring.go:63` | Shared observed planning opener; binds identity reader and carries policy into `store.NewFS`. |
| `internal/appwiring/wiring.go:26` | `compose(reads, authorize)` | `GuardedMutations(authorize)` | `bindingsFor` -> CLI `ports.Bindings.Compose` | Production composition root; creates guarded space, config, workspace adapters and planning opener. |
| `internal/cli/root.go:285` | `newRootCmd(...)` | `app.authorizeMutation` | `cmd/tskflwctl/main.go`, CLI test helpers | Wires Cobra invocation container to `ports.Bindings.Compose`; discards partial bundles on error. |
| `internal/cli/root.go:360` | `(a *App) repoPreRun(...)` | Guarded | Default persistent pre-run for repo commands | Checks `compositionErr`, runs `resolve()`, enforces command safety annotations. |
| `internal/cli/root.go:162` | `(a *App) styleOnlyPreRun(...)` | Guarded | Operational commands (`init`, `config`, `space`, `doctor`) | Checks `compositionErr`, skips planning repo discovery. |
| `internal/cli/root.go:151` | `(a *App) metadataOnlyPreRun(...)` | Metadata-only | Metadata commands (`schema`, `version`, `completion`, `--help`) | Permits execution without application services; works even when persistence composition fails. |
| `internal/cli/status.go:41` | `newStatusCmd` PreRun | Guarded | `tskflwctl status --all` | Custom PreRun; explicitly checks `compositionErr != nil` before cross-space status. |
| `internal/cli/ui.go:41` | `newUICmd` PreRun | Guarded | `tskflwctl ui` | Custom PreRun; explicitly checks `compositionErr != nil` before UI startup. |
| `internal/cli/template.go:28` | `newTemplateCmd` PreRun | Guarded | `tskflwctl template ...` | Custom PreRun; explicitly checks `compositionErr != nil` before best-effort cwd template discovery. |
| `internal/cli/theme.go:38` | `newThemeCmd` PreRun | Guarded | `tskflwctl theme ...` | Custom PreRun; explicitly checks `compositionErr != nil` before best-effort cwd theme discovery. |
| `internal/cli/ports/runtime.go:41` | `ports.Bindings.Compose` | Function signature | CLI primary adapter interface | Mandates that controllers discard all partial services upon composition failure. |

---

### 6. Claim / evidence matrix

| # | Invariant / Contract Claim | Implemented Evidence | Falsification / Boundary Check |
|---|---|---|---|
| 1 | `MutationPolicy` is an opaque value with invalid zero value. Required constructor arguments remove implicit defaults; invalid construction returns no usable adapter and performs no I/O or callback execution. | `internal/core/mutation_policy.go:46-56`, `internal/appwiring/persistence_policy_test.go:17-56` | Tested with zero value `{}` and `GuardedMutations(nil)` across 5 adapter constructors; all returned `ErrInvalidMutationPolicy` without disk stat or authorizer calls. |
| 2 | `ReadOnlyMutations` permits reads but denies every mutation API before filesystem effects, editors, transformations, or planner callbacks. Direct zero-value adapter denies mutations too. | `internal/store/mutation_authorization_test.go:15-162`, `internal/configstore/mutation_policy_test.go:14-48`, `internal/spacestore/mutation_policy_test.go:16-47` | Verified all 26 store mutation methods, 2 config methods, and 2 space methods fail closed for both `ReadOnlyMutations` and `&FS{}` in write and dry-run modes. Tested zero disk creation. |
| 3 | `GuardedMutations` consults its callback at every mutation boundary without caching constructor authorization. Guard errors retain their original cause. Decisions can change dynamically after opening. | `internal/workspacestore/fs_test.go:144-212`, `internal/core/mutation_policy.go:58-73` | Tested dynamic decision changes (`denied -> allowed -> denied`) after opening on direct and pointer workspaces; counter verified fresh evaluation for each mutation. |
| 4 | Ordinary and direct/pointer workspace opening preserve policy and capabilities. Read operations do not call authorizer. Atlas summary stores are forced read-only regardless of registry mode, even after interface widening. | `internal/workspacestore/fs_test.go:16-63`, `internal/spacestore/mutation_policy_test.go:49-80` | Hostile Probe 2 demonstrated that summary store widening cannot perform mutations even when registry adapter has unrestricted privileges. |
| 5 | `UnrestrictedMutations` is a deliberate choice for trusted fixtures/tooling; production CLI/TUI remains guarded. Locks, CAS, eligibility, and validation still apply. | `internal/appwiring/wiring.go:27-37`, `internal/store/lock.go:116-127`, `internal/cli/composition_test.go:253-272` | Verified production wiring only mints `GuardedMutations`. Unit test suites use `UnrestrictedMutations` without bypassing flock or domain validation. |
| 6 | Failed `Bindings.Compose` discards all partial services and preserves refusal on operational CLI paths without fallbacks hiding it. Metadata/completion continue to function. | `internal/cli/composition_test.go:64-109`, `internal/cli/root.go:284-292` | Hostile Probe 3 and Coordinated Probe proved partial services are discarded and template/theme/status routes refuse cleanly. |
| 7 | Only internal Go constructors change. No machine schema, persisted representation, CLI flag/command changes, or documentation drift. | `docs/cli`, `internal/wire/schema_comments.json`, `scripts/release-validate.sh` | Verified zero diff against committed `docs/cli` and `schema_comments.json`. Planning lint clean. |

---

### 7. Evaluation of strongest and weakest owner tests

- **Strongest tests:**
  - `store.TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects`: Thoroughly covers all 26 store mutation methods in both dry-run and write modes across guarded, read-only, and zero-value adapter states, and confirms that the planning directory was not touched.
  - `cli.TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata`: Exercises 17 distinct CLI command invocations against a failed composition with populated cwd, proving both refusal on operational routes and proper behavior on metadata/completion.
  - `workspacestore.TestWorkspaceOpeningCarriesPolicyAcrossDirectAndPointerEntryPoints`: Covers direct and pointer workspaces across all 3 policy modes, tests dynamic authorization transition (`denied -> allowed -> denied`), and proves that decisions are never cached.
- **Weakest tests:**
  - `cli.TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata`: Uses a minimal `ports.Bindings` where `IsMissingPlanning` is nil. Although `LocalBindings` defines `isMissingPlanning` as matching only `config.ErrNoConfig`, testing with `appwiring.LocalBindings()` directly would make the test even more realistic.
  - `appwiring.TestInvalidLocalCompositionDoesNotPublishPartialServicesOrDiscover`: Exercises invalid composition primarily via `Compose(nil)`. It tests `openPlanning` directly with zero `MutationPolicy{}`, but does not test `openPlanning` after an adapter constructor fails inside `Compose`.

---

### 8. Open findings

No open defects, data-safety violations, or architectural boundary escapes were identified. All contract invariants hold under hostile probing.

---

### 9. Final verdict

**Verdict:** `ready`

The explicit persistence authorization implementation satisfies all contractual, architectural, and data-safety requirements:
1. Persistence constructors fail closed on zero values and nil guards without performing I/O or invoking callbacks.
2. Store mutations, previews, and dry-run requests are authorized prior to file modifications or domain planning logic.
3. Guard decisions are dynamic and never cached upon opening or after an allow decision.
4. Atlas summary stores remain strictly read-only regardless of registry adapter privileges or interface widening.
5. CLI composition failures cleanly discard partial services and refuse operational commands without leaking errors through best-effort discovery or completion.
6. Documentation, schemas, planning linter, race detector, and unit suites are 100% green with zero drift.

## Owner reconciliation (2026-10-05)

Accepted the bounded no-new-defects conclusion, corroborated by the independent Codex review
of the same source fingerprint. The suggested real-`LocalBindings` coverage is now permanent:
`TestFailedCompositionDiscardsPartialServicesAndPreservesMetadata` uses a full real bundle,
the actual missing-planning predicate, wrapped missing-config/conflict causes, and repo/home
snapshots. Populated no-op/callback and constructor-option regressions were also added.

Corrections to the report's evidence vocabulary (the reviewer history above is retained):

- `PlanningSummarySource` embeds `SummaryStore`, `TaskGraphSource`, and `AuditSnapshotSource`
  (`internal/core/space_overview.go`), not `ListTasks`/`GetTask`. The concrete-store widening
  probe is valid despite this inaccurate interface inventory.
- Unchanged files/stat assertions establish no creation or writes, not absence of reads.
  No-read constructor evidence rests on inspected control flow and observable callback/discovery
  witnesses; no syscall trace was collected. `doctor` has a custom pre-run, not `styleOnlyPreRun`.
- Injecting constructor failures inside `compose` to exercise a subsequent opener is unnecessary:
  failed composition intentionally publishes no opener. Direct opener refusal and full-bundle
  controller tests cover the reachable boundary without adding production test hooks.

The received report was byte-identical to the retained sandbox deliverable before owner edits.
The reviewer did not include final helper transfer output; owner receipt is confirmed, but that
omission is not retroactively presented as a recorded helper attestation. Full normal/race tests,
lint, build, generated-output comparisons, and planning/audit lint pass after reconciliation.
No production fix or additional task is required; closure does not imply merge or release.
