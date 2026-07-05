---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: EditTask/EditEpic/EditAudit are near-verbatim resolve→loop→parse→CAS→write copies; fold the rule-of-three into one generic helper so the parse-before-accept + CAS contract lives once.
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [store, refactor]
created: "2026-06-28"
updated_at: "2026-06-28"
completed_at: "2026-06-28"
id: 6fgq1n003jf9
---
# Generify the store editor-loop across `EditTask` / `EditEpic` / `EditAudit`

## Objective

The three store editor entry points — `EditTask`, `EditEpic`, `EditAudit` in
`internal/store/edit.go` — are near-verbatim copies of one skeleton:

1. resolve the slug to a `(path, meta)` pair (`resolve` / `resolveEpic` /
   `resolveAudit`),
2. read the file, hand its content to the `edit` callback in a loop,
3. accept the result only if it still parses (`parseTask` / `parseEpic` /
   `parseAudit`) — reopen on a broken edit, give up if re-saved unchanged,
4. compare-and-swap the path before the write (guard against a concurrent move
   relocating the file during the editor window),
5. write atomically and return `(entity, changed, err)`.

They differ only in the resolve fn, the parse fn, and the entity type. Fold the
duplication into one generic helper so the parse-before-accept + CAS contract —
the subtle, security-relevant part — is defined and tested in exactly one place.

## Context

Surfaced by the 2026-06-28 adversarial review of the audit-edit work
(`audit edit` / `audit append`, shipped in
[audit-editing-faces-audit-edit-set-and-append](6ffr4wc01thc-audit-editing-faces-audit-edit-set-and-append.md)). The reviewer flagged the
editor-loop as the rule-of-three: it was a copy when there were two faces (task,
epic); `EditAudit` made it three. Deliberately deferred out of that PR's scope
(don't refactor shared infra inside a feature PR), filed here.

Relates to epic 21 (code-quality / architecture hardening).

## Acceptance criteria

- [ ] One generic helper (e.g. `editFile[T any]`) parameterised by a resolve fn,
      a parse fn, and the entity type, carrying the loop + parse-before-accept +
      CAS + atomic-write logic.
- [ ] `EditTask`/`EditEpic`/`EditAudit` become thin wrappers over it; the
      per-entity error wording (`task`/`epic`/`audit %q changed on disk…`) is
      preserved (pass the noun in, or wrap at the call site).
- [ ] The "no net change but the file was already broken on disk" branch and the
      "re-saved the same broken content → user gave up" branch both survive
      (they're easy to drop in a careless merge).
- [ ] Existing `TestEdit{Task,Epic,Audit}_*` all stay green unchanged — the
      refactor is behaviour-preserving. go build/test/lint green.

## Implementation sketch

- Go generics make this clean: `func editFile[T any](resolve func(string) (path string, meta M, err error), parse func([]byte, string, M) (T, error), noun, slug string, edit ...) (T, bool, error)`.
  `M` (the per-entity meta: status / "" / bucket) can itself be a type param or
  an `any` threaded opaquely to `parse`.
- Keep the helper in `edit.go`; the three public methods stay as the typed API
  the core ports expect (`TaskStore`/`EpicStore`/`AuditStore`).

## Risks / gotchas

- The CAS re-resolve compares `curPath != path` — the generic version must
  re-call the *same* resolve fn, not a captured path, or the guard rots.
- `EditBody` / `AppendAuditBody` (in `body.go`) are a *different* shape (no
  interactive loop) — out of scope; don't try to fold those in.

## Completed 2026-06-28

Shipped as planned. New generic `editFile[T any](noun, path, orig, parse, recheck,
edit)` in `store/edit.go` carries the whole loop (parse-before-accept, reopen-on-
broken, give-up, unchanged-but-pre-broken, and the optional pre-write recheck).
The three faces are now thin wrappers that supply a `parse` closure and a `recheck`
closure: `EditTask`/`EditAudit` pass a re-resolve CAS guard (the relocation
conflict message stays per-noun, baked into the closure); `EditEpic` passes
`nil` (epics never move directories). The per-noun write-error wording is preserved
via the `noun` param (`write task/epic/audit …`). Behaviour-preserving: every
existing `TestEdit{Task,Epic,Audit}_*` (incl. the relocation/CAS conflict and
broken→fixed-reopen cases) passes untouched; build/vet/test/lint green. `EditBody`/
`AppendAuditBody` deliberately left alone (different shape, as scoped).

## Done when

The resolve→loop→parse→CAS→write contract lives in one helper, the three faces
are thin wrappers, and every existing edit test passes untouched.
