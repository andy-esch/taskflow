---
schema: 1
status: completed
epic: 20-cli-ux-and-ergonomics
description: Accept titles with :, em-dashes, arrows on task/epic/audit new (Slugify the filename, keep the title in the body H1) instead of ValidateTitle rejecting them.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-25"
started_at: "2026-06-25"
completed_at: "2026-06-25"
---
# Slugify titles on create instead of rejecting filename-hostile characters

## Objective

Creating an entity with a perfectly reasonable title that happens to contain a
`:`, em-dash, or arrow is rejected outright — you have to retype a "clean" title.
(Felt repeatedly authoring the very tasks that spawned this one.) Make the friction
go away naturally: accept any title, derive the slug, and keep the full title intact.

## Context — this reverses a deliberate decision, read first

`domain.ValidateTitle` (slug.go) is an intentional 2026-06-13 guard. Its reasoning:
`Slugify` is an allowlist that *silently drops* non-ASCII punctuation (em-dashes,
arrows, bullets) and breaks on filesystem-reserved ASCII (`:` `/` `*` …), so a
title like `Foo: bar — baz` would silently become the slug `foo-bar-baz`. The guard
chose to **surface that loudly** (reject + suggest a clean title via `suggestTitle`)
rather than silently mangle. It runs on all three create paths:
`service_task.go:264`, `service_epic.go:41`, `service_audit.go:33` (the audit
*area*), each just before `Slugify`.

What makes reversing it safe today:
- The body scaffold already renders the **full original title** as the H1
  (`renderTemplate(tmpl, {"title": p.Title})`) — so the title is NOT lost when the
  slug drops characters.
- There's already an explicit **empty-slug** error (`title produced an empty slug`)
  for the genuinely-unusable case (e.g. a title of only punctuation).
- The create confirmation already prints the resulting **path/slug**, so the
  derivation is visible, not silent — which is the original concern's actual fix.

Relates to epic 20 (CLI UX).

## Acceptance criteria

- [ ] `task/epic/audit new` accept titles/areas with `:`, em/en dashes, arrows,
      curly quotes, etc.; the slug is `Slugify`'d and the full title is preserved
      (body H1 today; optionally a `title:` frontmatter field — see Decisions).
- [ ] Genuinely-unusable input still errors clearly: a title that slugifies to
      empty, and anything truly filesystem-illegal `Slugify` can't neutralize.
- [ ] The create confirmation makes the derived slug obvious (it already shows the
      path; consider a `→ slug: <slug>` line when it differs notably from the title).
- [ ] go build/test/lint green; `ValidateTitle`/`suggestTitle` tests updated or
      removed; docs/cli regenerated if help text changes.

## Decisions to settle

1. **Silent vs. surfaced.** Pure accept-and-slugify (rely on the printed path), or a
   one-line "created … (slug for '<title>')" note when the slug diverges? The note
   honors the original 2026-06-13 concern at no friction cost — recommended.
2. **`title:` frontmatter field?** Today the title lives only in the slug + body H1.
   Adding a machine-readable `title:` (schema bump) would let the slug be a pure id
   and the title a real field — nice, but scope. Decide in/out.
3. **Keep a narrowed `ValidateTitle`?** Drop it entirely (Slugify already neutralizes
   `/`, control chars, etc. as word breaks) vs. keep it for only the empty-slug /
   path-traversal backstop. Prefer the smallest guard that can't produce a bad path.

## Implementation sketch

- Remove the `ValidateTitle` calls from the three create paths (or narrow it to the
  empty-slug / illegal-path backstop the per-path empty-slug error already half-does).
- Keep/centralize the empty-slug error.
- Optionally thread the original title into the confirmation render + (if chosen) a
  `title:` field in the task/epic frontmatter and the audit body.

## Risks / gotchas

- **Slug collisions**: two different titles can slugify to the same id (`Foo: bar`
  and `Foo bar`). `duplicateSlugIssues`/create's exclusive-create should already
  catch a clash on write — confirm the create path fails cleanly (don't overwrite).
- Don't reopen the path-traversal hole the allowlist closed: verify `Slugify` still
  reduces `../x`, `a/b`, control chars to safe word-broken slugs (it does today).
- Update the `schema`/authoring guidance if it tells agents titles must be clean.

## Done when

`tskflwctl task new "Wire OAuth: PKCE + refresh"` just works — file
`wire-oauth-pkce-refresh.md`, H1 `# Wire OAuth: PKCE + refresh` — no rejection, and
the derived slug is visible in the confirmation.
