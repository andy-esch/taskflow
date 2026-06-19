## tskflwctl epic new

Create a new epic (auto-numbered NN-slug)

```
tskflwctl epic new <title> [flags]
```

### Examples

```
  tskflwctl epic new "Billing overhaul" --description "Replace the legacy pipeline"
```

### Options

```
      --body string          override the default body scaffold
      --body-file string     read the body from a file, or - for stdin (replaces --body)
      --description string   one-line description (required, <=150 chars)
  -h, --help                 help for new
      --priority string      high|medium|low (default "medium")
      --status string        epic status: planning|in-progress|completed|archived (default "planning")
      --tags strings         comma-separated tags
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

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

