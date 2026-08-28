## tskflwctl task blockers

Explain the actionable blockers for a task

### Synopsis

Explain a task's current derived role, gate, eligibility, and actionable blocker frontier. --causal selects the full forensic closure. Resolved legacy constraints participate in both projections, while graph health still reports degraded until they are migrated.

```
tskflwctl task blockers <task> [flags]
```

### Examples

```
  tskflwctl task blockers deploy
  tskflwctl task blockers deploy --causal --json
```

### Options

```
      --causal   show the full causal blocker closure instead of the actionable frontier
  -h, --help     help for blockers
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

