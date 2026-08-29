## tskflwctl task start

Move task(s) to in-progress

### Synopsis

Move task(s) to in-progress.

Refuses a task unless it is ready-to-start with a clear dependency gate. --force
bypasses only that dependency gate; it does not bypass lifecycle role or repair
the dependencies, and the receipt names every outstanding blocker.

```
tskflwctl task start <task>... [flags]
```

### Examples

```
  tskflwctl task start my-task
  tskflwctl task start task-a task-b
```

### Options

```
      --force   start despite outstanding dependency blockers
  -h, --help    help for start
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

