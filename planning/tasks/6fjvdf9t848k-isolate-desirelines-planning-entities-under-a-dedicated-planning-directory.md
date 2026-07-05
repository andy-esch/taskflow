---
schema: 1
id: 6fjvdf9t848k
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Move desirelines-planning tasks/epics/audits under a dedicated planning/ dir (taskflow_root=planning) to isolate tool entities from ancillary dirs. De-risks the flatten; independent of Phase B.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [migration, planning]
created: "2026-07-04"
updated_at: "2026-07-05"
completed_at: "2026-07-05"
---

# Isolate desirelines-planning entities under a dedicated planning/ directory

## Objective

`desirelines-planning` keeps `tasks/`/`epics/`/`audits/` at the **repo root**, siblings to
~10 unrelated dirs (`archive/`, `research/`, `routines/`, `incidents/`, `issues/`,
`projects/`, `other-ai/`, `scripts/`, `bin/`, `tests/`) plus cruft (`.DS_Store`,
`.pytest_cache/`, four READMEs). taskflow, by contrast, isolates everything under
`planning/`. Move desirelines' three entity dirs under a dedicated, **visible** `planning/`
dir (config `taskflow_root = "planning"`) so the tool owns a clean, self-contained namespace
— removing the directory-clobbering risk and, as a bonus, taking the ancillary dirs + the
stray `audits/HOWTO-execute.md` out of the tool's scope entirely.

**Why now / why separate:** this is **independent of Phase B** — it works with today's
`tasks/<status>/` layout (the tool already supports a config-pointed root), so it's a safe,
reversible de-risking step to do *before* the flatten. The flatten then happens *inside*
`planning/`, where a single flat `tasks/` dir is self-contained rather than a 648-file pile
at the repo root. Keep it **visible** (`planning/`, not a hidden `.tskflwctl/`) — the whole
value is git-native, browse-on-GitHub-without-the-tool.

## Procedure (validated against the code + the real repo)

Run from the `desirelines-planning` repo root. Commit a clean checkpoint first — **git is
the undo**.

1. `mkdir planning`
2. `git mv tasks epics audits planning/`
   — **`git mv` has NO `-r`**; it takes `<source>... <destination>` and moves directory
   trees natively, so one command moves all three into `planning/`. **Leave the ancillary
   dirs at root** (they are not tskflwctl entities; nothing scans them).
3. Edit `.tskflwctl.toml`: `taskflow_root = "planning"` (was `"."`). **Leave `tracked_repos`
   (`../desirelines`, `../desirelines-deploy`) unchanged** — they resolve relative to the
   `.tskflwctl.toml` (repo root), not to `taskflow_root`, so the pointers stay valid.
4. **Fix only the links the depth change breaks** (this is small — do NOT confuse it with the
   big flatten link-rewrite):
   - Entities dropped one level (`epics/x.md` → `planning/epics/x.md`), so links from an
     entity to a **root-level ancillary dir** need one more `../`:
     `](../research/…)` → `](../../research/…)`, same for `../archive/`, `../incidents/`, etc.
     (Confirmed live: `epics/15-…` links `../research/…`; `epics/04-…` links `../archive/…`.)
   - **Entity↔entity links are PRESERVED** — all entities move together, so their relative
     structure is unchanged (`](../completed/…)` within tasks, `](../tasks/…)` from epics
     still resolve). The 122-file entity↔entity relative-link rewrite is a **Phase B / flatten**
     concern, not this move.
5. **Docs / prose:** update desirelines' `CLAUDE.md`, `AI_README.md`, `GEMINI.md`, `README.md`
   — any `tasks/`, `ls tasks/in-progress/`, `epics/`, `audits/` mention → `planning/…`.
6. **Ops:** `bin/pm`, `scripts/`, `routines/` that hardcode `tasks/`/`epics/`/`audits/` paths
   → update (or note; `pm` is being retired).
7. **Stray non-entity file:** `audits/HOWTO-execute.md` (no frontmatter) moves to
   `planning/audits/`. Either relocate it out of the entity dir, or rely on the Phase B code
   fix (the flat scan skips non-frontmatter files). Decide.
8. **Cruft:** ensure `.DS_Store` and `.pytest_cache/` are gitignored.
9. **Verify:** from the repo root, `tskflwctl lint` (discovery reads `.tskflwctl.toml` →
   `planning/`) lists every entity; `tskflwctl task list` / `board` work; the impl repos
   (`../desirelines` via `tracked_repos`) still resolve. One rename-churn commit; git is undo.

## Acceptance criteria

- [ ] `tasks/`/`epics/`/`audits/` live under `planning/`; `.tskflwctl.toml` has
      `taskflow_root = "planning"`; `tskflwctl lint` + `task list` pass from the repo root.
- [ ] Entity→ancillary links (`../research`, `../archive`, …) fixed; entity↔entity links
      still resolve; prose docs updated.
- [ ] `tracked_repos` still resolve (the `../desirelines` pointer works).

## Out of scope

- The **flatten** itself (id-led filenames, retire status==directory) — that's Phase B
  ([flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory](6fhnydm03edq-flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory.md)); this move
  is deliberately layout-preserving and can ship first.
- The **122-file entity↔entity relative-link rewrite** — preserved by this move, rewritten by
  the flatten migration.
- Cleaning up / migrating the ancillary dirs themselves (archive/research/routines/…).

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- De-risks [flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory](6fhnydm03edq-flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory.md)
  (Phase B) and [one-time-migration-script-this-repo-desirelines](6fhnydm029gt-one-time-migration-script-this-repo-desirelines.md).

## Decision (2026-07-04) — resolves step 7 (the stray HOWTO)

Step 7 left `audits/HOWTO-execute.md` handling open ("relocate it, or rely on the Phase B
fix — decide"). The carveout decision
([curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate](6fjvr03mr9zg-curation-carveouts-tolerate-non-entity-files-in-tool-dirs-frontmatter-gate.md)) settles it:
**move it to a top-level `planning/meta/`** as part of this isolate move, rather than leaving
it in `planning/audits/`.

- This is the **first place the `meta/` convention appears** — it is just an unscanned sibling
  folder (the tool only scans `tasks/`/`epics/`/`audits/`), so it is safe pre-flatten and needs
  no code.
- Doing it here means the flatten's migration sweep has one less loose file to handle.
- Same treatment for any routine specs currently loose near the entity dirs -> `meta/routines/`.

## Verified complete (2026-07-05)

All acceptance criteria hold in `desirelines-planning`:
- `tasks/`/`epics/`/`audits/` live under `planning/`; root-level entity dirs are gone.
- `.tskflwctl.toml`: `taskflow_root = "planning"`, `tracked_repos = ["../desirelines", "../desirelines-deploy"]` (resolve).
- `tskflwctl lint` passes from the repo root (discovery via the toml).
- The stray `HOWTO-execute.md` landed in `planning/meta/` per the carveout decision.

Shipped ahead of Phase B as the layout-preserving de-risk it was scoped to be; the flatten
then ran inside `planning/`.
