---
schema: 1
id: 6g3ag8py12y9
status: completed
epic: 20-cli-ux-and-ergonomics
description: Settle how a candidate-task line links to its finding, so audit sync and drift lint become possible
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, audit, core, design]
created: "2026-08-24"
updated_at: "2026-09-19"
depends_on: [6g72wf39pyhb]
started_at: "2026-09-19"
completed_at: "2026-09-19"
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

- [x] The linkage convention is decided and written down — where the code sits, whether one
      line may name several findings, and whether an unlinked line stays legal.
- [x] The status→glyph mapping is decided, defined ONCE in code, and shared with whatever
      already renders finding status rather than transcribed beside it.
- [x] `audit new`'s scaffold emits the decided shape, and its comment is derived from the
      mapping rather than hand-listed.
- [x] A candidate line is written by the tool, not typed — a finding resolved through
      `audit finding` updates its mirror in the same atomic write.
- [x] `audit lint` flags candidate↔status drift, an unresolvable finding reference, and a
      duplicate line for one finding.
- [x] A decision is recorded on the 10 legacy audits, and honoured — migrated, tolerated on
      read, or explicitly left alone.
- [x] Errors wrap the domain sentinels; suite + lint green; README/docs updated.

## Notes

Supersedes criteria 3 and 4 of `audit-finding-write-surface-status-write-and-candidate-list-sync`
(`6feeygw00jmx`), which are deferred pointing at exactly this blocker.

## Decision (2026-09-19)

Candidate lists now have an explicit managed format rather than a heuristic parser:

- `candidate-tasks:v1` marks the section the tool owns. Unversioned sections remain readable legacy prose and are neither linted nor rewritten; there is no bulk migration.
- A canonical row is `- <glyph> <CODE> · <status> — <one-line candidate>`. A row names exactly one finding, no unlinked rows are legal, and the row itself is optional. Several findings may independently name the same eventual task.
- `domain` owns the ordered finding-status/glyph table. The audit scaffold derives its marker legend from it, and themed rendering consumes the same glyph mapping.
- `audit finding --candidate <one-line>` adds or replaces a row; an empty value removes it. A status edit refreshes an existing managed row, and combined status/note/candidate changes use the audit body’s single atomic transform.
- `audit lint` and top-level `lint` validate only managed v1 rows: exact marker, grammar, finding reference, glyph/status agreement, status projection, and one-row-per-finding uniqueness.

The deliberately strict v1 write path prevents silent inference from the ten historical formats while leaving those audits undisturbed.

## Implementation and validation (2026-09-19)

Implemented the v1 parser, writer, status synchronizer, and lint projection across domain, core, filesystem loading, CLI, templates, generated references, architecture guidance, and both scheduled audit routines. Fresh default and security audits emit the derived marker; legacy audits remain unchanged. The golden updater now explicitly honors ADR-0008’s exclusion of Markdown template bodies while retaining the revision gate for ordinary JSON payload changes.

Validation: the race-enabled full Go suite passed; golangci-lint reported 0 issues; generated CLI docs and the security-template golden were refreshed; top-level planning lint, audit lint, and `git diff --check` passed. A disposable planning-space smoke test exercised `audit new` → `audit append` → `audit finding --candidate` → status sync → scoped lint successfully.

That smoke test also made the remaining finding-creation gap concrete: raw append leaves a new finding after the scaffold’s Candidate section. Parsing and candidate synchronization are correct, but authoring is awkward. Follow-up `6gbpe6e8n87k` now owns a canonical finding-creation verb, depends on this task, and is Wave 4 of the same dogfooded Thread.

## Adversarial review closeout (2026-09-19)

Antigravity completed a substantive no-findings review with consumer inventory, disposable-space lifecycle exercises, concurrency checks, boundary probes, and mutations that proved the focused tests fail when status synchronization or duplicate detection is removed. The clean verdict is accepted without waiting on the unavailable second reviewer.

Closeout hardening added permanent CLI coverage for dry-run and CRLF preservation, plus a store/core regression proving that a combined status-and-candidate edit retries around a concurrent audit append without losing prose or duplicating the projection. The full race-enabled suite, golangci-lint, planning lint, audit lint, and git diff check are green.

## Claude review amendments (2026-09-19)

Claude found five valid gaps after the initial clean review. All were fixed in scope: the ADR-0008 golden exception now proves a body-only change against an explicit snapshot allowlist; managed candidate add/remove cycles are whitespace-stable; fenced content inside the section is ignored by the row grammar without disturbing raw spans; token-delimited marker parsing distinguishes unknown versions and damaged managed rows from ordinary legacy prose; and theme tests require intentional colour coverage for every domain finding status.

Focused regressions cover every finding. The full race-enabled Go suite passes, golangci-lint reports 0 issues, and planning lint, audit lint, and git diff checks are clean. Audit 6gbqd98pm8cw is closed with M1, M2, L1, L2, and L3 fixed.
