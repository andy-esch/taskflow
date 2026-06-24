---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: 'Cleanup from the v2 review: derive the description-cap strings from MaxDescriptionLen (kill the 9-place drift), decide schema --json-schema paging, and dedup the TUI/CLI bar constructors.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tech-debt, cli]
created: "2026-06-24"
updated_at: "2026-06-24"
started_at: "2026-06-24"
completed_at: "2026-06-24"
---
## Objective

Small, non-blocking cleanups surfaced by the post-v2 adversarial review (2026-06-23/24).

## Acceptance criteria

- [ ] **Description-cap string drift** — derive the description-cap help/guidance from `domain.MaxDescriptionLen` instead of hardcoding the number. Bumping 150→200 was a 9-place sweep (entity.go ×4, task.go/epic.go flag help ×4, dto.go jsonschema tag) plus goldens + docs. Flagged in the 2026-06-13 audit ("two enforcement paths for the same documented cap"). Fix: `fmt.Sprintf(..., MaxDescriptionLen)` for the CLI flag help; a small builder for the static schema literals in entity.go/dto.go.
- [ ] **`schema --json-schema` paging** — decide whether to exclude `--json-schema` (machine-ish output) from the pager. On a TTY with `--paginate` it can silently drop output if the pager doesn't read it; git-consistent and low-risk, just make it a deliberate call.
- [ ] **Bar constructor dedup** — `tui.miniBar` and `render.Style.Bar` hand-keep identical `progress.Model` constructions (the `theme.BarFill` anti-drift seam was removed in the v2 migration). Add a shared constructor. **Coordinate with `feat/gradient-lipgloss`**, which already rewrites both bars (and deletes `barColor`) — cleanest to dedup as that lands.

## Out of scope

- The MAJOR OSC-11 fix + README corrections (done in the review-fix change).

## Related

- Audit precedent: `planning/audits/closed/2026-06-13-codebase-quality-architecture.md` (the cap dual-source finding).
- Epic [[21-code-quality-architecture-hardening]].
