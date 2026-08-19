---
schema: 1
id: 6g1e947scvs9
bucket: closed
area: multi-space-config-foundation
date: "2026-08-18"
---

# Architectural audit — multi-space config foundation (Commit 86e3e1d)

Scope: Implementation and design review of commit `86e3e1d` ("feat: multi-space config foundation"), adding the user-scoped (home) configuration tier (`internal/userconfig`), CLI wiring and style initialization (`internal/cli/root.go`), pager and theme precedence mechanics (`internal/cli/pager.go`, `internal/cli/theme.go`), process-level test isolation via `TestMain` (`cmd/tskflwctl/main_test_env_test.go`, `internal/cli/main_test.go`), and the broader multi-space architecture outlined in `planning/research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md` and Epic 29.

Method: Four independent reviews across (1) Correctness and failure modes, (2) Go idiom and robustness, (3) Architectural integrity, and (4) Design and reasoning. Findings are deduplicated, grounded in `file:line`, verified through test execution and microbenchmarking, and prioritized by consequence rather than effort.

Overall health is high. The extraction of `internal/userconfig` as a pure leaf package with no dependencies on `internal/config` or repository discovery is architecturally clean. Startup latency measurements confirm that loading the home config adds only 1–11 µs per invocation (well below human or machine perception). The pointer-based field-by-field merge (`*bool` on `PagerConfig.Enabled`) handles table merges without collapsing unset keys. However, several failure-mode regressions and UX contradictions exist: missing `$HOME` in containerized/headless environments causes spurious warning spam; user config load warnings fire during shell completion; relative environment overrides bypass directory isolation; and committed repo configs override personal terminal preferences.

## High

No findings in this category — no panics, data corruption hazards, or race conditions were identified.

## Medium

#### M1. Unset `$HOME` emits spurious warnings on every command in headless and CI environments  · **Status:** fixed

**File:** `internal/userconfig/userconfig.go:93-98`, `internal/cli/root.go:80-82` | **Component:** userconfig / cli
**Effort:** XS · **Urgency:** soon

In clean, minimal, or headless environments (e.g., minimal Docker containers, CI build runners, daemons, cron tasks, or scripts executed under `env -i`), `$HOME` is frequently unset. When `TSKFLW_CONFIG_HOME` and `XDG_CONFIG_HOME` are also undefined, `os.UserHomeDir()` returns an error (`$HOME is not defined`).

Because the user config is an optional preference tier, the absence of a home directory is normal and should be treated as "no user config present". Currently, `userconfig.Dir()` wraps this error, `userconfig.Load()` returns it, and `setStyle()` prints a warning to `stderr` on every single command invocation.

**Evidence:**
Running `env -u HOME -u XDG_CONFIG_HOME -u TSKFLW_CONFIG_HOME ./bin/tskflwctl version` or `status --json` yields:
```
⚠ ignoring user config: locate home dir: $HOME is not defined
tskflwctl v0.15.0-29-g86e3e1d
```
This pollutes `stderr` for every automated tool, wrapper script, or CI pipeline running in a home-less environment.

**Recommendation:** In `userconfig.Load()`, catch errors from `Dir()` caused by a missing/unresolvable home directory when env overrides are unset and return `&Config{}, nil` (silent fallback with empty config).

**Resolution (2026-08-18, fixed).** Reproduced exactly via `env -i ./bin/tskflwctl
version`. `userconfig.Load` now returns the empty config with **no error** when `Dir()` cannot
resolve a home directory: there is no preferences file to miss, so the empty config is the
complete answer and a warning is noise the user cannot act on. Verified the fix does not silence
real problems — a file that EXISTS but is malformed still warns. Test: `TestLoad_NoHomeDirIsSilent`.

#### M2. User config warnings fire on the shell completion path and machine surfaces  · **Status:** fixed

**File:** `internal/cli/root.go:72-83`, `internal/cli/root.go:146` | **Component:** cli
**Effort:** S · **Urgency:** soon

Warnings for user config loading errors are emitted unconditionally inside `setStyle()`, which runs at the very start of `PersistentPreRunE` before the `isCompletionCommand(cmd)` guard.

The repo has an explicit pattern (documented in `warnUnknownTheme`: "not called on the completion path") to ensure that ambient warnings never reach shell completion streams or corrupt completion parsing. Because `setStyle()` prints before `isCompletionCommand(cmd)` checks whether cobra is running `__complete`, any user config syntax error or missing-home warning is emitted directly during shell tab-completion.

**Evidence:**
With a malformed TOML file in `$TSKFLW_CONFIG_HOME/config.toml`, running `./bin/tskflwctl __complete task` yields:
```
⚠ ignoring user config: read /tmp/badcfg/config.toml: toml: line 2: ...
task	Work with tasks
:4
Completion ended with directive: ShellCompDirectiveNoFileComp
```

**Recommendation:** Do not print `userErr` immediately in `setStyle()`. Instead, record the error on `App` (e.g. `app.userConfigErr`) and emit the warning in `PersistentPreRunE` after the `isCompletionCommand(cmd)` check (and alongside `warnUnknownTheme`).

**Partially disputed (2026-08-18).** The **completion-path** half is confirmed and
stands. The **"machine surfaces"/`--json`** half does not: verified that `--json` stdout remains
valid JSON because the warning goes to `ErrOut`. Warning on stderr under `--json` is this repo's
*documented convention*, not a defect — `runInitPointer` does the same and says so in a comment.
Left open for the completion fix only.

**Resolution (2026-08-18, fixed — completion half only).** The warning no longer
fires during `__complete`: `setStyle` now only *records* the load error on `App.userCfgErr`, and
a new `warnPresentation(cmd)` emits it after the `isCompletionCommand` guard. Verified with
separated streams — completion stderr now carries only cobra's own directive line. The `--json`
half of this finding is **not** a defect and was not changed: `--json` stdout was already valid
JSON because the warning goes to `ErrOut`, and warning on stderr under `--json` is this repo's
documented convention (`runInitPointer` does the same and says so). Test:
`TestUserConfig_SilentOnCompletionPath`.

#### M3. Relative `TSKFLW_CONFIG_HOME` and `XDG_CONFIG_HOME` paths cause cwd-dependent config leakage  · **Status:** fixed

**File:** `internal/userconfig/userconfig.go:87-92` | **Component:** userconfig
**Effort:** XS · **Urgency:** soon

`Dir()` reads `TSKFLW_CONFIG_HOME` and `XDG_CONFIG_HOME` directly via `strings.TrimSpace(os.Getenv(...))` without normalizing them to absolute paths. If a user or script sets either variable to a relative path (e.g., `TSKFLW_CONFIG_HOME=config`), `userconfig.Load()` resolves the path relative to the process's working directory (`cwd`).

This causes the CLI to load different "home" configs as the user navigates between directories or when commands are run with `-C / --chdir`. Furthermore, the XDG Base Directory specification explicitly mandates: *"All paths set in these environment variables, as well as the default values, must be absolute paths. If an implementation encounters a relative path in any of these variables it should consider the path invalid and ignore it."*

**Evidence:**
In `internal/userconfig/userconfig.go`:
```go
if v := strings.TrimSpace(os.Getenv(DirEnv)); v != "" {
    return v, nil
}
if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
    return filepath.Join(v, AppDir), nil
}
```
Passing `TSKFLW_CONFIG_HOME=fixtures` returns relative `fixtures`, which resolves against `cwd`.

**Recommendation:** In `userconfig.Dir()`, resolve `TSKFLW_CONFIG_HOME` to an absolute path via `filepath.Abs(v)` if set. For `XDG_CONFIG_HOME`, verify `filepath.IsAbs(v)` and ignore non-absolute values per the XDG specification, falling back to `~/.config`.

**Resolution (2026-08-18, fixed).** Confirmed: the same `TSKFLW_CONFIG_HOME=relcfg`
warned from one cwd and was silent from another, and the diagnostic printed a relative path. The
two vars now degrade **differently, on purpose**: `TSKFLW_CONFIG_HOME` is ours, so a relative value
is resolved with `filepath.Abs`; `XDG_CONFIG_HOME` is governed by the XDG Base Directory spec,
which says a non-absolute value **must be ignored** — so it falls through to `~/.config` rather
than being silently repaired. The recommendation as written (Abs on both) would have violated the
spec for the XDG var. Tests: two new `TestDir_Precedence` cases.

#### M4. Committed repository config overrides personal user preferences for local-terminal concerns  · **Status:** wontfix

**File:** `internal/cli/root.go:117-128`, `internal/cli/pager.go:73-87` | **Component:** cli / config
**Effort:** S · **Urgency:** soon

The resolution precedence for themes and pagers is `flag > env > repo config (.tskflwctl.toml) > user config (~/.config/tskflwctl/config.toml) > default`.

The package documentation in `internal/userconfig/userconfig.go:4-8` explicitly notes that `[theme]` and `[pager]` are "local-terminal concerns" that "belong to a person, not a repo." However, because `.tskflwctl.toml` is committed to git and shared across a team, placing repo config above user config means that if a repository maintainer commits `[theme] name = "neon"` or `[pager] command = "delta"`, it forcibly overrides every contributor's personal preference. A contributor with a light terminal or without `delta` installed cannot override this via their dotfiles and must set shell environment variables (`TSKFLW_THEME`, `TSKFLW_PAGER`) or pass CLI flags on every run.

**Evidence:**
`themeName` in `internal/cli/root.go:117-128`:
```go
if s := strings.TrimSpace(cfgName); s != "" {
    return s
}
return strings.TrimSpace(userName)
```
Committed `cfgName` deterministically beats dotfile `userName`.

**Recommendation:** For personal terminal concerns (`[theme]` and `[pager]`), make user config take precedence over repo config (`flag > env > user config > repo config > default`), or deprecate repo-level theming and paging so shared project files cannot hijack individual terminal presentation.

**Resolution (2026-08-18, wontfix — deliberate).** The finding did not engage
the strongest counter-argument: git's own precedence is system < global < **local**, and
`.gitignore` likewise has the committed repo file outrank the personal global one. Decision
(andy-esch): **follow the established convention** — repo overrides home, as shipped. Caveat
recorded for a future revisit: git's `.git/config` is per-clone and never committed, whereas
`.tskflwctl.toml` IS committed, so this ordering does let one contributor's checked-in preference
override another's personal one. Accepted knowingly; every escape hatch (`--theme`,
`TSKFLW_THEME`, `TSKFLW_PAGER`, `--no-pager`) outranks both config tiers.

#### M5. `TestMain` isolation relies on process-global env mutation, leaving other packages unprotected and preventing `t.Parallel()`  · **Status:** fixed

**File:** `cmd/tskflwctl/main_test_env_test.go:14-25`, `internal/cli/main_test.go:18-29` | **Component:** testing / cli
**Effort:** S · **Urgency:** eventually

Commit `86e3e1d` introduces `TestMain` to `cmd/tskflwctl` and `internal/cli` to set `TSKFLW_CONFIG_HOME` to a temporary directory.

This pattern has two latent drawbacks:
1. `TestMain` only isolates the packages where it is defined. If any other package (e.g. `internal/tui`, `internal/core`, or a future integration package) executes CLI logic or touches user preferences, it will silently read the developer's real `~/.config/tskflwctl/config.toml`.
2. `os.Setenv` in `TestMain` mutates process-global state. Individual test functions in `internal/cli` that use `t.Setenv(userconfig.DirEnv, ...)` (such as `userconfig_test.go`) cannot run alongside `t.Parallel()` without panicking.
3. This bypasses the project's non-negotiable rule of dependency injection via `*cli.App` with no package globals, coupling user config loading directly to process environment variables.

**Recommendation:** Plumb user config loading through `*cli.App` or an injectable configuration loader rather than relying solely on global `os.Getenv(DirEnv)` calls in `userconfig.Dir()`.

## Low

**Partially disputed (2026-08-18).** Sub-claim **1 is valid and is the real
finding**: `TestMain` protects only its own package, so a future package that reaches
`userconfig` would silently read a developer's real `~/.config` with nothing failing loudly.
Sub-claim 2 is misattributed — `t.Setenv` forbids `t.Parallel()` on its own, independent of
`TestMain`; no current test combines them. Sub-claim 3 does not hold: reading an env var in a
function is not a package global, and the repo already does exactly this for `TSKFLW_THEME`,
`TSKFLW_PAGER`, `TSKFLW_NO_INPUT`, `PAGER`, and `LESS` (10 `os.Getenv` calls across
`cli/color.go`, `cli/pager.go`, `cli/root.go`). The "no package globals" non-negotiable is about
package-level DI state, not env reads. Left open for sub-claim 1 only; the recommended
injectable-loader refactor is out of proportion to it.

**Resolution (2026-08-18, fixed — sub-claim 1).** The real risk (a future package
silently reading the developer's own `~/.config`) is now structural rather than per-package:
`userconfig.Dir()` calls `guardTestIsolation()`, which refuses the real-home fallback whenever
`testing.Testing()` is true and `$HOME` still matches its process-start value. A test that isolated
itself necessarily changed something (DirEnv, XDG, or HOME → `t.TempDir`), so only an *unisolated*
package trips it — and it fails loudly with the env var name in the message instead of passing
locally and behaving differently elsewhere. Test: `TestGuardTestIsolation_FiresWithoutIsolation`.
Sub-claims 2 and 3 stand as disputed above; the recommended injectable-loader refactor was not
done, as env reads are established practice here (10 `os.Getenv` calls across `cli/color.go`,
`cli/pager.go`, `cli/root.go`) and the "no package globals" rule is about DI state, not env.

#### L1. Documentation asserts a "compiler-enforced invariant" that is not enforced by the compiler  · **Status:** fixed

**File:** `internal/userconfig/userconfig.go:10-13`, `docs/ARCHITECTURE.md:99-101` | **Component:** docs / architecture
**Effort:** XS · **Urgency:** eventually

`docs/ARCHITECTURE.md` and `internal/userconfig/userconfig.go` state:
> *"Because config never imports this package, home-scope data physically cannot influence planning-root discovery — an invariant the compiler enforces rather than one the reader has to remember."*

In reality, `internal/userconfig` does not import `internal/config`. Because there is no dependency cycle, Go's compiler would freely permit `internal/config` to import `internal/userconfig` with no compilation errors. The invariant is upheld entirely by maintainer discipline, not the compiler.

**Recommendation:** Update the doc comments to describe this as an architectural layering boundary rather than a compiler-enforced invariant, or add a `golangci-lint` depguard rule prohibiting `internal/config` from importing `internal/userconfig`.

**Resolution (2026-08-18, fixed).** Correct, and the over-claim was in prose this
repo owns. Rather than merely weaken the sentence, the rule is now **machine-checked**: a
`depguard` rule in `.golangci.yml` (`config-must-not-read-home-scope`) denies
`internal/userconfig` from `internal/config/*.go`, and `docs/ARCHITECTURE.md` now reads "a
layering rule, kept honest by a depguard rule … `just lint` is what makes the claim true."
Verified by temporarily adding the import and confirming `just lint` fails with that rule's
message, then restoring.

#### L2. Subcommands overriding `PersistentPreRunE` (`init`, `doctor`) omit `warnUnknownTheme`  · **Status:** fixed

**File:** `internal/cli/init.go:36`, `internal/cli/doctor.go:21` | **Component:** cli
**Effort:** XS · **Urgency:** eventually

`newInitCmd` and `newDoctorCmd` override `PersistentPreRunE` to avoid requiring an existing planning repo, calling `app.setStyle()` directly. However, neither command invokes `app.warnUnknownTheme()`.

While `newThemeCmd` (`internal/cli/theme.go:37`) explicitly added `app.warnUnknownTheme()` to its own `PersistentPreRunE`, `init` and `doctor` were omitted.

**Evidence:**
If a user config specifies `[theme] name = "invalid"`, running `tskflwctl version` or `tskflwctl theme list` emits a warning (`⚠ unknown theme "invalid"; using "neon"`), whereas running `tskflwctl init` or `tskflwctl doctor` silently uses the default without notifying the user.

**Recommendation:** Add `app.warnUnknownTheme()` to `init` and `doctor`'s `PersistentPreRunE`, or encapsulate post-style warnings into a single helper method on `App`.

**Evidence corrected (2026-08-18).** The code observation is right but the stated
evidence is backwards, and it under-reports the scope. Verified: with an invalid theme in the home
config, `tskflwctl version` does **not** warn either — `version` overrides `PersistentPreRunE`
with a bare `setStyle()` exactly like `init` and `doctor` (`internal/cli/version.go:36`). The real
split is {`version`, `init`, `doctor`} silent vs. {everything under root's `PersistentPreRunE`,
plus `theme`} warning. Pre-existing, but made more reachable by this commit since a theme name can
now originate in the home config. The suggested single post-style warnings helper on `App` is the
right shape.

**Resolution (2026-08-18, fixed — and wider than reported).** Confirmed `version`
was silent too, so the gap covered `version`, `init` and `doctor`, not just the latter two. Rather
than add a third hand-rolled call, the warnings moved behind one seam: `warnPresentation(cmd)`
emits both the user-config and unknown-theme warnings, and `styleOnlyPreRun` is the shared
`PersistentPreRunE` for commands that run without a planning repo. `version`, `init` and `schema`
now use it; `doctor`, `theme` and `template` call `warnPresentation` after their own resolve. The
hand-rolling that caused this finding is what the shared helper removes. Test:
`TestWarnPresentation_ReachesStyleOnlyCommands` covers `version`, `schema`, `theme list`, `init`.

#### L3. Dangling symlinks at `config.toml` silently degrade to empty config without diagnostic feedback  · **Status:** fixed

**File:** `internal/userconfig/userconfig.go:116-121` | **Component:** userconfig
**Effort:** XS · **Urgency:** eventually

In `userconfig.Load()`, if `toml.DecodeFile` returns an error matching `errors.Is(err, fs.ErrNotExist)`, the function returns `&Config{}, nil` on the assumption that no config file exists.

When `config.toml` is a broken symlink (e.g. pointing to a relocated or unmounted dotfiles repository), `DecodeFile` encounters `ENOENT` on `os.Open`. `userconfig.Load()` treats this identically to a missing file and stays completely silent.

**Recommendation:** Use `os.Lstat(path)` before attempting to read: if a symlink exists but its target does not resolve, return a clear diagnostic error so `setStyle()` can inform the user that their configured symlink is broken.

**Confirmed mechanism (2026-08-18).** A dangling symlink yields `ENOENT` from
`os.Open`, which `errors.Is(err, fs.ErrNotExist)` cannot distinguish from a genuinely absent file.
Realistic for this audience specifically — people who commit their config and symlink it from a
dotfiles repo that may not be mounted. Note the fix costs an `os.Lstat` on every invocation; worth
pairing with M2 so the per-command syscall budget is considered once rather than twice.

**Resolution (2026-08-18, fixed).** `Load` now distinguishes the two things ENOENT
can mean: an absent file stays silent, while a symlink whose target is gone reports
`broken symlink -> <target>`. The `os.Lstat` runs **only on the error path**, so the common case
still costs no extra syscall — the per-invocation budget concern raised alongside this finding does
not apply. Test: `TestLoad_BrokenSymlinkIsReported`.

#### L4. Two-file split between `config.toml` and `spaces.toml` stems from missing TOML writer rather than design principle  · **Status:** fixed

**File:** `internal/userconfig/userconfig.go:31-34`, `planning/research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md:270-272` | **Component:** userconfig / design
**Effort:** S · **Urgency:** eventually

`userconfig.go` declares that `config.toml` is hand-edited and read-only to the tool, with machine-managed state intended for a separate `spaces.toml`.

This split is primarily a workaround for the fact that the codebase has no TOML AST/serializer library (only `toml.DecodeFile`) and performs mutations via string/regex slicing. If a user follows the sketch schema in `6g0ajre026c6-multi-space-home-registry-and-the-atlas.md` and places `[[space]]` tables in `config.toml`, `userconfig.Load()` silently discards them because `configFileTOML` only models `[theme]` and `[pager]`.

**Recommendation:** Before implementing Slice 2 (the registry model), settle on whether `config.toml` or `spaces.toml` owns spaces, document the file contract clearly, and evaluate whether a structured TOML modifier is needed.

**Resolution (2026-08-18, fixed — the doc defect).** The two-file split was
already the decision (recorded in this epic and chosen before slice 1 was built), so that half is
not a finding. But the finding surfaced a **real defect in the sketch**: the research doc's schema
block showed `[[space]]` inside `config.toml`, contradicting the decision and describing a layout
where entries would be **silently discarded**, since `configFileTOML` models only `[theme]` and
`[pager]`. Both the sketch and this epic's Shape section now show the two files explicitly, and
the sketch's open question 4 is marked settled with the engineering rationale (TOML here is
decode-only, so a tool-owned file can be re-marshalled wholesale). The residual — whether
`spaces.toml` needs a `schema_version` migration path — is carried forward to slice 2.

## What is solid (checked, deliberately not findings)

The following areas were inspected in detail and found robust:
- **Invocation latency:** Microbenchmarking `userconfig.Load()` demonstrated ~1.1 µs for missing configs and ~11.0 µs (5.6 KB alloc) for populated configs. Reading the home config on every command has negligible performance cost and does not introduce latency regressions.
- **Darwin XDG dotfile resolution:** Targeting `~/.config/tskflwctl` instead of `os.UserConfigDir()` (`~/Library/Application Support` on macOS) is dotfile-friendly and portable.
- **Field-by-field merge semantics:** Using `*bool` on `PagerConfig.Enabled` correctly preserves the distinction between "unset" (`nil`) and "explicitly disabled" (`false`), allowing proper tiered fallbacks without collapsing to defaults.
- **Malformed file resilience:** Logging a warning to `stderr` and continuing with the zero config ensures that a syntax error in preferences never blocks users or agents from running commands.
- **Leaf package isolation:** `internal/userconfig` has zero imports from `internal/config`, `internal/domain`, or `internal/store`, maintaining a clean leaf architecture.

## Candidate tasks

- ⏳ `tskflwctl task new "Silently ignore missing home directory in userconfig.Load" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags userconfig` — Degrade silently in userconfig.Load when os.UserHomeDir fails and no environment overrides are set.
- ⏳ `tskflwctl task new "Suppress userconfig warnings on shell completion path" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags cli` — Defer printing user config load warnings in root PersistentPreRunE until after isCompletionCommand.
- ⏳ `tskflwctl task new "Enforce absolute paths for userconfig directory resolution" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags userconfig` — Normalize TSKFLW_CONFIG_HOME with filepath.Abs and ignore relative XDG_CONFIG_HOME values.
- ⏳ `tskflwctl task new "Adjust theme and pager precedence for personal preferences" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags cli` — Allow user config terminal preferences to take precedence over committed repo configs.
- ⏳ `tskflwctl task new "Add import linter rules to guard internal/config isolation" --epic 21-code-quality-architecture-hardening --tags quality` — Add depguard/linter rules ensuring internal/config never imports internal/userconfig.
- ⏳ `tskflwctl task new "Emit unknown theme warnings in init and doctor commands" --epic 21-code-quality-architecture-hardening --tags cli` — Invoke warnUnknownTheme in init and doctor PersistentPreRunE hooks.
- ⏳ `tskflwctl task new "Warn on dangling symlinks in userconfig.Load" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags userconfig` — Distinguish broken symlinks from missing files using os.Lstat.
