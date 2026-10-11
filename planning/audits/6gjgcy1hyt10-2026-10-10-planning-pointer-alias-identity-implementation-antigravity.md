---
schema: 1
id: 6gjgcy1hyt10
bucket: closed
area: planning-pointer-alias-identity-implementation-antigravity
date: "2026-10-10"
---
# Audit: Planning pointer alias identity and CLI-contract closeout — antigravity — 2026-10-10

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

Challenge the final single-governing-marker pointer-identity fix and the CLI-contract
Thread closeout. Is an initialized/migrated pointer to a configured data directory
usable by its next discovery without borrowing an unrelated identity, creating another
marker, weakening expected-ID checks, or changing worktree anchoring?

Codex owns the identity/governance boundary and shared-consumer systemic pass.
Antigravity owns the actual operator workflow, diagnostic execution and independent
evidence quality. Both review the shared invariants and cross-check one hostile case
from the other lens. Review independently; do not read the sibling report first.

## Review target

- Read-only source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
- Branch `fix/planning-pointer-alias-identity`, base `fd0d97adf3ca` (main, PR #293).
  The implementation and planning closeout are **uncommitted**. Capture staged,
  unstaged and untracked files with the mandatory independent-clone helper; HEAD
  alone does not contain the review target.
- Implementing task:
  [6gjdgpq17sr6](../tasks/6gjdgpq17sr6-preserve-planning-identity-for-pointers-addressing-data-subdirectories.md).
  Thread:
  [6gbn4v2jpf2m](../threads/6gbn4v2jpf2m-make-cli-contracts-self-describing-and-bounded.md).
  Checked ACs and completed local planning state reflect owner evidence, not your verdict
  or a claim of merge/release. Report a contrary finding rather than trusting them.
- Production: `internal/config/config.go` (`Discover`, `discoverDataRoot`,
  `resolveRootIdentity`, `resolvePlanningTarget`, `governingConfig`,
  `verifyPlanningRepoID`, `InitPointer`, plus `configForRoot` link consumers);
  `internal/config/migrate.go` (`migrateUnlocked`, `migrationConfig`).
- Changed fixtures: `internal/config/planning_target_test.go` and
  `internal/cli/thread_recovery_test.go`. Existing discovery, identity, linkback and
  worktree tests are relevant adjacent evidence. Inventory actual consumers of the
  returned root/ID/config directory in `internal/configstore/fs.go`, local app wiring,
  configuration preference loading and identity-bound Thread replay. Verify names in
  the captured sandbox rather than assuming this inventory is exhaustive.
- Docs: README configuration lifecycle and the `internal/config` architecture entry.
  ADR-0008 governs machine revision policy. No JSON shape/selector/vocabulary was
  added; owner treats this as restoring documented single-marker alias semantics,
  keeping revision 1.85. Challenge this rationale if you demonstrate an incompatible
  documented meaning or newly accepted/refused valid input.
- Explicit separate follow-up:
  [6gjgccnzxj0a](../tasks/6gjgccnzxj0a-decide-and-enforce-identity-authority-for-overlapping-planning-markers.md).
  A real owner CLI probe initialized both `<repo>` (`taskflow_root=planning`) and
  `<repo>/planning` (`taskflow_root=.`), producing two markers/IDs for one physical
  tree; two pointers then reported different IDs for the same planning_root. This is
  known, not disproved by the new tests. It needs its own authority/recovery policy.
  You may challenge deferral, but distinguish that reproduced topology issue from a
  new failure for a single-governing-marker alias. Do not silently implement a policy.

## Intended contract to challenge

1. Config-directory, configured-data-directory and symlink addresses of a tree
   governed by one scaffold marker resolve the same physical root and durable ID.
   InitPointer and migration capture the same authority that later validation uses.
2. Ancestor identity requires an exact, containment-checked taskflow_root match.
   Nearest markers and physical Git-file/directory boundaries stop inheritance.
   Pointer ancestors are never followed as scaffold authority; direct pointer chains
   remain refused. Malformed governing TOML is loud, not a silent ID-less fallback.
3. Relative external pointers remain canonical-checkout anchored from worktrees.
   In-tree taskflow_root remains local to its own checkout. A plausible decoy cannot win.
4. Opted-in missing/mismatched target IDs remain conflict; legacy markerless targets
   remain usable without inventing an ID. Existing pointer init cannot treat a changed
   identity as an unchecked no-op. Missing-ID recovery targets the governing marker,
   preserves documents/unknown TOML and never rewrites the old pointer expectation.
5. Direct data-directory discovery retains governing marker identity/preferences and
   gives migration the actual marker location. It does not borrow unrelated parent
   preferences, silently adopt a pointer ancestor, or create another marker.
6. This is local configuration-adapter work, not filesystem logic in core/ports,
   a transactional lock across arbitrary external configuration edits, a pointer-chain
   implementation, or automatic overlapping-marker recovery.

## Mandatory evidence floor

Both reviewers execute focused tests and one real-command hostile probe in their own
sandbox. A green suite alone is not sufficient. Record exact commands/fixture bytes,
root and ID observations, before/after pointer/config/document snapshots, results and
limits. Build a fresh sandbox binary outside its tree; do not use an installed binary.

Codex:

- Probe one configured custom-subdirectory alias through init or legacy migration,
  then use direct and pointer discovery/replay. Separately try an unrelated parent ID,
  a nearest marker, and a Git boundary; settle whether any can lend an identity.
- Run **one compiler-valid mutation** disabling governingConfig's exact-root authority
  check (not the TOML parser or test). Require
  `TestPlanningTarget_AliasCannotBorrowUnrelatedIdentity/different-root` to fail for
  the intended successful-wrong-identity reason. Restore and run green. If another
  caller blocks the mutant first, coordinate the mutation or mark it unresolved;
  compilation failure is not killed evidence.

Antigravity:

- From an unrelated cwd and isolated home, use fresh CLI init/migration for a data
  alias and run Thread compose via that alias, then apply via the governing config
  directory. Check actual saved pointer expectation and plan binding, not only JSON
  command success. Use a non-default custom layout and unknown TOML/comment content.
- Remove only the governing marker's ID after binding. Execute the diagnostic's
  actual quoted recovery command. Prove it migrates the existing marker, preserves
  documents/home registry and the old pointer bytes, creates no inner marker, and
  leaves the bound pointer in conflict with the newly minted identity.
- Run **one compiler-valid mutation** restoring the old alias validation failure:
  pass an empty target ID to expected-ID verification only for the ancestor-marker
  branch (leave initialization/migration ID capture intact). Require an exact data
  alias case of `TestPlanningTarget_BindingRoundTrip` to fail on its next discovery,
  then restore and run green. Explain what bug class the fixture actually catches.

Each cross-checks one hostile case from the other lens. Do not invent findings to
meet a quota. Separate reproduced product bugs, test coverage gaps and preferences.
No changed optional wire branches exist here; do not pretend empty schema tests
establish identity behavior. If you discover a changed machine field, test it with
non-default values and trace the reachable producer.

## Required hostile angles

- Physical versus lexical ancestry: data-directory symlinks, an escaping configured
  root, markers in intermediate directories, custom/deep nesting and Git files.
  Does a marker govern the exact tree, or merely sit nearby? Is malformed TOML
  hidden by a convenience boolean helper in any binding path?
- Configless compatibility: bare tasks/ and conventional planning/tasks/, ID-less
  markers, unrelated valid/malformed parent markers, pointer ancestors/self-reference.
  Identify any changed fallback semantics and distinguish invalid topology from a
  previously documented valid input. Use bounded probes; never let a recursion probe
  hang without a timeout.
- Canonical-worktree anchoring with a decoy at the naive relative location; confirm
  ordinary in-tree worktree roots are not redirected to main. Verify existing linkback
  and CheckLinks consumers still select the correct marker without following chains.
- Snapshot seams: root and ID now share a resolution observation, but the code does
  not claim a global transaction. Examine subsequent config reads/preferences/replay;
  show a reachable wrong-target or wrong-identity consequence before asserting TOCTOU
  corruption. Label existing broader races and distinguish them from this change.
- Separate systemic pass: direct alias discovery now returns governing Config.Dir,
  preferences and tracked_repos rather than a bare root. Trace real startup, migration,
  description and style consumers for accidental preference/identity or write-target
  changes. Check that tests do not conflate caller cwd, explicit -C, pointer address,
  data root and governing marker directory.
- Challenge the closeout and overlapping-marker scope boundary with evidence, not
  merely by listing the known follow-up again. Is any unsupported claim in the docs
  broader than what single-marker tests prove?

## Validation and restoration

Owner evidence for this dirty snapshot: `just build`, `just test` (full race suite),
`just lint` (zero issues), `just docs-check`, planning lint and `git diff --check` pass.
You may inherit these broad gates, but independently execute:

```sh
GOCACHE=<scratch-cache> go test ./internal/config ./internal/cli ./internal/appwiring -count=1
```

Use separate scratch paths outside the sandbox for binary, temporary fixtures, Go
cache and isolated `TSKFLW_CONFIG_HOME`. No real registry changes; no network or source
Git operations are needed. Build via `go build -o <scratch>/tskflwctl ./cmd/tskflwctl`.
Restore all probe code to the helper's captured sandbox baseline; only your assigned
audit may differ. Verify and transfer with the mandatory helper protocol. Leave the
sandbox for owner cleanup. On inability to execute/restore/verify, report incomplete
with surviving uncertainty; never perform fallback probes in the shared checkout.

## Deliverable

Keep the brief. Fill only your assigned audit's report slot, findings first with open
statuses via the sandbox's local audit tooling. Give severity, root cause, reachable
consumer trace, reproduction, recommendation and residual uncertainty. Cite paths
and exact evidence; no invented symbols or claims that planning is executable code.

Include verdict/limits, verified consumer inventory, concise probe and mutation ledger
(exact edit/test, killed/survived/invalid, restored green result), a bounded surviving
hypothesis ledger, and isolation attestation (sandbox path, independent Git directory,
baseline commit, captured deliverable fingerprint and guarded transfer outcome).
Do not change tasks, ACs, Thread state, implementation or sibling audit. No unperformed
review may claim ready or no findings. A failed evidence floor means incomplete.

## Reviewer report

### Verdict and scope

- **Verdict:** Ready (no findings). The single-governing-marker planning pointer alias resolution, ancestor identity lookup, recovery guidance, and CLI-contract Thread closeout satisfy all safety and correctness invariants. An initialized or migrated pointer addressing a configured data subdirectory resolves to the physical data root while correctly acquiring the governing marker's durable identity and configuration preferences. All hostile hypotheses were settled as falsified with concrete operator and programmatic probes, and two compiler-valid mutations were killed by their target tests.
- **Scope reviewed:**
  - Production logic: `internal/config/config.go` (`Discover`, `discoverDataRoot`, `configuredRoot`, `resolveRootIdentity`, `resolvePlanningRepo`, `resolvePlanningTarget`, `planningTarget`, `planningRepoDirectory`, `verifyPlanningRepoID`, `planningIdentityRecoveryCommand`, `configForRoot`, `governingConfig`, `InitPointer`) and `internal/config/migrate.go` (`migrateUnlocked`, `migrationConfig`).
  - Consumer integration: `internal/configstore/fs.go` (`LoadConfiguration`, `DiagnoseConfiguration`), `internal/appwiring/local_sources.go`, `internal/store/threadapply.go` (`currentPlanningIdentity`), and `internal/workspacestore/fs.go`.
  - Machine contract: Revision remains 1.85 under ADR-0008 (restoring documented single-marker alias semantics without additive schema changes).
- **Review limits:** Confined to branch `fix/planning-pointer-alias-identity` at sandbox baseline commit `bdfb6fb5bfbbae114a262d5dfcb706626beeed11`. Overlapping-marker authority (when both `<repo>` and `<repo>/planning` carry separate markers) is an acknowledged distinct issue tracked separately under follow-up task [6gjgccnzxj0a](../tasks/6gjgccnzxj0a-decide-and-enforce-identity-authority-for-overlapping-planning-markers.md).

---

### Verified consumer inventory

Every production component, helper, and consumer was verified against exact sandbox paths and declarations:

| Component / Function | File / Line | Verified Role |
| :--- | :--- | :--- |
| `config.Discover` | [`internal/config/config.go#L107-L163`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L107-L163) | Climbs filesystem to find governing `.tskflwctl.toml`, `tasks/`, or `planning/tasks/`, returning physical `Root`, `Dir`, `ID`, and preferences. |
| `config.discoverDataRoot` | [`internal/config/config.go#L167-L177`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L167-L177) | Discovers a bare-looking data directory, delegating to `governingConfig` to attach ancestor marker identity/preferences. |
| `config.configuredRoot` | [`internal/config/config.go#L184-L203`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L184-L203) | Enforces physical containment of `taskflow_root` within `dir`, preventing escaping symlinks and requiring an existing `tasks/` dir. |
| `config.resolveRootIdentity` | [`internal/config/config.go#L213-L229`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L213-L229) | Unifies root and identity resolution; disallows conflicting `taskflow_root` + `planning_repo`, delegates to `resolvePlanningTarget`. |
| `config.resolvePlanningRepo` | [`internal/config/config.go#L243-L246`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L243-L246) | Public helper delegating to `resolvePlanningTarget` and returning `target.root`. |
| `config.planningTarget` | [`internal/config/config.go#L251-L254`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L251-L254) | Single resolution record (`root`, `configDir`, `recoveryDir`, `tree`, `config`) shared across discovery, init, and migration. |
| `config.resolvePlanningTarget` | [`internal/config/config.go#L256-L308`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L256-L308) | Resolves pointer targets: handles direct markers, disallows pointer chains, finds ancestor governing config, verifies expected ID. |
| `config.planningRepoDirectory` | [`internal/config/config.go#L313-L319`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L313-L319) | Canonicalizes pointer target paths relative to canonical checkout anchor for git worktrees. |
| `config.verifyPlanningRepoID` | [`internal/config/config.go#L329-L344`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L329-L344) | Enforces opt-in identity check; generates actionable recovery guidance pointing at governing config. |
| `config.planningIdentityRecoveryCommand`| [`internal/config/config.go#L348-L357`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L348-L357)| Emits POSIX-quoted `tskflwctl -C <root> config migrate` or `tskflwctl init --path <root> --taskflow-root <tree> --no-register`. |
| `config.configForRoot` | [`internal/config/config.go#L366-L369`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L366-L369) | Convenience wrapper over `governingConfig` used by link consumers and checks. |
| `config.governingConfig` | [`internal/config/config.go#L374-L405`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L374-L405) | Climbs from physical root to find exact governing marker; stops at `.git` boundary, ignores pointers, requires exact physical match. |
| `config.InitPointer` | [`internal/config/config.go#L858-L898`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L858-L898) | Initializes pointer config with captured target ID; validates target up front; verifies ID match on re-init. |
| `config.migrateUnlocked` | [`internal/config/migrate.go#L85-L147`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/migrate.go#L85-L147) | Backfills pointer target ID using `resolvePlanningTarget`; emits actionable recovery guidance when target has no ID. |
| `config.migrationConfig` | [`internal/config/migrate.go#L149-L168`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/migrate.go#L149-L168) | Locates governing config for migration via `Discover`. |
| `configstore.FS.LoadConfiguration` | [`internal/configstore/fs.go#L36-L70`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/configstore/fs.go#L36-L70) | Consumes `Config.Dir`, `Config.Root`, `Config.ID`, `Theme`, `Pager` to build `core.RepositoryConfiguration`. |
| `store.FS.currentPlanningIdentity` | [`internal/store/threadapply.go#L206-L221`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/store/threadapply.go#L206-L221) | Re-reads planning repository identity at apply time, guarding against root redirection under lock. |

---

### Mandatory evidence floor ledger

#### Antigravity Operator Probe: Data Alias Lifecycle & Recovery

Tested with fresh binary `/tmp/tskflwctl` in isolated scratch space (`/tmp/audit_probes_6gjgcy1hyt10`) with ambient `cwd` and isolated `TSKFLW_CONFIG_HOME`:

1. **Setup & Non-Default Topology:**
   - Planning repository created at `planning_repo` with custom layout: `taskflow_root = "custom/planning"`, `id = "custom-stable-id"`, comment `# Custom governing marker comment`, and unknown key `unknown_key = "custom_value_preserve"`.
   - Data directory populated: `planning_repo/custom/planning/tasks/6gje00000001-task1.md` and `epics/01-custom-epic.md`.
   - Implementation repository `impl_repo` initialized with pointer pointing directly at the **data alias** (`planning_repo = "<path>/planning_repo/custom/planning"`), including comment `# Pointer custom comment` and unknown key `impl_unknown = "preserve_impl"`.
2. **Pointer Migration:**
   - Executed: `tskflwctl -C impl_repo config migrate --json` (exit `0`).
   - Inspected `impl_repo/.tskflwctl.toml`: captured `planning_repo_id = "custom-stable-id"`. Comments and unknown key `impl_unknown = "preserve_impl"` preserved byte-for-byte.
3. **Thread Compose via Data Alias:**
   - Executed: `tskflwctl -C impl_repo thread compose --from manifest.yml --out plan.yml --json` (exit `0`).
   - Inspected `plan.yml`: successfully bound to `planning_repo_id: "custom-stable-id"` through the data alias.
4. **Thread Apply via Governing Marker:**
   - Executed: `tskflwctl -C planning_repo thread apply plan.yml --json` (exit `0`).
   - Verified thread document creation: `6gjgjftk64jx-alias-thread.md` written directly into `custom/planning/threads/`.
   - Proved compose via data alias and apply via governing config operate against the exact same durable repository identity.
5. **Direct Discovery via Data Alias:**
   - Executed: `tskflwctl -C planning_repo/custom/planning thread list --json` (exit `0`). Direct discovery correctly inherited governing marker identity and settings without creating a separate markerless space.
6. **Diagnostic & Recovery Execution:**
   - Removed `id = ...` from `planning_repo/.tskflwctl.toml`.
   - Executed: `tskflwctl -C impl_repo thread list --json` (exit `14`, `ErrConflict`).
   - Stderr reported target carries no ID while pointer expects `"custom-stable-id"`, recommending:
     `tskflwctl -C '<path>/planning_repo' config migrate`
     Notice: Advice correctly targeted the governing marker root, NOT the inner data directory.
   - Executed the printed recovery command from ambient directory: exit `0`, updated `planning_repo/.tskflwctl.toml` with newly minted ID (`6gjgjfv3x67q`), preserved comments and `unknown_key = "custom_value_preserve"`.
   - Verified NO inner marker `.tskflwctl.toml` was created inside `custom/planning`.
   - Verified NO space registration occurred in `$TSKFLW_CONFIG_HOME/spaces.toml`.
   - Verified `impl_repo/.tskflwctl.toml` was NOT modified and did not silently rebind.
   - Executed: `tskflwctl -C impl_repo thread list --json` (exit `14`, `ErrConflict`, `refusing to use a different planning repo`), proving fail-closed protection against unreviewed identity changes.

#### Compiler-Valid Mutation Ledger

| Invariant Class | Exact Mutation Location | Mutation Content | Intended Test | Observed Test Failure | Restoration Result |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Primary Invariant (Antigravity):** Opted-in verification for data aliases must receive the governing marker's ID, not an empty ID. | [`internal/config/config.go#L304`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L304) | Passed `""` to `verifyPlanningRepoID` when `target.configDir != "" && target.configDir != root`, simulating pre-fix behavior. | `TestPlanningTarget_BindingRoundTrip` | **FAIL:** All data alias and symlink subtests (`planning/data/init`, `planning/data/migrate`, `docs/work/data/init`, `docs/work/symlink/migrate`) failed on next discovery with: `bound pointer discovery = <nil>, conflict: planning_repo ... resolved to ..., which carries no id — this pointer expects "stable-id"`. | Restored baseline; test PASS (0.07s). |
| **Codex Cross-Check:** Ancestor marker identity requires an exact, containment-checked `taskflow_root` match. | [`internal/config/config.go#L388`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/internal/config/config.go#L388) | Bypassed exact-root authority check in `governingConfig` by replacing `if evalOr(...) != root` with `if false && ...`. | `TestPlanningTarget_AliasCannotBorrowUnrelatedIdentity/different-root` | **FAIL:** `planning_target_test.go:125: bound alias error = <nil>, want conflict`. Unrelated parent marker was incorrectly adopted as authority. | Restored baseline; test PASS (0.00s). |

---

### Surviving-hypothesis ledger

All required hostile angles were settled as **falsified**:

1. **"Physical versus lexical ancestry can allow an ancestor marker to borrow an unrelated identity, or malformed TOML is silently swallowed": FALSIFIED.**
   - In `governingConfig`, both the target `root` and the ancestor's resolved `d/rel` are evaluated via `evalOr`, ensuring symlinks resolve to their physical targets. Containment is physically checked via `configuredRoot`. When an ancestor marker has malformed TOML, `readConfigFile` returns an error which `governingConfig` and `resolvePlanningTarget` bubble up immediately (`ErrValidation`), never silently falling back to a missing-marker path.
2. **"Configless compatibility is broken for bare task directories": FALSIFIED.**
   - `TestPlanningTarget_LegacyBareTreeDoesNotInheritParentIdentity` proves that an un-opted-in legacy pointer to a bare data directory continues to resolve without inventing an ID or failing. An opted-in pointer to a bare directory fails closed (`ErrConflict`) and offers explicit `tskflwctl init --path ... --taskflow-root ... --no-register` guidance.
3. **"Worktree anchoring can be hijacked by a decoy at the relative path": FALSIFIED.**
   - `TestPlanningTarget_WorktreeAliasUsesCanonicalCheckout` proves that relative external pointers within a git worktree anchor from `anchorDir(dir)` (the canonical checkout), correctly resolving the intended external planning repository rather than a decoy placed relative to the worktree directory.
4. **"Snapshot seams allow TOCTOU drift between target root and target ID": FALSIFIED.**
   - `resolvePlanningTarget` captures `root`, `configDir`, `recoveryDir`, `tree`, and `config` in a single pass into `planningTarget`. The same target record is used for verification, pointer initialization, and migration, eliminating multi-read drift.
5. **"Direct alias discovery mutates preferences or writes to wrong targets": FALSIFIED.**
   - Direct discovery via `discoverDataRoot` inherits the governing config's `Dir`, `ID`, `Theme`, and `Pager` without creating a secondary `.tskflwctl.toml` in the data directory. Migration run from the data alias updates the governing marker at `eval(repo)/.tskflwctl.toml`.
6. **"Overlapping markers create undefined authority within the single-marker fix": FALSIFIED (Separate Scope).**
   - The single-marker alias model operates strictly under the invariant that one marker governs the tree. The separate follow-up [6gjgccnzxj0a](../tasks/6gjgccnzxj0a-decide-and-enforce-identity-authority-for-overlapping-planning-markers.md) is cleanly demarcated and does not invalidate single-governing-marker alias resolution.

---

### Mandatory isolation attestation

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.vIjmO2/.git
baseline_commit=bdfb6fb5bfbbae114a262d5dfcb706626beeed11
source_blob=0631a5196658b25c0023b67bf67d07e003c8350d
source_fingerprint=9a7331e95f5cd022ee74fcde0dc24c98bdb06a8b
deliverable=planning/audits/6gjgcy1hyt10-2026-10-10-planning-pointer-alias-identity-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

## Implementation-owner disposition (2026-10-10)

No new in-scope product findings were reported. Accepted the concrete operator
recovery and compiler-valid mutation evidence as corroboration of the single-marker
alias fix, with these limits preserved rather than rewriting the reviewer verdict:

- The broad "TOCTOU drift ... FALSIFIED" statement is not accepted as a general
  guarantee. Capturing root and parsed identity in one resolution removes a separate
  identity lookup; filesystem observation and later consumers are not one global
  transaction. The demonstrated root/ID replay refusals do not prove arbitrary
  concurrent external edits impossible. Codex's scoped seam probes and limits are
  the stronger evidence for this boundary.
- Overlapping-marker authority remains an actual unresolved topology issue in task
  `6gjgccnzxj0a`, not a globally falsified hypothesis. The single-marker scope and
  documented follow-up remain appropriate; no automatic marker cleanup or rebind
  was implemented.
- The received report matched the independent sandbox report, with only its assigned
  audit modified. Its attestation still says `transfer=pending`; a final guarded
  transfer receipt was not retained in this report. This owner comparison confirms
  receipt/restoration, not how the transfer was executed. Do not infer stronger
  procedural evidence from the ready verdict.

Restored the missing report heading and retained the original evidence. Closing this
audit records owner triage of a bounded, qualified review; it is not a claim that
every hostile permutation, external configuration race or future topology is safe.
