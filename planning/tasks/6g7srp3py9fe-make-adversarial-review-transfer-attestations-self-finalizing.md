---
schema: 1
id: 6g7srp3py9fe
status: completed
epic: 21-code-quality-architecture-hardening
description: Reconcile the audit's required transfer result with a protocol that must finalize the report before the guarded transfer runs.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 4
tags: [reviews, agents, scripts, safety]
created: "2026-09-07"
updated_at: "2026-10-12"
started_at: "2026-10-12"
completed_at: "2026-10-12"
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

- [x] A compliant reviewer can follow the generated instructions literally and produce every
      required attestation field in the place the contract names.
- [x] The persisted evidence distinguishes successful verification from successful transfer and
      cannot claim completion before the transfer occurs.
- [x] Source-deliverable drift, unrelated sandbox changes, and failed transfers remain fail-closed.
- [x] The disposable-repository smoke test exercises the complete report/verify/transfer/receipt
      sequence and guards against a persisted `transfer=pending` handoff being treated as complete.
- [x] `scripts/README.md` and generated adversarial-review briefs state one consistent evidence
      contract.

## Out of scope

- Moving the helper into the taskflow Go library or CLI.
- General multi-file synchronization, sandbox deletion, or remote review orchestration.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Predecessor [make adversarial review sandbox isolation executable](6g7grjz5ma57-make-adversarial-review-sandbox-isolation-executable.md)
- Triggering review [audit ID collision hardening implementation](../audits/6g7sb0wnv5t5-2026-09-07-audit-id-collision-hardening-implementation-claude.md)
- Repeated evidence: [final source-boundary review](../audits/6gg38rnydjbm-2026-10-03-final-semantic-entity-source-boundary-checkpoint-antigravity.md)
  arrived byte-identical to its independent sandbox but still persisted `transfer=pending`.

## Thread and repeated evidence (2026-10-12)

This task now belongs to
[review-workspace safety and delivery evidence](../threads/6gk06gktr1q5-harden-isolated-review-workspace-safety-and-delivery-evidence.md).
The [documentation ownership review](../audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md)
again finalized its attestation before transfer and deferred the result to a handoff message;
the owner confirmed byte-for-byte report delivery. That proves receipt for this run, not that the
persisted attestation contract has been redesigned. The sibling
[push-guard task](6gk06e2g21hv-disable-pushes-from-isolated-review-workspaces.md) is independent;
neither needs an artificial dependency on the other or on the docs Thread.

## Delivery and verification (2026-10-12)

Decision: transfer success belongs in a helper-owned receipt, not inside the audit. A report must
be final before `transfer` runs, so it can truthfully attest only verification.

- `verify` prints `verify=passed` and `transfer=pending`. After the transferred bytes are confirmed
  in the source, `transfer` atomically writes a receipt under the sandbox's private state and
  prints it. The receipt carries the verification fields plus `transfer=succeeded`, the transferred
  blob, the source path, and a UTC timestamp.
- The new `receipt` subcommand fails until a transfer has succeeded. Afterwards it also fails when
  the source or sandbox no longer holds the transferred bytes, and on success it reprints the
  receipt with `delivery=confirmed`. Once a receipt exists, `verify` and `transfer` refuse and point
  to `receipt`. Refused or failed transfers record nothing.
- The [generated brief](../../scripts/prepare-adversarial-review-audits.sh) now instructs:
  verify → copy the attestation into the report → verify again → transfer → paste the printed
  receipt into the handoff. The owner runs `receipt` before triage or cleanup.
  [Scripts guidance](../../scripts/README.md) states the same split contract and the limit that
  reviewer-writable sandbox state detects accidents, not a forged handoff.
- [Disposable-repository tests](../../internal/tools/reviewworkspace/review_workspace_test.go) run
  the whole report/verify/transfer/receipt sequence. They show that a report quoting the pending
  attestation still has no confirmable receipt, and that the receipt stops vouching once the owner
  edits the source. Drift, an unrelated workspace change, a reviewer commit, and an unwritable
  source directory each fail without a receipt. A generator dry run pins the brief's literal
  helper-call order (`create verify verify transfer receipt`) and rejects the old in-report
  transfer requirement.
- Mutations: writing a receipt at `verify`, skipping the source-match check, recording the receipt
  before the source write, and restoring `main`'s generator each fail the intended test.
  ShellCheck, vet, lint, and the package tests pass.
