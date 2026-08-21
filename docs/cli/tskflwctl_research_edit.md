## tskflwctl research edit

Open a research doc in your editor (whole file; re-validated on save)

### Synopsis

Open the doc's markdown file in $VISUAL/$EDITOR (falling back to vi). On save the
file is re-parsed: a frontmatter break reopens the editor with the error rather than
landing on disk. The human counterpart to `research set` / `research append`.

```
tskflwctl research edit <research> [flags]
```

### Examples

```
  tskflwctl research edit theming-libs
  tskflwctl research edit   # pick from a list
```

### Options

```
  -h, --help   help for edit
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path (conflicts with --space)
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
      --space string   select a registered entry point by label (also TSKFLW_SPACE; conflicts with -C)
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl research](tskflwctl_research.md)	 - Work with research docs

