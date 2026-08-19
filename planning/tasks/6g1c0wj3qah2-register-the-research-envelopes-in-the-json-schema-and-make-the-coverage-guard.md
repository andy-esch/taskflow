---
schema: 1
id: 6g1c0wj3qah2
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: research list/show --json have no $defs in schema --json-schema; the envelope coverage guard only checks registered-but-unvalidated, never declared-but-unregistered.
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [wire, contract]
created: "2026-08-18"
updated_at: "2026-08-19"
started_at: "2026-08-18"
completed_at: "2026-08-18"
---

# Register the research envelopes in the JSON schema, and make the coverage guard bidirectional

## Objective

`research list --json` and `research show --json` emit `schema_version: 1.30` but have
**no definitions in `schema --json-schema`** — a hole in the agent contract this repo is
otherwise strict about. Fix the omission, then close the class of bug so the next noun
can't repeat it.

## The defect

    $ tskflwctl schema --json-schema | jq '[.["$defs"]|keys[]|select(test("esearch"))]'
    []      # 60 envelope defs, zero research

`jsonEnvelopes` (`internal/wire/envelopes.go`) is a hand-maintained registry struct that
a single `Reflect` walks to build `$defs`. `ResearchListEnvelope` and
`ResearchShowEnvelope` were never added, so an agent validating `research list --json`
against the published schema cannot — the `$defs` key doesn't exist.

## Why no test caught it (the more valuable half)

The coverage guard in `envelopes_test.go` iterates the **registry** and requires every
registered envelope to have a validation case:

    rt := reflect.TypeOf(Envelopes())
    for i := range rt.NumField() { if !covered[def] { t.Errorf("... has no validation case") } }

It is **one-directional**. It catches "registered but unvalidated" and is blind to
"declared but unregistered" — exactly this bug. Its comment claims "a newly-added
envelope can't be silently left unvalidated", which only holds if you remember to
register it in the first place.

## Acceptance criteria

- [x] `ResearchListEnvelope` + `ResearchShowEnvelope` registered in `jsonEnvelopes`, with
      schema descriptions (re-run `internal/tools/schemacomments`).
- [x] `research list --json` and `research show --json` output validates against its own
      `$defs` entry, via a case in the existing table.
- [x] **The guard becomes bidirectional**: a test enumerates every `type <X>Envelope
      struct` declared in `internal/wire` (parse the package with `go/ast` rather than
      reflection, which cannot enumerate unreferenced types) and fails if one is absent
      from `jsonEnvelopes`. This is the durable fix — it protects the projects/routines/ADR
      nouns that come next.
- [x] Goldens regenerated (`schema_jsonschema.golden` grows the two defs).

## Open question

Does adding missing `$defs` warrant a `schema_version` bump? No emitted payload changes
shape — the schema is being corrected to describe what 1.30 already emits — so the
argument for leaving it at 1.30 is strong. Decide and record it in the wire changelog
either way.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md) — the bidirectional guard is a hardening
  change that outlives this noun
