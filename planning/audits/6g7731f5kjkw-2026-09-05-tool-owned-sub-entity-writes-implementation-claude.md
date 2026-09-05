---
schema: 1
id: 6g7731f5kjkw
bucket: closed
area: tool-owned-sub-entity-writes-implementation-claude
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: tool-owned sub-entity writes — Claude — 2026-09-05

> Reviewer assignment: Claude. This document is the review brief and the only file the
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
AUDIT_REL="$(ls planning/audits/*-tool-owned-sub-entity-writes-implementation-claude.md)"
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

## Deliverable

Replace the reviewer-report placeholder with an executive verdict (`ready`,
`ready with tracked follow-ups`, or `not ready`); branch/base/checkpoint, runtime, and isolation
attestation; findings grouped by severity using the exact grammar above with `**Status:** open`;
acceptance-criteria traceability; and an explicit separation of demonstrated defects,
source-supported risks, and unverified concerns. If there are no findings, say so plainly and
still supply the evidence. Do not pre-resolve findings — the implementation owner triages them
with `tskflwctl audit finding`.

## Reviewer report

### 1. Executive verdict

**not ready** — scoped: **Part A is ready as it stands; Part B's repair and render halves are
not.**

Part A survives everything I threw at it. Eight goroutines stamping eight findings on one audit,
under `-race`, `-count=5`: zero stamps lost, appends and stamps both survive, the file still
parses, and exhaustion under live contention yields only `domain.ErrConflict`. All four
substitutions in the `TransformAuditBody` twin are correct — I mutated the resolver, the CAS
resolver, the parse function and the no-op check and the shipped suite killed each one.

Part B's *detection* is also sound: the recognizer catches every drift shape the corpus incident
produced, it is fence-aware across all six fence shapes I tried, it is idempotent, and its
false-positive rate over the 64-audit corpus is genuinely zero — I re-derived that independently
rather than trusting the shipped fixture.

What is not ready is what Part B does with a hit. Two of the three headline defects strike the
Thread's own premise rather than its edges:

- `lint --fix` **rewrites ordinary English prose headings into fabricated finding codes**, and
  then reports the fabrications as findings missing a status (H1). I demonstrated four such
  rewrites end-to-end on real heading text taken from this repository. The "zero false positives"
  result that justified the narrowing was measured on the one document class in this repo whose
  headings are entirely tool-scaffolded; the same repo's hand-authored documents contain 17
  headings the recognizer would corrupt.
- The primary triage surface now **asserts `no findings` over an audit carrying eight real,
  evidence-backed, drifted findings** (H2) — byte-identical to a genuinely empty audit, on
  `audit list`, `audit show` and `--json`. Relative to the `0/0` incident this Thread exists to
  fix, that is a more confident false claim, not a less confident one.
- `audit new --body` accepts the exact header `audit append` refuses (H3), so the guard that AC5
  says "cannot be forgotten or bypassed" is bypassed by the first write an audit ever gets.

None of this is corrupting anything today: the live corpus is clean, `lint --fix` over it is a
true no-op, and every gate is green. The verdict is about the contract, not about present damage
— and specifically about generalizing this pattern to further sub-entities, which is the Thread's
stated direction.

### 2. Isolation, environment, and validation attestation

| | |
|---|---|
| Shared checkout | `/Users/andyeschbacher/git/andy-esch/taskflow` @ `58e99d9` — **not modified, not built in, not tested in** |
| Sandbox | `/private/var/folders/…/T/taskflow-review.dEm3R9` (`mktemp -d`) |
| Creation | `git clone --no-hardlinks` + `rsync -a --delete --exclude=.git`, per the brief |
| Checkpoint | `30e7e9ab2d89e998cb3afbe253fba30566ee0725` ("chore: sandbox baseline") |
| `git rev-parse --absolute-git-dir` | `…/taskflow-review.dEm3R9/.git` — inside the sandbox |
| `.git/objects/info/alternates` | absent; `.git` is a standalone 26M object store |
| Not a worktree / not a symlink | `git worktree list` shows one entry, the sandbox itself |
| Remote | `origin` **removed** after clone, so no path can reach the shared checkout |
| Runtime | go1.26.6 darwin/arm64, binary `v0.19.0-34-g30e7e9a` |
| Part A range | `git diff ee64841..de0b88e` (merge `de0b88e`, PR #194) |
| Part B range | `git diff 535b160..aa3ed22` (merge `aa3ed22`, PR #196) |

Every build, test, mutation and probe below ran in the sandbox. Throwaway planning spaces for the
CLI probes were created outside both trees, under the session scratchpad. All mutations were
restored with `git checkout --` and verified with `git status --short` before the gates were run.

Gates, all from the restored checkpoint:

| Command | Result |
|---|---|
| `go clean -testcache && go test -race ./...` | **PASS**, 25 packages, exit 0, no `DATA RACE` |
| `go test -race -count=10 ./internal/{domain,store,core,cli/...}` | **PASS**, exit 0, no `DATA RACE` |
| `just lint` (golangci-lint) | **0 issues** |
| `just tidy-check` | clean |
| `just docs-check` | clean |
| `git diff --check` | clean |
| `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint` |
| `./bin/tskflwctl lint --fix` over the real corpus | `nothing to fix`, `git status` unchanged — B3 confirmed |

### 3. Producer / consumer inventory

**Producers — every path that writes an audit file:**

| Path | Content CAS | Near-miss guard |
|---|---|---|
| `FS.CreateAudit` (`create.go:217`) — `audit new`, incl. `--body`/`--body-file` | n/a (O_EXCL create) | **none — see H3** |
| `FS.AppendAuditBody` (`body.go:90`) — `audit append` | yes, via `writeBody` | yes (`body.go:112`) |
| `FS.TransformAuditBody` (`body.go:173`) — `audit finding`, `lint --fix` | yes, via `writeBody` | none, **deliberate** (it is the repair path; `TestTransformAuditBody_RepairPathIsNotBlockedByTheWriteGuard`) |
| `FS.EditAudit` (`edit.go:283`) — `audit edit` | yes, via `editFile` | yes (`edit.go:312`), **untested — see M3** |
| `FS.MoveAudit` (`auditstore.go:148`) — close/reopen/defer | yes | n/a (frontmatter only) |
| `FS.FixFrontmatter` (`fix.go:92,117`) — `lint --fix` | yes | n/a (body passed through untouched) |

All six route through `writeFileAtomic`/`createFileAtomic`; I found no raw body write bypassing
them.

**Consumers of the near-miss surface:**

| Consumer | Call site |
|---|---|
| `parseAuditWithFindings` → `AuditWithFindings.NearMisses` (one-read sweep) | `auditstore.go:243` |
| `core.AuditLintIssues` → `NearMissFindingIssues` (both `audit lint` and top-level `lint`) | `finding.go:274` |
| `core.LintAudits` (single-audit path, derives from the body it fetched) | `finding.go:336` |
| `core.FixFindingHeaders` → `CanonicalizeFindingHeaders` (`lint --fix`) | `finding.go:306` |
| `AppendAuditBody` / `EditAudit` write guards | `body.go:112`, `edit.go:313` |

The shared `AuditLintIssues` check-set from #193 is intact — `audit lint` and top-level `lint`
cannot drift, and `NearMisses` is threaded through the same single read rather than re-derived.
`render.auditProgressCell` consumes only `Audit.Findings`; it does **not** consume `NearMisses`,
which is the root of H2.

### 4. Evidence

**Independent corpus false-positive sweep (B1).** I ran the shipped `NearMissFindingHeaders`
over every markdown file in the planning corpus and classified every hit by hand, rather than
reusing the pinned fixture:

| Corpus | Files | Hits | Finding-shaped | **Corrupting false positives** |
|---|---|---|---|---|
| `planning/audits` | 64 | 0 | 0 | **0** |
| `planning/research` | 31 | 38 | 27 | **11** |
| `planning/tasks` | 324 | 20 | 14 | **6** |
| `planning/threads`, `epics`, `meta`, `docs` | 113 | 0 | 0 | **0** |

The audit result confirms the claim in the code comment. The rest is the systemic problem: the
narrowing was validated against the only document class that cannot exhibit the failure, because
audit headings are tool-scaffolded (`## Findings`, `## Candidate tasks`) while everything else is
hand-authored prose. Representative corrupting hits, verbatim from the corpus:

| Real heading | What `lint --fix` would write |
|---|---|
| `## Why 3 and 5 are one task (and 4 was not)` | `## WHY3. and 5 are one task (and 4 was not)` |
| `## The 14 without tags` | `## THE14. without tags` |
| `### Tier 1 — high impact, low effort` (×9 across 4 files) | `### TIER1. high impact, low effort` |
| `### base16 / tinted-theming — the recommended foundation` | `### BASE16. / tinted-theming — the recommended foundation` |
| `## OSC-11 spike result (2026-06-23)` | `## OSC11. spike result (2026-06-23)` |
| `### Path 0 — Remote filesystem mount (works today, zero code)` | `### PATH0. Remote filesystem mount (works today, zero code)` |
| `## V1 contract` | `## V1. contract` |

**Hostile boundary fixtures (B2).** Beyond the corpus, the recognizer claims
`### A1 Introduction` (named in the brief), `## Step 1 - foo`, `## PR 194 notes`,
`## Top 5 risks` and `## Day 2 plan`. It correctly ignores `### 1. Lifecycle`,
`#### 3.1 Data Flow`, `## 2) Scope`, `## Go 1.24 upgrade`, `## RFC 7231 semantics`,
`## ADR 0003 recap`, list items, indented headings, block-quoted headings and mid-line text.
`#### Q3. results` is not a near-miss because `findingHeaderRe` already parses it as a finding —
pre-existing behaviour, not introduced here.

**Fence awareness (B2).** Correct on all six shapes tried: backtick, tilde, info-string
(` ```go `), nested four-backtick wrapping an inner three-backtick, unterminated, and indented.
Only the genuinely-outside-the-fence header was claimed. No finding.

**Auto-repair safety (B3).** `lint --fix` over the real corpus: `nothing to fix`, tree unchanged.
`CanonicalizeFindingHeaders` is idempotent (second pass reports 0 hits, output byte-identical).
The repair routes through `TransformAuditBody`, so it carries the content CAS — confirmed by
mutation B-M17 (killed) and by reading `finding.go:303-317`. Repairs are announced per line in
both `--fix` and `--fix --dry-run` output.

**Genuine concurrency (A2), no `testHookBeforeBodyWrite`.** 8 goroutines × `EditFinding` on one
audit, and a second probe racing 4 stamps against 4 append loops, both `-race -count=5`:

```
TestReviewProbe_RealConcurrentFindingStamps   lost=0 of 8            (×5)
TestReviewProbe_StampsRacingAppends           stamped=4 of 4; appends present; findings parsed=8  (×5)
TestReviewProbe_ExhaustionIsErrConflict       maxRetries=0: ok≈127 conflict≈73, no other error class
```

**No-op and dry-run semantics (A3, A4).** With a pinned clock: a re-applied identical edit
returns `changed=false` and leaves `updated_at` byte-identical; `--dry-run` returns
`changed=true` and touches nothing (`writeBody` returns before the lock and the CAS, so a dry run
neither locks nor rechecks); `retryOnConflict` returns immediately on `dryRun` so it never
retries. Applying the same note twice leaves exactly one `**Resolution:**` label. The
`normalizeBody` equality swallows only trailing blank lines and EOL style — interior whitespace
and trailing spaces on a line still write — and neither shipped transform can produce a
whitespace-only change, so A3 is a non-issue.

**Scope (A5).** Part A's diff does not touch `edit.go`; `EditAudit` and the interactive loop were
left alone and no repair policy was smuggled in. Part B later added a guard to `EditAudit`
deliberately; see M3 and L1.

**Mutation probes.** Applied one at a time to the restored checkpoint, `go test ./internal/...`,
then restored. Killed:

| Mutation | Killed by |
|---|---|
| A-M1 `resolveAudit` → `resolve` (task resolver) | `TestAuditFinding_StatusAndNoteInOneWrite` +5 |
| A-M2 CAS `resolveAuditPath` → `resolvePath` | `TestEditFinding_RetriesAroundConcurrentAppendPreservingBoth` +5 |
| A-M3 drop the no-op equality check | `TestTransformAuditBody_UnchangedBodyIsNoOpAndDoesNotStamp` |
| A-M5 `EditFinding` without `retryOnConflict` | `TestRetry_EditFindingRetries` |
| A-M18 take the lock/CAS on a dry run | `TestAppendAuditBody_DryRun_NoWrite` +5 |
| B-M6 drop `strings.ToUpper` | `TestNearMissFindingHeaders_CatchesEverySilentLossShape/lowercase_code` |
| B-M7 drop the leading-zero trim | `TestNearMissFindingHeaders_CatchesEverySilentLossShape/multi-letter_code` |
| B-M8 regex accepts a bare-digit code | `TestNearMissFindingHeaders_LiveCorpusIsClean` +3 |
| B-M9 drop the fence skip | `TestNearMissFindingHeaders_IgnoresFencedExamples` |
| B-M10 drop the already-canonical skip | `TestAppendAuditBody_AcceptsCanonicalHeader` +5 |
| B-M12 drop the `audit append` guard | `TestAppendAuditBody_RefusesIntroducedNearMissHeader` |
| B-M16 emit no near-miss lint issues | `TestLintFindingHeaders_MessageNamesLineAndReplacement` |
| B-M17 `lint --fix` never repairs | `TestLintFix_CanonicalizesNearMissFindingHeader` |

Five survived; they are reported as M3.

### 5. Acceptance-criteria traceability

Part A — [give finding status writes a non-interactive CAS path](../tasks/6g392b0rps7w-give-finding-status-writes-a-non-interactive-cas-path.md):

| AC | Verdict | Evidence |
|---|---|---|
| 1 port exposes a CAS-protected body transform | **met** | `store.go:258`; `body.go:173`; A-M1/A-M2 killed |
| 2 `EditFinding` inside `retryOnConflict`, `attempted` gone | **met** | `finding.go:256`; flag absent from the diff; A-M5 killed |
| 3 append/finding race preserves both; exhaustion → `ErrConflict` | **met** | real-parallelism probes, `lost=0 of 8`, `stamped=4 of 4`, exhaustion yields only `ErrConflict` |
| 4 `--dry-run` no lock/write; already-applied edit is a no-op | **met** | pinned-clock probe; A-M18 killed |
| 5 focused tests pin the five behaviours | **met** | `transformauditbody_test.go`, `occ_retry_test.go`; one uncovered substitution (noun) reported as M3 |

Part B — [make near-miss audit findings loud and self-repairing](../tasks/6g72wf39pyhb-make-near-miss-audit-findings-loud-and-self-repairing.md):

| AC | Verdict | Evidence |
|---|---|---|
| 1 narrow letter-led recognizer, line + canonical replacement | **met** | `finding.go:64`, `NearMissFindingHeaders`; message names both |
| 2 zero false positives across the audit corpus | **met as written; the claim it is used to support is not** | 0/64 audits re-derived independently — but see H1: the corpus is the wrong population for the conclusion drawn |
| 3 all silent-loss shapes caught by lint | **met for ≤4-letter, ≤3-digit codes** | all nine shapes claimed; ≥5-letter and ≥4-digit codes still dropped — M2 |
| 4 `lint --fix` canonicalizes through the body-transform CAS; FixFrontmatter stays frontmatter-only | **met** | `finding.go:303`; B-M17 killed; corpus `--fix` is a true no-op |
| 5 `audit append` and `audit edit` route through the check, which "cannot be forgotten or bypassed by a raw write" | **not met** | `audit new --body` bypasses it entirely (H3); the `audit edit` half is untested (M3) |
| 6 explicit no-findings state; broken-vs-empty carried by the lint rule | **met literally, defeated in practice** | the two states render identically on every surface (H2) |
| 7 brief boilerplate and schema audit state the real rule | **met** | `entity.go:126`; verified against the parser's actual behaviour |

### 6. Demonstrated defects, source-supported risks, unverified concerns

**Demonstrated** (reproduced end-to-end in the sandbox, with the exact commands): H1, H2, H3,
M1, M3. For H1 I ran the corruption through the real binary — hand-author four prose headings
into an audit, run `lint --fix`, and read the rewritten file back.

**Source-supported risks** (correct by reading, not yet reachable in this corpus): M2 (the
recognizer's `[A-Za-z]{1,4}`/`\d{1,3}` is narrower than the parser's `[A-Z]+\d+`; this corpus
only uses 1–2 letter codes, so nothing is dropped today) and L1 (`EditAudit` discards the
original file's frontmatter-split error; reachable only for a file that has *both* unterminated
frontmatter *and* pre-existing drift).

**Unverified concerns.** L2's claim that Part B's shape *should* generalize to candidate-task
markers is a design judgement, not a measurement — I verified only that no code reads those
markers and that the four near-miss functions have no injection seam. I did not attempt to build
the generalization. I also did not review #195 or #197, which are out of scope.

## Findings

#### H1. `lint --fix` rewrites ordinary prose headings into fabricated findings, and the zero-false-positive measurement was taken on the one corpus that cannot show it · **Status:** tracked by 6g77rn6b9wf8

`nearMissHeaderRe` (`internal/domain/finding.go:81`) accepts any 1–4 letter word followed by a
number, so ordinary English section headings are claimed as drifted finding codes and
`CanonicalizeFindingHeaders` rewrites them in place under `lint --fix`. Reproduced through the
real binary on a throwaway planning space, using heading text taken verbatim from this
repository:

```
$ tskflwctl lint --fix
fixed …/audits/…-2026-09-05-prose-headings.md
  - line 31: "## Why 3 and 5 are one task (and 4 was not)" → "## WHY3. and 5 are one task (and 4 was not)"
  - line 35: "## The 14 without tags" → "## THE14. without tags"
  - line 39: "## Tier 1 — cheap wins" → "## TIER1. cheap wins"
  - line 43: "### base16 / tinted-theming — the recommended foundation" → "### BASE16. / tinted-theming — the recommended foundation"
…
  WHY3: missing **Status:** — expected one of: deferred, fixed, in-progress, open, …
  THE14: missing **Status:** — expected one of: …
```

Two of those four destroy the sentence outright, and one leaves a dangling `/`. The tool then
reports four findings that never existed, each "missing a status", and the audit cannot reach
`ready to close` until a human hand-repairs the file — the exact loop the Thread exists to
remove, now caused by the repair rather than by the drift.

**This is the systemic finding.** The narrowing is defended in the code comment as "Requiring a
LETTER-LED code is what does that … measured at zero [false positives] over the 57-audit corpus",
and I confirmed that number independently: **0 hits across all 64 audits**. But audits are the
one document class in this repo whose headings are tool-scaffolded. Running the same shipped
recognizer over the rest of the corpus:

| Corpus | Files | Hits | **Corrupting false positives** |
|---|---|---|---|
| `planning/audits` | 64 | 0 | **0** |
| `planning/research` | 31 | 38 | **11** |
| `planning/tasks` | 324 | 20 | **6** |

Seventeen headings in this repository would be corrupted by the rule that measured clean. The
measurement was drawn from the same population the design was tuned on, so it certifies the
sample rather than the rule. Nothing is corrupted *today* — the audit corpus is clean and
`lint --fix` over it is a verified no-op — but audits here are hand-authored (this brief is), and
`## Tier 1 — …` is a habit already present in the neighbouring documents.

Mitigating, and worth weighing before choosing a severity: `--fix` prints every rewrite, and
`--fix --dry-run` previews them. This is a loud destructive edit, not a silent one.

Directions, cheapest first: require the code letters to be **uppercase** (kills `Why`, `The`,
`Tier`, `Step`, `Top`, `Day`, `base16`, and every hit in the table, while `h1.` stays reachable
through a separate all-lowercase branch); or gate `--fix` on the heading sitting under a
`## Findings` section; or derive the acceptable code prefixes from the audit's own existing
findings. Whichever is chosen, re-run the measurement over **all** entity types, not just audits.

**Resolution:** Confirmed by reproduction: the shipped recognizer rewrites '##
Why 3 and 5 are one task' to '## Why3. and 5 are one task'. The
zero-false-positive claim was measured only on the tool-scaffolded audit corpus,
which cannot show it. Requiring uppercase code letters clears every prose case
in the finding while retaining 8 of 9 drift shapes.

#### H2. An audit whose findings all failed to parse is indistinguishable from an empty one on every read surface, and now says so in words · **Status:** tracked by 6g77rn6em6n8

`auditProgressCell` (`internal/cli/render/render.go:641`) branches on `a.Findings == 0` alone, so
the two states render identically. Two audits in one throwaway space — one genuinely empty, one
carrying eight real drifted findings in the canonical incident shape:

```
$ tskflwctl audit list
BUCKET  AUDIT                       PROGRESS     AREA
open    2026-09-05-genuinely-empty  no findings  genuinely-empty
open    2026-09-05-all-drifted      no findings  all-drifted

$ tskflwctl audit show 2026-09-05-all-drifted | grep '^findings:'
findings: no findings

$ tskflwctl audit findings 2026-09-05-all-drifted
$ echo $?
0
```

`--json` agrees: both report `findings: 0`. `audit show` says nothing about the drift anywhere in
its output. The only surface that tells the truth is a separate `lint`/`audit lint` invocation.

This is a regression in *framing* against the incident the Thread was opened for. Audit
`6g5axb85endz` rendered `0% settled 0/0` — a number that at least looked wrong. The new cell
replaces it with the English assertion `no findings`, which is simply false, on the surface
CLAUDE.md names as the cheapest triage path (`audit show <id>` / `audit list -o table`). AC6 is
satisfied to the letter — the distinction *is* carried by the lint rule — but the read surface
now states the opposite of the truth confidently, and a reader has no cue to go run lint.

No new hand-authored magic string is needed to fix it: `AuditWithFindings.NearMisses` is already
computed on the same single read for the list path (`auditstore.go:243`, `store.go:206`), and the
show path already holds the body. `no findings (8 unparsed headings — run lint)` is derivable
from data the renderer is already being handed.

**Resolution:** Accepted as a framing regression I introduced and argued for.
NearMisses is already on the same read, so the truthful cell needs no new data.

#### H3. `audit new --body` bypasses the near-miss write guard that AC5 says cannot be bypassed · **Status:** tracked by 6g77rn6hvmh8

`FS.CreateAudit` (`internal/store/create.go:217`) is the one audit-body producer with no
`NearMissWriteError` check, and `audit new` exposes `--body`, `--body-file` and `--template`
(`internal/cli/audit.go`), so arbitrary body text reaches it. The same header that `audit append`
refuses with exit 11 is accepted at creation:

```
$ tskflwctl audit append 2026-09-05-guarded --body '#### H-1. seeded drift  · **Status:** open'
error: validation failed: this audit write would add 1 heading(s) that parse to nothing, …
  exit=11

$ tskflwctl audit new seeded --date 2026-09-05 --body '## Findings

#### H-1. seeded drift  · **Status:** open
…'
created audits/6g77jasqv5rw-2026-09-05-seeded.md
  exit=0

$ tskflwctl audit show 2026-09-05-seeded | grep '^findings:'
findings: no findings
```

AC5's wording is that the check "cannot be forgotten or bypassed by a raw write". Creation is the
first write an audit gets, and it is exactly how an agent seeds a review brief with findings —
the workflow this Thread is built around. Combined with H2 the result is fully silent: the
findings are invisible and the surface says there are none.

The fix is the same two lines `AppendAuditBody` already carries, against an empty `prev`:
`domain.NearMissWriteError("audit", domain.IntroducedNearMissHeaders("", body))`.

**Resolution:** Correct: the guard was added to append and edit but not to
CreateAudit, so AC5's 'cannot be bypassed' overstated it.

#### M1. Two doc comments were captured by newly inserted declarations, and the generated wire artifact has already enshrined one of them · **Status:** tracked by 6g77rn6n6b86

Both new declarations in Part B were inserted **between** an existing doc comment and the
function it documented, so Go re-attached the comment to the new declaration:

- `internal/domain/finding.go:225-234` — the `LintFindings` doc comment ("LintFindings validates
  an audit's parsed findings plus the bucket↔state invariant…") now documents
  `type NearMissHeader`, and `func LintFindings` (line 365) is left with no doc comment at all.
- `internal/cli/render/render.go:626-641` — the `auditStateNote` doc comment now documents
  `func auditProgressCell`, and `auditStateNote` (line 650) is left undocumented.

The first has already been captured by the generator: `internal/wire/schema_comments.json:56`
records `domain.NearMissHeader` as *"LintFindings validates an audit's parsed findings plus the
bucket↔state invariant, returning one Issue per problem (Field = the finding code, or "bucket"
for the audit-level check)."* Because it is a checked-in generated artifact, `just docs-check`
passes with the wrong text baked in, and the misattribution will survive future regeneration.

Scope check: I confirmed `NearMissHeader` is **not** currently reachable from
`schema --json-schema` output (0 occurrences), so nothing user-facing ships the wrong description
today. It would the moment near-misses appear in a `--json` envelope, which the `NearMisses`
field on `AuditWithFindings` makes plausible.

Two independent instances in one PR is what makes this worth a finding rather than a nit.

**Resolution:** Verified: the LintFindings comment now documents NearMissHeader,
and it reached schema_comments.json. This is why that artifact went stale twice
during the build.

#### M2. The recognizer's code token is narrower than the parser's, so drift on a longer code is still silently dropped · **Status:** tracked by 6g77rn6b9wf8

`findingHeaderRe` accepts `[A-Z]+\d+` — unbounded letters and digits. `nearMissHeaderRe` accepts
`[A-Za-z]{1,4}` and `\d{1,3}`. Anything canonical in the gap has no safety net:

| Line | Parses as a finding? | Recognized as a near-miss? |
|---|---|---|
| `#### AUDIT1. Title` | yes | — (canonical) |
| `#### AUDIT-1. Title` | **no** | **no — silently dropped** |
| `#### CRITICAL-1. Title` | **no** | **no — silently dropped** |
| `#### H1234. Title` | yes | — (canonical) |
| `#### H-1234. Title` | **no** | **no — silently dropped** |
| `#### H-1.Title` (no space) | **no** | **no — silently dropped** |
| `#### H-1.` (empty title) | **no** | **no — silently dropped** |

Part B also added the sentence *"the code must match `[A-Z]+[0-9]+` followed by `.`"* to
`schema audit` (`internal/domain/entity.go:126`). An author who follows that documented grammar
and picks a five-letter code gets no protection from the feature that documents it.

Latent in this corpus: every code in use is 1–2 letters (`L` 188, `M` 166, `H` 79, `WR` 6, `S` 1,
`N` 1), so nothing is being dropped today. The asymmetry is nonetheless a hole in the stated
contract, and it is cheap to close by mirroring the parser's bounds.

**Resolution:** Same surface as H1: the recognizer's bounds must be re-derived
against findingHeaderRe's unbounded code, not chosen independently.

#### M3. Five behaviours in the merged contract are unpinned — the mutations that remove them pass the whole suite · **Status:** tracked by 6g77rn6n6b86

Mutation probes applied to the restored checkpoint, `go test ./internal/...`, then restored:

| Mutation | Survived because |
|---|---|
| **B-M13** delete the entire `audit edit` near-miss guard (`edit.go:305-313`) | no test exercises it. `nearmiss_write_test.go` covers `AppendAuditBody` and `TransformAuditBody` only; no test in the repo pairs `EditAudit` with a near-miss. This is **half of AC5**. |
| **B-M14** `a.Findings == 0` → `a.Findings < 0` in `auditProgressCell` | the zero-findings cell — **the headline user-visible change of AC6** — has no test; reverting it to the old `0% settled 0/0` is invisible to the suite. |
| **B-M11** key `IntroducedNearMissHeaders` by `Line` instead of `Text` | the documented insert-above case is untested. I verified the shipped behaviour *is* correct with a direct probe (insert-above → 0 introduced; new drifted header → 1; duplicated header → 1), so this is coverage only, not a defect. |
| **B-M15** drop `if len(a.NearMisses) == 0 { continue }` in `FixFindingHeaders` | the "audits with nothing to repair are not written at all" claim in the doc comment is unpinned; the mutation is behaviour-preserving but does N transforms instead of 0. |
| **A-M4** noun `"audit"` → `"task"` in `TransformAuditBody`'s `writeBody` call | the noun is the one of Part A's four substitutions with no coverage. It reaches users as `"task body ends inside an unterminated ``` fence…"` on an **audit**. The resolver, CAS resolver and parse function are all pinned (A-M1/A-M2/A-M3 killed). |

B-M13 and B-M14 are the ones that matter: each is the only test-visible evidence that an
acceptance criterion was actually implemented.

**Resolution:** Overlaps the thin-coverage list in this brief; the mutation
ledger makes it concrete.

#### L1. `EditAudit` discards the original file's frontmatter-split error, so a repair edit can be refused for drift it did not introduce · **Status:** tracked by 6g77rn6hvmh8

`internal/store/edit.go:296` is `_, origBody, _ := splitFrontmatterStrict(orig)`. When that call
errors — an audit whose frontmatter opens with `---` and is never closed — `origBody` is `nil`,
so `IntroducedNearMissHeaders("", newBody)` reports **every** near-miss in the file as introduced
by this edit.

The reachable case is an audit with *both* unterminated frontmatter *and* a pre-existing drifted
header. An author who opens `audit edit`, fixes the frontmatter and leaves the header alone is
told their write "would add 1 heading(s) that parse to nothing" — which they did not do. They can
escape by also fixing the header, so the consequence is a confusing message rather than a lost
edit, but the guard's own comment ("Pre-existing drift elsewhere in the file is not this edit's
fault") states the intent it fails to keep here.

Source-supported: I read the path rather than driving `$EDITOR`. `splitFrontmatterStrict` returns
`nil, nil, err` on that branch, so the empty baseline is certain.

**Resolution:** Same write-guard surface: the discarded split error makes
IntroducedNearMissHeaders treat a whole file as introduced.

#### L2. Part B's pipeline is hard-coded to findings, so the third instance of this bug class cannot reuse it · **Status:** tracked by 6g3ag8py12y9

The brief asks whether Part B's design generalizes to the candidate-task markers
(`✅ ⏳ ⚠️ ⛔`). It does as a *shape* — detect narrowly, lint loudly, canonicalize under `--fix`,
refuse at the write, diffed against the previous body — but there is no seam to reuse. All four
functions name findings directly: `NearMissFindingHeaders` hard-codes `findingHeaderRe` and
`nearMissHeaderRe`; `IntroducedNearMissHeaders` calls it with no injection point;
`NearMissWriteError` takes a `noun` but a finding-typed `[]NearMissHeader`;
`CanonicalizeFindingHeaders` is finding-only. Applying the pattern to markers means copying all
four.

Confirming the premise: the markers appear only in the scaffold template
(`internal/domain/entity.go:351,414`) and in `stripLeadingDecoration`, which *strips* them when
parsing a resolution. No parser reads a candidate item's marker, no lint rule checks it, and
`audit close` does not consult it — so the convention is still purely advisory, exactly as the
brief describes.

This is a design observation, not a demonstrated defect; I did not attempt the generalization.
It is also the second clone in this contract — `TransformAuditBody` and `TransformTaskBody` are
near-identical, differing only in noun, resolver, parse function and CAS recheck (and A-M4 shows
the noun is the substitution that slipped through). Worth deciding deliberately whether the next
sub-entity gets a generic helper or a third copy.

**Resolution:** The seam belongs with the candidate-list convention task, which
is the third instance this generalization exists to serve.

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

- ⏳ `tskflwctl task new "Narrow the near-miss recognizer so it cannot claim prose headings" --epic <id> --tags audit,lint` — require uppercase code letters (or a findings-section gate), then re-measure false positives across **all** entity types, not audits alone. Fixes H1.
- ⏳ `tskflwctl task new "Distinguish an empty audit from one whose findings failed to parse" --epic <id> --tags audit,render` — feed the already-computed `NearMisses` count into `auditProgressCell` and the `--json` envelope. Fixes H2.
- ⏳ `tskflwctl task new "Guard audit creation against seeding unparseable finding headers" --epic <id> --tags audit,store` — apply `NearMissWriteError` in `CreateAudit` against an empty baseline. Fixes H3.
- ⏳ `tskflwctl task new "Repair the two captured doc comments and regenerate schema_comments.json" --epic <id> --tags docs` — restore `LintFindings` and `auditStateNote`'s comments. Fixes M1.
- ⏳ `tskflwctl task new "Match the near-miss code bounds to findingHeaderRe" --epic <id> --tags audit,lint` — widen letters/digits so drift on a long code is not silently dropped. Fixes M2.
- ⏳ `tskflwctl task new "Pin the untested branches of the sub-entity write contract" --epic <id> --tags testing` — cover the `audit edit` guard, the no-findings cell, the text-keyed introduced-header diff, the clean-audit skip, and the audit noun. Fixes M3.
- ⏳ `tskflwctl task new "Stop EditAudit discarding the original frontmatter-split error" --epic <id> --tags audit,store` — fall back to refusing nothing when the baseline is unreadable. Fixes L1.
- ⏳ `tskflwctl task new "Decide whether the near-miss pipeline gets a reusable seam" --epic <id> --tags design` — before candidate-task markers become the third copy. Addresses L2.
