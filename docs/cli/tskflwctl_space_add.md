## tskflwctl space add

Register a planning entry point (defaults to the current directory)

### Synopsis

Register a direct planning checkout or implementation pointer as an entry point.

The path is VALIDATED as a planning repo before anything is written, so a typo
fails with nothing left behind. Registering the same path twice is a no-op —
identity is the physical directory, so relative, absolute and symlinked
spellings of one repo collapse to a single entry.

```
tskflwctl space add [path] [flags]
```

### Examples

```
  tskflwctl space add
  tskflwctl space add ~/git/andy-esch/desirelines
  tskflwctl space add ../other --id other
```

### Options

```
  -h, --help        help for add
      --id string   label to address this entry point by (default: the directory name)
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

* [tskflwctl space](tskflwctl_space.md)	 - Manage planning spaces and their registered entry points

