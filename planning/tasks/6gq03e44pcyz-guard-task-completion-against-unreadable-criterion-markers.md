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
updated_at: "2026-11-08"
audit_sources: [2026-11-08-weekly-task-sweep]
audited: "2026-11-08"
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

## Sweep verification (2026-11-08)

Automated weekly sweep — first recorded audit. The defect **still reproduces** on a
binary built from this checkout, and one part of it has been narrowed by code that
landed the day after this task was filed.

### Reproduction re-run (isolated planning tree outside this repo)

A task whose canonical `## Acceptance criteria` section contains only
`- [] intended criterion`:

| Command | Result |
| --- | --- |
| `lint --json` | diagnoses `malformed acceptance checkbox "[]"` |
| `lint` | exits **11** |
| `task complete --dry-run` | `would move … -> completed`, exits **0** |

The Reproduction section is therefore accurate as filed, and ACs 1–4 are all unmet.
`LintAcceptanceCriteria` still says so in its own words at
`internal/domain/body.go:712–714`: "malformed checkboxes and invalid states remain
diagnostic-only here." The completion guard counts only `domain.UnexplainedCriteria`
(`internal/core/task_lifecycle.go:329`), which enumerates canonical markers, so the
malformed line is invisible to it.

### What moved since 2026-10-24: the live gap is narrower than filed

`fix(criteria): make section selection explicit and fail closed` (`67492d6`,
2026-10-25 — the same repair whose self-review surfaced this task) added
`malformedCheckbox` to the section-selection scan at
`internal/domain/criterion_section.go:62–65`, so a malformed marker now marks its
span as a checklist. Re-running the reproduction with the heading renamed to
`## AcceptanceCriteria` confirms the consequence:

- **non-canonical heading + malformed marker** → now fails closed on both
  `task info` and `task complete --dry-run`: "unrecognized acceptance checklist
  heading … rename it to `Acceptance criteria`".
- **canonical heading + malformed marker** → still silently completable.

The live hole is therefore exactly the canonical-heading case. That sharpens the
Policy checkpoint rather than changing it: the detector is already trusted to
*select* a section, so trusting it to *refuse a read* is a smaller step than it was
when this was filed. Scope and ACs unchanged; recorded here rather than rewritten
into the Reproduction section, which stays as filed.

## Progress Log

- 2026-11-08: automated weekly sweep — defect re-reproduced end to end; recorded
  that the 2026-10-25 section-selection repair already fails closed for
  non-canonical headings, narrowing the live gap to the canonical-heading case.
