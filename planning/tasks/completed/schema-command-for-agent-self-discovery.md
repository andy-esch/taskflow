---
status: completed
epic: 17-pm-go-cli
description: 'DRAFT: schema command — agent self-discovery + authoring guidance (sections, field descriptions); shape needs design thought before finalizing'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, agents, json, draft]
created: "2026-06-12"
updated_at: "2026-06-16"
completed_at: "2026-06-16"
---
# `schema` command for agent self-discovery

> 🚧 **DRAFT — not yet integrated into the overall plan.** Filed from the
> 2026-06-12 CLI-design discussion. This *revives a deliberately descoped
> item* — see the conflict note — so it needs an explicit planning yes
> before work starts.

## Resolution (shipped 2026-06-16)

Built v1. Decisions made for the open questions:

- **Reversal approved** (the descope is reversed); kept the name **`schema`**.
- **One command:** `schema` = global contract (A); `schema <task|epic|audit>` =
  authoring guidance (B).
- **Everything derived from domain**, no hand-copied lists: field *types* from
  the existing type maps (new `domain.FieldType` / `KnownTaskFieldNames`); the
  body template + section names from the live `new` scaffold (new
  `core.ScaffoldBody`). Only per-field description/example are hand-authored —
  pinned to the real field set by a sync test (`TestTaskAuthoringFieldsMatchRegistry`).
- **Did NOT enrich `fields.go` into a struct registry** — kept the load-bearing
  bool maps; descriptions live in a separate `domain.FieldDoc` table instead.
- **Runs with no planning repo** (overrides the root `resolve()`, like `version`)
  — the bundled-self-describe use case. Tested.
- `schema_version` bumped 1.3 → 1.4 (additive envelopes).

**Deferred (not in v1):** the per-repo `.tskflwctl.toml` guidance override — it
depends on the unsettled config relative/absolute path handling, so it's its own
follow-up. The "envelope inventory" was also left out.

Implemented across `domain/schema.go`, `core/scaffold.go`, `cli/schema.go`,
`render.go`; tests in `domain/schema_test.go`, `cli/schema_test.go`. Suite + lint
+ vet green.

## Objective

Make the tool self-describing in ONE call instead of parsing `--help` prose.
Two complementary halves:

**A. Machine contract (the original scope).** `tskflwctl schema --json` emits —
- task statuses (+ which are active) and the epic-status enum,
- the known-field registry with types (`domain/fields.go` exists precisely
  for this: int/list/known sets),
- exit codes and `--json` error codes (the D9 vocabulary),
- the current `schema_version` and the envelope inventory.

Human mode prints the same as a readable table. Nearly free now: every list
it would emit already lives in `domain` as data.

**B. Authoring guidance (folded in 2026-06-14).** The other half of
self-discovery: how to *compose* a well-formed task/epic/audit, aimed mainly at
an **AI drafting documents** — especially when `tskflwctl` is bundled into
another repo with no CLAUDE.md to lean on (the strongest reason to build this).
Per-kind: `tskflwctl schema task` / `schema epic` / `schema audit` would emit
the body **section template** (Objective / Acceptance criteria / Out of scope /
Related — already encoded in the `task new` scaffold at `core/service.go`),
each **frontmatter field with a description + example** (not just its type),
and any **conventions** (e.g. one-line `description`, required tags). The
contract half (A) tells an agent *what values are valid*; the authoring half
(B) tells it *how to write the document*.

> ⚠️ **B needs design thought before finalizing — do not implement off this
> sketch.** Open questions below.

## ⚠️ Conflicts to resolve before starting

- **`schema` was explicitly descoped** when the port task closed (see the
  closure note in [[port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec]],
  completed) — this draft proposes reversing that. Planning should confirm
  the reversal rather than inherit it silently.
- The old pm `schema` had different semantics (frontmatter schema dump);
  decide whether this is the same command grown up or a new name
  (`contract`? `capabilities`?) to avoid false continuity.
- Output is itself part of the versioned JSON contract (D7: one global
  version) — adding it is a minor bump and its shape should be strict-decode
  tested like the other envelopes.

## ⚠️ Open design questions for part B (authoring guidance)

These need resolving before B is implementable — it is **not** a clean derive
like A:

- **One command or two?** Does `schema` carry both contract (A) and authoring
  (B), or does B get its own verb (`guide`/`explain`/`template`)? A bare
  `schema` for the global contract + `schema <kind>` for per-kind authoring is
  one option.
- **Where does the prose live (single source of truth)?** Field *descriptions*
  and *examples* don't exist yet — the registry is type-only. Options: enrich
  `domain/fields.go` from bool-maps into a richer registry (type + description +
  example), which also upgrades A's output; and/or `go:embed` a small authoring
  doc. Section names should be **derived from the actual `task new` scaffold**,
  not hand-copied, so guidance and the generated skeleton can't disagree.
- **How opinionated?** Best-practices prose drifts and is taste-laden — decide
  how much normative guidance ships vs. just structure + field semantics.
- **Per-repo override (`.tskflwctl.toml`).** Each repo has different conventions,
  so the built-in guidance must be overridable: a config key (e.g.
  `guidance_dir`/`authoring_docs`) pointing at repo-specific guidance the command
  surfaces *instead of / layered over* the tool defaults. Open: replace vs. merge
  with the built-in baseline; what format the override is (markdown the command
  prints? structured?); and — crucially — the path is resolved by the same
  config machinery whose relative/absolute handling is itself in question (see
  the config-path note flagged 2026-06-14), so settle that first.
- **Per-entity coverage:** task first; epic/audit have their own (smaller)
  scaffolds and field sets — confirm the same command shape extends cleanly.
- **Interaction with in-flight schema changes:** if the readiness axis
  ([[task-readiness-state-draft-vs-finalized-in-frontmatter]]) or the scaffold
  version key ([[scaffold-schema-version-key-and-domain-level-audit-finding-counter]])
  land, B's field list and sections must reflect them — sequence accordingly.

## Acceptance criteria (draft)

- [ ] Planning conflicts above resolved; **both** the part-A reversal question
      and the part-B open design questions answered; task de-drafted.
- [ ] **A:** one `--json` call yields statuses, epic enum, field registry with
      types, exit/error codes, schema_version (strict-decode test).
- [ ] **B:** `schema <kind>` (or the chosen verb) emits the body section
      template + per-field description/example + conventions, in `--json` and
      human modes.
- [ ] The emitted sets are DERIVED from domain (sync-guard test) — statuses,
      fields, AND section names — never hand-copied lists.

## Related

- Epic [[17-pm-go-cli]] · [[2026-06-12-pending-decisions]] (D7/D9) ·
  `internal/domain/fields.go` (field registry to enrich) ·
  `internal/core/service.go` (the `task new` body scaffold = section source).
- Authoring half (B) interacts with
  [[scaffold-schema-version-key-and-domain-level-audit-finding-counter]] and
  [[task-readiness-state-draft-vs-finalized-in-frontmatter]] (both add
  fields/sections the guidance must reflect).