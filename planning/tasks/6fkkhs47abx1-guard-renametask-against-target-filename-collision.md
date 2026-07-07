---
schema: 1
id: 6fkkhs47abx1
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'RenameTask lacks a target-filename collision guard: renaming onto an existing <id>-<slug>.md silently overwrites it. Add an os.Stat check returning ErrConflict instead of clobbering.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [hardening, rename]
created: "2026-07-06"
updated_at: "2026-07-06"
started_at: "2026-07-06"
completed_at: "2026-07-06"
---
# Guard RenameTask against target-filename collision

Adversarial-review finding (Scheme-2 review, 2026-07-05) — the last untracked
finding from that pass; the rest were fixed in code or tracked elsewhere.

## Context

`internal/store/rename.go` — `RenameTask`. Severity: High (data-loss hazard),
but a **narrow trigger**: `newPath` is `<id>-<newslug>.md` with the id
preserved, so a collision requires *another file sharing the same id* — a
pre-existing duplicate-id corruption. The loss is "silently clobbering a corrupt
same-id sibling," not a routine path.

## Trigger

A rename produces a filename that already exists on disk (e.g. `<id>-new.md`
exists from a stray file or an aborted rename), with `newName != oldName`.

## Why it's a bug

`RenameTask` never checks whether `newPath` exists before the write loop. The
target's edit is `{newPath, renamedContent}`, so `writeFileAtomic(newPath, ...)`
overwrites whatever was there, then the old file is removed. No `os.Stat`, no
`ErrConflict` — silent overwrite.

## Proposed fix

Guard before compiling/committing edits (aligns with the project's fail-loud
philosophy — see the invalid-frontmatter handling):

```go
if newPath != oldPath {
    if _, err := os.Stat(newPath); err == nil {
        return domain.Task{}, 0, fmt.Errorf(
            "%w: target filename already exists: %s", domain.ErrConflict, newName)
    }
}
```

Add a regression test: rename onto an existing target file → `ErrConflict`,
nothing written, both files intact.
