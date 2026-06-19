## tskflwctl lint

Validate active task frontmatter (--fix to auto-repair)

```
tskflwctl lint [flags]
```

### Examples

```
  tskflwctl lint
  tskflwctl lint --fix --dry-run
  tskflwctl lint --json
```

### Options

```
      --fix    auto-repair frontmatter (quote ':' values, normalize list fields)
  -h, --help   help for lint
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

