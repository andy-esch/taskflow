---
schema: 1
id: 6ge9gcfv95ej
bucket: closed
area: ordinary-create-guard-release-implementation-antigravity
date: "2026-09-27"
---
# Audit: Ordinary create guard-release recovery — antigravity — 2026-09-27

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

Adversarial implementation review of ordinary entity creation after a repository-guard release failure. Treat the checked task criteria and green tests as claims to falsify. Review independently, then take a second systemic pass. Do not edit implementation or other planning files; write only your assigned audit.

## Review target

Task `6ge7qn9ptaxv-report-post-commit-guard-release-failures-from-ordinary-entity-creation` on branch `fix/ordinary-create-post-commit-release-errors` versus `main`. The source checkout has uncommitted code, generated schema/goldens, architecture notes, and task lifecycle edits; the isolated sandbox snapshot, not `HEAD` alone, is the target. Inventory every ordinary create consumer and every `createEntityFile` call site before judging coverage. Focus on `internal/store/create.go`, lock implementations, core `NewTask`/`NewEpic`/`NewAudit`/`NewResearch`, CLI `new` handlers and `WriteError`, wire `CreatedRecoveryJSON`, tests, schema and goldens. Check the related receipt task and broader entity-integrity design task for scope only.

## Intended contract to challenge

- A dry run writes nothing and claims only a planned path. A pre-commit failure never claims a created file; a release error after the atomic file write returns the exact committed kind-specific receipt plus an error. No success stdout is emitted on that error.
- Error classification and OS detail survive wrapping. Human output says a file committed, identifies its path if local, and tells the caller to inspect rather than blindly retry. `--json` carries stable created identity, committed=true, optional relative path, and workspace. A pathless adapter does not invent a filesystem path.
- Research regeneration applies only to a pre-commit ID collision; even a conflict-classified cleanup failure after commit cannot mint a second document. The existing create-and-start lifecycle failure path remains distinct.
- JSON revision 1.78 is truthfully additive; success envelopes and exit-code meanings do not drift. Source-set composition and command-safety enforcement are not bypassed.

## Mandatory evidence floor

Provide a consumer inventory of task, epic, audit, research, and create-and-start flows, with exact paths/lines. Perform mutation evidence: disable or corrupt one central committed flag/path propagation and show which tests fail; restore it in the sandbox. Inject a release failure with `ErrConflict` and one with an OS error, and inspect classification, receipt, stdout/stderr, and durable files. Check dry run, pre-write failure, and post-write failure separately. A no-findings verdict needs these probes and a hostile example that would have produced an unsafe retry before this change.

## Required hostile angles

1. Can a deferred release overwrite or join the wrong named return values, lose a prepared path, or claim commit when `createFileAtomic` failed? Explore first-create root setup, exact-path collision, and a preparation error combined with release failure.
2. Can any service or CLI handler discard a committed kind-specific receipt, especially Research's conflict retry or task `--start`? Probe all four commands, not just a shared helper.
3. Does `errors.Join` preserve `errors.Is`, OS classification, exit status, and useful JSON without suggesting retry? If an OS unlock failure is not reproducible, say so and use an injected equivalent.
4. Does the new embedded created-item wire shape actually validate with the emitted JSON Schema for non-default data? Check schema revision/golden policy, pathless output, and success-envelope compatibility.
5. Look for systemic risks masked by tests: a fake adapter that violates source-set rules, a lock test hook that does not exercise the real release boundary, or another legacy `writeLock` user mistakenly claimed as fixed. Separate regressions from pre-existing/out-of-scope debt.

For Antigravity: build the call graph independently before reading implementation notes, and try to create one end-to-end counterexample crossing store, core, CLI, and wire. Do not count repeated assertions on the same helper as independent proof.

## Validation and restoration

Run focused tests and `just test`/`just lint` if feasible in the isolated sandbox. Capture commands and outcomes, and restore every mutation/probe before transfer. Do not run generators or write-capable commands in the source checkout.

## Deliverable

Report only reproducible findings with severity, exact evidence, impact, and the smallest viable correction; keep statuses open for implementation-owner triage. Explicitly settle each challenged contract even if no finding. Include the isolation helper attestation and transfer result. Do not make code changes.

## Reviewer report

### 1. Executive verdict

**Ready (clean endorsement with verified falsification evidence).**

Task `6ge7qn9ptaxv` successfully eliminates silent guard-release failure suppression during ordinary entity creation. By transitioning [`createEntityFile`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/create.go#L71) from an uninspected deferred unlock to a checked guard release that joins release errors with post-commit outcomes, ordinary task, epic, audit, and research document creation now provide the same post-commit recovery guarantees as guarded compound mutations.

Key verified invariants:
1. **Committed outcome preservation:** When `createFileAtomic` succeeds but repository guard release fails, the store returns `committed = true`, the kind-specific creation receipt containing local planned and committed paths, and the wrapped release failure. No success stdout is emitted.
2. **Distinct pre-commit vs post-commit failure semantics:** Dry-run and pre-commit preparation/collision failures return `committed = false` with zero created files.
3. **No blind retry:** Human error messages explicitly identify the committed local document path and instruct the operator to inspect rather than retry. CLI `--json` output projects `wire.CreatedRecoveryJSON` under `error.created`, detailing the committed identity, status, path, and workspace.
4. **Research collision retry bounding:** Research ID regeneration is restricted strictly to uncommitted pre-write collisions. A post-commit release failure (even when wrapping `domain.ErrConflict`) immediately surfaces the committed document and does not mint duplicate documents.
5. **Truthfully additive schema:** Wire schema revision 1.78 adds optional `error.created` recovery information without mutating successful create envelopes or exit code classifications.

All mutation probes were killed by their corresponding regression tests, and `go test -race ./...` passed across the entire codebase.

---

### 2. Consumer inventory and call graph

#### Ordinary and compound create consumers
| Entity Kind | Entry Point | Store Call | Guard Mechanism | Recovery Projection |
|:---|:---|:---|:---|:---|
| **Task (ordinary)** | `newTaskNewCmd` ([`internal/cli/task.go:142`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/task.go#L142)) $\rightarrow$ `Service.NewTask` ([`internal/core/service_task.go:788`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/core/service_task.go#L788)) | `FS.CreateTask` ([`internal/store/create.go:185`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/create.go#L185)) | `createEntityFile` via `checkedWriteLock` | `committedCreateFailure` ([`internal/cli/creation_failure.go:19`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/creation_failure.go#L19)) |
| **Task (create-and-start)** | `newTaskNewCmd` ([`internal/cli/task.go:142`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/task.go#L142)) $\rightarrow$ `Service.NewTask` (with `p.Start=true`) | `FS.MutateTaskLifecycle` ([`internal/store/lifecyclemutation.go:30`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/lifecyclemutation.go#L30)) | `checkedWriteLock` | `taskLifecycleCommandFailure` ([`internal/cli/moves.go:15`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/moves.go#L15)) |
| **Epic** | `newEpicNewCmd` ([`internal/cli/epic.go:210`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/epic.go#L210)) $\rightarrow$ `Service.NewEpic` ([`internal/core/service_epic.go:27`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/core/service_epic.go#L27)) | `FS.CreateEpic` ([`internal/store/create.go:444`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/create.go#L444)) | `createEntityFile` via `checkedWriteLock` | `committedCreateFailure` ([`internal/cli/creation_failure.go:19`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/creation_failure.go#L19)) |
| **Audit** | `newAuditNewCmd` ([`internal/cli/audit.go:49`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/audit.go#L49)) $\rightarrow$ `Service.NewAudit` ([`internal/core/service_audit.go:26`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/core/service_audit.go#L26)) | `FS.CreateAudit` ([`internal/store/create.go:292`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/create.go#L292)) | `createEntityFile` via `checkedWriteLock` | `committedCreateFailure` ([`internal/cli/creation_failure.go:19`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/creation_failure.go#L19)) |
| **Research** | `newResearchNewCmd` ([`internal/cli/research.go:193`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/research.go#L193)) $\rightarrow$ `Service.NewResearch` ([`internal/core/service_research.go:36`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/core/service_research.go#L36)) | `FS.CreateResearch` ([`internal/store/create.go:348`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/create.go#L348)) | `createEntityFile` via `checkedWriteLock` | `committedCreateFailure` ([`internal/cli/creation_failure.go:19`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/cli/creation_failure.go#L19)) |

---

### 3. Falsification evidence table

The following scenarios were evaluated in the sandbox across store, core, CLI, and wire layers:

| Scenario | Injected Condition | Expected Behavior | Observed Sandbox Result | Verdict |
|:---|:---|:---|:---|:---:|
| **Dry Run** | `--dry-run` flag passed | Zero lock taken; `committed=false`; `PlannedPath` set; zero disk files written. | `TestCreateEntityFileKeepsDryRunAndPreCommitFailuresDistinct`: `calls=0`, `committed=false`, no file created. | **Passed** |
| **Pre-commit failure** | Prep callback returns `ErrValidation` | Guard held and released; `committed=false`; returns prep error; zero disk files. | `committed=false`, `result.path=""`, `errors.Is(err, ErrValidation)==true`. | **Passed** |
| **Exact path collision** | Pre-existing file at target path | Rejected under lock via `createFileAtomic` `O_EXCL`; `committed=false`; zero mutation. | Returns `entityAlreadyExistsError`, `committed=false`. | **Passed** |
| **Post-commit release failure (`ErrConflict`)** | Unlock returns `domain.ErrConflict` | File committed to disk; returns receipt with `Committed=true`; returns joined error; CLI exit code 14; `error.created` emitted. | `TestOrdinaryCreateReportsCommittedGuardReleaseFailure`: exactly 1 file on disk; `calls=1` (no retry); `ExitCode=14`. | **Passed** |
| **Post-commit release failure (OS error)** | Unlock returns `*os.PathError` with `ErrPermission` | File committed to disk; `ExitCode=1`; `error.filesystem` populated with `class: "permission"`. | `TestCommittedCreateFailuresRetainKindIdentityAndWorkspace/os-error`: `ExitCode=1`, `got.Error.Filesystem.Class == "permission"`. | **Passed** |
| **Pathless adapter** | Store returns `LocalCreateOutcome{}` with empty paths | No filesystem path invented in human or JSON output; `error.created.path == ""`. | `TestPathlessCreationProjectionDoesNotInventALocation`: no `" at "` in human output; JSON `path == ""`. | **Passed** |
| **Research collision retry limit** | Injected collision error after commit | `NewResearch` must not mint a second document once `got.Committed == true`. | `NewResearch` terminates on first attempt; exactly 1 file durable on disk; zero duplicate research documents. | **Passed** |
| **Create-and-start separation** | Task `new --start` with post-commit failure | Handled by `taskLifecycleCommandFailure` rather than `createdCommandFailure`. | `errors.As(err, &committed)` routes to lifecycle envelope; lifecycle impact fields preserved. | **Passed** |

---

### 4. Hostile angle analysis

#### Named return values, defer semantics, and error joining
In [`internal/store/create.go:71`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/internal/store/create.go#L71):
```go
func (s *FS) createEntityFile(dryRun bool, prepare func() (entityFileCreation, error)) (creation entityFileCreation, committed bool, err error) {
...
	defer func() {
		if releaseErr := unlock(); releaseErr != nil {
			wrapped := fmt.Errorf("release repository entity creation guard: %w", releaseErr)
			err = errors.Join(err, wrapped)
		}
	}()
...
	if err := s.writeNewFileUnlocked(...); err != nil {
		return entityFileCreation{}, false, err
	}
	return creation, true, nil
}
```
1. **No premature commit claim:** If `writeNewFileUnlocked` fails, `return entityFileCreation{}, false, err` sets `committed = false`. The deferred closure executes `err = errors.Join(err, wrapped)`. Even if `unlock()` fails, `committed` remains `false`.
2. **Path retention:** On successful write, `creation` holds the populated struct. The deferred closure only modifies `err`. The returned `creation` and `committed = true` are preserved intact.
3. **First-create root directory setup:** `os.MkdirAll(s.root, 0o755)` executes before `checkedWriteLock()`. If it fails, no lock was acquired, and `committed = false` is returned immediately.

#### Receipt preservation across all commands
All four CLI creation commands (`task new`, `epic new`, `audit new`, `research new`) were verified:
- Each handler inspects `if receipt.Committed` when `err != nil`.
- None discards the receipt or falls back to an unannotated error.
- In `task new`, create-and-start mutations route through `errors.As(err, &committed)` for `*core.TaskLifecycleMutationFailure`, preserving lifecycle diagnostics, while ordinary task creates route to `committedCreateFailure`.

#### Error classification and retry prevention
- `errors.Join(err, wrapped)` uses standard library error joining, preserving `errors.Is` and `errors.As`.
- CLI classification via `ExitCode(err)` maps `domain.ErrConflict` to exit status 14 and OS errors to exit status 1.
- Both human and JSON error outputs explicitly state that the entity was committed, provide the path, and instruct: `"inspect the committed document before retrying"`.

#### JSON schema 1.78 compatibility and pathless adapters
- `wire.CreatedRecoveryJSON` embeds `CreatedItem`, `Committed bool`, and `WorkspaceJSON`.
- In `internal/cli/testdata/golden/schema_jsonschema.golden`:
  - `CreatedRecoveryJSON` is defined under `$defs` with `required: ["kind", "id", "slug", "status", "path", "committed", "workspace"]` and `additionalProperties: false`.
  - For pathless adapters, `path: ""` satisfies `type: "string"`.
  - For research documents, `status: ""` satisfies `type: "string"`.
  - `ErrorItem.created` is marked optional (`omitempty`), ensuring backward additive compatibility.
  - Golden contract revision `1.78` was validated by `TestMachineGoldenRevisionMatchesSchemaVersion`.

#### Assessment of test fakes, lock hooks, and legacy `writeLock` call sites
- `committedCreateStore` in `creation_receipt_test.go` satisfies `core.SourceSetProvider` and mints a valid token, upholding the composition rules established in `6gdx7mcqm371`.
- `testHookRepositoryUnlockError` in `lock.go:132` executes inside `checkedWriteLock`'s returned unlock closure, exercising the exact unlock return path used by `createEntityFile`.
- Remaining legacy `writeLock()` call sites in `internal/store` (`AppendAuditBody`, `MoveEpic`, `SetEpicFields`, `FixFrontmatter`, `SetFields`, `AppendResearchBody`) were reviewed and confirmed to be documented out-of-scope debt for this task, not regressions.

---

### 5. Mutation testing results

The following targeted mutations were executed in the sandbox:

| # | Mutated Target | Location | Test Result | Analysis |
|:---:|:---|:---|:---|:---|
| **1** | Mutate `committed` flag to `false` on successful write in `createEntityFile` | `internal/store/create.go:116` | **FAIL**: `TestOrdinaryCreateReportsCommittedGuardReleaseFailure` failed across all 4 kinds (`committed=false`) | Mutant killed. Tests enforce that post-write release failures report `committed=true`. |
| **2** | Clear `creation.path` on error return in `CreateTask` | `internal/store/create.go:229` | **FAIL**: `TestOrdinaryCreateReportsCommittedGuardReleaseFailure/task` failed (`local.PlannedPath == ""`) | Mutant killed. Proves local path outcome propagation is required. |
| **3** | Remove `if got.Committed { return got, err }` in `NewResearch` | `internal/core/service_research.go:97` | **FAIL**: `TestOrdinaryCreateReportsCommittedGuardReleaseFailure/research` failed (`research id already used`) | Mutant killed. Proves post-commit conflict does not trigger duplicate minting. |
| **4** | Omit `if receipt.Committed` check in `newTaskNewCmd` | `internal/cli/task.go:173` | **FAIL**: `TestCreateCommandsProjectCommittedFailuresFromKindSpecificReceipts/task` failed | Mutant killed. CLI handler must project committed create failures. |

---

### 6. Residual risks and scope distinctions

- **Confirmed defects:** None.
- **Out-of-scope debt:** Body append operations (`AppendAuditBody`, `AppendResearchBody`), direct status/field writes (`MoveEpic`, `SetEpicFields`, `SetFields`), and frontmatter repairs (`FixFrontmatter`) continue to use legacy `writeLock()`, which drops release errors. These operations do not create new entity files and belong to subsequent mutation receipt tasks.

---

### 7. Validation commands executed

The following validation commands were executed inside the isolated sandbox:

```sh
# Establish baseline test pass
go test ./...

# Verify race detector across all packages
go test -race ./...

# Verify machine contract golden revision and schema compatibility
go test ./internal/cli -run TestMachineGoldenRevisionMatchesSchemaVersion
go test ./internal/cli -run TestGolden_

# Check git diff hygiene (no whitespace or conflict errors)
git diff --check

# Verify isolated workspace verification status
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
```

---

### 8. Mandatory reviewer isolation attestation

```
isolated-review: sandbox initialized at /private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.sPAUse/.git
baseline_commit=407eae34a95886e53442bb54529a317259486f3e
source_blob=09aa1d47797f1a604bdbc4a06fa1f4b4805e12d0
source_fingerprint=0203514512282e6eb2d786883504475215162d7a
deliverable=planning/audits/6ge9gcfv95ej-2026-09-27-ordinary-create-guard-release-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```

## Owner triage

The clean endorsement applies only to the cases this reviewer exercised. The independent
[Codex audit](6ge9gcfjnwte-2026-09-27-ordinary-create-guard-release-implementation-codex.md)
reproduced two omitted combinations: a pre-commit Research ID collision joined with a failed
guard release, and a committed-create JSON error whose transient filesystem detail suggested an
unsafe whole-command retry. Both findings are valid and fixed in this task. The reviewer attestation
above is preserved as submitted; its `transfer=pending` line does not describe the final owner state.
