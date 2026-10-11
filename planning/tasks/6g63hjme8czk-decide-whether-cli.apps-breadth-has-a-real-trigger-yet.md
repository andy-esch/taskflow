---
schema: 1
id: 6g63hjme8czk
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Decide whether current cli.App breadth has an evidenced split trigger after composition extraction, or record a bounded not-yet decision.
effort: 2-3 hours (decision, not refactor)
tier: 4
priority: low
autonomy_level: 3
tags: [cli, architecture, dx]
created: "2026-09-02"
updated_at: "2026-10-11"
audited: "2026-10-11"
---
# Decide whether cli.App's breadth has a real trigger yet

## Objective

Assess the current `cli.App` after application-port migration and composition extraction
(PRs #275/#277), not the old field-count analogy to the TUI root model.

`App` still combines invocation/IO state, resolved presentation, and named application
services. It now receives `ports.Bindings` and a lazy planning opener; concrete construction
and framework launch wiring live in `internal/appwiring`. `Fixer` and `Linter` fields are
retired; `Layout` is an explicit local watcher capability. Ordinary planning opening is
lazy after flags are parsed, and completion deliberately defers opening until its target
flags are known. There is no longer a whole-CLI-package composition exemption.

Those changes address real boundary defects, but do not settle whether `App`'s remaining
breadth imposes a concrete cost. Measure current command dependencies, focused-test setup,
and incomplete-binding failures. Compare an IO/invocation split, per-command contracts,
and doing nothing without assuming a field count or TUI precedent makes a split necessary.

The deliverable remains a decision, not a refactor. Either name an evidenced trigger and
scope the work, or record that the breadth is accepted with a trigger for reconsideration.

## Acceptance criteria

- [ ] The concrete costs are written down with evidence, not asserted: what a new command must know about `App` today, what a test must construct, and whether any command has reached for a field outside its concern
- [ ] Candidate seams are named and compared, including the do-nothing option — at minimum an IO/invocation vs service-container split, and a per-command narrow interface
- [ ] A decision is recorded either way, with the trigger that would reopen it if the answer is "not yet"
- [ ] If the answer is "act", the follow-up task is filed with a bounded scope; if "not yet", no code changes land under this task
- [ ] The outcome is reflected wherever `App` is described so the next reviewer reads the decision rather than re-deriving it

## Out of scope

- Refactoring `App` under this task — this is the decision, and any change is a separately filed follow-up
- Reopening the merged composition boundary or weakening its executable controller rules
- The TUI's root `Model`, which has its own extraction history

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- [Composition isolation](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md), merged in PR #277

## Scope refresh (2026-10-04)

Old field counts, retired ports, and whole-package-exemption assumptions were removed after
the merged boundary work. The decision criteria remain unchecked: extraction is not evidence
that this separate breadth question has been answered. This is not a required adapter-neutral
Thread member and does not expand that Thread's closeout scope.

## Sweep refresh (2026-10-11)

2026-10-11: automated weekly sweep — no drift; the 2026-10-04 scope refresh above still
describes `App` accurately field for field, and the decision is still unrecorded.

**Verified accurate (2026-10-11)** against `internal/cli/root.go:27`:
- `App` still receives `bindings ports.Bindings` and a lazy `openPlanning` opener, with a
  `compositionErr` that keeps operational commands off a partial bundle while leaving
  metadata-only commands usable.
- `Fixer` and `Linter` are gone — no reference survives anywhere in `internal/cli`.
- `Layout` is still present as an explicit local watcher capability, carrying exactly the
  comment the scope refresh claims ("planning repair, body-link checking, and completion
  are application use cases on `Svc`").
- `CompletionService` still defers composition until Cobra has parsed the completed
  command's flags, as the body states.
- There is still no whole-CLI-package composition exemption. `.golangci.yml:120-129`
  scopes `internal/cli` to two named local exceptions, `init.go` and `workspace.go`,
  matching `docs/ARCHITECTURE.md`'s "named local exceptions" wording.
- The prerequisite this task was sequenced behind is finished:
  `isolate-cli-composition-wiring-and-enforce-controller-boundaries` (PR #277) is
  `status: completed`, and `task blockers` reports this task `gate: clear`,
  `eligible: true`. Nothing is holding it except the decision itself.

**Measurement still outstanding.** The first acceptance criterion asks for concrete costs
"with evidence, not asserted", and that evidence has not been gathered — so the sweep did
not tick it. For whoever picks this up, the shape worth measuring is visible from the
struct: `App` carries 5 distinct named application services (`Svc`, `SpaceSvc`,
`SpaceOverviewSvc`, `WorkspaceSvc`, `ConfigSvc`) alongside invocation/IO state, resolved
presentation, and three deferred-error fields (`compositionErr`, `userCfgErr`,
`commandSafety`). Counting which commands actually reach outside their own concern is the
missing evidence, and this sweep did not do it — recorded here so the next reader knows the
criterion is genuinely open rather than overlooked.
