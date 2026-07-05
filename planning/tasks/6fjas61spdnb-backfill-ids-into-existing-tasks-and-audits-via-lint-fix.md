---
schema: 1
id: 6fjas61spdnb
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: lint flags missing id (fail-open) + lint --fix mints id.NewAt from created (tasks)/date (audits), deduped; extend fix walk to audits. Carved from the migration script. Per ADR-0003.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, storage]
created: "2026-07-03"
updated_at: "2026-07-05"
started_at: "2026-07-03"
completed_at: "2026-07-03"
---
# Backfill ids into existing tasks and audits via lint --fix

## Objective

Complete the id foundation: new entities get ids on create, but everything already
in planning/ (and desirelines) had none. Rather than a throwaway script, fold the
backfill into `lint --fix` — additive, idempotent, reusable across every planning
repo — leaving only the destructive rename/status-move/ref-rewrite in the one-time
migration script.

## Acceptance criteria

- [x] `lint` flags a missing id (fail-open, all statuses incl. archived) — domain
  `MissingIDIssue`, wired into `LintTask`, the archived branch, and `LintAudits`.
- [x] `lint --fix` mints an id via `id.NewAt`, timestamped from the entity's own
  date (created → audit slug date → updated_at → lifecycle stamps), deduped within
  the run against existing + assigned ids (`mintUniqueID`); no-date/unparseable →
  skipped, and the re-lint re-flags it.
- [x] Fix walk extended to the audit buckets; epics stay text-only (keep NN-slug).
- [x] Tests: store backfill (task/audit/lifecycle-stamp/skip-present/skip-no-date/
  dedup), `mintUniqueID` collision-retry + give-up, `firstDateMillis` preference;
  core lint fixtures; a CLI flag→fix→clean end-to-end.
- [x] Executed on this repo: 190/190 tasks + 5/5 audits backfilled; planning lint
  clean.

## Out of scope

- The destructive cutover (id-led renames, status/bucket → frontmatter, ref
  rewrites) — stays in [[one-time-migration-script-this-repo-desirelines]].
- The pre-existing audit finding-status typo ("partially") in the
  codebase-quality-architecture audit — a separate data fix.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]
- ADR [[0003-stable-key-id-addressed-storage]] section 6 + the backfill-timestamp amendment

## Superseded by filename-backfill (2026-07-05, Phase B)

Post-flatten every entity file is id-led, so `lint --fix` now backfills a missing
frontmatter `id:` **from the filename** (the canonical key resolveID/CAS match on)
instead of minting one from a date. The whole date-mint apparatus this task added
(`firstDateMillis`/`mintUniqueID`/`dateFromFilename`/`knownIDs` + `UnrepairedIDMessage`
+ its restatement) was deleted (~125 LOC), and a new drift lint (`IDDriftIssue`) guards
filename-vs-frontmatter id agreement.
