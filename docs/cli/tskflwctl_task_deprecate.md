## tskflwctl task deprecate

Move task(s) to deprecated

```
tskflwctl task deprecate <task>... [flags]
```

### Examples

```
  tskflwctl task deprecate my-task
  tskflwctl task deprecate task-a task-b
```

### Options

```
  -h, --help   help for deprecate
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
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

