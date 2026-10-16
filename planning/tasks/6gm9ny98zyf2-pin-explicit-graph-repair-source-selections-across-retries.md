---
schema: 1
id: 6gm9ny98zyf2
status: completed
epic: 21-code-quality-architecture-hardening
description: Prevent explicit graph-repair retries from switching selected source ID/local-path addresses while reauthorizing fresh defects.
effort: 1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, repair, concurrency]
created: "2026-10-16"
started_at: "2026-10-16"
updated_at: "2026-10-16"
completed_at: "2026-10-16"
---
# Pin explicit graph-repair source selections across retries

## Objective and evidence

An explicit repair selection must not switch its source ID/local-path address after
a pre-commit conflict. Preserve fresh defect authorization and convergent repair, not
a cached write set. This is a bounded follow-up to the three-task safety PR, not a
universal mutation framework or an implicit extension of `--auto` semantics.

The [Claude safety review](../audits/6gm8vm7aq7zt-2026-10-16-selected-identity-safety-implementation-claude.md)
raised H-a as unproven. Owner validation confirmed it on 2026-10-16:

1. Seed distinct task IDs A/alpha and B/bravo, both with `depends_on: [invalid-token]`.
2. Request `RepairTaskGraph` dropping that declaration using `Source.TaskSlug=alpha`.
3. At `testHookBeforeGraphRepairVerify`, rename A to A-renamed and B to B-alpha,
   keeping the IDs and contents unchanged. The first attempt detects the source CAS
   conflict before any write.
4. On retry, `normalizeRepairEditSource` reselects B by the old slug. The call returns
   nil and `AppliedFiles` names B-alpha; A retains its defect. Strict improvement and
   removal-only policy do not establish that B was the requested source.

The compiler-valid throwaway real-store probe passed its bug-reproduction assertion
under `go test -race -run '^TestOwnerProbeGraphRepairRebindsExactSlug$' -count=1 -v`.
Fixture IDs were `ekpd8vnydg2h` (A) and `h46gx48dzz67` (B). The probe was replaced
with the opposite red-first safety regression below. No real planning artifacts
were used as fixtures.

## Scope and constraints

- Inventory CLI drop/clear selectors, portable `TaskGraphSourceEdit` intent and
  `RepairTaskGraph` retry behavior. Resolve explicit selectors first inside the guard;
  after selection retain enough source evidence to refuse alias/path/ID replacement.
- Broken graphs may contain duplicate IDs and readable physical occurrences. Do not
  collapse source identity to task ID alone, use opaque Location as authority or lose
  the existing exact local-path disambiguation. Ask before choosing any new identity
  or absence policy that changes accepted repair contracts.
- Recompute current defects, requested declaration occurrence, strict improvement,
  residual problems and task/Thread impacts. Preserve unrelated concurrent changes,
  malformed-Thread tolerance and committed-prefix retry stops.
- Treat auto-cleanup separately: it deliberately rediagnoses the current graph and
  does not mean a user's prior explicit source selection may be replaced.

## Acceptance criteria

- [x] A real-store red-first regression reproduces the confirmed slug-swap failure;
  fixed behavior never repairs B when A was first selected.
- [x] Explicit ID/slug/path selections pin the source ID plus exact local path;
  moved, missing, unreadable, replaced-ID or indistinguishable occurrences are
  refused, with duplicate-ID disambiguation and same-address replacement limits
  documented.
- [x] Portable pre-planner/empty-result conflicts and real adapter tests preserve
  fresh authorization, unrelated changes, preview/no-op behavior and durable receipts.
- [x] Auto-cleanup semantics remain deliberate; source identity policy is documented
  beside owning core/adapter code and the focused planning-data/persistence guides.
- [x] Normal build, race tests, lint and independent hostile review pass.

## Sequencing

Completed as the fourth bounded fix in
[Finish selected identity safety](../threads/6gm4h71hm1yt-finish-selected-identity-safety.md).
The first three fixes shipped in PR #310; they are precedent, not artificial
dependencies. Broad entity-integrity/storage redesign stays separate.

## Implementation checkpoint (2026-10-16)

Implemented on `fix/pin-explicit-graph-repair-sources` from merged main.
`TestRepairTaskGraphRetryNeverReselectsReusedSlug` in
[store regressions](../../internal/store/graphrepair_target_test.go) first failed
with B's source in `AppliedFiles` and a nil error. It now requires refusal without
changing either source.

The user approved pinning the first selected source ID and exact local path for
this call, refusing moved/missing occurrences rather than following renames by
ID in a graph with possible duplicate IDs. Identity is the logical ID/path
address, not an inode or a guarantee against byte replacement at that address.
Moving a selected occurrence requires inspection and a new call.

Implementation retains explicit selections inside the first successful
planner callback, including when a subsequent adapter conflict returns an empty
result. Retry plans use fresh source content and diagnostics. Explicit caller
location constraints remain checks, not selectors; implicit snapshot locations
do not become identity. Auto selections are never added to the retained set.

Portable and real-store tests cover pre-planner/empty-result conflicts,
ID/slug/path move/deletion/unreadable refusal, portable path reuse by another ID,
duplicate-ID occurrences, current-content preservation, newly valid dependency
refusal, already-satisfied no-ops, auto-only/mixed behavior and durable retry stops.
Mixed real-store retries cover both current auto sources and all-files-unchanged
refusal when the explicit source moves.
The unreachable source-absent normalization branch was removed: declaration
absence is convergent; selected source absence is not.

No wire shape or published valid-input meaning changes: this corrects unsafe
wrong-target retry behavior rather than introducing a new repair selector or
machine vocabulary. No schema revision is proposed under ADR-0008.

The [new concurrency audit](../audits/6gm8g0f66k80-2026-10-16-concurrency-and-atomicity.md)
was triaged separately: M1/M2 go to
[lint relocation safety](6gmbfxy1d561-preserve-source-evidence-and-partial-outcomes-during-lint-id-relocation.md),
L1 to [watcher session binding](6gmbfxy96tpw-bind-tui-watcher-health-and-reload-messages-to-their-originating-session.md).
Neither adds a prerequisite or expands this Thread's selected-identity finish line.

## Validation and review checkpoint

- `just build`, full `just test` with race detection and `just lint` (0 issues) pass
  using scratch build/lint caches. Focused core/store repair race tests also pass
  with `-count=3`.
- CLI help and the generated repair reference describe the approved policy.
  Regenerating into a fresh temporary output and comparing all of `docs/cli`
  produces no differences; `just docs-check` also passes. No generated page was
  hand-edited.
- `./bin/tskflwctl -C . lint --links` passes. All criteria are met; the completed
  Thread is healthy and soundly drained at 4/4 members.

Both independent reviews used isolated clones and have confirmed delivery receipts.
Neither reported a product defect; both supplied discriminating compiler-valid
mutations and real production probes. Owner dispositions close the audits:

- [Claude — retry semantics and fresh authorization](../audits/6gmbkgc5e1w9-2026-10-16-explicit-graph-repair-source-pinning-implementation-claude.md).
- [Antigravity — real source targeting and preservation](../audits/6gmbkgcajfap-2026-10-16-explicit-graph-repair-source-pinning-implementation-antigravity.md).

Claude's small observations became a retained mixed-call store regression,
unreadable-source advice and a direct portable-test guide link. Pre-existing
location error classification, representative impacts and cross-call identity
are not changed or mislabeled as new product defects. The originating Claude
safety audit's H1 is now fixed rather than merely tracked.
