---
schema: 1
id: 6gk49rvj7rhe
bucket: closed
area: review-workspace-hardening-implementation-claude-delivery-evidence
date: "2026-10-12"
updated_at: "2026-10-13"
---
# Audit: Review-workspace hardening implementation — claude-delivery-evidence — 2026-10-12

> Reviewer assignment: claude-delivery-evidence. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

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
project command in `$SOURCE_ROOT`. The sandbox's push URLs are disabled; never re-enable them or
push to the source path. If creation fails, report the blocker; never fall back to the shared
checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run
the helper's fail-closed verification:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, a removed or
weakened push guard, source-deliverable drift, and empty reports; it copies back only the assigned
audit through a same-directory atomic rename. A successful transfer records a helper-owned receipt
and prints it (`transfer=succeeded` with the transferred blob); a refused or failed transfer records
none. Put that printed receipt in your handoff message. Do not copy anything else manually. Leave the
workspace in place and report its path until the implementation owner confirms delivery with
`"$SANDBOX/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"`. On refusal, preserve
it and report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

## Review brief

Independently challenge the review-workspace hardening slice before it merges. The question to
settle: can the general isolated-review helper still write into the source repository, and is its
new delivery evidence truthful under every success, refusal, and failure path? A green suite and
checked ACs are hypotheses here, not evidence.

There are two reviewers, with complementary primary lenses and one shared intended contract.

- **git-safety** owns Git targeting and source preservation: the push guard, push-destination
  resolution, linked-worktree capture, and whether any `create`/`verify`/`transfer` path writes
  source refs, objects, administration, index, configuration, or files other than the
  deliverable. Bounded cross-check: the receipt's fail-closed paths.
- **delivery-evidence** owns the attestation and receipt contract: the state transitions across
  verify, transfer, and receipt; receipt atomicity and truthfulness; repeated or interleaved
  commands; failure paths; and whether the generated brief can be followed literally. Bounded
  cross-check: push-guard false positives in legitimate review workflows.

Each reviewer works independently and must not read the sibling report until their own is final.
This change modifies the helper you will use to create your own sandbox, so verify your sandbox's
independence yourself immediately after creation rather than trusting the helper's success output.

## Review target

Source checkout: the linked Git worktree named in your handoff prompt, on branch
`worktree-harden-isolated-review-workspaces`. Base `origin/main` is
`4fdf6fccf3d0fc80a7c58c3c33458b1ca9c1de54`, and the implementation is three commits:
`f5e0aaf1` (push guard), `22a27650` (transfer receipts), and `ef77ffd5` (linked-worktree sources),
with HEAD `ef77ffd5b54a4407a88cf7d6b7157406d9618469`. The only uncommitted state is the two review
audits. The source is itself a linked worktree, so your sandbox creation exercises the newly fixed
path.

Implementing records:

- [Push guard task](../tasks/6gk06e2g21hv-disable-pushes-from-isolated-review-workspaces.md),
  including its 2026-10-12 linked-worktree note
- [Receipt task](../tasks/6g7srp3py9fe-make-adversarial-review-transfer-attestations-self-finalizing.md)
- [Review-workspace hardening Thread](../threads/6gk06gktr1q5-harden-isolated-review-workspace-safety-and-delivery-evidence.md)
- Motivating review: [documentation ownership audit](6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md)

Entry points to verify in the sandbox before relying on them:

- `scripts/isolated-review-workspace.sh`: `check_push_guard`, `create_workspace` (guard
  installation, the `/.git` overlay exclude, the post-overlay `assert_independent_clone`),
  `load_state`/`load_workspace`, `attest`, `write_receipt`, `receipt_value`, `transfer_deliverable`,
  `show_receipt`, and `main`.
- `scripts/prepare-adversarial-review-audits.sh`: the injected sandbox protocol, the handoff
  prompt, and the `[delivery]` note.
- `scripts/README.md`: the guard, the receipt, and the evidence contract.
- `internal/tools/reviewworkspace/review_workspace_test.go`: `TestPushGuardRefusesOrdinaryPushesWithoutChangingSource`,
  `TestVerifyAndTransferRejectAWeakenedPushGuard`, `TestLinkedWorktreeSourceStaysIndependent`,
  `TestPushGuardAllowsLocalUpstreams`, `TestReviewWorkspaceTransfersOnlyTheDeliverable`,
  `TestTransferReceiptIsTheOnlyDeliveryEvidence`, `TestFailedTransfersRecordNoReceipt`, and
  `TestGeneratedBriefFollowsTheReceiptContract`.

Build a verified consumer inventory of every live reference to the helper's commands, output keys
(`push_guard`, `verify`, `transfer`, `transferred_blob`, `delivery`), and evidence contract. Cover
scripts, docs, AGENTS.md, routines, and generated-brief text, and distinguish live instructions
from historical planning records. Out of scope unless you show a new in-scope consequence:
OS/network sandboxing, hostile-agent containment, moving the helper into Go, deleting sandboxes,
and the task-attached review entity model.

## Intended contract to challenge

1. An ordinary push from a fresh sandbox fails before contacting the source: default push,
   `origin`, new or existing branches, `--all`, and tags. Fetching from `origin` still works. The
   source's refs, objects, configuration, HEAD, index, and files are unchanged. Explicit-URL pushes
   and command-line or user-wide overrides are documented limits, not guarantees.
2. `verify` and `transfer` refuse a removed `origin`, any remote whose effective push URL (after
   `insteadOf` expansion) is not exactly the guard, and any workspace-local `remote.pushDefault` or
   `branch.<name>.pushRemote`/`remote` naming something other than a configured remote or `.`. Each
   refusal includes a working restoration command. Legitimate workflows such as local tracking
   branches are not refused.
3. A linked-worktree source cannot cause the clone's `.git` to be replaced. No workspace Git write
   happens before independence is re-asserted. The baseline reflects the worktree's own branch and
   dirty state, and transfer writes only that worktree's deliverable.
4. The report attests verification only (`verify=passed`, `transfer=pending`). A receipt exists only
   after the transferred bytes are confirmed in the source. `receipt` fails before any transfer,
   and fails once the source or sandbox diverges from the transferred blob. After a receipt exists,
   `verify` and `transfer` refuse. Refused or failed transfers leave no receipt and no partial
   source change.
5. The generated brief, its handoff prompt, the `[delivery]` note, and `scripts/README.md` state one
   evidence contract, and a compliant reviewer can follow the brief literally.
6. Tests drive the real script against disposable repositories, with global/system Git config
   redirected and inherited `GIT_*` variables removed. They can never touch a developer's checkout
   or configuration, even when run from a Git hook.

## Mandatory evidence floor

Record each probe as reproduced, falsified, or unresolved.

1. **Own-sandbox independence:** immediately after `create`, show your sandbox's
   `rev-parse --absolute-git-dir`, that `.git` is a directory, that there are no alternates, the
   `remote get-url --push --all origin` result, and that the source worktree's `git status`,
   branch head, and the shared repository's `worktree list` are unchanged by your creation.
2. **Push attempts:** in a disposable fixture (not your review sandbox's source), reproduce the
   push matrix and a source-bytes comparison of your own design. Then attempt at least two
   weakenings or destination settings not covered by `TestVerifyAndTransferRejectAWeakenedPushGuard`.
   Candidates: `include.path`, `remote.<name>.pushurl` set to an empty value, `url.*.pushInsteadOf`
   once pushurl is removed, a remote name containing regex or whitespace characters, a
   `branch.<name>.pushRemote` whose branch name has a dot, and `push.default` variants. Report
   whether each is refused, harmless, or a bypass.
3. **Linked-worktree path:** in a disposable repository, run `create` from a linked worktree whose
   branch differs from the main checkout's, and compare the shared repository's administration
   bytes before and after. If GNU rsync is available (`rsync --version` not reporting openrsync),
   also run the pre-fix exclude (`--exclude='/.git/'`) and record what happens; the owner could
   not verify that locally.
4. **Receipt state machine:** drive verify, transfer, receipt, a repeated verify/transfer, an owner
   edit, and a workspace edit after transfer. Exercise at least one failure that occurs after the
   source replacement (for example, an unwritable private state directory when the receipt is
   written) and report the resulting state and exit status. Then check whether a reviewer could
   mistake that state for success.
5. **Compiler-valid mutations:** choose two critical invariants and, for each, make one
   behavior-changing edit that ShellCheck and `bash -n` accept. Run the exact intended test,
   require it to fail for the intended reason, then restore the sandbox baseline and show that test
   green. A syntax error is an invalid probe, not a kill. Prefer invariants the owner's ledger did
   not cover: the post-overlay independence assertion, the `.`-upstream allowance, the receipt's
   workspace-blob check, and the generator's handoff prompt.
6. **Literal walkthrough:** follow the generated brief's command blocks verbatim against a
   disposable source, report any step that cannot succeed as written, and check that every field
   the brief names appears in `verify` or receipt output.

## Required hostile angles

- Can the guard check be satisfied while a push still reaches the source, through ordering of
  `remotes`, multi-line values, case folding of config keys, `--local` scope versus included
  files, or a remote URL equal to the guard plus whitespace?
- Does any refusal message suggest a command that fails, or that would re-enable pushing?
- Does `load_state` (used by `receipt`) skip a check that matters for trusting the receipt, such as
  independence, the push guard, or source identity?
- Is the receipt keyed so that a stale receipt from an earlier helper version or another
  deliverable could vouch for the wrong file? Can `receipt_value` be fooled by duplicate keys,
  `=` in values, or CRLF?
- Can `transfer` replace the source file and then exit non-zero without a receipt, leaving an owner
  unable to distinguish "not delivered" from "delivered, not recorded"?
- Do the tests' fixture choices (one deliverable path, plain names, macOS versus Linux tools) mask
  a class of failures? Are any assertions satisfied vacuously?
- Is the documentation's security language accurate: accident prevention, explicit limits, and no
  implied hostile-agent containment?

Second pass: inspect the shared pattern rather than repeating the checklist. Look for boundaries
that only appear to fail closed, state that changes between check and action (guard check, then
push; hash check, then rename), test helpers whose environment isolation could mask a defect
class, and evidence fields copied across README, brief, and helper that could drift. Settle each
important hypothesis with hostile evidence.

## Validation and restoration

Required checks inside your sandbox:

```sh
just build
go test ./internal/tools/reviewworkspace/ -count=1 -v
shellcheck scripts/isolated-review-workspace.sh scripts/prepare-adversarial-review-audits.sh
./bin/tskflwctl -C . lint --links
git diff --check
```

The owner passed full race-enabled `just test` (28 packages) and `just lint` on HEAD. You may
inherit those, labeled as inherited, but the focused probes above are mandatory and independent.

Put disposable repositories, mutation backups, and caches in separate `mktemp -d` directories
outside your sandbox's tree. Set `GIT_CONFIG_GLOBAL` to a fixture file and `GIT_CONFIG_NOSYSTEM=1`
for every disposable-repository command, so no probe reads or writes the user's Git configuration.
Restore every probe to the sandbox baseline before verification. Only your assigned audit may
differ at verify/transfer time.

## Deliverable

Fill only the Reviewer report slot, and leave the brief intact. Create real findings with the
sandbox's fresh `./bin/tskflwctl audit finding new` and leave them **open**; do not create
follow-up tasks or close the audit.

Lead with the verdict (ready / needs changes / incomplete), the exact snapshot, and your limits.
Then include:

- an isolation attestation from your final `verify` output, plus your own independence probe;
- a verified consumer inventory with file:line or command evidence;
- findings consolidated by root cause, each with its reachable consequence, reproduction, evidence,
  recommendation, scope, and residual uncertainty, separating product defects, coverage gaps,
  defense-in-depth, and taste;
- a probe and mutation ledger;
- a bounded surviving-hypothesis ledger separating tested negatives from unresolved concerns.

High severity means a reachable wrong-target write, a source mutation, or materially false
delivery evidence. Coverage gaps are normally Low unless a concrete consequence justifies more.

## Reviewer report

**Verdict: needs changes.** Three Medium and three Low findings, all open (below). For ordinary pushes and common review workflows, the push guard does what it claims. On the success path, the receipt state machine is truthful and atomic. It breaks down in three places:

- **M1.** A transfer that fails after the source rename leaves the report delivered but reported as undelivered, with no recovery path.
- **M2.** The generated create block falls through into the shared checkout when `create` fails, and the brief's own path wording makes `create` fail.
- **M3.** `verify` attests an intact guard while include-file or legacy-remote settings push ordinary pushes to the source.

**Snapshot.**
- Source: the linked worktree `/Users/andyeschbacher/git/andy-esch/taskflow/.claude/worktrees/harden-isolated-review-workspaces` on `worktree-harden-isolated-review-workspaces`, HEAD `ef77ffd5b54a4407a88cf7d6b7157406d9618469`.
- Commits `f5e0aaf1`, `22a27650`, and `ef77ffd5` confirmed by `git log`.
- Base `4fdf6fcc516a6312e243ce1092c798d1d4797f9d`. The brief's full base hash `4fdf6fccf3d0fc80a7c58c3c33458b1ca9c1de54` does not resolve in the source or the sandbox (`git cat-file -t` fails); only the 8-character prefix matches.
- Captured as sandbox baseline `ca19368528f99285692caf4bde23be51ea4ef46a`, with source audit blob `b810560908263274603cc47d5e2316576d06877e`. The only uncommitted source state at capture was the two review audits.

**Limits.**
- Platform: macOS, `/bin/bash` 3.2, Git 2.54.0, ShellCheck 0.11.0, rsync = openrsync only (`rsync --version` reports openrsync protocol 29; `which -a rsync` finds only `/usr/bin/rsync`).
- GNU rsync behaviour was emulated, not run. Linux was not run.
- Full race-enabled `just test` (28 packages) and `just lint` are inherited from the owner, not rerun.
- I did not read the sibling `claude-git-safety` audit.
- M3 lies in git-safety's primary lane; I filed it because evidence floor 2 applies to both reviewers.

### Isolation attestation

Final `verify` output (`transfer=pending` is the truthful value; delivery evidence is the receipt in the handoff message):

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.KVi1y8
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.KVi1y8/.git
baseline_commit=ca19368528f99285692caf4bde23be51ea4ef46a
source_blob=b810560908263274603cc47d5e2316576d06877e
source_fingerprint=89b672cbfde3869f0f533d066cff8c7638e339e1
deliverable=planning/audits/6gk49rvj7rhe-2026-10-12-review-workspace-hardening-implementation-claude-delivery-evidence.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

Own independence probe, run immediately after `create` (evidence floor 1, reproduced):

- `git rev-parse --absolute-git-dir` → `<sandbox>/.git`, and `git-common-dir` → `.git`.
- `.git` is a real directory, not a symlink, and there is no `objects/info/alternates`.
- `remote get-url --push --all origin` → only `isolated-review-push-disabled://source-is-read-only`. The fetch URL is the source worktree, and the only remote is `origin`.
- The sandbox HEAD is on `refs/heads/worktree-harden-isolated-review-workspaces`, and the baseline's parent is `ef77ffd5`.
- Before/after `create`, these source values were byte-identical:
  - the source's `status --porcelain=v1 --untracked-files=all`
  - HEAD and its symbolic ref
  - the shared repository's `worktree list --porcelain`
  - SHA-256 of every file under `.git/worktrees/harden-isolated-review-workspaces/` (including `index`)
  - all 138 refs
  - `.git/config`
  - the loose-object count
- The source deliverable blob was still `b8105609…` and HEAD still `ef77ffd5` when work resumed after an interruption, and again before transfer.

### Verified consumer inventory

Live instructions and code. Each anchor was checked in the sandbox at baseline.

- `scripts/isolated-review-workspace.sh`:
  - commands: `main` dispatch `:381-384`
  - output keys: `push_guard` `:220,276`; `verify`/`transfer` `:277`; `transferred_blob` `:292`; `delivery=confirmed` `:371`; receipt file name `:11`
  - entry points: `check_push_guard` `:130`; `create_workspace` `:151` (guard install `:182`, `/.git` exclude `:185`, post-overlay `assert_independent_clone` `:188`); `load_state` `:224`; `load_workspace` `:246`; `attest` `:273`; `write_receipt` `:281`; `receipt_value` `:301`; `transfer_deliverable` `:331`; `show_receipt` `:353`; `main` `:374`
- `scripts/prepare-adversarial-review-audits.sh`:
  - create block `:161-167`
  - verify/diff `:184-185`
  - attestation field list `:188-190`
  - verify/transfer `:195-196`
  - owner `receipt` `:205`
  - handoff prompt `:237`
  - `[delivery]` note `:245`
- `scripts/README.md`:
  - usage `:9-18`
  - receipt contract `:28-35`
  - guard `:37-43`
  - limits `:45-48`
  - split evidence contract `:85-93`
  - absolute-to-relative path translation `:142-145`
- `internal/tools/reviewworkspace/review_workspace_test.go`: all eight named tests exist (`:180, :223, :281, :338, :349, :409, :460, :511`) and pass with `-count=1 -v`. Fixture isolation (strip `GIT_*`; set `HOME`, `TMPDIR`, `GIT_CONFIG_GLOBAL`, `GIT_CONFIG_NOSYSTEM=1`, `GIT_TERMINAL_PROMPT=0`) is at `:54-65`.
- `AGENTS.md:83` routes review tooling to the scripts guide and requires isolated sandboxes. It names no keys.
- `routines/` and `docs/` have no references to the helper or its keys (`git grep`).

Historical planning records:

- 88 `planning/audits/*` files embed some generation of the protocol, and 78 of them carry the old in-report "transfer result" wording. These are frozen copies, not live instructions.
- Only the two current-round audits mention `receipt --sandbox`.
- Implementing records: tasks `6g7srp3py9fe` and `6gk06e2g21hv`, and Thread `6gk06gktr1q5`.
- Other mentions are historical: tasks `6g7grjz5ma57`, `6g7s4k845fsb`, `6g9cz5saenme`, and `6gdx7mcqm371`; Thread `6g9czp7g9pt3`.
- The task ledgers' mutation claims (four for receipts, two for the guard) are inherited and were not re-run by me.

### Findings overview

- **Product defects:** M1 (delivery evidence after post-rename failure), M2 (generated create block fall-through and path wording), M3 (guard verification misses include and legacy-remote destinations), L1 (non-working restoration advice).
- **Coverage gap:** L2 (receipt workspace-blob check, handoff prompt and `[delivery]` note, and post-overlay assertion all survive mutation).
- **Defense in depth:** L3 (the verifier runs from the reviewer-mutable copy).
- **Taste only, not filed:**
  - The README's receipt field list (`:31`) omits `sandbox_retained`, and the brief's attestation list omits `deliverable_changed`.
  - The receipt's `source_blob` is the pre-transfer capture, which a reader may mistake for the current source blob.
  - An in-sandbox `git worktree add` is refused by the independence check without a remediation hint.

### Probe and mutation ledger

1. **Own-sandbox independence: reproduced (pass).** See the attestation section.
2. **Push matrix: reproduced (pass), plus bypasses.**
   - In a disposable fixture, these pushes all fail at the guard (rc=128, output names the guard): `push`, `push origin`, new branch, existing `feature`, `--all`, `--tags`, `--mirror`, and a single tag push.
   - `push.default` matching/current/upstream/simple are refused at the guard; `nothing` refuses for lack of refspecs.
   - `fetch origin` works, the source state digest is unchanged, and `verify` passes afterwards.
   - Weakenings beyond the eight tested ones:
     - `include.path`/`includeIf` carrying `remote.pushDefault` or `branch.main.pushRemote`, and a legacy `.git/remotes` file: **bypass** (M3).
     - Empty `pushurl` reset, `pushInsteadOf` after removing `pushurl`, remote names `a b` and `a.b*`, a dotted branch `pushRemote`, `[Remote] PushDefault`, and guard plus trailing whitespace: **refused**.
     - `remote.pushDefault='origin '`: **harmless** (verify passes, push does not reach the source).
   - Every printed restoration command was executed (L1).
3. **Linked-worktree path: reproduced (pass); GNU behaviour emulated.**
   - Pass: my own `create` from this linked worktree left the shared administration bytes unchanged (attestation section). In the fixture, `TestLinkedWorktreeSourceStaysIndependent` passes.
   - Pre-fix exclude on openrsync: `--exclude='/.git/'` fails closed with `rsync … .git: unlinkat: Directory not empty` (exit 23).
   - GNU behaviour (emulated): a PATH rsync shim emulating GNU `--delete` replacing the clone's `.git` directory with the source's `.git` file. The fixed helper's `:188` assertion refuses before any Git write, and the shared administration digest is unchanged. With that assertion removed, the baseline commit lands on the linked worktree's branch in the shared repository.
   - Real GNU rsync remains unresolved.
4. **Receipt state machine: reproduced (pass), with one falsified path (M1).**
   - `receipt` refuses before transfer and after `verify`.
   - `transfer` prints exactly the recorded receipt (`cmp` equal).
   - `receipt` → `delivery=confirmed`.
   - Repeated `verify`/`transfer` refuse with a working pointer to `receipt`.
   - A workspace edit after transfer is refused; reverting it restores confirmation.
   - An owner edit is refused.
   - Post-rename failure (unwritable state dir, and a concurrent workspace save before the `cmp`): the source is replaced, rc=2, no receipt, and `receipt` says "has not succeeded" (M1). Nothing is mistakable for success; delivered bytes are reported as undelivered.
5. **Compiler-valid mutations.** Every mutation passed `bash -n` and ShellCheck; restoration was confirmed with `git diff --quiet -- scripts` and the intended test green.
   - Receipt workspace-blob check `:367`: **survived** the intended test and the package (L2).
   - Handoff prompt `:237` plus `[delivery]` `:245`: **survived** (L2).
   - Post-overlay assertion `:188` alone: **survived**. With the exclude regressed (with or without the assertion): killed on macOS, but only by openrsync's own error (L2).
   - `.`-upstream allowance `:146`: **killed** for the intended reason.
6. **Literal walkthrough: falsified (M2).**
   - Ran the real generator (non-dry-run) and sourced its `sh` blocks verbatim. Substituting the path the prompt printed leaves the shell in the source checkout with rc=0.
   - With a correct relative path, the full walkthrough succeeds: verify → paste → verify → transfer → the owner's `[delivery]` `receipt` from the source shell → `delivery=confirmed`.
   - Every field the brief names appears in `verify`, and `transfer=succeeded` and `transferred_blob` appear in the receipt.
7. **Second pass (systemic).**
   - A self-masking verifier mutation (L3) was reproduced.
   - The block fall-through (M2) is the "appears to fail closed" pattern applied to the instructions rather than the helper.

Validation inside the sandbox: `just build` ok, focused `go test … -count=1 -v` 8/8 PASS, `shellcheck` clean, `./bin/tskflwctl -C . lint --links` pass, `git diff --check` clean.

### Surviving-hypothesis ledger

Tested negatives:

- Two simultaneous transfers both succeed with identical bytes, leave one consistent receipt, and leave no `*.review-transfer.*` residue.
- The receipt is written by mktemp plus `mv` in the same directory (0600), so it is atomic.
- `receipt_value`: duplicates resolve first-wins (the helper never writes duplicates); `=` inside a value is preserved; CRLF and a missing final newline both fail closed.
- `load_state` skips the guard, HEAD, and staging checks. This does not affect whether the receipt's bytes are true, because the receipt is a point-in-time record.
- No push-guard false positives of concern in legitimate workflows:
  - These pass: `fetch`/`fetch --prune`, an auto-tracking `branch probe origin/feature`, `branch -u origin/main`, a local `--track`, `pushDefault=origin`, stash push/pop, `remote set-head`, and user-global `remote.pushDefault`/`push.autoSetupRemote`.
  - These are refused with a working fix: a renamed `origin`, and a fetch-only GitHub remote (which really could push).
- Production audits are untracked in the source, while the tests only use a tracked deliverable. The untracked case works (walkthrough and this review).

Documented limits or unresolved:

- **Check-then-rename window (`:341-343`).** An owner edit landed there (deterministic `mv` shim) is silently overwritten, and `receipt` later confirms delivery. This is within the README's "not a repository lock" limit (`:50-52`).
- **Not tested:** real GNU rsync; Linux; SIGKILL between the source-side `mktemp` and the rename (it would leave a temp file beside the source audit).
- **Receipt not bound to itself.** The receipt's own `deliverable`/`sandbox_path` fields are not cross-checked against state. This only matters for copied or forged receipts, which `README.md:92-93` disclaims.
- **Pre-change sandboxes.** A sandbox transferred by the pre-change helper reads as "has not succeeded" under the new `receipt` (folded into M1).

## Findings

#### M1. A failed post-rename transfer delivers the report but leaves no receipt and no recovery · **Status:** fixed in fc537a52

**File:** scripts/isolated-review-workspace.sh:343 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** S

**Class:** product defect (delivery evidence). **Lens:** delivery-evidence primary.

**Consequence.** `transfer` can replace the source deliverable and then exit 2 with no receipt. From then on, the helper reports the delivered report as undelivered, and no command can recover a receipt. Exit status 2 is the same as every pre-write refusal, so the status alone cannot tell "not delivered" from "delivered, not recorded". The owner follows the `[delivery]` note, runs `receipt`, and is told `the guarded transfer has not succeeded for this workspace`, while the source already holds the reviewer's bytes. A reviewer who retries `verify`/`transfer` is told `source deliverable changed after workspace creation; preserve the workspace`, which blames the owner for the helper's own write. This falsifies intended contract 4 ("Refused or failed transfers leave no receipt and no partial source change").

**Evidence (code).** `scripts/isolated-review-workspace.sh:343` renames into the source before two failure points that exit after the source write: `:345` (`cmp -s … || die "transferred deliverable does not match the workspace"`, which never says the source was replaced) and `:347` → `write_receipt` `:285-297` (mktemp/mv in the private state dir). `show_receipt` `:358-359` maps "no receipt file" to "has not succeeded". `load_workspace` `:266-267` then refuses every retry because the recorded pre-transfer `source-blob` no longer matches.

**Reproduction (disposable fixture, reproduced).**
- Unwritable state dir: `chmod 500 <sandbox>/.git/isolated-review-workspace`, then `transfer`. Output: `mktemp: mkstemp failed … Permission denied` / `error: deliverable transferred, but its receipt could not be created …`, rc=2. Source blob after = workspace blob (`c87c2836…`), no receipt file. `receipt` → "no transfer receipt: the guarded transfer has not succeeded…", rc=2. `verify` and `transfer` → "source deliverable changed after workspace creation", rc=2.
- Concurrent editor save between the copy and the `cmp` (PATH `mv` shim appends to the workspace report on the transfer rename): rc=2, `error: transferred deliverable does not match the workspace`. Source now holds the pre-save bytes (`c87c2836…`), no receipt, and `receipt` again says not succeeded.
- Transitional trigger, same root cause: a sandbox created and transferred by the base-commit helper (`4fdf6fcc`) transfers fine (rc=0). The new helper's `receipt` against it says "the guarded transfer has not succeeded", rc=2.

**Could a reviewer mistake it for success?** No: no `transfer=succeeded` is printed. The misleading direction is the opposite one: delivered bytes reported as undelivered, with no recovery path.

**Recommendation.** Before the rename, record an in-flight marker naming the intended blob. Give post-rename failures a distinct exit status and a message saying the source was replaced. Let `receipt` (or a `transfer --record-receipt` repair) recognize source blob == workspace blob == in-flight blob as "delivered, receipt missing" and record it. Reword receipt absence as "no receipt recorded" and print the source/workspace blob comparison. Add regressions for the unwritable state dir and the concurrent-save case.

**Scope / residual uncertainty.** Needs an unusual failure (permissions, disk full, concurrent save) after the rename. Severity is Medium, not High, because the evidence fails toward "not delivered" rather than claiming false delivery. The marker design is one option; I have not prototyped it.

**Recommendation:** Record an in-flight marker, give post-rename failures a distinct status, and let receipt recover a delivered-unrecorded transfer.

**Resolution:** Same root cause as git-safety M3: the pending receipt precedes
    the rename, post-rename failures exit 3, and receipt records or diagnoses
    the interrupted delivery. The unwritable state directory now refuses with
    the source unchanged, and the editor-save case reports delivery uncertain.

#### M2. A failed create in the generated brief falls through to the shared checkout, and the brief's path wording makes create fail · **Status:** fixed in 229426ab

**File:** scripts/prepare-adversarial-review-audits.sh:161 | **Component:** scripts/prepare-adversarial-review-audits.sh
**Effort:** S

**Class:** product defect in generated instructions (pre-existing block text; in scope because this change checks the AC "A compliant reviewer can follow the generated instructions literally" and claims contract 5). **Lens:** delivery-evidence primary (literal walkthrough).

**Consequence.** A literal follower whose `create` fails lands in the shared source checkout with a successful exit status. That is the wrong-target state the whole protocol exists to prevent, and the brief's own wording causes the failure.

**Evidence.**
- The brief says to "Substitute the repository-relative assigned audit path printed in the handoff prompt" (`scripts/prepare-adversarial-review-audits.sh:157`). The handoff prompt prints `audit path`, which is absolute (`:236-237`). `scripts/README.md:142-143` admits that the reviewer must translate it. The helper rejects an absolute deliverable (`scripts/isolated-review-workspace.sh:56`).
- The create block (`prepare-adversarial-review-audits.sh:161-167`) has no failure guard. When the command substitution fails, `SANDBOX` is empty and `cd "$SANDBOX"` is `cd ""`, a successful no-op in both bash and zsh.
- Block 2 (`:184-185`) ends with `git diff`, so a failed `verify` still yields status 0.
- `SOURCE_ROOT="$(git rev-parse --show-toplevel)"` (`:161`) depends on the reviewer's cwd. This review's own harness started in a different checkout, and the coordinator had to supply explicit values.

**Reproduction (reproduced).** Run the real generator (non-dry-run) in a disposable clone, then source its extracted `sh` block 1 verbatim, with the placeholder replaced by the path the prompt printed. Output: `error: --deliverable must be repository-relative`, block rc=0, `pwd` = `toplevel` = the source checkout, `SANDBOX=[]`, identical in bash and zsh. `touch walk-probe-write` then shows `?? walk-probe-write` in the source. A second run that also failed `create` went on through block 2 (rc=0). Its report edit was written into the source's audit file. With a correct relative path, the same literal walkthrough succeeds end to end (verify → paste → verify → transfer → owner `receipt` → `delivery=confirmed`). Every field the brief names is present.

**Recommendation.** Make the blocks fail closed, for example `SANDBOX="$(…)" || { echo "create failed: report the blocker" >&2; exit 2; }` and `cd "${SANDBOX:?create failed}"`. Have the generator print `SOURCE_ROOT` and the repository-relative `AUDIT_REL` in the handoff prompt (it knows both) instead of relying on cwd and hand translation. Separate block 2's `verify` from `git diff`, or chain them with `&&`. Pin this with a test that executes the generated blocks against a forced `create` failure.

**Scope / residual uncertainty.** A reviewer who reads stderr sees the error, and the brief says to report the blocker. The harm needs a reviewer who trusts exit status or cwd. Severity is Medium. Owner triage could argue High under the brief's "reachable wrong-target write" definition.

**Recommendation:** Guard the generated blocks (exit on empty SANDBOX) and print SOURCE_ROOT plus the repository-relative AUDIT_REL in the handoff prompt.

**Resolution:** The handoff prints exact SOURCE_ROOT and AUDIT_REL, every
    generated block is guarded with SANDBOX:?, and commands chain with &&. The
    generator test executes the blocks with the old absolute path (refused, no
    fall-through) and runs a full literal walkthrough.

#### M3. Verify attests an intact push guard while include-file or legacy-remote settings push to the source · **Status:** fixed in 1ece6c8c

**File:** scripts/isolated-review-workspace.sh:148 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** XS

**Class:** product defect (guard verification). **Lens:** git-safety owns the push guard. Filed here because evidence floor 2 applies to both reviewers and the consequence is a false `verify=passed`/`push_guard=` attestation.

**Consequence.** `verify` and `transfer` pass, printing `push_guard=isolated-review-push-disabled://…`, while an ordinary `git push` from the sandbox writes the source repository. `scripts/README.md:40-43` and contract 2 say verify refuses any workspace-local `remote.pushDefault`/`branch.<name>.pushRemote` that names a path. The documented limits (`README.md:45-47`) name explicit URLs, `-c`/environment overrides, and user-wide or system config. They do not name workspace-local include files or legacy remote files.

**Root cause.** `check_push_guard` reads a narrower view than Git's push resolution. `scripts/isolated-review-workspace.sh:148` uses `git config --local --get-regexp`, and `--local` defaults to `--no-includes`, so values from `include.path`/`includeIf` in the sandbox's own `.git/config` are invisible. `:134` enumerates remotes with `git remote`, which omits legacy `$GIT_DIR/remotes/<name>` and `$GIT_DIR/branches/<name>` definitions that Git 2.54 still honours.

**Reproduction (disposable fixture, Git 2.54.0, reproduced).** Each case starts from a fresh sandbox with a report edit. The include file lives outside the sandbox.
- `git config include.path <f>` with `[remote] pushDefault = <source>` → verify rc=0, and `git push` reaches the source.
- Same include with `[branch "main"] pushRemote = <source>` → verify rc=0, and push reaches the source.
- `includeIf.gitdir:<sandbox>/.git.path` with the same pushDefault → verify rc=0, and push reaches the source.
- `.git/remotes/src` containing `URL: <source>` → `git remote` lists only `origin`, verify rc=0, and `git push src` creates `refs/heads/legacy-leak` in the source.

Detected controls: include that adds an origin `pushurl`, empty `pushurl` reset, pushurl removal plus pushInsteadOf, remote names with space or `*`, dotted branch names, mixed-case `[Remote] PushDefault`, and guard plus trailing whitespace are all refused.

**Recommendation.** Use `git config --local --includes --get-regexp …`. I checked that it surfaces the included `remote.pushdefault` in the fixture. Also enumerate or refuse `$GIT_DIR/remotes/*` and `$GIT_DIR/branches/*`. If any of these is deliberately out of scope, add it to the README's limits. Add both weakenings to `TestVerifyAndTransferRejectAWeakenedPushGuard`.

**Scope / residual uncertainty.** These are deliberate-looking configurations, not likely accidents, so this is Medium rather than High. The guard URL itself is not bypassed for `origin`. The gap is that post-hoc verification attests a guard that is not in force.

**Recommendation:** Read push-destination settings with --local --includes and refuse legacy remotes/branches files, or document them as limits.

**Resolution:** Same root cause as git-safety M2: workspace-scoped reads now
    include include files and config.worktree, and legacy remotes and branches
    files are refused.

#### L1. Some push-guard restoration commands do not restore the guard or do not parse · **Status:** fixed in 1ece6c8c

**File:** scripts/isolated-review-workspace.sh:142 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** S

**Class:** product defect in repair advice, plus a coverage gap. **Lens:** hostile angle "Does any refusal message suggest a command that fails?"

**Consequence.** Contract 2 promises a working restoration command with each refusal. Some printed commands succeed without restoring anything, write unrelated config, or fail to parse.

**Evidence.**
- `scripts/isolated-review-workspace.sh:142` prints `config --replace-all remote.$remote.pushurl …` with `$remote` unquoted, followed by an "(and remove any url.*.insteadOf rule …)" note that names no command.
- `:147` wraps `$key` in single quotes without escaping.
- The test (`internal/tools/reviewworkspace/review_workspace_test.go:266`) only checks that the output contains `push guard` and `git -C '<sandbox>'`. It never runs the advice.

**Reproduction (each printed command executed, then `verify` re-run, in bash and zsh).**
- "disabled URL rewritten", one of the owner's eight tested cases (`url.<source>.insteadOf=<guard>`): the command exits 0 but `verify` still refuses (rc=2). The required `git config --unset url.<source>.insteadOf` is never printed.
- Remote `a b`: the command exits 0 but writes junk config `remote.a = b.pushurl`, and verify still refuses.
- Remote `a.b*`: works in bash; zsh fails with `no matches found: remote.a.b*.pushurl`.
- `branch.it's.pushRemote`: bash says `unexpected EOF while looking for matching '`; zsh says `unmatched '`.

The other eleven cases restore, and `verify` passes after them. That covers all seven remaining tested weakenings, the empty pushurl, and `a.b*` under bash. No printed command re-enables pushing.

**Recommendation.** Quote the interpolated names with `printf %q`. For insteadOf/pushInsteadOf, print the actual offending keys (from `git config --get-regexp '^url\..*\.(push)?insteadof$'`) as `--unset` commands. Make the weakening test eval each printed command and then require `verify` to pass.

**Scope.** Low. Remote names with spaces and apostrophes need hand-edited config. The insteadOf case is one of the shipped test cases.

**Recommendation:** Quote interpolated names, print the offending insteadOf keys as unset commands, and have the test execute each printed command.

**Resolution:** Same root cause as git-safety L1; the weakening test now runs
    each printed restore line, including the insteadOf and quoted-name cases.

#### L2. New fail-closed branches survive compiler-valid mutations · **Status:** fixed in fc537a52, 229426ab, 98cdd131

**File:** internal/tools/reviewworkspace/review_workspace_test.go:511 | **Component:** internal/tools/reviewworkspace
**Effort:** S

**Class:** coverage gap. **Lens:** evidence floor 5 (compiler-valid mutations). Each mutation passed `bash -n` and ShellCheck 0.11.0. Each was restored from a backup outside the sandbox, after which the intended test passed again.

**Surviving mutations.**
1. **Receipt workspace-blob check.** `scripts/isolated-review-workspace.sh:367`, `git -C "$sandbox_root" hash-object` changed to `git -C "$source_root" hash-object`. `TestTransferReceiptIsTheOnlyDeliveryEvidence` and the whole package still pass. In a fixture, a workspace edit after transfer then yields `delivery=confirmed`; at baseline it yields `workspace deliverable changed after transfer`. No test edits the workspace after transfer.
2. **Generator handoff prompt and `[delivery]` note.** `scripts/prepare-adversarial-review-audits.sh:237` was rewritten to "Put the transfer result in your report.", and `:245` to "trust the report's attestation." `TestGeneratedBriefFollowsTheReceiptContract` and the package still pass. The test runs only `--dry-run`, which exits at `:226-228` before either `printf`. The prompt and note this change adds are therefore never asserted.
3. **Post-overlay independence assertion.** Replacing `:188` with `true` alone survives `TestLinkedWorktreeSourceStaysIndependent`, because the fixed exclude at `:185` already prevents the hazard. The coordinated mutations (`--exclude='/.git/'` alone, or together with dropping `:188`) are killed on macOS. The failure comes from openrsync's own `rsync … error: .git: unlinkat: Directory not empty` (exit 23), not from the assertion. With a PATH rsync shim emulating GNU's replace-directory-with-file behaviour (emulated, not GNU rsync), the fixed helper's `:188` assertion stops `create` before any Git write. Without the assertion, the baseline commit lands on the linked worktree's branch in the shared repository (`bb455e4 chore: capture isolated review baseline`). The assertion matters and is unpinned on macOS.

**Killed for the intended reason (control).** Dropping the `"$value" == "."` allowance at `:146` makes `TestPushGuardAllowsLocalUpstreams` fail with `push guard bypassed: branch.local-copy.remote names '.'`.

**Recommendation.** Add a post-transfer workspace-edit case. Run the generator non-dry-run with the fake CLI also answering `audit path`, and assert the prompt and `[delivery]` text. Exercise the post-overlay assertion directly, for example with an rsync shim that replaces `.git`, so it is pinned on every platform.

**Scope.** Low: the behaviour is currently correct. These are regression gaps for newly shipped fail-closed branches.

**Recommendation:** Add a post-transfer workspace edit, a non-dry-run generator assertion, and a direct post-overlay assertion probe.

**Resolution:** The receipt workspace-blob check, handoff prompt, delivery note,
    and post-overlay assertion each gained a regression; the corresponding
    mutations now fail.

#### L3. The verifier runs from the reviewer-mutable sandbox copy and can attest its own mutation · **Status:** fixed in 229426ab

**File:** scripts/prepare-adversarial-review-audits.sh:184 | **Component:** scripts/prepare-adversarial-review-audits.sh
**Effort:** XS

**Class:** defense-in-depth (systemic, second pass). **Lens:** "boundaries that only appear to fail closed" and "test helpers that can mask a defect class".

**Consequence.** The brief, the README, and the tests all run `verify`/`transfer`/`receipt` from the sandbox's copy of the helper (`prepare-adversarial-review-audits.sh:184,195-196,205`; `scripts/README.md:15-17`; `review_workspace_test.go:112-113`). That is the very file reviewers of this helper are told to mutate. One unrestored, syntax-valid mutation can make the verifier attest its own cleanliness. The report then carries `verify=passed` while the sandbox still holds stray probes, and the brief relies on verify to catch exactly that.

**Reproduction (disposable fixture, reproduced).** In a fixture sandbox, leave `src/kept.txt` edited. Then change the dirty-tree check in the sandbox's copy of the helper (`:264`) to `[[ -z "" ]] || die` (passes `bash -n`). `git status` shows `M planning/audits/review.md`, `M scripts/isolated-review-workspace.sh`, and `M src/kept.txt`. The sandbox copy's `verify` prints `verify=passed`/`transfer=pending` (rc=0). The unmutated source copy, `"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox <sandbox>`, refuses with `workspace has changes outside the deliverable` and lists both stray files.

**Recommendation.** Have the protocol run `verify`/`transfer` from `$SOURCE_ROOT/scripts/isolated-review-workspace.sh`, since the source is read-only by protocol. Alternatively, have the helper refuse when its own bytes differ from `git show <baseline>:scripts/isolated-review-workspace.sh`, which catches accidental mutation but not a mutation of that check.

**Scope.** Low. Transfer still copies only the deliverable, so the source is not harmed. The false evidence is the verification attestation, and it matters most for reviews of this helper itself.

**Recommendation:** Run verify and transfer from the read-only source copy, or refuse when the running helper differs from the baseline blob.

**Resolution:** The protocol, README, and tests run verify, transfer, and
    receipt from the source checkout copy, which reviewers cannot mutate; the
    generator test rejects any brief that runs the sandbox copy.
