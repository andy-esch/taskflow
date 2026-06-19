## tskflwctl schema

Describe the tool's contract + per-kind authoring guidance (for agents)

### Synopsis

With no argument, emit the machine contract — statuses, the epic/bucket
enums, the task field registry with types, and the exit/error codes — so an
agent can drive the tool without parsing --help prose. With a kind, emit how
to author that document: the body section template, per-field guidance, and
conventions. Everything is derived from the tool's own data.

```
tskflwctl schema [task|epic|audit] [flags]
```

### Examples

```
  tskflwctl schema --json
  tskflwctl schema task
  tskflwctl schema audit --json
```

### Options

```
  -h, --help   help for schema
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

