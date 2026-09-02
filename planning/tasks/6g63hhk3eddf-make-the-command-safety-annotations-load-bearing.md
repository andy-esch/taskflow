---
schema: 1
id: 6g63hhk3eddf
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: '76 commands carry safety annotations that nothing reads: expose them in schema --json and make them verify something'
effort: 3-5 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, agents, schema, architecture]
created: "2026-09-02"
---
# Make the command safety annotations load-bearing

## Objective

Every CLI command carries `Annotations{"safety": "read-only"|"mutating"}` — 76 of
them today (41 read-only, 35 mutating) — and **nothing reads them**. There is no
consumer anywhere in `internal/`, no test asserting they exist, and `schema` does
not emit them.

`planning/research/6f9menr01t1n-tskflwctl-command-spec.md:278` marks the tagging
work complete, and line 248 states what it was for: an agent should introspect
safety tags rather than scrape `--help` prose. That consumer was never built.

A hand-maintained convention with no reader can only rot — a new command that
omits the tag, or tags itself wrong, is invisible today. Give the tag a consumer,
then let it enforce something. The end state worth aiming at is that the tag
*does* work rather than describing it: a `read-only` command that can reach a
mutating path should be a test failure, not a code-review catch.

## Acceptance criteria

- [ ] `schema --json` emits the command surface with each command's safety tag, and a golden pins the output
- [ ] A test asserts every registered command — including hidden and deprecated ones — carries a recognized `safety` value
- [ ] The tag gates or verifies something rather than only describing it: a `read-only` command that reaches a mutating service path fails a test
- [ ] Decide and record whether `--dry-run` applicability should derive from the tag rather than being restated per command
- [ ] Agent-facing guidance (`schema`, and the CLAUDE.md triage section) tells an agent the tag exists and is machine-readable

## Out of scope

- Reclassifying any command's read-only/mutating value — audit values only where a new test disagrees with reality
- A general command-metadata registry: this is one field gaining one consumer, not a new descriptor layer
- Surfacing the tag in human `--help` prose

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- `planning/research/6f9menr01t1n-tskflwctl-command-spec.md` — command safety tags, the original intent
