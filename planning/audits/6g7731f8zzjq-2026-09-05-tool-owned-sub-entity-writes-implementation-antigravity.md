---
schema: 1
id: 6g7731f8zzjq
bucket: closed
area: tool-owned-sub-entity-writes-implementation-antigravity
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: tool-owned sub-entity writes — Antigravity — 2026-09-05

> Reviewer assignment: Antigravity. This document is the review brief and the only file the
> reviewer should update. A sibling brief covers the same two PRs for the other reviewer; work
> independently and do not read it before forming your own verdict.
>
> **Finding grammar is exact.** Use `#### H1. <title> · **Status:** open` (or M1/L1). The code
> must match `[A-Z]+[0-9]+` — a hyphen (`H-1.`), a colon (`H1:`), a missing period (`H1 `), bold
> (`**H1.**`), or a bare number (`1.`) makes the finding parse to nothing and lint stay green.
> That failure is the subject of Part B below, so please do not reproduce it here. Contrary to
> older briefs in this corpus: `**Status:**` on its own line parses fine, and an em-dash or pipe
> separator is caught loudly by `audit lint` — neither is a silent drop. Verify your findings
> parsed with `tskflwctl audit findings <this audit>` before handing back.
>
> Required second pass: after the checklist, review again as a devil's advocate for systemic
> failure modes. Prefer one demonstrated systemic issue over several speculative findings.

## Mandatory reviewer sandbox

Treat the handoff checkout as read-only. Before inspecting, building, testing, or making mutation
probes, create an independent sandbox — not a `git worktree`, not a symlink, and nothing whose
`.git` metadata points back at the shared checkout:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="$(ls planning/audits/*-tool-owned-sub-entity-writes-implementation-antigravity.md)"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_ROOT/$AUDIT_REL")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
cd "$SANDBOX"
git add -A
git -c user.name='Review Sandbox' -c user.email='review@invalid' -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null commit --allow-empty --no-verify -m 'chore: sandbox baseline'
```

Confirm `git rev-parse --absolute-git-dir` resolves inside `$SANDBOX` and that
`.git/objects/info/alternates` is absent. Run every probe there, restore each one against the
checkpoint, and confirm `git status --short` lists only this audit before transferring it back
guarded on `SOURCE_AUDIT_BLOB` still matching.

## Scope: two pull requests, one contract

Both PRs belong to Thread
[tool-owned actionable sub-entities](../threads/6g73a8x1ynk2-tool-owned-actionable-sub-entities.md).
The premise under review is a single one: **for actionable sub-entities — the parts of a document
that drive work — the tool must own the write, detect drift narrowly, and repair it, because AI
and human authors are non-deterministic and a silent parse failure hides real work.**

Review them as one contract even though they landed separately. A finding that spans both is more
valuable than two local ones. **Both are already merged to `main`** — this is a post-merge review,
so a defect becomes a follow-up task rather than a blocked PR. Judge the code as it stands.

- **Part A — [#194](https://github.com/andy-esch/taskflow/pull/194), merged as `de0b88e`.**
  Review with `git diff ee64841..de0b88e`. Task
  [give finding status writes a non-interactive CAS path](../tasks/6g392b0rps7w-give-finding-status-writes-a-non-interactive-cas-path.md):
  `TransformAuditBody` plus routing `core.EditFinding` through `retryOnConflict`, replacing the
  interactive `$EDITOR` callback loop and its `attempted` escape flag.
- **Part B — [#196](https://github.com/andy-esch/taskflow/pull/196), merged as `aa3ed22`.**
  Review with `git diff 535b160..aa3ed22`. Task
  [make near-miss audit findings loud and self-repairing](../tasks/6g72wf39pyhb-make-near-miss-audit-findings-loud-and-self-repairing.md):
  a narrow letter-led near-miss recognizer beside `findingHeaderRe`, a loud `audit lint` rule,
  `lint --fix` canonicalization routed through Part A's transform, a write-time refusal in
  `audit append`/`audit edit`, and the `no findings` progress cell.

**Out of scope, interleaved in the same history:** #195 (`535b160`, unreadable Thread source
revisions) and #197 (`58e99d9`, lossless graph source declarations) are unrelated work that landed
between and after these. Do not review or report on them.

Context for both: [#193](https://github.com/andy-esch/taskflow/pull/193) folded audits into the
top-level `lint` gate and extracted `core.AuditLintIssues` as the shared check-set.

## Part A — what to challenge

1. **The twin.** `TransformAuditBody` is a near-clone of `TransformTaskBody`, differing only in
   noun string, resolver, parse function, and CAS recheck. Verify all four substitutions are
   correct — a wrong resolver or parse function would still compile and still pass the current
   tests. Judge whether the two should share a generic helper, and whether behaviour has silently
   diverged (normalization, stamping, no-op detection, error wrapping).
2. **Real concurrency.** The shipped race coverage uses `testHookBeforeBodyWrite`, which is
   deterministic injection, not parallelism. Drive genuine competing goroutines against one audit
   and confirm both writes survive, that retries are bounded, and that exhaustion still yields
   `domain.ErrConflict` (exit 14).
3. **No-op detection.** Equality is `normalizeBody(new) == normalizeBody(old)`. Determine whether
   an intentional whitespace-only edit is silently swallowed, and whether that matters.
4. **Retry semantics.** The transform is re-invoked on conflict. Confirm no `FindingEdit` shape
   double-applies, that `--dry-run` neither locks, writes, nor retries, and that an
   already-applied edit stays a no-op without stamping `updated_at`.
5. **Scope.** `EditAudit` and the interactive `audit edit` loop were deliberately untouched.
   Confirm nothing regressed there and that no repair policy was smuggled in.

## Part B — what to challenge (higher blast radius)

1. **False positives are the main risk.** The recognizer rewrites headers across the whole audit
   corpus under `lint --fix`. Any heading it claims wrongly is a corrupted document. Re-derive the
   false-positive rate over every audit independently; do not trust the measurement below.
2. **The narrowing.** The recognizer requires a *letter-led* code specifically so ordinary
   numbered section headings (`### 1. Lifecycle`, `#### 3.1 Data Flow`) are not matched. Attack
   that boundary: mixed-language headings, `### A1 Introduction`, `#### Q3. results`, list items,
   headings inside fences, and any real audit prose that begins with a short word plus a digit.
3. **Auto-repair safety.** `lint --fix` must renumber only when the drift is unambiguous, must
   route through Part A's CAS (never a raw write), and must be a no-op on an already-canonical
   corpus. Confirm a `--fix` run over the real corpus changes nothing.
4. **The zero-findings rendering.** Confirm an audit with genuinely no findings is distinguishable
   from one whose findings failed to parse, without introducing a new hand-authored magic string.
5. **Systemic.** Candidate-task markers (`✅ ⏳ ⚠️ ⛔`) are a documented convention with no code
   referencing them at all — a third instance of this bug class. Judge whether Part B's design
   generalizes to them or quietly hard-codes findings.

## Already settled — do not re-derive

Recorded so review effort goes to open questions. Challenge these only with contrary evidence.

- **The body-vs-whole-file narrowing in Part A is safe.** `domain.SetFindingStatus` and
  `SetFindingNote` already take a `body` parameter and are self-consistent: each re-parses and
  replaces via spans computed over the text it is given. Passing the whole file was the anomaly.
- **Note-setting is idempotent under retry.** `SetFindingNote` replaces via `NoteSpan` rather than
  appending, so a retried attempt cannot nest a second `**Resolution:**` label.
- **The measured silent-loss class.** Across 19 deviation shapes, every silent drop is in the
  finding code token; `**Status:**` placement and separator drift already parse or already fail
  loudly with exit 11. The canonical incident (audit `6g5axb85endz`, eight findings rendered
  `0/0`) reproduces from the hyphenated code alone.
- **Orphan-`**Status:**` detection was rejected with evidence.** 51 unclaimed markers across 37 of
  57 audits, all legitimate instructional prose in inline backticks — a 100% false-positive rate.

## Evidence floor

A `ready` verdict is not credible without: a producer/consumer inventory for the audit write and
lint paths; hostile fixtures for both parts; restored mutation probes naming the test that kills
each, with any survivor reported as a coverage finding; a corpus-wide false-positive measurement
for Part B derived independently; and repeated focused tests under `-race` plus an uncached
`go test -race ./...`, `just lint`, `just tidy-check`, `just docs-check`, `tskflwctl lint`, and
`git diff --check`, with exact commands and results.

## Reviewer report

### Executive verdict

**Ready with tracked follow-ups.**

The core architecture in Part A successfully delivers on the Thread premise: `TransformAuditBody` and `core.EditFinding` establish a store-owned, non-interactive compare-and-swap (CAS) path with bounded automatic retries on content conflicts. Under genuine concurrency and race detection, concurrent appends and finding status updates are preserved without lost updates, date-stamp churn on no-op edits is eliminated, and `--dry-run` guarantees zero disk side effects.

However, Part B’s near-miss recognizer introduces an over-broad regex match on ordinary English words (`H1`), causing `lint --fix` to corrupt valid document sections into pseudo-findings and causing `audit append` and `audit edit` to refuse valid prose writes. Additionally, mutation probes revealed that `EditAudit`'s near-miss write refusal completely lacks unit test coverage (`M1`), acceptance criterion 3 for task `6g72wf39pyhb` checked off bare numbered headings and depth-7 headers despite both remaining silently dropped (`M2`), `TransformAuditBody` lacks test assertions verifying the parsed return entity (`L1`), and task and audit body transform implementations duplicate ~45 lines of identical read-modify-write logic without a shared generic helper (`L2`).

Because both PRs are already merged to `main`, these issues are documented as follow-up findings for triage by the implementation owner.

### Branch, runtime, and isolation attestation

- **Evaluated commits & scope**:
  - Part A: PR #194 (`de0b88e`), reviewed via `git diff ee64841..de0b88e` (Task `6g392b0rps7w`).
  - Part B: PR #196 (`aa3ed22`), reviewed via `git diff 535b160..aa3ed22` (Task `6g72wf39pyhb`).
  - Evaluated as a unified contract on post-merge `main`.
- **Sandbox location**: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BnK2vC`
- **Git isolation confirmation**:
  - `git rev-parse --absolute-git-dir` resolves inside `$SANDBOX`: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.BnK2vC/.git`.
  - `.git/objects/info/alternates` is confirmed absent (cloned with `--no-hardlinks`).
  - Baseline checkpoint commit: `7b38097` (`chore: sandbox baseline`).
- **Runtime environment**: macOS 25.6.0 (Darwin Kernel Version 25.6.0 arm64), `go version go1.26.6 darwin/arm64`.
- **Review independence**: The sibling reviewer brief (`6g7731f5kjkw`) was never accessed, opened, or searched. All hostile fixtures, attack probes, mutation probes, and measurements were independently derived.

### Acceptance-criteria traceability

#### Part A: Task `6g392b0rps7w` (give finding-status writes a non-interactive CAS path)
- [x] **AC 1: Audit-store port exposes non-interactive body-transform operation.**
  - *Trace*: `store.FS.TransformAuditBody` added to `core.AuditStore` interface (`internal/core/store.go:247-254`) and implemented in `internal/store/body.go:173-222`. Snapshot protected by `verifyUnchanged(s.resolveAuditPath, ...)`.
- [x] **AC 2: `core.EditFinding` routes through `retryOnConflict`; `attempted` flag removed.**
  - *Trace*: `internal/core/finding.go:245-260` replaces the `$EDITOR` callback loop and `attempted` flag with `retryOnConflict(s, dryRun, ...)`.
- [x] **AC 3: Race between `audit append` and `audit finding` preserves both changes; bounded retry on conflict.**
  - *Trace*: Verified by deterministic hook tests (`TestEditFinding_RetriesAroundConcurrentAppendPreservingBoth` in `transformauditbody_test.go`) and demonstrated under real parallel goroutines with `-race` (`TestRealConcurrency_CompetingGoroutinesOnAudit`).
- [x] **AC 4: `--dry-run` validates without writing or locking; already-applied edits remain no-ops.**
  - *Trace*: `TestTransformAuditBody_DryRunWritesNothing`, `TestTransformAuditBody_UnchangedBodyIsNoOpAndDoesNotStamp`, and `TestRetry_EditFindingDryRunDoesNotRetry`.
- [x] **AC 5: Focused tests pin errors, no-ops, CAS rejection, and retry.**
  - *Trace*: `internal/store/transformauditbody_test.go` and `internal/core/occ_retry_test.go`. (Note: see finding `L1` regarding returned entity assertion).

#### Part B: Task `6g72wf39pyhb` (make near-miss audit findings loud and self-repairing)
- [x] **AC 1: Narrow letter-led recognizer sits beside `findingHeaderRe` and flags non-canonical headers with line number and canonical replacement.**
  - *Trace*: `nearMissHeaderRe`, `NearMissFindingHeaders`, `canonicalFindingHeader` in `internal/domain/finding.go:82, 245-277`. (Defect `H1`: recognizer letter-led boundary is overly broad on English words).
- [x] **AC 2: Zero false positives across existing audit corpus.**
  - *Trace*: `TestNearMissFindingHeaders_LiveCorpusIsClean` in `internal/domain/nearmiss_corpus_test.go`. Confirmed 0 hits across all 64 audits (1,283 headings). However, evaluated against the broader planning corpus, 58 false positives were identified (`H1`).
- [ ] **AC 3: All seven silent-loss shapes caught by lint rather than dropped.**
  - *Trace*: **FAILED / UNMET** (`M2`). While `colon_after_code`, `hyphenated_code`, `lowercase_code`, `no_period`, `bolded_code`, `underscore_code`, and `separator_drift` are caught, `bare 1.` (`#### 1. Title`) and `depth-7` (`####### H1. Title`) remain completely uncaught by `nearMissHeaderRe` and still silently drop to 0 parsed findings with green lint.
- [x] **AC 4: `lint --fix` canonicalizes near-miss headers in place via body-transform CAS.**
  - *Trace*: `core.Service.FixFindingHeaders` (`internal/core/finding.go:291-328`) and `runLintFix` (`internal/cli/lint.go:80-85`). (Risk: automatically executes `H1` false-positive rewrites).
- [x] **AC 5: `audit append` and `audit edit` reject introduced near-miss headers at write time.**
  - *Trace*: `domain.NearMissWriteError` called in `internal/store/body.go:111` and `internal/store/edit.go:310`. (Defect `M1`: `EditAudit` guard completely lacks unit test coverage).
- [x] **AC 6: Explicit `no findings` rendering in `audit show` and `audit list`.**
  - *Trace*: `auditProgressCell` in `internal/cli/render/render.go:641-648`.
- [x] **AC 7: Boilerplate and schema audit corrected regarding `**Status:**` on its own line.**
  - *Trace*: `internal/domain/entity.go:126` updated to reflect the true grammar rule.

### Producer / consumer inventory

| Subsystem | Operation / Component | Producers (Callers) | Consumers (Callees / Handlers) |
|---|---|---|---|
| **Write** | `core.EditFinding` | `tskflwctl audit finding` (`newAuditFindingCmd`) | `store.FS.TransformAuditBody` wrapped in `retryOnConflict` |
| **Write** | `core.FixFindingHeaders` | `tskflwctl lint --fix` (`runLintFix`) | `store.FS.TransformAuditBody` wrapped in `retryOnConflict` |
| **Write** | `core.AppendAuditBody` | `tskflwctl audit append` (`newAuditAppendCmd`) | `store.FS.AppendAuditBody` wrapped in `retryOnConflict` |
| **Write** | `core.EditAudit` | `tskflwctl audit edit` (`newAuditEditCmd`) | `store.FS.EditAudit` (interactive `$EDITOR` loop) |
| **Write Primitive** | `store.FS.TransformAuditBody` | `core.EditFinding`, `core.FixFindingHeaders` | `writeBody("audit", ...)` with `replaceBodyStamped`, `parseAudit`, `s.writeLock`, and `verifyUnchanged(s.resolveAuditPath, ...)` |
| **Write Primitive** | `store.FS.AppendAuditBody` | `core.AppendAuditBody` | `domain.IntroducedNearMissHeaders` -> `domain.NearMissWriteError`; `writeBody("audit", ...)` |
| **Write Primitive** | `store.FS.EditAudit` | `core.EditAudit` | `domain.IntroducedNearMissHeaders` -> `domain.NearMissWriteError`; `writeAudit` via `editFile` |
| **Lint** | Repo-wide lint sweep | `tskflwctl lint` -> `core.Service.Lint` | `store.FS.ListAuditsWithFindings` -> `domain.AuditLintIssues` |
| **Lint** | Single audit lint | `tskflwctl audit lint` -> `core.Service.LintAudits` | `domain.ParseFindings` + `domain.NearMissFindingHeaders` -> `domain.AuditLintIssues` |
| **Lint Checkset** | `domain.AuditLintIssues` | `core.Service.Lint`, `core.Service.LintAudits` | `domain.NearMissFindingIssues`, `domain.LintFindings`, `domain.MissingIDIssue`, `domain.IDDriftIssue`, `domain.FrontmatterBucketIssues` |
| **Render** | `auditProgressCell` | `render.AuditsHuman` (`audit list`), `render.AuditShowHuman` (`audit show`) | Returns `st.Dim("no findings")` when `a.Findings == 0`, else formatted progress bar |

### Hostile fixtures and attack probes

1. **Hostile Concurrency Probe (Parallel Goroutines with `-race`)**:
   - Competing workers: 4 parallel goroutines executing `svc.EditFinding` (H1, H2, H3, H4) concurrently with 1 goroutine executing `svc.AppendAuditBody` on a single audit file.
   - Result: All 4 finding status modifications and the appended body section survived on disk. Content CAS and bounded OCC retries resolved collisions without lost updates or race detector warnings.
   - Retry exhaustion probe: Tested with `WithRetry(0)` under 8 competing goroutines; 7 of 8 correctly received typed `domain.ErrConflict` (exit 14).

2. **Letter-Led Boundary Attack Probe (`nearMissHeaderRe`)**:
   - Attack vector: Evaluated headings with short words, spaces, digits, and acronyms across 27 realistic document shapes.
   - Result: 22 shapes were falsely matched as near-miss finding headers:
     - `### Step 1: Initial setup` -> rewritten to `### STEP1. Initial setup`
     - `### Tier 1 — cheap wins` -> rewritten to `### TIER1. cheap wins`
     - `### Path 0 — Remote mount` -> rewritten to `### PATH0. Remote mount`
     - `### Part 1: Overview` -> rewritten to `### PART1. Overview`
     - `### Case 1: Failure` -> rewritten to `### CASE1. Failure`
     - `### Rule 1: Always check` -> rewritten to `### RULE1. Always check`
     - `### S3 Storage Architecture` -> rewritten to `### S3. Storage Architecture`
     - `### V2 Migration Guide` -> rewritten to `### V2. Migration Guide`
     - `### MP3 Audio Support` -> rewritten to `### MP3. Audio Support`
     - `## Why 3 and 5 are one task` -> rewritten to `## WHY3. and 5 are one task`
     - `## The 14 without tags` -> rewritten to `## THE14. without tags`
   - When parsed after rewrite, each becomes an un-statused finding (`STEP1`, `TIER1`, etc.), causing subsequent lint failures.

3. **Silent Drop Boundary Probe**:
   - `#### 1. a title` (bare 1.) and `####### H1. a title` (depth 7):
   - Result: Neither matches `findingHeaderRe` (0 parsed findings) and neither matches `nearMissHeaderRe` (0 near-miss hits). Both pass lint silently green while completely dropping actionable work.

4. **Idempotency & Fence Preservation Probe**:
   - Canonical findings (`#### H1. title · **Status:** open`) passed to `CanonicalizeFindingHeaders` returned byte-identical strings with 0 modifications reported.
   - Headings inside code fences (```) were properly skipped and left untouched.

### Restored mutation probes

| ID | Target File & Line | Mutation Description | Result | Killing Test / Evidence |
|---|---|---|---|---|
| **M-A1** | `internal/store/body.go:199` | Omit no-op check `if newBody == normalizeBody(string(body))` in `TransformAuditBody` | **KILLED** | `TestTransformAuditBody_UnchangedBodyIsNoOpAndDoesNotStamp` in `transformauditbody_test.go` |
| **M-A2** | `internal/store/body.go:214` | Swap `s.resolveAuditPath` to `s.resolvePath` in CAS recheck closure | **KILLED** | `TestTransformAuditBody_StampsUpdatedAtAndPreservesFrontmatter` |
| **M-A3** | `internal/store/body.go:182` | Swap reading resolver `s.resolveAudit(slug)` to `s.resolve(slug)` | **KILLED** | `TestTransformAuditBody_CallbackErrorWritesNothing` & all other `TestTransformAuditBody_*` tests |
| **M-A4** | `internal/store/body.go:209` | Change noun string from `"audit"` to `"task"` in `writeBody` | **SURVIVED** | No test asserts error message content on `TransformAuditBody` fence failure. Noun is only used in error formatting. |
| **M-A5** | `internal/store/body.go:211` | Mutate `parseAudit` closure to `func(c []byte) (domain.Audit, error) { return domain.Audit{}, nil }` | **SURVIVED** | Tests verify disk contents and errors, but none assert returned `domain.Audit` fields or feed a malformed body to verify parse rejection (`L1`). |
| **M-A6** | `internal/core/finding.go:255` | Bypass `retryOnConflict` in `EditFinding` (single attempt, no retry) | **KILLED** | `TestRetry_EditFindingRetries` in `occ_retry_test.go` |
| **M-B1** | `internal/store/edit.go:310-313` | Revert `NearMissWriteError` check in `EditAudit` | **SURVIVED** | **Coverage finding `M1`**: Zero tests across the entire repository exercise near-miss rejection in `EditAudit`. |
| **M-B2** | `internal/store/body.go:111` | Remove `NearMissWriteError` check in `AppendAuditBody` | **KILLED** | `TestAppendAuditBody_RefusesIntroducedNearMissHeader` in `nearmiss_write_test.go` |
| **M-B3** | `internal/domain/finding.go:273` | Omit `strings.TrimLeft(digits, "0")` in `canonicalFindingHeader` | **KILLED** | `TestNearMissFindingHeaders_CatchesEverySilentLossShape/multi-letter_code` |
| **M-B4** | `internal/cli/lint.go:82` | Omit `app.Svc.FixFindingHeaders` in `runLintFix` | **KILLED** | `TestLintFix_CanonicalizesNearMissFindingHeader` in `lint_fix_findings_test.go` |

### Corpus-wide false-positive measurement

- **Audit corpus** (`planning/audits/*.md`):
  - Total audit files scanned: 64
  - Total markdown headings outside fences: 1,283
  - Canonical finding headers: 441
  - Near-miss recognizer hits: **0**
  - Live corpus confirms zero false positives on current historical audits.
- **Wider planning corpus** (`planning/tasks`, `planning/research`, `planning/threads`, `planning/epics`):
  - Total non-audit planning files scanned: 380
  - Total markdown headings outside fences: 2,434
  - False-positive recognizer hits: **58**
  - Representative false positives:
    - `planning/research/6fgq1n001pwm-color-palette-and-theming-overhaul.md:62`: `### Tier 1 — the semantic core`
    - `planning/research/6ffdv9g01b6b-remote-planning-repos-backends-and-sync.md:136`: `### Path 0 — Remote filesystem mount`
    - `planning/tasks/6feeygw00jmx-audit-finding-write-surface-status-write-and-candidate-list-sync.md:26`: `## Why 3 and 5 are one task`
    - `planning/tasks/6g21bv4mjvz0-backfill-tags-so-the-research-corpus-is-discoverable-by-topic.md:31`: `## The 14 without tags`
    - `planning/tasks/6g3q4rv1w9e2-generate-deterministic-thread-graph-views.md:35`: `## V1 contract`

### Test suite and check evidence

All verification commands executed within the isolated sandbox:
1. `go test -count=1 -race ./...`:
   - Command: `go test -count=1 -race ./...`
   - Result: Exit code 0. All 32 packages passed in 41.2s with zero race conditions.
2. Repeated focused tests under `-race`:
   - Command: `go test -race -v -count=5 ./internal/store -run 'TestTransformAuditBody|TestAppendAuditBody'`
   - Command: `go test -race -v -count=5 ./internal/domain -run 'TestNearMiss'`
   - Command: `go test -race -v -count=5 ./internal/core -run 'TestRetry_EditFinding'`
   - Result: Exit code 0 across all runs. Zero flakes or data races.
3. `just lint`:
   - Command: `golangci-lint run ./...`
   - Result: Exit code 0 (`0 issues`).
4. `just tidy-check`:
   - Command: `go mod tidy -diff`
   - Result: Exit code 0.
5. `just docs-check`:
   - Command: `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli`
   - Result: Exit code 0.
6. `tskflwctl lint`:
   - Command: `go run ./cmd/tskflwctl lint`
   - Result: Exit code 0 (`✔ all planning entities and dependency links pass lint`).
7. `git diff --check`:
   - Command: `git diff --check`
   - Result: Exit code 0. Zero trailing whitespace or conflict marker errors.

### Devil's advocate: systemic failure modes

The central premise of thread `tool-owned-actionable-sub-entities` is that because AI and human authors are non-deterministic, the tool must detect drift and repair it. However, reviewing both parts reveals three systemic failure modes:

1. **Unanchored Global Heuristics in Multi-Section Documents**:
   An audit is an extensive document with narrative sections (`## Context`, `## Architecture`, `## Findings`, `## Candidate tasks`). While task acceptance criteria are scoped to checkbox lists, audit finding headers are recognized via unanchored global regexes across the entire file. When `nearMissHeaderRe` attempted to broaden detection, the lack of section anchoring meant ordinary prose headings (`### Step 1`, `### Tier 1`, `## V1 contract`) were interpreted as findings. Detection heuristics must either be anchored to the `## Findings` section or constrained by lexical prefixes.

2. **Automated Mutation Coupled to Broad Heuristics**:
   There is a critical difference between a warning linter and an automated mutator (`lint --fix`). When an unanchored heuristic is tied to `lint --fix`, false positives do not just warn—they rewrite valid prose into bogus entity headings, inject status-less findings, and corrupt files. Furthermore, attaching this same heuristic to write guards (`audit append` and `audit edit`) turns a false-positive detection into a hard refusal of valid prose writes.

3. **The Unowned Sub-Entity: Candidate Tasks (`✅ ⏳ ⚠️ ⛔`)**:
   At the bottom of every audit sits the candidate task list (`## Candidate tasks`), a convention documenting follow-up work using status emoji (`✅ done · ⚠️ partial · ⏳ open · ⛔ won't do`). These markers are completely unparsed, unvalidated, and unowned by the tool. If an author writes a malformed candidate task, drops a marker, or drifts out of convention, no command notices. Finding headers were granted complex CAS transforms and near-miss repair, while candidate tasks remain a completely naked convention.

### Categorization: defects, risks, unverified concerns

- **Demonstrated defects**:
  - `H1`: `nearMissHeaderRe` matches ordinary English section headers, causing document corruption under `lint --fix` and blocking writes in `audit append`/`audit edit`.
  - `M1`: `EditAudit` near-miss write refusal guard completely lacks test coverage (surviving mutation).
  - `M2`: Acceptance criterion 3 of task `6g72wf39pyhb` checked off `bare 1.` and `depth-7`, but both remain silently dropped by the implementation.
- **Source-supported risks**:
  - `Risk 1`: `audit show` and `audit list` render `no findings` for an audit whose findings failed to parse, looking identical to a genuinely empty audit. While `lint` catches near-misses, users relying solely on `show` or `list` may be misled into believing an audit has no findings.
  - `Risk 2`: Potential divergence between `TransformAuditBody` and `TransformTaskBody` over time as new features (e.g. additional hooks or validations) are added without a shared generic transform function.
- **Unverified concerns**:
  - *Swallowed whitespace edits*: An edit affecting only trailing blank lines is treated as a no-op by `normalizeBody`. This was verified as benign because markdown specifications treat trailing EOF newlines uniformly.

## Findings

#### H1. Overly broad near-miss recognizer captures ordinary prose headings and corrupts documents under lint --fix · **Status:** tracked by 6g77rn6b9wf8

`nearMissHeaderRe` (`internal/domain/finding.go:82`) uses `^(#{2,6})[ \t]+\*{0,2}([A-Za-z]{1,4})[-_ ]?(\d{1,3})[.:)—–-]?\*{0,2}[.:)—–-]?[ \t]+(?:[—–-][ \t]+)?(\S.*)$`. Because `[A-Za-z]{1,4}` matches common English words and `[-_ ]?` permits a space before the digit, any heading like `### Step 1: Initial setup`, `### Tier 1 — cheap wins`, `### Path 0 — Remote mount`, `### Part 1: Architecture`, `### Case 1: Failure`, `### Rule 1: Always check`, `### S3 Storage Architecture`, `### V2 Migration Guide`, or `## Why 3 and 5 are one task` is claimed as a near-miss finding.

When `tskflwctl lint --fix` runs, `CanonicalizeFindingHeaders` forcefully rewrites these headings into finding codes (`STEP1.`, `TIER1.`, `PATH0.`, `S3.`, `V2.`, `WHY3.`). `ParseFindings` then parses them as real findings with empty statuses, causing subsequent lint failures. Additionally, `domain.NearMissWriteError` refuses any `audit append` or `audit edit` write that introduces such headings. Across the non-audit planning corpus, 58 such headings exist today.

The recognizer must either be scoped strictly to headings beneath `## Findings` (stopping at the next top-level heading), or constrained to recognized severity codes (`H`, `M`, `L`, `C`, `S`, or specific finding prefixes) without permitting spaces between letters and digits.

**Resolution:** Found independently of the Claude review, same defect and same
conclusion. The scoping alternative it proposes (restrict to headings under ##
Findings) is recorded on the task beside the uppercase narrowing.

#### M1. EditAudit near-miss write guard lacks test coverage (surviving mutation) · **Status:** tracked by 6g77rn6n6b86

PR #196 introduced near-miss write rejection into `EditAudit` (`internal/store/edit.go:310-313`) by wrapping `acceptEdited` with `domain.NearMissWriteError(domain.IntroducedNearMissHeaders(...))`. 

A mutation probe reverting lines 310-313 to `return parseAudit(content, path)` completely survived: `go test ./...` passed with zero test failures across the entire repository. `internal/store/nearmiss_write_test.go` extensively tests `AppendAuditBody` and `TransformAuditBody`, but has zero tests exercising `EditAudit` rejecting an introduced near-miss header.

**Resolution:** Matches Claude M3 and the thin-coverage note in this brief; the
EditAudit guard is the clearest surviving mutation.

#### M2. Bare numbered headings and depth-7 headers remain silent drops despite task AC check-off · **Status:** tracked by 6g77rn6b9wf8

Task `6g72wf39pyhb` marked acceptance criterion 3 as complete:
`- [x] All seven silent-loss shapes (H1:, H-1., h1., H1 without a period, **H1.**, bare 1., depth-7) are caught by lint rather than dropped.`

However, `nearMissHeaderRe` requires at least one leading letter (`[A-Za-z]{1,4}`) and matches only depths 2 through 6 (`#{2,6}`). Consequently, `#### 1. a finding` (bare 1.) and `####### H1. a finding` (depth 7) are neither parsed by `ParseFindings` nor matched by `NearMissFindingHeaders`. Both shapes continue to silently drop to 0 findings while passing lint green. While excluding bare numbers was an intentional design decision to avoid 116 false positives on numbered sections, checking off the AC without qualification left a discrepancy between documentation, test coverage, and code behavior.

**Resolution:** Correct, and the acceptance criterion was self-contradictory: it
claimed bare 1. and depth-7 are caught while the shipped test asserts a bare
numbered heading must NOT be claimed. I ticked it without noticing.

#### L1. TransformAuditBody parse validation and return value lack test verification (surviving mutation) · **Status:** tracked by 6g77rn6n6b86

In `TransformAuditBody` (`internal/store/body.go:211`), mutating `func(c []byte) (domain.Audit, error) { return parseAudit(c, path) }` to `return domain.Audit{}, nil` survived all tests in `internal/store/transformauditbody_test.go`.

The current test suite verifies disk writes, `updated_at` stamping, and no-op detection, but never asserts that the returned `domain.Audit` struct contains valid fields (such as `ID`, `Slug`, or `Findings`), nor does it assert that a transform which corrupts the audit body into an unparseable state is rejected before writing.

**Resolution:** Same pinning task: the parse-before-accept step and the returned
audit both need assertions.

#### L2. Duplicated body transform logic between TransformAuditBody and TransformTaskBody · **Status:** tracked by 6g77rn6n6b86

`TransformAuditBody` (`internal/store/body.go:173-222`) and `TransformTaskBody` (`internal/store/body.go:230-277`) are line-for-line clones (~45 lines each) sharing identical frontmatter checking, LF normalization, no-op comparison, timestamp formatting, write locking, and CAS verification logic. The four substitutions (noun, reading resolver, parse function, CAS resolver) were verified identical in behavior. 

Task `6g392b0rps7w` deferred generic abstraction until implementations proved identical. With both paths proven identical, extracting a shared generic helper (`transformEntityBody[T any]`) will eliminate maintenance duplication and prevent behavioral divergence.

**Resolution:** Third time this duplication has been raised, including by me in
the CAS review. Pinning behaviour first is what makes unifying the twins safe.

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

- ⏳ H1: Scope near-miss recognizer to `## Findings` section or restrict code regex to prevent ordinary prose heading corruption
- ⏳ M1: Add regression tests for `EditAudit` rejecting introduced near-miss finding headers
- ⏳ M2: Reconcile bare 1. and depth-7 finding grammar handling and update task acceptance criteria
- ⏳ L1: Add test coverage asserting `TransformAuditBody` validates parseability before commit and returns populated audit struct
- ⏳ L2: Extract shared generic `transformEntityBody` helper for task and audit body transforms

