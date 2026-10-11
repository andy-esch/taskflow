---
schema: 1
id: 6gjgcy198g1m
bucket: closed
area: planning-pointer-alias-identity-implementation-codex
date: "2026-10-10"
---
# Audit: Planning pointer alias identity and CLI-contract closeout — codex — 2026-10-10

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

### Findings

No new in-scope product findings were demonstrated. The audit remains open. The known overlapping-marker issue remains unresolved in task 6gjgccnzxj0a; this review does not dispose of it.

### Verdict and limits

**Technical verdict: ready within the single-governing-marker scope.** I completed the checklist pass and an explicitly separate systemic pass, executed the required compiler-valid mutation, replayed actual recovery advice, and restored all probe code. Single-marker aliases retained governing identity without copying a marker, borrowing unrelated preferences, weakening saved pointer expectations, or selecting a worktree decoy.

This is a bounded review of the captured dirty implementation on fix/planning-pointer-alias-identity against fd0d97adf3ca3a9fb258422738c25288a82d6bfc. The owner’s full race suite, lint, docs-check and planning gates are inherited evidence, not independently rerun broad gates. My independent gates are below. I did not inspect the sibling audit. Arbitrary external edits during multiple filesystem reads are not a globally serialized transaction.

Procedural limit: before copying, source access additionally included pwd, rg --files to locate policy/deliverable/helper paths, and reading the isolation helper. That exceeds the brief’s literal “brief and initial copy only” restriction. No implementation inspection, tests, generation, mutation, or editing occurred in the source; all substantive review work used the independent clone. This attestation does not erase that narrow pre-copy deviation.

### Verified implementation and consumer inventory

| Boundary | Verified implementation and reachable consumer |
| --- | --- |
| Discovery and identity observation | internal/config/config.go:112 (Discover) uses resolveRootIdentity at line 213. Direct data fallback uses discoverDataRoot at line 167 and retains governing Dir, ID, TrackedRepos, Pager and Theme. |
| Shared pointer resolution | planningTarget at internal/config/config.go:251 carries root, config/recovery directories, tree and parsed config. resolvePlanningTarget at line 256 is used by discovery, InitPointer (line 868), and pointer migration (internal/config/migrate.go:117). Its identity check is at line 304; existing-pointer init checks the saved expectation at line 890. |
| Authority | governingConfig at internal/config/config.go:375 physically resolves the root, stops at the nearest marker, refuses pointer ancestry as authority, checks exact root equality at line 390, then containment/tasks at line 393. Git directory/file boundaries stop climbing at line 399. configuredRoot at line 184 checks physical containment at line 193. |
| Recovery and migration writes | verifyPlanningRepoID at internal/config/config.go:329 classifies missing/mismatched opted-in IDs as conflict. planningIdentityRecoveryCommand at line 348 emits shell quoting and governing-marker migration. migrationConfig at internal/config/migrate.go:144 uses discovered Dir; backfillID at internal/config/config.go:793 and insertTopLevelKey at line 815 insert the key surgically. internal/config/lock_unix.go:14 locks the config directory itself with flock; there is no invented lock-file path. |
| Canonical anchoring and link consumers | planningRepoDirectory at internal/config/config.go:313 uses anchorDir (internal/config/worktree.go:51) for relative external addresses. In-tree roots use their local config directory. configForRoot at line 366 is an error-discarding wrapper used by LinkBack (line 932) and checkBackLink (line 1064); identity-binding paths use the error-returning authority helper. CheckLinks (line 1046) and checkTrackedRepo (line 1086) are shipped consumers. |
| Startup and style | internal/appwiring/local_sources.go:25 selects real discovery; readRepository at line 30 and repositorySettings at line 49 map root/ID/config/preferences. openPlanning at internal/appwiring/wiring.go:57 uses one initial observation for metadata, service and layout. internal/cli/root.go:457 publishes those capabilities and resolves the theme at line 475; pager/theme data remain repository scoped. |
| Configuration description/preferences | internal/configstore/fs.go:36 loads discovery, raw description and pending migrations; config.Describe is at internal/config/config.go:729. Diagnosis at internal/configstore/fs.go:121 calls link checks. SetPreference at line 145 delegates repository edits to config.SetPresentation (internal/config/preference.go:25), which selects the discovered governing config path at line 33. |
| Workspace and guarded Thread replay | internal/workspacestore/fs.go:26 opens the same discovered corpus. NewPlanningStore at internal/workspacestore/planning.go:17 anchors re-reading at governing Dir or the legacy root. internal/store/threadapply.go:206 re-reads identity and rejects a different physical guarded root at line 214; internal/core/thread_apply.go:348 rejects a plan/current ID mismatch. No filesystem logic was added to core/ports. |
| Fixtures captured with the implementation | internal/config/planning_target_test.go:18 covers binding round trips across layouts/addresses/modes; line 77 covers hostile authority, line 135 canonical-worktree aliases, line 171 legacy bare trees, line 194 direct-alias migration. Changed CLI fixtures are internal/cli/thread_recovery_test.go:196 and line 230. Adjacent discovery/link/worktree fixtures are in discover_test.go, links_test.go, repoid_test.go and worktree_test.go; they ran in the focused config gate. These are actual files in the dirty snapshot, not claims of committed/merged delivery. |
| Documentation and wire scope | README lines 323–327 and docs/ARCHITECTURE.md:645 qualify parity by one governing marker. The Thread closeout explicitly says independent reviews are pending and is not merge/release evidence. The diff command below returned no changes to wire/core/render or these downstream adapters. Revision remains 1.85 at internal/wire/wire.go:365. |

The revision rationale survives this review: restored alias root/identity semantics were already described by the resolver’s same-tree contract (internal/config/config.go:236–242). No key, selector, vocabulary or optional wire branch was added. ADR-0008 covers value semantics as well as shapes, so this conclusion depends on restoration of documented meaning, not merely a shape-only argument. I found no demonstrated incompatible documented meaning requiring another revision. Optional-wire semantic mutation/validation is **not applicable** here; empty schema tests would not establish identity behavior.

### Independent probe and mutation ledger

Evidence root: /private/tmp/taskflow-alias-audit.MfpSIE. Build and tests ran from the sandbox with separate GOCACHE, GOTMPDIR and GOMODCACHE directories there and GOPROXY=off. Dependency-cache bytes were copied from the preceding independent review’s scratch cache; no shared module/cache writes, installed binary or network were used.

Exact independent commands:

~~~sh
GOCACHE=/private/tmp/taskflow-alias-audit.MfpSIE/go-cache \
GOTMPDIR=/private/tmp/taskflow-alias-audit.MfpSIE/go-tmp \
GOMODCACHE=/private/tmp/taskflow-alias-audit.MfpSIE/go-mod-cache GOPROXY=off \
go test ./internal/config ./internal/cli ./internal/appwiring -count=1

# Same environment:
go build -buildvcs=false -o /private/tmp/taskflow-alias-audit.MfpSIE/tskflwctl ./cmd/tskflwctl
python3 /private/tmp/taskflow-alias-audit.MfpSIE/cli_probes.py
python3 /private/tmp/taskflow-alias-audit.MfpSIE/worktree_probes.py
python3 /private/tmp/taskflow-alias-audit.MfpSIE/topology_probes.py

git diff HEAD^ --name-only -- internal/wire internal/core internal/cli/render internal/appwiring internal/configstore internal/workspacestore
~~~

| Probe | Observation and receipt |
| --- | --- |
| Custom alias init plus legacy migration | Actual CLI target init used --taskflow-root notes/deep/work --no-register; pointer init used its data address with --no-register --no-link-back. Legacy pointer preview and migration preserved comments/unknown TOML. Saved expectations, config/direct-data/symlink/pointer views and the saved plan all bound ID 6gjgh7zae2y9. Direct aliases retained governing config directory and catppuccin/explicit false pager/repo-pager; pointer presentation remained local to the pointer. executed/custom-views.json, marker-before.toml, cli-ledger.json and raw per-command streams preserve exact paths/results. |
| Real cross-address Thread replay | Compose through the bound pointer, apply through governing config, then retry through the direct symlink all exited 0. Original task bytes stayed unchanged; the expected Thread document was created. The script records plan bytes and binding in executed/custom-recovery-summary.json. |
| Cross-lens missing-ID recovery | Removed only governing id assignment after binding. Actual JSON conflict advice contained a quoted tskflwctl -C '<governing config directory>' config migrate; /bin/sh parsed it, then the fresh binary executed that exact argv from unrelated cwd. Spaces, double quotes and an apostrophe were present in the path. Recovery minted 6gjgh80g2btn; all old config content survived, with only the inserted ID assignment and its separating blank line added. Document/topology snapshots, both pointer bytes and isolated home remained unchanged; no inner marker appeared. Old pointers, old plans and repeat pointer init returned 14; a fresh direct-alias plan applied successfully. executed/recovery-before.toml, recovery-before-documents.json, custom-recovery-summary.json and streams are receipts. |
| Recovery with an existing registry | A separate actual CLI registration created an unrelated ambient entry in isolated spaces.toml. Replayed governing-marker remedy again; registry hash c24b9ab3ca4389abe40bf062dc7b56feea40f9b829dfb7cc899215fd908683d0, user-config hash, documents and pointer bytes all remained unchanged. Pointer still conflicted. topology verified/observed.json and ledger.json record this distinct later identity recovery. |
| Hostile authority matrix | Nine real-command cases tested unrelated exact-root mismatch, intermediate nearest marker, Git directory/file boundaries, ancestor self-pointer, malformed governing TOML, replaced/missing ID and an escaping symlink. Discovery, migration and repeat-init refused without config/pointer/home writes: conflict 14, except malformed TOML validation 11. For unrelated/nearest/Git/pointer/escaping cases, legacy pointer initialization and direct discovery succeeded with neither inherited ID nor parent preferences. executed/hostile-summary.json, hostile-*-before.json, cli-ledger.json and streams contain exact fixture bytes/observations. |
| Configless and recursion bounds | Bare tasks/ and conventional planning/tasks/ direct and pointer discovery succeeded without ID or marker creation; pointer migration refused with 11 and no writes. Literal self-pointer target was refused with 11; ancestor self-pointers did not recurse. Every CLI subprocess had a 20-second timeout. |
| Canonical-worktree/decoy and links | Pointer main checkout and linked worktree both selected custom data ID 6gjghfbmtb37, ignoring valid naive-relative decoy ID 6gjghfbqkfv3. Worktree repeat-init/migration were safe no-ops; compose through worktree then apply through governing config succeeded and left decoy/home bytes unchanged. Actual linkback wrote ../impl in the governing marker; config doctor was clean from main, worktree and direct-data addresses. Separate in-tree planning worktree retained its own local data root. worktree probes/ledger.json includes scratch Git setup and CLI argv; observed.json and streams record roots/IDs. Fixture Git repositories are separate scratch repositories, not commits/worktrees of the review/source repository. |
| Required compiler-valid mutation | In governingConfig, replaced only if evalOr(filepath.Join(d, filepath.FromSlash(rel))) != root { with if false { // reviewer mutation: exact-root authority disabled. Ran go test ./internal/config -run '^TestPlanningTarget_AliasCannotBorrowUnrelatedIdentity$/^different-root$' -count=1 -v. **Killed**: compilation succeeded; exit 1 at planning_target_test.go:125, `bound alias error = <nil>, want conflict`, from the first Discover attempt. Mutant successfully lent unrelated parent ID, the intended wrong-identity reason. No coordinated caller mutation needed. Restored exact bytes; same test passed exit 0. exact-root-mutation.log, exact-root-restored.log, mutation-receipts.json. |
| Separate systemic consumer/snapshot probe | Temporary TestReviewerDirectAliasConsumersAndSnapshotSeam used actual local bindings/configstore readers and edits, without repository fixture helpers. Startup/full configuration selected the same governing marker; typed preference write preserved opaque TOML and created no inner marker. After opening via direct data alias and composing, changed governing root to a decoy while keeping the same ID: dry-run and real apply rejected changed guarded root without writes to either corpus. Preference write through the now-independent old alias refused validation instead of changing unrelated parent settings. Restored root with replacement ID: both replay modes conflicted without document writes. Test passed; source archived as reviewer_alias_probe_test.go beside systemic-alias-consumers.log, then removed from sandbox. |
| Restored gates | Initial focused gate passed (focused.log). After all production/temporary test probes were restored/removed, go test ./internal/config ./internal/cli ./internal/appwiring ./internal/configstore ./internal/workspacestore -count=1 passed (restored-focused.log). Mandatory mutation’s restored gate also passed independently. |

The CLI harness was corrected during construction for nested/omitted JSON fields, decoding the error stream before shell parsing, and the migration writer’s intentional ID separator; only the final executed/ run is the complete primary ledger. An initial registration fixture with a space-containing basename produced the documented registration warning; the completed registry probe used ambient-registration. These were reviewer harness corrections, not product findings.

### Explicit second pass and surviving hypotheses

The second pass challenged the shared resolution helper and consequential changes from direct aliases returning governing metadata:

- **Wrong preference/write target:** settled with actual startup/configuration readers, configstore preference mutation, governing-marker byte checks and an independent alias after root redirection. Returning governing marker enables the intended preference write; unrelated nearest markers cannot receive that alias’s write.
- **Changed state between projection/opening and action:** settled for reachable root/ID replacement with same-ID decoy and replacement-ID replay in both apply modes. Production re-reader remains anchored to governing marker; physical-root and plan-ID guards rejected those changes without corpus writes. This does not establish global atomicity against arbitrary editor races after a read; that broader race is not newly claimed solved.
- **Test helpers concealing authority bugs:** required mutant failed on successful wrong-identity discovery, not compilation or incidental caller guard. Out-of-process fixtures separated cwd, explicit target, pointer address, physical data directory and governing config. Systemic temporary test used direct fixture bytes and production consumers.
- **Boolean helper hiding malformed binding TOML:** identity-bearing discovery/init/migration use governingConfig’s error path. Malformed-marker CLI attempts were loud validation errors. configForRoot remains diagnostic/linkback convenience; LinkBack first performs target validation. Concurrent edits between later reads are broader existing uncertainty, not demonstrated new identity corruption.
- **Legacy malformed ancestry:** well-formed unrelated parents and bare/conventional roots were verified. A malformed candidate ancestor is deliberately loud under the authority walk; I did not establish a documented valid-input guarantee that malformed nearby topology must be ignored. Exhaustive invalid-TOML/permission/broken-symlink permutations were not performed.
- **Overlapping-marker boundary and closeout:** independently initialized both outer repo (taskflow_root=planning) and its data directory (taskflow_root=.). Two pointers reported the same physical root with IDs 6gjgj10xk7cj and 6gjgj1107c3y; plan composed through outer identity was refused through inner identity with exit 14 and no document write. This reproduces known topology and its actual replay consequence. It requires two markers, and no comparable single-marker failure survived. README/architecture and Thread closeout qualify the guarantee and name pending reviews/follow-up; deferral is defensible within this brief’s scope. Topology issue is not fixed, merged, released, or covered by a ready verdict for overlapping trees.
- **Residual bounds:** no global config transaction, automatic overlapping-marker policy, pointer-chain support, exhaustive OS matrix, or independently rerun full race/lint/documentation gates is claimed.

### Isolation attestation and guarded delivery

~~~text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.TwqBq8
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.TwqBq8/.git
baseline_commit=bdfb6fb5bfbbae114a262d5dfcb706626beeed11
source_blob=764ea424feff3aa32bdc9c3181954e1f8c53622f
source_fingerprint=9a7331e95f5cd022ee74fcde0dc24c98bdb06a8b
deliverable=planning/audits/6gjgcy198g1m-2026-10-10-planning-pointer-alias-identity-implementation-codex.md
~~~

General helper created independent --no-hardlinks clone and captured dirty overlay in its sole sandbox baseline commit. All probe edits were restored to that baseline. No subsequent review-repository commit, staging or branch switch was made. verify-clean.log attests clean restoration before report editing; final verification, audit lint and inspected audit diff are retained as verify-final.log, audit-lint.json and audit-final.diff.

Delivery postcondition: **transfer=succeeded**, established only by scripts/isolated-review-workspace.sh transfer --sandbox <sandbox_path> and its authoritative /private/tmp/taskflow-alias-audit.MfpSIE/transfer.log. Final pre-transfer verification necessarily reports transfer=pending; on refusal this draft is undelivered and conflict must be reported without bypassing guard. Helper alone transfers this assigned file. Retain sandbox and scratch evidence until owner confirms receipt.

## Implementation-owner disposition (2026-10-10)

Accepted the bounded ready verdict: no new in-scope product findings to fix or
transfer. Checked the retained mutation/restoration receipts, systemic probe and
guarded-transfer log; the received report matched the independent sandbox report,
whose only working-tree change was its assigned audit. The disclosed pre-copy
metadata/helper reads do not demonstrate shared implementation mutation and remain
recorded as a procedural qualification, not erased by this disposition.

Preserved the replay seam as
`TestPlanningDataAliasRevalidatesGoverningMarkerBeforeThreadApply`: ordinary and
workspace stores opened via a data alias reject both same-ID root redirection and
replacement IDs, in preview and apply, without writing either corpus. No production
fix was necessary. The separately reproduced overlapping-marker replay refusal
remains in task `6gjgccnzxj0a`; closing this audit does not resolve that topology or
claim a global transaction against arbitrary configuration edits.
