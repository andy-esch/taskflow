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
      --no-link-back           pointer mode: don't add this repo to the planning repo's tracked_repos
      --path string            directory to initialize (default ".")
      --planning-repo string   point this repo at an external planning repo (relative to --path, or absolute): writes a pointer config, no tree
      --track strings          record an impl repo this planning repo tracks (repeatable; scaffold mode only)
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
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

