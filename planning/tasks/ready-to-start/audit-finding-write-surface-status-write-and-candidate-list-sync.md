---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'audit finding --status write + audit sync + candidate drift lint — items 3+5 carved from the finding-level read task (blocked on external HOWTO)'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, audit, core]
created: "2026-06-21"
---

# Audit finding write surface — status write + candidate-list sync

> Carved out of [[audit-finding-level-operations-query-write-lint-sync]] (its
> items 3 + 5) on 2026-06-21. The finding **read** surface — `ParseFindings`,
> `audit findings` (query), `audit lint` — shipped and closed out epic 17. This is
> the finding **write** surface, a *feature*, **not** part of retiring Python `pm`,
> so it lives outside the (now completed) port epic.

## Why 3 and 5 are one task (and 4 was not)

`audit finding --status` (item 3) and `audit sync` (item 5) are **both surgical
writes** that share machinery the current read-only parser lacks:
- a **body-offset extension** to `domain.Finding` (today it's read-shaped — no
  span info) so a `**Status:**` line can be rewritten in place;
- a **candidate-list parser** (the bottom-of-audit `✅⚠️⏳⛔` list), which `sync`
  rewrites and the drift check reads.

`audit lint` (item 4) was read-only validation over the existing parser, so it
shipped independently. Item 5's *drift-detection* half (a `fixed` finding still
marked `⏳`) is conceptually a lint check, but it needs the candidate-list parser
`sync` introduces — so it rides here, not in `audit lint`.

## Scope

1. **`audit finding <slug> <code> --status <value> [--pr N] [--note]`** — surgically
   rewrite one finding's `**Status:**` line, byte-preserving the rest of the body
   (atomic write, `--dry-run`, exit codes), stamping the cheat-sheet format. On
   `--status fixed`, append the resolution block. The interactive prompt face is
   [[interactive-prompt-layer-gh-style-pickers]]; here it's the non-TTY append.
2. **`audit sync <slug>`** — re-derive the candidate-list `✅⚠️⏳⛔` symbols from the
   finding `**Status:**` lines (atomic, `--dry-run`).
3. **Candidate-list drift check** — folded into `audit lint`: a finding whose status
   and candidate mark disagree is a warning (needs the candidate-list parser).

## ⚠️ Blocked — needs the external grammar

Item 3's exact stamp format (`in-progress (since YYYY-MM-DD)`, `fixed YYYY-MM-DD
(PR #N)`, and the formats for `landed`/`deferred`/`superseded`/`wontfix`) plus the
resolution-block shape are **fixed by `desirelines-planning/audits/HOWTO-execute.md`,
which is not in this repo**. The status *vocabulary* is known
(`domain.FindingStatuses`), but the per-status line format and the resolution block
must be sourced from HOWTO before item 1 can be built. The candidate
`✅ done · ⚠️ partial · ⏳ open · ⛔ won't do` mapping IS in-repo
(`core.auditBodyTemplate`), so the sync rewrite is unblocked once the body-offset
extension lands.

## Acceptance criteria

- [ ] `domain.Finding` (or a sibling) carries the Status-line span for in-place rewrite.
- [ ] `audit finding <slug> <code> --status <v> [--pr N]` rewrites one Status line in
      the HOWTO format; body otherwise byte-identical; `--status fixed` appends the
      resolution block; `--dry-run` + exit codes.
- [ ] `audit sync <slug>` re-derives the candidate-list symbols from Status lines.
- [ ] `audit lint` flags candidate↔status drift.
- [ ] Errors wrap the domain sentinels; suite + lint green; README/docs updated.

## Related

- Source: [[audit-finding-level-operations-query-write-lint-sync]] (items 3+5).
- Format contract: `desirelines-planning/audits/HOWTO-execute.md` (external — blocks item 1).
- Interactive resolution-block prompt: [[interactive-prompt-layer-gh-style-pickers]].
- Epic [[20-cli-ux-and-ergonomics]].
