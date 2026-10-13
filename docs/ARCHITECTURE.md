# tskflwctl architecture

Taskflow is a local-first planning CLI and TUI over Markdown/frontmatter. Both
primary adapters use the same application services; the filesystem is a secondary
adapter, not a controller dependency.

## The rule: CLI/TUI are primary adapters over a shared core; the filesystem is a secondary adapter

```text
CLI / TUI  ──▶  core services + pure domain  ──ports──▶  secondary adapters
                           ▲
              binary-selected composition
```

## Read for the change

| Concern | Canonical guide | Owning code |
| --- | --- | --- |
| Package direction, composition and exceptions | [Boundaries](architecture/boundaries.md) | [Invocation ports](../internal/cli/ports/runtime.go), [wiring](../internal/appwiring/wiring.go), [lint rules](../.golangci.yml) |
| Identity, snapshots, optional capabilities and graph projections | [Planning data](architecture/planning-data.md) | [Record evidence](../internal/core/entity_read.go), [core ports](../internal/core/store.go) |
| Guarded writes, YAML preservation, CAS and recovery | [Persistence](architecture/persistence.md) | [Filesystem adapter](../internal/store/fsstore.go), [mutation policy](../internal/core/mutation_policy.go) |
| Root discovery, preferences, registry and workspace opening | [Configuration and spaces](architecture/configuration-and-spaces.md) | [Discovery](../internal/config/config.go), [workspace service](../internal/core/workspace.go) |
| Commands, output and machine compatibility | [CLI and wire](architecture/cli-and-wire.md) | [Command tree](../internal/cli/root.go), [wire contract](../internal/wire/wire.go) |
| Shell navigation, async state, graph views and watching | [TUI](architecture/tui.md) | [Model](../internal/tui/model.go), [entity registry](../internal/tui/entity.go) |

Detailed API guarantees belong beside their types and methods. These guides explain
cross-package relationships and link to implementation and regression evidence;
they do not reproduce every method, filename, command, or wire revision.
[AGENTS.md](../AGENTS.md#where-truth-lives) owns contributor routing and the
source-of-truth matrix. Accepted decisions live in [ADRs](../planning/adrs/),
not in a feature-status roster here.

## Keeping guides and code connected

When changing a contract, update its owning comment and relevant guide in the same
change. A guide links to the owning Go source and the checks that exercise its claims;
the package comment points back using a repository-relative guide path. Use Go
`[Symbol]` links for API references rather than copied declarations. Repository paths
in Go comments are source-navigation cues, not promises that `go doc` resolves local
Markdown links.

Navigation checks catch broken routes, not semantic contradictions. Behavior tests
and lint enforce specific guarantees; review still judges prose, exceptions and
compatibility. The [architecture navigation test](../internal/tools/architecturecheck/architecture_test.go)
runs with `just test` and checks the index, focused guides, root contributor routes
and linked package-comment return paths, plus task/Thread links to retained index
anchors. It ignores fenced/inline-code examples and does not prove semantic prose
or implement every Markdown dialect. Selective package overviews, compiled examples and the executable
import-map check are [sequenced work](../planning/threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md),
not claimed complete by this extraction.

## Current package map and dependency direction

The [boundaries guide](architecture/boundaries.md#package-map) owns the role table
and descriptive import map. The map is still manual; its checker follows this split.

## Planning-data change checklist

Use the [planning-data checklist](architecture/planning-data.md#change-checklist)
for identity, snapshots, capability binding, authorization and recovery changes.

## Enforced fitness rules

See [executable boundary checks](architecture/boundaries.md#enforced-fitness-rules).
Import lint does not prove snapshot consistency, no-effects behavior or recovery.

## Explicit composition and named local exceptions

See [composition](architecture/boundaries.md#composition-and-local-exceptions);
a named exception is not permission for general adapter leakage.

## Non-negotiable patterns

The [boundaries](architecture/boundaries.md), [planning-data](architecture/planning-data.md),
[persistence](architecture/persistence.md) and [CLI](architecture/cli-and-wire.md)
guides own the relevant rules. Consult the guide for the change, not all six.

## Why these boundaries (and why not collapse them)

See [boundary rationale](architecture/boundaries.md#why-these-boundaries).
Consolidation needs demonstrated friction, not hypothetical future adapters.

## The TUI (`internal/tui`)

See the [TUI guide](architecture/tui.md), including async/session rules and layout checks.

## Testing

Run `just test`, `just lint` and `just build` for code changes. The focused guides link
to regression evidence. [CLI/wire validation](architecture/cli-and-wire.md#validation)
covers generated references, golden updates and process-exit tests.

## Status (2026-08-25)

This old anchor remains for incoming links, not as a current feature snapshot.
The original section, including later edits under its dated heading, is archived in
[the restructuring task](../planning/tasks/6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md#archived-architecture-snapshot).
Current work lives in [planning](../planning/); historical design and review context
is recorded there and in the [pre-split reference](https://github.com/andy-esch/taskflow/blob/4fdf6fcc/docs/ARCHITECTURE.md).
