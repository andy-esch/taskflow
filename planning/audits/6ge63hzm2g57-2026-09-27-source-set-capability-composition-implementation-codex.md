---
schema: 1
id: 6ge63hzm2g57
bucket: closed
area: source-set-capability-composition-implementation-codex
date: "2026-09-27"
---
# Audit: Source-set capability composition implementation — codex — 2026-09-27

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

**Verdict: no reproducible finding in the reviewed implementation.** I completed a construction/consumer pass and an adversarial pass against the shared validator, its callers, the filesystem identity mint, and the direct CLI channels. This is a review of the sandbox baseline (including the uncommitted handoff overlay), not a claim that the queued entity path split has shipped.

### Verified inventory and evidence

| Boundary | Code evidence | Result |
| --- | --- | --- |
| Identity and constructor | `internal/core/source_set.go:15-34,41-68`; `internal/core/service.go:228-305` | The ID is opaque, comparable, nonzero, process local; each selected planning port must publish a stable matching ID. `NewService` returns `(nil, typed error)` before handing out a service. `ErrIncompatibleCapabilities` wraps `domain.ErrValidation`. |
| Independently injected ports | `internal/core/service.go:50-222,290-305` | Task graph reads; lint reads and audit snapshots; graph mutations, repairs, and lifecycle mutations; Thread reads, paths, creation, mutation, and apply all appear in the one final validation matrix. `WithLintSource` defaults audit snapshots unless the explicit audit option was selected. Typed nil options remain absent. `WithThreadStore` detaches only implicit aggregate paths; an explicit path survives either option order. |
| Aggregate discovery and fallback | `internal/core/service.go:232-268`; `internal/core/service_task.go:150-165`; `internal/core/store.go:289-302` | Aggregate task/epic/audit/research persistence is one validated `Store`; its discovered lint, audit, graph, Thread, and mutator ports are selected before options. The fallback graph source delegates its witness to that store. A store without a provider fails construction. Entity-specific `Resolve*Path` methods still live on that same aggregate port; their extraction is the queued task `6gcwcf88z57p`. |
| Production construction | `internal/cli/root.go:428-460`; `internal/core/workspace.go:74-117`; `internal/workspacestore/fs.go:34-47`; `internal/core/service.go:282-286,334` | The CLI creates one `FS`, checks `NewService`, then assigns the same instance to its direct ports. Workspace discovery supplies one `FS` to all planning fields and `WorkspaceService.Open` propagates constructor failure. `MustNewService` calls the checked constructor and its only production caller is the repo-less built-in-template service. A repository-wide search for production `NewService`/`MustNewService` calls found these sites. |
| Filesystem witness | `internal/store/fsstore.go:36-67,107-115`; `internal/store/source_set_test.go:11-68`; `internal/workspacestore/fs_test.go:89-121` | Each `NewFS` mints its own ID. Identical root/record/location/revision and a different corpus with the same record ID both reject when paired; explicit `WithSourceSetID` permits a known same-corpus pair. Independent workspaces get distinct IDs and reject cross-wiring. |
| Direct channels and wire | `internal/cli/lint.go:42-55,79-85`; `internal/cli/root.go:453-459`; `internal/core/store.go:310-333`; `internal/cli/ui.go:141`; `internal/tui/tui.go:42-49`; `internal/tui/atlas.go:175-183` | Direct `Fixer` mutation and `Linter` link scans are excluded from `NewService` but production assigns both from its validated `FS` after success. `Layout` supplies only watcher paths and comes from the same `FS` in CLI/workspace production. Checkout/root/planning ID select or assert the workspace (`workspace.go:86-101`); they are not substituted for `SourceSetID`. No `SourceSetID` reference exists in `internal/domain` or `internal/wire`; the scoped CLI schema tests passed. |
| Deliberately deferred source | `internal/core/service.go:90-100`; `internal/core/template.go:5-31` | `WithTemplateSource` is not validated. The current production source is compiled-in `builtinTemplates`; repository-local templates are described as future epic 22 work and have no production option caller. A future repo-local source must be bound to the planning set when it becomes selectable. |
| Planning status | `planning/tasks/6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md`; `planning/tasks/6gdx7mcqm371-bind-split-planning-capabilities-to-one-source-set.md`; `planning/tasks/6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md` | The design is completed, this implementation task is in progress with unchecked acceptance boxes, and the source/path split is ready-to-start. The latter's proposed path ports and domain cleanup are requirements, not shipped evidence. |

### Runtime and mutation controls

- Focused source-set, workspace, and filesystem tests passed. `TestFS_MultiWorkspaceSourceSetsCannotBeCrossWired` passed separately. `go test -race ./...` passed after probe restoration. The temporary positive probe constructed a split service with a separate graph reader and Thread path provider sharing an honest token; its first `Board` read succeeded. The pathless read-only regression passed separately. Foreign path providers in both option orders returned a nil service and errors matching both `ErrIncompatibleCapabilities` and `domain.ErrValidation`, with the graph read counter unchanged. An unstable guarded mutator also rejected. Explicit lint/audit overrides retained their selected sources in both orders. A temporary CLI test confirmed exit code 11 for `ErrIncompatibleCapabilities`. All temporary tests were removed.
- Exact matrix-entry mutations were run separately for task graph, lint, audit, graph mutation, repair, lifecycle, Thread read, path, creation, mutation, and apply. Each named mismatch subtest failed at its assertion. The lint mutation removed both lint and its derived audit validation entries, so the sibling default could not preserve the result accidentally. Removing both aggregate and fallback graph entries made the missing-provider test fail. Disabling the provider-presence, zero, unstable, or cross-ID checks each failed its matching test. Replacing per-instance FS minting with a shared token failed the FS identity test. Removing implicit Thread path detachment or explicit-path marking failed their respective option tests. Breaking fallback token delegation failed the Thread-detach construction test with an empty-identity error. Making `WorkspaceService.Open` drop the constructor error failed its cross-corpus test; bypassing the constructor's final validation failed the graph mismatch test. Every probe restored its original file bytes; these are assertion failures, not build-only results.
- `internal/core/source_set_test.go:20-152`, `internal/store/source_set_test.go:11-68`, `internal/core/workspace_test.go:153-165`, and `internal/workspacestore/fs_test.go:89-121` are committed-in-baseline regression evidence. The shared default fake token at `internal/core/service_epic_test.go:20-31` can mask an accidentally cross-corpus *test fixture*, but explicit foreign-token fakes and real two-corpus filesystem tests provide negative controls. No production adapter uses that default. No repair command or new optional wire field is emitted by this source-set change; the repair-command and non-default-wire branches of the generic review floor have no target here. A repository search found no token in domain or wire types.

**Residual uncertainty:** The token witnesses adapter wiring, not physical equality or later staleness; the API explicitly relies on honest, stable adapters. A future entity-specific path port or repo-local template source can bypass this guarantee if a later change omits it from the constructor's selected-capability matrix. The current tree has no such shipped path port or repo-local template adapter. No finding is opened on future work.

### Isolation attestation

- Workspace path: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.QuER6Y`
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.QuER6Y/.git` (independent in-tree directory, no worktree link or object alternates; checked by helper)
- Sandbox baseline commit: `c059493b36f3bb0ab03f2f6b44cd89ba94869c5d`
- Captured source blob: `2537b1beb702b104a77f8faa88b5d64b16167dca`
- Captured source fingerprint: `736c54880b5d796b4badb7a0ccdcf8a46ee63bfe`
- Sole deliverable: `planning/audits/6ge63hzm2g57-2026-09-27-source-set-capability-composition-implementation-codex.md`
- Transfer result: `succeeded` via `scripts/isolated-review-workspace.sh transfer` after its fail-closed verification; sandbox retained for owner confirmation.
