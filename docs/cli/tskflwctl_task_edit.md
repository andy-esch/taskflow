## tskflwctl task edit

Open a task in your editor (whole file; re-validated on save)

### Synopsis

Open the task's markdown file in $VISUAL/$EDITOR (falling back to vi). On
save the file is re-parsed: a frontmatter break or bad field reopens the
editor with the error rather than landing on disk. The human counterpart to
`task set` — agents and scripts should drive `set` (deterministic) instead.

```
tskflwctl task edit <task> [flags]
```

### Examples

```
  tskflwctl task edit add-retry-backoff
  tskflwctl task edit   # pick from a list
```

### Options

```
  -h, --help   help for edit
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

