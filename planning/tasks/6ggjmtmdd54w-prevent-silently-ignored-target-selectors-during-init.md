---
schema: 1
id: 6ggjmtmdd54w
status: next-up
epic: 21-code-quality-architecture-hardening
description: Reject or explicitly honor init target selectors so -C or --space cannot silently initialize the caller's working directory.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, safety, contracts]
created: "2026-10-04"
depends_on: [6g1xp8qymz1m]
updated_at: "2026-10-04"
---
# Prevent silently ignored target selectors during init

## Objective

Make initialization target selection explicit and fail closed rather than accepting a selector that is silently ignored.

## Evidence

During 2026-10-04 throwaway-space dogfood for task 6ggdkztshmzz, running `tskflwctl -C /tmp/new-space init --taskflow-root planning --no-register` from the real checkout repaired that checkout rather than initializing `/tmp/new-space`. It removed only the empty legacy `planning/projects/.gitkeep`; the owner restored it exactly. Re-running with `init --path /tmp/new-space` initialized the intended target.

`internal/cli/init.go` uses `styleOnlyPreRun` and `filepath.Abs(path)`, with `--path` defaulting to `.`. Global `-C` remains accepted but does not select this bootstrap target. Investigate explicit `--space` and `TSKFLW_SPACE` too; those variants have not yet been reproduced.

## Scope

- Decide and document whether bootstrap supports `-C` or rejects it in favor of `--path`; never silently ignore an explicit selector.
- Define conflict rules for `--path`, `-C`, and explicit/environment space selection without requiring an existing planning repo to scaffold a new one.
- Preserve config repair, pointer mode, explicit no-register behavior, and mutation authorization. No broad config redesign.
- Audit other style-only commands for the same explicit-selector ambiguity; track independent mutations separately.

## Acceptance criteria

- [ ] A two-directory test proves init cannot mutate caller cwd after an explicit different target selector is supplied.
- [ ] Unsupported or conflicting explicit target combinations fail before any write, with actionable guidance.
- [ ] Accepted target combinations work for fresh scaffold and existing-config repair; pointer mode remains covered.
- [ ] Environment space defaults are deliberate and documented; no hidden registry mutation occurs in no-register tests.
- [ ] Human help, generated docs, and JSON/exit behavior match the chosen policy; normal validation passes.

## Sequencing

Independent CLI safety followup, not a blocker for the adapter-neutral refactor closure or task 6ggdkztshmzz. Pick it up through the CLI contract-hardening Thread.

## Related

- [Core impact/recovery task](6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md)
- [Configuration lifecycle](6g1xp8qymz1m-consolidate-the-configuration-lifecycle-under-one-config-command-hub.md)
