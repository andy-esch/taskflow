---
schema: 1
id: 6g1dnnfgyjap
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: core.NewResearch never dedupes a same-day id collision; both colliding docs then become permanently unwritable with a futile retry message and no lint diagnostic.
effort: Unknown
tier: 1
priority: high
autonomy_level: 3
tags: [core, domain]
created: "2026-08-18"
updated_at: "2026-08-19"
completed_at: "2026-08-18"
---

# Duplicate stable ids brick research docs, silently and unrecoverably

## The defect

`id.NewAt` is stateless and its own doc is explicit: *"Callers minting several ids at the
same millisecond … must dedupe: regenerate on a clash."* `internal/tools/researchmigrate`
does exactly that (`mintUnique`). **`core.NewResearch` does not** — and because it keys on
a DAY (`service_research.go:69`, `s.newIDAt(day.UnixMilli())`), every doc sharing a
`created` date draws from the same 2^17 random slot.

Nothing downstream catches it:

- `store/create.go` claims "the id makes the flat filename unique, so writeNewFile's
  O_EXCL is the whole collision guard" — **false** for an id collision on a *different*
  slug: different path, so O_EXCL never fires.
- `domain.LintResearch` has no duplicate-id check, and the old duplicate lint was retired
  on the premise "id-led filenames are unique by construction" — true of filenames, not ids.

## Reproduced end to end

Two docs sharing an id, different slugs:

    research list          -> both listed, no complaint
    research show <id>     -> ambiguous match (exit 13)
    research set alpha     -> "changed on disk during update; retry: conflict" (exit 14)
    lint                   -> ✔ all active tasks and epics pass lint

**Both docs become permanently unwritable.** The message says retry; retrying can never
work, because the CAS re-resolve returns `ErrAmbiguous` and `verifyUnchanged` collapses
that into a conflict (see the sibling epic-21 task). No diagnostic anywhere. Recovery
needs a hand `mv`, and research has no `rename`.

Probability: ~1/131072 per same-day pair; ~0.9% across a scripted 50-doc same-date
import; ~0.007% for one doc backdated onto the corpus's 9-doc 2026-01-03 cluster. Low
odds, silent onset, unrecoverable.

## Acceptance criteria

- [x] `NewResearch` regenerates on a clash against ids ALREADY ON DISK, not just within
      one process — the on-disk set is the collision domain, since minting is per-command.
- [x] A duplicate id is a lint finding (it is currently invisible to `lint`), so an id
      that slips in by any route is diagnosable rather than silent.
- [x] Test: two docs forced onto the same id (injected id generator) — creation is
      rejected or deduped, and if one exists on disk lint reports it.
- [x] Consider whether `CreateResearch` should reject a colliding id at the store layer
      too, as the last line of defence for a direct store caller.

## Out of scope

- Widening the id (more random bits) — the bug is the missing dedupe, not the width.
- `verifyUnchanged`'s error collapsing, and create-path `id.Valid` checking — both shared
  across entities, filed on epic 21.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Found by an independent adversarial correctness review, 2026-08-18
