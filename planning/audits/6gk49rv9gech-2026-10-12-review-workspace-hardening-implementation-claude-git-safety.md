---
schema: 1
id: 6gk49rv9gech
bucket: open
area: review-workspace-hardening-implementation-claude-git-safety
date: "2026-10-12"
updated_at: "2026-10-12"
---
# Audit: Review-workspace hardening implementation — claude-git-safety — 2026-10-12

> Reviewer assignment: claude-git-safety. This document is the review brief and the only file the reviewer should update.
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

**Verdict: needs changes.** Reviewer: claude-git-safety. Primary lens: Git targeting and source
preservation. Bounded cross-check: the receipt's fail-closed paths.

**Snapshot:** source linked worktree `.claude/worktrees/harden-isolated-review-workspaces`, branch
`worktree-harden-isolated-review-workspaces`, HEAD `ef77ffd5b54a4407a88cf7d6b7157406d9618469`
(`f5e0aaf1`, `22a27650`, `ef77ffd5` over `4fdf6fcc`). It was captured as sandbox baseline
`7dd8cdd1523ec3b89893efa2e0e7b128dabdbb1a`, with this audit's source blob
`240f7343c4a9af65c07191695ecc9bce17527ea0`.

The slice's fixes work for the cases they target:

- every ordinary push form fails at the guard without touching the source;
- the eight tested weakenings are refused;
- the corrected overlay keeps a linked-worktree source's shared administration byte-identical.

The helper can still write into the source repository through paths the tests do not model:

- **H1:** the overlay copies nested Git metadata (for example, ignored linked worktrees under this
  repository's `.claude/`). One ordinary `git commit` inside such a copy advanced a source branch,
  while `verify` kept attesting independence.
- **M1:** `create` rewrites the source index through `git diff` auto-refresh.
- **M2:** the guard's destination check ignores included and worktree-scoped config.
  `push_guard=…` and `verify=passed` can then coexist with a plain `git push` that creates a source
  branch.

On the receipt cross-check:

- **M3:** a failure after the source rename leaves an unrecoverable "delivered, not recorded"
  state. `receipt` then reports it as "not succeeded".

**Limits:**

- Host: macOS (Darwin 27), `/bin/bash` 3.2.57, Git 2.54.0, and `/usr/bin/rsync` = openrsync
  (protocol 29).
- No GNU rsync was found on PATH, in `/opt/homebrew`, `/usr/local`, or `/opt/local`. GNU-specific
  overlay behavior and the Linux CI run are therefore not reproduced.
- Full race `just test` and `just lint` are inherited from the owner, not re-run.
- I did not read the sibling delivery-evidence report.

**Probe environment:** all probes ran in disposable repositories under `$P` =
`/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/8903e980-d39f-489c-bec5-5af1565e0a2b/scratchpad/probe.eyfFDV`.
Git configuration was redirected (`GIT_CONFIG_GLOBAL=$P/home/gitconfig`,
`GIT_CONFIG_NOSYSTEM=1`), inherited `GIT_*` variables were unset, and `HOME` and `TMPDIR` were
redirected.

### Isolation attestation

Final `verify` output, copied verbatim. `transfer=pending` is the truthful value when the report is
finalized; delivery evidence is the receipt in the handoff message.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.TnG02W
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.TnG02W/.git
baseline_commit=7dd8cdd1523ec3b89893efa2e0e7b128dabdbb1a
source_blob=240f7343c4a9af65c07191695ecc9bce17527ea0
source_fingerprint=89b672cbfde3869f0f533d066cff8c7638e339e1
deliverable=planning/audits/6gk49rv9gech-2026-10-12-review-workspace-hardening-implementation-claude-git-safety.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

Own independence probe, run immediately after `create` (evidence floor 1, reproduced):

- `git rev-parse --absolute-git-dir` → `<sandbox>/.git`, and `--git-common-dir` → `.git`.
  `.git` is a plain directory, not a symlink, and there is no `objects/info/alternates` file.
- `git remote get-url --push --all origin` → `isolated-review-push-disabled://source-is-read-only`
  (one line). The fetch URL is the source worktree path, local config sets no
  `pushDefault`/`pushRemote`, and `worktree list` shows only the sandbox.
- `find . -name .git -not -path ./.git` is empty, so H1 does not affect this review's own workspace.
- Creation left the source unchanged. Before/after `--no-optional-locks` reads matched:
  - `status --porcelain=v1 -b --untracked-files=all` (the two untracked audits only);
  - HEAD `ef77ffd5…` on `refs/heads/worktree-harden-isolated-review-workspaces`;
  - the shared `worktree list --porcelain` (three worktrees).

  SHA-256 hashes also matched for the shared `.git/config` and `packed-refs`, the loose refs, every
  file under `.git/worktrees/harden-isolated-review-workspaces/` (including `index`), the
  loose-object count (101), and this audit's bytes.

### Verified consumer inventory

Live instructions and executable consumers, verified with `git grep` in the sandbox:

- `scripts/isolated-review-workspace.sh`:
  - usage text at 13-42;
  - `create` output at 218-220;
  - `attest` at 273-278: `sandbox_path`, `git_dir`, `baseline_commit`, `source_blob`,
    `source_fingerprint`, `deliverable`, `push_guard`, `deliverable_changed`, `verify=passed`,
    `transfer=`;
  - `write_receipt` at 281-299: `transferred_blob`, `source_deliverable`, `sandbox_retained`,
    `transferred_at`;
  - `show_receipt` at 353-372, with `delivery=confirmed` at 371;
  - refusal messages at 136, 142, 147, 249 and 359.
- `scripts/prepare-adversarial-review-audits.sh`:
  - usage at 20-21;
  - injected protocol at 147-209: `create` 163, `verify` 184 and 195, `transfer` 196, `receipt`
    205, field list 188-191, receipt contract 199-206;
  - handoff prompt at 237, `[handoff freeze]` at 241, `[delivery]` at 245.
  - The `--dry-run` exit at 226-229 comes before 237-245.
- `scripts/README.md`: example 9-18 (all four commands), overlay and linked worktree 20-26,
  receipt 28-35, guard 37-43, limits 45-48, generator 65-73, evidence contract 85-93, path
  translation 142-145.
- `AGENTS.md:83`: a routing pointer to the scripts guide; it names no command or key.
- `internal/tools/reviewworkspace/review_workspace_test.go`:
  - output keys asserted at 372, 376, 415-416, 424, 428, 436, 497 and 552;
  - receipt path at 392;
  - all eight named tests exist (180, 223, 281, 338, 349, 409, 460, 511) and pass.
- `.github/workflows/ci.yml:112` runs the package through `go test -p 4 -race ./...` on
  `ubuntu-latest` (line 95). That this exercises GNU rsync is inferred, not verified.
- Live generated briefs for this round: this audit, and the sibling `6gk49rvj7rhe`. I saw the
  sibling only in a `git grep -l` listing.
- No references exist in `docs/`, `routines/`, the root `README.md`, or `justfile`.

Historical records, not live instructions:

- 86 earlier audits embed earlier generated protocols. 72 of them carry the superseded
  "transfer result—in the report" wording.
- Tasks `6g7grjz5ma57`, `6g7srp3py9fe`, `6g9cz5saenme` and `6gk06e2g21hv`.
- Threads `6g9czp7g9pt3` and `6gk06gktr1q5`.

Contract consistency:

- The brief's field list (prepare script 188-191, README 85-88) matches `attest` output.
- The README's receipt keys (31-34) match `write_receipt` and `show_receipt`.
- Usage text, README 37-48 and the injected protocol state the same guard limits: explicit URL,
  `-c` or environment overrides, and user-wide or system config.
- None of them says that included or worktree-scoped files go unchecked (M2), or that ignored
  content, including nested Git metadata, is copied (H1).

### Evidence floor

#### 1. Own-sandbox independence: reproduced

See the attestation above.

#### 2. Push attempts: matrix reproduced; "verify refuses any workspace-local destination setting" falsified

The fixture `$P/src` is a non-bare repository with branches `main` and `feature` and tag `v0`.

From a fresh sandbox, all 17 ordinary push forms fail with rc=128 and the guard's remote-helper
error. The source's full SHA-256 inventory of admin and working files (40 entries) is identical
before and after. The forms:

- default `git push`, `git push origin`;
- `HEAD:refs/heads/leak`, `HEAD:refs/heads/feature`, a new branch, `-u`;
- `--all` with and without a remote, `--tags` both ways, `--mirror`, `--force`;
- a delete refspec, a tag, `--delete`, `--atomic`, `--dry-run`.

`git fetch origin` exits 0. The control `git push <source path> HEAD:refs/heads/explicit-control`
does create the ref, so the probe is live. A first matrix attempt was invalid (zsh does not
word-split `$args`); it was rerun with `${=args}`.

Weakenings beyond the existing test, each in a fresh fixture:

- **Refused (verify rc=2):**
  - an appended empty `pushurl` (Git 2.54 resets the list, so `get-url` falls back to the source);
  - `pushurl` removed, plus `url.<src>.pushInsteadOf`;
  - `pushInsteadOf` to another disabled scheme;
  - the guard URL plus a trailing space;
  - an unguarded remote named `r.*`;
  - a dotted branch `fix/v1.2` with a path `pushRemote`;
  - case-folded `[Remote] PushDefault`;
  - a multi-line `pushDefault` value;
  - an included `[remote "origin"] pushurl = <src>`.
- **Harmless (verify rc=0, push fails at the guard):**
  - `push.default` set to nothing, current, upstream, simple or matching;
  - a guarded remote named `a b`, used as `pushDefault`;
  - a guarded remote listed before `origin`.
- **Bypass (verify rc=0 with `push_guard=` and `verify=passed`; plain `git push` targets the source
  path):**
  - `include.path` → `remote.pushDefault`;
  - `include.path` → `branch.main.pushRemote`;
  - `includeIf.onbranch:main.path` → `pushRemote`;
  - `extensions.worktreeConfig` with `config --worktree remote.pushDefault`.

  End to end on a new branch, the source received `* [new branch] review-notes -> review-notes`
  (M2).

#### 3. Linked-worktree path: fix reproduced; new negative for the pre-fix code

With the fixed helper, I ran `create` from a linked worktree on `review-branch` (the main checkout
is on `main`) that had a branch commit, a dirty edit, and an untracked file:

- The shared `.git` inventory (49 files, including `worktrees/linked/index`) and the worktree's
  files are unchanged.
- The sandbox owns its `.git`, and its baseline parent is the linked worktree's head.
- The sandbox branch is `review-branch`, and the dirty and untracked state was captured.

GNU rsync is unavailable, so the pre-fix `--exclude='/.git/'` was run with openrsync only:

- On a cloned-repository fixture it fails closed: `.git: unlinkat: Directory not empty`, rc=23, the
  sandbox is removed, and the admin is unchanged.
- On a `git init` fixture shaped like the Go test's, openrsync replaced `.git` under the real
  pre-fix helper `22a27650`. The helper then committed `chore: capture isolated review baseline`
  onto the source's `review-branch` (`da4f45e` → `4face93`) before dying at the post-baseline
  assertion (L3).

I did not isolate why the outcome depends on repository shape.

#### 4. Receipt state machine: success paths reproduced; "no partial source change" falsified

| Step | Result |
|---|---|
| `receipt` before any transfer | rc=2, "no transfer receipt" |
| `transfer` with no delta | rc=2 |
| `verify` | rc=0, `transfer=pending` |
| `receipt` after verify | rc=2 |
| `transfer` | rc=0, `transfer=succeeded`, `transferred_blob=c87c283…` |
| `receipt` | rc=0, `delivery=confirmed` |
| repeated `verify` and `transfer` | rc=2, "already transferred" |
| workspace edit, then `receipt` | rc=2, "workspace deliverable changed after transfer" |
| workspace restored to the transferred bytes, then `receipt` | rc=0 |
| owner edit, then `receipt` | rc=2, "no longer matches" |

Failure after the source rename: I made the private state directory unwritable (`chmod 0500`)
after `verify`.

- `transfer` exits rc=2 with "deliverable transferred, but its receipt could not be created".
- The source now holds the workspace blob, and there is no receipt and no stray temp file.
- `receipt` reports "the guarded transfer has not succeeded", both before and after permissions are
  restored.
- Retrying `verify` or `transfer` reports "source deliverable changed after workspace creation".

A reader of `receipt` or of the retry output can mistake a delivered report for one that was never
delivered, or for owner drift (M3).

#### 5. Compiler-valid mutations

Two kills (X1, and X3 when coordinated) and four survivors (X2, X3 alone, X4, X5); see the ledger below. Mutation labels X1-X5 are distinct from finding codes.

#### 6. Literal walkthrough: reproduced

I followed the generated brief's three command blocks verbatim, substituting only `AUDIT_REL`,
from a shell whose cwd was inside a disposable source:

- create, verify, `git diff`, paste, verify, transfer and receipt all exit 0;
- `verify` fields are byte-identical before and after pasting the attestation;
- every field the brief names appears. `verify` prints `sandbox_path`, `git_dir`,
  `baseline_commit`, `source_blob`, `source_fingerprint`, `deliverable`, `push_guard`,
  `deliverable_changed`, `verify` and `transfer`. `transfer` adds `transferred_blob`,
  `source_deliverable`, `sandbox_retained` and `transferred_at`. `receipt` adds `delivery`.

Caveat: block 1 derives `SOURCE_ROOT` from the cwd. This round's coordinator had to supply exact
values, because the reviewer started in the main checkout.

#### Contract 6, test isolation: reproduced

I ran the package with `GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE` and `GIT_OBJECT_DIRECTORY`
pointed at a decoy repository, plus a decoy `GIT_CONFIG_GLOBAL`. It passes and leaves the decoy
byte-identical. `newFixture` drops `GIT_*`, `HOME` and `TMPDIR` (test lines 54-65).

### Probe and mutation ledger

Each mutation was applied with `sed` in the sandbox, and `bash -n` and ShellCheck 0.11.0 accepted
it. Each was then restored with `/bin/cp -f` from a backup outside the sandbox. After every
restore, `git diff --quiet` was clean and the targeted test was green.

| # | Invariant | Edit | Intended test | Result |
|---|---|---|---|---|
| X1 | dot-upstream allowance | drop the `"$value" == "."` clause (146) | `TestPushGuardAllowsLocalUpstreams` | **killed**: verify rejects `branch.local-copy.remote` = `.` |
| X2 | `branch.<name>.remote` destination check | remove the `remote` alternative from the 148 regexp | `TestVerifyAndTransferRejectAWeakenedPushGuard` (8 subtests) and the upstream test | **survived**. Under the mutant, `branch.leak.remote=<src>` with `merge` passes verify, and `git push` creates `refs/heads/leak` in the source |
| X3 | post-overlay independence assertion | 188 → `:` | `TestLinkedWorktreeSourceStaysIndependent` | **survived** alone. With the pre-fix exclude at 185 also restored, it is killed, but only at the post-baseline assertion (203), after the baseline commit ran against the shared repository |
| X4 | receipt workspace-blob check | 367 `$sandbox_root` → `$source_root` | `TestTransferReceiptIsTheOnlyDeliveryEvidence`, `TestFailedTransfersRecordNoReceipt`, linked-worktree test | **survived** |
| X5 | receipt sentence in the handoff prompt | delete it from 237 | `TestGeneratedBriefFollowsTheReceiptContract` | **survived**: the dry run exits at 226-229, before 237 and 245 |

An earlier X1 attempt through the Edit tool did not apply. Its green run was discarded as an
invalid probe.

Other probes:

- the `create` index rewrite (finding M1);
- nested-worktree copies (H1);
- every restoration command run literally (L1);
- `verify`, `transfer` and `receipt` against a stat-dirty source, which left its admin unchanged.

### Validation (sandbox, independent)

- `just build`: ok (`v0.23.0-48-g7dd8cdd1`).
- `go test ./internal/tools/reviewworkspace/ -count=1 -v`: ok. All 8 tests and 8 weakening
  subtests pass.
- `shellcheck` on both scripts: clean.
- `./bin/tskflwctl -C . lint --links`: rc=0.
- `git diff --check`: clean.
- Full `just test` and `just lint`: inherited from the owner.
- A disposable `GOCACHE` outside the sandbox was used.

### Classification summary

- **Product defects:**
  - H1: nested Git metadata is copied, while `verify` attests independence. Pre-existing overlay
    scope.
  - M1: `create` writes the source index. Pre-existing `source_fingerprint`.
  - M2: the destination check misses included and worktree config. New in this slice.
- **Delivery-evidence cross-check:** M3, a failure after the rename is unrecoverable and
  misreported.
- **Refusal advice:** L1.
- **Coverage:** L2.
- **Planning-record accuracy:** L3.
- No taste findings.

### Surviving-hypothesis ledger

Tested negatives:

- `verify`, `transfer` and `receipt` do not write source admin, even on a stat-dirty source.
- The tests cannot reach a developer's checkout or configuration, even with hook-style `GIT_*`
  variables set.
- The guard URL fails before contacting anything, for every ordinary push form.
- Regex or whitespace characters in remote names, and every `push.default` variant, are refused or
  harmless.
- The empty-`pushurl` reset is refused, and its printed remedy works.
- The remedies for a removed origin, an extra or empty `pushurl`, and a path-valued
  `pushDefault`/`branch.<name>.remote` work when run literally.
- `load_state` (used by `receipt`) calls `assert_independent_clone` (228). It skips the push-guard
  and HEAD checks, which do not bear on receipt truth.
- `receipt_value` takes the first matching key, and the first `transfer=` line is the one `attest`
  writes.
- None of the 49 source branches captured in this sandbox has a stray baseline commit. Reflogs and
  dangling objects were not inspected.

Unresolved:

- GNU rsync behavior for both excludes, and whether CI actually exercises it.
- `index.lock` contention between M1's refresh and a concurrent owner command: 0 collisions in 193
  attempts, and the result is timing-dependent.
- Source-repository `core.fsmonitor`, textconv or LFS filters running during the source-side
  `git diff` or `hash-object`: not probed.
- An owner edit landing between the hash check and the rename (341-343) is overwritten. The README
  documents this as a handoff-freeze limit; I did not reproduce it.
- Submodules whose `.git` file uses an absolute gitdir are in H1's class but were not built.

## Findings

#### H1. Overlay copies nested Git metadata that writes source refs while verify attests independence · **Status:** open

**File:** scripts/isolated-review-workspace.sh:185 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** M · **Urgency:** soon

**Classification:** product defect in Git targeting and source preservation. The overlay scope is
pre-existing (`main`'s overlay also copies everything except the root `.git`), but the defect bears
directly on whether the helper can still write into the source.

**Consequence:** `create` overlays the source with `rsync -a --delete --exclude='/.git'`
(`scripts/isolated-review-workspace.sh:185`). That excludes only the root entry and also copies
ignored content. A nested linked worktree's `.git` file, with its absolute `gitdir:` into the
shared repository, therefore lands in the sandbox still bound to the source.
`assert_independent_clone` (105-124) inspects only the root `.git`. `verify` then attests
`verify=passed` while the sandbox contains Git metadata that writes the source's refs, objects and
worktree administration.

This repository's own layout reaches it. `.gitignore:24` ignores `.claude/`, and the main checkout
holds `.claude/worktrees/harden-isolated-review-workspaces/.git` and
`.claude/worktrees/tui+dashboard-counts-staleness/.git`, both with absolute gitdirs. A review whose
`SOURCE_ROOT` is the main checkout would copy both, including another agent's in-progress worktree.

**Reproduction, ignored case** (`$P/nested.xOpkZ2`):

1. The main checkout ignores `.claude/` and holds
   `git worktree add -b other-agent-work main/.claude/worktrees/other`.
2. `create --source main` exits 0. The sandbox contains `.claude/worktrees/other/.git` =
   `gitdir: $P/nested.xOpkZ2/main/.git/worktrees/other`, and `verify` exits 0.
3. `git -C <sandbox>/.claude/worktrees/other commit --allow-empty` advanced the source branch
   `other-agent-work` from `21b0bba` to `33f4a36`. It also changed the source's
   `refs/heads/other-agent-work` and its reflog, wrote a new object, and changed
   `worktrees/other/{index,logs/HEAD,COMMIT_EDITMSG}`.
4. `verify` still passes after that write.

**Reproduction, non-ignored case** (`$P/nested2.PQWYjK`), with no reviewer command:

1. The nested worktree `wt/other` is not ignored.
2. `source_fingerprint` (77-91) prints `fatal: Unable to hash wt/other/` three times, yet `create`
   exits 0. bash 3.2 clears `errexit` inside `$(...)`, which I verified directly: `x="$(f)"`
   continued past `false`.
3. The baseline commits a gitlink (`160000 ... wt/other`).
4. A later plain `verify`, whose `git status` (263) recurses into the gitlink, rewrote the source's
   `worktrees/other/index`.

**Suggested change:** After the overlay, reject or strip every nested `.git`, file or directory, or at
least any whose resolved gitdir lies outside the sandbox. Better, overlay only tracked and
non-ignored untracked paths: `git ls-files -z --cached --others --exclude-standard` fed to rsync
`--files-from` with `--from0`, plus deletions. Make `source_fingerprint` check each command's status
and fail closed. Add fixtures with an ignored and a non-ignored nested worktree.

**Scope and residual uncertainty:** This review's own sandbox is unaffected: its source is the
linked worktree, and it has no nested `.git`. The ignored case needs a Git command run inside the
copied directory, possibly implicitly (VCS stamping, an editor or LSP, a shell prompt); I did not
test those implicit callers. Absolute-gitdir submodules are in the same class but were not built.

**Recommendation:** Reject or strip nested .git entries after the overlay (or overlay only non-ignored paths) and fail the fingerprint closed.

#### M1. create rewrites the source index through git diff auto-refresh · **Status:** open

**File:** scripts/isolated-review-workspace.sh:83 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** XS · **Urgency:** soon

**Classification:** product defect, pre-existing (`4fdf6fcc` lines 65 and 69 have the same
`source_fingerprint`). It contradicts this slice's linked-worktree contract and the invariant that
`TestLinkedWorktreeSourceStaysIndependent` states.

**Consequence:** `source_fingerprint` runs porcelain `git -C "$root" diff --binary ... HEAD` (83)
twice against the source (178 and 189). With `diff.autoRefreshIndex` at its default, Git rewrites
the source index, taking `index.lock`, whenever a tracked file is stat-dirty but content-identical.
That is common after a checkout or rebase, or after a formatter, an editor re-save, or `touch`. For a
linked worktree, the rewritten file is the shared repository's `worktrees/<name>/index`. The handoff
explicitly expects the owner to keep using the source concurrently.

**Reproduction:**

- In `$P/idx.fhEnGI`, after `touch src/kept.txt scripts/isolated-review-workspace.sh`, `create`
  changed the source `.git/index` SHA-256 from `a600c097...` to `7c1e39cf...`.
- Isolated: the exact `git diff` invocation rewrites the index, and adding `--no-optional-locks`
  still rewrites it; `-c diff.autoRefreshIndex=false` does not.
- Linked worktree (`$P/lw3.1MshSt`): the shared `worktrees/linked/index` changed from
  `6d95ed2e...` to `34fa0b29...`.

**Suggested change:** Pass `-c diff.autoRefreshIndex=false --no-optional-locks` to every source-side
Git read, or use plumbing that never writes. Add a regression that makes source files stat-dirty
before `create`, then compares the index bytes.

**Scope and residual uncertainty:** Only the stat cache changes; the staged tree is identical. Lock
contention with an owner command is timing-dependent: 0 of 193 attempts collided.

**Recommendation:** Run source-side Git reads with -c diff.autoRefreshIndex=false --no-optional-locks and add a stat-dirty regression.

#### M2. Push-destination check ignores included and worktree config so a passing verify coexists with a source push · **Status:** open

**File:** scripts/isolated-review-workspace.sh:148 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** S · **Urgency:** soon

**Classification:** product defect in push-destination resolution, new in this slice.

**Consequence:** `check_push_guard` reads destination settings with
`git config --local --get-regexp ...` (148). A specific-file scope disables `include.*` processing
and excludes `config.worktree`, while `git push` resolves the effective value. A workspace-local
`include.path`, `includeIf` or `extensions.worktreeConfig` setting can therefore name the source path
as `remote.pushDefault` or `branch.<name>.pushRemote`. `verify` and `transfer` then print
`push_guard=isolated-review-push-disabled://source-is-read-only` and `verify=passed`, while a plain
`git push` writes source refs. This falsifies intended contract 2 and `scripts/README.md:40-43`.

Separately, the `branch.<name>.remote` alternative is unpinned. Mutation X2 (dropping it from the
regexp) survives every test, and the code it removes guards a real source write.

**Reproduction** (`$P/w.XoVN9i`):

1. `inc.cfg` contains `[remote] pushDefault = <source>`, and `git config include.path` points at
   it.
2. The `--local` regexp sees only `branch.main.remote origin`, and `verify` exits 0.
3. `git switch -c review-notes && git push` creates `* [new branch] review-notes -> review-notes` in
   the source.

`verify` also passes for an include that sets `branch.main.pushRemote`, for
`includeIf.onbranch:main.path`, and for `config --worktree remote.pushDefault`.

**Suggested change:** Resolve the destination the way `git push` does. Options:

- `git for-each-ref --format='%(push:remotename)' refs/heads`, which reports the source path here;
- `git config --includes --get-regexp ...` without a file scope, which catches both the include and
  the worktree-config cases (`--local --includes` misses `config.worktree`).

Add these cases, plus `branch.<name>.remote=<path>`, to the weakening table.

**Scope:** This is an accident-prevention guard, and the bypass needs an indirection unlikely to
happen by accident. When it does happen, the attestation is false. User-wide and command-line
overrides remain documented limits.

**Recommendation:** Resolve effective push destinations (%(push:remotename) or --includes without a file scope) and pin branch.<name>.remote.

#### M3. A failure after the source rename leaves delivered-but-unrecorded state that receipt reports as not succeeded · **Status:** open

**File:** scripts/isolated-review-workspace.sh:343 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** S · **Urgency:** eventually

**Classification:** delivery-evidence product defect on the receipt's fail-closed path (bounded
cross-check; it overlaps the sibling lens).

**Consequence:** `transfer_deliverable` renames into the source (343) before calling
`write_receipt` (347). If the receipt cannot be written, or `cmp` fails at 345, the command exits 2
with the source already replaced. Intended contract 4 says a failed transfer leaves no partial
source change.

Afterwards, no helper command distinguishes "delivered, not recorded" from "not delivered":

- `receipt` says "the guarded transfer has not succeeded";
- `verify` and `transfer` blame "source deliverable changed after workspace creation", which reads
  as owner drift.

**Reproduction** (`$P/rcptfail.GBpM82`):

1. After `verify`, run `chmod 0500 <sandbox>/.git/isolated-review-workspace`.
2. `transfer` exits rc=2 with "deliverable transferred, but its receipt could not be created".
3. The source blob changed from `9b9ac30b...` to the workspace's `c87c2836...`. There is no receipt
   and no stray temp file.
4. `receipt` exits rc=2 with "not succeeded", both then and after permissions are restored.
5. Retrying `verify` or `transfer` exits rc=2 with "source deliverable changed".

**Suggested change:** Create and fill the pending receipt before the rename, and publish it after
`cmp`. Teach `receipt`, and the retry refusals, to recognize source == workspace != captured
`source_blob`. Report that as a distinct `transfer=unrecorded` state with a recovery command. Add a
regression for a failure after the rename.

**Scope:** The failing transfer's own stderr is truthful; the evidence available afterwards is not.

**Recommendation:** Prepare the receipt before the rename and make receipt report an unrecorded delivery distinctly.

#### L1. Several printed restoration commands fail when run literally · **Status:** open

**File:** scripts/isolated-review-workspace.sh:147 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** XS · **Urgency:** eventually

**Classification:** refusal advice. None of these remedies re-enables pushing.

**Reproduction:** I ran each printed remedy literally, then re-ran `verify`:

- **Multi-valued `remote.pushDefault`** (a path plus `origin`): `config --unset 'remote.pushdefault'`
  exits 5 ("has multiple values"), and `verify` still refuses.
- **Unguarded remote named `a b`:** the key at 142 is unquoted, so the command splits into
  `remote.a b.pushurl`. It writes the junk value `remote.a=b.pushurl`, and `verify` still refuses.
- **Multi-line `pushDefault` value:** the line-based parse at 144-148 prints "`<path>` names ''",
  and `config --unset '<path>'` fails with "invalid key".

The remedies for a removed origin, an extra or empty `pushurl`, and a path-valued `pushDefault` or
`branch.main.remote` all worked.

**Suggested change:** Use `--unset-all`, quote keys (`printf %q`), and parse `--get-regexp -z`
output (key, newline, value, NUL-terminated).

**Recommendation:** Use --unset-all, quote keys, and parse --get-regexp -z output.

#### L2. Mutation survivors and fixture shapes leave new invariants unpinned · **Status:** open

**File:** internal/tools/reviewworkspace/review_workspace_test.go:36 | **Component:** internal/tools/reviewworkspace
**Effort:** S · **Urgency:** eventually

**Classification:** coverage gap. It masks the defect classes behind H1, M1 and M2.

**Evidence:** four compiler-valid mutations survive. Each passed `bash -n` and ShellCheck and was
restored afterwards:

- dropping the `branch.<name>.remote` alternative from the regexp at 148;
- removing the post-overlay assertion at 188. It survives alone; restoring the pre-fix exclude as
  well gets it caught only at 203, after the baseline commit has run against the shared repository;
- pointing the receipt's workspace-blob check (367) at the source instead;
- deleting the receipt sentence from the handoff prompt (237).

`TestGeneratedBriefFollowsTheReceiptContract` uses `--dry-run`, which exits at prepare-script lines
226-229, before the handoff prompt (237) and the `[delivery]` note (245). Neither is pinned.

`newFixture` (36-93) never creates stat-dirty files, ignored content, nested repositories or
worktrees, or include- or worktree-scoped config. That is why the "index unchanged" assertions in
`TestLinkedWorktreeSourceStaysIndependent` (291-306) and `requireSameSource` pass despite M1, and
why H1 and M2 stay invisible.

**Suggested change:**

- Add those fixture shapes.
- Add a receipt case with a workspace edit after transfer.
- Add a non-dry-run generator capture, using the fake CLI, that asserts the handoff and
  `[delivery]` text.
- Add a test that the post-overlay assertion fires before any Git write, for example by simulating a
  replaced `.git`.

**Recommendation:** Add stat-dirty, nested, include-scoped fixtures plus receipt workspace-edit and non-dry-run handoff assertions.

#### L3. Task note says the pre-fix linked-worktree overlay fails closed on macOS but it commits into the source · **Status:** open

**File:** planning/tasks/6gk06e2g21hv-disable-pushes-from-isolated-review-workspaces.md | **Component:** planning
**Effort:** XS · **Urgency:** soon

**Classification:** planning-record accuracy, with a live consequence on `main`.

**Evidence:** Task `6gk06e2g21hv` ("Linked-worktree sources") says the pre-fix overlay fails closed
on macOS openrsync, and that whether it could commit into the shared repository is unverified.

I ran the real pre-fix helper (`git show 22a27650:scripts/isolated-review-workspace.sh`, exclude
`'/.git/'` at line 183) on this host's openrsync. The source was a linked worktree of a `git init`
fixture (`$P/prefix-go.glmPl9`). The overlay replaced `.git`. The helper then committed
`chore: capture isolated review baseline` onto the source's `review-branch` (`da4f45e` →
`4face93`) and exited 2 at the post-baseline assertion. A cloned-repository fixture did fail closed
(rc=23), so the outcome depends on repository shape.

`main`'s helper (`4fdf6fcc`, line 139) has the same exclude. Until this change lands,
linked-worktree reviews launched with `main`'s helper can commit into the source. This slice's fix
holds (evidence floor 3).

**Suggested change:** Correct the task note. Before relying on earlier linked-worktree reviews, check
source branches and reflogs for stray `capture isolated review baseline` commits; the 49 branches
captured in this sandbox have none.

**Recommendation:** Correct the record and check source branches and reflogs for stray baseline commits from main's helper.
