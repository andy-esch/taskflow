---
schema: 1
id: 6g9zev1epg76
bucket: open
area: arch-machine-contract
date: "2026-09-14"
updated_at: "2026-09-19"
---

# Weekly Architecture Audit: machine-contract — 2026-09-14

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `machine-contract` · ISO week
`2026-W38` (mod 4 = 2).
ADRs consulted: ADR-0003 (§3 identity, §4 flat id-led layout), ADR-0006
(Threads as task DAGs), ADR-0007 (planning state vocabularies), ADR-0001 (ADR
process) · Files read: `internal/wire/wire.go`, `internal/wire/envelopes.go`,
`internal/wire/schema.go`, `internal/wire/dto.go`, `internal/cli/schema.go`,
`internal/cli/exit.go`, `internal/cli/render/columns.go`,
`internal/wire/registry_test.go`, `internal/wire/wire_changelog_test.go`,
`internal/wire/schema_descriptions_test.go`, `internal/cli/golden_test.go`,
`docs/ARCHITECTURE.md` · Sources cited: 7 (4 fetched, 3 search-surfaced — see
the note under Best-practice comparison)

## Executive summary

This lens is in **good structural health and poor declarative health**. The
mechanics of the machine contract are among the strongest things in the
repository: one global `schema_version`, a hand-maintained 66-entry changelog
that a test proves contiguous and ascending, an envelope registry that a
source-parsing test proves complete in both directions, 37 byte-stable goldens,
and reflected JSON-Schema descriptions pinned to the live vocabularies. What is
missing is the **decision layer**. No ADR governs the `--json` contract at all,
so its one stated compatibility rule lives in a Go doc comment — and the
practice has diverged from that rule at least five times without anyone being
able to point at a decision that permits it (H1). Three smaller gaps follow the
same shape: things the contract publishes about itself are incomplete (M1, L1)
or unevenly applied across entities (M2, M3). Read H1 first; the rest are cheap
and independent.

## State of the architecture (this lens)

**The wire package is a genuine neutral leaf.** `internal/wire` imports only
`core` and `domain` (`.golangci.yml` depguard enforces it, per
`docs/ARCHITECTURE.md:62-63`), and every envelope is produced by a
`ToXEnvelope(...) XEnvelope` value constructor
(`internal/wire/envelopes.go:38, 80, 108, …`). `render`'s `*JSON` emit funcs
build that value and encode it; a future `serve` adapter can obtain the same
value and embed it in a response body. The split is real, not aspirational —
`wire.EncodeJSON` (`internal/wire/wire.go:283`) is the single encoder, so every
envelope is compact with one trailing newline.

**Version.** `SchemaVersion = "1.66"` (`internal/wire/wire.go:276`) is one
version for the whole CLI output schema, decided 2026-06-12. The 255-line
changelog above it is the de-facto compatibility record.
`TestSchemaVersionChangelogIsAscending`
(`internal/wire/wire_changelog_test.go:63`) parses that comment block out of the
package's own source and asserts the entries start at 1.1, stay contiguous, and
end exactly at the `SchemaVersion` constant — a genuinely unusual guard that
makes a prose changelog executable.

**Envelope registry.** `jsonEnvelopes` (`internal/wire/envelopes.go:1141-1195`)
registers 48 envelopes so one `Reflect` pulls them into a single `$defs`
document. `TestJSONEnvelopes_RegistryIsComplete`
(`internal/wire/registry_test.go:26`) closes the inverse gap by *parsing the
package source*: any `type <Name>Envelope struct` that is not a field of
`jsonEnvelopes` fails the test. Its doc comment names the bug that motivated it
(the research envelopes shipped, emitted `schema_version`, and appeared in no
schema, with every test green).

**Self-description.** `tskflwctl schema` runs anywhere — it overrides the root
`resolve()` with `app.styleOnlyPreRun` (`internal/cli/schema.go:52`) so an agent
can learn the contract without a planning repo. `runSchemaContract`
(`internal/cli/schema.go:83-130`) assembles every value from its real source:
`domain.AllStatuses()`, `domain.AllEpicStatuses()`,
`domain.AllThreadStatuses()`, `domain.AllAuditBuckets()`,
`domain.FindingStatuses()`, `domain.CriterionSuffixStates()`,
`domain.KnownTaskFieldNames()`, and the CLI's own `errCodes` table. Nothing is
transcribed. ADR-0007's Consequences name this as the structural fix for the
`landed` drift.

**Errors.** `errCodes` (`internal/cli/exit.go:22-31`) maps four domain classes
to exit codes 10/11/13/14 and to the stable machine name that becomes the
`code` field of the error envelope. `ExitCode` and `errorCodeName` both read
that one table, so a code and its name cannot drift; `schema.go` iterates the
same table for the published `exit_codes`. `WriteError`
(`internal/cli/exit.go:66`) emits `ErrorEnvelope` under `--json` and prose
otherwise, always on stderr, and enriches the envelope with typed recovery
receipts (`dependency_mutation`, `graph_repair`, `task_lifecycle`,
`task_rename`, `thread_*`, `filesystem`) via `errors.As`.

**Projections.** `--json -c` is a deliberately *separate* surface from the
typed envelopes: `ProjectedListJSON` (`internal/cli/render/columns.go:250`)
emits `schema_version` + the same list key + string-valued rows in `-c` order,
and its doc comment states plainly that it "does NOT validate against
`schema --json-schema`". Schema 1.66 made canonical wire selectors
(`updated_at`, `open_findings`) available while keeping the old `updated`/`open`
spellings as aliases with their own keys — handled additively and covered by
focused tests (`internal/cli/pipeline_test.go:128-239`).

**Which ADR governs this?** None. ADR-0003 governs storage identity and layout;
ADR-0006 governs Threads; ADR-0007 governs body-level vocabularies and names
the wire versions that carry them (1.45–1.48) without deciding wire policy.
ADR-0004 and ADR-0005 are `proposed`. The `--json` contract — arguably this
tool's most-consumed public surface — is governed by a Go doc comment and
`docs/ARCHITECTURE.md` prose. That absence is H1.

## ADR reconciliation

### ADR-0003 §3 — a 12-char time-sortable id is the entity's stable identity

**Quoted:** "Stable identity that survives re-titles, moves, and a future
backend swap (it *is* the identity)" (ADR-0003 Consequences, §187-192), deciding
"Identity — a 12-char time-sortable id" (§101).

**Implementation:** follows, in the typed envelopes. `TaskJSON`, `AuditJSON`,
and `TaskInfoJSON` all carry `id` (schema 1.24), and schema 1.32 corrected
`created.id` to carry the stable id with `created.slug` added alongside —
`internal/wire/wire.go:124-128` calls the previous behaviour "a defect, not a
deliberate contract … the one command where an agent captured the wrong handle,
on exactly the command where capturing the durable one matters most."

**Divergence:** the *cheap* machine path does not follow. `TaskColumns()`
(`internal/cli/render/columns.go:296-315`) and `AuditColumns()` (`:374-385`)
declare no `id` column, so `task list --json -c id` and
`audit list --json -c id` both fail with exit 11 while
`research list --json -c id` succeeds (`ResearchColumns()`, `:369`). See M2.

**Class if divergent:** unanchored — ADR-0003 decides storage identity, not the
wire projection surface, and no ADR governs the latter (see H1). The ADR is
nonetheless the reason the id matters.

### ADR-0007 §2 / §4 — closed vocabularies are spelled once and published

**Quoted:** "The vocabularies are now derived into `schema`/`schema audit`
rather than transcribed, so the M3 failure cannot recur in the same shape."

**Implementation:** follows. `finding_statuses` and `criterion_states` are read
from `domain.FindingStatuses()` / `domain.CriterionSuffixStates()` at
`internal/cli/schema.go:116-117`, and `schema_descriptions_test.go:47` pins the
one description that cannot be derived. Verified live: `schema --json` emits all
seven finding statuses and all four criterion states.

### ADR-0007 Consequences — "Schema 1.45–1.48 carry the wire half. 1.47 is **not additive**"

**Quoted:** "1.47 is **not additive** — `landed` is no longer accepted — though
no audit ever used it."

**Implementation:** drift against the contract's own stated versioning rule.
`internal/wire/wire.go:23` states "Adding a field bumps the minor;
renaming/removing bumps the major." 1.47 removed a vocabulary value and bumped
the minor. ADR-0007 records the *change* honestly and records no decision about
the *versioning*, so the exception has no stated scope. See H1.

**Class if divergent:** ADR gap.

### ADR-0006 — Threads as first-class initiative views

**Quoted:** ADR-0006 makes Threads a first-class planning entity; schema
1.54–1.65 carry twelve wire releases of Thread contract.

**Implementation:** follows for the typed envelopes (`thread_list`,
`thread_show`, `thread_frontier`, `thread_graph`, `thread_plan`, plus mutation
and apply receipts, all registered and goldened). Silent on the projection
surface: `thread list` has neither `-o` nor `-c` while the other five planning
entities have both. See M3.

**Class if divergent:** ADR gap — ADR-0006 does not say a Thread must be
triage-projectable, and nothing else does either.

### No ADR governs the machine contract

Stated explicitly, per the routine: the `--json` envelope policy (one global
version; every envelope carries `schema_version`; mutation receipts carry
`workspace`; projections are a view, not the contract; the exit-code taxonomy)
is an architectural decision set that was made — twice datable, 2026-06-12 and
the 1.47 exception — and never recorded. This is the lens's headline finding.

## Best-practice comparison

> **Research note (material to this audit's depth).** This environment's network
> egress policy blocked `WebFetch` for most hosts; only `json-schema.org` and
> `raw.githubusercontent.com` resolved. Sources 1–4 below were fetched and
> quoted directly. Sources 5–7 are cited from WebSearch result summaries because
> their hosts were blocked (`anthropic.com`, `infoq.com`, `doc.rust-lang.org`);
> they are treated as corroborating, never as the sole anchor for any finding.

### 1. A minor version promises backward compatibility

**Source:** <https://raw.githubusercontent.com/semver/semver/master/semver.md>
**Guidance:** "Minor version Y (x.Y.z | x > 0) MUST be incremented if new,
backward compatible functionality is introduced to the public API." "Major
version X (X.y.z | X > 0) MUST be incremented if any backward incompatible
changes are introduced to the public API."
**This codebase:** diverges. `SchemaVersion` is shaped as semver and
`wire.go:23` restates the rule in the project's own words, but at least five
releases shipped backward-incompatible changes as minor bumps (1.25, 1.26, 1.32,
1.47, 1.59 — enumerated in H1).
**Justified?** The *practice* is defensible; the *silence* is not. A single-user,
local-first tool whose only consumers are agents re-reading a freshly published
`schema` has little to gain from a 2.0 and much to lose in changelog legibility.
But the number is still shaped like semver, so a consumer pinning `>=1 <2` — the
only thing a semver-shaped number invites — is given a promise the project does
not keep. Recording the real policy costs nothing and removes the trap.

### 2. Machine-readable output on every data-returning command; exit codes mapped to failure modes

**Source:**
<https://raw.githubusercontent.com/cli-guidelines/cli-guidelines/main/content/_index.md>
(clig.dev)
**Guidance:** "Display output as formatted JSON if `--json` is passed." "Have
machine-readable output where it does not impact usability." "Return zero exit
code on success, non-zero on failure" — and map non-zero codes to important
failure modes.
**This codebase:** follows on coverage, partial on the taxonomy.
`--json` is a global flag; `internal/cli/output_coverage_test.go` exercises the
`--json` variants of show/transition/schema; stdout stays empty on failure and
the error envelope goes to stderr (`internal/cli/exit.go:63-65`). But the
*published* taxonomy is incomplete: `schema --json`'s `exit_codes` lists only
`{10 not-found, 11 validation, 13 ambiguous, 14 conflict}`, omitting 0, the
catch-all 1, and 130. See M1.
**Justified?** No. The omission is an oversight of the same kind the repo
already fixed elsewhere (`finding_statuses` at 1.21, `criterion_states` at 1.48
were both added precisely so an agent need not "trigger an error and parse
prose to learn the set").

### 3. Version the schema through `$id`, not through prose

**Source:** <https://json-schema.org/draft/2020-12/release-notes> and
<https://json-schema.org/understanding-json-schema/structuring>
**Guidance:** `$defs` is "a standardized place to keep subschemas intended for
reuse in the current schema document"; `$ref` "contains a URI-reference that is
resolved against the schema's Base URI"; `$id` in Draft 2020-12 identifies a
schema resource, and bundled documents keep their references stable.
**This codebase:** follows on structure, diverges on identity.
`wire.JSONSchema()` (`internal/wire/envelopes.go:1205`) emits a correct
2020-12 document: `$schema`, one `$ref` to `#/$defs/jsonEnvelopes`, and 147
`$defs`. But `$id` is
`https://github.com/andy-esch/taskflow/internal/wire/json-envelopes` — a
constant URI whose content changes on every release. The version appears only
inside the human `title` string ("tskflwctl --json output (schema_version
1.66)"). See L1.
**Justified?** Partly: the schema is generated on demand by a local binary, not
served, so nobody dereferences the `$id`. It still means a consumer that caches
or diffs schemas by `$id` cannot tell 1.30 from 1.66, and a future 2.0 would
silently replace content at the same identifier.

### 4. Remove an enum value and you break every consumer that switches on it

**Source:** <https://doc.rust-lang.org/cargo/reference/semver.html> *(cited from
search summary — host blocked by this environment's egress policy)*
**Guidance:** renaming or removing an enumeration member is a breaking change;
the hazard is specifically a consumer whose `switch` covers the whole member
list and errors in the default case.
**This codebase:** diverges at 1.47 (`landed` removed from `finding_statuses`)
and 1.26 (`misfiled` / `declared_status` fields retired), both minor bumps.
**Justified?** On blast radius, yes — `internal/wire/wire.go:193` records that
"no audit in the corpus ever used" `landed`. On contract, no: a consumer cannot
audit the corpus, only the version number.

### 5. Publish the handle a caller should store, cheaply

**Source:** <https://www.anthropic.com/engineering/writing-tools-for-agents>
*(cited from search summary — host blocked)*; corroborated by
<https://www.infoq.com/articles/ai-agent-cli/> *(blocked)*
**Guidance:** tool responses should return the identifier a caller will need
later rather than forcing a second round-trip; implement filtering/projection so
an agent is not overwhelmed by data; error responses should communicate
"specific and actionable improvements, rather than opaque error codes."
**This codebase:** follows on filtering (`-c`/`-o`/`-q` are exactly this) and on
error actionability (`FilesystemErrorJSON` carries class/operation/path and an
explicit `retryable` flag; `internal/wire/envelopes.go:1062-1070`). Partial on
the handle: the cheap path can project `slug` but not `id` for tasks and audits,
so an agent that wants durable references must either pay for full `--json` or
issue one `task info` per row. See M2.

## Tensions and trade-offs

- **One global version vs. per-envelope versions.** The single `schema_version`
  was decided 2026-06-12 and is the right call for a tool shipped as one binary:
  it gives a consumer one number to reason about instead of 48. The cost is that
  every envelope's compatibility is only as strong as the weakest change in the
  release, which is exactly what makes H1 bite — a Thread-only reshape (1.59)
  and a vocabulary removal (1.47) both spend the same minor bump as an additive
  field.
- **Curated columns vs. DTO parity.** The Thread `harden-cli-machine-contracts`
  records the decision that "a column registry need not mirror every full DTO
  field," and it is right — reflecting every field into `-c` would make the
  triage surface useless. M2 does not argue against that decision; it argues
  that `id` is the one field the contract itself singles out as the thing to
  store, and so is the one exception the curation should make.
- **Agent-first vs. human-first output.** `--json -c` being a string-valued view
  rather than a typed envelope is a genuine simplification that a networked API
  could not make. It is documented in three places
  (`columns.go:243-249`, `schema.go:28-31`, CLAUDE.md) and tested. No finding
  here; it is correctly scoped and correctly disclaimed.
- **`schema:` (on-disk) vs. `schema_version` (wire).** Two version handles with
  similar names for genuinely different things. `6fkkz41cax80` Q10 draws the
  line explicitly — this audit's H1 is about the half that task parks.

## Findings

#### H1. The `--json` contract's stated versioning rule and its practiced rule have diverged, and neither is recorded as a decision  · **Status:** fixed 2026-09-19

**File:** `internal/wire/wire.go:23` (rule) · `internal/wire/wire.go:94-98, 124-128, 192-194, 251-253` (the exceptions) | **Component:** wire / machine contract
**Effort:** S · **Urgency:** soon
**Class:** ADR gap
**Anchored to:** SemVer clauses 7–8
(<https://raw.githubusercontent.com/semver/semver/master/semver.md>);
ADR-0007 Consequences ("1.47 is **not additive**"); ADR-0001 (a decision of this
weight is recorded, not commented)

The contract states its own compatibility rule in one sentence:

> "Adding a field bumps the minor; renaming/removing bumps the major."
> — `internal/wire/wire.go:23`

At least five releases contradict it, each shipped as a minor bump:

| Release | Change | Why it is not additive |
| --- | --- | --- |
| 1.25 | `misfiled` / `declared_status` "inverted meaning" | same field names, opposite predicate |
| 1.26 | "retired the task `misfiled`/`declared_status` fields and the `status` summary `misfiled` count" | field removal |
| 1.32 | "`created.id` now carries the STABLE id" (was the slug) | same field, different referent; the entry itself warns "CONSUMERS: anything that stored `created.id` as a reference held a slug and must re-read" |
| 1.47 | "`landed` is no longer accepted" | enum value removal; the entry says "NOT additive" |
| 1.59 | unreadable-diagnostic shape "replace[d]" `{path,message}` | shape replacement |

Every one of these is documented with unusual honesty in the changelog — that is
not the defect. The defect is that the **number** carries the compatibility
promise for any consumer that has not read 255 lines of Go comment, and the
number says "backward compatible" five times when the prose beside it says
otherwise. SemVer clause 8: "Major version X MUST be incremented if any backward
incompatible changes are introduced to the public API."

The project's actual policy appears to be *never bump major; record the break
loudly in the changelog*, and for this project that is probably correct — the
only consumers are agents that re-read `schema` on each run, and five 2.0-through-6.0
bumps would have destroyed the changelog's legibility for no reader's benefit.
But that policy has never been written down, so: the next vocabulary removal has
no precedent to follow, a reviewer cannot tell an intended exception from a
mistake, and the 2.0 that eventually honours the rule will be
indistinguishable in kind from 1.47.

Note this is **not** covered by `6fkkz41cax80` (`adr-close-frontmatter-schema-policy-questions`),
whose Q10 explicitly scopes itself to the on-disk handle: "Relation to the
envelope `schema_version`: distinct handles — `schema:` versions the **on-disk
file shape**, `schema_version` versions the **`--json` output**."

**Compounding cost:** `wire.go`'s changelog grows roughly one entry per feature
(66 in ~15 months, five in the last month alone: 1.62–1.66). Each release
spent under an unstated policy is another precedent that a later ADR must either
ratify retroactively or declare a mistake. The two adapters that will consume
this contract without a human in the loop — `tskflwctl serve` (epic 19) and any
external agent integration — are exactly the consumers that would apply
semver range semantics mechanically. Deciding now costs one ADR; deciding after
a web adapter ships means deciding it against installed consumers.

**Recommendation:** Write an ADR recording the `--json` contract policy that is
actually practiced, and amend `wire.go:21-25` to match it. Minimum content: (a)
what `schema_version` versions and what it does not (explicitly: not the
projected `-c` view, not the on-disk `schema:` key); (b) the real increment
rule, including whether major is ever bumped and, if not, what marks a
non-additive release instead (a `NOT ADDITIVE` changelog convention already
exists in practice at 1.47 — make it a required marker and consider publishing
it as a field); (c) the stability promise for `exit_codes`, the published
vocabularies, and envelope key names; (d) what a consumer is entitled to assume
from the number alone. No code change is required to make the contract honest —
only for the parts of (b) that publish the marker.

**Proposed ADR amendment:** a new ADR (next number after 0007), *"The `--json`
machine contract: one version, additive by default, breaks recorded not
promoted"*, recording: `schema_version` is a monotonic contract *revision*
counter shaped as `major.minor`, not a semver compatibility range; the major
component is reserved and has never been incremented; a release that is not
backward compatible MUST carry an explicit non-additive marker in the changelog
entry, and consumers are directed to `schema --json` on every run rather than to
version-range pinning. Human decision — nothing was edited.

**Follow-up:** if the decision goes the other way (honour semver and ship a
2.0), that is a much larger change with consumer-migration consequences and
belongs in its own task, not in this ADR.

**Resolution:** ADR-0008 now defines schema_version as one monotonic all-JSON
revision rather than SemVer. Revision 1.68 publishes the policy, requires
ADDITIVE or NOT ADDITIVE changelog classifications from that boundary, and tests
the declaration against the executable constants.

#### M1. The published `exit_codes` contract omits every code an agent is most likely to actually receive  · **Status:** open

**File:** `internal/cli/exit.go:22-31`, `internal/cli/schema.go:100-103` | **Component:** cli / machine contract
**Effort:** XS · **Urgency:** soon
**Class:** unanchored
**Anchored to:** clig.dev, "Return zero exit code on success, non-zero on
failure" + map non-zero codes to important failure modes
(<https://raw.githubusercontent.com/cli-guidelines/cli-guidelines/main/content/_index.md>);
the repo's own stated principle at `internal/cli/exit.go:63-65` ("an agent
driving --json must never have to parse prose to learn why a command failed")

`tskflwctl schema --json` publishes exactly four exit codes:

```
[{"code":10,"name":"not-found"},{"code":11,"name":"validation"},
 {"code":13,"name":"ambiguous"},{"code":14,"name":"conflict"}]
```

Three codes the CLI can return are absent:

- **0** — success, and per `ExitCode`'s own comment "also covers idempotent
  no-ops" (`exit.go:34`). An agent building a dispatch table from `exit_codes`
  has no entry for the case that happens most.
- **1** — the fallback for every error `domain.Classify` does not classify
  (`exit.go:48`). Its wire code name is the bare string `"error"`
  (`exit.go:59`), which appears nowhere in the published contract. This is not a
  rare path: `FilesystemErrorJSON`'s own doc comment states "Exit codes are
  deliberately unchanged (these stay exit 1)"
  (`internal/wire/envelopes.go:1060-1061`), so *every* permission-denied,
  not-found-directory, read-only and disk-full failure lands here — as does
  every Cobra flag-parse error. Reproduced live:
  `tskflwctl thread list --json -c slug` →
  `{"schema_version":"1.66","error":{"code":"error","message":"unknown shorthand flag: 'c' in -c"}}`.
- **130** — `128 + SIGINT`, returned for `prompt.ErrAborted` (`exit.go:40`).
  Only reachable on a TTY, so the weakest of the three, but it is a code the
  tool returns and does not publish.

The repo has twice fixed this exact shape of gap for other vocabularies —
`finding_statuses` (1.21) and `criterion_states` (1.48) were both added so an
agent would not have "to trigger an error and parse prose to learn the set"
(`wire.go:79-81`, `:196-199`). The exit-code table is the same kind of closed
vocabulary and did not get the same treatment.

**Compounding cost:** small but real and asymmetric. `exit.go:19-20` declares
that the table's "order and names are part of the wire golden and must not
change," so this is append-only — and each release that ships without the
missing entries is another window in which an integration hard-codes the
four-code world. More importantly, the same table is the natural source for the
`serve` adapter's HTTP status mapping (epic 19): a status map derived from a
taxonomy with no "unclassified" member has to invent one, and it will not match.

**Recommendation:** Append the missing rows to the published contract without
disturbing the existing four. Because `errCodes` is also the `Class → code`
lookup, the two roles want separating rather than overloading: keep `errCodes`
as-is for classification and give `runSchemaContract` a small published table
that includes `{0, "ok"}`, `{1, "error"}` and `{130, "aborted"}` alongside it —
or add a `class` field so unclassified rows can carry an empty class. Either is
a minor bump (additive field / additive rows). Consider also publishing `12` as
`{"code":12,"name":"invalid-transition","status":"retired"}`, which CLAUDE.md
and `domain/errors.go` both document as reserved and the contract does not
mention at all.

#### M2. `--json -c` cannot project `id` for tasks or audits, so the cheap machine path cannot return the durable handle  · **Status:** fixed

**File:** `internal/cli/render/columns.go:296-315` (tasks), `:374-385` (audits) | **Component:** cli/render — column registry
**Effort:** XS · **Urgency:** soon
**Class:** unanchored
**Anchored to:** ADR-0003 §3 (the 12-char id "*is* the identity"); schema 1.32's
own rationale (`internal/wire/wire.go:124-128`); stable-identifier practice
(<https://www.anthropic.com/engineering/writing-tools-for-agents>, cited from
search summary — host blocked)

Reproduced against this repo's own planning tree:

```
$ tskflwctl task list --json -c id,slug
{"schema_version":"1.66","error":{"code":"validation","message":"validation failed:
 unknown column \"id\" (available: slug, status, tier, priority, epic, updated_at,
 description, revisit_at)"}}

$ tskflwctl audit list --json -c id,slug      # same error
$ tskflwctl research list --json -c id,slug   # works — ResearchColumns has `id`
```

The contract is emphatic that the id is the handle to store. `CreatedItem.ID`'s
doc comment: "This is the handle to store: it survives renames"
(`envelopes.go:692-694`). `CreatedItem.Slug`: "it changes when the entity is
renamed, so it is for display and for typing at a prompt, **never for saving as
a reference**" (`:697-698`). Schema 1.32 exists solely because `created.id` once
carried the slug, which the changelog calls "a defect … on exactly the command
where capturing the durable one matters most."

Yet the path CLAUDE.md names as "the cheap machine path"
(`task list --json -c slug,status,description`) can only project the mutable
handle. The two workarounds are both bad: full `--json` (every frontmatter
field for every task — the cost `-c` exists to avoid), or one `task info` per
row (N invocations; `TaskInfoJSON` does carry `id`, `dto.go:78`).

This is *not* an argument for column/DTO parity. The Thread
`harden-cli-machine-contracts` records the opposite and is right: "a column
registry need not mirror every full DTO field." `id` is the single field the
contract itself designates as the thing a caller must store, which makes it the
one principled exception.

**Compounding cost:** `task rename` shipped at schema 1.62/1.63, so slugs are
now genuinely mutable by a supported verb rather than only by hand-editing.
Every agent workflow, script, and planning-body cross-link written against
`-c slug` in the meantime holds a handle that a rename invalidates — and this
repo self-hosts its planning, so its own corpus accumulates them. Adding the
column later does not retroactively fix references already stored. Cost to fix
now: two `column(...)` lines plus a golden regen.

**Recommendation:** Add an `id` column to `TaskColumns()` and `AuditColumns()`,
appended **last** in both — the file's existing convention for additive columns
(see the `revisit_at` and `percent`/`deprecated` comments at `:311-313`,
`:327-329`) keeps the default `-o table`/`csv` layout byte-stable, so this is a
pure minor bump. `EpicColumns()` already leads with `id` and needs nothing.
Coordinate with `enforce-projected-column-registry-invariants`, which will want
the new selectors covered by its collision/parity fixtures.

**Resolution:** Task and audit list registries now expose stable IDs as
trailing, explicitly selectable columns. Schema 1.67, command-level projections,
completion, machine-text goldens, and the shared full-wire fidelity harness pin
the behavior without changing slug-first quiet output.

#### M3. `thread list` has neither `-o` nor `-c`, so the newest first-class entity is absent from the triage contract  · **Status:** open

**File:** `internal/cli/thread.go` (`thread list` flag set) | **Component:** cli — Thread commands
**Effort:** S · **Urgency:** eventually
**Class:** ADR gap
**Anchored to:** ADR-0006 (Threads as a first-class planning entity); clig.dev,
"Have machine-readable output where it does not impact usability"

Surveyed across the planning entities:

| Command | `-c/--columns` | `-o/--output` |
| --- | --- | --- |
| `task list` | yes | yes |
| `epic list` | yes | yes |
| `audit list` | yes | yes |
| `audit findings` | yes | yes |
| `research list` | yes | yes |
| **`thread list`** | **no** | **no** |

`thread list` offers only `--status` and the global `--json`. Threads are a
first-class entity by ADR-0006 with twelve schema releases of wire contract
behind them (1.54–1.65), and `ThreadsEnvelope` is registered and goldened — so
the *typed* contract is complete. What is missing is the projection surface that
CLAUDE.md's triage guidance assumes ("`--json` is compact … and also takes `-c`
to project just the fields you need — the cheap machine path"). For Threads that
sentence is false, and an agent triaging Threads must take the full envelope,
which for Threads is unusually heavy (per-Thread membership, nominal *and*
sound rollups, external gates, graph and projection health, completed-
inconsistency codes).

`space list`, `template list` and `theme list` also lack projection, but those
are tool-metadata lists rather than planning entities; excluding them looks
principled. Excluding Threads does not.

**Compounding cost:** low and linear rather than compounding — nothing gets
harder if this waits. The reason to record it is consistency drift: each new
first-class entity (epic 28 names routines, projects and ADRs as candidates)
that ships without a column registry makes "every planning entity is
`-c`-projectable" less true, until the triage guidance has to be rewritten
around the exceptions instead of the rule. Hence *eventually*, not *soon*.

**Context:** severity Medium (a documented contract promise is false for one
entity) exceeds urgency (nothing degrades while it waits, and the full envelope
is a correct if expensive substitute).

**Recommendation:** Add a `ThreadColumns()` registry and wire `-o`/`-c` into
`thread list`, following `ResearchColumns()` as the template. A reasonable
starting set: `slug`, `status`, `done`, `total`, `drained`, `frontier`,
`description`, `id`. Use `contractColumn` for any column whose canonical wire
key differs from its display header, per the 1.66 convention.

#### L1. The published JSON Schema's `$id` carries no version  · **Status:** fixed 2026-09-19

**File:** `internal/wire/envelopes.go:1205-1219` | **Component:** wire — JSON Schema generation
**Effort:** XS · **Urgency:** eventually
**Class:** unanchored
**Anchored to:** <https://json-schema.org/draft/2020-12/release-notes> and
<https://json-schema.org/understanding-json-schema/structuring>

`schema --json-schema` emits a well-formed 2020-12 document, but its identity is
version-free:

```json
"$id": "https://github.com/andy-esch/taskflow/internal/wire/json-envelopes"
```

The version appears only inside the human-facing `title`
("tskflwctl --json output (schema_version 1.66)"), which is prose, not a
machine field. A consumer that caches, diffs, or keys schemas by `$id` — the
purpose `$id` serves in 2020-12, where it identifies a schema resource — sees one
identifier whose content silently changes on every release.

Kept as a Low rather than dropped because it is the same defect as H1 viewed
from the other end: the contract's version is reliably present in every
*payload* (`schema_version` on all 48 envelopes) and absent from the *schema
document* that describes them.

**Compounding cost:** none today — the schema is generated on demand by a local
binary, so nothing dereferences the `$id`. It becomes real the moment `serve`
(epic 19) publishes the schema at a URL, at which point a version-free `$id`
means served schemas cannot be cached or content-addressed correctly.

**Recommendation:** Embed the version in `$id` (for example
`…/json-envelopes/1.66` or `…/json-envelopes?v=1.66`) and/or add a
`schema_version` annotation at the schema root. Both are additive to the schema
document. Fold into whatever H1's ADR decides about how the version is
published, rather than fixing separately.

**Resolution:** The generated Draft 2020-12 schema now has a revision-qualified
$id and root annotations for revision, scheme, and compatibility; focused tests
and machine-contract goldens pin them.

## What audited clean

- **The changelog is executable.** `TestSchemaVersionChangelogIsAscending`
  (`wire_changelog_test.go:63`) parses `wire.go`'s own comment block and enforces
  contiguity, ascent, and agreement with the `SchemaVersion` constant — so
  concurrent branches must append rather than interleave. Satisfies
  ADR-0007's "derived rather than transcribed" principle for the version record.
- **Envelope registration is complete in both directions.**
  `TestJSONEnvelopes_RegistryIsComplete` (`registry_test.go:26`) parses the
  package source for `type *Envelope struct` and fails any that is not a
  `jsonEnvelopes` field, closing a gap reflection cannot reach — and its comment
  names the real bug (the research envelopes) that motivated it.
- **Every published vocabulary is read from its owner, never transcribed.**
  `runSchemaContract` (`cli/schema.go:83-130`) sources all six enums from
  `domain`, directly implementing ADR-0007 §2's "spelled once".
- **Exit code and machine name cannot drift.** One `errCodes` table backs
  `ExitCode`, `errorCodeName`, and the published `exit_codes`
  (`cli/exit.go:22-31`) — the shape of the contract is right even though its
  membership is incomplete (M1).
- **Errors are structured, not prose.** `ErrorEnvelope` carries a stable `code`
  plus typed recovery receipts attached by `errors.As` for seven distinct
  failure families, and `FilesystemErrorJSON` classifies OS failures with an
  explicit `retryable` flag rather than making an agent read English
  (`envelopes.go:1054-1070`). Matches the cited practice on actionable,
  self-correcting error responses.
- **The projected view is disclaimed where a consumer will look.**
  `ProjectedListJSON`'s comment (`columns.go:243-249`), `schema`'s `Long` text
  (`cli/schema.go:28-31`), and CLAUDE.md all say the same thing: `--json -c` is
  a string-valued view and only bare `--json` validates against the schema. Three
  places, one message, no drift.
- **1.66 shows the right way to make a breaking change additive.** Canonical
  wire selectors (`updated_at`, `open_findings`) were introduced while the old
  `updated`/`open` spellings kept working *and* kept their own output keys
  (`columns.go:104-128`), with focused tests on both paths
  (`pipeline_test.go:128-239`). This is the pattern H1 asks the project to make
  its stated default.
- **Reflected schema descriptions are pinned to the live vocabularies.**
  `schema_descriptions_test.go` guards both that per-field descriptions exist at
  all and that the one hand-written enum description tracks
  `domain.AllEpicStatuses()`.
- **Nothing presentational leaks into the contract.** `wire` imports only
  `core` and `domain`, enforced by depguard rather than by memory
  (`docs/ARCHITECTURE.md:62-63`); `ToXEnvelope` returns values so a web adapter
  gets the same payload the CLI encodes.

## Candidate tasks (human to triage)

- `tskflwctl task new "ADR: record the --json machine-contract versioning policy" --epic 21-code-quality-architecture-hardening --tags adr,json,contract --tier 2 --priority high --description "Record what schema_version versions, the real increment rule, and the stability promise for exit codes and published vocabularies; amend wire.go's stated rule to match."` — finding H1
- `tskflwctl task new "Publish the complete exit-code taxonomy in the schema contract" --epic 20-cli-ux-and-ergonomics --tags cli,json,contract --tier 3 --priority medium --description "Add 0/1/130 (and retired 12) to schema --json exit_codes without disturbing the existing four rows or the wire golden."` — finding M1
- `tskflwctl task new "Project the stable id from task and audit list columns" --epic 20-cli-ux-and-ergonomics --tags cli,json,contract --tier 3 --priority medium --description "Append an id column to TaskColumns and AuditColumns so --json -c can return the durable handle without a full envelope."` — finding M2
- `tskflwctl task new "Give thread list the -o/-c projection surface" --epic 30-threads-and-task-dependency-graphs --tags cli,threads,json --tier 4 --priority low --description "Add ThreadColumns and wire -o/-c into thread list so Threads join the cheap triage contract with the other planning entities."` — finding M3
- ⚠️ H1 partially adjacent to `planning/tasks/6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md` — its Q10 explicitly parks the wire handle as out of scope
- ⚠️ M2 partially adjacent to `planning/tasks/6g9txf8x9m70-enforce-projected-column-registry-invariants.md` — an invariant pass that explicitly declines to add columns

## Related-task observations (propose-only)

- Scope boundary worth confirming: `6fkkz41cax80` (`adr-close-frontmatter-schema-policy-questions`, next-up, epic 26) draws a clean line at Q10 between the on-disk `schema:` key and the wire `schema_version`. H1 lives on the far side of that line and needs its own ADR; if the human would rather widen `6fkkz41cax80` to cover both handles, that is a reasonable alternative to a new task — but the line is currently drawn deliberately, so widening it should be explicit.
- `6g9txf8x9m70` (`enforce-projected-column-registry-invariants`, ready-to-start, epic 20, tier 2/high) states "do not require every DTO field to become a column." M2 agrees with that principle and asks for one named exception. If M2 lands as its own task it should be sequenced **before** the invariant pass so the new selectors are covered by its fixtures, or folded into it explicitly — sequencing is the human's call.
- The Thread `harden-cli-machine-contracts` (in-progress, 1/2 done) is the obvious home for M1 and M2 if the human wants them tracked as Thread members rather than loose tasks. This audit does not mutate Thread membership.
