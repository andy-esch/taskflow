---
schema: 1
id: 6g6x7e2ef37r
status: completed
epic: 21-code-quality-architecture-hardening
description: Revise the oversized architecture document into a smaller navigable structure with clear ownership and durable cross-links.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [docs, architecture, maintainability]
created: "2026-09-04"
depends_on: [6g9cz5saenme]
updated_at: "2026-10-12"
started_at: "2026-10-12"
completed_at: "2026-10-12"
---

# Restructure the architecture documentation into focused guides

## Objective

Make the architecture documentation easier to navigate and maintain by replacing the oversized
single-file structure with a concise entry point and smaller, clearly owned guides organized around
stable architectural concerns. Use the source-of-truth matrix from the documentation-ownership task
to decide what belongs in the index, a focused guide, generated output, package documentation, an
ADR, or planning history.

## Acceptance criteria

- [x] Define a simple target information architecture before moving content.
- [x] Break the current architecture document into focused guides with clear scopes and useful
  navigation from a central architecture entry point.
- [x] Preserve meaningful architectural decisions and context while removing obsolete or duplicate
  prose deliberately rather than accidentally.
- [x] Repair repository links and contributor guidance that reference the current document layout.
- [x] Documentation generation/checks, link validation where available, and planning lint pass.

## Out of scope

- Changing product or code architecture merely to match the new document layout.
- Re-litigating accepted ADR decisions unless the restructuring exposes a concrete contradiction
  that should be tracked separately.
- A comprehensive rewrite of unrelated user or CLI documentation.

## Research handoff (2026-10-11)

The [ownership research](../research/6gjggjw01raw-documentation-ownership-and-portable-agent-routing.md)
measured the current reference at 1,087 lines / 11,466 words (baseline 09974fb), larger than the
Thread's original estimate. Inventory its dated `Status (2026-08-25)` and review-history sections
separately from current contracts; do not leave historical feature snapshots presented as current
guidance. Preserve accepted decision history in ADRs and useful review provenance in planning.
Keep the planning-data checklist's owner and old incoming links explicit during extraction.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- Follows [Establish documentation ownership and agent routing](6g9cz5saenme-establish-documentation-ownership-and-agent-routing.md)

## Layout decision (2026-10-12)

Keep `docs/ARCHITECTURE.md` as the short entry point and compatibility-anchor router. Six canonical guides in `docs/architecture/` own boundaries/composition, planning reads/projections, persistence/recovery, configuration/spaces, CLI/wire, and the TUI. Guides link to the owning Go types and regression tests; relevant package comments point back using repository-relative guide paths. Detailed API contracts stay beside their symbols. No new product behavior or compatibility promise is intended.

Retain the manual import map in the boundaries guide for the already-sequenced import-check task. Retain historical feature/review provenance here or via immutable baseline links, not as current guide policy. Expanded `doc.go`, executable examples, and broader drift enforcement stay with existing sibling tasks. Validate this slice's navigation and package-comment routes without claiming prose is mechanically proven.

## Archived architecture snapshot

The following is the old `Status (2026-08-25)` section as found at baseline 4fdf6fcc, not a verified snapshot of that date; later edits had accumulated beneath its dated heading. Preserved as history rather than current implementation guidance.

```markdown
## Status (2026-08-25)

The layered shape now has two substantial primary consumers and several reusable
application seams; it is no longer architecture held in reserve for a hypothetical UI:

- The CLI exposes full task lifecycle and body mutation, epic/audit/research authoring,
  schema and wire discovery, configuration lifecycle, planning-space registry verbs,
  global `--space` selection, and the cross-space `status --all` overview.
- The Bubble Tea app provides its dashboard plus task/epic/audit/research tabs,
  a cross-space atlas with reversible entry-point navigation, filtering/sorting/find,
  structured navigation, task and audit lifecycle actions, task field editing,
  active-space filesystem reloads, and the embedded configuration editor. Reads and
  writes remain asynchronous commands over core services; production TUI code does not
  open planning/config stores.
- The configuration and cross-space features have dedicated core services and composite
  filesystem adapters. `SpaceRegistryService` owns catalog/selection/mutation policy and
  `SpaceOverviewService` owns cross-space summaries; Cobra is one primary adapter over
  both rather than the owner of either use case.
- Writes use atomic replacement/exclusive create, a repo-wide lock, content-version CAS,
  bounded retry for agent mutations, and parse-before-accept editor loops. Machine output
  is a versioned `internal/wire` contract with generated JSON Schema and golden coverage.
- The planning model's closed vocabularies now have write verbs rather than only linters.
  An acceptance criterion carries a state and a reason (`task ac`), an audit finding
  carries a status, resolution paragraph, and optional versioned candidate projection
  (`audit finding`), and the words the two
  share are declared once in `domain/resolution.go`. Two invariants are enforced at write
  time rather than reported after: `task complete` refuses a task with an unexplained
  unmet criterion, mirroring `audit close` requiring terminal findings and complete parsing.

The atlas decision has now activated the narrow reusable workspace-opening boundary:
`core.WorkspaceService` and `internal/workspacestore` can open an explicit local entry
point for a primary adapter without exposing discovery or concrete persistence. The
broader `doctor`/fix promotion remains deferred because the atlas does not consume it.
Thread `context.Context` when an HTTP request path exists; reshape findings pagination
when a web findings view is scoped. Until those triggers occur, package movement or
speculative interfaces are out of scope.
```

## Extraction and preservation evidence (2026-10-12)

| Old reference content | New owner / disposition |
| --- | --- |
| Rule, package roles, manual import map, fitness policy, named exceptions and consolidation rationale | [Boundaries](../../docs/architecture/boundaries.md); preserve explicit exceptions and followup limits, remove volatile type counts |
| Planning-data checklist and core read/projection contracts | [Planning data](../../docs/architecture/planning-data.md); owning types and behavioral tests remain authoritative |
| Store, source-preserving YAML, CAS, creates, rename, graph/lifecycle/apply and audit-body guards | [Persistence](../../docs/architecture/persistence.md); distinguish structured guards from tracked generic-body gaps |
| Discovery, preference scope, configuration/registry services and workspace opening | [Configuration/spaces](../../docs/architecture/configuration-and-spaces.md); replace the stale apply-parity-gap claim with shipped regression evidence |
| CLI output/DI, queries, machine revision and golden/process validation | [CLI/wire](../../docs/architecture/cli-and-wire.md); no changes to command behavior, schema revision or compatibility promises |
| TUI shell, identity, atlas, contention, graph focus, watchers and layout | [TUI](../../docs/architecture/tui.md); remove exhaustive file/key roster in favor of owner/test routes |
| Dated Status section and review history | Archived section above, immutable pre-split reference, and original linked planning records; not new current policy |

The original architecture heading anchors remain in the index as forwarding sections. README,
AGENTS, the weekly architecture routine, and eight package comments now route to the focused
owners. After review, the root index is 99 lines / 613 words; index plus six guides is 5,794 words versus the
11,466-word baseline. These are measurements, not a prose-size policy.

The new test-only architecture navigation check runs in the ordinary suite. It checks local inline
links/ATX heading anchors in the index, guides, AGENTS and README; each guide must link back to the
index and to an owning package comment with a return route. It also checks incoming task/Thread
links to retained index anchors and excludes fenced/inline-code decoys. It does not parse all Markdown dialects,
validate network resources, prove prose/API compatibility, or replace import-map/example work.
Compiler-valid owner probes broke the core guide anchor and then removed its return route; each
failed TestArchitectureNavigation for the intended reason. The source comment was restored and
focused race-enabled tests passed.

Owner validation: just test (race-enabled), just lint (0 issues), just build, just docs-check,
planning lint --links and git diff --check passed. go doc ./internal/core shows the focused
checklist route. The independent review and its dispositions are recorded below; the checks
remain bounded evidence, not proof of semantic prose fidelity.

## Independent review handoff (2026-10-12)

Prepared empty, open audits for [Claude](../audits/6gk3mt5hz2vj-2026-10-12-focused-architecture-guides-implementation-claude.md)
and [Antigravity](../audits/6gk3mt5v19dw-2026-10-12-focused-architecture-guides-implementation-antigravity.md).
Claude leads semantic preservation; Antigravity leads hostile navigation/test maintenance, with
cross-checks of shared safety contracts. Both require independent captured sandboxes, a tested
compiler-valid navigation mutation, source/guide/test traces, and guarded single-audit transfer.
At handoff, review had not run and the task remained in progress pending disposition. No
implementation commits or pushes had been made in this slice.

## Review disposition (2026-10-12)

Claude's four Low findings were verified and fixed via audit finding verbs: correct owner/test routes, navigable-link assertions, relocated consumer/forwarding-anchor coverage, and restored lifecycle task+Thread snapshot CAS. Nearby userconfig wording now names the actual lint rule rather than compiler enforcement. Three owner mutations (fenced index link, inline-code back-link, renamed retained anchor) fail at the intended assertions and pass after restoration; focused source-precedence/dashboard and lifecycle raw-Thread-race tests also pass.

Antigravity contributed independent checks and two killed mutations, but its exhaustive verdict and two nonexistent destination references were not accepted. The original report is preserved with an explicit owner correction. Future reviews should challenge the actual assertions with decoys and verify every destination, not generalize a helper test to all callers. Both landed reports matched their retained sandbox copies before owner edits; pending transfer attestations were not rewritten into invented final results.

Full race tests, lint (0 issues), build, generated CLI drift check, planning link lint and whitespace checks pass. Production behavior and wire revision are unchanged. Broader Markdown/routine/semantic reconciliation and selective package overviews remain in the existing threaded tasks; the two stale open-task ACs were re-routed without checking them off. This slice is ready for PR handoff; no commit or push was made during review disposition.

## CI portability followup (2026-10-12)

[PR #303](https://github.com/andy-esch/taskflow/pull/303) exposed lowercase justfile links that macOS accepted but Linux rejected; Git tracks Justfile. Corrected the AGENTS, README and research routes. The navigation guard now checks exact directory-entry spelling for both file and parent-directory components, without resolving away symlink names. The new regression failed on all three wrong-case variants before the fix and passes afterward.

The CI-equivalent go test -p 4 -race -coverprofile=<temporary output> ./... run passed locally, as did lint, build and generated-doc checks. Coverage stayed outside the worktree. This corrects navigation/test portability, not product behavior or the wire contract.
