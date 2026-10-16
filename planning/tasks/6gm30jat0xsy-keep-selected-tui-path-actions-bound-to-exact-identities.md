---
schema: 1
id: 6gm30jat0xsy
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Prevent delayed editor and path-copy actions from resolving another record when their selected source ID disappears.
effort: 1-2 days
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, identity, tui, core, store]
created: "2026-10-15"
updated_at: "2026-10-16"
started_at: "2026-10-15"
---

# Keep selected TUI path actions bound to exact identities

## Objective

Keep a selected source ID distinct from a fresh human-reference query across the
optional local-path boundary. Preserve parse-free repair access and ordinary CLI
path ergonomics without letting a delayed TUI action open/copy another record.

## Evidence and scope

During [selected task action work](6gkt3fq0pv26-preserve-selected-tui-task-identities-through-service-actions.md),
owner probes queued real `E` and `Y` actions, removed the selected task, and created
a different-ID task whose slug was the vanished ID. Both the initial and confirmation
lookups resolved the impostor; the reducer accepted its path and returned an editor
or clipboard effect with no error. Effects were deliberately not executed. The two
lookups protect rename timing, not identity, when both use the fuzzy path port.

Local evidence (not a durable test dependency):
`/private/tmp/taskflow-selected-action-neighbors.uYLhPU/selected_identity_neighbors_probe_test.go`.
The race-enabled probes passed their bug-reproduction assertions; they were removed
from the repository rather than committed as passing safety coverage.

Inventory `localPathLookup`, `load*Detail`, and each optional entity path capability.
Give already-selected lookup explicit identity-only semantics without requiring a
typed decode of broken frontmatter, opening opaque `Location`, or changing CLI
reference policy globally. Stop for input if a broader capability/compatibility
decision is needed. Existing detail source-ID guards already refuse a mismatched
read; do not claim an unprobed wrong-detail acceptance defect.

## Acceptance criteria

- [x] Selected editor and path-copy actions follow rename but refuse deletion,
  duplicate IDs and ID-shaped or case-folded impostor labels without launching
  an effect on another record.
- [x] Selected local lookup has an explicit identity-only application/adapter
  contract; CLI path commands retain normal human-reference matching and
  parse-free access to malformed documents.
- [x] Real queued E and Y command tests cover drift before each lookup, stale
  selection/reload/session results and optional pathless capabilities; no editor
  or clipboard effect runs during safety tests.
- [x] Inventory every entity path family and selected-detail read: retain
  existing source-mismatch refusal and track any separately confirmed exposure
  without parsing diagnostic locations.

## Out of scope

- New navigation UX or a universal entity/action framework.
- Task field/lifecycle writes, handled by the selected-task action task.
- Non-task mutation entry points, tracked in
  [non-task target pinning](6gkt3fphx1ts-pin-non-task-mutation-retries-to-selected-source-identities.md).

## Sequencing

Belongs to Refine Thread and TUI navigation. Independently actionable: the existing
source-ID and optional-path contracts are its prerequisites, not completion of an
unrelated navigation design or task-field helper implementation. Coordinate with
[structured action targets](6g86g03zfj2f-define-consistent-action-targets-for-structured-tui-detail-selections.md)
on eventual parent/child UX, but do not block this correctness fix on that design.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)

## Implementation and evidence

All five optional entity path ports now distinguish ordinary CLI references from
exact selected `*PathByID` lookups. Service validates nonempty source evidence;
the filesystem adapter resolves unique ownership without typed decoding. Existing
source-set/detachment policy is shared by both modes. TUI `localPathLookup`, all four
ordinary detail loaders and the Thread projection detail loader use selected mode.
Semantic detail reads retain their existing source-ID mismatch refusal; no new
wrong-detail acceptance defect is claimed. Thread mutation materialization/CAS
also use the exact Thread resolver rather than a human-reference lookup.

[Queued local-action tests](../../internal/tui/selected_local_path_test.go) exercise
76 missing/impostor/case-folded/duplicate cases across E/Y and both lookups, ten
parse-free repair/same-ID rename controls, and four same-ID/list-generation workspace
session collisions. Epics use complete-stem identity and cannot duplicate a stem in
one directory. Existing [local-path tests](../../internal/tui/local_path_test.go)
cover selection/reload staleness, rename between lookups and detached/pathless sources.
No editor or clipboard effect is executed by these safety tests.

A compiler-valid owner mutation restoring fuzzy task lookup failed all eight
task impostor/case-folded cases at the intended effect assertion; restoring exact
lookup passed the race-enabled safety tests. Core optional-capability tests separately
pin opaque IDs and exact-mode dispatch without semantic decoding or local fallback.

Owner build/full race tests/lint/docs checks pass. Member of
[Finish selected identity safety](../threads/6gm4h71hm1yt-finish-selected-identity-safety.md);
independent review remains pending. Neither the final lookup-to-external-effect
window nor delete/recreate with the identical canonical ID is made transactional.
