## tskflwctl task set

Set one or more frontmatter fields (validated, single atomic write)

```
tskflwctl task set <task> [flags]
```

### Options

```
      --autonomy int         autonomy level 1-5
      --description string   one-line description (<=150 chars)
      --effort string        effort estimate
      --epic string          epic id
      --force                allow --set of a field tskflwctl doesn't know
  -h, --help                 help for set
      --priority string      high|medium|low
      --set stringArray      key=value (repeatable); known fields are typed+validated, unknown keys need --force
      --tags strings         comma-separated tags
      --tier int             tier 1-5
      --unset stringArray    remove a frontmatter key (repeatable)
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

