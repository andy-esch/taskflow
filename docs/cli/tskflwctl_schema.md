## tskflwctl schema

Describe the tool's contract + per-kind authoring guidance (for agents)

### Synopsis

For triage, lead with the terse path: `epic show <id>` for an epic's task
roster, and `task list -o table -c slug,status,description` for a compact,
byte-stable table. --json is compact and also takes -c to project just the
fields you need; reach for full --json (no -c) when you need every frontmatter
field. A --json -c projection is a string-valued column view — only full --json
validates against --json-schema.

With no argument, emit the machine contract — statuses, the epic/bucket
enums, the task field registry with types, and the exit/error codes — so an
agent can drive the tool without parsing --help prose. With a kind, emit how
to author that document: the body section template, per-field guidance, and
conventions. With --json-schema, emit a JSON Schema for the full --json output
envelopes so an agent can validate the tool's output. Everything is derived
from the tool's own types and data.

```
tskflwctl schema [task|epic|audit] [flags]
```

### Examples

```
  # Triage first: cheap, scannable views
  tskflwctl epic show <id>
  tskflwctl task list -o table -c slug,status,description
  # Full frontmatter only when you need every field:
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
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

