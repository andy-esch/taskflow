---
status: completed
epic: 20-cli-ux-and-ergonomics
description: huh TTY-only pickers for missing inputs (epic/tags, bare transition verbs, ambiguity); human-only - agents always get exit codes
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, interactive]
created: "2026-06-12"
updated_at: "2026-06-20"
started_at: "2026-06-19"
completed_at: "2026-06-20"
---
# Interactive prompt layer (gh-style pickers)

## Decided (2026-06-17)

De-drafted. Resolution: **build it (option a), strictly human-only.** Prompts
appear ONLY when stdin+stderr are TTYs, `--json` is off, and `--no-input` (flag
/ `TSKFLW_NO_INPUT=1`) is unset; agents and pipelines always get today's exit
11/13, and every prompt has a flag twin (a guardrail test enforces it). The
required-input rule is unchanged for agents — humans just get a recovery face.
Ambiguity-recovery (item 4) still sequences after the fuzzy-resolution task.

## Architecture (decided 2026-06-19)

Grounded in the hexagonal setup (adapters → core → store, core pure): the prompt
layer is **CLI-adapter-internal**. The core never learns input might come from a
human — it receives complete params either way. TTY detection picks the *face* of
a capability, never the capability.

1. **`Prompter` port, not direct huh calls.** A small interface defined where it's
   consumed (`internal/cli/prompt`), with two implementations: a huh-backed one
   (the ONLY file importing huh) and a test fake. Buys: huh isolated to one file
   (swap/upgrade in one place); prompt *flow* testable in CI **without a TTY** (the
   fake); new pickers = new option lists over the same methods.
   ```go
   type Option struct{ Label, Value string }
   type Prompter interface {
       SelectOne(title string, opts []Option) (string, error)
       Text(title, placeholder string) (string, error)
   }
   ```
   (No `SelectMany` was needed: tags are free-form, so they ship as a
   comma-separated `Text` prompt — `fillTags` parses + dedups — not a multiselect
   over a fixed vocabulary. A richer "multiselect of suggestions" stays a future
   enhancement.)
2. **One `Gate`, resolved once, injected** — not scattered `isTerminal` checks.
   `on = stdin TTY && stderr TTY && !--json && !--no-input` (`TSKFLW_NO_INPUT=1`),
   resolved in App setup next to `setStyle`, stored on the App (DI, no globals).
3. **Flag-twin contract made structural** via a fill-helper tri-state, so a
   prompt-only path is impossible to write:
   `flag set → use it · else gate.On() → prompt · else → today's required error
   (exit 11/13)`. A guardrail test runs every prompting command non-interactively
   and asserts the exit code.
4. **Thin RunE / centralized abort.** Commands fill missing params via the helpers,
   then call core with complete params. huh's `ErrUserAborted` → clean exit in the
   wrapper.

**Layout:** `internal/cli/prompt/{prompt.go (Option, Prompter, Gate, ErrAborted),
tty.go (ttyPrompter — the real-TTY impl), picker.go (bubbles/list picker),
fake.go (test prompter)}`. App gains `Gate` + `Prompter` fields. Prompts render to
**stderr**; stdout stays byte-stable.

**Picker backend (decided 2026-06-19):** list pickers use **`bubbles/list`** — the
SAME component the TUI browses with — not `huh.Select`. huh.Select's built-in
filter has a viewport bug (after filtering it scrolls to the clamped selection and
hides earlier matches); bubbles/list's fuzzy filter (on the item's `FilterValue`,
matching the TUI) is battle-tested and correct, and reusing it unifies CLI/TUI
list-picking. huh is kept only for **text** prompts (no list, no quirk). This was a
swap *behind the unchanged `Prompter` port* — the interface and all contract tests
were untouched, which is exactly the payoff of the port. The picker model's key
handling (enter/esc/ctrl-c, filter-aware) is unit-tested by driving `Update`
directly (no TTY needed).

**Build order:** gate + `task new --epic` picker first (flagship, establishes the
pattern + contract test); then bare-transition pickers; then `--description`/title
text; the **tags picker**; **item 4 (ambiguity) stays deferred** (blocked on
[[fuzzypartial-slug-resolution]]). _Done so far (2026-06-19): foundation + epic
picker + all six bare-transition pickers, contract-tested._

**Tags (D1 resolved 2026-06-19 — prompt on a TTY):** tags are **free-form**, not a
controlled vocabulary — the model has no fixed tag list. So the picker leads with
**free text entry** (comma-separated); the tags already in use are offered as
*suggestions* (a multiselect over discovered tags), never a closed set. Agents/
pipes/`--json` still get exit 11 when `--tags` is missing. _Future (noted, not
scoped): a tag registry / normalization / `tag` management surface could turn the
suggestions into something richer — file a followup if wanted._

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

## Acceptance criteria

- [x] Planning conflicts resolved; task de-drafted (D1 = prompt for tags; epic
      moved to 20; ambiguity/item-4 deferred behind fuzzy-resolution).
- [x] No prompt is reachable when piped, under `--json`, or with `--no-input`
      (contract tests: `task new` missing epic/tags, `--start` missing
      description, bare transition → exit 11; the `gateOpen` truth table).
- [x] Prompts write to stderr; stdout byte-stable on the flag-driven path.
- [x] Each prompt has a flag equivalent (the `fillX` tri-state makes a
      prompt-only path impossible to write).

## Shipped (2026-06-20)

`task new` interactive flow: **epic** picker → **tags** (free-form comma text) →
**description** (only for `--start`/`--next`). **Bare transition verbs** → task
picker. Pickers run on **bubbles/list** (substring filter, alt-screen) behind the
`Prompter` port; text on huh. Gate + `--no-input` + the `fillSelect`/`fillText`/
`fillTags` helpers enforce the contract; hardened via a two-agent adversarial
review (fixed a Unicode-filter panic + a SIGINT→130 mapping). **Deferred:** item 4
(ambiguity, blocked on [[fuzzypartial-slug-resolution]]); a richer
multiselect-of-suggestions for tags; the TUI filter toggle
([[toggle-tui-list-filter-between-fuzzy-and-substring]]).

## Related

- Epic [[17-pm-go-cli]] · builds on [[fuzzypartial-slug-resolution]] ·
  see [[2026-06-12-pending-decisions]] (D1) and the 2026-06-12 design
  discussion in session notes.