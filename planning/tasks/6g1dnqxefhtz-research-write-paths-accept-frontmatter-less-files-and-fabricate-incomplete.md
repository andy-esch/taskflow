---
schema: 1
id: 6g1dnqxefhtz
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: set/append succeed on a doc the read path rejects, writing a block with no id and no created; also LintResearch validates created with ValidateDate not ValidateMintableDate.
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [store, domain]
created: "2026-08-18"
updated_at: "2026-08-19"
completed_at: "2026-08-18"
---

# The research write paths accept a file the read path rejects, and fabricate incomplete frontmatter

## The defect

`splitFrontmatterStrict` returns `(nil, wholeFile, nil)` for a doc with no `---` block, and
`documentMapping` then CREATES an empty mapping — so a write succeeds where the read
refuses. The parse-before-commit blame branch (`researchstore.go:130-135`) is dead here,
because the fabricated content parses fine.

Reproduced on an id-led doc with no frontmatter — **exactly the shape 10 legacy docs had**:

    $ research show legacy
    error: malformed frontmatter: missing frontmatter — a research doc must open with a `---` YAML block

    $ research append legacy --body "## Added"
    ✔ appended to legacy

    $ head -3 research/6dr29v000zzr-legacy.md
    ---
    updated_at: "2026-08-18"
    ---

The doc now claims **no `id` and no `created`** — and `created` is the chronology anchor
the id is minted from. Its prose `**Created**: 2026-01-03` is left stranded in the body.
`research set --description` does the same, producing a block whose only key is
`description`.

Net effect: the file degrades from "loudly unreadable" (a FileProblem naming the fix) to
"parses but violates the contract" — quieter and worse.

## Also: LintResearch validates `created` with the weaker rule

`domain/lint.go` uses `ValidateDate`, while `core.NewResearch` and `researchmigrate` both
use `ValidateMintableDate`. `research edit` is whole-file editing and is therefore the one
path that CAN change `created` — so a hand-edited `created: 1026-06-15` passes lint
silently even though the doc's id cannot encode that date. Bundled here because it is the
same guarantee: a research doc always carries an id and a representable `created`.

## Acceptance criteria

- [x] `SetResearchFields` and `AppendResearchBody` REFUSE a doc with no frontmatter block,
      with the same shape-naming error the read path gives — a write must not be more
      permissive than a read.
- [x] Refusal is `ErrValidation` (exit 11) and writes nothing.
- [x] `LintResearch` uses `ValidateMintableDate` for `created`, so an out-of-range date
      introduced by `research edit` is flagged.
- [x] Tests for both, including that the file is byte-unchanged after a refusal.

## Note on scope

The permissive-write behaviour is shared with `task`'s `EditBody`, but research is the
entity with a documented frontmatter-less legacy corpus, so it is materially more likely
here. Fix research; if the shared helper is the right place, note it for the siblings.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Found by an independent adversarial correctness review, 2026-08-18
