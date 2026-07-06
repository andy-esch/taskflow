---
status: completed
epic: 17-pm-go-cli
description: completion.go still globs tasks/<status>, audits/<bucket>, epics directly; derive the subdirs from one domain helper shared with WatchPaths/Init
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [go, architecture, refactor]
created: "2026-06-14"
started_at: "2026-06-14"
updated_at: "2026-06-14"
completed_at: "2026-06-14"
id: 6fc6vcr00hs7
---

# Consolidate planning-tree layout into one domain source (completion globs)

## Objective

Finish the layout-consolidation started in
[put-storage-layout-knowledge-back-behind-the-port](6fbj87001q7p-put-storage-layout-knowledge-back-behind-the-port.md). That work made
`store.WatchPaths()` the canonical dir set and had `config.Init` derive its dirs
from `domain.AllStatuses()`/`AllAuditBuckets()`. But `cli/completion.go` is still
a third place that encodes the tree shape — it globs `tasks/<status>`,
`audits/<bucket>`, and `epics/` directly (`completion.go:90-94,114,137`). It
can't call `WatchPaths()` (shell completion runs before DI and must tolerate
malformed frontmatter / a missing service), so the fix is to push the *relative*
subdir enumeration down into `domain`, where all three consumers can share it.

## Acceptance criteria

- [x] A `domain` helper returns the canonical planning subdirs. — `domain/layout.go`:
      `TasksDir`/`EpicsDir`/`AuditsDir`/`ProjectsDir` constants + `TaskStatusDirs()`
      / `AuditBucketDirs()` derived from `AllStatuses()`/`AllAuditBuckets()`.
- [x] `completion.go`, `config.Init`, `store.NewFS`/`WatchPaths` (and config's repo
      discovery) all build from it — verified no `"tasks"`/`"audits"`/`"epics"`/
      `"projects"` literals remain in those files.
- [x] Sync-guard tests: new `TestTaskStatusDirs`/`TestAuditBucketDirs` pin the
      derivation; the existing `TestInitScaffoldsEveryStatusAndBucket` and
      `TestFS_WatchPaths` still cover init + watcher.
- [x] `go build/test/vet` green; `gofmt` clean.

## Out of scope

- `completion.go` switching to the store/service — it deliberately stays
  service-free (works on a half-broken tree, no YAML parse). Only the dir
  enumeration is shared, not the I/O.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md)
- Completes the layout thread from [put-storage-layout-knowledge-back-behind-the-port](6fbj87001q7p-put-storage-layout-knowledge-back-behind-the-port.md)
- Touches `internal/domain/`, `internal/cli/completion.go`,
  `internal/config/config.go`, `internal/store/fsstore.go`
