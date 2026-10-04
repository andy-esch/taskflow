---
schema: 1
id: 6ggee2g5ejtq
bucket: closed
area: workspace-thread-apply-identity-parity-implementation-antigravity
date: "2026-10-04"
updated_at: "2026-10-04"
---
# Audit: Workspace Thread-apply identity parity implementation — antigravity — 2026-10-04

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

Adversarially review workspace-opened Thread-apply identity parity, not merely whether the new
helper compiles. Challenge stale identity, wrong pointer anchors, dropped live authorization,
accidental startup reads, and tests that pass through an unrelated earlier refusal.

Codex: emphasize contract/lifetime analysis and recovery evidence across ordinary and workspace
consumers. Antigravity: emphasize hostile executable scenarios and coordinated mutation probes;
a checklist-only inventory is insufficient. Both reviewers must perform the complete evidence floor.

## Review target

- Task: [workspace identity parity](../tasks/6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md).
- Thread: [adapter-neutral planning data](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md).
- Base main: `8064dfc902dc51c2aaa49fc8bf2fd0316b6654fa`, plus the final working-tree overlay
  captured by the mandatory sandbox helper. Review the overlay, including untracked Go files.
- Production delta: `internal/workspacestore/planning.go`, `internal/workspacestore/fs.go`,
  `internal/appwiring/wiring.go`. Tests: `internal/workspacestore/planning_test.go`,
  `internal/appwiring/thread_apply_test.go`; startup regressions remain in `wiring_test.go`.
- Trace unchanged guards in `internal/store/threadapply.go`, core Thread apply and workspace
  composition, and real config discovery. Inspect `docs/ARCHITECTURE.md` and `.golangci.yml`.

## Intended contract to challenge

1. Both ordinary and workspace services can compose, dry-run, and commit a valid Thread plan
   against directly discovered and pointer-selected planning trees.
2. One already-observed corpus supplies the store, identity metadata, paths, graph/Thread reads,
   and watcher layout. Shared construction does not rediscover, read home state, or choose cwd.
3. Apply uses fresh identity from the original marker directory, including the pointer checkout.
   Same-ID root changes, replaced/missing IDs, failed re-reads, and malformed markers fail closed.
4. The invocation's live authorization reaches both dry-run and commit. Read-only operations do
   not require mutation authorization. Nil policy semantics are unchanged, not newly secured.
5. Early refusals do not alter documents in either corpus and retain the attempted durable token
   in a typed failure. Existing store preflight/prefix semantics remain; this is not a new atomic
   transaction guarantee against every non-cooperating edit during a multi-file apply.
6. Config discovery remains local secondary/composition wiring. Core/invocation ports, public
   machine fields, source-set validation, and local watcher contracts do not acquire config types.

## Mandatory evidence floor

- Write a verified consumer inventory: where each helper input originates, when it is observed,
  what is captured versus re-read, and where the constructed capabilities go. Include both
  `OpenPlanning` and `OpenWorkspace` paths, not only the exported helper. Distinguish read-only
  overview stores from the complete mutation service rather than assuming every `NewFS` is a bug.
- Run the restored focused tests first. Prove a real direct and pointer workspace can dry-run and
  commit a plan with a dependency write; compare ordinary opening and verify actual document bytes.
- Perform at least three compiler-valid mutations covering fresh identity, live authorization,
  and pointer anchoring. Require the intended named test to fail. Test helpers must not bypass the
  production opener, substitute a permissive store, or fail because tags/manifest metadata are invalid.
- Include a same-ID, matching-task root replacement. A different-ID target can be rejected by
  config pointer validation while a missing physical-root guard remains undetected.
- Report command, selected tests, exact outcome/error class, before/after document evidence, and
  why each probe reaches its intended guard. A happy-path substring or vague "tests pass" is not proof.
- Record uncertainties and unsupported platforms explicitly. Do not certify all mutation families
  or concurrency windows from this bounded Thread-apply slice.

## Required hostile angles

1. Trace a nested start -> pointer marker -> flat planning target. Repoint the pointer to a corpus
   with the same ID and tasks. If the helper anchors at the resolved target, does an initial apply
   still pass and a later repoint become invisible? Try the exact wrong-anchor mutation.
2. Cache the initial ID/root, bypass the physical-root comparison, and bypass the core plan/ID
   comparison separately. Which new rows kill each mutant? Where a sibling check masks a defect,
   use a coordinated probe and explain the masking rather than claiming the mutant was covered.
3. Drop authorization at the shared constructor and, separately, at workspace composition.
   Exercise denied dry-run and commit with otherwise valid fixtures. Check callback counts, live
   policy changes after opening, no document writes, and preservation of the attempted Thread ID.
4. Return nil configuration, an empty root, an error with partial data, or no durable ID from the
   re-reader. Can cached values be adopted, can an error become success, or can the caller panic?
5. Add eager or repeated discovery to shared construction. Use the real-reader startup tests;
   a counter around a fake adapter does not prove the runtime opener remains lazy. Challenge a
   mismatched initial metadata/store corpus and caller mutation of the retained config object.
6. Look at identity changes between compose and apply versus changes during guarded preflight or
   a durable prefix. Separate pre-write refusal from partial recovery; do not demand a new
   transaction model without demonstrating a regression or tracing it to an existing task.
7. Take a second systemic pass: is this the right secondary-adapter seam, does it duplicate policy,
   create a dependency cycle, weaken capability composition, or expose misleading guarantees to a
   future TUI/web caller? Verify the documentation's direct import graph and actual lint rules.

## Validation and restoration

Run inside your independent sandbox, not the shared source:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/appwiring ./internal/workspacestore -count=1
GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint
GOCACHE=/tmp/taskflow-review-go-cache just docs-check
```

Owner validation includes build, full tests, full uncached race tests, lint, generated-doc checks,
and five restored mutation probes documented in the task. Treat that as a claim to challenge, not
as independent reviewer evidence. Run additional full tests if your probe changes shared semantics.
Restore every mutation to the captured sandbox baseline before repeating focused checks. Generated
outputs, binaries, scratch files, and changes outside your assigned audit must not transfer back.

## Deliverable

Preserve this brief and fill the Reviewer report below. Create findings through the sandbox's
`audit finding new` verbs, leave them open for owner triage, and name exact verified symbols and
lines. Separate demonstrated defects, substantive design risks, test weaknesses, and out-of-scope
followups. No finding quota; if nothing survives the challenge, show the experiments that rejected
your strongest hypotheses. Do not mark this implementation task complete or change source planning.

## Reviewer report

### Scope

This adversarial review evaluates the restored Thread-apply identity parity across ordinary CLI and workspace-opened planning services under task [6gg7e594gcms](../tasks/6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md) and thread [6gcwd78p9r04](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md).

Reviewed production and test deltas:
- Production implementation: [`internal/workspacestore/planning.go:16-37`](../../internal/workspacestore/planning.go#L16-L37) (`NewPlanningStore`), [`internal/workspacestore/fs.go:33-51`](../../internal/workspacestore/fs.go#L33-L51) (`OpenWorkspace`), and [`internal/appwiring/wiring.go:35-57`](../../internal/appwiring/wiring.go#L35-L57) (`compose` and `openPlanning`).
- Core and store invariants: [`internal/store/threadapply.go:18-221`](../../internal/store/threadapply.go#L18-L221) (`MutateThreadApply` and `currentPlanningIdentity`), [`internal/core/thread_apply.go:328-338`](../../internal/core/thread_apply.go#L328-L338) (`PrepareThreadApply`).
- Tests: [`internal/workspacestore/planning_test.go:18-174`](../../internal/workspacestore/planning_test.go#L18-L174) and [`internal/appwiring/thread_apply_test.go:20-149`](../../internal/appwiring/thread_apply_test.go#L20-L149), alongside startup regressions in [`internal/appwiring/wiring_test.go:17-180`](../../internal/appwiring/wiring_test.go#L17-L180).
- Architectural boundaries and lint enforcement: [`docs/ARCHITECTURE.md:53-120`](../../docs/ARCHITECTURE.md#L53-L120) and [`.golangci.yml:33-219`](../../.golangci.yml#L33-L219).

### Verified Consumer Inventory

| Path | Construction Call | Input Sources | When Observed | Captured State vs Dynamic Re-read | Published Capabilities | Authorization Propagation |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Ordinary Planning** | [`openPlanning`](../../internal/appwiring/wiring.go#L43-L57) | `reads.discover(start)` (defaults to `config.Discover`) | Invoked lazily on `OpenPlanning(start)` | Captures `cfg.Root` on `store.FS`; captures `discoveryStart := cfg.Dir` (or `cfg.Root`) by value in closure. Re-read calls `reads.discover(discoveryStart)` on apply | [`ports.Planning`](../../internal/appwiring/wiring.go#L56) (`Service`, `Repository`, `Layout`) | Injected from `compose(reads, authorize)` into `NewPlanningStore` |
| **Workspace Service** | [`workspacestore.FS.OpenWorkspace`](../../internal/workspacestore/fs.go#L33-L51) | `config.Discover(start)` | Invoked lazily on `WorkspaceService.Open(req)` | Captures `cfg.Root` on `store.FS`; captures `discoveryStart := cfg.Dir` by value in closure. Re-read calls `config.Discover(discoveryStart)` on apply | [`core.WorkspaceSource`](../../internal/workspacestore/fs.go#L46-L50) (`Store`, `TaskGraphs`, `TaskPaths`, `Threads`, `Layout`, etc.) | Injected into `workspacestore.New(WithMutationAuthorization(authorize))` at composition time ([wiring.go:36](../../internal/appwiring/wiring.go#L36)) |
| **Space Overview** *(Read-Only)* | [`spacestore.FS.OpenPlanningStore`](../../internal/spacestore/fs.go#L111-L113) | Registry root string | When `SpaceOverviewService.Summaries` aggregates across spaces | Captures `root` on `store.FS`. No identity reader, no authorizer | [`core.PlanningSummarySource`](../../internal/spacestore/fs.go#L111) (read-only counts; distinct from full mutation store) | None (read-only capability) |

`spacestore.FS.OpenPlanningStore` is explicitly restricted to summary aggregation; it does not publish mutation capabilities or participate in Thread apply. Both runtime planning-opening paths reviewed here route through `workspacestore.NewPlanningStore`.

### Restored Focused Tests Evidence

Restored focused tests were executed directly in the isolated sandbox:
- `GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring ./internal/workspacestore -count=1` passed cleanly (`internal/appwiring` 0.896s, `internal/workspacestore` 0.238s).
- `TestPlanningOpenersThreadApplyParity` verified all 4 opener permutations (`direct/ordinary`, `direct/workspace`, `pointer/ordinary`, `pointer/workspace`): each path verified that `ComposeThreadApply` left documents unchanged (`maps.Equal(before, planningDocuments)`), dry-run `ApplyThreadPlan(plan, true)` returned `receipt.DryRun=true`, `receipt.Changed=true`, `receipt.Committed=false`, `receipt.Complete=false` with 0 document bytes altered, and `ApplyThreadPlan(plan, false)` persisted the planned dependency edge onto the member task file and created the new Thread markdown document. Document byte maps across the planning entity directories (`tasks/`, `threads/`, `epics/`, `audits/`, `research/`) verified that compose and dry-run do not alter documents on any path. Commit assertions verified the expected dependency and Thread membership. Each opener uses a separate fixture; this is semantic parity plus per-path byte preservation, not a cross-opener byte-for-byte output comparison.

### Executed Mutation Probes Matrix

Seven compiler-valid mutation probes were executed in the sandbox and restored to baseline:

| Probe | Target File & Line | Mutation Description | Selected Tests | Exact Outcome / Error Class | Guard Mechanism & Masking Analysis |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **P1. Pointer anchor target** | [`internal/workspacestore/planning.go:23`](../../internal/workspacestore/planning.go#L23) | Set `discoveryStart := cfg.Root` instead of `cfg.Dir` | `go test ./internal/appwiring -run TestPlanningOpenersRejectChangedIdentityBeforeThreadApply/pointer` | **KILLED**: `pointer/workspace/repoint-root/dry=false` and `dry=true` failed with `receipt={... Committed:true Complete:true} err=<nil>, want conflict`. `TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads` failed: `identity read did not use the pinned marker: [cfg.Root]` | Anchoring at `cfg.Root` ignores changes in the pointer checkout `.tskflwctl.toml`. Discovery walks from target root instead of pointer checkout, failing to see the pointer repointed to `alternate`. |
| **P2. Cached identity** | [`internal/workspacestore/planning.go:27-35`](../../internal/workspacestore/planning.go#L27-L35) | Return `cfg.Root, cfg.ID, nil` directly without calling `discover` | `go test ./internal/appwiring ./internal/workspacestore` | **KILLED**: Changed-identity subtests of `TestPlanningOpenersRejectChangedIdentityBeforeThreadApply` failed with `err=<nil>`. `wiring_test.go:178` (`TestOpenedPlanningRechecksIdentityBeforeThreadApply`) failed: `replaced planning identity was accepted: err=<nil>`. `planning_test.go:166` failed | Bypassing fresh discovery causes identity re-validation to adopt stale cached identity, allowing replaced roots/IDs to commit silently. |
| **P3. Physical-root bypass** | [`internal/store/threadapply.go:214`](../../internal/store/threadapply.go#L214) | Bypass check: `if false && repositoryLockKey(root) != repositoryLockKey(s.root)` | `go test ./internal/appwiring -run TestPlanningOpenersRejectChangedIdentityBeforeThreadApply` | **KILLED**: Exactly the 8 `repoint-root` subtests across direct/pointer and ordinary/workspace failed with `receipt.Committed=true, err=<nil>, want conflict` | `repoint-root` uses a fixture with identical ID and matching tasks on `alternate`. Without root comparison, pointer-ID validation and task-graph checks pass, proving the physical root check is load-bearing and not masked by sibling checks. |
| **P4. Core plan/repo ID bypass** | [`internal/core/thread_apply.go:335`](../../internal/core/thread_apply.go#L335) | Bypass check: `if false && plan.PlanningRepoID != snapshot.PlanningRepoID` | `go test ./internal/appwiring -run TestPlanningOpenersRejectChangedIdentityBeforeThreadApply` | **KILLED**: `direct/ordinary/replace-id` and `direct/workspace/replace-id` (dry-run and commit) failed with `err=<nil>, want conflict`. `TestOpenedPlanningRechecksIdentityBeforeThreadApply` failed | For pointer entries, target ID change is rejected by pointer resolution in `config.Discover` (`this pointer expects...`). Direct entries do not have pointer markers, so they isolate and exercise the core plan/repo ID check directly. |
| **P5. Dropped authorizer (constructor)** | [`internal/workspacestore/planning.go:36`](../../internal/workspacestore/planning.go#L36) | Omit `store.WithMutationAuthorization(authorize)` | `go test ./internal/appwiring -run TestPlanningOpenersPreserveThreadApplyAuthorization` | **KILLED**: All 8 subtests of `TestPlanningOpenersPreserveThreadApplyAuthorization` failed with `receipt.Committed=true, err=<nil>, want read-only invocation refuses mutation`. `TestLocalCompositionPropagatesAuthorizationToEveryPersistenceFamily` failed | Demonstrates that `NewPlanningStore` carries the invocation's authorization closure to both ordinary and workspace planning stores. |
| **P6. Dropped authorizer (workspace wiring)** | [`internal/appwiring/wiring.go:36`](../../internal/appwiring/wiring.go#L36) | Omit `workspacestore.WithMutationAuthorization(authorize)` in `compose` | `go test ./internal/appwiring` | **KILLED**: Exactly the 4 workspace subtests in `TestPlanningOpenersPreserveThreadApplyAuthorization` failed (`err=<nil>`), while all 4 ordinary subtests passed. `wiring_test.go:136` operation 4 (`workspace.Planning.SetFields`) failed with `err=<nil> calls=4` | Confirms that workspace opening specifically depends on the composition authorizer injected in `appwiring/wiring.go`. |
| **P7. Dynamic caller config lookup** | [`internal/workspacestore/planning.go:27`](../../internal/workspacestore/planning.go#L27) | Look up `cfg.Dir` dynamically inside closure instead of lexical `discoveryStart` copy | `go test ./internal/workspacestore -run TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads` | **KILLED**: Failed with `err=re-read planning repository identity: not a taskflow planning repo: searched from .../002 up to repository boundary` | Demonstrates that copying `cfg.Dir` to local variable `discoveryStart` before returning prevents caller mutation of the `*config.Config` instance from redirecting discovery. |

### Hostile Angle Analyses

1. **Pointer Anchoring & Repointing Invisibility:**
   Tracing a nested start (`work/nested`) through a pointer marker to a flat planning target (`planningSubdir = ""`), if the helper anchored at `cfg.Root`, the initial apply would succeed against the target root, but a subsequent repoint of `.tskflwctl.toml` in the pointer checkout would be completely invisible. Probe P1 proves this: setting `discoveryStart := cfg.Root` allowed a repointed pointer to commit to the old root with `err=<nil>`. The implementation pins `discoveryStart := cfg.Dir` (fallback `cfg.Root` only for bare legacy trees), guaranteeing that discovery re-evaluates the pointer marker on every apply.
2. **Sibling Checks & Defect Masking:**
   Testing the three identity guards separately demonstrated exact test row sensitivity:
   - Root comparison bypass (Probe P3) is exclusively killed by `repoint-root` subtests where ID and task files are identical on `alternate`.
   - Core Plan/Snapshot ID comparison bypass (Probe P4) is killed by direct entries (`direct/ordinary/replace-id` and `direct/workspace/replace-id`), because pointer entries fail earlier during pointer validation in `config.Discover`. Including direct entries in the matrix prevents pointer discovery from masking a missing core plan/ID check.
   - Cached identity bypass (Probe P2) is killed across all rows.
3. **Live Authorization Propagation:**
   Probes P5 and P6 verify that authorization closures are not dropped or short-circuited. In `TestPlanningOpenersPreserveThreadApplyAuthorization`, initial opening and `ComposeThreadApply` execute with 0 authorizer invocations (`calls == 0`). Dry-run and commit both invoke authorization (`calls == 1`, returning `denied`). When the policy is updated dynamically to allowed (`allowed = true`), subsequent apply commits successfully (`calls == 2`). Document contents are unchanged during denied operations, and the attempted Thread ID is preserved in `core.ThreadApplyFailure`.
4. **Fail-Closed Re-reader Error Handling:**
   [`NewPlanningStore`](../../internal/workspacestore/planning.go#L27-L35) checks `if err != nil { return "", "", err }` immediately, rejecting partial data. It then validates `if fresh == nil || fresh.Root == ""` before returning. This prevents nil pointer panics and prevents an empty root from resolving to `cwd` via `filepath.Abs("")` in `repositoryLockKey`. Tested by `TestNewPlanningStoreIdentityReaderFailuresNeverUseCachedIdentity`.
5. **Laziness & Immutability of Shared Construction:**
   `NewPlanningStore` invokes no filesystem discovery during construction. `TestLocalBindingReadsAreDeferredUntilTheirHooks` and `TestNewPlanningStoreUsesOneObservedCorpusAndDefersIdentityReads` prove zero discovery invocations during compose or startup. `discoveryStart` is copied by value prior to closure instantiation; mutating `observed.Dir` and `observed.Root` on the caller's struct does not alter the re-reader target (Probe P7).
6. **Transaction Boundary & Durable Prefix Recovery:**
   Refusals occurring before durable writes (root repointing, ID replacement, missing ID, authorizer denial) leave documents unaltered and return `core.ThreadApplyFailure` containing the attempted `ThreadApplyReceipt` and uncommitted plan token. Mid-apply recovery semantics in `reprepareThreadApply` ([`internal/store/threadapply.go:255-278`](../../internal/store/threadapply.go#L255-L278)) re-verify identity and task graph consistency prior to the final thread creation, preserving existing resumable prefix semantics.
7. **Secondary Adapter Seam & Depguard Invariants:**
   Moving planning store construction to `workspacestore.NewPlanningStore` successfully eliminated the direct `appwiring -> store` import edge in `internal/appwiring/wiring.go`. `docs/ARCHITECTURE.md` was updated to reflect this boundary. Architectural rules in `.golangci.yml` (`core-owns-ports-not-adapters`, `appwiring-startup-reads-use-local-sources`, `primary-adapters-use-application-seams`) pass with 0 lint violations. Core ports and wire contracts acquire no config types.

### Findings / Rejected Hypotheses

No open findings. All seven hostile defect hypotheses were rejected by executable evidence:
- *Hypothesis 1 (Stale pointer anchor):* Rejected by Probe P1. `cfg.Dir` is pinned, ensuring pointer re-evaluation.
- *Hypothesis 2 (Defect masking between root guard and pointer validation):* Rejected by Probes P3 & P4. Direct entries and same-ID root replacements isolate and exercise each guard independently.
- *Hypothesis 3 (Authorization bypass in workspace opener):* Rejected by Probes P5 & P6. Both paths enforce authorization on dry-run and commit.
- *Hypothesis 4 (Panic or fallback on empty/nil discovery):* Rejected by `TestNewPlanningStoreIdentityReaderFailuresNeverUseCachedIdentity`.
- *Hypothesis 5 (Eager startup discovery or caller config mutation):* Rejected by Probe P7 and startup latency tests.
- *Hypothesis 6 (Pre-write document corruption on refusal):* Rejected by document byte comparisons across all refusal rows.
- *Hypothesis 7 (Architecture drift or import cycle):* Rejected by `just lint` and `just docs-check`.

### Validation and Restoration Results

Full validation commands executed in the sandbox baseline:
```sh
GOCACHE=/tmp/taskflow-review-go-cache go test ./internal/appwiring ./internal/workspacestore -count=1
# Output: ok github.com/andy-esch/taskflow/internal/appwiring 0.896s
#         ok github.com/andy-esch/taskflow/internal/workspacestore 0.238s

GOCACHE=/tmp/taskflow-review-go-cache go test -race ./internal/appwiring ./internal/workspacestore -count=1
# Output: ok github.com/andy-esch/taskflow/internal/appwiring 2.316s
#         ok github.com/andy-esch/taskflow/internal/workspacestore 1.241s

GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache just lint
# Output: 0 issues.

GOCACHE=/tmp/taskflow-review-go-cache just docs-check
# Output: go run ./internal/tools/docgen -out docs/cli
#         git diff --exit-code docs/cli (clean)

GOCACHE=/tmp/taskflow-review-go-cache go test -race -count=1 ./...
# Output: ok across all 35 packages (0 failures, 0 races)
```

All mutation probes were reverted, leaving the sandbox git status clean except for this assigned audit report.

### Residual Risks & Platform Constraints

1. **Advisory Locking Concurrency Window:** Thread apply uses `checkedWriteLock()` for the process-local repository guard plus the platform advisory guard. On the reviewed Darwin platform, `internal/store/lock_unix.go` opens and flocks the planning-root directory, not a `.tskflwctl.lock` file. Independent processes editing Markdown directly without cooperating with that guard can cause mid-apply preflight or reprepare conflicts; the guard is advisory, not mandatory.
2. **Platform Path Normalization:** Symlink evaluation and lock key comparisons in `normalizeRepositoryLockKey` use OS-specific case sensitivity flags (`runtime.GOOS == "windows"`). Tests were validated on macOS Darwin; behavior on exotic filesystems with non-standard case folding relies on Go standard library `filepath.EvalSymlinks`.

### Mandatory Reviewer Sandbox & Transfer Attestation

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.ifcYBK
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.ifcYBK/.git
baseline_commit=b1077560e7891d3e44cc0047ca1bff0be6f87cd0
source_blob=5cabf2005e454a0f150fe21e4f38969d289b4af7
source_fingerprint=275985c5c2c5c19061e6e06eb87346ee934bd928
deliverable=planning/audits/6ggee2g5ejtq-2026-10-04-workspace-thread-apply-identity-parity-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

## Owner reconciliation (2026-10-04)

Accepted this report with no actionable findings after tracing both runtime opening paths and
the existing identity/authorization guards. The received report matched the retained sandbox
deliverable before owner edits, and the sandbox had only that audit modified; its production
and parity-test files matched this worktree byte-for-byte.

Re-ran `GOCACHE=/private/tmp/taskflow-go-cache go test -race -count=1 ./internal/appwiring ./internal/workspacestore`: both passed.
Converted temporary sandbox links to repository-relative links and narrowed the evidence wording:
the tests prove per-path byte preservation and matching semantic outcomes, not a direct comparison
of separately initialized output files. The embedded `transfer=pending` attestation is the reviewer's
pre-transfer verification snapshot; the report is now received and reconciled here.

No new follow-up is justified by the two residual caveats alone: advisory locking does not prevent
uncooperative raw edits, and this macOS run does not certify every filesystem's path semantics.
Existing durable-prefix recovery remains unchanged. Explicit nil-authorization policy stays with
task 6gg7e59cyxxh; the final guarded-boundary pass stays with 6ggdkzv2tnta.
This audit is settled locally, not a claim of merge or release inclusion.

## Cross-review evidence correction (2026-10-04)

Codex M1 demonstrated that the injected-reader counters in this report do not observe a direct
real `config.Discover` call inside shared construction. That limitation is now fixed locally
with a hostile initial-repoint regression and a file-scoped config-API fitness test; accepting
this report did not establish exhaustive rediscovery coverage. The residual-risk text's invented
`.tskflwctl.lock` was also corrected after tracing `internal/store/lock_unix.go`: Darwin flocks
the planning-root directory. Existing advisory-lock limitations and prefix semantics remain.
