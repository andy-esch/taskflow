---
schema: 1
id: 6fq9zy126tas
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: A non-Crockford id is misreported as 'no leading id', aborts the whole command (exit 11), and lint --fix can't repair it.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [id, store]
created: "2026-07-18"
---
# Invalid (non-Crockford) entity id is mishandled

## Objective

An entity file whose leading id contains a non-Crockford char (Crockford base32
excludes i/l/o/u) is misdiagnosed and unrepairable. `splitFlatName` calls
`id.Valid` (internal/store/flatname.go:32); a bad char fails it, so the file is
reported as *not id-led* even though a 12-char id is right there. Reproduced on
current code with an `l` in the id.

## Acceptance criteria

- [ ] Error names the offending char / rule (e.g. `id "…" contains non-Crockford char 'l'`) instead of "has no leading id"
- [ ] One bad file no longer forces exit 11 for the whole command — skip-with-warning (or gate the abort behind `--strict`) so unrelated queries succeed with a trustworthy exit code
- [ ] `lint --fix` repairs/re-mints an invalid id, not only a missing one
- [ ] Ids are validated at write time so a bad id can't persist

## Notes

- Today: `audit list` prints good rows then `error: … 1 file(s) with unreadable
  frontmatter` (exit 11); `lint --fix` "fixes 1 file" but leaves the bad-id file untouched.
- Manual fix that worked: swap the illegal char for its Crockford decode-alias
  (l→1, i→1, o→0) in both filename and `id:` — same decoded value, canonical, keeps sort order.
- Loci: internal/store/flatname.go, internal/id/id.go (Valid/decode), scanDir error surfacing.
- Source: https://github.com/andy-esch/taskflow/issues/105 (P2, High)
