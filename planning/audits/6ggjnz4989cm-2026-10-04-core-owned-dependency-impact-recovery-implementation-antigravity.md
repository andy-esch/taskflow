---
schema: 1
id: 6ggjnz4989cm
bucket: open
area: core-owned-dependency-impact-recovery-implementation-antigravity
date: "2026-10-04"
---
# Audit: Core-owned dependency impact and recovery implementation — antigravity — 2026-10-04

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Challenge the core-owned dependency impact/recovery implementation, not the entire adapter-neutral
refactor. Find demonstrated data-safety, semantic ownership, compatibility, or recovery regressions.
Do not award readiness based on a green suite alone. Do not implement fixes or settle findings.

## Review target

- Task: `planning/tasks/6ggdkztshmzz-make-dependency-impact-and-recovery-semantics-core-owned.md`.
- Thread: `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.
- Source audit: `planning/audits/6ga93gvffkpc-2026-09-15-adapter-hygiene.md`, M1/L1.
- Baseline HEAD is merged PR #280 (`289b0ba`); the implementation is the captured uncommitted
  handoff overlay, including newly added tests. Do not inspect only the source branch's commits.
- Production: `internal/core/{dependency_operations,service_task,task_lifecycle,thread_mutation}.go`,
  `internal/wire/{dependency,wire}.go`, `internal/cli/render/{dependency,render}.go`, and
  `internal/cli/task_dependency.go`. Review README, architecture, generated CLI reference, and
  machine goldens alongside code. Verify every name here exists before using it as evidence.
- Build a repository-wide consumer inventory: dependency and lifecycle receipts/failures, wire
  conversions, CLI success/error routes, lifecycle moves renderer, TUI receipt/remedy consumption,
  and repair envelopes reusing these DTOs. Identify each warning/guidance decision's owner.

## Intended contract to challenge

1. Core owns warning decisions. `NewlyUnsafe()` preserves the old predicate: a different non-clear
   gate OR newly inconsistent state. Broken → blocked still warns; this is not a severity scale.
   `NewlyInconsistent()` warns only on false → true. Safe, unchanged unsafe, cleared, and role-only
   changes do not manufacture new warnings.
2. Shared core task recovery applies to dependency and lifecycle receipts. Dependency human/JSON
   consumers receive equivalent guidance, including runnable stable-ID blocker commands. Wire
   copies core decisions; renderers do not rederive policy from string gates or boolean pairs.
3. Dependency impacts represent the full proposed plan, even on refusal or partial failure.
   `dry_run`, `applied_task_ids`, `remaining_task_ids`, and lifecycle `committed` are separate
   durability evidence. Failed writes must not claim planned state landed. Partial/all-applied
   conflict failures stay visible and never silently auto-retry after the first durable write.
4. Pre-write CLI failures deliberately omit `error.dependency_mutation`; error text remains
   classified/actionable. Core callers retain the typed receipt. This exclusion is not a missing
   payload bug. Committed-prefix errors carry the structured receipt and the same core advice.
5. Schema 1.81 adds warning flags to shared impact DTOs and optional dependency `remedy`, including
   nested recovery envelopes. No existing field meanings, persisted Markdown schema, eligibility,
   authorization, override, repair transaction, or legacy migration semantics change.

Non-goals: a generic recovery framework, severity ordering, a second backend, TUI unreadable-marker
redesign, constructor mutation-policy design, or fixing the separately tracked init selector issue.
Evidence-backed discoveries outside this task should be proposed as followups, not imposed as
unbounded refactor closure blockers.

## Mandatory evidence floor

- Run baseline focused tests, full tests, race tests, lint, build, and generated-output comparisons
  in your sandbox. Show exact commands and results; report unavailable tooling as unverified.
- Independently reproduce preview, applied warning, no-op, clear, refusal, partial prefix, and
  all-applied failure. For prefix failures use the real filesystem migration tests as well as
  portable service tests; inspect resulting files and retry counts, not just error prose.
- Mutate at least three distinct owner/caller boundaries with compiler-valid probes: core
  predicate, mapper/renderer, and durable-prefix retry. Each named test must fail for the semantic
  reason claimed. Include one coordinated mapper+renderer mutation where their agreement could
  mask policy leakage. Restore and rerun before proceeding. A compile error is not a killed probe.
- Populate schema branches with `newly_unsafe=true`, `newly_inconsistent=true`, and non-empty
  dependency `remedy`, including `error.dependency_mutation`. Validate the emitted envelope against
  its own definition. Test false flags and omitted empty remedy too; empty arrays are not evidence.
- Run the blocker commands actually recommended by a receipt. Confirm the referenced stable IDs
  resolve; on preview distinguish the current graph's blockers from the proposed graph's blockers.
- Produce a claim/evidence matrix and an explicit verdict (ready, ready with tracked followups,
  or blocked). A no-findings verdict must show hostile tests, not paraphrase owner evidence.

## Required hostile angles

### Semantic ownership and presentation

Trace the full core → wire → CLI/TUI path. Can a non-CLI consumer obtain the warning and guidance
without parsing prose or reconstructing graph policy? Does advice survive when flags are false or
when a renderer receives a copied flag that disagrees with raw state? Is there any remaining adapter
policy duplication hiding behind a helper? Separate genuine recovery policy from output formatting.

### Durability and retry

Challenge failures before planning, after planning but before a write, after each atomic write,
after the last write, and during guard cleanup. Change the durable graph between a partial failure
and resuming intent. Do receipts and advice still distinguish current evidence from the abandoned
full plan? Can a committed conflict be swallowed into a successful no-op? Do error wrapping,
`errors.Is`/`errors.As`, exit classification, and structured recovery agree?

### Compatibility and all consumers

Review schema descriptions and the 1.81 additive claim. Shared DTOs appear in repair/lifecycle as
well as ordinary dependency output. Verify all populated shapes, required boolean properties,
optional remedy omission, and older field meanings. Look for legacy manually constructed receipts
or failures whose default values could create inconsistent guidance; identify actual reachable
callers rather than treating arbitrary fabricated invalid objects as production bugs.

### Systemic second pass, particularly for Antigravity

After the first pass, pick two apparently trustworthy abstractions and try to disprove them with
concrete counterexamples. Good candidates: tests compare a mapper to the same mutated predicate;
a fake port reports stronger durability than the real store; a generated schema branch is never
populated; error formatting invents guidance that the structured receipt lacks. Report for each
hypothesis the exact hostile input, path, observed output, and why it survived or was falsified.
Do not merely label patterns as anti-patterns. One demonstrated systemic issue is more useful than
several speculative redesigns. Also explain which owner-authored tests are strongest and weakest.

## Validation and restoration

Follow the mandatory independent-clone protocol injected above. Never execute mutation probes or
generators in the shared source. The sandbox baseline commit is the only permitted review commit;
do not stage, commit again, switch branches, push, or create a PR. Restore probes only in your sandbox.

Suggested commands (verify test names first):

```sh
go test ./internal/core -run 'TestTaskGraphStateImpactNewlyUnsafe|TestThreadProjectionImpactNewlyInconsistent|TestServiceDependency|TestDependencyPreviewRefusal|TestLifecycleThreadRecovery' -count=1
go test ./internal/wire -run 'TestImpactWireCopiesCoreDecisionsAndDependencyRemedy|TestJSONSchema_ValidatesRealOutput' -count=1
go test ./internal/cli/render -run 'TestDependencyHumanUsesCoreGuidance|TestMovesHumanUsesProjectedImpactFlags' -count=1
go test ./internal/cli -run 'TestDependencyMutationPublishesCoreRecoveryIntent|TestWriteErrorCarriesStructuredDependencyMutationRecovery|TestTaskDependAddRejectsCycle' -count=1
go test -race ./internal/store -run 'TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes' -count=20
go test ./...
go test -race ./...
golangci-lint run ./...
just build
./bin/tskflwctl -C . lint
```

Generate CLI docs and schema comments into a private temporary directory and compare them with
`docs/cli` and `internal/wire/schema_comments.json`; do not use `-update` to conceal drift.
For throwaway dogfood use `init --path "$SCRATCH" --taskflow-root planning --no-register`.
Global `-C` does not select init's bootstrap destination today; the separate followup owns that.
Do not mutate real planning states, source audit dispositions, task ACs, or machine fixtures.

## Deliverable

Update only your assigned audit, retaining this brief. Record severity-coded open findings with
reproduction, exact implementation locations, user/machine impact, a bounded recommended fix, and
the regression that would pin it. Keep falsified hypotheses and protocol/test limitations separate
from findings. Include commands/results, mutation outcomes and restoration, consumer inventory,
claim/evidence matrix, isolation attestation, and verdict. Do not duplicate the same root cause
across several bands or invent findings to meet a quota.

## Reviewer report

Pending external review. No findings have been recorded by the implementation owner.
