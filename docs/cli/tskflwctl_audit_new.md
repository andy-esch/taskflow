## tskflwctl audit new

Create a new audit (open bucket, scaffolded findings)

```
tskflwctl audit new <area> [flags]
```

### Examples

```
  tskflwctl audit new dispatcher
  tskflwctl audit new arch-data-flow --date 2026-06-16
```

### Options

```
      --body string        override the default scaffold
      --body-file string   read the body from a file, or - for stdin (replaces --body)
      --date string        audit date YYYY-MM-DD (default today)
  -h, --help               help for new
      --template string    body scaffold to use (default "default"); e.g. "security". completes the available names
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

