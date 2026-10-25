---
schema: 1
id: 6gpxh6ncaqpw
status: completed
epic: 21-code-quality-architecture-hardening
description: Prevent narrative headings from hiding real task criteria in reads, writes, lint and completion guards.
effort: half day
tier: 2
priority: high
autonomy_level: 2
tags: [criteria, validation, lint, tui]
created: "2026-10-24"
started_at: "2026-10-24"
updated_at: "2026-10-25"
completed_at: "2026-10-25"
---
# Prevent narrative headings from hiding task criteria

## Objective

Keep a narrative heading mentioning acceptance from hiding the real criterion
section, disabling criterion edits, or bypassing the unmet-criterion completion
guard. Use one deliberate selection contract for reads, writes and lint.

## Reproduction and cause

Discovered during the [cross-kind TUI creation check](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md)
on 2026-10-24. In a fresh throwaway space, create a task through `task new
--body-file` with this body:

```markdown
# Shadowed criteria

### Integrated acceptance notes

Narrative evidence, not criteria.

## Acceptance criteria

- [ ] This must remain unmet.
```

Observed with the real binary: `task info --json` reports 0/0; `task ac --check 1`
refuses with exit 11; `lint --links` passes; `task complete --dry-run` permits
completion with exit 0. A control changing only the narrative heading to
"Integrated verification notes" reports 0/1 and refuses completion with exit 11.
No task was actually completed in the reproduction.

Before this repair, `scanAcceptanceCheckboxes` in [body.go](../../internal/domain/body.go) selected the
first heading containing "acceptance", while lint's `isAcceptanceSectionHeading`
used a narrower predicate. Canonical criteria after that narrative section were
not scanned. The integration document used a neutral heading as a workaround;
that closeout alone did not repair the parser.

## Selection contract checkpoint

The broad matcher is documented legacy behavior, not an incidental spelling
mistake. Decide canonical-heading precedence and treatment of legacy aliases and
multiple genuine sections before implementation. Prefer a canonical section over
incidental prose; diagnose ambiguous source structure rather than silently treating
an unread section as zero criteria. Review compatibility under
[ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md),
including changed tally/exit behavior. Do not silently migrate documents or alter
unrelated uses of the general-purpose `Section` helper.

### Approved policy — 2026-10-24

The owner approved all three choices:

- Recognize case-insensitive `Acceptance criteria`, including parenthetical notes
  such as `(draft)`, and the supported `Acceptance` alias. Incidental narrative
  mentions are not criterion sections.
- Multiple recognized sections are ambiguous, not first-wins. With no recognized
  section, a nonstandard acceptance heading containing a checklist must be
  diagnosed rather than silently counted as zero. Criterion edits and completion
  refuse until the document is clarified; lint explains the repair. No automatic
  document rewrites.
- Completion `--force` may bypass known unmet criteria, never ambiguous criterion
  structure. The structural check also applies to previews and completion no-ops.

Implementation shares selection between reads, edits, lint and completion;
raw body reads/editor repair remain available. The machine-contract assessment
must distinguish a formerly misleading zero tally from a valid empty section.

## Acceptance criteria

- [x] An explicit section-selection and legacy/ambiguity policy aligns criterion reads, mutations, lint and completion eligibility.
- [x] Narrative headings before and after canonical criteria cannot hide unmet criteria; real CLI tests pin tallies, refusal before effects and completion without force.
- [x] Regressions cover fenced examples, nested headings, CRLF, accepted legacy headings and multiple genuine sections without mutating unrelated Markdown.
- [x] Owning comments/help and any required machine-contract revision describe the actual policy; focused tests, normal gates and planning checks pass.

## Out of scope

TUI criterion-editor design, tracked-reason policy, a general Markdown parser
replacement, and bulk document rewrites.

## Implementation checkpoint

- [The selector](../../internal/domain/criterion_section.go) finds one explicit
  section before enumerating criteria. Reads now return validation errors rather
  than success-shaped zero/first-section results for invalid structure; all
  criterion transforms and lint use the same selection policy. General `Section`
  substring matching remains unchanged.
- Core completion checks structure before force handling and same-status no-ops.
  Criterion transforms revalidate inside the existing CAS callback and on retries,
  retaining selected source identity and operation time. Raw body repair remains
  available; no local-path capability, new persistence seam or automatic rewrite
  was introduced.
- CLI task info/ac propagate the error; the TUI keeps the body visible and renders
  an unavailable-criteria diagnostic. Archived task structure is linted without
  reintroducing unrelated archived field/state nags. Repair previews retain their
  existing proposal-only semantics; ordinary lint and real `--fix` diagnose these
  unrepairable structural errors without rewriting the headings.
- Machine revision `1.90` is explicitly `NOT ADDITIVE`: previously accepted
  nonstandard/ambiguous inputs can now fail, and tallies no longer follow narrative
  headings. Global revision goldens and task authoring conventions were regenerated;
  persisted schema/body shape is unchanged. Owning help, comments and the
  [planning-data guide](../../docs/architecture/planning-data.md#criterion-section-selection)
  describe the same contract.

Evidence: domain selection/edit and fuzz regressions, portable core read/lint and
fresh-body CAS retry tests, real CLI read/edit/forced-preview/no-op byte-and-mode
checks, and TUI diagnostic/body visibility tests. Compiler-valid overlays restoring
first-wins selection or bypassing structure with force failed the intended tests;
normal focused race runs passed afterward. No source file was mutated by the
overlays. A ten-second fuzz run executed over 424,000 cases without failure.
The original real-binary reproduction now reports 0/1 and refuses completion;
explicit force still permits a preview of known unmet criteria.

`just build`, full `just test` (race), `just lint`, regenerated CLI docs,
planning/link/audit lint and whitespace checks passed. Initial fixture assumptions
about fake-store body reads and asynchronous TUI detail loading were corrected.
The initial repair-preview assertion was corrected against its existing contract,
not by changing that unrelated repair behavior. Generated documentation was
committed before rerunning `docs-check`; final review evidence follows below.

Self-review separately reproduced the preexisting
[unreadable-marker completion gap](6gq03e44pcyz-guard-task-completion-against-unreadable-criterion-markers.md):
canonical `- []` is diagnosed by lint but omitted from the tally/completion gate.
It is not ambiguous section selection, so its policy is tracked separately rather
than silently included here. That follow-up is a member of the authoring Thread
and another prerequisite of criterion UI work; neither parser repair depends on
the other.

## Independent review checkpoint

[Codex](../audits/6gq0xcjm5eqx-2026-10-24-criterion-section-selection-implementation-codex.md)
completed both review passes with no new in-scope findings. Owner confirmation of
the source-owned helper receipt reported `verify=passed`, `transfer=succeeded`,
and `delivery=confirmed` for transferred blob
`f96bd597b56505d1fc8a80000ab0b5f0da0ca668`; the empty audit is now closed.
The reviewer disclosed a pre-sandbox filename inventory (no implementation reads
or writes). Its later implementation inspection and probes used the independent
captured workspace; this procedural deviation does not invalidate the technical
evidence, but is not an unqualified protocol pass.

Independent evidence includes hostile literal CLI fixtures, forced/unforced
completion and previews for active and completed tasks, explicit editor repair,
and actual filesystem CAS conflicts. A compiler-valid stale-callback mutation
failed the intended retry regression; restoration passed. The owner inspected
that probe and mutation evidence and reran focused criterion race tests across
domain/core/CLI/TUI and wire tests successfully. The initial rerun used an
inaccessible default Go cache; rerunning with the existing writable temporary
cache passed. No product changes were indicated by this review. The malformed
marker and tracked-reason tasks remain open exclusions, not repaired claims.

[Antigravity](../audits/6gq0xcjtcej9-2026-10-24-criterion-section-selection-implementation-antigravity.md)
found two valid gaps in the same captured implementation: H1 titles and narrative
ancestors could falsely own checklists (M1), and the domain test oracle checked
counts without exact selected text or diagnostic locations (L1). The helper receipt
confirmed delivery of blob `5195b14391dfeccae08a2a67651a341ffec8e749` before
owner changes. Both findings are fixed and the audit is closed. Its original
verdict and citations describe the reviewed baseline, not the post-review fix.

H1s are now document titles only. Recognized sections at levels 2–6 still include
deeper subsections; an unknown heading directly owns only the checklist before
its next heading, so ordinary child checklists do not implicate narrative
ancestors. Diagnostics identify the actual offending heading and line. This
preserves the existing any-title creation contract rather than restricting names.
New domain regressions failed against the reviewed selector before the fix, then
passed. They assert complete literal criterion values and precise diagnostics;
CLI tests cover default scaffolds titled `Acceptance` / `Acceptance criteria`,
ordinary descendants, refusal without effects and heading-only repair. Fresh
binary smoke checks independently confirmed those paths and that a repaired
unmet criterion still blocks completion. Source comments, guide, help and schema
conventions describe these boundaries; the unshipped revision remains 1.90.

Post-review fuzzing passed over 1.4 million executions in ten seconds. The CLI
goldens were regenerated from the committed 1.89 baseline after updating the
unshipped conventions; no revision guard was weakened. Neither review claims
the malformed-marker or tracked-reason follow-ups are repaired here.

Final gates passed after the review fixes: `just build`, full `just test` (race),
`just lint` (zero issues), `just docs-check`, planning/link lint, both audit lints
and whitespace checks. All four criteria are met; the task is complete. The
Thread remains healthy with no in-flight work. The separate unreadable-marker
task is eligible, while TUI criterion editing still waits for that repair and the
tracked-reason policy task.

## Related

- Independent implementation reviews: [Codex — closed, no new findings](../audits/6gq0xcjm5eqx-2026-10-24-criterion-section-selection-implementation-codex.md); [Antigravity — closed, M1/L1 fixed](../audits/6gq0xcjtcej9-2026-10-24-criterion-section-selection-implementation-antigravity.md). Both used isolated dirty-tree capture, hostile evidence, mutation sensitivity and a separate systemic pass.
- [TUI criterion actions](6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion.md): establish trustworthy criterion enumeration before shipping more editing UI.
- [Tracked reason alignment](6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md): different policy defect; neither repair requires the other.
- [Historical body structure work](6fpfcecdymca-structure-aware-body-mutation-and-metadata-reads.md): documents the broad matcher and its narrower lint guard.
