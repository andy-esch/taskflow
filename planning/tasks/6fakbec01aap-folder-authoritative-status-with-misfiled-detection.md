---
status: completed
epic: 17-pm-go-cli
description: Directory is the source of truth for status; flag a recognized status that disagrees with its folder via warning glyph and lint, realign with lint fix
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [pm-tooling, go, data-integrity]
created: "2026-06-09"
completed_at: "2026-06-09"
updated_at: "2026-06-09"
id: 6fakbec01aap
---

# Folder-authoritative status with misfiled detection

## Objective

A `completed/` task was listed as active `ready-to-start` because `parseTask`
trusted a *valid* frontmatter status over the directory — contradicting the
"status == directory" invariant. Make the directory authoritative and surface
the drift (it shows up when reading legacy/pm repos cross-repo).

## Done

- **Directory authoritative** (`store/fsstore.go parseTask`): `Status` is always
  the folder; the frontmatter value is kept as `domain.Task.Declared`.
- **`Task.Misfiled()`** = `Declared.Valid() && Declared != Status` — only a
  *recognized* status in the wrong folder counts; legacy/foreign words
  (`superseded`/`done`/`blocked`) are tolerated (folder governs, no false alarm).
- **`⚠` marker** in `task list` (row + footer count) and `task show`.
- **Lint** flags misfiled tasks — including *archived* ones (a completed/ file
  with a stale active label), via `domain.MisfiledIssues` run on inactive tasks
  too (active tasks still get full field lint).
- **`lint --fix` realigns** the frontmatter to the folder (`store/fix.go`
  `realignStatus`) — safe, since the folder already governs behavior; foreign
  words and unparseable files are left alone.

## Acceptance

- [x] Active list excludes a misfiled completed/ file; `--all` shows it
      `⚠ completed`; lint flags it; `--fix` realigns. Tests at domain/store/cli.
- [x] **Real data:** against `desirelines-planning` (read-only), the rule flags
      11 genuinely-drifted tasks and tolerates the 5 legacy foreign-vocab files.

## Out of scope

- Cleaning `desirelines-planning`'s data — separate pm-managed repo, the user's
  call (`tskflwctl lint --fix -C ../desirelines-planning`, or pm).
- Realigning on a `move` no-op (use `lint --fix`).

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); builds on the M2 crash-safe `Move` and the color work
  [cli-color-glyphs-table-headers-render-styling](6fakbec017kj-cli-color-glyphs-table-headers-render-styling.md).
