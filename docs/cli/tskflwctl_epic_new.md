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
      --template string      body scaffold to use (default "default"); completes the available names
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
```

### SEE ALSO

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

