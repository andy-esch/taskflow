## tskflwctl space

Manage planning spaces and their registered entry points

### Synopsis

The spaces registry records which repo checkouts enter each planning space, so
they can be addressed by name instead of by path. Direct planning checkouts and
implementation pointers with the same durable planning id form one logical space.

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
* [tskflwctl space add](tskflwctl_space_add.md)	 - Register a planning entry point (defaults to the current directory)
* [tskflwctl space forget](tskflwctl_space_forget.md)	 - Drop an entry point from the registry (the repo itself is untouched)
* [tskflwctl space list](tskflwctl_space_list.md)	 - List planning spaces, entry points, and current health

