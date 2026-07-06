---
status: reference
created: "2026-06-21"
tags: [cli, ux, fang, spike, decision]
---

# Fang evaluation spike — findings & decision

Spike for [evaluate-fang-for-styled-help-errors-and-manpages](../tasks/6fdtbb4037bs-evaluate-fang-for-styled-help-errors-and-manpages.md) (epic
[20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)). Branch `spike/fang-eval`, worktree-isolated. The
question: adopt `charmbracelet/fang` to upgrade the human face (styled help,
errors, manpages) **without** touching the agent/machine contract.

## TL;DR — recommendation: **adopt, narrowly and TTY-gated** (low risk, real lift)

The two fears in the task brief were both wrong:

1. **"fang v2 wants lipgloss v2; assess the bump's blast radius on the TUI."**
   There is **no bump**. lipgloss v1 and v2 have different module paths, so they
   coexist. Adding fang left *every* existing v1 charm dep (lipgloss v1,
   bubbletea v1.3.10, bubbles, glamour, huh) **byte-for-byte unchanged** in
   `go.mod` — fang pulls lipgloss/v2 **alongside** v1, for its own use
   only. The TUI never migrates. (See "Dependency reality".)
2. **"Pin an older fang to avoid the v2 dep."** Impossible — *every* fang release
   back to v0.2.0 requires lipgloss v2. But it doesn't matter, per (1).

**Update (2026-06-21, after the v2.0.4 release was flagged):** lipgloss v2 is now
**stable** (`charm.land/lipgloss/v2` v2.0.0 … v2.0.4; canonical module path is the
`charm.land` vanity, both import paths resolve to it). fang v1.0.0 *pins* the
beta in its own go.mod, but consumer MVS lets us require **v2.0.4** directly:
`go get charm.land/lipgloss/v2@v2.0.4` — and fang v1.0.0 **builds + tests clean
against stable v2.0.4** (suite green, golangci-lint 0, contract still
byte-identical). So the original "beta dependency" objection is **largely
retired**: fang is a stable major, lipgloss is a stable major, cobra is stable.

The machine contract is provably intact (10/10 byte-identical outputs, all exit
codes preserved). Cost is now ~0.96 MB binary + one pre-1.0 *transitive*
rendering lib (`ultraviolet`, never imported directly) and the long-stable
muesli mango/roff manpage libs (v0.1.0).

## What fang actually buys us (net-new only)

Mapped against what the repo already has, the genuine net-new surface is narrow:

| fang feature | verdict |
| :-- | :-- |
| **Styled help pages** | ✅ net-new, nice lift (boxed usage, colored sections) |
| **Styled errors (TTY)** | ✅ net-new for humans — pink `ERROR` badge block |
| **Manpages (mango/roff)** | ✅ net-new — `docgen` only emits markdown; this is real roff |
| Automatic `--version` | already have (`version` subcmd + JSON) → suppressed via `WithoutVersion()` |
| Completion command | already have richer dynamic slug completion → keep ours (don't pass `WithoutCompletions()`) |
| Light/dark theming | partially have (`theme` pkg) |

## Integration (the seam)

`fang.Execute` sets `SilenceUsage/Errors`, installs a styled help func, adds a
hidden `man` command, runs the command, and on error calls an overridable
`errHandler` **and returns the error** — it never sets exit codes or calls
`os.Exit`. So the wiring that preserves the contract is a **TTY gate in `main`**:

- `useFang()` is **closed** for any machine context — non-TTY stderr (pipe /
  redirect / CI) **or** an explicit `--json` run. In those cases `main` runs the
  *original* path verbatim → machine output is byte-identical **by construction**.
- On a real TTY (no `--json`), fang renders help/errors. Its error handler
  delegates prompt-aborts (exit 130) back to the existing quiet writer and routes
  everything else to `fang.DefaultErrorHandler`. `os.Exit(cli.ExitCode(err))`
  still owns the semantic codes (10/11/13/14).

`TSKFLW_FANG_FORCE=1` bypasses only the TTY check (demo aid), never the `--json`
exclusion. See `cmd/tskflwctl/main.go` on this branch.

## Evidence

**Machine contract (non-TTY, fang binary vs HEAD binary, both anchored at the
repo):** all byte-identical, all exit codes equal.

```
                    head   fang
task list --json      0      0     stdout+stderr IDENTICAL
task show <nope> --json 10   10    IDENTICAL (JSON envelope: code=not-found)
task show <nope>      10     10    IDENTICAL (prose "error:")
--help                 0      0     IDENTICAL
task list              0      0     IDENTICAL
```

**`--json` beats the force** (contract guard): with `TSKFLW_FANG_FORCE=1` *and*
`--json`, output is still the exact machine envelope, exit 10.

**Human path (forced):** styled colored help; pink `ERROR` badge on failures;
`man` emits real roff (`.TH TSKFLWCTL 1 …`).

**Quality:** `go test ./...` green, `golangci-lint run` 0 issues, `gofmt`/`vet`
clean.

## Dependency reality

`go get github.com/charmbracelet/fang@v1.0.0` — diff of `go.mod`: pure
*additions*, zero changes to existing entries.

- Added: `charm.land/lipgloss/v2` (**v2.0.4 stable** after the explicit upgrade;
  fang's own go.mod pins beta.3, overridden by consumer MVS),
  `charmbracelet/fang v1.0.0`, `charmbracelet/ultraviolet` (fang's renderer),
  `x/exp/charmtone`, `muesli/mango` + `mango-cobra` + `mango-pflag` + `roff`
  (manpages), `x/windows`, `golang.org/x/sync`.
- Binary: **16.89 MB → 17.85 MB (+0.96 MB, ~+6%)**, same toolchain/flags.

## Cons / watch-items

- **One pre-1.0 transitive dep remains.** With lipgloss pinned to stable v2.0.4,
  the only pre-release left is `charmbracelet/ultraviolet` (a `v0.0.0-…`
  pseudo-version) — fang's rendering engine, never imported directly. Lower risk
  than a beta lipgloss (you don't touch its API, and fang's stable v1.0.0 vendors
  a working version), but still pre-1.0. mango/roff are v0.1.0 (years-stable).
  Pin exactly and re-run the contract check on upgrades.
- **fang restyles our help text — and it's NOT configurable.** It title-cases
  flag/command descriptions ("machine-readable JSON output" → "Machine-Readable
  JSON output") and upper-cases section headers ("usage" → "USAGE"), overriding
  the repo's deliberate lowercase house style. Confirmed hardcoded in fang's
  unexported `makeStyles` (`Transform(titleFirstWord)` / `Transform(strings.ToUpper)`,
  theme.go:130/135/189). `WithTheme`/`WithColorSchemeFunc` control **colors only**,
  not these text transforms — so the only escapes are (a) accept it, or (b) skip
  fang's help func and use fang for errors+manpages only. Sticky con.
- **Manpage shorthand quirk.** mango renders the `-C` shorthand as `--C` in roff
  (`\fB--C --chdir\fP`). Minor; the markdown `docgen` reference is unaffected.
- **Two lipgloss majors in the tree.** Coexistence is fine for Go, but it's two
  versions to keep an eye on at upgrade time.

## If adopted — wiring checklist

- [ ] Land `main.go` TTY gate (this branch). Keep the original path for
      machine/`--json`. Do **not** pass `WithoutCompletions()` (keep our dynamic
      completion). Pass `WithoutVersion()` (keep our version string + subcmd).
- [ ] Add a `main_test.go` case asserting machine-path byte-identity + exit codes
      (codify the evidence above so a fang change can't silently break the contract).
- [ ] Decide on the description title-casing: accept it, or skip fang help and use
      fang for errors+man only.
- [ ] Route the manpage angle here (supersedes the man note in
      [auto-generate-cli-reference-docs-with-a-ci-sync-check](../tasks/6fdtbb403ba4-auto-generate-cli-reference-docs-with-a-ci-sync-check.md)); wire `man` into
      goreleaser if shipping manpages.
- [ ] Pin fang + lipgloss/v2 exactly; add a renovate/CI note re: beta churn.

## In-scope low-hanging fruit (research 2026-06-21)

Cheap additions that fit *this* task (fang human-face, TTY-gated) and the
agent/pipeline contract, ranked by value-for-effort:

1. **Align fang's palette to the repo `theme` via `WithColorSchemeFunc`.** Best
   bang-for-buck — directly defuses the "second styling authority" con. fang's
   `ColorScheme` is 16 color fields (`Title`, `Command`, `Flag`, `ErrorHeader`
   `[fg,bg]`, `Description`, …). Feed the repo's existing hex palette
   (`internal/theme`, which already maps status/priority → `{Glyph, Color}`) so
   help/errors use the *same* Charple/yellow/etc. as `list`/`status`/the TUI.
   One function; makes fang look like *our* tool, not generic Charm. (Colors
   only — does NOT fix the title-casing transform; see cons.)
2. **Ship manpages via goreleaser.** The hidden `man` command already emits roff
   (mango). Wire `tskflwctl man > tskflwctl.1` into the release build + nfpm so
   brew/apt/deb installs land a real man page — the genuine net-new this task
   names. (File/accept the `-C`→`--C` roff quirk.)
3. **`main_test.go` contract guard (mandatory companion, not optional).** Codify
   the byte-identity + exit-code evidence so a future fang/lipgloss bump can't
   silently break the machine contract. The gate is bespoke; this test is what
   makes it safe.
4. **`WithNotifySignal(os.Interrupt)` for graceful Ctrl-C** → context cancellation
   on the human path. ⚠️ Verify it composes with the huh prompt's own Ctrl-C →
   `prompt.ErrAborted` (exit 130) before adopting — could be free, could be a
   footgun. Evaluate, don't assume.
5. **`WithCommit(<sha>)` / `WithVersion`** — fold the goreleaser commit/version
   ldflags into fang's `--version`. Trivial, low value (we keep our own `version`
   subcmd + `--json` regardless via `WithoutVersion()`); do only if cheap.

## Out of scope → captured as a follow-up task

Bigger UI/rendering ideas the now-present **lipgloss v2** (`table`/`tree`/`list`)
and the **stable bubbletea v2 / bubbles v2** unlock are *not* part of this
human-face task. Captured separately as
[explore-lipgloss-v2-and-charm-ecosystem-ui-enhancements](../tasks/6feeygw02vzb-explore-lipgloss-v2-and-charm-ecosystem-ui-enhancements.md) (epic 20), covering:
lipgloss/v2 `tree` for `epic show`/planning hierarchy, `table` for list output
(gated, golden-test risk), `list` for audit/candidate lists; a **bubbletea v2 +
lipgloss v2 migration of the TUI** (epic 18) that would also *consolidate* the
two-lipgloss-majors situation onto v2; `wish` (TUI-over-SSH, cf. epic 19); `vhs`
(doc GIFs); and `charmbracelet/log` (considered, likely declined — would muddy
the `--json`/stderr discipline).

## Out of scope (unchanged)

TUI stays on bubbletea/lipgloss v1 in *this* task (untouched). No exit-code or
envelope changes.
