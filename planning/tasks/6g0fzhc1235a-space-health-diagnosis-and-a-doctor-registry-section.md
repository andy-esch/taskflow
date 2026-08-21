---
schema: 1
id: 6g0fzhc1235a
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Share typed ok/empty/missing/not-a-repo/unreadable/mismatch diagnoses across space list and doctor, with remedies. Broken entries stay visible and are never auto-forgotten.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, multi-repo]
created: "2026-08-15"
updated_at: "2026-08-20"
started_at: "2026-08-20"
completed_at: "2026-08-20"
---
# Space health diagnosis and a `doctor` registry section

## Objective

Registered paths rot — repos move, get deleted, get un-inited. Make that a **kind
diagnosis with a remedy**, never a bad error, and put the logic in one place both the
CLI and (later) the TUI read, so the two can never disagree about what is wrong.

This is the "delightful missing config" requirement of epic 29, and it is a
generalization of machinery that already exists: epic 23's `LinkProblem` →
`CheckLinks` → `doctor` is the same pattern one scope down.

## Notes

Widen `LinkProblem` into a `SpaceProblem{ID, Path, Kind, Message, Remedy}`, resolved
lazily per entry:

| state | detection | message | remedy |
| --- | --- | --- | --- |
| `ok` | discovery succeeds | — | — |
| `missing` | path gone | "not found at ~/…" | forget · repoint |
| `not-a-repo` | exists, no marker | "no `.tskflwctl.toml` — init here?" | init · forget |
| `unreadable` | TOML / permission error | verbatim parse error | `$EDITOR` · forget |
| `empty` | valid, zero entities | "no tasks yet" | — |
| `mismatch` | path resolves, durable id differs from `verify_id` | expected/found ids | restore · forget/re-add |

Three principles separate "delightful" from "nagging":

- **A broken entry is never fatal** — it is reported alongside the healthy ones, never
  an error screen and never a blocked command. Same philosophy `lint` already applies
  one scope down: a task with an unrecognized status is *listed and flagged*, never
  moved or dropped.
- **Never auto-forget.** Removal is always explicit.
- **One diagnosis function** behind every surface.

`moved` detection was deliberately rejected for this slice: guessing where a repo went is
the kind of cleverness that misfires. A truthful `missing` plus an explicit forget/re-add
remedy is safer; a path that resolves to the wrong repo is detected mechanically through
the recorded durable `verify_id` and reported as `mismatch`.

## Acceptance criteria

- [x] One exported diagnosis function returns typed `SpaceProblem`s for a registry
- [x] `doctor` gains a registry section: human + `--json` + the existing nonzero-exit
      CI contract, without duplicating the linkback logic
- [x] Every state in the table has a test with a real temp-dir fixture
- [x] A broken entry never blocks `space list` or any other command
- [x] Nothing auto-removes an entry
- [x] `just test` + `just lint` green

## Out of scope

- TUI rendering of the states (later, if the board proceeds)
- Automatic repair of any kind

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
- The pattern this generalizes: epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)

### 2026-08-20 — implementation closeout

A new internal/spacehealth projection now owns one typed diagnosis per registered checkout: ok, empty, missing, not-a-repo, unreadable, and mismatch. Empty is healthy; missing and configuration failures remain visible data with explicit remedies; a recorded verify_id mismatch mechanically detects a wrong repo at a still-resolving path. The proposed nearby moved-repo scan was deliberately omitted because guessing is less safe than an honest missing diagnosis and explicit forget/re-add flow.

space list and doctor consume the same projection. List remains successful for every entry and never mutates the registry. Doctor retains its existing linkback problems and adds a human/JSON registry section; only actionable space problems join the existing validation exit contract. The JSON schema advances to 1.38, with empty/mismatch/remedy on space entries and typed doctor registry problems. Real temp-dir fixtures cover every supported state, doctor behavior, ordinary-command isolation, and no-auto-forget behavior.

Documentation was corrected as part of the feature: README now includes a practical multi-repo workflow and current rollout boundary; ARCHITECTURE names userconfig registry storage and the shared health layer; generated CLI docs and machine-contract goldens are refreshed; and proposed ADR-0005 now reflects the shipped two-file layout, local-label plus durable verify-id model, explicit registration, and the planned status of global --space/init auto-registration. Full race-enabled tests, golangci-lint, module tidiness, planning lint, and diff hygiene are green.
