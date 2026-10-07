---
schema: 1
id: 6ggjmtmdd54w
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Reject or explicitly honor init target selectors so -C or --space cannot silently initialize the caller's working directory.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, safety, contracts]
created: "2026-10-04"
depends_on: [6g1xp8qymz1m]
updated_at: "2026-10-06"
started_at: "2026-10-06"
audit_sources: [planning/audits/6gh1frgyz2fh-2026-10-06-adapter-hygiene.md]
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

- [x] A two-directory test proves init cannot mutate caller cwd after an explicit different target selector is supplied.
- [x] Unsupported or conflicting explicit target combinations fail before any write, with actionable guidance.
- [x] Accepted target combinations work for fresh scaffold and existing-config repair; pointer mode remains covered.
- [x] Environment space defaults are deliberate and documented; no hidden registry mutation occurs in no-register tests.
- [x] Human help, generated docs, and JSON/exit behavior match the chosen policy; normal validation passes.

## Sequencing

Independent CLI safety followup, not a blocker for the adapter-neutral refactor closure or task 6ggdkztshmzz. Pick it up through the CLI contract-hardening Thread.

## Related

- [Core impact/recovery task](6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md)
- [Configuration lifecycle](6g1xp8qymz1m-consolidate-the-configuration-lifecycle-under-one-config-command-hub.md)

## Implemented selector policy (2026-10-06)

Init now honors global -C without requiring planning discovery, and --space/environment
defaults use the ordinary exact registered-entry selector. Explicit --path or -C
overrides TSKFLW_SPACE. --path/-C/--space are mutually exclusive explicit targets;
blank selectors fail validation before writes. Stale or missing space selections
never fall back to caller cwd.

Two-directory snapshots prove caller/registry remain untouched for selected scaffold,
repair, pointer mode, and invalid combinations. Pointer-relative planning paths
resolve from the selected directory. JSON receipts and real-process smoke tests pass.
Existing init test helpers injected -C implicitly; they now distinguish real cwd from
the selector, avoiding the same blind spot in future tests.

The only other style-only namespace is space: it intentionally operates on registry
entries identified by explicit subcommand arguments, not on an implicit planning tree.
No independent ignored-planning-target mutation was found there. Help and generated
reference document the policy. Race suite, lint, build, and planning lint passed.

Machine-contract revision 1.82 is NOT ADDITIVE: formerly accepted combinations now refuse. Migration: choose exactly one explicit --path, -C, or --space; --path/-C override the ambient space default.

## External review handoff (2026-10-06)

Prepared two independent audits with no findings:
[Codex](../audits/6gh86jxj4sve-2026-10-06-shared-write-and-audit-safety-implementation-codex.md)
and [Antigravity](../audits/6gh86jxtyx5k-2026-10-06-shared-write-and-audit-safety-implementation-antigravity.md).

Codex leads init/audit/machine-contract checks; Antigravity leads YAML preservation.
Both cross-check the other lens and must use independent dirty-state-capturing sandboxes,
bounded compiler-valid mutation evidence, and guarded one-audit transfer. Review has
not run; implementation remains in-progress pending owner triage.

## Review triage (2026-10-07)

Codex M1 was confirmed and fixed: scaffold-repair advice now includes the shell-quoted
resolved -C target and preserves flag/environment registration opt-out. A real-process
regression executes the emitted command verbatim from another cwd for -C, --path,
registered --space, and environment selection. Quote/expansion-like path contents stay
literal; only the selected scaffold changes, not caller or registry. Full race suite,
lint, build, generated comparisons, and planning lint pass. Codex findings are settled;
Antigravity remains pending and this task stays in-progress until final review closeout.

## Additional baseline audit evidence (2026-10-06)

Reinforced by audit 2026-10-06-adapter-hygiene: H1. Both selector variants this task records as not yet reproduced now reproduce deterministically — `-C` plus `--space` bypasses the mutual-exclusion guard in `startDir()`, and an unknown `--space` is swallowed; each scaffolds the caller's cwd at exit 0. The finding also notes that `git -C` resolves later relative path options against the `-C` target, which is a ready precedent for the open "support `-C` or reject it" decision: `init --path` would resolve relative to `-C` rather than the process cwd.

This is pre-implementation corroboration. The implemented policy above honors -C
while rejecting combined explicit selectors, rather than adopting git's combined
-C/relative-path convention. H1 is already tracked by this task; it is not a new
unresolved implementation finding.
