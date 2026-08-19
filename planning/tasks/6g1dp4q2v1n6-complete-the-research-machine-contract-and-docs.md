---
schema: 1
id: 6g1dp4q2v1n6
status: in-progress
epic: 28-first-class-entities-new-planning-nouns
description: schema --json lacks research_fields, the known-field list advertises unsettable keys, the JSON Schema falsely claims research edit --json, six prose lists and the README omit research.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [contract, docs]
created: "2026-08-18"
started_at: "2026-08-19"
updated_at: "2026-08-19"
---

# Complete the research machine contract and docs

## Objective

`research` ships without the contract surface and documentation its siblings have. None of
these is a bug in behaviour; each is the tool describing itself wrongly, which for an
agent-facing tool is its own class of defect.

## Findings

**1. `schema --json` has no `research_fields`.** The contract exposes `task_fields` and
`epic_fields` only. The precedent is explicit and on the record: the wire changelog bumped
to 1.15 to add `epic_fields` *because* `epic set` landed. 1.31 landed `research set`
without the analogue, so an agent must trigger an error and scrape prose to learn which
keys are settable — exactly what `schema` exists to prevent.

**2. `knownResearchFields` seeds the tool-managed stamps, and it is dead logic.** It seeds
`{schema, id, updated_at}`, three of the four fields `ProtectedResearchField` ALWAYS
rejects — and since the protected check runs FIRST in `SetResearchFields`, the seeding is
unreachable. Its only effect is to pollute the advertised list:

    unknown research field "tier" (known: created, description, id, schema, tags, updated_at)

Four of six advertised "known" fields are unsettable. Task's registry deliberately excludes
`id`/`schema`; epic's excludes all three. Drop the seed.

**3. The shipped JSON Schema falsely claims `research edit --json` emits the mutation
envelope.** The doc comment on `ResearchMutationEnvelope` names `research edit`, that
comment is harvested into `schema_comments.json`, and it ships in `schema --json-schema`.
`newResearchEditCmd` has no `app.JSON` branch. It also contradicts the wire changelog,
which correctly says set/append. The MISSING envelope is defensible (task/audit/epic `edit`
all behave identically); the false claim is not.

**4. Six hand-maintained prose lists that omit research**, all shipped:
- `cli/schema.go` — `Use: "schema [task|epic|audit]"`, two lines above a registry-driven
  `ValidArgs`. Propagates into `docs/cli/tskflwctl_schema.md`.
- `cli/render/schema_render.go` — `schema`'s human footer says `<task|epic|audit>` while
  the SAME command's `kinds` array emits `["task","epic","audit","research"]`.
- `cli/template.go` — `--kind` help `"(task|epic|audit)"`, with a registry-driven completion
  func beside it.
- `cli/lint.go` — "`--fix` repairs tasks/audits" and "backfill missing task/audit ids",
  both stale since research joined `fixDir`.
- `store/fix.go` ("walks every task, epic, and audit file") and `cli/audit.go` ("`lint`
  (which covers tasks/epics)").
- `wire/envelopes.go` — "CreatedEnvelope is `task/epic/audit new --json`"; `research new`
  emits it. Harvested and shipped.

**5. README documents zero research commands** — command blocks for task/epic/audit, with
research appearing only as a directory. Its opening line still says "task/epic/audit
files"; `CLAUDE.md` still says `schema task|epic|audit`.

**6. (low) `ResearchShowHuman` re-implements `fieldPrinter`** verbatim with a different
width instead of calling it, and drops the relative-date suffix task's `updated` carries.
It matches `AuditShowHuman`'s hand-rolled copy, making it the third copy of an extracted
helper.

## Acceptance criteria

- [x] `research_fields` in the schema contract, shaped like `epic_fields`, with a wire
      changelog entry and a version bump if the envelope shape changes.
- [x] `knownResearchFields` no longer seeds protected stamps; the advertised list contains
      only settable keys.
- [x] The envelope doc comment names only the commands that actually emit it; regenerate
      `schema_comments.json` and the goldens.
- [x] All six prose lists corrected, ideally sourced from `domain.SchemaKinds()` where the
      string is assembled at runtime so they cannot drift again.
- [x] README gains a research command block; its opening line and CLAUDE.md's `schema`
      line include research.
- [ ] `ResearchShowHuman` uses `fieldPrinter`.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Found by an independent adversarial architecture/contract review, 2026-08-18
