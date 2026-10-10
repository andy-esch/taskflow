---
schema: 1
id: 6gje9v15sjjm
bucket: open
area: bounded-planning-queries-and-identity-recovery-implementation-antigravity
date: "2026-10-10"
---
# Audit: Bounded planning queries and identity recovery — antigravity — 2026-10-10

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

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

Challenge the bounded planning-list contract and executable ID-less recovery fixes before merge.
Codex's primary lens is deterministic query/cursor semantics and neutral boundary ownership;
Antigravity's primary lens is real CLI/operator behavior and recovery that preserves identity and
tree topology. Both cross-check one critical invariant from the other lens. Work independently;
do not read the sibling report until yours is final. A high finding count is not the goal.

## Review target

- Read-only source: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch
  `feat/bounded-planning-reads-and-recovery`, base `8c1a86e`. Implementation ends at
  `883ff23`; capture the complete current source including later review/planning-only commits.
- [Bounded queries](../tasks/6gbn4g1j40pj-bound-and-page-agent-facing-list-queries.md#query-contract),
  [ID-less recovery](../tasks/6ggfd81jg0qg-make-id-less-thread-planning-recovery-instructions-executable.md),
  [CLI-contract Thread](../threads/6gbn4v2jpf2m-make-cli-contracts-self-describing-and-bounded.md),
  [machine revision policy](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md).
- Verify a consumer inventory: `core.ListQuery`, `queryList`, `Service.Query*` in
  `internal/core/service_list_query.go`; task/epic/audit/research/Thread list and audit findings
  in CLI; `renderListPage`, projected JSON, wire list envelopes and `ListPageJSON`.
  Existing TUI/List callers deliberately keep their original unbounded contract.
- Recovery producers: `core.MissingPlanningIdentityError`, `ComposeThreadApplyPlan`,
  `PrepareThreadApply`, `store.FS.currentPlanningIdentity`, configuration pointer verification,
  and `migrateUnlocked`. Verify the paths/tests before treating these names as evidence.
- The release task closeout is approved bookkeeping, not a claim that another retrospective
  smoke pass ran. The source-installed v0.23.0 binary is not the published archive's exact bytes.
- Known, pre-existing data-directory pointer identity mismatch is tracked by
  [6gjdgpq17sr6](../tasks/6gjdgpq17sr6-preserve-planning-identity-for-pointers-addressing-data-subdirectories.md).
  Do not present it as introduced or fixed here. You may challenge the new recovery advice
  for this case if it still recommends a second marker or misrepresents safe action.

## Intended contract to challenge

- Zero query preserves legacy order/filter/default wire shape. Any query option opts into stable
  sorting/filtering; bounds apply to returned rows, not filesystem scans. Source IDs, not declarations
  or location parsing, identify rows. Findings use audit source ID plus finding code.
- Stable primary order, ascending source-ID/value-digest ties, preservation of duplicate occurrences,
  and no caller-list mutation. Limit 1–1000; metadata search <=256 UTF-8 bytes. Findings reject dates.
- Cursor v1 is opaque and versioned, containing only digests and offset. Matching full values and
  complete diagnostics bind continuation; changes refuse with conflict/restart. Changed query/limit/
  kind/filter or malformed tokens refuse validation. It is neither write authority nor snapshot
  isolation; restored ABA views and unchanged nonmatching edits need not invalidate it.
- Typed/projected JSON publish the same optional page metadata. Changing output format/columns
  does not invalidate continuation. Human/CSV/name output stays rows with diagnostics on stderr.
  Empty/final/exact-boundary pages have no token; full read diagnostics and partial-read failure stay
  visible on every page. An empty Thread query retains the repository graph verdict.
- Date filters are inclusive, own updated_at falling back to creation (audit: date), not rollup
  activity or a reliable edit feed. Missing/invalid dates do not match.
- Conditional core recovery uses migration for an existing marker and explicit target/layout init
  for markerless trees. Configuration-local hints can use existing topology evidence; core cannot
  discover paths. Recovery preserves semantic documents, does not register a home space or silently
  rebind a pointer, and a regenerated identity refuses the old materialized plan.
- Revision 1.85 is ADDITIVE for opt-in flags/page metadata. Challenge any changed default semantics.

## Mandatory evidence floor

Do focused independent checks, not just the owner's full green suite. Minimum:

1. One real CLI paging workflow through each of the six surfaces, including nondefault metadata,
   typed-to-projected continuation, and one hostile record edit or new load diagnostic between pages.
   Record stdout/stderr and process exit classification; an in-process shape alone is insufficient.
2. Replay printed recovery from outside the target with an unrelated ambient space, for an ID-less
   marker, a markerless nested tree, and an opted-in missing target marker. Snapshot original document
   bytes/topology, pointer bytes, and home registration independently; prove old-plan refusal and
   fresh recomposition rather than accepting only a useful-looking message.
3. Perform a compiler-valid mutation of your primary invariant and one cross-check from the other
   lens. Select from `TestListQueryRefusesChangedViewsAndForeignQueries`,
   `TestListQueryUsesPortableSourceIdentityNotDeclaredIdentity`,
   `TestPointerRecoveryHintsDoNotSilentlyRebind`, and
   `TestThreadRecoveryHintsExecuteWithoutForkingExistingDocuments`. Name the exact edit, intended
   failing test, observed failure, and restored green result. Compile errors are invalid probes.

The owner's self-review already killed three mutations (view guard disabled, declared research
identity substituted, markerless pointer tree forced to '.'). You may reuse their baseline evidence,
but challenge a distinct route or independently reproduce the result rather than citing it as yours.

## Required hostile angles

Codex: prioritize query normalization/binding, same-ID duplicates and permutations, edits before
the cursor offset, filtering that changes membership, unsupported/oversized cursor versions, and
aliasing/caller mutations. Is the fingerprint the actual claimed view, or does omitted/internal
adapter data create false promises? Trace typed/projected metadata and schema with nondefault values.
Cross-check one recovery command against a nested tree, not only its string.

Antigravity: settle each hypothesis as reproduced/falsified/unresolved with a reachable producer:

- A cursor succeeds after a matching row or read diagnostic changed, or loses a row after rename.
- `--json -c` drops the continuation or makes it depend on the chosen columns.
- A page hides a failed record, produces clean exit status for partial reads, or hides unhealthy
  graph state when its Thread results are empty.
- An emitted `config migrate` is unusable in the state recommending it, or printed initialization
  silently selects an empty flat tree rather than existing nested data. Include paths with spaces
  and quotes, and pointer/registry context rather than making cwd rescue the instruction.
- Recovery changes pointer identity/unknown configuration fields, registers a space despite
  --no-register, or permits an old plan to be adopted after regenerating identity.

Both: perform a separate systemic pass over shared helpers and adjacent consumers. Look for
apparently safe generic abstractions masking adapter assumptions; flaky tests relying on source
order; clear bounds but misleading cost claims; date filtering confused with a complete change feed;
and compatibility changes slipped into default calls. Distinguish product defects from preferences,
coverage gaps, intentional opt-in tradeoffs, and the already tracked alias bug.

## Validation and restoration

Owner evidence at implementation checkpoint: `just test` (race), `just lint`, `just build`, generated
CLI docs/schema drift tests, planning lint, diff check, and targeted mutation probes all passed.
100-task fixture: full JSON 19,836 bytes vs five projected rows plus continuation 548 bytes (97.2%
smaller). Verify that as output-volume evidence, not tokenization or avoided-reader-work evidence.

In your independent sandbox run focused core/config/CLI/wire tests and the hostile probes above.
Broader inherited owner gates need not be rerun unless your evidence suggests a gap. Keep build
outputs, fixtures, logs, and Go/lint caches outside the sandbox's tracked/untracked tree. Restore
every temporary mutation to the captured baseline; only your assigned audit may differ. Use the
mandatory helper verify/transfer protocol. Never copy implementation fixes back; report them.

## Deliverable

Keep the brief, fill the report below, and leave findings open using canonical audit grammar/tooling.
Findings first: root cause, production path, reachable consequence, reproduction, concrete evidence,
recommendation, and uncertainty. Consolidate by root cause. Include verified consumer inventory,
hostile-probe/mutation ledger, a bounded surviving-hypothesis ledger, and the required isolation
attestation. A no-findings verdict needs hostile evidence; a blocked review is incomplete, not clean.

## Reviewer report

Review not yet performed. Replace this line with your report; do not pre-settle findings or turn
the owner's checked task criteria into evidence that invariants hold.
