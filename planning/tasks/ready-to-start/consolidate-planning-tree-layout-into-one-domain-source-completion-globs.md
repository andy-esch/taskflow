---
status: ready-to-start
epic: 17-pm-go-cli
description: completion.go still globs tasks/<status>, audits/<bucket>, epics directly; derive the subdirs from one domain helper shared with WatchPaths/Init
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, architecture, refactor]
created: "2026-06-14"
---

# Consolidate planning-tree layout into one domain source (completion globs)

## Objective

Finish the layout-consolidation started in
[[put-storage-layout-knowledge-back-behind-the-port]]. That work made
`store.WatchPaths()` the canonical dir set and had `config.Init` derive its dirs
from `domain.AllStatuses()`/`AllAuditBuckets()`. But `cli/completion.go` is still
a third place that encodes the tree shape — it globs `tasks/<status>`,
`audits/<bucket>`, and `epics/` directly (`completion.go:90-94,114,137`). It
can't call `WatchPaths()` (shell completion runs before DI and must tolerate
malformed frontmatter / a missing service), so the fix is to push the *relative*
subdir enumeration down into `domain`, where all three consumers can share it.

## Acceptance criteria

- [ ] A `domain` helper returns the canonical planning subdirs (e.g.
      `TaskDirs()`/`AuditDirs()` or one `LayoutDirs()` yielding relative paths),
      derived from `AllStatuses()`/`AllAuditBuckets()`.
- [ ] `completion.go`, `config.Init`, and `store.WatchPaths()` all build from that
      helper — no independent `tasks/<status>` / `audits/<bucket>` literals or
      ad-hoc enumeration remain.
- [ ] A sync-guard test (or extend the existing ones) so a new status/bucket
      flows to completion automatically.
- [ ] `go build/test/vet` green; `gofmt` clean.

## Out of scope

- `completion.go` switching to the store/service — it deliberately stays
  service-free (works on a half-broken tree, no YAML parse). Only the dir
  enumeration is shared, not the I/O.

## Related

- Epic [[17-pm-go-cli]]
- Completes the layout thread from [[put-storage-layout-knowledge-back-behind-the-port]]
- Touches `internal/domain/`, `internal/cli/completion.go`,
  `internal/config/config.go`, `internal/store/fsstore.go`
