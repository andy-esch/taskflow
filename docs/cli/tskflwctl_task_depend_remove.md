## tskflwctl task depend remove

Remove one or more hard prerequisites

```
tskflwctl task depend remove <task> [flags]
```

### Examples

```
  tskflwctl task depend remove deploy --on build --on verify
```

### Options

```
  -h, --help         help for remove
      --on strings   prerequisite task reference (repeat or comma-separate)
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

