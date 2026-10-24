---
schema: 1
id: 6gpxh6ncaqpw
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Prevent narrative headings from hiding real task criteria in reads, writes, lint and completion guards.
effort: half day
tier: 2
priority: high
autonomy_level: 2
tags: [criteria, validation, lint, tui]
created: "2026-10-24"
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

`scanAcceptanceCheckboxes` in [body.go](../../internal/domain/body.go) selects the
first heading containing "acceptance", while lint's `isAcceptanceSectionHeading`
uses a narrower predicate. Canonical criteria after that narrative section are
not scanned. The integration document uses a neutral heading as a workaround;
the underlying parser is unchanged by that closeout.

## Selection contract checkpoint

The broad matcher is documented legacy behavior, not an incidental spelling
mistake. Decide canonical-heading precedence and treatment of legacy aliases and
multiple genuine sections before implementation. Prefer a canonical section over
incidental prose; diagnose ambiguous source structure rather than silently treating
an unread section as zero criteria. Review compatibility under
[ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md),
including changed tally/exit behavior. Do not silently migrate documents or alter
unrelated uses of the general-purpose `Section` helper.

## Acceptance criteria

- [ ] An explicit section-selection and legacy/ambiguity policy aligns criterion reads, mutations, lint and completion eligibility.
- [ ] Narrative headings before and after canonical criteria cannot hide unmet criteria; real CLI tests pin tallies, refusal before effects and completion without force.
- [ ] Regressions cover fenced examples, nested headings, CRLF, accepted legacy headings and multiple genuine sections without mutating unrelated Markdown.
- [ ] Owning comments/help and any required machine-contract revision describe the actual policy; focused tests, normal gates and planning checks pass.

## Out of scope

TUI criterion-editor design, tracked-reason policy, a general Markdown parser
replacement, and bulk document rewrites.

## Related

- [TUI criterion actions](6g3cp32y9149-let-the-tui-flip-an-acceptance-criterion.md): establish trustworthy criterion enumeration before shipping more editing UI.
- [Tracked reason alignment](6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md): different policy defect; neither repair requires the other.
- [Historical body structure work](6fpfcecdymca-structure-aware-body-mutation-and-metadata-reads.md): documents the broad matcher and its narrower lint guard.
