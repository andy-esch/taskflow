---
status: completed
epic: 17-pm-go-cli
description: TUI watcher and config.Init both hardcode the store directory layout; assorted ARCHITECTURE.md and comment drift from the 2026-06-12 review
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, architecture, docs]
created: "2026-06-12"
updated_at: "2026-06-14"
started_at: "2026-06-14"
completed_at: "2026-06-14"
id: 6fbj87001q7p
---
# Put storage-layout knowledge back behind the port

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], finding M15 + doc-drift
> lows). The hexagonal import graph verified clean; these are the two spots
> where *layout* knowledge escaped the store, plus doc drift to sweep.

## Objective

1. **M15a — The TUI watcher duplicates the store's directory layout.**
   `watchDirs` reconstructs `<root>/tasks/<status>` and `<root>/audits/…`
   (`internal/tui/watch.go:44-55`), and `cli/ui.go:17` leaks `Cfg.Root` into
   the TUI to make that possible — contradicting "the TUI never touches the
   fs" (ARCHITECTURE.md, `tui/model.go` header). Expose `WatchPaths()
   []string` from the store through `core`; fsnotify mechanics stay in `tui`.
2. **M15b — `config.Init` hardcodes the status-dir list** as string literals
   (`config.go:104-107`) instead of deriving from `domain.AllStatuses()`,
   with no sync-guard test (the TUI's `statusViews` has one). A new status
   would ship with `init` not scaffolding its dir while the watcher watches
   it.
3. **Doc/comment drift:** ARCHITECTURE.md calls `cmd/tskflwctl` "the sole
   composition root" but the wiring lives in `cli/root.go:99`;
   `cli/root.go:2` still says "a future TUI" — it shipped; `tui/model.go:34`
   documents `m.root` as "reserved for the S3 watch" but the field is dead
   (only the `Run` parameter is used) — delete field or comment;
   ARCHITECTURE.md quotes `core.TaskStore` where the code asserts
   `core.Store`; acknowledge the pragmatic `Task.Path` fs leak instead of
   claiming domain purity unqualified.
4. **Cheap forward-compat while in here:** consider reserving a schema/
   version key in `init`/`task new` scaffolds (files currently carry no
   version marker); consider `domain.CountFindings(body)` so the audit
   "what counts as open" rule (`auditstore.go:15-24` regexes) becomes a
   testable domain invariant.

## Acceptance criteria

- [x] One source of truth for the directory layout, with a sync-guard test
      covering `init` scaffolding and watcher paths. — `store.WatchPaths()` (M15a)
      is the single layout source, exposed through `core.Store`/`Service`;
      `config.Init` derives its dirs from `domain.AllStatuses()`/`AllAuditBuckets()`
      (M15b). Guards: `TestFS_WatchPaths`, `TestInitScaffoldsEveryStatusAndBucket`.
- [x] The TUI no longer receives a raw fs root. — `New(svc)`/`tui.Run(svc)` dropped
      the `root` param; the watcher takes `svc.WatchPaths()`; `ui.go` no longer
      passes `Cfg.Root`; the dead `Model.root` field is gone.
- [x] Listed doc/comment drift corrected. — `root.go` "future TUI"→shipped;
      ARCHITECTURE.md `core.TaskStore`→`core.Store`, composition-root wording,
      `Task.Path` purity caveat, watcher-paths note. (`m.root` comment resolved by
      deleting the field.)

> Item #4 (schema/version key in scaffolds; `domain.CountFindings`) was **not**
> done — it's an optional "consider," and `CountFindings` belongs with the audit
> regexes. Left for a separate task if wanted.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/tui/watch.go`, `internal/cli/ui.go`,
  `internal/config/config.go`, `internal/core/`, `docs/ARCHITECTURE.md`.