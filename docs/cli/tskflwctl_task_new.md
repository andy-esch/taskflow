## tskflwctl task new

Create a new task (validated, handoff-ready scaffold)

```
tskflwctl task new <title> [flags]
```

### Examples

```
  tskflwctl task new "Add retry backoff" --epic 17-pm-go-cli --tags net
  tskflwctl task new "Triage flaky test" --epic 17-pm-go-cli --next
```

### Options

```
      --autonomy int         autonomy level 1-5 (default 3)
      --body string          override the default body scaffold
      --body-file string     read the body from a file, or - for stdin (replaces --body)
      --description string   one-line description (<=150 chars)
      --effort string        effort estimate (default "Unknown")
      --epic string          epic id (required)
  -h, --help                 help for new
      --next                 create in next-up instead of ready-to-start
      --priority string      high|medium|low (default "medium")
      --start                create directly in in-progress
      --tags strings         comma-separated tags (at least one required)
      --template string      body scaffold to use (default "default"); completes the available names
      --tier int             tier 1-5 (default 3)
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

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

