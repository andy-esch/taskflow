## tskflwctl space

Manage the registry of planning repos on this machine

### Synopsis

The spaces registry records which planning repos exist on this machine, so
they can be addressed by name instead of by path.

It is ADVISORY: nothing in it changes what a command run in a directory
resolves to. With no registry, everything behaves exactly as it did before one
existed, and deleting it costs convenience — never data, never addressability.

### Options

```
  -h, --help   help for space
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits, research) over markdown
* [tskflwctl space add](tskflwctl_space_add.md)	 - Register a planning repo (defaults to the current directory)
* [tskflwctl space forget](tskflwctl_space_forget.md)	 - Drop a space from the registry (the repo itself is untouched)
* [tskflwctl space list](tskflwctl_space_list.md)	 - List registered planning repos with their current health

