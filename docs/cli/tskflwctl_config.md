## tskflwctl config

Inspect, migrate, diagnose, and edit configuration

### Synopsis

Inspect and maintain repository and user configuration. Bare `config` is
the deterministic alias for `config show`; it never changes behavior based on
whether stdin is a terminal.

```
tskflwctl config [flags]
```

### Options

```
  -h, --help   help for config
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
* [tskflwctl config doctor](tskflwctl_config_doctor.md)	 - Audit linkback integrity and the home space registry
* [tskflwctl config edit](tskflwctl_config_edit.md)	 - Edit safe user or repository preferences interactively
* [tskflwctl config migrate](tskflwctl_config_migrate.md)	 - Apply safe, idempotent configuration upgrades
* [tskflwctl config show](tskflwctl_config_show.md)	 - Show repository, user, and effective configuration with provenance

