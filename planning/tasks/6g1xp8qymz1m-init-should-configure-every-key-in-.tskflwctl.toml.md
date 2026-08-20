---
schema: 1
id: 6g1xp8qymz1m
status: next-up
epic: 20-cli-ux-and-ergonomics
description: Theme and pager are hand-edit only; init reaches every other key. Add them, with a palette picker that repaints live as the selection moves — and decide repo vs home first.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, ux, config, theme]
created: "2026-08-20"
---
# `init` should configure every key in .tskflwctl.toml

## Objective

Nobody should have to hand-edit `.tskflwctl.toml` to reach a setting the tool itself
owns. `init` grew placement (`taskflow_root`), pointers, tracking and durable ids; what
remains is the presentation block — and the palette picker in particular should be
**delightful, repainting the theme live as the selection moves**.

## Coverage today

| key | reachable from `init`? |
| --- | --- |
| `taskflow_root` | ✅ `--taskflow-root` + the placement prompt |
| `id` | ✅ minted on create, backfilled on re-run |
| `planning_repo` | ✅ `--planning-repo` + the here-vs-elsewhere prompt |
| `planning_repo_id` | ✅ recorded on create, backfilled on re-run |
| `tracked_repos` | ✅ `--track` |
| `[theme].name` | ❌ **hand-edit only** |
| `[pager].enabled` | ❌ **hand-edit only** |
| `[pager].command` | ❌ **hand-edit only** |

Three keys. The scaffold template already ships them as commented-out examples, which is
the tell: we knew people would want them and made them read the file to get there.

## The palette selector

**Repaint as the cursor moves** — the point is seeing the theme, not reading its name.

Two things to know before starting:

- **`huh.Select` cannot do this.** It exposes no highlight-change hook (no `OnChange`,
  nothing equivalent), so `prompt.SelectOne` is the wrong tool. This needs a small
  Bubble Tea model of its own — closer to the TUI's `list.Model` usage than to the huh
  prompt layer.
- **The renderer already exists.** `theme preview` draws "color swatches + a sample bar"
  for a named theme, and `internal/design` holds the registry. The work is driving that
  renderer from a moving cursor, not inventing the preview.

Keep the non-TTY path intact: a flag (`--theme <name>`) must set it headlessly, and the
`prompt.Gate` contract means an agent or pipe never blocks.

## The question worth settling first

**Should these land in the repo config at all?**

`[theme]` and `[pager]` are documented in `internal/config` as *local-terminal concerns*,
and the home-config tier exists precisely so a preference about your own terminal is not
repeated per project. A theme written into a **committed** `.tskflwctl.toml` imposes one
contributor's taste on everyone in the repo — the exact problem the home tier solved.

So the honest shape may be: `init` asks, then writes to `~/.config/tskflwctl/config.toml`
by default, with the repo config as an explicit opt-in for a project that genuinely wants
to pin its look (screenshots, docs). Decide this before building the picker — it changes
what the picker writes, not just where a flag points.

## Notes

- Re-running `init` in an existing repo is a **repair**, and it reports the layout rather
  than re-asking settled questions. Any new prompt must respect that: offer to change a
  setting only where changing it is actually supported.
- `[pager].enabled` is a `*bool` on purpose — "unset" must stay distinct from an explicit
  `false`, or the tier merge breaks. A three-way prompt (default / on / off), not a
  toggle.
- Surgical TOML writes: preserve comments, key order, unknown keys. The scaffold's
  commented-out `# [theme]` block should become real, not be duplicated.

## Acceptance criteria

- [ ] Where `[theme]`/`[pager]` are written (repo vs home) is decided and recorded
- [ ] Every key in the table is reachable without hand-editing the file
- [ ] The palette picker repaints the theme as the selection moves
- [ ] Each is settable headlessly by flag; no prompt blocks a non-TTY run
- [ ] `[pager].enabled` can be returned to *unset*, not just true/false
- [ ] Writes are surgical — comments and key order survive
- [ ] `just test` + `just lint` green

## Out of scope

- Changing `taskflow_root` on an existing repo — that is a migration (move the tree,
  rewrite the key), not a settings edit.
- A general `config set` command. This task is about `init` being complete; a separate
  editing verb is a different question.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- The theme registry and preview renderer: epic
  [25-design-system-coherent-palette-and-selectable-themes](../epics/25-design-system-coherent-palette-and-selectable-themes.md)
- Why the home tier exists (and why a committed theme is questionable):
  [multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
