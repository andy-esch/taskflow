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
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
```

### SEE ALSO

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits
* [tskflwctl doctor](tskflwctl_doctor.md)	 - Audit planning_repo <-> tracked_repos linkback integrity
* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics
* [tskflwctl init](tskflwctl_init.md)	 - Scaffold a planning tree here, or point at an external planning repo
* [tskflwctl lint](tskflwctl_lint.md)	 - Validate active task and epic frontmatter (--fix auto-repairs tasks)
* [tskflwctl schema](tskflwctl_schema.md)	 - Describe the tool's contract + per-kind authoring guidance (for agents)
* [tskflwctl status](tskflwctl_status.md)	 - At-a-glance project dashboard (counts, in-progress, epic progress)
* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks
* [tskflwctl template](tskflwctl_template.md)	 - List and inspect the body scaffolds `new --template` can use
* [tskflwctl ui](tskflwctl_ui.md)	 - Launch the interactive TUI (Bubble Tea)
* [tskflwctl version](tskflwctl_version.md)	 - Print the tskflwctl version

