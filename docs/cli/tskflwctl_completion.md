## tskflwctl completion

Generate the autocompletion script for the specified shell

### Synopsis

Generate the autocompletion script for tskflwctl for the specified shell.
See each sub-command's help for details on how to use the generated script.


### Options

```
  -h, --help   help for completion
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, Threads, epics, audits, research) over markdown
* [tskflwctl completion bash](tskflwctl_completion_bash.md)	 - Generate the autocompletion script for bash
* [tskflwctl completion fish](tskflwctl_completion_fish.md)	 - Generate the autocompletion script for fish
* [tskflwctl completion powershell](tskflwctl_completion_powershell.md)	 - Generate the autocompletion script for powershell
* [tskflwctl completion zsh](tskflwctl_completion_zsh.md)	 - Generate the autocompletion script for zsh

