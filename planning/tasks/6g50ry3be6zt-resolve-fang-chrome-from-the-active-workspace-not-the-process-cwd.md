---
schema: 1
id: 6g50ry3be6zt
status: ready-to-start
epic: 25-design-system-coherent-palette-and-selectable-themes
description: 'Chrome ignores -C/--space: an error box renders in the cwd repo''s theme while the body uses the retargeted repo''s'
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [design, palette]
created: "2026-08-29"
updated_at: "2026-08-29"
---
# Resolve fang chrome from the active workspace, not the process cwd

## Objective

`cli.ChromeTheme` discovers the repo config by walking up from `os.Getwd()`, so a run
that retargets with `-C` or `--space` paints its styled error box in the *cwd* repo's
theme while the command body uses the retargeted repo's.

Reproduced with two scratch repos, cwd pinned to `catppuccin` and the `-C` target to
`neon`:

```
$ tskflwctl -C ./neonrepo task show nope
error badge bg: 48;2;243;139;168   # #f38ba8 — Mocha red, not neon's #FF4242
```

## Why it is worth fixing

On the **help** path this is inherent: cobra returns before `PersistentPreRunE`, so
nothing has resolved a workspace and cwd is the only signal available.

On the **error** path it is not. By the time an error renders, `PersistentPreRunE` has
already run and `App.Th` holds the `-C`-aware answer — chrome discards it and
re-resolves from cwd. That also repeats a home-config read and a directory walk the App
already did.

This is the same class of defect as the one that motivated
`route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`: chrome
disagreeing with the body about which theme is active. That task fixed the `[theme]`
axis and left this one.

## Scope

- Give the fang colorscheme closure access to the resolved `App.Th`, preferring it when
  populated and falling back to `ChromeTheme` when it is not (help, and any error raised
  from `PersistentPreRunE` itself before theme resolution completes).
- Keep one precedence contract: whatever the App resolved must win, rather than a second
  parallel resolution path.

## Acceptance criteria

- [ ] A styled error from a `-C`-retargeted run renders in the target repo's theme, verified under a pty against two repos pinned to different themes.
- [ ] Styled help still renders outside a planning repo, with no home config, and with an unreadable one.
- [ ] An error raised from `PersistentPreRunE` before theme resolution still renders styled rather than panicking or falling back to unpainted output.
- [ ] Chrome performs at most one theme resolution per run — no duplicate home-config read or repo walk on the error path.

## Out of scope

- The help path's cwd dependency, which no available signal can improve.

## Related

- Epic [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Follows `route-fang-chrome-through-the-resolved-theme-and-fix-codeblock-contrast`, which introduced `cli.ChromeTheme`.
