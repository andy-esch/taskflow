## tskflwctl thread compose

Compile existing tasks and dependency edges into a durable Thread apply plan

### Synopsis

Read one strict literal YAML/JSON manifest, resolve its exact stable task IDs, and validate the proposed global DAG without mutation. A real run creates a no-clobber materialized plan; --dry-run prints the same plan without creating the output file.

```
tskflwctl thread compose [flags]
```

### Options

```
      --from string   strict authoring manifest path, or - for stdin
  -h, --help          help for compose
      --out string    new durable apply-plan path (must not already exist)
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

* [tskflwctl thread](tskflwctl_thread.md)	 - Work with initiative Threads over the task DAG

