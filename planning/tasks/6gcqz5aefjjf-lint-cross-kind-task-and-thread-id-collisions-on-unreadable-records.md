---
schema: 1
id: 6gcqz5aefjjf
status: completed
epic: 21-code-quality-architecture-hardening
description: Keep cross-kind identity collisions visible in lint when either task or Thread source is malformed.
effort: S
tier: 3
priority: medium
autonomy_level: 4
tags: [lint, diagnostics, identity]
created: "2026-09-22"
audit_sources: [planning/audits/6gch6xep2mh2-2026-09-22-correctness-and-errors.md]
depends_on: [6g5vm4efjcdv]
updated_at: "2026-10-04"
started_at: "2026-10-04"
completed_at: "2026-10-04"
---

# Lint cross-kind task and Thread ID collisions on unreadable records

## Objective

Make the planning-space-wide task/Thread ID invariant fail closed in repository lint. When either
document is malformed, consume the stable identity already recovered by the read boundary instead of
silently dropping the cross-kind collision from the report.

## Scope

- Seed task and Thread identity sets from safe recovered IDs on unreadable records as well as decoded
  entities; do not infer identity from arbitrary prose or an invalid filename.
- Preserve the existing create-time guards and duplicate-within-kind checks while making the lint
  rule symmetric for readable/unreadable task and Thread pairs.
- Land after the adapter-neutral lint-load diagnostic shape so the rule consumes one portable
  identity contract rather than adding another filesystem-specific path dependency.

## Current boundary (2026-10-04)

The lint-read prerequisite and workspace identity-parity task are merged. At the start of this
slice, `core.Service.Lint` omitted safely recovered IDs carried by task load problems and Thread
read problems. The local implementation now consumes those identity-bearing diagnostics without
requiring a source path. PR #280 merged the work; this task is **completed**. Thread-projection
behavior is not silently added to this lint slice.

## Implementation and evidence (2026-10-04)

- Seed cross-kind identity sets from explicitly recovered, strictly valid stable IDs. Never infer
  identity from locations, optional repair paths, slugs, messages, or untrusted frontmatter.
- Report each readable or unreadable collision owner while retaining the original load problem.
  A malformed Thread's within-kind duplicate and cross-kind collision share one result, in stable
  check order. Source occurrence order and independent opaque locations are preserved.
- Keep recovered task ownership separate from valid Thread membership. Preserve readable
  source/frontmatter drift checks, readable collision policy, creation guards, and wire shapes.
- Document the trust rule inline on `LoadProblem` and in the architecture's lint-port section.

The new core readability matrix failed on all three missing combinations before the fix. Permanent
core regressions cover pathless records, source versus declared-ID drift, malformed/missing IDs,
untrusted declarations, distinct IDs, combined duplicate/collision attribution, and no rescanning.
CLI regressions exercise real malformed YAML and invalid filenames, preserved load diagnostics,
exit 11, and repeatable human/JSON ordering with multiple IDs and duplicate Thread occurrences.
Existing guarded creation and unreadable-Thread refusal tests also pass unchanged.

Self-review used `/private/tmp/taskflow-collision-probes.qE2oqz`, an independent copy without Git
metadata. Four compiler-valid mutations each failed the intended regression: omit recovered task
identity; omit recovered Thread identity; accept malformed recovered IDs; and promote unreadable
tasks into valid membership. All mutations were restored, mutated source files compared byte-for-byte
with the worktree, and restored focused tests passed. This is owner verification, not external review.

`go test ./... -count=1`, `go test -race -count=1 ./...`, `just lint`, `just build`, and
`just docs-check` pass with temporary caches. Repository planning lint is clean and the Thread
frontier reports healthy graph/projection state. The source audit's M1 is fixed in PR #280;
release inclusion remains a separate milestone.

## Acceptance criteria

- [x] A readable task colliding with an unreadable Thread remains reported, and the inverse case is
      covered by a regression test.
- [x] Two unreadable records with the same safely recovered cross-kind ID are diagnosed when both
      read boundaries provide authoritative identity.
- [x] Missing, malformed, or untrusted recovered IDs do not create false collision reports.
- [x] Ordering and de-duplication remain deterministic in human and `--json` lint output.
- [x] Creation guards and readable-record collision behavior remain unchanged.

## Out of scope

- Redesigning the lint diagnostic vocabulary or `lint --fix` applicability.
- Expanding stable-ID uniqueness beyond the currently decided task/Thread invariant.
- Repairing malformed records or changing guarded create behavior.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-22 correctness and errors](../audits/6gch6xep2mh2-2026-09-22-correctness-and-errors.md), finding M1
- Predecessor [adapter-neutral lint load diagnostics](6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md)
- [Adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Final guarded-contract regression pass](6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md)
