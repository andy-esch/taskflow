# Demo planning fixture

A small, **curated** planning tree the demo GIFs ([`../`](../)) record against —
authored to show `tskflwctl`'s symbology off in one screen, rather than recording
this repo's own [`planning/`](../../planning/) (whose audits are all closed, so the
open-audit section and the segmented bar wouldn't appear).

It's a self-contained planning root (`taskflow_root = "."` in
[`.tskflwctl.toml`](./.tskflwctl.toml)); the tapes `cd` into it.

## What's here, and why

| | Contents | Shows off |
| :-- | :-- | :-- |
| **Epics** | `01-api-gateway` (75%), `02-observability` (50%), `03-data-pipeline` (25%), `04-search-indexing` (0%) | the rollup bars at a spread of completion |
| **Tasks** | 14 across every active + archived status (in-progress, next-up, ready, completed, deferred) | the status glyphs and the dashboard's count line |
| **Audits** | one **open** (`2026-06-20-api-gateway`, 8 findings) + one **closed** (`2026-06-10-data-pipeline`) | the bucket glyphs and the Open-audits dashboard section |

The open audit's eight findings deliberately span **fixed · landed ·
in-progress · open · deferred · wontfix**, so the **segmented finding bar** shows
all of its bands (`█` done · `▓` in-progress · `▒` dropped · `░` open) and the
`audit show` **finding tree** shows every status group.

## Regenerating

It's committed static data — edit the markdown in place, or re-run the
`tskflwctl epic/task/audit new` (+ `complete`/`defer`/`close`) commands that
generated it. Keep it [lint-clean](../../) (`tskflwctl -C assets/demo-planning
lint`). Dates are baked in, so relative-date labels ("today") age until the GIFs
are re-recorded.
