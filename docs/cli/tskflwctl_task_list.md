## tskflwctl task list

List tasks (active by default)

```
tskflwctl task list [flags]
```

### Examples

```
  tskflwctl task list
  tskflwctl task list -q --tag tui | xargs tskflwctl task start
  tskflwctl task list -o table -c slug,status,epic
  tskflwctl task list --unblocked --json
  tskflwctl task list --revisit-due -q | xargs tskflwctl task next   # resume snoozed tasks now due
```

### Options

```
      --all                    include completed/deprecated/deferred
  -c, --columns strings        select columns for -o table/csv/json, comma-separated (-o table when no format is pinned); available: slug,status,tier,priority,epic,updated_at,description,revisit_at,id,location
      --cursor string          resume page.next_cursor with the same filters, sort, and limit; restart if results changed
      --epic string            filter by epic
  -h, --help                   help for list
      --limit string           maximum returned rows (integer 1-1000; omitted means unbounded)
  -o, --output string          output format: human|json|name|table|csv
  -q, --quiet                  concise command handles, one per line (alias for -o name)
      --revisit-due            only deferred tasks whose revisit date has arrived (composes with --epic/--tag/-c)
      --search string          case-insensitive substring in list metadata, not full bodies (maximum 256 UTF-8 bytes)
      --sort string            sort by id|name|created|updated; prefix with - for descending (findings: id|name only)
      --status string          filter by status
      --tag string             filter by tag
      --unblocked              only next-up or ready-to-start tasks with a clear dependency gate
      --updated-since string   own updated date >= YYYY-MM-DD, falling back to creation; not a complete change feed
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

