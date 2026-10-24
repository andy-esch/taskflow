---
schema: 1
id: 6gp1ytg50tev
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Exercise the shared creation host with audit area/date and research title/date semantics, typed receipts and origin-bound recovery.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-21"
depends_on: [6gpb4y5yqmjp]
updated_at: "2026-10-23"
started_at: "2026-10-23"
completed_at: "2026-10-23"
---
# Create audits and research from shared TUI authoring forms

## Objective

Create audits and research on their respective TUI pages using the shared staging host proven by tasks and epics, without transplanting task-parent or epic lifecycle assumptions.

## Acceptance criteria

- [x] Page-local c and named palette actions expose audit and research creation using typed NewAuditParams and NewResearchParams through captured services in tea.Cmd.
- [x] Audit collects area/date and keeps core open-bucket/default-date policy; research collects title/description/tags/created date with no lifecycle and leaves collision retry to core.
- [x] Shared mechanics are tested across tasks, epics and these materially different forms; typed receipts, capability/default policy and exact destination selection remain per kind.
- [x] Validation, absent/read-only/pathless capabilities, pending/unknown/committed outcomes, session drift, stale widget results and watcher landing retain input and never replay or retarget writes.
- [x] Independent durable metadata/scaffold expectations, real keyboard/paste flows, narrow rendering and normal code/planning gates pass with owning docs updated.

## Out of scope

Thread graph/membership creation, universal creation API, bulk/template workflows and long Markdown editing.

## Implementation checkpoint — 2026-10-23

- [Audit](../../internal/tui/audit_create.go) and
  [research](../../internal/tui/research_create.go) forms are concrete consumers of
  the value-owned host. Their page-local `c` and named palette actions use the
  existing controls/presentation; the root model/session routes are unchanged.
- Dates stay blank until core defaulting at submission. Audit always starts open;
  research description/tags are optional, has no lifecycle, and retains core's
  date-based identity and bounded collision retry. No universal creation port or
  UI-owned schema/default policy was added. Both create methods already belong to
  the required Store; optional capability absence here means local navigation,
  not an invented optional audit/research creation port.
- Typed receipts control exact-ID landing and retained non-replayable recovery.
  Submission captures copied inputs and the originating Service. Same-ID incoming
  records cannot steal late outcomes or cause incoming reads/writes; issued origin
  work is not cancelled or retargeted on workspace activation.
- [Independent durable tests](../../internal/tui/document_create_test.go) cover
  metadata/scaffolds, blank/default and explicit dates, validation/repair, real
  read-only persistence, pathless opaque IDs, core collision handling, actual stale
  and duplicate results, watcher/filter/bucket landing, narrow modes and running
  Bubble Tea keyboard/paste creation. Shared mechanical, non-empty-buffer and
  Darwin native-provider tests now cover all four kinds. Malformed preview,
  unconfirmed and partial receipts are adapter-defense probes, not claims about
  actual filesystem producers.
- Compiler-valid Go overlays removing the audit origin-result guard or discarding
  the research created-date input fail the intended origin-isolation and durable
  backdating regressions. Source files were not mutated by these probes.
- README, runtime key help and the
  [owning architecture guide](../../docs/architecture/tui.md#audit-and-research-creation)
  describe the new surfaces. No CLI metadata, core, wire or persisted schema changed.

Validation passed: `just build`, full `just test` race suite, `just lint` (zero
issues), focused race checks for the final receipt/origin/backdating cases,
architecture navigation, planning/link/audit lint and whitespace checks. The build
emitted the existing nonfatal module stat-cache permission warning. The first
test pass corrected fixture expectations for scaffold headings, newly created
entity directories and consumed terminal-output chunks; no product failure was
silently waived.

Implementation and both reviews are complete for PR/integration handoff. Thread creation remains
the separately scoped successor, with its initial-membership design checkpoint.

## Independent review checkpoint

Prepared on 2026-10-23:

- [Codex](../audits/6gpgy3n31042-2026-10-23-tui-audit-research-creation-implementation-codex.md):
  completed with no findings; owner accepted the bounded evidence and closed the
  audit after confirming the guarded delivery receipt. Independent producer,
  same-ID incoming-space, explicit recovery/retry, literal date/default,
  pathless-selection and actual native-provider probes passed. Removing the
  recovery-Enter displacement guard exposed incoming reads; restoration passed.
  Full project gates were attributed to the owner, not claimed as independent.
- [Antigravity](../audits/6gpgy3n9hp5z-2026-10-23-tui-audit-research-creation-implementation-antigravity.md):
  completed with no findings; guarded delivery confirmed and audit closed.
  Independent keyboard/default/backdating, resize/native-input and live-program
  probes passed; audit-date and text-only-enum mutations failed as intended and
  restored green. The owner note corrects receipt source references and bounds
  claims about layered entry guards, already-issued writes and input ownership.

Both reviews used independent sandboxes, cross-lens probes, compiler-valid
mutations and a second systemic pass. Their bounded evidence warranted no code
fix or new follow-up. Review reports retain their captured-baseline attestations;
owner dispositions distinguish confirmed delivery from historical transfer-pending
report text. Neither review certifies every asynchronous terminal interleaving.

## Related

- [Additional creation integration](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md).
- [TUI architecture](../../docs/architecture/tui.md#staged-task-creation).
