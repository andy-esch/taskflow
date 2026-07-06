---
schema: 1
status: active
description: Point an impl repo at an external planning repo (planning_repo) + reverse tracked_repos, init pointer mode, and linkback integrity (warnings + doctor)
priority: medium
tags: [config, cli, multi-repo]
created: "2026-06-22"
---
# Decoupled planning — point an impl repo at an external planning repo

Today `tskflwctl` assumes the planning tree lives in the repo you run it from
(or a `planning/` subdir). The `taskflow_root` key is deliberately **locked
inside the repo's own tree** — `configuredRoot` rejects any value that escapes
via `..` or an absolute path, as a "don't fork the data" guardrail. So an
*implementation* repo (e.g. `desirelines`) has no sanctioned way to say "my
planning lives in the sibling `desirelines-planning` repo," and `tskflwctl init`
always scaffolds a task tree you don't want there.

This epic adds **decoupled planning**: an impl repo points at an external
planning repo, the planning repo records the impl repos it tracks, and both
sides stay honest.

## Design (decided 2026-06-22)

- **`planning_repo` key** (new): in an impl repo's `.tskflwctl.toml`,
  `planning_repo = "../desirelines-planning"` — *allowed* to escape the tree
  (relative or absolute; remote backends are phase 2 — see **Remote backends**
  below). `taskflow_root` keeps its strict
  in-tree meaning and typo guardrail; `planning_repo` wins when both are set.
- **Validate-or-error**: `init` and discovery resolve the target and **error**
  if it isn't a real planning root (no `tasks/`).
- **Both directions**: the planning repo's `tracked_repos = [...]` records the
  impl repos that point at it. `init --planning-repo X` auto-appends this repo
  to X's `tracked_repos` (opt out with `--no-link-back`).
- **Linkback integrity**: ambient `⚠` warnings on normal runs when the two
  sides disagree — compared on **physical** paths (`Abs` + `EvalSymlinks`) to
  avoid false positives — plus `tskflwctl doctor` for an explicit bidirectional
  audit with a nonzero exit for CI.

## Config, both sides

```toml
# desirelines/.tskflwctl.toml          (impl repo — pointer, NO tree scaffolded)
planning_repo = "../desirelines-planning"

# desirelines-planning/.tskflwctl.toml (planning repo — owns the tree)
taskflow_root = "."
tracked_repos = ["../desirelines"]
```

## Sequencing

1. **Config schema + real TOML parser** (foundation) — `--next`.
2. **Discovery honors + validates `planning_repo`** — `--next`, after 1.
3. **`init` pointer mode + interactive flow** — after 1–2.
4. **`tracked_repos` seeding + auto-link-back** — after 3.
5. **Linkback integrity — ambient warnings + `doctor`** — after 4.
6. **JSON envelopes + docs/schema regen** — closes the docs-check gate.

Steps 1–2 are independently shippable and unblock the rest.

## Notes / risks

- **New dependency**: reading `tracked_repos` (a string array) plus a second
  scalar key pushes past the deliberately hand-rolled one-key TOML scanner
  (which refuses escapes rather than guess). Adopt a real TOML parser
  (`BurntSushi/toml` or `pelletier/go-toml/v2`; neither is in `go.mod` yet).
- **The correctness trap is normalization**: the linkback detector must compare
  physical paths, or it cries wolf whenever one side stores `../planning` and
  the other an absolute path or a symlinked checkout. A noisy false positive on
  every command is worse than no warning.
- **docs-check gate**: every new flag/command requires
  `go run ./internal/tools/docgen -out docs/cli` + committing the result.

## Remote backends (phase 2)

The local phase (sibling-dir `planning_repo`, steps 1–6 above) is **shipped**.
The "remote later" half — pointing `planning_repo` at a git remote, an object
store, or a service — is scoped in the research doc:

- `planning/research/2026-06-24-remote-planning-repos-backends-and-sync.md` —
  backend & sync options, grounded in the current storage seam.
- `planning/research/2026-06-24-task-storage-model-files-logs-or-versioned-db.md`
  — the on-disk data model (status-as-directory vs frontmatter / event-log / DB).
  Shares this work's root cause: state encoded in the path. Decide it *before*
  locking the OCC shape.

Headline: `core.Store` is a clean port (a backend swap leaves the Service +
domain untouched), but **discovery** (root is a physical path everywhere),
**write atomicity** (POSIX rename / `O_EXCL`), and **concurrency** are FS-coupled
— and today's concurrency is *path*-CAS only, with **no content/hash check**
(a concurrent same-file edit is last-writer-wins). The doc recommends a
backend-agnostic **version-aware `Store` port** as the foundation (optimistic
concurrency: read-version + `ifVersion`-on-write → `ErrConflict`), then
**git-sync over a local cache** as the first real remote — with the git-touching
kept *opt-in* so the "tool writes files, you commit" default survives. Object
stores and a `serve` daemon are later forks. Open questions for the maintainer
(chiefly: is "the tool never touches git" a hard line?) are listed at the end of
the doc.

The storage / read-model / OCC **foundation** this phase depends on is now its own
(design-first) epic —
[24-data-model-evolution-stable-key-storage-read-model-content-occ](24-data-model-evolution-stable-key-storage-read-model-content-occ.md); remote
backends ride on top of it once that foundation is decided and built.
