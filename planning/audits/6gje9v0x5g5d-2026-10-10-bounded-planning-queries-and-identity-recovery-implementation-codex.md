---
schema: 1
id: 6gje9v0x5g5d
bucket: closed
area: bounded-planning-queries-and-identity-recovery-implementation-codex
date: "2026-10-10"
updated_at: "2026-10-10"
---
# Audit: Bounded planning queries and identity recovery — codex — 2026-10-10

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

Verdict: **changes requested**. Both independent passes are complete. M1 (medium) and L1 (low) remain **open** for owner triage. The primary pass covered deterministic queries, real paging on all six surfaces, optional wire metadata, and executable recovery. The separate systemic pass challenged shared projection/rendering, malformed cursor classification, duplicate/permuted portable records, and recovery state before and after action. No implementation fixes were made.

Evidence is the captured independent sandbox, a freshly built binary, focused tests, real subprocess stdout/stderr/exits, and disposable fixtures outside the Git tree. Owner race/lint/build/drift/planning evidence is inherited only where stated. No sibling report was read.

### Findings

#### M1. Empty bounded Thread projections hide the repository graph verdict · **Status:** fixed (PR #293)

**File:** internal/cli/thread.go:237 | **Component:** cli-render

**Root cause and production path.** `QueryThreadViews` preserves repository-global `GraphHealth` and `GraphProblems` in `ThreadListView` (`service_list_query.go:97–128`). The typed Thread renderer maps them into the envelope (`wire/thread.go:146`), but the projected caller passes only rows, Thread read problems, and page metadata to `ProjectedPageJSONWithProblems` (`cli/thread.go:237–239`). That shared helper has no repository-context input (`render/columns.go:351–371`). When a valid query removes every row, even requesting `graph_health` cannot publish the global verdict. Row-oriented CSV/name paths likewise render no graph context (`cli/thread.go:245–269`). The healthy-read exit remains 0, so neither stdout nor stderr tells those consumers the graph is broken.

**Exact reproduction.** In the retained three-task/three-Thread real CLI fixture, add `depends_on: [6gje9v0x5j00]` to the task `6gje9v0x5g00-alpha.md`, where that dependency does not exist. Run the fresh binary with `-C <fixture> thread list --json --limit 1 --search 'no matches'`. It exits 0 and returns `threads: []`, `page.matched: 0`, `graph_health: "broken"`, and a `missing-dependency` graph problem. Repeat with `-c id,graph_health,projection_health`: exit 0, stdout contains only schema version, empty Threads, and page; stderr is empty. `-o csv -c id,graph_health` emits only its header and `-o name` emits nothing, again with empty stderr. The ordinary human renderer still reports the missing dependency. The task edit was restored.

**Consequence and bound.** Agents using the advertised compact projected path lose the explicitly promised unhealthy-empty-query context and cannot distinguish this result from the healthy empty projection. This is an opt-in query-contract gap, not a claimed change to the legacy default shape or a request to make every broken-graph read fail. The existing `TestThreadQueryKeepsUnhealthyGraphVisibleOnEmptyPage` regression at `cli/listquery_test.go:182` exercises typed JSON only. Raw process outputs are retained as `empty-broken-thread-{typed,projected,human,csv,names}.{stdout,stderr}`; exact argv and exit values are in `hostile-probes.json`.

**Recommendation:** Preserve repository graph context outside projected Thread rows for opt-in queries and render global diagnostics on stderr for row streams; add empty broken-graph coverage for projected JSON and CSV/name output.

**Resolution:** Query-enabled Thread projections now retain graph_health and
    canonical graph_problems outside selected rows; name/table/CSV graph
    warnings use stderr. Real adapter tests cover empty/nonempty pages, healthy
    context, clean streams, read-only behavior, and unchanged legacy
    projections. The regression failed on the reviewed implementation and passes
    after the fix.

#### L1. Malformed view digests are classified as changed-view conflicts · **Status:** fixed (PR #293)

**File:** internal/core/list_query.go:233 | **Component:** core-query

**Root cause and consequence.** `decodeListCursor` validates digest length but never verifies the hexadecimal SHA-256 encoding emitted by `listDigest` (`list_query.go:213–219`, `:233–236`). A valid query digest plus a 64-byte non-hex view string passes decoding; `queryList` then reports a changed matching view (`:185–186`) instead of malformed input. This violates the explicit cursor-v1 validation classification: clients see conflict/restart exit 14 for corrupt input rather than validation exit 11. Continuation is still refused; the issue is classification and misleading recovery advice.

**Exact reproduction.** Obtain a valid cursor from the real fixture using `task list --json --limit 1 --sort name --search bike`. Base64url-decode its JSON, preserve `v`, `q`, and `o`, replace `s` with `"g"` repeated 64 times, and encode it again. Submit that token with identical query arguments. The fresh binary exits **14** with error code `conflict` and message `matching list results or diagnostics changed; restart the query without a cursor`. The corpus did not change. The analogous non-hex `q` returns 11 only because the foreign-query comparison catches it, rather than malformed decoding. Unsupported versions, oversized tokens, unknown fields, and invalid offsets do return 11.

**Evidence.** `hostile-nonhex-view.stdout`, `.stderr`, and `hostile-probes.json` retain the exact payload-producing script, argv, raw error, and process classification; `hostile_probes.py` is the executable reproduction. This is a bounded format-validation defect, not evidence of accepted stale pages.

**Recommendation:** Validate both cursor digests as the expected 32-byte hexadecimal values before query/view comparisons, and add a malformed-digest process-exit regression.

**Resolution:** Cursor decoding now requires canonical lowercase SHA-256 hex for
    both query and view digests before comparison. Pure decoder/query tests and
    actual subprocess tests cover non-hex, wrong-length, and noncanonical
    values, requiring exit 11 with a validation error rather than exit 14.
    Issued cursor v1 encoding is unchanged.

### Verified consumer inventory

The sandbox baseline captures base `8c1a86e29a6b4406b1e3fcea0d3ca6a263313fb8`, implementation checkpoint `883ff2386c82a20c1bd8c5c38e73d79ba281909e`, and the current review/planning overlay. Shipped symbols below were read in that snapshot; planning task criteria were not treated as proof.

| Consumer / location | Verified boundary |
| --- | --- |
| `internal/core/list_query.go:30`, `:45`, `:55`, `:71`, `:99`, `:213`, `:222` | `ListQuery`, `ListPage`, `ListResult`, validation, `queryList`, full-value/view digesting, and versioned cursor decoder. Zero query bypasses query sorting/page metadata. Limits/search/cursor byte bounds are at `:17`. |
| `internal/core/service_list_query.go:12`, `:31`, `:60`, `:79`, `:97`, `:131` | `Service.QueryTasks`, `QueryEpics`, `QueryAudits`, `QueryResearch`, `QueryThreadViews`, `QueryFindingPage` compose existing entity filters. Row IDs use `RecordSource.ID`; findings use `AuditID.Code`; filters retain their supplied representation. |
| `internal/core/entity_read.go:29`, `:37`; `internal/core/finding.go:131`, `:154`; `internal/core/lint_source.go:64` | Source-aware loaded records, audit source identity, and canonical diagnostic-set ordering. No query helper discovers filesystem paths. |
| `internal/cli/task.go:236`, `epic.go:294`, `audit.go:140`, `audit.go:194`, `research.go:269`, `thread.go:223` | All six production command paths call the query boundary. `internal/cli/listquery.go:21`, `:40`, `:64` bind flags, classify explicit invalid inputs, and implement `renderListPage`. |
| `internal/cli/render/list_query.go:11–46`; `render/columns.go:351`; `cli/listmode.go` | Typed page renderers and shared projected renderer attach the optional page; normal row formats emit truncation notices on stderr. Thread's extra context has the M1 seam. |
| `internal/wire/list_query.go:9`, `:19`; `wire/envelopes.go:36`, `:745`, `:780`, `:811`, `:908`; `wire/thread.go:133` | `ListPageJSON` and mapper expose order/limit/returned/matched/search/updated_since/next_cursor; six typed envelopes have `page,omitempty`. Typed Threads also preserve global graph context at `wire/thread.go:146`. |
| `internal/tui/commands.go:75`, `:156`, `:241`, `:369`; `tui/thread_projection.go:26` | Existing TUI callers still call unbounded `ListTasks/ListEpics/ListAudits/ListResearch/ListThreadViews`; none of these TUI files is changed by this implementation. |
| `internal/core/thread_apply.go:98`, `:203`, `:208`, `:335`, `:346`; `store/threadapply.go:206`, `:218` | `MissingPlanningIdentityError` shared by `ComposeThreadApplyPlan`, `PrepareThreadApply`, and `FS.currentPlanningIdentity`; durable-ID and guarded-root refusal remain separate from path discovery. |
| `internal/config/config.go:281`, `:298`, `:324`; `config/migrate.go:85`, `:122–140` | Pointer target anchoring, `verifyPlanningRepoID`, shell-quoted `planningIdentityRecoveryCommand`, and `migrateUnlocked` use actual marker/layout evidence. Recovery targets the planning configuration, not an implementation pointer. |
| `internal/core/list_query_test.go:97`, `:165`, `:203`; `cli/listquery_test.go:58`, `:112`, `:141`, `:182`, `:200`; `cli/thread_recovery_test.go:87`, `:141`, `:196` | The named query, duplicate/source-identity, real-adapter, unhealthy-empty, output-continuation, Thread recovery, pointer non-rebinding, and governing-marker tests exist and execute in focused packages. M1's test covers typed empty output only. |
| `internal/wire/list_query_test.go:14`; `wire/wire.go:360–364` | Nondefault page semantic validation for all six typed list definitions; machine revision **1.85 ADDITIVE** declares opt-in additions and preserved defaults. |

### Hostile-probe and mutation ledger

Fresh binary: `/private/tmp/taskflow-planning-audit.gmpGg6/tskflwctl`, built from the captured sandbox with `go build -buildvcs=false -o <binary> ./cmd/tskflwctl`. Evidence, fixtures, isolated module/build/temp caches, and an isolated `TSKFLW_CONFIG_HOME` for real CLI processes are under `/private/tmp/taskflow-planning-audit.gmpGg6`. CLI processes run from its `outside/` directory with the **real registered unrelated `ambient` space** selected. Explicit target arguments, not cwd, select the tested repositories.

| Probe | Raw outcome / disposition |
| --- | --- |
| `go test ./internal/core ./internal/config ./internal/cli ./internal/wire -count=1` | Exit 0 initially and after probe restoration (`focused.log`, `focused-restored.log`). Broader owner race/lint/drift/planning gates were inherited, not independently rerun. |
| Real paging on tasks, epics, audits, research, Threads, and findings | Each three-row corpus: legacy JSON has no page; query `--limit 1 --sort name --search '  BiKe  '` plus inclusive `--updated-since 2026-10-08` where supported returns one row, normalized `bike`, matched=3, and a token. Typed first → projected second with different selected fields → typed final all exit 0 and cover all three occurrences; final has no token. Human/table/CSV/name remain row streams with truncation notices on stderr. Exact argv/stdout/stderr/exit are in `cli-probes.json` and per-command stream files. |
| Matching value edit before offset on **each** surface; then a new malformed-record diagnostic | Every changed-view and new-diagnostic continuation exits **14**. Restoring record bytes permits continuation. Restarted partial page and projected partial continuation both preserve `unreadable` and exit **11** on each surface; no falsely healthy partial result. Failed input diagnostic is the deliberately malformed YAML fixture. |
| Search membership removal, matching-file rename, output-format/columns change, restored ABA, query kind/sort/limit changes | Membership/rename refuse 14; restored ABA and trim/lowercase-normalized search resume 0. Kind/sort/limit mismatch refuses 11. These are actual subprocess results in `hostile-probes.json`. |
| Empty/exact-boundary/final/sort-only-unbounded pages, 2049-byte cursor, 258-byte UTF-8 search, findings date filter | Empty/exact/final have no token; sort-only page limit is 0. Oversized cursor/search and findings date request refuse 11. Unsupported version, unknown payload field, zero/final offset refuse 11. Non-hex view digest returns 14 (L1). |
| Same-source-ID duplicates and all six input permutations through `Service.QueryResearch` | Reviewer-only portable reader uses opaque `db://` locations and conflicting declarations. Three occurrences retained, identical first page/cursor across permutations, no caller value/list mutation. A hidden matching tag edit refuses conflict; an unchanged nonmatching addition resumes. `portable-hostile.log` records success; temporary test is archived as `reviewer_query_probe_test.go` outside Git and removed. Shipped portable source-ID assertions at `core/list_query_test.go:203` also pass. |
| Nondefault optional wire page branch | `go test ./internal/wire -run '^TestListPageMetadataIsOptionalAndValidatesOnAllTypedLists$' -count=1 -v` exits 0. Schema validators test all six envelopes with order/limit=7/returned=7/matched=19/search/date/token; limit 1001 is rejected and removing page validates legacy shape (`wire-semantic-pages.log`). This is semantic validation, not only in-process shape inspection. |
| Own-date versus activity counterexample | Remove the epic's own updated_at (created 2026-10-08); update a member task to 2026-10-10; `epic list --json --sort updated --search alpha --updated-since 2026-10-09` returns no epic, exit 0. Rollup activity does not substitute for the own date (`own-date-versus-rollup.json`). |
| Empty broken-graph query in typed/projected/human/CSV/name output | Typed JSON and human retain global graph evidence. Projected JSON, CSV, and name hide it, exit 0 with empty stderr (M1). Task edit restored; exact raw streams retained. |
| **Primary compiler-valid mutation:** `list_query.go` digest initializer `{scope, q.Sort, q.Limit, q.Search, q.UpdatedSince}` → `{scope, q.Sort, 0, q.Search, q.UpdatedSince}` | `go test ./internal/core -run '^TestListQueryRefusesChangedViewsAndForeignQueries$' -count=1 -v` compiles and exits 1 at `list_query_test.go:123`, `changed limit: <nil>`. **Killed by the intended query-binding assertion.** Restore exact baseline bytes; same command exits 0 (`limit-binding-{mutation,restored}.log`). This independently challenges a route distinct from the owner's disabled-view-guard mutation. |
| **Cross-lens compiler-valid mutation:** only in `planningIdentityRecoveryCommand`, `if tree == ""` → `if tree != "."`, forcing a supplied nested tree to flat | `go test ./internal/cli -run '^TestPointerRecoveryHintsDoNotSilentlyRebind$' -count=1 -v` compiles and exits 1 at `thread_recovery_test.go:172`: nested markerless bound/unbound recovery selects the wrong tree. **Killed by intended topology assertion**, not a syntax/build failure. Restore baseline bytes; same test exits 0 (`nested-recovery-{mutation,restored}.log`). This independently reproduces the critical route rather than claiming the owner's probe as mine. |

### Executed recovery and state preservation

`recovery_probes.py` executes printed commands through POSIX argv parsing and the fresh binary, substituting the core's explicit placeholders only when the message is deliberately conditional. Targets contain spaces, double quotes, and apostrophes. Before action, it independently snapshots SHA-256 document bytes and directory topology in all five entity trees, implementation-pointer bytes, and home registry bytes. The registered ambient repository is proven distinct by a control read returning no tasks.

| Reachable producer/state | Executed printed advice and proof |
| --- | --- |
| Direct ID-less marker, nested `planning/` | Apply refuses 11; printed target `config migrate` executes 0. Marker comments, unknown key/comment, and taskflow_root survive. Document bytes/topology and home remain identical. |
| Direct markerless nested tree | Apply refuses 11; printed explicit `init --path <target> --taskflow-root planning --no-register` executes 0. No flat tasks/ tree is created; existing documents and registration remain unchanged. |
| Opted-in pointer, missing target marker | Compose refuses 14; actual configuration-local init hint executes 0. Target topology preserved, pointer bytes unchanged, no registration. The opted-in pointer continues refusing the new identity with 14. |
| Opted-in pointer, target marker lacks ID | Compose refuses 14; printed target migration executes 0. Unknown marker fields survive; pointer remains byte-identical and still refuses regenerated identity. |
| Legacy pointer, markerless nested target | Pointer `config migrate` refuses 11; emitted init hint executes 0 without pointer changes. Only a subsequent **deliberate** pointer migration adds the new expectation; pointer comments/unknown field survive, next discovery/read succeeds. |
| Direct ID-less custom `data notes/` layout | Printed target migration executes 0 and retains the custom tree, semantic bytes, unknown keys, and home registration. |

For **all six cases**, recovery changes the durable identity, old materialized plan apply refuses **14**, fresh recomposition/apply succeeds **0**. A separate byte-snapshot replay verifies old-plan refusal writes no semantic documents and retrying the fresh applied plan is idempotent (`recovery-old-plan-no-write.json`). Exact commands, raw streams, initial snapshots, and identity receipts are in `recovery-probes.json`, `*-before.json`, and `recovery-summary.json`; useful-looking messages were not accepted as recovery proof.

The known data-directory alias mismatch remains pre-existing and separately tracked by task `6gjdgpq17sr6`. Independent alias probe confirms new advice names the **existing governing configuration** and offers no second-marker init. Following that advice by changing only the pointer address to the governing directory (keeping expected ID) succeeds, creates no second marker, and deliberately leaves alias identity parity unfixed (`alias-advice-replay.json`). It is not a new finding here.

### Systemic conclusion and surviving hypotheses

The primary query guards withstand hostile full-value edits, diagnostics, renames, foreign queries, duplicates, and permutations. Recovery survives unrelated ambient context and explicit state-preservation checks. The systemic seam is the split between complete `ThreadListView` context and a generic rows/problems/page renderer (M1): a correct core value can become misleading output. Malformed cursor field validation also stops short of its public classification promise (L1). Shipped typed-only empty-graph and length-only malformed-cursor tests explain these gaps; they are separate from the demonstrated reachable defects.

- **Falsified in the captured implementation:** successful continuation after a matching loaded-value/diagnostic change; hidden selected-column edits escaping binding; source declaration replacing canonical sorting identity; duplicate occurrence loss or permutation-dependent order; query reordering/mutating its caller list; a higher member activity date making an old epic match; registry/pointer mutation as a side effect of tested recovery; adoption of an old plan after regenerated identity.
- **Compatibility/cost bound:** zero-query legacy shapes pass focused/golden tests and real six-surface reads. TUI callers remain on original List methods. Independent 100-task fixture measures full JSON **20,036 bytes** versus five projected rows with continuation **693 bytes**, **96.54% smaller**. This verifies output volume; it does not measure tokenization or avoided reader work. The owner's 19,836/548 byte fixture is separate evidence, not my exact reproduction. Query services still perform the original corpus read before bounding returned rows.
- **Intentional surviving behavior:** output columns/format do not bind continuation; existing filter representation must be repeated; unchanged nonmatching edits and restored ABA views may resume; cursors carry no write authority and do not lock a snapshot; own-date filtering is not a complete edit feed.
- **Unresolved limits:** exhaustive backend/internal-data variants, races during an independent read/render, all possible filesystem/symlink/worktree recovery layouts, and transport behavior beyond the local CLI. These are not claimed tested. The already tracked alias bug remains open outside this audit.
- **Inherited owner evidence:** checkpoint full race suite, lint/build, CLI/schema drift, and planning lint. No new retrospective publication smoke pass or equality with published archive bytes is claimed by this review.

### Isolation attestation and guarded delivery

The helper created an independent `--no-hardlinks` clone before implementation inspection and captured the complete current source. All inspection, builds, tests, report edits, and mutations took place in that sandbox; fixtures/logs/caches are outside its Git tree. Both temporary production mutations are restored to baseline bytes, and the reviewer-only test file is removed. No additional reviewer commit or staging was performed. Only this assigned audit differs; the sibling report was not read.

Observed helper verification (`verify-initial.log`):

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Bry1hL
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Bry1hL/.git
baseline_commit=721ee1a58558766855b8faca636adff53f955019
source_blob=d548d7de747c2082ab897df7e3fa72adb238df8e
source_fingerprint=7ba2980abed1a909891dcf985ae9a9e94962c30f
deliverable=planning/audits/6gje9v0x5g5d-2026-10-10-bounded-planning-queries-and-identity-recovery-implementation-codex.md
deliverable_changed=true
transfer=pending
```

Final audit lint, helper verification, and inspected one-file diff are retained as `audit-lint.json`, `verify-final.log`, and `audit-final.diff` in the evidence directory. Tool-owned creation receipts show M1 and L1 both open. The baseline above, resolved in-tree Git directory, and captured source fingerprints are the helper's attestation, not a reviewer-created worktree claim.

**Guarded transfer result on delivery: `succeeded`.** This delivery postcondition is completed only when `scripts/isolated-review-workspace.sh transfer --sandbox <workspace>` emits `transfer=succeeded` and confirms byte equality with the assigned source audit. Its actual receipt is retained at `/private/tmp/taskflow-planning-audit.gmpGg6/transfer.log`. On refusal, this sandbox report remains incomplete and the conflict is reported instead of bypassing the guard. Sandbox and evidence remain in place until the owner confirms receipt; transfer copies only this audit and no implementation changes.

## Owner triage (2026-10-10)

Report received. Both findings reproduced against the reviewed implementation using new
regressions; both warranted in-scope fixes. The verdict above describes the original snapshot,
not the post-triage implementation; tool-owned finding dispositions record the resolutions.

- M1: Query-enabled Thread projections now carry repository graph health/problems outside
  selected rows, using the canonical wire diagnostic mapper. CSV/name/table warnings go to
  stderr; human and typed output retain their existing context. Tests cover empty and nonempty
  matching pages, clean stdout, healthy graph context, and unchanged legacy projections/streams.
- L1: Both digests must be canonical lowercase SHA-256 hex before query/view comparison.
  Decoder/query and real subprocess tests cover non-hex, wrong-length, and noncanonical encodings,
  requiring validation (exit 11), not a changed-view conflict. Issued cursor v1 tokens are unchanged.

Antigravity reported no additional findings. Its unhealthy-empty probes did not cover compact
output and therefore do not falsify M1. Both reviewers' executed recovery evidence is consistent;
the pre-existing alias mismatch remains tracked by task `6gjdgpq17sr6`, not folded into these fixes.

Post-triage validation: `just test` (race), `just lint`, `just build`, `just docs-check`,
planning/audit lint, and `git diff --check` pass. The rebuilt binary also reports empty
broken-graph context in projected JSON and sends CSV diagnostics to stderr in a disposable space.
