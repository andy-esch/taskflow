---
schema: 1
id: 6gcwcf8gzn50
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Stop lint repair, link checking, and entity completion from reaching directly into the filesystem adapter.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, cli, ports, lint]
created: "2026-09-23"
depends_on: [6gcwcf88z57p]
updated_at: "2026-10-03"
audited: "2026-09-27"
audit_sources: [2026-09-28-arch-hexagonal-boundaries]
started_at: "2026-10-03"
---

# Route CLI planning data operations through application ports

## Objective

Close the remaining direct primary-to-secondary planning-data paths: frontmatter repair and body
link checking invoked outside application orchestration, and completion constructing `store.FS`
or globbing entity directories in Cobra controllers. Portable application use cases must own these
workflows so satisfying the semantic ports is sufficient without a hidden filesystem fallback.

## Scope

- Expose application-level use cases for safe frontmatter repair and optional link integrity checks,
  retaining dry-run, partial-durability reporting, and command-safety enforcement.
- Define a parse-free completion capability that can include malformed local records without making
  Cobra know the directory layout.
- Inject these capabilities through composition rather than constructing `store.FS` inside command
  controllers.
- Build on the settled entity read/source vocabulary and optional-local-capability composition;
  do not reintroduce paths through a CLI-only convenience store.
- Keep completion failure silent and fast as required by shell completion UX.

## Acceptance criteria

- [x] CLI command controllers do not import or construct `internal/store` for lint repair, link
      checking, or entity completion.
- [x] Repair and link checks run through named application use cases preserving human/JSON formats,
      exit-code and dry-run policy, and reporting the completed prefix on later failures.
- [x] Completion still offers malformed id-led local records and preserves duplicate-slug
      disambiguation without filesystem globs in the Cobra adapter.
- [x] A non-filesystem fake can drive all three workflows without providing directories or paths.
- [x] Tests pin call counts and prove no hidden fallback opens the filesystem adapter.

Implementation criteria and both review dispositions verified locally; lifecycle remains
**in-progress** pending merge.

## Out of scope

- Moving user-supplied body/manifest file reading out of the CLI adapter.
- Moving TUI filesystem watching behind a planning-data service.
- Changing completion vocabulary or repair authorization.

The older reusable-workspace task remains the owner of init, doctor, and discovery orchestration.
This task is the narrower planning-data boundary triggered by the now-explicit multi-adapter
contract; it does not reactivate that task's unrelated deferred scope.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
- [Reusable workspace discovery seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md)

## Sweep verification (2026-09-27)

Historical baseline below; the implementation described after it closes these sites.

Automated weekly sweep. Every premise in `## Objective` re-checked against HEAD and
**all three still hold** — nothing was rescoped and no acceptance criterion was ticked.
Refs verified accurate (2026-09-27):

- *"invokes frontmatter repair and link lint ports directly"* — `internal/cli/root.go:80-85`
  still declares `Fixer`/`Layout`/`Linter` as narrow ports that bypass `Service`, with the
  comment saying so outright ("none route through the Service"); they are bound at
  `root.go:454-455`, and `config.CheckLinks` is called straight from `root.go:477`.
- *"constructs `store.FS` for status-aware completion"* — `store.NewFS(root)` at
  `internal/cli/completion.go:222` (`taskCompleter`) and `:259` (`auditCompleter`), plus
  `root.go:446`.
- *"globs entity directories itself"* — `filepath.Glob` at `completion.go:190`, `:229`
  (tasks), `:241` (threads), `:266` (audits).

`docs/ARCHITECTURE.md:89` still lists this exact edge as an intentional exception
(*"`cli -> store` through `Fixer`, `Linter`, `Layout`, and completion"*), so the task and
the documented architecture continue to agree on what is being closed.

### Note for the implementer: three stale comments inside the target file

Not a change to this task's scope, recorded so the next reader does not take them as
current. `internal/cli/completion.go` carries three doc comments that still assert the
pre-ADR-0003 premise, while the code bodies directly beneath them have already been
updated to the flat layout:

- `:179-182` (`slugsFromGlobs`) — *"Because status/bucket **is** the directory, the caller
  selects which dirs to glob to filter by state."* Doubly stale: its only remaining caller
  is epic completion (`:288`), which passes a single `EpicsDir` glob and filters no state.
- `:207-208` (`taskCompleter`) — *"whose status (== their directory)"*, immediately above a
  body whose own comment correctly cites ADR-0003 §4 and parses frontmatter instead.
- `:245-246` (`auditCompleter`) — *"whose bucket (== their directory)"*, same shape.

Left untouched (this sweep does not edit `internal/`), but they sit in the exact functions
this task rewrites, so fixing them is nearly free once the work starts.

## Progress Log

- 2026-09-27: automated weekly sweep — all three cited seams re-verified present at current line numbers; no drift, nothing rescoped; flagged three stale `status == directory` comments inside `completion.go` for the implementer.
- 2026-10-03: implemented the application-port migration on `refactor/cli-planning-application-ports`;
  acceptance criteria verified locally, with external review still outstanding.
- 2026-10-03: reconciled Antigravity's no-findings report with qualified evidence. Owner reproduced
  a surviving real cwd-fallback mutation and hardened the negative-completion test with a populated
  throwaway cwd; the corrected test kills that mutant and passes restored production code. No
  production changes or new design calls; companion Codex review still pending on its captured baseline.
- 2026-10-03: Codex arrived during that reconciliation. Accepted and fixed M1 (suggestions must resolve
  to their observed sources) and M2 (selected-record suppression must not hide same-alias siblings).
  Added explicit source-certified alias safety, actual resolver enumeration, canonical-ID references,
  and completion-to-resolution/lifecycle regressions. Both new defect mutants are killed; no design
  call or separate out-of-scope implementation task is needed.

Reinforced by audit 2026-09-28-arch-hexagonal-boundaries: M4 (tracked here). The audit confirms the two `store.NewFS` construction sites (`internal/cli/completion.go:222`, `:259`) and the three direct entity-directory globs (`:229`, `:239`, `:266`), and additionally records that three contract comments in that file still document the `status == directory` model ADR-0003 §2 retired.

## Implementation notes (2026-10-03)

- `RepairPlanning` owns ordinary frontmatter → finding headers → post-write lint. It checks known
  required capabilities before writing, preserves completed/proposed results on errors, and never
  retries a multi-document operation wholesale. Dry-run does not lint an unwritten prospective state.
- `LintWithLinks` owns opt-in body-link diagnostics; ordinary lint never calls that capability.
  Ambient repository linkback warnings separately use `ConfigurationService.RepositoryLinkProblems`,
  without scanning the home registry or invoking `config.CheckLinks` from the CLI controller.
- `CompletionSource` returns canonical identity, human search label, adapter-owned resolution reference,
  explicit `SlugIsReference` evidence, and optional observed state. A label is not a selector merely
  because it appears once. `CompleteEntities` owns prefix filtering, duplicate disambiguation,
  already-typed suppression, and deterministic order. The filesystem adapter reads names only for
  plain completion; state-aware requests read each candidate's own metadata. Broken frontmatter
  remains addressable by bare canonical ID. Resolver candidate enumeration is shared, including
  case-insensitive README exclusion and symlink rejection. Flat kinds exclude non-id-led strays;
  epics retain legacy filename keys. Aliases must be safe under case folding and ID precedence;
  duplicate canonical IDs are omitted rather than given a misleading selector.
- All three optional capabilities participate in the existing checked source-set composition;
  absent, typed-nil, missing-witness, and mismatched capabilities cannot create a local fallback.
  Neither the aggregate `Store` nor the public machine schema was widened.
- Cobra's initial `__complete` hook precedes parsing the completed command's flags. Composition
  therefore runs through an injected deferred factory at the candidate callback, preserving `-C`,
  explicit `--space`, pointer spaces, and flag/environment precedence.

Deliberate correctness improvements alongside the refactor: state exclusion applies to a source,
not its slug (so a damaged duplicate is still offered by its unambiguous reference); already-typed
bare IDs and references suppress only the same record, not same-label siblings; and a post-lint
failure now renders the completed repair prefix that the previous CLI discarded. Existing result
envelopes and exit classification stay intact.

Evidence: portable fakes exercise actual Cobra commands for all five entity completions, repair,
and link lint; source-set tests cover the new capabilities; core tests execute real body transforms
and fail a later document; existing filesystem repair, safety, completion, and wire suites remain.
Validation passed: `go test ./...`, `go test -race ./...`, `just lint`,
`just docs-check`, planning lint, and the dogfooded Thread frontier.

Remaining boundary enforcement belongs to
[isolate CLI composition wiring](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md),
not this task. Root composition and the explicitly local watcher `Layout` are retained here.

Completed external reviews (original briefs retained; owner dispositions recorded):

- [Codex implementation review](../audits/6gg59f816kn3-2026-10-03-cli-planning-application-ports-implementation-codex.md)
- [Antigravity implementation review](../audits/6gg59f89zk42-2026-10-03-cli-planning-application-ports-implementation-antigravity.md)
