---
schema: 1
status: ready-to-start
epic: 22-selectable-template-library
description: Generalize the single per-kind body template into a named, selectable library (built-in + repo-local), chosen via --template and agent-discoverable!
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [feature, templates, dx]
created: "2026-06-22"
started_at: "2026-06-22"
updated_at: "2026-09-06"
id: 6fes83r03rhs
audited: "2026-09-06"
---
# Design a selectable template library

## Objective

Generalize the single per-kind body scaffold (now single-sourced on
`domain.Descriptor.BodyTemplate`, post-M1) into a **named, selectable template
library**: a growable, user-/agent-extensible reference of task/epic/audit (and
future-entity) templates, chosen per project need. M1 put one default template per
kind on the entity descriptor; this turns "one default" into "a set of named
templates" with built-in + repo-local resolution and a selection surface.

## Why

Different projects and different work want different scaffolds — a security audit
vs an architecture audit; a bug vs a feature vs a spike; a lightweight epic vs a
full charter. Today the scaffold is one-size-fits-all. A library lets a repo (or an
agent) pick the right shape, and grows into a shared reference of good templates.

## Sketch (for discussion — not locked)

**Layers / resolution (precedence high → low):**
- **Repo-local** — templates a project defines in its planning tree (e.g.
  `templates/<kind>/<name>.md`), tailoring scaffolds to its needs. Dogfooded:
  markdown + frontmatter like everything else, reusing the split/parse machinery.
- **Built-in** — the curated set embedded in the binary (`go:embed`), the growable
  "reference library" the tool ships. `default` == today's template.
- Repo-local overrides built-in by `<kind>/<name>`.

**Template format:** a markdown file with a small frontmatter header (`name`,
`description`, `kind`, maybe `frontmatter-defaults`/`placeholders`); body is the
scaffold.

**Selection (a flag twin, like the rest of the CLI):**
- `task new "X" --template bug`, `audit new sec --template security`,
  `epic new "Y" --template charter`. Omitted → `default`.
- `--template` is shell-completable (offer resolved names per kind).
- Off-TTY/agent: the flag is the only path (no prompt); a bad name fails exit 11
  listing the available names.

**Discovery (agent-facing):**
- `tskflwctl template list [--kind audit] [--json]` and `template show <kind> <name>`
  — list/inspect built-in + repo-local templates, with their source.
- `schema <kind>` references the available templates so an agent discovers them.
- Authoring face: `template new <kind> <name>` (+ `template edit`) scaffolds a
  repo-local template, mirroring the task/audit authoring surface.

**Architecture fit:** the template store is a new secondary-adapter concern (read
built-ins via embed + repo-local via the fs store) behind a port the cli/create
paths consume; the descriptor resolves a *named* template instead of a single
`BodyTemplate`. Core stays pure.

## Open questions

- Repo-local location: `templates/` at the planning root vs `.tskflwctl/templates/`?
- Do templates carry frontmatter defaults (e.g. a "bug" task pre-tags `bug`), or
  body-only to start?
- Placeholder model: keep per-kind Printf args, or named placeholders (`{{title}}`)
  so custom templates control their own fills?
- How `schema --json` advertises templates; precedence/versioning rules.

## Out of scope (for now)

- Remote/shared template registries (git-backed or URL library) — local built-in +
  repo-local first; remote is a later layer.
- Per-template validation beyond the existing frontmatter lint.

## Suggested first increment

1. Descriptor resolves a *named* template (built-in only to start; `default` ==
   today's template).
2. `--template` on `task/epic/audit new` + completion + off-TTY error.
3. `template list`/`show` read surface (`--json`).
4. Repo-local override via the store (embed built-ins; scan `templates/`).
5. Authoring surface (`template new`/`edit`) + docs.

## Progress / handoff (as of 2026-06-22)

**Steps 1–3 are IMPLEMENTED; steps 4–5 remain.** The design above still holds;
deviations + decisions are noted at the end. Pick up at step 4 (repo-local).

**Done — step 1 (named templates on the descriptor):**
- `Descriptor.BodyTemplate` (one string) → `Descriptor.Templates []NamedTemplate`
  (`{Name, Description, Body}`) in `internal/domain/entity.go`; `DefaultTemplate =
  "default"`.
- Lookups: `domain.Template(kind, name)` (body), `LookupTemplate(kind, name)`
  (metadata), `TemplatesFor(kind)`, `TemplateNames(kind)`. Empty name = default;
  unknown kind/name → `ErrValidation` listing the available names.
- First real second template shipped: audit `security` (`auditSecurityBodyTemplate`).

**Done — step 2 (`--template` selection):**
- `--template` on `task/epic/audit new` (`internal/cli/{task,epic,audit}.go`);
  `New{Task,Epic,Audit}Params` gained `Template`; create paths call
  `domain.Template(kind, p.Template)`. Mutually exclusive with `--body`/`--body-file`;
  shell-completable (`completeTemplateNames`); off-TTY bad name → exit 11.

**Done — step 3 (`template list`/`show`):**
- `internal/cli/template.go`: `template list` (`--kind`, `--json`) + `template show
  <kind> [name]` (`--json`). Runs with no planning repo (overrides PersistentPreRunE,
  like `schema`). Render + two `--json` envelopes: `internal/cli/render/templates.go`
  + `envelopes.go` (registered in `jsonEnvelopes`, so `schema --json-schema` covers
  them); golden snapshots for both. `core.TemplateBody(kind, name)` renders a named
  scaffold with placeholder labels (generalized from `ScaffoldBody`).
- Tests: `internal/domain/{template,entity}_test.go`,
  `internal/cli/template_test.go`, golden cases in `integration_golden_test.go`.

**Next — step 4 (repo-local templates): the first piece that needs the repo.**
- Embed the built-ins (keep them as the floor); add a repo-local source scanning a
  `templates/` dir; repo-local overrides built-in by `<kind>/<name>`.
- Resolution must route through a store-backed port — today the cli reads
  `domain.Template*` directly (fine for built-in, but repo-local needs the fs).
  Likely shape: a `TemplateSource` port + a `core.Service` method that merges
  built-in + repo-local; `template list/show` and the create paths consume it.
- `template`'s PersistentPreRunE currently skips repo resolution ("runs anywhere").
  Make the repo *optional-but-used*: resolve best-effort so built-ins still work
  repo-less, and layer repo-local when a repo is present.
- Settle first: repo-local location (`templates/` vs `.tskflwctl/templates/`); the
  template file format (frontmatter header `name`/`description`/`kind`).

**Next — step 5 (authoring surface):** `template new <kind> <name>` / `template edit`
mirroring task/audit authoring. (`schema <kind>` now already advertises templates.)

**Post-review hardening (2026-06-22 adversarial review):**
- ✅ **Placeholder model RESOLVED — named placeholders.** Templates use `{{key}}`
  (not Printf `%s`), filled by `core.renderTemplate`; the per-kind keys+labels live on
  `Descriptor.Placeholders`. A literal `%` is now safe, a template may use any subset,
  and an undeclared `{{token}}` is left visible (guarded by
  `TestTemplates_OnlyDeclaredPlaceholders` + a no-leftover-`{{` create test). **Step 4
  must validate author/repo-local templates on load** (reject undeclared `{{tokens}}`).
- ✅ Closed-`switch kind` in the renderer retired — `TemplateBody`/`RenderLabels` are
  descriptor-driven, so a new kind lights up without a renderer edit.
- ✅ `template show --raw` added (the unrendered `{{...}}` source, for forking); core
  mutual-exclusion guard for `--body`+`--template`; `schema <kind>` advertises templates.
- ➡️ **Two review themes are now their own tasks under epic 22** (do before/with step 4):
  - `route-template-resolution-through-a-core-port` — the CLI still reads `domain.Template*`
    directly; step 4 needs a store-backed `TemplateSource` behind core.Service.
  - `template-list-output-modes-and-source-provenance` — add `source` (built-in vs
    repo-local) to the envelope + route `template list` through the shared `-o/-c/-q`.
- Still open: repo-local location (`templates/` vs `.tskflwctl/templates/`); template
  file format (frontmatter header `name`/`description`/`kind`).

## Related

- Builds directly on epic 21's M1 entity descriptor (`internal/domain/entity.go`) —
  the single-default seed this generalizes. Shipped in increments; this task tracks
  the remaining steps (4–5).

## Sweep verification (2026-09-06)

Automated weekly sweep. Every "Done" claim in the 2026-06-22 handoff was
re-verified against HEAD `84b3798` and **all of steps 1–3 still hold**:

- Step 1 — `Descriptor.Templates []NamedTemplate` + `Placeholders`
  (`internal/domain/entity.go:27–28`), `DefaultTemplate = "default"` (`:33`),
  `LookupTemplate` (`:234`), `Template` (`:255`), `TemplatesFor` (`:272`),
  `TemplateNames` (`:281`); the `security` audit template is live at `entity.go:138`
  / `auditSecurityBodyTemplate` (`:389`).
- Step 2 — `completeTemplateNames` (`internal/cli/completion.go:143`), wired on
  `audit new` at `audit.go:89`.
- Step 3 — `internal/cli/template.go` (with `--raw` at `:103`) and
  `internal/cli/render/templates.go`; `core.TemplateBody` / `RenderLabels` /
  `renderTemplate` all at `internal/core/scaffold.go:44/36/14`;
  `TestTemplates_OnlyDeclaredPlaceholders` at `internal/domain/template_test.go:75`.

### Step 4's stated prerequisite has since SHIPPED — read that block with this correction

The `## Next — step 4` block says:

> Resolution must route through a store-backed port — today the cli reads
> `domain.Template*` directly (fine for built-in, but repo-local needs the fs).
> Likely shape: a `TemplateSource` port + a `core.Service` method that merges
> built-in + repo-local

and the `➡️` note at the end of "Post-review hardening" lists
`route-template-resolution-through-a-core-port` as a prerequisite to do
"before/with step 4". **That sibling task is now `completed`** (epic 22 is at
33%, 1/3). The port exists:

- `core.TemplateSource` — `internal/core/template.go:11`, with `Templates(kind)` and
  `Lookup(kind, name)`.
- `builtinTemplates` as the default source (`template.go:23`), so `template
  list/show` still run with no planning repo.
- `WithTemplateSource(src)` on the service (`internal/core/service.go:57–60`),
  whose own doc comment says *"epic 22 wires a repo-local source"* — the swap seam
  step 4 asked for is already cut and named.

So step 4 is **no longer a CLI refactor plus a source**; it is now a source
implementation plus the wiring. The port's own comment states this intent
verbatim: *"which makes step 4 (repo-local templates) a source swap here rather
than a CLI refactor."*

Two residual direct reads of `domain.Template*` remain in the CLI, both outside
the create/list paths and both arguably fine:
`internal/cli/completion.go:135,145` (`TemplateNames` for shell completion) and
`internal/cli/schema.go:145` (`TemplatesFor`, for the repo-less `schema <kind>`
surface). Whether repo-local names must appear in completion and in `schema` output
is an unstated step-4 decision — flagged, not decided.

### This task has no acceptance criteria

`tskflwctl task ac design-a-selectable-template-library` reports **none** — the file
has a "Suggested first increment" list and a progress narrative, but no
`## Acceptance criteria` section, so `task complete` would not refuse it on any
unmet criterion. Given steps 1–3 are done and 4–5 are not, this is worth giving
real criteria before pickup. Not added by this sweep: writing acceptance criteria
is scoping, and scoping is the human's call.

The two "Still open" design decisions are genuinely still open and untouched:
repo-local location (`templates/` vs `.tskflwctl/templates/`) and the template file
format.

## Progress Log

- 2026-09-06: automated weekly sweep — steps 1–3 re-verified intact; recorded that step 4's prerequisite `TemplateSource` port has shipped (sibling task completed), leaving a source swap rather than a CLI refactor; flagged the missing acceptance-criteria section.
