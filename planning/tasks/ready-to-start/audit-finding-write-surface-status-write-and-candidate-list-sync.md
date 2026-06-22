---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'audit finding --status write + audit sync + candidate drift lint — items 3+5 carved from the finding-level read task (grammar transcribed in-repo)'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, audit, core]
created: "2026-06-21"
updated_at: "2026-06-22"
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

## Authoring grammar (transcribed in-repo 2026-06-21 — no longer blocked)

The earlier "blocked on an external HOWTO" note was wrong: that file
(`audits/HOWTO-execute.md` in the sibling `desirelines-planning` repo) is reachable
from this workspace, and its grammar is transcribed here so the task is
**self-contained** — it no longer depends on an out-of-repo file.

**`**Status:**` line values** (item 1 stamps these in place; vocabulary is already
`domain.FindingStatuses`, this pins the per-status *line format*):

| Value | Stamp format |
|---|---|
| open | `open` (default, written with the audit) |
| in-progress | `in-progress (since YYYY-MM-DD)` |
| fixed | `fixed YYYY-MM-DD (PR #N)` + a 1–3 line resolution block underneath (what landed + where the tests live) |
| deferred | `deferred (reason)` — cite the deciding doc (epic, ADR, task, thread) |
| superseded | `superseded by <link>` |
| wontfix | `wontfix (reason)` |

**Candidate-list symbols** (item 2 `sync` derives these from the Status lines):
`✅ fixed · ⚠️ partial-with-follow-up · ⏳ still-open · ⛔ deferred or wontfix`. The
in-repo `core.auditBodyTemplate` mapping (`✅ done · ⚠️ partial · ⏳ open · ⛔ won't
do`) already encodes the same axis, so the `sync` rewrite is unblocked once the
body-offset extension lands.

**Open design call (not a blocker):** whether the generic tool adopts this
desirelines house format verbatim or generalizes it (e.g. drops the
desirelines-specific "merged to `main`" gloss). Decide during implementation.

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
- Format source: desirelines `audits/HOWTO-execute.md` (transcribed into the grammar
  section above on 2026-06-21; reachable in-workspace — no longer a blocker).
- Interactive resolution-block prompt: [[interactive-prompt-layer-gh-style-pickers]].
- Epic [[20-cli-ux-and-ergonomics]].
