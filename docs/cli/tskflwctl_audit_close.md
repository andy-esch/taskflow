## tskflwctl audit close

Move audit(s) to the closed bucket

### Synopsis

Move audit(s) to the closed bucket. Close/defer refuse while parsed open findings or unparsed finding-like
headers remain, including ambiguous headings. Run audit lint <audit> and repair or
clarify the headings first; there is no --force bypass. Reopen remains available
for repair. Empty audits may close/defer. Dry-run applies the same validation.

```
tskflwctl audit close <audit>... [flags]
```

### Examples

```
  tskflwctl audit close 2026-06-06-schemas-scripts
  tskflwctl audit close   # pick from a list
```

### Options

```
  -h, --help   help for close
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

