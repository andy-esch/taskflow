---
schema: 1
id: 6gg7h65sfjrb
bucket: closed
area: cli-composition-boundary-implementation-codex
date: "2026-10-03"
updated_at: "2026-10-03"
---
# Audit: CLI composition boundary implementation — codex — 2026-10-03

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

Adversarial implementation review of CLI composition isolation and executable controller boundaries.
Challenge whether this is a real boundary rather than construction moved behind a prettier name.
Do not implement fixes or change planning state; report actionable findings for owner triage.

## Review target

- Branch: `refactor/cli-composition-boundary`, based on main `ec5185e` (PR #275).
- Review the captured staged/unstaged/untracked snapshot, not just the commit diff. The independent
  sandbox baseline includes the new packages; record its commit and source fingerprint.
- Task: `planning/tasks/6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md`.
- Main surfaces: `internal/appwiring`, `internal/cli/ports`, `internal/cli/root.go`, `ui.go`,
  `config.go`, `completion.go`, neutral config consumers, command guards, binary/tool entrypoints,
  `.golangci.yml`, and the matching architecture guide. Many existing tests only rename the helper
  used to select real adapters; inspect those deltas for weakened assertions.
- Baseline architecture audit: `planning/audits/6gefs7wbbcyt-2026-09-28-arch-hexagonal-boundaries.md`.
  M1/L1 are locally fixed; M4 shipped in PR #275. M2/M3 have explicit followup tasks, not claimed fixes.

## Intended contract to challenge

1. Only the binary selects local wiring; production CLI controllers cannot import it, concrete
   persistence, or UI frameworks. Invocation contracts depend only on core and presentation values.
2. Composition is per command tree, with no singleton or local fallback. Every actual persistence
   family and later workspace-opened store receives that invocation's authorization closure.
3. Planning opening is lazy, publishes one observed corpus atomically on success, and retains source-set
   checks, local watcher layout, and the ordinary CLI identity reader. Cobra completion opens only
   after completed-command flags are parsed; failures stay silent without discovering cwd instead.
4. Startup/home preferences use neutral values without extra configuration inspection or registry
   scans. Theme/pager precedence, pointer/space targeting, receipts, errors, schema, and command safety
   stay compatible. Missing operational bindings fail explicitly rather than panic or open local data.
5. Standard lint rejects unknown internal adapters and nested sources, not merely four named stores.
   Exact package allowances must not grant blanket prefix access. Init topology and local checkout
   receipts are file-specific exceptions, not unrestricted controller or config fallbacks.
6. UI gates and selection stay in controllers; concrete Bubble Tea options/execution stay in wiring.
   CLI docs and manpage generators construct metadata-only trees without local composition.

## Mandatory evidence floor

- Build a **consumer inventory**: every production `NewRootCmd`/`ChromeTheme` caller, all wiring hooks,
  persistence constructor/identity-reader sites, launch paths, retained `config` imports, and lint rule
  file scopes/allowances. Name verified symbols/lines; distinguish a constructor from a use-case call.
- Run clean baseline `go test ./...`, `go test -race ./...`, `just lint`, and `just docs-check`.
  Record tool versions, exact commands, observed outcomes, and any blocked verification.
- Independently reproduce compiler-valid forbidden-import probes under actual standard lint: known
  persistence, `appwiring`, a new adapter not named in config, a child of an allowed package, nested
  CLI/TUI sources, neutral ports, and both named exception files. A type-check/formatting failure is
  not evidence that depguard works. Restore and rerun clean lint.
- Execute at least three runtime mutation experiments against focused regressions: early planning
  opening, lost/misbound authorization, and failed/absent composition falling back to real populated
  cwd data. Use actual local candidates for fallback, not a hard-coded invented candidate. Pair each
  mutant with the intended test and restoration result. Report surviving mutants honestly.
- Inspect every new regression and challenge the precise defect it claims to detect. If broad tests
  only prove fake composition and cannot observe extra I/O inside local wiring, state that limit;
  do not claim every discovery or constructor regression is covered because the suite is green.

## Required hostile angles

**Codex lens: runtime identity, safety, and semantic compatibility.** Cross-check authorization
reachability for ordinary planning writes, dry-run, repair, config, registry, init, UI launch, and
later workspace opening. Reuse bindings across two command trees and try to exchange their policies.
Stress pointer/space selection and configuration replacement before Thread apply. Check startup
metadata mapping, unset versus false pager settings, home-config errors, help/error chrome, UI atlas
startup, and local receipt provenance. Seek concrete behavior drift, partial publication, or a second
corpus read—not only direct imports.

**Antigravity lens: escape routes and deceptive tests.** Start with actual compiler-valid mutations
rather than reading the happy-path tests. Try moving a forbidden import into an unknown CLI child,
under an allowed package prefix, or into `init.go`/`workspace.go`; verify the responsible depguard
rule really fires. Try removing a launch hook, returning an incomplete service bundle, and failing an
explicit opener while cwd has valid records. Execute real Cobra commands and capture stdout/stderr
and error class. Then audit the test-only composition helpers: do they inadvertently install guards,
resolve the wrong target, or override hooks in ways that conceal the changed production wiring?

**Both reviewers:** do a second systemic pass. Challenge whether factories/launch requests are a
service locator in disguise, whether consumer ownership is honest, and whether allow-list/exception
growth is maintainable. Any objection must name a current failure or a justified future extension
with a bounded remedy, not a preference for a larger framework. Check docs against actual imports.

Known scoped followups: workspace opening already lacked the Thread-apply identity reader before
this extraction (6gg7e594gcms); adapter authorization omission outside actual binary wiring remains
optional (6gg7e59cyxxh, design checkpoint); the policy ADR is proposed, not accepted (6gg7e59mm68g).
Verify those claims. Report new regressions or materially incomplete scoping, but do not count the
same recorded residual issue as a newly introduced bug. Existing cwd-only chrome behavior is not
intended to change in this task.

## Validation and restoration

The owner ran full/race tests, standard lint, generated-doc drift checks, planning lint, and a fresh
binary's throwaway config/status and TUI launch checks. Sixteen formatted forbidden imports were
rejected specifically by depguard in an independent checkout. These are claims to verify, not your
evidence. Reproduce with scratch probes only in your assigned independent sandbox.

Prefer targeted tests in `internal/cli/composition_test.go` and `internal/appwiring/wiring_test.go`
for experiments, then the integration suites. Never change the shared source/index, weaken production
guards to make fixtures pass, or transfer a code fix. Restore probes to your captured sandbox baseline;
only the assigned audit may differ at transfer. Record unavailable TTY/tool tests as unavailable.

## Deliverable

Preserve this brief and fill the Reviewer report. Include the consumer inventory, command/evidence
matrix, mutation-to-regression results (including survivors), systemic second-pass conclusions,
readiness verdict, and mandatory isolation attestation. Findings use exact canonical headers with
status **open**, verified path/line, executable reproduction or clearly labeled inference, impact,
and a bounded recommendation. Do not close findings, alter task ACs, or manufacture a finding quota.
Separate real regressions, known followups, test gaps, and unsupported preferences.

## Reviewer report

Reviewed by Codex on 2026-10-03. Both required passes are complete: an implementation/compatibility pass, followed by a systemic pass using compiler-valid import probes, runtime mutations, populated-cwd commands, and real terminal launches. Reviewed the captured working snapshot, including the new packages, against main `ec5185ea41f6b0f05d0ba2bae80afa4d5689ba48`; the sandbox baseline is the restoration authority.

**Verdict: owner triage required; findings remain open.** M1 demonstrates an incomplete UI composition accepted at the new invocation boundary. M2 demonstrates an unenforced external framework boundary. L1 records a demonstrated systemic regression-test gap. I found no wrong-target write or authorization bypass in the ordinary, fully populated `LocalBindings` binary. That narrower result does not support a no-findings signoff.

### Consumer inventory — verified production code

The inventory below comes from the sandbox's actual Go sources, not proposed tasks. Searches excluded `*_test.go`; test helpers are inventoried separately.

| Consumer or boundary | Verified implementation and distinction |
| --- | --- |
| Every production root constructor caller | `cmd/tskflwctl/main.go:23-24` selects `appwiring.LocalBindings` and calls `cli.NewRootCmd`. `internal/tools/docgen/main.go:27` and `internal/tools/mangen/main.go:31` are the other two callers; both pass `ports.Bindings{}`. No other production caller was found by the `NewRootCmd(` search. |
| Every production chrome caller | `cmd/tskflwctl/main.go:38` calls `cli.ChromeTheme` with those bindings. Its definition is `internal/cli/root.go:209`; repository/user presentation reads occur at lines 212/218. These are presentation reads, not service construction. Existing cwd-only chrome targeting remains unchanged. |
| Invocation contracts and ownership | `internal/cli/ports/runtime.go:15` defines `Planning`; lines 25 and 38 define `Services` and `Bindings`. They currently import only `core`, `design`, and standard `io`. The six binding hooks are `Compose`, `ReadRepository`, `ReadUser`, `IsMissingPlanning`, `RunBrowser`, and `RunConfiguration`; `Services.OpenPlanning` is the lazy seventh operational hook. |
| Per-tree composition | `internal/cli/root.go:273-277` calls `Compose(app.authorizeMutation)` once and assigns the four named services and opener to that App. This is constructor wiring. `resolveFrom` at line 439 invokes the opener and validates its result before publishing configuration/service/layout at line 450. No local fallback or singleton exists in the reviewed production implementation. |
| All local binding implementations | `internal/appwiring/wiring.go:23-43` supplies all six hooks. `ReadRepository` calls `config.Discover` at line 30; `ReadUser` calls `userconfig.Load` at line 37. Missing planning is classified with `errors.Is(err, config.ErrNoConfig)` at line 26. Ordinary style reads user values in `internal/cli/root.go:118`; its warning path is separate from completion. |
| Persistence-family constructors | `internal/appwiring/wiring.go:47` constructs `spacestore.FS`; line 51 constructs `configstore.FS`; line 56 constructs `workspacestore.FS`. Each receives the supplied authorizer. The neighboring `core.New*Service` calls assemble use-case services; they are not persistence constructors. The configuration service also gets the same registry service and `design.Names()`. |
| Ordinary planning opening and identity reader | `internal/appwiring/wiring.go:64` discovers configuration; line 72 constructs one `store.FS`. Lines 72-78 install an apply-time rediscovery reader anchored at the observed checkout/root and the invocation authorizer. Line 79 calls `core.NewService(fs)`, which checks selected source sets (`internal/core/service.go:387`). That same FS is supplied as local layout at wiring line 83. |
| Later workspace opening | `internal/workspacestore/fs.go:34-45` discovers the requested entry and constructs a new FS at line 39 with its stored authorizer. `core.WorkspaceService.Open` validates source completeness and expected identity before constructing a service (`internal/core/workspace.go:78-134`). This opener still lacks the Thread-apply identity reader: independently reproduced known followup, not a new extraction regression. |
| Read-only overview child constructor | `internal/spacestore/fs.go:111-112` constructs `store.NewFS(root)` without an authorizer and returns only `core.PlanningSummarySource`. `internal/core/space_overview.go:263` consumes that read-only source. Do not interpret the four guarded family constructors as proof that every nested FS constructor is guarded; no mutating capability is exposed by this overview use case. This constructor also predates the extraction. |
| Mutation authorization reachability | `internal/cli/command_safety.go:27-59` binds the selected command and rejects unbound/read-only mutations. Ordinary pre-run binds before discovery at `internal/cli/root.go:346`; style-only commands bind at line 150. Storage authorization is consulted in `internal/store/fsstore.go:86`, `internal/store/fix.go:26`, `internal/configstore/fs.go:112,158`, and `internal/spacestore/fs.go:85,98`. Later workspace stores inherit the closure above. Denied planning writes, repair, config preferences, registry addition, and later-workspace writes were exercised in both dry-run and committing modes by the committed authorization regression. |
| Init topology exception | `internal/cli/init.go:54` authorizes before topology operations. Verified config uses: `Describe` at 64/187/306, `PendingMigrations` at 111, `Init` at 126/228, `AddTrackedRepo` at 236, `InitPointer` at 336, `LinkBack` at 344, plus typed topology receipts. `registerInitializedSpace` checks the injected registry at 403 and calls the application use case. This is a named topology exception, not planning-store construction. |
| Checkout receipt exception | `internal/cli/workspace.go:67` calls `config.DescribeCheckout`; line 75 uses `config.ConfigFile`. Root/identity/selection values come from neutral `App.Cfg` at lines 55-77. No planning discovery or store constructor remains in this file. |
| Every retained production config import | Exact source sites: `internal/appwiring/wiring.go:11`, `internal/cli/init.go:12`, `internal/cli/workspace.go:9`, `internal/configstore/fs.go:13`, `internal/spacestore/fs.go:11`, `internal/workspacestore/fs.go:7`, and `internal/spacehealth/diagnose.go:12`. CLI has only the two named exceptions. |
| UI gates and launch requests | `internal/cli/ui.go:56-77` checks authorization, dry-run, terminal gates, startup, and the launch hook, then sends `ports.Browser`. `internal/cli/config.go:48-76` owns config-editor gates and sends `ports.ConfigurationEditor`. Concrete options/execution are `internal/appwiring/launch.go:9-23`: configuration/workspace/overview/atlas-theme options, optional atlas landing, `tui.Run`, and `configui.Run`. M1 identifies the missing service validation on the UI path. |
| Other neutral consumers and missing-service guards | Theme/user/pager fields in `internal/cli/root.go:163-172,495-506`, `pager.go:80-106`, `ui.go:125-143`, and graph repair path normalization in `task_dependency_repair.go:108` use neutral values. Explicit service guards are at `config.go:54,97,126`, `doctor.go:55`, `space.go:48,110,128`, `status.go:53`, and `root.go:416,440,447`. Those commands fail with `domain.ErrValidation`; they do not construct adapters. |
| Completion | `internal/cli/root.go:268-271,351-354` defers planning resolution through `CompletionService`. `internal/cli/completion.go:19-24` handles a missing registry silently; entity completion uses the injected application service. Real opaque-target and populated-cwd completion probes verified no local fallback. |

### Lint policy inventory and actual probes

`.golangci.yml:118-147` covers direct/recursive CLI sources and excludes direct/recursive render, prompt, ports, the two named exception files, and tests. Its exact internal allowances are ports, prompt, render, core, design, domain, editor, graphfmt, id, and wire, all ending in `$`. Lines 152-169 reapply that policy to only `init.go`/`workspace.go`, adding exact `config$`. Lines 171-180 give direct/recursive neutral ports only exact core/design allowances.

The primary-adapter rule at lines 74-100 covers direct/recursive TUI, config UI, render, and prompt sources, excluding tests. Exact allowances are configui, core, design, domain, editor, listfilter, progressbar, theme, themepreview, and wire. The focused rule at lines 102-114 prevents config UI/render/prompt from importing configui; full TUI embedding is intentional. Domain/core/wire rules at lines 17-67 cover direct/recursive sources, excluding tests; config/home separation covers direct/recursive config sources at lines 188-197, including tests.

All final probes below were formatted and independently compiled with `go test ./... -run '^$'` (exit 0) before standard `just lint`. Scratch imports were restored after every probe. A preliminary scratch launcher formatting error was removed and the affected probes rerun; it is **not** counted as guard evidence.

| Compiler-valid import probe | Standard lint result |
| --- | --- |
| CLI → store, configstore, spacestore, workspacestore | Four separate exit-1 runs, each specifically `cli-controllers-use-injected-application (depguard)`. |
| CLI → appwiring; new internal/reviewadapter; core/reviewchild; config; tui; configui | Six separate exit-1 depguard rejections under the controller rule. The new adapter and allowed-package child were actual temporary packages. |
| New internal/cli/reviewchild source → store | Exit 1, controller depguard rule; recursive coverage proven. |
| New internal/tui/reviewchild source → store | Exit 1, `primary-adapters-use-application-seams (depguard)`. |
| CLI ports → store | Exit 1, `cli-composition-contracts-stay-neutral (depguard)`. |
| init.go → configstore; workspace.go → appwiring | Two separate exit-1 depguard rejections under `cli-local-topology-and-checkout`. |
| New TUI child → configui | Exit 0; intentional full-TUI embedding allowance. |
| CLI controller and CLI ports → Bubble Tea constructor value | Both exit 0, zero issues: M2. |
| init.go and workspace.go → Bubble Tea constructor value | Both exit 0 after stdin formatting; zero issues: M2 also covers the exceptions. |
| New core child → new domain/reviewallowed child | Exit 0. The foundational domain/core/wire allowances remain prefixes, unlike the tightened CLI/presentation allowances. The same prefix entries are visible in `git show ec5185e:.golangci.yml`; this is preexisting, not a newly introduced defect or proof of a shipped adapter. |
| Clean restoration | `just lint`: exit 0, `0 issues.` after probe batches and at final restoration. |

Fifteen distinct compiler-valid import cases were rejected specifically by depguard. Their cases are listed above; this is independent evidence, not a relabeling of the owner's claimed sixteen probes. Detailed local-only evidence is retained under the sandbox's `.git/review-evidence/`: `lint-probes.py`, `lint-results.json`, `lint-extra-probes.py`, `lint-extra-results.json`, `lint-framework-exceptions.py`, and `lint-framework-exceptions-results.json`.

### Command and compatibility evidence matrix

Tools: `go version go1.27.1 darwin/arm64`; `golangci-lint 2.14.0`, built with Go 1.27.1, revision `114493f`; `just 1.58.0`.

Initial literal runs of all four required commands were blocked by permissions on the default `/Users/andyeschbacher/Library/Caches/go-build`; these were setup failures, not code failures. Reran the same commands after:

```sh
export GOCACHE="$PWD/.git/review-cache/go"
export GOLANGCI_LINT_CACHE="$PWD/.git/review-cache/lint"
```

| Exact command or exercise | Observed outcome |
| --- | --- |
| `go test ./...` | Clean sandbox baseline exit 0, all packages pass. Restored final runs, including `go test ./... -count=1`, also exit 0. |
| `go test -race ./...` | Baseline exit 0, all packages pass, no race report. No production code changes were retained. |
| `just lint` | Baseline and restored final exit 0, zero issues. |
| `just docs-check` | Baseline and restored final exit 0; generation produces no CLI-reference diff. |
| `go run ./internal/tools/mangen -out .git/review-evidence/manpages` | Exit 0 using the metadata-only constructor; output remains sandbox-only. |
| `go build -o .git/review-evidence/tskflwctl ./cmd/tskflwctl` | Exit 0, fresh baseline binary. Go emitted denied self-version stat-cache writes under the default module cache, but produced the binary successfully; those warnings are not test failures. |
| `go test ./internal/cli ./internal/appwiring -run '^TestReview' -count=1 -v` | Three scratch hostile tests passed before removal. Failed explicit opener: actual local control candidates `alpha\nbeta\n:4\n`; command returns the sentinel with empty stdout; completion returns only `:4\n`, with no application warning. Cobra's own directive diagnostic is still present on stderr. |
| Non-default startup mapping | Scratch `TestReviewStartupMetadataNonDefaults` verifies scaffold/pointer/bare modes, ID/root/checkout/config path, tracked repos, theme, pager command, **false versus unset** pager pointer, detached copies, home preferences, and malformed home-config error/zero-values. `repositorySettings` performs value mapping, not migration/registry inspection. |
| Identity changed after projection | Scratch `TestReviewPointerReplacementAndWorkspaceParity` repoints a real pointer after composing a Thread plan. Both dry-run and committing applies return `domain.ErrConflict`, with no committed receipt. The committed direct-ID replacement regression also kills cached identity. |
| Machine compatibility | `git diff ec5185e -- internal/wire` has zero lines: this extraction adds no optional wire fields/branches. Baseline tests include the committed machine/projection goldens (`internal/cli/integration_golden_test.go:63,149`) and semantic schema validation (`internal/wire/envelopes_test.go:313`). The scratch non-default checks above target the newly mapped neutral values; no new wire coverage is falsely claimed. |
| Missing browser/editor hooks, real Cobra with terminal streams | The sandbox-only launcher uses `os.Stdin/Stdout/Stderr`. Removing `RunBrowser` returns exit/class 11 with `browser launch is unavailable from this invocation`; removing `RunConfiguration` returns 11 with the corresponding editor error. |
| Incomplete UI bundle, real Cobra with terminal streams | With Configuration/Workspaces/Overview nil, valid local planning opening, and a recording launch hook, `ui -C <actual repo>` prints `BROWSER_REACHED configuration_nil=true workspaces_nil=true overview_nil=true planning_nil=false layout_nil=false atlas=false`, then `RESULT exit=0 err=<nil>`: M1. |
| Fresh real binary terminal smoke | `tskflwctl -C <scratch repo> ui` rendered Overview; `config edit` rendered Configuration/About; ambient `ui` with a registered scratch space rendered `[atlas]` and one space. Each quit with `q`, exit 0. OSC background responses were supplied to the pseudo-terminal. This verifies launch and the non-default atlas-landing option, not every interactive mutation. |
| Every repair command emitted by the reviewed no-space startup case | In two isolated, repository-boundary scratch directories with separate empty homes, `ui` returned exit 10 and suggested `tskflwctl init` or `tskflwctl space add <path>`. Executed each exact remedy against its recommending state: init succeeded and subsequent UI reached the in-repo launcher; space add of the real scratch corpus succeeded and subsequent UI reached atlas with `LandOnAtlas=true`. No hypothetical candidate was substituted. |

Logs are sandbox-only, not committed fixtures: `baseline-0.log` through `baseline-3.log`, `final-tests.log`, `final-tests-uncached.log`, `final-lint.log`, `final-docs.log`, `hostile-runtime.log`, `remedy-init.log`, and `remedy-space-add.log`. Reproduction sources `review-cli_test.go`, `review-appwiring_test.go`, and `reviewlaunch.go` are retained in that evidence directory; their temporary production/test source copies were removed. Existing goldens at `internal/cli/testdata/golden/` are shipped fixtures; these review probes are not.

### Exact runtime mutations paired with regressions

Each row changed only sandbox source, executed the named test with `-count=1`, restored the source to the captured baseline, and reran that test successfully. Mutation diffs/logs and restoration logs remain in `.git/review-evidence/`. Compilation failures are not counted as kills.

| Exact mutation | Intended regression and result |
| --- | --- |
| Call `app.openPlanning("")` while constructing the controller tree | `TestCommandTreeConstructionDoesNotReadInvocationData`: **killed**, reads becomes 1. |
| Call `openPlanning(".", authorize)` inside real local `compose`, ignoring its result | Same construction regression: **survives** because its fake Compose replaces local wiring. |
| Resolve planning in the initial completion pre-run before completed flags | `TestInjectedPlanningOpenerRunsAfterCompletionFlagsAndSafety`: **killed**, starts include cwd and then opaque:entry instead of one target. |
| Replace the ordinary FS authorizer specifically with nil at wiring line 78 | `TestLocalCompositionPropagatesAuthorizationToEveryPersistenceFamily`: **killed**, operation 0 gets validation rather than the denial sentinel, zero authorizer calls. |
| Omit the configstore authorization option | Same persistence-family regression: **killed**, operation 2 returns nil instead of the sentinel. |
| Omit the spacestore authorization option | Same regression: **killed**, operation 3 returns nil instead of the sentinel. |
| Omit the workspacestore authorization option | Same regression: **killed**, operation 4 reaches ordinary validation rather than the denial sentinel. |
| Route every fresh bundle's policy through one overwritten package-level callback | `TestLocalCompositionKeepsAuthorizationInvocationScoped`: **killed**; after the second tree runs, a read-tree mutation is incorrectly accepted. Reused the same `LocalBindings` value across both real Cobra trees. |
| Select `LocalBindings` when Compose is absent | `TestMetadataOnlyCommandTreeAndMissingBindingsHaveNoLocalFallback`: **killed**; real populated cwd returns Alpha instead of validation failure. |
| Fill missing Configuration from local composition | `TestMissingNamedServicesFailWithoutPersistenceFallback`: **killed** across config/show/migrate/doctor cases. |
| Publish opener fields before checking error/completeness | `TestFailedPlanningOpenerDoesNotPublishPartialInvocation`: **killed** in both failed and incomplete cases. |
| Return cached root/ID from the identity-reader closure after rediscovery | `TestOpenedPlanningRechecksIdentityBeforeThreadApply`: **killed**; replaced ID wrongly accepts dry-run. |
| Change only `store.NewFS(cfg.Root, ...)` to `store.NewFS(".", ...)` | `TestLocalCompositionSelectsOneObservedDirectOrPointerCorpus`: **survives**, because the empty corpora and five-path count never inspect actual service data or actual watcher roots. Other tests can detect some wrong-root outcomes; this named regression does not. |
| Add a second ignored `config.Discover(start)` after successful discovery | Same direct/pointer regression: **survives**; matching metadata is not a discovery count. |
| Add an extra injected `ReadUser` call in ChromeTheme | `TestChromeThemeUsesOnlyExplicitPresentationReaders`: **killed**, user reads becomes 2. |
| Install local ReadRepository when ChromeTheme has none | Same chrome regression: **survives**, because the package cwd has no non-default repository theme in its fixture. |
| On explicit opener error, substitute `LocalBindings().Compose(a.authorizeMutation).OpenPlanning(actual cwd)` | Scratch `TestReviewFailedOpenerPopulatedCWD`: **killed in both Cobra command and completion subtests**. Actual control candidates were alpha/beta; mutant returns Alpha and `alpha\nbeta\n:4\n` respectively. Restoration passes. |

The planning-option probe initially matched a `spacestore` substring; that trial is not planning-family evidence. The exact ordinary-FS expression was then mutated separately (`planning-mutation.py`, `corrected-planning-mutation.json`) with the operation-0 result above.

The denial regression's task fixture lacks tags; removing planning/workspace authorization therefore reaches unrelated task validation before a write. That still kills the policy omission, but is not proof of a valid unauthorized write succeeding. The scratch identity tests use tagged valid tasks. The two-tree policy-exchange mutant does demonstrate an otherwise valid task-field operation accepted under the wrong invocation policy.

For the systemic pass, coordinated three mutations: early opening in real Compose, duplicate discovery in the local opener, and chrome's missing-reader local fallback. **`go test ./...` still exited 0, including an uncached `go test ./... -count=1` rerun.** Evidence: `fallback-and-systemic.py`, `systemic-survivor.diff`, `systemic-survivor.log`, `systemic-uncached.py`, `systemic-survivor-uncached.log`. Each violation lies beyond what the fake composition/read-count tests observe. All were then restored; no mutant is being shipped.

### Regression-delta review and systemic conclusions

Inspected all three new appwiring regressions and all seven new composition regressions, then executed their precise challenged mutations above. Existing CLI test deltas principally replace constructors with `newTestRootCmd`/`newTestRootCmdWithApp`/`newTestChromeTheme` (`internal/cli/composition_test.go:21-30`), which explicitly select binary wiring. They do not install a permissive authorizer, replace the selected target, or override launch hooks. Pager/theme/startup assertions map to neutral fields without weaker expected outcomes. `newUITestApp` at `internal/cli/ui_test.go:171` now gets real services from the same helper; its registry fixture mutation runs through a separately classified real command at line 121 rather than bypassing authorization on its unbound App. Startup helper tests still do not exercise the final launch hook or incomplete in-repo bundles: M1.

The factories are a finite named bundle, not a runtime key lookup or global service registry. Concrete constructors currently stay in wiring/secondary adapters; the neutral request definitions honestly describe the CLI's consumers. Reusing bindings retains independent bundles/policies, verified by the two-tree test and hostile policy exchange. There is no demonstrated need for a larger framework.

The controller's publication check is atomic on opener error/incompleteness; actual failed opener commands stay on their selected error rather than discovering populated cwd. Ordinary direct/pointer Thread apply retains the apply-time identity check. The local opener's source-set check comes from constructing one FS-backed service; a green metadata assertion alone is not evidence of the chosen service corpus or read count.

The boundary is therefore substantially real, but two fail-closed claims are incomplete: missing UI service validation depends on startup location (M1), and external Bubble Tea can bypass the import policy even in neutral contracts/exceptions (M2). The shared testing pattern misses real-wiring discovery and fallback violations (L1). Exact CLI/presentation package allowances and recursive globs work for the demonstrated internal cases. Exceptions remain file-specific, but depguard cannot restrict an allowed package to particular functions; current config call sites were inspected rather than inferred from lint alone. No unsupported objection to the bundle's style or a future unbounded architecture rewrite is counted as a finding.

### Verified scoped followups and limitations

- `planning/tasks/6gg7e594gcms-preserve-planning-identity-revalidation-in-workspace-opened-thread-apply.md` is still ready-to-start. Scratch direct Thread dry-run succeeded while workspace dry-run returned `validation failed: planning repository identity cannot be re-read for Thread apply`. `git show ec5185e:internal/workspacestore/fs.go` confirms the same missing identity-reader constructor before extraction. This fails closed and is not counted again.
- `planning/tasks/6gg7e59cyxxh-require-an-explicit-mutation-authorization-policy-at-persistence-composition.md` records the optional-policy hardening and a constructor compatibility decision checkpoint. The current nil-policy returns are visible at `store/fsstore.go:87-88`, `configstore/fs.go:40-41`, and `spacestore/fs.go:41-42`; workspacestore propagates nil into that FS option. This is existing residual work, not a claimed fix. The read-only overview child constructor noted in the inventory also prevents an unqualified “all constructors guarded” assertion.
- `planning/tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md` is a ready-to-start task to propose an ADR, with unchecked acceptance criteria; `docs/ARCHITECTURE.md:116` links this task. That is planned authority, not an accepted or already shipped policy ADR.
- The historical architecture audit's M1/L1 are labeled locally fixed, M4 fixed in PR #275, and M2/M3 tracked (`planning/audits/6gefs7wbbcyt-2026-09-28-arch-hexagonal-boundaries.md:316,377,443,503,565,657`). The new probes verify the relevant internal guard repairs, without closing that audit or changing its dispositions.
- Terminal smoke is available and passed, but no exhaustive interactive editing/terminal-layout review is claimed. Local-only probes/logs are preserved in the independent sandbox. No fixes, task lifecycle/AC changes, commits beyond the helper's baseline, staging, or edits to the other reviewer's audit were performed.

### Mandatory isolation attestation

Created with the repository's general `scripts/isolated-review-workspace.sh create --source "$SOURCE_ROOT" --deliverable "$AUDIT_REL" --print-path`, using an independent clone with sandbox-owned Git metadata and captured staged/unstaged/untracked/deleted state. All implementation inspection, project commands, probes, and report editing occurred there; the source was used only for the brief, initial workspace setup, and guarded transfer.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Vj8NtK
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.Vj8NtK/.git
baseline_commit=9a329e2a2ff63ff534c10b0042a7e5442bb85b92
source_blob=67fe69261b51959b08bc05428dd5d0fb6edc06f8
source_fingerprint=c9c375912de370587f42c02afe82b993301f5db4
deliverable=planning/audits/6gg7h65sfjrb-2026-10-03-cli-composition-boundary-implementation-codex.md
deliverable_changed=true
transfer=succeeded
```

The final helper verification checks unchanged baseline HEAD/index, independent Git metadata, no unrelated tracked/untracked deltas, nonempty report, clean audit diff, and unchanged source-deliverable blob. The transfer result above applies to the guarded atomic transfer of this audit only; on refusal the local report must remain untransferred and the conflict reported. The workspace and local evidence are retained until the implementation owner confirms receipt.

## Findings

#### M1. In-repo UI accepts missing operational services at the launch boundary · **Status:** fixed locally (2026-10-03)

**File:** internal/cli/ui.go:94 | **Component:** cli-composition

**Class:** current invocation-boundary failure for incomplete injection; the ordinary full LocalBindings binary supplies these services.

The new contract says missing operational services fail explicitly, but the same incomplete bundle fails or succeeds according to startup location. `internal/cli/ui.go:94-99` returns an in-repo runtime workspace without checking configuration/workspace/overview services. The launch at lines 72-76 forwards those nil pointers. Only the ambient atlas path checks some named services at lines 101-102. `internal/cli/composition_test.go:111-137` exercises missing named services and the ambient startup case, but omits an already opened in-repo UI.

Executable reproduction, using sandbox-only `.git/review-evidence/reviewlaunch.go`: take LocalBindings, wrap Compose to clear Configuration, Workspaces, and Overview, retain its valid local OpenPlanning, and install a recording RunBrowser. With terminal streams, run `REVIEW_MODE=incomplete TSKFLW_CONFIG_HOME=<sandbox home> <reviewlaunch> -C <real scratch repo> ui`.

Observed:
```text
BROWSER_REACHED configuration_nil=true workspaces_nil=true overview_nil=true planning_nil=false layout_nil=false atlas=false
RESULT exit=0 err=<nil>
```

Removing just the launch hook correctly returns validation exit 11, so the acceptance above is not a terminal-gate bypass in the harness. The real launcher forwards nil services to TUI options (`internal/appwiring/launch.go:11-13`); downstream configuration/workspace actions report unavailable (`internal/tui/command_dispatch.go:302`, `internal/tui/atlas.go:376`) rather than rejecting the partial CLI composition at launch. This makes the advertised failure boundary dependent on launch location and postpones service omissions into an already running browser.

**Impact:** new callers or a wiring regression can launch a partially functioning UI successfully despite the documented missing-service contract. No persistence bypass in the full shipped binary was demonstrated. Validate the UI's required named capabilities consistently before RunBrowser, and test both in-repo and atlas starts; if a partial UI is deliberately supported, make that explicit in the contract and test its promised behavior.

**Recommendation:** Validate the UI's required service bundle before either startup path reaches the launch hook, and cover incomplete in-repo and atlas invocations.

**Resolution:** Validate all four browser services before either landing route;
regression covers each omitted service in repo and atlas starts. Moving the
guard after repo landing now fails that regression.

#### M2. External Bubble Tea imports bypass controller and neutral-contract lint · **Status:** fixed locally (2026-10-03)

**File:** .golangci.yml:143 | **Component:** dependency-policy

**Class:** demonstrated enforcement gap for the extracted framework boundary.

The controller, exception, and neutral-port rules deny only the repository's internal package prefix (`.golangci.yml:143-144,169,180`). They have no denial for the actual external Bubble Tea framework. Consequently a controller or even the supposedly framework-neutral invocation contract can import it directly while standard lint remains green. The exception comment at lines 149-151 says those files do not permit UI framework launches, but the same escape works there.

Exact probe, separately placed in `internal/cli/reviewprobe.go` and `internal/cli/ports/reviewprobe.go`:
```go
package cli // use ports in the contract probe

import tea "charm.land/bubbletea/v2"

var ReviewFrameworkConstructor = tea.NewProgram
```

Format, run `go test ./... -run '^$'`, then `just lint`. Each command exits 0; standard lint prints `0 issues.`. Separately insert that import and exported constructor value into `init.go` and `workspace.go`, format with `golangci-lint fmt --stdin`, and repeat: compilation and standard lint also exit 0. Evidence is in sandbox-only `lint-framework-controller.log`, `lint-framework-ports.log`, `lint-framework-exceptions-results.json`, and the probe scripts. All probes were removed and clean lint rerun.

**Impact:** direct framework construction/types can re-enter controllers or their neutral contracts without the intended executable boundary failing. Internal tui/configui imports are correctly rejected, but that rejection does not protect against direct external framework use. The reviewed production contract currently stays neutral; this is a live guard omission, not a claim that it already contains Bubble Tea types. Add explicit framework denials for the controller/exception scopes and neutral contracts while preserving deliberately allowed presentation libraries; pin compiler-valid external framework probes.

**Recommendation:** Deny direct launch-framework dependencies in controllers, named exceptions, and neutral invocation contracts, with compiler-valid standard-lint probes.

**Resolution:** Deny direct Bubble Tea imports in controllers and both
exceptions; restrict invocation contracts to stdlib and exact core/design. Ten
compiler-valid import probes now produce ten depguard failures under standard
lint.

#### L1. Real-wiring discovery and fallback violations survive the regression suite · **Status:** fixed locally (2026-10-03)

**File:** internal/cli/composition_test.go:33 | **Component:** regression-coverage

**Class:** systemic test gap; no retained production mutant or assertion of existing extra discovery.

The new regressions observe injected fake hooks or metadata, not I/O and corpus choice inside real local wiring. `internal/cli/composition_test.go:33-57` supplies fake Compose, so adding an early `openPlanning(".", authorize)` inside the real `internal/appwiring/wiring.go:46` leaves it green. `internal/appwiring/wiring_test.go:16-34` opens empty direct/pointer corpora, compares repository metadata, and checks only that WatchPaths has five entries. Changing the actual FS root at wiring line 72 to `"."`, or adding duplicate discovery after line 67, leaves that precise regression green. `composition_test.go:183-201` also fails to catch a missing chrome reader falling back to local repository discovery because its missing-reader fixture has no non-default local theme.

Independently executed all four precise mutants, restored each, and reran its named test. Evidence is in sandbox-only `runtime-mutations.py`, `mutation-early-local-open.log`, `mutation-wrong-store-root.log`, `mutation-duplicate-discovery.log`, and `mutation-chrome-local-fallback.log`.

For the systemic second pass, coordinated early local opening, duplicate discovery, and chrome fallback in `fallback-and-systemic.py`; **the entire `go test ./...` suite exited 0, also with `-count=1`**. The coordinated diff/output are retained as `systemic-survivor.diff`/`systemic-survivor.log`. All production files were then restored and the clean suite/lint/docs checks passed. Wrong-root can fail neighboring tests; that does not make the named direct/pointer test sensitive to its claimed corpus guarantee.

**Impact:** a green suite gives false assurance about lazy real-wiring reads, one observed corpus, and missing-reader discovery boundaries. Add a bounded observation seam or hostile local fixtures that exercise actual LocalBindings: assert discovery count/laziness, distinguish actual service records and watcher roots from two corpora, and give missing-reader chrome a populated non-default-theme cwd. Require these exact mutations to fail before claiming those guarantees are regression-covered.

**Recommendation:** Exercise actual local wiring with read-count observation and distinguishable populated corpora, then require the documented surviving mutants to fail.

**Resolution:** Observe real startup readers and adapter composition; use
populated, distinguishable corpora and non-default chrome fixtures. All
documented surviving mutants now fail named regressions; failed-opener cwd
fallback is also rejected. See owner reconciliation for scope and restoration
evidence.

### Guarded audit-only count correction

The technical review and evidence above remain tied to the original Vj8NtK workspace and its captured baseline. After the first successful transfer, the summary count was corrected from sixteen to fifteen distinct depguard-rejected cases (the enumerated results already showed fifteen). This second independent clone was used only to correct this audit and record its guarded handoff; no new implementation review or code change is claimed. Original evidence remains in Vj8NtK. Both workspaces are retained until receipt.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.qErgQo
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.qErgQo/.git
baseline_commit=d5ccc29e31e1c836351dcf2686edffbb899010b5
source_blob=47bfc5ce990daba2d4da9bc4585ae7606e9e5277
source_fingerprint=a18434a90fcdfe178f03556d8a332fd09584aa24
deliverable=planning/audits/6gg7h65sfjrb-2026-10-03-cli-composition-boundary-implementation-codex.md
deliverable_changed=true
transfer=succeeded
```

## Owner reconciliation (2026-10-03)

The report above reviews its captured pre-fix snapshot. The owner accepted M1, M2, and L1;
their fixes are local to this composition slice, not evidence that the original report was wrong.

- M1: UI startup validates configuration, registry, workspace, and overview services before
  either repository or atlas landing. All four omissions are tested on both routes.
- M2: controller and named-exception rules reject direct Bubble Tea imports. Invocation contracts
  allow only the standard library and exact core/design packages. Ten formatted, compiler-valid
  probes produced ten specific depguard failures under standard `just lint`: direct/nested
  controller and ports framework imports, both named exceptions, Cobra/Lipgloss contract imports,
  and config/userconfig imports outside the local startup-reader boundary.
- L1: per-binding observation wraps real discovery/preference readers and actual adapter
  composition, without globals. Populated direct/pointer targets and a distinct populated cwd
  assert both service record identity and exact watcher roots. Chrome fallback fixtures have
  non-default repository/home themes; failed-opener command/completion fixtures have real cwd tasks.

In independent workspace `/private/tmp/isolated-review.VJp0PO` (baseline
`5c9a04aef6104b5719091eebb1ed59e983b1106d`), each of these runtime mutants failed its named
regression: moving UI validation below repository landing; eager real composition opening;
duplicate startup discovery; selecting a cwd store; implicit repository chrome fallback;
implicit home chrome fallback; and failed explicit opening falling back to populated cwd.
Authorization-loss probes also failed for the missing guard itself, with valid tagged task
fixtures rather than unrelated task validation. All probes were restored; focused tests and
standard lint passed again, and the isolation helper verified a clean sandbox with no task delta.

This observation contract covers declared startup readers, not every physical syscall inside a
secondary adapter. A production composition file cannot bypass it by importing config/userconfig
directly. Already tracked workspace identity parity, explicit authorization policy, and the policy
ADR remain separate followups; no additional task or design decision was needed here.

The later Antigravity report independently confirms the eager-discovery gap (its M1 overlaps L1).
Its original snapshot and review evidence remain historical; the owner reproduced the gap and
verified the new guard against it rather than treating corroboration as required for a fix.
