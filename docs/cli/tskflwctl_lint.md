## tskflwctl lint

Validate entity frontmatter, audit findings, and task-dependency graph integrity

### Synopsis

Validate task, epic, research, and Thread frontmatter plus audit findings, then validate the
repository-global task-dependency graph. Exactly resolved legacy dependency fields are visible
advisories; missing, ambiguous, or structurally unsafe references are errors.

--fix repairs ordinary frontmatter and missing ids. It never normalizes or changes
graph-owned task fields (depends_on, blocked_by, dependencies, or blocks); a
would-be graph repair is skipped and reported for deliberate remediation.

```
tskflwctl lint [flags]
```

### Examples

```
  tskflwctl lint
  tskflwctl lint --fix --dry-run
  tskflwctl lint --links
  tskflwctl lint --json
```

### Options

```
      --fix     auto-repair ordinary frontmatter and missing ids; graph-owned task fields are skipped
  -h, --help    help for lint
      --links   also check body cross-links: flag any [..](path.md) whose target file is missing (opt-in — a tree can carry pre-existing danglers)
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

