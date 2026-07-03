---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: Drive per-entity bits (dir/fields/parse/columns) from a descriptor so a new entity (project/adr) isn't a ~15-file edit.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, growth]
created: "2026-06-22"
started_at: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r02hh1
---
# Entity descriptor to collapse per-entity fan-out

## Objective

Adding the already-scaffolded project/adr entity is a ~15-file shotgun edit across
6 packages because task/epic/audit are hand-enumerated in every layer. Introduce a
data-driven entity descriptor (dir name, field order, parse/serialize, columns) and
drive the already-generic machinery (scanDir[T], Column[T], resolveID, writeNewFile)
from it, so a new entity is a registry entry, not a cross-package edit.

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M1** (the headline growth lever). This is the biggest sustainable-growth
item; ProjectsDir is already scaffolded in layout.go/Init, so it's imminent.

## Acceptance criteria

- [ ] A new entity type can be added without per-layer enumeration in domain/store/core/render/tui.
- [ ] ARCHITECTURE.md reflects the real (now-small) fan-out.
- [ ] just test + just lint green.

## Implementation plan

**Approach.** This is the largest item; do it as a *staged, store-first* descriptor, not
a big-bang. The generic seams already exist (`scanDir[T]` in store/resolve.go,
`Column[T]` in render/columns.go, `resolveID`/`markdownCandidates`, `writeNewFile`,
the domain `layout.go` dir helpers, the `entityTab` registry in tui/entity.go). What
is *not* generic is the per-entity field order + parse/serialize and the hand-written
human/JSON render funcs. The lowest-risk, highest-value first step is to land an
`audit`/`task`/`epic`-shaped descriptor that drives the **store layout + create/list
scan + candidate gather** (the part that is genuinely uniform), then collapse the
domain enumeration sites (schema/authoring/conventions/scaffold) onto it. Leave the
render `*Human` funcs and the TUI delegates entity-specific for now (M9/M10 are
separate tasks); the win is that lighting up `project`/`adr` no longer touches store
or the domain dir/schema sites. Reflection-based generic frontmatter parse is
explicitly out of scope — too magical for this codebase's "grounded, explicit" voice.

**Steps.**
1. **Domain descriptor.** Add `internal/domain/entity.go` defining `type EntityKind`
   and a `Descriptor{ Kind, Dir string; Buckets []string (status/bucket subdirs, nil
   for flat like epics); AuthoringFields []FieldDoc; Conventions []string;
   BodyTemplate string }`. Populate one registry `var entities = []Descriptor{…}` for
   task/epic/audit, seeded from the data already scattered across `domain/schema.go`
   (`taskAuthoringFields`, `epicAuthoringFields`, `auditAuthoringFields`,
   `Conventions`, `SchemaKinds`) and `domain/layout.go` (`TasksDir`/`EpicsDir`/
   `AuditsDir`/`ProjectsDir`). Re-point `SchemaKinds()`, `AuthoringFields(kind)`,
   `Conventions(kind)` to read the registry instead of their own `switch`. Keep the
   `TaskStatusDirs()`/`AuditBucketDirs()` helpers but back them with the descriptor's
   `Buckets`.
2. **Scaffold.** Move the three body templates (`taskBodyTemplate`/`epicBodyTemplate`/
   `auditBodyTemplate`, currently in `core/service.go`) onto the descriptor's
   `BodyTemplate`, and make `core.ScaffoldBody(kind)` (core/scaffold.go) look them up
   by kind rather than `switch`. Watch the placeholder arity (task takes title+epic,
   epic takes title+description, audit takes area+date) — keep a per-descriptor
   `fillTemplate(args…)` or just keep `Sprintf` at the call sites and store only the
   format string.
3. **Store layout from the descriptor.** Rework `FS.WatchPaths()` (store/fsstore.go),
   `taskCandidates`/`auditCandidates` (fsstore.go/auditstore.go), and the `ListTasks`/
   `ListAudits` scan loops to iterate `domain.AuditBucketDirs()`/`TaskStatusDirs()`
   (already derived) — they already do, so this step is mostly verifying that a new
   flat-or-bucketed entity drops in by adding a descriptor + a `scanDir[T]` call. Add
   `ProjectsDir` to the layout helpers' coverage and to `config.Init`'s dir list
   (already appends `ProjectsDir`).
4. **Doc.** Update `docs/ARCHITECTURE.md` §"Why these boundaries" / the M1 framing:
   state the *real* remaining fan-out after this change (domain descriptor entry +
   store `scanDir` call + render funcs + tui tab), and note render/tui are
   deliberately still per-entity (tracked by M9/M10).

**Tests.** Add `internal/domain/entity_test.go` asserting every `SchemaKind` has a
descriptor with non-empty Dir, AuthoringFields, BodyTemplate, and that
`AuthoringFields`/`Conventions`/`ScaffoldBody` agree with the registry for all kinds
(a table over `SchemaKinds()`). Reuse the existing drift guards
(`schema_descriptions_test.go`, `TestInitScaffoldsEveryStatusAndBucket`) — they must
stay green unchanged. No golden churn expected (output funcs untouched); if any
`schema <kind>` text moves, regenerate with `go test ./internal/cli -update` and
diff intentionally.

**Risks / gotchas.** (a) Scope creep — resist folding render/tui into this task; gate
those behind M9/M10. (b) `schema_comments.json` is build-generated from Go-doc
comments (`go run ./internal/tools/schemacomments`); moving types between files can
shift the comment map — regenerate it and let `schema_comments_test.go` confirm.
(c) The epic auto-numbering (`nextEpicNumber`) and the flat-vs-bucketed distinction
must survive: epics have no `Buckets`, audits/tasks do — the descriptor's
`Buckets==nil` is the flat marker. (d) Don't break the `status==directory` invariant:
the descriptor names dirs, it must not become a second source of truth for a task's
status.

**Done when.** A `project`/`adr` entity can be lit up in store + domain by adding one
descriptor + one `scanDir` call (demonstrated by a test that scans a temp tree for a
new descriptor), the doc's fan-out claim is accurate, and `go build ./...`,
`go test ./...`, `golangci-lint run ./...` are green.

## Outcome (2026-06-22)

Done. The `domain.Descriptor` registry (`internal/domain/entity.go`) collapsed the
DOMAIN fan-out — schema kinds, authoring fields, conventions, and body templates are
one registry entry per kind; the scaffold renderer is descriptor-driven (no per-kind
switch). The selectable-template library (epic 22) was built on it. The remaining
render/TUI per-entity fan-out is tracked separately by audit findings M9
(split-render.go-and-service.go-god-files) and M10 (make-tui-lifecycle-action-machinery-
registry-driven) — so the audit's broader M1 theme stays in-progress while this
descriptor task is complete.
