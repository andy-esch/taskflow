## tskflwctl task demote

Move task(s) to ready-to-start

```
tskflwctl task demote <task>... [flags]
```

### Examples

```
  tskflwctl task demote my-task
  tskflwctl task demote task-a task-b
```

### Options

```
  -h, --help   help for demote
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

