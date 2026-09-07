---
schema: 1
id: 6g7srp3py9fe
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Reconcile the audit's required transfer result with a protocol that must finalize the report before the guarded transfer runs.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 4
tags: [reviews, agents, scripts, safety]
created: "2026-09-07"
---

# Make adversarial review transfer attestations self-finalizing

## Objective

Make the isolated-review protocol's persisted attestation truthful and achievable. The current
generated brief requires the audit itself to say that transfer succeeded, but the reviewer must
finish the audit before invoking `transfer`; the 2026-09-07 audit therefore arrived with
`transfer=pending` despite a successful guarded transfer.

## Scope

- Decide whether transfer success belongs inside the audit, in a helper-owned receipt, or in the
  implementation owner's disposition evidence.
- Update the helper, generated brief, and scripts guidance so the required evidence can be produced
  without manually editing the transferred audit or weakening its source-hash guard.
- Preserve the general standalone-shell-tool boundary and one-deliverable transfer model.
- Add a disposable-repository regression for the chosen end-to-end attestation flow.

## Acceptance criteria

- [ ] A compliant reviewer can follow the generated instructions literally and produce every
      required attestation field in the place the contract names.
- [ ] The persisted evidence distinguishes successful verification from successful transfer and
      cannot claim completion before the transfer occurs.
- [ ] Source-deliverable drift, unrelated sandbox changes, and failed transfers remain fail-closed.
- [ ] The disposable-repository smoke test exercises the complete report/verify/transfer/receipt
      sequence and guards against a persisted `transfer=pending` handoff being treated as complete.
- [ ] `scripts/README.md` and generated adversarial-review briefs state one consistent evidence
      contract.

## Out of scope

- Moving the helper into the taskflow Go library or CLI.
- General multi-file synchronization, sandbox deletion, or remote review orchestration.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Predecessor [make adversarial review sandbox isolation executable](6g7grjz5ma57-make-adversarial-review-sandbox-isolation-executable.md)
- Triggering review [audit ID collision hardening implementation](../audits/6g7sb0wnv5t5-2026-09-07-audit-id-collision-hardening-implementation-claude.md)
