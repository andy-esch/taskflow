## tskflwctl task depend repair

Diagnose or repair broken graph-owned declarations

### Synopsis

With no selector, diagnose exact graph-owned source defects and print copyable repair commands. --auto applies only canonical deduplication, self-edge removal, and empty legacy-key cleanup. Use repeatable --drop/--dedupe selectors or a YAML --plan for explicit, reauthorized source removals. Invalid and dangling values, cycle choices, and ambiguous legacy intent are never guessed.

```
tskflwctl task depend repair [flags]
```

### Examples

```
  tskflwctl task depend repair
  tskflwctl task depend repair --auto --dry-run
  tskflwctl task depend repair --drop 'tasks/ID-task.md:depends_on=RAW#0'
  tskflwctl task depend repair --plan repair.yaml --json
```

### Options

```
      --auto                 apply only canonical dedupe, self-edge, and empty legacy-key repairs
      --dedupe stringArray   deduplicate one canonical source value: <task-or-path>:depends_on=<task-id>
      --drop stringArray     drop an exact defective source declaration: <task-or-path>:<field>=<raw-value>[#occurrence]
  -h, --help                 help for repair
      --plan string          read additional convergent repair operations from YAML
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

* [tskflwctl task depend](tskflwctl_task_depend.md)	 - Change repository-global task dependencies through the graph guard

