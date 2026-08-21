## tskflwctl config doctor

Audit linkback integrity and the home space registry

### Synopsis

Audit the cross-repo links: an impl repo's planning_repo pointer should be
matched by the planning repo tracking it back, and every tracked_repos entry
should exist and point its planning_repo back here. Also diagnose every home
space-registry entry. Reports each inconsistency and exits non-zero when any
is found — usable as a CI gate; nothing is repaired or forgotten.

```
tskflwctl config doctor [flags]
```

### Examples

```
  tskflwctl config doctor
  tskflwctl config doctor --json
```

### Options

```
  -h, --help   help for doctor
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

* [tskflwctl config](tskflwctl_config.md)	 - Inspect, migrate, diagnose, and edit configuration

