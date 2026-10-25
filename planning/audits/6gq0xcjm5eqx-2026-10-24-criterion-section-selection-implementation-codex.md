---
schema: 1
id: 6gq0xcjm5eqx
bucket: closed
area: criterion-section-selection-implementation-codex
date: "2026-10-24"
---
# Audit: Criterion section selection and completion guards — codex — 2026-10-24

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

No new in-scope product findings from either review pass. No findings were settled,
and no task/Thread state was changed. This is a bounded technical result, with the
protocol qualification and unexecuted checks recorded below, not an unrestricted
readiness claim.

### Snapshot and execution boundaries

- Independent workspace: `/private/tmp/isolated-review.Nfb7Eg`.
- Captured baseline: `5a9747fb4e71bfe1dbcc657d66e837441de9a638`; parent/source
  HEAD: `0d13d616584e9b4fe7949fe501a5709bfc281f4d`. Reviewed the captured dirty-tree
  implementation, including the new selector and tests, rather than only HEAD.
- Evidence/scripts/logs remain outside the workspace at
  `/private/tmp/criterion-codex-evidence.NgRTdH`. The temporary real-adapter Go test
  was copied there and removed from the workspace; the production mutation was
  restored from the helper baseline. No agents were launched or sibling report read.
- Protocol qualification: the initial source command also ran `pwd` and a
  filename-only `rg --files` inventory before sandbox creation. That exceeded the
  brief's initial-operation restriction. It did not inspect implementation contents,
  run tests/generators, or change source files. All implementation inspection,
  builds, tests, probes, restoration, and report editing followed successful helper
  creation in the independent sandbox. Only the source-owned helper performs transfer.
- Initial build/test commands used the existing module download cache read-only;
  Go printed one denied module stat-cache write during the first successful build.
  Final build and focused/wire reruns used reviewer-owned `GOCACHE`, `GOTMPDIR`, and
  `GOMODCACHE`, populated through the existing local file proxy with `GOSUMDB=off`.
  No external network lookup was needed.

### Pass 1 — independent behavior and authority checks

Executed fresh `just build`, then the required checks (all passed):

```text
go test -race ./internal/domain ./internal/core ./internal/cli ./internal/tui -run '^TestCriterion' -count=1
go test ./internal/wire
just docs-check
```

The final focused race rerun passed all four packages after restoration; wire
passed again. `just docs-check` regenerated the reference and found no drift.
Full `just test`, `just lint`, and the ten-second fuzz campaign are owner-reported
evidence in the implementing task, not independently executed here. No goldens
were regenerated. The focused tests actually present and executed include
`TestCriterionSectionSelection` (domain/criterion_section_test.go:9),
`TestCriterionStructureRefusesEveryEdit` (:56),
`TestCriterionRetryRevalidatesFreshBodyStructure` (core/criterion_section_test.go:45),
`TestCriterionSectionPortableReadsLintAndCompletion` (:64),
`TestCriterionStructureCLIRefusesBeforeEffects` (cli/task_criterion_section_test.go:13),
`TestCriterionCanonicalCLISelectionDoesNotRewriteNarrative` (:80), and
`TestCriterionStructureIsVisibleWithoutHidingTaskBody` (tui/criterion_section_test.go:13).
Paths in this paragraph are under `internal/`.

**Independent real CLI oracle.** `cli_probe.py` creates its own planning space,
epic and tasks using the fresh binary. Its accepted literal body is:

````markdown
# Independent hostile source

## Acceptance notes
- [ ] narrative before

## Work
### aCcEpTaNcE cRiTeRiA (review) ###
- [x] shipped
#### Edge behavior
- [ ] must block
```md
## Acceptance
- [ ] fake
```
### Status of acceptance
- [ ] narrative after
````

The outer code block above depicts source containing a fenced example. The
independently expected criterion texts were exactly `shipped`, `must block`,
in that order, with tally 1/2. Real `task ac` and `task info --json` agreed.
Unforced real/preview completion exited 11; forced preview exited 0 with the
whole fixture tree unchanged. Checking criterion 2 produced exactly the literal
body with only `- [ ] must block` replaced by `- [x] must block`, verified against
`task show --raw --json`; completion then succeeded. No expected result used the
production selector, its regex, or the domain test wrappers.

**Refusal classes.** Literal invalid bodies were:

```markdown
# Refused
## Acceptance
- [x] decided narrative
## Acceptance criteria
- [ ] must block
```

and:

```markdown
# Refused
## Acceptance evidence log
- [ ] must block
```

For each body, both an active task and an already-completed task were exercised.
Every combination of `task complete` / `task move <id> completed`, force on/off,
and real/preview execution exited 11. Actual guarded check, add, remove, replace
and defer operations also exited 11, including previews and the first body's
already-checked criterion 1. Snapshots independently compared exact file bytes,
permissions (including a task chmodded to 0640), and directory entries before and
after every refusal/preview. No effects were observed.

`task info` and `task ac` emitted no success payload on invalid input. Their real
process stderr contained revision `1.90`, error code `validation`, and actionable
acceptance diagnostics, with exit 11. Lifecycle stdout contained per-item errors;
lint stdout contained acceptance issues and stderr its validation summary.
Ordinary lint and actual `lint --fix` diagnosed both classes without changing
bytes/modes; `lint --fix --dry-run` exited 0 and preserved them, consistent with
its proposal-only contract. Raw body reads remained available.

**Repair advice executed.** Unknown-heading cases were explicitly renamed to
`Acceptance criteria` through `task set --body`; the same read/edit/completion
route recovered. The separate `repair_archived_probe.py` explicitly merged both
recognized sections into one canonical section while retaining both original
criterion texts. It read both criteria, still refused unforced completion while
the second was unmet, then accepted checking criterion 2 and completion. No
implicit migration was needed. A real `task edit` run in a PTY with a small editor
script also renamed an unknown heading; subsequent criterion read, check and
completed status/tally were verified. The initial noninteractive `task edit`
probe correctly refused; it was rerun with the required terminal.

**Fresh-body challenge beyond the portable fake.** The reviewer-only
`TestReviewerCriterionActualFSConflict` used the actual `store.FS` and a second
application `ReplaceBody` producer inside the existing pre-write test hook.
Between the first transform and CAS, that producer replaced a valid section with
a body containing `Acceptance criteria` plus `Acceptance`. The stale CAS failed;
the retry refused the newly ambiguous body. Checkbox, multi-state, add and tracked
operations each took two transform attempts, retained the exact same selected
source ID and operation time (clock called once), and left precisely the competing
producer's bytes/mode. Executed with the race detector alongside the shipped
`TestAcceptanceCriterionTransformRetriesWithoutLosingConcurrentBodyWrite`
(`internal/store/body_test.go:220`) and the portable retry regression; all passed.
This is a deterministic real-adapter interleaving, not a claim of exhaustive
concurrent filesystem scheduling or a second OS process.

**Compiler-valid mutation.** Announced intended regression before execution:
`TestCriterionRetryRevalidatesFreshBodyStructure`. In
`internal/core/service_task.go:488`, wrapped `transform` with a memoizer:

```go
original := transform
var cached string
var prepared bool
transform = func(body string) (string, error) {
    if prepared { return cached, nil }
    out, err := original(body)
    if err == nil { cached, prepared = out, true }
    return out, err
}
```

Expected: reusing the first successful projection on retry bypasses validation of
the second body, and the intended test must fail. Observed: the package compiled,
the intended regression failed with `err=<nil>`, two identical targets, and
`commits=1` (`mutation.log`). Restored the production file from the captured
baseline; the intended test and the real-FS probes passed under `-race`
(`fs-conflict-restored.log`), followed by the full required focused rerun.
This attacks callback freshness rather than repeating the owner's force-bypass
or first-wins mutations. `mutation.patch` records the exact edit.

### Verified repository-wide consumer inventory

Repository-wide Go searches excluded test files, then a broader criterion/section
search checked adjacent consumers (`consumer-inventory.txt`,
`completion-inventory.txt`, `systemic-search.txt`). The inventory is of existing
code, not planned capabilities:

| Consumer/operation | Verified route and error behavior |
| --- | --- |
| Count | `internal/domain/body.go:307` uses `scanAcceptanceCheckboxes` (:250); sole external caller `internal/cli/task.go:336` returns the error before path resolution/rendering. |
| List | `body.go:533`; external reads are `internal/core/service_task.go:392` and archived lint `internal/core/service.go:840`. Internal list consumers are tally (:513), unexplained (:488), and state lint (:742); each handles errors. |
| Tally | `body.go:512`; sole external caller `internal/tui/detail.go:824` renders unavailable plus the sanitized diagnostic, not zero. The shipped 160×40 Model fixture retains the raw body; no live TUI claim is made. |
| Unexplained/completion | `body.go:487` → `internal/core/task_lifecycle.go:329`; structural errors return before the same-status return (:335) and unmet-criteria override (:355). |
| Every criterion transform | Domain add/remove/replace (`body.go:387/:416/:444`), state/multi (:564/:581), and checkbox (:682) use the scanner before targeting/no-op. CLI `criteriaEdit` (`task.go:845`) selects add/remove/replace; CLI state/tracked routes (:545/:547) reach core callbacks (`service_task.go:470/:476`). Checkbox core callback is :403; generic criterion edits :484. All run through `TransformTaskBody` (`internal/store/body.go:252`), whose callback (:276) precedes the unchanged-body return and content CAS (:296). |
| Lifecycle entry paths | Named verbs and generic move converge at `internal/cli/task.go:919`; batch `internal/cli/moves.go:32` attempts each through this route. Core `Move` (:588) and `MoveSelectedTask` (:602) in service_task.go reach `runTaskLifecycleMutation` (:650), including selected-plan retry (:667). TUI `internal/tui/entity.go:364` uses selected-ID movement. Actual FS validates via `internal/store/lifecyclemutation.go:72`; create-and-start is separately restricted to in-progress (`internal/core/task_lifecycle.go:260`). |
| Lint and repair | Active/archived body checks are `internal/core/service.go:832/:840`; `LintWithLinks` (`internal/core/planning_maintenance.go:91`) starts with ordinary lint. Real `RepairPlanning` (:46) re-lints (:84); dry-run returns earlier (:82). CLI `internal/cli/lint.go:112` explicitly skips post-repair lint for preview. Actual and preview commands did not rewrite headings. |
| Raw repair | `ShowTask`/raw show remains independent of criterion enumeration; `ReplaceBody`, `AppendBody`, and editor routes (`service_task.go:561/:572/:553`) remain available. Editor accepts same-status body repair and rejects lifecycle/dependency edits (`internal/store/edit.go:175/:181`). General `Section` substring behavior (`body.go:216`) is unchanged and independently pinned by domain/criterion_section_test.go:86. |

Machine propagation was checked through real subprocesses as well as
`internal/cli/exit.go:96/:137` (classification/JSON stderr).
`internal/wire/wire.go:383/:390/:405` declares 1.90 and not-additive, matching the
observable refusal/tally change, authoring convention in
`internal/domain/entity.go:72`, reference help and
`docs/architecture/planning-data.md:119`. The revision/task-schema goldens exist
at `internal/cli/testdata/golden/machine_contract_revision.txt:1` and
`schema_task_json.golden:1`. `go test ./internal/wire` executed semantic/schema
validators; the wire diff adds no optional DTO fields, so there is no new optional
branch requiring a non-default fixture. Full CLI golden testing was not rerun
beyond the required focused CLI tests.

### Pass 2 — systemic falsification and bounded surviving hypotheses

Re-reviewed section ownership and every adjacent consumer after completing the
first-pass checklist; checked that a shared selector cannot make all expected
values agree incorrectly. `systemic_probe.py` used literal documents and explicit
expected lists, independent of the production regex/helpers:

| Hostile arrangement | Independent expected criterion texts | Observed |
| --- | --- | --- |
| `Acceptance journal` parent containing narrative checklist, then child `Acceptance criteria (RC)`, followed by a higher-level heading | `actual` only | Match |
| Narrative acceptance checklist followed by empty canonical section and same-level Notes checklist | Empty | Match |
| Four-backtick fenced example containing a three-backtick line and alias heading, then real canonical section | `actual` only | Match |
| CRLF `ACCEPTANCE ##`, nested Edge heading, then same-level Notes | `first`, `second` | Match |
| Ordinary Notes checklist with no acceptance heading | Empty | Match |

A two-item `task complete ... --force` batch refused both independently invalid
bodies with per-item diagnostic errors and exact unchanged bytes/modes. The
separate archived probe removed description/tags/tier/autonomy through the CLI
and added a checked/deferred state contradiction: plain lint remained clean.
Adding a second recognized section produced exactly one acceptance issue, with
no resurrected active-field or criterion-state nags (`repair-archived.log`).

- **Falsified in executed scope:** narrative shadowing via parent/child ownership,
  fake criteria in differently sized fences, success-shaped empty reads for either
  invalid class, force/no-op/preview/batch completion bypass, idempotent check
  bypass, stale callback reuse on a real adapter retry, unusable explicit heading
  repair, and automatic heading rewrite by repair.
- **Bounded, not proved universally:** arbitrary OS-process races, all Markdown
  dialects, all terminal sizes/live TUI ordering, every retry schedule, and all
  golden payloads. The existing line-oriented ATX/fence language was reviewed;
  this does not assert universal Markdown support. No invented lock path or new
  persistence/capability requirement is part of this result.
- **Tracked exclusions remain open elsewhere:** canonical malformed-marker
  completion is task `6gq03e44pcyz` (its file:22 records the reproduction), and
  tracked-reason semantics are task `6gkkzvr92waq` (its file:26 records the scope).
  Their files and open work were verified; neither is presented as shipped repair
  or relabeled as a new finding.
- Probe harness assumptions about lint envelope keys and the editor terminal gate
  were corrected against actual outputs before claiming their results. These were
  probe mistakes, not product findings.

### Final validation and isolation attestation

Executed after removing/restoring probes with the fresh CLI:

```text
./bin/tskflwctl -C . lint --links
✔ all planning entities and dependency links pass lint
./bin/tskflwctl -C . audit lint 6gq0xcjm5eqx
✔ all audit findings pass lint
git diff --check
(exit 0, no output)
```

Only this assigned audit differs from the captured baseline. The helper output
below is copied verbatim; `transfer=pending` is intentional because these report
bytes must be final before delivery. The transfer receipt belongs in the handoff.

```text
sandbox_path=/private/tmp/isolated-review.Nfb7Eg
git_dir=/private/tmp/isolated-review.Nfb7Eg/.git
baseline_commit=5a9747fb4e71bfe1dbcc657d66e837441de9a638
source_blob=a0637403221ff7fe4552ca869440d9408369542d
source_fingerprint=798dd4c06055582289c6c3c6b558a596a49c24f6
deliverable=planning/audits/6gq0xcjm5eqx-2026-10-24-criterion-section-selection-implementation-codex.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```
