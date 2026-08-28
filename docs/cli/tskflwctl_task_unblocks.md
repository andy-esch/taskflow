## tskflwctl task unblocks

Show every task transitively downstream of this task

### Synopsis

Show the queried task's current derived state and every transitive downstream task with deterministic shortest paths. Resolved legacy constraints participate in the projection. This is current impact, not a promise that completing the source alone makes every result eligible.

```
tskflwctl task unblocks <task> [flags]
```

### Examples

```
  tskflwctl task unblocks build
  tskflwctl task unblocks build --json
```

### Options

```
  -h, --help   help for unblocks
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

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

