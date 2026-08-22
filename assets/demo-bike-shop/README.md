# Demo pointer fixture

A config file and nothing else — deliberately.

This is an **impl repo** whose [`.tskflwctl.toml`](./.tskflwctl.toml) points at
[`demo-planning/`](../demo-planning/) rather than holding a planning tree of its own.
It exists so the atlas has a card reachable **two ways**: the planning checkout itself,
and a repo that routes to it. That is the one atlas behavior nothing else can
demonstrate — a card is a planning *identity*, not a directory, so two registered paths
sharing one durable id collapse into a single card with two entry points, which is what
`h`/`l` selects between.

Both configs carry the same planning id (`demo-planning` holds it as `id`, this one
verifies it as `planning_repo_id`), and `demo-planning` lists this repo in
`tracked_repos`. That linkback is what keeps `tskflwctl doctor` reporting
`Linkbacks ✔ consistent` in both directions; if you move or rename either directory,
re-run `init`/`config migrate` rather than hand-editing, or the demo will record a
warning banner.

Used only by [`atlas.tape`](../vhs/atlas.tape), through
[`atlas-setup.sh`](../vhs/atlas-setup.sh), which stages a copy and rewrites the relative
`planning_repo` to match the renamed sibling.
