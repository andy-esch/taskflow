## tskflwctl init

Scaffold a planning tree (tasks/ epics/ projects/ audits/) + config

```
tskflwctl init [flags]
```

### Examples

```
  tskflwctl init
  tskflwctl init --path ./planning
```

### Options

```
  -h, --help          help for init
      --path string   directory to initialize (default ".")
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

