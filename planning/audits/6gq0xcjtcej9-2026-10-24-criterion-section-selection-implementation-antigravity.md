---
schema: 1
id: 6gq0xcjtcej9
bucket: closed
area: criterion-section-selection-implementation-antigravity
date: "2026-10-24"
updated_at: "2026-10-25"
---
# Audit: Criterion section selection and completion guards — antigravity — 2026-10-24

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
final guarded audit transfer. Your handoff prompt prints exact `SOURCE_ROOT` and `AUDIT_REL`
values. Substitute them below; never derive them from your current directory, which may be a
different checkout. Then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="<SOURCE_ROOT from your handoff prompt>"
AUDIT_REL="<AUDIT_REL from your handoff prompt>"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)" && cd "${SANDBOX:?create failed: stop and report the blocker}"
```

Every later block begins from `${SANDBOX:?…}`, so it cannot run in the shared checkout after a
failed `create`. Run the helper from `$SOURCE_ROOT`, not from the sandbox: you may mutate the
sandbox's copy while probing, and only the read-only source copy can be trusted to verify your work.

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. The sandbox's push URLs are disabled; never re-enable them or
push to the source path. If creation fails, report the blocker; never fall back to the shared
checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run
the helper's fail-closed verification:

```sh
cd "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX" &&
  git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent or nested `.git`, a
removed or weakened push guard, source-deliverable drift, and empty reports; it copies back only
the assigned audit through a same-directory atomic rename. A successful transfer records a
helper-owned receipt and prints it (`transfer=succeeded` with the transferred blob). A refusal exits
2 and records nothing. Exit 3 means the report was already delivered when a later step failed: run
`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"` and put its
output, or its refusal, in your handoff. Put the printed receipt in your handoff message. Do not
copy anything else manually. Leave the workspace in place and report its path until the
implementation owner confirms delivery with that same `receipt` command. On refusal, preserve it and
report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

## Review brief

Challenge whether the criterion-section repair actually removes silent loss of
criteria without inventing a new write target, hiding diagnostics, or breaking
the approved legacy/empty-section behavior. Checked ACs and owner green tests are
claims to challenge, not evidence of correctness by themselves.

The assigned reviewer is named in this audit's preamble. Primary lanes:

- **Codex:** mutation authority, fresh-body CAS/retry, lifecycle force/preview/no-op
  gates, and real CLI refusal-before-effects. Cross-check one parser fixture and
  the error-to-machine-output boundary.
- **Antigravity:** independently derive section boundaries, test the language and
  false-positive/false-negative policy, inventory consumers, and challenge wire,
  help and TUI truthfulness. Cross-check one forced completion and one edit.

Both own the shared intended contract. Do not read the sibling report until your
independent report is final. No finding quota; a demonstrated systemic defect is
more useful than speculative style complaints or a blanket clean verdict.

## Review target

- Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
- Base HEAD: `0d13d616584e9b4fe7949fe501a5709bfc281f4d`. The implementation is
  **uncommitted**, including new untracked Go tests and selector code. Capture the
  dirty tree with the mandatory sandbox helper before inspecting implementation.
  HEAD alone contains the earlier creation closeout, not this repair.
- Implementing task: `planning/tasks/6gpxh6ncaqpw-prevent-narrative-headings-from-hiding-task-criteria.md`.
  Its approved-policy section records the owner's three decisions. Governing
  machine policy: `planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md`.
- Start at `internal/domain/criterion_section.go` (`selectCriterionSection`,
  `isAcceptanceSectionHeading`) and `internal/domain/body.go` (enumeration,
  count/list/tally/unexplained reads, every criterion edit, and lint).
- Follow `internal/core/service_task.go`, `internal/core/task_lifecycle.go`
  (`validateExistingLifecyclePlan`) and `internal/core/service.go` for application
  reads, CAS callbacks/retries, force/no-op ordering and archived-task lint.
- Primary adapters: `internal/cli/task.go` and `internal/tui/detail.go`
  (`criterionRollup`). Wire/authoring owners: `internal/wire/wire.go`,
  `internal/domain/entity.go`, `internal/cli/testdata/golden/`.
- Focused shipped tests are in `internal/domain/criterion_section_test.go`,
  `internal/core/criterion_section_test.go`,
  `internal/cli/task_criterion_section_test.go`, and
  `internal/tui/criterion_section_test.go`. The domain fuzz target is
  `FuzzAcceptanceCriteria` in `internal/domain/body_fuzz_test.go`.
- Produce a verified repository-wide **consumer inventory**, not just these entry
  points: find all count/list/tally/unexplained callers, all criterion mutations,
  every path into completion, lint/repair and raw-body repair. Changed reader
  signatures require error handling; compilation alone does not prove honest output.
- Explicitly tracked, not fixed here: `planning/tasks/6gq03e44pcyz-guard-task-completion-against-unreadable-criterion-markers.md`
  covers malformed `- []` under a canonical heading being linted but omitted from
  enumeration/completion. `planning/tasks/6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md`
  owns tracked-reason policy. You may challenge these boundaries only with a
  concrete new in-scope consequence; do not relabel their known repros as new bugs.
- Ignore unrelated captured planning work, including
  `planning/tasks/6gnqadwhhz54-run-through-tui-task-creation-acs.md` and any parallel
  performance task. Do not review the already-shipped TUI creation feature again.

## Intended contract to challenge

1. Exactly one explicit case-insensitive `Acceptance criteria` heading, allowing
   parenthetical notes such as `(draft)`, or the supported `Acceptance` alias
   selects the criterion section. Both headings together are ambiguous, not a
   precedence rule that discards one. Fenced examples do not select sections.
2. Narrative mentions before/after a recognized section cannot shadow it. With no
   recognized section, a nonstandard acceptance heading containing a checklist is
   a validation problem, not a success-shaped zero. Genuinely absent and explicitly
   empty criterion sections remain valid empty reads.
3. Reads, criterion transforms, lint and completion use the same selection policy.
   Invalid structure yields actionable validation diagnostics. The TUI keeps the
   task body/repair context visible rather than disguising failure as 0/0.
4. Every criterion edit revalidates the actual guarded body, including on conflict
   retry and idempotent edits. Refusals and previews preserve source bytes/modes;
   selected source identity and operation time remain bound by existing contracts.
5. Completion validates structure before force handling and same-status no-ops.
   Force may override known unmet criteria, not unrecognized/ambiguous structure.
   Both `task complete` and generic `task move ... completed` must agree.
6. Raw body reads/replacement and editor repair remain possible. No guessed merge,
   automatic heading rewrite, persisted schema migration, new persistence port or
   changed generic `Section` substring semantics is intended.
7. Revision `1.90` is declared `NOT ADDITIVE`; read/exit/tally behavior changes are
   machine-visible despite unchanged JSON shapes. Goldens, authoring conventions,
   generated reference and source-adjacent guide must accurately describe this.

## Mandatory evidence floor

Both reviewers must perform these bounded checks in their own sandbox:

- Build the fresh local CLI, then run
  `go test -race ./internal/domain ./internal/core ./internal/cli ./internal/tui -run '^TestCriterion' -count=1`
  and `go test ./internal/wire`. Distinguish executed checks from inherited owner
  gates. Do not regenerate goldens to turn an unexpected failure green.
- Derive at least one accepted hostile document and one refused document yourself,
  using literal source and an independently stated expected criterion set. Exercise
  the actual CLI/core caller, not only the selector or shared preparation helpers.
  Check metadata/JSON/exits as applicable, and exact unchanged bytes/modes after
  refused mutations. Helpers must not calculate expectations using the production
  selector, its regex, or the domain test `countCriteria/listCriteria` wrappers.
- **Codex minimum:** execute completion with/without force and preview on invalid
  active and already-completed tasks; execute one actual guarded criterion edit
  refusal. Inspect and independently challenge fresh-body retry validation, rather
  than treating `criterionRetryStore`'s simulated conflict as a real-FS race proof.
- **Antigravity minimum:** make a compact input/expected-output matrix not copied
  from `TestCriterionSectionSelection`, with at least two additional lexical or
  nesting arrangements. Run one accepted edit through the CLI and compare the
  entire expected body independently. Compare task info/ac, ordinary lint, and the
  TUI diagnostic for an invalid document. A direct Model test is not a live TUI run.
- Execute **one compiler-valid mutation per reviewer**, with the intended regression
  named before execution. Codex: attack structural validation before force/no-op or
  fresh-body callback validation. Antigravity: attack a lexical/range/diagnostic
  invariant other than the owner's first-wins mutation. Report exact edit, test,
  expected failure, observed result and restored green. A surviving mutation is
  evidence to analyze; compilation failure is not a killed mutation.
- Execute the proposed explicit heading repair on each of your refused diagnostic
  classes and rerun the same read/edit/completion check. Confirm that the advice is
  usable and does not require a hidden automatic migration. Keep repair user-driven.

Owner evidence may be inherited, clearly attributed: full `just test` race suite,
`just build`, `just lint`, planning/link/audit lint, focused criterion race tests,
a ten-second fuzz run exceeding 424,000 cases, and two overlay mutations restoring
first-wins selection or force bypass. Your mutation must add useful independent
evidence rather than simply quote this ledger.

## Required hostile angles

### Codex: lifecycle and mutation authority

- Can any early return, completed no-op, move alias, batch command, preview or
  typed override avoid the structural check? Separate intended unmet-criteria
  force from guessing which section exists. Confirm no-op/idempotent writes still
  verify the structure they would target.
- Can a valid projected body become ambiguous before persistence or during CAS
  retry? Are the same source/clock retained while the new body is parsed afresh?
  Keep a portable callback probe distinct from a real adapter/producer trace.
- Are structural failures propagated or swallowed by count/list/tally wrappers,
  command rendering, error envelopes, or lint? Inspect repair access and ensure an
  error did not accidentally become permission to follow a different source/path.

### Antigravity: parser and consumer falsification

Settle each hypothesis as reproduced, falsified by an executed probe, or unresolved:

- **A narrative wins indirectly.** Parent/child heading levels, later canonical
  sections, surrounding checklists or an empty canonical section make selection
  count the wrong lines or falsely reject an ordinary note.
- **The language description overpromises.** Parenthetical suffixes, capitalization,
  closing hashes, spacing, fences or newline variants differ between help, selector,
  edits and lint. Probe a real mismatch, not just an aesthetic regex preference.
  Distinguish a changed in-scope behavior from a preexisting unsupported Markdown
  form; a general parser replacement is not a prerequisite to a bounded fix.
- **All tests share the same mistaken oracle.** The helper wrappers, fabricated
  portable bodies, or line-based expectations hide a broad selection/edit class.
  Use independent literal bodies and expected bytes, not another regex clone.
- **Compatibility labels hide semantic drift.** Revision/classification, schema task
  conventions, task info/ac error output and snapshots disagree with reachable
  behavior. Check actual JSON producers, not just changed revision strings. No new
  optional wire fields are claimed; don't invent an optional-field test requirement.
- **Presentation launders the failure.** The TUI or another consumer still says 0/0,
  hides body repair, or truncates away all actionable diagnostic context. Check one
  constrained viewport in addition to the existing 160×40 fixture; avoid a general
  interaction redesign or unsupported claims about live runtime order.

### Both: separate systemic pass and scope discipline

After the checklist, re-review shared helpers and adjacent consumers for a class
of failures the examples might miss. Test section ownership and error propagation
as contracts, not individual strings. Check actual producer reachability before
claiming corruption, a permission failure, or universal Markdown support.

Archived-task structure is newly linted; unrelated archived field/state nags should
not return. Ordinary `lint` and real `lint --fix` diagnose residual structural errors;
`lint --fix --dry-run` is a proposal-only repair preview with no post-repair lint
pass. Do not treat that preexisting preview contract as a new finding here. Check
that neither real nor preview repair silently rewrites headings. Do not conflate
the tracked malformed-marker gap with header ambiguity or demand a vocabulary change.

## Validation and restoration

- Follow the injected isolation protocol before implementation inspection. Build
  and run tests only in the independent sandbox; the source checkout is read-only.
  Do not launch agents or change branches/history beyond the helper baseline.
- Put caches and scratch evidence under a reviewer-owned temporary directory
  outside the sandbox tree. Scratch probe files added to the sandbox must be
  removed after use; restore only reviewer-owned mutations to the captured helper
  baseline, never to the source checkout's HEAD or its current dirty files.
- Owner `docs-check` sees intentionally uncommitted generated pages in the source.
  Your helper baseline commits those pages: run `just docs-check` normally there.
  Any new generated drift is evidence, not an excuse to waive the check. Restore
  generator/probe changes before transfer, including accidental golden updates.
- With the sandbox's fresh CLI run planning/link lint and audit lint for the assigned
  report, plus `git diff --check`. Unrelated captured changes must remain byte-identical.
  Restore probes so only the assigned audit differs; use the source-owned helper's
  verify/transfer protocol and return its printed receipt. Preserve the sandbox on
  refusal or an incomplete review; never copy reports manually into the source.

## Deliverable

Keep the brief intact and replace the empty report slot below. Report findings
first, then a concise evidence ledger and bounded surviving-hypothesis list.
Leave every finding open, using the injected exact finding grammar and repository
finding verbs where practical. Do not settle findings or alter task/Thread states.

For each finding give root cause, reachable consequence, exact path/line or command,
minimal reproduction, recommendation and uncertainty. Consolidate shared causes.
Distinguish product defects, test gaps, preferences, defense-in-depth and tracked
out-of-scope work. Verify named symbols/tests in the captured sandbox before citing.

A ready/no-findings result needs independent hostile probes, a meaningful mutation,
restored green and a separate systemic pass. Permission, isolation, probe or transfer
failure means incomplete/degraded review, not unqualified readiness. Include snapshot
identity, checks executed versus inherited, verification attestation with truthful
`transfer=pending`, and the transfer receipt in your handoff. Do not edit the report
after transfer to claim delivery; the owner confirms it with the helper receipt.

## Reviewer report

### Findings

#### M1. Ancestor and H1 headings mentioning acceptance false-reject ordinary checklists and misdirect heading repair · **Status:** fixed

**File:** internal/domain/criterion_section.go:39-63, 75-80 | **Component:** domain/criterion-section
**Effort:** S · **Urgency:** soon

**Classification:** Product defect (false-positive validation refusal on valid tasks, misattributed repair advice that corrupts the task H1 and sweeps non-criterion checkboxes into criteria, and immediate lint/read failure on default-scaffolded tasks titled `Acceptance` or `Acceptance criteria`).

**Root cause:**
`selectCriterionSection` (`internal/domain/criterion_section.go:31-82`) scans every unfenced ATX heading matched by `bodyHeadingRe` (`internal/domain/body.go:19`, which matches `#{1,6}`), including the level-1 document H1 (`# <Title>`) that `taskBodyTemplate` (`internal/domain/entity.go:304`), `Service.NewTask` (`internal/core/service_task.go:1007-1018`), and `FS.RenameTask` (`internal/store/tasks.go:130-156`) write at the top of task bodies. Whenever any heading contains `"acceptance"` case-insensitively (`strings.Contains(strings.ToLower(h[2]), "acceptance")` at `internal/domain/criterion_section.go:47`), `selectCriterionSection` appends it to `spans` and pushes its index onto `open` (`:48-50`).

Two interacting defects follow:
1. **Non-acceptance child headings do not shadow open ancestor spans for checklist attribution (`internal/domain/criterion_section.go:41-62`).** The `open` stack only pops when `spans[open[len(open)-1]].level >= level` (`:41`). A level-1 H1 (`# Improve acceptance test runner`, `level == 1`) stays on `open` across every `##` section (`level == 2`), and a level-2 narrative heading (`## Integrated acceptance notes`, `level == 2`) stays on `open` across every `###` subsection (`level == 3`), even when a deeper child heading (`## Notes` or `### Implementation steps`) does not mention `"acceptance"`. When any checkbox line appears under that non-acceptance child heading (`:57-62`), `for _, index := range open { spans[index].checklist = true }` marks **all** open ancestors as owning an acceptance checklist, and `:75-79` reports the **first** span in `spans` (`spans[0]`, the outermost ancestor) rather than the innermost heading that directly owns the checklist.
2. **Level-1 H1 document titles are treated as candidate/recognized criterion sections (`internal/domain/criterion_section.go:22-25, 47-53`).** `isAcceptanceSectionHeading` does not check heading level (`level >= 2`), whereas canonical criterion sections are `##` (or deeper) headings (`taskBodyTemplate` at `internal/domain/entity.go:308` and `AddCriterion` at `internal/domain/body.go:404` both materialize `## Acceptance criteria`).

**Reachable consequences:**
1. **Valid tasks with no criterion section are falsely rejected (`exit 11`).** Any task whose H1 title contains `"acceptance"` (or whose body has a narrative `## ... acceptance ...` heading with a `### Notes` child subsection) and that uses an ordinary Markdown checklist under a non-acceptance heading (`## Notes` or `### Notes`) has no criterion section (`total: 0`). Instead of returning an empty criterion set, `CountAcceptanceCriteria` (`internal/domain/body.go:307`), `ListAcceptanceCriteria` (`:533`), `TallyCriteria` (`:512`), `UnexplainedCriteria` (`:487`), and `LintAcceptanceCriteria` (`:715`) all return `ErrValidation`:
   - CLI `task info <ref> [--json]`, `task ac <ref> [--json]`, `lint [--json]`, `task ac` mutations, and `task complete <ref> [--force]` / `task move <ref> completed [--force]` fail with exit code `11`.
   - TUI `criterionRollup` (`internal/tui/detail.go:824`) renders `unavailable: invalid task body: unrecognized acceptance checklist heading ...`.
2. **Misattributed line-1 diagnostic and destructive repair advice.** When a task whose H1 title mentions `"acceptance"` (`# Align acceptance lint` at line 1) contains an ordinary checklist under `## Objective` (line 5) and a genuinely nonstandard `## AcceptanceCriteria` checklist at line 9, `selectCriterionSection` blames line 1 (`unrecognized acceptance checklist heading "Align acceptance lint" at line 1; rename it to Acceptance criteria (or Acceptance) before reading, editing or completing`) instead of line 9 (`"AcceptanceCriteria" at line 9`). Following that repair advice literally by renaming line 1 to `# Acceptance criteria`:
   - Overwrites the task's H1 document title (`domain.FirstH1` at `internal/domain/body.go:880`),
   - Leaves `## AcceptanceCriteria` at line 9 unfixed, and
   - Causes `selectCriterionSection` to select the level-1 `# Acceptance criteria` span covering the entire file (`lines 1..end`), sweeping both the ordinary note under `## Objective` (`#1: ordinary objective note`) and the item under `## AcceptanceCriteria` (`#2: real criterion`) into `task info` (`"total":2`) and `task ac`.
3. **Default `task new` scaffold fails its own linter and readers.** Creating a task titled `"Acceptance"` or `"Acceptance criteria"` via `tskflwctl task new "Acceptance" --epic <epic> --tags <tag>` writes `# Acceptance` at line 2 and `## Acceptance criteria` at line 8 from `taskBodyTemplate` (`internal/domain/entity.go:304-309`). `selectCriterionSection` sees `len(recognized) == 2` (`lines 2, 8`), so `lint`, `task info`, `task ac`, and `task complete` immediately fail with exit code `11` (`2 acceptance-criteria sections at lines 2, 8`), violating `internal/core/service_task.go:1019` (`new must not scaffold a file its own linter rejects`) and `internal/domain/entity.go:71-72` (`any title is accepted ... and the full title is kept as the body H1`).

**Minimal reproduction (verified in isolated sandbox):**
```sh
# 1. Valid task with H1 mentioning "acceptance" and an ordinary checklist under ## Notes:
./bin/tskflwctl -C "$SCRATCH/repo" task new "Improve acceptance test runner" --epic 01-quality --tags testing --no-input
./bin/tskflwctl -C "$SCRATCH/repo" task set improve-acceptance-test-runner \
  --body $'# Improve acceptance test runner\n\n## Notes\n\n- [ ] ordinary note checkbox\n' --no-input
./bin/tskflwctl -C "$SCRATCH/repo" task info improve-acceptance-test-runner --json
# -> exit 11: "invalid task body: unrecognized acceptance checklist heading \"Improve acceptance test runner\" at line 1; rename it to `Acceptance criteria` (or `Acceptance`) before reading, editing or completing"

# 2. Valid task with narrative H2 and ordinary checklist under child H3:
./bin/tskflwctl -C "$SCRATCH/repo" task set alpha \
  --body $'# Normal title\n\n## Integrated acceptance notes\n\nProse only.\n\n### Notes\n\n- [ ] ordinary note checkbox\n' --no-input
./bin/tskflwctl -C "$SCRATCH/repo" task info alpha --json
# -> exit 11: "invalid task body: unrecognized acceptance checklist heading \"Integrated acceptance notes\" at line 3; rename it to `Acceptance criteria` (or `Acceptance`) before reading, editing or completing"

# 3. Default `task new` with title "Acceptance":
./bin/tskflwctl -C "$SCRATCH/repo" task new "Acceptance" --epic 01-quality --tags testing --no-input
./bin/tskflwctl -C "$SCRATCH/repo" lint --json
# -> exit 11: field "acceptance": "invalid task body: 2 acceptance-criteria sections at lines 2, 8; merge them into one `Acceptance criteria` section before reading, editing or completing"
```

**Recommendation:**
1. Exclude level-1 document headings (`level == 1`) from criterion-section selection and candidate spans (`if level >= 2 && strings.Contains(...)`), or only treat `level >= 2` headings as section candidates so `# <Title>` remains the document H1 owned by `FirstH1`.
2. Track every unfenced ATX heading's level on the active heading stack (or close/suppress outer non-recognized spans when entering a child heading, or attribute a checkbox only to the innermost active heading) so a checkbox under `### Notes` is attributed to `### Notes` rather than an ancestor `## Integrated acceptance notes`, and nested nonstandard acceptance headings report the innermost checklist-owning heading rather than `spans[0]`.
3. Add regression cases covering `# <Title>` H1s containing `"acceptance"`, `# Acceptance` with `## Acceptance criteria`, and `## Integrated acceptance notes` followed by `### Notes` containing an ordinary checklist.

**Uncertainty:** Low. All three consequences were reproduced end-to-end through the sandbox binary (`./bin/tskflwctl`) and domain/core entry points.

---

**Resolution:** Confirmed the captured-baseline failures before fixing them.
    Excluded H1 document titles from criterion section candidates and attributed
    nonstandard checklists only to their immediately owning heading, while
    preserving nested subsections inside recognized criteria. Domain and CLI
    regressions cover scaffolded Acceptance/Acceptance criteria titles, ordinary
    narrative descendants, exact diagnostic heading/line, refusal before
    effects, and heading-only repair without sweeping ordinary notes into
    criteria.

#### L1. Criterion section selection tests omit H1 titles and do not assert selected criterion text or diagnostic target · **Status:** fixed

**File:** internal/domain/criterion_section_test.go:10-54 | **Component:** domain/criterion-section-tests
**Effort:** XS · **Urgency:** soon

**Classification:** Test gap.

**Root cause:**
All test bodies in `internal/domain/criterion_section_test.go:14-30`, `internal/core/criterion_section_test.go:30-33, 65-68`, `internal/cli/task_criterion_section_test.go:14-17, 82`, and `internal/tui/criterion_section_test.go:16` use headless `##`/`###` fragments without a leading `# <Title>` H1. In addition:
- In `TestCriterionSectionSelection` (`internal/domain/criterion_section_test.go:49`), valid selection cases check only `count.Total != tc.total || len(criteria) != tc.total || len(LintAcceptanceCriteria(tc.body)) != 0`. In every multi-section valid fixture (`canonical after narrative`, `nested headings`, `fenced headings` at `:14, :17, :18`), both the canonical section and each decoy section contain exactly 1 unchecked checkbox (`tc.total = 1`), and the test never asserts `criteria[0].Text == "actual"`.
- In the nonstandard-heading cases (`:26-29, 38`), the test checks only substring `tc.problem = "unrecognized acceptance checklist heading"` without asserting the blamed heading title or 1-based line number.

**Reachable consequence:**
A selector bug that picks a 1-item narrative/fenced decoy section instead of the 1-item canonical section, or that blames an outer/H1 heading at line 1 instead of the offending heading line (as demonstrated in M1), is not caught by `TestCriterionSectionSelection`.

**Minimal reproduction:**
Inspect `internal/domain/criterion_section_test.go:14-52`: `tc.total` is `1` for `canonical after narrative` (`- [ ] narrative`, `- [ ] actual`, `- [ ] also narrative`), `nested headings` (`- [ ] actual`, `- [ ] outside`), and `fenced headings` (`- [ ] example`, `- [ ] actual`), and `criteria[0].Text` is unread.

**Recommendation:**
Include realistic `# <Title>` H1 headings (including titles mentioning `"acceptance"`) in `TestCriterionSectionSelection`, assert exact selected criterion text and state (`criteria[0].Text == "actual"`), and assert the exact quoted heading and line number in rejected cases.

**Uncertainty:** None; verified directly in `internal/domain/criterion_section_test.go`.

---

**Resolution:** Strengthened the domain selection table to compare complete
    literal Criterion values and exact tallies, including text, index, checked
    state and disposition. Invalid cases assert heading title and line across
    every reader and lint. Added realistic H1/nested fixtures and CLI scaffold,
    diagnostic, byte/mode preservation and minimal repair regressions; the new
    cases failed against the reviewed selector and passed after the ownership
    fix.

### Verdict and scope discipline

- **Verdict:** **Findings open (`M1`, `L1`).** On headless `##`/`###` bodies and bodies whose `# <Title>` H1 and ancestor headings do not contain `"acceptance"`, the implementation satisfies the intended contract: canonical `Acceptance criteria` (with optional parenthetical suffix) and legacy `Acceptance` are selected regardless of narrative mentions before or after; duplicate recognized sections fail closed; nonstandard acceptance headings with checklists fail closed across reads, edits, lint (including archived tasks), CLI completion (`task complete` and `task move ... completed` with/without `--force` and `--dry-run`), and TUI detail rendering; and fresh-body CAS retries revalidate structure before persisting. However, `M1` is a reachable in-scope product defect in `selectCriterionSection` when a task's `# <Title>` H1 or an ancestor heading mentions `"acceptance"`, and `L1` is the corresponding unit-test oracle gap.
- **Tracked out-of-scope boundaries respected:**
  - `planning/tasks/6gq03e44pcyz-guard-task-completion-against-unreadable-criterion-markers.md` (malformed `- []` under a canonical heading flagged by `LintAcceptanceCriteria` at `internal/domain/body.go:728-739` while omitted from `scanAcceptanceCheckboxes` enumeration/completion when another valid item exists) is not relabeled as a finding here. Note that under a *nonstandard* heading with no canonical section, `selectCriterionSection` (`internal/domain/criterion_section.go:57-58`) does treat `malformedCheckbox(line)` as checklist evidence.
  - `planning/tasks/6gkkzvr92waq-align-tracked-criterion-writes-and-lint-reason-semantics.md` (tracked criterion reason semantics) is not relabeled here.
  - `lint --fix --dry-run` (`internal/core/planning_maintenance.go:88-96`) is a proposal-only repair preview with no post-repair `s.Lint()` pass; verified that neither real `lint --fix` nor `lint --fix --dry-run` rewrites headings.

### Isolation attestation and captured snapshot

Final verification executed from `/private/tmp/isolated-review.q9iXJq` via `"/Users/andyeschbacher/git/andy-esch/taskflow/scripts/isolated-review-workspace.sh" verify --sandbox "/private/tmp/isolated-review.q9iXJq"`:

```text
sandbox_path=/private/tmp/isolated-review.q9iXJq
git_dir=/private/tmp/isolated-review.q9iXJq/.git
baseline_commit=42ffe4533889f1733c9d58f132db4b9200b4e4a3
source_blob=09005241e50fdc48d1053b0e2a743a0d28624a5e
source_fingerprint=798dd4c06055582289c6c3c6b558a596a49c24f6
deliverable=planning/audits/6gq0xcjtcej9-2026-10-24-criterion-section-selection-implementation-antigravity.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

- **Source checkout (`SOURCE_ROOT`):** `/Users/andyeschbacher/git/andy-esch/taskflow` (untouched during review; pre-capture base `HEAD~1`: `0d13d616584e9b4fe7949fe501a5709bfc281f4d`).

### Checks executed vs. inherited

- **Independently executed in `/private/tmp/isolated-review.q9iXJq`:**
  - `just build` -> built fresh `./bin/tskflwctl` (`v0.23.0-160-g42ffe453`).
  - `go test -race ./internal/domain ./internal/core ./internal/cli ./internal/tui -run '^TestCriterion' -count=1` -> PASS (`internal/domain` 1.143s, `internal/core` 1.132s, `internal/cli` 1.669s, `internal/tui` 2.033s).
  - `go test ./internal/wire` -> PASS (0.315s).
  - `just docs-check` (`go run ./cmd/gendocs` + `git diff --exit-code docs/cli`) -> PASS (exit 0, zero diff).
  - Independent CLI/core/domain/TUI hostile probes in external scratch directory `/private/tmp/antigravity-review-scratch-6gq0xcjtcej9` (with isolated `XDG_CONFIG_HOME`, subsequently removed) and temporary TUI viewport test (subsequently removed and verified clean).
  - Independent compiler-valid mutation in `internal/domain/criterion_section.go:41`, test kill verification, and restoration to clean baseline.
  - Sandbox planning/link/audit lint and whitespace check: `./bin/tskflwctl -C . lint`, `./bin/tskflwctl -C . lint --links`, `./bin/tskflwctl -C . audit lint 6gq0xcjtcej9`, and `git diff --check` -> PASS.
- **Inherited from owner baseline (attributed, not re-run in full):**
  - Full repository `just test` race suite across unrelated packages, `just lint` (`golangci-lint`), 10-second `FuzzAcceptanceCriteria` run (>424,000 cases), and owner's two overlay mutations (first-wins restoration and force-bypass restoration).

### Verified repository-wide consumer inventory

Every producer and caller of criterion section selection, criterion counts/lists/tallies, criterion mutations, lifecycle completion, lint/repair, and raw-body repair was verified in `/private/tmp/isolated-review.q9iXJq`:

1. **Domain selector and section boundaries (`internal/domain/criterion_section.go`, `internal/domain/body.go`):**
   - `isAcceptanceSectionHeading` (`internal/domain/criterion_section.go:22-25`) and `selectCriterionSection` (`:31-82`).
   - `scanAcceptanceCheckboxes` (`internal/domain/body.go:248-303`) calls `selectCriterionSection(lines)` at `:252`; returns `(lines, section, items, err)`.
   - Generic `Section` (`internal/domain/body.go:100-137`) and `ReplaceSection` (`:160-208`) remain unchanged substring section helpers used by CLI `--section` scoping (`narrowBody` at `internal/cli/bodyscope.go:33-53`), not criterion selection.
2. **Criterion read helpers and all callers:**
   - `CountAcceptanceCriteria(body string) (ACCount, error)` (`internal/domain/body.go:307-325`):
     - Called by `newTaskInfoCmd` (`internal/cli/task.go:336-339`), which aborts before `render.TaskInfoJSON` / `render.TaskInfoHuman` (`:350-353`) when `err != nil`. Wire converters `wire.ToTaskInfoEnvelope` / `wire.ToTaskInfoJSON` (`internal/wire/wire.go:932-960`) receive `ac` only on success.
   - `ListAcceptanceCriteria(body string) ([]AcceptanceCriterion, error)` (`internal/domain/body.go:533-552`):
     - Called by `UnexplainedCriteria` (`internal/domain/body.go:487-498`), `TallyCriteria` (`:512-527`), `lintCriterionStates` (`:741-754`), `Service.AcceptanceCriteria` (`internal/core/service_task.go:392-398` -> CLI `newTaskAcCmd` read path at `internal/cli/task.go:510-519`), and the archived-task lint branch in `Service.Lint` (`internal/core/service.go:840-845`).
   - `TallyCriteria(body string) (CriterionTally, error)` (`internal/domain/body.go:512-527`):
     - Called by TUI `criterionRollup(body string)` (`internal/tui/detail.go:824-856`), rendered into the `Criteria` metadata row in `renderTaskMeta` (`:806-808`). On error, `criterionRollup` returns `errorStyle.Render("unavailable: " + err.Error())` (`:826-828`) while `renderTaskDetail` (`:682-700`) still renders the task body below and `taskDetailContent.rawBody()` (`:184-189`) preserves raw body access for `e` (`editSelected` in `internal/tui/commands.go:467-513`).
   - `UnexplainedCriteria(body string) ([]AcceptanceCriterion, error)` (`internal/domain/body.go:487-498`):
     - Called by `validateExistingLifecyclePlan` (`internal/core/task_lifecycle.go:329-336`), reached from `ValidateTaskLifecyclePlan` (`:244-273`) and `Service.validateMove` (`internal/core/service_task.go:636-652`).
     - Crucially, `validateExistingLifecyclePlan` checks `UnexplainedCriteria(currentBody)` at `:329-331` **before** `plan.Override != TaskLifecycleOverrideAcceptanceCriteria` (`:332`) and **before** the same-status no-op return `if task.Status == plan.To` (`:338-340`).
     - Reached by `FS.MutateTaskLifecycle` (`internal/store/lifecyclemutation.go:72`), `Service.Move` / `Service.MoveSelectedTask` (`internal/core/service_task.go:580-634`), CLI `task complete` (`internal/cli/task.go:257-291`), CLI `task move` (`internal/cli/task.go:196-255`), and TUI `applyMove` (`internal/tui/commands.go:292-343`).
3. **Criterion mutations and fresh-body CAS retry paths:**
   - Domain transforms in `internal/domain/body.go`: `AddCriterion` (`:387-412`), `RemoveCriterion` (`:416-439`), `ReplaceCriterionText` (`:444-472`), `SetAcceptanceCriterion` (`:564-566`), `SetCriterionState` (`:581-603`), and `SetCriteriaState` (`:682-707`). Every one calls `scanAcceptanceCheckboxes(body)` first and returns `("", err)` on structural error before any index or no-op check (`:596-598`, `:699-706`).
   - Core service wrappers in `internal/core/service_task.go`: `SetAcceptanceCriterion` (`:401-409`), `SetCriterionState` (`:413-421`), `SetCriteriaState` (`:425-433`), `SetCriteriaTracked` (`:438-464`), and `EditCriteria` (`:469-494`). Each binds source identity once via `s.selectTaskMutationTarget(ref)` (`:504-514`) and runs `s.transformSelectedTaskBody(target, now, dryRun, ...)` inside `s.retryMutation` (`:516-532`).
   - Store persistence in `internal/store/body.go`: `FS.TransformTaskBody` (`:252-298`) acquires `s.writeLock()`, resolves the bound target via `s.resolveTaskMutationTarget(target)` (`:263`), reads fresh disk bytes via `os.ReadFile(loc.path)` (`:272`), invokes `transform(doc.Body)` (`:279`), verifies CAS (`s.verifyUnchanged` at `:290`), and writes atomically (`writeFileAtomic` at `:293`).
4. **Lint, repair, and raw-body escape hatches:**
   - `LintAcceptanceCriteria(body string) []string` (`internal/domain/body.go:715-739`) calls `selectCriterionSection(lines)` at `:717` and returns `[]string{err.Error()}` on structural error.
   - `Service.Lint` (`internal/core/service.go:838-878`) runs `LintAcceptanceCriteria(tb.Body)` on active tasks (`:875`) and runs `ListAcceptanceCriteria(tb.Body)` on archived tasks (`:840-845`, emitting `Field: "acceptance"` without reviving field/state nags on archived tasks).
   - `Service.RepairPlanning` (`internal/core/planning_maintenance.go:46-96`): for `dryRun == true`, calls `fixer.FixAll(true)` (`:89`, which only normalizes frontmatter/filenames in `internal/store/fix.go:24-95` and never rewrites Markdown headings) and returns without running post-fix lint; for `dryRun == false`, calls `fixer.FixAll(false)` and then `s.Lint()` (`:93`), failing with `ErrValidation` (exit 11) when residual structural issues remain.
   - Raw-body read and repair paths remain unblocked by criterion structure: `Service.ShowTask` (`internal/core/service_task.go:100-106` -> `task show --raw`), `Service.ReplaceBody` / `AppendBody` (`:561-577` -> `task set --body` / `task body`), and `Service.EditTask` (`:553-558` -> `task edit` / TUI `e`).

### Independent input/expected-output matrix and CLI/TUI probes

All probes below were derived independently using literal document bytes and hand-calculated expectations (without calling `selectCriterionSection`, `criterionHeadingRe`, or domain test wrappers) and executed against `/private/tmp/isolated-review.q9iXJq/bin/tskflwctl`:

| Case | Literal body arrangement (abbreviated) | Independent expectation | Observed CLI / core / TUI result |
| --- | --- | --- | --- |
| **M-A1 (Accepted hostile nesting + fences + CRLF)** | `# Task alpha\r\n\r\n### Acceptance criteria (notes)\r\n~~~md\r\n## Acceptance\r\n- [ ] fenced decoy 1\r\n~~~\r\n\r\n## Implementation\r\n\r\n### aCcEpTaNcE CrItErIa (Phase 2: API/CLI) ###\r\n\r\n#### Sub-heading inside canonical\r\n- [ ] first real criterion\r\n` + "```markdown\r\n### Other heading\r\n- [x] fenced decoy 2\r\n```" + `\r\n- [x] second real criterion\r\n\r\n### Sibling non-acceptance subsection\r\n- [ ] sibling checklist outside canonical\r\n\r\n## Post-acceptance notes\r\n- [ ] trailing narrative checkbox\r\n` | Rejected as duplicate because `### Acceptance criteria (notes)` is also a recognized parenthetical `Acceptance criteria` heading at line 3 alongside line 11 (`2 acceptance-criteria sections at lines 3, 11`). | Confirmed: `task info --json` exited `11` with `2 acceptance-criteria sections at lines 3, 11`. |
| **M-A2 (Accepted hostile nesting + fences + CRLF after changing line 3 to narrative `### Acceptance criteria notes`)** | Same as M-A1 with line 3 `### Acceptance criteria notes\r\n- [ ] narrative decoy\r\n` | Valid; selects `### aCcEpTaNcE CrItErIa (Phase 2: API/CLI) ###` (lines 11..22, closed by sibling `### Sibling non-acceptance subsection` at line 23); criteria = `[#1 unchecked "first real criterion", #2 checked "second real criterion"]`; `checked:1, total:2`. | Confirmed: `task info --json` returned `{"checked":1,"total":2}`; `task ac --json` returned exact 2 items; `lint --json` returned `"issue_count":0`. |
| **M-A3 (Accepted CLI edit + full byte comparison)** | Ran `task ac alpha --check 1 --replace 2 --text "second real criterion updated"` on M-A2 | Only line 14 (`- [ ] first real criterion` -> `- [x] first real criterion`) and line 20 (`- [x] second real criterion` -> `- [x] second real criterion updated`) change; narrative decoy, fenced decoys, sibling checklist, and trailing narrative checkbox remain byte-identical. | Confirmed via `cmp -s` against independently constructed expected file (exit `0`, byte-for-byte match). `task complete alpha` then succeeded (`exit 0`). |
| **M-R1 (Refused nonstandard + byte/mode preservation + explicit repair)** | `# Task beta\n\n## Acceptance checklist\n- [ ] hidden item one\n- [x] hidden item two\n` (`chmod 0600`) | Refused (`exit 11`) blaming `"Acceptance checklist" at line 3` across `task info --json`, `task ac --json`, `lint --json`, `lint --fix --json`, `task ac --check 1 [--dry-run]`, `task complete [--force] [--dry-run]`, `task move beta completed --force [--dry-run]`; file SHA-256 and mode `0600` unchanged. | Confirmed: all 10 commands exited `11` with `"Acceptance checklist" at line 3`; SHA-256 `407eb849...` and mode `600` preserved; renaming heading to `## Acceptance criteria` via `task set --body` restored `{"checked":1,"total":2}` and allowed `task complete beta --force` (`exit 0`, mode `600` preserved). |
| **M-R2 (Refused duplicate canonical + alias + explicit repair)** | `# Task beta\n\n## Acceptance\n- [x] legacy item\n\n## Notes\n### Acceptance criteria (v2)\n- [ ] canonical item\n` (`chmod 0600`) | Refused (`exit 11`) blaming `2 acceptance-criteria sections at lines 3, 7`; bytes and `0600` mode preserved; merging into one `## Acceptance criteria` section restores `{"checked":1,"total":2}`. | Confirmed: exited `11` across `task info`, `task ac`, `lint`, `task complete --force`; SHA-256 `15de98eb...` and mode `600` preserved; explicit merge restored `{"checked":1,"total":2}` and allowed completion. |
| **M-R3 / M-R4 / M-R5 (Hostile H1 & ancestor probes -> Finding M1)** | `# Improve acceptance test runner` + `## Notes\n- [ ] ordinary note`; `# Normal title` + `## Integrated acceptance notes` + `### Notes\n- [ ] ordinary note`; `task new "Acceptance"` | Should be valid `0/0` (or valid empty scaffold on `task new "Acceptance"`). | **Failed (Finding M1):** exited `11` with false-positive/misattributed diagnostics. |
| **TUI diagnostic & constrained viewports** | Invalid body `## Acceptance checklist\n- [ ] hidden item one\n` loaded in TUI at `160x40`, `90x20` (minimum two-pane split, `m.twoPane = m.width >= 90` in `internal/tui/view.go:39`, 54-col detail pane), and `72x20` (single-pane detail focused via `Enter`) | `criterionRollup` shows `unavailable: invalid task body: unrecognized acceptance checklist heading "Acceptance checklist" at line 1...`, never `0/0`, wraps diagnostic cleanly in narrow viewports, and preserves `rawBody()` + body rendering below metadata. | Confirmed at `160x40`, `90x20`, and `72x20`: `0/0` absent, `unavailable:` and wrapped `rename it to Acceptance criteria` + `## Acceptance checklist` body visible, `m.detail.content.rawBody() == body`. |

### Mutation testing ledger

- **Target invariant:** Section-closing boundary in `selectCriterionSection` (`internal/domain/criterion_section.go:41`: a same-level sibling heading closes an open criterion section so checklists in subsequent sibling sections are not absorbed into the criterion section).
- **Named regressions before execution:**
  - `TestCriterionSectionSelection/canonical_after_narrative` (`internal/domain/criterion_section_test.go:14`)
  - `TestCriterionSectionSelection/nested_headings` (`internal/domain/criterion_section_test.go:17`)
  - `TestCriterionSectionSelection/empty_canonical` (`internal/domain/criterion_section_test.go:20`)
  - `TestCriterionCanonicalCLISelectionDoesNotRewriteNarrative` (`internal/cli/task_criterion_section_test.go:80`)
- **Compiler-valid mutation applied to `internal/domain/criterion_section.go:41`:**
  ```go
  // Baseline:
  for len(open) > 0 && spans[open[len(open)-1]].level >= level {
  // Mutated:
  for len(open) > 0 && spans[open[len(open)-1]].level > level {
  ```
- **Observed failure (`go test ./internal/domain ./internal/cli -run '^TestCriterion' -count=1`):**
  ```text
  --- FAIL: TestCriterionSectionSelection (0.00s)
      --- FAIL: TestCriterionSectionSelection/canonical_after_narrative (0.00s)
          criterion_section_test.go:50: wrong selection: count={Checked:0 Total:2} criteria=[{Index:1 Text:actual Checked:false State:unmet Reason: Target:} {Index:2 Text:also narrative Checked:false State:unmet Reason: Target:}] lint=[]
      --- FAIL: TestCriterionSectionSelection/nested_headings (0.00s)
          criterion_section_test.go:50: wrong selection: count={Checked:0 Total:2} criteria=[{Index:1 Text:actual Checked:false State:unmet Reason: Target:} {Index:2 Text:outside Checked:false State:unmet Reason: Target:}] lint=[]
      --- FAIL: TestCriterionSectionSelection/empty_canonical (0.00s)
          criterion_section_test.go:50: wrong selection: count={Checked:0 Total:1} criteria=[{Index:1 Text:outside Checked:false State:unmet Reason: Target:}] lint=[]
  FAIL	github.com/andy-esch/taskflow/internal/domain	0.264s
  --- FAIL: TestCriterionCanonicalCLISelectionDoesNotRewriteNarrative (0.06s)
      task_criterion_section_test.go:86: narrative shadowed canonical criteria: {"schema_version":1,"revision":"1.90",...,"acceptance":{"checked":0,"total":2}}
  FAIL	github.com/andy-esch/taskflow/internal/cli	0.415s
  ```
- **Restoration:** Restored `>= level` in `internal/domain/criterion_section.go:41` and re-ran `go test -race ./internal/domain ./internal/core ./internal/cli ./internal/tui -run '^TestCriterion' -count=1` -> all 4 packages PASS.

### Required hostile angles and second systemic pass

1. **A narrative wins indirectly:** **Reproduced (`M1`).** While a narrative section before or after a canonical section cannot shadow the canonical section (`len(recognized) == 1` wins at `internal/domain/criterion_section.go:72-73`), when a task has *no* canonical section, an ancestor heading (including the document's `# <Title>` H1 or a narrative `## Integrated acceptance notes`) remains on `open` across deeper non-acceptance child headings (`## Notes` / `### Notes`) and falsely claims ordinary child checklists as unrecognized acceptance checklists (`M1`). Furthermore, when a task's H1 title is `# Acceptance` or `# Acceptance criteria`, `isAcceptanceSectionHeading` treats the H1 itself as a recognized section.
2. **The language description overpromises:** **Partially reproduced via `M1`; otherwise falsified.** Closing ATX hashes (`bodyATXClosingRe`), case-insensitivity, parenthetical suffixes (`(?i)^acceptance criteria(?:[ \t]+\([^\r\n]*\))?$`), tilde/backtick fences (`fenceScanner`), and `\r\n` normalization (`splitBodyLines`) behave consistently across selector, edits, lint, CLI help (`docs/cli/tskflwctl_task_ac.md`, `docs/cli/tskflwctl_task_complete.md`), and `schema task` conventions (`internal/domain/entity.go:84-93`). The only mismatch between documented conventions and implementation is `M1`: `entity.go:71-72` promises that any task title is kept as the body H1, whereas `selectCriterionSection` scans the H1 as a criterion section heading.
3. **All tests share the same mistaken oracle:** **Reproduced (`L1`).** All shipped unit/integration fixtures omit the `# <Title>` H1 line that real tasks have, and `TestCriterionSectionSelection` checks only `count.Total` (where decoys and canonical sections both have `1` item) and error substrings rather than selected criterion text or diagnostic line/title.
4. **Compatibility labels hide semantic drift:** **Falsified by verification.** `internal/wire/wire.go:25, 199` declares `CurrentRevision = "1.90"` with classification `"NOT ADDITIVE"` in `docs/architecture/planning-data.md` and `planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md`. Actual JSON producers (`task info --json`, `task ac --json`, `lint --json`, `schema task --json`) were exercised in the sandbox and emit revision `"1.90"` and exit `11` (`"code":"validation_failed"`) on invalid criterion structure rather than `0/0`.
5. **Presentation launders the failure:** **Falsified by verification.** `criterionRollup` (`internal/tui/detail.go:824-828`) surfaces `unavailable: <err>` and never `0/0`; tested at `160x40`, `90x20`, and `72x20`, confirming the error message wraps cleanly and `rawBody()` remains intact for editor repair.
6. **Second systemic pass across shared helpers (`fenceScanner`, `Section`, `ReplaceSection`, `FirstH1`, `RenameTask`, `TransformTaskBody`, `MutateTaskLifecycle`):**
   - Verified that `Section` / `ReplaceSection` (`internal/domain/body.go:100-208`) and `narrowBody` (`internal/cli/bodyscope.go:33-53`) were not altered and remain raw-body section helpers.
   - Verified that `TransformTaskBody` (`internal/store/body.go:252-298`) and `MutateTaskLifecycle` (`internal/store/lifecyclemutation.go:20-140`) always pass fresh disk-loaded body bytes into the domain callback under `writeLock()` and verify CAS before atomic rename.
   - The systemic pass directly uncovered the cross-helper interaction in `M1` between `FirstH1` / `taskBodyTemplate` / `RenameTask` (which own the level-1 `# <Title>` H1) and `selectCriterionSection` (which scans level-1 headings as candidate acceptance sections).
