---
schema: 1
id: 6g1c0zpgeejk
status: ready-to-start
epic: 28-first-class-entities-new-planning-nouns
description: research/ was never added to FixFrontmatter or sweepStaleTemps, so lint promises an id repair --fix won't do and temp orphans linger; plus 'matches 3 researchs'.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, core]
created: "2026-08-18"
---

# Extend lint --fix and the temp sweep to research, and fix the mass-noun plural

## Objective

Three small defects that share one root cause: `research/` was added as an entity dir but
never added to the repair/housekeeping surfaces that iterate the entity dirs by hand.
Bundled because they are one change locus (`store/fix.go`, plus a one-line message) and
one test setup.

## 1. `lint` promises a repair `--fix` will not perform

    $ tskflwctl lint
    no-id
      id: missing stable id — `lint --fix` assigns one
    $ tskflwctl lint --fix
    nothing to fix

    could not auto-repair:
    no-id
      id: missing stable id — `lint --fix` assigns one

`MissingIDMessage` is shared with tasks/audits, where `--fix` *does* backfill the id from
the filename. But `FixFrontmatter` only walks `tasksDir`/`epicsDir`/`auditsDir`
(`store/fix.go:82-90`), so for research the message names a command that cannot help.

Not corruption — `--fix` honestly reports "could not auto-repair" rather than claiming
success — but it sends the user in a circle. Research is id-led, so the same
backfill-from-filename repair applies and should just work.

## 2. Crash-orphaned temp files in `research/` are never swept

`sweepStaleTemps` runs over `tasksDir`/`epicsDir`/`auditsDir` only, but
`createFileAtomic` stages its temp in the *target* dir — so a crashed `research new`
leaves a `.tskflwctl-*.tmp` in `research/` forever. Confirmed: an aged orphan survives
`lint --fix`. Invisible (the `.md` scan filter hides it), so it is pure litter, but it is
litter the tool promises to clean.

## 3. `matches 3 researchs`

    error: "same-title" matches 3 researchs: same-title (6dr29v0005cd), ...

`resolve.go:190` builds the plural as `%ss` from the kind name. Correct for
task/epic/audit, wrong for a mass noun. Fix by giving the kind a plural rather than
suffixing — the registry is the natural home, and projects/routines land next.

## Acceptance criteria

- [ ] `FixFrontmatter` covers `research/`: text normalization + id backfill, so the
      `lint --fix` the message promises actually repairs a missing research id.
- [ ] `sweepStaleTemps` covers `research/`; an aged orphan there is removed and reported.
- [ ] Ambiguous-match wording is correct for research, from a per-kind plural rather than
      `%ss`.
- [ ] Tests for all three (the fixer ones as store tests, the wording as a CLI assertion).

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Sibling: [fold-audits-into-the-top-level-lint-command](6fm8p1cj11qf-fold-audits-into-the-top-level-lint-command.md) — same genre of
  hand-maintained entity-dir list drifting behind the entity set
