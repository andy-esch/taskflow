## tskflwctl task start

Move task(s) to in-progress

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
  -h, --help   help for start
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

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

