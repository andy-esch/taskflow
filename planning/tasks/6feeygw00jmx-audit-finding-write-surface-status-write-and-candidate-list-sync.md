---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: 'audit finding --status write + audit sync + candidate drift lint — items 3+5 carved from the finding-level read task (grammar transcribed in-repo)'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, audit, core]
created: "2026-06-21"
updated_at: "2026-08-24"
id: 6feeygw00jmx
started_at: "2026-08-24"
completed_at: "2026-08-24"
---

# Audit finding write surface — status write + candidate-list sync

> Carved out of [audit-finding-level-operations-query-write-lint-sync](6fd5r5c03v5y-audit-finding-level-operations-query-write-lint-sync.md) (its
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
   [interactive-prompt-layer-gh-style-pickers](6fbj870019vt-interactive-prompt-layer-gh-style-pickers.md); here it's the non-TTY append.
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

- [x] `domain.Finding` (or a sibling) carries the Status-line span for in-place rewrite.
- [x] `audit finding <slug> <code> --status <v> [--pr N]` rewrites one Status line in
      the HOWTO format; body otherwise byte-identical; `--status fixed` appends the
      resolution block; `--dry-run` + exit codes.
- [ ] `audit sync <slug>` re-derives the candidate-list symbols from Status lines. · **tracked:** carried by 6g3ag8py12y9 — the linkage convention is a design call, not a coding gap
- [ ] `audit lint` flags candidate↔status drift. · **tracked:** carried by 6g3ag8py12y9 — drift lint falls out of the convention decided there
- [x] Errors wrap the domain sentinels; suite + lint green; README/docs updated.

## Related

- Source: [audit-finding-level-operations-query-write-lint-sync](6fd5r5c03v5y-audit-finding-level-operations-query-write-lint-sync.md) (items 3+5).
- Format source: desirelines `audits/HOWTO-execute.md` (transcribed into the grammar
  section above on 2026-06-21; reachable in-workspace — no longer a blocker).
- Interactive resolution-block prompt: [interactive-prompt-layer-gh-style-pickers](6fbj870019vt-interactive-prompt-layer-gh-style-pickers.md).
- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md).

## Progress notes (2026-08-24)

`audit finding <audit> <code> --status <v>` exists and is in use — this audit's own H1 was
resolved with it rather than the search-and-replace every previous status change used.

**Surgical by construction, not by care.** `ParseFindings` now records each finding's
status span, and the write replaces exactly that range, so prose elsewhere containing the
same word and every other finding are untouched without the writer taking any precautions.

**The span covers the whole VALUE, not just the token.** The line formats carry decoration
(`fixed 2026-08-24 (PR #12)`, `deferred (see ADR-0003)`, `superseded by <link>`), and a
token-only span would have left the old decoration stranded after a new value. Caught by
probing a re-stamp rather than by reading the code.

**Only the token is normalised.** A first cut lowercased the whole value and turned
`(see ADR-0003)` into `(see adr-0003)` — decoration holds dates, links, and document names
whose spelling is not the tool's to flatten. The token is still lowercased, so it parses
back as vocabulary.

**No port change was needed.** The write goes through `EditAudit`, the existing
parse-before-accept path. That path retries its callback on a rejected edit, so the callback
refuses the second attempt: this rewrite is deterministic, and a retry would be rejected
identically rather than succeeding. `--dry-run` validates and computes without ever entering
the write path.

`--status` is required and rejects an unknown value naming the legal set; a missing code is
`ErrNotFound`; re-stamping the same value reports "already" and writes nothing.

### What is deferred, and why

Criteria 3 and 4 are blocked on a convention that does not exist yet. `sync` has to map a
candidate-list line back to its finding, and the corpus spells that link two ways:

```
- ✅ S1 — dead `Style.Enabled()` removed.                      ← code leads
- ⏳ `tskflwctl task new "…" --epic …` — H1; mirrors `task set`  ← code in trailing prose
```

Deriving symbols from statuses needs one of these to be authoritative, or a third explicit
form. That is a decision about the audit format, not an implementation detail, so it is
recorded rather than guessed — and criterion 4 falls out of it, since drift cannot be
detected before the linkage is defined.

**Handed off 2026-08-24 to `6g3ag8py12y9`** (*Decide the candidate-list convention and make
the tool own it*). Both criteria are now `tracked` rather than `deferred`: the remaining
work is a format decision with its own migration question, not a coding gap in this task.
That task also records a third spelling this note missed — codes in parens at the end of
the line, naming several findings at once — and a second blocker beyond linkage: the
scaffold offers four glyphs against seven statuses, so even a solved mapping would collapse
`deferred`, `superseded`, and `wontfix` onto one symbol.

Criterion 2's substance (in-place status write, byte-identical body, `--dry-run`, exit
codes) is done; the `--pr` sugar and the resolution block remain, and both want a `--note`
flag to carry their content.
