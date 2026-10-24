---
schema: 1
id: 6gjzzvdy5jnw
bucket: closed
area: arch-machine-contract
date: "2026-10-12"
updated_at: "2026-10-23"
---

# Audit: arch-machine-contract — 2026-10-12

> Create findings with `audit finding new`; update them with `audit finding`.

## Findings

<!-- Example grammar; let `audit finding new` allocate and render real findings: -->

```
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

<what's wrong, why it matters, evidence>

**Recommendation:** <minimum fix>

**Resolution:** <how it was resolved — written by `audit finding --note`, not by hand>
```

#### H1. The generated JSON Schema is silent on the closed vocabularies that most often force a NOT ADDITIVE revision · **Status:** tracked by 6gppr2k482rm

**File:** internal/wire/envelopes.go:1291 | **Component:** wire — JSON Schema generation
**Effort:** M · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** ADR-0008 §3 (closed vocabulary → `NOT ADDITIVE`), ADR-0008 §4
(the generated schema is the validation artifact), ADR-0007 §4 (tolerant reads);
`https://json-schema.org/blog/posts/custom-annotations-will-continue`

ADR-0008 §3 makes closed-vocabulary membership the primary trigger for an
incompatible revision: a revision is `NOT ADDITIVE` when "a published closed
vocabulary adds, removes, or renames a member that an exhaustive consumer may
switch on." The corpus bears that out — 1.47 dropped `landed` from the finding
statuses and 1.70 extended the exit-code taxonomy, and both are classified
incompatible for exactly this reason.

ADR-0008 §4 designates the generated Draft 2020-12 document as the validation
artifact, and the 2026-09-19 amendment took real trouble to make it fail closed
on revision identity: `constrainEnvelopeSchemaVersions`
(`internal/wire/envelopes.go:1319`) pins every registered envelope's
`schema_version` to a `const`, so pairing a cached schema with a payload from
another revision is rejected.

That artifact is silent on the axis the ADR says drives incompatibility. The
published document carries 154 definitions and exactly two enum-constrained
properties — `SchemaCommand.safety` and `SchemaExitCode.state`
(`internal/wire/schema.go:36,46`), both two-valued literals on the
schema-contract DTOs themselves. None of the six vocabularies the binary
publishes under `schema --json` (`statuses`, `epic_statuses`,
`thread_statuses`, `audit_buckets`, `finding_statuses`, `criterion_states`) is
expressed in the schema at all. A payload carrying the retired `landed`, or
`"status": "banana"`, validates clean against 1.87.

So a consumer that did exactly what §6 asks — validate typed envelopes against
the exact-revision schema — would have seen **no validation difference
whatsoever** across the 1.46 → 1.47 boundary, the one revision whose
`NOT ADDITIVE` label exists to warn it.

What makes this an ADR gap rather than plain drift is that the obvious fix is
wrong. Reads are deliberately tolerant so `lint` can report malformed data
already on disk (ADR-0007 §4), and the wire honours that. Verified against a
scratch planning repo built from this checkout: a task whose frontmatter was
hand-edited to `status: banana` is emitted verbatim by `task list --all --json`
as `"status":"banana"`, and `lint --json` reports it as `frontmatter status
missing or unrecognized`. An `enum` assertion on `TaskJSON.status` would make
that legitimate diagnostic payload fail validation. The schema therefore cannot
simply enumerate — and ADR-0008 never says what it should do instead.

The gap is exactly the distinction the tolerant-reader literature draws.
ADR-0008 publishes `reader_expectation: ignore-unknown-object-fields`, which is
the *field* half of the pattern. The arc42 write-up of Fowler's tolerant reader
(surfaced via search at `https://quality.arc42.org/approaches/tolerant-reader`;
the page did not resolve from this sandbox, so this is secondary attribution)
holds that tolerance applies to structure the consumer never reads, not to
values it does read, and that an unknown member of a vocabulary the consumer
switches on should be surfaced rather than silently defaulted. taskflow
publishes no statement of that value half, so a consumer cannot tell which
string fields are closed-vocabulary-with-tolerant-echo and which are free text.

An `x-` annotation is the mechanism the spec already blesses, and this document
already uses the convention at its root (`x-taskflow-revision-compatibility`,
`internal/wire/envelopes.go:1311`). Per the JSON Schema blog
(`https://json-schema.org/blog/posts/custom-annotations-will-continue`,
2023-02-22): "going forward, JSON Schema will treat any keyword that starts
with `x-` as an annotation", whose values are "returned to the user or
application without any processing by the schema", and "there can be no
collision with keywords defined by vocabularies, now or in the future." An
annotation asserts nothing, so tolerant echo survives, while a strict consumer
gains a machine-readable list of the fields it must check against the published
registry.

**Compounding cost:** every entity kind that arrives with a lifecycle brings a
vocabulary — threads brought `thread_statuses` at 1.54 (research deliberately
brought none) — and epic 28 `first-class-entities-new-planning-nouns` is active
at 63% with routines, projects and ADRs still to land. Annotating six registries
across today's 154 definitions is one bounded additive pass. Doing it after
those nouns ship means re-deriving annotations over a wider DTO surface and
regenerating all 41 machine-contract goldens behind a revision bump that must
itself be classified. Left alone, the cost never stops being paid: every
validating consumer needs a second out-of-band read of `schema --json` plus
hand-maintained knowledge of which strings are closed — and each future
vocabulary change ships invisible to the validator, which is the specific
failure 1.47 and 1.70 were labelled to prevent.

**Proposed ADR amendment** (for a human to accept or reject; nothing was
edited):

```
### Closed vocabularies are annotated, not asserted, in the generated schema

Reads are deliberately tolerant (ADR-0007 §4), so the generated Draft 2020-12
document MUST NOT constrain a vocabulary-bearing property with `enum`: a
payload reporting an unrecognized on-disk value is correct output and must
validate. Instead, every property whose value is drawn from a closed
vocabulary published by `schema --json` carries an `x-taskflow-vocabulary`
annotation naming that registry.

`reader_expectation` is extended to state the value half of the tolerant-reader
contract: a consumer switching exhaustively on such a property MUST resolve the
annotated registry from the same binary, and MUST surface an unrecognized
member rather than silently defaulting it.
```

**Recommendation:** Do NOT add enum — it would reject the tolerant reads lint depends on. Annotate each vocabulary-bearing property in the generated schema with an x- extension keyword naming its schema --json registry (e.g. x-taskflow-vocabulary: "statuses"), derived from the same domain owners runSchemaContract already uses, and guard it in the shape of TestDispositionDescriptionsReferencePublishedVocabularies so a vocabulary-typed DTO field with no annotation fails the build.

**Resolution:** Re-verified 2026-10-23: no x-taskflow vocabulary annotations
    exist. Tracked by a decision-first task whose first criterion is accepting,
    revising or rejecting the proposed ADR-0008 amendment before any schema
    change.

#### M1. The list-projection selector registries are declared contract surface that the binary never publishes · **Status:** tracked by 6gppr2k0exeb

**File:** internal/cli/schema.go:85 | **Component:** cli — schema contract / render column registry
**Effort:** S · **Urgency:** soon

**Class:** implementation drift
**Anchored to:** ADR-0008 §2 (the revision covers projected list JSON
"including selector names, value semantics, and requested column order"),
ADR-0008 §5 bullet 4 (derive published registries from their owners rather than
a second transcription), ADR-0008 §6 (consumers inspect the invoked binary's
contract); `https://www.anthropic.com/engineering/writing-tools-for-agents`

ADR-0008 §2 states the revision covers projected list JSON "including selector
names, value semantics, and requested column order", and §6 tells consumers to
read `schema --json` from the binary they invoke. `schema --json` publishes 14
top-level keys and none of them is a selector registry (verified on this
checkout: `schema_version`, `revision_policy`, `statuses`, `epic_statuses`,
`thread_statuses`, `audit_buckets`, `finding_statuses`, `criterion_states`,
`task_fields`, `epic_fields`, `research_fields`, `exit_codes`, `commands`,
`kinds`).

The only machine-readable record of the selectors is
`internal/cli/testdata/golden/projection_contract_json.golden`, generated by
`TestGolden_ProjectionContract` (`internal/cli/integration_golden_test.go:148`).
That is a test artifact: it ships with the source tree, not with the binary. A
consumer doing what §6 asks therefore cannot obtain contract surface §2 says
the revision covers.

What an agent must do instead is provoke a deliberate validation failure per
entity kind and parse the prose tail of the error message:

```
$ tskflwctl task list --json -c bogus
{"schema_version":"1.87","error":{"code":"validation","message":"validation
failed: unknown column \"bogus\" (available: slug, status, tier, priority,
epic, updated_at, description, revisit_at, id, location)"}}
```

Six registries exist — task (10), epic (9), audit (9), finding (9), research
(7), thread (12), 56 selectors in total — so full discovery costs six failing
calls and six prose parses. The error-message list is also an incomplete
substitute: it omits the four compatibility aliases 1.66 kept alive (`updated`,
`open`, and two on audit), which the golden does record. The trial-and-error
path under-reports the contract.

Two further properties of the projection contract are undiscoverable from the
binary. Values are stringified — full `--json` emits `"tier":3` where the
projection emits `"tier":"3"` (verified) — which CLAUDE.md states as policy but
nothing machine-readable does, even though §2 names "value semantics" as
covered. And the registry is a deliberately curated subset: `effort`,
`autonomy_level`, `created` and `tags` are full-envelope fields with no
selector. §5 expressly permits that ("A curated projection may remain small"),
but an agent can only learn the boundary by elimination.

This is the shape the project has already corrected three times —
`finding_statuses` at 1.21, `research_fields` at 1.34, `criterion_states` at
1.48 — each landing with the same reason recorded in the changelog: without it
"an agent had to trigger an error and parse prose to learn the set." The
selectors are the last registry still on the old footing and the most expensive
one, because the set is per-entity rather than global. It is also the registry
the Anthropic tool-design guidance speaks to most directly: that article
recommends making implicit knowledge explicit, including "definitions of niche
terminology, relationships between underlying resources", and treats
prompt-engineered errors as a fallback for mistakes rather than the designed
discovery path.

The assembled data already exists in exactly the publishable shape.
`render.Specs()` (`internal/cli/render/columns.go:151`) returns
`[]ColumnSpec{Name, Desc, Aliases}` and is already consumed by both the help
layer and the golden; `runSchemaContract` (`internal/cli/schema.go:85`) already
assembles the contract from domain owners alongside
`commandSafetySurface(root)`. The publication is a call, not a new source of
truth — so it does not create the second transcription §5 bullet 4 forbids.

**Compounding cost:** every new entity kind with a list view adds a registry
nobody can discover. Thread's landed three weeks ago at 1.72, and epic 28
`first-class-entities-new-planning-nouns` (active, 63%) is committed to
routines, projects and ADRs as further nouns. Publishing six registries now is
one additive revision reusing an existing accessor; publishing nine or ten
later is the same work over a wider surface. Meanwhile every agent written in
the interval hard-codes a selector list copied out of an error message it had
to provoke — which is precisely the second transcription §5 bullet 4 exists to
prevent, except that it accumulates in consumers where this repo cannot see or
fix it.

**Recommendation:** Add a per-entity selector registry to SchemaContract, populated from render.Specs(...) in runSchemaContract — the same assembly point that already derives exit codes and command safety — carrying each selector's name, description and aliases plus an explicit string value type, and reduce TestGolden_ProjectionContract to asserting the published contract instead of re-deriving it.

**Resolution:** Re-verified 2026-10-23: schema --json still publishes no
    selector registry. Tracked with L1, and the residual of 2026-07-24 M3, by
    the schema-contract publication task.

#### L1. The contract publishes the command surface and the payload surface but not the edge between them · **Status:** tracked by 6gppr2k0exeb

**File:** internal/wire/envelopes.go:1289 | **Component:** wire / cli — schema contract
**Effort:** S · **Urgency:** eventually

**Class:** unanchored (no ADR clause governs the mapping; adjacent to M1)
**Anchored to:** the generated schema's own root instruction,
`internal/wire/envelopes.go:1306`

The schema document tells a consumer to "validate a command's `--json` output
against the matching definition" in `$defs`, but nothing published says which
definition matches which command. `schema --json` lists 100 runnable command
paths with their safety capability; the schema lists 54 envelopes. The join is
left to name-guessing, and name-guessing fails on the commands an agent is most
likely to be driving: `task start` emits `moves`, `task ac --check` emits
`task_mutation`, `task path` emits `path`, and `thread add` emits
`thread_update`.

Recorded as Low and kept tight because it is adjacent to M1 rather than
independently urgent: both are the same incompleteness in the self-description,
and validation is the minority consumer path. The cost grows linearly — one
unmapped command per new leaf — and nothing compounds structurally.

**Recommendation:** Carry the emitted envelope definition name on SchemaCommand, set where each command's renderer is wired, so the already-published commands list becomes the join.

**Resolution:** SchemaCommand still carries only path, safety, hidden and
    deprecated. Tracked with M1 by the schema-contract publication task.

## Routine metadata

Routine: `weekly-architecture-audit` · lens `machine-contract` · ISO week
`2026-W42` (mod 4 = 2). On-schedule Monday run.

ADRs consulted: ADR-0008 (monotonic revisions for the JSON machine contract,
Decision §1–§6 and the 2026-09-19 Amendments in full), ADR-0007 (planning state
vocabularies, §2/§4), ADR-0003 (referenced for the frontmatter `schema:` key
that ADR-0008 §2 excludes from the wire contract).

Files read end-to-end: `internal/wire/wire.go` (the classified changelog and the
policy constants), `internal/wire/schema.go`, `internal/wire/wire_changelog_test.go`,
`internal/wire/registry_test.go`, `internal/wire/schema_descriptions_test.go`,
`internal/wire/schema_comments_test.go`, `internal/cli/exit.go`,
`internal/cli/command_safety.go`, `internal/cli/golden_test.go`,
`internal/cli/schema.go`, plus `internal/cli/render/columns.go:80-160` and
`internal/wire/envelopes.go:1124-1340` as the two edges that test whether the
pattern holds.

Sources cited: 3 fetched and quoted directly, 3 surfaced by search and
attributed as secondary (see Best-practice comparison).

Live probes were run against a scratch planning repo created with this
checkout's binary under the session scratchpad; nothing in `planning/` was
mutated to produce them.

## Executive summary

The machine contract is the healthiest lens this repo has. ADR-0008's policy is
not merely written down but executable at four independent points — a changelog
that must stay contiguous and classified, a registry guard that parses its own
package source to catch an envelope nobody registered, a golden-update gate that
refuses to rewrite a machine snapshot without first advancing and classifying
the revision, and a process-exit taxonomy with one owner and a fail-closed
mapping. The four findings the 2026-09-14 run of this lens raised are all fixed
or tracked.

The headline gap is a mirror image of that strength. ADR-0008 §3 makes
closed-vocabulary change the primary trigger for a `NOT ADDITIVE` revision, and
§4 makes the generated Draft 2020-12 document the validation artifact — but the
document expresses none of the six vocabularies the binary publishes, so the one
axis that most often forces incompatibility is the one axis the validator cannot
see (H1). The naive fix is wrong, because tolerant reads are a deliberate
decision in ADR-0007 §4 and the wire genuinely echoes an unrecognized on-disk
status; that is what makes this an ADR gap with a proposed amendment rather than
a defect to patch. Secondarily, the six list-projection selector registries are
declared covered by the revision in §2 but exist in machine-readable form only
inside a test golden, so the cheap agent path is still discovered by provoking an
error and parsing prose (M1) — the last registry still on the footing that 1.21,
1.34 and 1.48 each corrected.

Read H1 first; its proposed amendment is the only item needing a human decision.

## State of the architecture (this lens)

**The revision and its policy.** One global `SchemaVersion`, currently `"1.87"`
(`internal/wire/wire.go:373`), preceded by an 87-entry changelog in the same doc
comment. Every entry from 1.68 carries an `ADDITIVE` or `NOT ADDITIVE` label —
20 additive, 4 not (1.70, 1.82, 1.83, 1.84 at `wire.go:289,343,350,355`). Seven
constants beside the version pin the policy itself: scheme
(`monotonic-revision`), scope (`all-json-output`), the narrower generated-schema
scope (`typed-envelopes`), the classification boundary (`1.68`), the current
classification, the default, the reader expectation
(`ignore-unknown-object-fields`), and the validation mode (`exact-revision`).

That record is executable rather than descriptive.
`TestSchemaVersionChangelogIsAscending`
(`internal/wire/wire_changelog_test.go:131`) reads `wire.go` as bytes, slices out
the changelog between the doc comment and the `const`, and requires entries to be
contiguous, ascending, ending exactly at `SchemaVersion`, classified from the
boundary forward, and in agreement with `SchemaRevisionCompatibility`. It even
diagnoses an ASCII hyphen where an em dash is required
(`TestParseSchemaClassificationsExplainsDelimiter`). A separate test pins the
policy constants and refuses major-revision movement, which ADR-0008 §1 reserves
for a superseding ADR.

**The envelope layer.** 54 `*Envelope` types across `envelopes.go` (40),
`thread.go` (9), `dependency.go` (3), `dependency_repair.go` and
`task_rename.go`. All 54 carry `schema_version`. They are registered in one
`jsonEnvelopes` struct (`envelopes.go:1226`) so a single `Reflect` pulls them and
their shared types into one `$defs` registry — 154 definitions at 1.87.

Registration is guarded from both directions. `envelopes_test.go` walks the
registry demanding a validation case per registered envelope; `registry_test.go`
closes the inverse hole by parsing the package's own source files with `go/ast`
and failing any `type <Name>Envelope struct` that is not a field of
`jsonEnvelopes`. Its comment records the bug that motivated it: the research
envelopes shipped, emitted `schema_version`, appeared in no schema, and every
test passed. Reflection could not have caught it, so the package reads itself.

**Workspace coverage.** Changelog 1.31 gave every mutation envelope a
`workspace` object because external planning (epic 23) means the directory you
run in is not necessarily the tree you write to. Resolving embedded structs
transitively, that rule has held as the surface grew: `thread_mutation`,
`thread_update`, `thread_apply`, `dependency_mutation`, `graph_repair` and
`task_rename` all carry it, through their inner `*JSON` payloads. The two
mutation envelopes without it are `space_mutation` and `init` — `space` writes
the home-scoped registry rather than a planning tree, and `init` is what creates
the tree it would otherwise name, so both absences are principled.

**The error shape.** One `ErrorEnvelope` of `{schema_version, error}`, where
`ErrorItem` (`envelopes.go:1124`) is `{code, message}` plus nine optional typed
recovery payloads — `created`, `dependency_mutation`, `graph_repair`,
`task_lifecycle`, `task_rename`, `thread_mutation`, `thread_update`,
`thread_failure`, `thread_apply` — and `filesystem`, which classifies an OS
failure into `permission|not-found|read-only|no-space|io` with an explicit
`retryable` bit rather than leaving an agent to read English. `WriteError`
(`internal/cli/exit.go:136`) attaches each via `errors.As`, and notably clears
`retryable` when a create already committed, so a retry loop cannot be told to
repeat a call that would make a second document.

Streams are clean: on a genuine failure stdout is empty and the envelope goes to
stderr (verified). `lint --json` is the one deliberate exception — it writes its
report to stdout, exits 11, and writes a summary error envelope to stderr,
because its findings are data and the exit code is the signal. That is coherent,
though `WriteError`'s doc comment still asserts "stdout stays empty on failure",
which is now literally untrue for that one command; a comment-level inaccuracy,
routed to the next `code-quality-audit` rather than inflated here.

**The exit taxonomy.** `processExitCodes` (`internal/cli/exit.go:33`) is the
single owner of every number, machine name, meaning, and reservation state. The
original four domain rows stay first and in order so a consumer comparing that
historical prefix sees no churn; code 12 is published as `reserved` with the
explicit meaning "retired and reserved; this binary never emits it". `ExitCode`
composes `classifiedExitCode` with `activeExitCode`, so an unpublished or
reserved number cannot become a process contract by accident, and
`errorCodeName` derives the envelope's `error.code` from the same rows —
`schemaExitCodes()` copies them into the wire DTO rather than retyping them.

**Command safety.** 1.71 publishes all 100 runnable command paths with a
`read-only` or `mutating` capability (`internal/cli/command_safety.go:70`),
hidden and deprecated leaves included, in byte-stable path order. This is not
documentation: `commandSafetyState.bind` is fail-closed on an unrecognized
annotation, and `authorizeMutation` makes a read-only command reaching a
mutating persistence path a hard error. Cobra's private `__complete` transport,
which cannot be annotated alongside the registered tree, is handled explicitly.

**Projections.** Caller-selected `-c` projections are the second form of machine
JSON that §2 covers and the static schema deliberately cannot describe. Six
typed column registries back them (`internal/cli/render/columns.go`), validated
at construction by `validateColumnRegistry`, which rejects an empty selector, a
missing raw projector, or a name that would shadow an earlier canonical name or
alias. 1.66 moved projections onto canonical wire selectors with raw values and
kept the old names as aliases; 1.67 made stable `id` explicitly selectable.
Projected values are strings by design, so `tier` is `3` in a full envelope and
`"3"` in a projection.

**Goldens.** 41 machine-contract goldens, gated by
`internal/cli/golden_test.go`. A golden whose name ends `_json` must parse as a
JSON object and declare a revision identity; `validateMachineGoldenUpdate`
refuses to rewrite it at an unchanged revision, and `finalizeGoldenRevision`
advances the committed marker only after the full suite has exercised every
machine golden — so a partial `-update` cannot ratchet the marker. The single
exception is allowlisted by name and proves, by decoding both payloads and
removing `body`, that nothing adjacent rode along.

## ADR reconciliation

### ADR-0008 §1 — `schema_version` is a revision, not SemVer

**Quoted:** "`schema_version` is a monotonically increasing identifier for the
complete machine-readable JSON contract shipped by one `tskflwctl` binary… The
major component is reserved. Changing its meaning or starting a new major line
requires another ADR."

**Implementation:** follows. `SchemaRevisionScheme = "monotonic-revision"` with
the explicit comment "intentionally not semver" (`wire.go:376`);
`TestSchemaRevisionPolicyBoundary` fails if the major component ever moves off 1.
There are no per-envelope versions: all 54 carry the one global value.

### ADR-0008 §2 — the revision covers all JSON output; the generated schema covers typed envelopes

**Quoted:** "The revision covers… projected list JSON, including selector names,
value semantics, and requested column order."

**Implementation:** drift. The typed-envelope half follows exactly —
`JSONSchemaScope = "typed-envelopes"`, and the document says so. The projected
half is covered by the revision in policy and is revision-gated in the test
suite, but neither its selector names nor its value semantics are published by
the binary, so a consumer following §6 cannot read the contract §2 claims. See
M1.

**Class if divergent:** implementation drift.

### ADR-0008 §3 — every new revision declares compatibility

**Quoted:** "**`NOT ADDITIVE`** — a key, type, referent, requiredness, selector,
error code, or documented meaning changes incompatibly; or a published closed
vocabulary adds, removes, or renames a member that an exhaustive consumer may
switch on."

**Implementation:** follows as to the declaration, with a gap as to
verifiability. Every revision from 1.68 is labelled and the labels agree with
the constants, enforced by test. The ADR is candid that this is the limit of
automation: "A test can require a declaration but cannot infer semantic
compatibility." Spot-reading the four `NOT ADDITIVE` entries against their
diffs, the labels are truthful and the consumer action is named in each. The gap
is that the vocabulary clause quoted above — the clause that classified 1.47 and
1.70 — has no expression in the validation artifact §4 designates. See H1.

**Class if divergent:** ADR gap.

### ADR-0008 §4 — the generated schema has revision identity

**Quoted:** "`schema --json-schema` uses a `$id` containing the exact revision
and publishes the same revision and compatibility classification as root
annotations."

**Implementation:** follows, and the 2026-09-19 amendment went further than the
clause. `$id` is
`https://github.com/andy-esch/taskflow/internal/wire/json-envelopes/1.87`; the
root carries `x-taskflow-schema-version`, `x-taskflow-revision-scheme` and
`x-taskflow-revision-compatibility`; and `constrainEnvelopeSchemaVersions`
additionally pins each envelope's `schema_version` to a `const` so a mispaired
cached schema fails closed rather than silently passing.

### ADR-0008 §5 — consumer-facing output remains agent-usable

**Quoted:** "expose a readable slug or title for comprehension and the durable ID
needed for stored references or later calls"; "offer bounded filtering or
projection where an unbounded response would waste context"; "return structured,
specific recovery information instead of requiring prose parsing"; "derive
published vocabularies and registries from their owners rather than maintaining a
second transcription."

**Implementation:** follows on three of four, drift on the fourth. Readable slug
plus durable `id` is satisfied across the list and show envelopes and, since
1.67, in projections too. Bounded responses are satisfied by 1.85's `--limit` /
`--sort` / `--search` / `--updated-since` with a versioned `next_cursor`, and
`page` metadata appears only when a bound was requested, so the default shape is
unchanged (verified both ways). Structured recovery is satisfied richly — nine
typed recovery payloads plus classified filesystem errors. Derivation from owners
is where it slips: five of the six registries in `schema --json` are derived from
`domain` accessors in `runSchemaContract` and `schemaExitCodes()`, but the
projection selector registries are published nowhere, which leaves consumers to
transcribe them from an error message (M1).

**Class if divergent:** implementation drift.

### ADR-0008 §6 — consumers inspect the invoked binary's contract

**Quoted:** "Consumers should read `schema --json` and, when validating typed
envelopes, `schema --json-schema` from the binary they invoke… Unknown additive
object fields must be ignored; closed vocabulary changes and revisions marked
`NOT ADDITIVE` require deliberate handling."

**Implementation:** silent on the hard half. Both commands run without a planning
repo, so the inspection path exists. But the clause asks a consumer to handle
closed-vocabulary change deliberately while giving it no machine-readable way to
know which fields carry a closed vocabulary (H1) and no published selector
registry at all (M1). `ignore-unknown-object-fields` is published as the reader
expectation; the corresponding rule for unknown *values* is not stated anywhere.

### ADR-0007 §2/§4 — words spelled once; every closed vocabulary gets a validated atomic write verb

**Quoted:** "a value the tool cannot write is a value nobody can be held to";
"Reads stay tolerant so `lint` can REPORT malformed data already on disk; writes
refuse to create it."

**Implementation:** follows, and this clause is load-bearing for H1. The
vocabularies in `schema --json` are derived from `domain.FindingStatuses()`,
`domain.CriterionSuffixStates()`, `domain.AllStatuses()` and friends rather than
transcribed. Tolerant reads are confirmed empirically: `status: banana` round-trips
into `task list --all --json` verbatim and is reported by `lint`. That is why H1
recommends an annotation and not an `enum`.

### No ADR governs

Two parts of this lens have no governing clause, recorded here as ADR-0008 §2
asks. First, the mapping from a command to the envelope it emits (L1) — the
schema's root instruction assumes it and nothing publishes it. Second, the
stream and exit-code contract for a command whose *report* is a success-shaped
envelope but whose exit is non-zero, which `lint --json` relies on. The second
is coherent in practice and small enough to leave undecided; it is noted rather
than raised.

## Best-practice comparison

### Custom annotation keywords are the sanctioned route for schema metadata that asserts nothing

**Source:** `https://json-schema.org/blog/posts/custom-annotations-will-continue`
(JSON Schema project blog, 2023-02-22) — fetched and quoted directly.

**Guidance:** "going forward, JSON Schema will treat any keyword that starts with
`x-` as an annotation"; such values are "returned to the user or application
without any processing by the schema"; "Moving forward, prefix your custom
annotation keywords with `x-`" because then "there can be no collision with
keywords defined by vocabularies, now or in the future." The post adds that `x-`
keywords already work with published versions, where they "will still just be
collected as annotations."

**This codebase:** partial. The convention is already adopted at the document
root — `x-taskflow-schema-version`, `x-taskflow-revision-scheme`,
`x-taskflow-revision-compatibility` (`internal/wire/envelopes.go:1308-1312`) —
and is exactly the right choice there. It stops at the root: no property-level
annotation carries contract semantics, which is where H1 proposes extending it.

**Justified?** No. The mechanism is already in the file, costs nothing at
validation time, and is the only way to express vocabulary membership without
breaking the tolerant reads ADR-0007 §4 requires.

### `enum` is an assertion, so it cannot express a vocabulary a tolerant reader must still accept

**Source:** `https://json-schema.org/understanding-json-schema/reference/enum` —
fetched and quoted directly.

**Guidance:** "The `enum` keyword is used to restrict a value to a fixed set of
values. It must be an array with at least one element, where each element is
unique." A non-member value is marked "not compliant to schema." The page offers
no guidance on choosing `enum` over an annotation.

**This codebase:** diverges, and correctly. Only two properties are
enum-constrained, both two-valued literals on the schema-contract DTOs
themselves. Everything drawn from a planning vocabulary is an unconstrained
string.

**Justified?** Yes, for the DTOs that echo on-disk data — verified above, an
`enum` on `TaskJSON.status` would invalidate the diagnostic payload `lint`
depends on. Not justified as a reason to publish *nothing*: the choice to avoid
the assertion does not imply the choice to omit the annotation, and H1 separates
the two.

### Agent tool surfaces should make implicit knowledge explicit rather than teach by error

**Source:** `https://www.anthropic.com/engineering/writing-tools-for-agents`
(Anthropic Engineering) — fetched and quoted directly.

**Guidance:** make implicit knowledge explicit, including "definitions of niche
terminology, relationships between underlying resources"; "input parameters
should be unambiguously named"; errors should be prompt-engineered "to clearly
communicate specific and actionable improvements" and may "give examples of
correctly formatted tool inputs"; large responses want "some combination of
pagination, range selection, filtering, and/or truncation"; and agents handle
natural-language names "significantly more successfully than they do with
cryptic identifiers" while still needing "technical IDs to chain follow-up tool
calls".

**This codebase:** follows on three counts, partial on one. Bounded responses
arrived at 1.85 with exactly the recommended combination. The slug-plus-id
pairing is the article's identifier guidance implemented directly, and the
changelog at 1.32 records fixing the one command where an agent captured the
wrong handle. Errors are genuinely actionable — the lint diagnostic names the
repair command, and the space-registry conflict prints the two commands to run.
The partial is discovery: the selector sets are learnable *only* through an
error (M1), which makes the error the designed discovery path rather than a
fallback for mistakes.

**Justified?** No. The data is already assembled by `render.Specs()` and already
revision-gated in a golden; publishing it is a call at an existing assembly
point.

### Tolerance belongs to structure the consumer ignores, not to values it switches on

**Source:** `https://quality.arc42.org/approaches/tolerant-reader` — surfaced by
search; the page did not resolve from this sandbox, so this is secondary
attribution to the arc42 quality model's write-up of Fowler's 2011 pattern, and
the quotes below are the search summary's rather than verified page text.
Corroborated independently by GraphQL's "Robust Applications" guidance
(`https://graphql.org/learn/robust-applications/`, also unreachable here), which
per the same search holds that treating every enum value as exhaustively known
causes crashes or silent data loss and prescribes an explicit default branch.

**Guidance:** a consumer should discard fields its schema does not know, but
tolerance applies to structure the consumer never reads, not to values it does;
an unknown member of a vocabulary the consumer reads should be rejected, alerted
on, or routed to a fallback — never silently defaulted.

**This codebase:** partial. The field half is published as policy
(`reader_expectation: ignore-unknown-object-fields`) and enforced in prose by
§3's "Consumers must ignore unknown object keys." The value half is unstated,
and the producer does emit unrecognized vocabulary values by design, so the
situation the guidance warns about is reachable: a consumer switching on
`status` meets `"banana"` with no published rule for what to do.

**Justified?** No, and it is the substance of H1's proposed amendment. The
tolerance itself is justified — it is how `lint` reports bad data — but leaving
the consumer rule unstated is not.

### Runtime contract publication is a tooling affordance, not a production contract

**Source:** `https://grpc.io/docs/guides/reflection/` and the gRPC server-reflection
design notes — surfaced by search; neither resolved from this sandbox, so this is
secondary attribution.

**Guidance:** reflection lets a server "declare the protobuf-defined APIs they
export… including all types referenced by the request and response messages" so
clients "can then use this information to encode requests and decode responses
in a human-readable manner." The design notes frame it as "an optional extension
for servers to assist clients in runtime construction of requests without having
stub information precompiled into the client", whose "primary usecase… is to
write (typically) command line debugging tools", and are explicit that a server
"is under no obligation to return a complete list of all methods it supports."

**This codebase:** follows, and exceeds the framing. `schema --json` is not an
optional debugging aid here — it is the designed discovery path for the tool's
primary consumer, and 1.71 deliberately publishes the *complete* runnable
command surface including hidden and deprecated leaves, which the gRPC design
explicitly declines to promise.

**Justified?** The divergence is in the project's favour and worth keeping. It
does, though, temper M1's severity: the precedent literature treats runtime
contract publication as a convenience rather than an obligation, which is part
of why M1 is Medium and not High.

## Tensions and trade-offs

**Tolerant reads versus a validating contract.** This is the real tension behind
H1 and it has no costless resolution. ADR-0007 §4 makes reads tolerant so `lint`
can report malformed data; ADR-0008 §4 makes the generated schema a fail-closed
validator. Those pull opposite ways on any field whose value comes off disk. The
project cannot assert the vocabulary without breaking the diagnostic, and cannot
stay silent without leaving the `NOT ADDITIVE` axis unvalidatable. An annotation
splits the difference rather than resolving it — the consumer still has to do the
check — which is why H1 is framed as a decision for a human rather than a patch.

**A single global revision versus per-surface honesty.** ADR-0008 §1 rejected
per-envelope versions for good reasons, and the consequence is accepted
pessimism: a thread-only change advances the number for every consumer, and
four of the twenty classified revisions are `NOT ADDITIVE` for reasons most
consumers will never touch (1.83 and 1.84 are both audit close/defer semantics).
A consumer comparing exact revisions therefore sees more churn than it
experiences. For a single-binary, single-user tool this is the right trade — one
number is inspectable in one call — but it is a real cost and it grows with the
surface.

**The static schema cannot describe the dynamic half.** §2 is candid that
caller-selected rows "cannot have one static shape", so the contract has two
forms with one revision. The project's answer — a committed golden recording
ordered selectors and aliases — is sound engineering for *regression*, but a
golden is a source-tree artifact and a contract needs to travel with the binary.
That mismatch is M1's root, and it is a consequence of the two-form design rather
than a mistake in it.

**Local-first legitimises what would elsewhere be gaps.** No negotiated content
type, no deprecation window, no parallel-version serving, no schema registry.
For a tool whose consumer invokes the same binary it inspects, every one of those
is correctly absent, and the project should not be scored against them. The
exit-code taxonomy is the pattern to emulate here: a closed vocabulary, one
owner, published with reservation state, and fail-closed at the boundary — a
networked service's rigour applied at exactly the one place where the contract
actually crosses a process boundary.

## What audited clean

- **The changelog is executable, not documentary.** 87 contiguous entries,
  classified from 1.68, with `wire_changelog_test.go` parsing `wire.go`'s own
  bytes to enforce contiguity, ascent, terminal agreement with `SchemaVersion`,
  classification presence, and constant agreement — ADR-0008 §3.
- **Major-revision movement is test-blocked.** `TestSchemaRevisionPolicyBoundary`
  fails if the major component leaves 1, making ADR-0008 §1's "requires another
  ADR" a build-time fact rather than a convention.
- **The envelope registry is guarded in both directions.**
  `registry_test.go` parses the package with `go/ast` to catch a declared but
  unregistered envelope — the one failure mode reflection is blind to, and one
  that actually shipped once. ADR-0008 §2's "the generated schema covers typed
  envelopes" is only true because of this.
- **Exact-revision validation fails closed.**
  `constrainEnvelopeSchemaVersions` pins every envelope's `schema_version` to a
  `const`, so a cached schema paired with the wrong payload is rejected rather
  than silently accepted — ADR-0008 §4 and its 2026-09-19 amendment.
- **The golden gate cannot be ratcheted by accident.** A machine golden may not
  be rewritten at an unchanged revision, and the committed marker advances only
  after the full suite exercised all 41 — with the single body-only exception
  allowlisted by name and proved by payload comparison.
- **The process-exit taxonomy has exactly one owner.** `processExitCodes` holds
  every number, name, meaning and reservation state; `activeExitCode` makes a
  reserved or unpublished number unreachable at the boundary; `errorCodeName`
  and `schemaExitCodes()` are consumers, not transcriptions — ADR-0008 §5
  bullet 4.
- **Command safety is an enforced invariant, not a label.** `bind` is
  fail-closed on an unrecognized annotation and `authorizeMutation` turns a
  read-only command touching a mutating persistence path into an error; the
  published list covers all 100 runnable leaves including hidden and deprecated
  ones — ADR-0008 §5, and stricter than the gRPC-reflection precedent cited
  above.
- **Recovery information is structured, never prose.** Nine typed recovery
  payloads on one error shape, plus `filesystem` classifying an OS failure with
  an explicit `retryable` bit that is deliberately cleared once a create has
  committed — ADR-0008 §5 bullet 3, and the Anthropic guidance on actionable
  errors.
- **Bounded responses arrived in the recommended shape.** 1.85's limit, sort,
  search, updated-since and versioned cursor, with `page` metadata opt-in so the
  unbounded default shape is unchanged — ADR-0008 §5 bullet 2.
- **`workspace` coverage survived the surface tripling.** The 1.31 rule still
  holds across every mutation envelope added since, through embedded payloads;
  the two exceptions (`space`, `init`) are principled rather than oversights.
- **Reflected descriptions are pinned to their owners where they can be.**
  `schema_comments.json` is regenerated and byte-compared with a floor guard
  against a silently-empty comment map, `EpicMetaJSON.status`'s literal tag is
  pinned to `domain.AllEpicStatuses()`, and the disposition fields must name
  their published registry. This is the prose form of the machinery H1 asks for
  structurally — which is also why H1 is a small step rather than a new idea.

## Related-task observations (propose-only)

- Possibly ready to close: `planning/audits/6ggqwdz14nzx-2026-10-05-arch-data-model-and-storage.md`
  reports `ready_to_close: true` in `audit list --json` (1 finding, 1 done, 0
  open). Noted in passing while running this run's backpressure check; not acted
  on, since this was not a triage run.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- → H1 · tracked — Decide how the JSON Schema expresses closed vocabularies | tracked by 6gppr2k482rm
- → M1 · tracked — Publish selector registries in schema --json | tracked by 6gppr2k0exeb
- → L1 · tracked — Name each command's emitted envelope | tracked by 6gppr2k0exeb
