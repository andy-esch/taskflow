---
schema: 1
id: 6g7qc8qd00xe
bucket: closed
area: arch-data-model-and-storage
date: "2026-09-07"
updated_at: "2026-09-07"
---

# Weekly Architecture Audit: data-model-and-storage — 2026-09-07

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `data-model-and-storage` · ISO week
`2026-W37` (mod 4 = 1).

ADRs consulted: ADR-0003 (§2/§3/§4/§6 + all three amendments), ADR-0007 (§4),
ADR-0006 (Thread/task identity namespace only) ·
Files read: `internal/domain/entity.go`, `internal/domain/fields.go`,
`internal/domain/schema.go`, `internal/domain/validate.go`,
`internal/domain/lint.go`, `internal/domain/layout.go`,
`internal/store/frontmatter.go`, `internal/store/create.go`,
`internal/store/atomic.go`, `internal/store/fix.go`,
`internal/core/service.go` (Lint), `internal/core/service_research.go`,
`internal/core/service_audit.go`, `internal/domain/epic.go`,
`internal/domain/research.go` · Sources cited: 5

## Executive summary

The storage half of this lens is in genuinely good shape: surgical frontmatter
editing, atomic replacement, the single `validateIDUpdate` choke point, and the
entity carveout contract are all faithful to ADR-0003 and hold up under
adversarial reading. The **identity** half does not. ADR-0003 §3 requires
creation to "collision-check the id and regenerate"; only `research` implements
it. I reproduced a duplicate stable id on two audit documents: `lint` prints
"✔ all planning entities and dependency links pass lint" while every write verb
on either document fails permanently with `ambiguous match` (H1). Underneath it
sits an ADR gap — nothing in any accepted ADR says whether the 12-char id
namespace is global across kinds or per-kind, and the code has quietly taken
both positions (M1). Read H1 first; then note that epic 26's open policy ADR
(`6fkkz41cax80`) covers the schema-marker and field-registry findings (M2, M3)
but its Q7 referential-rules list omits identity uniqueness entirely.

Off-schedule: no. Run is on Monday 2026-09-07, ISO week 2026-W37.

## State of the architecture (this lens)

**Domain.** `internal/domain/entity.go:58` is the entity registry: one
`Descriptor` per kind (task/epic/audit/research/thread) carrying directory,
authoring fields, conventions, templates, and placeholders. `SchemaKinds`,
`AuthoringFields`, `Conventions`, `Template`, and `Placeholders` all read it, so
the *metadata* fan-out for a new kind really is one entry.

Storage **types** are a separate axis and are not registry-driven in the same
way. Tasks declare theirs in a typed table (`internal/domain/fields.go:24`,
23 rows) from which `knownTaskFields`/`intFields`/`listFields`/`dateFields` are
derived (`fields.go:65`). Epics declare theirs as a hand-written map literal
(`internal/domain/epic.go:106`) with a hardcoded list-field predicate
(`epic.go:130` — `return f == "tags"`). Research derives its set *from the
Descriptor* (`internal/domain/research.go:81`), with the comment naming this the
intended charter: "a noun's fields ride one registry". Audits and Threads have
no known-field registry, which is principled — neither has a `set` verb
(`audit --help` exposes no `set`; Thread membership mutations have not landed).

**Identity.** Ids are 12-char Crockford base32, 43 bits ms-timestamp + 17 bits
random (ADR-0003 §3 as amended 2026-07-02). Tasks, audits, epics and Threads
mint at creation time via `newID` (`internal/core/service.go:31-36`); research
alone mints *from its `created` day* via `newIDAt`
(`internal/core/service_research.go:89`), which is why same-day research docs
share a single 2^17 random slot.

**Store.** `internal/store/frontmatter.go:118` (`updateFrontmatter`) is the one
field-write path: split the fence strictly, unmarshal to a `yaml.Node`, mutate
nodes in place, re-assemble. `setMapNode` (`frontmatter.go:357`) replaces a
value node while carrying the old node's head/line/foot comments forward unless
the replacement brings its own. `detectLineEnding` (`frontmatter.go:82`)
re-emits in the file's own EOL style. `validateIDUpdate` (`frontmatter.go:94`)
sits at the top of `updateFrontmatter` and is explicitly documented as the choke
point that `task set --force` was observed bypassing.

Writes are atomic: `stageTemp` → `Sync` → `Chmod` → `Rename` → `syncDir`
(`internal/store/atomic.go:13-62`), with the destination's existing mode
preserved on overwrite, and `createFileAtomic`'s `O_EXCL`
(`atomic.go:84`) documented as the version-CAS `ifVersion == ""` precondition.

**Repair.** `FixFrontmatter` (`internal/store/fix.go:25`) does text-level
normalization, backfills a missing `id:` from the id already leading the flat
filename (`fix.go:280`), and canonicalises a misspelled Crockford id in both
filename and field — but *refuses* when the id is referenced elsewhere in the
tree, because there is no rename cascade (`fix.go:168-180`). That refusal is the
right shape: a repair that trades one broken file for several dangling links is
not a repair.

**Which ADR governs this?** ADR-0003 governs paradigm (§1), the stable key (§2),
identity form (§3), filename (§4), references (§5), and migration (§6), plus the
three amendments. ADR-0007 §4 governs the body-level closed vocabularies.
ADR-0006 governs Thread membership. **Nothing governs the uniqueness *scope* of
a stable id** — see M1.

## ADR reconciliation

### ADR-0003 §3 — creation collision-checks the id

**Quoted:** "**Creation collision-checks** the id and regenerates on the
astronomically-rare intra-millisecond clash (or generates monotonically within a
tick)."

**Implementation:** *drift, partial.* Only `NewResearch` implements it
(`internal/core/service_research.go:82-98`: mint, attempt, regenerate on
`ErrConflict`, bounded at `maxIDMintAttempts`), backed by a store-side scan of
ids already on disk (`internal/store/create.go:275-292`). `CreateTask` and
`CreateAudit` do not check the id at all; both carry the comment "The id makes
the flat filename unique, so writeNewFile's `O_EXCL` is the whole collision
guard" (`create.go:155-157`, `create.go:230-231`) — the exact claim
`CreateResearch` documents as **false** twenty lines later: "a duplicate id on a
DIFFERENT slug is a different path, which O_EXCL never sees" (`create.go:275`).

**Class if divergent:** implementation drift → H1.

### ADR-0003 §4 — id-led flat filenames, one directory per kind

**Quoted:** "`<id>-<slug>.md`, in one flat directory per entity (`tasks/`,
`epics/`, `audits/`); no status, bucket, or epic subdirectories."

**Implementation:** *follows.* Verified across all five kinds
(`internal/domain/layout.go:7-14`, `internal/store/paths.go`); the corpus is
334 tasks, 69 audits, 31 research docs, 2 Threads, 15 epics, all conforming.
Epics keep `NN-<slug>` per the 2026-07-04 amendment.

### ADR-0003 §4 amendment (2026-07-04) — the carveout contract

**Quoted:** "**Entity = a filename-shape test.** A file directly in a scanned
bucket is an entity iff its name leads with a valid stable key… A *positive*
signal, so a real entity that lost its frontmatter still fails loud… while a
non-id-led file is simply not one." And: "**`meta/` — the sanctioned home.**"

**Implementation:** *follows.* `planning/meta/` exists and holds exactly the
non-entity material it is meant to (`no-op-log.md`); resolution gates on the
same shape.

### ADR-0003 §6 — migration is a one-time throwaway script

**Quoted:** "the migration is a **one-time throwaway script**, *not* a permanent
`tskflwctl migrate` command… **Run once per repo, then discarded** — it need not
be general, configurable, or supported."

**Implementation:** *drift, benign.* Three migration tools remain in the tree
and in the `go build ./...` / `go test -race ./...` surface —
`internal/tools/flatmigrate` (433 lines), `internal/tools/researchmigrate`
(523), `internal/tools/wikimigrate` (261) — none referenced outside historical
planning docs. More importantly the §6 posture ("hard cutover, no coexistence")
is about to be contradicted by work already scoped: the open task
`enforce-reserved-document-schema-versions-across-entity-writers` requires "an
explicit migration contract before the first real schema bump, including
dry-run, idempotency, partial-failure recovery, and Git-native rollback
expectations."

**Class if divergent:** ADR pressure → L1.

### ADR-0007 §4 — every closed vocabulary gets a validated, atomic write verb

**Quoted:** "Every closed vocabulary gets a validated, atomic write verb. `task
ac` owns criterion state; `audit finding` owns finding status and its
`**Resolution:**` paragraph. Reads stay tolerant so `lint` can REPORT malformed
data already on disk; writes refuse to create it."

**Implementation:** *follows* for the vocabularies themselves. Worth recording
under this lens because H1 breaks the *reachability* of that guarantee rather
than the guarantee: on a duplicate-id audit, `audit finding` is not tolerant or
strict — it is unreachable (`re-resolve audit … : ambiguous match`).

### No ADR governs: the uniqueness scope of a stable id

ADR-0003 §3 mints an id per entity and calls it "collision-safe at planning
scale", but never says *within what scope* two ids may not collide. The code has
taken both positions: `ARCHITECTURE.md` asserts "Task creation and Thread
creation check one cross-kind stable-ID namespace", `internal/core/service.go:513`
and `:592` tell the user "task and Thread identities must be globally unique",
while `crossKindIdentityOwner` (`internal/store/fix.go:225-245`) returns `""`
for any directory other than tasks/threads and `ensureTaskIDNotThread`
(`internal/store/create.go:191`) checks Threads only. An audit and a task may
share an id today with no diagnostic anywhere. **Class: ADR gap → M1.**

## Best-practice comparison

### Every persisted object carries a version the reader honours

**Source:** <https://raw.githubusercontent.com/kubernetes/community/master/contributors/devel/sig-architecture/api-conventions.md>

**Guidance:** "All JSON objects returned by an API MUST have the following
fields: kind: a string that identifies the schema this object should have;
apiVersion: a string that identifies the version of the schema the object should
have. **These fields are required for proper decoding of the object.**"

**This codebase:** *partial.* `domain.FileSchemaVersion = 1`
(`internal/domain/layout.go:23`) is stamped first into every new task, audit,
research, epic, and Thread (`store/create.go:91,207,251,351`,
`store/threadcreation.go:150`) and its own doc states the purpose: "so a future
format migration has an in-file signal to branch on." No non-test code reads it.
Corpus coverage is partial and nothing backfills it: **248/334 tasks, 67/69
audits, 10/15 epics, 31/31 research, 2/2 Threads** carry `schema:`; `lint --fix`
reports "nothing to fix" on a file that lacks it.

**Justified?** Partly. ADR-0003 §6 deliberately chose a hard cutover with no
dual-layout read path, so *not branching today* is consistent. What is not
justified is a marker whose 26% gap is unrecorded and un-backfilled — see M2.

### Name uniqueness must have a declared scope

**Source:** <https://raw.githubusercontent.com/kubernetes/community/master/contributors/devel/sig-architecture/api-conventions.md>

**Guidance:** a name is "A string that uniquely identifies this object **within
the current namespace**", and a UID is "A unique in time and space value…used to
distinguish between objects with the same name". Uniqueness is always stated
against an explicit scope.

**This codebase:** *diverges.* The scope is never stated, and three different
scopes are implemented simultaneously (see the ADR-reconciliation entry above).

**Justified?** No. Single-user and local-first do not remove the need to say
what a primary key is unique against; they only reduce how often it is tested.

### A consistency checker must cover the invariants the store relies on

**Source:** <https://raw.githubusercontent.com/git/git/master/Documentation/git-fsck.adoc>

**Guidance:** git-fsck "Verifies the connectivity and validity of the objects in
the database", performing "full tracking of the resulting reachability" and
testing "SHA-1 and general object sanity" — i.e. the checker's job is the
invariants the object store depends on, not a subset of them.

**This codebase:** *partial.* `lint` is thorough where it looks: dangling links,
graph health, status/bucket vocabularies, id drift, duplicate epic `NN`
(`internal/domain/lint.go:55`), duplicate research ids
(`internal/core/service.go:556`), duplicate Thread ids (`:580`), and duplicate
*task* ids via the dependency-graph source analysis (verified: two tasks on one
id produce "duplicate stable task id … no source is uniquely authoritative").
Audits are the hole: no intra-kind duplicate-id check and no cross-kind check.

**Justified?** No. The check is one line of wiring over an existing helper whose
own doc comment (`internal/domain/lint.go:335-349`) describes exactly this
failure mode.

### Adding one field must not mean editing N places kept in lockstep by eye

**Source:** <https://raw.githubusercontent.com/kubernetes/community/master/contributors/devel/sig-architecture/api_changes.md>

**Guidance:** "you have to touch a number of pieces to make a complete change" —
versioned types, internal types, defaults, conversions, validation, generated
code — and "It must be possible to round-trip your change (convert to different
API versions and back) with **no loss of information**."

**This codebase:** *partial.* The round-trip half is satisfied (see below). The
"one place" half is three places with three different shapes: a derived typed
table (task), a hand-written literal (epic), and a registry-derived set
(research). Drift tests in `internal/domain/schema_test.go` (`TestTask
AuthoringFieldsMatchRegistry`, `TestEpicAuthoringFieldsMatchRegistry`,
`TestAuditAuthoringFieldsMatchStruct`, `TestTaskFieldsMatchStruct`) hold them in
sync today, so this is a structural cost, not live drift — see M3.

### YAML round-tripping via a node tree loses textual representation

**Source:** <https://pkg.go.dev/gopkg.in/yaml.v3>

**Guidance:** "although Node offers access into details such as line numbers,
colums, and comments, **the content when re-encoded will not have its original
textual representation preserved**. An effort is made to render the data
plesantly, and to preserve comments near the data they describe, though."

**This codebase:** *follows, with the risk understood and bounded.* This is the
single biggest latent hazard in the lens — every surgical write re-encodes the
whole frontmatter block through `assembleFile` (`store/frontmatter.go:273`) — so
I tested it rather than assuming. Round-tripping a realistic frontmatter block
(a 122-character plain `description`, a flow-style `tags` list, a double-quoted
scalar, a `!!int` and a `!!str` id) through the exact
`Unmarshal → Node → Encoder{SetIndent(2)} → Encode` path reproduced the input
**byte-identically**: no folding of the over-80-column scalar, quoting style
preserved, flow sequence preserved, key order preserved. `setMapNode`'s comment
carry-forward (`frontmatter.go:357-376`) covers the comment case the upstream
doc warns about. The residual exposure the doc names is real but currently
unexercised.

**Justified?** Yes — and the mitigation is the right one: the repo's own fuzz
targets (`store/fuzz_test.go`) plus the CRLF/blockstyle regression already
recorded as a task (`6g1dp77y0g5n-researchmigrate-corrupts-block-style-tags-and-crlf-files`).

### Timestamp-prefixed ids need same-instant disambiguation

**Source:** <https://raw.githubusercontent.com/ulid/spec/master/README.md>

**Guidance:** "Within the same millisecond, sort order is not guaranteed", and
the monotonic variant specifies that "if the same millisecond is detected, the
`random` component is incremented by 1 bit in the least significant bit position
(with carrying)."

**This codebase:** *partial.* ADR-0003 §3 chose the same shape (43 time + 17
random) and explicitly accepted the weaker option — regenerate on clash rather
than monotonic increment — which is fine. The divergence is that the regenerate
is implemented for one kind of five.

**Justified?** For *tasks/audits/epics*, ms-precision minting makes a real mint
collision effectively impossible, so the absence of a regenerate loop is
defensible. The absence of any *detection* is not (H1): the realistic source of
a duplicate id is a copied file or a merge, which no minting policy prevents.

## Tensions and trade-offs

- **Single-user scope vs. declared invariants.** Several guards here are absent
  on the reasoning that concurrent creation "doesn't occur in practice"
  (`store/create.go:319-326`, `nextEpicNumber`). That reasoning is sound for
  *races*. It does not transfer to H1/M1, whose trigger is a copy, a merge, or a
  hand-edit — all of which a single user does more often than a daemon would.
- **Five typed entities vs. one generic one.** `ARCHITECTURE.md` argues, and I
  agree, that a further data-driven persistence collapse "trades clarity for
  machinery" for five heterogeneous entities. M3 is deliberately *not* a request
  for that collapse; it is a request that the three kinds which already have a
  field registry share one shape.
- **Hard cutover (ADR-0003 §6) vs. a corpus with in-the-wild repos.** The ADR's
  throwaway-migration posture was correct for a two-repo internal tool. Epic 23
  (external planning repos) and epic 26 have since widened that, and L1 is where
  the seam shows.

## Findings

#### H1. A duplicate stable id on an audit is invisible to `lint` and permanently bricks both documents  · **Status:** fixed

**File:** `internal/store/create.go:230` · `internal/core/service.go:556-580` |
**Component:** store/create, core/lint
**Effort:** S · **Urgency:** soon
**Class:** implementation drift
**Anchored to:** ADR-0003 §3 ("Creation collision-checks the id and
regenerates") and <https://raw.githubusercontent.com/git/git/master/Documentation/git-fsck.adoc>

`CreateAudit` performs no id-uniqueness check — only `writeNewFile`'s `O_EXCL`
on the full `<id>-<slug>.md` path — and `Service.Lint` wires
`domain.DuplicateIDIssues` for research (`service.go:556`) and Threads
(`service.go:580`) but never for audits. Tasks are covered by a different
mechanism (the dependency-graph source analysis), and epics by
`DuplicateEpicNNIssues`. **Audits are the only id-led kind with neither.**

Reproduced end to end in a scratch planning repo (`tskflwctl init`, one audit,
`cp audits/<id>-alpha-area.md audits/<id>-beta-area.md`):

    tskflwctl lint            -> ✔ all planning entities and dependency links pass lint
    tskflwctl lint --fix      -> nothing to fix
    tskflwctl audit list      -> both listed, no marker, indistinguishable from healthy
    tskflwctl audit close <id>       -> "6g7qb0qv1m5z" matches 2 audits by id: ambiguous match
    tskflwctl audit close alpha-area -> re-resolve audit … : ambiguous match
    tskflwctl audit finding beta-area H1 --status fixed --note "done"
                                     -> re-resolve audit … : ambiguous match

Both documents are permanently unwritable through every audit verb, by id *and*
by unique slug, because the write path's CAS re-resolve goes `ErrAmbiguous`.
There is no `audit rename`, so recovery is a hand `mv` plus a hand-edit of the
frontmatter `id:` — and nothing tells the operator that. The hygiene command
that exists to catch this affirmatively reports health.

This is the *same defect* the completed tier-1 task
`6g1dnnfgyjap-duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably`
diagnosed, reproduced, and fixed — for research only. Its "Out of scope" section
lists the siblings it deliberately deferred (`verifyUnchanged` error collapsing,
create-path `id.Valid` checking) and does **not** list generalizing the check to
the other id-led kinds. The class was left open when the instance was closed.

**Compounding cost:** Two costs compound separately. (1) Every week this routine
and `code-quality-audit` add audits; the corpus is at 69 and grows by ~2/week,
and each new document is another that a `cp`-based duplication or a branch merge
can silently brick. (2) Epic 26's `enforce-reserved-document-schema-versions-across-entity-writers`
is about to route "parsing/lint diagnostics and guarded/generic writers through
one declared policy" across every kind; adding the uniqueness rule now makes it
one row in that policy, while adding it later means retrofitting a rule the
policy has already been designed without.

**Recommendation:** Two lines of wiring plus a guard, mirroring what research
already has. (a) In `Service.Lint`, build `auditIDs` from the records already
loaded at `service.go:573` and pass them through `domain.DuplicateIDIssues`, the
same shape as `service.go:556`. (b) In `CreateAudit`, scan `auditCandidates()`
for the minted id before writing, mirroring `CreateResearch`
(`store/create.go:283-292`). No regenerate loop is needed — audits mint at
ms precision (`core/service_audit.go:48`), so a mint clash is not the realistic
trigger; detection is.

**Follow-up:** The general question — should `DuplicateIDIssues` be applied
uniformly over the union of all id-led kinds rather than kind-by-kind — is M1's,
not this finding's. H1 is fixable without waiting for M1.

**Resolution:** CreateAudit now serializes canonical ID collision detection with
file creation under the repository write lock, and ordinary lint reports every
duplicate audit ID with all conflicting sources. Store race coverage, core lint
coverage, the original CLI reproduction, the full race-enabled suite, static
analysis, docs/module checks, planning lint, and diff hygiene pass. Implemented
by task 6g7s4k845fsb; the analogous research serialization gap is tracked by
6g7s6hr3qnfq.

#### M1. The cross-kind stable-id namespace is asserted in three places and enforced between two kinds  · **Status:** tracked by 6fkkz41cax80

**File:** `internal/store/fix.go:225-245` · `internal/store/create.go:191-202` ·
`internal/core/service.go:513,592` | **Component:** domain/identity
**Effort:** S (decision) / M (enforcement) · **Urgency:** soon
**Class:** ADR gap
**Anchored to:** <https://raw.githubusercontent.com/kubernetes/community/master/contributors/devel/sig-architecture/api-conventions.md>

No accepted ADR states the scope within which a 12-char stable id must be
unique. ADR-0003 §3 decides the id's *form* and calls it "collision-safe at
planning scale"; ADR-0006 introduces Threads into the same id space. Neither
says whether tasks, audits, research docs and Threads share one namespace.

The code answers "global" in its user-facing text and "task↔Thread only" in its
enforcement:

- `internal/core/service.go:513` and `:592` tell the user "task and Thread
  identities must be **globally unique**".
- `docs/ARCHITECTURE.md` asserts "Task creation and Thread creation check one
  cross-kind stable-ID namespace under the same canonical-root guard."
- `crossKindIdentityOwner` (`store/fix.go:225`) switches on `s.tasksDir` and
  `s.threadsDir` and returns `"", nil` for everything else.
- `ensureTaskIDNotThread` (`store/create.go:191`) is named for, and checks, only
  Threads.

So an audit and a task, or a research doc and a Thread, may share an id today
with no create-time guard, no lint diagnostic, and no repair-time refusal.
Whether that *matters* depends on the undecided question: if ids are per-kind,
"globally unique" is the wrong message and the task↔Thread check is
over-enforcement; if they are global, the enforcement is 40% complete. Scheme 2
(ADR-0003 §5) leans global — body cross-links are plain relative paths, but an
id-prefix resolve that has to be told which kind it is looking in is a weaker
primitive than one that does not.

The open policy ADR (`6fkkz41cax80-adr-close-frontmatter-schema-policy-questions`)
is the natural home, but its **Q7 — Cross-field / referential rules** candidate
list (`epic` exists · pointer validity · `related_tasks` resolvable · `blocks`
symmetry · date ordering) does **not** include identity uniqueness — the one
referential rule whose violation has already been reproduced as unrecoverable in
this repo's own corpus.

**Compounding cost:** Cheap to decide now, expensive later in exactly one way:
every additional first-class noun (epic 28 has ADRs and projects scoped) doubles
the pairwise cross-kind checks someone must write by hand if the answer is
"global" and the enforcement stays ad-hoc. Two kinds is one check; five kinds is
ten, or one uniform pass. Deciding after those nouns land means writing the
pairwise version first and the uniform version afterwards.

**Recommendation:** Decide the scope in the epic-26 policy ADR, then implement
it once. If global: replace the three ad-hoc checks with one pass over the union
of id-led records at create time and in `Lint`. If per-kind: fix the two lint
messages and drop `ensureTaskIDNotThread`/`crossKindIdentityOwner`'s special
case.

**Proposed ADR amendment:** to ADR-0003, under `## Amendments` — *"§3 decides the
id's form but not its uniqueness scope. Stable ids are unique across the whole
planning space, not per entity kind: `<id>` resolves to at most one document of
any kind. Consequence: creation and `lint` check the id against the union of
id-led records (tasks, audits, research, Threads), not only against the writer's
own directory."* Stated as the direction the code's own messages already claim;
the ADR owner may of course choose per-kind instead, in which case the amendment
should say that and the messages should be corrected.

**Follow-up:** Add identity uniqueness to Q7's candidate rule list in
`6fkkz41cax80` so the ADR survey can decide it rather than inherit it.

**Resolution:** The frontmatter-schema policy ADR task now makes stable-ID
uniqueness scope an explicit Q7 decision and records both possible enforcement
consequences.

#### M2. `schema:` is stamped into every new document, read by nothing, and absent from 26% of the corpus  · **Status:** tracked by 6g7f0tqgftg3

**File:** `internal/domain/layout.go:16-23` · `internal/store/create.go:91,207,251,351` |
**Component:** domain/layout, store/create
**Effort:** S · **Urgency:** soon
**Class:** ADR gap
**Anchored to:** <https://raw.githubusercontent.com/kubernetes/community/master/contributors/devel/sig-architecture/api-conventions.md>
("required for proper decoding of the object")

`FileSchemaVersion = 1` is stamped as the **first** frontmatter key of every new
task, audit, research doc, epic, and Thread, and a golden test
(`store/schema_version_test.go:31`) pins that position. Its doc comment states
the purpose: "so a future format migration has an in-file signal to branch on."

No non-test code reads it. More consequentially, nothing backfills it, and the
corpus predates it:

| kind | with `schema:` | total | missing |
| :-- | --: | --: | --: |
| tasks | 248 | 334 | **86** |
| epics | 10 | 15 | **5** |
| audits | 67 | 69 | **2** |
| research | 31 | 31 | 0 |
| threads | 2 | 2 | 0 |

`lint --fix` reports "nothing to fix" on a schema-less entity — a principled gap
in one reading (the marker is not a domain field, and `--fix` deliberately stays
narrow) but an unrecorded one, since `backfillMissingID` (`store/fix.go:280`)
establishes that `--fix` *does* backfill exactly this class of tool-managed key.

The result is a marker that cannot do its stated job. On the day a schema 2
lands, absence has to be read as "1", which means the marker never distinguishes
anything until schema 3 — and the 93 currently-unmarked documents become a
permanent special case in every migration that follows.

**Compounding cost:** Linear and monotone. The unmarked set is frozen at 93
today and can only be repaired while the tool still treats absence as benign;
once a real bump exists, backfilling becomes a semantic claim ("these were
version 1") that nobody can verify from the file alone. Every week of delay adds
no new unmarked files but shortens the window in which the fix is trivially
safe.

**Recommendation:** Either (a) backfill `schema: 1` through `lint --fix` while
absence is still unambiguously "pre-marker", or (b) record explicitly — in the
epic-26 ADR — that absence means 1 forever, and drop the pretence that the
marker is a branch point. Do not ship a schema 2 before one of the two.

**Proposed ADR amendment:** none for ADR-0003 (§6's scope is the 2026-07 flatten).
This belongs in the epic-26 policy ADR's **Q10 — Schema versioning & evolution**,
which already asks "Reuse the `schema:` frontmatter field as the version
handle?" — the answer should be accompanied by the backfill decision, not just
the reuse decision.

**Follow-up:** ⏳ tracked — see the cross-reference below.

**Resolution:** Corpus inventory for this repo (93 schema-less documents)
gathered by the 2026-09-07 weekly architecture audit; the
backfill-vs-declare-absence decision belongs to this task's Q10 dependency.

#### M3. Three entity kinds have a known-field registry, in three different shapes  · **Status:** tracked by 6fkkz41cax80

**File:** `internal/domain/fields.go:24-70` · `internal/domain/epic.go:106-130` ·
`internal/domain/research.go:81-92` | **Component:** domain
**Effort:** M · **Urgency:** eventually
**Class:** unanchored
**Anchored to:** <https://raw.githubusercontent.com/kubernetes/community/master/contributors/devel/sig-architecture/api_changes.md>

The entity `Descriptor` registry (`domain/entity.go:58`) successfully collapsed
the *metadata* fan-out. The *storage-type* fan-out did not follow it uniformly:

- **task** — a typed table (`fields.go:24`) with four derived sets
  (`fields.go:65`). Not derived from the Descriptor; a sync test
  (`TestTaskAuthoringFieldsMatchRegistry`) pins the two together.
- **epic** — a hand-written map literal (`epic.go:106`) plus a hardcoded
  predicate `IsEpicListField(f) { return f == "tags" }` (`epic.go:130`).
- **research** — derived *from* the Descriptor at package init
  (`research.go:81`), with the comment naming the intent: "a noun's fields ride
  one registry."
- **audit, thread** — none, correctly: neither has a `set` verb.

Research implements the stated charter; task and epic predate it and were never
migrated. Being fair to the code: `internal/domain/schema_test.go` holds all
three in sync, so this is **not live drift** — it is a structural cost. The
cost is that "how does this kind declare its fields?" has three answers, so
adding a `set` verb to audits or Threads (epic 28's trajectory) means choosing
one, and epic 26's single-declared-registry work has to unify three shapes
rather than generalize one.

**Compounding cost:** Bounded but real. Each new noun that gains a mutable field
surface adds a fourth (fifth…) choice point, and each new *field* on task or
epic is a two-place edit held together by a test rather than a one-place edit.
The tests make the cost maintenance rather than correctness — which is why this
is Medium and not High.

**Recommendation:** Fold this into epic 26's declared-registry work rather than
fixing it standalone: whatever shape Q9 ("One engine, per-entity schemas over a
shared core") lands on, migrate task and epic onto research's Descriptor-derived
form so there is one answer.

**Follow-up:** ⏳ tracked — see the cross-reference below.

**Resolution:** Q9 (entity coverage & sharing) is the deciding question; the
three existing registry shapes are inventoried in the finding.

#### L1. ADR-0003 §6's throwaway-migration posture is the project's only accepted migration policy, and open work now needs a different one  · **Status:** tracked by 6fkkz41cax80

**File:** `internal/tools/{flatmigrate,researchmigrate,wikimigrate}` |
**Component:** tools
**Effort:** XS (the ADR entry) · **Urgency:** eventually
**Class:** ADR pressure
**Anchored to:** ADR-0003 §6 and the scope of
`enforce-reserved-document-schema-versions-across-entity-writers`

§6 decided "**Run once per repo, then discarded** — it need not be general,
configurable, or supported." Three such scripts (1,217 lines total) remain in
the tree and in the build/test surface, unreferenced outside historical planning
docs — a benign drift on its own. The pressure is that the open epic-26 task
requires "an explicit migration contract before the first real schema bump,
including dry-run, idempotency, partial-failure recovery, and Git-native
rollback expectations", which is precisely the *permanent, general, supported*
migration facility §6 rejected. Epic 23 (external planning repos) makes "a small,
known set of repos" less true than it was in June.

Listed as Low, and adjacent to M2, because nothing is broken today — but an ADR
whose stated policy contradicts the next piece of scoped work should say so in
its own Amendments rather than be quietly outvoted by a task.

**Compounding cost:** Near-zero this quarter. It compounds only at the first real
schema bump, when someone reads §6 as the governing policy and either builds a
throwaway for a facility that needs to be durable, or builds the durable one and
leaves the ADR stating the opposite.

**Recommendation:** No code change. Add the amendment below when the epic-26 ADR
is written, so the two land together.

**Proposed ADR amendment:** to ADR-0003, under `## Amendments` — *"§6's
throwaway-script decision governs the 2026-07 flatten specifically, not
migration policy in general. A durable migration contract for on-disk schema
bumps (dry-run, idempotency, partial-failure recovery, Git-native rollback) is
epic 26's to decide; §6 does not preclude one."* The 2026-07 migration tools
under `internal/tools/` may then be retired or kept deliberately, as a separate
call.

**Resolution:** The schema-evolution question now distinguishes ADR-0003's
one-off flat-layout migration from the durable migration contract required
before a real schema bump.

## What audited clean

- **Surgical frontmatter editing survives a real round-trip.** Verified
  byte-identical output for long plain scalars, quoted scalars, flow sequences,
  and key order through the exact production encode path
  (`store/frontmatter.go:273`) — the hazard `gopkg.in/yaml.v3`'s own docs warn
  about is present in the library but not realised here.
- **Comments are carried forward, not clobbered.** `setMapNode`
  (`frontmatter.go:357-376`) moves head/line/foot comments onto a replacement
  node only when the replacement has none of its own — satisfies "no loss of
  information" (k8s api_changes).
- **One choke point for id writes.** `validateIDUpdate`
  (`frontmatter.go:94-124`) sits above every field write including
  `task set --force`, which its comment records as having previously bypassed
  the known-field registry. This is the right response to a real incident.
- **Atomic replacement is complete, not partial.** `stageTemp` fsyncs and chmods
  before rename, `syncDir` follows, destination mode is preserved on overwrite,
  and `sweepStaleTemps` is scoped by the tool's own prefix *and* an hour of age
  (`atomic.go:13-139`) — ADR-0003 §2's "makes content-OCC a one-file check".
- **`createFileAtomic`'s `O_EXCL` is correctly identified as the CAS empty
  precondition** (`atomic.go:82-84`), so creates need no separate
  `verifyUnchanged`.
- **Unterminated frontmatter fails loud rather than silently becoming a body.**
  `splitFrontmatterStrict` (`frontmatter.go:60-66`) exists specifically to stop a
  surgical edit prepending a second fence — a subtle failure caught by design.
- **The carveout contract is implemented as a positive test.** ADR-0003's
  2026-07-04 amendment ("Entity = a filename-shape test", `meta/` as the
  sanctioned home) holds across resolution and listing, so a stray cannot shadow
  a real entity.
- **Repair refuses when it would dangle references.** `repairInvalidID`
  (`fix.go:168-180`) declines to canonicalise an id referenced elsewhere and
  names the referrers instead — the correct posture for a repo with no rename
  cascade.
- **Research's day-minted id is handled exactly right.** Regenerate-on-clash,
  bounded attempts, store-side scan, and a lint diagnostic
  (`service_research.go:82-98`, `store/create.go:275-292`,
  `service.go:556`) — the ULID spec's same-instant problem, solved. H1 asks only
  that the *detection* half be generalized.
- **`ValidateMintableDate`** (`validate.go:92-105`) guards the one place a
  backdated id can silently destroy id-order-is-date-order, with the bounds
  derived from `id.MaxMillis` rather than retyped.

## Candidate tasks (human to triage)

- `tskflwctl task new "Detect duplicate stable ids on audits at create and lint time" --epic 28-first-class-entities-new-planning-nouns --tags store,domain,lint --tier 1 --priority high --description "CreateAudit does no id-uniqueness check and Lint wires DuplicateIDIssues for research/threads only; two audits on one id are unwritable with lint reporting clean."`
- `tskflwctl task new "Decide and enforce the uniqueness scope of a stable id" --epic 26-frontmatter-schema-declared-validation-contract --tags domain,identity,adr --tier 2 --priority medium --description "No ADR states whether the 12-char id namespace is global or per-kind; the code claims global in user text and enforces task-Thread only."`
- ⏳ M2 tracked in `planning/tasks/6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md`
- ⏳ M3 tracked in `planning/tasks/6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md`

## Related-task observations (propose-only)

- **Class left open when the instance was closed:**
  `planning/tasks/6g1dnnfgyjap-duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably.md`
  (completed, tier 1, high) fixed this exact defect for research and enumerated
  its deferred siblings in "Out of scope" — generalizing the duplicate-id check
  to the other id-led kinds is not among them. H1 is the same defect on audits.
  No status change proposed; flagged because a reader of that task would
  reasonably believe the class was handled.
- **Survey gap:**
  `planning/tasks/6fkkz41cax80-adr-close-frontmatter-schema-policy-questions.md`
  Q7's candidate referential-rule list omits stable-id uniqueness. Suggest adding
  it as a Q7 option before the survey is answered — it is the one rule in that
  family with a reproduced unrecoverable failure in this corpus. Proposal only;
  the survey is Andy's to fill in.
- **Inventory now available:** the same task's sibling
  `6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers`
  scopes "Inventory existing schema-less fixtures and real planning data so
  rollout is deliberate rather than an accidental mass failure." M2's table above
  is that inventory for this repo (93 documents), gathered 2026-09-07.
