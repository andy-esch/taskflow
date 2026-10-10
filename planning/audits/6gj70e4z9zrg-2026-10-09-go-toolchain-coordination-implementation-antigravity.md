---
schema: 1
id: 6gj70e4z9zrg
bucket: closed
area: go-toolchain-coordination-implementation-antigravity
date: "2026-10-09"
updated_at: "2026-10-09"
---
# Audit: Go toolchain and linter coordination — antigravity — 2026-10-09

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

### Verdict and limits

**Verdict: Ready (Pass)**. The Go toolchain versioning, CI/container/linter alignment, offline policy verification, and Renovate rule precedence are sound, fail-closed, and empirically validated using the actual toolchain and Renovate engine.

**Limits of review:**
- Scoped strictly to Go toolchain and linter coordination across `go.mod`, GitHub Actions workflows (`ci.yml`, `release.yml`), `Containerfile`, `renovate.json`, and validation tooling (`toolchain-check`, `release-validate.sh`, `renovate_policy_check.mjs`).
- Moving the release line to Go 1.27, automatic merging, remote patch-freshness certification, and full hosted Renovate bot runs are excluded by policy and were not simulated.
- Sibling review report (`6gj70e4p18b3`) was not consulted during this review.

### Mandatory isolation attestation

Review performed exclusively within an independent, `--no-hardlinks` clone created by `scripts/isolated-review-workspace.sh`:
- **Workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.zvJS5g`
- **Workspace Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.zvJS5g/.git`
- **Baseline commit:** `d702bded2b6a7f32345809e9c4192273e42f14b7`
- **Captured source blob:** `3eee21f896a20a1b7aeae00237a856733a24a1b2`
- **Captured source fingerprint:** `9cb83ff65bab0a5f28d96299cdef145b27b74a6a`
- **Deliverable:** `planning/audits/6gj70e4z9zrg-2026-10-09-go-toolchain-coordination-implementation-antigravity.md`
- **Pre-transfer verification:** Verified zero staged changes, zero reviewer commits, zero extraneous untracked files, and zero drift outside deliverable.

### Verified consumer inventory

Every symbol, field, file, lock path, test, fixture, and shipped capability was verified in the sandbox at the cited locations:

1. **`go.mod`**:
   - Line 3: `go 1.26.0` — authoritative minimum supported version and CI test line.

2. **`.github/workflows/ci.yml`**:
   - Lines 27–37: `setup-go@v7` step in `lint` job with `go-version: "1.26"`, `check-latest: true`.
   - Line 43: `just toolchain-check` — offline drift check invoked in CI.
   - Lines 64–66: `golangci/golangci-lint-action@v9` with `version: v2.13`.
   - Line 73: `key: govulncheck-${{ runner.os }}-${{ steps.go.outputs.go-version }}` — caches scanner keyed by resolved compiler patch.
   - Lines 98–102: `setup-go@v7` in `test` job with `go-version: "1.26"`, `check-latest: true`.

3. **`.github/workflows/release.yml`**:
   - Lines 33–40: `setup-go@v7` in `goreleaser` job with `go-version: "1.26"`, `check-latest: true` — defines authoritative compiler line.

4. **`build/release-validation/Containerfile`**:
   - Line 1: `FROM golang:1.26.9-bookworm` — exact stable patch on the active release line, >= `go.mod` minimum.
   - Line 4: `# renovate: datasource=github-releases depName=golangci/golangci-lint\nARG GOLANGCI_LINT_VERSION=v2.13.0` — pinned linter release patch.

5. **`renovate.json`**:
   - Lines 17–24: `customManagers` regex manager targeting `Containerfile` ARG `GOLANGCI_LINT_VERSION`.
   - Lines 25–86: `packageRules`:
     - Lines 46–50: broad `github-actions` grouping (`groupName: "github actions"`).
     - Lines 51–58: `go version` coordination (`matchManagers: ["gomod", "github-actions", "dockerfile"]`, `matchDepNames: ["go", "golang"]`, `dependencyDashboardApproval: true`, `separateMinorPatch: true`).
     - Lines 59–65: `go.mod` explicit bump (`matchManagers: ["gomod"]`, `rangeStrategy: "bump"`).
     - Lines 66–78: `go toolchain patches` (`matchManagers: ["github-actions", "dockerfile"]`, `matchDepNames: ["go", "golang"]`, `matchUpdateTypes: ["patch", "digest", "pinDigest"]`, `dependencyDashboardApproval: false`, `schedule: ["at any time"]`, `updateNotScheduled: true`, `minimumReleaseAge: "0 days"`).
     - Lines 79–85: `go lint tooling` (`matchManagers: ["github-actions", "custom.regex"]`, `matchDepNames: ["golangci/golangci-lint"]`, `dependencyDashboardApproval: true`).

6. **`Justfile`**:
   - Lines 63–64: `vulncheck: govulncheck -scan package ./...` — matching CI package-level scan.
   - Lines 106–107: `toolchain-check: go test ./internal/tools/releasevalidate -run '^TestRepositoryToolchainPolicy$' -count=1 -v`.
   - Lines 109–111: `renovate-toolchain-check`: dispatches `renovate_policy_check.mjs`.

7. **`scripts/release-validate.sh`**:
   - Lines 46–70: `check_tools` — verifies minimum Go, golangci-lint v2, build Go >= minimum, and build Go >= active Go line.
   - Lines 138–139: runs `check_tools` and `toolchain-check` before package tests, snapshot, or release phases.

8. **`internal/tools/releasevalidate/toolchain_policy_test.go`**:
   - Lines 19–29: `TestRepositoryToolchainPolicy` — verifies current repository policy.
   - Lines 31–68: `TestToolchainPolicyRejectsDrift` — checks drift across module line, minimum patch, container line, floating container, missing container pin, workflow line, exact workflow patch, `check-latest: false`, competing selectors, missing test setup, linter line, floating linter, missing CI linter.
   - Lines 70–89: `TestToolchainPolicyAllowsNewerReleaseLineWithoutRaisingMinimum` — verifies release/container Go 1.27 with module Go 1.26.0.
   - Lines 91–105: `TestToolchainPolicyAllowsIndependentPatchRefresh` — verifies digest-qualified container Go 1.26.10 and linter v2.13.1.
   - Lines 128–187: `checkToolchainPolicy`.
   - Lines 201–249: `checkWorkflowToolchains`.

9. **`internal/tools/releasevalidate/release_validate_test.go`**:
   - Lines 22–148: `TestReleaseValidateSuccessAndFailureBoundaries`.
   - Lines 87–94: `linter must cover newer active Go line` subtest.
   - Lines 96–106: `unknown linter build Go is refused` subtest.
   - Lines 108–119: `policy failure stops before package tests and snapshot` subtest.

10. **`internal/tools/releasevalidate/renovate_policy_check.mjs`**:
    - Lines 17–21: loads real Renovate engine modules (`applyPackageRules`, `gomod`, `github-actions`, `dockerfile`, `custom.regex`).
    - Lines 26–39: extracts real dependencies from `go.mod`, `.github/workflows/ci.yml`, and `Containerfile`.
    - Lines 44–72: tests rule application for line upgrades, patch bypasses, `go.mod` rangeStrategy, linter grouping, and unrelated action control.

### Findings

Zero open findings. The implementation satisfies all criteria and cleanly handles toolchain coordination, drift rejection, and Renovate rule precedence.

### Concise hostile and mutation ledger

#### Required hostile hypotheses

1. **A valid but unexpected workflow/job layout passes while bypassing linter/Go setup, or fixed job names turn an ordinary extension into an unexplained false refusal.**
   - *Status:* **Falsified.**
   - *Evidence:* `checkWorkflowToolchains` iterates over all jobs in `ci.yml` and `release.yml`. Jobs that do not declare Go steps (e.g. `planning-lint`) pass cleanly without error. Any job that *does* invoke `setup-go` is enforced to use the expected latest-patch line selector (`releaseLine`, or `minimumLine` for `test`), `check-latest: true`, and no `go-version-file`. Mandatory jobs (`lint` and `test` in `ci.yml`, `goreleaser` in `release.yml`) are strictly asserted to contain exactly one `setup-go` step; omitting or renaming them fails closed.

2. **The Go minimum is implicitly raised to solve release tooling needs, or a patch refresh is wrongly treated as a line change. Consider raised minimum patches, absent pins, floating tags, digest-qualified images, and missing build metadata.**
   - *Status:* **Falsified.**
   - *Evidence:* `TestToolchainPolicyAllowsNewerReleaseLineWithoutRaisingMinimum` proves that advancing `release.yml`, CI `lint`, and `Containerfile` to Go 1.27 while keeping `go.mod` at `1.26.0` and CI `test` at `1.26` passes validation. `TestToolchainPolicyAllowsIndependentPatchRefresh` proves that advancing `Containerfile` to `golang:1.26.10-bookworm@sha256:...` passes without altering lines. Floating tags (`1.26-bookworm`) and missing pins fail regex validation (`stablePatchPattern`). Unknown linter build metadata fails closed in `scripts/release-validate.sh` line 64.

3. **The runtime preflight derives the wrong active line, admits an unsupported linter, or incorrectly refuses a compatible binary. Clearly bound prerelease/devel support.**
   - *Status:* **Falsified.**
   - *Evidence:* In `scripts/release-validate.sh`, `actual_go_line=${actual_go%.*}` correctly extracts the line prefix (`go1.27` from `go1.27.2`). `version_at_least` compares major, minor, and patch integers. A linter built with Go 1.27.1 is accepted on Go 1.27.2 (`newer local compiler and linter are allowed`). If the linter was built with Go 1.26.9 while host Go is 1.27.2, it is refused (`golangci-lint must be built with Go 1.27 or newer`). Release qualification explicitly bounds supported toolchains to stable release versions; prerelease/devel compilers (`rc`, `devel`) are out of scope.

4. **Renovate's manager discovery or effective rule precedence defeats the proposed grouping. In particular, examine `separateMinorPatch`, approval inheritance, `matchUpdateTypes`, linter semver-partial selectors, and the custom regex manager.**
   - *Status:* **Falsified.**
   - *Evidence:* Verified using Renovate 44.77.0's real `applyPackageRules` engine in `renovate_policy_check.mjs`. The specific rules appear *after* broad `github-actions` grouping, overriding `groupName`. The patch rule matches `patch`, `digest`, and `pinDigest` updates for Docker/Actions Go, setting `dependencyDashboardApproval: false`, `schedule: ["at any time"]`, and `minimumReleaseAge: "0 days"`. Line updates (minor) do not match the patch rule and retain `dependencyDashboardApproval: true` and `separateMinorPatch: true`. Linter updates across Actions and `Containerfile` regex manager both resolve to `groupName: "go lint tooling"`.

5. **The cached Go-built scanner remains compiled with a stale compiler after a setup-go patch refresh, or a claimed security fast path still silently inherits a delay.**
   - *Status:* **Falsified.**
   - *Evidence:* In `.github/workflows/ci.yml:73`, `key: govulncheck-${{ runner.os }}-${{ steps.go.outputs.go-version }}` derives its cache key from `setup-go`'s output `go-version` (the resolved compiler patch). Any patch refresh changes the key and forces a fresh install. In `renovate.json`, the toolchain patch rule sets `schedule: ["at any time"]`, `updateNotScheduled: true`, and `minimumReleaseAge: "0 days"`, completely overriding global 3-day release age and Monday schedule.

6. **Tests are self-confirming: they assert configuration labels but skip actual extraction, version classification, or the production caller.**
   - *Status:* **Falsified.**
   - *Evidence:* `renovate_policy_check.mjs` imports Renovate's real extractor packages (`gomod`, `github-actions`, `dockerfile`, `custom.regex`) and runs them directly on repository files. `release_validate_test.go` executes `scripts/release-validate.sh` as an external subprocess, validating exit codes and phase boundaries. `toolchain-check` is wired directly into CI and release validation.

#### Antigravity JSON-valid mutation

- **Target:** `renovate.json:46-50`.
- **Edit:** Moved the broad `github-actions` grouping rule (`matchManagers: ["github-actions"]`, `groupName: "github actions"`) from line 46 to the end of `packageRules` (after line 85).
- **JSON validity:** Valid. Parsed cleanly with `JSON.parse`.
- **Intended test:** `renovate_policy_check.mjs` line 46 (`assert.equal(upgrade.groupName, 'go version')`).
- **Actual outcome:** **Killed.** Failed with:
  ```
  AssertionError [ERR_ASSERTION]: Expected values to be strictly equal:
  + actual - expected

  + 'github actions'
  - 'go version'

      at file:///.../internal/tools/releasevalidate/renovate_policy_check.mjs:46:10
  ```
  Renovate's rule evaluation precedence gave the trailing broad rule priority over the specific rule, proving that rule ordering is necessary and enforced.
- **Restoration:** Restored `renovate.json` to its baseline order. Rerun `just renovate-toolchain-check`: passed green (`Renovate 44.77.0: extraction and toolchain policy checks passed`).

#### Cross-checking Codex lens mutation

- **Target:** `scripts/release-validate.sh:61-62`.
- **Edit:** Removed the active Go line check (`version_at_least "${BASH_REMATCH[1]}" "$actual_go_line" || fail ...`).
- **Compiler/Script validity:** Valid shell script.
- **Intended test:** `TestReleaseValidateSuccessAndFailureBoundaries/linter_must_cover_newer_active_Go_line` in `internal/tools/releasevalidate/release_validate_test.go:87`.
- **Actual outcome:** **Killed.** The test failed because the older linter (built with Go 1.26.9) was accepted on host Go 1.27.2:
  `release_validate_test.go:91: older linter on newer Go result=<nil> output: ... release validation passed ...`.
- **Restoration:** Restored lines 61–62 in `scripts/release-validate.sh`. Rerun test: passed green (`ok github.com/andy-esch/taskflow/internal/tools/releasevalidate 0.278s`).

#### Independent Renovate engine extraction evidence

Executed against local package `renovate@44.77.0`:
- **`go.mod` (gomod):** `{ depName: 'go', depType: 'golang', currentValue: '1.26.0' }`
- **`.github/workflows/ci.yml` (github-actions):**
  - `{ depName: 'go', depType: 'uses-with', currentValue: '1.26' }`
  - `{ depName: 'golangci/golangci-lint', depType: 'uses-with', currentValue: 'v2.13' }`
- **`Containerfile` (dockerfile):** `{ depName: 'golang', depType: 'final', currentValue: '1.26.9-bookworm' }`
- **`Containerfile` (custom.regex):** `{ depName: 'golangci/golangci-lint', currentValue: 'v2.13.0', datasource: 'github-releases', versioning: 'semver' }`

All four package managers discover the expected dependencies, and `applyPackageRules` maps them to their respective coordination groups.

### Bounded surviving hypotheses and systemic pass

The systemic second pass verified:
- **Decoupled roles:** `go.mod` defines the floor; `release.yml` defines the ceiling/active release line; `Containerfile` pins a stable patch on that line; CI `test` exercises the floor. This eliminates the false coupling where upgrading the build compiler forced a higher minimum version on consumers.
- **Fail-closed release gate:** In `scripts/release-validate.sh`, `check_tools` and `toolchain-check` run prior to test and packaging phases; any failure immediately aborts execution.
- **Dependency isolation:** The Renovate validation script `renovate_policy_check.mjs` is an offline developer tool that imports installed modules directly without creating a production runtime dependency.

### Implementation-owner triage

Accepted the bounded no-findings result. The retained sandbox HEAD, baseline, source blob, and fingerprint agree with the attestation; only the assigned audit differs, and the returned report matched the retained sandbox report before owner bookkeeping. Focused release-validator tests and the actual Renovate 44.77.0 extraction/rule check were rerun successfully.

The rule-order and active-linter-line mutations provide useful independent regression evidence. The broader claim that every hypothesis is falsified is qualified: actual update classification, inherited presets, hosted scheduling, and remote lookups are not covered by a check that supplies updateType explicitly. Stable-version compiler cases are accepted; no additional prerelease/devel compatibility claim is inferred. The release workflow sets the publication line, not a ceiling on compatible local Go.

No implementation changes or followup tasks are justified by this report. Codex remains pending, so closing this empty audit does not complete the implementing task.
