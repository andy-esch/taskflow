# TUI architecture

[Architecture index](../ARCHITECTURE.md) · [Planning projections](planning-data.md#threads-and-graph-presentation)

[The Bubble Tea model](../../internal/tui/model.go) is the second primary adapter
over core services. Reads and writes run through `tea.Cmd`, returning messages;
`Update` and `View` do not perform planning I/O or open persistence.

## Shell and entity extension seams

[The entity registry](../../internal/tui/entity.go) owns list state, loaders, filters,
sort choices and declared lifecycle delegates. The shell owns keys, focus, overlays,
history and canonical cross-entity navigation. Add a browsable entity through this seam;
a read-only orientation screen uses [the dashboard](../../internal/tui/dashboard.go)
rather than pretending to be a filterable/mutable entity.

Rows carry adapter-established source IDs separately from display labels.
All families use those keys for selection restore, delayed-result guards, navigation,
edits/lifecycle, dashboard/atlas jumps and cached sessions. No path, filename,
frontmatter declaration or position is an identity fallback. Duplicate labels get
stable shortest-unique prefixes; invalid or duplicate canonical-key snapshots are
quarantined before filtering, not silently reduced to one row.
[Identity tests](../../internal/tui/identity_test.go) and
[cross-view tests](../../internal/tui/identity_cross_view_test.go) exercise this.

Task field set/unset and move/defer commands pass `TaskMutationTarget` to explicit
selected-identity Service entry points, never the reference-based CLI methods.
Submitted commands capture the originating service in `tea.Cmd`; workspace activation
closes unsubmitted forms/menus, which do not capture their opening service. Matching
IDs and list generations cannot carry an open form into a different checkout.
Generation/session envelopes protect result delivery, not cancellation of an
already-issued write. Validation, eligibility, timestamps, retry bounds and durable
receipts stay in core/persistence.
[Selected-action tests](../../internal/tui/selected_task_action_test.go) exercise
real command closures and forms before watcher refresh, impostor labels, rename,
collisions, denied writes and stale/current-space results.
[Activation tests](../../internal/tui/selected_task_session_test.go) exercise a
pending Atlas open landing under an edit, action, confirm or defer form.

Audit lifecycle and epic field/lifecycle commands also use selected-identity
Service entry points, retaining ordinary CLI validation and reference matching.
[Non-task action tests](../../internal/tui/selected_entity_action_test.go) cover
missing selected sources and impostor labels before the first Service call.
This is not yet a universal selected-action guarantee:
[local-path actions](../../planning/tasks/6gm30jat0xsy-keep-selected-tui-path-actions-bound-to-exact-identities.md)
still require an identity-only optional lookup. Rechecking a fuzzy ID string twice
does not prevent both lookups selecting another record's ID-shaped human label.

Lifecycle menus and command verbs use the registry's transition table and delegate.
Tasks, audits and epics have their respective core transitions; Threads remain
read-only here and research has no lifecycle. Task inline editing uses descriptor
fields and core validation, not a second policy path.

The generic [detail pane](../../internal/tui/detail.go) caches raw/pretty Markdown and
owns alternate-view and structured-navigation seams. Presentations supply row order
and selected identity; the shell handles open/history. Runtime
[help sections](../../internal/tui/help.go) own bindings, not a copied key inventory
in this guide. Structured follow and back-stack behavior lives in
[nav.go](../../internal/tui/nav.go).

## Threads and local graph focus

[Thread loaders](../../internal/tui/thread_projection.go) use shared core projections,
not a parallel root-model list/detail cache. Task-only edits invalidate loaded Threads
because readiness is derived. One graph/body read supplies detail semantics; optional
Thread paths join independently. Pathless adapters retain browsing while path copy,
clickable local titles and editor actions degrade explicitly.
[Local-action tests](../../internal/tui/local_path_test.go) prevent opaque diagnostic
locations becoming filesystem authority.

Graph ranks are dependency explanations, not lifecycle waves or execution barriers.
Spatial and one-hop focus views consume the same pure neighborhood selector as CLI
exports. Full/focused layout caches stay separate; selection/focus state survives
generic detail reloads. Layout, route geometry, fan-choice controls, clipping and scope
cards are presentation concerns; core supplies exact topology and boundary evidence.
Do not reparse Mermaid/DOT to invent semantics.

## Asynchronous state and refresh

[Commands/messages](../../internal/tui/commands.go) carry selection/list generations;
workspace sessions add an outer generation so old results cannot land in a new space.
The active-space watcher uses the supplied Layout, not reconstructed directory names.

[Transient-read retry](../../internal/tui/read_retry.go) treats planner-window conflict
as contention: retain the last coherent model, retry once after quiet, supersede
stale retries, then surface repeated conflict. First loads remain loading rather than
inventing emptiness. Mutation conflict is different: report and reload. Lifecycle
feedback independently presents task/Thread impacts and exact core remedies.
Committed cleanup failure still warns and reloads.

[Atlas](../../internal/tui/atlas.go) groups durable planning identities and exposes
diagnosed entry points. Repo launches open that repo; no-repo launches use atlas
landing with a seeded runtime workspace, not an implied selection. Successful opening
swaps services/config-start/Layout atomically; failed/stale opening preserves the
current workspace. [Session caches](../../internal/tui/session.go) are per checkout
and do not retain watcher descriptors. Atlas has home-scoped chrome and explicit
sorting independent of the launch repo's theme.

Partial overview refresh advances healthy groups while contended groups retain
visibly stale coherent summaries; only affected groups retry once. First-load/durable
failures do not invent retained data. Group retry and workspace/load generations are
separate from whole-overview dirtiness: a partial success cannot clear unrelated
staleness or open errors. Retained summary slices are independently owned.

## Watcher and terminal discipline

[Watcher implementation](../../internal/tui/watch.go) and
[tests](../../internal/tui/watch_test.go) cover canonicalized
deepest-existing paths, leaf attachments, nearest-parent sentinels and quiet-period
reconciliation. Missing/nested/replaced directories recover without polling.
Partial coverage is degraded; no useful coverage is off; manual refresh retries.
Symlink aliases deduplicate, but layout-controlled symlinks remain degraded because
retarget observation is not portable; manual refresh re-resolves and reattaches.
Workspace switch closes old leaves/sentinels and watches only the new active space.

Subtract border frames before sizing children, guard rendering before window size,
truncate rather than wrap joined elements, and clamp output to terminal dimensions.
The view-fit regression in [model tests](../../internal/tui/model_test.go) pins
height/width limits. [Design research](../../planning/research/6faxn1800y6n-tui-design-decisions.md)
retains the broader checklist.

TUI tests inject messages and assert state/rendering, with selected full-program
tests and real watcher coverage. Pure reducer tests avoid relying on backend timing.
[Navigation redesign](../../planning/tasks/6g8vxcnezktm-reassess-tui-navigation-and-information-architecture-at-current-scale.md)
remains future design, not a promise that today's bindings/layout are permanent.

## Presentation ownership

[Theme](../../internal/theme/theme.go) owns semantic glyph/color roles;
[design](../../internal/design/palette.go) chooses palette/chrome hues.
Theme never imports design. CLI and TUI map the same semantics to their different
renderers. Model-scoped styles avoid mutable global theme state; configui and CLI
palette preview use the shared ordered token projection, not copied token rosters.
