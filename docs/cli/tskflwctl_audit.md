## tskflwctl audit

Work with code audits

### Options

```
  -h, --help   help for audit
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
* [tskflwctl audit close](tskflwctl_audit_close.md)	 - Move audit(s) to closed/
* [tskflwctl audit defer](tskflwctl_audit_defer.md)	 - Move audit(s) to deferred/
* [tskflwctl audit findings](tskflwctl_audit_findings.md)	 - Query findings across audits (or one) by status/effort/urgency/component
* [tskflwctl audit lint](tskflwctl_audit_lint.md)	 - Validate audit findings (status vocabulary, missing status, bucket↔state)
* [tskflwctl audit list](tskflwctl_audit_list.md)	 - List audits (open by default)
* [tskflwctl audit new](tskflwctl_audit_new.md)	 - Create a new audit (open bucket, scaffolded findings)
* [tskflwctl audit reopen](tskflwctl_audit_reopen.md)	 - Move audit(s) back to open/
* [tskflwctl audit show](tskflwctl_audit_show.md)	 - Show an audit's metadata and body

