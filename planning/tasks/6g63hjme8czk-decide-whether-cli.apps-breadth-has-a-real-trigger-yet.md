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
updated_at: "2026-10-04"
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
