---
schema: 1
id: 6g3ag8py12y9
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Settle how a candidate-task line links to its finding, so audit sync and drift lint become possible
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, audit, core, design]
created: "2026-08-24"
updated_at: "2026-08-24"
---
## Context

Every audit ends with a `## Candidate tasks` list that mirrors the findings above it,
under a scaffold comment that names four glyphs:

```
<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->
```

"Mirror" is doing a lot of work there. Nothing enforces it, nothing derives it, and the
two halves of the document drift the moment a finding is re-stamped — which is now cheap
and frequent, because `audit finding --status` exists. `audit sync` was meant to re-derive
those glyphs from finding status, and `audit lint` to flag the drift. Neither can be
written, for two independent reasons.

## Reason 1: there is no finding↔line linkage

The corpus spells the link three different ways, sometimes within one file:

| Spelling | Example |
| :-- | :-- |
| trailing prose, after the first em-dash | `` - ⏳ `tskflwctl task new "…" --tags lint` — M1; XS, `FindingStatuses()` is already there. `` |
| leading, right after the glyph | `- ⛔ L1 — no standalone task; falls out of H1.` |
| in parens, at the end | `- ✅ ~~`tskflwctl task new "…"`~~ — fixed directly (H1, H2); regression tests added.` |

Note the third row maps ONE line to TWO findings. The inverse also occurs: findings with
no candidate line, and candidate lines that describe follow-up work belonging to no
finding. Any parser guessing at this would be wrong quietly, which is worse than absent.

10 of the 12 audits in `planning/audits/` carry such a list.

## Reason 2: four glyphs, seven statuses

Even with linkage solved, the mapping is not total:

- statuses: `open · in-progress · fixed · tracked · deferred · superseded · wontfix`
- glyphs: `✅ done · ⚠️ partial · ⏳ open · ⛔ won't do`

`in-progress` has no glyph (⚠️ "partial" is close but means something else). `tracked` is
neither done-here nor open. `deferred`, `superseded`, and `wontfix` all collapse onto ⛔,
which erases the distinction the vocabulary work was about. A sync that flattens three
statuses into one symbol is a lossy mirror, and readers reason from the mirror — the exact
failure named in M3 of `2026-08-17-finding-status-surface`.

## Direction to consider (not decided)

Stop parsing the convention; start emitting it. The tool writes the candidate line with the
finding code in a fixed slot and derives the glyph from a status→glyph table it owns —
`theme.FindingStatus` already holds most of that mapping for the TUI and would be the
natural single source.

That is a bigger call than it looks, which is why this is its own task rather than folded
into `audit-finding-write-surface-status-write-and-candidate-list-sync`: it changes the
`audit new` scaffold every future audit inherits, and leaves 10 existing audits
non-conforming. Whatever is decided should say what happens to them — migrate via
`lint --fix`, leave them as legacy, or accept both spellings on read and emit only the new
one on write.

## Open questions

1. **Slot or free text?** A fixed leading slot (`- ⏳ H1 — …`) is greppable and trivially
   parsed, but it is also the spelling the corpus uses LEAST.
2. **One line, many findings** — allowed (`H1, H2`) or forbidden? Forbidding it is simpler
   to sync and truer to "mirror each finding"; allowing it matches how the work actually
   gets grouped into tasks.
3. **Lines with no finding.** A candidate list also carries follow-up work that no finding
   raised. Does that stay legal, and does lint distinguish it from a broken link?
4. **The glyph table.** Does `tracked` get its own symbol (the finding glyph is `→`)? Do
   `deferred`/`superseded`/`wontfix` stay collapsed, or does the mirror widen to match the
   vocabulary?
5. **Migration.** What happens to the 10 existing audits, and is `lint --fix` the vehicle?

## Acceptance criteria

- [ ] The linkage convention is decided and written down — where the code sits, whether one
      line may name several findings, and whether an unlinked line stays legal.
- [ ] The status→glyph mapping is decided, defined ONCE in code, and shared with whatever
      already renders finding status rather than transcribed beside it.
- [ ] `audit new`'s scaffold emits the decided shape, and its comment is derived from the
      mapping rather than hand-listed.
- [ ] A candidate line is written by the tool, not typed — a finding resolved through
      `audit finding` updates its mirror in the same atomic write.
- [ ] `audit lint` flags candidate↔status drift, an unresolvable finding reference, and a
      duplicate line for one finding.
- [ ] A decision is recorded on the 10 legacy audits, and honoured — migrated, tolerated on
      read, or explicitly left alone.
- [ ] Errors wrap the domain sentinels; suite + lint green; README/docs updated.

## Notes

Supersedes criteria 3 and 4 of `audit-finding-write-surface-status-write-and-candidate-list-sync`
(`6feeygw00jmx`), which are deferred pointing at exactly this blocker.
