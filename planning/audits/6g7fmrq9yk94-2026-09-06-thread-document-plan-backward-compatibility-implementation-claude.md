---
schema: 1
id: 6g7fmrq9yk94
bucket: closed
area: thread-document-plan-backward-compatibility-implementation-claude
date: "2026-09-06"
updated_at: "2026-09-06"
---
# Audit: Thread document and plan backward compatibility implementation — claude — 2026-09-06

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

| Item | Value |
| --- | --- |
| Sandbox path | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.22qBo9` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.22qBo9/.git` (real directory; `git rev-parse --absolute-git-dir` resolves inside the sandbox) |
| Isolation proof | `git clone --no-hardlinks` + `rsync -a --delete --exclude=.git`; no `objects/info/alternates`, no symlink, no `git worktree`. Verified with `ls -ld .git` (directory, not gitfile) and an empty alternates read. |
| Sandbox baseline commit | `4605371e334880b4169195a08fb8160a4f3bcb7d` (`chore: capture review sandbox baseline`) — the only commit created |
| Reviewed baseline | Full copied working tree of `test/thread-backward-compatibility` (staged + unstaged + untracked), merge-base with `origin/main` = `81da388f09138985b6dcb688c7fc23bb4962a4c0` |
| Captured source-audit blob | `5d9a6f2804a26554d2c88e5754bdc4d5be799899` |
| Guarded transfer | Hash guard re-checked immediately before transfer: current source blob `5d9a6f28…` == captured blob → **guard passed**, single-file atomic `mv` performed, `cmp -s` confirmed identical. |
| Build/test cache | Redirected to `…/T/taskflow-review-cache` (`GOCACHE`, `GOTMPDIR`, `GOLANGCI_LINT_CACHE`); no writes to user-scoped config, no global installs. |
| Final sandbox state | `git status --short` empty; `git diff --stat HEAD` empty — every mutation probe restored to the baseline checkpoint. The audit delta is carried by the transfer only. |

No write-capable command was run in `$SOURCE_ROOT`. The only source-checkout operations were reading
this brief, `git hash-object` on the audit, and the final guarded `mv`.

Tool versions: `go1.26.6 darwin/arm64` · `just 1.51.0` · `golangci-lint 2.12.2 (built with go1.26.2)` ·
`git 2.54.0`.

### Consumer inventory (verified paths and symbols)

Every symbol below was read in the sandbox at the cited line; nothing here is inferred from planning prose.

| Consumer | Location | Relevance |
| --- | --- | --- |
| Historical Thread parser | `internal/store/threadstore.go` `parseThread`; `internal/store/frontmatter.go:126` `splitFrontmatterStrict`, `documentMapping` | Reads retained v0.18/v0.19 documents |
| Surgical writer | `internal/store/threadmutation.go:128` `materializeThreadMutation` → `internal/store/frontmatter.go:118` `updateFrontmatter` → `setMapNode:357`, `deleteMapNode`, `assembleFile:273`, `detectLineEnding`; atomic commit at `threadmutation.go:102` `writeFileAtomic` | Comment/unknown-key/key-order/body durability |
| Guarded mutation entry | `internal/store/threadmutation.go:19` `MutateThread` (CAS via `ifVersion: hashContent(content)`, path re-resolution, reload + `domain.ValidateThreadDocument` before write) | State change between projection and action |
| list / show / graph projection | `internal/wire/thread.go:147` `ToThreadsEnvelope`, `:169` `ToThreadShowEnvelope`, `:216` `ToThreadGraphProjectionJSON`, `ThreadGraphEdgeJSON:196`, `ThreadGraphWaveJSON:202` | Wire fields the goldens pin |
| Wire version | `internal/wire/wire.go:259` `const SchemaVersion = "1.61"` | Historical labels 1.58 / 1.60 |
| Manifest decoder | `internal/cli/thread_apply.go:144` `decodeStrictThreadYAML` (`KnownFields(true)` + trailing-document guard) | Fail-closed authoring parse |
| Manifest schema gate | `internal/core/thread_apply.go:197` (`Schema != 0 && Schema != ThreadApplyPlanSchema`) | Accepts omitted/0/1 |
| Plan validator | `internal/core/thread_apply.go:322` `PrepareThreadApply`; `:14` `const ThreadApplyPlanSchema = 1`; `:326` strict plan-schema gate | Strict schema-one |
| Repository identity | Three guards: `internal/store/threadapply.go:216` (in-lock re-read), `internal/core/thread_apply.go:332` (missing id), `:335` (mismatch → `ErrConflict`) | Refusal before mutation |
| Resumable apply | `internal/core/thread_apply.go:366-451` (durable-prefix `ThreadApplySkipped` detection, Thread-last op append at `:450`) | Interrupted-prefix convergence |
| Mutation-side envelopes | `internal/wire/thread.go:273` `ToThreadMutationJSON`, `:317` `ToThreadUpdateJSON`, `:429` `ToThreadApplyComposeEnvelope`, `:450` `ToThreadApplyJSON`, `:464` `ToThreadApplyEnvelope` | The four receipts |
| Error-envelope path | `internal/wire/envelopes.go:1049` `ThreadApply *ThreadApplyJSON` (+ `ThreadMutation`, `ThreadUpdate` siblings) | Structured failures |
| Test harness | `internal/cli/workspace_test.go:18` `runIn` (uses `-C root`, no `chdir`; stdout/stderr captured separately) | Order- and parallel-safe |
| **Accidental-pass risk** | `internal/cli/integration_golden_test.go:116` `TestGolden_MachineContract` (refreshable via `-update`) | See H1 |
| **Accidental-pass risk** | `internal/store` `TestUpdateFrontmatter_PreservesCommentsAndOrder` | See L1 |

### Fixture provenance (evidence floor 1)

Independently verified with `git hash-object` against `git rev-parse <tag>:<path>` — not by trusting the README.

| Fixture | Claim | Result |
| --- | --- | --- |
| `v0.18.0/6fjangd7kvh4-fixture-thread.md` | exact v0.18.0 bytes | blob `648b70a1…` == `v0.18.0:internal/cli/testdata/planning/threads/…` **IDENTICAL** |
| `v0.19.0/6fjangd7kvh4-fixture-thread.md` | exact v0.19.0 bytes | blob `648b70a1…` == tag blob **IDENTICAL** |
| `v0.18.0/thread-show.json` | exact tag golden | `9db213e0de…` == `v0.18.0:…/golden/thread_show_json.golden` **IDENTICAL** |
| `v0.18.0/thread-graph.json` | exact tag golden | `c4f6af043d…` **IDENTICAL** |
| `v0.19.0/thread-show.json` | exact tag golden | `38f1bab9c1…` **IDENTICAL** |
| `v0.19.0/thread-graph.json` | exact tag golden | `c2eb70e2b8…` **IDENTICAL** |
| `planning/tasks/*.md`, `planning/.tskflwctl.toml` | "byte-identical at both tags" | all four blobs equal at **both** tags **CONFIRMED** |
| README commit/wire labels | `89cfb851…`/1.58, `d6bf8dd8…`/1.60 | `git rev-list -n1` matches both; `git show <tag>:internal/wire/wire.go` gives `SchemaVersion = "1.58"` / `"1.60"` **CONFIRMED** |
| `artifacts/*.yml` | reconstructed, *not* claimed byte-exact | Every field present in the tagged structs: `git show <tag>:internal/core/thread_apply.go \| grep 'yaml:'` is **byte-identical across v0.18.0, v0.19.0 and HEAD** (31 tagged fields). `thread.title` correctly absent from the plan; `nodes[].member` and `dependencies[].from/to` all exist at both tags. **Reconstruction is honest and the README does not overclaim.** |

The "throwaway dogfood plan" claim is consistent with history: no committed plan artifact exists at
either tag (`git ls-tree -r <tag> | grep -i thread.*plan` returns only source, docs and planning
prose — no materialized `.yml` plan). The README states the reconstruction plainly rather than
implying byte identity, which is the correct disclosure.

### Validation ledger (evidence floor 2)

All commands run inside the sandbox with sandbox-local caches. Re-run after every probe was restored:

| Command | Result |
| --- | --- |
| `go test ./internal/cli/ -run TestThreadPreview -v -count=1` | **ok** — 4 tests / 15 subtests, all PASS (0.49s) |
| `go test ./internal/cli/ -count=1` | **ok** (2.95s) |
| `go test -race ./... -count=1` | **ok** — 0 failures across all packages |
| `just lint` (`golangci-lint run ./...`) | **0 issues** |
| `just docs-check` (`docgen` + `git diff --exit-code docs/cli`) | **PASS** |
| `git diff --check` | **clean** |
| `just build` → `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `go test ./internal/cli/ -shuffle=on` | **ok** — order independent |
| `go test ./internal/... -p 8` | **ok** — safe under parallel package execution |
| `go test ./internal/cli/ -count=3` then fixture re-hash | fixture tree hash unchanged (`322800c33b2a7118`); `git status` on the fixture tree empty |

### Black-box command ledger (evidence floor 5)

Repositories were built by **copying** the fixtures into `$TMPDIR/taskflow-bb*/planning` and driving
the freshly built `./bin/tskflwctl` (v0.19.0-47-g4605371), never the fixture directory in place.
Repository bytes were snapshotted with `shasum -a 256` over the whole tree before and after each
failure — no "no mutation" claim rests on a missing Thread file.

Reads (both v0.18.0 and v0.19.0 repositories, identical results):

| Command | Result |
| --- | --- |
| `thread list` | exit 0 — `unstarted / fixture-thread / 0/2 done / 0/2 drained / 1 eligible / healthy/healthy` |
| `thread show fixture-thread` | exit 0 — 2 members, 1 external gate, `satisfied nominally-complete/clear` |
| `thread graph … --format mermaid` | exit 0 — `n2 --> n0` (prerequisite→dependent preserved) |
| `thread frontier … --json` | exit 0 — `schema_version 1.61`, full view |
| `thread compose --from <omitted\|zero\|one>` | exit 0 for all three manifest forms, both repos |

Apply lifecycle:

| Scenario | Result |
| --- | --- |
| Apply from clean state | exit 0 · `changed=true complete=true committed=true` · ops `[(dependency, applied), (thread, applied)]` · exactly two files touched: `tasks/6fjangd7kvh1-beta-task.md` + new `threads/6fjangd7kvh5-…md` (Thread created last) |
| Repeat apply (same plan) | exit 0 · `changed=false complete=true committed=false` · ops `[(dependency, skipped), (thread, skipped)]` · **repository bytes IDENTICAL** — a true no-op, not merely a skipped Thread |
| Apply after simulated durable prefix (`depends_on` pre-landed) | exit 0 · ops `[(dependency, skipped), (thread, applied)]` · converges from the same plan |

Refusals — every one byte-clean:

| Refusal | Exit | Message | Repo bytes |
| --- | --- | --- | --- |
| plan `schema: 0` | 11 | `unsupported Thread apply-plan schema 0` | IDENTICAL |
| plan `schema: 2` | 11 | `unsupported Thread apply-plan schema 2` | IDENTICAL |
| plan `schema: 99` | 11 | `unsupported Thread apply-plan schema 99` | IDENTICAL |
| manifest `schema: 2` | 11 | `unsupported Thread authoring manifest schema 2` | IDENTICAL |
| mismatched repo id | 14 | `apply plan belongs to planning repository "6fjangd7kvhz", current repository is "6fjangd7kvhy"` | IDENTICAL |
| missing repo id | 11 | `planning repository has no durable id; run ``tskflwctl config migrate`` before applying a Thread plan` | IDENTICAL |
| Thread id collides with a task id | 14 | `planned Thread id 6fjangd7kvh0 is already used by a task: conflict` | IDENTICAL |
| planned Thread advanced since apply | 14 | `already exists and has advanced since this plan was applied (status); the plan will not overwrite it` | IDENTICAL |

Permissive-parsing probes (all correctly fail closed, exit 11, no writes):

- unknown top-level manifest field → `field future_semantics not found in type core.ThreadComposeManifest`
- unknown **nested** plan field → `field future_mode not found in type core.ThreadApplyThread`
- two YAML documents in one plan → `Thread apply plan must contain exactly one YAML/JSON document`

Retained-document hazards not covered by the suite but verified sound by hand:

- **Duplicate frontmatter key** → read refused with `malformed frontmatter: line 9: mapping key "tags" already defined at line 8`; no mutation.
- **CRLF Thread document** → read OK, `thread start` OK, CRLF **preserved** (`detectLineEnding`/`assembleFile`); file remains `CRLF line terminators` after the surgical write.
- **`schema: 2` Thread document** → reads, mutates, and lints identically to `schema: 1` (marker preserved verbatim). The advisory-marker contract holds; see M1 for the missing pin.

### Mutation-test matrix (evidence floor 3)

Each row: mutate production code (or fixture bytes) in the sandbox, run the named test, restore, and
confirm `git status --short` is empty. Compile failures were treated as inadmissible and reworked.

| # | Invariant | Mutation | Expected | Actual |
| --- | --- | --- | --- | --- |
| M1b | Historical wire-version provenance | v0.18.0 golden `schema_version` 1.58→1.59 | fail | **KILLED** — `…:122: v0.18.0 thread show schema versions: historical=1.59 current=1.61` |
| M1c | Release-table label honesty | table `wireVersion` 1.58→1.59 | fail | **KILLED** — same assertion |
| M2 | Removed historical JSON field | graph node `description` → `json:"-"` | fail | **KILLED** — `…nodes.description: historical field is missing` (both releases) |
| M3 | Edge direction (prerequisite→dependent) | `ThreadGraphEdgeJSON{From: edge.To, To: edge.From}` | fail | **KILLED** — `…edges.from: current "6fjangd7kvh0", historical "6fjangd7kvh2"` |
| M4 | Wave meaning (1-based index) | `Index: wave.Index - 1` | fail | **KILLED** — `…waves.index: current 0, historical 1` |
| M6 | Unknown scalar/map/list preservation | `updateFrontmatter` drops keys outside a known set | fail | **KILLED** — `surgical update lost "future_scalar: keep"` |
| M7 | Stable Thread identity | mutation restamps `id` to `6fjangd7kvh9` | fail | **KILLED** — membership receipt `ThreadID:6fjangd7kvh9` |
| M8 | Body bytes preserved | `assembleFile` trims the body | fail | **KILLED** — `surgical update lost "# Thread: Fixture Thread\n\n…"` |
| M9 | Advisory `schema: 1` marker retained | mutation unsets `schema` | fail | **KILLED** — `surgical update lost "schema: 1"` |
| M10 | Manifest schema 0/omitted accepted | require explicit `1` | fail | **KILLED** — `unsupported Thread authoring manifest schema 0` |
| M11 | Unsupported manifest refused before write | accept any manifest schema | fail | **KILLED** — `unsupported manifest error = <nil>` |
| M12 | Plan strictly schema-one | accept any plan schema | fail | **KILLED** — `unsupported plan error = <nil>` |
| M13 | Repo-identity refusal (core guard alone) | delete `snapshot.PlanningRepoID == ""` branch | fail | *survived* — shadowed by the store guard |
| M13b | Repo-identity refusal (store guard alone) | delete `repoID == ""` branch | fail | *survived* — shadowed by the core guard |
| **M13c** | **Repo-identity refusal (coordinated)** | **delete both guards together** | fail | **KILLED** — `legacy repository error = conflict: apply plan belongs to planning repository "6fjangd7kvhz", current repository is ""` (a third guard still refuses, but with the wrong class/remedy, which the test correctly rejects) |
| M15 | Durable-prefix detection | treat already-landed edges as pending | fail | **KILLED** — `planned dependencies changed before final Thread convergence; retry the same plan: conflict` |
| M16 | Thread-last ordering | emit the Thread operation first | fail | **KILLED** — recovery receipt shows `thread` op at index 0 |
| M17 | Second-apply idempotency | existing Thread reported `pending` | fail | **KILLED** — `converged apply = … Changed:true … (thread, pending)` |
| M18 | Apply receipt carries workspace | drop `Workspace` from `ToThreadApplyJSON` | fail | **KILLED** — both `thread_apply_test.go:196` and `thread_compatibility_test.go:367` |
| M19 | Apply receipt carries `plan_path` | drop `PlanPath` | fail | **KILLED** — both tests |
| M20 | `thread_mutation` receipt workspace | drop `Workspace` from `ToThreadMutationJSON` | fail | **KILLED** — `thread_test.go:70` and `:499` |
| M21 | `thread_update` receipt workspace | drop `Workspace` from `ToThreadUpdateJSON` | fail | **KILLED** — `thread_test.go` recovery + compatibility membership receipt |
| M22 | `thread_compose` receipt workspace | drop `Workspace` from `ToThreadApplyComposeEnvelope` | fail | **KILLED** — `thread_compatibility_test.go:252` |
| P2 | Suite pins schema **one**, not "the constant" | `ThreadApplyPlanSchema` 1→2 | fail | **KILLED** — `unsupported Thread authoring manifest schema 1`; the fixtures pin the literal, not the constant |
| M5 | Comment preservation on an **edited** node | `setMapNode` drops `LineComment` carry-over | fail | *survived the new suite* (killed only by `internal/store`'s `TestUpdateFrontmatter_PreservesCommentsAndOrder`) → **L1** |
| P1 | `thread list` historical field retained | `ThreadsEnvelope.GraphHealth` → `json:"-"` | fail | *survived the new suite* → **H1** |

Two mutations that survived a single-site edit (M13/M13b) were re-run as one coordinated mutation
(M13c) because the two guards defend the same invariant from adjacent layers; the coordinated form
kills the test, so the invariant is genuinely pinned and the redundancy is real defence in depth,
not a gap.

### Matcher strength: attacking `assertHistoricalJSONSubset` (evidence floor 4)

Additive evolution must stay legal; every historical semantic change must fail. Probes mutate either
the current wire structs or the fixture bytes, running `TestThreadPreviewReleaseWireSemanticsRemainCompatible`.

| Probe | Expected | Actual |
| --- | --- | --- |
| A1 — **additive** new object field (`future_field` on every graph node) | pass | **ok** — additive evolution remains legal |
| A2 — nested scalar flipped (`state.eligible` true→false) | fail | **FAIL** |
| A3 — list **order** reversed (`waves[].task_ids`) | fail | **FAIL** |
| A4 — list **cardinality** (extra node) | fail | **FAIL** |
| A5 — vocabulary reinterpretation (`role` member→owner) | fail | **FAIL** |
| A6 — `false` → `null` (empty/null distinction) | fail | **FAIL** |
| A7 — scalar **type** change (bool → string) | fail | **FAIL** |
| A8 — numeric value (`rollup.total` 2→3) | fail | **FAIL** |
| A9 — object → scalar (`state` becomes a string) | fail | **FAIL** |

I could not make the helper accept a meaningful incompatibility **within the payloads it is given**.
`assertJSONSubset` recurses object-wise over historical keys only (additive-safe), enforces exact list
order and cardinality, treats a missing historical key as fatal, and compares leaves with
`reflect.DeepEqual` after `encoding/json` normalisation, so `null`/`false`/`0`/`""` stay distinct.
The helper is sound. **Its weakness is which payloads reach it — see H1.**

One tautology worth recording, which is *not* a defect: line 71 compares
`got["schema_version"]` against `wire.SchemaVersion`, so it cannot detect a regressed current
version (mutating `SchemaVersion` 1.61→1.60 leaves this suite green). That invariant is held
elsewhere — the same mutation fails `TestGolden_MachineContract/task_list_json` and eleven sibling
goldens — so no finding is warranted; the historical half of the same comparison (M1b/M1c) is real.

### Test isolation (evidence floor 6)

- **No runtime Git or historical binary.** `grep -nE 'exec\.|os/exec|git |LookPath'` over
  `internal/cli/thread_compatibility_test.go` returns nothing. The committed bytes are the evidence.
- **Committed fixtures are never mutated.** Tree hash before and after `go test -count=3` is
  identical (`322800c33b2a7118`); `git status --short` on the fixture tree stays empty. This holds
  even though `TestThreadPreviewApplyPlanRemainsStrictAndRetryable` passes the *committed*
  `artifacts/plan-schema-one.yml` path straight to `thread apply`.
- **Order independence / parallelism.** `-shuffle=on` and `-p 8` both green. `runIn`
  (`workspace_test.go:18`) uses `-C root` rather than `chdir`, so the relative fixture paths are
  stable and no test mutates process-global state.
- **`os.CopyFS`, modes, symlinks.** All 15 fixture files are `100644` in the index and on disk; the
  tree contains no symlink, FIFO or socket, so `os.CopyFS` (which rejects irregular files and copies
  the source mode) always yields writable copies. `compatibilityRepo` correctly copies the dotfile
  `.tskflwctl.toml` — proven by the compose assertions resolving `Workspace.RepoID = 6fjangd7kvhz`.
- **macOS/Linux boundary.** Nothing in the new test is platform-conditional: no mode bits beyond
  0644/0755, no case-sensitivity assumption (all fixture names are lowercase and distinct), no path
  separator literals. Verified on `darwin/arm64`; nothing observed would differ on Linux.

### Planning graph and prose (evidence floor 7)

Persisted `depends_on` chains, read from frontmatter, match the ADR-0006 diagram exactly:

```
6g6wdvfjdaaa (graduation contract)
   └─> 6g7ddeyp773z  compatibility fixtures      status: in-progress
         └─> 6g7fhfpmy032  v0.20 preview          status: next-up
               └─> 6g7ddfhh2jc2  graduation      status: next-up
```

- **Spatial boundary holds.** `6g6dw5js81f3` depends on `[6g5rwjr0dz4p, 6g6scc9jgxae]` and **no task
  declares a dependency on it** (checked across all of `planning/tasks/`), so it gates neither the
  v0.20 checkpoint nor graduation. The ADR now additionally requires it to enter "through an optional
  CLI/TUI presentation adapter … core must not import it".
- **Production Thread membership.** All four tasks (`6g7ddeyp773z`, `6g7fhfpmy032`, `6g7ddfhh2jc2`,
  `6g6dw5js81f3`) are members of `6g503c6pfqeb-complete-production-threads`.
- **Preview notice intact.** `README.md:7-10` still carries the Threads preview block; the v0.20 task
  makes retaining it an acceptance criterion and lists preview removal as out of scope.
- **No contradictory release promise.** `docs/THREADS_COMPATIBILITY.md` now says passing the gates
  "makes graduation supportable; it does not force it", renumbers the decision procedure to five
  steps with v0.20 as step 1, and rewrites G6 as "real dogfood and preview soak". The v0.20 task's
  six acceptance criteria are all unticked, consistent with `status: next-up`.
- **Evidence integrity of the prose.** All 16 `../planning/...` links in
  `docs/THREADS_COMPATIBILITY.md` resolve to existing files, and all five test symbols named as G5
  evidence exist (`TestServiceThreadReadsComposeIndependentGraphAndThreadPorts` →
  `internal/core/service_thread_test.go`, `TestServiceComposeThreadApplyRendersDefaultTemplate` →
  `internal/core/service_thread_apply_test.go`,
  `TestServiceShowThreadGraphDetailUsesOnePairedReadInOrder` → `internal/core/thread_graph_test.go`,
  `TestTaskGraphReadAttributesUnreadableRecordWithoutFilesystemPath` →
  `internal/core/dependency_graph_test.go`,
  `TestThreadTopologyCursorOpensSelectedTaskByStableIdentity` →
  `internal/tui/thread_projection_test.go`). `just release-snapshot`, cited by the v0.20 task, exists
  in the justfile.

### Findings

#### H1. The compatibility suite pins only two of the six Thread machine surfaces shipped at both tags, so a destructive change to `thread list`/`frontier`/`path`/`plan` survives a routine golden refresh · **Status:** fixed

**Reproduction.** In the sandbox, drop a top-level field that both preview releases emitted from the
list envelope:

```sh
# internal/wire/thread.go:129
-	GraphHealth   string `json:"graph_health" jsonschema:"…"`
+	GraphHealth   string `json:"-"`

go test ./internal/cli/ -run TestThreadPreview -count=1   # ok — compatibility suite is GREEN
go test ./internal/cli/ -count=1                          # only TestGolden_MachineContract fails
go test ./internal/cli/ -run TestGolden_MachineContract -update -count=1
go test ./internal/cli/ -count=1                          # ok — the whole package is now GREEN
git status --short internal/cli/testdata/golden/          # 2 goldens rewritten
```

**Impact.** `thread_compatibility_test.go` feeds `assertHistoricalJSONSubset` only `thread show`
(line 122) and `thread graph` (line 132). `thread list` is checked at lines 138-141 for exactly one
property — `list.Threads[0].Thread.ID == "6fjangd7kvh4"` — and `thread frontier`, `thread path` and
`thread plan` are never invoked. Yet all four shipped exact goldens at both tags
(`v0.19.0:internal/cli/testdata/golden/thread_{list,frontier,path,plan}_json.golden`), and the list
envelope carries four top-level fields that appear nowhere in show/graph: `graph_health`,
`graph_problems`, `threads`, `unreadable` (`internal/wire/thread.go:128-132`).

The only net covering those surfaces is `TestGolden_MachineContract`, whose whole purpose is drift
*detection* — its failure message is literally "regenerate with `-update` if intended". A maintainer
who makes a breaking change and refreshes goldens (the routine, sanctioned move) gets **no**
compatibility signal at all. That inverts the task's stated purpose: the fixtures exist precisely so
that historical evidence is the thing you may *not* regenerate. The stated contract names this
surface explicitly — "Current **list**/show/graph reads retain every field, value, task role,
graph/projection health value, prerequisite-to-dependent edge, wave, and body emitted by those
releases" — so this is a current-scope gap, not a follow-up.

**Evidence.** P1 in the mutation matrix (survived); the golden-refresh sequence above; tagged golden
inventory via `git ls-tree -r v0.19.0 | grep 'golden/thread_'`.

**Smallest robust fix.** Retain the four remaining tagged goldens beside the existing two
(`git show v0.18.0:internal/cli/testdata/golden/thread_list_json.golden`, etc.) and extend the
`threadPreviewReleases` table with `listGolden`/`frontierGolden`/`pathGolden`/`planGolden`, routing
each through the existing `assertHistoricalJSONSubset`. The helper already has the right semantics
(A1-A9); only the inputs are missing. `thread path` needs its one-argument form
(`thread path <thread>`), and `thread plan` reuses the graph projection shape. No new helper, no new
fixture repository, roughly one table column and four calls.

**Resolution:** Retained exact tagged list, frontier, path, and plan JSON
alongside show and graph, and routed all six release surfaces through the
additive-only historical subset assertion.

#### M1. The "advisory document `schema`" guard the task is explicitly required not to break has no executable pin · **Status:** fixed

**Reproduction.** Copy a release repository, set the Thread document's marker to a future value, and
drive the CLI:

```sh
sed 's/^schema: 1$/schema: 2/' …/v0.19.0/6fjangd7kvh4-fixture-thread.md > $R/threads/6fjangd7kvh4-fixture-thread.md
tskflwctl -C $R thread show fixture-thread   # reads fine
tskflwctl -C $R thread add fixture-thread gamma-task   # mutates fine
grep '^schema:' $R/threads/6fjangd7kvh4-fixture-thread.md   # schema: 2 — preserved
tskflwctl -C $R lint   # identical output to the schema: 1 repo (2 pre-existing unknown-epic issues)
```

**Impact.** The behaviour is **correct today** — I verified read, mutate, marker preservation and
lint are all byte-identical between `schema: 1` and `schema: 2`. But nothing pins it. The task body
(`6g7ddeyp773z`, line 25) states the work must not turn "the repository's currently advisory document
`schema` marker into a Thread-only policy", and AC 3 is ticked partly on that basis; the intended
contract restates it as "this work must not accidentally create a Thread-only schema enforcement
rule". Every Thread document in the fixture tree carries `schema: 1`, and the only `schema: 2` in the
suite is a *manifest* (line 266), which is a different, deliberately strict boundary. A future change
that added a Thread-document schema gate would break every retained preview Thread and the entire
suite would stay green.

The documentation half of AC 3 is genuinely satisfied (`docs/THREADS_COMPATIBILITY.md:15,29,43`;
ADR-0006:1367; task lines 25 and 54). It is the executable half that is absent — and "make the
upgrade promise executable" is this task's whole objective.

**Evidence.** Black-box probe above; `grep -rn 'schema: 2' --include='*_test.go' internal/` returns
only `thread_compatibility_test.go:266` (the manifest case).

**Smallest robust fix.** One subtest in
`TestThreadPreviewReleaseDocumentsRemainSurgicallyMutable`: rewrite the copied Thread's marker to a
non-1 value before the existing `thread add`/`thread start` sequence and assert the read succeeds,
the mutation commits, and the marker survives verbatim. The fixture and helper (`compatibilityRepo`)
already provide everything needed; this is a few lines, not a new fixture.

**Resolution:** Added an executable schema-2 document probe proving current show
and guarded membership mutation accept and preserve the advisory marker.

#### L1. Comment preservation is only demonstrated on a node no Thread mutation touches · **Status:** fixed

**Reproduction.**

```sh
# internal/store/frontmatter.go:364 — delete the LineComment carry-over in setMapNode
-			if val.LineComment == "" { val.LineComment = old.LineComment }
go test ./internal/cli/ -run TestThreadPreviewReleaseDocumentsRemainSurgicallyMutable -count=1  # ok
go test ./... -count=1   # FAIL: internal/store TestUpdateFrontmatter_PreservesCommentsAndOrder
```

**Impact.** The test plants its comment on `tags` (line 162:
`tags: [fixture, graph] # retained tag comment`) and asserts it at line 206. But `tags` is not in the
update set for either exercised mutation — `materializeThreadMutation`
(`internal/store/threadmutation.go:151-172`) writes only `tasks`, `status`, `updated_at`,
`started_at` and `ended_at`. The comment therefore survives because its node is never rebuilt, so
this assertion exercises yaml.Node round-tripping, not `setMapNode`'s comment carry-over — the
"comments on edited nodes" case the brief names. The invariant *is* covered, one layer down, by
`internal/store`'s own test, which is why this is low rather than medium: no regression class is
unguarded, but the compatibility suite is weaker than it reads.

**Evidence.** M5 in the mutation matrix (survived the named test, killed by the store test).

**Smallest robust fix.** Move or duplicate the planted comment onto `tasks:` — a key both `thread
add` and the lifecycle verbs rewrite — and keep the existing assertion. One changed anchor string.

**Resolution:** Moved the retained inline comment onto tasks, the frontmatter
node rewritten by membership mutation, and asserted it survives.

#### L2. The two release directories are byte-identical apart from the version label, which the README does not say · **Status:** fixed

**Reproduction.**

```sh
git hash-object …/v0.18.0/6fjangd7kvh4-fixture-thread.md …/v0.19.0/6fjangd7kvh4-fixture-thread.md
# 648b70a14772ab9f91fb41c4bcd1a2e0bf4457f3  (both)
diff <(python3 -m json.tool …/v0.18.0/thread-graph.json) <(python3 -m json.tool …/v0.19.0/thread-graph.json)
# 2c2 < "schema_version": "1.58"  ---  > "schema_version": "1.60"
```

**Impact.** Not a correctness defect — these *are* the honest tagged bytes, and pinning both labels is
exactly right. The problem is legibility. The README explains that `planning/` is retained once
because it "is byte-identical at both tags, so retaining one labelled copy avoids fake format
differences", then applies the opposite treatment to the Thread documents and payloads, which are
equally identical. A future maintainer reading `threadPreviewReleases` will reasonably infer that the
two rows pin two different persisted shapes; they pin one shape and two version strings. That
invites either a wrong conclusion about what regression the v0.19.0 row can catch, or a
half-refresh at the next release.

**Evidence.** Hashes and diff above; the README rationale at lines 8-11.

**Smallest robust fix.** One sentence in `internal/cli/testdata/thread_compatibility/README.md`:
state that v0.18.0 and v0.19.0 emitted identical Thread documents and identical show/graph payloads,
that the pair therefore pins the wire-version labels and the shared shape, and name the command a
maintainer runs to retain a new release's bytes.

**Resolution:** Documented the identical v0.18.0/v0.19.0 shapes, distinct wire
labels, and the procedure for retaining a later release.

### Acceptance-criterion disposition

| AC | Verdict | Basis |
| --- | --- | --- |
| 1 — provenance-covered fixtures read and projected | **Met** | All six release artifacts and four shared planning files byte-verified against the tags; black-box reads green on both copied repositories |
| 2 — guarded update preserves ID, body, comments, key order, unknown frontmatter | **Met with a caveat** | M6/M7/M8/M9 kill identity, body, unknown scalar/map/list and the advisory marker; key order is asserted at lines 212-220. The *comment* half is demonstrated only on an untouched node — **L1** |
| 3 — fixtures prove emitted fields/meanings; advisory `schema` documented | **Partially met** | The proof half is strong (M2/M3/M4 + A1-A9). The advisory-marker half is documented in four places but has no executable pin — **M1** |
| 4 — manifest 0/omitted/1 and strict plans; unsupported fails without writes; interrupted plan retryable | **Met** | M10/M11/M12/M15/M17 all killed; P2 proves the fixtures pin schema *one*, not the current constant; every refusal byte-clean in black-box |
| 5 — replay against migrated identity; legacy config fails before mutation with the remedy | **Met** | M13c (coordinated) kills; black-box shows exit 11 with the `config migrate` remedy and identical repository bytes |
| 6 — stable JSON fields/meanings/roles/health/edges/lifecycle/errors covered; four mutation-side envelopes with non-default values and a structured failure | **Not met** | M18-M22 confirm all four envelopes carry non-default workspace/path values and two structured failures. But "stable Thread JSON fields … have command-level compatibility coverage" is false for `thread list`, `frontier`, `path` and `plan` — **H1** |
| 7 — focused, race, lint, docs/schema, planning lint pass | **Met** | Full ledger above, all green, re-verified after every probe was restored |

### Verdict

**Not ready.** The implementation itself is sound: I could not find a way to make the current binary
misread, silently rewrite, or partially mutate a retained v0.18.0/v0.19.0 artifact, and every
refusal I could construct was byte-clean and correctly classified. The fixtures are genuinely
honest — every "exact" claim verified against the tags, and the reconstructed artifacts do not
overclaim. The blocking gap is coverage, not behaviour: **H1** leaves four of six shipped machine
surfaces outside the compatibility net, and the one safety net they do have is explicitly designed
to be regenerated. **M1** leaves the task's own named non-goal unguarded. Both fixes are additive and
small; nothing here calls for redesign.

### Systemic pass (second review)

Challenged deliberately, with the result and why it held:

1. **Are the fixtures tautological — generated by the code they test?** No. The four goldens and both
   Thread documents are byte-identical to the tag blobs (`git hash-object` vs `git rev-parse
   <tag>:<path>`), so current code cannot have authored them. The reconstructed manifests/plan *are*
   self-authored, which is the real risk, so I checked whether they merely echo current behaviour:
   P2 (`ThreadApplyPlanSchema` 1→2) kills the suite, proving the fixtures pin the literal schema-one
   contract rather than tracking the constant. The YAML tag surface is byte-identical across
   v0.18.0, v0.19.0 and HEAD, so the reconstruction cannot contain a field that did not exist.

2. **Do shared helpers mask defect classes?** `assertHistoricalJSONSubset` was attacked from nine
   directions (A1-A9) and held on all of them while still admitting additive fields. The masking is
   not in the helper but in its call sites — H1.

3. **Boundaries that only appear to fail closed.** `decodeStrictThreadYAML`
   (`internal/cli/thread_apply.go:144`) sets `KnownFields(true)` *and* rejects trailing documents;
   both verified black-box against unknown top-level fields, unknown nested fields and a two-document
   plan. `requireThreadApplyPlanAbsent` uses `os.Lstat` (symlink-safe) and the writer uses
   `O_EXCL|0600`. Duplicate frontmatter keys are refused with a precise diagnostic. This is real
   fail-closed behaviour, not the appearance of it.

4. **State changing between projection and action.** `materializeThreadMutation` re-resolves the
   Thread path, CAS-stamps `ifVersion: hashContent(content)`, reparses the rendered document and
   re-runs `domain.ValidateThreadDocument` before `writeFileAtomic`; `currentPlanningIdentity`
   re-reads repository identity *inside* the guarded region and refuses if the root moved. M15
   surfaced the apply-side counterpart: perturbing durable-prefix detection produces `planned
   dependencies changed before final Thread convergence; retry the same plan: conflict`. The
   projection→action window is genuinely revalidated.

5. **Redundant guards hiding a gap.** Repository-identity refusal is defended at three layers, so
   single-site mutations (M13, M13b) survived and could have been misread as an untested invariant.
   The coordinated mutation (M13c) settles it: with both remedy-naming guards gone the test fails,
   and a third guard still refuses — with `ErrConflict` instead of `ErrValidation`, which the test
   correctly rejects. Defence in depth, not a hole.

6. **Would the contract block legitimate additive evolution?** No — A1 confirms a new object field on
   every graph node passes. The suite is correctly asymmetric: additive is legal, destructive is not,
   *for the payloads it covers*.

7. **Coverage the suite cannot see.** Beyond H1: the fixture repository only ever produces
   `graph_health: healthy`, `projection_health: healthy`, `problems: []`, `graph_problems: []`, so no
   degraded/broken vocabulary is pinned. I am **not** filing this as a finding, because no exact
   historical bytes exist for those states at either tag — pinning them would require reconstruction,
   which is a different (and weaker) kind of evidence. If the owner wants it, the natural follow-up is
   a task under epic 30 depending on `6g7ddeyp773z` and sequenced before `6g7fhfpmy032`, pinning
   degraded/broken projection vocabulary against reconstructed-but-labelled artifacts. Similarly,
   CRLF documents and duplicate keys are uncovered by the suite but verified sound by hand above.

8. **Does the v0.20 soak weaken the compatibility promise or silently satisfy graduation?** No. The
   promise text in `docs/THREADS_COMPATIBILITY.md:15` is unchanged in substance; what changed is the
   decision procedure, which now inserts a preview-labelled checkpoint *before* the graduation
   decision and states that passing gates "does not force" graduation. The persisted `depends_on`
   chain matches the ADR diagram, `6g7ddfhh2jc2` remains `next-up` with its own gates, and the v0.20
   task lists preview removal as out of scope.

9. **Does experimental presentation leak toward core?** No. `6g6dw5js81f3` has no dependents, is not
   referenced by the v0.20 or graduation tasks' criteria, and ADR-0006 now requires it to sit behind
   an optional adapter that core must not import. `ThreadGraphProjection`
   (`internal/wire/thread.go:207-215`) remains renderer-neutral and is consumed identically by
   `thread graph` and `thread plan`.

10. **Test maintainability.** Dates and IDs are deterministic where asserted (the composed Thread's
    minted ID and `composed_at` are deliberately not asserted; the plan's `created` == `composed_at`
    == `2026-01-04` is fixed). Failure messages name the exact JSON path
    (`v0.18.0 thread graph.projection.edges.from: …`), which made every mutation above diagnosable
    without a debugger. Runtime is negligible (~0.5s for all four tests). The one legibility problem
    is L2. Critically, the compatibility fixtures are **not** reachable by `-update`: the regeneration
    run rewrote only `testdata/golden/`, leaving `testdata/thread_compatibility/` untouched — the
    separation between deliberate historical evidence and ordinary goldens is real and I verified it.
