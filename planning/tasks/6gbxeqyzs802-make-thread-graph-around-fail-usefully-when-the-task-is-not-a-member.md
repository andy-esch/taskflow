---
schema: 1
id: 6gbxeqyzs802
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Name the Thread, distinguish missing from non-member from stale projection, and print the thread add command that fixes it.
effort: XS
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, ux]
created: "2026-09-20"
---
# Make `thread graph --around` fail usefully when the task is not a member

## Objective

`thread graph --around <task>` refuses a task that is not a member of the Thread with a message
that names neither the Thread nor the fix. The refusal itself is right: it exits 10, writes to
stderr, and prints nothing to stdout. What is missing is the sentence that turns the refusal into
an action, which matters most for the agents that call this command while assembling a document.

## What happened

Assembling a pull-request body in an implementation repo, an agent ran:

```
tskflwctl thread graph 6g9nftpx0cwc --around 6gbpgbrn57m9 --depth 1 --format mermaid
```

and got:

```
error: task "6gbpgbrn57m9" is not a node in the supplied Thread graph: not found
```

The task existed and was in the same epic as the Thread's members; it had simply never been added
to the Thread, and it had no dependency edges to any member. Neither fact is visible in the
message, so the reader is left to guess between a bad id, a bad Thread, a stale projection, and a
membership gap.

The agent had captured the command with `2>&1`, pasted the result into a ```mermaid fence, and
published it, so GitHub rendered "No diagram type detected matching given configuration for text:
error: task ... is not found". That part is the caller's bug, not this one: the CLI put the error
on stderr and left stdout empty, which is exactly right. It is recorded here only because it shows
what a terse refusal costs downstream, and because a message naming the fix would have been caught
in review rather than in a published document.

## Scope

- Name the Thread in the message, not just the task: which Thread was searched, by id and slug.
- Distinguish the cases the current wording merges:
  - the task does not exist at all
  - the task exists but is not a member of this Thread
  - the task is a member but the projection is stale
- For the membership case, name the command that fixes it (`thread add <thread> <task>`), in the
  same shape as the `--on` hint `task depend add` already prints. That hint is a good model: it
  restates the corrected command in full.
- Consider whether a non-member with dependency edges into the Thread should render as an external
  gate rather than refusing, since `graph` already draws external prerequisites in amber. If so,
  that is a behavior change and wants its own decision, not a silent default.

## Acceptance criteria

- [ ] `thread graph --around` names the Thread and the task in its refusal
- [ ] Missing task, non-member task, and stale projection produce distinguishable messages
- [ ] The non-member message prints the `thread add` command that would fix it
- [ ] Exit code stays 10, stderr stays the only stream written, stdout stays empty
- [ ] Tests cover all three cases, including the exact hint text
- [ ] Docs and generated CLI reference updated; tests, lint, and planning lint pass

## Out of scope

- Changing what `graph` renders for members, or the Mermaid and DOT output itself
- Auto-adding a task to a Thread as a side effect of asking for a graph

## Notes for whoever picks this up

Two adjacent observations from the same session, both optional:

- Nothing prompts an agent to add a follow-up task to the Thread its parent belongs to. A
  `thread add` hint on `task new`, when the new task's declared prerequisite is a Thread member,
  would have prevented this entirely.
- A one-hop graph around a freshly created task renders a single node with no edges, which reads as
  a broken diagram. Rendering a note in that case ("no neighbours at depth 1") would be clearer
  than an isolated node.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
