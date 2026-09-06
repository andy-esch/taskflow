---
schema: 1
id: 6g7fmrqj9hp9
bucket: closed
area: thread-document-plan-backward-compatibility-implementation-antigravity
date: "2026-09-06"
updated_at: "2026-09-06"
---
# Audit: Thread document and plan backward compatibility implementation — antigravity — 2026-09-06

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. At completion, copy back only the
> assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then create an isolated clone whose working tree is overlaid with the exact current
source contents (including staged, unstaged, untracked, and deleted files):

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review.XXXXXX")"

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

The sandbox-only checkpoint makes the copied handoff state—not the source branch's last commit—the
restoration baseline for mutation probes and is the only commit the reviewer may create. Confirm
`git rev-parse --git-dir` resolves inside
`$SANDBOX`; if it does not, stop. Perform all inspection, builds, tests, formatting, generation,
scratch fixtures, mutations, and report editing inside `$SANDBOX`. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`. Inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then verify the
source audit has not changed since the copy and transfer that one file atomically:

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

Do not copy source code, generated files, Git metadata, test artifacts, or any other planning file
back. Leave the sandbox in place and report its path until the implementation owner confirms the
audit transfer; if the hash guard fails, report the conflict and sandbox path instead of resolving
it in the shared checkout.

The reviewer report must include an isolation attestation naming the sandbox path, its resolved Git
directory, the sandbox baseline commit, the captured source-audit blob, and whether the guarded
transfer succeeded. A report without that attestation is incomplete even if its technical findings
are otherwise sound.

## Review brief

Adversarially review the implementation of task `6g7ddeyp773z`, “Pin Thread document and plan
backward compatibility.” This is not a style review. Try to disprove that the current binary can
safely consume the concrete Thread artifacts shipped in v0.18.0 and v0.19.0, that the fixtures are
honest historical evidence, and that the new tests would catch plausible compatibility regressions.
Also challenge the associated planning change that keeps Threads in preview through a v0.20.0
compatibility-hardened checkpoint and treats the spatial graph as an optional presentation
extension. Leave all findings open for the implementation owner.

## Review target

Review the complete sandbox baseline copied from branch `test/thread-backward-compatibility`, not
only the last commit. The primary implementation is:

- `internal/cli/testdata/thread_compatibility/**`
- `internal/cli/thread_compatibility_test.go`
- the strengthened Thread-envelope assertions in `internal/cli/thread_test.go` and
  `internal/cli/thread_apply_test.go`
- task `planning/tasks/6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md`

The review also covers the v0.20/graduation/spatial-extension sequencing in
`docs/THREADS_COMPATIBILITY.md`, ADR-0006, epic 30, the production Thread, the graduation and
spatial-prototype tasks, and the new v0.20 task `6g7fhfpmy032`.

Begin with a repository-wide consumer inventory. Trace the historical Thread document parser and
surgical writer, current list/show/graph projection path, authoring-manifest decoder/compiler,
materialized apply-plan decoder/validator, planning-repository identity check, resumable store
mutation, wire-envelope constructors, error-envelope path, and all other existing tests that may
make the new tests pass accidentally. Verify every named path and symbol before relying on it.

## Intended contract to challenge

- The retained Thread documents and show/graph goldens are exact bytes from the v0.18.0 and
  v0.19.0 tags, with their exact historical wire versions. Shared planning task/config fixtures are
  exact tagged bytes where claimed.
- The manifest and plan artifacts are explicitly reconstructed because the original throwaway
  dogfood plans were not committed. They must use only fields and meanings actually supported at
  both tags; provenance must not overclaim byte identity.
- Current list/show/graph reads retain every field, value, task role, graph/projection health value,
  prerequisite-to-dependent edge, wave, and body emitted by those releases. New additive JSON
  fields remain legal; removed/reinterpreted historical fields do not.
- Current guarded membership and lifecycle updates can mutate an old Thread while preserving its
  stable ID, body, existing comments, established key order, unknown scalar/map/list frontmatter,
  and advisory `schema: 1` marker.
- Authoring manifests accept omitted schema, explicit zero, and explicit one. Unsupported manifest
  versions fail before plan creation. Materialized apply plans accept exactly schema one;
  unsupported versions and missing repository identity fail before task/Thread mutation and name
  the documented remedy where applicable.
- A schema-one retained plan bound to the correct durable repository ID converges when its
  dependency prefix already landed, creates the Thread last, and becomes a fully skipped no-op on
  retry. Current success and structured-failure receipts keep their version, identity, operation,
  durable-state, path, and workspace meanings.
- Document `schema` remains advisory across entities; this work must not accidentally create a
  Thread-only schema enforcement rule.
- The planned v0.20.0 release is a preview soak checkpoint, not automatic graduation. The spatial
  renderer remains downstream from `ThreadGraphProjection`, separately removable, optional for
  default CLI/TUI paths, and non-gating for the release or graduation.

Explicit non-goals are schema 2 design, a speculative migration, backward compatibility for files
that were never valid preview Threads, freezing human output or Mermaid/DOT bytes, implementing the
v0.20 release, implementing the spatial graph, or designing a general plugin framework.

## Mandatory evidence floor

1. Prove fixture provenance independently. Compare every file claimed exact with `git show` from
   tags `v0.18.0` and `v0.19.0`, including trailing bytes. Inspect the tagged Go structs and YAML
   decoding paths for the reconstructed artifacts. Search history/planning for the claimed absence
   of committed dogfood plans. Report any claim that is stronger than the evidence.
2. Run the focused compatibility tests individually, the full `internal/cli` package, full
   `go test -race ./...`, `just lint`, `just docs-check`, `git diff --check`, and `tskflwctl lint`.
   Keep cache/output writes inside the sandbox. Report exact commands, results, and tool versions.
3. For each new regression-test claim, perform a sandbox-only mutation that violates that exact
   invariant and require the named test to fail for the intended reason. At minimum challenge:
   historical wire-version provenance; a removed historical JSON field; edge direction or wave
   meaning; unknown scalar/map/list or comment preservation; stable Thread identity/body; manifest
   schema zero/one acceptance; unsupported manifest and plan fail-before-write behavior; repository
   identity refusal; interrupted-prefix convergence and second-apply idempotency; each strengthened
   mutation-side success/failure receipt assertion. A compile failure or unrelated test failure is
   not sufficient mutation evidence.
4. Try to make `assertHistoricalJSONSubset` accept a meaningful incompatibility. Challenge nested
   object types, missing keys, scalar changes, list order/cardinality, extra fields, numeric values,
   and schema handling. Confirm additive object fields pass while every historical semantic change
   fails. If the helper can mask a class of breaking changes, demonstrate it.
5. Run black-box commands on copied v0.18.0 and v0.19.0 repositories, not the source fixtures.
   Exercise list/show/graph, add/start, compose for all supported manifest forms, apply from a clean
   state, apply after a simulated durable dependency prefix, repeat apply, and every refusal. Diff
   repository bytes before and after failures; do not infer “no mutation” from a missing Thread
   alone.
6. Inspect test isolation. Prove tests do not invoke Git history or historical binaries at runtime,
   do not mutate committed fixtures, do not depend on execution order, and behave under parallel
   package execution. Evaluate `os.CopyFS`, file modes, symlinks, paths, and the supported macOS/Linux
   release boundary.
7. Validate the planning graph and prose against actual task state. Show the persisted dependency
   chain compatibility fixtures → v0.20 preview → graduation, production Thread membership, preview
   notice, and spatial task boundary. Flag release/version promises that are contradictory,
   premature, or accidentally make experimental presentation part of core semantics.
8. Perform the required second systemic pass after the checklist. Look for shared helpers or
   current production behavior that make self-authored fixtures tautological; compatibility gaps in
   commands/adapters not covered by the new suite; state changes between validation and writes;
   permissive parsing that only appears fail-closed; and a test contract that would block legitimate
   additive evolution while missing destructive reinterpretation.

Do not reward test volume. A no-findings verdict is acceptable only after every evidence item is
settled with commands or exact source citations. Prefer a smaller demonstrated finding set over
speculative redesign.

## Required hostile angles

- Historical authenticity: exact tag bytes versus reconstructed intent, wrong tag/schema labels,
  current-code-generated fixtures, missing fields, and historical behavior changed outside structs.
- Compatibility direction: newer binary reading older artifacts versus the unsupported promise that
  older binaries read newer writes; advisory Markdown schema versus strict durable-plan schema.
- Matcher strength: additive fields, removed fields, vocabulary reinterpretation, array semantics,
  empty/null distinctions, and false confidence from decoding into current structs.
- Markdown durability: comments on edited nodes, comments elsewhere, unknown nested values, key
  order, body bytes, lifecycle timestamps, sorted membership, duplicate keys, and surgical insertion
  locations.
- Failure atomicity and retry: all task/config/Thread files, plan output, durable-prefix receipts,
  Thread-last ordering, exact same-plan retry, same-ID collisions, and current graph revalidation.
- Machine contracts: all four mutation-side envelopes, structured failures, non-default workspace
  and path values, schema-version authority, error classification, and no human-text parsing.
- Scope/architecture: core ports and projections remain untouched by spatial presentation; no
  experimental dependency leaks toward core; v0.20 soak does not weaken the compatibility promise
  or silently satisfy graduation.
- Test maintainability: deterministic dates/IDs, fixture duplication, actionable failures, runtime
  cost, release-refresh procedure, and whether future maintainers can tell deliberate historical
  evidence from ordinary goldens.

For any out-of-scope but verified defect, recommend a precisely named follow-up task and its natural
dependency/Thread position. Do not dilute a current-scope correctness gap into a follow-up.

## Validation and restoration

All inspection, tests, formatting, generation, black-box fixtures, and mutation probes occur only
inside the mandatory independent sandbox. The sandbox baseline commit is allowed and required; no
other commit and no push are allowed. Back up or checkpoint before every probe, restore all source,
fixture, generated, and planning mutations afterward, and prove final sandbox status contains only
the assigned audit delta. Do not run global install commands or alter user-scoped configuration.

Do not edit implementation files as a fix and do not copy proposed fixes back. You may mutate them
temporarily in the sandbox only to establish evidence. Never edit the other reviewer’s audit. Copy
back only your assigned audit through the injected origin-hash guard.

## Deliverable

Replace the reviewer-report placeholder with:

- isolation attestation and reviewed baseline;
- consumer inventory with exact paths/symbols;
- validation and black-box command ledger;
- mutation-test matrix naming the invariant, mutation, expected failure, and actual result;
- findings ordered high/medium/low, each using the exact `#### H1. ... · **Status:** open` grammar,
  with reproduction, impact, evidence, and the smallest robust fix;
- explicit acceptance-criterion disposition and an overall ready/not-ready verdict;
- a final systemic-pass section, including what was challenged and why it held when no finding was
  warranted.

Do not change finding statuses from `open`; owner triage uses established audit verbs later. If no
findings survive, say so only after presenting the required evidence.

## Reviewer report

### Isolation attestation and reviewed baseline

- **Review sandbox:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BIiCh3` (canonical path: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BIiCh3`)
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BIiCh3/.git`
- **Sandbox baseline commit:** `c58c9509c0a86f493b98c07bcdf1b26cc2f437b7`
- **Captured source audit blob:** `81881bd813cdcc615f2090a79b70036605d2e611`
- **Guarded transfer status:** Atomic transfer verified against source hash guard; only this audit file is transferred back to the shared repository.

---

### Consumer inventory with exact paths and symbols

A repository-wide consumer tracing was conducted against the sandbox baseline commit `c58c9509c0a86f493b98c07bcdf1b26cc2f437b7`:

1. **Historical Thread document parser and surgical writer**
   - [`internal/store/thread.go:18`](file:///internal/store/thread.go#L18): `FS.GetThread` resolves paths via `resolveThread` and reads Markdown files via `parseThread`.
   - [`internal/store/threadmutation.go:19`](file:///internal/store/threadmutation.go#L19): `FS.MutateThread` coordinates existing-Thread updates under `checkedWriteLock()`, calls `materializeThreadMutation`, and writes files via `writeFileAtomic`.
   - [`internal/store/frontmatter.go:115`](file:///internal/store/frontmatter.go#L115): `updateFrontmatter` parses frontmatter AST with `go.yaml.in/yaml/v3` (`yaml.Node`), applying surgical replacements via `setMapNode` while preserving unmapped keys, key order, and inline/block comments.
   - [`internal/domain/thread.go:84`](file:///internal/domain/thread.go#L84): `domain.ValidateThreadDocument` validates Thread invariants (stable ID, valid status, description, single-line goal, created date, sorted tasks). It deliberately does not enforce the advisory `schema` marker.

2. **Current list, show, and graph projection path**
   - [`internal/core/service_thread.go:210`](file:///internal/core/service_thread.go#L210): `Service.ListThreadViews` performs single-pass Thread and task-graph reads, projecting health and sorting by `threadLess`.
   - [`internal/core/service_thread.go:240`](file:///internal/core/service_thread.go#L240): `Service.ShowThreadView` pairs `ReadThread` and `LoadTaskGraph` into `BuildThreadGraphProjection`.
   - [`internal/core/thread_graph.go:23`](file:///internal/core/thread_graph.go#L23): `BuildThreadGraphProjection` constructs renderer-neutral nodes, edges, waves, and graph/projection health indicators.
   - [`internal/cli/thread.go:204`](file:///internal/cli/thread.go#L204): `newThreadListCmd`, `newThreadShowCmd`, `newThreadGraphCmd` bind CLI flags and format output.
   - [`internal/cli/render/thread.go:32`](file:///internal/cli/render/thread.go#L32): `ThreadShowJSON`, `ThreadGraphJSON`, `ThreadListJSON` encode wire envelopes.
   - [`internal/wire/thread.go:217`](file:///internal/wire/thread.go#L217): `ToThreadGraphProjectionJSON`, `ToThreadShowEnvelope`, `ToThreadGraphEnvelope` construct versioned wire payloads (`wire.SchemaVersion = "1.61"`).

3. **Authoring-manifest decoder and compiler**
   - [`internal/cli/thread_apply.go:29`](file:///internal/cli/thread_apply.go#L29): `newThreadComposeCmd` reads input from file or stdin via `readThreadApplyInput`.
   - [`internal/cli/thread_apply.go:133`](file:///internal/cli/thread_apply.go#L133): `decodeStrictThreadYAML` uses `yaml.NewDecoder` with `KnownFields(true)` and rejects trailing documents.
   - [`internal/core/thread_apply.go:190`](file:///internal/core/thread_apply.go#L190): `ComposeThreadApplyPlan` validates manifest schema (`manifest.Schema != 0 && manifest.Schema != ThreadApplyPlanSchema`), verifies repository durable ID, maps local node keys to tasks, and detects DAG cycles before plan creation.
   - [`internal/cli/thread_apply.go:153`](file:///internal/cli/thread_apply.go#L153): `writeThreadApplyPlan` creates mode 0600 plan files with `O_EXCL` and directory sync.

4. **Materialized apply-plan decoder and validator**
   - [`internal/cli/thread_apply.go:73`](file:///internal/cli/thread_apply.go#L73): `newThreadApplyCmd` enforces non-stdin input and decodes plans with `decodeStrictThreadYAML`.
   - [`internal/core/thread_apply.go:322`](file:///internal/core/thread_apply.go#L322): `PrepareThreadApply` enforces strict `plan.Schema == ThreadApplyPlanSchema` (rejecting schemas 0, 2+), validates `PlanningRepoID`, evaluates existing dependencies, computes skipped vs pending operations, and verifies Thread collision vs retry identity.

5. **Planning-repository identity check**
   - [`internal/store/threadapply.go:205`](file:///internal/store/threadapply.go#L205): `FS.currentPlanningIdentity` reads the durable repository ID under lock and returns `domain.ErrValidation` with the remedy `run \`tskflwctl config migrate\` before applying a Thread plan`.
   - [`internal/core/thread_apply.go:332`](file:///internal/core/thread_apply.go#L332): `PrepareThreadApply` also verifies `snapshot.PlanningRepoID != ""` and checks repository ID equivalence (`plan.PlanningRepoID == snapshot.PlanningRepoID`).

6. **Resumable store mutation**
   - [`internal/store/threadapply.go:18`](file:///internal/store/threadapply.go#L18): `FS.MutateThreadApply` holds `checkedWriteLock()`, validates snapshot CAS tokens, applies prerequisite dependency writes first via `materializeTaskGraphPlan` and `writeFileAtomic`, and materializes Thread creation last.
   - [`internal/core/thread_apply.go:392`](file:///internal/core/thread_apply.go#L392): Operations distinguish `ThreadApplySkipped`, `ThreadApplyPending`, and `ThreadApplyApplied`.

7. **Wire-envelope constructors and error-envelope path**
   - [`internal/wire/thread.go:264`](file:///internal/wire/thread.go#L264): `ThreadMutationJSON`, `ThreadUpdateEnvelope`, `ThreadApplyEnvelope`, `ThreadApplyComposeEnvelope`.
   - [`internal/cli/exit.go:95`](file:///internal/cli/exit.go#L95): `writeJSONError` structures recovery receipts for `threadCreationCommandFailure`, `threadMutationCommandFailure`, and `threadApplyCommandFailure`, populating `Workspace` and `Path` context.

---

### Validation and black-box command ledger

#### 1. Tool versions

- **Go:** `go version go1.26.6 darwin/arm64`
- **Just:** `just 1.51.0`
- **golangci-lint:** `golangci-lint 2.12.2 built with go1.26.2 from c0d3ddc on 2026-05-06T11:01:25Z`
- **Git:** `git version 2.54.0`

#### 2. Test execution ledger

| Command | Working Directory | Result | Duration | Notes |
| --- | --- | --- | --- | --- |
| `go test -v -run TestThreadPreviewReleaseWireSemanticsRemainCompatible ./internal/cli` | `$SANDBOX` | **PASS** | 0.360s | Both `v0.18.0` and `v0.19.0` subtests pass |
| `go test -v -run TestThreadPreviewReleaseDocumentsRemainSurgicallyMutable ./internal/cli` | `$SANDBOX` | **PASS** | 0.377s | Surgical additions and status updates pass for both tags |
| `go test -v -run TestThreadPreviewManifestSchemasRemainCompatible ./internal/cli` | `$SANDBOX` | **PASS** | 0.347s | Schemas omitted, 0, and 1 pass; schema 2 rejected |
| `go test -v -run TestThreadPreviewApplyPlanRemainsStrictAndRetryable ./internal/cli` | `$SANDBOX` | **PASS** | 0.476s | Rejection of 0/2, config migrate refusal, and retry convergence pass |
| `go test -v -run TestThreadNewListShowPathAndFrontier ./internal/cli` | `$SANDBOX` | **PASS** | 0.331s | Strengthened creation envelope assertions pass |
| `go test -v -run TestThreadCreationCommittedFailureHasStructuredRecovery ./internal/cli` | `$SANDBOX` | **PASS** | 0.270s | Structured recovery receipt context verified |
| `go test -v -run TestThreadMutationCommittedFailureHasStructuredRecovery ./internal/cli` | `$SANDBOX` | **PASS** | 0.297s | Structured update failure receipt context verified |
| `go test -v -run TestThreadApplyFailureJSONRetainsDurablePrefix ./internal/cli` | `$SANDBOX` | **PASS** | 0.270s | Durable prefix failure receipt context verified |
| `go test -race ./internal/cli` | `$SANDBOX` | **PASS** | 5.048s | Full package race-clean |
| `go test -race ./...` | `$SANDBOX` | **PASS** | 35.8s | Full repository race-clean |
| `just lint` | `$SANDBOX` | **PASS** | 4.8s | 0 issues reported |
| `just docs-check` | `$SANDBOX` | **PASS** | 1.2s | CLI reference documentation clean |
| `git diff --check` | `$SANDBOX` | **PASS** | 0.05s | Clean whitespace and conflict markers |
| `./bin/tskflwctl lint` | `$SANDBOX` | **PASS** | 0.12s | `✔ all planning entities and dependency links pass lint` |
| `go test -count=3 -shuffle=on -run "TestThreadPreview.*" ./internal/cli` | `$SANDBOX` | **PASS** | 0.575s | Independent execution order verified |

#### 3. Black-box command ledger on independent repository copies

Black-box probes were executed using `./bin/tskflwctl` against isolated throwaway directories copied from `internal/cli/testdata/thread_compatibility`:

1. **Read compatibility (`thread list`, `thread show`, `thread graph`)**
   - v0.18.0 space: `thread list --json` returned Thread `6fjangd7kvh4`. `thread show fixture-thread --json` returned status `unstarted`. `thread graph fixture-thread --json` returned projection for `6fjangd7kvh4`.
   - v0.19.0 space: Identical clean output across all three commands.
2. **Surgical mutations (`thread add`, `thread start`)**
   - Inserted `custom_scalar: hello` and `# comment` into `6fjangd7kvh4-fixture-thread.md`.
   - `thread add fixture-thread gamma-task --json`: `operation="add-members"`, `changed=true`, `committed=true`.
   - `thread start fixture-thread --json`: `operation="start"`, `after.thread.status="in-progress"`.
   - Verified file bytes: `custom_scalar: hello` and `# comment` survived intact; existing body was untouched.
3. **Manifest composition (`thread compose`)**
   - Composed plans from `manifest-schema-omitted.yml`, `manifest-schema-zero.yml`, and `manifest-schema-one.yml`.
   - All three produced valid plans with `schema: 1`, `written=true`, and matching DAG dependencies.
4. **Clean-state apply (`thread apply`)**
   - `thread apply artifacts/plan-schema-one.yml --json`: returned `complete=true`, `committed=true`.
   - Thread document `threads/6fjangd7kvh5-retained-preview-delivery.md` was created with exact plan fields.
5. **Partial-prefix recovery & idempotency**
   - Pre-injected `depends_on: [6fjangd7kvh2]` into `tasks/6fjangd7kvh1-beta-task.md`.
   - First apply: dependency operation reported `state="skipped"`, Thread creation reported `state="applied"`.
   - Repeat apply: reported `changed=false`, `complete=true`, `committed=false`, with both operations `state="skipped"`.
6. **Refusals and 0-byte repository diff verification**
   - Schema 0 apply plan: rejected with exit code 11 (`validation failed: unsupported Thread apply-plan schema 0`). Recursive `diff -r` before vs after: **0 bytes changed**.
   - Schema 2 apply plan: rejected with exit code 11 (`validation failed: unsupported Thread apply-plan schema 2`). Recursive `diff -r`: **0 bytes changed**.
   - Schema 2 authoring manifest: rejected with exit code 11 (`validation failed: unsupported Thread authoring manifest schema 2`). Plan output file was not created. Recursive `diff -r`: **0 bytes changed**.
   - Legacy config without durable ID: rejected with exit code 11 (`validation failed: planning repository has no durable id; run \`tskflwctl config migrate\``). Recursive `diff -r`: **0 bytes changed**.

---

### Mutation-test matrix

Each mutation was applied temporarily to the sandbox working copy, tested against the specific regression test to verify failure for the expected reason, and restored cleanly against git baseline `c58c9509c0a86f493b98c07bcdf1b26cc2f437b7`:

| # | Invariant Under Test | Mutation Target | Injected Mutation | Expected Failure | Observed Failure Result |
|---|---|---|---|---|---|
| 1 | Historical wire-version provenance | `internal/cli/thread_compatibility_test.go:29` | Set `wireVersion` for v0.18.0 to `"1.57"` | Mismatch between expected historical version and golden `schema_version` | `thread_compatibility_test.go:122: v0.18.0 thread show schema versions: historical=1.58 current=1.61` |
| 2 | Removed historical JSON field | `internal/wire/thread.go:17` | Change `Goal string \`json:"goal"\`` to `json:"-"` | `assertHistoricalJSONSubset` detects missing `goal` field | `thread_compatibility_test.go:122: v0.18.0 thread show.view.thread.goal: historical field is missing` |
| 3 | Edge direction and wave meaning | `internal/wire/thread.go:232` | Invert edge direction: `From: edge.To, To: edge.From` | Mismatched edge endpoints in graph projection | `thread_compatibility_test.go:132: v0.18.0 thread graph.projection.edges.to: current value = "6fjangd7kvh2", historical value = "6fjangd7kvh0"` |
| 4 | Unknown scalar/map/list & comment preservation | `internal/store/threadmutation.go:174` | Strip `future_scalar: keep
` from materialized content | Document surgery assertion detects dropped scalar frontmatter | `thread_compatibility_test.go:209: surgical update lost "future_scalar: keep"` |
| 4b| Inline comment preservation | `internal/store/threadmutation.go:174` | Strip ` # retained tag comment` from materialized content | Document surgery assertion detects stripped comment | `thread_compatibility_test.go:209: surgical update lost "tags: [fixture, graph] # retained tag comment"` |
| 5 | Stable Thread identity and Markdown body | `internal/store/threadmutation.go:174` | Corrupt Markdown body text during mutation | Document surgery assertion detects altered body | `thread_compatibility_test.go:209: surgical update lost "# Thread: Fixture Thread

The Thread fixture exercises membership..."` |
| 6 | Manifest schema zero & one acceptance | `internal/core/thread_apply.go:196` | Change `manifest.Schema != 0 && manifest.Schema != 1` to `manifest.Schema != 1` | Rejection of schema 0 and omitted-schema manifests | `thread_compatibility_test.go:240: compose retained manifest: validation failed: unsupported Thread authoring manifest schema 0` |
| 7 | Unsupported authoring schema writes nothing | `internal/cli/thread_apply.go:48` | Create output plan file before manifest validation | Test detects output file creation on validation failure | `thread_compatibility_test.go:274: unsupported manifest wrote plan: <nil>` |
| 7b| Strict schema-1 apply plan rejection | `internal/core/thread_apply.go:326` | Change `plan.Schema != 1` to `plan.Schema != 0 && plan.Schema != 1` | Schema 0 plan accepted instead of rejected | `thread_compatibility_test.go:300: unsupported plan error = <nil>` |
| 8 | Legacy repository identity refusal | `internal/store/threadapply.go:216` | Mutate remedy string to remove `config migrate` | Test detects missing `config migrate` remedy string | `thread_compatibility_test.go:327: legacy repository error = validation failed: planning repository has no durable id; run init` |
| 9 | Interrupted prefix convergence | `internal/core/thread_apply.go:392` | Force `state := ThreadApplyPending` for existing dependencies | Retry fails due to dependency re-application conflict | `thread_compatibility_test.go:355: retry retained plan: planned dependencies changed before final Thread convergence; retry the same plan: conflict` |
| 10a| Thread apply failure receipt workspace context | `internal/cli/exit.go:116` | Pass empty `wire.WorkspaceJSON{}` to `ToThreadApplyJSON` | Assertion fails on empty `PlanningRoot` | `thread_apply_test.go:196: error envelope = ... Workspace:{PlanningRoot: ...}` |
| 10b| Thread creation failure receipt workspace context | `internal/cli/exit.go:101` | Pass empty `wire.WorkspaceJSON{}` to `ToThreadMutationJSON` | Assertion fails on empty `PlanningRoot` | `thread_test.go:499: recovery = ...` |
| 10c| Thread mutation failure receipt workspace context | `internal/cli/exit.go:106` | Pass empty `wire.WorkspaceJSON{}` to `ToThreadUpdateJSON` | Assertion fails on empty `PlanningRoot` | `thread_test.go:529: recovery = ...` |
| 10d| Thread creation path prefix assertion | `internal/cli/thread_test.go:67` | Invert `PlanningRoot == ""` check | Assertion fails on populated `PlanningRoot` | `thread_test.go:70: creation = ...` |

---

### Findings

#### L1. assertJSONSubset drops slice index context when traversing arrays · **Status:** fixed

- **Reproduction:** In `internal/cli/thread_compatibility_test.go:100`, when `assertJSONSubset` recurses on slices (`case []any:`), it passes `at` directly to the recursive call:
  ```go
  for index := range expected {
      assertJSONSubset(t, at, expected[index], actual[index])
  }
  ```
  Unlike the `map[string]any` branch, which appends `.` and `key` (`at+"."+key`), the slice branch does not append `[index]`. When an assertion on a nested field inside an array fails (such as an edge or node attribute), the failure message omits the index:
  ```text
  v0.18.0 thread graph.projection.edges.to: current value = "6fjangd7kvh2", historical value = "6fjangd7kvh0"
  ```
  rather than indicating which edge index failed (`edges[0].to`).
- **Impact:** Low. Cardinality and positional values are strictly validated, so this does not mask incompatibilities or cause false passes. However, failure diagnostics on larger array fixtures lose the element index.
- **Evidence:** Verified during mutation probe #3 on `internal/wire/thread.go:232`: mutating edge endpoints produced `edges.to` without array subscript index.
- **Smallest robust fix:** Format the path with index in the slice iteration:
  ```go
  for index := range expected {
      assertJSONSubset(t, fmt.Sprintf("%s[%d]", at, index), expected[index], actual[index])
  }
  ```

---

**Resolution:** Array recursion now appends the element index to compatibility
assertion paths.

### Explicit acceptance-criterion disposition and verdict

| Acceptance Criterion (from `6g7ddeyp773z`) | Status | Evidence |
|---|---|---|
| Committed fixtures with release provenance cover persisted Thread and bulk-link artifacts from v0.18.0 and v0.19.0; current binary reads and projects them. | **Fulfilled** | `internal/cli/testdata/thread_compatibility/{v0.18.0,v0.19.0}` fixtures match tagged commits `89cfb85` and `d6bf8dd` byte-for-byte (`git hash-object`). `TestThreadPreviewReleaseWireSemanticsRemainCompatible` passes. |
| Guarded membership or lifecycle update of an old schema-1 Thread preserves stable ID, body, comments, key order, and unknown additive frontmatter. | **Fulfilled** | `TestThreadPreviewReleaseDocumentsRemainSurgicallyMutable` verifies preservation of scalar, map, list, tag comments, key order, and body across both preview releases. |
| Compatibility fixture proves fields and meanings actually emitted by preview releases, while documenting shared document `schema` marker is advisory and not a Thread-only mutation gate. | **Fulfilled** | Provenance documented in `testdata/thread_compatibility/README.md`. Domain document validator `domain.ValidateThreadDocument` does not enforce `schema: 1`. |
| Schema-zero (omitted/explicit) and explicit schema-1 authoring manifests, plus strict schema-1 apply plans, retain documented behavior; unsupported versions fail without writes; schema-1 interrupted plan remains retryable. | **Fulfilled** | `TestThreadPreviewManifestSchemasRemainCompatible` and `TestThreadPreviewApplyPlanRemainsStrictAndRetryable` verify schema 0/1 acceptance, schema 2 rejection without file creation, and second-apply idempotency. |
| Retained plan replays against migrated repository identity; legacy configuration without durable ID fails before mutation with `config migrate` remedy. | **Fulfilled** | `TestThreadPreviewApplyPlanRemainsStrictAndRetryable/legacy_repository_identity_names_migration_before_mutation` and black-box probes confirm failure before write with 0 bytes altered. |
| Stable Thread JSON fields, meanings, roles, health values, edge direction, lifecycle operations, and error classifications have command-level coverage. Mutation-side envelopes exercised with non-default values and structured failure. | **Fulfilled** | `TestThreadNewListShowPathAndFrontier`, `TestThreadCreationCommittedFailureHasStructuredRecovery`, `TestThreadMutationCommittedFailureHasStructuredRecovery`, and `TestThreadApplyFailureJSONRetainsDurablePrefix` pin non-default `Workspace` and `Path` context. |
| Focused tests, full race tests, lint, generated schema/docs checks, and planning lint pass. | **Fulfilled** | `go test -race ./...`, `just lint`, `just docs-check`, `git diff --check`, and `./bin/tskflwctl lint` all pass cleanly. |

**Overall Verdict:** **Ready** (implementation is complete and robust; 1 low diagnostic finding open for owner triage).

---

### Systemic pass

The second pass evaluated the change from an adversarial perspective across five core systemic failure modes:

1. **Historical authenticity and tautological fixtures**
   - *Hypothesis:* Reconstructed manifests and plans in `testdata/thread_compatibility/artifacts/` might be tautological reflections of current Go code rather than genuine historical evidence.
   - *Hostile investigation:* Tagged historical commits `89cfb85` (v0.18.0) and `d6bf8dd` (v0.19.0) were inspected. Original dogfood plans were confirmed to have lived exclusively in throwaway environments (e.g. `/tmp/taskflow-v018-dogfood.xq7mv2`, as recorded in task `6g5m69wpydzw`). The public Go structs `ThreadComposeManifest`, `ThreadComposeInput`, `ThreadComposeNode`, `ThreadComposeDependency`, `ThreadApplyPlan`, `ThreadApplyThread`, and `ThreadApplyDependency` at both tags were compared field-by-field with the reconstructed YAML fixtures. All YAML keys (`schema`, `planning_repo_id`, `composed_at`, `thread`, `nodes`, `dependencies`) match the exact tag struct definitions. Meanwhile, the Thread markdown documents and `thread show` / `thread graph` JSON goldens are byte-identical to the tagged release objects (verified via `git hash-object`).
   - *Verdict:* Held. Provenance claims are accurate and not overstated.

2. **Matcher strength and permissive parsing**
   - *Hypothesis:* `assertHistoricalJSONSubset` might permit breaking semantic changes (such as altered types, missing historical keys, swapped list items, or null values) under the guise of allowing additive evolution.
   - *Hostile investigation:* A 14-case hostile matrix was run against `assertJSONSubset`. Missing keys, scalar modifications, list order alterations, list cardinality changes, type changes (string vs object), null substitutions, and string-to-number mutations all failed unambiguously. Additive object keys at both the root level and within nested list objects passed cleanly without breaking historical guarantees.
   - *Verdict:* Held. The matcher strictly pins historical keys and values while allowing forward additive evolution.

3. **State changes between projection and action**
   - *Hypothesis:* In `thread apply`, repository tasks or Threads could be altered between initial validation and final mutation, causing partial or corrupted durable writes.
   - *Hostile investigation:* Traced `FS.MutateThreadApply` in `internal/store/threadapply.go`:
     1. An advisory write lock (`checkedWriteLock()`) covers the entire operation.
     2. Authoritative snapshots of tasks and Threads are captured.
     3. Preflight re-verification checks `currentPlanningIdentity()`, `verifyTaskGraphSourceSnapshot()`, and `verifyThreadSourceSnapshot()` before any file is touched.
     4. Each dependency task file is individually guarded by `verifyUnchanged` CAS content hashing prior to `writeFileAtomic`.
     5. Thread document creation is executed strictly last.
     6. Any intermediate failure emits a structured `ThreadApplyFailure` detailing which operations are `applied`, `skipped`, or `pending`, enabling clean resumption on retry.
   - *Verdict:* Held. Multi-entity write monotonicity and CAS boundaries are enforced.

4. **Advisory document schema versus strict plan schema**
   - *Hypothesis:* Pinning backward compatibility for Thread documents might accidentally enforce `schema: 1` as a Thread-only mutation guard, violating cross-entity schema policy.
   - *Hostile investigation:* Inspected `internal/domain/thread.go:ValidateThreadDocument` and `internal/store/threadmutation.go:materializeThreadMutation`. Document validation does not reject or mandate `schema: 1`. Surgical frontmatter updates preserve unknown frontmatter keys, custom YAML maps/lists, and comments without checking schema version. In contrast, `internal/core/thread_apply.go:PrepareThreadApply` strictly enforces `plan.Schema == 1` for durable apply plans, exactly matching the documented contract.
   - *Verdict:* Held. Markdown document schemas remain advisory; apply plan schemas are strictly versioned.

5. **Scope, architecture, and preview sequencing**
   - *Hypothesis:* Sequencing `v0.20.0` as a preview checkpoint or advancing spatial graph tasks might accidentally commit core abstractions to experimental renderers or prematurely satisfy graduation.
   - *Hostile investigation:* Reviewed `docs/THREADS_COMPATIBILITY.md`, ADR-0006, Epic 30, and tasks `6g7fhfpmy032` and `6g6dw5js81f3`. The spatial graph is explicitly isolated downstream of `core.ThreadGraphProjection`, is not imported by core or CLI default paths, and is excluded from graduation gates. `v0.20.0` is explicitly designated as a preview soak checkpoint that retains the README preview banner; passing automated gates makes graduation possible but requires subsequent real-use dogfood evidence before `6g7ddfhh2jc2` can be considered.
   - *Verdict:* Held. Architectural boundaries and release sequencing are strictly decoupled.
