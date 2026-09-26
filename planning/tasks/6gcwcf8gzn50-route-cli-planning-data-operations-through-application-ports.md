---
schema: 1
id: 6gcwcf8gzn50
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Stop lint repair, link checking, and entity completion from reaching directly into the filesystem adapter.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, cli, ports, lint]
created: "2026-09-23"
depends_on: [6gcwcf88z57p]
updated_at: "2026-09-26"
---

# Route CLI planning data operations through application ports

## Objective

Close the remaining direct primary-to-secondary planning-data paths. The CLI currently invokes
frontmatter repair and link lint ports directly, constructs `store.FS` for status-aware completion,
and globs entity directories itself. Those are deliberate exceptions today, but they bypass
application orchestration and make a second storage adapter incomplete even when it satisfies the
documented semantic ports.

## Scope

- Expose application-level use cases for safe frontmatter repair and optional link integrity checks,
  retaining dry-run, partial-durability reporting, and command-safety enforcement.
- Define a parse-free completion capability that can include malformed local records without making
  Cobra know the directory layout.
- Inject these capabilities through composition rather than constructing `store.FS` inside command
  controllers.
- Build on the settled entity read/source vocabulary and optional-local-capability composition;
  do not reintroduce paths through a CLI-only convenience store.
- Keep completion failure silent and fast as required by shell completion UX.

## Acceptance criteria

- [ ] CLI command controllers do not import or construct `internal/store` for lint repair, link
      checking, or entity completion.
- [ ] Repair and link checks run through named application use cases with unchanged human/JSON,
      exit-code, dry-run, and partial-progress behavior.
- [ ] Completion still offers malformed id-led local records and preserves duplicate-slug
      disambiguation without filesystem globs in the Cobra adapter.
- [ ] A non-filesystem fake can drive all three workflows without providing directories or paths.
- [ ] Tests pin call counts and prove no hidden fallback opens the filesystem adapter.

## Out of scope

- Moving user-supplied body/manifest file reading out of the CLI adapter.
- Moving TUI filesystem watching behind a planning-data service.
- Changing completion vocabulary or repair authorization.

The older reusable-workspace task remains the owner of init, doctor, and discovery orchestration.
This task is the narrower planning-data boundary triggered by the now-explicit multi-adapter
contract; it does not reactivate that task's unrelated deferred scope.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
- [Reusable workspace discovery seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)
