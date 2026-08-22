# Demo kitchen fixture

The **second** planning identity, and the only reason the atlas has more than one card.
[`demo-planning/`](../demo-planning/) is a bike workshop; this is a vegetarian kitchen —
deliberately unrelated, because the atlas exists to answer "what am I working on across
*unrelated* projects", and two variations on one theme would not show that.

Only [`atlas.tape`](../vhs/atlas.tape) records against it, via
[`atlas-setup.sh`](../vhs/atlas-setup.sh), which stages a copy into a throwaway registry.

It's a self-contained planning root (`taskflow_root = "."` in
[`.tskflwctl.toml`](./.tskflwctl.toml)) carrying its own durable id, which is what lets
the registry group entry points by planning identity rather than by path.

## What's here, and why

| | Contents | Shows off |
| :-- | :-- | :-- |
| **Epics** | `01-dal` (42%), `02-banh-mi-completely-from-scratch` (37%), `03-potato-dominant-recipes` (42%) | three rollup bars mid-progress, so the card's counts and the dashboard both have something to render |
| **Tasks** | 24 across every status — 9 completed, 4 next-up, 4 ready, 3 in-progress, 2 deferred, 2 deprecated | the status glyphs, and an atlas card whose "3 in progress" is not zero |
| **Audits** | one **open** (`2026-08-15-pantry-staples`, 6 findings) | the open-audit count on the atlas card and the segmented bar on the dashboard |
| **Research** | 2 docs (pressure-cooker vs stovetop for legume texture; high-hydration dough with rice flour) | the research tab, which shipped after the bike fixture was authored |

The audit's six findings span **fixed · in-progress · open · wontfix · deferred**, so the
segmented bar shows several bands rather than one solid block.

## Why the contrast with `demo-planning` matters on camera

The bike tree's dates are a month old and this one's are fresh, so in the atlas GIF one
space reads "1mo ago" and the other reads "today". That is what a real machine looks
like: a project you touched last month beside one you are in the middle of. It is a
side effect of when each fixture was authored rather than a design, but it is worth
preserving if this tree is ever regenerated.

## Regenerating

Committed static data — edit the markdown in place, or re-run the `tskflwctl
epic/task/audit/research new` (+ lifecycle) commands that generated it. Change status
with the lifecycle verbs, never by hand-editing frontmatter. Keep it lint-clean
(`tskflwctl -C assets/demo-kitchen lint`). As with the bike fixture, dates are baked in,
so relative-date labels age until the GIFs are re-recorded.
