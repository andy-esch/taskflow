---
schema: 1
id: 6gppr2k0exeb
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Expose each entity's -c selectors (name, description, aliases, string type) and each command's emitted envelope in the schema contract, so agents stop learning them from errors.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 3
tags: [schema, agents, contracts]
created: "2026-10-23"
audit_sources: [planning/audits/6gjzzvdy5jnw-2026-10-12-arch-machine-contract.md, planning/audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md]
updated_at: "2026-10-23"
---
# Publish selector registries and command output envelopes in schema --json

## Objective

ADR-0008 §2 says the revision covers projected list JSON, including selector names and value
semantics, yet the six selector registries exist in machine form only in a test golden; agents
learn them by provoking an error. Separately, `schema --json` lists commands and the JSON Schema
lists envelopes, but nothing joins them. Publish both from their existing owners, without a second
transcription.

## Acceptance criteria

- [ ] `schema --json` carries a per-entity selector registry derived from `render.Specs(...)`: name, description, aliases, and an explicit string value type.
- [ ] `TestGolden_ProjectionContract` asserts the published contract rather than re-deriving it.
- [ ] Each published command names the envelope definition its `--json` output validates against, set where its renderer is wired; commands without a typed envelope say so explicitly.
- [ ] A test fails if a runnable command or a projection registry is missing from the published contract.
- [ ] The additive revision is classified per ADR-0008, and the generated CLI reference is refreshed.
- [ ] A recorded decision on whether per-command argument and flag inventory belongs in the contract or stays with `--help`.

## Out of scope

- Vocabulary annotations in the generated JSON Schema (separate ADR-0008 decision task).

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Audit [2026-10-12 machine contract](../audits/6gjzzvdy5jnw-2026-10-12-arch-machine-contract.md), findings M1 and L1
- Audit [2026-07-24 AI-agent CLI ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md), finding M3 (paths and safety shipped in [6g63hhk3eddf](6g63hhk3eddf-make-the-command-safety-annotations-load-bearing.md))
