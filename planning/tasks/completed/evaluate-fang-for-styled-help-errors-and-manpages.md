---
status: completed
epic: 20-cli-ux-and-ergonomics
description: Evaluate charmbracelet/fang for styled help/errors/manpages on the human face; gated so it never touches the --json envelope or exit codes
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux]
created: "2026-06-19"
updated_at: "2026-06-21"
completed_at: "2026-06-21"
---
## Objective

Evaluate (and, if it earns its keep, adopt) [charmbracelet/fang](https://github.com/charmbracelet/fang)
to upgrade the **human face**: wrap the root with `fang.Execute(ctx, root)` for
styled help pages, fully-styled errors, automatic `--version`, manpage generation
(mango), a completion command, and adaptive light/dark theming — a big visual lift
for very little code.

This is explicitly an **evaluation with a hard gate**, because fang reshapes global
help/error rendering and our machine contract is non-negotiable.

## Risks / gates (the whole point of scoping this as eval-first)

- **Must never touch the agent contract:** the `--json` error envelope, the exit-code
  mapping (10/11/13/14), and any non-TTY/piped output must be byte-identical with and
  without fang. fang's fancy errors are for humans on a TTY only.
- **Dep cost:** fang v2 wants lipgloss v2; we're on lipgloss v1 (via the TUI). Assess
  the bump's blast radius on the TUI before committing.
- **`SilenceUsage`/`SilenceErrors`:** we already set these and map errors to exit codes
  ourselves — confirm fang composes with that rather than fighting it.
- Experimental library — pin a version, watch churn.

## Acceptance criteria

- [ ] A spike behind a flag/build shows styled help + errors on a TTY with the
      `--json` envelope and exit codes provably unchanged (test both paths).
- [ ] lipgloss v2 bump assessed (TUI still builds/renders) or a path to avoid it found.
- [ ] Decision recorded: adopt (with the gates wired) or decline (with reasons).
- [ ] If adopted: manpage generation replaces/augments any hand-written man content.

## Out of scope

- Replacing the TUI (bubbletea stays; fang is for the cobra CLI surface).
- Changing exit codes or the error-envelope schema.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Sibling human-face work: [[glamour-render-markdown-bodies-in-show]]
- fang's `man` overlaps the manpage angle noted in
  [[auto-generate-cli-reference-docs-with-a-ci-sync-check]].

## Decision (2026-06-21): ADOPTED

Evaluated via a worktree spike (`spike/fang-eval`), then shipped on branch
`feat/fang-styled-cli`. Findings + cons + the in-scope follow-ups live in
[[2026-06-21-fang-evaluation-spike]].

**Outcome:** the central risk (a lipgloss v2 *bump* on the TUI) was a non-issue —
v1 and v2 coexist (different module paths), so the TUI is untouched; and lipgloss
v2 has since gone **stable (v2.0.4)**, which fang v1.0.0 builds against cleanly.
The machine contract is preserved by a TTY/`--json` gate (`useFang`) — fang wraps
the human face only; piped/agent output is byte-identical, guarded by a unit test
+ the existing subprocess smoke.

**Shipped:** gated `fang.Execute` (styled help + errors), a repo-aligned 16-color
ANSI `ColorScheme` (consistent with `render`/TUI, not charmtone truecolor),
roff manpages via `internal/tools/mangen` + goreleaser archive wiring, `just man`.

**Knowingly accepted:** fang title-cases help descriptions/headers (hardcoded,
not configurable). **Deferred:** `WithNotifySignal` (needs prompt-abort
interaction check) and the broader lipgloss-v2 UI ideas →
[[explore-lipgloss-v2-and-charm-ecosystem-ui-enhancements]].
