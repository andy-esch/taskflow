## tskflwctl audit reopen

Move audit(s) back to open/

```
tskflwctl audit reopen <audit>... [flags]
```

### Examples

```
  tskflwctl audit reopen 2026-06-06-schemas-scripts
```

### Options

```
  -h, --help   help for reopen
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

