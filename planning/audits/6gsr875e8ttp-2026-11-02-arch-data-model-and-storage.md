---
schema: 1
id: 6gsr875e8ttp
bucket: open
area: arch-data-model-and-storage
date: "2026-11-02"
updated_at: "2026-11-02"
---

# Weekly Architecture Audit: data-model-and-storage — 2026-11-02

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.
> Architecture audits are propose-only — no code, docs, or ADR edits.

Routine: `weekly-architecture-audit` · lens `data-model-and-storage` · ISO week
`2026-W45` (mod 4 = 1).
ADRs consulted: ADR-0003 (§2, §3, §4, §5, §6 + all four amendments), ADR-0006
(§6, §8, §11), ADR-0007 (§4, §5) · Files read end-to-end:
`internal/domain/entity.go`, `internal/domain/fields.go`, `internal/domain/layout.go`,
`internal/domain/research.go`, `internal/domain/thread.go`, `internal/domain/validate.go`,
`internal/store/frontmatter.go`, `internal/store/frontmatter_preserve.go`,
`internal/store/fix.go` (repair paths), plus `internal/domain/epic.go`,
`internal/domain/schema_test.go` and `internal/store/schema_version_test.go` as the
guard inventory · Sources cited: 5

## Executive summary

The write machinery under this lens is in good shape and the surgical-preservation
guarantee is stronger than the YAML library it is built on promises. What this run
turned up is not a persistence defect but a **coverage asymmetry across entity kinds**:
the same storage question is answered differently for each of the five registered
nouns, and the two places where it is answered most weakly are the two the ADRs
speak about most clearly. The headline finding is implementation drift — ADR-0006 §8's
accepted CLI surface specifies `thread set|edit|rename` and `thread info`, and none of
the four exists, leaving Threads the only entity kind with no write path to its own
frontmatter at all. Two Medium findings concern the storage-machinery keys `id` and
`schema`: both are reachable through `--force` on tasks and epics, while the research
kind refuses all three of its machinery keys with stated reasons. Read
"State of the architecture" for the per-kind inventory, then M1 — its guard's own
code comment states a justification that the repair path no longer supports.

## State of the architecture (this lens)

**Layout and identity (ADR-0003 §3/§4) are implemented uniformly and are not where
the asymmetry lives.** `internal/domain/layout.go:7` names the six planning
directories; the flat, id-led layout holds for tasks, audits, research and Threads,
with epics keeping their `NN` prefix per ADR-0003's 2026-07-04 amendment. Identity
validity is strict lowercase Crockford base32, enforced at the single frontmatter
write choke point (`internal/store/frontmatter.go:117`) and empirically confirmed:
`--set id=NOTANID` and `--set id=iiiiiiiiiiii` are both refused with position-accurate
messages, even under `--force`.

**The surgical write path is one choke point with a two-tier fidelity strategy.**
`updateFrontmatter` (`frontmatter.go:112`) edits a `yaml.Node` in place, replacing
existing keys where they sit and appending only new keys in lexical order
(`frontmatter.go:135-148`). `assembleEditedFile`
(`internal/store/frontmatter_preserve.go:23`) then adds a second tier **only for
documents containing a block scalar** (`:28-30`): it re-splices the original source
text of untouched top-level entries, decodes both sides, and refuses the write
outright if the decoded values differ from the requested edit (`:96-98`). Documents
without a block scalar take `assembleFile` (`frontmatter.go:270`) directly, with no
verify-after-encode step.

I probed that ungated path rather than assume it, because the prior same-lens audit's
fidelity probe carried a literal scalar and therefore exercised only the *gated* tier.
A task document carrying a head comment, an inline trailing comment, a free-floating
foot comment, single- and double-quoted scalars, a flow mapping, an empty value, and a
YAML anchor/alias pair survived a `task set --set tier=4` with every one of those
intact, in original key order, with `updated_at` appended; the only deltas were inline
comment-space normalization (documented in the persistence guide) and one inserted
blank line before the appended key. A 189-character description was not re-wrapped
despite `assembleFile` setting no encoder width. So the ungated tier holds in practice
for the shapes the corpus actually contains.

**"Which fields exist" is answered five different ways, one per noun.** This is the
centre of gravity for this lens today:

| Kind | Field registry | Machinery-key protection | AuthoringFields drift guard | Generic write verb |
| --- | --- | --- | --- | --- |
| task | declarative `taskFields` table, type sets derived (`fields.go:24`, `:66`) | `ValidateField` rejects `status`, `updated_at` only (`validate.go:194`) | yes (`schema_test.go:13`, `:80`) | `task set` |
| epic | hand-written 5-key map (`epic.go:105`) | `ValidateEpicField` rejects `status` (`epic.go:135`) | yes (`schema_test.go:34`) | `epic set` |
| audit | none (identity is immutable) | n/a | yes, vs the struct (`schema_test.go:56`) | none by design |
| research | **derived from the registry** (`research.go:65`) | `ProtectedResearchField` refuses `created`, `id`, `schema`, `updated_at`, each with its reason (`research.go:123`) | **none** | `research set` |
| thread | none | n/a | **none** | **none** |

Only research realises the epic-28 charter its own comment states — "a noun's fields
ride one registry". Task keeps a separate declarative table that a test pins to the
`Task` struct but that nothing reconciles against the task Descriptor's typed
`AuthoringFields`; epic keeps a hand-written map whose comment claims it is "every
frontmatter key tskflwctl itself reads or writes on an epic" while the tool also
writes `schema` (`create.go:313`) and `updated_at`, neither of which is in it.

**Validation strength runs opposite to write-path strength.** `ValidateThreadDocument`
(`thread.go:72`) is the most complete document-level validator in the tree — it
enforces ADR-0006 §6's duplicate-free, stable-ID-only, sorted `tasks` set, required
description and goal, and date shape on four optional timestamps. Its own doc comment
says it is "shared by guarded creation and **future** guarded Thread mutations". There
are no Thread mutations for it to guard: `thread` offers membership verbs, lifecycle
verbs, compose/apply, and read verbs, and nothing that writes `description`, `goal`,
`target_date` or `tags`.

**The reserved on-disk format marker is write-only.** `FileSchemaVersion`
(`layout.go:23`) is stamped first into every new document by five writers
(`create.go:143`, `:313`, `:368`, `:459`, `threadcreation.go:161`). Its own comment
states "The loader ignores it". Nothing reads it, nothing lints it, and `lint --fix`
does not backfill it — all three confirmed, and the first two restated by the code
comment itself. The only reader of a `schema` key anywhere is
`internal/cli/task_dependency_repair.go:31`, for an apply-plan document, not a planning
entity.

**Repair is deliberately narrow and converges on the filename.** `lint --fix` assigns a
*missing* id from the id that already leads the filename (`fix.go:344`), and
canonicalises a *misspelled* id by moving the filename and the field together
(`fix.go:224`, `:273`). Both therefore write a value that agrees with the name by
construction. A frontmatter id that is well-formed but *disagrees* with the filename is
reported and explicitly declined: `lint --fix` prints "could not auto-repair" with a
"rename the file or fix the field" remedy. That boundary is principled — automatic
repair should not acquire rename authority — but see M1 for what it implies about the
write gate.

## ADR reconciliation

### ADR-0006 §8 — the accepted Thread CLI surface

**Quoted:** "`tskflwctl thread set|edit|rename <thread> ...`" and
"`tskflwctl thread info|path <thread>`" (`planning/adrs/0006-adopt-threads-as-task-dags.md:441-442`).
**Implementation:** drift — the implemented `thread` tree provides `add`, `apply`,
`cancel`, `complete`, `compose`, `frontier`, `graph`, `list`, `new`, `path`, `plan`,
`remove`, `reopen`, `show`, `start`. `set`, `edit`, `rename` and `info` are absent.
Those four strings occur exactly once in the ADR — in §8 — and neither §11
("Explicitly Out of Scope") nor any of the fifteen dated amendments retracts them.
**Class if divergent:** implementation drift → H1.

### ADR-0006 §6 — Thread identity reuses the task/research conventions

**Quoted:** "Thread IDs reuse ADR-0003's task/research conventions: 12-character
stable ID in both filename and frontmatter, exact-ID resolution, drift lint, and
rename-safe identity."
**Implementation:** partial — identity, resolution and drift lint hold. "Rename-safe
identity" is the half that does not: ADR-0003 §5's Scheme 2 makes rename safety an
*active* cascade (`task rename` rewrites inbound body links), and Threads have no
rename verb to cascade with. 165 inbound relative-path links to thread files exist
across 152 files, including `docs/ARCHITECTURE.md`.
**Class if divergent:** implementation drift → folded into H1.

### ADR-0003 §3/§4 — the id is the stable key and leads the filename

**Quoted:** "Each entity carries a stable **`id`** (new frontmatter field)" (§3);
"`<id>-<slug>.md`, in one flat directory per entity … **Id-led** (not slug-led) so the
filename sorts chronologically and a known id resolves by prefix-glob" (§4).
**Implementation:** follows on layout and format; **silent** on whether a generic
forced write may change the declared id. The codebase answered that question twice and
differently — `research.go:128` refuses it citing "the immutable key [that] must match
the filename", `validateIDUpdate` (`frontmatter.go:96`) permits any well-formed value.
**Class if divergent:** ADR gap → M1.

### ADR-0003 — the `schema` key is governed by no clause at all

ADR-0003 decides the paradigm, organizing key, identity, filename and reference scheme,
and §6 decides migration. It never mentions a per-document format-version key. The
`schema:` marker is therefore an **undocumented architectural decision**: five writers
stamp it, one entity kind protects it, and no accepted decision states what it means or
what a reader owes it. Recording that is the explicit purpose of the unstarted ADR task
`6fkkz41cax80`.
**Class:** ADR gap → M2, M3.

### ADR-0007 §4/§5 — validated write verbs, tolerant reads

**Quoted:** "Every closed vocabulary gets a validated, atomic write verb. … Reads stay
tolerant so `lint` can REPORT malformed data already on disk; writes refuse to create
it." (§4) · "Invariants are enforced at write time where a linter would only report
them." (§5)
**Implementation:** follows within the ADR's declared scope. ADR-0007 is explicitly
about the two *body-level* closed vocabularies (criterion state, finding status), and
both are tool-owned and validated. I am **not** scoring the frontmatter findings as
drift against it: extending §4/§5 from bodies to frontmatter machinery keys would be
my extension of an accepted decision, not the decision itself. The principle is cited
in M1/M2 as this project's own stated posture, which is a different and weaker claim
than drift.

### ADR-0003 §6 — migration is a throwaway script, not a CLI command

**Quoted:** "the migration is a **one-time throwaway script**, *not* a permanent
`tskflwctl migrate` command … **Run once per repo, then discarded**."
**Implementation:** follows on the load-bearing half. `flatmigrate`, `researchmigrate`
and `wikimigrate` are `package main` programs under `internal/tools/` with no reference
from `internal/cli/` or `cmd/` — grep-confirmed, so no permanent migrate verb exists.
1,217 lines of one-time migration code are retained in-tree rather than literally
discarded; that is a reasonable reading of "discarded" (not shipped as a command) and
not scored as drift. Noted in Tensions.

## Best-practice comparison

### A declared format version is only load-bearing if the reader acts on it

**Source:** https://github.com/open-telemetry/opentelemetry-go/pull/4538
**Guidance:** the OpenTelemetry file-format specification is quoted in that review as
saying consumers "SHOULD NOT attempt to interpret the schema file if the MINOR version
number is higher", and the implementation under review documents that "If r contains a
Schema with a file format version higher than FileFormat, an error will be returned."
The thread is instructive in both directions: MrAlias argues strict single-version
parsing "[is not] the user experience we want to provide", and pellared proposes the
middle path of accepting a newer-version file that uses no unknown fields.
**This codebase:** diverges — `schema: 1` is stamped by five writers and read by none
(`layout.go:23` states this outright). No reader refuses, warns about, or branches on
a value it does not support, and the key is itself writable to an arbitrary string
through `--force` (M2).
**Justified?** Partly, and shrinking. No incompatible bump has happened, so there is
nothing yet to refuse, and ADR-0003 §6's hard-cutover posture deliberately rejects a
dual-layout read path. But the OTel thread's disagreement is about *which* refusal rule
to adopt, not about whether the reader participates. taskflow's corpus inventory (in
task `6g7f0tqgftg3`) shows the absence case is a closed, non-growing backlog of 94
pre-marker documents, so the decision is cheap now and gets no cheaper.

### Round-tripping YAML through a node tree does not preserve source text

**Source:** https://pkg.go.dev/go.yaml.in/yaml/v3#Node
**Guidance:** the library documents `Node` as "an intermediate representation that
allows detailed control over the content being decoded or encoded", and then disclaims
fidelity directly: "the content when re-encoded will not have its original textual
representation preserved", with only "an effort … made to render the data plesantly,
and to preserve comments near the data they describe" (typo in the source).
**This codebase:** follows, and exceeds the library's guarantee where it matters.
Rather than trusting the encoder, `assembleEditedFile` re-splices the original source
bytes of untouched entries and then **verifies by decoding both sides and refusing the
write** when they disagree (`frontmatter_preserve.go:96-98`) — converting the library's
"an effort is made" into a hard pre-commit check. That is the right shape of answer to
a documented non-guarantee.
**Justified?** n/a — follows. The one honest caveat is that the verify tier is gated on
`containsBlockScalar`, so the stronger guarantee is scoped to the shapes that provoked
the 2026-10-05 H1 rather than applied universally; the empirical probe above is why I
am not raising that as a finding.

### One declared schema per document type, reused by validation and authoring

**Sources:** https://github.com/jackchuka/mdschema ·
https://documentation.platformos.com/developer-guide/platformos-check/checks/valid-frontmatter
**Guidance:** mdschema states that "Frontmatter is required once a `frontmatter` block
exists; fields are required unless marked `optional`", that "A schema has one
`structure` tree plus optional document-wide rules", and that entries "are checked
against the declared `type`/`format` when the schema loads" — schema-load-time
validation from one declaration. platformOS's `valid_frontmatter` check flags keys "that
are not part of the file type's frontmatter schema", i.e. a closed-by-default key set
per document type.
**This codebase:** partial — one *authoring* registry exists and is genuinely good
(`entity.go:52` carries directory, fields, conventions, templates and placeholders per
kind, and drives `schema <kind>`), and research derives its known-field set from it
(`research.go:65`). But the *validating* registry is per-kind and hand-maintained for
task and epic and absent for thread, so the schema a drafter reads and the rules a
write enforces are reconciled by three bespoke tests covering three of five kinds.
**Justified?** Not fully. The project constraint that legitimately applies here is
tolerant reads over a legacy corpus — a closed-by-default key set cannot simply be
switched on when 18 research docs carry a vestigial `status:` and unknown-field
preservation is a stated guarantee. That argues for *where* strictness lands, not for
five registries. Epic 26 exists precisely to decide this and its ADR is unstarted → M3.

### Checking and repairing are separate capabilities, and repair may legitimately decline

**Source:** https://djwong.org/docs/xfs-online-fsck-design.html (read via search summary;
see the research caveat below)
**Guidance:** the XFS online-fsck design separates a facility that checks metadata from
one that repairs it, pursues "targeted checking and repair", and explicitly concedes
that online fsck "will never fix 100% of what offline fsck can repair" — a repair tool
declining a class of damage is a design position, not a gap.
**This codebase:** follows. `lint` reports frontmatter/filename id disagreement with a
concrete remedy while `lint --fix` declines it, and the persistence guide states the
reason ("automatic repair does not gain case-only rename authority"). Repair converges
on the filename rather than guessing intent (`fix.go:344`, `:273`).
**Justified?** Yes — and this is the practice M1 leans on: if repair legitimately
declines a damage class, the write path should not be the thing that manufactures it.

### Enforce invariants where values are constructed, not with a later check

**Source:** https://deviq.com/principles/parse-dont-validate/ (read via search summary;
see the research caveat below)
**Guidance:** the "parse, don't validate" principle, as summarised there from Alexis
King's essay, shifts responsibility for correctness "from scattered validation checks to
type construction", with parsing applied at system edges such as config loading, so
that downstream code needs no re-checking.
**This codebase:** partial. Creation paths call typed validators with native types and
`validate.go:19-21` states the intent plainly ("Typed validators are the canonical field
rules … so the rules live in exactly one place"). The `--force` escape hatch is the
documented exception, and it is the surface both M1 and M2 sit on.
**Justified?** The escape hatch itself is justified — it was added deliberately
(`6fbj87001d81`, 2026-06-12) so unknown keys need an explicit opt-in. What is not
justified is its *reach*: `--force`'s own help text promises "allow --set of a field
tskflwctl doesn't know", and `id` and `schema` are not fields the tool doesn't know.

**Research caveat.** Five searches were run. Direct fetches succeeded for
pkg.go.dev, github.com/open-telemetry and github.com/jackchuka. Three hosts
(`images.lsst.io`, `pipelines.lsst.io`, `deviq.com`) do not resolve under this
session's network policy, so the two sources marked above are cited from search-result
summaries including their quoted fragments, not from a page I fetched. The LSST
schema-versioning model (a `schema version` plus a separate `minimum reader version`,
so an old reader refuses a file only when told to) is relevant prior art for M2 and is
named here without being cited as read.

## Tensions and trade-offs

**Tolerant reads versus a closed key set.** ADR-0003's carveout amendment and the
research corpus's vestigial `status:` key both commit the tool to preserving frontmatter
it does not model. A closed-by-default key set (the platformOS practice) therefore
cannot be adopted wholesale, which is exactly why `--force` exists. The resolution is
not to close the key set but to separate *unknown* keys (preserve, require `--force`)
from *reserved machinery* keys (refuse, with a reason) — the distinction research
already draws and task/epic do not.

**Single-user scope legitimately defers the format-version question, but not forever.**
A networked multi-tenant service would have to refuse an unsupported format on day one.
taskflow has one author and git as its undo, so "stamp now, decide later" was a sound
trade. The cost of the deferral is that the decision must land *before* the first bump
to be worth anything, and released binaries already in the wild will never check.

**Retaining the throwaway migrators.** ADR-0003 §6 said "run once per repo, then
discarded"; 1,217 lines remain under `internal/tools/`. Retention is cheap and useful
as executable provenance for how the corpus reached its current shape, and the ADR's
real constraint (no permanent `migrate` verb) holds. Worth a one-line retention note in
the ADR's amendments eventually; not worth a finding.

## Findings

Severity order: Critical → High → Medium → Low. Every entry below was created
through `tskflwctl audit finding new`, which owns its code, structure and
`**Status:**`; nothing in this section is hand-written.

#### H1. Threads have no guarded write path to their own frontmatter, and ADR-0006 §8's accepted set|edit|rename|info verbs are all absent · **Status:** open

**File:** internal/domain/thread.go:72 | **Component:** cli/thread + domain/thread (write surface)
**Effort:** M · **Urgency:** soon

**Class:** implementation drift
**Anchored to:** ADR-0006 §8 (`planning/adrs/0006-adopt-threads-as-task-dags.md:441-442`), ADR-0006 §6 ("rename-safe identity"), ADR-0003 §5 (Scheme 2 link cascade)

ADR-0006 §8 specifies the accepted Thread CLI surface. Two of its lines read
`tskflwctl thread set|edit|rename <thread> ...` and `tskflwctl thread info|path <thread>`.
The implemented command tree provides `add`, `apply`, `cancel`, `complete`, `compose`,
`frontier`, `graph`, `list`, `new`, `path`, `plan`, `remove`, `reopen`, `show`, `start`.
`set`, `edit`, `rename` and `info` do not exist. Those four verb names occur exactly
once in the ADR — in §8 — and nothing in §11 ("Explicitly Out of Scope") or in any of
the fifteen dated amendments through 2026-09-12 retracts them.

The consequence is structural, not cosmetic. `domain.Thread`
(`internal/domain/thread.go:52-67`) persists `description`, `goal`, `target_date` and
`tags` — ordinary mutable initiative metadata, declared as authoring fields on the
registry (`internal/domain/entity.go:164-170`) and settable at creation. After creation
there is no tool route to any of them. Thread is the only one of the five registered
kinds with neither a `set` verb nor an `edit` escape hatch: task, epic, audit and
research all have `edit`, and task, epic and research have `set`.

So correcting a Thread's goal requires hand-editing frontmatter, which `AGENTS.md`
forbids ("do not hand-edit tool-owned sections or frontmatter to bypass guards") and
which bypasses `ValidateThreadDocument` (`thread.go:72`) — the most complete
document-level validator in the tree, whose own comment says it is "shared by guarded
creation and **future** guarded Thread mutations". The validator names the gap; there
are no mutations for it to guard.

The missing `rename` is the sharper half, because ADR-0006 §6 claims Threads reuse
ADR-0003's conventions including "rename-safe identity". Under ADR-0003 §5's Scheme 2,
rename safety is not a property of the id — it is an active cascade: body cross-links
embed real relative paths, so `task rename` rewrites every inbound link and `lint`
flags danglers (ADR-0003's 2026-09-08 amendment specifies that cascade as a guarded,
resumable multi-document mutation). Threads have the links but not the cascade. There
are 165 inbound relative-path links to thread files across 152 files today, including
one from `docs/ARCHITECTURE.md`. Re-titling any of the 13 Threads therefore means a
manual file rename plus up to 165 manual link rewrites, with `lint --links` available
only to report the danglers afterward.

**Compounding cost:** this is cheapest to fix while there are 13 Threads and before a
TUI write route is built on top of the hole. Three costs grow instead. (1) Every Thread
created accumulates metadata that only a forbidden hand-edit can correct, so the
hand-edit practice normalises against the project's own stated posture. (2) The inbound
link count grows with the corpus — it has reached 165 across 152 files with no cascade
to maintain it, and each new audit or doc that links a Thread adds to what a future
`thread rename` must rewrite correctly on its first run. (3)
`6gmshzxchdzc` plans Thread curation from the TUI; a TUI route cannot precede the
guarded write path it would call, so the missing verbs are a latent blocker on
in-flight epic-18 work rather than an isolated gap. Fixing `rename` later also means
writing the cascade against a larger link graph than the one `task rename` was
validated against.

**Proposed ADR amendment:** none required if the verbs ship as specified. If the
project instead decides Threads are deliberately write-once except for membership and
lifecycle, ADR-0006 needs an `## Amendments` entry retracting the §8 `set|edit|rename`
and `info` line, stating that Thread metadata is fixed by recreating the Thread, and
withdrawing §6's "rename-safe identity" claim for Threads — because that phrase
currently promises a cascade that does not exist. Either outcome is a human decision;
nothing was edited.

**Recommendation:** Ship guarded `thread set` for description/goal/target_date/tags routed through ValidateThreadDocument, then `thread rename` reusing RenameTask's cascade; or amend ADR-0006 §8 to retract the four verbs and state how Thread metadata is corrected instead.

#### M1. The id write gate's own justification is stale: no repair path needs the permission it grants, so --force is the only route to an unrepairable identity drift · **Status:** open

**File:** internal/store/frontmatter.go:92 | **Component:** store/frontmatter (surgical write gate)
**Effort:** XS · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** ADR-0003 §3/§4 (the id is the stable key and leads the filename) · https://deviq.com/principles/parse-dont-validate/ · https://djwong.org/docs/xfs-online-fsck-design.html

`validateIDUpdate` (`internal/store/frontmatter.go:96`) is the single gate every
frontmatter field write passes, including `task set --force`
(`frontmatter.go:113-119`). It checks the id's *format* only, and its comment states
why that is deliberate:

> "It deliberately checks the FORMAT only: writing a valid id that disagrees with the
> filename is a different problem, already reported by lint as drift with a rename
> remedy, and forbidding it here would break the fix pass that repairs exactly that."

The second clause of that justification no longer holds, and the first is weaker than
it reads. Both internal repair writers write an id that agrees with the filename *by
construction*: `backfillMissingID` (`internal/store/fix.go:344`, called at `:362`)
writes `fnID`, the id already leading the filename, and `repairInvalidID`
(`fix.go:224`, writing at `:275`) canonicalises the id in the filename and the field
together — its own comment says "both move together". Neither ever writes a value that
disagrees with its name. Requiring `id == filename id` at the gate would therefore
break no repair pass.

Meanwhile the "already reported by lint" half describes detection, not recovery.
Verified empirically against a scratch planning repo built with this binary:

- `task set probe-task --set id=000000000000` → refused, "unknown field".
- `task set probe-task --force --set id=000000000000` → **succeeded**, desyncing the
  declaration from `tasks/6gsr5qag2jpk-probe-task.md`.
- `lint` → correctly reports `frontmatter id "000000000000" disagrees with the filename
  id "6gsr5qag2jpk" — rename the file or fix the field`.
- `lint --fix` → `nothing to fix` / `could not auto-repair`, and the drift persists.

So the write path can manufacture exactly the damage class that repair deliberately
declines, and the gate permits it on the strength of a repair capability that is not
exercised. The remedy lint offers is not reachable through the tool either: `task set`
rejects `id` as unknown without `--force`, and there is no `task rename --id`, so the
only routes back are a hand-edit or a second `--force` write. The XFS online-fsck
design's position — check and repair are separate, and repair may legitimately decline
a class of damage — is the right posture, and it is precisely why the writer should not
create what repair won't fix.

This also sits badly with `--force`'s advertised contract. Its help text reads "allow
--set of a field tskflwctl doesn't know". `id` is not a field the tool doesn't know: it
is the canonical resolution key, the CAS/`resolveID` match target, and the flat
filename's lead (ADR-0003 §4). A user reaching for `--force` to add a custom key is not
asking for authority over identity. The same surface is closed for the fields the
project considered worth closing — `IsGraphOwnedTaskField` (`internal/domain/fields.go:89`)
holds `depends_on` and its three legacy aliases read-only *including* under `--force`,
and `task set --help` says so.

**Compounding cost:** XS to fix now, and the cost of deferral is paid in corpus damage
and in copies. ADR-0003 §3/§4 make the id the future database primary key; a corpus that
can carry declaration/filename disagreement means any later backend migration must
reconcile two identity sources per document rather than trust one, and each occurrence
needs a hand-fix because repair declines the class. The gate is also the pattern a sixth
entity kind will copy: epic 28 is adding routines, projects and ADRs, and each new
id-led noun inherits this gate's semantics for free. Narrowing one function now is
cheaper than auditing five write surfaces later.

**Proposed ADR amendment:** ADR-0003 is silent on whether a generic forced write may
change a declared `id`, and the codebase has answered twice and differently — research
refuses it as "the immutable key [that] must match the filename"
(`internal/domain/research.go:128`) while `validateIDUpdate` permits any well-formed
value. Propose an `## Amendments` entry to ADR-0003: "The `id` field is tool-owned.
A write that would leave the declared `id` disagreeing with the id leading the
document's filename is refused at the write path, including under `--force`; the two
are only ever changed together, by the rename and repair verbs that move both.
`--force` grants authority over unmodelled keys, never over the stable key." For a human
to accept or reject; nothing was edited.

**Recommendation:** Narrow validateIDUpdate to require that an id write equal the id leading the target filename, and update its comment; both internal repair writers already satisfy that by construction, so no repair path changes.

#### M2. The reserved on-disk schema marker is stamped by five writers, read by none, and settable to an arbitrary wrongly-typed value through --force on tasks and epics · **Status:** open

**File:** internal/domain/layout.go:16 | **Component:** domain/layout + per-kind field registries
**Effort:** XS · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** no accepted ADR governs this key (see below) · https://github.com/open-telemetry/opentelemetry-go/pull/4538

`FileSchemaVersion` (`internal/domain/layout.go:23`) is the on-disk frontmatter format
version, stamped as the first frontmatter key into every new planning document by five
writers: `internal/store/create.go:143`, `:313`, `:368`, `:459` and
`internal/store/threadcreation.go:161`. Its own doc comment states the design — "a
future format migration has an in-file signal to branch on" — and then concedes "The
loader ignores it (it is not a domain field) and surgical edits preserve it".

Nothing reads it. The only `schema` reader anywhere in the tree is
`internal/cli/task_dependency_repair.go:31`, which parses a Thread apply-plan document,
not a planning entity. Nothing lints it. `lint --fix` does not backfill it. Two of those
three are restated by the code's own comments; the third is restated in the sweep notes
of task `6g7f0tqgftg3`.

That much is already tracked. What is not is that the marker is **writable to anything**
through the generic set paths, so the signal is not merely inert but corruptible.
Verified against a scratch planning repo built with this binary:

- `task set probe-task --force --set schema=99` → succeeded; the file now carries
  `schema: "99"` — note the quotes: written as a YAML **string**, not the int the
  writers stamp.
- `epic set 01-probe-epic --force --set schema=99` → succeeded, same string coercion.
- `research set probe-research --force --set schema=99` → **refused**: "schema is the
  on-disk format version, managed by the tool".
- `lint` after both successful writes → silent about the schema key.

The asymmetry is the finding. `ProtectedResearchField`
(`internal/domain/research.go:123-133`) refuses `created`, `id`, `schema` and
`updated_at` and returns the *reason* for each, and its comment notes these are rejected
before the known-field set is consulted — which is why `--force` cannot reach them.
Task and epic have no equivalent: `schema` is simply absent from `taskFields`
(`internal/domain/fields.go:24-48`) and `knownEpicFields` (`internal/domain/epic.go:105`),
so it falls through the unknown-key gate that `--force` is designed to open. The
`knownEpicFields` comment compounds this by claiming it lists "every frontmatter key
tskflwctl itself reads or writes on an epic" while the tool writes both `schema`
(`create.go:313`) and `updated_at`, neither of which appears in the map.

Against external practice, a declared format version only earns its keep if the reader
participates. The OpenTelemetry Go review of exactly this question quotes its file-format
spec as saying consumers "SHOULD NOT attempt to interpret the schema file if the MINOR
version number is higher", and the implementation documents that "If r contains a Schema
with a file format version higher than FileFormat, an error will be returned". That
thread's own disagreement is instructive and worth reading before deciding taskflow's
rule — MrAlias argues strict single-version parsing "[is not] the user experience we
want to provide", and pellared proposes accepting a newer file that uses no unknown
fields. The disagreement is about *which* refusal rule, never about whether the reader
reads the field.

**Compounding cost:** the deferral is defensible — no incompatible bump has happened, so
there is nothing to refuse yet — but the asymmetry is not, and it is the half that gets
more expensive. A format marker is worth exactly what it is trusted for at the moment of
the first bump. Task `6g7f0tqgftg3`'s own corpus inventory shows why the clock matters:
the 94 pre-marker documents are a closed, non-growing backlog, so every document written
since the marker shipped carries it and the inventory is currently reliable. A corpus that
can also carry forced-set or wrongly-typed values makes that inventory untrustworthy in a
way counting cannot detect, and the migration planned in that task has to branch on it.
Closing the write side is XS and independent of the read policy; it makes the eventual
decision cheaper rather than pre-empting it. Released binaries already in the wild will
never check the field at all, which is a further argument for the refusal landing before,
not after, a bump.

**Proposed ADR amendment:** none yet — this belongs to the unstarted policy ADR
`6fkkz41cax80`, not to ADR-0003 directly. Recorded here as an ADR gap because ADR-0003
decides paradigm, organizing key, identity, filename, reference scheme and migration, and
never mentions a per-document format-version key at all. Five writers stamp a key that no
accepted decision describes. The input this audit adds for that ADR: whichever way the
read policy goes, the key should be tool-owned on the write side first, matching the
reason-returning refusal research already implements.

**Recommendation:** Add `schema` to the protected-key refusal for task and epic the way ProtectedResearchField already does, so the marker is trustworthy when 6g7f0tqgftg3 gives it a reader; the read/branch policy itself stays with that task.

#### M3. Five entity kinds answer "which frontmatter fields exist" five different ways; only research derives from the registry, and two kinds have no drift guard · **Status:** open

**File:** internal/domain/fields.go:24 | **Component:** domain (per-kind field registries)
**Effort:** M · **Urgency:** soon

**Class:** ADR gap
**Anchored to:** no accepted ADR governs the field contract (epic 26's ADR `6fkkz41cax80` is unstarted) · https://github.com/jackchuka/mdschema · https://documentation.platformos.com/developer-guide/platformos-check/checks/valid-frontmatter

The *authoring* side of the data model is genuinely well factored: one registry
(`internal/domain/entity.go:52`) carries each kind's directory, authoring fields,
conventions, templates and placeholders, and `schema <kind>` derives its output from it
rather than transcribing inventories. The *validating* side has not followed, and the
result is five bespoke implementations of one question:

- **task** — a declarative `taskFields` table with the known/int/list/date sets derived
  from it (`internal/domain/fields.go:24`, `:66`). Protection: `ValidateField`
  (`internal/domain/validate.go:194`) refuses `status` and `updated_at`.
- **epic** — a hand-written five-key map (`internal/domain/epic.go:105`) whose comment
  claims it is "every frontmatter key tskflwctl itself reads or writes on an epic" while
  the tool also writes `schema` (`internal/store/create.go:313`) and `updated_at`.
  Protection: `ValidateEpicField` refuses `status`.
- **audit** — no field registry and no `set` verb; `area`/`date` are immutable identity.
  Defensible by design.
- **research** — the known-field set **derived from** `AuthoringFields("research")` plus
  its one declared stamp (`internal/domain/research.go:65`), with
  `ProtectedResearchField` returning a *reason* per refused key (`:123`) and
  recognized-vs-settable kept deliberately distinct (`:99`).
- **thread** — no field registry, no protection, and no write verb at all (H1).

Only research realises the charter its own comment states — "a noun's fields ride one
registry (the epic-28 charter)". Nothing reconciles task's `taskFields` against the task
Descriptor's typed `AuthoringFields`; `TestTaskFieldsMatchStruct`
(`internal/domain/schema_test.go:80`) pins the table against the `Task` struct instead,
and `TestTaskAuthoringFieldsMatchRegistry` (`:13`) checks the documented fields are known
but not the converse.

The drift guards are themselves asymmetric. `schema_test.go` covers task (`:13`), epic
(`:34`) and audit (`:56`). **Thread and research have none** — no test asserts that the
five fields the thread descriptor documents (`entity.go:164-170`) correspond to real
`domain.Thread` yaml tags, or that research's do. `TestAuthoringFields_Kinds` (`:111`)
and `entity_test.go:35-40` only check resolvability and matching length, which cannot
catch a renamed or mistyped field name. For thread this is currently harmless because no
write path consumes the set; it stops being harmless the moment H1's `thread set` ships,
and that is the wrong order to discover it.

Against external practice, the pattern in comparable tools is one schema per document
type reused by both validation and authoring. mdschema states "A schema has one
`structure` tree plus optional document-wide rules" and that entries "are checked against
the declared `type`/`format` when the schema loads"; platformOS's `valid_frontmatter`
check flags keys "that are not part of the file type's frontmatter schema" — a
closed-by-default key set per type. taskflow is **partial** on this: the authoring
declaration exists, the validating one is per-kind.

The project constraint that legitimately applies is tolerant reads over a legacy corpus.
A closed-by-default key set cannot simply be switched on when 18 research docs carry a
vestigial `status:` that is deliberately undeclared and unlinted, and when unknown-field
preservation is a stated guarantee. But that argues about *where strictness lands*, not
about how many registries express it — and it is the reason this finding proposes
derivation and guards, not strictness.

**Compounding cost:** the cost is paid per new noun, and new nouns are in flight. Epic 28
is at 63% and chartered for "routines, projects, ADRs"; task `6g7f0tqgftg3`'s sweep notes
already record that `planning/adrs/` holds 8 documents waiting to become a registered
kind. Each new noun currently costs a bespoke known-field set, a bespoke protection
function, and a bespoke drift guard — three artefacts with three existing shapes to copy
from and no stated preference among them, which is how epic.go's map came to misdescribe
itself. Every noun registered before the pattern is settled adds one more call site to
the eventual consolidation and one more chance to copy the weakest of the three. The
derivation and the two missing guards are ordinary groundwork that does not need the ADR
to land first, and doing them first makes the ADR's scope smaller.

**Proposed ADR amendment:** none. Epic 26's charter already owns this decision
("replacing today's split between hand-written LintTask checks and the envelope JSON
schema") and its ADR task `6fkkz41cax80` is `next-up` and unstarted. This finding is
evidence for that ADR, not a competing decision: the inventory above is what the ADR has
to reconcile, and the research kind is the candidate reference implementation. The
strictness and field-registry policy questions epic 26 lists as open remain open.

**Recommendation:** As groundwork the ADR does not need to wait for: derive each kind's known-field set from Descriptor.AuthoringFields plus its declared tool-managed stamps the way research.go:65 does, give thread and research the AuthoringFields drift guard task/epic/audit have, and correct the knownEpicFields comment.

## What audited clean

- **Verify-after-encode on the block-scalar write path.** `frontmatter_preserve.go:96-98`
  decodes both sides and refuses rather than persisting a changed value — a hard check
  where `go.yaml.in/yaml/v3`'s own docs promise only that "an effort is made".
- **Fidelity on the *un*gated write path, measured not assumed.** Head/inline/foot
  comments, two quoting styles, a flow mapping, an empty value and an anchor/alias pair
  all survived a `task set` byte-intact in original key order, with no block scalar
  present to trigger the stronger tier.
- **One id choke point, and it holds under `--force`.** `frontmatter.go:117` gates every
  field write; `--set id=NOTANID` and `--set id=iiiiiiiiiiii` are refused with
  position-accurate messages — ADR-0003 §3's lowercase-Crockford identity, enforced once.
- **Repair converges on the filename rather than guessing.** Both id repair paths write a
  value that agrees with the name by construction (`fix.go:344`, `:273`) — the XFS
  design's "targeted repair" posture, and the reason declining id *drift* is principled.
- **`lint` detects the drift the write path can create.** The disagreement is reported
  with a concrete two-option remedy, so the defect is never silent — ADR-0007 §4's
  "reads stay tolerant so `lint` can REPORT malformed data".
- **Research is the reference implementation for a noun's field contract.** Known fields
  derived from the registry (`research.go:65`), protected fields that return the *reason*
  (`research.go:123`), and recognized-vs-settable kept distinct (`research.go:99`) — the
  epic-28 charter, actually realised.
- **Graph-owned fields are read-only through generic paths, including `--force`.**
  `IsGraphOwnedTaskField` (`fields.go:89`) covers the canonical field and all three legacy
  aliases, and `task set --help` states the `--force` exclusion explicitly — ADR-0006 §2.
- **`ValidateThreadDocument` enforces ADR-0006 §6 at the document level.**
  `thread.go:72` pins the duplicate-free, stable-ID-only, sorted `tasks` set plus required
  description/goal — the strongest document validator in the tree.
- **Derived-not-transcribed authoring prose.** The task and audit descriptors build their
  vocabulary sentences from `CriterionSuffixStates()` and `FindingStatuses()`
  (`entity.go:77`, `:126`) with the drift incident that motivated it named in the comment —
  ADR-0007 §2's "words that mean the same thing in two places are spelled once".
- **Three of five kinds have an AuthoringFields drift guard.** `schema_test.go:13`, `:34`,
  `:56` pin task, epic and audit against their registries or structs. Named here as the
  pattern M3 asks to complete, not as a gap in itself.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ H1 · open — tskflwctl task new "Give Threads the guarded metadata write verbs ADR-0006 §8 specifies" --epic 30-threads-and-task-dependency-graphs --tags thread,cli,store,adr-drift
- ○ M1 · open — tskflwctl task new "Require an id write to match the filename id at the frontmatter write gate" --epic 24-data-model-evolution-stable-key-storage-read-model-content-occ --tags store,frontmatter,identity,validation
- ○ M2 · open — tskflwctl task new "Make the reserved schema key tool-owned on the write side" --epic 26-frontmatter-schema-declared-validation-contract --tags frontmatter,schema,validation (PARTIAL overlap: read/branch policy is 6g7f0tqgftg3's; this is the write-side refusal it does not cover)
- ○ M3 · open — tskflwctl task new "Derive every kind's known-field set from the registry and add the missing thread/research drift guards" --epic 26-frontmatter-schema-declared-validation-contract --tags domain,frontmatter,registry (PARTIAL overlap: policy is ADR task 6fkkz41cax80's; this is the pre-ADR groundwork)

## Related-task observations (propose-only)

- Possibly under-scoped: `planning/tasks/6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md`
  — its scope covers read behaviour and whether a writer may mutate a document of an
  unsupported format, but not whether the `schema` key is itself writable through generic
  `--set --force`. M2's probe wrote `schema: "99"` (a string, not an int) to both a task
  and an epic. Worth one added acceptance criterion rather than a separate task.
- Scope note: `planning/tasks/6gmshzxchdzc-manage-thread-membership-and-lifecycle-from-the-tui.md`
  covers Thread membership and lifecycle curation from the TUI. It does not cover the
  missing CLI metadata verbs in H1, and a TUI route cannot precede the guarded write path
  it would call. Sequencing observation only; no change proposed.
