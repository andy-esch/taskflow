---
status: completed
epic: 17-pm-go-cli
description: 'Robustness + docs pass after dropping pm: conflict exit code, exclusive create, audit-count hardening, fakeStore cleanup, CLAUDE.md'
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, robustness, docs]
created: "2026-06-09"
updated_at: "2026-06-09"
completed_at: "2026-06-09"
id: 6fakbec00fke
---

# Harden create-loop and fill docs (post-pm-retirement review)

## Objective

A review + hardening pass once the daily loop stopped needing Python `pm`:
tighten the new create path, fix latent anti-patterns the earlier reviews
flagged, and fill the doc gaps (notably: the repo had no `CLAUDE.md`).

## What changed

**Robustness**
- **`ErrConflict` → exit 14** (`domain/errors.go`, `cli/exit.go`): "already
  exists" is now a distinct conflict, not a generic validation error — scripts
  can tell a name collision (14) from bad input (11).
- **Exclusive atomic create** (`store/atomic.go`): factored a shared `stageTemp`
  and added `createFileAtomic` (temp + hard-link, fails with `os.IsExist`).
  `CreateTask`/`CreateEpic` use it, removing the stat-then-write **TOCTOU**;
  collisions map to `ErrConflict`.
- **Audit finding counts hardened** (`store/auditstore.go`): strip ```-fenced
  blocks before counting (example syntax in docs no longer inflates `Findings`),
  and the open-status regex no longer matches `open-ish`/`openness`.

**Anti-patterns**
- **`fakeStore` god-stub** (`core/service_epic_test.go`): replaced 13 hand-stubs
  with an embedded `nopStore` (+ `var _ Store = nopStore{}`), so a new `Store`
  method now updates one place, not every fake.
- De-duplicated the two atomic-write functions via `stageTemp`.

**Docs**
- New **`CLAUDE.md`**: build/test/lint, the adapter architecture, the
  `tskflwctl`-not-`pm` planning workflow, git rules, exit-code conventions.
- `README.md`: a "Daily workflow" command reference + a "`pm` is retired" note.
- `docs/ARCHITECTURE.md` + epic 17: exit codes 10–14, atomic-create path.

## Acceptance

- [x] `go build/test ./...` + `golangci-lint` green; new tests for conflict
      exit 14 and audit fence/open-ish counting.
- [x] `CLAUDE.md` exists and routes agents to `tskflwctl` for planning.
- [x] This task itself was created with `tskflwctl task new` (pm unused).

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); follows
  [add-create-verbs-task-new-and-epic-new](6fa91vg03g4z-add-create-verbs-task-new-and-epic-new.md) and the autocomplete work.
