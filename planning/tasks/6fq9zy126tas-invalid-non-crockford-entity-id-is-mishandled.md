---
schema: 1
id: 6fq9zy126tas
status: next-up
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: A non-Crockford id is misreported as 'no leading id', aborts the whole command (exit 11), and lint --fix can't repair it.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [id, store]
created: "2026-07-18"
updated_at: "2026-08-23"
---
# Invalid (non-Crockford) entity id is mishandled

## Objective

An entity file whose leading id contains a non-Crockford char (Crockford base32
excludes i/l/o/u) is misdiagnosed and unrepairable. `splitFlatName` calls
`id.Valid` (internal/store/flatname.go:32); a bad char fails it, so the file is
reported as *not id-led* even though a 12-char id is right there. Reproduced on
current code with an `l` in the id.

## Acceptance criteria

- [x] Error names the offending char / rule (e.g. `id "…" contains non-Crockford char 'l'`) instead of "has no leading id"
- [x] One bad file no longer *hides the rest*: the listing completes best-effort, and the
      trailing error NAMES the offending files rather than only counting them.
      **Amended 2026-08-23:** the exit code deliberately stays 11. The original wording
      ("no longer forces exit 11") was written before the partial-failure convention
      settled; `status --all` already renders everything and then exits non-zero, and a
      caller that received an incomplete result has to be able to tell. Exit codes are a
      documented contract (10/11/13/14) that agents branch on, so the fix is a better
      message, not a quieter exit. `--strict` was considered and rejected: it would add a
      global flag and still change the default for existing callers.
- [x] `lint --fix` repairs/re-mints an invalid id, not only a missing one
- [ ] Ids are validated at write time so a bad id can't persist

## Notes

- Today: `audit list` prints good rows then `error: … 1 file(s) with unreadable
  frontmatter` (exit 11); `lint --fix` "fixes 1 file" but leaves the bad-id file untouched.
- Manual fix that worked: swap the illegal char for its Crockford decode-alias
  (l→1, i→1, o→0) in both filename and `id:` — same decoded value, canonical, keeps sort order.
- Loci: internal/store/flatname.go, internal/id/id.go (Valid/decode), scanDir error surfacing.
- Source: https://github.com/andy-esch/taskflow/issues/105 (P2, High)

## Corroboration (2026-08-19)

Hit first-hand while stress-testing the research entity. Test fixtures were written with
ids containing `l`, `o`, and `u` — all excluded by Crockford base32 — and every one came
back as:

    not an entity file: "6dr29v000crl-crlf.md" has no leading id — move it to meta/ or delete it

A 12-char id is right there in the filename; the message says there isn't one. It cost a
full round of debugging before the real cause (the excluded characters) was obvious, which
is direct evidence for the "names the offending char / rule" criterion above — the current
wording actively misdirects.

Also folded in here from an epic-21 task (see
[verifyunchanged-collapses-every-resolve-error-into-a-futile-retry-conflict](6g1dpb5gwtt9-verifyunchanged-collapses-every-resolve-error-into-a-futile-retry-conflict.md)), which had
independently picked up the same write-time-validation fix: **no create path validates
`id.Valid` before writing.** `CreateResearch(Research{Slug:"bogus", ID:"abcde"})` returns
nil and writes `research/abcde-bogus.md`, which then reads back as `not an entity file`
and cannot be resolved. Identical in `CreateTask`/`CreateAudit`, reachable via `WithIDGen`
or any direct store caller. Every MUTATE path parses before committing; no CREATE path
does. That is this task's last acceptance criterion, so it belongs here.
