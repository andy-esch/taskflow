---
schema: 1
status: completed
epic: 22-selectable-template-library
description: CLI reads domain.Template* directly; route resolution through a store-backed TemplateSource behind core.Service so step 4 (repo-local) is a port swap.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, templates, dx]
created: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
---
# Route template resolution through a core port

## Objective

Template DATA resolution currently bypasses core: `internal/cli/template.go` calls
`domain.LookupTemplate`/`TemplatesFor`/`SchemaKinds` directly (like `schema`, which is
fine for built-in-only data), while every other read/create surface goes through
`core.Service`. Step 4 (repo-local templates) needs the filesystem, so resolution
must move behind a store-backed port. Build that seam now so step 4 is a port swap,
not a CLI refactor.

## From the 2026-06-22 adversarial review of the template work (theme 2, ranks 4, 11)

- Add a `TemplateSource` port + `core.Service` methods (e.g. `ListTemplates(kind)`,
  `ShowTemplate(kind,name)`, and a raw-body accessor for the create paths) that today
  delegate to the built-in registry and later merge repo-local.
- Repoint `template list`/`show` and the create paths (`NewTask/NewEpic/NewAudit`) to
  the service method; keep `domain.Template*` as the built-in source the service reads.
- Make `template`'s PersistentPreRunE resolve the repo *best-effort* (built-ins still
  work repo-less; repo-local layers on when a repo is present).
- NOTE: the closed-`switch kind` renderer drift (original rank 2) is already fixed —
  `TemplateBody`/`RenderLabels` are descriptor-driven. This task is only the
  layering/port move.

## Acceptance criteria

- [x] `template list/show` and the create paths resolve via core.Service, not domain directly.
- [x] Built-in templates still work with no planning repo.
- [x] A fakeStore/port test proves the seam; `go build`/`go test`/`golangci-lint`/docs-check green.

## Related

- Sequenced before/with step 4 of [[design-a-selectable-template-library]].
