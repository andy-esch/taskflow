---
schema: 1
id: 6g7s4k845fsb
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: CreateAudit accepts an ID already used by another audit and lint stays green, leaving both records permanently unwritable.
effort: 2-4 hours
tier: 1
priority: high
autonomy_level: 4
tags: [audit, identity, lint, store, robustness]
created: "2026-09-07"
started_at: "2026-09-07"
updated_at: "2026-09-07"
completed_at: "2026-09-07"
---
# Detect duplicate audit IDs at creation and lint time

## Objective

Close the same-kind audit identity hole reproduced by the 2026-09-07 architecture audit. Refuse a
new audit whose minted stable ID already belongs to another audit, and make ordinary repository lint
report duplicate audit IDs with enough evidence to repair the files deliberately.

## Scope

- Add an audit-ID collision guard to the store creation path before any file is written.
- Reuse the existing duplicate-ID domain diagnostic in repository lint for parsed audit records.
- Preserve resilient audit listing while making the identity collision loud through lint.
- Provide deterministic diagnostics that name the duplicate ID and affected audit sources.
- Mutation-test the guard and lint wiring against the exact copied-audit reproduction.

## Acceptance criteria

- [x] Audit creation refuses a minted ID already owned by another audit before writing any new file.
- [x] Ordinary `tskflwctl lint` reports every same-kind duplicate audit ID and identifies all
      conflicting documents deterministically.
- [x] A clean audit corpus remains unchanged and lint-clean; audit listing remains available for
      diagnosis even when duplicates exist.
- [x] Focused tests fail if either the create-time scan or lint wiring is removed and cover a
      collision whose slugs differ.
- [x] The architecture-audit H1 finding is marked fixed with the implementation evidence, and full
      tests, race tests, lint, docs/schema checks, planning lint, and diff hygiene pass.

## Out of scope

- Deciding whether stable IDs must be unique across every entity kind. That planning-space policy is
  tracked separately by the frontmatter-schema ADR task.
- Automatically rewriting either duplicate ID; references make that a deliberate repair.
- Adding an audit rename command or changing stable-ID generation.

## Related

- Audit [2026-09-07 architecture, data model, and storage](../audits/6g7qc8qd00xe-2026-09-07-arch-data-model-and-storage.md)
- Prior research-ID fix [duplicate-stable-ids-brick-research-docs](6g1dnnfgyjap-duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably.md)
- Policy task [ADR — declared frontmatter-schema contract](6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md)
- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)

## Implementation evidence (2026-09-07)

`CreateAudit` now checks the canonical audit ID under the repository write lock before its atomic create, so different-slug collisions cannot pass either the identity check or a concurrent check/create race. Ordinary repository lint applies the shared duplicate-ID diagnostic to canonical audit filename IDs and names every conflicting audit deterministically while keeping audit reads available for repair. Regression coverage includes direct store rejection, two concurrent creators with exactly one winner, core lint wiring, and the original CLI-level copied-audit reproduction.

Validation passed: `go test -race ./...`, `just lint`, `just tidy-check`, `just docs-check`, `tskflwctl lint`, and `git diff --check`. The analogous research check/create serialization gap discovered during review is deliberately tracked by `6g7s6hr3qnfq` rather than widened into this task.

## Adversarial review closeout (2026-09-07)

The external review confirmed the repository lock serializes audit ID detection and creation across independent store values, alternate and symlinked roots, and separate processes. It found one real diagnostic split: duplicate lint used parsed records while mutation used parse-free filename identities, so an unreadable colliding record could hide or undercount the collision.

The correction keeps identity portable and out of public wire formats: resilient read problems now carry optional adapter-recovered stable identity, and core combines readable and unreadable sources without parsing paths. The shared duplicate check now covers audits, research, and Threads, names every available source deterministically, and uses wording valid for any collision count. Focused mixed-readability tests cover all three kinds; the CLI regression covers three audits with one malformed document. All review findings are fixed. The review-protocol mismatch that left a successfully transferred report stamped `transfer=pending` is tracked separately by `6g7srp3py9fe`.
