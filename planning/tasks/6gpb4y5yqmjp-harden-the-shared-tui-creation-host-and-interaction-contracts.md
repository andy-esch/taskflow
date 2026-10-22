---
schema: 1
id: 6gpb4y5yqmjp
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Consolidate creation lifetime and result routing, make finite widget delivery testable, and unify visibility/layout bounds before more entity forms.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [tui, authoring, architecture, testing]
created: "2026-10-22"
depends_on: [6gp1ytfvm652]
updated_at: "2026-10-22"
started_at: "2026-10-22"
completed_at: "2026-10-22"
---
# Harden the shared TUI creation host and interaction contracts

## Objective

Consolidate the wiring now proven by task and epic creation before audit/research
and Thread forms extend it. Reduce kind-specific root-model edits without turning
UI staging into a universal entity schema or changing portable mutation contracts.

## Scope

- Give a value-owned creation host one active-kind/lifetime boundary, monotonic
  generations and one displacement route. Keep concrete widget state and typed
  requests/receipts per kind; do not store a mutable form pointer across Model copies.
- Centralize creation-result routing, including the intentional late-origin-result
  exception to session rejection. A departed result may enrich its retained
  recovery copy, never navigate/reload/write the incoming space. Catalog and widget
  messages do not gain that exception.
- Improve the interaction harness so finite widget results are delivered through
  Model.Update rather than silently discarded. Bound settling, explicitly stop
  cursor maintenance, and exercise safe clipboard results without the user clipboard.
- Share mechanical contract scenarios across both forms while keeping expected
  typed requests, durable metadata and receipt policy independently asserted.
- Use one layout sizing contract for supported input, resize notice and modal
  bounds, accounting for shell padding exactly once.

## Acceptance criteria

- [x] One creation host owns the active draft, generation allocation, result dispatch and workspace displacement; task/epic controllers retain typed requests/receipts and Model copies retain independent draft text.
- [x] Common routing rejects stale, cross-kind and unrelated-session results; late origin create outcomes can update only retained recovery, preserving duplicate-submit and no-replay protections through executed commands.
- [x] The bounded interaction harness delivers finite widget/application results, excludes only explicit maintenance and fails usefully on unbounded work; regressions demonstrate that dropping finite results is detected.
- [x] Shared task/epic contract cases cover opening, cancellation, pending/recovery, stale input and source displacement, with independent durable assertions and existing real-program/native clipboard coverage retained.
- [x] Rendering and input visibility use one supported-size rule; boundary and resize tests prove that a visible form accepts input and a resize notice cannot edit a hidden draft.
- [x] Build, full race suite, lint, planning/link/audit lint and owning architecture guidance pass; targeted compiler-valid mutations fail the intended regressions and restored code is green.

## Sequencing

After the completed epic/shared-form slice 6gp1ytfvm652. Both audit/research creation
6gp1ytg50tev and guarded Thread creation 6gp1ytgea8dt follow this foundation, not
each other. The older editor/palette/command-bar paste task 6gp2qmse4prt stays
independent; do not make all TUI authoring wait for it.

## Boundaries

No new entity form, universal creation port, form-schema registry, generic receipt,
core API/wire change, dependency upgrade or broad TUI navigation rewrite. Preserve
task parent/default policy, epic explicit defaults, origin service binding,
adapter-established identity and optional local capabilities. Any genuinely wider
finding needs a scoped follow-up rather than silently expanding this task.

## Implementation checkpoint

- [creationHost](../../internal/tui/creation_host.go) now owns concrete value
  slots, active-kind selection, opening/action wiring, monotonic generations and
  displacement. The shell stores one host and delegates page/palette actions;
  form pointers are borrowed synchronously, never retained as copied host state.
- Typed create results implement a private routing seam. The host checks kind,
  pending attempt and original session before each per-kind controller interprets
  its unchanged receipt policy. One shell exception delivers late origin outcomes
  for recovery; catalog/widget results do not acquire that privilege.
- [The shared harness](../../internal/tui/staged_creation_helpers_test.go) now
  delivers finite widget results through Model.Update, including real opaque
  native clipboard results in the Darwin fixture. Only explicit cursor maintenance
  is excluded; a command-work budget reports unbounded follow-up rather than
  silently accepting it. This is synchronous test settling, not a replacement for
  running-program delivery-order tests or a wall-clock timeout.
- [Shared mechanical contracts](../../internal/tui/creation_host_test.go) cover
  active-draft replacement refusal, clean cross-kind generations, dirty Keep/Discard,
  prior-copy independence, displaced recovery, result guards, finite input delivery,
  loop-budget failure and supported-size boundaries. Existing typed durable,
  portable, no-replay, origin and running-program tests remain independent.
- One modal-size helper accounts for shell gutter; the same creation fit rule
  selects resize rendering and input authority. Supported dimensions stay unchanged.
- Compiler-valid Go overlays removing attempt matching or restoring the old
  drop-all-widget helper fail the intended stale-attempt and finite-delivery
  regressions for both kinds. Production source was never mutated by these probes.
  Normal routing/boundary/runtime regressions also passed three race repetitions.

No core API, wire, persisted entity or CLI metadata changed. The existing
older-input paste follow-up remains independent. Implementation and review are
complete for PR/integration handoff; start the successor creation slices only
after this foundation lands.

Validation: `just build`, full `just test` race suite, `just lint`, planning/link
lint, audit lint and whitespace checks pass. The build emitted the existing
nonfatal Go module stat-cache permission warning. No generated CLI docs changed.

The [Codex implementation review](../audits/6gpb8hr5rjk1-2026-10-22-tui-creation-host-contracts-implementation-codex.md)
completed on 2026-10-22 with no new findings. The source helper confirmed delivery
from the independent sandbox before owner triage. The review independently tested
both real origin-create flows against same-ID incoming records, counted incoming
reads, compared bytes/modes and executed recovery/duplicate-result commands. It
also covered validation retry, non-default durable fields, native-provider/runtime
input and boundary rendering, and killed then restored both selected mutations.
Full project gates were explicitly inherited, not claimed as independent runs.
The disclosed preliminary filename-only source discovery did not inspect code or
write source; substantive review and all probes stayed isolated. The owner accepted
the bounded evidence and closed the audit. No code fixes or new follow-ups were
warranted; no claim is made about unimplemented entity forms or every interleaving.

## Evidence and starting points

- `internal/tui/creation_host.go`: active draft, opening/action wiring, result matching and displacement.
- `staged_creation.go`, `staged_creation_view.go`: mechanics, input ownership and shared sizing.
- `model.go`, `session.go`: shell delegation and the narrow late-origin-result exception.
- `staged_creation_helpers_test.go`: bounded finite-result delivery; the old task-only helper dropped all widget results.
- `staged_creation_test.go`, `epic_create_test.go`: reusable mechanical scenarios and typed evidence.
- [TUI architecture](../../docs/architecture/tui.md#epic-creation-and-shared-staging).
- [Prior Codex review](../audits/6gp2rf2gyp1a-2026-10-21-tui-epic-creation-shared-staging-implementation-codex.md): native messages and shallow widget-buffer copies escaped overly narrow tests; both product fixes already landed.
