---
schema: 1
id: 6g0fzhbtk0gy
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'Add the [[space]] array to spaces.toml plus space list|add|forget. Registry stays advisory: nothing in it may change what Discover resolves from a cwd. Vocabulary settled; entry schema is not.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [config, cli, multi-repo]
created: "2026-08-15"
updated_at: "2026-08-19"
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

- [ ] `space list` (human + `--json` with `schema_version`) shows id, path, state
- [ ] `space add [path]` validates via discovery before writing; a bad path errors
      with nothing left behind
- [ ] `space forget <id>` removes only the entry; the repo is untouched
- [ ] Physical-path dedup: `../x`, an absolute path, and a symlink resolve to one entry
- [ ] Surgical write: comments, key order, and unknown keys survive a round trip
- [ ] Local discovery behavior is provably unchanged with and without a registry
- [ ] `just test` + `just lint` green

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
