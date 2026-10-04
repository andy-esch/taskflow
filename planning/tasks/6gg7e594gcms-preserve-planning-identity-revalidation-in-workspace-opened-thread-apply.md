---
schema: 1
id: 6gg7e594gcms
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Restore guarded Thread apply parity between ordinary and workspace-opened planning services.
effort: 4-8 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, ports, threads, safety]
created: "2026-10-03"
updated_at: "2026-10-04"
depends_on: [6gcwcf8rxe72]
started_at: "2026-10-04"
---
## Objective

Give a workspace-opened planning service the same guarded Thread-apply identity revalidation as
the ordinary CLI opener, without moving filesystem discovery into a primary adapter.

## Baseline evidence

`internal/workspacestore/fs.go` constructs `store.FS` with mutation authorization but without
`WithPlanningIdentityReader`. `internal/appwiring/wiring.go` supplies both to ordinary CLI opening.
The workspace service consequently exposes Thread apply, but a valid plan fails even in dry-run:
`validation failed: planning repository identity cannot be re-read for Thread apply`.

An isolated `TestWorkspaceIdentityParityProbe` opened the same initialized corpus both ways,
composed the same one-member manifest, and called `ApplyThreadPlan(plan, true)`. Direct opening
passed; workspace opening failed with that message. This predates the composition extraction and
fails closed; no current TUI apply workflow is being declared broken.

## Scope

- Reuse or consolidate local opening/identity-reader wiring at the secondary/composition boundary.
- Preserve per-invocation mutation authorization and checked source-set composition.
- Do not turn optional local discovery into a semantic `core` dependency.
- Own the direct/pointer identity-replacement matrix here: change the discovered root, replace
  the durable planning identity, or fail the identity re-read between compose and apply; all
  must fail closed without persistence. The final Thread regression pass consumes this evidence
  rather than duplicating the matrix.

## Acceptance criteria

- [x] Direct and pointer workspace opening can compose and dry-run a valid Thread apply plan.
- [x] Direct and pointer workspace applies reject repointed roots, replaced
  durable planning identity, and unavailable or failed identity re-reads before
  persistence; no permissive fallback is used.
- [x] Authorization is preserved for read-only, dry-run, and committing callers.
- [x] Tests compare the ordinary and workspace opening paths; no permissive identity fallback exists.
- [x] Shared opening code, if introduced, retains lazy discovery and one observed initial corpus.

## Related

- [Composition isolation](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Final guarded-contract regression pass](6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md)

## Out of scope

- Adding Thread apply to the TUI or changing its preview UX.
- Changing graph or persisted Thread formats.

## Implementation and verification (2026-10-04)

Reproduced the baseline in `TestPlanningOpenersThreadApplyParity`: ordinary direct/pointer
opening passed; both workspace paths failed with the missing identity-reader validation error.
The initial fixture contains valid tagged tasks and a real dependency addition plus Thread create.

`workspacestore.NewPlanningStore` now constructs one filesystem store from the already-observed
configuration, installing both the invocation authorizer and an identity re-reader. Ordinary binary
wiring and workspace opening share it. It does no initial rediscovery and copies the marker/root
anchor before retaining the callback; caller mutation of the config object cannot redirect it.
Pointer identity is re-read from the pointer checkout, not the resolved planning target or cwd.
Missing construction inputs and empty/nil fresh discovery results fail explicitly rather than
falling back to cached identity. Legacy id-less trees remain readable but cannot apply durable plans.

The helper stays in the local secondary adapter. Core ports, source-set checking, watcher layout,
wire fields, and graph/prefix transaction semantics are unchanged. Architecture documentation now
records this shared construction and removes the retired direct `appwiring -> store` import edge.
Nil authorization retains its existing behavior: choosing an explicit policy is still owned by
6gg7e59cyxxh and its user design checkpoint, not pre-decided here.

### Regression evidence

- `TestPlanningOpenersThreadApplyParity`: both ordinary/workspace routes and direct/pointer
  entries compose, dry-run without document changes, then commit the dependency and Thread.
- `TestPlanningOpenersRejectChangedIdentityBeforeThreadApply`: the same four opening paths reject
  root repointing, replaced IDs, removed markers, and malformed markers, for dry-run and commit.
  Same-ID root replacements contain matching tasks so the root guard itself must be load-bearing;
  the flat pointer target also prevents a wrong discovery anchor from failing for an unrelated reason.
- `TestPlanningOpenersPreserveThreadApplyAuthorization`: reads/compose remain usable for a denied
  caller, both apply modes are denied before writes, and the retained live policy can subsequently
  authorize the operation. Refusals preserve the typed failure and attempted Thread token.
- Shared-constructor tests observe real discovery only at apply, preserve the original watcher root
  and non-zero source-set witness, reject missing readers/nil fresh results, preserve reader failure
  causes, and re-read a genuinely marker-less legacy root without inventing an ID.

Five compiler-valid mutations were tested one at a time in
`/private/tmp/taskflow-workspace-parity-probes.5Gona7`, a disposable copy with no shared Git metadata:
cached identity instead of discovery; dropping the authorizer; anchoring at the resolved pointer
target; bypassing the physical-root comparison; and bypassing the plan/repository ID comparison.
Each named regression failed on the unsafe success/receipt, not fixture validation. All probes were
restored, the three mutated source files compared byte-for-byte with the worktree, and restored
focused tests passed. These are owner probes, not independent review results or permanent mutations.

### Validation and lifecycle

`just build`, `go test ./...`, `go test -race -count=1 ./...`, `just lint`, and
`just docs-check` pass (temporary Go/lint caches). Planning/audit lint and the Thread frontier are
checked at handoff. Local implementation criteria are met, but lifecycle remains **in-progress**
pending merge. Implementation, regression tests, and architecture documentation are committed in
`8edd869`; both external reviews are reconciled below. No merge, release inclusion, TUI Thread-apply
UX, or public-constructor authorization policy is claimed here.

## External review handoff

- [Codex](../audits/6ggee2fx4053-2026-10-04-workspace-thread-apply-identity-parity-implementation-codex.md)
- [Antigravity](../audits/6ggee2g5ejtq-2026-10-04-workspace-thread-apply-identity-parity-implementation-antigravity.md)

At handoff these were unreviewed briefs, not independent ready verdicts. Each reviewer captured
the working-tree overlay in an independent clone and transferred only its assigned audit after
restoring probes. The target included untracked Go files; the handoff HEAD alone did not contain
that work. Review evidence describes the pre-commit overlay; the implementation is now committed above.

## External review reconciliation (2026-10-04)

[Antigravity](../audits/6ggee2g5ejtq-2026-10-04-workspace-thread-apply-identity-parity-implementation-antigravity.md)
reported no actionable findings with seven compiler-valid mutation probes. Owner verification
accepted the conclusions, confirmed restored sandbox source/test parity, and reran focused uncached
race tests successfully. Temporary report links and overbroad evidence wording were corrected;
no production change or new follow-up is warranted by this report.

[Codex](../audits/6ggee2fx4053-2026-10-04-workspace-thread-apply-identity-parity-implementation-codex.md)
found two valid local issues, now fixed: M1 demonstrated that injected-reader counters missed
direct real rediscovery; L1 found the missing `workspacestore -> domain` edge in the architecture
inventory. No production persistence defect was demonstrated.

`TestPlanningOpenRetainsInitialCorpusWhenMarkerChangesDuringDiscovery` now repoints real direct
and pointer markers after the first observation, before construction. Metadata, task reads,
watchers, and compose must retain that observation; dry-run and committing applies then refuse
the changed root, preserve the token, and leave both corpora unchanged. The file-scoped
`TestSharedPlanningConstructorUsesConfigOnlyAsData` prohibits direct config API references in
the shared constructor file, including function-value references and renamed imports. It is
not a general static proof against every possible indirect I/O path.

In `/private/tmp/taskflow-workspace-review-fixes.RToQZE`, an independent copy without Git metadata,
the discarded direct call fails the fitness test; adopting the rediscovered config fails both
the real-opener regression and fitness test; an aliased discovery function reference fails the
fitness test. Mutations were restored, the constructor compared byte-for-byte to this worktree,
and restored focused race tests passed. Partial valid discovery data accompanied by an error is
also a permanent refusal row, rather than only a temporary reviewer probe.

The review's pre-existing markerless migration hint is tracked as L2 in
[6ggfd81jg0qg](6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md),
in the CLI-contract Thread. It does not extend this refactor's closure prerequisites. Antigravity's
invented lock-file reference was corrected to the actual Darwin root-directory flock.

After reconciliation, full uncached race tests, lint, build, and generated-doc checks pass. Both
audits are settled locally; this task remains in-progress awaiting integration, not merge/release.
