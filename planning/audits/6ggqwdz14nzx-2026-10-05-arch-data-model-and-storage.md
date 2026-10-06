---
schema: 1
id: 6ggqwdz14nzx
bucket: open
area: arch-data-model-and-storage
date: "2026-10-05"
updated_at: "2026-10-05"
---

# Audit: arch-data-model-and-storage — 2026-10-05

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

Routine: `weekly-architecture-audit` · lens `data-model-and-storage` · ISO week
`2026-W41` (mod 4 = 1).
ADRs consulted: ADR-0003 (§2, §3, §4, §5, §6 + all four amendments), ADR-0006 (§2, §3, §6, §9),
ADR-0007 (§4) · Files read: `internal/store/frontmatter.go`, `internal/store/create.go`,
`internal/store/fix.go`, `internal/store/threadcreation.go`, `internal/domain/fields.go`,
`internal/domain/entity.go`, `internal/domain/epic.go`, `internal/domain/research.go`,
`internal/domain/thread.go`, `internal/domain/lint.go`, `internal/id/id.go`,
`internal/core/service.go`, `internal/core/thread_creation.go`, `internal/tools/{flatmigrate,researchmigrate,wikimigrate}/main.go`
· Sources cited: 4

#### H1. A surgical frontmatter write silently appends one newline per write to a folded block scalar's decoded value · **Status:** tracked by 6g1dhhk6721x

**File:** internal/store/frontmatter.go:273 | **Component:** store/frontmatter (surgical write path)
**Effort:** S · **Urgency:** soon

**Class:** implementation drift
**Anchored to:** ADR-0003 §2 (frontmatter is the sole authority for mutable state once the directory is retired) · CLAUDE.md non-negotiable "Frontmatter is edited **surgically** — preserve unknown fields, comments, and key order" · https://github.com/yaml/go-yaml/issues/337 (open) · https://pkg.go.dev/go.yaml.in/yaml/v3

`updateFrontmatter` (`internal/store/frontmatter.go:118`) reassembles every field write through `assembleFile` (`:273`), which re-encodes the whole frontmatter mapping with `yaml.NewEncoder`. For a **folded** block scalar (`>` / `>-`) whose body contains **more-indented** lines, that encode inserts one extra blank line inside the scalar each time, and the next parse reads that blank as a real newline in the decoded value. The growth is unbounded and silent: it compounds once per write, forever, on a field the writer never intended to touch.

Reproduced on HEAD through the product CLI, in an isolated planning tree, on a task carrying:

```yaml
rationale: >
  Heading:

    * first item
    * second item
```

Five `tskflwctl task set <slug> --set tier=N` calls later, the file holds five blank lines inside that scalar. Decoded through the store's own reader and `go.yaml.in/yaml/v3`:

```
before: rationale decodes to 41 bytes: "Heading:\n\n  * first item\n  * second item\n"
after:  rationale decodes to 46 bytes: "Heading:\n\n\n\n\n\n\n  * first item\n  * second item\n"
```

One `\n` per write. This is value corruption, not formatting churn — the string an agent or a future reader gets back is not the string that was authored, and nothing reports it. `lint` is clean on the corrupted file, because the document still parses and the field is unknown to the registry, so no validator has an opinion about it.

Two scoping facts, stated so the severity is not overread:

- **No document in the live corpus triggers it today.** The only folded scalar in any planning frontmatter is `purpose: >-` in `planning/research/6fe4my001bdk-adrs-and-projects-format-design.md`, which is uniformly indented; three `research set` round-trips against a copy reflowed it onto one line (the known cosmetic defect) and left the decoded value byte-identical. Literal (`|`) scalars are stable across round-trips, matching the upstream issue's own finding.
- **Nothing the tool writes emits a folded scalar**, so the exposure is hand-authored or migration-written frontmatter — which is precisely the content the surgical-edit promise exists to protect, and precisely what `researchmigrate` produces (see `6g1dp77y0g5n`, still open, for its block-style-YAML and CRLF defects).

**Compounding cost:** the guard is one conditional at one choke point today, because `updateFrontmatter`/`assembleFile` is the single path every writer already shares — tasks, epics, research, lifecycle verbs, `lint --fix`, and the guarded graph/Thread mutations. That is as cheap as this fix will ever be. Two planned changes make it structurally more expensive: epic 26's declared frontmatter schema will admit multi-line text fields (a recommendation, a rationale, a goal) as *known* fields, at which point folded scalars stop being an accident of hand-authoring and become part of the contract; and every new entity kind registered under epic 28 inherits the same writer. Cost of deferring is also measured in undetectability — a value that has silently grown newlines for months is indistinguishable from an authored one without reading git history, and there is no lint that would ever flag it.

**Minimum fix:** at the one choke point, leave a scalar node untouched when its value is unchanged (re-emit only the nodes the update actually names), or fail closed with `ErrValidation` naming the field when a folded scalar is present. Either satisfies the first acceptance criterion already written on `6g1dhhk6721x`. Pin it with a round-trip regression fixture covering `>`, `>-`, `|`, and a more-indented body, and assert on the *decoded* value rather than the file text — the text-level assertion is what let this hide behind the reflow.

**Follow-up:** whether to carry a vendored patch or a version pin for yaml/go-yaml#337 while it is open is a dependency-policy call, not part of this fix; the choke-point guard is correct regardless of what upstream does.

**Recommendation:** At updateFrontmatter/assembleFile, re-emit only the nodes an update names (or fail closed on a folded scalar), pinned by a decoded-value round-trip fixture covering >, >-, | and a more-indented body.

**Resolution:** Independently reproduced during boundary closeout: five
updateFrontmatter tier writes changed a more-indented folded scalar from 41 to
46 decoded bytes. Existing task now covers decoded-value fidelity, is high
priority/tier 2, and remains ready in epic 21 outside the adapter-neutral
migration. Not fixed; prioritize before the next release.

## Executive summary

This lens was last audited four weeks ago (`2026-09-07-arch-data-model-and-storage`,
now closed). Four of that run's five findings were ADR gaps about identity scope, the
reserved `schema:` marker, and registry shape; all four are still correctly `tracked by`
live design tasks, and none has drifted further — so this run deliberately does **not**
re-report them. What it adds is one new, reproduced defect in the write path that ADR-0003
§2 makes load-bearing: **a surgical frontmatter write silently appends one newline per
write to the decoded value of a folded block scalar elsewhere in the same file**, without
bound. Measured through the real reader: 41 bytes → 46 bytes after five `task set` calls
(H1). The repo already has a task for the *cosmetic* half of this behaviour
(`6g1dhhk6721x`), and that task's load-bearing premise — "**No data is lost**" — is false
for folded scalars carrying more-indented content, which is why it sits at priority `low`,
tier 4. Read H1 first, then the Related-task observation about that triage. Everything else
in this lens audited clean, including a verified 3-round-trip fidelity check of comments,
key order, unknown fields, and literal (`|`) scalars.

## State of the architecture (this lens)

**One choke point for every field write.** `updateFrontmatter`
(`internal/store/frontmatter.go:118`) is the single path through which every frontmatter
field mutation passes — `task set`, `epic set`, `research set`, `lint --fix`, the lifecycle
verbs, and the guarded graph/Thread mutations. It validates the `id` before anything
touches disk (`validateIDUpdate`, `frontmatter.go:93`), with the comment explicitly naming
`task set --force` as the bypass it closes. It parses the frontmatter into a `yaml.Node`,
mutates that node, and reassembles through `assembleFile` (`frontmatter.go:273`). The body
is carried verbatim; `detectLineEnding` (`frontmatter.go:82`) re-emits the frontmatter in
the file's own ending so a CRLF file does not come back mixed. `setMapNode`
(`frontmatter.go:357`) replaces a value node in place and transplants the old node's
head/line/foot comments onto the replacement unless the new node carries its own. This is
a genuinely careful surgical-edit implementation, and it is the right shape: one place to
guard, one place to test.

**Identity.** `internal/id/id.go` mints a 60-bit id as 43 time bits + 17 random bits
(ADR-0003 §3 as amended 2026-07-02). `New()` (`id.go:69`) draws from `crypto/rand` and is
strictly monotonic within a process; `NewAt(unixMilli)` (`id.go:105`) is deliberately
stateless and documents its caller obligation in the code — "callers that mint many ids at
one timestamp (migrations backfilling historical files) must dedupe: a shared-ms collision
has probability 1/2^randBits". Day-precision minting is therefore the real collision
surface, exactly as ADR-0003's amendment anticipated. Empirically the live corpus holds 586
entity ids (tasks + audits + research + threads) across 388 distinct 43-bit time buckets,
all unique; the largest single-day bucket holds 25 ids drawing from one 17-bit space, and
summing over the observed distribution gives 1,323 same-bucket pairs, i.e. ≈1.0% prior
probability that this corpus would already contain a day-minted collision. It does not.

**Identity enforcement is per-kind, with one cross-kind exception.** Creation checks are
same-kind: `ensureCandidateIDUnique` is called for tasks (`create.go:218`), audits
(`create.go:327`), and research (`create.go:375`), each against its own directory's
candidates. Thread creation is the only path that looks across kinds — `ValidateThreadCreationPlan`
rejects an id already owned by a task *and* by another thread (`internal/core/thread_creation.go:145`,
`:149`). The repair path checks one pair: `crossKindIdentityOwner` (`internal/store/fix.go:237`)
switches on `tasksDir`/`threadsDir` and returns nothing for every other directory. Lint
mirrors that: `DuplicateIDIssues` (`internal/domain/lint.go:358`) is invoked once per kind
for research (`service.go:892`), audits (`:925`), and threads (`:937`); tasks get duplicate
detection separately through the graph (`ProblemDuplicateTaskID`,
`internal/core/dependency_graph.go:36`); epics get `DuplicateEpicNNIssues`; and the only
cross-kind lint is the task↔Thread pair (`service.go:840`, `:951`). So five kinds, three
same-kind mechanisms, and one of ten kind-pairs covered cross-kind.

**Creation ordering is correct.** `createEntityFile` (`create.go:71`) takes the repository
write lock *first* and runs the identity scan inside that critical section, with a doc
comment that names the race it exists to prevent: "This ordering prevents a future entity
from accidentally recreating the audit/research race by scanning identity first and
acquiring the write lock later." Dry-run runs the same preparation and path checks without
creating the root or taking the lock. Thread creation goes further, re-reading the task
graph and Thread set after planning and refusing on any change
(`threadcreation.go:87`–`:97`).

**Field model.** `taskFields` (`internal/domain/fields.go:24`) is a typed table and the
single source of truth for task fields; the int/list/date/known sets are all derived from
it. `knownResearchFields` (`internal/domain/research.go:65`) is *derived from the entity
descriptor* — the strongest pattern in the tree. `knownEpicFields`
(`internal/domain/epic.go:105`) is a hand-written map literal, and list-ness for both epic
and research is the string literal `f == "tags"`. Threads and audits have no known-field
map, which is principled: neither has a generic `set` verb. Per-entity metadata otherwise
lives in one registry (`internal/domain/entity.go:57`), which is what makes a new kind a
table row rather than a scattered `switch`.

**The reserved on-disk marker.** `FileSchemaVersion = 1` (`internal/domain/layout.go:23`)
is stamped as the first frontmatter key by all five create paths (`create.go:144`, `:280`,
`:335`, `:426`; `threadcreation.go:161`). Its own doc comment states the design: "The
loader ignores it (it is not a domain field) and surgical edits preserve it". Corpus
coverage is partial by age — 86 of 403 tasks and 5 of 15 epics carry no `schema:` key.

**Migration.** Three self-contained migration programs remain in-tree —
`internal/tools/flatmigrate` (433 lines), `researchmigrate` (523), `wikimigrate` (261) —
each opening with "Throwaway by design" in its own doc comment, each dry-run-by-default and
refusing a dirty git tree. `researchmigrate` carries its own frontmatter writer
(`main.go:289`) rather than the store's.

Then, explicitly: **which ADR governs this?** ADR-0003 governs the layout, identity form,
filename, and reference scheme, and ADR-0003 §2 is what makes frontmatter fidelity
load-bearing rather than cosmetic. ADR-0006 §2/§6 govern `depends_on` and Thread
membership. No ADR governs the *declared field contract* — that is epic 26's open subject,
and the 2026-09-07 run already recorded it as an ADR gap.

## ADR reconciliation

### ADR-0003 §2 — status/bucket move into frontmatter, and the directory stops being the authority

**Quoted:** "The directory stops being the authority for mutable state. `status` (tasks)
and `bucket` (audits) become frontmatter source-of-truth; the tree is organized only by
*stable* things."

**Implementation:** follows — and this is the clause that makes H1 a defect rather than a
diff annoyance. With the directory retired, frontmatter is the *only* authority for mutable
state, so the fidelity of the frontmatter write path is the fidelity of the data model.
Status/bucket live in frontmatter throughout (`domain/entity.go:20` is the one stale
reference, below), the flat layout holds for `tasks/`, `audits/`, `research/`, `threads/`,
and `lint` reports rather than relocates a bad status.

**Class if divergent:** implementation drift → H1 (the write path, not the layout).

### ADR-0003 §3 — a 12-char time-sortable id, collision-checked at creation

**Quoted:** "**Creation collision-checks** the id and regenerates on the
astronomically-rare intra-millisecond clash (or generates monotonically within a tick)."

**Implementation:** follows, within the scope the clause states. `New()` is monotonic
within a process (`id.go:69`), creation collision-checks same-kind
(`create.go:218`/`:327`/`:375`), and `NewAt`'s day-precision caller obligation is documented
at the mint site (`id.go:99`–`:105`). The clause says nothing about whether the id namespace
is planning-space-wide or per-kind, which is the gap the 2026-09-07 run recorded as its M1
and which `6g7wsr8yvt1a` now names explicitly as "its unresolved planning-space-wide versus
per-kind stable-ID namespace decision". **Not re-reported here** — it is tracked, the
destination task is live, and nothing about it has moved in the wrong direction since.

**Class if divergent:** ADR gap — already recorded and tracked (`6fkkz41cax80`).

### ADR-0003 §4 — id-led, flat filenames, one directory per entity

**Quoted:** "`<id>-<slug>.md`, in one flat directory per entity (`tasks/`, `epics/`,
`audits/`); no status, bucket, or epic subdirectories."

**Implementation:** follows. Verified against the live tree: every scanned directory is
flat, and the 2026-07-04 carveout amendment holds — a non-id-led `.md` is a `FileProblem`
and is excluded from resolution candidacy, with `README.md` silently carved.

### ADR-0003 §6 — migration is a one-time throwaway script, not a CLI command

**Quoted:** "**Run once per repo, then discarded** — it need not be general, configurable,
or supported."

**Implementation:** drift in the letter, already recorded in substance. Three throwaway
tools are still in-tree and still compiled and tested by `go build ./...` / `go test ./...`.
The 2026-09-07 run recorded this as L1 ("ADR-0003 §6's throwaway-migration posture is the
project's only accepted migration policy, and open work now needs a different one"),
`tracked by 6fkkz41cax80`. **Not re-reported.** One factual update for the record: the
`researchmigrate` defects are still open as `6g1dp77y0g5n` (`ready-to-start`) — "emits a
duplicate `tags` key on block-style YAML (unparseable file), mishandles CRLF frontmatter,
and can mint an id already on disk" — so the retained tool is still a second, weaker
frontmatter writer outside the `updateFrontmatter` choke point.

**Class if divergent:** ADR pressure — already recorded and tracked.

### ADR-0006 §2 — `depends_on` is graph-owned; generic set/edit paths may not touch it

**Quoted:** "Generic set/edit paths treat both `depends_on` and the legacy dependency
fields as graph-owned, including under `--force`; guarded dependency commands are the only
product write path."

**Implementation:** follows. `IsGraphOwnedTaskField` (`fields.go:89`) names `depends_on`
plus the three legacy aliases, and the repair counterpart
(`updateDependencySourceEdits`, `frontmatter.go:155`) removes only exact selected YAML
occurrences while retaining every unselected sequence node, key, comment, and body — it
refuses to clear a whole legacy field or manufacture a desired set.

### ADR-0006 §6 — Thread documents own metadata and membership only; `tasks` is a sorted, duplicate-free set

**Quoted:** "`tasks` is a duplicate-free semantic set serialized in stable ID order; list
position carries no execution meaning."

**Implementation:** follows. `ValidateThreadDocument` (`domain/thread.go:72`) enforces
id validity, no duplicates, and sortedness, and rejects a non-stable member id. Thread
frontmatter carries no dependency edges.

### No ADR governs the declared frontmatter field contract

Stated explicitly, as the template asks. `schema --json` publishes `task_fields` (23
entries, typed `{name,type}`), `research_fields` (4, typed), and `epic_fields` (5, bare
strings, and missing `updated_at`, which the `Epic` struct persists and the tool stamps);
it publishes no `thread_fields` and no `audit_fields`, while `kinds` lists all five. This
is epic 26's subject and was recorded by the 2026-09-07 run as M3 ("Three entity kinds have
a known-field registry, in three different shapes"), `tracked by 6fkkz41cax80`. **Not
re-reported.** Noted here only because the `epic_fields`/`thread_fields` asymmetry is the
concrete, agent-visible shape of that gap and is worth the destination task having on file.

## Best-practice comparison

### A `yaml.Node` round-trip does not preserve the original text, by design

**Source:** https://pkg.go.dev/go.yaml.in/yaml/v3
**Guidance:** the library's own documentation for `Node` states that "the content when
re-encoded will not have its original textual representation preserved", while "an effort
is made to render the data plesantly, and to preserve comments near the data they describe".
**This codebase:** partial — and the split is exactly where the docs put it. Measured by
writing a probe task carrying a head comment, a trailing line comment, a comment on a list
item, a folded scalar, a literal scalar, and three unknown fields, then issuing three
`task set` calls: head comment, line comments, list-item comment, key order, all unknown
fields, and the literal (`|`) scalar survived byte-identically; the whitespace run before a
trailing `#` normalized to a single space; and the folded scalar was reflowed onto one line.
**Justified?** Partly. Taking the documented formatting loss in exchange for a dependency-light
surgical editor is a reasonable trade for a local-first tool whose own writers only ever emit
single-line scalars. What is not justified is the *semantic* loss in H1, which the same
documentation does not license and which the library tracks as a bug.

### Folded block scalars accumulate a newline per `Node` round-trip (open upstream defect)

**Source:** https://github.com/yaml/go-yaml/issues/337
**Guidance:** the issue reports that parsing a document containing a folded block scalar
(`>`) into a `yaml.Node` and re-emitting it via `yaml.Encoder` inserts an extra blank line
inside the scalar body, which the next parse captures as a real newline — "round 0: decoded
(41 bytes) … round 1: decoded (42 bytes)" — and records that literal (`|`) scalars remain
stable, establishing it as a folded-style-specific defect. The issue is **open** and names
`gopkg.in/yaml.v3` and `go.yaml.in/yaml/v4 v4.0.0-rc.4` as affected.
**This codebase:** diverges — reproduced on HEAD through the product CLI, not just the
library. `taskflow` imports `go.yaml.in/yaml/v3 v3.0.5` directly (`go.mod:23`) and carries
`go.yaml.in/yaml/v4 v4.0.0-rc.2` indirectly (`go.mod:66`). See H1 for the measurement.
**Justified?** No. The guard belongs at the one choke point the project already maintains
for exactly this class of promise.

### Short random id tails need an explicit uniqueness strategy, not just entropy

**Source:** https://github.com/ulid/spec
**Guidance:** the ULID specification pairs its timestamp with **80 bits** of randomness —
"1.21e+24 unique ULIDs per millisecond" — and, for the same-timestamp case, specifies that
"if the same millisecond is detected, the `random` component is incremented by 1 bit in the
least significant bit position (with carrying)", failing generation rather than risking a
repeat.
**This codebase:** follows, with a deliberate and documented reduction. ADR-0003 §3 chose 17
random bits over ULID's 80 as the explicit price of a 12-character id, and discharged the
difference the way the spec recommends — monotonic-within-tick for the live path
(`id.go:69`) plus a creation collision check. `NewAt` then names the residual obligation at
the mint site for the day-precision case (`id.go:99`).
**Justified?** Yes for the entropy reduction: at planning scale the measured exposure is
≈1.0% over the whole corpus history, and the monotonic live path removes the common case
entirely. The part the spec does not cover — whether the namespace being protected is
per-kind or planning-space-wide — is the tracked ADR gap above, not a divergence from this
source.

### Entropy choices and concurrency safety must be explicit at the mint site

**Source:** https://github.com/oklog/ulid
**Guidance:** the reference Go implementation stresses that "care should be taken when
providing a source of entropy", that `math/rand.Rand` "is not safe for concurrent use by
multiple goroutines", that "security-sensitive use cases should always use
cryptographically secure entropy provided by `crypto/rand`", and that same-millisecond
ordering requires an explicit monotonic entropy source.
**This codebase:** follows. `internal/id/id.go` uses `crypto/rand` (`id.go:27`), treats an
entropy failure as unrecoverable for an identity mint rather than degrading, serializes the
monotonic counter under a lock while deliberately keeping the `crypto/rand` read outside it
(`id.go:71`), and pins the bit budget with compile-time assertions that fail on both
overflow and underflow (`id.go:51`–`:52`).
**Justified?** n/a — no divergence.

## Tensions and trade-offs

**Surgical editing versus a dependency-light stack.** The project's non-negotiable
("frontmatter is edited surgically — preserve unknown fields, comments, and key order") is
a stronger promise than `gopkg.in/yaml.v3`'s `Node` API actually offers; the library
documents that it does not preserve original text. Meeting the promise exactly would mean
a text-level frontmatter editor (splice the changed line, never re-emit the block) or a
round-trip-preserving YAML library, both of which cost more than the current ~450-line
file. The defensible middle — and what H1 asks for — is to keep the formatting loss and
close the *semantic* loss, because a value that changes meaning is a different kind of
promise than a value that changes shape.

**A reserved marker versus a hard cutover.** ADR-0003 §6 chose "hard cutover, no
coexistence", which is the right call for two known repos; the `schema:` key then exists to
give a future migration "an in-file signal to branch on". Those two decisions pull against
each other: the marker only pays off in a world where coexistence is read, and the ADR
chose a world where it is not. That tension is `6g7f0tqgftg3`'s to resolve and is already
tracked; it is recorded here because this lens is where it shows.

**Per-kind enforcement versus a global primary key.** ADR-0003 §3 describes the id as
"the DB primary key", which reads as a global namespace, while every enforcement mechanism
in the tree except Thread creation is scoped to one kind. Each per-kind check is individually
cheap and correct; a global index is the cheaper thing to own once epic 28 adds more nouns.
Again tracked, not re-reported.


## What audited clean

- **The single frontmatter write choke point.** Every field mutation — including
  `task set --force` — passes `validateIDUpdate` before touching disk
  (`frontmatter.go:93`, `:118`), with the comment naming the force-bypass it was built to
  close. ADR-0003 §3's "the id is the canonical key" has exactly one gate.
- **Surgical fidelity, measured rather than assumed.** Three `task set` round-trips over a
  probe carrying a head comment, two trailing comments, three unknown fields, and a literal
  scalar left all of them byte-identical and in their original key order — the CLAUDE.md
  non-negotiable, verified empirically rather than by reading the code.
- **CRLF handling.** `detectLineEnding` (`frontmatter.go:82`) plus `replaceBodyWith`'s body
  re-encoding mean a CRLF file does not come back with a mixed-ending diff — the same class
  of promise as H1, correctly kept.
- **Creation takes the lock before scanning identity.** `createEntityFile` (`create.go:71`)
  documents and enforces the ordering that prevents the audit/research
  scan-then-lock race from recurring in a future entity kind.
- **Thread creation is the strictest create path in the tree.** It validates against both
  tasks and threads (`core/thread_creation.go:145`, `:149`), then re-reads the graph and
  Thread set after planning and refuses on any change (`threadcreation.go:87`–`:97`) —
  ADR-0006 §2's "one planning-repository mutation guard", honoured.
- **Graph-owned fields are read-only through generic paths.** `IsGraphOwnedTaskField`
  (`fields.go:89`) covers the canonical field and all three legacy aliases, and the repair
  writer removes only exact selected occurrences (`frontmatter.go:155`) — ADR-0006 §2,
  including its `--force` clause.
- **Thread membership invariants are enforced at the document level.**
  `ValidateThreadDocument` (`domain/thread.go:72`) enforces ADR-0006 §6's duplicate-free,
  stable-ID-only, sorted `tasks` set.
- **`id.NewAt` names its own caller obligation.** The day-precision collision probability
  and the "callers must dedupe" requirement are documented at the mint site
  (`id.go:99`–`:105`) rather than left for a reader to derive — which is why the research
  create path has a dedupe check at all (`create.go:375`).
- **Research derives its field registry from the entity descriptor.**
  `knownResearchFields` (`research.go:65`) is built from `AuthoringFields("research")`, so
  it cannot drift from the descriptor. This is the shape epic 26 is heading toward, already
  working in one kind.
- **Last cycle's gaps are tracked, not drifting.** All four open findings from
  `2026-09-07-arch-data-model-and-storage` still carry live destinations
  (`6fkkz41cax80` next-up, `6g7f0tqgftg3` ready-to-start, reconciled by `6g7wsr8yvt1a`),
  and its H1 (duplicate audit id) is `fixed`. Re-auditing the same lens four weeks later
  found no regression in any of them.

## Related-task observations (propose-only)

- **Possibly under-triaged:** `planning/tasks/6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md`
  — priority `low`, tier 4. Its triage rests on a stated premise, "**No data is lost** — a
  `>-` folded scalar joins its lines with spaces, so the value is byte-identical (verified
  by round-tripping both versions through the YAML parser and comparing)". That holds for
  the `>-` case it reproduces, and is false for a `>` scalar containing more-indented
  content: H1 measures the decoded value growing 41 → 46 bytes over five writes. Its
  2026-09-27 sweep note re-confirmed the reflow but not this case. The task's first
  acceptance criterion ("byte-identical, including its wrap width and its chomping
  indicator") would incidentally cover H1 if implemented as written, so the gap is in the
  severity assessment rather than the scope. Severity and priority are the author's call —
  nothing was changed.
- **Scope note, not a conflict:** `planning/tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md`
  (`ready-to-start`, tier 2) already owns reconciling identity, schema, locking, and
  atomic-write foundations into one sequence, and explicitly names the unresolved
  per-kind-versus-space-wide stable-ID namespace decision. This run found nothing that
  argues for pulling any of that forward independently, which is why three of last cycle's
  four gaps are referenced above rather than re-filed.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- → H1 · tracked — Preserve untouched YAML scalar values and wrapping | tracked by 6g1dhhk6721x | high priority; decoded-value corruption reproduced
