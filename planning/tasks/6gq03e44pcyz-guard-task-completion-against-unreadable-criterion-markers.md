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
updated_at: "2026-11-15"
audit_sources: [2026-11-15-weekly-task-sweep]
audited: "2026-11-15"
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

## Sweep verification 2026-11-15

First automated weekly sweep of this task (no prior `audited:`). The reproduction was
re-run with a binary built from `main` (`5513c3d`) against a throwaway planning tree
outside this repository — no task in `planning/` was touched by the reproduction.

**Verified accurate (2026-11-15) — the gap is still open, exactly as described.**
With a canonical criterion section containing `- [] intended criterion`:

| Surface | Observed |
| --- | --- |
| `task info --json` | `"ac":{"checked":0,"total":0}` — the intended criterion is invisible |
| `lint --json` | diagnoses `malformed acceptance checkbox "[]"`, exits **11** |
| `task complete --dry-run` | `would move … -> completed`, exits **0** |

Code citations hold: `malformedCheckbox` exists at `internal/domain/body.go:768` and is
consumed only by `LintAcceptanceCriteria` (`:715`, call site `:730`) and the
section classifier `criterion_section.go:62` — it is still diagnostic-only and is not consulted by
criterion enumeration. The completion guard reads
`domain.UnexplainedCriteria` (`internal/core/task_lifecycle.go:329`, defined at
`body.go:487`), which sees no unmet criteria, so the refusal at `task_lifecycle.go:355-358`
never fires.

**Finding — the reproduction understates the impact.** The Reproduction section records
only that `--dry-run` permits completion ("No task was actually completed by the
reproduction"). This sweep ran the real verb in the throwaway tree: `task complete`
succeeded with exit 0 and wrote `status: completed`. The malformed marker is therefore
not merely a preview-level gap; a task whose only intended criterion is unreadable can
actually be completed. This strengthens AC 2 rather than changing it.

**Detector conservatism — evidence for AC 3.** The current detector already behaves the
way the Policy checkpoint asks. In the same throwaway tree, `- [-] dash marker`,
`- [1] numeric marker` and a link-only bullet (`- see [docs](../y.md) …`) were all left
alone — neither diagnosed nor counted as criteria — while `- [x ] trailing space in
marker` was diagnosed as malformed and excluded from the tally (`checked:0, total:1`
beside one valid criterion). So the `[1]`/`[-]`/Markdown-link edges AC 3 names are
already safe today, and `[x ]` is a second instance of the same uncounted-but-diagnosed
class as `[]`.

**Nothing ticked.** AC 1 is owner-approval-gated and the policy is undecided, so no
criterion was checked or given a state.

## Progress log

- 2026-11-15: automated weekly sweep — first audit; reproduction re-run on `5513c3d` in a throwaway tree and extended: the real `task complete` (not just `--dry-run`) completes a task whose only criterion is `- []`, and the detector is already conservative around `[-]`/`[1]`/links.
