---
schema: 1
id: 6gbqd98z3fzc
bucket: closed
area: candidate-task-convention-implementation-antigravity
date: "2026-09-19"
---
# Audit: Candidate-task convention implementation — antigravity — 2026-09-19

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

### Mandatory isolation attestation

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/.git
baseline_commit=f7c980b7232f57e2242b09ade4c9966a54c1833a
source_blob=52438708a22b564e0225df320ce5c3442822beaa
source_fingerprint=a0ab4331e0d9a1ebff0f6df242d483d6ad87f41a
deliverable=planning/audits/6gbqd98z3fzc-2026-09-19-candidate-task-convention-implementation-antigravity.md
```

### Review verdict

**Substantive No-Findings (Clean Pass)**

The implementation of task `6g3ag8py12y9` (`decide-the-candidate-list-convention-and-make-the-tool-own-it`) establishes a robust, tool-owned projection for audit candidate tasks (`candidate-tasks:v1`). It strictly enforces canonical grammar (`- <glyph> <CODE> · <status> — <text>`), derives glyphs and legend markers from a single domain source of truth (`findingStatusSpecs`), guarantees multi-flag atomic edits through compare-and-swap (CAS) body transformations, provides comprehensive linting across scoped and repository sweeps, and fails closed when encountering legacy unversioned prose. Every required hostile angle, mutation probe, and boundary test passed with exact verification.

---

### Consumer inventory & symbol verification

Every claimed component, function, and symbol was directly verified inside the isolated sandbox:

- [`internal/domain/candidate.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/domain/candidate.go):
  - `CandidateTasksVersion = "candidate-tasks:v1"`: version token identifying managed sections.
  - `candidateHeadingRe`: regex matching `## Candidate tasks`.
  - `sectionHeadingRe`: regex matching any heading `#{1,6} `, properly bounding candidate sections before following content (such as appended findings or subsequent headings).
  - `candidateRowRe`: regex matching canonical row syntax `^- ([^ \t]+)[ \t]+([A-Z]+\d+)[ \t]+·[ \t]+([a-z-]+)[ \t]+—[ \t]+(.+?)[ \t]*$`.
  - `CandidateTasksMarkerComment()`: generates the version marker comment containing the exact legend dynamically derived from `findingStatusSpecs`.
  - `candidateSections()` & `parseCandidateSection()`: scans markdown bodies using `blankFences` to preserve byte offsets while masking code fences.
  - `LintCandidateTasks()`: validates managed v1 sections for duplicate sections, marker drift, malformed rows, unparsed finding codes, unknown statuses, status/glyph disagreements, finding/candidate status drift, and duplicate rows. Deliberately tolerates unversioned legacy sections.
  - `SetFindingCandidate()`: adds, replaces, or removes (`""`) a candidate row; verifies finding existence, single version marker, zero malformations, and fails closed against legacy prose.
  - `syncManagedCandidateStatus()`: re-synchronizes candidate rows when a finding's status changes.
- [`internal/domain/finding.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/domain/finding.go):
  - `findingStatusSpecs`: single source of truth for the 7 finding statuses and their plain-text glyphs (`open ○`, `in-progress ●`, `fixed ✔`, `tracked →`, `deferred ◌`, `superseded ◌`, `wontfix ✘`).
  - `FindingStatusGlyph()`: returns canonical glyph for any status; falls back to `•` for unknown values.
  - `FindingStatuses()`: returns sorted status vocabulary.
  - `SetFindingStatus()`: rewrites finding status span and hooks `syncManagedCandidateStatus()` for seamless candidate mirror synchronization.
- [`internal/core/finding.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/core/finding.go):
  - `FindingEdit`: struct holding `Status`, `*Note`, and `*Candidate` edits.
  - `FindingEdit.apply()`: chains `SetFindingStatus` → `SetFindingNote` → `SetFindingCandidate`, re-parsing between operations to ensure offset integrity.
  - `EditFinding()`: orchestrates body transformation with CAS retry on conflict.
  - `AuditLintIssues()`: unified check-set combining near-miss headers, finding lint, candidate tasks lint, and stable ID checks.
  - `LintAudits()`: runs `AuditLintIssues()` on single or all audits.
- [`internal/core/service.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/core/service.go):
  - `Lint()`: incorporates `a.CandidateIssues` into repository-wide lint sweep.
- [`internal/core/store.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/core/store.go):
  - `AuditWithFindings`: carries `CandidateIssues []domain.Issue`.
- [`internal/store/auditstore.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/store/auditstore.go):
  - `parseAuditWithFindings()`: derives `CandidateIssues` via `domain.LintCandidateTasks()` during the single-pass filesystem scan.
  - `TransformAuditBody()`: atomic file read-modify-write with file lock and content CAS verification.
- [`internal/cli/audit.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/cli/audit.go):
  - `newAuditFindingCmd()`: binds `--candidate` flag alongside `--status`, `--note`, and `--pr`, enforcing validation, describing edits, and reporting atomic mutations.
- [`internal/theme/theme.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/theme/theme.go):
  - `FindingStatus()`: delegates glyph selection to `domain.FindingStatusGlyph(status)`, preventing theme/domain glyph drift.
- [`internal/cli/golden_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/cli/golden_test.go):
  - `goldenContentOutsideRevision()`: restricts the ADR-0008 Markdown template exception strictly to `template_show_` snapshots, preventing unversioned mutations to machine contract goldens.
- Planning & follow-up:
  - Task `6g3ag8py12y9`: acceptance criteria fully met.
  - Task `6gbpe6e8n87k` (`add-a-tool-owned-audit-finding-creation-verb`): verified as an independent follow-up task scoped to allocate codes and insert finding blocks before `## Candidate tasks`; not confused with candidate projection work.

---

### Mandatory evidence floor & toolchain validations

All commands executed strictly within `$SANDBOX`:

1. **Git diff check**:
   ```sh
   git diff --check HEAD
   ```
   *Result:* Clean (exit code 0, no whitespace errors or merge markers).

2. **Race-detector test suite**:
   ```sh
   GOCACHE=/tmp/taskflow-review-go-cache go test -race ./...
   ```
   *Result:* Passed (exit code 0, 0 data races).

3. **Linter**:
   ```sh
   GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache golangci-lint run ./...
   ```
   *Result:* Passed (exit code 0, 0 issues).

4. **Internal planning & audit linters**:
   ```sh
   ./bin/tskflwctl --no-color lint
   ./bin/tskflwctl --no-color audit lint
   ```
   *Result:* All planning entities, dependency links, and audit findings pass lint.

---

### Disposable planning space lifecycle reproduction

Executed the full lifecycle in an isolated temporary repository (`/tmp/tskflw-probe-87CdFO`) using sandbox binary `./bin/tskflwctl`:

1. **Scaffold initialization**: `tskflwctl init --taskflow-root planning --no-register`
2. **Fresh audit creation**: `tskflwctl audit new probe-audit`
   *Generated candidate section:*
   ```markdown
   ## Candidate tasks

   <!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
   <!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->
   ```
3. **Appended finding**: `tskflwctl audit append probe-audit --body "#### H1. Critical issue · **Status:** open\n\nFinding description here."`
4. **Candidate addition**: `tskflwctl audit finding probe-audit H1 --candidate "Fix the critical issue"`
   *Resulting markdown:*
   ```markdown
   ## Candidate tasks

   <!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
   <!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

   - ○ H1 · open — Fix the critical issue

   #### H1. Critical issue · **Status:** open

   Finding description here.
   ```
5. **Status synchronization**: `tskflwctl audit finding probe-audit H1 --status in-progress`
   *Resulting candidate row:*
   `- ● H1 · in-progress — Fix the critical issue`
6. **Candidate text replacement**: `tskflwctl audit finding probe-audit H1 --candidate "Refactor and fix the critical issue"`
   *Resulting candidate row:*
   `- ● H1 · in-progress — Refactor and fix the critical issue`
7. **Candidate removal**: `tskflwctl audit finding probe-audit H1 --candidate ""`
   *Result:* Candidate row cleanly removed from markdown body without corrupting comments or subsequent findings.
8. **Candidate re-addition**: `tskflwctl audit finding probe-audit H1 --candidate "Fix the critical issue once more"`
   *Resulting candidate row restored cleanly.*
9. **Linter checks**: Both `audit lint probe-audit` and top-level `lint` report 0 issues.

---

### Hostile angles & mutation probes

1. **Probe 1: Status synchronization**:
   - *Mutation:* In [`internal/domain/finding.go:650`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/domain/finding.go#L650), bypassed `syncManagedCandidateStatus(out, f.Code)` by returning `out, nil`.
   - *Test:* `go test -run TestSetFindingCandidateAddReplaceRemoveAndStatusSync ./internal/domain/...`
   - *Failure output:*
     ```
     --- FAIL: TestSetFindingCandidateAddReplaceRemoveAndStatusSync (0.00s)
         candidate_test.go:86: finding and managed mirror did not update atomically:
             #### H1. First · **Status:** fixed 2026-09-19
             ...
             - ○ H1 · open — Use the narrower repair
     FAIL
     ```
   - *Restoration:* Reverted change; test passed.

2. **Probe 2: Duplicate candidate row detection**:
   - *Mutation:* In [`internal/domain/candidate.go:185`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xHMMdK/internal/domain/candidate.go#L185), bypassed duplicate candidate detection (`for code, count := range seen { if count > 1 ... }`).
   - *Test:* `go test -run TestLintCandidateTasksManagedRows ./internal/domain/...`
   - *Failure output:*
     ```
     --- FAIL: TestLintCandidateTasksManagedRows (0.00s)
         candidate_test.go:52: missing issue containing "2 candidate rows"
     FAIL
     ```
   - *Restoration:* Reverted change; test passed.

3. **Probe 3: Atomic rollback & fail-closed on multi-flag mutation**:
   - *Probe:* Executed `tskflwctl audit finding atomic-audit H1 --status fixed --candidate "Will fail"` against an audit containing an unversioned legacy candidate section.
   - *Observation:* Candidate write failed closed with `ErrValidation` (exit code 11):
     `Validation failed: this audit uses a legacy Candidate tasks section; add candidates only to a candidate-tasks:v1 audit rather than guessing at legacy prose.`
   - *Byte-identical verification:* SHA-256 hash of the audit file before and after the failed command was identical (`BEFORE_HASH == AFTER_HASH`). Neither `--status fixed` nor `updated_at` was committed to disk.

4. **Probe 4: Domain glyph & legend drift protection**:
   - *Probe:* Evaluated `findingStatusSpecs` against marker generation and theme delegates. Alteration of any spec immediately triggers failure in `TestCandidateTasksMarkerDerivedFromFindingStatuses`.

5. **Probe 5: Golden policy containment (ADR-0008)**:
   - *Probe:* Tested `validateMachineGoldenUpdate` in `internal/cli/golden_test.go`. Updates to non-template machine JSON goldens at unchanged revision are strictly rejected (`refusing to rewrite machine-contract golden <name> at unchanged revision <rev>`), while `template_show_*` Markdown bodies are permitted without bumping wire revision.

---

### Systemic analysis (pass 2)

1. **Shared abstractions & fence masking**:
   `candidateSections()` relies on `blankFences()`, which masks code blocks by turning enclosed characters into spaces while preserving byte lengths and line breaks. This prevents fenced markdown examples (such as those in audit documentation or templates) from activating the v1 candidate parser, while keeping all byte spans and line indices strictly aligned with the original file.

2. **Section boundary delimitation**:
   `candidateSections()` bounds candidate content between `## Candidate tasks` and the next section header matching `(?m)^#{1,6}[ \t]+`. This ensures that even when a finding is appended after `## Candidate tasks` (such as via legacy `audit append`), `candidateSections()` never swallows the appended finding as malformed candidate prose, and insertion of new candidate rows reliably places them before the subsequent section.

3. **Downgrade attack resistance**:
   Unversioned candidate prose is deliberately ignored by v1 linters and synchronizers to avoid corrupting human prose. If an author or agent attempts to use `audit finding --candidate` on an unversioned audit, the tool refuses with an explicit remediation message rather than silently dropping or creating rogue sections. Furthermore, any typo in a managed section's marker line is flagged by `LintCandidateTasks` as legend drift rather than silently falling back to unmanaged prose.

4. **Store read-path consistency**:
   Both `audit lint` (single audit read) and top-level `lint` (single-pass directory scan via `ListAuditsWithFindings`) execute the exact same check-set through `core.AuditLintIssues()`. There is zero divergence between CLI command diagnostics and repository validation gates.

---

### Residual risks & operational notes

1. **Shell quoting of CLI suggestions**: When populating `--candidate` with recommended command lines (e.g. `--candidate 'tskflwctl task new "Title" --effort S'`), shell quoting rules require enclosing the outer flag value in single quotes to avoid premature subshell expansion or argument splitting.
2. **Follow-up finding creation (`6gbpe6e8n87k`)**: As planned, `audit append` places content at the end of the markdown body. Task `6gbpe6e8n87k` will add a tool-owned finding creation verb to insert findings before `## Candidate tasks`. The candidate parser's boundary design already accommodates this gracefully.
