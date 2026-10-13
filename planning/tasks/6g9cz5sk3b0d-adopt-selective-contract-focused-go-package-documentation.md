---
schema: 1
id: 6g9cz5sk3b0d
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Centralize package contracts in doc.go where they add value, using a risk-based standard instead of mandatory mini-READMEs for every package.
effort: 1 day
tier: 3
priority: medium
autonomy_level: 4
tags: [documentation, go, architecture, godoc]
created: "2026-09-12"
depends_on: [6g9cz5saenme]
updated_at: "2026-10-12"
---
# Adopt selective, contract-focused Go package documentation

## Objective

Make `go doc` a dependable, code-adjacent source for package ownership and usage without reproducing the monolithic architecture document inside every directory.

The Desirelines `shared`, `dispatcher`, and `apigateway` modules demonstrate the useful shape: focused `doc.go` files use Go headings, linkable symbols, small usage examples, and explicit concurrency, lifecycle, error, or composition contracts. They also show why coverage should be selective: many simple packages need only a precise package comment, while a port boundary or stateful adapter benefits from an expanded contract.

Taskflow currently has package comments in arbitrary implementation files. This can make package documentation depend on file ordering and encourages multiple file headers to compete with the package overview.

## Acceptance criteria

- [ ] Define a risk-based package documentation standard with three intentional outcomes: a short package comment, an expanded `doc.go` contract, or generated-package provenance guidance.
- [ ] Expanded docs are selected for architectural boundaries and packages with non-obvious lifecycle, concurrency, persistence, compatibility, failure, or composition rules; there is no quota requiring `doc.go` in every package.
- [ ] For selected packages, `doc.go` becomes the single package overview and covers ownership, allowed dependencies, safe composition, important invariants, and links to the relevant symbols or scoped guides.
- [ ] Existing package comments are audited so `go doc` has one coherent synopsis rather than comments distributed across arbitrary first source files.
- [ ] Package documentation avoids volatile file inventories, current feature rosters, copied interface method lists, and decision history better owned by generated inventories, type docs, ADRs, or planning.
- [ ] Generated packages, if any, name their source artifact and regeneration command; ordinary packages use Go doc links such as `[Service]` instead of textual symbol copies.
- [ ] `go doc` output for the selected packages is reviewed for useful headings, working symbol links, and an agent-readable first screen.

## Out of scope

- Creating `doc.go` mechanically for every directory.
- Moving detailed subsystem manuals or accepted decision history into Go comments.
- Changing package boundaries merely to make the documentation diagram tidier.
- Treating uncompiled snippets as sufficient verification; executable examples are owned by the sibling example task.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- Reference reviewed: `../desirelines/packages/{shared,dispatcher,apigateway}`

## Architecture extraction handoff (2026-10-12)

Focused [architecture guides](../../docs/ARCHITECTURE.md) now own cross-package explanations and link
to owning source/regression evidence. Existing package comments in core, store, appwiring, config,
cli, wire and tui route back to the relevant guide; their paths are repository source-navigation
cues, not local Markdown links resolved by go doc. The navigation test derives the participating
source routes from guide links without a second package inventory. No doc.go standard or complete
package-comment audit was delivered by that extraction. Preserve those two-way links as this task
chooses selective package overviews, and do not duplicate the focused manuals inside comments.

## Review correction handoff (2026-10-12)

Userconfig is now an eighth guide-linked package overview: its comment points to the [preferences/registry owner](../../docs/architecture/configuration-and-spaces.md#preferences-and-home-registry) and correctly credits depguard, not the compiler, for the config-to-userconfig ban. Preserve that route during selective doc.go work. This small correction does not constitute the broader package-comment audit or documentation standard.
