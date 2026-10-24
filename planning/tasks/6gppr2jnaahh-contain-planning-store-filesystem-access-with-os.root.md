---
schema: 1
id: 6gppr2jnaahh
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Enforce directory containment in store.FS with os.OpenRoot, validate create slugs at the store boundary, and keep dangler probes from following symlinks out of the root.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [store, security, hardening]
created: "2026-10-23"
audit_sources: [planning/audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md]
updated_at: "2026-10-23"
---
# Contain planning store filesystem access with os.Root

## Objective

Store traversal relies on lexical checks: `validQueryName`, regular-file filters, and `filepath.Rel`
in `DanglingLinks`. Go's `os.Root` enforces containment at the OS level. Decide where it fits beside
the store's guarded and atomic write helpers, adopt it for reads and probes, and close the two
concrete gaps the audit named.

## Acceptance criteria

- [ ] Create paths refuse a slug that is not a valid single path component at the store boundary, independent of `domain.Slugify`.
- [ ] `DanglingLinks` cannot stat or follow a symlink that resolves outside the planning root.
- [ ] Store reads and existence probes go through an `os.Root` handle, or the reason a path stays lexical is recorded beside it; absolute paths needed for `$EDITOR` and `path` output are preserved.
- [ ] Atomic write, CAS, and lock behavior are unchanged, with regressions for symlink escape and hostile slugs.

## Out of scope

- Unifying the three `writeFileAtomic` implementations, though this design must not block it.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-14 Go 1.27 modernization](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md), finding M2
- [Unify writeFileAtomic](6g63jj1dh0sb-unify-the-three-divergent-writefileatomic-implementations.md) and [shared integrity foundations](6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md)
