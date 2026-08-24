## tskflwctl task complete

Move task(s) to completed

### Synopsis

Move task(s) to completed.

Refuses a task whose acceptance criteria are still unmet with no reason given —
the task counterpart of `audit close` refusing while findings are open. A criterion
carrying a state (`task ac --defer|--wontfix|--tracked|--na`) has been DECIDED and
does not block; only a silently unticked box does. --force completes anyway.

```
tskflwctl task complete <task>... [flags]
```

### Examples

```
  tskflwctl task complete my-task
  tskflwctl task complete task-a task-b
```

### Options

```
      --force   complete even with unmet, unexplained acceptance criteria
  -h, --help    help for complete
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

