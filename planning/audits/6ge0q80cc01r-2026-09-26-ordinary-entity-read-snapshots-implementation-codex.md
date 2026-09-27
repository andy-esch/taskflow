---
schema: 1
id: 6ge0q80cc01r
bucket: closed
area: ordinary-entity-read-snapshots-implementation-codex
date: "2026-09-26"
updated_at: "2026-09-26"
---
# Audit: Ordinary entity read snapshots implementation — codex — 2026-09-26

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Adversarially review the implementation of task `6gcwcf80v8hg`,
`make-ordinary-entity-list-diagnostics-adapter-neutral`. The intended outcome is an
entity-specific, adapter-neutral read boundary, not merely a rename of `FileProblem`.
Treat checked acceptance criteria and passing tests as hypotheses. Find concrete runtime
defects, compatibility regressions, hidden rereads, or systemic gaps; distinguish them
from work explicitly assigned to the later source/path, TUI, and source-set tasks.

## Review target

Review the entire uncommitted delta on branch `refactor/portable-entity-read-snapshots`
against base `6a3e6f49df605ad7e2bf0e233ea7a494b1184b5f`. Include untracked files.
Primary planning sources are:

- `planning/tasks/6gcwcf80v8hg-make-ordinary-entity-list-diagnostics-adapter-neutral.md`;
- `planning/tasks/6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md`;
- `planning/threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md`.

Trace the new `internal/core/entity_read.go` values through `core` ports/services,
`internal/store` scans and conversions, `internal/cli` human and projected output,
`internal/wire` DTOs/schema/goldens, and `internal/tui` list/detail adapters. Review
tests as possible sources of false confidence. Do not implement fixes or edit any file
other than your assigned audit.

## Intended contract to challenge

1. `RecordSource.ID` is the canonical readable-record identity; declared `ID`,
   `FilenameID`, slug, path, and opaque `Location` cannot override it. Domain-carried
   fields remain temporarily for lint and older presentation consumers only.
2. Task, epic, audit, and research ports remain entity-specific while sharing
   `LoadedRecord` and `LoadProblem`. Ordinary list/show results expose no
   `domain.FileProblem`; no new generic entity repository or kind switch appears.
3. A filesystem task request performs one authoritative body-bearing/versioned scan
   and projects graph, ordinary, and lint views from it. Board, Summary, finding
   queries, and projected CLI output do not hide a second scan or combine generations.
4. Unreadable records retain kind, recoverable ID/slug, optional opaque location,
   independently optional local repair path, and message. Core never parses a
   location to manufacture identity. A URI never becomes the historical JSON `path`.
5. Local partial reads still show available records and actionable errors, then
   exit 11. Full and column-projected JSON preserve a required `path` field (empty
   when no local path), optional portable fields, and the correct schema revision.
6. Explicit source identity survives graph analysis, epic joins, lint attribution,
   Board/status, TUI list identity, full and projected JSON, and each show/info DTO.
   Duplicate or drifting IDs remain visible and cannot silently become actionable.
7. The migration has not weakened graph health, unreadable-source revision evidence,
   optimistic mutation/CAS, snapshot ordering, or the existing Thread read contract.
8. Transitional bare-domain projections are local and explicit, not a second
   competing source-of-truth that future adapters will accidentally depend on.

## Mandatory evidence floor

- Produce a **consumer inventory** for all changed ordinary read ports and their
  downstream consumers. Classify every remaining core-facing `FileProblem`,
  `Path`, `FilenameID`, `CanonicalID()`, `LocationIsPath`, and `SourceVersion` use as
  intentional guarded/local compatibility, deferred work, or current violation.
  Do not infer completeness from a single `rg` result or one interface definition.
- Build a field-lineage table for source ID, declared ID, slug, location, local path,
  message, and source revision for readable and unreadable task/epic/audit/research
  records. Cite producer, conversion, core consumer, human output, wire output,
  and mutation evidence separately.
- Use independent fixtures/fakes for all four kinds: matching identity; missing or
  drifting declaration; stale `FilenameID`; explicit ID plus misleading URI/path;
  malformed document with recoverable ID+slug; malformed filename with neither;
  duplicate canonical IDs with different locations; empty location; and both
  opaque location and local repair path. State where current behavior is necessarily
  limited by the later source/path task.
- Measure adapter calls and filesystem reads for graph, ordinary task list, lint,
  Board, Summary, audit finding queries, and show/info. Check normal and partial
  snapshots; compare list/body/graph projections from the same bytes. Look for
  double hashing, duplicate parsing, hidden rereads, and mutation-revision loss.
- Independently exercise full JSON, `--json -c`, human output, stderr/stdout, and
  exit 11 for local and pathless failures. Validate non-empty `unreadable` arrays
  against generated JSON Schema; inspect the 1.75→1.76 changelog/golden delta for
  unrelated drift and confirm generated CLI docs remain current.
- Trace source-ID precedence through task graph and dependency lint attribution,
  epic filter/join/rollup, audit findings, Board blocked ordering, current and
  cross-space status, TUI registry/detail, and every show/info converter. Include
  contradictory embedded IDs and locations, not only happy-path equal values.
- Perform mutation testing in the sandbox. At minimum try reverting source-ID
  precedence in one graph/Board path, reconstructing a problem ID from location,
  emitting opaque location as JSON path, adding a second task or audit read,
  dropping unreadable source revision, and changing one projected `id` column
  back to the domain field. For each probe, report whether an existing focused
  test fails for the intended reason. Do not count compile failure as detection;
  coordinate cross-layer mutations where needed to keep the program compiling.

## Required hostile angles

- **False portability:** Can a pathless adapter implement the new ports without
  supplying fake local paths, while list/show/lint/Board/status/TUI remain usable?
  Does a compatibility branch still consult `Value.Path` or `FilenameID` when
  `Source` disagrees?
- **Mixed-generation reads:** Could an ordinary list combine a graph from one
  generation with epics/audits from another and claim a coherent answer? Identify
  existing guarantees versus aspirations; only report a regression or concrete
  untracked requirement, not generic cross-entity non-atomicity.
- **Corrupt corpus:** Are duplicate source IDs, malformed frontmatter, absent IDs,
  legacy epic names, and unreadable source versions retained and attributed without
  suppressing good records or granting unsafe mutation?
- **Contract drift:** Do generated schema, DTO comments, projected columns, golden
  fixtures, docs, and exit-code behavior describe the same observable result?
- **Future migration pressure:** Does any new helper make the sequenced removal
  of `Path`/`FilenameID`/`SourceVersion` harder or accidentally create a generic
  repository abstraction under an innocuous name?

For Antigravity in particular, do not stop at a broad inventory or a quick
ready/no-findings verdict. Pick at least three seams above, construct contradictory
cross-layer inputs, run focused probes, and show the exact observation that
supports or falsifies each claim. If a probe is blocked, say precisely why and
choose another. For Codex, emphasize systemic failure modes and whether narrow
tests mask a neighboring consumer that still trusts domain or path metadata.

## Validation and restoration

Work only in the mandatory independent sandbox created by the generated audit's
instructions. Establish the baseline with `go test ./...`, then run focused
tests/probes with a sandbox-local or `/tmp` Go cache. Inspect `git diff --check`,
generated CLI docs/schema, and the relevant linter if useful. Restore every
mutation, generated file, and fixture after each experiment. Before transfer,
the assigned audit must be the only difference from the sandbox baseline.
Never commit to or alter the shared source checkout.

## Deliverable

Preserve this brief and append your report to your assigned audit. Use exact
finding headings such as `#### H1. ... · **Status:** open` and leave every
finding open for owner triage. Each finding needs a concrete reproduction or
code path, actual versus expected result, severity, consequence, and the smallest
plausible remedy. Separate confirmed defects, untested risks, and intentionally
deferred work. A no-findings verdict requires an evidence-backed inventory and
probe results, not simply a passing suite. Include the sandbox attestation and
transfer result required above.

## Reviewer report

_Reviewer: append verdict, evidence inventory, probes, findings (if any),
scope distinctions, and isolation/transfer attestation below this line._

### Verdict

**Changes requested.** The filesystem path and the common nonempty-source-ID path work, but a malformed portable read can promote a document-declared task ID into an authoritative, eligible graph node. Independently, readable records' opaque locations are discarded before duplicate diagnostics and ordinary wire rows, leaving physically different duplicate occurrences indistinguishable. Both were reproduced with contradictory source/domain inputs in the isolated workspace. No implementation file was changed in the deliverable.

### Consumer inventory and field lineage

The checked implementation is the captured, uncommitted delta against `6a3e6f49df605ad7e2bf0e233ea7a494b1184b5f`, including `internal/core/entity_read.go` and `internal/store/entity_read.go` (both were untracked in the handoff). The planning task's checked boxes were treated as requirements, not evidence.

| Entity-specific port and producer | Core consumers and primary projections | Evidence |
| --- | --- | --- |
| `TaskStore.ReadTasks/ReadTask`, `TaskGraphSource.ReadTaskGraph`, `LintSource.ReadLintTasks`; filesystem `ReadTasks` delegates to `ReadTaskGraph`, both use `scanTaskDocuments` | `ListTasks`, `ShowTask`, Board, Summary/status, strict dependency graph, lint, epic rollups/show, Thread task views, CLI task list/show/info and Board/status, TUI task list/detail | `internal/core/store.go:15-19`; `internal/core/service_task.go:22-97,139-171,197-220`; `internal/store/fsstore.go:125-198`; `internal/store/entity_read.go:56-70`; `internal/core/board.go:42-80`; `internal/core/service.go:363-393`; `internal/tui/commands.go:25-68`; `internal/wire/envelopes.go:40-48,82-93,110-127` |
| `EpicStore.ReadEpics/ReadEpic`, `SummaryStore.ReadEpics`, `LintSource.ReadLintEpics`; filesystem `ListEpics/GetEpic` | task epic validation, `ListEpics`, `ShowEpic`, Summary/status, lint, CLI epic list/show, TUI epic list/detail | `internal/core/store.go:181-184,304-306`; `internal/core/service_task.go:37-48`; `internal/core/service_epic.go:154-192,378-405`; `internal/core/service.go:372-379,641-654`; `internal/store/entity_read.go:74-92`; `internal/tui/commands.go:124-167`; `internal/wire/envelopes.go:734-759` |
| `AuditStore.ReadAudits/ReadAudit`, `AuditSnapshotSource.ReadAuditSnapshot`; filesystem `ListAudits/GetAudit`, `ListAuditsWithFindings` | `ListAudits`, `ShowAudit`, Summary/status, `QueryFindings`, lint, CLI audit list/show/info/findings, TUI audit list/detail | `internal/core/store.go:233-236`; `internal/core/audit_source.go:4-22`; `internal/core/service_audit.go:67-99`; `internal/core/finding.go:153-178`; `internal/core/service.go:383-413,691-713`; `internal/store/entity_read.go:96-115`; `internal/store/lintsource.go:30-68`; `internal/tui/commands.go:173-205`; `internal/wire/envelopes.go:768-793,895-910` |
| `ResearchStore.ReadResearch/ReadResearchDocument`, `LintSource.ReadLintResearch`; filesystem `ListResearch/GetResearch` | `ListResearch`, `ShowResearch`, lint, CLI research list/show, TUI research list/detail | `internal/core/store.go:267-269`; `internal/core/service_research.go:110-145`; `internal/core/service.go:659-682`; `internal/store/entity_read.go:119-138`; `internal/tui/commands.go:284-313`; `internal/wire/envelopes.go:798-824` |

| Field, readable and unreadable lineage | Conversion and core use | Human, wire, mutation evidence |
| --- | --- | --- |
| **Source ID and declared ID.** Task/audit/research parsers recover filename ID into `FilenameID`, while YAML `id` stays `ID`; epic parser sets `ID` from its whole filename stem. `taskRecord`, `epicRecord`, `auditRecord`, `researchRecord` copy the adapter-resolved ID to `RecordSource.ID`. For unreadable id-led files, `scanDirSourcesWithReader` recovers ID/slug before conversion; epic `ListEpics` recovers the stem. | `internal/store/fsstore.go:348-393`; `internal/store/auditstore.go:188-216`; `internal/store/researchstore.go:74-104`; `internal/store/epicstore.go:18-42,257-276`; `internal/store/resolve.go:66-78`; `internal/store/entity_read.go:9-47`. Core list/show retains `LoadedRecord`; graph/Board/status/lint temporarily inject source ID into domain `FilenameID` (`internal/core/service_task.go:210-220`; `internal/core/service.go:393,603,678,710`). The empty-ID exception is finding M1. | Full ordinary wire and show/info converters take `Source.ID` (`internal/wire/dto.go:53-56,246-269,370-373`; `internal/wire/envelopes.go:110-127,150-155,751-824`). Projected task/audit/research `id` uses `loadedRecordColumns` (`internal/cli/render/columns.go:400-405,461-486,493-517`); epic rollup copies source ID into its bare display value (`internal/core/service_epic.go:186-192`). Human list labels use slug or epic ID; TUI projects source IDs before registry validation (`internal/tui/commands.go:43-68,190-205,294-307`; `internal/tui/entity.go:72-94`). Filesystem CAS still resolves canonical filename identity (`internal/store/cas.go:155-175`). |
| **Slug, location, local repair path, message.** Readable slug is parsed from filename for task/audit/research; epic uses its stem. Readable `Source.Location` is the adapter's opaque context. Unreadable `EntitySlug`, `Location`, `LocalPath`, `Message` come from `scanDirSourcesWithReader` and `loadedProblems`; a malformed non-id-led task/audit/research filename has no recoverable ID/slug. | `internal/store/resolve.go:53-82`; `internal/store/entity_read.go:49-56`; `internal/core/entity_read.go:17-47`. `taskGraphLoadProblems` and `canonicalLoadProblems` keep explicit fields and never parse location for identity (`internal/core/service_task.go:258-293`; `internal/core/lint_source.go:22-65`). Readable location is lost in graph/bare-domain projection; see M2. | `render.LintProblemsHuman` prints identity, location, and distinct repair path (`internal/cli/render/render.go:964-998`); `portableProblemsError` leads with slug/ID (`internal/cli/problems.go:57-91`). `ToLintLoadProblemsJSON` maps only `LocalPath` to required `path`, separately from optional `location` (`internal/wire/envelopes.go:950-974`). No ordinary revision reaches wire. Repair/path navigation remains local (`internal/core/store.go:18-24,235-239,270-274`), pending the later path-capability task. |
| **Source revision.** Task parser hashes readable exact bytes into `Task.SourceVersion`; versioned scan hashes unreadable bytes into `TaskGraphLoadProblem.SourceVersion`. Epic/audit/research ordinary reads carry no revision. | `internal/store/fsstore.go:140-184,393`; `internal/store/resolve.go:61-79`; `internal/core/service_task.go:101-119,230-236`. `SameSourceSnapshot` requires equal nonempty readable/unreadable revisions (`internal/core/dependency_graph.go:940-971`). | Version is absent from human/JSON DTOs; guarded graph mutations compare it (`internal/store/graphmutation_test.go:249-282`). Ordinary audit/research/epic writes instead gather fresh evidence in their store write paths (`internal/store/auditstore.go:128`; `internal/store/researchstore.go:171`; `internal/store/epicstore.go:132`). |

Remaining crossing classification: `domain.FileProblem` is absent from the changed ordinary ports but remains in local compatibility constructors/conversions (`internal/core/service_task.go:178-245`, `internal/core/dependency_graph.go:323`) and the legacy `render.ProblemsHuman` (`internal/cli/render/render.go:955`). `Path`, `FilenameID`, and `CanonicalID()` in graph, Board, Summary, lint and TUI are guarded compatibility projections where explicit source identity is injected first (`internal/core/service_task.go:210-220`, `internal/core/service.go:393,603,678,710`, `internal/tui/commands.go:43-47,190-193,294-298`); the empty-ID fallback and discarded readable location are current violations. `LocationIsPath` and `Path` on task/Thread load problems are local compatibility (`internal/core/service_task.go:109-117,248-274`; `internal/core/service.go:571-577`); graph `Path` is still a local diagnostic field. `SourceVersion` stays guarded comparison evidence (`internal/core/dependency_graph.go:940-971`), not a public list field. Thread-specific `Path`/`FilenameID`/`SourceVersion` and CLI/TUI local editor/path uses are deferred to the sequenced source/path and TUI tasks, not findings against this slice. No generic entity repository or kind-switched read method was found: `LoadedRecord`/`LoadProblem` are values, and the four ports remain distinct (`internal/core/entity_read.go:17-80`, `internal/core/store.go:14-19,181-184,233-236,267-269`).

### Adversarial probes and observed limits

- **Contradictory identity and duplicate location.** A sandbox-only `TestReviewEmptySourceID` gave `TaskGraphRead.Records` a task with declared `6g0000000001`, stale `FilenameID`, `Source.ID=""`, and `Source.Location="db://tasks/other"`. `NewTaskGraphRead` returned `health=healthy`, `ids=["6g0000000001"]`, no problems, and `Eligible=true`; Board showed the declared ID. Changing that probe to assert broken health failed: `missing source ID must fail closed; health=healthy eligible=true`. A second probe supplied two same-ID/same-slug tasks at `db://a` and `db://b`: graph health was broken, but both duplicate messages said `across <unknown path>, <unknown path>` and exposed empty `Path`. A wire probe with same-source-ID, same-value records at those two locations produced **identical full JSON rows** for task, epic, audit, and research (two rows per kind). These temporary probe files were removed.
- **Matching versus contradictory source identity.** `TestOrdinaryReadEnvelopesPreferSourceIdentityForEveryEntity` covers source ID against stale declared/filename IDs in all four kinds (`internal/wire/envelopes_test.go:80-119`); `TestEpicReadIdentityAndRollupUseExplicitSource`, `TestSummaryBareProjectionsUseLoadedSourceIdentity`, `TestBoard_BareProjectionUsesExplicitSourceIdentity`, and TUI's registry check cover neighboring consumers (`internal/core/service_epic_test.go:178`; `internal/core/usecases_test.go:231`; `internal/core/board_test.go:128`; `internal/tui/entity.go:72-94`). Empty `Source.Location` with nonempty ID is accepted by the source value, as intended; only task graph behavior for that case was exercised independently. Epic legacy filenames retain a recoverable stem even when not id-led (`internal/store/epicstore.go:29-35`), so the “neither ID nor slug” filename case applies to the other three kinds.
- **Local partial corpus.** A temporary independent fixture put one good and one malformed-frontmatter document in each of task/epic/audit/research, plus a task `stray.md` without recoverable identity. For each kind, `list --json`, `list --json -c id`, and human `list` retained the good row and returned exit 11. JSON stdout was parseable, contained revision `1.76` and nonempty `unreadable`, with no stderr text; human stdout held the good row and stderr held the bad-record detail. The task stray printed `unidentified task record`, while the id-led bad task retained slug and ID. The output recommended moving `stray.md` to `meta/` or deleting it, but emitted no executable repair command to run. No remote CLI/TUI adapter exists in the captured implementation; pathless behavior beyond the core/wire fakes remains unproved end-to-end.
- **Schema and generated evidence.** A temporary schema probe compiled `TasksEnvelope`, `EpicsEnvelope`, `AuditsEnvelope`, and `ResearchListEnvelope` and validated each with a nonempty unreadable array whose `Location="db://wrong/opaque"` and `LocalPath="/tmp/repair.md"`; all four passed, with `path` equal to the local path. `internal/wire/envelopes_test.go:53-77` separately checks pathless `path=""`. The 40 changed golden files are revision-only after normalizing `schema_version`, except the generated JSON Schema: its four ordinary `unreadable` refs change from `FileProblem` to `LintLoadProblemJSON`, the unused `FileProblem` definition disappears, and schema revision markers change. `internal/wire/wire.go:315-321` documents 1.76. Generated CLI docs and schema comments matched their generators (`go run ./internal/tools/docgen -out /private/tmp/taskflow-codex-generated-docs` plus `diff -ruN`; `go run ./internal/tools/schemacomments -out /private/tmp/taskflow-codex-schema-comments.json` plus `diff -u`).
- **Reads and mutation evidence.** Task graph, ordinary list, lint, and Board each enter one `scanTaskDocuments`; its scanner executes one `readFile(path)` per regular Markdown file (`internal/store/fsstore.go:125-198`; `internal/store/resolve.go:36-82`). Summary makes one graph, epic, and body-aware audit snapshot call; findings make one audit snapshot call; selected show/info uses one `Get*` body read and no second body read (`internal/core/service.go:363-385`; `internal/core/finding.go:153-166`; `internal/store/entity_read.go:65-138`). Existing call-count tests pass: `internal/core/board_test.go:82-98,128-142`, `internal/core/lint_source_test.go:46-73`, `internal/core/usecases_test.go:194-215,259-280`, `internal/core/finding_test.go:83-107`, and `internal/core/service_epic_test.go:611-633`. The scanner hashes every task before parse and `parseTask` hashes a valid task again (`internal/store/resolve.go:61-65`; `internal/store/fsstore.go:393`): ordinary list/lint now pay two SHA-256 passes, although still one file read/parse. No content-generation mixing within one entity request was observed; separately read task/epic/audit families in Summary have no cross-family atomicity promise in this slice.

**Mutation-test ledger.** All edits below were applied one at a time to the sandbox, compiled, tested, and restored. Replacing graph's source-ID injection with declared ID failed `TestBoard_BareProjectionUsesExplicitSourceIdentity` with `board identity = "stale-declaration"`. Recovering a missing problem ID from `db://tasks/6g0000000009-wrong.md` failed `TestBoard_DoesNotInferIdentityFromOpaqueLocation` with manufactured `EntityID=6g0000000009`. Emitting `Location` as wire `path` failed `TestToLintLoadProblemsJSONKeepsOpaqueLocationsOutOfPath`. Adding a second Board graph read failed its explicit-source test with `graph read calls = 2`; adding a second Summary audit read failed `TestService_Summary_ReadsEachAuditOnce` with `called 2 times`. Dropping the unreadable task revision failed `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` on empty revision. Changing the shared projected `id` extractor back to the domain field **passed** the existing `TestColumnRegistriesMatchFullWireValues`, because its source and declared IDs agree (`internal/cli/render/columns_contract_test.go:158-213`); a temporary contradictory `TestReviewProjectedSourceID` failed with `projected id="declared", want canonical`. This is a demonstrated test blind spot, not a shipped projection defect. No mutation was credited for a compile error.

### Confirmed findings

#### M1. Empty explicit task source ID falls back to declared identity and becomes eligible · **Status:** fixed

**Reproduction:** The sandbox-only `TaskGraphRead{Records: []LoadedRecord[domain.Task]{{Value: Task{ID:"6g0000000001", FilenameID:"stale-id", Status:ready-to-start}, Source: RecordSource{ID:"", Location:"db://tasks/other"}}}}` produced healthy graph state and an eligible task under `6g0000000001`; the fail-closed assertion failed as quoted above. `taskGraphTasks` overwrites `FilenameID` with the empty source ID (`internal/core/service_task.go:210-220`); `canonicalTaskID` then falls back to the declared ID (`internal/core/dependency_graph.go:571-576`). Board uses the resulting `CanonicalID()` and graph eligibility (`internal/core/board.go:58-70`). TUI also injects an empty source ID before using `CanonicalID()` (`internal/tui/commands.go:43-63`), so its ostensibly strict `validateEntityItems` can receive a declared-ID key (`internal/tui/entity.go:72-94`).

**Actual versus expected:** A readable record without canonical source identity is treated as addressable and startable. The design requires it to become a `LoadProblem`; declared ID and stale filename metadata must never override `Source.ID` (`planning/tasks/6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md`, “Decision”). **Consequence/severity:** Medium: a faulty or incomplete pathless adapter can make Board/graph/TUI claim a task is safe to select; the guard is illusory at the application boundary. Existing filesystem reads always populate the source ID, so this does not reproduce with the current local store. **Smallest remedy:** Reject a `Records` entry with empty `Source.ID` at the graph/read boundary or convert it to an explicit load problem before graph, Board and UI projections; never use declared ID as fallback for an explicit record. Add the contradictory failing probe as a permanent focused test.

**Resolution:** Explicit loaded task records without a canonical source ID now
become load problems before graph, Board, list, and TUI projection; ordinary
show/list and audit bulk-fix boundaries also reject missing source identity.
Added contradictory-ID regressions.

#### M2. Readable source locations vanish before duplicate attribution · **Status:** tracked by 6ge1bacd3bd2

**Reproduction:** Two readable task records with identical valid source ID and slug but `Source.Location` values `db://a` and `db://b` yield two graph duplicate-ID problems that both say `<unknown path>` and expose no distinguishing location. Full JSON list constructors for all four kinds emit two identical rows when supplied equal semantic values and the same source ID at those two locations; the sandbox-only wire probe logged `identical=true` for task, epic, audit and research. `taskGraphTasks` carries only source ID into the bare task (`internal/core/service_task.go:210-220`), while graph duplicate text and source refs read `Task.Path` (`internal/core/dependency_graph.go:385-424`; `internal/core/dependency_source.go:296-298`). Ordinary DTOs include the source ID but no readable source location (`internal/wire/dto.go:24-40,190-210,217-269,352-385`; `internal/wire/envelopes.go:40-48,734-820`).

**Actual versus expected:** Graph correctly marks the duplicate as broken and TUI rejects duplicate registry keys, but CLI list/lint cannot tell a user which of the two physical records needs repair. The design requires CLI list/lint to attribute every duplicate occurrence with available source context, without making location an identity or local path. **Consequence/severity:** Medium: a pathless adapter can report corruption but provide two indistinguishable occurrences for inspection; this undermines the migration's diagnostic goal. Existing filesystem files normally have distinct path-derived slugs, so the failure appears with a portable adapter that can return repeated semantic slugs. **Smallest remedy:** Preserve `RecordSource.Location` through graph problem attribution and expose optional readable location context in duplicate diagnostics/list rows (with a wire revision as needed); keep canonical ID as the only selectable identity. Add four-kind duplicate-location fixtures and verify list/lint output, not merely the count of problems.

**Resolution:** Confirmed location loss in graph duplicate attribution and
indistinguishable ordinary list rows. Sequenced a dedicated location-contract
task after this read migration and before domain path removal; it covers graph,
list/lint, human, wire, and schema evidence.

### Untested risks and deferred work

The existing projected-column contract fixtures use equal source and declared IDs, so a shared helper regression survives that test; the mutation ledger demonstrates this exactly. Add divergent IDs to each kind's column fixture and test display plus `--json -c id`. This is a **test gap**, not a third runtime finding. Readable task hashing twice is a measured code-path cost; no workload benchmark was run, so no performance severity is assigned. The later source/path, source-set, and TUI tasks own removal of domain `Path`/`FilenameID`/`SourceVersion`, optional local navigation, and independently composed capabilities; this review does not call their absence a current regression. Actual remote-adapter filesystem-call counts and delayed TUI local-action races remain untested because no such adapter or capability split is shipped here.

### Validation and isolation attestation

`GOCACHE=/private/tmp/taskflow-codex-go-cache go test ./...` passed at the captured baseline. Focused core/store/wire/TUI tests, the temporary contradictory probes, ordinary CLI partial-mode probes, `go vet ./...`, generated-doc/schema-comment comparisons, and `git diff --check` passed where applicable. Each mutation was restored before the next; all four temporary `review_probe_test.go` files were removed. The final guarded verification requires this audit to be the sandbox's only difference from its baseline. I did run read-only `git status` and an `AGENTS.md` search on the source before creating the sandbox, contrary to the brief's stricter first-two-operations rule; no source implementation inspection, tests, generators, or writes occurred there. All substantive review work used the independent sandbox.

- Workspace: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CurZdy`
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CurZdy/.git` (independent clone)
- Baseline commit: `fff26a80670901f9c053c639a6486a3afe341395`
- Captured source deliverable blob: `ae520f712e3bc33498fe0dcd2ea61b237add250a`
- Captured source fingerprint: `ab1579ab06195a8cda5d41acbc9e59b5c4f4c7c3`
- Deliverable: `planning/audits/6ge0q80cc01r-2026-09-26-ordinary-entity-read-snapshots-implementation-codex.md`
- Guarded transfer result: **succeeded** (verified and transferred with `scripts/isolated-review-workspace.sh` after final review).
