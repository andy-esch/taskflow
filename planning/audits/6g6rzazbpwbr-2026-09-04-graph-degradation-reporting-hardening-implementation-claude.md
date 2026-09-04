---
schema: 1
id: 6g6rzazbpwbr
bucket: open
area: graph-degradation-reporting-hardening-implementation-claude
date: "2026-09-04"
---

# Audit: graph degradation reporting hardening — claude — 2026-09-04

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> Required second pass: after completing the checklist, review again as a devil's advocate for systemic failure modes. Challenge shared abstractions, adapters that only happen to be filesystem-backed, test helpers that conceal divergent snapshots, and error paths that appear noisy but still report a false clean state. Prefer one demonstrated systemic issue over several speculative findings, and settle every challenged pattern with hostile evidence.
>
> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only source. Before inspecting implementation, running tests or generators, or making mutation probes, create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement whose `.git` metadata points back to the shared checkout. At completion, copy back only this assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Create an isolated clone whose working tree includes the exact current
source contents, including staged, unstaged, untracked, and deleted files:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g6rzazbpwbr-2026-09-04-graph-degradation-reporting-hardening-implementation-claude.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review-claude.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
test -d "$SANDBOX/.git"
cd "$SANDBOX"

git add -A
git -c user.name='Taskflow Review Sandbox' \
  -c user.email='review-sandbox@invalid' \
  -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null \
  commit --allow-empty --no-verify -m 'chore: capture review sandbox baseline'
```

The checkpoint is the only commit you may create. Confirm `git rev-parse --git-dir` resolves inside
`$SANDBOX`. Perform all inspection, builds, tests,
formatting, generation, fixtures, mutations, and report editing there. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`; inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then transfer
only the audit, guarded against concurrent source edits:

```sh
test "$(git -C "$SOURCE_ROOT" hash-object "$SOURCE_AUDIT")" = "$SOURCE_AUDIT_BLOB" || {
  printf 'source audit changed; do not overwrite it; preserve sandbox at %s\n' "$SANDBOX" >&2
  exit 1
}
TRANSFER="$(mktemp "${SOURCE_AUDIT}.review-transfer.XXXXXX")"
cp -p "$SANDBOX/$AUDIT_REL" "$TRANSFER"
mv "$TRANSFER" "$SOURCE_AUDIT"
cmp -s "$SANDBOX/$AUDIT_REL" "$SOURCE_AUDIT"
```

Do not copy anything else back. Leave the sandbox in place and report its path until the
implementation owner confirms receipt. If the hash guard fails, report the conflict and sandbox
path instead of resolving it in the shared checkout.

## Review brief

Perform an independent adversarial implementation review of
[report graph degradation in status and lint](../tasks/6g697mp8s4tx-report-graph-degradation-in-status-and-lint.md)
on branch `feat/graph-degradation-reporting-hardening`, based on `main` at commit `92f9893`.
The implementation is in commits `0e8a25c`, `2ed59b6`, `1ab44d0`, and `198d585`.
Review the complete `git diff 92f9893...HEAD`; use this assigned audit as the review brief. Judge the work
against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), and the task acceptance criteria.

Assume the implementation can be systemically wrong despite green tests. It moves repository-wide
task-graph health into the ordinary summary path and then publishes it through CLI text, JSON, the
TUI dashboard, and the cross-space Atlas. It also relies on existing lint ownership for legacy
dependency diagnostics. Re-derive those contracts from production code; comments, golden churn,
and filesystem-backed test doubles are not proof.

Do not edit implementation or other planning files. The newly tracked
[`preserve-portable-load-diagnostics-in-board-and-status`](../tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
follow-up acknowledges that Board and Summary currently collapse neutral load problems to
`domain.FileProblem`. Do not duplicate that as a finding unless you demonstrate a broader defect,
incorrect scope/sequencing, or present data loss beyond the recorded debt.

## Intended contract to challenge

- `core.Service.Summary` loads task records exactly once through its injected `TaskGraphSource`.
  Counts, in-progress tasks, unreadable-task problems, graph health, and graph detail all describe
  that one read; the legacy store is not a second task authority.
- `SummaryStore` owns only non-task summary reads. `PlanningSummarySource` explicitly composes the
  non-task summary port and `TaskGraphSource`, so cross-space/local composition cannot silently omit
  graph health. Alternate and future adapters can satisfy the capability without filesystem paths.
- Healthy, degraded, and broken retain the same core meanings used by mutation guards, graph reads,
  Thread projections, and Board. Non-healthy detail is actionable and truthful; healthy output does
  not acquire a spurious warning.
- Human `status`, `status --all`, dashboard, and Atlas make non-healthy graphs conspicuous without
  turning a read-only status command into a failing process or independently deriving graph state.
- Plain `status --json` exposes optional top-level `graph`; `status --all --json` exposes optional
  `spaces[].summary.graph`. Non-healthy values carry health and detail through the shared wire type.
  The schema bump to 1.60 and all generated/golden artifacts accurately describe compatibility.
- Ordinary lint reports graph-invalid states through one established owner per defect. Missing and
  ambiguous legacy references are visible once through grouped legacy diagnostics, remain advisory
  under the accepted policy, and cannot produce a false clean message. Strict graph mutations still
  refuse every degraded/broken graph independently of lint exit policy.
- Existing status counts, audit/epic problems, cross-space failure retention, TUI attention counts,
  other JSON commands, and non-filesystem portability do not regress.
- Planning/docs state only what shipped. The portable diagnostic follow-up is appropriately scoped,
  sequenced after its lint prerequisite, and dogfooded in the production Threads graph.

## Mandatory evidence floor

A `ready` verdict is not credible unless the report contains all of the following:

1. A consumer and composition inventory for `Summary`, `SummaryStore`, `PlanningSummarySource`,
   `TaskGraphSource`, `Service.Summary`, `summarize`, `OpenPlanningStore`, `SummaryJSON`,
   `GraphHealthJSON`, `RenderSummary`, dashboard summary rendering, Atlas attention/detail, and every
   schema encoder/decoder or version declaration affected by 1.60. Classify each production use;
   grep counts alone are not an inventory.
2. End-to-end throwaway-space probes for healthy, degraded, and broken graphs, including unreadable
   task input, missing canonical dependency, ambiguous legacy slug reference, duplicate/missing task
   identity, invalid status, and a cycle. Compare `status`, `status --json`, `status --all`, TUI, and
   `lint`; also attempt an ordinary dependency mutation against each invalid state.
3. A split-authority adapter probe where `TaskGraphSource` intentionally returns a different task
   set from any legacy `ListTasks` capability. Prove counts, in-progress rows, problems, health, and
   detail all use the graph-source snapshot. Repeat with a pathless in-memory or remote-shaped
   adapter and with errors from each composed capability.
4. Multi-space probes containing healthy, degraded, broken, unavailable, and stale-last-good spaces.
   Check Atlas attention, detail text, sorting/selection stability, refresh recovery, and whether one
   space's failure can hide or relabel another space's graph verdict.
5. JSON/schema compatibility evidence: exact plain versus `--all` field placement, healthy omission,
   non-healthy inclusion, stable health vocabulary, detail omission rules, schema 1.60 declaration,
   JSON Schema output, and representative unrelated command envelopes. Check both pretty and compact
   output and any decode/round-trip consumers found in the inventory.
6. Lint output/exit evidence for every `TaskGraphProblemCode`, with special attention to missing and
   ambiguous legacy references, unreadable input, duplicate messages, false `✓ no issues found`, and
   advisory versus blocking policy. Trace each problem to exactly one presentation owner.
7. At least these temporary, restored mutation probes, naming the test that kills each mutation:
   - make `Summary` count tasks from the store while health comes from `TaskGraphSource`;
   - remove either half of the `PlanningSummarySource` composite capability;
   - suppress degraded or broken output in one of CLI, dashboard, or Atlas;
   - put `graph` at the wrong JSON nesting level or emit an empty healthy object;
   - stop bumping the schema or leave one representative golden envelope on 1.59;
   - print a legacy dependency defect twice or let it end with a false clean message; and
   - weaken the normal mutation guard because lint treats resolved legacy debt as advisory.
   A surviving mutation is a coverage finding even if production currently looks correct.
8. Repeated focused tests under `-race`, an uncached full `go test -race ./...`, static analysis,
   generated-doc drift, module tidiness, planning/audit lint, and `git diff --check`, with exact
   commands, Go version, durations, and cached/uncached distinction. If resource limits prevent an
   item, record that rather than silently substituting weaker evidence.

## Required adversarial angles

1. **Snapshot coherence and split authority.** Look for a second task scan, a hidden `ListTasks`
   dependency, or task data reconstructed from graph problems. Challenge mutation between task,
   epic, and audit reads and distinguish unavoidable cross-entity skew from a false claim that task
   graph fields share a snapshot.
2. **Port strictness and portability.** Inspect compile-time interface satisfaction and every
   `OpenPlanningStore` implementation/test double. Challenge nil or partially implemented graph
   sources, filesystem assumptions, pathless diagnostics, remote latency/errors, and whether the
   composite interface is too broad or accidentally narrower than the production need.
3. **Health semantic drift.** Compare Summary with Board, mutation guards, graph queries, and Thread
   projections for every problem code. Look for zero-value health, nil reads, detail chosen from an
   unstable order, degraded reported as broken (or vice versa), and health/detail contradictions.
4. **Presentation honesty.** Confirm all human surfaces are noisy enough without double counting a
   graph plus its unreadable file, hiding the warning below routine content, or claiming a repair
   action unsupported by an adapter. Check no-color, narrow TUI, empty repository, and recovery.
5. **Wire compatibility.** Treat 1.60 as a public contract. Challenge `omitempty`, top-level versus
   nested placement, shared object reuse, generated comments, consumers that switch on exact schema,
   and the broad unrelated-golden churn. Decide whether the versioning policy was followed rather
   than assuming a bump makes the change safe.
6. **Lint ownership and policy.** Re-derive why each graph code is omitted from or added to domain
   lint. Attack the grouped legacy path with multiple owners/tasks, mixed missing and ambiguous refs,
   resolved legacy refs, unreadable documents, and simultaneous cycle/identity faults. Distinguish
   advisory exit behavior from missing visibility.
7. **Error containment and refresh.** Force graph load, non-task summary, per-space open, and render
   failures. Look for stale graph warnings surviving recovery, last-good data being erased too soon,
   partial success presented as healthy, or one source error masking more actionable evidence.
8. **Regression blast radius.** Exercise ordinary status counts, epics/audits, dashboard focus,
   Atlas attention, all registered spaces, other JSON envelopes, and direct core callers. Search for
   mocks whose easy filesystem conformance masks a production capability break.
9. **Planning truthfulness.** Compare task closeout, ADR amendments, README/CLI docs, architecture
   claims, follow-up scope, dependency direction, and Thread placement with actual code and CLI
   behavior. Flag materially stronger claims, not stylistic omissions.
10. **Systemic second pass.** After the checklist, deliberately seek a shared helper, zero value,
    interface composition, projection/presentation split, golden-update mechanism, or diagnostic
    ownership convention that could hide an entire class of defects. Demonstrate it or record the
    evidence that settles it.

## Validation and restoration

Run proportionate validation and hostile scratch-space probes inside the reviewer sandbox. Restore
every probe and generated artifact to the sandbox checkpoint. Do not install dependencies, push,
edit implementation permanently, create follow-up tasks, change finding statuses, close this audit,
or edit the other reviewer's audit. At finish, sandbox `git status --short` must show only this
assigned audit before its guarded one-file transfer back to the source checkout.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/HEAD/worktree state, runtime, and exact validation results;
- a compact data-flow and port-boundary re-derivation;
- findings grouped by severity, each with stable code and `**Status:** open` in the heading;
- acceptance-criteria traceability, hostile-evidence, and restored-mutation ledgers;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but the evidence and mutation ledgers are still required.
Do not pre-resolve findings; the implementation owner will triage them with
`tskflwctl audit finding`.

## Reviewer report

_Reviewer: replace this line and append your report here._

## Findings

<!-- Use the exact heading grammar from the assignment note. Omit this section if there are no findings. -->

## Candidate tasks

<!-- Do not create tasks. Suggest one command per open finding only when deferral is warranted. -->
