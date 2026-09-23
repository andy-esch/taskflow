---
schema: 1
id: 6gcqz5aqt2sg
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Close silent enum-switch fallthroughs and make future vocabulary drift fail mechanically in tests and lint.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [correctness, lint, testing]
created: "2026-09-22"
audit_sources: [planning/audits/6gch6xep2mh2-2026-09-22-correctness-and-errors.md, planning/audits/6g8zy840z6gj-2026-09-11-test-rigour.md]
---

# Make closed-vocabulary switches fail closed and exhaustively checked

## Objective

Turn closed-vocabulary switch coverage from reviewer memory into an enforced invariant. Fix the three
known silent or misleading fallthroughs, cover the lifecycle-override vocabulary, and adopt the
`exhaustive` linter narrowly enough that deliberate defaults stay explicit rather than becoming
blanket suppressions.

## Scope

- Make `blockerReason` total over persisted task statuses so a future valid status cannot be
  mislabeled as terminal `invalid-status` and truncate the blocking frontier.
- Make relative revisit-date units and legacy dependency resolutions reject or diagnose unknown
  values explicitly; neither may fall through to data that looks valid.
- Give `TaskLifecycleOverride` an enumerable or mechanically checked vocabulary, including a negative
  test for an out-of-vocabulary value.
- Enable and configure the existing golangci-lint `exhaustive` analyzer after inventorying affected
  switches. Every suppression must name why a default is intentionally safe.

## Acceptance criteria

- [ ] Drift tests or mechanical checks fail when a persisted `domain.Status` is omitted from graph
      role or blocker classification.
- [ ] An unknown relative-date unit returns validation failure rather than `0001-01-01`.
- [ ] An unknown legacy dependency resolution cannot be reported as an empty field.
- [ ] `TaskLifecycleOverride` rejects an unknown value in a focused test and cannot silently drift
      from its validators.
- [ ] `golangci-lint run ./...` enforces exhaustive closed-vocabulary switches without broad package-
      level or analyzer-wide suppression.
- [ ] Existing status, revisit-date, dependency-lint, and lifecycle behavior remains unchanged for
      every declared value.

## Out of scope

- Adding new statuses, lifecycle overrides, date units, or legacy resolution modes.
- Reworking graph traversal or dependency lint beyond the unsafe fallthroughs.
- Treating every Go switch as an enum switch.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-22 correctness and errors](../audits/6gch6xep2mh2-2026-09-22-correctness-and-errors.md), findings L1 and L2
- Audit [2026-09-11 test rigour](../audits/6g8zy840z6gj-2026-09-11-test-rigour.md), finding M2
