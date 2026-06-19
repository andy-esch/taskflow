## tskflwctl

Local-first planning CLI (tasks, epics, audits) over markdown

### Options

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
  -h, --help           help for tskflwctl
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
```

### SEE ALSO

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits
* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics
* [tskflwctl init](tskflwctl_init.md)	 - Scaffold a planning tree (tasks/ epics/ projects/ audits/) + config
* [tskflwctl lint](tskflwctl_lint.md)	 - Validate active task frontmatter (--fix to auto-repair)
* [tskflwctl schema](tskflwctl_schema.md)	 - Describe the tool's contract + per-kind authoring guidance (for agents)
* [tskflwctl status](tskflwctl_status.md)	 - At-a-glance project dashboard (counts, in-progress, epic progress)
* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks
* [tskflwctl ui](tskflwctl_ui.md)	 - Launch the interactive TUI (Bubble Tea)
* [tskflwctl version](tskflwctl_version.md)	 - Print the tskflwctl version

