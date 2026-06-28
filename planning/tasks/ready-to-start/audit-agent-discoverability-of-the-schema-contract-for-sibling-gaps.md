---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'Review found epic_fields missing from schema --json + a brittle envelope count-guard (both fixed). Sweep for siblings: audit-field discoverability, other magic-count guards, missing no-drift tests.'
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
updated_at: "2026-06-28"
started_at: "2026-06-26"
---
## Objective

The 2026-06-25 adversarial review found two agent-discoverability/contract gaps (epic_fields absent from `schema --json`; a literal-count envelope-validation guard that silently skipped envelopes) — both fixed. Sweep for siblings.

## Acceptance criteria
- [ ] `schema --json` exposes every gate an agent needs (audit-settable fields, finding statuses), not just task/epic fields.
- [ ] Grep for other brittle 'magic count' / literal-list guards that can silently drift.
- [ ] Every authoring-field registry has a no-drift test (tasks + epics do; audits?).

## Notes
Hardening, not urgent — the review closed the worst cases.

**Status check 2026-06-28.** The two named findings (epic_fields missing from `schema --json`; the brittle envelope count-guard) are fixed — `schema audit` exists and `schema --json-schema` covers the envelopes. Open work is the sibling SWEEP only: audit-field discoverability in the JSON schema, other magic-count guards, and the missing no-drift tests.
