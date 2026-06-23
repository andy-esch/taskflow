## tskflwctl init

Scaffold a planning tree here, or point at an external planning repo

```
tskflwctl init [flags]
```

### Examples

```
  tskflwctl init
  tskflwctl init --path ./planning
  tskflwctl init --planning-repo ../desirelines-planning
```

### Options

```
  -h, --help                   help for init
      --path string            directory to initialize (default ".")
      --planning-repo string   point this repo at an external planning repo (relative to --path, or absolute): writes a pointer config, no tree
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

