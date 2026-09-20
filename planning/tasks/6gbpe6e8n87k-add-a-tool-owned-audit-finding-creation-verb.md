---
schema: 1
id: 6gbpe6e8n87k
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Create canonical audit findings through one validated atomic verb that allocates codes, preserves section order, and can establish the managed candidate projection.
effort: 4-8 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, findings, cli, agents, robustness]
created: "2026-09-19"
depends_on: [6g3ag8py12y9]
updated_at: "2026-09-19"
---
# Add a tool-owned audit finding creation verb

## Context

`audit finding` can safely update a finding only after that finding already parses. Creation still relies on raw `audit append`, where an author must reproduce the header and metadata grammar and a fresh scaffold places the appended finding after its final `## Candidate tasks` section. Near-miss recognition makes mistakes loud, but it does not give agents the canonical write path that acceptance criteria and candidate rows already have.

Build the remaining writer without turning audits into a hidden database or migrating historical prose. The new verb should allocate identity, construct one canonical block, and use the existing audit body CAS so creation is indivisible. On `candidate-tasks:v1` audits it should be able to create the finding and its optional candidate row in the same transform.

## Acceptance criteria

- [ ] A dedicated non-interactive verb creates one canonical finding block without requiring the caller to type its finding code or Markdown grammar.
- [ ] The tool validates the severity band and finding metadata vocabularies, allocates the next collision-free code in that band deterministically, and reports the chosen code.
- [ ] A new finding is placed in the audit findings area before a trailing Candidate tasks section when one exists; arbitrary historical body content and whitespace outside the insertion point remain unchanged.
- [ ] An optional one-line candidate value creates the new finding and its `candidate-tasks:v1` row in the same atomic CAS write; legacy candidate sections remain untouched and receive an explanatory refusal if that option is requested.
- [ ] Dry-run and JSON/human receipts expose the planned or created finding identity and preserve the repository error/exit-code contract.
- [ ] Concurrent audit edits cannot clobber either the finding or candidate row; retry behavior, duplicate allocation, malformed audits, and section-boundary cases have focused tests.
- [ ] Schema guidance, CLI docs, and the audit routines use the writer for new findings; raw `audit append` remains available for unstructured sections.

## Related

- Completes the writer half intentionally left out of `make-near-miss-audit-findings-loud-and-self-repairing` (`6g72wf39pyhb`).
- Builds on the managed candidate convention in `decide-the-candidate-list-convention-and-make-the-tool-own-it` (`6g3ag8py12y9`).
- Keep audit findings as Markdown-first prose; this is a safe authoring path, not a storage-model migration.