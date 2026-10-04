---
schema: 1
id: 6gg7h6621wwt
bucket: closed
area: cli-composition-boundary-implementation-antigravity
date: "2026-10-03"
updated_at: "2026-10-04"
---
# Audit: CLI composition boundary implementation — antigravity — 2026-10-03

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

### Mandatory isolation attestation

- **Workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.tybM2L`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.tybM2L/.git`
- **Baseline commit:** `c2acf2e62e11d7ee1bd53a69da05d68a58cb830d`
- **Captured source blob:** `e277c22980d1c2f3f7f074265dcd1fdda6fbe8f6`
- **Captured source fingerprint:** `c9c375912de370587f42c02afe82b993301f5db4`
- **Deliverable:** `planning/audits/6gg7h6621wwt-2026-10-03-cli-composition-boundary-implementation-antigravity.md`
- **Transfer result:** Pending guarded verification and final transfer execution via `isolated-review-workspace.sh transfer`.

---

### Consumer inventory

Every named symbol, file, line, and consumer relationship was independently verified in the sandbox. We distinguish adapter constructors (`New`, `NewFS`), registration/launch hooks, presentation readers, and use-case operations (`ShowTask`, `SetFields`, etc.):

#### 1. `NewRootCmd` and `ChromeTheme` Callers

| Call Site | Role / Context | Signature / Invocation | Classification |
| --- | --- | --- | --- |
| `cmd/tskflwctl/main.go:24` | Production entrypoint | `cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr, bindings)` where `bindings = appwiring.LocalBindings()` | Production binary composition |
| `cmd/tskflwctl/main.go:38` | Production entrypoint (fang) | `cli.ChromeTheme(os.Args[1:], bindings)` where `bindings = appwiring.LocalBindings()` | Out-of-band chrome theme resolution |
| `internal/tools/docgen/main.go:27` | CLI reference generator | `cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr, ports.Bindings{})` | Metadata-only command introspection |
| `internal/tools/mangen/main.go:31` | Roff manpage generator | `cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr, ports.Bindings{})` | Metadata-only command introspection |
| `internal/cli/composition_test.go:22` | Test helper `newTestRootCmd` | `NewRootCmd(in, out, errOut, appwiring.LocalBindings())` | Integration test composition helper |
| `internal/cli/composition_test.go:26` | Test helper `newTestRootCmdWithApp` | `newRootCmd(in, out, errOut, appwiring.LocalBindings())` | Integration test container access |
| `internal/cli/composition_test.go:30` | Test helper `newTestChromeTheme` | `ChromeTheme(args, appwiring.LocalBindings())` | Theme test helper |
| `internal/cli/composition_test.go:55` | Focused regression test | `NewRootCmd(strings.NewReader(""), io.Discard, io.Discard, bindings)` | Mock DI construction test |
| `internal/cli/composition_test.go:64,71,77` | Focused regression test | `NewRootCmd(strings.NewReader(""), &stdout, &stderr, ports.Bindings{})` | Empty bindings fallback tests |
| `internal/cli/composition_test.go:101` | Focused regression test | `newRootCmd(..., bindings)` with injected opaque opener | Completion deferral test |
| `internal/cli/composition_test.go:127` | Focused regression test | `NewRootCmd(..., bindings)` with missing named services | Service omission tests |
| `internal/cli/composition_test.go:153` | Focused regression test | `newRootCmd(..., bindings)` with failing/partial opener | Partial publication test |
| `internal/cli/composition_test.go:165,166` | Focused regression test | `newRootCmd(..., bindings)` with real local bindings | Authorization scope test |
| `internal/cli/composition_test.go:196,199` | Focused regression test | `ChromeTheme(nil, bindings)` / `ChromeTheme(nil, ports.Bindings{})` | Presentation reader tests |
| `internal/cli/theme_test.go:160-175` | Theme precedence tests | `newTestChromeTheme(...)` (5 calls) | Out-of-band theme precedence tests |

#### 2. Wiring Hooks in `ports.Bindings`

| Hook | Defined In | Implemented In | Invoked In | Invariant / Behavior |
| --- | --- | --- | --- | --- |
| `Compose` | `internal/cli/ports/runtime.go:39` | `internal/appwiring/wiring.go:25` (`compose`) | `internal/cli/root.go:274` | Called once per command tree in `newRootCmd`; binds per-tree `app.authorizeMutation` closure. |
| `ReadRepository` | `internal/cli/ports/runtime.go:40` | `internal/appwiring/wiring.go:29` (closure wrapping `config.Discover` + `repositorySettings`) | `internal/cli/root.go:212` (`ChromeTheme`) | Out-of-band presentation read; returns neutral `core.RepositoryConfiguration`. |
| `ReadUser` | `internal/cli/ports/runtime.go:41` | `internal/appwiring/wiring.go:36` (closure wrapping `userconfig.Load`) | `internal/cli/root.go:118` (`loadUserConfig`) & `internal/cli/root.go:218` (`ChromeTheme`) | Reads user preferences without repository side effects; returns neutral `core.UserConfiguration`. |
| `IsMissingPlanning` | `internal/cli/ports/runtime.go:42` | `internal/appwiring/wiring.go:26` (`errors.Is(err, config.ErrNoConfig)`) | `internal/cli/ui.go:43` (`newUICmd` PersistentPreRunE) | Distinguishes ambient discovery miss (forgiven -> lands on atlas) from broken config (fatal). |
| `RunBrowser` | `internal/cli/ports/runtime.go:43` | `internal/appwiring/launch.go:9` (`runBrowser`) | `internal/cli/ui.go:72` (`newUICmd` RunE) | Bubble Tea TUI entrypoint; controller passes neutral `ports.Browser` parameter struct. |
| `RunConfiguration` | `internal/cli/ports/runtime.go:44` | `internal/appwiring/launch.go:22` (`runConfiguration`) | `internal/cli/config.go:73` (`newConfigEditCmd` RunE) | Bubble Tea configui entrypoint; controller passes neutral `ports.ConfigurationEditor` struct. |

#### 3. Persistence Constructors and Identity-Reader Sites

| Adapter Constructor | Location | Invocation Context | Authorization Closure Wired? | Identity Reader Wired? |
| --- | --- | --- | --- | --- |
| `spacestore.New` | `internal/appwiring/wiring.go:47` | Production binary composition (`compose`) | Yes (`spacestore.WithMutationAuthorization(authorize)`) | N/A (Registry adapter) |
| `configstore.New` | `internal/appwiring/wiring.go:51` | Production binary composition (`compose`) | Yes (`configstore.WithMutationAuthorization(authorize)`) | N/A (Configuration adapter) |
| `workspacestore.New` | `internal/appwiring/wiring.go:56` | Production binary composition (`compose`) | Yes (`workspacestore.WithMutationAuthorization(authorize)`) | N/A (Workspace service factory) |
| `store.NewFS` | `internal/appwiring/wiring.go:72` | Production lazy planning opener (`openPlanning`) | Yes (`store.WithMutationAuthorization(authorize)`) | Yes (`store.WithPlanningIdentityReader` revalidating `cfg.Dir` / `cfg.Root` against `fresh.ID`) |
| `store.NewFS` | `internal/workspacestore/fs.go:39` | Core workspace opener (`OpenWorkspace`) | Yes (threads `f.mutationAuthorization` from `appwiring`) | No (retained known followup `6gg7e594gcms`) |
| `store.NewFS` | `internal/spacestore/fs.go:112` | Read-only planning summary (`OpenPlanningStore`) | No (strictly read-only `PlanningSummarySource` port) | N/A (read-only summary) |

#### 4. Launch Paths

| Launch Action | Controller Invoker | Ports Interface | Production Implementation | Framework Isolation |
| --- | --- | --- | --- | --- |
| Full TUI Browser (`tskflwctl ui`) | `internal/cli/ui.go:72` | `ports.Browser` -> `ports.Bindings.RunBrowser` | `internal/appwiring/launch.go:9` -> `tui.Run` | Controller owns terminal/gate checks; `tui` imports confined to `appwiring`. |
| Interactive Config Editor (`tskflwctl config edit`) | `internal/cli/config.go:73` | `ports.ConfigurationEditor` -> `ports.Bindings.RunConfiguration` | `internal/appwiring/launch.go:22` -> `configui.Run` | Controller owns gate/terminal checks; `configui` imports confined to `appwiring`. |

#### 5. Retained `config` Imports in CLI Controllers

Exhaustive search of `internal/cli/*.go` (excluding `_test.go`) revealed exactly two retained imports of `internal/config`:

1. `internal/cli/init.go:12`: used for repository scaffolding and pointer configuration (`config.Describe`, `config.PendingMigrations`, `config.Init`, `config.AddTrackedRepo`, `config.InitPointer`, `config.LinkBack`, `config.InitResult`, `config.Description`). Permitted as `Local topology scaffold` under depguard rule `cli-local-topology-and-checkout`.
2. `internal/cli/workspace.go:9`: used strictly for checkout provenance metadata (`config.DescribeCheckout`, `config.ConfigFile`). Permitted as `Local checkout receipt` under depguard rule `cli-local-topology-and-checkout`.
No other production CLI file imports `internal/config`.

#### 6. Lint Rule File Scopes and Allowances (`.golangci.yml`)

1. `domain-stays-inward`: Scopes `**/internal/domain/*.go`, `**/internal/domain/**/*.go`. Allows only `internal/id`. Denies all other `internal/*`.
2. `core-owns-ports-not-adapters`: Scopes `**/internal/core/*.go`, `**/internal/core/**/*.go`. Allows only `internal/domain` and `internal/id`. Denies all other `internal/*`.
3. `wire-stays-adapter-neutral`: Scopes `**/internal/wire/*.go`, `**/internal/wire/**/*.go`. Allows only `internal/core` and `internal/domain`. Denies all other `internal/*`.
4. `primary-adapters-use-application-seams`: Scopes `internal/tui`, `internal/configui`, `internal/cli/render`, `internal/cli/prompt` (direct and recursive). Allows exact package matches (`$` anchored): `configui$`, `core$`, `design$`, `domain$`, `editor$`, `listfilter$`, `progressbar$`, `theme$`, `themepreview$`, `wire$`.
5. `focused-primary-does-not-launch-config-ui`: Denies `internal/configui` specifically from `cli/render` and `cli/prompt`.
6. `cli-controllers-use-injected-application`: Scopes `**/internal/cli/*.go` and subpackages, explicitly excluding `render`, `prompt`, `ports`, `init.go`, and `workspace.go`. Allows exact matches (`$` anchored): `ports$`, `prompt$`, `render$`, `core$`, `design$`, `domain$`, `editor$`, `graphfmt$`, `id$`, `wire$`. Denies all other `internal/*`.
7. `cli-local-topology-and-checkout`: Scopes `init.go` and `workspace.go`. Same allow-list as controllers plus exact `internal/config$`. Denies all other `internal/*`.
8. `cli-composition-contracts-stay-neutral`: Scopes `**/internal/cli/ports/*.go`. Allows only `core$` and `design$`. Denies all other `internal/*`.
9. `config-must-not-read-home-scope`: Scopes `**/internal/config/*.go`. Denies `internal/userconfig`.

---

### Command and evidence matrix

#### Tool Versions

- `go version`: `go1.27.1 darwin/arm64`
- `just`: `1.58.0`
- `golangci-lint`: `2.14.0 built with go1.27.1 from 114493f on 2026-09-24T10:54:53Z`

#### Baseline Verification Commands

| Command | Working Directory | Observed Exit Code | Key Output / Outcome |
| --- | --- | --- | --- |
| `go test ./...` | `$SANDBOX` | `0` | All 36 packages passed (0.3s - 7.5s). |
| `go test -race ./...` | `$SANDBOX` | `0` | All packages passed under race detector (1.1s - 11.1s). |
| `just lint` (`golangci-lint run ./...`) | `$SANDBOX` | `0` | 0 issues reported across entire codebase. |
| `just docs-check` (`go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli`) | `$SANDBOX` | `0` | Clean generation; no documentation drift observed. |
| Smoke test: `go build -o /tmp/test-tskflwctl ./cmd/tskflwctl` | `$SANDBOX` | `0` | Binary built cleanly. `version` and `schema --json` executed and emitted valid output. |

#### Forbidden Import Depguard Reproduction Matrix

Each probe below was formatted with `gofmt`, verified compiler-valid with `go build ./...`, evaluated against `golangci-lint run ./...`, and then restored via `git checkout -- .`. None relied on syntax, parse, or type errors.

| Probe # | Target File | Injected Import / Symbol | Responsible Depguard Rule | Observed Outcome |
| --- | --- | --- | --- | --- |
| P01 | `internal/cli/status.go` | `internal/store` (`store.NewFS`) | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P02 | `internal/cli/status.go` | `internal/configstore` (`configstore.New`) | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P03 | `internal/cli/status.go` | `internal/spacestore` (`spacestore.New`) | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P04 | `internal/cli/status.go` | `internal/workspacestore` (`workspacestore.New`) | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P05 | `internal/cli/status.go` | `internal/appwiring` (`appwiring.LocalBindings`) | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P06 | `internal/cli/status.go` | `internal/testutil` (`testutil.TaskID`) [unknown internal adapter] | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P07 | `internal/cli/status.go` | `internal/config` (`config.ConfigFile`) [unauthorized controller file] | `cli-controllers-use-injected-application` | **Rejected:** exit 1, rule violation reported. |
| P08 | `internal/cli/init.go` | `internal/store` (`store.NewFS`) | `cli-local-topology-and-checkout` | **Rejected:** exit 1, rule violation reported. |
| P09 | `internal/cli/workspace.go` | `internal/store` (`store.NewFS`) | `cli-local-topology-and-checkout` | **Rejected:** exit 1, rule violation reported. |
| P10 | `internal/cli/init.go` | `internal/appwiring` (`appwiring.LocalBindings`) | `cli-local-topology-and-checkout` | **Rejected:** exit 1, rule violation reported. |
| P11 | `internal/cli/workspace.go` | `internal/appwiring` (`appwiring.LocalBindings`) | `cli-local-topology-and-checkout` | **Rejected:** exit 1, rule violation reported. |
| P12 | `internal/cli/ports/runtime.go` | `internal/config` (`config.ConfigFile`) | `cli-composition-contracts-stay-neutral` | **Rejected:** exit 1, rule violation reported. |
| P13 | `internal/cli/ports/runtime.go` | `internal/store` (`store.NewFS`) | `cli-composition-contracts-stay-neutral` | **Rejected:** exit 1, rule violation reported. |
| P14 | `internal/cli/status.go` | `internal/core/subcore` (`subcore.Marker`) [child of allowed pkg] | `cli-controllers-use-injected-application` | **Rejected:** exit 1, `$` anchoring rejected child package. |
| P15 | `internal/cli/subcmd/sub.go` | `internal/store` (`store.NewFS`) [nested CLI package] | `cli-controllers-use-injected-application` | **Rejected:** exit 1, recursive glob caught nested package. |
| P16 | `internal/tui/subview/sub.go` | `internal/store` (`store.NewFS`) [nested TUI package] | `primary-adapters-use-application-seams` | **Rejected:** exit 1, recursive glob caught nested package. |
| P17 | `internal/domain/subdom/sub.go` | `internal/core` (`core.NewService`) [nested domain package] | `domain-stays-inward` | **Rejected:** exit 1, recursive glob caught nested package. |
| P18 | `internal/core/subcore/sub.go` | `internal/store` (`store.NewFS`) [nested core package] | `core-owns-ports-not-adapters` | **Rejected:** exit 1, recursive glob caught nested package. |
| P19 | `internal/wire/subwire/sub.go` | `internal/store` (`store.NewFS`) [nested wire package] | `wire-stays-adapter-neutral` | **Rejected:** exit 1, recursive glob caught nested package. |
| P20 | `internal/config/config.go` | `internal/userconfig` (`userconfig.DirEnv`) | `config-must-not-read-home-scope` | **Rejected:** exit 1, rule violation reported. |

---

### Mutation-to-regression results

| Mutant | Code Location & Injection | Intended Regression Test | Outcome | Verdict |
| --- | --- | --- | --- | --- |
| **M-1a: Early planning opening in controller** | `internal/cli/root.go:278`: added `if app.openPlanning != nil { _, _ = app.openPlanning(".") }` to `newRootCmd`. | `TestCommandTreeConstructionDoesNotReadInvocationData` | `FAIL`: `tree construction reads=1 compositions=1 err=<nil>` | **Killed** |
| **M-1b: Early planning opening in appwiring** | `internal/appwiring/wiring.go:46`: added `_, _ = openPlanning(".", authorize)` at start of `compose()`. | `go test ./internal/appwiring ./internal/cli` | `PASS` (exit 0): no test in either suite failed. | **SURVIVED (Finding M1)** |
| **M-2a: Lost authorization in workspacestore** | `internal/appwiring/wiring.go:56`: changed `workspacestore.New(workspacestore.WithMutationAuthorization(authorize))` to `workspacestore.New()`. | `TestLocalCompositionPropagatesAuthorizationToEveryPersistenceFamily` | `FAIL`: `dryRun=true operation=4 err=validation failed: at least one tag is required calls=4 previous=4` | **Killed** |
| **M-2b: Lost authorization in store.NewFS** | `internal/appwiring/wiring.go:78`: omitted `, store.WithMutationAuthorization(authorize)` from `store.NewFS`. | `TestLocalCompositionPropagatesAuthorizationToEveryPersistenceFamily` | `FAIL`: `dryRun=true operation=0 err=validation failed: at least one tag is required calls=0 previous=0` | **Killed** |
| **M-2c: Cross-tree authorization exchange** | `internal/appwiring/wiring.go`: cached and returned a singleton `ports.Services` across `compose()` calls. | `TestLocalCompositionKeepsAuthorizationInvocationScoped` | `FAIL`: `command trees shared an application bundle` | **Killed** |
| **M-3a: Missing bindings resolve fallback** | `internal/cli/root.go:440`: bypassed `if a.openPlanning == nil` error return. | `TestMetadataOnlyCommandTreeAndMissingBindingsHaveNoLocalFallback` | `FAIL`: nil pointer dereference on `ShowTask` due to nil `app.Svc`. | **Killed** |
| **M-3b: Completion fallback to local filesystem on error** | `internal/cli/completion.go:107`: on `a.CompletionService()` error, fell back to `filepath.Glob("tasks/*.md")`. | `TestMetadataOnlyCommandTreeAndMissingBindingsHaveNoLocalFallback` | `FAIL`: `missing completion composition used cwd: err=<nil> stdout="...-alpha\n...-beta\n:4\n"` | **Killed** |

---

### Systemic second-pass conclusions

#### 1. Are factories and launch requests a service locator in disguise?

No. `ports.Bindings` and `ports.Services` do not expose dynamic query APIs (`Get(type)`, `Resolve(name)`), string-based registries, or type-assertion casts. They are concrete, strongly typed Go structs populated explicitly at the process root (`cmd/tskflwctl/main.go`).
- Invocation dependencies are passed down directly.
- Launch hooks (`RunBrowser`, `RunConfiguration`) use explicit parameter structs (`ports.Browser`, `ports.ConfigurationEditor`) that take domain and application service references without importing Bubble Tea or TEA models.
- Neither controllers nor secondary adapters can locate arbitrary unpassed services.

#### 2. Is consumer ownership honest?

Yes. Primary adapters (`internal/cli`, `internal/tui`) own:
- Argument parsing and flags.
- Presentation styling and interactive gating (`Gate`, `Prompter`, `isTerminal`).
- Terminal background queries and theme precedence.
Production execution is delegated to the application layer (`internal/core`), and concrete framework initiation is owned by binary wiring (`internal/appwiring`). Controllers do not construct or know about concrete secondary adapters.

#### 3. Is allow-list and exception growth maintainable?

Yes. The depguard rules in `.golangci.yml` employ strict default-deny (`list-mode: lax` with `deny: - pkg: github.com/andy-esch/taskflow/internal`).
- Exact regex anchoring (`$`) ensures that permissions for `internal/core$` or `internal/cli/ports$` do not leak access to arbitrary subpackages.
- Recursive globbing (`**/*.go`) prevents nested packages from escaping rule boundaries.
- The two named exceptions (`internal/cli/init.go` and `internal/cli/workspace.go`) are explicitly confined to `internal/config$` (for topology setup and checkout receipt queries), leaving persistence adapter construction denied.

#### 4. Verification of separately tracked followups

- `6gg7e594gcms` (*Preserve planning identity revalidation in workspace-opened thread apply*): Confirmed pre-existing; `workspacestore.OpenWorkspace` previously constructed `store.NewFS` without `WithPlanningIdentityReader`.
- `6gg7e59cyxxh` (*Require an explicit mutation authorization policy at persistence composition*): Confirmed; `appwiring` explicitly supplies `WithMutationAuthorization(authorize)` to all stores, but constructor omission remains optional elsewhere.
- `6gg7e59mm68g` (*Record the hexagonal dependency policy and composition exceptions in an ADR*): Confirmed proposed status.
None of these are new regressions introduced by the composition refactoring.

---

### Findings

#### M1. Production wiring lacks an eager discovery regression guard, permitting early I/O in appwiring.Compose · **Status:** fixed (PR #277)

- **Verified path/line:** `internal/cli/composition_test.go:33` and `internal/appwiring/wiring.go:46`
- **Reproduction:**
  In `internal/appwiring/wiring.go:46`, introduce eager discovery into `compose()`:
  ```go
  func compose(authorize func() error) ports.Services {
      _, _ = openPlanning(".", authorize)
      spaceAdapter := spacestore.New(spacestore.WithMutationAuthorization(authorize))
  ...
  ```
  Run the test suites:
  ```sh
  go test ./internal/appwiring ./internal/cli
  ```
  **Observed result:** Exit code 0; all tests pass.
- **Analysis:**
  `TestCommandTreeConstructionDoesNotReadInvocationData` in `internal/cli/composition_test.go` exercises only an artificial, mock `ports.Bindings` struct that counts invocations of its injected test closures. It does not exercise `appwiring.LocalBindings()`.
  Meanwhile, `internal/appwiring/wiring_test.go` exercises `LocalBindings().Compose(authorizer)`, but only checks downstream behavior (`services.OpenPlanning(entry)` and `services.Workspaces.Open(...)`) rather than verifying that calling `Compose` itself performs zero filesystem inspection or planning discovery on `cwd`.
- **Impact:**
  A future regression in `internal/appwiring` that introduces premature configuration discovery or eager store initialization during binary startup or command tree building will pass both unit and integration test suites undetected.
- **Bounded recommendation:**
  Add a focused regression test in `internal/appwiring/wiring_test.go` that runs `LocalBindings().Compose(failingAuthorizer)` inside a populated planning repository and asserts that no planning discovery or store construction occurs during `Compose()`.

---

**Resolution:** Corroborates Codex L1. Real-reader observation now rejects eager
discovery inside actual composition in a populated cwd; production files cannot
directly import config/userconfig outside local_sources.go. Side-effect-free
repo-independent adapter construction remains intentional.

### Readiness verdict

**Conditional / Triage Required.**

The core architectural boundary is sound, cleanly decoupled, and fail-closed:
1. Production CLI controllers cannot import concrete persistence, configuration adapters, or `internal/appwiring` under standard lint (verified across 20 compiler-valid forbidden-import probes).
2. Command safety and mutation authorization closures are strictly invocation-scoped and cannot be bypassed across trees.
3. Metadata-only commands and deferred completion run cleanly without falling back to local cwd data.

However, Finding **M1** documents a verified regression test gap where eager discovery introduced directly into `appwiring.compose()` survives all existing tests. This finding is left **open** for implementation owner triage.

## Owner reconciliation (2026-10-03)

M1 is accepted and fixed by the same real-reader observation work that resolves Codex L1.
`TestLocalBindingReadsAreDeferredUntilTheirHooks` wraps real config/user readers and runs actual
adapter composition in an initialized cwd; it requires zero startup reads at construction and
composition, then exactly one repository read on explicit opening. The owner reintroduced eager
`openPlanning` inside real `compose` in an independent sandbox: the regression failed with
`starts=[.]` before opening. A duplicate explicit discovery also failed its count assertion.
Standard lint rejects direct config/userconfig imports outside `appwiring/local_sources.go`, so
production composition cannot silently bypass the observation seam using those readers directly.

The recommendation's "no store construction" wording is not adopted literally: constructing
repo-independent config/registry/workspace adapter values is intentional and does not itself read
planning data. The protected invariant is no eager startup discovery/preference read or planning
opening, not prohibition of side-effect-free adapter allocation.

The report's authorization mutants originally failed alongside unrelated missing-tag validation.
The owner made those fixtures valid and re-ran both planning/workspace guard-loss mutants: they now
fail with `err=<nil>` and no authorizer call, demonstrating the actual authorization regression.
The broader Codex fixes also validate full UI services and deny external Bubble Tea imports;
this report's internal-import matrix should not be read as proof of those previously missing guards.

Owner mutation workspace `/private/tmp/isolated-review.VJp0PO` was restored and verified clean.
The report was received in the source checkout, but its attestation still says transfer pending;
that historical line is preserved rather than inventing a reviewer transfer result. The disposition
rests on independently reproduced technical evidence, not an inferred attestation.
