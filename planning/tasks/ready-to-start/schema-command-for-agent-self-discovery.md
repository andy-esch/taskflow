---
status: ready-to-start
epic: 17-pm-go-cli
description: 'DRAFT: revive the descoped schema command - one --json call emitting statuses, epic enum, field registry, exit/error codes, schema_version'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [cli, agents, json, draft]
created: "2026-06-12"
---
# `schema` command for agent self-discovery

> 🚧 **DRAFT — not yet integrated into the overall plan.** Filed from the
> 2026-06-12 CLI-design discussion. This *revives a deliberately descoped
> item* — see the conflict note — so it needs an explicit planning yes
> before work starts.

## Objective

Let an agent configure itself in ONE call instead of parsing `--help` prose:
`tskflwctl schema --json` emits the machine contract —
- task statuses (+ which are active) and the epic-status enum,
- the known-field registry with types (`domain/fields.go` exists precisely
  for this: int/list/known sets),
- exit codes and `--json` error codes (the D9 vocabulary),
- the current `schema_version` and the envelope inventory.

Human mode prints the same as a readable table. Nearly free now: every list
it would emit already lives in `domain` as data.

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

## Acceptance criteria (draft)

- [ ] Planning conflicts above resolved; task de-drafted.
- [ ] One `--json` call yields statuses, epic enum, field registry with
      types, exit/error codes, schema_version (strict-decode test).
- [ ] The emitted sets are DERIVED from domain (sync-guard test), never
      hand-copied lists.

## Related

- Epic [[17-pm-go-cli]] · [[2026-06-12-pending-decisions]] (D7/D9) ·
  `internal/domain/fields.go`.