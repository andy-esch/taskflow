## tskflwctl audit info

Show an audit's metadata + file path + finding tally (no body)

```
tskflwctl audit info <audit> [flags]
```

### Examples

```
  tskflwctl audit show 2026-06-20-api-gateway --frontmatter-only
  tskflwctl audit info 2026-06-20-api-gateway --json
```

### Options

```
  -h, --help   help for info
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

