## tskflwctl schema

Describe the tool's contract + per-kind authoring guidance (for agents)

### Synopsis

With no argument, emit the machine contract — statuses, the epic/bucket
enums, the task field registry with types, and the exit/error codes — so an
agent can drive the tool without parsing --help prose. With a kind, emit how
to author that document: the body section template, per-field guidance, and
conventions. With --json-schema, emit a JSON Schema for the --json output
envelopes so an agent can validate the tool's output. Everything is derived
from the tool's own types and data.

```
tskflwctl schema [task|epic|audit] [flags]
```

### Examples

```
  tskflwctl schema --json
  tskflwctl schema task
  tskflwctl schema --json-schema
```

### Options

```
  -h, --help          help for schema
      --json-schema   emit a JSON Schema (Draft 2020-12) for the --json output envelopes
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

