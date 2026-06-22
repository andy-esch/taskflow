---
schema: 1
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: render.go (703 LOC) and core/service.go (~693 LOC) grow per entity x use-case; split by concern behind the same facade.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [maintainability, refactor]
created: "2026-06-22"
---
# Split render.go and service.go god-files by concern

## Objective

Two files grow monotonically and are the future merge-conflict epicenters. Finish
the by-concern split already started (envelopes.go/columns.go/finding.go): move JSON
DTOs + schema-contract types out of render.go (dto.go/schema.go), keep render.go for
generic table renderers; move per-entity use-cases out of service.go into
service_task/epic/audit.go behind the same Service facade (port unchanged).

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M9** (render.go) and **L1** (service.go). No behavior change; caps
cognitive size before the reporting/entity roadmap lands. Pairs naturally with M1.

## Acceptance criteria

- [ ] render.go holds only generic renderers; DTOs/schema types live elsewhere.
- [ ] service.go split into entity-scoped files; Service facade + core.Store port unchanged.
- [ ] just test + just lint green; golden snapshots unchanged.

## Implementation plan

**Approach.** A pure file-level move within each package — no API, no behavior, no
byte-of-output change. `render.go` (703 LOC) and `core/service.go` (688 LOC) are both
single `package render` / `package core` files, so moving funcs/types into new files
in the same package is mechanical and the existing tests (golden + unit) are the
safety net. Follow the by-concern split already begun (`envelopes.go`, `columns.go`,
`finding.go`). Do NOT attempt the field-descriptor `*ShowHuman` consolidation the
audit floats as a "consider" — that is a behavior-risking rewrite better folded into
M1's descriptor work; keep this task to the zero-risk move.

**Steps (render, M9).**
1. Add `internal/cli/render/dto.go`: move the unexported JSON DTO structs +
   their mappers — `taskJSON`/`toJSON`, `auditJSON`/`auditToJSON`,
   `epicJSON`/`epicMetaJSON`/`toEpicMeta`, `statusCountJSON`, `findingJSON`,
   `lintTaskJSON` (currently lines 36–66, 270–296 partial, 340–410, 444–468 of
   render.go). The `*Envelope` structs already live in `envelopes.go` — leave them.
2. Add `internal/cli/render/schema_render.go` (name to avoid colliding with the
   cli-package `schema.go`): move the `Schema*` contract types + their human/JSON
   funcs — `SchemaStatus`, `SchemaField`, `SchemaExitCode`, `SchemaContract`,
   `SchemaJSON`/`SchemaHuman`, `KindSchema`, `SchemaKindJSON`/`SchemaKindHuman`
   (render.go 607–703).
3. Leave `render.go` holding the generic table renderers + the `*Human`/`*JSON`
   list/show funcs and `encodeJSON`. `style.go` (the `Style` type + `writeTable`)
   stays as-is. Fix the noted copy-drift while files are open: `TaskShowHuman` uses
   `%-12s`, `AuditShowHuman` `%-9s`, `EpicShowHuman` `%-12s` — extract a shared
   `fieldLabel(st, w, label)` helper IF it doesn't change emitted bytes; if widths
   genuinely differ in golden output, leave them and just note the divergence (do not
   churn goldens for cosmetics in a "no behavior change" task).

**Steps (service, L1).**
4. Split `core/service.go` into `service_task.go` (ListTasks/ShowTask/EditTask/
   ReplaceBody/AppendBody/Move/SetFields + NewTask/NewTaskParams + `coerceField`/
   `splitList`/`unknownFieldErr`/the task body template), `service_epic.go`
   (NewEpic/NewEpicParams/ListEpics/rollupEpics/EpicSummary/ShowEpic/epicExists +
   epic body template), `service_audit.go` (NewAudit/NewAuditParams/ListAudits/
   ShowAudit/MoveAudit + audit body template), and keep `service.go` for the
   `Service` struct, `NewService`, `WatchPaths`, `Summary`/`StatusCount`,
   `Lint`/`LintFix`/`LintResult`, and the shared helpers (`hasTag`, `stringify`).
   `finding.go` already holds QueryFindings/LintAudits — leave it. The `core.Store`
   port (store.go) and `Service`'s method set are unchanged.
   NOTE: if M1 (entity descriptor) lands first it will relocate the body templates to
   the domain descriptor — sequence this after M1, or expect a small merge there.

**Tests.** No new tests needed; this is covered by the existing golden snapshots
(`internal/cli/testdata/golden/`, run in-process against `testdata/planning/`) and the
core/cli unit suites. The whole point is they pass byte-for-byte. Run
`go test ./internal/cli` first — any golden diff means an accidental behavior change,
not an expected update; do NOT run with `-update` unless a width-drift fix in step 3
was deliberately chosen.

**Risks / gotchas.** (a) `schema_comments.json` is generated from Go-doc comments by
`internal/tools/schemacomments`; moving the `Schema*`/DTO types to new files can shift
the comment map keyed by `package.Type.Field` — regenerate with
`go run ./internal/tools/schemacomments` and let `render/schema_comments_test.go` +
`schema_descriptions_test.go` confirm no drift. (b) Keep every moved symbol's
exported/unexported casing identical — a stray capitalization is an API change. (c)
gofmt/goimports each new file (golangci-lint's `gofmt`/`goimports` will flag import
blocks). (d) Don't touch `envelopes.go`'s `jsonEnvelopes` registration — the
JSONSchema reflection depends on it.

**Done when.** `render.go` contains only generic renderers + list/show funcs;
DTOs and schema-contract types live in `dto.go`/`schema_render.go`; `service.go` is
the facade + cross-entity use-cases with per-entity files beside it; the `core.Store`
port and `Service` method set are unchanged; and `go build ./...`,
`go test ./...` (goldens unchanged), `golangci-lint run ./...` are green.
