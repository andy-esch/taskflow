---
schema: 1
id: 6gk3mt5hz2vj
bucket: closed
area: focused-architecture-guides-implementation-claude
date: "2026-10-12"
updated_at: "2026-10-12"
---
# Audit: Focused architecture guides and code-documentation navigation — claude — 2026-10-12

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Review the extraction of the monolithic architecture reference into a short index and six focused
guides with two-way package-comment routes and a bounded navigation test. The central question:
did condensation silently lose or strengthen a safety/compatibility contract, or build a misleading
maintenance guard? Finding count and word-count reduction are not success metrics.

Claude's primary lens is semantic preservation and code/guide agreement. Antigravity's primary lens
is hostile navigation/test maintenance, plus one source-authority or composition cross-check.
Both independently inspect the common invariants; neither reads the sibling report before finalizing.
Give a findings-first verdict, distinguishing defects from preferences and explicitly deferred work.

## Review target

Source checkout: /Users/andyeschbacher/git/andy-esch/taskflow
Branch: docs/focused-architecture-guides
Base: 4fdf6fcc516a6312e243ce1092c798d1d4797f9d

The implementation is dirty/untracked, not HEAD alone. Capture the full checkout using the required
independent sandbox protocol. Compare the old guide with git show <base>:docs/ARCHITECTURE.md in
the sandbox, and read the captured index plus every docs/architecture/*.md guide.

Implementing task:
planning/tasks/6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md
Thread:
planning/threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md

Changed routes include AGENTS.md, README.md, routines/weekly-architecture-audit.md and package
comments in internal/core/store.go, internal/store/frontmatter.go, internal/appwiring/wiring.go,
internal/config/config.go, internal/cli/root.go, internal/wire/wire.go and internal/tui/model.go.
New checks: internal/tools/architecturecheck/architecture_test.go:
TestArchitectureNavigation, TestHeadingAnchorsIgnoreFencedExamples,
TestCheckTargetDetectsBrokenRoutes.

Build a verified consumer inventory of incoming guide anchors, package comment routes, relevant
type contracts and the actual regression tests cited by the guides. Verify every named file/symbol.
Do not assume planning requirements or linked tests prove the prose true.

Decision owners: AGENTS.md's ownership matrix, accepted ADR-0003/0007/0008 and
docs/THREADS_COMPATIBILITY.md. The dependency-policy ADR is still future work.
Selective doc.go expansion, compiled examples, exact import-map enforcement and broad documentation
reconciliation remain existing sequenced tasks; the extraction only establishes routing and a
scoped link/anchor guard. The recent arch-machine-contract audit's proposals are not adopted here.

## Intended contract to challenge

- Current cross-package rules and named local exceptions remain accurate; no new behavior,
  compatibility promise or dependency policy is silently enacted.
- Type/method comments own exact API guarantees. Guides explain relationships, link to source and
  evidence, and owning package comments route back; no copied method/command/revision inventory.
- Preserve source identity versus declarations/labels, opaque locations versus local authority,
  checked source-set composition versus snapshot consistency/security, explicit policy including
  denied previews/no-ops, CAS limits with raw editors, and committed/partial receipts.
- Preserve Thread graph/rank/frontier semantics, structured audit settlement versus remaining
  generic-body guards, initial-corpus and later identity revalidation, and TUI async/session rules.
- Old architecture headings remain useful forwarding anchors; useful history survives in planning
  or an immutable reference rather than a stale feature snapshot being presented as current policy.
- Test-only navigation checks ordinary inline local links/ATX headings and linked package-comment
  routes. It is not a full Markdown parser, external-URL validator or semantic contract proof.
- No production executable logic, wire revision or emitted output is intended to change.

## Mandatory evidence floor

1. Independently run GOCACHE=<external scratch cache> go test -race ./internal/tools/architecturecheck
   -count=1, plus focused checks for any semantic claim you dispute. Full owner test/lint/build/docs
   results may be inherited if clearly labeled, not reported as independently executed.
2. Trace at least three risky statements from old guide to new guide to owning code/test. Claude:
   policy/preview authorization, durable-prefix/CAS limits, source/Location authority. Antigravity:
   choose one of those plus actual incoming anchors and package-comment navigation.
3. Execute a compiler-valid mutation that must fail TestArchitectureNavigation. Claude: remove the
   core package-comment return route while leaving the guide's source link. Antigravity: break the
   planning-data Change checklist anchor or a local owner-source link while leaving other links
   intact. Record exact edit, intended assertion, failure, restore and green result. A compile error
   is invalid, not a killed probe. Optional wire-branch/repair-command probes are not applicable
   unless you first demonstrate such a change in this slice.
4. Challenge one plausible blind spot: headings inside examples, a same-file #fragment, a removed
   guide/return route, misleading package ownership, or preservation of an old forwarding anchor.
   Identify whether it is reproduced, falsified or deliberately outside the guard's documented
   grammar/scope. Do not file unsupported Markdown syntax as a bug unless a current/reasonably
   expected supported authoring route is affected and the claim says it is checked.
5. Check go doc on at least one changed package. Local paths are repository navigation cues,
   explicitly not clickable relative Markdown links promised by go doc.

## Required hostile angles

- Would an agent following the new guide mutate using a display slug, derive a path from Location,
  retry committed work, assume read-only preview is allowed, or treat source-set equality as a
  consistent snapshot? Show the exact misleading sentence and reachable owning path if so.
- Did a named local exception become a general exemption? Are internal-import depguard checks
  confused with complete standard-library/framework bans? Is future policy marked accepted?
- Did shortening the core/store sections delete important maintenance/completion behavior,
  Thread impact attribution, unreadable/duplicate evidence, or CAS/raw-editor limitations?
- Does guide drift still get checked after harmless headings/file moves? Does the checker pass
  falsely because an example/string contains the expected link, or because another package
  accidentally supplies the return route? Assess concrete consequence and stated coverage.
- Does the handmade anchor normalizer agree with the actual checked headings? Does fence stripping
  ignore real routes or treat example headings as valid anchors? No generic parser rewrite demand.
- Distinguish obsolete implementation counts/file rosters from useful rationale. Compare old
  review history and the dated Status section with its archive and immutable baseline.
- Systemic second pass: pick one shared pattern (source navigation, cross-package contract
  ownership, return-route derivation, or test evidence labeling), inspect an adjacent consumer,
  and falsify a concrete hypothesis. Do not confuse layout preference with a safety finding.
- Pressure-test scoping: demonstrate an in-scope failure before demanding implementation of
  doc.go/examples/import-map or the unaccepted machine-schema vocabulary proposals.

## Validation and restoration

Owner baseline: just test (race-enabled), just lint, just build, just docs-check, planning lint
--links and git diff --check passed. Two owner compiler-valid comment mutations (missing anchor,
removed return route) failed the intended navigation test; restored focused race tests passed.
These are inherited owner observations, not independent corroboration.

Only inspect, run generators/tests, edit or probe in the independent sandbox. Keep caches/fixtures
outside its captured tree. Restore every probe to its captured baseline; only your assigned audit
may differ at verify/transfer. The helper's sandbox baseline commit is allowed; no other commits,
index changes or pushes. Do not work in the source checkout, use git worktrees, or read the sibling
report. On isolation/verification/transfer failure, retain the sandbox and report incomplete.

The helper currently leaves a source-path origin; pushes are forbidden. Accident-prevention
hardening is separately tracked in 6gk06e2g21hv, not claimed implemented by this change.
Preserve the helper attestation and use guarded one-file transfer, never manual copyback.

## Deliverable

Preserve the brief; fill only your assigned audit with verdict/limits, verified inventory, open
findings, concise evidence/mutation ledger, and a bounded surviving-hypothesis ledger. Consolidate
root causes and state reachable consequence, reproduction, recommendation and residual uncertainty.
Coverage gaps/preferences/defense-in-depth are not automatically high severity.
Use repository finding verbs where possible, exact canonical grammar otherwise; leave findings
open and audit open for implementation-owner triage. No placeholders count as findings.

Include sandbox path, independent Git directory, baseline commit, captured source audit hash/
fingerprint, verification and guarded transfer outcome. Report technical completion separately
from the helper's actual transfer outcome; do not invent a successful future transfer.

## Reviewer report

### Isolation attestation

- Workspace: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.WSnqDn`,
  created by `scripts/isolated-review-workspace.sh create` (independent `--no-hardlinks` clone; no
  worktree or symlink).
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.WSnqDn/.git`.
  Its `origin` is the source path; nothing was fetched or pushed.
- Baseline commit: `624df34579d99e88fadd046a093d4cc1eaa033e0` (sandbox-only capture, parent `4fdf6fcc`).
- Captured source audit blob `752714523e7a86cf4d19f5d661033171866eea54`; source fingerprint
  `d9da022674f5dbaa42b759eace9f547999971a3f`.
- Deliverable: `planning/audits/6gk3mt5hz2vj-2026-10-12-focused-architecture-guides-implementation-claude.md`.
- Pre-write `verify`: exit 0, `deliverable_changed=false`, `transfer=pending`. The final `verify`
  and guarded `transfer` ran after this text was frozen; their output is reported in the reviewer's
  handoff message, not asserted here.
- Every probe was restored with `git show HEAD:<path> > <path>` and checked with `git status`. No
  staging, commits, stash or index changes. Go cache and the built binary were kept in an external
  scratch directory.

### Verdict

Findings first: four Low findings, no High or Medium. I found no safety or compatibility contract
that the condensation silently lost or strengthened along the brief's hostile angles. The defects
are:

- L1: guide evidence and owner links that point at files that neither pin nor own the claim.
- L2: two navigation assertions that accept non-navigable mentions.
- L3: consumers of the relocated content outside the checked set that were neither re-routed nor
  guarded.
- L4: one test-pinned CAS scope (lifecycle Thread snapshot) that lost its only prose statement.

None changes production behavior, a wire revision or emitted output; the diff touches only comments
in seven Go files plus a test-only package. L1 to L3 bear on the implementing task's checked AC
"Repair repository links and contributor guidance that reference the current document layout" and
on the guard's stated coverage. I recommend fixing them before closing 6g6x7e2ef37r. L4 needs one
phrase.

Technical completion: both review passes and evidence-floor items 1 to 5 are done. For the transfer
outcome, see the attestation above.

Hostile-angle results (the exact sentences and owners are listed in the traces below):

- Display slug as identity: no. See planning-data.md:15-16 and tui.md:17-21.
- Path derived from `Location`: no. See planning-data.md:17-18 and 25-26, and tui.md:43-46. Owned by
  core/entity_read.go:23-28 and 50-54.
- Committed work retried: no. See persistence.md:54-55 and 66-68, and planning-data.md:142-144.
  Owned by core/retry.go:60-66 and core/task_lifecycle.go:92-94.
- Read-only preview assumed allowed: no. See boundaries.md:108-109. Owned by
  core/mutation_policy.go:31.
- Source-set equality treated as a snapshot: no. See planning-data.md:39-40 and
  cli-and-wire.md:28-29.
- A named exception generalized, depguard read as a complete ban, or future policy marked accepted:
  no. See boundaries.md:76-79 and 115-129, and the ADR task status.
- Unrestricted mutation policy: "not a production default" moved to its owner comment
  (mutation_policy.go:40-41). Grep finds no production `UnrestrictedMutations()` caller.

## Findings

#### L1. Guide evidence and owner links name files that neither pin nor own the cited claims · **Status:** fixed 2026-10-12

**File:** docs/architecture/planning-data.md:88 | **Component:** docs/architecture
**Effort:** XS · **Urgency:** soon

`docs/architecture/planning-data.md:82-88` makes four claims and then says `[Entity-read regressions](../../internal/core/entity_read_test.go) cover these joins`:

- LintSource and AuditSnapshotSource share one snapshot.
- An explicitly injected audit source takes precedence over the lint default "independently of option order".
- Dashboard reads use portable snapshots.
- Those reads preserve diagnostics without another scan.

That file contains only two tests: `TestOrdinaryReadServicesDoNotPublishSourceLessRecords` (`entity_read_test.go:48`) and `TestThreadReadValidateSourcesFailsClosedWithoutCanonicalIdentity` (`:77`). Neither touches LintSource, AuditSnapshotSource, option order, or Summary/board diagnostics.

The regressions that do cover these claims all pass with the race detector in the sandbox:

- `TestExplicitAuditSnapshotSourceWinsRegardlessOfOptionOrder` (`internal/core/lint_source_test.go:226`)
- `TestLintSourceSuppliesDefaultAuditSnapshotSource` (`lint_source_test.go:256`)
- `usecases_test.go:260` and `:303`
- `board_test.go:89`, `:102` and `:181`

The pre-split guide stated these claims without any test link (pre-split lines 416-446), so this misattribution is new.

`docs/architecture/configuration-and-spaces.md:44-47` has the same problem. It links the claim that userconfig "owns home preferences and the separate advisory spaces registry. Its directory is resolved from `TSKFLW_CONFIG_HOME`, XDG config home, then the documented config fallback" to `internal/userconfig/paths.go`. That file only handles registry-entry tilde and physical paths (`paths.go:9-17`). The package contract and the `Dir` resolution are in `internal/userconfig/userconfig.go:1-13` and `:82-83`.

Consequence: a contributor changing audit-source precedence, dashboard diagnostics, or the home config directory is routed to a file that neither pins nor implements the claim. They may conclude the behavior is unpinned, or edit the wrong file. The navigation test only proves that link targets exist, as its own comment says, so it cannot catch this.

Residual uncertainty: I checked every test citation in the six guides by matching test names against claims. These are the only two misroutes I found.

**Recommendation:** Cite lint_source_test.go (plus usecases_test.go/board_test.go for dashboard joins) at planning-data.md:88, and route configuration-and-spaces.md:44 to internal/userconfig/userconfig.go.

**Resolution:** Corrected the userconfig owner to userconfig.go and linked
    lint_source_test.go, usecases_test.go and board_test.go for their actual
    source-precedence, scan-reuse and diagnostic guarantees. Independently reran
    the cited focused core regressions with the race detector; they pass.
    Existing-file checks are not presented as proof of claim-to-test agreement.

#### L2. Index and back-link assertions accept non-navigable mentions inside fences or plain text · **Status:** fixed 2026-10-12

**File:** internal/tools/architecturecheck/architecture_test.go:50 | **Component:** architecturecheck
**Effort:** S · **Urgency:** eventually

`TestArchitectureNavigation` validates link targets only after fences are stripped (`internal/tools/architecturecheck/architecture_test.go:41,61`). Its two structural assertions, however, are raw `strings.Contains` checks over the whole file:

- The index must contain `"(architecture/"+guide+")"` (`:50`).
- Each guide must contain `"(../ARCHITECTURE.md)"` (`:54`).

Neither check is restricted to parsed inline links or to unfenced prose. A fenced example, inline code, or a bare parenthesized path therefore satisfies "architecture index must link directly to this guide".

Reproduced in the sandbox:

1. Replaced both live `architecture/tui.md` links in `docs/ARCHITECTURE.md` (lines 24 and 83) with plain text.
2. Appended a fenced `text` code block containing `[TUI](architecture/tui.md)`.
3. Ran `go test -race ./internal/tools/architecturecheck`. `TestArchitectureNavigation/tui.md` still passed (`ok ... 1.183s`), even though the index no longer navigated to the guide.
4. Restored from the baseline; the suite is green again.

This is the brief's hostile case of a check passing falsely because an example or string contains the expected link, and it falls within the test's own stated claim. Today every guide has a live table link, so the consequence is low, but the guard is weaker than its failure message implies.

Residual uncertainty: the back-link check at `:54` has the same shape but was not separately probed.

**Recommendation:** Derive both assertions from parsed localLink matches over withoutFences(content) and compare targets, adding a fenced-mention negative test beside TestHeadingAnchorsIgnoreFencedExamples.

**Resolution:** Both structural assertions now use parsed prose links rather
    than raw substring mentions. Fences, inline code, bare paths, images and
    escaped labels cannot satisfy them; wrapped labels remain checked. Added
    regressions and reproduced the real fenced-index and inline-code-backlink
    mutations: each fails at its intended assertion, with a restored
    race-enabled suite passing.

#### L3. Consumers of relocated content outside the checked set were neither re-routed nor guarded · **Status:** fixed 2026-10-12

**File:** .golangci.yml:207 | **Component:** docs/planning routing
**Effort:** S · **Urgency:** soon

The extraction re-routed AGENTS, README, the weekly routine, seven package comments, and three sibling tasks. Its acceptance criterion "Repair repository links and contributor guidance that reference the current document layout" is checked. It missed the following consumers outside the navigation test's document set:

1. `.golangci.yml:207` still says "docs/ARCHITECTURE.md states this as an invariant" in the `config-must-not-read-home-scope` rule comment. The index no longer states that invariant; it now lives in `docs/architecture/boundaries.md:71-72` and `configuration-and-spaces.md:54`.
2. Open task `6g63hjmb3d8s` is now out of date in three places:
   - Its body quotes the removed sentence "for five heterogeneous entities it trades clarity for machinery" (lines 32-33).
   - Its last AC targets "`docs/ARCHITECTURE.md`'s fan-out paragraph" (line 45).
   - Its scope evidence asserts that "`docs/ARCHITECTURE.md:240` still carries the declined data-driven collapse verbatim" (lines 79-80).

   That paragraph is gone. `boundaries.md:144-149` rewrote it and already calls this unstarted task "a bounded exception", which partly pre-empts the AC.
3. An AC of open task `6ffdv9g01q40` (lines 44-45) requires "docs/ARCHITECTURE.md updated if a new view/key is added". That now conflicts with `docs/architecture/tui.md:33-35` ("help sections own bindings, not a copied key inventory in this guide").
4. Nothing guards the forwarding heading `## Planning-data change checklist` (index line 55). Six planning records link to `docs/ARCHITECTURE.md#planning-data-change-checklist`:
   - Open: 6g9cz5t37dc0:25 and 6gg7e59mm68g:53, plus Thread 6g9czp7g9pt3:59.
   - Completed: 6g9cz5saenme:92 and 6ggdkzv2tnta:198, plus Thread 6gcwd78p9r04:171.

   Reproduced: I renamed that heading to `## Planning-data routing`. `TestArchitectureNavigation` stayed green, and `tskflwctl lint --links` exited 0 with byte-identical output. `lint --links` only checks that target files exist; its help says `flag any [..](path.md) whose target file is missing`. Restored afterwards. Meanwhile, the task (line 121) says old anchors "remain in the index as forwarding sections", and the Thread (line 71) says the anchor "remains a forwarding route".

Consequence: navigation and maintenance only; no safety contract changes. An agent picking up `6g63hjmb3d8s` or `6ffdv9g01q40` is sent to content that no longer exists or that now contradicts guide policy. A future index edit could also silently break the six incoming anchors.

Residual uncertainty: I did not audit closed historical records beyond the anchor sweep. Leaving those as history is reasonable.

**Recommendation:** Re-route .golangci.yml:207 to boundaries.md#enforced-fitness-rules, add handoff notes to 6g63hjmb3d8s and 6ffdv9g01q40, and either guard retained forwarding anchors via derived incoming links or track that gap explicitly in 6g9cz5t37dc0.

**Resolution:** Re-routed the depguard comment and both open-task
    criteria/handoffs to their focused owners; clarified that zero-variance
    factory work is queued, not shipped. Added a derived incoming-index-anchor
    check for task and Thread links. Renaming the retained checklist forwarder
    now fails on all six current consumers. Wider
    Markdown/routine/historical-citation coverage remains explicitly scoped to
    task 6g9cz5t37dc0, not claimed delivered here.

#### L4. Lifecycle Thread-snapshot CAS lost its only prose statement · **Status:** fixed 2026-10-12

**File:** docs/architecture/persistence.md:91 | **Component:** persistence guide
**Effort:** XS · **Urgency:** eventually

The pre-split guide (lines 569-571) said: "Task lifecycle authorization also scans and validates every Thread under the guard. Its receipt compares all before/after Thread projections, and its whole-snapshot CAS covers the Thread set so a concurrent membership edit cannot make the attribution stale."

The new table row (`docs/architecture/persistence.md:91`) keeps the attribution ("full Thread impact attribution under the same guard") but drops the Thread-set CAS. The owning port comment (`internal/core/store.go:117-122`) does not state it either. By contrast, the `ThreadMutationStore` comment (`store.go:237-239`) does: "neither can validate against a stale task/Thread snapshot".

The behavior itself is intact:

- Implemented in `internal/store/lifecyclemutation.go:109-114`.
- Pinned by `TestTaskLifecycleRejectsRawThreadRaceByWholeSnapshotCAS` (`internal/store/threadmutation_test.go:489`), which passes with the race detector in the sandbox.

"Under the same guard" alone only describes serialization among cooperating writers, and the guide itself says raw editors do not join that guard (`persistence.md:45-47`).

Consequence: low. A regression would fail the pinned test. However, a reader of the guide or the port no longer learns that lifecycle Thread attribution is safe against raw edits. That is the brief's intersection of Thread impact attribution and CAS/raw-editor limits. Every other dropped contract I sampled has a code-comment owner; see the ledger.

**Recommendation:** Add whole task+Thread snapshot and target CAS to the lifecycle table row or to the TaskLifecycleMutationStore port comment.

**Resolution:** Restored whole task+Thread snapshot and immediate target CAS in
    the lifecycle guide and its owning port comment, retaining the raw-editor
    final-window limitation. Linked threadmutation_test.go and independently
    reran TestTaskLifecycleRejectsRawThreadRaceByWholeSnapshotCAS with the race
    detector; it passes.

## Reviewer evidence

### Verified inventory

Read in full in the workspace: `docs/ARCHITECTURE.md` (97 lines), the six
`docs/architecture/*.md` guides, and `git show 4fdf6fcc:docs/ARCHITECTURE.md` (1,087 lines).

- **Local link targets:** every inline local target in the index and six guides exists. An
  independent shell sweep found 0 missing; the navigation test also passes.
- **Incoming anchors:** the only anchored incoming link to `docs/ARCHITECTURE.md` in the repository
  is `#planning-data-change-checklist`, from six planning records (L3). All ten pre-split headings
  survive as index headings with identical slugs, for example `#the-tui-internaltui` and
  `#status-2026-08-25`.
- **Non-planning consumers:** AGENTS.md, README.md and routines/weekly-architecture-audit.md were
  re-routed. The routine's change of `internal/cli/app.go` to `root.go` is a valid repair, because
  `app.go` does not exist. `.golangci.yml:207` is stale (L3).
- **Package-comment routes (seven, all in file doc comments of guide-linked files):**
  - core/store.go:8 → planning-data.md#change-checklist
  - store/frontmatter.go:3 → persistence.md
  - appwiring/wiring.go:5 → boundaries.md#composition-and-local-exceptions
  - config/config.go:2 → configuration-and-spaces.md
  - cli/root.go:4 → cli-and-wire.md
  - wire/wire.go:14 → cli-and-wire.md
  - tui/model.go:6 → tui.md

  Each guide has exactly one return route except cli-and-wire, which has two (cli and wire).
- **Named symbols, each verified by its declaration:**
  - core/entity_read.go: `RecordSource` :29 (field `LocationIsPath` :32), `LoadedRecord` :37,
    `VersionedRecord` :56 (`LocalPath` :59), `LoadProblem` :67.
  - Source sets and service: `SourceSetID` core/source_set.go:15; `NewService` service.go:321;
    `TaskGraphSource` service_task.go:148.
  - Task graph: `TaskGraph` dependency_graph.go:347; `SourceDeclarations` dependency_source.go:112;
    `CanonicalDependencies` and `Prerequisites` dependency_graph.go:1076 and :1095.
  - Lint, audit and maintenance: `LintSource` lint_source.go:16; `AuditSnapshotSource`
    audit_source.go:20; `RepairPlanning` and `LintWithLinks` planning_maintenance.go:23 and :53;
    `CompleteEntities` completion.go:45.
  - Thread projections: `ThreadView` thread_projection.go:65; `ThreadGraphProjection`
    thread_graph.go:42.
  - Ports in core/store.go: the six mutation ports at :96, :108, :120, :233, :240 and :248;
    `RenameTask` :50; `Layout` :403.
  - Opening and configuration: `NewPlanningStore` workspacestore/planning.go:17; `DescribeCheckout`
    config/worktree.go:134; the `ConfigFile` constant, used at config.go:125.
  - Space services: `SpaceRegistryService` space_registry.go:24; `SpaceOverviewService`
    space_overview.go:105; `WorkspaceService` workspace.go:67.
  - Other: `FileProblem` domain/problem.go:5; `TestModel_ViewFitsTerminal`
    tui/model_test.go:347.
- **Cited test files:** all exist, and their test names were matched against the claims citing them.
  Only entity_read_test.go misroutes (L1). Supporting anchors:
  - source_set_test.go:132 and :158 (typed nil, detach)
  - frontmatter_fidelity_test.go:81, :103 and :222 (refusal)
  - graphrepair_test.go:262 (durable prefix)
  - threadapply_test.go:411-638 (raw CAS and prefix)
  - mutation_authorization_test.go:151, :184 and :241
  - occ_retry_test.go:119 (dry run is not retried)
  - appwiring/thread_apply_test.go:21-218
  - appwiring/persistence_policy_test.go:17 and :58
  - workspacestore/planning_architecture_test.go:14
  - cli/command_safety_test.go:19 and :71
  - cli/golden_test.go:179 and :270
  - tui/watch_test.go:52-397 (real watcher)
- **Planning-status claims:**

  | Task | Subject | Status |
  | --- | --- | --- |
  | 6gg7e59mm68g | Dependency-policy ADR | ready-to-start; no ADR file (0001 to 0008 only) |
  | 6ggxymjbf86r | Direct-I/O ban | ready-to-start |
  | 6g63hjm7cp6w | Import map | ready-to-start |
  | 6g77rn6hvmh8 | Body-write gaps | next-up |
  | 6ghht05bzxkk | Settlement | completed |
  | 6gg7e594gcms | Apply parity | completed |

  The guides describe each accurately. In particular, apply parity replaces the old guide's stale
  "gap" claim with shipped evidence.
- **Not shipped and not claimed:** doc.go overviews, compiled examples, the import-map check and
  the proposals of audit 6gjzzvdy5jnw (cli-and-wire.md:52-55).

### Old to new to owner traces

1. **Policy and preview authorization.**
   - Old guide 81-83 and 159-168 → boundaries.md:106-113 and planning-data.md:140-141.
   - Owner core/mutation_policy.go:
     - :9-12: every mutation, previews included, before effects or callbacks.
     - :31: read-only denies even dry runs.
     - :34: guarded checks run at each mutation, never at construction.
     - :40-42: unrestricted is for trusted callers and fixtures; production uses guards.
     - :58-59: decisions are not cached.
   - Tests: store/mutation_authorization_test.go:151 and :184
     (`TestPopulatedFSMutationsRefuseBeforeCallbacksAndNoops`; no-op and dry-run cases at :23-127)
     and appwiring/persistence_policy_test.go:17.
   - The new "no-ops" wording is backed by those tests rather than a strengthening. Summary stores
     use `ReadOnlyMutations` (spacestore/fs.go:105).
   - Preserved.
2. **Durable prefix and CAS limits.**
   - Old guide 520-526, 575-577 and 597-608 → persistence.md:45-57 and 97-107.
   - Owners:
     - core/store.go:69-80: write order is recovery data; `AppliedTaskIDs` is the durable prefix.
     - core/dependency_graph_mutation.go:26-41 and :92-95: "sound" means non-broken per prefix;
       the edge-only superset optimization.
     - store/cas.go:19: the hash is computed on read and never stored.
     - store/lock.go:69-101: planner-window conflict.
     - core/retry.go:54-66: bounded retry; previews never retried.
     - core/task_lifecycle.go:92-94: no blind retry of committed results.
   - Tests: threadapply_test.go:411, :447, :539 and :638; graphrepair_test.go:262;
     occ_retry_test.go:119; service_research_test.go:361.
   - The raw-editor verify-to-rename limit is preserved at persistence.md:45-47.
   - Preserved, except L4.
3. **Source and `Location` authority.**
   - Old guide 73-77 and 399-415 → planning-data.md:15-29, and tui.md:17-21 and 43-46.
   - Owner core/entity_read.go: :23-28 (Location is opaque; `LocationIsPath` is not a capability)
     and :50-54 (`LocalPath` is explicit; Location never authorizes repair).
   - Tests: dependency_repair_test.go:62, :119 and :169; entity_read_test.go:77;
     tui/local_path_test.go.
   - Preserved.

### Evidence and mutation ledger

Everything ran in the workspace, with `GOCACHE` in external scratch, on Go 1.27.2 darwin/arm64.

- **Navigation suite:** `go test -race ./internal/tools/architecturecheck -count=1 -v` passed: six
  guide subtests plus two helper tests.
- **Required mutation (core return route):**
  - Edit at core/store.go:8: replaced `// docs/architecture/planning-data.md#change-checklist; detailed contracts remain beside their types.`
    with `// the planning-data guide; detailed contracts remain beside their types.` The
    planning-data.md:7 link to core/store.go was left intact.
  - `go build ./internal/core` succeeded. Intended assertion: architecture_test.go:84-86.
  - Result: `--- FAIL: TestArchitectureNavigation/planning-data.md` with
    `architecture_test.go:85: guide needs a linked owning package comment that routes back to it`.
    No other subtest failed.
  - Restored; status clean; rerun passed. Killed.
- **Coordinated masking probe:** replaced the wire/wire.go:14 route with "See the CLI/wire guide".
  The build succeeded and the suite passed, because cli/root.go:4 also routes to cli-and-wire.md.
  This matches the documented "at least one" rule at architecture_test.go:57-59, so it is accepted
  design. Restored.
- **Blind-spot probe A (same-file fragment):** appended `See [map](#package-mapp).` to
  boundaries.md. The suite failed with `missing heading anchor "package-mapp"`. Falsified; restored.
- **Blind-spot probe B (fenced link):** appended a fenced block containing a link to
  `missing-in-fence.md` in tui.md. The suite passed, as designed. Restored.
- **Blind-spot probe C (forwarding anchor):** renamed the index heading. The suite passed and
  `lint --links` produced identical output. Reproduced (L3); restored.
- **Blind-spot probe D (fenced index link):** see L2. Reproduced; restored.
- **Focused race tests for disputed or traced claims, all passing:**
  - core: `TestExplicitAuditSnapshotSourceWinsRegardlessOfOptionOrder`,
    `TestLintSourceSuppliesDefaultAuditSnapshotSource`,
    `TestService_Summary_PreservesMixedPortableLoadDiagnostics`,
    `TestBoard_PreservesPathlessTaskLoadProblemIdentity`, both entity_read tests,
    `TestRetry_DryRunNotRetried`, `TestNewResearch_DoesNotRetryCommittedConflict`,
    `TestTaskGraphRepairDoesNotSelectOpaqueLocations`.
  - store: `TestTaskLifecycleRejectsRawThreadRaceByWholeSnapshotCAS`,
    `TestThreadApplyImmediateTargetCASRejectsRawEditAndRetryPreservesIt`,
    `TestThreadApplyRawEditAfterDurablePrefixIsReportedAndResumable`,
    `TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects`,
    `TestPopulatedFSMutationsRefuseBeforeCallbacksAndNoops`,
    `TestInvalidPolicyDoesNotExecuteConstructorOptions`.
  - appwiring: `TestPersistenceConstructorsRequireValidPoliciesWithoutIOOrAuthorization`,
    `TestPlanningOpenersPreserveThreadApplyAuthorization`.
- **`go doc`** on `./internal/core`, `./internal/wire` and `./internal/store`: routes render as plain
  repository paths and `[SchemaVersion]` renders as a doc link, consistent with index lines 38-40.
- **Planning lint:** `tskflwctl -C <workspace> lint --links`, using a binary built from the
  workspace into scratch, exited 0.
- **Wire and output:** only package-comment and documentation changes appear under internal/; there
  are no golden, schema or wire-revision changes in `git diff 4fdf6fcc HEAD --stat`.

### Surviving-hypothesis ledger

| Hypothesis | Outcome |
| --- | --- |
| Guide sentences lead an agent to the unsafe actions in the brief's hostile angles | Falsified; see the verdict and traces. |
| A named exception became a general license, or depguard reads as a full stdlib/framework ban | Falsified: boundaries.md:76-79 and 115-129 match `.golangci.yml:19-215` (default-deny, exact allows, `init.go`/`workspace.go` only). |
| The future dependency-policy ADR is presented as accepted | Falsified: boundaries.md:128-129; the task is ready-to-start; no ADR file exists. |
| "Every ordinary graph prefix must remain sound" strengthens non-broken to healthy | Falsified: the owner uses "sound" for non-broken (dependency_graph_mutation.go:26 and :93). |
| "Previews must not be blindly retried" invents a rule | Falsified: retry.go:60-66; occ_retry_test.go:119. |
| "Only a complete successful update advances the marker" strengthens the golden policy | Falsified: golden_test.go:179 and :270. |
| persistence.md:63-64 settles the open ID-uniqueness question in 6fkkz41cax80 | Falsified: it describes the enforced task/Thread pair and warns against inferring a union rule; it does not choose per-kind or global. |
| Dropped sentences lost owner-less contracts | Falsified for each sampled sentence. The owners are listed after this table. The surviving exception is L4. |
| Old "synthetic fs-event tests" versus new "real watcher coverage" | The old text was stale; watch_test.go has real watcher tests. The new text is accurate. |
| Another package masks a removed return route | Reproduced only for cli-and-wire, which the test documents as intended. Not filed. |
| The link regex misses wrapped inline links in checked docs | Reproduced: the README.md:44-45 link text spans a newline and is skipped (`[^\]\n]*`). The target exists today. The index, guides and AGENTS have no such links. Not filed separately; it could be folded into L2. |
| The Go-route regex truncates a malformed fragment (for example uppercase) and passes | Plausible from reading `guideRoute`, but not probed; there is no current instance. Not filed. |
| Pre-existing drift outside this slice | userconfig/userconfig.go:12-13 says the compiler enforces the config/userconfig rule, contradicting `.golangci.yml:204-209`. This slice did not change it, and boundaries.md:71-72 lists the rule as lint. Not filed; a candidate for 6g9cz5sk3b0d. |

Owners of the sampled dropped sentences:

- `projects/` retirement: config.go:668-670.
- EditFinding no-op and preview: finding.go:229-231.
- `ThreadPathSource` being parse-free: store.go:207-211.
- Rename receipt paths: task_rename.go:9-16.
- The pathless wire `path` field: wire.go:307-316.
- Dashboard duplicate suppression: dashboard.go:109.
- Committed cleanup that wraps `ErrConflict`: task_lifecycle.go:92-94 and
  service_research_test.go:361.

### Limits and inherited evidence

- I executed only the commands listed above. The owner ran the following and I did not re-run them:
  full `just test`, `just lint`, `just build` and `just docs-check`, plus planning lint on the source
  checkout and `git diff --check`.
- I did not validate network targets. The pre-split GitHub URL pins `4fdf6fcc`, which is the
  PR #301 merge commit on local `main`.
- I compared safety and compatibility claims sentence by sentence and sampled the
  presentation/history prose. I did not account for every deliberately removed roster, such as TUI
  keys or render type counts.
- I did not read the sibling Antigravity report or audit closed historical planning records.

## Owner disposition (2026-10-12)

All four Low findings were independently verified and fixed; the CLI resolution paragraphs record the exact dispositions. No production behavior or machine-contract revision changed. Full race tests, lint, build and generated-doc drift checks pass; focused evidence tests and three live navigation mutations also pass after restoration. Before owner edits, the landed audit was byte-identical to the report retained in Claude's attested sandbox. This corroborates the delivered artifact, not an assertion that the frozen report's transfer=pending attestation was self-finalizing. Broader documentation reconciliation remains task 6g9cz5t37dc0; selective package contracts remain task 6g9cz5sk3b0d.
