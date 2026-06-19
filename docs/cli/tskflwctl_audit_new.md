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
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
```

### SEE ALSO

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

