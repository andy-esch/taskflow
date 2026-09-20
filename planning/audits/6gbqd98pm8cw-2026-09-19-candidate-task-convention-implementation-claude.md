---
schema: 1
id: 6gbqd98pm8cw
bucket: closed
area: candidate-task-convention-implementation-claude
date: "2026-09-19"
updated_at: "2026-09-19"
---
# Audit: Candidate-task convention implementation — claude — 2026-09-19

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Adversarially review the implementation of the versioned, tool-owned audit Candidate tasks
projection. Treat green tests as claims to challenge, not proof. Perform one contract/correctness
pass and then a separate systemic pass looking for failure classes hidden by shared helpers,
permissive fixtures, or the current corpus. Report only evidence-backed findings; a substantive
no-findings report is acceptable when every required challenge is settled.

## Review target

Review the complete working-tree change for task `6g3ag8py12y9`,
`decide-the-candidate-list-convention-and-make-the-tool-own-it`, against `main`. Start with a
repository-wide consumer inventory and verify every claimed consumer in the sandbox:

- `internal/domain/candidate.go`, `finding.go`, `entity.go`, their tests, and fence/section/span
  helpers;
- `internal/core/finding.go`, `service.go`, and `store.go`;
- `internal/store/auditstore.go` plus the existing atomic audit-body transform/CAS path;
- `internal/cli/audit.go`, `lint.go`, CLI integration tests, generated help, and the template
  golden/update guard;
- `internal/theme/theme.go` and all other finding-status glyph consumers;
- README, CLAUDE, architecture guidance, and scheduled audit routines;
- the planning decision and acceptance criteria in task `6g3ag8py12y9`.

Also inspect the newly scoped follow-up `6gbpe6e8n87k` only to distinguish intentionally deferred
finding creation from defects in the candidate projection. Do not credit follow-up requirements as
implemented behavior.

## Intended contract to challenge

- A managed section is explicitly marked `candidate-tasks:v1`. Unversioned Candidate tasks prose
  remains readable, is not linted as v1, and is never guessed at or rewritten.
- A canonical row is exactly one finding projection:
  `- <glyph> <CODE> · <status> — <one-line text>`. Rows are optional, unlinked rows and duplicate
  rows are invalid, and multiple findings may independently carry text pointing to the same task.
- The ordered status/glyph mapping is declared once in the domain. Template legends and themed
  status rendering consume that source rather than transcribing glyphs.
- `audit finding --candidate <text>` adds or replaces one row; an explicitly empty value removes
  it. A status change synchronizes any existing managed row. Status, note, and candidate edits in
  one invocation either persist together through the existing CAS transform or do not write.
- `audit lint` and top-level `lint` report malformed v1 rows, marker drift, unknown finding refs,
  status/glyph drift, and duplicate rows. Fenced examples do not activate the grammar.
- `audit new` default and security scaffolds emit a clean, derived v1 marker. The supported fresh
  workflow (`audit new`, append a finding, set its candidate, change status, lint) works even though
  a dedicated finding-creation verb is intentionally deferred.
- Markdown template body changes remain outside the JSON revision under ADR-0008. The golden
  updater exception must apply only to template-show body goldens and must not weaken revision
  enforcement for other JSON payloads.

Non-goals: migrating historical Candidate sections; inferring legacy linkage; storing derived rows
outside Markdown; permitting many finding codes in one row; building the follow-up finding-creation
verb; changing the seven-status vocabulary; or exposing candidate rows as a new JSON DTO.

## Mandatory evidence floor

Before reaching a verdict:

1. Run `git diff --check`, `go test -race ./...`, and golangci-lint with writable caches under
   `/tmp`. Run the planning and audit linters with the sandbox-built binary.
2. Reproduce in a disposable planning space: fresh default audit; appended canonical finding;
   candidate add, replace, status sync, and removal; scoped and top-level lint. Capture the actual
   Markdown after each meaningful transition.
3. Exercise legacy/no-section refusal and prove a combined status+candidate failure leaves the
   audit byte-identical. Exercise dry-run as well as a real write.
4. Challenge boundaries with fenced fake sections, CRLF input, Unicode text, a following H1/H2/H4
   section, a candidate section at EOF, multiple managed sections, duplicate rows, a missing code,
   wrong glyph, stale status, a drifted marker legend, explanatory comments, malformed list rows,
   and status decorations such as `tracked by <id>`.
5. Verify candidate diagnostics survive both store read paths: scoped `audit lint` and the
   single-read top-level lint aggregation. Confirm malformed/unreadable audits remain file problems
   rather than being silently reclassified.
6. Run repeated or coordinated concurrency tests against the actual audit-body transform. Establish
   whether two candidate/status edits, or an append racing a candidate edit, preserve CAS/retry
   semantics without duplicate rows or lost prose. Do not infer atomicity merely from one helper's
   name.
7. Inspect generated CLI docs and both audit routine specs for commands that work as written. Check
   that no guidance tells an agent to hand-type a managed row or falsely claims top-level lint
   excludes audit findings.
8. Verify ADR-0008's body-template exclusion directly, then prove the golden exception cannot bless
   an ordinary changed JSON golden at an unchanged revision.

For every claimed defect, provide an exact path/line and a minimal reproduction or mutation result.
Separate a shipped behavior from a planning-only intention.

## Required hostile angles

- **Mutation probes:** disable managed-row status synchronization and require its focused test to
  fail; make duplicate or missing-code lint permissive and require the corresponding focused test
  to fail; split the combined edit into an intermediate write or force candidate validation to
  fail after a status change and require the atomicity test to catch it; alter one domain glyph and
  show the legend/render drift protections that fail. Restore every mutation.
- **Span and boundary safety:** look for byte-offset drift caused by fenced content, multibyte
  glyphs, CRLF, section-end detection, insertion/removal at EOF, or replacements performed after an
  earlier edit changed offsets. Prefer property/fuzz evidence if a compact target is feasible.
- **Legacy downgrade attacks:** determine whether a small marker typo can silently turn previously
  managed corrupt data into ignored legacy prose, and distinguish forward-compatible unknown
  versions from accidental v1 corruption. Report only if the implementation can identify the
  intent without heuristic false positives.
- **Single-source claims:** inventory every persisted or rendered finding glyph and every status
  legend. Look for a second mapping that can drift, including tests, templates, docs, and TUI/CLI
  presentation.
- **Atomicity and ownership:** verify the core owns the multi-field operation, the filesystem
  adapter owns durable CAS, and Cobra is only an adapter. Look for a success receipt based on a
  stale/reloaded body or a failure path that partially changes `updated_at`.
- **Golden-policy containment:** try to use the template exception with a misleading golden name or
  another payload. Decide whether prefix matching is adequately bounded by the committed golden
  inventory or needs a stricter allowlist.
- **Operational ergonomics:** run the exact routine commands, including shell quoting of a proposed
  `task new` command inside `--candidate`. Flag unusable guidance, not aesthetic preference.

On the second pass, explicitly ask whether candidate projection should have reused an existing body
grammar abstraction, whether `AuditWithFindings.CandidateIssues` can become stale or diverge across
fake/real stores, whether optional rows undermine any stated mirror invariant, and whether a future
status addition can compile while leaving a consumer wrong.

## Validation and restoration

Work only in the mandatory independent sandbox injected above. Before each mutation, record the
focused test expected to fail; after the mutation, run that test and record the failure; restore the
file from the sandbox baseline and rerun it green. Do not weaken production code or tests to make a
probe convenient. Use writable caches such as:

```sh
GOCACHE=/tmp/taskflow-review-go-cache go test -race ./...
GOCACHE=/tmp/taskflow-review-go-cache \
  GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache golangci-lint run ./...
```

Build `./bin/tskflwctl` inside the sandbox before planning-space validation. End with only the
assigned audit modified; the isolation helper's verify and transfer steps are mandatory.

## Deliverable

Preserve this brief and replace the Reviewer report placeholder with one of:

- findings in exact repository grammar, all left `open`, each with severity, effort, urgency,
  component, exact evidence, reproduction/mutation, affected contract, and minimum recommendation;
  or
- a substantive no-findings report listing the consumer inventory, hostile cases, mutation probes,
  concurrency results, commands, and residual risks.

Do not implement fixes, edit the task, change finding statuses, create other planning files, commit,
or push. Include the required isolation attestation and transfer result.

## Reviewer report

Adversarial review of the versioned, tool-owned Candidate tasks projection for task
`6g3ag8py12y9`, against `origin/main`, performed entirely in the mandatory independent
sandbox. Five evidence-backed findings, all left `open`. Every named symbol, path, test,
and command below was executed or read in the sandbox; planning-only intentions are
labelled as such and the deferred follow-up `6gbpe6e8n87k` is credited nowhere as shipped
behaviour.

### Isolation attestation

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.NaIw88
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.NaIw88/.git
baseline_commit=579acb96841e5a7b5101db597298f714886af401
source_blob=1181fd2667f3aa38bfd6ad1d9a9aa32b24b7fbcb
source_fingerprint=2e581a62c337d8dfc42f194db48705dc166d7a04
deliverable=planning/audits/6gbqd98pm8cw-2026-09-19-candidate-task-convention-implementation-claude.md
deliverable_changed=true
transfer=<recorded in the handoff message>
```

`create` reported an independent `--no-hardlinks` clone: `.git` is a real in-tree directory,
`objects/info/alternates` is absent, and `git worktree list --porcelain` names exactly this
one checkout. HEAD never moved off `579acb96` — no reviewer commit, stage, branch switch,
restore, clean, stash, or reset. Every build, test, lint, generator, mutation, disposable
planning space, and edit happened under `$SANDBOX`. The helper permits one transfer per
sandbox (it re-checks `source_blob` immediately before the atomic rename), so the transfer
line is necessarily recorded in the handoff message rather than inside the file being
transferred; the workspace is retained at the path above pending owner confirmation.

One side effect outside both repositories: `tskflwctl init` in the disposable space
registered a space in the user's global registry (`space list` shows several pre-existing
stale `tmp.*` entries from earlier reviews). The disposable tree was deleted; the registry
entry `tmp-review-space` can be dropped with `space forget tmp-review-space`.

### Baseline (sandbox, writable caches under /tmp)

| Gate | Result |
| --- | --- |
| `git diff --check` | clean |
| `GOCACHE=/tmp/taskflow-review-go-cache go test -race ./...` | all packages ok |
| `golangci-lint run ./...` (both caches under /tmp) | `0 issues.` |
| `./bin/tskflwctl lint` (sandbox-built binary) | `✔ all planning entities and dependency links pass lint` |
| `./bin/tskflwctl audit lint` | `✔ all audit findings pass lint` |
| `go run ./internal/tools/docgen -out docs/cli` | no diff — generated CLI docs match the committed ones |

Re-run after every mutation was restored: identical, `git status --porcelain` empty.

---

#### M1. The ADR-0008 golden exception exempts a whole file by name prefix rather than the Markdown body  · **Status:** fixed

**File:** internal/cli/golden_test.go:84,96-98 | **Component:** cli/golden-policy
**Severity:** medium · **Effort:** XS · **Urgency:** soon

`goldenContentOutsideRevision` is `strings.HasPrefix(name, "template_show_")` (line 97) and
is consulted at line 84 as a blanket bypass of the revision gate for the golden's *entire*
payload. But `template_show_security_json` is not body-only: its JSON payload is
`{schema_version, template:{kind,name,description}, body}`, and `template.kind` /
`template.name` / `template.description` are ordinary `--json` machine-contract fields
surfaced by `template show` and `template list` — not Markdown scaffold content.

**Reproduction (Mutation F, through the real `-update` path).** Changed only the security
template's description string at `internal/domain/entity.go:139` — a contract field, not a
body template — then ran `go test ./internal/cli -run TestGolden -count=1 -update` at
`wire.SchemaVersion = "1.68"` (internal/wire/wire.go:286). Result:
`template_show_security_json.golden` was **rewritten** carrying
`"description":"CHANGED contract description — not a Markdown body."` at unchanged
`schema_version 1.68`, while the sibling `template_list_json` — which carries the *same*
field — was correctly refused: `refusing to rewrite machine-contract golden
template_list_json at unchanged revision 1.68; advance SchemaVersion and append its
classified changelog entry first`. The identical contract change is therefore gated in one
golden and blessed in another, leaving the corpus in a mixed state after one `-update`.

A direct probe of `validateMachineGoldenUpdate` at baseline `1.68` accepted all three of:
a changed `template.description`; a **deleted** `template.kind`; and the name
`template_show_but_actually_task_list_json` (prefix match only). The control
`task_list_json` was refused, so ordinary goldens do remain gated.

**Affected contract:** the brief's "The golden updater exception must apply only to
template-show body goldens and must not weaken revision enforcement for other JSON
payloads", and the exception's own comment, "Keep this exception narrow; ordinary JSON
payload changes remain gated." Prefix matching is **not** adequately bounded by the
committed inventory: the one golden it matches today already carries contract fields
outside the body.

**Recommendation:** scope the exemption to the payload it names — compare the candidate and
committed goldens with the `body` key excluded and bypass the gate only when nothing else
differs — and replace the prefix test with an explicit allowlist of committed golden names.

**Resolution:** The updater now explicitly allowlists the security template
snapshot and permits an unchanged-revision rewrite only when decoded JSON is
identical after removing body; metadata changes, removed fields, misleading
names, and missing baselines remain gated.

#### M2. Candidate add/remove cycles grow the managed section's whitespace without bound, and lint never sees it  · **Status:** fixed

**File:** internal/domain/candidate.go:273-289 | **Component:** domain/candidate
**Severity:** medium · **Effort:** XS · **Urgency:** eventually

`insertCandidateLine` rewinds `insertAt` back over the trailing whitespace run (line 285)
but leaves that run in the tail it re-appends, then unconditionally prepends `"\n\n"`
(line 288). `removeCandidateLine` takes the row plus exactly one newline (lines 275-279).
Net **+1 newline per set/remove cycle**, monotonically, whenever anything follows the
section. `LintCandidateTasks` classifies blank lines as allowed (candidate.go:101), so the
guard that owns this section never reports the drift.

**Reproduction (CLI, disposable planning space, on the documented layout).** `audit new` →
`audit append` a canonical finding — which is the supported fresh workflow, and which
leaves the managed section *followed* by that finding. Four
`audit finding … --candidate "cycle N"` / `--candidate ''` cycles took the blank-line count
inside the section from 4 to 5 to 6 to 7, and `od -c` showed six consecutive `\n` between
the marker comment and `#### H1.`. `audit lint` reported
`✔ all audit findings pass lint` (exit 0) at every step. Unit probe on
`"…marker\n\n## Related\n\n- ctx\n"`: 5 cycles grew 9 → 14 newlines, +1 each.

`store.normalizeBody` (internal/store/body.go:283) trims only *trailing* newlines, so it
masks this when the candidate section is last — but not in the layout the scaffold and the
routines actually produce.

The committed test encodes the growth instead of catching it:
internal/domain/candidate_test.go:69-71 asserts
`strings.HasSuffix(got, "\n\n\n## Related\n\n- context\n")` after a *single* insert, and
the test never re-adds a row after removing one, so the accumulation is invisible to it.

**Affected contract:** the section is tool-owned and every other write in this file is
byte-surgical; unbounded whitespace churn in a managed region produces diff noise on every
candidate edit and contradicts that ownership.

**Recommendation:** have `insertCandidateLine` normalize the gap it owns — drop the
whitespace run it skipped instead of re-appending it — so set → remove → set is byte-stable,
and replace the whitespace `HasSuffix` assertion with a round-trip byte-equality check.

**Resolution:** Candidate insertion and single-row removal now normalize the
whitespace owned by the managed section; repeated add/remove cycles are covered
by a byte-stability regression.

#### L1. A fenced example inside a managed section activates the row grammar and blocks the tool's own write path  · **Status:** fixed

**File:** internal/domain/candidate.go:81 | **Component:** domain/candidate
**Severity:** low · **Effort:** XS · **Urgency:** eventually

`candidateSections` derives heading and section-end offsets from `blankFences(body)`
(line 65) — fence-aware — and then calls `parseCandidateSection(body, &section)` (line 81)
on the **raw** body, so lines 103-115 classify every line of a fenced block inside a managed
section as candidate grammar.

**Reproduction.** A managed section containing a fenced shell example plus one valid row
yields three lint issues, one per fence line: ``line 7 is not a canonical candidate row …
"```sh"``, `line 8 … "tskflwctl audit finding my-audit H1 --candidate \"do the thing\""`,
``line 9 … "```"``. `SetFindingCandidate` then refuses outright:
`validation failed: the managed Candidate tasks section is malformed; run `audit lint` and
repair it before writing a candidate`. A fenced `## Findings` inside the section likewise
fails to end it — the section end correctly stays at the later real heading, because that
scan *is* fence-aware — while the fence's four lines are reported malformed.

**Affected contract:** "Fenced examples do not activate the grammar." That holds for a
fence *outside* the section (confirmed: `TestLintCandidateTasksIgnoresFencedManagedExample`
passes, and a fenced `## Candidate tasks` heading is correctly ignored) but not for one
inside it.

This fails **closed** — loud lint, refused write, no corruption — so severity is low. The
cost is that the section becomes permanently unwritable by the tool until the fence is
deleted by hand, and the refusal names neither the cause nor the remedy.

**Recommendation:** pass the fence-masked `prose` to `parseCandidateSection` for line
classification while keeping spans against `body`. This is safe by construction:
`blankFences` overwrites fenced bytes with spaces in place and never touches newlines, so
offsets are byte-identical — verified (`len(blankFences(body)) == len(body)` and section
offsets still land on `## Candidate tasks` across multibyte fenced content).

**Resolution:** Candidate section classification now reads fence-masked prose
while retaining raw byte spans, so fenced examples neither lint as rows nor
block tool-owned writes.

#### L2. Managed-marker detection is an unanchored prefix test, so a one-character typo silently downgrades corrupt managed data to ignored legacy prose  · **Status:** fixed

**File:** internal/domain/candidate.go:97 | **Component:** domain/candidate
**Severity:** low · **Effort:** S · **Urgency:** eventually

Line 97 tests `strings.HasPrefix(trimmed, "<!-- "+CandidateTasksVersion)` with no delimiter
after the version, which fails in both directions.

**Downgrade.** Starting from a managed section holding `- ○ H1 · open — stale row` against a
`fixed` finding — which correctly reports
`candidate status open disagrees with finding status fixed` — each of `candidate-tasks:v2`,
`candidate-task:v1`, and `<!--candidate-tasks:v1` (missing space) makes the section
unmanaged and `LintCandidateTasks` return **0 issues**. The stale mirror becomes invisible.

**Mis-classification.** Conversely `candidate-tasks:v10`, `candidate-tasks:v19`, and
`candidate-tasks:v1-draft` are all accepted **as v1** and reported as
`managed marker legend drifted — expected "<!-- candidate-tasks:v1 · … -->"`, so a genuinely
newer version is diagnosed as v1 corruption rather than as a forward-compatible unknown.

**The brief's precondition for reporting this is met.** A detector would not need
heuristics with false positives: `candidateRowRe` rejects every legacy row shape tested
(`- ⏳ \`tskflwctl task new …\` — <one line>`, `- ✅ done · wire the thing`,
`- ⚠️ partial — H1 mostly handled`, `- ⏳ H1 — handwritten legacy row`,
`- ⛔ won't do (see ADR-0003)`), and a scan of all 34 committed audits carrying a
`## Candidate tasks` section — **every one of which is legacy; zero managed v1 audits are
committed** — found **zero** lines matching the canonical grammar inside an unmanaged
section. "An unmanaged Candidate tasks section containing a canonical-grammar row" is
therefore a zero-false-positive signal on today's corpus.

**Affected contract:** "Unversioned Candidate tasks prose remains readable, is not linted
as v1", plus the forward-compatible-vs-corrupt distinction the brief asks for.

**Recommendation:** require a delimiter after the version
(`"<!-- "+CandidateTasksVersion+" "`, or accept `-->` immediately after), and add one lint
check that reports an unmanaged `## Candidate tasks` section containing a canonical-grammar
row as suspected marker corruption.

**Resolution:** The v1 marker now requires a token delimiter; explicit unknown
versions are diagnosed separately, and canonical rows under a damaged or missing
marker produce a focused suspected-corruption diagnostic while ordinary legacy
prose remains ignored.

#### L3. Themed finding-status colour is still transcribed beside the domain table, with no registry-coverage test  · **Status:** fixed

**File:** internal/theme/theme.go:108-125 | **Component:** theme
**Severity:** low · **Effort:** XS · **Urgency:** eventually

`FindingStatus` now reads the glyph from `domain.FindingStatusGlyph` (line 107) — the
single-source change this task is about — but keeps a hand-written `switch` for the
**colour**, whose `default:` arm (lines 123-124) is documented as the *unknown*-status case.
`internal/theme/theme_test.go:95-118` is a hand-transcribed seven-row table plus `""` and
`"bogus"`; nothing asserts coverage of `domain.FindingStatuses()`.

**Reproduction (Mutation E, coordinated).** Added
`{Status: "escalated", Glyph: "▲"}` to `findingStatusSpecs`. `go build ./...` was clean.
`domain.TestTallyFindings_EveryStatusLandsInABand` and
`render.TestFindingStatusOrder_CoversRegistry` both failed — those consumers *are* guarded —
and the machine-contract golden failed. But `theme.TestFindingStatus` and
`theme.TestCriterionStateReusesFindingGlyphs` both **passed** while the new status rendered
in the gray "unknown" colour, which is the colour this palette uses for the dropped/parked
band. `audit lint` no longer flags the value either, because the vocabulary now accepts it,
so nothing anywhere reports the mis-render.

This **narrows** a pre-existing gap rather than introducing one — before this change both
glyph and colour were transcribed here — and `theme.CriterionState` delegates correctly
(theme.go:128-152). It is reported because it is the precise answer to the brief's
second-pass question, "can a future status addition compile while leaving a consumer
wrong?", and because the change's stated contract is that "themed status rendering consume[s]
that source rather than transcribing glyphs": the glyph half is now shared, the colour half
is the one finding-status consumer with no exhaustiveness guard.

**Recommendation:** add a coverage assertion to `theme_test.go` requiring every
`domain.FindingStatuses()` value to resolve to a non-default colour, or move the colour into
`findingStatusSpec` beside `Glyph`.

---

**Resolution:** Finding-status colour lookup now reports whether a status was
intentionally classified, with a registry-coverage test requiring every domain
status to avoid the unknown fallback.

### Consumer inventory (every entry verified in the sandbox)

| Claimed consumer | Verified |
| --- | --- |
| `internal/domain/candidate.go` (new, 333 lines) | parser, lint, writer, status synchronizer; reuses the **existing** `blankFences` fence masker (finding.go:142) and `Span` (finding.go:54) rather than a parallel abstraction |
| `internal/domain/finding.go` | `findingStatusSpec`/`findingStatusSpecs` (lines 206-221) replace the old `findingStatuses` map, which is now derived from them; `FindingStatusGlyph`; `SetFindingStatus` calls `syncManagedCandidateStatus` (line 650) |
| `internal/domain/entity.go` | both audit templates became `var` to compose `candidateTasksScaffold()`; audit authoring guidance updated |
| `internal/core/finding.go` | `FindingEdit.Candidate *string`; `apply` chains candidate third, re-parsing between steps; `AuditLintIssues` gained `candidateIssues`; a misplaced `LintAudits` doc comment was moved onto the function |
| `internal/core/service.go`, `store.go` | `Lint()` folds `a.CandidateIssues`; `AuditWithFindings.CandidateIssues` added |
| `internal/store/auditstore.go` | `parseAuditWithFindings` returns candidate issues from the **same single body read** as findings (one `bodyText`), preserving the one-read contract |
| `internal/theme/theme.go` | `FindingStatus` consumes `domain.FindingStatusGlyph`; colour still local (**L3**) |
| All other finding-glyph consumers | `render.Style.FindingStatus` (style.go:155) → `theme.FindingStatus`; `theme.CriterionState` delegates; `render.go:811,822,865`; `tui/detail.go:1722`; `tui/help.go:142-147`. Grepping all seven glyphs across non-test Go found **no second finding-status mapping** in production — remaining literals belong to task-status / bucket / atlas vocabularies |
| `internal/cli/audit.go`, `lint.go` | `--candidate` flag, flag-required guard, `describeFindingEdit`, command help; both lint long-descriptions corrected |
| Generated help + docs | `go run ./internal/tools/docgen -out docs/cli` produced **no diff** |
| Golden/update guard | `goldenContentOutsideRevision` (**M1**) |
| README, CLAUDE.md, ARCHITECTURE, routines | read; no guidance tells an agent to hand-type a managed row, and no surviving text claims top-level `lint` excludes audit findings (grepped both claims across README/CLAUDE/docs/routines/adrs — empty) |
| Task `6g3ag8py12y9` | all 7 criteria `[x]`; decision + validation + closeout sections present |
| Follow-up `6gbpe6e8n87k` | `ready-to-start`, depends on `6g3ag8py12y9`; its criteria explicitly own finding **creation** and the placement of a new finding before a trailing Candidate section — credited nowhere below as shipped |

### Mutation probes (each restored from the sandbox baseline; focused test re-run green after)

| # | Mutation | Focused test predicted | Result |
| --- | --- | --- | --- |
| A | Drop `syncManagedCandidateStatus` from `SetFindingStatus` (finding.go:650) | `TestSetFindingCandidateAddReplaceRemoveAndStatusSync` | **FAILED** as predicted; `cli.TestAuditFinding_ManagedCandidateLifecycle` also failed |
| B1 | Duplicate-row check permissive (`count > 1` → `> 99`) | `TestLintCandidateTasksManagedRows` | **FAILED**: `missing issue containing "2 candidate rows"` |
| B2 | Missing-code check permissive (`if !ok` → `if false && !ok`) | `TestLintCandidateTasksManagedRows` | **FAILED**: `missing issue containing "references no parsed finding"`; `cli.TestLintFoldsManagedCandidateIssues` also failed |
| C | Split the combined edit into two sequential `TransformAuditBody` calls | `TestAuditFinding_CandidateRefusesLegacyWithoutWriting` | **FAILED**: the intermediate write left `**Status:** fixed` and a stamped `updated_at` after the candidate step errored |
| D | Change `open`'s glyph `○` → `◯` | legend/render drift protections | **FAILED** in 5 packages: `cli.TestGolden_MachineContract`, `render.TestStyle_Enabled_WrapsANSI`, three `domain` candidate tests, `theme.TestFindingStatus`, `tui.TestAuditDetailFindingIndex` |
| E | Add an 8th status `escalated`/`▲` (coordinated) | exhaustiveness guards | Compiles clean; tally + render-order + golden guards fire; **theme colour guard does not** → **L3** |
| F | Change a template **description** (contract field, not a body) + `-update` | revision gate | Gate **bypassed** for `template_show_security_json`, enforced for `template_list_json` → **M1** |

### Hostile boundary cases

Settled clean: CRLF input (domain writes bare LF, but `store.normalizeBody`/`replaceBodyStamped`
fold to LF and re-emit in the file's own ending — `TestAuditFinding_ManagedCandidatePreservesCRLFAndDryRun`
asserts zero bare LF at the CLI boundary, and `SetFindingNote` behaves identically on `main`,
so this is contained, not a candidate-specific defect); multibyte/emoji/Cyrillic candidate
text, including replacement after a status edit moved offsets; byte-offset preservation
across multibyte fenced content; a following `#`, `##`, `###`, `####`, `######` heading all
terminating the section with the row inserted before it; a candidate section at EOF (insert
and remove); explanatory `<!-- … -->` comments; duplicate rows; a missing finding code; a
wrong glyph; a stale status; a drifted marker legend; malformed list rows; status decorations
(`tracked by 6gbn4g1bb7wr` → row carries the bare token `tracked`, and lint compares tokens,
so a decorated status lints clean); two findings carrying the **same** candidate text (allowed,
as the contract requires); a fenced `## Candidate tasks` example outside the section; two
Candidate sections where one is managed and one legacy (the row lands in the managed one,
lint stays clean); and a body whose legacy section contains the literal string
`candidate-tasks:v1` **not** in comment position (correctly not treated as a marker).

Found: fenced content **inside** a managed section (**L1**), marker typo/version handling
(**L2**), whitespace accumulation (**M2**).

Prose (non-comment) inside a managed section is reported
(`line N is not a canonical candidate row …`) and blocks `--candidate` until removed. I
judged this intentional strictness rather than a defect: it fails closed and loudly, and the
contract allows only rows, blanks, and comments. Worth noting only that neither the lint
message nor the write refusal states the remedy (wrap it in an HTML comment or move it out).

Line numbers in candidate diagnostics are **body**-relative, not file-relative (reported
`line 27` for file line 35, an audit with 8 frontmatter lines). I initially took this for a
defect, then confirmed the sibling `findings:` near-miss diagnostic does exactly the same
(`line 31` for file line 39) — `NearMissFindingHeaders` enumerates body lines too. It is the
pre-existing repo-wide convention, so candidate lint is consistent with its neighbour and
this is not a defect of this change.

### Atomicity, ownership, and concurrency

`EditFinding` (core/finding.go:250-265) passes **one** closure to
`store.TransformAuditBody`, which reads the body, runs the whole transform, and returns
before any write on error (store/body.go:194-197); the write goes through `writeBody` under
`s.writeLock` with a `verifyUnchanged` content CAS (body.go:208-217). Core owns the
multi-field operation, the filesystem adapter owns durable CAS, Cobra only parses flags —
verified by Mutation C, which the committed test caught.

Proven end-to-end in the disposable space: a combined `--status fixed --candidate …` against
a **legacy** section exits 11 and leaves the file byte-identical (`shasum -c` OK, status still
`open`); a combined `--status open --candidate $'line1\nline2'` exits 11 with the multiline
refusal and leaves the prior status intact. `--dry-run` removal printed
`✔ would set H1's candidate row removed` with the file byte-identical, then the real removal
applied exactly one line. No success receipt is built from a stale or reloaded body — the
receipt is derived from the flags and the audit the transform returned.

Concurrency: `go test ./internal/store -race -count=25` over
`TestEditFindingCandidate_RetriesAroundConcurrentAppendPreservingAll`,
`TestEditFinding_RetriesAroundConcurrentAppend`, and all `TestTransformAuditBody_*` passed —
no flake, no duplicate row, no lost prose. The candidate edit recomputes against fresh prose
on retry because the transform is a closure over whatever body the store hands back, not
precomputed text. I did not infer atomicity from a helper's name.

### Second-pass questions, answered

- **Should candidate projection have reused an existing body-grammar abstraction?** It did:
  `blankFences` and `Span`, both pre-existing. The offset-validity invariant it depends on
  (fenced bytes overwritten in place, newlines untouched) holds and is verified. The one
  place the reuse is *incomplete* is row classification — **L1**.
- **Can `AuditWithFindings.CandidateIssues` become stale or diverge across fake/real stores?**
  No. Three implementations exist: `store.FS` (auditstore.go:30) derives it from the same
  `bodyText` as `ParseFindings` in one read; `core.fakeStore` (service_epic_test.go:167-176)
  calls the same `domain.LintCandidateTasks(body, findings)`; `countingAuditStore` delegates
  to the fake; `nopStore` returns nil, as a nop should. Nothing caches it, and scoped
  `audit lint` and the single-read top-level aggregation produced **byte-identical**
  diagnostics on the same drifted file.
- **Do optional rows undermine a stated mirror invariant?** No. The recorded decision makes
  rows explicitly optional and one-per-finding; lint enforces uniqueness and linkage but
  never absence, and the same candidate text on two findings is permitted by design
  (verified). No stated invariant claims completeness.
- **Can a future status addition compile while leaving a consumer wrong?** Yes — **L3**.
- **Do malformed/unreadable audits stay file problems?** Yes. An audit with broken
  frontmatter and a bogus v1 marker was counted as `1 unreadable file(s)` by both lint
  entry points and was never reclassified into candidate issues.

### Operational ergonomics (commands run as written)

Both routine specs' exact commands work, including the shell quoting of a proposed
`task new` inside `--candidate`:
`--candidate '`tskflwctl task new "Bound the retry queue" --epic 21-… --tags gateway --tier 2 --priority p2 --description "..."`'`
produced a well-formed row that re-lints clean; the tracked-handoff form
`--status "tracked by <id>" --note "…" --candidate "Tracked in planning/tasks/<id>-<slug>.md"`
landed all three fields in one write with the receipt
`✔ set H1 tracked by 6gbn4g1bb7wr, its resolution note and its candidate row`. Receipts
were exercised across the full flag matrix; `--pr` without `--status` still fails closed
(`--pr decorates a status — pass --status too`). Fresh default **and** security scaffolds
emit the derived marker and lint clean. The full documented fresh workflow
(`audit new` → `audit append` → `--candidate` → status change → scoped and top-level lint)
succeeds.

`audit append` does place a new finding *after* the scaffold's trailing Candidate section.
The row still lands correctly (any heading ends the row-only section) and lint stays clean,
and this is **explicitly owned by the deferred follow-up `6gbpe6e8n87k`**, so it is recorded
here as intentionally deferred authoring ergonomics, not a defect in the projection.

### Residual risks

- **No managed v1 audit is committed.** All 34 audits with a Candidate tasks section are
  legacy, so every piece of v1 evidence comes from synthetic fixtures and disposable spaces.
  The first real managed audit will be the first corpus exercise of **M2** and **L1**.
- **M1** is the highest-leverage item: it silently weakens the machine-contract gate that
  ADR-0008 exists to enforce, and it does so inconsistently between two goldens carrying the
  same field.
- Criterion 6 of `6g3ag8py12y9` says "the 10 legacy audits"; the corpus actually has 34
  audits carrying such a section. The *decision* (tolerate on read, never rewrite) is
  honoured for all of them, so this is a stale figure in the criterion's wording rather than
  an implementation gap. Flagged for the owner; I did not edit the task.
- Not exercised: a real multi-process race (only the in-process `testHookBeforeBodyWrite`
  seam), and property/fuzz coverage of the row grammar — the compact boundary probes above
  were used instead.
