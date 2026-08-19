---
schema: 1
id: 6g1c0ssbbxam
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: research new accepts dates outside id.NewAt's 43-bit window (1970..2248), silently inverting the id-order-equals-date-order invariant; a year typo mis-orders a doc forever.
effort: Unknown
tier: 1
priority: high
autonomy_level: 3
tags: [core, domain]
created: "2026-08-18"
updated_at: "2026-08-19"
started_at: "2026-08-18"
completed_at: "2026-08-18"
---

# Guard the representable date range when minting a backdated id

## Objective

`research new --created <date>` mints the doc's id FROM that date so lexical id order
is authorship order (ADR-0003 §3). That invariant silently inverts for dates outside
the representable window, with no error and no warning.

## The defect

`id.NewAt` packs 43 bits of `UnixMilli`, so ids only order correctly for
**1970-01-01 → 2248-09-08** (bisected). Outside it:

- **before 1970** — a negative `UnixMilli` is cast to `uint64` and masked to 43 bits,
  producing an enormous value, so the doc sorts *last*.
- **after ~2248-09-08** — the value overflows 43 bits and wraps to near-zero, so the
  doc sorts *first*.

`domain.ValidateDate` accepts any well-formed `YYYY-MM-DD`, so nothing rejects it:

    id            created     slug
    5wacqvr03gnh  2300-01-01  far-future    <- sorts FIRST
    6g14aqm01z1g  2026-08-18  today
    r02s5kc02ky7  1900-05-05  year-1900
    zwagkpm03t7f  1969-01-01  pre-epoch     <- sorts LAST

The realistic trigger is a **year typo** — `--created 1026-06-15` or `9026-…` is
accepted and the doc is mis-ordered permanently. `id.NewAt`'s own doc comment notes
"room through ~year 2248"; it is simply never enforced.

## Acceptance criteria

- [x] A date outside the representable window is rejected with `ErrValidation` (exit 11)
      naming the window — at the same seam that already validates the date, so every
      caller inherits it rather than each minting site remembering.
- [x] The boundaries are pinned by test: 1970-01-01 and the last representable day are
      accepted; the days either side are rejected.
- [x] `internal/tools/researchmigrate` gets the same guard (it mints from arbitrary
      recovered dates, including a prose `**Created**:` that could be mistyped).
- [x] The window is documented where an author will see it — the `created` field doc in
      the research Descriptor, so `schema research` states it.
- [x] A property/fuzz-style test asserts that for in-range dates, sorting by id equals
      sorting by created.

## Out of scope

- Widening the id layout (more time bits) — 2248 is not a real constraint; the bug is
  the missing guard, not the range.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md) §3 (id-from-date policy)
