---
schema: 1
id: 6g0fzhbtk0gy
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'Add the [[space]] array to spaces.toml plus space list|add|forget. Registry stays advisory: nothing in it may change what Discover resolves from a cwd. Vocabulary settled; entry schema is not.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [config, cli, multi-repo]
created: "2026-08-15"
updated_at: "2026-08-20"
started_at: "2026-08-20"
completed_at: "2026-08-20"
---
# Space registry model and the `space` verbs (list/add/forget)

## Objective

Add the `[[space]]` array to the home config plus the CLI surface to manage it: the
user-level record of which planning repos exist on this machine.

The **vocabulary is settled** (`space` / `atlas`, decided 2026-08-18 — see
[decide-the-multi-space-vocabulary](6g1erb0p5893-decide-the-multi-space-vocabulary-blocks-slice-2.md)),
so `[[space]]`, `space list|add|forget`, and `--space` are the committed spellings. The
**schema** below is still where the remaining calls get made — id collision policy, entry
ordering, and whether `space.id` becomes a durable minted identity (see the 2026-08-18
note at the end).

## Notes

Sketch schema (expect it to move):

```toml
[[space]]
id     = "taskflow"                    # stable key; how you address it
path   = "~/git/andy-esch/taskflow"    # the repo dir, NOT the resolved planning root
label  = "tskflwctl"                   # optional display name
accent = "cyan"                        # optional; else derived from id
added  = "2026-08-15"
```

- **Anchor `path` at the repo** (the `.tskflwctl.toml` dir), not the resolved root:
  registration then means one thing in both config modes, and a pointer repo resolves
  through its `planning_repo` to the real root via ordinary discovery. Epic 23
  compatibility falls out for free.
- **`id` is a stable key** (ADR-0003's philosophy), defaulted from the dir basename.
  Collisions resolved at registration, never silently.
- **Dedup on PHYSICAL paths**, reusing the helper `tracked_repos` already dedups with,
  so `../x`, an absolute path, and a symlinked checkout collapse to one entry.
- **Store `~` unexpanded** so the file stays portable and committable.
- **Atomic + surgical writes** — `setTrackedReposInText` is the model.
- `space add` validates through discovery **before** writing anything (the
  "require + error, leave nothing behind" contract `InitPointer` already follows).
- `space forget` never touches the repo on disk — it only drops the entry.

The invariant worth protecting: **the registry is advisory.** Nothing in it may
change what `Discover` resolves from a given cwd. No home config ⇒ today's behavior
exactly; deleting the file costs convenience, never data or addressability.

## Acceptance criteria

- [x] `space list` (human + `--json` with `schema_version`) shows id, path, state
- [x] `space add [path]` validates via discovery before writing; a bad path errors
      with nothing left behind
- [x] `space forget <id>` removes only the entry; the repo is untouched
- [x] Physical-path dedup: `../x`, an absolute path, and a symlink resolve to one entry
- [x] Surgical write: comments, key order, and unknown keys survive a round trip
- [x] Local discovery behavior is provably unchanged with and without a registry
- [x] `just test` + `just lint` green

## Out of scope

- `--space` (separate task), `init` auto-registration (separate task)
- `space scan` — walking the filesystem for planning repos
- Any TUI surface

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)

### 2026-08-18 — inherited decision from audit H1

Audit 2026-07-24 H1 (the `workspace` read + receipt) deliberately defined a
workspace's identity as its **absolute resolved planning root** — enough to prove which
tree a mutation touched, but it does NOT survive moving the repo.

`space.id` here is the natural home for a **durable** identity if one is wanted: an id
minted into the planning repo's own `.tskflwctl.toml` at `init`, which a registry entry
then references instead of matching on path. That would also make `moved` detection
trivial and honest rather than a heuristic (open fork 6).

Decide it with this task, not separately — the alternative is inventing two identities
for the same thing. If a durable id is adopted, `wire.WorkspaceJSON` should carry it too
so the two surfaces agree.

### 2026-08-19 — the identity decision now has a fourth consumer, and a hard case

[worktrees-and-relative-paths-in-cross-repo-planning](../research/6g1emag02srv-worktrees-and-relative-paths-in-cross-repo-planning.md)
demonstrated that a committed relative `planning_repo` can **silently resolve to an
unrelated planning repo** when the checkout is not where the path assumed — no warning,
no error, every mutation landing in the wrong tree. Absolute paths are no escape; they
break on every other machine.

The only fix that closes it in the general case is **layer C**: mint a durable id into a
planning repo's own committed `.tskflwctl.toml`, have pointers record it alongside the
path, and **verify it after following the path**. Path = hint, id = assertion.

That is the same primitive as `space.id`, which now has **four** consumers:

1. linkback verification (epic 23),
2. `--space` selection (the wrong-repo guard that replaced `--expect-root`),
3. the `workspace` object on mutation receipts,
4. `moved` detection in the registry (open fork 6).

**So this task's identity choice is load-bearing well beyond the registry.** Deciding
`space.id` as a *path-keyed* identity does not merely make `moved` detection a
heuristic — it leaves the silent-wrong-tree hazard permanently open and the registry
inherits exactly the fragility the research doc documents.

Concretely, to settle here:

- Is `space.id` the **minted durable id** read from the target repo's config, or a
  local label keyed to a path?
- If durable: where is it minted (`init`), how are the ~2 existing repos backfilled
  (`lint --fix`? `doctor`?), and does `wire.WorkspaceJSON` carry it so the receipt and
  the registry agree?
- If not durable: record explicitly that the wrong-tree hazard is accepted, so it is a
  decision rather than an omission.

### 2026-08-19 — identity settled; the entry schema is now decided

[decide-space.id-identity](6g1m8mc8p46h-decide-space.id-identity-durable-minted-id-vs-path-keyed.md)
is closed. **Two identities**: path addresses the checkout, a durable committed id asserts
the repo. The entry shape is therefore:

```toml
[[space]]
id        = "taskflow"                  # LOCAL LABEL — the address, what --space takes
path      = "~/git/andy-esch/taskflow"  # the repo dir, not the resolved planning root
verify_id = "6g1durable0aa"             # the target repo's durable id, checked after resolving
label     = "tskflwctl"                 # optional display name
accent    = "cyan"                      # optional
added     = "2026-08-19"
```

`id` is a **local label**, not the durable id: every worktree of a repo shares one committed
id, so a durable id cannot address a specific checkout — and labels are typable.

Still open in this task (unchanged): **collision policy** on a derived label, and **entry
ordering** in the file.

**Dedup gains a second axis.** Two entries may now share a `verify_id` (two checkouts of one
repo) — which is legitimate and must not be collapsed. Dedup on **physical path**, never on
`verify_id`.

### 2026-08-19 — labels must be shell-completable, and that is part of why they won

`--space <label>` should complete from the registry, like every other addressable property
here (`completion.go` already does command / flag / status-aware slug completion).

This is a second, independent argument for the label over the durable id as the address:
a completion menu of `taskflow · desirelines · dotfiles` is usable, whereas one of
`6g1durable0aa · 6g1kzy4pgt2f` is not — the ids are not recognizable even when shown.
Completability is a property of the *address*, and it should be treated as a requirement of
this task, not a nicety bolted on later.

Constraint to carry over from `completion.go`: completion funcs do their own forgiving
discovery and must **not** fail outside a planning repo. Registry completion reads the home
config, which exists independently of any planning repo, so `--space` should complete even
where nothing else does.

### 2026-08-20 — implementation closeout

The two remaining schema calls are settled:

- **Derived-label collisions are refused.** Registration never invents a suffix; the error
  names the existing path and asks for an explicit `--id`.
- **Entries retain insertion order.** `space add` appends one table and `space forget`
  removes one table. Existing tables are never reordered or re-encoded, so comments, key
  order, unknown keys, whitespace, file mode, and a dotfiles-managed registry symlink all
  survive.

Registration stores the discovered repo marker directory rather than the caller's exact
subdirectory or the resolved planning root. That keeps pointer repos addressable as their
own checkout while ordinary discovery continues to route them to the external planning
tree. Physical path dedup covers relative, absolute, and symlink spellings; `verify_id`
records the resolved planning repo's durable id without collapsing multiple checkouts.

The CLI now covers human and schema-versioned JSON list output, explicit state and path
display, add/forget receipts, bad-path no-write behavior, label completion outside any
planning repo, and broken entries that remain listed. The JSON envelope registry, generated
schema comments, machine-contract goldens, and CLI reference are updated for schema 1.37.
Focused tests also compare `config.Discover` before and after populating the registry to pin
the advisory invariant directly.

### 2026-08-20 — adversarial review amendment

The post-implementation review found and closed four correctness gaps. Add and forget dry-runs now execute the same snapshot planning and validation as real mutations, so path dedup is a true no-op, label collisions remain conflicts, previews return the exact would-be entry, and neither creates or rewrites the registry. Add/forget read-modify-write transactions now take the same directory-level Unix flock pattern already used by the repository store, eliminating cooperating-writer lost updates and stale decoded-index edits on the darwin/linux release targets.

Surgical forget scanning is now quote-aware for TOML headers, so a hash inside a quoted key cannot hide an unrelated following table; ambiguous blank/comment trivia before the next table and trailing file comments are preserved. Registry content/collision errors are typed at the userconfig boundary and mapped to validation/conflict by the CLI, while filesystem and lock failures retain generic operational status. Regressions cover CLI preview semantics and byte-for-byte no-write behavior, malformed registry classification, quoted-table/comment preservation, and concurrent additions with no lost entries.

No new decision follow-up is needed: the cross-platform lock choice follows the repository existing lock design and published darwin/linux target set. Durable verify_id enforcement and fuller health diagnosis remain owned by their existing follow-up tasks rather than being widened into this task.
