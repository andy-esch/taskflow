---
schema: 1
id: 6gbn4g1bb7wr
status: completed
epic: 20-cli-ux-and-ergonomics
description: Publish success, fallback, abort, and reserved transition codes alongside classified domain failures from one test-enforced machine contract.
effort: XS
tier: 2
priority: high
autonomy_level: 3
tags: [cli, json, agents, contract]
created: "2026-09-19"
updated_at: "2026-09-21"
started_at: "2026-09-20"
audit_sources: [planning/audits/6gcb5n1q8w97-2026-09-21-complete-cli-exit-code-taxonomy-implementation-claude.md, planning/audits/6gcb5p5jh8ct-2026-09-21-complete-cli-exit-code-taxonomy-implementation-antigravity.md]
completed_at: "2026-09-21"
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

- [x] `schema --json` exposes every active process exit the CLI deliberately returns, including
      success, unclassified fallback, and interactive abort.
- [x] The code-12 reservation is represented truthfully or its omission is explicitly justified
      where agents discover exit behavior.
- [x] One source owns the published code/name/meaning metadata, while domain classification still
      uses only classifiable failures.
- [x] Focused process-level tests prove representative commands produce the published codes and
      JSON error names; schema and golden tests reject drift.
- [x] The correctly classified machine-contract revision, generated docs/schema, and agent guidance are
      updated; focused tests, full tests, lint, planning lint, and diff checks pass.

## Out of scope

- Inventing new failure classes or changing existing command exit behavior.
- Mapping CLI failures to HTTP status codes for a future server adapter.
- Building the executable command manifest owned by the adjacent command-safety task.

## Related

- Audit [Machine contract, M1](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md)
- ADR [Monotonic JSON machine-contract revisions](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
- Task [Make command safety annotations load-bearing](6g63hhk3eddf-make-the-command-safety-annotations-load-bearing.md)

## Implementation evidence (2026-09-21)

Implemented one CLI-owned process taxonomy that publishes stable code, name, meaning, and active/reserved state while keeping domain failure classification separate. The existing 10/11/13/14 rows remain the stable prefix; 0 success, 1 generic error, 130 prompt abort, and retired code 12 append. Code 12 is explicitly reserved and cannot be returned by the current ExitCode paths.

Schema revision 1.70 is classified NOT ADDITIVE because the published closed vocabulary gains members, even though the historical four-row prefix remains stable. The schema JSON and generated Draft 2020-12 schema expose the complete taxonomy and constrain state to active or reserved. README, CLAUDE.md, architecture guidance, the command-spec research record, and generated CLI help now point agents to schema --json.

Focused unit tests pin registry uniqueness, exact order/content, domain mappings, fallback names, and reservation semantics. A real-binary test compares published rows with observed exits and JSON error names for success, generic failure, not-found, validation, ambiguity, and conflict. The race-enabled suite, golangci-lint, planning lint, audit lint, and git diff --check pass.

## Adversarial review closeout (2026-09-21)

Claude found five valid gaps; all were fixed in scope. Revision 1.70 is now truthfully NOT ADDITIVE under ADR-0008, emitted outcomes pass through an active-only registry guard, filesystem failure is pinned to 1/error at unit and process boundaries, exit-row wording is precise, the human table is numerically sorted and tested, and agent guidance uses the real wire vocabulary. Antigravity reported no findings. Both audits are closed.

Validation: the full race-enabled Go suite passes, golangci-lint reports zero issues, generated CLI docs and schema comments reproduce exactly, planning and both audit lints pass, and git diff --check is clean.
