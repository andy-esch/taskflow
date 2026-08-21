## tskflwctl config show

Show repository, user, and effective configuration with provenance

### Synopsis

Show raw repository and user scopes separately, followed by each effective
theme and pager value and the source that won precedence. Also reports planning
topology, durable identity, config paths, tracked repos, and pending migrations.

```
tskflwctl config show [flags]
```

### Options

```
  -h, --help   help for show
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

* [tskflwctl config](tskflwctl_config.md)	 - Inspect, migrate, diagnose, and edit configuration

