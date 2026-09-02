---
schema: 1
id: 6g63hjme8czk
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: 'App is ~30 fields across flags, presentation, and seven services: name the trigger for a split or record that it is accepted'
effort: 2-3 hours (decision, not refactor)
tier: 4
priority: low
autonomy_level: 3
tags: [cli, architecture, dx]
created: "2026-09-02"
---
# Decide whether cli.App's breadth has a real trigger yet

## Objective

`cli.App` is ~30 fields spanning three distinct concerns: invocation flags
(`JSON`, `DryRun`, `Color`, `NoInput`, `Chdir`, `Space`…), resolved presentation
state (`Style`, `Th`, `Gate`, `Prompt`, `User`, `userCfgErr`), and seven injected
services (`Svc`, `SpaceSvc`, `SpaceOverviewSvc`, `WorkspaceSvc`, `ConfigSvc`, plus
the narrow `Fixer`/`Layout`/`Linter` ports). Every command receives all of it.

It is the CLI's counterpart to the TUI's ~50-field root `Model` — the same
breadth, arrived at the same way, and the TUI's answer (extract per-concern
sub-models: `spaceSession`, `entityTab`, the modal stack) is the established one.
The mainstream CLI equivalent is a split between an IO/invocation context and a
service container, as in kubectl's `IOStreams` + factory.

But this epic's stated principle is trigger-scoped consolidation **instead of
speculative rewrites**, and no concrete pain has been recorded here: DI is
genuinely global-free, `PersistentPreRunE` resolves everything once, and tests
construct `App` without complaint. So the deliverable is a decision, not a
refactor. Either name the trigger and scope the work, or record that the breadth
is accepted and stop re-litigating it in reviews.

## Acceptance criteria

- [ ] The concrete costs are written down with evidence, not asserted: what a new command must know about `App` today, what a test must construct, and whether any command has reached for a field outside its concern
- [ ] Candidate seams are named and compared, including the do-nothing option — at minimum an IO/invocation vs service-container split, and a per-command narrow interface
- [ ] A decision is recorded either way, with the trigger that would reopen it if the answer is "not yet"
- [ ] If the answer is "act", the follow-up task is filed with a bounded scope; if "not yet", no code changes land under this task
- [ ] The outcome is reflected wherever `App` is described so the next reviewer reads the decision rather than re-deriving it

## Out of scope

- Refactoring `App` under this task — this is the decision, and any change is a separately filed follow-up
- The composition-root exception in `docs/ARCHITECTURE.md`; that edge is already classified and is not what this examines
- The TUI's root `Model`, which has its own extraction history

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
