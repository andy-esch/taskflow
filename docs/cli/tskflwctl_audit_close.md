## tskflwctl audit close

Move audit(s) to closed/

```
tskflwctl audit close <audit>... [flags]
```

### Examples

```
  tskflwctl audit close 2026-06-06-schemas-scripts
```

### Options

```
  -h, --help   help for close
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
```

### SEE ALSO

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

