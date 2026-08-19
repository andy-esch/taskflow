---
schema: 1
id: 6g0fzhbz239k
status: next-up
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'After a successful scaffold or pointer init, append a [[space]] entry. Best-effort like LinkBack: warns to stderr, never fails the init. Opt out via --no-register/TSKFLW_NO_REGISTER.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, config, multi-repo]
created: "2026-08-15"
---
# `init` auto-registers the new space (best-effort, `--no-register`)

## Objective

Close the loop so the registry populates itself: after a successful scaffold **or**
pointer `init`, append a `[[space]]` entry, reported as one more `+` line in the
existing output. Without this, every new repo needs a second command nobody will
remember to run.

## Notes

- **Best-effort, exactly like `LinkBack`**: the init has already succeeded by the
  time this runs, so an unwritable or missing `$HOME` warns to **stderr** after the
  success line and never fails the command. The warning must not corrupt `--json`
  stdout (same ordering discipline `runInitPointer` already uses).
- Opt out with `--no-register` and `TSKFLW_NO_REGISTER=1`, mirroring
  `--no-link-back`. The env form matters for CI.
- Both modes register — scaffold and pointer. Because the entry stores the *repo*
  dir, a pointer repo registers as itself and resolves through `planning_repo`
  naturally.
- The `--json` init envelope grows a field for what was registered (the envelope is
  golden-tested; regenerate with `go test ./internal/cli -update`).
- Deliberately **not** in scope: auto-registering any repo you merely run a command
  in. A read-only command writing to `$HOME` is surprising, and throwaway clones and
  worktrees would accumulate.

## Acceptance criteria

- [ ] Scaffold and pointer `init` both append a `[[space]]` and report it
- [ ] An unwritable/missing `$HOME` warns to stderr; the init still succeeds (exit 0)
- [ ] `--no-register` / `TSKFLW_NO_REGISTER=1` suppress it
- [ ] Re-running `init` does not duplicate the entry (physical-path dedup)
- [ ] `--json` envelope carries the registration; goldens regenerated
- [ ] No test writes to a real `$HOME`
- [ ] `just test` + `just lint` green

## Out of scope

- Auto-registration on ordinary command runs
- The board's "unregistered current repo" prompt (TUI, later)

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
