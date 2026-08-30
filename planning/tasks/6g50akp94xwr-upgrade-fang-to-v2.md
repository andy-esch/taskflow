---
schema: 1
id: 6g50akp94xwr
status: completed
epic: 20-cli-ux-and-ergonomics
description: Upgrade fang to charm.land/fang/v2 v2.0.1 for native lipgloss/v2 alignment and multiline error formatting
effort: 1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, dependencies]
created: "2026-08-29"
updated_at: "2026-08-29"
started_at: "2026-08-29"
completed_at: "2026-08-29"
---
# Upgrade fang to v2 (charm.land/fang/v2 v2.0.1)

## Objective

Upgrade the CLI Cobra wrapper dependency from `github.com/charmbracelet/fang v1.0.0` to `charm.land/fang/v2 v2.0.1`, aligning module paths with Charm's canonical `charm.land/*/v2` vanity domain, adopting upstream error formatting fixes, and verifying that the human/machine TTY isolation boundary remains intact.

## Background & Motivation

Renovate proposed updating `github.com/charmbracelet/fang` from `v1.0.0` to `v2.0.1`. In Charm's v2 ecosystem, `fang` was transitioned to the canonical module path `charm.land/fang/v2`.

Upstream changes in `v2.0.1` include:
1. **Vanity Import Path Alignment:** All other Charm dependencies in `go.mod` (`bubbles/v2`, `bubbletea/v2`, `glamour/v2`, `huh/v2`, `lipgloss/v2`) use `charm.land/*`. Upgrading `fang` unifies the module namespace across the repository.
2. **Multiline Error Formatting (#86):** In v1.0.0, multiline error strings could have newlines stripped or flattened in styled error boxes. Version 2.0.1 preserves newlines in error details, improving the presentation of multi-item validation failures and cycle diagnostic trees on human terminals.
3. **Native Lip Gloss v2 Styling Engine:** Fang v1 was pinned to an older/beta Lip Gloss v2 dependency in its own `go.mod`. Version 2.0.1 builds natively against stable `charm.land/lipgloss/v2` (v2.0.4+), removing indirect dependency drift.
4. **Command Group Evaluation Optimization:** Improved preallocation during Cobra command group evaluation reduces unnecessary allocations on large CLI command trees.

## API & Seam Analysis

Inspection of `cmd/tskflwctl/main.go` confirms that the methods and types used by `tskflwctl` remain compatible in `charm.land/fang/v2`:
- `fang.Execute(ctx, root, opts...)` — unchanged entry point.
- `fang.WithoutVersion()` — suppresses default `--version` flag (keeping custom version handling).
- `fang.WithoutManpage()` — suppresses runtime manpage command (using `./internal/tools/mangen` instead).
- `fang.WithColorSchemeFunc(repoColorScheme)` — accepts `func(lipgloss.LightDarkFunc) fang.ColorScheme` to adapt to terminal light/dark background.
- `fang.WithErrorHandler(fangErrorHandler)` — routes prompt aborts (exit 130) quietly and delegates other human errors to `fang.DefaultErrorHandler`.
- `fang.Styles` and `fang.ColorScheme` — struct types and field names are preserved.

The machine/human contract isolation in `main.go` (`useFang` TTY gate) remains unchanged:
- Off-TTY (pipes, CI, scripts) and `--json` runs bypass `fang` entirely, preserving byte-identical output and semantic exit codes (10–14).

## Scope

- Update `go.mod` and `go.sum`: replace `github.com/charmbracelet/fang v1.0.0` with `charm.land/fang/v2 v2.0.1`, then run `go mod tidy`.
- Update imports in `cmd/tskflwctl/main.go` (`github.com/charmbracelet/fang` → `charm.land/fang/v2`).
- Update comments/references in `cmd/tskflwctl/main_test.go`, `internal/tools/mangen/main.go`, and documentation where relevant.
- Verify that `cmd/tskflwctl/main_test.go` (`TestUseFang`, `TestRepoColorScheme`) passes cleanly.
- Verify that all CLI golden tests, linters, and race tests pass without drift.

## Acceptance Criteria

- [x] `go.mod` requires `charm.land/fang/v2 v2.0.1` and contains no legacy `github.com/charmbracelet/fang` entries.
- [x] `cmd/tskflwctl/main.go` imports `charm.land/fang/v2` and compiles cleanly.
- [x] Styled help (`tskflwctl --help` on TTY) and styled error boxes still
  render, unchanged from v1, in the DEFAULT theme's colors. Chrome resolving the
  *active* theme is not this task — `repoColorScheme` hardcodes
  `design.Default()`, so `[theme]` and `--theme` do not reach fang before or
  after this upgrade;
  `route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`
  (6g4ykwm5svfb) owns that and depends on this task.
- [x] Multiline error messages on a TTY preserve line breaks and diagnostic formatting inside the error box.
- [x] Machine contract tests in `cmd/tskflwctl/main_test.go` confirm that `--json` and non-TTY execution bypass `fang` with byte-identical output and semantic exit codes (10–14, 130).
- [x] `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, and `tskflwctl lint` pass with 0 issues.

## Verification Commands

```bash
go get charm.land/fang/v2@v2.0.1
go mod tidy
go test -race ./...
golangci-lint run ./...
go mod tidy -diff
go run ./cmd/tskflwctl lint
```
