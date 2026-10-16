# Persistence and recovery

[Architecture index](../ARCHITECTURE.md) · [Planning data](planning-data.md)

This guide owns filesystem transaction relationships and recovery boundaries.
The [store adapter](../../internal/store/fsstore.go) implements consumer-owned
[core ports](../../internal/core/store.go); it does not set application policy.

## Markdown and source preservation

Frontmatter is authoritative for status/bucket; storage is flat and ID-led for
tasks, audits, research and Threads. Epics retain their sequence-led names; research
has no lifecycle. The [entity descriptor](../../internal/domain/entity.go) owns
directory/authoring metadata, not a duplicate list in controllers.
[ADR-0003](../../planning/adrs/0003-stable-key-id-addressed-storage.md) owns this decision.

Missing/unrecognized statuses are listed and lint-flagged; unreadable files remain
source-problem evidence, not candidates for silent deletion/movement.
Non-entity Markdown belongs in `meta/`; scanned
`README.md` has the explicit carveout. Local `FileProblem` paths remain appropriate
for filesystem repair; portable read adapters translate them once into neutral
record problems in the same scan.

[Frontmatter editing](../../internal/store/frontmatter.go) and
[the preservation encoder](../../internal/store/frontmatter_preserve.go) keep
unknown keys, comments, order and untouched block-scalar source. The decoded result
must equal the requested edit or the writer refuses before persistence. Inline
comment spacing may normalize. Body writes and dependency repair share the safeguard;
fresh creates use the ordinary encoder.
[Fidelity tests](../../internal/store/frontmatter_fidelity_test.go) exercise refusal
and surgical preservation, not merely parseability.
Existing mapping keys stay in place; newly added keys are appended in lexical order
because update maps have no caller order. Task field validation uses the same stable
key order for deterministic first-error advice, not whole-file sorting.

ID validity stays strict lowercase. Case-only spelling diagnostics suggest the same
lowercase identity, never a replacement ID; automatic repair does not gain case-only
rename authority. Manual corrections must account for filename, declaration and
inbound references. Alias repair retains its existing guarded ownership/refusal rules.

Generic create/set/edit cannot change graph-owned dependencies; generic set/edit
cannot change task status. Guarded dependency and lifecycle verbs own those changes.
Ordinary task creation accepts only non-start states; guarded create-and-start uses
the lifecycle capability. Text-level lint repair skips unchecked dependency normalization.

## Locking and source CAS

[Repository guards](../../internal/store/lock.go) combine a canonical-root process
mutex with directory `flock` on Unix. The supported release matrix is macOS/Linux;
other source builds explicitly refuse mutation until a runtime-tested shared-root
lock identity exists. See [platform lock implementations](../../internal/store/lock_unix.go).

Cooperating writers serialize verification and replacement. Raw editors do not join
the advisory lock: whole-source and immediate per-target checks detect drift where
possible but do not eliminate the final verify-to-rename race.

[Content CAS](../../internal/store/cas.go) uses an internal whole-file hash, never
persisted or published. Guarded snapshot equality includes duplicate readable
occurrences and opaque unreadable revisions, not just representative entity IDs.
Interactive edit checks the source after the editor returns; it does not hold the
lock through an editor session. Scriptable mutations have bounded pre-commit retries
where their use case supports them; previews, editor conflicts and committed results
must not be blindly retried. See [retry tests](../../internal/core/occ_retry_test.go)
and [recovery tests](../../internal/core/dependency_recovery_test.go). A general shared
retry/durability policy remains [design work](../../planning/tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md).

Task body/field ports take an explicit selected identity, not a fuzzy reference.
Their initial lookup and CAS both use exact ID ownership; a missing ID is not permission
to try another record's slug. Core requests initial user-reference resolution through
the adapter's parse-free source lookup, preserving field repair without treating a
malformed typed value as missing identity. Field writes validate the updated document;
body writes still require their resulting task to decode.
[Target regressions](../../internal/store/task_target_test.go) exercise rename, old-alias
reuse, disappearance, collisions and preview/no-op boundaries. Audit/research/epic
scriptable ports use the same explicit-target rule, with entity-specific parse-free
source lookup. Their write and CAS lookups are exact; an epic uses its whole stem,
not its numeric prefix. [Entity regressions](../../internal/store/entity_target_test.go)
cover all nine retry families, duplicate ownership and occurrence-based header repair.
Graph retries retain semantic endpoints while recomputing current changes; they do
not replay cached write sets. See [graph retry regressions](../../internal/store/graph_retry_target_test.go)
and the [planning-data guide](planning-data.md).

## Ordinary create and task rename

[Ordinary creation](../../internal/store/create.go) holds the canonical guard before
identity scans/sequence allocation through no-clobber create. Dry-run performs checks
but reserves nothing. Task/Thread creation shares a cross-kind ID namespace; do not
infer a universal cross-entity namespace from that rule.

Kind-specific receipts retain planned/committed local paths outside domain values.
A post-commit guard-release failure returns the committed receipt and error; callers
inspect the created record rather than remint/retry. Research creation cannot treat
a committed/finalization failure as an ID collision. CLI JSON suppresses a whole-call
retry hint for committed creation. Future adapters may provide native uniqueness
without copying filesystem lock ordering.

[RenameTask](../../internal/store/rename.go) captures the source before waiting,
replans inbound links under the guard, and CAS-checks every write. Links commit first,
exclusive destination create second, source deletion last; destination permissions
come from the source. A pre-destination prefix is convergent and resumable by stable ID.
A destination-written/source-retained failure preserves both for inspection, not
destructive retry. Preview claims no reservation. See
[rename regressions](../../internal/store/rename_test.go) and ADR-0003's amendment.

## Guarded graph and lifecycle writes

All planner callbacks are pure over taskflow-owned immutable snapshots. While a
planner runs, any Store call at that canonical root—including through another FS or
an unrelated concurrent caller—fails with conflict instead of escaping the snapshot
or deadlocking. Adapters can later retry a read after the planner window.

| Capability | Transaction relationship |
| --- | --- |
| `TaskGraphMutationStore` | Strict snapshot, core plan/prefix validation, surgical dependency writes, whole-source and target CAS |
| `TaskLifecycleMutationStore` | One task transition plus full Thread impact attribution under the same guard; whole task+Thread snapshot CAS and immediate target CAS; distinct gate/AC override values |
| `ThreadCreationMutationStore` | Guarded initial identity/membership validation; only unstarted creation |
| `ThreadMutationStore` | Membership/lifecycle plan over task+Thread evidence; surgical atomic Thread replacement |
| `ThreadApplyMutationStore` | Identity revalidation, dependency writes first, no-clobber Thread creation last; no nested sibling transaction |
| `TaskGraphRepairStore` | Broken-graph-specific removals and preservation proof, not a mode on ordinary mutation |

Write order is semantic recovery data, not necessarily stable-ID sorting. Every
ordinary graph prefix must remain sound before writing; canonical edge-only supersets
can use the validated-final-graph optimization, while removal/migration retain full
prefix checks. Each durable replacement advances the typed receipt. Applied/remaining
IDs or lifecycle committed state—not proposed impacts—describe progress.

Bulk apply rechecks identity, tasks and Threads after each changed dependency prefix,
including when the intended Thread already exists. An out-of-band edit cannot make
the operation falsely complete. Authoring `member: false` describes upstream
compose-time context; runtime external gates are direct prerequisites outside
membership, not all transitive ancestors.

Core owns newly-unsafe/newly-inconsistent warning decisions and recovery advice.
CLI/TUI present that advice rather than deriving it from strings. Committed cleanup
failures remain failures with durable receipts; TUI reloads and retains warnings.
[Lifecycle/Thread tests](../../internal/store/threadmutation_test.go),
[Thread-apply tests](../../internal/store/threadapply_test.go) and
[authorization tests](../../internal/store/mutation_authorization_test.go) pin no-effects,
CAS and receipt behavior.

## Broken-graph repair

[Repair](../../internal/store/graphrepair.go) alone admits broken source graphs:
exact drops, deduplication and explicitly selected empty legacy-key cleanup.
Core simulates each source-file step, proves componentwise structural non-regression
separately from source preservation, and permits residual damage when selected repairs
strictly improve it. The adapter materializes exact YAML-node removals, reauthorizes
each next group against the durable snapshot, and performs whole-source and fresh
per-step CAS.

Malformed Threads make impact evidence incomplete but cannot prevent repairing
underlying tasks. Pathless declarations remain diagnostic, not locally executable
targets. A manifest is convergent intent, not a stale replacement snapshot.
See [repair regressions](../../internal/store/graphrepair_test.go).

## Managed audit bodies

Finding recognition is not repair authority.
[The domain classifier](../../internal/domain/finding_header.go) distinguishes
canonical, repairable drift, ambiguous diagnostics and ordinary prose. Only
repairable drift is eligible for automatic rewriting. Prose/code/fence content does
not gain status authority by resembling a finding. Near-misses remain separate from
parsed tallies; parsed 100% does not mean complete evidence.

Guarded close/defer requires complete parsing and terminal parsed findings before
same-bucket/preview returns; reopen and genuinely empty audits remain supported.
Finding status edits revalidate against the bucket in the CAS-protected transform.
Terminal corrections and note/candidate-only edits remain available; unrelated
pre-existing defects stay lintable. Generated notes cannot manufacture headings,
status labels or fences. Tracked-token settlement still needs independent destination
metadata lint. Detailed policy/evidence lives in
[the settlement task](../../planning/tasks/6ghht05bzxkk-align-audit-lifecycle-and-finding-writes-with-parsed-settlement-policy.md);
[remaining body-write gaps](../../planning/tasks/6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md)
are not claimed solved by these structured guards.

Finding and criterion vocabularies share overlap, not a subset relation.
[Resolution owners](../../internal/domain/resolution.go) and validated write verbs
own those words; managed candidate rows have a versioned marker while unversioned
legacy sections remain prose. Tolerant reads expose damaged data for lint; guarded
writes enforce their documented, baseline-aware boundaries.
