---
schema: 1
id: 6g9g0jf3rsja
bucket: open
area: tui-overview-parity-and-miami-vice-implementation
date: "2026-09-12"
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

_No findings recorded yet._

## Candidate tasks

_Follow-up work you would file, if any. Reference task ids once created._
