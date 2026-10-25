---
schema: 1
id: 6gq03e44pcyz
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Prevent malformed checkbox markers from making task completion treat an intended criterion list as empty.
effort: half day
tier: 2
priority: high
autonomy_level: 2
tags: [criteria, validation, lint, tui]
created: "2026-10-24"
---
# Guard task completion against unreadable criterion markers

## Objective

Keep a malformed checkbox from silently disappearing from the completion guard,
while distinguishing actual damaged criterion syntax from citations, links and
ordinary prose. Establish an explicit policy before broadening validation.

## Reproduction

The [section-selection repair](6gpxh6ncaqpw-prevent-narrative-headings-from-hiding-task-criteria.md)
self-review confirmed this preexisting adjacent gap with a real binary: a task
whose canonical criterion section contains `- [] intended criterion` reports
0/0 in `task info --json`; normal `lint --json` diagnoses the malformed checkbox
and exits 11, but `task complete --dry-run` permits completion with exit 0.
No task was actually completed by the reproduction.

Section selection is unambiguous here. `malformedCheckbox` already detects the
broken marker, but criterion enumeration omits it and completion sees no unmet
criteria. This is not repaired by changing which heading selects the section.

## Policy checkpoint

Recommend refusing completion when criterion markers cannot be interpreted,
including forced completion, previews and no-ops. Decide the read/tally behavior
explicitly: never imply that an unread intended criterion is a known empty list.
Keep the detector conservative around `[1]`, `[-]` and Markdown links, and do not
expand this into a new criterion-state/reason policy. Headings and marker spelling
require explicit repair; no guessed checkbox conversions or document migration.
Classify any machine behavior change under
[ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md).

## Acceptance criteria

- [ ] An owner-approved malformed-marker policy aligns criterion reads, lint, edits and completion without conflating non-checkbox Markdown with criteria.
- [ ] Malformed intended criteria cannot bypass completion, including force, previews and already-completed no-ops; refusals preserve bytes and modes.
- [ ] Tests cover damaged empty/spaced/x markers, valid checked/unchecked markers, fenced examples, citations, links and explicit repair.
- [ ] Owning comments/help, machine-contract classification, focused tests and normal gates agree with the implemented policy.

## Out of scope

Criterion-section selection, tracked-reason policy, bulk rewriting and TUI editor design.

## Related

- [TUI criterion actions](6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion.md): establish readable criteria before extending native authoring.
- [Historical lint guard](6fpfcecdymca-structure-aware-body-mutation-and-metadata-reads.md): malformed-checkbox detection exists, but is currently diagnostic-only.
- [Domain body policy](../../internal/domain/body.go) and [completion policy](../../internal/core/task_lifecycle.go).
