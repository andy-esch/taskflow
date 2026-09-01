## tskflwctl thread

Work with initiative Threads over the task DAG

### Options

```
  -h, --help   help for thread
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
* [tskflwctl thread add](tskflwctl_thread_add.md)	 - Atomically add task members to a Thread
* [tskflwctl thread apply](tskflwctl_thread_apply.md)	 - Converge dependencies and one new Thread from a durable plan
* [tskflwctl thread cancel](tskflwctl_thread_cancel.md)	 - Cancel a Thread without changing member tasks
* [tskflwctl thread complete](tskflwctl_thread_complete.md)	 - Complete a soundly drained Thread
* [tskflwctl thread compose](tskflwctl_thread_compose.md)	 - Compile existing tasks and dependency edges into a durable Thread apply plan
* [tskflwctl thread frontier](tskflwctl_thread_frontier.md)	 - Show active work and graph-clear pending members
* [tskflwctl thread graph](tskflwctl_thread_graph.md)	 - Export a deterministic Mermaid or DOT Thread graph
* [tskflwctl thread list](tskflwctl_thread_list.md)	 - List Threads with nominal and sound progress
* [tskflwctl thread new](tskflwctl_thread_new.md)	 - Create an unstarted Thread with optional initial task members
* [tskflwctl thread path](tskflwctl_thread_path.md)	 - Print the absolute path to a Thread file
* [tskflwctl thread plan](tskflwctl_thread_plan.md)	 - Show explanatory member dependency waves and external gates
* [tskflwctl thread remove](tskflwctl_thread_remove.md)	 - Atomically remove task members from a Thread
* [tskflwctl thread reopen](tskflwctl_thread_reopen.md)	 - Reopen a completed Thread
* [tskflwctl thread show](tskflwctl_thread_show.md)	 - Show Thread progress, members, gates, frontier, and body
* [tskflwctl thread start](tskflwctl_thread_start.md)	 - Start a Thread with at least one live member

