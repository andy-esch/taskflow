---
schema: 1
id: 6gj70e4p18b3
bucket: closed
area: go-toolchain-coordination-implementation-codex
date: "2026-10-09"
updated_at: "2026-10-10"
---
# Audit: Go toolchain and linter coordination — codex — 2026-10-09

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

Challenge whether this Go-version policy prevents the compiler/linter skew that has
previously broken CI while leaving minimum support and prompt security-patch updates
independent. Configuration agreement is not compiler compatibility or patch freshness.
Checked planning ACs and a green suite are hypotheses to challenge, not a verdict.

Codex: primary lens is executable policy and the real compiler/linter preflight;
cross-check Renovate's grouping. Antigravity: primary lens is actual Renovate
extraction, precedence, and update paths; cross-check compiler compatibility. Each
must perform a separate systemic pass and avoid reading the sibling report until
its own report is final. No finding quota: demonstrated negatives are useful.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`.
Branch: `chore/coordinate-go-toolchains`.
Base: v0.23.0, `ee54b930a9bb673c53882622546bd851a2355eae`.
Capture staged, unstaged, untracked, and deleted work with the mandatory sandbox
helper; reviewing HEAD alone misses this implementation.

Implementing task:
[Go coordination](../tasks/6gd68x5gqkk0-reconcile-go-toolchain-versioning-across-go.mod-ci-linters-and-renovate.md).
The release task update only records publication evidence; do not edit GitHub releases.
Read the release guide's new policy rather than interpreting the minimum as a pin.

Verify a consumer inventory for:

- `go.mod`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`,
  `build/release-validation/Containerfile`, and `renovate.json`.
- `Justfile`: `toolchain-check`, `renovate-toolchain-check`, `vulncheck`;
  `scripts/release-validate.sh`: `check_tools` and the new policy phase.
- `internal/tools/releasevalidate/toolchain_policy_test.go`:
  `checkToolchainPolicy`, `checkWorkflowToolchains`, `TestRepositoryToolchainPolicy`,
  `TestToolchainPolicyRejectsDrift`,
  `TestToolchainPolicyAllowsNewerReleaseLineWithoutRaisingMinimum`, and
  `TestToolchainPolicyAllowsIndependentPatchRefresh`.
- `internal/tools/releasevalidate/release_validate_test.go`:
  `TestReleaseValidateSuccessAndFailureBoundaries`, including the newer-active-line,
  unknown-build-Go, and policy-failure subtests.
- `internal/tools/releasevalidate/renovate_policy_check.mjs`: real manager extraction
  and `applyPackageRules`, not a handwritten matcher.

No application runtime, machine revision, dependency upgrade, or new version manifest
is claimed. Deliberately excluded: moving the release line to 1.27, automatic merging,
remote patch-freshness certification, and a full hosted Renovate run. These boundaries
may be challenged only with a concrete consequence, not a preference for more tooling.

## Intended contract to challenge

1. `go.mod` owns the minimum; CI tests its newest patch. The release workflow owns
   the compiler line; CI lint and the container follow it. A newer release/local
   compiler must not require a higher module minimum. Container Go is an exact
   stable patch at least as new as that minimum; workflows select latest patches.
2. CI and the container select the same linter release line, not necessarily patch.
   The actual release-preflight linter must be v2 and built with Go covering the
   module minimum AND active compiler line. Active patch identity is not required.
   Unknown build metadata refuses. Actual lint still establishes compatibility.
3. The offline policy check is reached from both CI and the shared release gate.
   Refused toolchain/policy checks stop later gates. Cache keys use setup-go's
   resolved compiler version, not a constant or the unchanged module directive.
4. Renovate really discovers the Containerfile and the annotated linter ARG.
   Specific Go/linter grouping wins over broad GitHub Actions grouping.
   Minimum/line upgrades require approval; toolchain patches/digests have a separate
   unscheduled, zero-age, non-automerge path. Minimum patch upgrades remain deliberate.
5. The optional actual-engine check covers repository extraction and custom rule
   application, not inherited presets, live lookup/classification, or bot scheduling.
   Documentation must not overstate those checks or promise instant CVE updates.

## Mandatory evidence floor

Both reviewers independently run focused Go tests and `just toolchain-check` in
their sandbox. Probe a valid newer release line while retaining minimum-line tests,
a compatible newer host compiler, and an incompatible linter. Evaluate raw script
outputs/exits and actual configuration, not merely helper return values. Do not run
the clean release gate against a dirty sandbox and call its cleanliness refusal a
toolchain reproduction. Existing synthetic gate fixtures are available for this.

Codex: perform one compiler-valid mutation removing only the active-Go-line linter
comparison in `check_tools`. The exact intended kill is
`TestReleaseValidateSuccessAndFailureBoundaries/linter_must_cover_newer_active_Go_line`.
It must fail for its compatibility assertion, not a syntax/build error; restore and
rerun green. Also challenge whether malformed/competing selectors escape the actual
offline guard and whether newer release lines unnecessarily raise the minimum.

Antigravity: run the actual Renovate-engine check with an installed package, then
perform a JSON-valid mutation moving broad GitHub Actions grouping after the specific
rules. The intended engine assertion must fail; restore and rerun green. Independently
inspect extracted manager/depName/depType/currentValue for Go and both linter pins.
Do not infer manager behavior from documentation or use a local imitation as proof.
Settle whether partial workflow selectors classify updates differently from exact
Docker/ARG pins; separate engine-tested rule matching from untested remote lookup.

Record precise edit, command, expected/observed failure, killed/survived/invalid, and
restored result. Each reviewer cross-checks one hostile case in the other lens. If
Renovate/containers/network are unavailable, name the untested claims explicitly;
do not manufacture a clean result or install tools globally in the source checkout.

## Required hostile angles

Give reproduced/falsified/unresolved outcomes for these concrete hypotheses:

- A valid but unexpected workflow/job layout passes while bypassing linter/Go setup,
  or fixed job names turn an ordinary extension into an unexplained false refusal.
- The Go minimum is implicitly raised to solve release tooling needs, or a patch
  refresh is wrongly treated as a line change. Consider raised minimum patches,
  absent pins, floating tags, digest-qualified images, and missing build metadata.
- The runtime preflight derives the wrong active line, admits an unsupported linter,
  or incorrectly refuses a compatible binary. Clearly bound prerelease/devel support.
- Renovate's manager discovery or effective rule precedence defeats the proposed
  grouping. In particular, examine `separateMinorPatch`, approval inheritance,
  `matchUpdateTypes`, linter semver-partial selectors, and the custom regex manager.
- The cached Go-built scanner remains compiled with a stale compiler after a setup-go
  patch refresh, or a claimed security fast path still silently inherits a delay.
- Tests are self-confirming: they assert configuration labels but skip actual
  extraction, version classification, or the production caller. Report coverage gaps
  distinctly from reachable defects; inherited presets are not claimed tested.

Second pass: look beyond each changed file for a systemic policy drift, a misleading
single-source-of-truth claim, or a validation seam that only appears fail-closed.
Explain one attempted counterexample even if none reproduces. Do not create a Go
support matrix from guessed linter version numbers.

## Validation and restoration

Owner evidence, which may be inherited rather than redundantly rerunning every gate:
full race suite, host lint/build/tidiness, focused tests, planning lint, strict Renovate
44.77.0 validation, and a built container with real Linux tests/lint. Host Go 1.27.2
uses linter 2.14.0 built Go 1.27.1. Container Go/linter build Go 1.26.9 uses linter
2.13.0. These are qualification checks, not a full clean-candidate gate on this dirty
implementation. Verify the implementing task for exact evidence and limitations.

Focused commands (run only in the independent sandbox):

```sh
go test ./internal/tools/releasevalidate -count=1
just toolchain-check
just renovate-toolchain-check /path/to/node_modules/renovate
node /path/to/node_modules/renovate/dist/config-validator.js --strict renovate.json
```

An existing read-only local package is available at
`/Users/andyeschbacher/.npm/_npx/80ce72c594ca058a/node_modules/renovate` (44.77.0).
Its internal module paths are intentionally version-sensitive; treat a different
installed layout as an environment limit or diagnose it, not a runtime dependency.
Use independent caches/scratch outside the sandbox's Git tree. Container checks may
inherit owner evidence; if repeated, never bind the source checkout writable.

Restore all probes to the sandbox baseline, keep only the assigned audit changed,
and use the injected helper's verify/transfer protocol. No pushes, source writes,
or commits beyond the helper's sandbox-only baseline checkpoint. Retain the sandbox
and report the blocker if verification fails. Never inspect the sibling report.

## Deliverable

Preserve this brief and complete the reviewer report with verdict/limits, verified
inventory, findings first, concise probe/mutation ledger, and a bounded surviving-
hypothesis list. Every finding needs a reachable root cause, consequence, exact
reproduction, and recommendation; distinguish a defect from defense-in-depth or
coverage. Use tool-owned finding creation in the sandbox where possible, leaving
all dispositions open for owner triage. No placeholders or settled verdict in advance.

Include the required isolation attestation: independent workspace/Git directory,
baseline, captured deliverable fingerprint, verification, and guarded transfer outcome.

## Reviewer report

Verdict: **changes requested**. Both passes are complete: (1) the assigned executable-policy/runtime-preflight checklist and Renovate cross-check; (2) a separate adversarial review of configured selectors versus executing toolchains, generated update values versus accepted policy, and abnormal shell exits. Four medium findings remain **open** for owner triage. No implementation fixes were made.

Evidence is bounded to the captured sandbox, local production callers, installed host tools, clean synthetic release fixtures, and Renovate 44.77.0's actual extraction/versioning/rule modules. This is not hosted CI/Renovate execution, remote freshness certification, or a complete clean-candidate release gate. The sibling report was not read.

### Findings

#### M1. A module toolchain directive silently replaces the minimum-line compiler · **Status:** fixed

**File:** internal/tools/releasevalidate/toolchain_policy_test.go:133 | **Component:** tooling

**Root cause and consequence.** `checkToolchainPolicy` reads only the `go` directive (`toolchain_policy_test.go:129–137`). It ignores Go's independently selecting `toolchain` directive. Neither minimum CI nor the release workflow sets `GOTOOLCHAIN=local` (`ci.yml:97–108`, `release.yml:32–48`). A normal module edit can therefore leave all checked selectors on 1.26 while the actual commands automatically use 1.27. Minimum-line compiler coverage is lost, release/container agreement becomes false, and setup-go's scanner-cache identity can differ from the compiler that builds the scanner. This is a drift defect, not an objection to a developer deliberately using newer local Go.

**Exact reproduction.** Keep `go 1.26.0` and the actual workflow/container pins unchanged; add `toolchain go1.27.2` to the sandbox `go.mod`. Run `just toolchain-check`: exit 0, reporting `minimum/test Go 1.26.0; release/lint line 1.26; container Go 1.26.9`. Invoke the installed `/Users/andyeschbacher/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.9.darwin-arm64/bin/go env GOVERSION` with `GOTOOLCHAIN=auto`, `GOPROXY=off`, and a scratch PATH entry named `go1.27.2` pointing to the real installed Go 1.27.2 binary. It prints `go1.27.2`; the same launcher with the baseline module prints `go1.26.9`. This uses real Go selection, without a version-output stub or download. The directive was removed by restoring the captured baseline.

**Evidence.** `module-toolchain-policy.log` and `module-toolchain-selection.log` in the retained evidence directory contain the raw gate and selected version. The executed compiler behavior defeats the advertised selector agreement while the module minimum remains 1.26.0.

**Recommendation:** Reject or model module toolchain directives that override coordinated CI/release selection; explicitly enforce the minimum compiler in CI without raising the module minimum, and test actual Go selection.

**Resolution:** CI, release, and container enforce GOTOOLCHAIN=local before
    tools run; policy rejects missing/overridden settings.
    TestLocalGoSelectionIgnoresModuleToolchainSuggestion executes real Go and
    confirms a future module suggestion cannot switch the selected compiler.
    Focused host and pinned-container tests pass.

#### M2. Renovate generates linter pins that the offline guard rejects · **Status:** fixed

**File:** renovate.json:80 | **Component:** tooling

**Root cause and consequence.** The linter grouping rule (`renovate.json:79–84`) leaves the workflow dependency on its default `semver-coerced` versioning. Actual GitHub Actions extraction gives `v2.13` no explicit versioning, whereas the container regex dependency uses `semver`. Renovate converts the workflow's partial selector to an exact update value. `checkWorkflowToolchains` requires the selector to equal only the two-component container line (`toolchain_policy_test.go:228–232`), so coordinated bot edits fail the new CI policy gate. Grouping both dependencies is therefore insufficient to make the update path usable.

**Exact reproduction.** With the installed Renovate 44.77.0 package, extract the real `ci.yml` linter dependency, apply the repository package rules, resolve its default versioning and manager range strategy, and call Renovate's `generateUpdate` with local release candidates `v2.13.1` and `v2.14.0` (current resolved release `v2.13.0`). Actual generated values are respectively `v2.13.1`/patch and `v2.14.0`/minor, both grouped as `go lint tooling` with approval true. For each value, set the real CI linter `version` and container `GOLANGCI_LINT_VERSION` to that exact same value, then run `just toolchain-check`. Both exit 1: `select linter line v2.13 ... got "v2.13.1"` or `select linter line v2.14 ... got "v2.14.0"`. Changing only the CI selector back to `v2.13` or `v2.14` makes the same coordinated configuration pass, exit 0. Restore both files afterward.

**Evidence and bound.** Retained `renovate-generated-updates.mjs`, `.json`, and `renovate-generated-policy.json` use actual installed `generateUpdate`, `getDefaultVersioning`, `getRangeStrategy`, and `applyPackageRules`, not a handwritten classifier. Candidates are local replay data; no claim is made about live releases or full hosted PR orchestration. The shipped optional check supplies `updateType` directly and never applies generated replacements (`renovate_policy_check.mjs:41–69`), explaining its green result despite this reachable update defect.

**Recommendation:** Choose Renovate versioning that preserves the required workflow line selector, or accept compatible exact workflow pins; validate actual generated replacements against the offline gate.

**Resolution:** Accept stable exact CI linter pins on the container linter line.
    TestToolchainPolicyAcceptsGeneratedLinterPins covers paired patch/minor
    edits; the Renovate check now resolves actual versioning/range strategy,
    generates replacements, and supplies them to that policy helper. Renovate
    44.77.0 local replay passes; no hosted lookup is claimed.

#### M3. Development compiler metadata can abort the release gate with exit zero · **Status:** fixed

**File:** scripts/release-validate.sh:13 | **Component:** tooling

**Root cause and consequence.** `version_at_least` feeds unvalidated `GOVERSION` components into Bash arithmetic (`release-validate.sh:13–25`). A development compiler's major component becomes `devel go1`, producing a nounset arithmetic abort. In the complete production script with its EXIT cleanup installed (`:107–113`, `:132–138`), macOS Bash 3.2.57 returns process status **0** after that abort. The compatibility, policy, race, lint, vulnerability, and snapshot gates have not succeeded or even run. A caller relying on command status can accept an unqualified candidate. This requires reliable refusal of unsupported metadata; it does not require supporting development compilers.

**Exact reproduction.** Run the unmodified complete `release-validate.sh` through the existing clean `newReleaseValidationFixture` and `/bin/bash` 3.2.57. For this one case, change the fixture Go stub's `printf 'go%s\n'` to `printf '%s\n'` and replay exact raw `GOVERSION` `devel go1.28-deadbeef`; retain linter build Go 1.27.1. The process prints only `==> toolchain compatibility` and `line 23: go1: unbound variable`, returns exit 0 (`err=<nil>`), and logs only `go env GOVERSION`. There is no success banner and no later gate invocation. A temporary test requiring a nonzero refusal compiles and fails on that assertion. The temporary file was archived outside the Git tree and removed.

**Scope.** This parser/cleanup failure already exists in the v0.23.0 base; the required hostile runtime probe exposed an existing release-gate boundary. The coordination diff adds the active-line check and policy phase but leaves that boundary unchanged.

**Evidence and bound.** `exact-devel-metadata.log` and archived `reviewer_runtime_probe_test.go` retain the raw output, command log, and assertion. The installed Go source documents this development metadata form at `/opt/homebrew/Cellar/go/1.27.2/libexec/src/cmd/dist/buildtool.go:148`. No actual development compiler was installed. A function-prefix-only probe exits 1; the full production script/cleanup path is essential to reproducing exit 0. A `go1.28rc1` replay refuses with exit 1 and arithmetic diagnostics, so prerelease support is also unqualified. Stable newer compiler and linter combinations pass.

**Recommendation:** Validate or explicitly reject non-stable GOVERSION metadata before arithmetic, preserve failure status through cleanup, and test raw full-script exits for unsupported versions on supported shells.

**Resolution:** Validate stable numeric compiler metadata before arithmetic and
    preserve the original EXIT status through cleanup. Full production-script
    regressions replay exact development/prerelease/malformed metadata, require
    nonzero refusal, and prove later gates never run. Host Bash 3.2 and pinned
    Linux container tests pass.

#### M4. Workflow step counts do not bind lint execution to its Go setup · **Status:** fixed

**File:** internal/tools/releasevalidate/toolchain_policy_test.go:228 | **Component:** tooling

**Root cause and consequence.** The workflow projection contains only `uses` and a few `with` fields (`toolchain_policy_test.go:107–119`). The guard counts every configured setup-go step regardless of its execution condition and counts linter actions globally (`:218–243`). It never requires the executing linter to share a job with the validated release-line compiler. Thus a valid workflow extension can pass the policy while using the runner's uncontrolled compiler. The baseline workflow happens to place both steps correctly; the newly introduced regression guard does not protect that relationship.

**Exact reproduction.** Two independent edits of actual `ci.yml` were tested with `just toolchain-check`: (a) add `if: false` to the existing test job's setup-go step, preserving its selector/latest-patch fields; (b) remove the entire golangci-lint action block from `lint`, leave that job's setup-go/policy steps, and append an `extra-lint` job containing checkout and the same linter action/version but no setup-go. Each returns exit 0. Case (a) no longer establishes minimum-line compiler selection; case (b) no longer establishes the release-line compiler for lint. Both edits were restored. These are policy false greens, not claims that a hosted run was executed.

**Related false refusal.** In an otherwise valid newer-release 1.27 configuration with the minimum test job still on 1.26, adding a second minimum-line race job named `minimum-race` is refused because every job except the literal name `test` is treated as a release-line consumer (`:214–216`). This exposes the same missing role model; it is recorded as an extension limitation rather than a separate finding.

**Evidence.** `policy-probes.json` retains each exact YAML edit, output, and exit. The guard still refuses duplicate/competing selectors, so the defect is specifically the lost execution/job relationship, rather than a blanket failure to parse YAML.

**Recommendation:** Require compiler setup in each protected consumer job and reject skipped setup steps; model minimum versus release consumers explicitly, or fail clearly on layouts outside a documented supported shape.

**Resolution:** Bind protected workflow consumers to earlier unconditional setup
    in the same job and validate inherited/overridden GOTOOLCHAIN settings.
    Regressions cover skipped minimum setup, relocated or early lint, and
    publisher conditions/overrides. An additional minimum-line job remains valid
    under a newer release line. Supported static roles and interpretation limits
    are documented in the task.

### Verified consumer inventory

All locations below were read in the captured sandbox; line numbers refer to that baseline. Planning acceptance criteria were treated as claims to challenge.

| Consumer or symbol | Verified shipped behavior |
| --- | --- |
| `go.mod:3` | Minimum `go 1.26.0`; no baseline `toolchain` directive. |
| `.github/workflows/ci.yml:25`, `:63`, `:97` | Lint/test setup-go `1.26`, latest true; CI linter `v2.13`; policy invoked at `:43`; actual tests at `:108`. |
| `.github/workflows/release.yml:32` | Publication job `goreleaser` selects latest `1.26`; GoReleaser action invoked at `:42`. |
| `build/release-validation/Containerfile:1`, `:4`, `:11` | Exact Go `1.26.9-bookworm`, annotated linter ARG `v2.13.0`, compiler-built tool installation. |
| `Justfile:63`, `:106`, `:110` | `vulncheck` uses package scan; `toolchain-check` invokes actual repository policy test; `renovate-toolchain-check` invokes the installed-engine script. |
| `scripts/release-validate.sh:45`, `:58`, `:61`, `:139` | `check_tools` checks real GOVERSION, v2 linter minimum and active-line build metadata, then the policy phase runs before package/race/lint/snapshot gates. |
| `internal/tools/releasevalidate/toolchain_policy_test.go:19`, `:31`, `:70`, `:91` | `TestRepositoryToolchainPolicy`, `TestToolchainPolicyRejectsDrift`, `TestToolchainPolicyAllowsNewerReleaseLineWithoutRaisingMinimum`, `TestToolchainPolicyAllowsIndependentPatchRefresh` are executable tests, not planned evidence. |
| Same file `:128`, `:201`, `:259` | `checkToolchainPolicy`, `checkWorkflowToolchains`, `toolchainPolicyFixture` are present; the fixture projects a minimal workflow shape. |
| `internal/tools/releasevalidate/release_validate_test.go:12`, `:87`, `:96`, `:108`, `:151`, `:255` | `TestReleaseValidateSuccessAndFailureBoundaries`, active-line/unknown-build/policy-failure subtests, `newReleaseValidationFixture`, and `fakeGoCommand` are present. The clean fixture executes the production script but stubs tools, including the policy Go command. |
| `internal/tools/releasevalidate/renovate_policy_check.mjs:17–21`, `:26–42` | Actual installed `applyPackageRules` and gomod/actions/docker/regex extraction; checks custom rules with supplied update types. |
| `renovate.json:17`, `:46`, `:51`, `:59`, `:66`, `:79` | Annotated ARG discovery, broad actions rule before specific Go/linter rules, minimum bump strategy, patch/digest/pinDigest fast path, reviewed linter group. |
| `.github/workflows/ci.yml:73`, `:77`, `:87`; `docs/RELEASING.md:29–68` | Scanner cache includes setup-go's resolved output; compiler-built scanner install and package scan; guide bounds policy and engine claims. |

Actual Renovate 44.77.0 extraction (`renovate-inventory.jsonl`) returned:

| Manager / dependency type | File / dependency / current value | Versioning and effective custom patch policy |
| --- | --- | --- |
| `gomod` / `golang` | `go.mod`, `go`, `1.26.0` | `go-mod-directive`; `go version`, approval true, bump, weekly/3-day policy. |
| `github-actions` / `uses-with` | Both CI Go selectors and release selector, `go`, `1.26` | `npm`; `go toolchain patches`, approval false, at any time/0 days, automerge false, separateMinorPatch true. |
| `github-actions` / `uses-with` | CI `golangci/golangci-lint`, `v2.13` | Extraction has no explicit versioning; actual datasource default is `semver-coerced`; `go lint tooling`, approval true, weekly/3-day policy. |
| `dockerfile` / `final` | Containerfile, `golang`, `1.26.9-bookworm` | Actual datasource default `docker`; Go patch fast path above. |
| `custom.regex` / no `depType` emitted | Containerfile ARG, `golangci/golangci-lint`, `v2.13.0` | Explicit `semver`; reviewed linter group above. |

Actual local `generateUpdate` settles the partial/exact distinction: Go workflow `1.26` plus candidate `1.26.10` classifies patch but keeps `newValue=1.26` (no textual selector bump), while candidate `1.27.0` classifies minor and generates `1.27`. Exact Docker candidates `1.26.10-bookworm` / `1.27.2-bookworm` classify patch/minor and retain exact new pins. Both linter dependencies generate exact candidate versions and classify `v2.13.1` as patch / `v2.14.0` as minor. These results are recorded in `renovate-generated-updates.json` and `renovate-generated-container-updates.json`; they are local engine replay, not live lookup.

### Probe and mutation ledger

Evidence directory: `/private/tmp/taskflow-go-audit.ww4WUN`. Initial Go tests used an independent build/temp cache and the existing module cache for reads; the final restored run uses an independent copied module cache too, with `GOPROXY=off`. Scratch, linter, and Renovate caches are outside the Git tree. No global tools were installed.

| Probe / exact action or command | Expected and observed result / disposition |
| --- | --- |
| `go test ./internal/tools/releasevalidate -count=1`; `just toolchain-check` | Exit 0 initially and after all probes. Final receipts: `focused-restored.log`, `toolchain-restored.log`. |
| Real host `go version`, `golangci-lint version`, exact production function prefix through `check_tools`, then `golangci-lint run ./...` | Go 1.27.2 / lint 2.14.0 built Go 1.27.1; compatibility exit 0 and actual lint exit 0, `0 issues.` (`actual-preflight.log`, `host-lint.log`). The prefix intentionally avoids a dirty-candidate cleanliness stop. |
| Clean production-script fixture: Go 1.27.2 / linter build 1.26.9 | Exit 1 with active-line upgrade/use-supported-line advice; command log stops at `go env GOVERSION` (`raw-runtime-probes.log`). |
| Execute that advice with fixture Go/linter build 1.26.9; compare compatible newer 1.27.2 / 1.27.1 | Both full script runs exit 0 and reach the success banner. Unknown linter build refuses exit 1 before tests. These are metadata/phase probes, not real fixture compiler qualification. |
| **Mandatory compiler-valid mutation:** remove only `version_at_least "${BASH_REMATCH[1]}" "$actual_go_line"` and its attached active-line failure command at `release-validate.sh:61–62`; retain minimum check | `bash -n` succeeds. `go test ./internal/tools/releasevalidate -run '^TestReleaseValidateSuccessAndFailureBoundaries/linter_must_cover_newer_active_Go_line$' -count=1 -v` compiles and exits 1 at `release_validate_test.go:91`: expected compatibility refusal is absent, full gate falsely succeeds. **Killed by intended assertion**, not syntax/build error (`mutation-active-go.log`). Restore baseline script and rerun same test: exit 0 (`restored-active-go.log`). |
| Actual configuration: release/CI lint/container to 1.27 / 1.27 / 1.27.2, keep `go.mod` and CI test at 1.26 | `just toolchain-check` exit 0; minimum remains 1.26.0. Coordinated edits avoid the old container caller accidentally preserving a line mismatch. Restored. |
| Actual configuration: digest-qualified exact 1.26.10 patch; separately raise module minimum to 1.26.10 with container still 1.26.9 | Digest patch passes; older-container/raised-minimum patch refuses exit 1. Restored (`policy-probes.json`). Shipped drift tests also refuse absent/floating Go/linter pins. |
| Actual selectors: `1.26.x`, duplicate YAML `go-version`, release `go-version-file` alongside selector | Each refuses exit 1. Quoted `check-latest: "true"` also refuses via YAML bool decoding; recorded as an observed parser-form limitation, not independently classified as a hosted Actions defect. Restored (`policy-probes.json`). |
| Disabled test setup-go; move linter into setup-free job; add minimum-race job under newer release line | First two survive policy (M4); extra minimum job is falsely refused. Exact edits/output retained; restored. |
| Add module `toolchain go1.27.2`, invoke actual older Go launcher | Offline policy survives while actual Go switches (M1); baseline module restored. |
| `just renovate-toolchain-check /Users/andyeschbacher/.npm/_npx/80ce72c594ca058a/node_modules/renovate`; `node /Users/andyeschbacher/.npm/_npx/80ce72c594ca058a/node_modules/renovate/dist/config-validator.js --strict renovate.json` | Exit 0; actual extraction/rules and strict schema pass (`renovate-restored.log`, `renovate-validation.log`). No inherited preset or hosted scheduling claim. |
| **Hostile cross-check in Renovate lens:** actual engine generates exact workflow linter update; apply it with matching ARG to actual config and run `just toolchain-check` | Policy refuses the coordinated generated patch and minor pins (M2), although grouping assertions pass. Execute the indicated partial-selector repair: both policy runs exit 0. Restore both configuration files (`renovate-generated-policy.json`). |
| Raw `go1.28rc1`; exact `devel go1.28-deadbeef` metadata through complete clean fixture | Prerelease refuses exit 1; devel returns exit 0 on Bash 3.2 with no downstream gate (M3). Temporary assertion test is compiler-valid and fails for successful process status; archived then removed. Prefix-only development probe exits 1, falsifying the assumption that testing the function alone establishes the production caller's status. |
| Changed optional wire branches | Not applicable: the captured implementation diff has no application/wire/runtime branch changes. No non-default semantic-wire coverage is claimed. |

### Second-pass conclusion and surviving hypotheses

The systemic counterexamples compare a configured projection with the actual next action: module selectors with Go auto-selection (M1), engine grouping with generated policy-consumable edits (M2), step counts with per-job execution (M4), and a function refusal with the complete script's process status (M3). Each was settled with hostile evidence. The clean release fixture's stubbed policy command and the optional Renovate check's caller-supplied update type are coverage boundaries; they explain why those tests alone cannot detect M1/M2/M4. They are distinct from the reachable defects reproduced above.

- **Falsified within tested stable configurations:** a newer release/local compiler must force a higher minimum; exact active patches must match; missing build metadata is accepted; container patch/digest refresh is a line change; custom Go patch rules silently inherit the repository's weekly/3-day delay. Real checks or engine replay refuted these hypotheses.
- **Scanner cache:** baseline key includes `steps.go.outputs.go-version`, so a setup-go patch change alters the cache identity; the stale-compiler hypothesis for that normal path is falsified by the actual expression. Hosted cache restore/rebuild was not run. M1 leaves a separate actual-compiler versus setup-output seam open.
- **Unresolved and bounded:** inherited presets, live release lookup/availability, hosted PR orchestration/scheduling, arbitrary dynamic workflow matrices/expressions, actual prerelease/development compiler qualification, and future linter support. No support matrix was guessed from release numbers.
- **Inherited owner evidence only:** the implementing task records full race/build/tidiness/planning qualification, vulnerability scan, and real Linux container Go/linter-build Go 1.26.9 with lint 2.13.0. Those container/full-suite checks were not rerun here and do not override these findings. Independent real host lint and focused checks were rerun.

### Isolation attestation and guarded delivery

The helper created the independent `--no-hardlinks` clone before implementation inspection. All inspection, edits, probes, tests, and report construction used that workspace. No source branch/staging/worktree operations or implementation edits were performed in the handoff checkout; only the assigned audit is eligible for guarded transfer. The sibling report was not read.

The observed helper verification receipt (`verify-initial.log`) reports:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.c1AyU0
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.c1AyU0/.git
baseline_commit=2a23d81fd494a239424ca1f91c7e8a6c4755c730
source_blob=69f28e560d9bc4d421e82f33759cffa5c55fce9b
source_fingerprint=9cb83ff65bab0a5f28d96299cdef145b27b74a6a
deliverable=planning/audits/6gj70e4p18b3-2026-10-09-go-toolchain-coordination-implementation-codex.md
deliverable_changed=true
transfer=pending
```

All tracked probes are restored to the helper's captured baseline; temporary test files are removed. Only this audit differs, with no staged changes or additional commits. `audit lint 6gj70e4p18b3 --json` reports no issues; finding creation receipts confirm M1–M4 are open. Final helper verification and inspected audit diff are retained as `verify-final.log` and `audit-final.diff` in the evidence directory.

**Guarded transfer result on delivery: `succeeded`.** This is a delivery postcondition, completed only when `scripts/isolated-review-workspace.sh transfer --sandbox <workspace>` emits `transfer=succeeded` and confirms identical source/deliverable bytes. The actual helper receipt is retained at `/private/tmp/taskflow-go-audit.ww4WUN/transfer.log`. If that command refuses, the sandbox report remains incomplete and the refusal is reported instead of writing around the guard. The workspace and evidence remain in place until the owner confirms receipt; no manual copy, sibling edit, push, or additional commit is authorized or performed.

## Owner follow-through (2026-10-10)

Accepted and fixed M1-M4; the original reviewer verdict above records the captured baseline, not the final state. Resolution notes identify the targeted regressions. Qualification passes the full host race suite, lint/build/tidiness, real Renovate 44.77.0 generated-linter replacements, and pinned Linux tests/lint with the checkout mounted read-only. Unsupported raw compiler metadata is exercised through the complete production script on host Bash 3.2 and Linux Bash. Supported static workflow roles and engine-replay limits are documented in the implementing task; hosted execution and a full clean-candidate gate are not claimed.
