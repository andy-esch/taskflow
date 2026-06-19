---
status: in-progress
epic: 20-cli-ux-and-ergonomics
description: cobra/doc GenMarkdownTree to docs/cli with DisableAutoGenTag; CI fails on drift so the command reference never goes stale (LLM-readable)
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, docs, dx]
created: "2026-06-19"
updated_at: "2026-06-19"
started_at: "2026-06-19"
---
## Objective

Auto-generate the CLI command reference from the cobra tree instead of hand-writing
it, and gate drift in CI. cobra's `doc.GenMarkdownTree(root, "docs/cli")` emits one
markdown page per command (flags, examples, subcommands); `root.DisableAutoGenTag =
true` makes output reproducible (no timestamp footer). A CI step regenerates and
fails on `git diff`, so the reference can never go stale. cobra.dev's
["LLM-ready CLI docs"](https://cobra.dev/docs/how-to-guides/clis-for-llms/) guide
recommends exactly this — predictable one-command-per-file structure is ideal for
agents, and it complements our existing `schema` self-description.

This pays down doc upkeep and reinforces the agent-first thesis (a always-current,
machine-chunkable command reference).

## Acceptance criteria

- [ ] A small generator (e.g. `internal/tools/docgen` or `cmd/docgen`) writes
      `docs/cli/*.md` from the root command with `DisableAutoGenTag = true`.
- [ ] A `just docs` (or make) target regenerates; output is deterministic across runs.
- [ ] CI regenerates and fails if the working tree differs (drift guard).
- [ ] Examples are rich enough to be useful (we already populate `Example`); audit
      for gaps while wiring this up.
- [ ] Watch the hyphen-in-command-names caveat from cobra/doc (our verbs are flat,
      so likely fine — confirm).

## Out of scope

- Manpage/ReST generation (fang's `man` would cover manpages — see the fang task).
- Publishing to a docs site (just commit the markdown for now).

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Pairs with [[publish-json-schema-for-the-json-envelopes]] (the machine contract;
  this is the human/agent prose reference).
