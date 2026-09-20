---
schema: 1
id: 6gbn4g1bb7wr
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Publish success, fallback, abort, and reserved transition codes alongside classified domain failures from one test-enforced machine contract.
effort: XS
tier: 2
priority: high
autonomy_level: 3
tags: [cli, json, agents, contract]
created: "2026-09-19"
---

# Publish the complete CLI exit-code taxonomy

## Objective

Make `schema --json` describe every process exit an agent can intentionally branch on, not only the
four domain-classified failures. Preserve the existing error-classification table while publishing
success, fallback, interactive abort, and the retired transition reservation from one explicit,
test-enforced taxonomy.

## Scope

- Publish codes `0` (`ok`), `1` (`error`), `10` (`not-found`), `11` (`validation`), `13`
  (`ambiguous`), `14` (`conflict`), and `130` (`aborted`) with semantics that match executable paths.
- Decide whether reserved code `12` belongs in the same list with an explicit retired/reserved
  state, or in separate reservation metadata; do not imply the current binary emits it.
- Keep domain error classification separate from the larger published process taxonomy so rows
  without a domain class do not distort `errors.Is` behavior.
- Advance and classify the machine-contract revision, update generated artifacts, and pin the
  published order, names, and actual exit behavior.

## Acceptance criteria

- [ ] `schema --json` exposes every active process exit the CLI deliberately returns, including
      success, unclassified fallback, and interactive abort.
- [ ] The code-12 reservation is represented truthfully or its omission is explicitly justified
      where agents discover exit behavior.
- [ ] One source owns the published code/name/meaning metadata, while domain classification still
      uses only classifiable failures.
- [ ] Focused process-level tests prove representative commands produce the published codes and
      JSON error names; schema and golden tests reject drift.
- [ ] The additive machine-contract revision, generated docs/schema, and agent guidance are
      updated; focused tests, full tests, lint, planning lint, and diff checks pass.

## Out of scope

- Inventing new failure classes or changing existing command exit behavior.
- Mapping CLI failures to HTTP status codes for a future server adapter.
- Building the executable command manifest owned by the adjacent command-safety task.

## Related

- Audit [Machine contract, M1](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md)
- ADR [Monotonic JSON machine-contract revisions](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
- Task [Make command safety annotations load-bearing](6g63hhk3eddf-make-the-command-safety-annotations-load-bearing.md)
