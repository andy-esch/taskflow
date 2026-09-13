---
schema: 1
id: 6g9g0jf3rsja
bucket: open
area: tui-overview-parity-and-miami-vice-implementation
date: "2026-09-12"
updated_at: "2026-09-12"
---
# Audit: TUI Overview parity + the miami-vice theme — 2026-09-12

> **This audit is the review brief AND the deliverable.** Scope, commands and known
> traps are below. Record your findings in the Findings section of THIS file, using the
> tool. Do not review in prose elsewhere.

## Your task

Perform an adversarial implementation review of two changesets. Both were authored by a
Claude Code session; neither has had an independent reviewer.

This repo dogfoods its own tool for planning. Read `CLAUDE.md` at the root first — it is
binding. `docs/ARCHITECTURE.md` is the one-screen orientation.

## Scope

**A. Merged — PR #231, branch `worktree-theme+miami-vice`**

```
e16a21d  test(design): cover every registered theme's find highlights
c037150  feat(design): add the miami-vice theme
c556735  feat(cli): complete theme names for --theme and theme preview
2d62136  docs(planning): record the 80s design directions and deliver miami-vice
```

**B. Open — branch `worktree-tui+dashboard-counts-staleness`**

```
e613b27  feat(tui): show the active/archived task counts on the Overview
5fa7199  feat(tui): colour the Overview's in-progress ages by staleness
3aabf7a  docs(planning): close the two Overview parity tasks
```

Changeset A is in `main`'s history; for B use `git log --oneline origin/main..HEAD` on
that branch. Review the two as one body of work.

## How to record findings

Findings go in the Findings section below, one `####` block each. Get the exact
authoring contract from the tool before writing any:

```
tskflwctl schema audit
```

**The one trap that fails silently:** a finding's code must match `[A-Z]+[0-9]+` followed
by a literal `.` — `H1.` parses, while `H-1.`, `H1:`, `h1.`, `**H1.**` and a bare `1.`
parse to NOTHING, with no error. Every other malformation is caught loudly.

A second trap, hit while scaffolding this file: an HTML comment does **not** shield an
example finding from the parser — only a code fence does. Keep any illustrative finding
fenced.

Run `tskflwctl audit lint` when you finish and confirm it is clean, and
`tskflwctl audit info 2026-09-12-tui-overview-parity-and-miami-vice-implementation` to
check the tally matches the number of findings you actually wrote. That is the check
that proves they registered.

Prefer `tskflwctl audit append` over hand-editing where it fits.

**Never hand-edit a Status or Resolution line.** `tskflwctl audit finding <audit> <code>
--status <v> [--note <text>]` owns both and writes them in one validated, atomic edit.
Statuses: `open | in-progress | fixed | tracked | deferred | superseded | wontfix`;
`tracked` requires naming the destination task.

Leave your findings **open** and do not close this audit — triage is the maintainer's
call. If a finding warrants follow-up work, file it with `tskflwctl task new "Title"
--epic <id> --tags <tag>` and reference it rather than widening scope yourself.

`tskflwctl lint` must pass before you finish.

## Where the implementing agent thinks the risk is

Offered so you don't spend budget rediscovering it — **not** a scope limit. Findings
outside this list are worth more than confirmations inside it.

1. `internal/cli/theme_test.go` — `TestThemeEntries` is partly self-referential. It
   derives `want := design.Names()`, the same source `themeEntries` calls internally, so
   it can no longer catch a theme vanishing from the registry. Believed defused by
   `TestNames` and the three slot-pin tests. Verify that reasoning, and judge the trade.
2. `internal/cli/render/status.go` — `countLine` no longer filters zero buckets; it
   depends on `Summary.SplitCounts` having done so upstream. Correct for both current
   callers. Is the documented contract enough, or should the guard return?
3. `internal/core/service.go` — `Summary.SplitCounts` placement. Presentation-adjacent
   filtering now sits in the application layer beside the domain-ish active/archived
   split, and the CLI's private `splitCounts` was deleted. Widest blast radius here.
   Does it respect the boundary `docs/ARCHITECTURE.md` sets out?
4. `internal/tui/column.go` — the `relDateCells` / `staleDateCells` split: three
   functions where there was one, so per-item ages colour by staleness while rollup
   columns stay neutral. Right shape? Is excluding the atlas space column defensible?
5. `internal/design/theme.go` — the miami-vice palette. Two deliberate calls to probe:
   `ColorYellow` is left unmoved (it carries warn/revisit/audit-open/priority alongside
   in-progress), and `#b026ff` is confined to the gradient because it measures 4.07:1 as
   text. Note `--theme miami-vice` on a LIGHT terminal renders plain Latte, visually
   indistinguishable from `neon`.

## Worth your independent attention

- Whether the new TUI tests have teeth. Only the staleness pair was spot-checked by
  reverting the implementation; treat the rest as unproven. One assertion there
  originally passed against a no-op and had to be rewritten to compare rendered styling
  rather than cell text.
- Terminal-width safety of the new dashboard rows at narrow sizes, and their interaction
  with the cursor-keyed scroll window in `dashboard.view` / `scrollTo`.
- Whether the `--theme` and `theme preview` completions work in a real shell; they were
  exercised only through cobra's `__complete`.
- Anything in changeset A, which merged without review.

## Validation the implementer reported as green

`just test` (`go test -race ./...`) · `just lint` (0 issues) · `just docs-check` ·
`tskflwctl lint`. Re-run rather than trust; report any disagreement as a finding.

## Findings

One `####` block per issue. Un-fence this shape — it is fenced so the example does not
itself parse as a live finding:

```text
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

<what's wrong, why it matters, evidence>

**Recommendation:** <minimum fix>
```

Codes are conventionally H/M/L plus a number, by severity.

#### M1. Cross-space status now prints every zero-valued bucket  · **Status:** open

**File:** internal/cli/render/status.go:99 | **Component:** CLI status renderer
**Effort:** XS · **Urgency:** soon

The zero filter was removed from `countLine` on the premise that all callers receive `Summary.SplitCounts` output, but `renderCompactSpaceSummary` still calls it with raw `summary.Counts` at line 172. That slice explicitly contains every status and may contain zeroes (`internal/core/service.go:306`). On this branch, `go run ./cmd/tskflwctl --no-color status --all` renders `0 next-up · 1 ready-to-start · 0 in-progress · 14 completed · 0 deprecated · 1 deferred` for `andy-esch-infra`, and an empty `dotfiles` summary renders six zero buckets. Running the same command from the merged Miami Vice worktree renders only non-zero buckets and `tasks none`, respectively. This is a user-visible regression outside the two single-space dashboards and violates the new zero-omission contract.

**Recommendation:** Restore the defensive zero check in `countLine`, or pass a zero-filtered projection at the compact caller; add a `StatusAllHuman` regression test built from a canonical Summary containing all status buckets.

#### M2. Short Overviews can make the new task counts permanently unreachable  · **Status:** open

**File:** internal/tui/dashboard.go:79 | **Component:** TUI dashboard viewport
**Effort:** S · **Urgency:** soon

The new task heading and count lines deliberately have no navigation target, while `dashboard.view` centers its window exclusively on the selected target row (`internal/tui/dashboard.go:299-315`). When there is no in-progress task, the first selectable row is normally an epic below the counts and the empty in-progress widget. In a real 60x12 Zsh PTY, this branch opened with `in progress (0)` as the first visible line; `tasks`, `active`, and `archived` were all clipped above the window. The dashboard key handler moves only through `d.nav`, so no key can focus or reveal those non-target rows: moving up from the first epic wraps to the last target. The feature is therefore absent, not merely initially scrolled, on a supported short terminal. Existing count tests all use height 40, while the short-screen test checks only that the selected last row remains visible.

**Recommendation:** Give informational rows a reachable vertical viewport policy, or pin/reserve the count band while cursor-scrolling the remaining widgets; add a no-in-progress short-height test proving both the counts and current selection can be reached.

#### L1. Theme preview completes an impossible second theme name  · **Status:** open

**File:** internal/cli/theme.go:70 | **Component:** shell completion
**Effort:** XS · **Urgency:** eventually

`theme preview` accepts at most one positional name, but its new `ValidArgsFunction` reuses `completeThemeNames`, which ignores `args` and always returns the registry. This survives the Cobra driver and is visible in a real shell: after typing `tskflwctl theme preview neon ` in Zsh and pressing Tab, the generated completion offered `catppuccin`, `miami-vice`, and `neon`; selecting one produces a command rejected by `cobra.MaximumNArgs(1)`. The flag-value completion itself works, and prefix filtering works, so the defect is specifically the already-filled optional positional.

**Recommendation:** Use a positional wrapper that returns no candidates once `len(args) >= 1`, while keeping the registry-driven function for `--theme`; cover both the first and second positional completion requests.

#### L2. Staleness tests do not prove the Overview uses the colored helper  · **Status:** open

**File:** internal/tui/column_staleness_test.go:11 | **Component:** TUI regression coverage
**Effort:** XS · **Urgency:** soon

The new tests call `staleDateCells` directly and prove that helper differs from `relDateCells`, but no test observes styling through `dashboard.setSummary`. The existing dashboard age test at `internal/tui/dashboard_test.go:286` strips ANSI before checking dates and alignment. Consequently, changing the production call at `internal/tui/dashboard.go:105` back to `relDateCells` would leave the helper tests and every rendered-dashboard assertion satisfied while removing the delivered feature. The current wiring is correct, but its load-bearing integration is unguarded.

**Recommendation:** Render an Overview fixture with fresh and stale in-progress tasks and compare the two date cells styling before stripping ANSI, so the test fails if the dashboard call site returns to the neutral helper.

## Candidate tasks

_No separate tasks filed. M1, M2, and L2 are bounded corrections to the open Overview branch; L1 is small enough to fold into its review follow-up rather than creating planning overhead before maintainer triage._
