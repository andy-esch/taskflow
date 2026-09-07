---
schema: 1
id: 6g7sb0wnv5t5
bucket: closed
area: audit-id-collision-hardening-implementation-claude
date: "2026-09-07"
updated_at: "2026-09-07"
---
# Audit: Audit ID collision hardening implementation — claude — 2026-09-07

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Adversarially review the implementation that prevents two audit documents from owning the same
canonical stable ID and makes the collision visible through ordinary repository lint. Treat the
implementation and its tests as claims to challenge, not proof. Complete a correctness pass and
then a distinct systemic pass looking for an invariant that appears local but fails across another
creation path, filesystem instance, process boundary, malformed document, or lint projection.

## Review target

Review the complete uncommitted patch on branch `chore/v020-closeout-audit-id-hardening` relative to
`origin/main`, concentrating on:

- `internal/store/create.go` and `internal/store/create_test.go`;
- `internal/core/service.go` and `internal/core/finding_test.go`;
- `internal/cli/lint_audits_test.go`;
- task `6g7s4k845fsb`, follow-up `6g7s6hr3qnfq`, and finding H1 in audit
  `6g7qc8qd00xe`.

Produce a verified consumer inventory covering every production caller of `CreateAudit`, the repository-lock implementation used by
`writeLock`, every audit candidate/list path feeding creation and lint, and all consumers of
`domain.DuplicateIDIssues`. The release-closeout documentation elsewhere in the patch is context,
not the primary review target; report only factual contradictions there.

## Intended contract to challenge

- A non-dry-run audit creation holds the repository-wide writer guard across the canonical ID scan
  and atomic file creation. Concurrent cooperating writers using different slugs but the same ID
  cannot both succeed; exactly one owns the ID.
- An existing audit whose canonical filename ID matches the proposed ID causes `CreateAudit` to
  return `domain.ErrConflict` before a new file is written. A distinct ID with a duplicate slug
  remains legal.
- Dry-run performs the same collision validation without writing or taking a lock whose only value
  would be coordinating a mutation.
- Ordinary `tskflwctl lint` reports same-kind duplicate canonical audit IDs for every readable
  conflicting audit and identifies the duplicate ID and affected sources deterministically. Reads
  remain available for diagnosis; the change must not make audit listing fail closed.
- Canonical identity is the valid ID leading the flat filename. This task does not decide global
  cross-kind uniqueness, rewrite duplicate IDs, add rename semantics, or repair malformed audits.
- The separately noticed research check/create race is genuinely bounded to task `6g7s6hr3qnfq`
  and is not evidence that the new audit path repeats the same race.

## Mandatory evidence floor

Work only in the independent sandbox required by the injected protocol. Establish the exact patch
and enumerate the symbols above from code rather than trusting this brief. At minimum:

1. Reproduce different-slug/same-ID creation rejection and the copied-audit ordinary-lint failure
   through public behavior, checking exit/error classification and absence of a refused file.
2. Run the focused tests under `-race`, then repeat the concurrent creation test enough times to
   challenge scheduling assumptions. Add a sandbox-only probe using two `FS` instances for the
   same root; if practical, exercise separate CLI processes or explain precisely which existing
   cross-process lock tests establish that boundary.
3. Exercise duplicate slugs with distinct IDs, exact-path collision, dry-run collision, and a clean
   corpus. Confirm no legal case is newly rejected and no dry-run write occurs.
4. Challenge canonical-ID selection with filename/frontmatter drift, an unreadable audit, a
   non-ID-led carveout file, and more than two duplicates. Determine whether each result matches the
   stated readable-record and canonical-filename contract and whether every actionable source is
   named.
5. Mutation-test both guarantees: remove or move the locked collision scan so check and create are
   separable, and remove the `DuplicateIDIssues` audit wiring. Identify the exact newly added test
   that fails for each mutation; a failure caused only by an unrelated test is insufficient.
6. Inspect error ordering and candidate ordering for filesystem nondeterminism. Verify that repeated
   lint runs produce stable source ordering and that duplicate diagnostics compose with existing
   audit finding/near-miss problems rather than replacing them.
7. Run `go test -race ./...`, `just lint`, `just tidy-check`, `just docs-check`, `tskflwctl lint`,
   `tskflwctl audit lint`, and `git diff --check` in the sandbox with writable isolated caches where
   needed.

For every finding, cite an exact path/line or command result and include a minimal reproduction or
failing test. Verify names and capabilities from the sandbox. Do not treat planning requirements as
already implemented evidence.

## Required hostile angles

- Look for a check-then-create window hidden behind helper names, including different `FS` values,
  alternate root spellings, symlinks, and separate processes.
- Determine whether lock acquisition and directory creation preserve the pre-existing permissions,
  failure semantics, O_EXCL behavior, and no-partial-file guarantee.
- Challenge whether `auditCandidates` is the right identity inventory for both creation and lint,
  especially when a document is malformed or its filename ID disagrees with frontmatter.
- Check whether `FilenameID == ""`, duplicated IDs on three or more records, or list/read problems
  can silently suppress or misattribute the diagnostic.
- Inspect whether service-level ID minting should retry an astronomically rare collision under the
  accepted ADR, distinguishing that policy question from the task's explicit reject-before-write
  contract.
- Search sibling entity creation/lint paths for an abstraction asymmetry caused or exposed by this
  patch. File a finding only when demonstrated; the already tracked research race should not be
  duplicated without new evidence.
- Review the tests as attack surfaces: shared fixtures, same-instance locking, substring-only output
  assertions, or cached ordering may make them pass while the public invariant is false.

Do not broaden the implementation into global cross-kind uniqueness, automatic repair, schema
migration, or a generic storage rewrite. Track a real out-of-scope issue explicitly when it cannot
be safely addressed here.

## Validation and restoration

Do not edit, stage, commit, restore, reset, clean, rebase, push, or switch branches in the source
checkout. The sandbox helper's baseline commit is the only permitted commit. Sandbox-only probes
may temporarily mutate code or fixtures, but restore them to the captured baseline before transfer
and retain command evidence in the audit. Run the helper's `verify` and `transfer` steps exactly as
instructed; if the source audit changed, report the conflict and leave the sandbox intact.

## Deliverable

Replace the reviewer-report placeholder in your assigned audit with:

- an isolation attestation and reviewed baseline;
- a concise verdict;
- a verified consumer/invariant inventory;
- findings ordered by severity, each left `open`, with evidence, impact, and a bounded correction;
- hostile tests and mutations attempted, including clean results where meaningful;
- validation results and any limitations.

A no-findings verdict is acceptable only after the mandatory probes are evidenced. Do not change
finding statuses or any file other than the assigned audit.

## Reviewer report

### Isolation attestation

All inspection, builds, tests, generators, mutation probes, scratch fixtures, and this report were
performed in an independent sandbox created by `scripts/isolated-review-workspace.sh`. The source
checkout was used only to read this brief and to perform the helper's initial copy; no
state-changing command was run there.

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.C7TXwa
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.C7TXwa/.git
baseline_commit=343c770c604f8e18f4a2c05384d8ceba83a7d533
source_blob=00f3550bca6ad2bfc2b0d2a94c59386b333ae276
source_fingerprint=4a591c6420f9ca2af3281669b08994bed0c31880
deliverable=planning/audits/6g7sb0wnv5t5-2026-09-07-audit-id-collision-hardening-implementation-claude.md
deliverable_changed=false
transfer=pending
```

Independence was checked directly, not assumed: `git -C "$SANDBOX" rev-parse --absolute-git-dir`
resolves to the in-tree `.git`, `.git/objects/info/alternates` is absent, and
`git worktree list --porcelain` reports exactly one worktree — the sandbox itself. The sandbox is
therefore not a `git worktree`, symlink, or alternates-backed clone of the source.

**Reviewed baseline.** Sandbox baseline commit `343c770c` (parent `7ed39c8`, the v0.20.0 tag
commit), which captures the source's staged, unstaged, and untracked state at copy time. The
reviewed patch is `git diff origin/main HEAD` — 16 files, +545/-42, comprising
`internal/store/create.go`, `internal/core/service.go`, three test files, and planning/release
documentation. Every probe was restored to this baseline before transfer; `git status --porcelain`
and `git diff --stat HEAD` were both empty immediately before the `verify` step above.

### Verdict

**The central contract holds.** Every guarantee the brief asked me to challenge was reproduced and
survived adversarial probing: the canonical-ID scan and the atomic create are genuinely serialized
under the repository write lock — verified across separate `*FS` values, alternate root spellings,
a symlinked root, and **separate OS processes**; a different-slug/same-ID create is rejected with
`domain.ErrConflict` before any file is written; duplicate slugs with distinct IDs remain legal;
dry-run performs the same validation while writing nothing and creating no directories; and
ordinary `tskflwctl lint` now reports duplicate canonical audit IDs deterministically where
`origin/main` reported the same corpus as fully clean. Both mutations required by the brief are
killed by exactly the newly added tests, with no unrelated failures. The separately noticed
research check/create race is genuinely bounded to `6g7s6hr3qnfq` and is **not** repeated on the
audit path.

**One demonstrated systemic gap remains (M1).** Creation and lint answer the same identity question
from two different inventories. Creation and the write-path CAS use the *filename* inventory
(`auditCandidates`), which sees every id-led file; lint uses the *parsed* inventory
(`ListAuditsWithFindings`), which silently drops any audit that fails to parse. When one colliding
audit is unparseable, lint reports the surviving audit as clean while that audit is permanently
unwritable, and when three files collide it undercounts and leaves one actionable source unnamed. I
followed lint's own emitted repair instruction against the state that recommends it and the corpus
stayed broken with the duplicate diagnostic entirely gone. This is a diagnostic and repair-guidance
defect, not data corruption — writes still fail closed — so I rate it medium, not high.

No finding contradicts the decision to ship; M1 narrows a claim that the completed acceptance
criterion and the H1 resolution state absolutely (L2).

### Verified consumer and invariant inventory

Every name below was read from code in the sandbox at the cited path and line.

**Production callers of `CreateAudit`** — exactly one: `Service.NewAudit`
(`internal/core/service_audit.go:61`), reached from the `audit new` command. The audit ID is minted
by `s.newID()` (`internal/core/service_audit.go:48`), which is `id.New` (`internal/core/service.go:199`)
— process-monotonic (`internal/id/id.go:68-77`), so within one process an audit ID cannot repeat.
`--date` backdates the *slug and frontmatter*, not the ID, so audits do **not** cluster in one
millisecond slot the way research does. All other references are the `core.Store` interface
declaration (`internal/core/store.go:252`) and test fakes.

**Repository lock used by `writeLock`** (`internal/store/lock.go:103`) — two composed halves:
a process-local keyed mutex (`processRepositoryLock`, `internal/store/lock.go:63-66`) keyed by
`repositoryLockKey`, which resolves symlinks, absolutizes, and cleans the root
(`internal/store/lock.go:37-49`); and `flock(LOCK_EX)` on the planning root directory
(`internal/store/lock_unix.go:24-31`), which is authoritative across processes. Non-unix targets
refuse mutation outright rather than silently no-op (`internal/store/lock_other.go`). `CreateAudit`
takes this same lock at `internal/store/create.go:252` — the identical primitive used by all 17
production mutation sites.

**Audit candidate/list paths.** Two inventories, and the split is the substance of M1:

| Inventory | Implementation | Sees unparseable files? | Consumers |
| --- | --- | --- | --- |
| Filename (`auditCandidates`) | `internal/store/auditstore.go:176` → `flatCandidates`, `internal/store/resolve.go:139-163` (ReadDir + `splitFlatName`, no parsing) | **Yes** | `ensureAuditIDUnique` (`create.go:269`), `resolveAuditPath` CAS (`auditstore.go:163`), `resolveAudit` (`auditstore.go:184`) |
| Parsed (`ListAuditsWithFindings`) | `internal/store/auditstore.go:30` → `scanDir`, `internal/store/resolve.go:32-77` | **No** — a parse failure becomes a `FileProblem` and the record is skipped at `resolve.go:62-67` | `Service.Lint` (`core/service.go:570`), `Summary` (`core/service.go:319`), `FixFindingHeaders` (`core/finding.go:292`) |

`scanDir` is fatal only on a genuine read error (`resolve.go:52`); `README.md` is carved out by name
(`resolve.go:46`), and non-id-led files are rejected by `splitFlatName` in both inventories, so the
carveout gates agree.

**Consumers of `domain.DuplicateIDIssues`** (`internal/domain/lint.go:350`) — three, all in
`Service.Lint`: research (`core/service.go:556`, pre-existing), **audits (`core/service.go:579`,
added by this patch)**, threads (`core/service.go:589`, pre-existing). The new audit wiring is a
faithful copy of the research idiom. Tasks deliberately do not use it: they carry a stronger
graph-level duplicate-ID diagnostic that names full paths, which I verified independently
(`task show` → exit 13 ambiguous; `lint` → `duplicate stable task id … across <path>, <path>; no
source is uniquely authoritative`).

**Invariants confirmed by probe, not by reading:** file mode `0644` and directory mode `0755` are
unchanged; `createFileAtomic`'s `O_EXCL` remains the last-resort guard behind the scan; no partial
or temp artifact is left in `audits/` on either the success or the refusal path; and a dry-run
against a non-existent planning root neither writes nor creates the root.

### Findings

#### M1. Duplicate-audit-ID lint is suppressed or undercounts when a colliding audit is unparseable, so the readable victim is reported clean while permanently unwritable · **Status:** fixed

**File:** `internal/core/service.go:570-586` · `internal/store/resolve.go:62-67` |
**Component:** core/lint

**Evidence.** `Lint` builds the duplicate inventory from `ListAuditsWithFindings`, which skips any
audit whose frontmatter fails to parse (`internal/store/resolve.go:62-67`, `continue`). Creation
(`ensureAuditIDUnique`, `internal/store/create.go:268-280`) and the write-path CAS instead use the
filename inventory, which sees those files. Two audits sharing canonical ID `6g7s4k845fsb`, one with
broken YAML:

```
$ tskflwctl -C $W lint --color=never
! …/audits/6g7s4k845fsb-2026-09-07-beta.md
    validation failed: malformed frontmatter: invalid YAML — line 2: did not find expected ',' or ']'
error: validation failed: 0 item(s) with issues, 1 unreadable file(s)      # exit 11
```

Zero duplicate diagnostics. The surviving audit is nonetheless bricked, and the write path *can*
name both files because it uses the filename inventory:

```
$ tskflwctl -C $W audit show 6g7s4k845fsb                                   # exit 13
error: "6g7s4k845fsb" matches 2 audits by id: 2026-09-07-alpha …, 2026-09-07-beta …: ambiguous match
$ tskflwctl -C $W audit close 2026-09-07-alpha                              # exit 13
✘ 2026-09-07-alpha: re-resolve audit … id "6g7s4k845fsb" is claimed by 2 files: … : ambiguous match
```

With three colliding files where one is malformed, the diagnostic undercounts — `shared by 2 docs`
— and never names the third actionable source. **Running lint's own emitted repair against the
state that recommends it** ("rename one file and its frontmatter id") renames one of the two named
files and leaves the repository still broken, now with *no* duplicate diagnostic at all:

```
$ mv $W/audits/6g7s4k845fsb-2026-09-07-alpha.md $W/audits/6g7s4k845fzz-2026-09-07-alpha.md
$ sed -i '' 's/^id: 6g7s4k845fsb/id: 6g7s4k845fzz/' $W/audits/6g7s4k845fzz-2026-09-07-alpha.md
$ tskflwctl -C $W lint --color=never
! …/audits/6g7s4k845fsb-2026-09-07-gamma.md
    validation failed: malformed frontmatter: invalid YAML — line 1: …
error: validation failed: 0 item(s) with issues, 1 unreadable file(s)      # exit 11
$ tskflwctl -C $W audit close 2026-09-07-beta                              # exit 13, still bricked
```

**Impact.** Lint is the hygiene gate this repository actually runs, and in the mixed
readable/unreadable corpus it attributes an identity collision solely to the other file's YAML. A
maintainer who fixes the named defect and re-runs lint is told the corpus is clean on the audit axis
while an audit remains unwritable. This is the specific failure mode the new diagnostic exists to
explain. Writes still fail closed, so no corruption occurs and severity is bounded to diagnosis and
repair guidance.

**Bounded correction.** Make lint answer the identity question from the same inventory creation
uses. The cheapest fix stays inside the code this patch already touches: `Lint` receives the audit
`FileProblem`s as `ap` (`core/service.go:570`) and each carries `Path`, so folding the leading
canonical ID of each unreadable audit path into `auditIDs` before calling `DuplicateIDIssues`
restores the union — this needs a small exported filename-ID helper, since `splitFlatName` is
store-private (`internal/store/flatname.go:28`). The more principled alternative is a `Store` method
exposing canonical audit IDs from `auditCandidates` and keying lint off that. Either way the
unreadable file should also be named as a conflicting source so the count and the repair advice are
complete. Note the identical shape is pre-existing for research and threads (I reproduced it for
research), so whichever fix is chosen is worth applying to all three call sites rather than audits
alone.

**Resolution:** Resilient read problems now retain optional adapter-recovered
entity identity outside the public wire contract. Ordinary lint unions readable
records with unreadable identity sources, attaches the collision to both, and
never parses an adapter path for meaning. The shared path covers audits,
research, and Threads; focused tests cover mixed readability for all three and a
three-way CLI audit reproduction.

#### L1. The shared duplicate-ID message says "both" while reporting three or more documents, and names a count rather than the conflicting sources · **Status:** fixed

**File:** `internal/domain/lint.go:363-368` | **Component:** domain/lint

**Evidence.** With three colliding audits, human and `--json` output both read:

```
duplicate stable id "6g7s4k845fsb" shared by 3 docs — both are unresolvable by id and unwritable …
```

"shared by 3 docs — **both**" is self-contradictory. The message also identifies only a count; the
conflicting sources are recoverable solely from the surrounding `LintResult` labels. That is
sufficient and deterministic today (the JSON envelope carries a per-slug row, and 25 consecutive
lint runs produced one byte-identical output), but it is weaker than the task-graph equivalent,
which names full paths: `duplicate stable task id … across <path>, <path>; no source is uniquely
authoritative`.

**Impact.** Cosmetic-to-mild: the wording undercuts confidence in an otherwise accurate diagnostic,
and the count-only phrasing is what makes M1's undercount invisible rather than obviously wrong.
Slug labels are also not a unique key — duplicate slugs with distinct IDs are legal by this patch's
own contract — so two rows can share a label.

**Bounded correction.** Reword to a count-agnostic phrase ("all are unresolvable by id and
unwritable") and, ideally, pass the conflicting paths into the message as the task diagnostic does.
This is pre-existing shared code that the patch newly routes audits through; it is not a regression.

**Resolution:** Duplicate-ID diagnostics now use count-neutral 'all' wording,
list every available source location in deterministic order, and instruct the
operator to assign distinct filename/frontmatter IDs to all but one owner.
Domain and CLI tests pin three-way wording, source completeness, and ordering.

#### L2. The completed acceptance criterion and the H1 resolution state duplicate-ID lint coverage absolutely, which the shipped behavior does not meet · **Status:** fixed

**File:** `planning/tasks/6g7s4k845fsb-detect-duplicate-audit-ids-at-creation-and-lint-time.md:36-37`
· `planning/audits/6g7qc8qd00xe-2026-09-07-arch-data-model-and-storage.md` (H1 resolution) |
**Component:** planning

**Evidence.** The ticked criterion reads "Ordinary `tskflwctl lint` reports **every** same-kind
duplicate audit ID and identifies **all** conflicting documents deterministically", and the H1
resolution repeats "ordinary lint reports every duplicate audit ID with all conflicting sources".
M1 demonstrates both claims fail once a conflicting document is unparseable. The brief's own
contract is worded more carefully ("for every *readable* conflicting audit"), so the code matches
the intended contract while the planning record overstates it.

**Impact.** Audit `6g7qc8qd00xe` H1 was moved to `fixed` on the strength of the absolute claim. A
later reader has no signal that the mixed readable/unreadable corpus is uncovered.

**Bounded correction.** Qualify both statements to readable/parsed records and reference M1 (or its
successor task) as the remaining gap. No status change is proposed here — I have left every finding
`open` for implementation-owner triage as instructed.

**Resolution:** The implementation now includes unreadable canonical identity
sources and names all conflicting documents, so the completed acceptance
criterion and architecture-audit H1 resolution are accurate without
qualification. The review closeout evidence is recorded on task 6g7s4k845fsb.

### Hostile tests and mutations attempted

**Required mutations — both killed by exactly the newly added tests.**

| Mutation | Result | Killed by |
| --- | --- | --- |
| Hoist `ensureAuditIDUnique` outside the write lock in `CreateAudit` (check and create separable) | **393/393 iterations failed**, all `successes=2 conflicts=0` — a deterministic kill, not a flaky one | `TestCreateAudit_SerializesDuplicateIDCheckWithCreate` (`internal/store/create_test.go:189`) |
| Remove the `DuplicateIDIssues` audit wiring from `Lint` | Failed | `TestServiceLintReportsDuplicateAuditIDs` (core) and `TestLintReportsEveryAuditSharingAStableID` (cli) — and **no unrelated test failed**, so the kill is attributable |

My first attempt at the unlock mutation asserted and refused to apply because the pattern also
matched `CreateTask`; I re-scoped it to the `CreateAudit` function body. That is worth recording:
the two functions share the mutation's exact shape, so a careless coordinated edit would have
produced a misattributed result.

**Sandbox-only probes written, run, and then removed** (`internal/store/zz_review_probe_test.go`,
`zz_review_proc_test.go`). Each was itself validated against the unlock mutation to prove it is a
real detector rather than a tautology — all three concurrency probes failed against the mutant with
two files sharing one ID:

- **Two independent `*FS` instances, same root, 60 raced iterations** — exactly one success each
  time. Fails against the mutant at iteration 0.
- **Separate OS processes, 25 raced iterations** — two re-executed test binaries synchronised on a
  filesystem barrier, exactly one `created` and one `conflict` each time. Fails against the mutant
  (`created=2`, two files on disk). This closes the cross-process boundary directly rather than by
  inference. It is corroborated by the pre-existing
  `TestRepositoryLockReleasesWhenProcessTerminates`, which proves the product `writeLock` path holds
  the cross-process flock via a non-blocking probe, and by
  `TestMutateTaskGraphSerializesOppositeEdgesAcrossProcesses`.
- **Alternate root spellings raced against the canonical root**, 60 iterations — exactly one
  success. Also verified non-concurrently that a trailing separator, `.`, `..`-round-trip, and a
  **symlinked** root all still observe the collision, and that none of them wrote a second file.
- **Legal cases** — duplicate slug with a distinct ID is accepted; exact-path collision remains
  `ErrConflict`; a non-id-led carveout file in `audits/` neither becomes a candidate nor blocks
  creation.
- **Dry-run** — refuses the duplicate ID, writes nothing, and does not create the planning root on a
  virgin path.
- **Permissions / no-partial-file** — root `0755`, `audits/` `0755`, file `0644`, exactly one entry,
  no temp artifacts.

**Clean results worth recording (challenged, found sound).**

- **Filename/frontmatter drift.** A file named with ID `…fzz` whose frontmatter claims `…fsb` is
  correctly flagged by `IDDriftIssue` (audits do get it, via `AuditLintIssues`,
  `internal/core/finding.go:277`), no false duplicate is raised, and both audits stay resolvable by
  their canonical filename ID. The canonical-filename contract holds.
- **`FilenameID == ""` cannot arise for a listed audit** — `parseAuditWithFindings` errors out
  before constructing a record when `splitFlatName` fails (`internal/store/auditstore.go:207-210`),
  so the empty-ID skip in `DuplicateIDIssues` is unreachable defensive code on this path, not a
  suppression channel.
- **Ordering determinism** — `os.ReadDir` sorts by filename and `scanDir` preserves that order; 25
  consecutive lint runs on a three-way collision produced exactly **one** distinct output hash.
- **Diagnostic composition** — the duplicate issue is *appended* to near-miss and bucket issues,
  replacing nothing (verified with an audit carrying a near-miss header, a closed bucket with an
  open finding, and a duplicate ID simultaneously).
- **Research race is genuinely bounded.** `CreateResearch` performs its scan at
  `internal/store/create.go:319-328`, *before* `writeNewFile` takes the lock at line 335 — the race
  `6g7s6hr3qnfq` describes. `CreateAudit` does not repeat it. I found no new evidence justifying a
  duplicate finding.
- **Service-level mint retry is a defensible policy difference, not a defect.** `NewResearch`
  regenerates on collision (`internal/core/service_research.go:87-97`) because research IDs are
  minted from a *day* and same-day docs share one 2^17 tail. Audit IDs come from `id.New` at current
  time and are process-monotonic, so the collision is confined to a cross-process same-millisecond
  event at 2^-17. Rejecting rather than retrying is consistent with the task's explicit
  reject-before-write contract; I raise it as inventory context, not a finding.
- **No sibling creation-path asymmetry was demonstrated.** Tasks have a stronger path-naming
  duplicate-ID diagnostic through the task graph; epics use `NN-` numbering with a documented,
  accepted non-serialization; threads share the `DuplicateIDIssues` idiom. Nothing here rose above
  the shared-idiom gap already recorded as M1.
- **Release-closeout documentation contains no factual contradiction.** I verified the claims that
  are checkable: tag `v0.20.0` resolves to `7ed39c86c6b597cbb3614cb1cf5f893c1d69cbec`, exactly the
  candidate commit the release task records; and README's new "the CLI owns guarded Thread
  mutations, repair, and bulk apply" matches the shipped surface (`thread add|remove|start|
  complete|cancel|reopen|apply` are all present in `thread --help`).

**Pre-existing behavior confirmed unchanged by this patch** (compared against a binary built from
`origin/main` via `git archive`): a duplicate audit ID that is *unreadable by permission* makes
`lint` fail closed with exit 1 and no diagnostics at all. This is identical on `origin/main`
(`scanDir` returns fatally at `internal/store/resolve.go:52`), so the patch does not make audit
listing fail closed — the brief's contract on that point is met. The same comparison confirms the
patch's value: `origin/main` reports a two-way readable audit ID collision as
`✔ all planning entities and dependency links pass lint`, exit 0.

### Validation results

Run in the sandbox at the restored baseline:

| Command | Result |
| --- | --- |
| `go test -race ./...` | **PASS** — every package ok, exit 0 |
| `just lint` (`golangci-lint run ./...`) | **PASS** — 0 issues |
| `just tidy-check` (`go mod tidy -diff`) | **PASS** — no diff |
| `just docs-check` (docgen + `git diff --exit-code docs/cli`) | **PASS** — no diff |
| `tskflwctl lint` on `planning/` | **PASS** — exit 0, all entities clean |
| `tskflwctl audit lint` | **PASS** — all audit findings pass lint |
| `git diff --check` | **PASS** — clean |
| Focused new tests under `-race` | PASS; concurrency test repeated `-count=300` under `-race`, no failures and no race reports |

### Limitations

- Findings are left `open` for implementation-owner triage; I changed no finding status and no file
  other than this audit.
- The cross-process probe re-executes the Go test binary rather than the `tskflwctl` CLI. It
  exercises the same `store.FS.CreateAudit` and `writeLock` code path in genuinely separate
  processes, which is the boundary in question; a CLI-level race cannot be driven to a duplicate ID
  without an ID-generator seam the CLI does not expose.
- Verified on `darwin/arm64`, Go 1.26.6, under `//go:build unix` lock semantics. The `!unix`
  branch's refusal-to-mutate stance was read but not executed.
- M1's severity assumes the current fail-closed write behavior. I did not attempt to construct a
  case where a duplicate audit ID passes the CAS and commits.
- Probe corpora were built under `/tmp` outside both the source and the sandbox working tree; the
  sandbox was restored to `343c770c` (empty `git status --porcelain`, empty `git diff --stat HEAD`)
  before verification and transfer.
