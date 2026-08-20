---
schema: 1
id: 6g1sbcbfb4mk
status: next-up
epic: 21-code-quality-architecture-hardening
description: anchorDir stats per resolution and CheckLinks iterates. Measured at well below noise (doctor 8.1ms/run); file the observation, act only on a profile.
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [perf, config]
created: "2026-08-19"
---
# Cache worktree anchoring if it ever shows up in a profile

## Objective

`anchorDir` performs an `Lstat` and, for a `.git` file, a `ReadFile` on **every**
`resolveRepoPath` / `resolvePlanningRepo` call. `CheckLinks` iterates `tracked_repos`, so
the same directory is re-stat'd several times per command
(audit 2026-08-19-worktree-aware-resolution, L1).

## Measured, not assumed

Timed after the change, 10 runs each, in this repo:

```
task list    45.3 ms/run
doctor        8.1 ms/run
```

`doctor` is the linkback-heavy path and is the cheapest command measured. The extra work is
a handful of syscalls against warm page cache — far below noise, and nowhere near the
dominant cost (`task list` is dominated by reading and parsing ~200 task files).

## Urgency

**None on its own.** File it so the observation is not lost, and do it only if a profile
implicates it. Premature caching here would add invalidation questions — worktrees are
created and removed mid-session — for no measured gain.

Two things that would change the call:

- A **cross-space `status --all`** (epic 29) that resolves N spaces per invocation, turning
  a per-command cost into a per-space-per-command cost.
- Any future path that calls `CheckLinks` in a loop.

The honest pairing is with the other per-invocation cost already noted (the user-config
read added in 1.31): if startup latency is ever profiled, measure both at once rather than
guessing at either.

## Acceptance criteria

- [ ] Only acted on if a profile implicates it; otherwise closed as wontfix with the numbers
- [ ] If cached: invalidation covers worktrees created or removed during a session

## Related

- Audit [2026-08-19-worktree-aware-resolution](../audits/6g1s9jpzr1yk-2026-08-19-worktree-aware-resolution.md) L1
- Prior startup-latency work: [osc11-startup-latency-spike](../research/6ff3hpm02x9p-osc11-startup-latency-spike.md)
