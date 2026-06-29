## tskflwctl theme

Inspect color themes

### Synopsis

Inspect color themes. Select one with --theme, the TSKFLW_THEME env, or the
[theme] table in .tskflwctl.toml (precedence: flag > env > config).

On a truecolor terminal the theme drives every colored surface — status glyphs,
bars, the TUI, and the picker. On a 16-color terminal the semantic colors fall
back to your terminal's own palette (so they look the same across themes there).

### Options

```
  -h, --help   help for theme
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown
* [tskflwctl theme list](tskflwctl_theme_list.md)	 - List the available color themes
* [tskflwctl theme preview](tskflwctl_theme_preview.md)	 - Preview a theme's palette (color swatches + a sample bar)

