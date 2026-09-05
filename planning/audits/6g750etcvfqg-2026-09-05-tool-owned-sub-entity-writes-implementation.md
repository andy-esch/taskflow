---
schema: 1
id: 6g750etcvfqg
bucket: open
area: tool-owned-sub-entity-writes-implementation
date: "2026-09-05"
---
# Audit: tool-owned sub-entity writes — 2026-09-05

> Reviewer assignment: UNASSIGNED. Copy this file per reviewer, appending `-<reviewer>` to the
> area and filename, before starting. This document is the review brief and the only file the
> reviewer should update.
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
AUDIT_REL="planning/audits/<your copy of this file>"   # per-reviewer copy, see the header
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

Review them as one contract even though they land separately. A finding that spans both is more
valuable than two local ones.

- **Part A — [#194](https://github.com/andy-esch/taskflow/pull/194), landed or in review.**
  `TransformAuditBody` plus routing `core.EditFinding` through `retryOnConflict`, replacing the
  interactive `$EDITOR` callback loop and its `attempted` escape flag.
- **Part B — upcoming.** Task
  [make near-miss audit findings loud and self-repairing](../tasks/6g72wf39pyhb-make-near-miss-audit-findings-loud-and-self-repairing.md):
  a narrow letter-led near-miss recognizer beside `findingHeaderRe`, a loud `audit lint` rule, and
  `lint --fix` canonicalization routed through Part A's transform.

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

## Part B — decisions taken during implementation

Recorded as they were made, so the reviewer inherits the reasoning rather than reverse-engineering
it from the diff. Each is a deliberate choice with a plausible alternative; all are fair game.

1. **A separator dash is discarded, not preserved.** `#### BTA-01 — title` canonicalizes to
   `#### BTA1. title`, dropping the em-dash. The first implementation kept it and produced
   `#### BTA1. — title` with a dangling dash. **This is the only place the repair discards author
   bytes.** The judgement is that a dash between code and title is a separator standing in for the
   period, not title text. Attack it: a title that legitimately opens with a dash, an en-dash used
   as a bullet, `#### H1 - - double`. Mitigation in place: the lint message shows the exact
   replacement before `--fix` is ever run, so nothing changes silently.
2. **The recognizer captures the title rather than slicing the line.** The first version computed
   the title by byte offset and split a multi-byte em-dash mid-rune (`#### M2. \x94 a title`),
   caught by a test rather than review. Any future widening must keep the capture-group form —
   verify no byte arithmetic returns.
3. **`AuditWithFindings` gained a `NearMisses` field** rather than carrying the raw body. Near
   misses are derived on the same scan that parses findings, because a dropped finding is by
   construction absent from the parsed set and no later pass over parsed findings can see it. The
   alternative — carrying every body — was rejected as wasteful for `summarize`, which shares this
   type. Judge whether a derived diagnostic belongs on a read-model struct, and whether the
   single-audit path (`LintAudits`) and the sweep path can drift now that they derive it separately.
4. **`AuditLintIssues` takes near-misses as a third parameter.** It could have recomputed them from
   a body, but the sweep does not carry one. Confirm both call sites pass equivalent input and that
   nothing bypasses the shared check-set.
5. **`FixFindingHeaders` returns the durable prefix on failure.** A mid-run write error surfaces
   after the audits already repaired, matching the frontmatter fixer's partial-progress contract.
   Verify the prefix is accurate and that a conflict mid-sweep is retried rather than lost.
6. **Body repairs run after the frontmatter pass** in `runLintFix`, because the frontmatter pass can
   rename a file to heal a broken id and the body pass resolves by slug. Check the ordering holds
   when a single run both renames a file and repairs a header inside it.
7. **Heading depth is preserved, not normalized.** A `##### h-3:` becomes `##### H3.`, not `####`.
   The parser accepts depths 2–6, so normalizing would be a gratuitous rewrite — but it does mean a
   canonicalized corpus can carry mixed finding depths.
8. **A write is refused, not auto-repaired.** `audit append` and `audit edit` reject a heading that
   would parse to nothing, naming the canonical replacement, rather than silently rewriting what
   the author typed. This follows the unterminated-fence guard sitting beside it in `writeBody`;
   `lint --fix` remains the explicit opt-in for rewriting text. Challenge whether refusing is right
   for an agent workflow, where auto-repair would save a round trip.
9. **Only INTRODUCED drift is refused.** `IntroducedNearMissHeaders(prev, next)` diffs by header
   text, so pre-existing drift never blocks an unrelated append or a status stamp — otherwise one
   bad header would freeze the document. Attack the diff: duplicate identical headers (the count is
   decremented per occurrence), a header moved rather than added, and a rename that both removes
   and adds. The repair path (`TransformAuditBody`) is deliberately NOT covered by this guard.
10. **An acceptance criterion was narrowed rather than met.** AC 6 originally also required lint to
    flag an open audit with no findings and no near-misses. Measured against the live corpus that
    rule fires on exactly one audit — this brief, legitimately empty while it waits for a reviewer —
    and would fire on every fresh `audit new` scaffold. It was dropped as a false-positive generator
    of the same kind this work exists to remove; the broken-vs-empty distinction is carried by the
    near-miss rule, which names the offending line. **Reviewer: this is a scope reduction I made
    unilaterally — overturn it if you disagree.**

## Part B — where the coverage is thin

Stated plainly rather than left for the reviewer to discover, because the author knows best where
he did not look.

- **`audit edit`'s guard has no test.** The append path is covered in both directions; the editor
  path (`internal/store/edit.go`, inside `acceptEdited`) is wired but only exercised indirectly. It
  reads the pre-edit body via a second `splitFrontmatterStrict` on `orig` — confirm that matches
  what the editor callback actually receives, and that a broken-frontmatter edit still surfaces the
  parse error rather than the near-miss refusal.
- **`FixFindingHeaders` partial-progress is unverified under failure.** The happy path, dry run, and
  idempotence are tested; a mid-sweep write failure across several drifted audits is not. The
  durable prefix it returns on error is asserted by construction only.
- **Retry inside the fix sweep is untested.** `FixFindingHeaders` wraps each audit in
  `retryOnConflict`, but no test races a concurrent writer against a `--fix` run.
- **No fuzzing of the recognizer.** `internal/domain/body_fuzz_test.go` exists as a precedent for
  the criterion scanner; the near-miss recognizer has none. A fuzz target over heading-shaped input
  asserting "never claims a line `ParseFindings` would accept" would be cheap and valuable.
- **Multi-byte coverage is one case.** The em-dash regression is pinned, but CJK titles, RTL text,
  and combining marks in a heading are untested.

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
- **The recognizer's corpus measurement is now a test, not a claim.**
  `TestNearMissFindingHeaders_LiveCorpusIsClean` walks `planning/audits/` on every run and fails
  with the exact heading it would have corrupted; `TestNearMissFindingHeaders_DoesNotShadowCanonicalFindings`
  asserts `--fix` is a no-op over the same tree. Re-derive independently if you like, but a
  regression here cannot land silently.

## Autonomy disclosure

This slice was implemented with unusually little supervision, so the things a reviewer would
normally infer from a conversation are recorded instead:

- Every acceptance criterion was ticked by the author. **AC 6 was narrowed before being ticked**
  (decision 10 above) — that is the one place the delivered scope is smaller than the agreed scope.
- Three shipped behaviours change existing output or refuse previously-accepted input, and each
  deserves an explicit yes/no: the `no findings` progress cell (porcelain change), the
  `audit append`/`audit edit` refusal (previously-accepted writes now fail), and the separator-dash
  discard under `--fix` (the only place author bytes are dropped).
- The corpus measurements quoted throughout (58 audits, 426 findings, 51 orphan markers, 116
  false positives for the rejected pattern) were taken by the author against the live tree. The
  first two are pinned as tests; the last two are not, and are the ones to re-derive if any of the
  design rests on them for you.

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

## Findings

<!-- Reviewer: add findings here, one per issue, in the exact grammar from the header. -->

_No findings recorded yet — this brief is queued for a reviewer._

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

- _none yet_
