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
```

### SEE ALSO

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

