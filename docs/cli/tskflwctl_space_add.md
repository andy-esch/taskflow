## tskflwctl space add

Register a planning repo (defaults to the current directory)

### Synopsis

Register a planning repo so it can be addressed by name.

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
      --id string   label to address this space by (default: the directory name)
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
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl space](tskflwctl_space.md)	 - Manage the registry of planning repos on this machine

