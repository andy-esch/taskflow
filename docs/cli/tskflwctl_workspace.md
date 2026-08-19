## tskflwctl workspace

Print the planning tree this directory resolves to

### Synopsis

Print the planning tree a command run from here would read and write.

External-planning routing is deliberately transparent: a repo whose
.tskflwctl.toml carries a planning_repo resolves into ANOTHER repo, so the
directory you are standing in is not necessarily the tree you would change.
This reports the resolved root, the config that selected it, and which
mechanism won — cheaply, before a mutation rather than after.

```
tskflwctl workspace [flags]
```

### Examples

```
  tskflwctl workspace
  tskflwctl workspace --json
  tskflwctl task complete foo --expect-root "$(tskflwctl workspace --json | jq -r .workspace.planning_root)"
```

### Options

```
  -h, --help   help for workspace
```

### Options inherited from parent commands

```
  -C, --chdir string         anchor to the planning repo at this path
      --color string         colorize output: auto|always|never (default "auto")
      --dry-run              preview the mutation without writing (validation still runs)
      --expect-root string   fail (exit 14) unless this directory resolves to this planning root — a wrong-repo write guard for agents
      --json                 machine-readable JSON output
      --no-color             disable colored output (alias for --color=never)
      --no-input             never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager             do not pipe long human output through a pager
      --paginate             page long human output through $PAGER (on a TTY), even if disabled in config
      --theme string         color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits, research) over markdown

