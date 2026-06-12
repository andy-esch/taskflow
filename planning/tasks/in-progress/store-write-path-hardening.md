---
status: in-progress
epic: 17-pm-go-cli
description: Move parses after mutating disk, set-after-move resurrects a moved task, CRLF edits mix line endings, lint --fix silent on unrepairable files
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [go, store, data-safety]
created: "2026-06-12"
started_at: "2026-06-12"
updated_at: "2026-06-12"
---
# Store write-path hardening

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], findings M1/M3 + B3/B4
> carried from [[2026-06-11-critical-review-and-polish-research]]). These are
> the data-safety siblings of the in-progress
> [[harden-task-set-against-silent-frontmatter-corruption]] — same
> parse-before-commit principle, more call sites.

## Objective

1. **M1 — `Move` mutates disk, then can report failure.**
   `internal/store/fsstore.go:139-145` parses the assembled content *after*
   writing the new file and removing the old; a task that round-trips as
   YAML but fails typed decode (e.g. `tier: "4"`) moves on disk yet returns
   an error — a retrying agent sees a "failed" move that succeeded.
   `MoveAudit` has the same shape (`auditstore.go:102-105`, parse after
   `os.Rename`). Parse before committing, exactly as the new `SetFields`
   guard does.
2. **M3 — No read-modify-write concurrency control.** `SetFields` writes
   back to the originally-resolved path (`fsstore.go:150-174`); a concurrent
   `Move` in between means the write resurrects the slug in its old status
   dir → permanent `ErrAmbiguous` with no repair tooling. Re-resolve before
   rename (compare-and-swap) or the advisory `flock` ARCHITECTURE.md already
   lists. Related: `nextEpicNumber` TOCTOU (`create.go:81-132`) — `O_EXCL`
   only dedupes identical `NN-slug`.
3. **B3 — CRLF files get mixed line endings on surgical edit.**
   `assembleFile` (`frontmatter.go:86-92`) always emits LF frontmatter over
   a CRLF body (same in `fix.go:114-119`). Detect the dominant ending and
   normalize, or document that edits normalize to LF. No CRLF round-trip
   test exists (fuzz seeds only assert no-panic).
4. **B4 — `lint --fix` exits 0 and says nothing about unrepairable files.**
   `runLintFix` (`cli/lint.go:56-66`) returns nil unconditionally;
   `FixFrontmatter` (`store/fix.go:19-67`) only reports files it changed.
   Surface still-unreadable files and exit non-zero.
5. **Lows while in the area:** unterminated frontmatter (`---` never closed)
   parses as "no frontmatter" with no `FileProblem`, and a later `SetFields`
   double-fences the file (`frontmatter.go:45`, `fsstore.go:202-209`);
   `config.Init` writes non-atomically with a check-then-write race
   (`config.go:121-123`) against the repo's own atomic-helper convention;
   `writeFileAtomic` never fsyncs the parent dir, so `Move`'s crash-safety
   comment (`fsstore.go:130-133`) overstates the guarantee — fix or reword.

## Acceptance criteria

- [ ] No store mutation can succeed on disk while reporting failure
      (parse-before-commit everywhere; tests per call site).
- [ ] set-after-move cannot resurrect a task (test with an interleaved move).
- [ ] CRLF file round-trips with consistent endings (behavioral test).
- [ ] `lint --fix` on an unrepairable file prints it and exits non-zero.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/store/fsstore.go`, `auditstore.go`, `frontmatter.go`,
  `fix.go`, `atomic.go`, `create.go`, `internal/config/config.go`,
  `internal/cli/lint.go`.