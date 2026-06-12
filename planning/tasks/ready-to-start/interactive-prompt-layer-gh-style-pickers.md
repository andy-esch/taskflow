---
status: ready-to-start
epic: 17-pm-go-cli
description: 'DRAFT: huh-based pickers as TTY-only fallback for missing inputs (task new epic/tags, bare transition verbs, ambiguity recovery); needs plan review'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, interactive, draft]
created: "2026-06-12"
---
# Interactive prompt layer (gh-style pickers)

> 🚧 **DRAFT — not yet integrated into the overall plan.** Filed from the
> 2026-06-12 CLI-design discussion. Needs a planning pass to resolve the
> conflicts called out below before it's actionable; defer to the existing
> tasks named there wherever they disagree with this sketch.

## Objective

Give humans the `gh pr create` experience without ever compromising the
agent/pipeline contract. Governing rule: **interactivity is a recovery path
for missing required input, never a requirement** — TTY detection picks the
*face* of a capability, never the capability.

1. **Library + guardrails.** Use `huh` (Charm stack, matches bubbletea/
   lipgloss). Prompt ONLY when stdin AND stderr are TTYs, `--json` is off,
   and `--no-input` (flag and/or `TSKFLW_NO_INPUT=1`) is not set. Prompts
   render to **stderr** so stdout stays clean. Non-TTY behavior is exactly
   today's exit-11/13 errors — no agent-reachable path may ever block.
2. **`task new` fallback prompts:** missing `--epic` → select from epics;
   missing `--tags` → multiselect over tags in use + free entry (softens the
   D1 required-tags rule for humans; agents still get exit 11); missing
   title/description → text input.
3. **Bare transition verbs:** `task start` with no args → picker over
   ready-to-start; `task complete` → over in-progress (etc.).
4. **Ambiguity recovery:** when slug resolution yields multiple candidates —
   on a TTY, picker among them; piped/`--json`, today's `ErrAmbiguous`
   (exit 13) listing candidates.
5. Architecture: a `prompt` sub-package inside the `cli` adapter that fills
   missing params before calling `core`; the core never knows.

## ⚠️ Conflicts to resolve before starting

- **[[fuzzypartial-slug-resolution]] owns resolution semantics** (exact >
  unique prefix > unique substring; explicit ErrAmbiguous with candidates).
  Item 4 here is a *presentation layer on its output* — sequence that task
  first, and if anything here implies different resolution behavior, that
  task wins.
- **D1 interplay:** the tags picker changes the *felt* contract of
  [[align-task-new-scaffold-with-lint]] (completed) for humans. Confirm the
  team wants prompt-instead-of-error before building.
- Epic fit: filed under 17 by default, but 17 is the port epic and nearly
  done — planning should decide whether interactive UX warrants its own epic.

## Acceptance criteria (draft)

- [ ] Planning conflicts above resolved; task de-drafted.
- [ ] No prompt is reachable when piped, under --json, or with --no-input
      (test: every prompting command, non-TTY → current error codes).
- [ ] Prompts write to stderr; stdout byte-identical to the flag-driven run.
- [ ] Each prompt has a flag equivalent (no prompt-only capability).

## Related

- Epic [[17-pm-go-cli]] · builds on [[fuzzypartial-slug-resolution]] ·
  see [[2026-06-12-pending-decisions]] (D1) and the 2026-06-12 design
  discussion in session notes.