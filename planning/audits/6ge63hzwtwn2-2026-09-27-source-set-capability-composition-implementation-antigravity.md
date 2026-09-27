---
schema: 1
id: 6ge63hzwtwn2
bucket: closed
area: source-set-capability-composition-implementation-antigravity
date: "2026-09-27"
---
# Audit: Source-set capability composition implementation — antigravity — 2026-09-27

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

Adversarial implementation review of the source-set composition boundary. Treat the task's
acceptance criteria and the implementation summary as claims to falsify, not evidence of safety.
Both reviewers should make an independent first pass and a second pass for a common-mode defect.
Do not edit code or planning outside your assigned audit; report reproducible findings only.

## Review target

Review task `6gdx7mcqm371-bind-split-planning-capabilities-to-one-source-set` on
`feat/source-set-capability-composition` against its main-branch base. The handoff includes
uncommitted implementation and planning changes: the sandbox baseline, not `HEAD` alone, is the
review target. Inventory all production construction sites and independently injectable planning
capabilities before forming a verdict. In particular inspect `internal/core/source_set.go`,
`service.go`, `service_task.go`, `workspace.go`, `store.go`, `internal/store/fsstore.go`,
`internal/workspacestore`, `internal/cli/root.go`, TUI workspace consumers, all new source-set
tests and the test-fake changes. Check the design task `6gcwcf7rgxef`, this task, and the queued
source/path split `6gcwcf88z57p` only for contract and sequencing, not as shipped behavior.

## Intended contract to challenge

- One opaque, comparable, nonzero `SourceSetID` witnesses an adapter-owned planning corpus.
  It is process-local composition evidence, not authorization or a durable planning-space key;
  paths, record IDs, source locations, and revisions must not substitute for it. A composition
  root may explicitly share a token between adapter objects it knows address one corpus.
- `NewService` rejects every selected aggregate or independently injected planning-data reader,
  path resolver, or mutator with missing, zero, unstable, or differing witness before returning a
  usable service. `WorkspaceService.Open` and CLI resolution propagate that typed error.
- Implicit Thread paths detach when Thread reads are explicitly replaced. An explicit compatible
  path source survives either option order; a mismatched one fails. Absent optional ports, typed
  nils, and pathless/read-only services remain valid. Unsupported operations still report their
  existing capability errors.
- Filesystem capabilities on one `FS` instance share a token. Two independent instances do not
  gain an identity match merely because their roots, record IDs, locations, or revisions match.
  All currently selected task, epic, audit, research, and Thread data paths use the shared
  construction rule; future entity-specific path ports must reuse it.

Do not turn the honest-adapter assumption into a claim that tokens cryptographically prove physical
equality. Conversely, an adapter's ability to lie is not by itself a finding. Show an accidental
or plausible miscomposition that this implementation claims to prevent but permits, or a valid
composition that it rejects without a supported escape hatch.

## Mandatory evidence floor

1. Build a consumer inventory from code, not from the task prose: every production `NewService`
   or `MustNewService` call, every `With*` planning-data option, aggregate discovery fallback,
   workspace field, and direct local-path or mutation channel that might bypass the validator.
   Name any deliberately excluded capability and test whether the exclusion is safe.
2. Execute at least one real negative construction probe (mismatched source sets) and one positive
   split-capability probe (honestly shared set); verify the error classification with `errors.Is`
   and that rejection precedes a planning-data read, editor launch, or mutation. Include a
   pathless/read-only probe and at least one multi-workspace case.
3. For each new regression test that purports to protect a distinct invariant, mutate its exact
   guard in the sandbox (for example, omit one capability from validation, use root text as an
   ID, keep implicit Thread paths, or stop checking a missing provider). Record whether the
   named test fails; restore code afterward. A broader unrelated failure does not count.
4. Run focused tests and `go test -race ./...` if feasible. Any skipped command needs a reason.
   Inspect `git diff --check` and prove only the assigned audit differs before transfer.
5. Cite exact path/line and a command result for each finding. Separate an observed defect from
   a future-path-port concern. If there are no findings, include the verified inventory, negative
   controls, mutation-kill results, and residual risks rather than a checklist-only endorsement.

## Required hostile angles

- Try all option orders for explicit Thread read/path replacement, including an aggregate Store
  that discovers a path and a separate task graph reader. Test typed-nil interfaces and a store
  that lacks `SourceSetProvider`; test missing, zero, and changing IDs for a guarded mutator too.
- Challenge the `taskStoreGraphSource` fallback, lint/audit defaulting, repair/lifecycle/Thread
  mutation injection, and the `MustNewService` escape hatch. Can any production path obtain a
  service or mutate without the selected source sets agreeing?
- Inspect `WorkspaceSource.Layout`, checkout/root/planning ID, watcher use, and any CLI direct
  `Fixer`/`Linter` wiring. Decide with evidence whether they are safely outside this task's
  planning-data boundary or an unguarded cross-corpus local capability. Do not demand an
  unrelated new feature to call this change defective.
- Challenge the test fakes' shared default token: can it mask a real cross-corpus pairing or make
  a test pass for the wrong reason? Verify that same-ID/different-corpus fixtures would fail if
  the boundary were weakened and that explicit same-corpus pairing still works.
- Probe error semantics and lifecycle: constructor return value on error, CLI exit mapping,
  workspace open failure, and whether merely composing ports performs a read/mutation first.
  Check that no token leaks into domain records, JSON, or generated CLI schemas.

Codex emphasis: trace the architecture end to end and look for a bypass or contract mismatch in
direct adapters, legacy fallback ports, and future port extension. Try a coordinated mutation of
the validator and its immediate caller; compilation alone is not a useful invariant proof.

Antigravity emphasis: act as a falsification engineer. Construct a small table of selected port
combinations with predicted accept/reject outcomes, run several surprising cases, and seek one
counterexample. Reproduce at least one proposed issue with an executable sandbox test or minimal
probe. Before reporting no findings, explain why omitted `Layout`/direct CLI capabilities and the
shared test token are not counterexamples. Prefer a concrete flaw over a broad speculative list.

## Validation and restoration

All inspection, tests, mutation probes, and report edits must happen inside the independent
reviewer sandbox described above. Restore all probe code and generated changes to the sandbox
baseline without touching the shared handoff checkout. Do not commit, push, or edit the other
reviewer's audit. The only transferable delta is your assigned audit.

## Deliverable

Preserve this brief. Under `## Reviewer report`, provide a verdict and evidence table, then findings
using the exact audit heading grammar and leaving each finding `open`. For each finding include
severity, affected contract, reproduction, likely cause, and a bounded repair direction. Include
the isolation attestation and transfer result. If there are no findings, give a reasoned clean
verdict with residual uncertainty and the required negative and mutation evidence.

## Reviewer report

### 1. Executive verdict

**Ready (clean endorsement with verified falsification evidence).**

Task `6gdx7mcqm371` establishes an adapter-neutral, fail-closed source-set composition boundary across all production and injectable planning capabilities. Construction through [`core.NewService`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L228) strictly validates that every non-nil planning capability publishes a non-zero, stable, identical [`core.SourceSetID`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/source_set.go#L15) witness before returning a usable service.

Adversarial exploration across all option permutations, typed-nil interfaces, cross-corpus fixtures, and mutation probes demonstrated that:
1. Rejection is fail-closed, occurs at construction time, and strictly precedes any planning-data read, editor launch, or disk mutation.
2. The error [`core.ErrIncompatibleCapabilities`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/source_set.go#L34) wraps [`domain.ErrValidation`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/domain/errors.go), preserving CLI exit status 11 classification.
3. Pathless and read-only services remain supported without requiring dummy local path or mutation capabilities.
4. Replacing Thread reads cleanly detaches implicit aggregate-store thread path resolution, while explicitly supplied paths survive arbitrary option ordering.
5. Filesystem instances mint unique process-local nonces and cannot be paired across distinct corpora or independent un-bound instances, even when record IDs, locations, revisions, or root paths match.
6. Direct CLI capabilities (`Fixer`, `Linter`), directory watch hints (`Layout`), and test fake default tokens (`testSourceSetID`) were examined under hostile scrutiny and proved safe from cross-corpus contamination.
7. Zero token leakage occurs into domain entities, JSON DTOs, or CLI golden schemas.

All 6 targeted mutation probes were killed by their corresponding regression tests, and `go test -race ./...` passed across the entire repository.

---

### 2. Consumer inventory and capability mapping

The following production construction sites, functional options, and planning capabilities were inventoried from code:

#### Production construction sites
| Call Site | Location | Passed Capabilities | Error Propagation |
|:---|:---|:---|:---|
| `App.resolveFrom` | [`internal/cli/root.go:453`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/cli/root.go#L453) | `fs := store.NewFS(...)` (monolithic aggregate adapter) | Error returned immediately; halts command execution with CLI exit code 11. |
| `WorkspaceService.Open` | [`internal/core/workspace.go:103`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/workspace.go#L103) | `source.Store`, `WithTaskGraphSource(source.TaskGraphs)`, `WithThreadStore(source.Threads)`, `WithThreadPathSource(source.ThreadPaths)` | Returns zero `Workspace{}` and wrapped error before returning to caller. |
| `NewBuiltinTemplateService` | [`internal/core/service.go:334`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L334) | `MustNewService(nil)` | Uses built-in templates only; zero data capabilities; fails closed on any data operation. |

#### Injectable planning capabilities validated by `validateSourceSets`
| Port Name | Interface Type | Discovery from `Store` | Functional Option |
|:---|:---|:---:|:---|
| **aggregate store** | [`core.Store`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L297) | Direct parameter | `NewService(store, ...)` |
| **lint reads** | [`core.LintSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/lint_source.go) | `store.(LintSource)` | [`core.WithLintSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L67) |
| **audit snapshots** | [`core.AuditSnapshotSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/audit_source.go) | `store.(AuditSnapshotSource)` | [`core.WithAuditSnapshotSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L81) |
| **task graph reads** | [`core.TaskGraphSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service_task.go#L144) | `store.(TaskGraphSource)` (fallback: `taskStoreGraphSource`) | [`core.WithTaskGraphSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L53) |
| **task graph mutations** | [`core.TaskGraphMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L76) | `store.(TaskGraphMutationStore)` | [`core.WithTaskGraphMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L134) |
| **task graph repairs** | [`core.TaskGraphRepairStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L95) | `store.(TaskGraphRepairStore)` | [`core.WithTaskGraphRepairStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L146) |
| **task lifecycle mutations** | [`core.TaskLifecycleMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L112) | `store.(TaskLifecycleMutationStore)` | [`core.WithTaskLifecycleMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L156) |
| **Thread reads** | [`core.ThreadStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L225) | `store.(ThreadStore)` | [`core.WithThreadStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L170) |
| **Thread paths** | [`core.ThreadPathSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L237) | `store.(ThreadPathSource)` | [`core.WithThreadPathSource`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L188) |
| **Thread creation** | [`core.ThreadCreationMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L248) | `store.(ThreadCreationMutationStore)` | [`core.WithThreadCreationMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L198) |
| **Thread mutations** | [`core.ThreadMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L261) | `store.(ThreadMutationStore)` | [`core.WithThreadMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L207) |
| **Thread apply** | [`core.ThreadApplyMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L274) | `store.(ThreadApplyMutationStore)` | [`core.WithThreadApplyMutationStore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L217) |

---

### 3. Falsification matrix: Port combinations and outcomes

A systematic matrix of port combinations was evaluated with predicted vs observed outcomes:

| # | Selected Ports & Options | Token Assignment | Predicted Outcome | Observed Sandbox Result | Error / Behavior |
|:---:|:---|:---|:---:|:---:|:---|
| **1** | `NewService(store)` | `store`: $T_1$ | **Accept** | **Accept** | Svc initialized; auto-discovers all ports with token $T_1$. |
| **2** | `NewService(store)` without `SourceSetProvider` | `store`: no interface | **Reject** | **Reject** | `aggregate store does not publish a source-set identity` |
| **3** | `NewService(store)` with zero token | `store`: $T_0$ (`IsZero()==true`) | **Reject** | **Reject** | `aggregate store publishes an empty source-set identity` |
| **4** | `NewService(store)` with unstable token | `store`: returns new token each call | **Reject** | **Reject** | `aggregate store publishes an unstable source-set identity` |
| **5** | `NewService(nil)` | None | **Accept** | **Accept** | Pure template service (`NewBuiltinTemplateService`); data calls fail closed. |
| **6** | `NewService(nil, WithTaskGraphSource(g))` | `g`: $T_1$ | **Accept** | **Accept** | Pathless/read-only board works; task listing fails closed with clean error. |
| **7** | `NewService(store, WithTaskGraphSource(g))` | `store`: $T_1$, `g`: $T_2$ | **Reject** | **Reject** | `aggregate store and task graph reads address different source sets` |
| **8** | `NewService(store, WithThreadStore(th))` | `store`: $T_1$, `th`: $T_2$ | **Reject** | **Reject** | `aggregate store and Thread reads address different source sets` |
| **9** | `NewService(store, WithThreadPathSource(p))` | `store`: $T_1$, `p`: $T_2$ | **Reject** | **Reject** | `aggregate store and Thread paths address different source sets` |
| **10** | `NewService(store, WithThreadStore(th))` (where `store` has paths, `th` doesn't) | `store`: $T_1$, `th`: $T_1$ | **Accept** | **Accept** | Implicit aggregate paths detached (`s.threadPaths == nil`). Pathless Thread service. |
| **11** | `NewService(store, WithThreadStore(th), WithThreadPathSource(p))` | `store`: $T_1$, `th`: $T_1$, `p`: $T_1$ | **Accept** | **Accept** | Explicit paths preserved; `ResolveThreadPath` succeeds. |
| **12** | `NewService(store, WithThreadPathSource(p), WithThreadStore(th))` | `store`: $T_1$, `p`: $T_1$, `th`: $T_1$ | **Accept** | **Accept** | Reverse option order retains explicit paths (`s.threadPathsExplicit == true`). |
| **13** | `NewService(nil, WithThreadStore(th), WithThreadPathSource(p))` | `th`: $T_1$, `p`: $T_2$ | **Reject** | **Reject** | `Thread reads and Thread paths address different source sets` |
| **14** | `NewService(nil, WithLintSource(lint))` | `lint`: $T_1$ | **Accept** | **Accept** | Auto-defaults `s.auditReads = lint` (both $T_1$). Audit queries succeed. |
| **15** | `NewService(nil, WithLintSource(lint), WithAuditSnapshotSource(audit))` | `lint`: $T_1$, `audit`: $T_2$ | **Reject** | **Reject** | `lint reads and audit snapshots address different source sets` |
| **16** | `NewService(nil, WithAuditSnapshotSource(audit), WithLintSource(lint))` | `audit`: $T_2$, `lint`: $T_1$ | **Reject** | **Reject** | Reverse order; `auditReadsExplicit=true` prevents overwrite. Rejected. |
| **17** | `NewService(nil, WithTaskGraphMutationStore(mut))` | `mut`: $T_1$ | **Accept** | **Accept** | Guarded mutation service without store. |
| **18** | `NewService(nil, WithTaskGraphMutationStore(mut))` without provider | `mut`: no interface | **Reject** | **Reject** | `task graph mutations does not publish a source-set identity` |
| **19** | `NewService(store, WithTaskGraphSource(typedNil))` | `store`: $T_1$, `typedNil`: `(*gFake)(nil)` | **Accept** | **Accept** | Typed nil ignored; keeps default store task graphs. |
| **20** | `NewService(typedNilStore)` | `typedNilStore`: `(*fakeStore)(nil)` | **Accept** | **Accept** | Evaluates as nil store; behaves identically to `NewService(nil)`. |
| **21** | `WorkspaceService.Open` with cross-corpus threads | `Store`: $T_1$, `Threads`: $T_2$ | **Reject** | **Reject** | Fails before use with `ErrIncompatibleCapabilities`; 0 reads executed. |
| **22** | Two `FS` instances on same root without `WithSourceSetID` | `fs1`: $T_1$, `fs2`: $T_2$ | **Reject** | **Reject** | Independent instances do not match by path or record ID. |
| **23** | Two `FS` instances on same root with `WithSourceSetID(T1)` | `fs1`: $T_1$, `fs2`: $T_1$ | **Accept** | **Accept** | Explicit deliberate binding succeeds. |
| **24** | `FS` configured with `WithSourceSetID(SourceSetID{})` | `fs`: $T_0$ | **Reject** | **Reject** | `aggregate store publishes an empty source-set identity` |

---

### 4. Hostile analysis: Omitted capabilities and shared test tokens

#### Why `WorkspaceSource.Layout` and watcher use are safely outside the planning-data boundary
1. **Interface definition:** [`core.Layout`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/store.go#L331-L333) declares a single method: `WatchPaths() []string`.
2. **Behavioral scope:** `Layout` is strictly an operating system notification hint used by [`internal/tui/atlas.go:181`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/tui/atlas.go#L181) and [`internal/tui/tui.go:47`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/tui/tui.go#L47) to register directory paths with `fsnotify`.
3. **Absence of planning-data access:** `Layout` cannot read, write, parse, filter, project, or validate any planning document.
4. **Guarded execution:** Even if an alien `Layout` were configured, a file watcher notification merely triggers an interactive re-read through `workspace.Planning` ([`core.Service`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go)). Because `workspace.Planning` is validated by `SourceSetID`, all actual planning data accesses remain strictly within the validated corpus.

#### Why direct CLI `Fixer` and `Linter` wiring cannot bypass composition
1. **Direct adapter binding:** In [`internal/cli/root.go:457-459`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/cli/root.go#L457-L459):
   ```go
   a.Svc, err = core.NewService(fs)
   if err != nil { return err }
   a.Fixer = fs
   a.Layout = fs
   a.Linter = fs
   ```
2. **No capability splitting:** Neither `Fixer` nor `Linter` is a functional option on `core.NewService`. Neither exists on [`core.Service`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go).
3. **Execution context:** CLI commands `tskflwctl lint --fix` and `tskflwctl lint --links` invoke `app.Fixer.FixFrontmatter` and `app.Linter.DanglingLinks` directly against the single filesystem store instance resolved for the repository. They cannot be split, mixed, or cross-wired with external adapters.

#### Why the test fake shared token (`testSourceSetID`) does not mask defects
1. **Compilation boundary:** `var testSourceSetID = NewSourceSetID()` is declared in `internal/core/service_epic_test.go:20` within test files. It is never compiled into production binaries.
2. **Production isolation:** Production adapters ([`store.NewFS`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/store/fsstore.go#L108)) mint independent `SourceSetID` nonces via `core.NewSourceSetID()` for every instance.
3. **Test discrimination:** Tests that verify cross-corpus rejection ([`workspace_test.go:156`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/workspace_test.go#L156), [`source_set_test.go:52`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/source_set_test.go#L52), [`fs_test.go:89`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/workspacestore/fs_test.go#L89)) explicitly mint new IDs or instantiate separate workspaces. When the validator was disabled in mutation probe 1, the test suite immediately caught the bypass.

#### Evidence of zero token leakage
Grep verification confirmed zero occurrences of `SourceSet` or `source_set` across:
- `internal/domain` (domain structs: `Task`, `Epic`, `Audit`, `Research`, `Thread`)
- `internal/wire` (wire DTOs and JSON envelopes)
- `internal/cli/testdata/golden` (JSON Schemas and golden command outputs)

---

### 5. Mutation testing results

Six targeted mutation experiments were executed in the sandbox to verify test suite sensitivity:

| # | Mutated Guard | Target File / Subsystem | Test Result | Analysis |
|:---:|:---|:---|:---|:---|
| **1** | Omit `lifecycleMutations` from `validateSourceSets` | `internal/core/service.go:298` | **FAIL**: `TestNewServiceRejectsEveryMismatchedSplitCapability/lifecycle_mutation` failed (`want typed mismatch before use`) | Mutant killed. Explicit validation of lifecycle mutations is enforced. |
| **2** | Keep implicit Thread paths when `WithThreadStore` is called | `internal/core/service.go:178` | **FAIL**: `TestNewServiceRetainsImplicitThreadPathDetach` failed (`implicit aggregate paths after explicit Thread reads`) | Mutant killed. Detachment of implicit paths on semantic replacement is enforced. |
| **3** | Bypass missing provider check in `validateSourceSet` | `internal/core/source_set.go:49` | **FAIL**: `TestNewServiceRejectsMissingOrEmptySourceSetIdentity/missing_provider` failed (`want "does not publish"`) | Mutant killed. Unwitnessed capabilities fail closed. |
| **4** | Derive filesystem `SourceSetID` from root directory path | `internal/store/fsstore.go:111` | **FAIL**: `TestFilesystemSourceSetIsInstanceOwnedNotDerivedFromRootOrRecordID` failed (`must be unique per adapter instance`) | Mutant killed. Identity cannot be spoofed by path equality. |
| **5** | Ignore `NewService` error in `WorkspaceService.Open` | `internal/core/workspace.go:107` | **FAIL**: `TestWorkspaceService_OpenRejectsCrossCorpusThreadCapabilitiesBeforeUse` failed (`cross-corpus workspace error = <nil>`) | Mutant killed. Workspace open propagates composition failure. |
| **6** | Allow zero `SourceSetID` in `validateSourceSet` | `internal/core/source_set.go:53` | **FAIL**: `TestNewServiceRejectsMissingOrEmptySourceSetIdentity/empty_provider` failed (`want "empty source-set"`) | Mutant killed. Empty zero-value witness is rejected. |

---

### 6. Residual risks and sequencing with task `6gcwcf88z57p`

1. **Sequencing with split path capabilities (`6gcwcf88z57p`):**
   - The subsequent task `6gcwcf88z57p` will split `ResolveTaskPath`, `ResolveEpicPath`, `ResolveAuditPath`, and `ResolveResearchPath` out of semantic read stores into dedicated path ports (`TaskPathSource`, `EpicPathSource`, etc.).
   - Those new path ports must implement [`core.SourceSetProvider`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/source_set.go#L28) and be registered in [`s.validateSourceSets()`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/internal/core/service.go#L290), mirroring `ThreadPathSource`. The composition architecture in `6gdx7mcqm371` is ready to receive them without structural changes.
2. **Adapter honesty:**
   - As documented in the brief, `SourceSetID` is process-local composition evidence, not a cryptographic proof. An adapter that deliberately copies another adapter's nonce can fool the check within the same process. This is the intended contract boundary.

---

### 7. Validation commands executed

The following validation commands were executed inside the isolated sandbox:

```sh
# Establish baseline suite pass
go test ./...

# Verify race-free concurrency across all packages
go test -race ./...

# Check git diff hygiene (no trailing whitespace or conflict markers)
git diff --check

# Verify isolated workspace verification status
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
```

---

### 8. Mandatory reviewer isolation attestation

```
isolated-review: sandbox initialized at /private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.gdgxqa/.git
baseline_commit=26dbef171f6754377bce4e0654a67c77e9c6f5c9
source_blob=4b685353b5db7ecb9402db4ec04837a79672df94
source_fingerprint=736c54880b5d796b4badb7a0ccdcf8a46ee63bfe
deliverable=planning/audits/6ge63hzwtwn2-2026-09-27-source-set-capability-composition-implementation-antigravity.md
deliverable_changed=true
transfer=pending
```
