## tskflwctl status

At-a-glance project dashboard (counts, in-progress, epic progress)

```
tskflwctl status [flags]
```

### Examples

```
  tskflwctl status
  tskflwctl status --json
```

### Options

```
  -h, --help   help for status
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

