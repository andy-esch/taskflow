---
schema: 1
id: 6gbpe6e8n87k
status: completed
epic: 20-cli-ux-and-ergonomics
description: Create canonical audit findings through one validated atomic verb that allocates codes, preserves section order, and can establish the managed candidate projection.
effort: 4-8 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [audit, findings, cli, agents, robustness]
created: "2026-09-19"
depends_on: [6g3ag8py12y9]
updated_at: "2026-09-20"
started_at: "2026-09-20"
audit_sources: [planning/audits/6gbxf5yz9gja-2026-09-20-audit-finding-creation-verb-implementation-claude.md, planning/audits/6gbxf5yz94xv-2026-09-20-audit-finding-creation-verb-implementation-antigravity.md]
completed_at: "2026-09-20"
---
# Add a tool-owned audit finding creation verb

## Context

`audit finding` can safely update a finding only after that finding already parses. Creation still relies on raw `audit append`, where an author must reproduce the header and metadata grammar and a fresh scaffold places the appended finding after its final `## Candidate tasks` section. Near-miss recognition makes mistakes loud, but it does not give agents the canonical write path that acceptance criteria and candidate rows already have.

Build the remaining writer without turning audits into a hidden database or migrating historical prose. The new verb should allocate identity, construct one canonical block, and use the existing audit body CAS so creation is indivisible. On `candidate-tasks:v1` audits it should be able to create the finding and its optional candidate row in the same transform.

## Acceptance criteria

- [x] A dedicated non-interactive verb creates one canonical finding block without requiring the caller to type its finding code or Markdown grammar.
- [x] The tool validates the severity band and finding metadata vocabularies, allocates the next collision-free code in that band deterministically, and reports the chosen code.
- [x] A new finding is placed in the audit findings area before a trailing Candidate tasks section when one exists; arbitrary historical body content and whitespace outside the insertion point remain unchanged.
- [x] An optional one-line candidate value creates the new finding and its `candidate-tasks:v1` row in the same atomic CAS write; legacy candidate sections remain untouched and receive an explanatory refusal if that option is requested.
- [x] Dry-run and JSON/human receipts expose the planned or created finding identity and preserve the repository error/exit-code contract.
- [x] Concurrent audit edits cannot clobber either the finding or candidate row; retry behavior, duplicate allocation, malformed audits, and section-boundary cases have focused tests.
- [x] Schema guidance, CLI docs, and the audit routines use the writer for new findings; raw `audit append` remains available for unstructured sections.

## Related

- Completes the writer half intentionally left out of `make-near-miss-audit-findings-loud-and-self-repairing` (`6g72wf39pyhb`).
- Builds on the managed candidate convention in `decide-the-candidate-list-convention-and-make-the-tool-own-it` (`6g3ag8py12y9`).
- Keep audit findings as Markdown-first prose; this is a safe authoring path, not a storage-model migration.

## Design decision (2026-09-20)

The creation surface will be `audit finding new <audit> <title> --band H|M|L`. It is a child of the existing singular finding command, so existing `audit finding <audit> <code>` edits remain backward compatible. New findings always begin open. The next code is the highest existing numeric suffix in that band plus one; gaps are not recycled because finding codes are durable audit-local identities.

Optional `--file`, `--component`, `--effort`, `--urgency`, `--body`/`--body-file`, `--recommendation`, and `--candidate` inputs let the domain renderer own all structural Markdown. Band, effort, and urgency use explicit domain vocabularies; free-form fields cannot inject headings or reserved lifecycle blocks. The pure domain transform locates or creates one real Findings section, core recomputes allocation on every CAS retry, and the existing audit body transform remains the only durable write boundary.

Creation returns a dedicated compact finding receipt rather than repeating the full audit body. Its additive JSON envelope will advance the monotonic machine-contract revision. A supplied candidate is added only to a managed v1 section inside the same transform; legacy or missing candidate sections fail before any write.

Creation is valid only while the audit is open. The audit metadata and body are supplied to the domain callback from the same store snapshot guarded by the content CAS, so a concurrent close cannot race an earlier bucket check and leave a non-open audit with a newly open finding. Existing malformed findings or managed candidate projections likewise block creation; diagnostics route through `audit lint` for inspection and `audit edit` for explicit repair.

## Implementation evidence (2026-09-20)

Implemented the audit finding new command across domain, core, store, wire, CLI, generated reference docs, schema guidance, and both scheduled audit routines. Codes are allocated monotonically inside the retried CAS callback; the callback now receives audit metadata and body from the same guarded snapshot, which blocks creation in closed or deferred audits without a check-then-write race. Creation refuses malformed existing findings, near-miss headers, ambiguous sections, drifted managed candidate projections, structural metadata injection, and legacy candidate writes.

Focused domain/core/store/wire/CLI tests pass, including real-filesystem concurrent allocation, candidate atomicity, dry-run, compact JSON, whitespace preservation, malformed inputs, and non-open audit refusal. The full race-enabled suite and golangci-lint pass. A disposable planning repo round-trip created H1 with a managed candidate row, previewed H2 via JSON without writing it, and passed ordinary lint.

## Adversarial review closeout (2026-09-20)

Claude reported four findings; Antigravity completed a substantive clean pass. All four were addressed. The guarded-snapshot contract now has a real-filesystem close/defer race test that was mutation-verified against a fabricated-open adapter. LF and CRLF creation now preserve identical vertical spacing. Rendered location and planning metadata again match the canonical templates.

Candidate corruption remains fail-closed by design. The documented audit append path now inserts narrative before a trailing managed Candidate tasks section, preventing the tool from poisoning its own projection, with store and CLI regression coverage. Pre-existing malformed projections name audit lint for diagnosis and audit edit for explicit recovery. Both review audits are closed with their dispositions recorded.
