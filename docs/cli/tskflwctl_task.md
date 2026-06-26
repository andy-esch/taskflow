## tskflwctl task

Work with tasks

### Options

```
  -h, --help   help for task
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
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown
* [tskflwctl task append](tskflwctl_task_append.md)	 - Append a section to a task's body (atomic; agent-facing)
* [tskflwctl task complete](tskflwctl_task_complete.md)	 - Move task(s) to completed
* [tskflwctl task defer](tskflwctl_task_defer.md)	 - Move task(s) to deferred (optionally with a revisit date)
* [tskflwctl task demote](tskflwctl_task_demote.md)	 - Move task(s) to ready-to-start
* [tskflwctl task deprecate](tskflwctl_task_deprecate.md)	 - Move task(s) to deprecated
* [tskflwctl task edit](tskflwctl_task_edit.md)	 - Open a task in your editor (whole file; re-validated on save)
* [tskflwctl task list](tskflwctl_task_list.md)	 - List tasks (active by default)
* [tskflwctl task move](tskflwctl_task_move.md)	 - Transition task(s) to <status> (generic escape hatch)
* [tskflwctl task new](tskflwctl_task_new.md)	 - Create a new task (validated, handoff-ready scaffold)
* [tskflwctl task promote](tskflwctl_task_promote.md)	 - Move task(s) to next-up
* [tskflwctl task set](tskflwctl_task_set.md)	 - Set one or more frontmatter fields (validated, single atomic write)
* [tskflwctl task show](tskflwctl_task_show.md)	 - Show a task's metadata and body
* [tskflwctl task start](tskflwctl_task_start.md)	 - Move task(s) to in-progress

