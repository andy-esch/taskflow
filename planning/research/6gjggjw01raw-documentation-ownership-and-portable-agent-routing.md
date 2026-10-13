---
schema: 1
id: 6gjggjw01raw
created: "2026-10-11"
description: Primary-source documentation practices and shared AGENTS.md ownership using native Claude and Antigravity discovery.
tags: [documentation, agents, maintainability]
updated_at: "2026-10-11"
---
# Documentation ownership and portable agent routing

## Question

How can taskflow reduce contradictory guidance and make the right contracts easier to find,
without introducing another documentation system or assuming identical coding-agent loaders?

## Findings

### Current repository (2026-10-11, baseline 09974fb)

- `CLAUDE.md` is the sole root agent guide: 136 lines / 1,204 words. It mixes durable safety
  rules with a command catalog, lifecycle vocabularies, storage layout, and prototype history.
  No root `AGENTS.md`, `GEMINI.md`, or nested agent guides exist. The user has selected
  `AGENTS.md` as the shared owner. After the loader research, they selected native Claude
  and Antigravity discovery, removing `CLAUDE.md` and declining a `GEMINI.md` wrapper.
- [README](../../README.md) has two development sections; its later `just test` description
  omits the race detector present in [Justfile](../../Justfile). The large manual command roster
  competes with [generated CLI reference](../../docs/cli/tskflwctl.md). Generation derives command
  metadata, not tested runtime semantics; both are needed when a promise is in dispute.
- [Architecture](../../docs/ARCHITECTURE.md) is now 1,087 lines / 11,466 words, with a
  dated status section and review history among current contracts. Its planning-data checklist
  already routes risky changes to owning ports and regressions; preserve that route.
- There are no `doc.go` files yet, but meaningful package/type contracts already exist, including
  [core ports](../../internal/core/store.go), [source sets](../../internal/core/source_set.go),
  [invocation bindings](../../internal/cli/ports/runtime.go), and
  [wire contracts](../../internal/wire/wire.go). Do not describe package contracts as absent.
- Executable ownership is already substantial: justfile, `.golangci.yml`, core/adapter tests,
  wire/schema tests, generated-doc checks, toolchain policy checks, and planning lint. There is
  not yet a general prose/link/agent-wrapper drift gate; don't claim one.
- Route dogfooding also found `schema task` describing guarded dependency commands as future work
  in [entity metadata](../../internal/domain/entity.go). Nearby comments referenced removed status
  directories and overstated automatic prose checking. Correct these at their owning sources in
  the ownership slice; the published JSON convention requires a classified revision/fixture update
  under ADR-0008, even though the command behavior is unchanged.

### Primary-source practices and loader differences

[Google's documentation guidance](https://google.github.io/styleguide/docguide/best_practices.html)
favors a small maintained set, deleting known-dead content, updating documentation alongside code,
and linking instead of duplicating. This supports removing the agent guide's mutable command
inventory, not discarding useful product examples or accepted decision history.

[Diátaxis](https://diataxis.fr/) separates tutorial, how-to, reference, and explanation by reader
need. Apply that distinction to content ownership; a four-directory site migration is not required.

[Codex's instruction guide](https://developers.openai.com/codex/guides/agents-md) describes a
root-to-working-directory instruction chain, per-directory override/fallback discovery, and a
context-size limit. Root guidance should be short and sufficient; nested instructions are not
uniformly discovered simply because a task touches a file there.

[Claude Code's memory guide](https://code.claude.com/docs/en/memory) documents hierarchical
`CLAUDE.md` loading, relative `@path` imports, and current native `AGENTS.md` support with
version/configuration qualifications: native discovery requires v2.1.277 or later and may be
superseded by project/ancestor `CLAUDE.md` or `CLAUDE.local.md`, or disabled by settings. Global
user/managed Claude guidance does not by itself suppress native project discovery. A wrapper
import is an available fallback, but the maintainer chose native discovery rather than retaining
that extra file. Imports load content into context; importing a whole architecture manual would
not be progressive disclosure. Normal Markdown links remain on-demand.

[Gemini CLI's context guide](https://geminicli.com/docs/cli/gemini-md/) documents global,
workspace, and just-in-time `GEMINI.md` context plus configurable `context.fileName` discovery.
Its [import processor](https://geminicli.com/docs/reference/memport/) supports `@./file.md`,
relative to the importing file, with code-span/fence exclusions and import safety checks.
A GEMINI wrapper can therefore import the same root guide without changing user settings.
This is an available compatibility option, not the repository's chosen configuration.
Gemini CLI's default discovery should not be conflated with Antigravity's native support.

[Antigravity's rules guide](https://www.antigravity.google/docs/rules/) explicitly documents
native `AGENTS.md` discovery at workspace and directory scopes across its IDE and CLI surfaces.
It also distinguishes inlining `@[label](path)` from a plain `@filename` reference; Gemini CLI's
import behavior is not interchangeable. The maintainer uses native discovery, so no `GEMINI.md`
wrapper is needed for this workflow.

## Recommendation (as of 2026-10-11)

1. Put shared contributor/agent working rules, a source-ownership matrix, and a task-based reading
   router in root `AGENTS.md`. Keep it ordinary Markdown with no tool-specific import syntax.
2. Remove repository `CLAUDE.md` in favor of native Claude `AGENTS.md` support (maintainer
   refinement after the initial wrapper proposal). Use Antigravity's native discovery too:
   do not add `GEMINI.md`. Keep only the shared root guide, with no symlinks, generated copies,
   or machine/user configuration changes. Gemini CLI users may need explicit filename
   configuration or to read the guide themselves; native support is not promised for every harness.
3. Separate three authorities: code/tests describe implemented behavior; accepted ADRs and
   compatibility promises describe intended policy; planning describes work/history. A mismatch
   is a defect or decision to resolve, not permission to rewrite policy to match code.
4. Keep a few high-risk rules immediately visible: architecture boundaries, identity/evidence,
   guarded writes, managed planning mutations, validation, user-owned changes, and Git authority.
   Route detailed procedures to their owners; do not require every guide for a typo fix.
5. Correct the README's race-test description and collapse its duplicated contributor-command
   roster into a link to the shared guide/justfile. Preserve user-facing workflows in this slice.
6. Use existing downstream tasks for the architecture split, selective package docs, executable
   examples/import inventory, and final reconciliation. Add routing/link regressions there rather
   than creating a separate overlapping enforcement project.

### Verification and limits

Walk the routes for a planning-data write, machine-contract change, CLI-help change, release,
planning triage, and documentation edit. Check destinations/anchors and command recipes against
the current checkout. Confirm the shared root guide exists, its links resolve, and neither
repository `CLAUDE.md` nor `GEMINI.md` introduces a competing entry point. Inspect that shared
guide rather than snapshotting its prose.

The installed Claude Code reports v2.1.293, above the documented native-discovery minimum;
this does not verify local settings or loaded context. Gemini CLI is not installed here.
Live client verification remains installation-specific: Claude `/context`, Gemini CLI `/memory show`,
Antigravity's active rules, and a fresh Codex session's instruction-source report can help confirm
actual loading. No model session, settings change, or Gemini installation was performed. These checks do
not prove adherence, and a smaller guide alone does not prove productivity gains. Revisit routing
after real changes and reviews; add instructions for recurring local failures, not hypothetical ones.

## Related

- [Research task](../tasks/6gjthp9frrhq-research-documentation-ownership-and-portable-agent-routing.md)
- [Ownership implementation](../tasks/6g9cz5saenme-establish-documentation-ownership-and-agent-routing.md)
- [Documentation Thread](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- [Architecture split](../tasks/6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md)
- [Package contracts](../tasks/6g9cz5sk3b0d-adopt-selective-contract-focused-go-package-documentation.md)
- [Reconciliation and drift checks](../tasks/6g9cz5t37dc0-reconcile-documentation-projections-and-enforce-drift-checks.md)
