## tskflwctl thread plan

Show explanatory member dependency waves and external gates

### Synopsis

Rank Thread members into deterministic topological waves and list immediate external gates separately. Member waves preserve ordering paths through included gates without treating those gates as Thread-owned work. Waves explain dependency order; they do not authorize dispatch or impose execution barriers. Under --json, emit the same neutral projection used by graph export.

```
tskflwctl thread plan <thread> [flags]
```

### Options

```
  -h, --help   help for plan
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

