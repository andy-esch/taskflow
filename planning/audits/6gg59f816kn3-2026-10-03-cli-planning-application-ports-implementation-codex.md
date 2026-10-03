---
schema: 1
id: 6gg59f816kn3
bucket: closed
area: cli-planning-application-ports-implementation-codex
date: "2026-10-03"
updated_at: "2026-10-03"
---
# Audit: CLI planning application ports implementation — codex — 2026-10-03

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

Adversarial implementation review of
[Route CLI planning data operations through application ports](../tasks/6gcwcf8gzn50-route-cli-planning-data-operations-through-application-ports.md).
Challenge the actual runtime boundary and compatibility, not just the absence of imports. Do not
implement fixes, settle findings, or turn this into the next task's composition-root extraction.

Codex emphasis: composition consistency, mutation authorization and partial durability, real Cobra
flag timing, production consumer coverage. Antigravity emphasis: independent hostile filesystem
fixtures, test-helper blind spots, and coordinated mutations that expose falsely passing tests.
Both reviewers must inspect the complete contract, and neither should rely on the other's report.

## Review target

Branch `refactor/cli-planning-application-ports`, including staged, unstaged, and untracked changes
captured by the mandatory sandbox. Baseline before this implementation: main `a7707bc` (PR #274).
Review the captured implementation, not only `git show HEAD`; new files are part of the target.

Start with these verified implementation locations, then inventory every production consumer:

- `internal/core/planning_maintenance.go`: `RepairPlanning`, `LintWithLinks`, result/error contracts.
- `internal/core/completion.go`: optional port, candidate vocabulary, selection policy.
- `internal/core/service.go`, `store.go`, `source_set.go`: automatic selection, explicit options,
  checked source-set wiring, known missing-capability preflight.
- `internal/store/completion.go`, existing `fix.go`, `danglers.go`, `auditstore.go`, and
  `internal/core/finding.go`: naming/parse boundary, write guards, real body transforms.
- `internal/cli/completion.go`, `research.go`, `lint.go`, `root.go`: controller migration,
  deferred completion factory, runtime safety binding, ambient linkback warnings.
- `internal/core/configuration.go`, `internal/configstore/fs.go`: repository-only link diagnosis
  must not become a home-registry scan.
- New `planning_ports_test.go`, completion and maintenance tests; extended source-set, command
  safety, configuration and space-selection tests; existing lint/repair/wire fixtures.
- `docs/ARCHITECTURE.md`, task and Thread prose: claims must match actual shipped contracts.

## Intended contract to challenge

1. Ordinary frontmatter repair, opt-in body-link lint, and entity completion are named core use
   cases. Controllers neither instantiate another store nor open directories when a selected
   capability is unavailable. Root composition still legitimately constructs secondary adapters;
   local watcher `Layout` still bypasses semantic use cases explicitly.
2. The new optional repair, link, and completion capabilities share the existing source-set witness
   with every selected reader/mutator. Missing, empty, unstable, or mismatched witnesses fail at
   construction, before reads/writes. Typed-nil options remain absent. The aggregate `Store` and
   public wire schema are not widened.
3. Repair orders frontmatter before body repair (frontmatter can rename records), then post-lints
   real writes only. Known missing required capabilities fail before the first write. A later
   failure returns the completed/proposed prefix; no wholesale retry can replay durable work.
   Persistence retains its existing authorization, contention, and graph-owned-field protections.
4. Completion needs source identity, human alias, an adapter-owned unambiguous resolution reference,
   and optional observed state—not a path or parsed semantic record. Plain local completion reads
   names only. State-aware exclusion applies to individual records, not aliases, and alias counts
   include excluded records. Malformed id-led regular records remain addressable. README, stray
   non-id-led files, directories and symlinks are not ordinary candidates.
5. Real Cobra `__complete` callbacks compose after completed-command flags are parsed. `-C`,
   registered direct and pointer spaces, `--space` vs environment precedence, and invalid explicit
   selection cannot accidentally complete cwd. Missing services/ports/read failures are silent,
   and file completion remains disabled. Test injection must require no usable local directory.
6. Existing human/JSON envelopes, exit classification, dry-run behavior and safety remain intact.
   Intentional fixes: damaged same-slug siblings survive state exclusion; already-typed bare IDs
   suppress the same record; a post-lint failure now renders the repair prefix previously discarded.
   Distinguish these from unintended compatibility changes.
7. Ambient linkback warnings use a narrow configuration use case, do not scan the registry, remain
   best-effort, and preserve the entry-point/linkback context. Input body/manifest files, TUI
   watching, new graph/repair authorization rules, and broad init/doctor extraction are non-goals.

## Mandatory evidence floor

- Consumer inventory: trace every production repair, body-link, and five-entity completion call
  from Cobra through core to the selected capability. Search the entire repo for old App fields,
  direct store construction, glob helpers, fallback routes, and stale contract comments. Classify
  permitted wiring/process integration instead of treating all filesystem access as a defect.
- Use actual `__complete` invocations, not only direct calls to the new policy helper. Build two
  independent planning trees with disjoint records; complete one while cwd points at the other.
  Exercise both flag positions, environment precedence, selected pointer, invalid registry entry,
  no planning tree, all five kinds, duplicate aliases, and malformed metadata. Record exact stdout,
  stderr, directive and suggested selector resolution. Do not assume Cobra hook ordering.
- Independently seed an already-in-progress task and a broken same-slug sibling. Verify `task start`
  completes the damaged sibling by unambiguous reference, not an ambiguous alias, and verify the
  readable sibling still resolves independently. Include a missing/invalid status and an audit
  bucket case. Do not construct expected results by calling the production helper under test.
- Exercise actual repairs that rename an audit and canonicalize its finding body; compare before/
  after bytes and dry-run output. Force a second-document body failure and a post-lint failure:
  inspect durable/proposed prefix, rendered evidence, exit class, subsequent calls, and absence of
  whole-operation retries. Verify graph-owned fields and read-only command authorization still
  block before writing, including dry-run. A zero-audit fixture cannot prove body repair ordering.
- Execute at least three concrete mutations in the sandbox, restore each, and require the named
  regression test to fail for its claimed behavior. Suggested mutations: remove the new optional
  capabilities from the source-set check; prefilter aliases before counting or exclude by slug;
  compose during the early Cobra hook and reuse that service; post-lint during dry-run; discard
  completed results on a later failure. Compilation failure alone is not behavioral coverage.
- Include a coordinated architectural probe: remove validation for one new selected capability
  while supplying a foreign witness, or introduce a local fallback behind a failed injected
  completion factory. Run the exact production-controller test claimed to pin it. Report surviving
  mutations even if the unmodified suite is green, with an honest coverage/scope assessment.
- Demonstrate parse-free plain completion and bounded state-aware reads with instrumentation or
  a convincing hostile fixture, not a timing claim alone. Inspect whether shared test stubs return
  empty bodies or permissive no-op results that mask whole classes of production failures.
- For any ready/no-findings conclusion, provide at least three challenged hypotheses rejected by
  executed evidence. Do not invent findings to meet a quota. Unverified hypotheses remain unknown,
  not demonstrated defects; reports lacking the evidence floor are incomplete.

## Required hostile angles

First pass: source-set identity vs observed record identity; option ordering and auto-discovery;
typed nils and missing capabilities; malformed data and resolver-compatible selectors; collision
counting before state filtering; deferred Cobra parsing; dry-run/write parity; actual body mutations
after frontmatter renames; retained prefixes after read/write failure; safety guards after moving
calls into core; opaque diagnostics vs local opening authority; output/exit compatibility.

Second systemic pass: can a same-source-set adapter still supply mismatched naming/reading intent?
Does a fake's embedded interface hide unsupported calls? Are candidate suggestions guaranteed to
resolve to the observed source, or only superficially unique? Can stale flags or discovery context
retarget completion/warnings? Are errors suppressed only on the intended protocol? Are optional
ports honest capabilities rather than a service locator or newly mandatory broad interface? Does
the documentation claim enforcement that is actually deferred to the next task? Settle concrete
failure paths with evidence; label trusted-adapter assumptions and future enhancements accurately.

Antigravity: specifically attempt to falsify these assertions rather than restating them:
“the early hook is safe for flags”; “one remaining duplicate makes a slug unambiguous”; “a zero-body
stub proves repair”; “read-only safety is irrelevant to dry-run”; “checking only graph ports also
checks repair ports”; “silencing an error permits local fallback.” Re-run with independent fixtures
and record which test/command proves each assertion false or establishes the intended alternative.

## Validation and restoration

Only run inside the independent mandatory sandbox. Use task-specific caches outside the source:

```sh
export GOCACHE="$SANDBOX/.git/review-go-cache"
export GOLANGCI_LINT_CACHE="$SANDBOX/.git/review-lint-cache"
go test ./...
go test -race ./internal/core ./internal/store ./internal/cli ./internal/tui ./internal/wire
just lint
just docs-check
go build -o "$SANDBOX/.git/review-tskflwctl" ./cmd/tskflwctl
"$SANDBOX/.git/review-tskflwctl" -C . lint
git diff --check
```

Separate environment/toolchain blockers from product failures. Temporary planning trees belong in
scratch directories, never the real user's planning space or registry. In the sandbox, restore every
mutation, generated change and fixture before the guarded one-audit transfer. Do not stage or create
additional commits; the helper's initial baseline is the only permitted checkpoint. No push or PR
writes. If required evidence is incomplete, say so and keep the verdict appropriately bounded.

## Deliverable

Update only your assigned audit, preserving this brief. In `Reviewer report`, provide captured
baseline/isolation attestation, consumer inventory, exact validation results, an invariant-to-test
matrix, mutation outcomes, independent fixture/command transcripts, rejected hypotheses, and a
bounded ready/not-ready verdict. Any finding must have the exact required header grammar, concrete
path/line, reproduction, violated contract, impact and smallest useful remediation; leave it open.
Distinguish in-scope fixes from follow-ups/accepted scope and do not edit tasks or another audit.
Transfer only the assigned audit using the helper, retain the sandbox, and report its path.

## Reviewer report

> The report below describes the captured review baseline. Finding statuses and resolution notes
> have subsequently been updated by the implementation owner; see Owner reconciliation below.

### Verdict and review scope

**Both adversarial passes completed; not ready for acceptance pending owner triage of M1 and M2.**
The application-port migration, deferred Cobra composition, mutation authorization, and durable-prefix
handling held under the executed probes. The completion vocabulary does not yet guarantee that a
suggestion addresses the observed record, and already-typed suppression has a new alias-wide regression.
All findings remain **open**. No implementation fixes, task changes, other-audit updates, additional
commits, staging, pushes, or PR operations were performed.

Pass 1 traced the captured production consumers, executed real completion/repair commands, and tested
missing/foreign capabilities, malformed records, ordering, dry-run, authorization, and late failures.
Pass 2 challenged resolver compatibility, adapter honesty despite a shared witness, fake-method coverage,
case/ID alias collisions, already-typed siblings, early hook reuse, and fallback behind an injected
factory failure. This review did not read or rely on the other reviewer's report.

### Isolation and transfer attestation

The source checkout was used only to read this brief, invoke the prescribed creation helper, and perform
the final guarded helper transfer. Implementation inspection started in the independent clone. The
helper captured the complete working state, rather than reviewing only the source HEAD. Comparison
base: `a7707bc3fea1ce4b420b27241af3e2c95e93617f`. The sole reviewer baseline commit is below.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8SsE4q
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8SsE4q/.git
baseline_commit=e62b8371da3e031e46df3a40bcf3e8d9e724c681
source_blob=741b8368c4146124b2063dd50b9d2b0ac03bbbab
source_fingerprint=6ea8c8d30e0fd0281cfdbb38cb647d5108b42700
deliverable=planning/audits/6gg59f816kn3-2026-10-03-cli-planning-application-ports-implementation-codex.md
transfer=succeeded
```

The delivered report's transfer attestation applies to the successful guarded transfer; a refusal would
leave this draft undelivered and be reported separately. The helper verified the independent in-tree
Git directory, unchanged baseline, no staging, no unrelated delta, and original source-deliverable blob
before atomic replacement. Only this audit transferred. The sandbox remains in place until owner receipt.

Evidence scripts, full stdout/stderr, original/changed repair bytes, and mutation logs are retained under
`$SANDBOX/.git/review-evidence/`; these are reviewer scratch artifacts, not shipped regression fixtures.
The abbreviated commands and outcomes below are sufficient to reproduce the important observations.
`review-base-tskflwctl` was built from an archive of `a7707bc` inside the sandbox's `.git/review-base/`.

### Findings

#### M1. Completion suggestions are not guaranteed to resolve to their observed source · **Status:** fixed

**Location:** `internal/store/completion.go:51` initializes `Reference` to the full filename stem;
`:57` replaces ID/Slug but leaves that reference untouched. `internal/core/completion.go:81`–`:87`
selects aliases or references without the selected resolver's equivalence/precedence rules.
`internal/store/resolve.go:234` prioritizes canonical IDs, then `:255` matches case-insensitive aliases;
it never matches a flat record's combined `id-slug` stem. Task, audit, Thread, and research resolution
use those ID/alias candidates (`internal/store/fsstore.go:354`, `auditstore.go:210`,
`threadstore.go:119`, `researchstore.go:87`).

**Reproduction:** independent tree A contains `6fjangd7kve5-dup.md` with `status: in-progress` and
`6fjangd7kvf6-dup.md` with the literal body `damaged metadata\n`. While cwd is disjoint tree B:

```text
__complete -C A task start du
stdout: 6fjangd7kvf6-dup\n:4\n
stderr: Completion ended with directive: ShellCompDirectiveNoFileComp\n
exit: 0
-C A task show 6fjangd7kvf6-dup --json
stdout: empty; stderr error.code=not-found; exit: 10
-C A task show 6fjangd7kvf6 --json
stdout: empty; stderr error.code=validation, naming the actual damaged file; exit: 11
-C A task show 6fjangd7kve5 --json
stdout: task.id=6fjangd7kve5, task.slug=dup, status=in-progress; stderr: empty; exit: 0
-C A task show dup --json
stdout: empty; stderr error.code=ambiguous; exit: 13
```

The damaged record is enumerated, but its promised unambiguous suggestion cannot locate it. Starting it
also retains the existing broken-graph refusal; completion is not permission to mutate malformed data.
To separate that guard from the selector defect, a second tree contains two readable, graph-healthy
`dup` tasks in `ready-to-start`, with `tags: [review]`. Completion emits
`6fjangd7kva1-dup` and `6fjangd7kvb2-dup`. Executing **both emitted** `task start <suggestion>
--dry-run --json` commands returns exit 10 with a not-found move/error. Bare-ID control invocations
return exit 0 with lifecycle receipts for the corresponding source IDs. Full references returned for
Thread/audit/research ID-prefix completion likewise fail their actual `show` resolvers with exit 10.

The same boundary also permits a wrong source: a task with ID `6fjangd7kvc3` and alias
`6fjangd7kva1` completes that alias, but `show 6fjangd7kva1` returns the other task whose canonical ID
is `6fjangd7kva1`. A pair with aliases `Case`/`case` completes `case`, which resolves ambiguously
(exit 13), since core counts aliases case-sensitively and the local resolver does not.

**Contract/impact:** `CompletionCandidate.Reference` explicitly promises an adapter-understood
unambiguous selector (`internal/core/completion.go:13`). Duplicate disambiguation and the claimed
damaged-sibling fix therefore do not deliver addressability; an alias collision can even suggest a
selector for a different source. This affects all four flat entity kinds, not just one controller.

**History and coverage:** full-stem rejection, case ambiguity, and canonical-ID alias shadowing were
also reproduced using `a7707bc`; they are inherited behavior. Nevertheless the new optional port ships
a false reference guarantee and newly claims that the repaired exclusion policy makes damaged siblings
addressable. `TestComplete_StatusAware_DamagedDuplicateRemainsAddressable`
(`internal/cli/completion_test.go:103`) compares text only. The new core policy fixtures in
`internal/core/completion_test.go:28` similarly assume their local stem references resolve. None invokes
the production resolver with the suggestion. This is an in-scope contract gap, not a request to extract
composition wiring.

**Smallest useful remediation:** publish selectors the local adapter actually resolves (bare canonical
IDs where uniquely owned), use a record reference when a human alias is unsafe under that adapter's
resolution rules, and add completion-to-resolution/action tests with distinct IDs, damaged duplicates,
case collisions, and ID-shadowing aliases. Retain fail-closed behavior for duplicate canonical IDs.
Evidence: `fixture-transcripts.json`, `supplement-transcripts.json`, `final-fixture-transcripts.json`,
`base-alias-transcripts.json` and their independently seeded fixture scripts.

**Resolution:** The filesystem completion source now uses the actual resolver
candidate enumeration, emits bare canonical IDs, and certifies friendly aliases
only when case folding and ID precedence select the same source. Duplicate
canonical IDs are omitted rather than publishing a false reference. Permanent
store and real Cobra show/lifecycle tests execute emitted selectors across all
four flat kinds, malformed siblings, case collisions, ID shadowing, and epic
README carveouts.

#### M2. Typing one record suppresses every sibling sharing its alias · **Status:** fixed

**Location:** `internal/core/completion.go:66`–`:67` marks the whole slug as taken when a record's ID or
reference was typed; `:76` then excludes every candidate with that slug.

**Reproduction:** with the same two different-ID `dup` tasks:

```text
__complete -C A task complete 6fjangd7kve5 du
captured implementation stdout: :4\n
base a7707bc stdout: 6fjangd7kve5-dup\n6fjangd7kvf6-dup\n:4\n
__complete -C A task complete 6fjangd7kve5-dup du
captured implementation stdout: :4\n
base a7707bc stdout: 6fjangd7kvf6-dup\n:4\n
```

All four runs exit 0 with only Cobra's normal directive diagnostic on stderr. The base bare-ID behavior
also reoffered the already-typed record, as acknowledged in the task; correcting that must not hide the
other record. The independent reviewer test `TestReviewTypedRecordMustNotHideSibling` uses portable
IDs `first`/`second`, the same alias `dup`, and valid opaque references `record:first`/`record:second`.
With either `Args=[first]` or `Args=[record:first]`, it expects only `record:second`; both assertions fail
on the captured implementation (`got=[]`). This isolates the regression from M1's local reference bug.

**Contract/impact:** intended improvement #6 and task prose at
`planning/tasks/6gcwcf8gzn50-route-cli-planning-data-operations-through-application-ports.md:141` promise
suppression of the same record. Multi-record lifecycle commands instead lose eligible siblings; this
also affects any primary adapter using the new generic policy with valid opaque selectors.

**Smallest useful remediation:** track already-selected records by source ID/reference and keep alias
collision counting separate. Do not promote one selected record to an alias-wide exclusion. Add the
portable sibling test and a real multi-argument Cobra regression. Evidence: `sibling-regression.txt`,
`review_composition_test.go`, and `supplement-transcripts.json`. Leave owner triage open.

**Resolution:** Already-selected source IDs are tracked separately from
search-label counts and suggestion deduplication. Typing an ID or opaque
reference suppresses only that source, not same-slug siblings. Portable
ID/reference sibling regressions and real multi-argument Cobra lifecycle
completion cover the fix.

### Production consumer inventory

Repository-wide searches covered the named ports, `store.NewFS`, old App fields/helpers, glob calls,
filesystem reads, and fallback routes. Retained inventories: `full-port-inventory.txt`,
`consumer-inventory.txt`, `controller-access.txt`, and `stale-inventory.txt`.

| Consumer | Verified application-to-selected-port path |
| --- | --- |
| `lint --fix` | Cobra `internal/cli/lint.go:28` → `runLintFix` at `:76` → `Service.RepairPlanning` (`internal/core/planning_maintenance.go:23`) → selected `Fixer.FixFrontmatter`, then selected audit snapshot + aggregate `TransformAuditBody`, then selected lint reads for writes only. Production FS implementations: `internal/store/fix.go:25`, `body.go:189`. |
| `lint --links` | Cobra flag/RunE (`internal/cli/lint.go:35`) → `runLint` at `:40` → `LintWithLinks` (`internal/core/planning_maintenance.go:53`) → ordinary `Lint`, then selected `Linter.DanglingLinks` (`internal/store/danglers.go:18`). Ordinary lint does not call the optional link port. |
| Task completion | `internal/cli/task.go:275,320,361,456,569,610,721,770,806,860,945`; `edit.go:35`; `task_dependency.go:84,105,139,163`; cross-entity `thread.go:70,191,442`. All use `completeTaskSlugs`/`taskCompleter` → `entityCompleter` (`internal/cli/completion.go:99`) → `CompleteEntities` → selected `CompletionSource.ReadCompletionCandidates`. |
| Thread completion | `internal/cli/thread.go:68,100,299,323,341,362,389` → `completeThreadSlugs` (`completion.go:128`) → the same core use case, `EntityThread`. |
| Epic completion | `internal/cli/epic.go:35,66,143,190,339`; task `--epic` registrations at `task.go:210,250,691` → `completeEpicIDs` (`completion.go:145`) → the same core use case, `EntityEpic`. |
| Audit completion | `internal/cli/audit.go:171,238,320,435,477,522,560,582,618,678` → `completeAuditSlugs`/`auditCompleter` (`completion.go:133,141`) → the same core use case, `EntityAudit`; lifecycle exclusion supplies the destination bucket. |
| Research completion | `internal/cli/research.go:51,103,148,292,332` → wrapper at `:348` → the same core use case, `EntityResearch`. |
| Ambient linkback warnings | `repoPreRun` → `warnLinks` (`internal/cli/root.go:480`) → `ConfigurationService.RepositoryLinkProblems` (`internal/core/configuration.go:283`) → `configstore.FS.DiagnoseConfiguration` (`internal/configstore/fs.go:133`) → `config.CheckLinks`. No registry Catalog scan. |

`App` no longer has Fixer/Linter fields. `planningRoot`, `flatCompletions`, and `slugsFromGlobs` have no
production implementation left. A stale comment still names `planningRoot` at `internal/cli/root.go:365`;
historical planning documents also reference removed helpers. These are not shipped capabilities.
`Store` remains the same aggregate (`internal/core/store.go:354`); sibling ports are at `:369,380` and
`internal/core/completion.go:27`. No public wire shape was added by this change.

Permitted concrete wiring remains in `internal/cli/root.go:268`–`:276` and `:452`,
`internal/workspacestore/fs.go:39`, and `internal/spacestore/fs.go:112`. `Layout` use in
`internal/cli/ui.go:141`, `internal/tui/tui.go:46`, and `internal/tui/atlas.go:181` is local watcher
integration. `internal/store/atomic.go:143` globs its own crash-temp files. The remaining controller
`os.ReadFile` calls are explicit input/process integration: task body files (`task.go:47`), external
editor results (`edit.go:105`), Thread manifests/plans (`thread_apply.go:107,142`), and dependency repair
manifests (`task_dependency_repair.go:180`). No hidden planning completion/repair fallback was found.

### Validation results and invariant matrix

All commands ran against the isolated captured implementation, with sandbox Git-local Go/lint caches.

| Validation | Result |
| --- | --- |
| `go test ./...` | PASS; `go-test.txt`. |
| `go test -race ./internal/core ./internal/store ./internal/cli ./internal/tui ./internal/wire` | PASS for all five packages; `race.txt`. |
| `just lint` | PASS, `0 issues.`; `lint.txt`. |
| `just docs-check` | PASS; generated CLI reference unchanged; `docs-check.txt`. |
| `go build -o "$SANDBOX/.git/review-tskflwctl" ./cmd/tskflwctl` | PASS. Go printed a denied home module-cache stat write; build exited 0 and all binary probes ran successfully. This was an environment diagnostic, not a product failure. |
| `review-tskflwctl -C . lint` | PASS, `✔ all planning entities and dependency links pass lint`; `planning-lint.txt`. Repeated after report editing before transfer. |
| Nondefault semantic wire validation | PASS: `FixEnvelope` with nonempty/pathless fixes, skipped graph guard, unreadable opaque record, remaining issues, dry-run and pointer/space context; `LintEnvelope` with nonempty link problems. Actual rename dry/write stdout also validated against `FixEnvelope`. `wire-validation.txt`; reviewer test `review_schema_test.go`. |
| Final restored focused core/store/CLI regressions | PASS with ordinary Go temporary-directory discovery; `final-focused-default-tmp.txt`. |
| Restoration and `git diff --check` | PASS; only the assigned audit differs, no staged changes, baseline HEAD unchanged; helper verification before transfer. |

An initial scratch validator used a nonexistent `LintResult.Kind` field and failed to compile. It was
corrected after reading `internal/core/service.go:677`; only the subsequent successful run counts.
No compilation failure is counted as a mutation kill.

A final focused run with `TMPDIR` forced beneath the sandbox failed `TestComplete_OutsideRepo_Quiet`: its supposedly bare `t.TempDir()` then legitimately discovered the ancestor sandbox planning config. Repeating the same restored regressions without that temporary-directory override passed all three packages. This is fixture/environment coupling, not evidence of a product fallback; the independent `-C /` no-tree control returned only `:4`. The initial full and race suites used ordinary temporary-directory discovery and passed. Both logs are retained.

| Invariant | Executed evidence / practical limit |
| --- | --- |
| New ports share selected source-set witness before I/O | `TestNewServiceRejectsEveryMismatchedSplitCapability` and `TestNewServiceRejectsMissingOrEmptySourceSetIdentity` (`internal/core/source_set_test.go:61,96`); reviewer `TestReviewOptionalPortComposition` exercised empty, unstable, foreign witnesses for each new port with zero business calls. Same-witness option order and typed-nil options retained the explicit selection. |
| Typed nil and missing workflows remain absent/preflighted | `TestMaintenanceTypedNilOptionsStayAbsent`, `TestMaintenanceMissingCapabilitiesFailBeforeMutation` (`internal/core/planning_maintenance_test.go`); selected-write preflight mutants failed before any write. Automatic FS discovery verified through the actual binary repair/link/completion paths. |
| Completion remains portable and errors do not fallback | `TestCLICompletionUsesInjectedApplicationForEveryEntity`, `TestCLICompletionFailuresStaySilentWithoutLocalFallback` (`internal/cli/planning_ports_test.go:71,95`); injected service works with `/no/local/planning/corpus`. Failed-source/composition/missing-port paths yield only `:4`. Coordinated fallback mutant killed by the production-controller test. |
| Flags compose after the completed command is parsed | Three actual `-C` placements, direct/pointer spaces, env/flag precedence, invalid selections; transcripts below. Early-hook + reuse coordinated mutation killed by `TestComplete_GlobalSpaceAndSelectedEntities` and `TestComplete_TaskSlugs_IncludesMalformed`. |
| Collision counting precedes observed state exclusion | Actual task/audit damaged-sibling fixtures; missing/invalid task statuses and audit buckets remain offered. Prefilter/global-slug mutants killed. Addressability itself fails M1; selected-record suppression fails M2. |
| Plain completion reads no record bodies | Reviewer `TestReviewCompletionReadBoundaries` instrumented the existing `completionState` read boundary without changing semantics: plain request = 2 candidates / 0 file reads; state-aware = exactly 1 read per eligible task, including damaged one. No README/stray/directory/symlink reads. `instrumentation.txt`; production source restored. |
| Frontmatter precedes body and post-lint is write-only | `TestRepairPlanningOwnsOrderingAndDryRun` and actual renamed-audit fixture; body-before-frontmatter and dry-run-postlint mutants killed. Real renamed body repaired after the new name exists. |
| Durable/proposed prefix survives late failure, no whole replay | Actual two-audit native body failure; reviewer FS-backed `TestReviewActualBodyFailureAndPostLintPrefix`; permanent `TestRepairPlanningRetainsActualBodyRepairPrefix` and `TestRepairPlanningRetainsPrefixAndDoesNotRetryFailures`. Prefix-discard and CLI-render-discard mutants killed. |
| Existing authorization/content CAS/graph guards retained | `TestReadOnlyCommandCannotReachAnyMutatingApplicationBoundary` (`internal/cli/command_safety_test.go:71`), reviewer nonempty-audit read-only test for both modes, actual graph-field refusal for both modes, full/race suites including `TestTransformAuditBody_ConflictsOnConcurrentContentEdit` and `TestTransformAuditBody_TransformSeesFreshBodyAfterConcurrentAppend` (`internal/store/transformauditbody_test.go:91,117`). |
| Warnings preserve pointer context without registry scan | `TestRepositoryLinkProblemsDoesNotScanRegistry` (`internal/core/configuration_test.go:29`) plus actual missing-linkback/invalid-registry fixture. Both emitted alternative init commands executed against a one-sided state and cleared the warning. Completion suppresses that warning. |

### Independent command and byte evidence

Retained scripts: `fixtures.py`, `supplement.py`, `final-fixtures.py`, `native-failure.py`.
Here `CLI=$SANDBOX/.git/review-tskflwctl`, `A/B/pointer` are under
`$SANDBOX/.git/review-evidence/fixtures/`, and every command uses a private `TSKFLW_CONFIG_HOME`.
A and B were independently initialized with `init --path <tree> --no-register`; pointer was initialized
with `--planning-repo A`. Their ordinary records have aliases `alpha` and `bravo`, respectively.
Metadata/body fixture bytes and explicit expected IDs were seeded independently, without consulting
the policy helper. `README`/stray files, a directory named `.md`, and a symlink were also planted.

All completion runs below exit 0, disable file completion with `:4\n`, and have exact stderr
`Completion ended with directive: ShellCompDirectiveNoFileComp\n`, unless otherwise stated. This normal
Cobra diagnostic is not an application error or warning. Full argument vectors and unabridged output
are in the transcript JSON files.

| Actual invocation, cwd B | Exact stdout (escaped newlines) |
| --- | --- |
| `__complete -C A task show al` | `alpha\n:4\n` |
| `-C A __complete task show al` | `alpha\n:4\n` |
| `__complete task show -C A al` | `alpha\n:4\n` |
| `__complete --space direct task show al` | `alpha\n:4\n` |
| `__complete --space pointer task show al` | `alpha\n:4\n` |
| `TSKFLW_SPACE=pointer __complete task show al` | `alpha\n:4\n` |
| `TSKFLW_SPACE=other __complete --space direct task show al` | `alpha\n:4\n` |
| `TSKFLW_SPACE=gone __complete -C A task show al` | `alpha\n:4\n` |
| `__complete --space gone task show ''` / unknown label / malformed registry | `:4\n` (missing explicit target never falls back to B) |
| `__complete -C / task show ''` | `:4\n` (actual no-tree discovery). |
| `__complete -C A thread show ''` | `alpha\n:4\n` |
| `__complete -C A research show ''` | `alpha\n:4\n` |
| `__complete -C A epic show ''` | `17-project\nreadme\nstray\n:4\n` |
| `__complete -C A task show ''` / `audit show ''` | `6fjangd7kve5-dup\n6fjangd7kvf6-dup\nalpha\ninvalid\nmissing\n:4\n` |
| `__complete -C A task start ''` | `6fjangd7kvf6-dup\nalpha\ninvalid\nmissing\n:4\n` |
| `__complete -C A audit close du` | `6fjangd7kvf6-dup\n:4\n` |
| `__complete -C A audit reopen ''` | `6fjangd7kve5-dup\n6fjangd7kvf6-dup\ninvalid\nmissing\n:4\n` |

Epics intentionally keep legacy filename stems as candidates (`internal/store/resolve.go:120`), so the
non-id-led epic exception is distinct from flat records. Lowercase `readme.md` is erroneously offered
by both base and capture, although the resolver skips it case-insensitively (exit 10); it is another
inherited illustration of M1's enumeration/resolution mismatch. A first attempted no-tree fixture under
sandbox `.git` discovered the legitimate ancestor sandbox planning tree. It was rejected as a no-tree
control and replaced by the `-C /` command above; no cwd-retargeting finding is based on that fixture.

**Rename + body repair:** `6fjangd7kvii-rename.md` contained schema/id, open bucket, area/date, and a
nonempty findings body with a hyphenated code. Dry-run exits 0, reports one canonical-ID filename
repair, leaves original bytes unchanged, and creates no target. Real `lint --fix --json` exits 0,
removes the old file, creates `6fjangd7kv11-rename.md`, updates its ID, and canonicalizes the body header.
Its stdout has two fixes (old-path frontmatter evidence, then new-path body evidence). The next call
returns `fixed: []`, exit 0. Original and final bytes are in `repair-before.md`/`repair-after.md`.
Dry-run does not preview the body of a record whose invalid filename has not actually been renamed;
this omission was reproduced on `a7707bc` too and is recorded as retained compatibility, not a new
orchestration regression or proof that dry-run models prospective state.

**Native second-document body failure:** two valid audit filenames contain near-miss headers; the
second also ends in an unterminated code fence. A task needs ID backfill. Actual binary `lint --fix
--json` exits 11, stdout renders exactly two completed fixes (task + first audit), and stderr contains:

```text
error.code=validation
canonicalize finding headers in second: validation failed: audit body ends inside an
unterminated ``` fence opened at line 5; everything after it is invisible to the structure scanners
```

The first audit and task bytes changed, the second audit did not. `--dry-run` on an independently
seeded twin also exits 11 and renders two proposed fixes while all three byte strings remain unchanged.
After closing the second fence, the next write repairs only the second audit; the following call has
zero fixes. Both later calls still exit 11 because the deliberately incomplete task retains unrelated
required-field lint issues. The original body failure is gone and its durable prefix is not replayed.
See `native-failure-transcripts.json`.

**Forced post-lint failure and human rendering:** reviewer `TestReviewActualBodyFailureAndPostLintPrefix`
uses the real FS frontmatter/body implementations behind a counting wrapper that forces either the
second body operation or `ReadLintTasks` to return `ErrValidation`. Real Cobra execution covers
human/JSON × dry/write modes with two nonempty audit bodies. Body failure renders 2 fixes and exits
class 11; post-lint failure renders 3 durable fixes and exits class 11; dry-run never calls post-lint
and exits 0 in that case. The frontmatter operation runs once per invocation. After clearing the fault,
a write invokes only the still-unrepaired body (or no body after a post-lint failure). Human output
lists the same changes and count as JSON. `hostile-repair.txt` records exact stdout/stderr, call order,
and byte assertions. This is a forced adapter-error probe, not a claimed natural on-disk post-lint fault.

**Safety:** a read-only-classified actual Cobra lint command with nonempty audit body and repairable
frontmatter rejects both write and dry-run before touching bytes. Direct graph-owned-field fixtures
with scalar `depends_on` and ordinary comma-list `tags` return `skipped: true`; every byte remains
unchanged, including tags. Dry-run exits 0 and real write exits 11 after reporting residual lint.
The recommended `lint` commands were run against those same broken trees. Body persistence still
uses authorization before reads (`internal/store/body.go:195`), content revalidation before write
(`:233`), and the repository-root directory flock on Unix (`internal/store/lock_unix.go:23`), plus the
same-process root-keyed guard (`internal/store/lock.go:51`). No separate lock-file path is assumed.

**Warnings:** deleting A's tracked pointer link while leaving an invalid private registry yields a
one-sided-link warning through `-C pointer task show alpha --json`; stdout remains a valid show envelope,
exit 0. Completion from the pointer yields only `alpha\n:4\n` and normal Cobra stderr. Executing
`init --path A --track pointer --no-register`, and separately reseeding then executing
`init --path pointer --planning-repo A --no-register`, both exits 0 and clears the subsequent warning.
This verifies both emitted repair alternatives rather than merely inspecting their prose.

### Executed mutations and effectiveness limits

Each mutation was compiled and executed in the sandbox, then restored byte-for-byte before the next
probe. Logs/scripts record exact replacement strings and named `go test -run ... -count=1 -v` commands.
**27 compiling mutation variants: 26 behavioral kills; 1 surviving single-site mutation explained and
then killed by a coordinated variant.** No build failure counts toward that total.

| Mutation | Named regression(s) and outcome |
| --- | --- |
| Remove each new port from `validateSourceSets` (3 variants) | `TestNewServiceRejectsEveryMismatchedSplitCapability` + `TestNewServiceRejectsMissingOrEmptySourceSetIdentity`: all KILLED with admitted invalid construction. |
| Count aliases only after state exclusion | `TestCompleteEntitiesPreservesResolutionAndMalformedCandidates`: KILLED, damaged sibling offered as ambiguous `dup`. |
| Remove only the first typed-ID check | Same regression: SURVIVED. The later `taken[candidate.ID]` check still suppressed the unique fixture. This is masking by another enforcement site, not evidence the ID invariant is untested. |
| Remove both typed-ID checks, coordinated | Same regression: KILLED, typed unique record reappears. Its sibling blind spot remains independently demonstrated by M2. |
| Exclude every record with an excluded sibling's slug, coordinated | Same regression: KILLED, damaged sibling disappears. |
| Request state on all core completion queries | `TestCompleteEntitiesAcceptsOpaqueSelectorsForEveryKind` + `TestCompleteEntitiesPreservesResolutionAndMalformedCandidates`: KILLED. |
| Drop retained fixes on late body failure | `TestRepairPlanningRetainsActualBodyRepairPrefix` + `TestRepairPlanningRetainsPrefixAndDoesNotRetryFailures`: KILLED. |
| Post-lint a dry-run | `TestRepairPlanningOwnsOrderingAndDryRun`: KILLED by observed call sequence. |
| Remove known workflow capability preflight | `TestMaintenanceMissingCapabilitiesFailBeforeMutation`: KILLED after a forbidden frontmatter call. |
| Drop CLI rendering of the late prefix | `TestCLILintMaintenanceUsesPortableApplicationPorts`: KILLED, missing machine output. |
| On failed injected factory, call `a.resolve()` and use its local service | `TestCLICompletionFailuresStaySilentWithoutLocalFallback`: KILLED by actual controller output becoming local candidates. This is the required coordinated architectural probe: factory failure plus local composition, with the exact production-controller regression. |
| Compose in early `__complete` hook AND reuse `app.Svc` in deferred factory | `TestComplete_GlobalSpaceAndSelectedEntities` + `TestComplete_TaskSlugs_IncludesMalformed`: KILLED by wrong corpus. |
| Parse state for plain FS requests | `TestCompletionSourcePlainQueriesDoNotParseState`: KILLED by nonempty state. Separate instrumentation verifies actual zero reads rather than only output. |
| Admit symlink entries | `TestCompletionSourceSkipsSymlinks`: KILLED. |
| Replace observed ID with parsed task declaration | `TestCompletionSourceRetainsMalformedRecordsAndIndependentIdentity`: KILLED. |
| Select typed-nil repair/link/completion options (3 variants) | `TestMaintenanceTypedNilOptionsStayAbsent`: all KILLED. |
| Silently succeed without core completion port / return failed-source partial candidates (2 variants) | `TestCompleteEntitiesDoesNotFallbackOnMissingOrFailedCapability`: both KILLED. |
| Route ambient repository diagnosis through registry-composing `Diagnose` | `TestRepositoryLinkProblemsDoesNotScanRegistry`: KILLED by the intentionally unusable registry. |
| Run body links after failed ordinary lint | `TestLintWithLinksUsesOptionalCapabilityOnceAfterLint`: KILLED. |
| Repair bodies before frontmatter | `TestRepairPlanningOwnsOrderingAndDryRun` + `TestRepairPlanningRetainsActualBodyRepairPrefix`: KILLED by ordering/prefix assertions. |
| Reject opaque candidate IDs/references | `TestCompleteEntitiesAcceptsOpaqueSelectorsForEveryKind`: KILLED for all kinds. |
| Route every CLI entity kind as task | `TestCLICompletionUsesInjectedApplicationForEveryEntity`: KILLED for the other kinds. |

Mutation details are in `mutations.py`, `coordinated.py`, `extra-mutations.py`,
`remaining-mutations.py`, their result JSON, and `mutant-*.txt`/coordinated logs.

Shared fake inspection: `cliPlanningStub` embeds `core.Store` (`internal/cli/planning_ports_test.go:18`)
and returns an empty audit snapshot (`:49`), so its maintenance tests prove controller routing/prefix
rendering but cannot prove any audit body transform. `maintenanceStore`'s ordering test similarly has
no audits by default. `headerMaintenanceStore` does execute the callback and force a later body failure
(`internal/core/planning_maintenance_test.go:24`), but it is a memory model rather than filesystem
rename/CAS evidence. `sourceSetProbe` embeds unsupported interfaces intentionally only for constructor
checks (`internal/core/source_set_test.go:13`); its methods must never be treated as operational coverage.
The real byte fixtures, native fence failure, bounded-read instrumentation, and actual suggestion
resolution above address those limits. The green string-only completion tests remain insufficient for
M1/M2 despite the many other killed mutations.

### Second-pass conclusions, rejected hypotheses, and boundaries

- **Rejected: an early successful discovery makes reuse safe.** The coordinated early-hook/reuse
  mutation fails actual Cobra regressions; independent B-cwd commands prove deferred composition
  honors all exercised selectors. Invalid explicit selection yields `:4`, not cwd candidates.
- **Rejected: one eligible duplicate makes a human alias unambiguous.** The full-set count mutant
  offers `dup` and fails; actual `show dup` is ambiguous. The retained full reference still fails M1.
- **Rejected: silent factory failure permits a local fallback.** Adding that fallback makes the exact
  injected-controller regression fail. Errors are suppressed on the completion protocol, while normal
  failed shows/lifecycle/repair commands retain error envelopes and exit classes.
- **Rejected: moving repair into core bypasses read-only authorization, including dry-run.** Real
  nonempty-audit probes reject before bytes change. Graph-owned repair guards likewise remain active.
- **Rejected: a zero-body fake establishes body durability/order.** Inspection showed that blind spot;
  native two-body failure and actual rename bytes independently establish the successful orchestration.
- **Rejected: witness equality certifies adapter honesty or selector intent.** Reviewer
  `TestReviewWitnessDoesNotCertifyAdapterHonesty` explicitly constructed an FS over a foreign tree using
  `WithSourceSetID(reader.SourceSetID())`. Construction succeeds, completion returns `foreign`, and the
  selected reader's `ShowTask(foreign)` returns not-found. This is intentional trust in the composition
  owner: `internal/core/source_set.go:14` calls the token a wiring witness, not a security credential;
  `internal/store/fsstore.go:59` permits explicit binding. It is not reported as a new enforcement defect.
  Actual root wiring passes one FS, rather than falsely binding foreign roots.

Documentation correctly leaves executable controller-import enforcement to the next task
(`docs/ARCHITECTURE.md:97` and Thread step 8 at
`planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md:53`). The Thread's step 7 still
says implementation is in progress pending adversarial review. The task's completion-addressability and
same-record-suppression claims at `:127`–`:141` are stronger than the executed behavior (M1/M2).
A future composition-root extraction, remote resolver verification, or stronger untrusted-adapter
security mechanism is outside this review. No findings were settled or automatically converted to work.

## Owner reconciliation — 2026-10-03

Both findings are accepted and fixed in this implementation, using the normal `audit finding` verbs
to record their statuses and resolution notes. No changes to resolver vocabulary or relaxed mutation
guards were introduced to make invalid full-stem suggestions work.

- **M1:** completion uses the same `flatCandidates`/`epicCandidates` enumeration as resolution.
  Flat references are bare canonical IDs. `SlugIsReference` is explicit adapter evidence, defaulting
  false, that a human label resolves to the same source; local certification checks case-folded
  alias collisions, canonical-ID precedence, and plain-query validity. Duplicate canonical IDs
  are omitted rather than presented as safe selectors. Epic README exclusion is case-insensitive.
- **M2:** selected source IDs, label occurrence counts, and emitted-suggestion deduplication are
  separate. An ID or reference selects only its source. Shared labels do not exclude siblings.
- Permanent tests resolve every returned reference against the actual filesystem resolver for all
  four flat kinds, then execute actual Cobra `show` for duplicate/case/ID-shadowing fixtures and
  dry-run lifecycle actions for emitted task references. Damaged-record completion reaches the
  expected validation failure rather than not-found. Opaque ID/reference sibling tests and a real
  batch-command completion test cover selected-record behavior independently of filesystem naming.
- Owner mutation probes in independent sandbox `isolated-review.XYYvDr` reintroduced full-stem
  references and alias-wide suppression. `TestCompletionSourceReferencesMatchActualResolver` failed
  with actual not-found errors for all four kinds; `TestCompleteEntitiesSelectedRecordDoesNotHideSibling`
  failed for both typed ID and opaque reference. Restoring production code made those tests and the
  actual Cobra regressions pass. The earlier populated-cwd fallback hardening remains in place.

The captured original review remains valid evidence; these fixes and the fixture-only fallback change
are explicit post-capture deltas, not silently attributed to the reviewer. Full `go test -race ./...`,
standard lint and generated CLI-doc checks pass after the fixes. No further design call or new
out-of-scope implementation issue was demonstrated. Existing prospective dry-run limitations and
the next composition-boundary enforcement task remain scoped as documented, not reopened here.

Owner checked byte-identical audit delivery before disposition, the recorded independent baseline,
and the helper's retained `transfer-attestation.txt` with `transfer=succeeded`. Review receipt is
confirmed. Both audits are reconciled; current task remains in progress for the commit/PR handoff.
