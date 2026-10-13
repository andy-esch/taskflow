---
schema: 1
id: 6gka14zenf4x
bucket: closed
area: review-workspace-hardening-fixes-claude-fix-verification
date: "2026-10-13"
updated_at: "2026-10-13"
---
# Audit: Review-workspace hardening fixes — claude-fix-verification — 2026-10-13

> Reviewer assignment: claude-fix-verification. This document is the review brief and the only file the reviewer should update.
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
final guarded audit transfer. Your handoff prompt prints exact `SOURCE_ROOT` and `AUDIT_REL`
values. Substitute them below; never derive them from your current directory, which may be a
different checkout. Then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="<SOURCE_ROOT from your handoff prompt>"
AUDIT_REL="<AUDIT_REL from your handoff prompt>"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)" && cd "${SANDBOX:?create failed: stop and report the blocker}"
```

Every later block begins from `${SANDBOX:?…}`, so it cannot run in the shared checkout after a
failed `create`. Run the helper from `$SOURCE_ROOT`, not from the sandbox: you may mutate the
sandbox's copy while probing, and only the read-only source copy can be trusted to verify your work.

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
cd "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX" &&
  git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent or nested `.git`, a
removed or weakened push guard, source-deliverable drift, and empty reports; it copies back only
the assigned audit through a same-directory atomic rename. A successful transfer records a
helper-owned receipt and prints it (`transfer=succeeded` with the transferred blob). A refusal exits
2 and records nothing. Exit 3 means the report was already delivered when a later step failed: run
`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"` and put its
output, or its refusal, in your handoff. Put the printed receipt in your handoff message. Do not
copy anything else manually. Leave the workspace in place and report its path until the
implementation owner confirms delivery with that same `receipt` command. On refusal, preserve it and
report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

## Review brief

This is a focused second round. Settle two questions: did each fix for the thirteen first-round
findings actually close its reachable consequence, and did the fixes introduce a new way for the
helper or its generated protocol to write into the source or report false delivery? One reviewer
covers both earlier lenses (git-safety and delivery-evidence) on the fix delta only. Do not
re-review unchanged first-round code except where a fix depends on it.

The helper under review creates your own sandbox from a linked-worktree source, and its new
protocol prints your exact `SOURCE_ROOT` and `AUDIT_REL`. Check your sandbox's independence
yourself after creation instead of trusting the helper's output.

## Review target

Source checkout: the linked Git worktree named in your handoff prompt, on branch
`worktree-harden-isolated-review-workspaces`, with HEAD
`d279f6d5ee6f9bd061f823fedd06cca9fe50bdf6`. HEAD is a merge of `origin/main`
`86e255be59f7389dde526d139bfc9ef42f0f5582`, which brought the unrelated architecture-guide split.
The only uncommitted state is this audit.

Fix delta under review:

| Commit | Change |
| --- | --- |
| `98cdd13183ec0ad5bba3e5df8960de89f68d935b` | Nested `.git` excluded at any depth and refused; fingerprint fails closed; source reads skip index auto-refresh. |
| `1ece6c8cc891da937a9c7b443910791a768edd93` | Push destinations read from every workspace config scope; legacy remote files refused; quoted `restore:` lines. |
| `fc537a52731826c19907e833740d43878c4fb8ad` | Pending receipt written before the source rename; exit 3 after it; `receipt` recovery and diagnosis. |
| `229426ab5ad44164a13bffa9abd1406fa4e23d3e` | Handoff prints `SOURCE_ROOT`/`AUDIT_REL`; guarded blocks; helper run from the source copy. |
| `ea68919c3498fbd2854baba82ab26722ae959930` | Finding dispositions and the corrected task note. |

First-round evidence and dispositions:
[git-safety](6gk49rv9gech-2026-10-12-review-workspace-hardening-implementation-claude-git-safety.md)
and [delivery-evidence](6gk49rvj7rhe-2026-10-12-review-workspace-hardening-implementation-claude-delivery-evidence.md).
Each of their findings is marked fixed with a commit and a regression test.

Entry points: `scripts/isolated-review-workspace.sh` (`read_git`, `source_fingerprint`,
`refuse_nested_git`, `restore_line`, `workspace_config`, `origin_file`, `check_push_guard`,
`write_pending`, `publish_receipt`, `transfer_deliverable`, `show_receipt`, `load_workspace`),
`scripts/prepare-adversarial-review-audits.sh`, `scripts/README.md`, and
`internal/tools/reviewworkspace/review_workspace_test.go`. Verify every name before relying on it.

Build a short verified consumer inventory of the changed contracts: exit statuses 2 and 3,
`restore:` lines, the pending-receipt file, `SOURCE_ROOT`/`AUDIT_REL` in the handoff, and
source-copy invocation. Cover scripts, README, and tests.

## Intended contract to challenge

1. No `create`, `verify`, `transfer`, or `receipt` path writes the source repository's refs,
   objects, worktree administration, index, or configuration, or any file other than the
   deliverable. This holds for plain and linked-worktree sources, with nested worktrees,
   stat-dirty files, and a replaced `.git`.
2. `verify` passes only if no workspace-scoped setting or file can direct an ordinary push past
   the guard. Every printed `restore:` line runs as written and leads to a passing `verify`.
3. Exit 2 means the source report is unchanged and nothing was recorded. Exit 3 means it was
   replaced. `receipt` confirms only bytes that agree in source, workspace, and the record, and
   never reports a partial or diverged transfer as delivered.
4. A literal follower of the generated brief cannot end up working in the shared checkout, and
   always verifies with the source copy of the helper.

## Mandatory evidence floor

Record each item as reproduced, falsified, or unresolved.

1. **Re-run each first-round reproduction against the fix:** nested worktree commit, non-ignored
   nested repository, stat-dirty index, the include, `includeIf`, worktree-config and
   legacy-remote bypasses, the unwritable state directory, editor save after the rename, and a
   failed `create` in the literal brief. Report the new outcome.
2. **New hostile cases for the fixes:**
   - a nested `.git` deeper than one level or inside an ignored directory created after `create`;
   - a symlink named `.git`;
   - a relative `include.path`;
   - an `insteadOf` rule in an included file;
   - a pending receipt left by a killed transfer, followed by `verify`, `transfer`, and `receipt`;
   - concurrent `receipt` and `transfer`;
   - a `restore:` line for a path with a newline or `$`.
3. **Two compiler-valid mutations of your choice** on fix-delta invariants the owner's ledger did
   not cover. Run the exact intended test, require a failure for the intended reason, and restore.
4. **Exit-status audit:** list every `die`/`die_delivered` call after the source rename and confirm
   no exit-2 path follows a source replacement.

## Required hostile angles

- Can `refuse_nested_git`'s `find` miss metadata (symlinks, permission errors, `-prune` scope), or
  run slowly enough on large ignored trees to make reviewers bypass it?
- Can `--worktree --includes` or `--show-origin` parsing mis-pair records, so a value is checked
  under the wrong key or a restore line targets the wrong file?
- Does `show_receipt` deleting a stale pending record, or publishing one, create a check-then-act
  window, or let a pending record from another deliverable vouch for this one?
- Do the new tests' shims (rsync, mv, fake CLI) mask the behavior they claim to pin, for example by
  passing because the shim itself failed?
- Does the merge with `origin/main` interact with these files or with the architecture check?

Second pass: look for a fix that relocated a defect class instead of removing it, and for evidence
copied across the README, the brief, and the helper that could drift.

## Validation and restoration

Run inside your sandbox:

```sh
just build
go test ./internal/tools/reviewworkspace/ -count=1 -v
shellcheck scripts/isolated-review-workspace.sh scripts/prepare-adversarial-review-audits.sh
./bin/tskflwctl -C . lint --links
git diff --check
```

The owner passed full race-enabled `just test` (29 packages, including `architecturecheck`),
`just lint`, and `just docs-check` on HEAD. Inherit those if you label them as inherited.

Put disposable repositories, backups, and caches in separate `mktemp -d` directories outside the
sandbox. Use `GIT_CONFIG_GLOBAL=<fixture file>` and `GIT_CONFIG_NOSYSTEM=1` for every
disposable-repository command. Restore every probe so only your assigned audit differs.

## Deliverable

Fill only the Reviewer report slot. Create real findings with the sandbox's
`./bin/tskflwctl audit finding new` and leave them open.

Lead with the verdict (ready / needs changes / incomplete), the snapshot, and your limits. Then
include:

- the isolation attestation and your own independence probe;
- a per-finding re-test table covering all thirteen first-round findings;
- new findings, consolidated by root cause;
- a probe and mutation ledger;
- a bounded surviving-hypothesis ledger.

Separate regressions introduced by the fixes from residual gaps and taste. High severity means a
reachable wrong-target write or materially false delivery evidence.

## Reviewer report

**Verdict: needs changes.** Every reproduction from the first-round findings now behaves as
claimed, and the exit-status contract holds inside `transfer`. The fixes still leave two
Medium problems:

- a regression: concurrent `receipt` or `verify` can corrupt the new pending-receipt state (M2);
- a residual of H1's class: copied absolute symlinks still bind the workspace to the source (M1).

Three Low findings cover push-guard residuals, mutation survivors, and fail-closed usability
regressions.

**Snapshot:** source worktree `worktree-harden-isolated-review-workspaces` at `d279f6d5` plus this
uncommitted audit. The sandbox baseline is `ec443cd6`. The helper blob `e9d881c0` was identical in
the source, the baseline, and every probe fixture.

**Limits:** run on macOS (case-insensitive APFS) with Git 2.54.0, bash 3.2.57, openrsync,
ShellCheck 0.11.0 and Go 1.27.2. I did not test GNU rsync or Linux. I did not inspect the main
checkout. A clone cannot show the source's reflogs.

### Isolation attestation

Final `verify` output, run with the source checkout's helper copy:

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.UwP634
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.UwP634/.git
baseline_commit=ec443cd69f820a19f5dab76513288f0c74985d86
source_blob=50fc37f4e19826d2ad5b5dcc000a2ce0235b8af8
source_fingerprint=f65dbad7d46ae1a1eb094834c02ad7f7a45eaa8e
deliverable=planning/audits/6gka14zenf4x-2026-10-13-review-workspace-hardening-fixes-claude-fix-verification.md
push_guard=isolated-review-push-disabled://source-is-read-only
deliverable_changed=true
verify=passed
transfer=pending
```

`transfer=pending` is the truthful value here. Delivery evidence is the helper's receipt, which
goes in the handoff message.

**Own independence probe** (run inside the sandbox):

- `.git` is a directory, and `--git-dir`, `--git-common-dir` and `--absolute-git-dir` all name the
  sandbox's own `.git`.
- There is no `objects/info/alternates`, no `commondir`, no `config.worktree`, and no
  `.git/remotes` or `.git/branches`.
- No object file has more than one link, and `git worktree list` shows only the sandbox.
- The push URL is the guard, and `git push --dry-run` fails at
  `remote-isolated-review-push-disabled`.
- The baseline differs from `d279f6d5` only by this audit.

Every probe ran in disposable repositories under `/tmp/fixverify-probe.Mokj8I`, with
`GIT_CONFIG_GLOBAL` set to a fixture file and `GIT_CONFIG_NOSYSTEM=1`. Mutations ran in the
sandbox, from backups in `/tmp/fixverify-backup.gMr2C8`, and every one was restored. Only this
audit differs from the baseline.

### Re-test of the thirteen first-round findings

| First-round finding | Outcome against the fix | Residual or new |
| --- | --- | --- |
| git-safety H1 (nested Git metadata) | **Fixed for the stated shapes.** Ignored nested worktree: its `.git` is not copied. `git -C <copy> rev-parse --show-toplevel` is the sandbox, a commit there lands in the sandbox, `verify` then refuses "HEAD changed", and the shared administration digest is unchanged. Non-ignored nested worktree: `create` exits 2 with "cannot fingerprint" and leaves no sandbox. | Copied absolute symlinks and a `.GIT` case variant still reach the source (M1). |
| git-safety M1 (stat-dirty index) | **Fixed.** The index SHA was unchanged across `create`, `verify`, `transfer` and `receipt`, for a plain checkout and for a linked worktree's `worktrees/linked/index`, with files still stat-dirty afterwards. | None |
| git-safety M2 / delivery M3 (push settings in other scopes) | **Fixed for the stated shapes.** Refused: `include.path`, `includeIf.onbranch:main`, `includeIf.gitdir:`, a `config --worktree` push default, and legacy `remotes` and `branches` files. Each printed restore line ran literally (bash or zsh) and then `verify` passed. Mutating `branch.*.remote` is killed. | A dormant `includeIf` and a symlinked legacy remote both pass `verify` and then push into the source (L1). |
| git-safety M3 / delivery M1 (failure after the rename) | **Fixed for the stated shapes.** Unwritable state directory: exit 2, source unchanged, no stray file, and the retry succeeds. Editor save after the rename: exit 3, `receipt` reports "delivery uncertain", and `verify` points to `receipt`. SIGKILL after the rename: `receipt` records `recorded_by=receipt …`, then `delivery=confirmed`. | Concurrent `receipt` or `verify` corrupts the receipt (M2, a regression). |
| git-safety L1 / delivery L1 (restore lines) | **Fixed.** Every first-round shape restores literally under zsh (`zsh -f -c`): a multi-valued `pushDefault`, remote `a b`, remote `a.b*`, `branch.it's.pushRemote`, a multi-line value, and a local `insteadOf`. So does remote `a$HOME`. | An included or worktree-scoped `pushurl`, and a legacy file name containing a newline, print lines that do not restore (L1). |
| git-safety L2 / delivery L2 (mutation survivors) | **Fixed.** Every named survivor is now killed: `branch.*.remote`, the receipt workspace blob, the handoff sentence, and the first-round `[delivery]` rewrite (each for the intended reason). The post-overlay assertion is killed by the shared-administration check at `review_workspace_test.go:514`. | Two new fix-delta survivors (L2). |
| git-safety L3 (task note) | **Fixed.** A correction section is appended to task `6gk06e2g21hv` (lines 111–129). No baseline commit other than this sandbox's appears in the clone's 77 refs, which include every source branch. | Reflogs were not checkable. The note's earlier "excludes `/.git`" line is now stale (taste). |
| delivery M2 (failed `create` in the literal brief) | **Fixed.** I took the blocks from this generated brief and ran them in bash and zsh, using the old absolute `AUDIT_REL`. `create` exits 2, block 2 aborts at `${SANDBOX:?…}`, block 3 never runs, the source is clean, and no sandbox is created. My own handoff printed exact `SOURCE_ROOT` and `AUDIT_REL`. | None |
| delivery L3 (sandbox copy as verifier) | **Fixed in protocol, README and tests.** With the sandbox copy mutated to report `verify=passed`, literal block 2 still refuses through the source copy and lists both stray files. | The regression test does not pin it (L2, mutation 1). |

### New findings by root cause

- **Regressions introduced by the fixes:**
  - M2 (`fc537a52`): readers delete the live pending record, and the publish step neither
    validates the record nor refuses to overwrite an existing receipt.
  - L3 (`98cdd131`): fail-closed usability regressions with misleading diagnoses.
- **Residual gaps in fixed defect classes:**
  - M1 (H1's class): the workspace binding is still identified by the name `.git`, not by where
    paths resolve.
  - L1 (M2, M3 and L1's class): the guard enumerates push inputs only as they resolve now.
  - L2: string-based and outcome-only regression pins.
- **Taste, not filed:**
  - The generated brief says "Exit 3 means the report was already delivered" (generator line 208).
    The helper (`:36-38`) and README (`:48`) say "replaced", and the `:468` exit 3 means
    "delivery is unconfirmed".
  - Restore lines for an included setting edit the included file wherever it lives, possibly
    user-wide configuration. Unsetting the local include directive would keep the edit inside the
    workspace.
  - `${SANDBOX:?create failed…}` also fires when a non-persistent agent shell has lost the
    variable. That is fail-closed but misleading.
  - No test asserts exit status 2 numerically.
  - A `[delivery]` rewording that keeps the pinned substring survives.

### Changed-contract consumer inventory (verified)

- **Exit statuses 2 and 3:**
  - Helper: usage `:36-39`; `die`/`die_delivered` `:50-60`; exits at `:468` and `:470`.
  - README: `:47-51`.
  - Generated brief: generator `:206-210`.
  - Tests: `review_workspace_test.go:739-795` assert exit 3 for both paths. Refusal tests check
    only for a non-nil error.
- **`restore:` lines:**
  - Helper: `restore_line` `:172-178`, used at `:208`, `:221`, `:224`, `:231` and `:235`;
    `origin_file` `:192-196`.
  - README: `:61-62`.
  - Tests: `:319-350` execute the lines through `bash -c` only.
- **Pending-receipt file:**
  - Helper: `:14`, `:342-347`, `:384-399`, `:403-410`, `:450-466` and `:481-496`.
  - README: `:46-51`.
  - Tests: `:580` and `:681`.
- **`SOURCE_ROOT`/`AUDIT_REL` in the handoff:**
  - Generator: brief `:157-166`; derivation and `printf` `:246-252`.
  - README: `:161-166`.
  - Tests: `:854-905`.
- **Source-copy invocation:**
  - Generator: `:164`, `:189`, `:200-201`, `:209` and `:259`.
  - README: `:12-23`.
  - Tests: `helper()`/`workspace()` `:116-145`, the string check `:870`, and `:901`.

### Exit-status audit

In `transfer_deliverable`, the source rename is `scripts/isolated-review-workspace.sh:462`; a
failure there is `die`, with the source unchanged. After a successful rename:

- `:464-466` clear `transfer_file` and `pending_cleanup`, then the EXIT trap.
- `:468` and `:470` are the only error exits, and both are `die_delivered` (exit 3).
- `:471` `cat` can fail only with cat's own status.

No exit-2 path follows a source replacement. The pre-rename `die` calls at `:456-462` run the trap
installed at `:454`, which removes the staged copy and the pending record. `:449` and
`write_pending` (`:389-397`) run before that trap exists, and `write_pending` removes its own
temporary file.
After an interrupted delivery, `verify` and `transfer` exit 2 and name `receipt`. That refusal is
the command's own outcome; the replacement came from the earlier transfer.

### Probe and mutation ledger

| Probe | Result |
| --- | --- |
| Deep nested `.git` file, created after `create` inside ignored `ignored/a/b/c/` | Refused by `verify` (falsified gap). |
| Nested symlink named `.git` pointing at the source `.git` | Refused (falsified gap). |
| Unreadable directory in the sandbox | Fail closed: "cannot scan the workspace". |
| `.GIT` gitfile | Escapes, and a commit through it advances the source branch (M1). |
| Absolute symlinks to a nested worktree or the source root | Escape: commits advance `other-agent-work` and `main` while `verify` passes (M1). |
| Untracked absolute symlink to a source file | Committed into the baseline; a write through it changes the source file (M1). |
| Relative `include.path` (`rel.cfg`, `../../relup.cfg`) | Refused; restore works. |
| `insteadOf` in an included file | Refused; both restore lines work. |
| Included or worktree `pushurl` | Refused; restore line loops (L1). |
| Dormant `includeIf.onbranch` | Passes `verify`, then a push reaches the source (L1). |
| Symlinked legacy remote | Passes `verify`; `git push src` reaches the source (L1). |
| Restore line for a remote named with `$` | Restores. |
| Legacy file name containing a newline | `rm` restore line fails (L1). |
| SIGKILL during staging | `verify` clears the pending record and passes; a stray `*.review-transfer.*` stays in the source (pre-existing, noted in M2). |
| SIGKILL after the rename | `verify`/`transfer` exit 2 and point to `receipt`; `receipt` records and confirms. |
| Concurrent `receipt`/`verify` and `transfer` (shims at staging, after the rename, and during staging) | Corrupted receipt; transfer exits 0 (M2). Without shims: 1 of 40 trials. |
| Sandbox parent containing `[1]`, `*` or `?` | `[` makes `create` refuse itself (L3); `*` and `?` still detect planted metadata (no fail-open). |
| Untracked dangling or directory symlink | `create` refuses with a misleading hint (L3). |
| `find` cost on this tree | 1539 files, `real 0.00`. |
| Merge `ea68919c..d279f6d5` | Shares no file with the fix delta; `architecturecheck` passes. |

| Mutation | Result |
| --- | --- |
| Generator `:189`: verify with the sandbox copy | **Survived** the whole package (L2). |
| Helper `:458`: staged-copy hash check removed | **Survived** the whole package; exit 2 becomes exit 3 with the source replaced (L2). |
| Control, helper `:470`: `die_delivered` to `die` | Killed for the intended reason. |
| First-round L2 survivors (5) | All killed. |

Validation run in the sandbox, all passing:

- `just build`
- `go test ./internal/tools/reviewworkspace/ -count=1 -v` (15 of 15)
- ShellCheck on both scripts
- `./bin/tskflwctl -C . lint --links`
- `git diff --check`
- the full race-enabled `just test` (29 packages ok, including `architecturecheck`)

`just lint` and `just docs-check` are inherited from the owner's HEAD run; I did not re-run them.

### Bounded surviving-hypothesis ledger

1. A SIGTERM between `:462` and `:465` would run the EXIT trap and delete the pending record after
   delivery. The window is two assignments; not probed.
2. A stale pending record from a SIGKILL before the rename, followed by a hand copy of identical
   bytes, would be recorded by `receipt` as an interrupted transfer. The bytes agree, so this is
   wording, not false bytes; reasoned only.
3. Cross-deliverable vouching: falsified as reachable. State is per sandbox. The pending record's
   `deliverable` field is never cross-checked, which would matter only if state were copied
   between sandboxes.
4. GNU rsync and Linux handling of `--exclude='.git'` and of symlinks: untested locally. CI runs
   on `ubuntu-latest` (`.github/workflows/ci.yml:95,112`).
5. The main checkout as a source may contain absolute symlinks (M1 reachability): unknown, not
   inspected.
6. Submodules: their `.git` is now excluded by name, and a dirty submodule probably fails the
   fingerprint closed. Not built.
7. The handoff prints `SOURCE_ROOT`/`AUDIT_REL` with `%s` into a double-quoted template, so a path
   containing `"`, `$` or a backtick would misparse when pasted. No such path exists here; not
   probed.
8. `includeIf.hasconfig:` conditions: not probed.

## Findings

#### M1. Absolute symlinks and case-variant .GIT entries still bind the workspace to the source while verify passes · **Status:** fixed in 1e8332bf

**File:** scripts/isolated-review-workspace.sh:131 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** M · **Urgency:** soon

**Classification:** residual gap in the H1 fix (`98cdd131`), same consequence class. Not a regression: the pre-fix overlay also copied symlinks. The fix removes Git bindings by the exact name `.git`, but `rsync -a` copies symlinks verbatim (`scripts/isolated-review-workspace.sh:274`), and `refuse_nested_git` (`:128-135`) neither follows symlinks nor matches case variants. `scripts/README.md:27-29` claims the overlay "cannot bind part of the workspace to the shared repository".

**Consequence:** Git commands and ordinary file writes reached through a copied absolute symlink act on the source repository or its working files. `verify` passes before and after.

**Reproduction** (disposable fixtures under `/tmp/fixverify-probe.Mokj8I`, Git 2.54.0, macOS APFS):

- `symlnk`: the source ignores `ignored/` and `.claude/`. It holds a nested worktree `.claude/worktrees/other` on branch `other-agent-work`, plus `ignored/other-wt -> <source>/.claude/worktrees/other` and `ignored/source-root -> <source>`. `create` exits 0, and the sandbox holds both links with unchanged absolute targets. `verify` prints `verify=passed`. `git -C <sandbox>/ignored/other-wt rev-parse --absolute-git-dir` returns `<source>/.git/worktrees/other`. `git -C <sandbox>/ignored/other-wt commit --allow-empty` advanced the source's `other-agent-work` (`5dd3112`). The same through `ignored/source-root` advanced the source's `main` (`36511e6`). The shared administration digest changed, and `verify` still printed `verify=passed`.
- `fp-abs-file-into-source`: a non-ignored untracked `src/kept-link -> <source>/src/kept.txt` is committed into the baseline as mode `120000` with its absolute target. `printf ... >> <sandbox>/src/kept-link` changed the source's `src/kept.txt`. Sandbox `git status` lists only the deliverable, and `verify` passes.
- Case variant: a nested `ignored/up/.GIT` gitfile written after `create` resolves to `<source>/.git/worktrees/other` on case-insensitive APFS. `find -name .git` misses it, so `verify` passes, and a commit there advanced the source branch.

**Reachability:** needs an absolute symlink into a Git repository (or a case-variant gitfile) in the source or the workspace, plus a Git command or write run through it. This worktree's own tree has none (`find -type l` in the sandbox: 0 of 1539 files), so this review was not exposed. I did not inspect the main checkout. The README presents the helper as a general tool for "any review". Severity is Medium because no such link was shown in this repository. If any intended source contains such links, treat it as H1's High.

**Recommendation:** Refuse workspace symlinks that resolve outside the workspace (at least absolute and source-resolving ones), match .git case-insensitively, and scope the README isolation claim.

**Resolution:** refuse_external_bindings refuses any nested .git in any letter
    case and any symlink whose target is absolute or climbs out of the
    workspace, in create, verify, and transfer.
    TestSymlinksLeadingOutOfTheWorkspaceAreRefused covers create-time and
    planted links and a .GIT gitfile; disabling the escape check or the
    case-insensitive match fails it.

#### M2. A concurrent receipt or verify deletes the live pending record, so transfer exits 0 with a field-less receipt that can never be confirmed · **Status:** fixed in 5fafeab1

**File:** scripts/isolated-review-workspace.sh:409 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** S · **Urgency:** soon

**Classification:** regression introduced by `fc537a52`. It relocates first-round M3 (git-safety) / M1 (delivery-evidence) into a concurrency window. Before the fix, `receipt` and `verify` never mutated receipt state.

**Mechanism:** `load_workspace` (`scripts/isolated-review-workspace.sh:342-347`) and `show_receipt` (`:485-489`) delete `transfer-receipt.pending` whenever the source still equals the captured blob. Nothing coordinates that delete with an in-flight `transfer`. `publish_receipt` (`:403-410`) appends with `>>`, which creates the file if it is missing, then `mv`s it over any existing receipt. `transfer` prints whatever results (`:471`) and exits 0.

**Deterministic reproduction:** PATH shims run the concurrent command at an exact point, then `exec` the real `cp`/`mv`. Each fired once, logged below.

- A. `receipt` during the staging window (`cp -p`, before the rename) prints "no transfer receipt" and deletes the pending record. `transfer` then exits 0 and prints only `transferred_at=2026-10-13T11:39:44Z`, and the source holds the workspace bytes. Afterwards `receipt` exits 2 with "transfer receipt does not record a successful transfer", and `verify` and `transfer` exit 2 with "deliverable already transferred".
- B. `receipt` right after the rename publishes a correct receipt and prints `delivery=confirmed`. `transfer`'s own publish then overwrites it with a timestamp-only file and exits 0. A later `receipt` exits 2 as in A, so the owner's earlier confirmation is contradicted.
- C. A concurrent `verify` during staging gives the same outcome as A.

**Natural reproduction (no shims):** three `receipt` pollers over 40 transfers produced 1 corrupted outcome: `transfer` rc=0, printed only `transferred_at=2026-10-13T11:41:01Z`, and the receipt file's only key is `transferred_at`. The pending-to-publish window measured about 20 ms; one `receipt` call takes about 43 ms.

**Consequence:** the delivered report becomes permanently unconfirmable. A receipt exists, so `verify` and `transfer` refuse, and `receipt` has no recovery for a malformed receipt. The reviewer's handoff carries an exit-0 "receipt" without `transfer=succeeded` or `transferred_blob`. This fails toward "not delivered", not false delivery, hence Medium, as in the first round.

**Adjacent residual (pre-existing):** a transfer killed by SIGKILL before the rename leaves `planning/audits/review.md.review-transfer.WrZRGs` untracked in the source. The pending record now shows that a transfer was interrupted. `verify` nevertheless discards it and passes, and the later `transfer` and `receipt` succeed with the stray file still in the source.

**Recommendation:** Serialize verify/transfer/receipt state changes with a lock, make publish refuse unless the pending record names the expected blob and never overwrite a receipt, and validate the printed receipt.

**Resolution:** verify and receipt never delete a pending record and report
    leftovers with restore lines; receipts are built complete from the pending
    record and hard-linked without overwriting, idempotent for one blob.
    TestConcurrentCommandsCannotCorruptATransfer and
    TestKilledTransferLeavesRepairableState pin it; restoring the deletion or
    the append-and-rename publish fails them.

#### L1. Push-guard scope fix misses dormant conditional includes and symlinked legacy remotes, and some restore lines do not restore · **Status:** fixed in f2b7b068

**File:** scripts/isolated-review-workspace.sh:233 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** S · **Urgency:** eventually

**Classification:** residual gaps in the M2/M3/L1 fixes (`1ece6c8c`). `check_push_guard` enumerates some of Git's push inputs as they are resolved for the current HEAD, by file type. These shapes escape that. Contract 2 and `scripts/README.md:59-62` ("read from every workspace scope … `include.path`/`includeIf` files"; "Legacy … files are refused"; restore lines "restore the guard when run as written") are falsified for them.

**Reproduction** (fixtures under `/tmp/fixverify-probe.Mokj8I`):

- Dormant conditional include. `includeIf.onbranch:notes/**.path = notes.cfg` sets `remote.pushDefault = <source>`. On `main`, `verify` exits 0 with `push_guard=…` and `verify=passed`. `git switch -c notes/scratch` leaves HEAD at the baseline, so verify's commit check still holds. A plain `git push -u` then creates `notes/scratch` in the source. `verify` refuses only after the switch.
- Symlinked legacy remote. `.git/remotes/src` is a symlink to a file containing `URL: <source>`. `verify` exits 0, and `git push src HEAD:refs/heads/legacy-symlink-leak` creates that ref in the source. `find … -type f` (`:233`) skips symlinks.
- Pushurl in another scope. `remote.origin.pushurl = <source>` from `include.path`, or from `config --worktree --add`, is refused. The only printed restore line is `git -C <sandbox> config --replace-all remote.origin.pushurl <guard>`, which edits the local file. It exits 0, and `verify` keeps refusing with the same line.
- Newline in a legacy file name. `find | head -n 1` truncates `.git/remotes/a\nb` (a newline) at the newline. The printed `rm -- …/remotes/a` exits 1, and `verify` keeps refusing.

**Controls that now hold:** active include.path, includeIf `onbranch:main` and `gitdir:`, worktree-scoped `pushDefault`, relative includes (`rel.cfg` and `../../relup.cfg`), insteadOf inside an included file, and legacy remotes/branches regular files are all refused. Every restore line for them ran literally in bash or zsh and was followed by a passing `verify`. So was every first-round L1 shape under zsh, including a remote named `a$HOME`.

**Severity:** Low. Each shape needs one deliberate indirection beyond the first-round Medium cases, and the README scopes the guard to accidents. The owner may align it with the first-round Medium.

**Recommendation:** Refuse include/includeIf keys and worktreeConfig in workspace config, refuse any non-directory entry under .git/remotes and .git/branches, and print --file restore lines for non-local pushurl origins.

**Resolution:** Conditional includes are refused, any non-directory legacy
    remote entry is refused NUL-delimited, and push URLs from other workspace
    scopes get --file restore lines. Four new weakening cases run their restore
    lines; each mutation fails its case.

#### L2. Two fix-delta invariants survive compiler-valid mutations · **Status:** fixed in 5fafeab1, 85df971e

**File:** internal/tools/reviewworkspace/review_workspace_test.go:870 | **Component:** internal/tools/reviewworkspace
**Effort:** XS · **Urgency:** eventually

**Classification:** coverage gap. Two fix-delta invariants survive compiler-valid mutations (each passed `bash -n` and ShellCheck 0.11.0, ran against the whole `internal/tools/reviewworkspace` package, and was restored from `/tmp/fixverify-backup.gMr2C8`).

- **Mutation 1 (`229426ab`, L3 invariant):** in `scripts/prepare-adversarial-review-audits.sh:189` (generated block 2), `"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify` becomes `./scripts/isolated-review-workspace.sh verify`. The block has already run `cd "${SANDBOX:?…}"`, so this verifies with the reviewer-mutable sandbox copy, which is exactly first-round L3. `TestGeneratedBriefFollowsTheReceiptContract` and the whole package pass. The test rejects only the literal string `"$SANDBOX/scripts/isolated-review-workspace.sh"` (`review_workspace_test.go:870`). Its call-order check (`:908-922`) accepts any path ending in the helper name, and the fixture's sandbox copy equals the source copy, so the walkthrough passes too. L3's resolution, "the generator test rejects any brief that runs the sandbox copy", does not hold.
- **Mutation 2 (`fc537a52`, contract 3):** in `scripts/isolated-review-workspace.sh:458`, the staged-copy check `[[ "$(git hash-object -- "$transfer_file")" == "$transferred" ]]` becomes `[[ -n "$transferred" ]]`. The whole package passes. A `cp` shim appends to the workspace report just before staging. The baseline helper then exits 2 with "workspace report changed during transfer; the source was not changed": the source is unchanged and no pending record remains. The mutant exits 3 with the source replaced, a pending record left, and `receipt` reporting "delivery uncertain". No test edits the workspace between hashing and staging, so the exit-2/exit-3 boundary at this branch is unpinned.

**Control:** changing `:470` from `die_delivered` to `die` is killed by `TestInterruptedTransferIsRecordedByReceipt` for the intended reason ("transfer should exit 3 …; err=exit status 2"). All five first-round L2 survivors are now killed as well.

**Recommendation:** Require every generated helper call to start with the SOURCE_ROOT path, and add a staging-time workspace-edit case requiring exit 2 and an unchanged source.

**Resolution:** TestWorkspaceEditDuringStagingLeavesTheSourceUnchanged pins the
    staged-copy re-hash (exit 2, source untouched), and the generator test now
    requires every helper call to use the source copy; both reported survivors
    are killed.

#### L3. New fail-closed checks refuse legitimate sources with misleading messages · **Status:** fixed in 1e8332bf

**File:** scripts/isolated-review-workspace.sh:110 | **Component:** scripts/isolated-review-workspace.sh
**Effort:** XS · **Urgency:** eventually

**Classification:** regressions introduced by `98cdd131`. Both fail closed, so this is usability with a misleading diagnosis, not a safety issue.

- **Untracked symlinks fail the fingerprint.** `source_fingerprint` (`scripts/isolated-review-workspace.sh:108-114`) runs `git hash-object` on each untracked path, and that follows symlinks. `create` refused, rc=2, for a non-ignored untracked symlink that dangles (`src/latest -> does-not-exist`: "could not open 'src/latest' for reading") or targets a directory (`src/plan-link -> ../planning`: "Unable to hash src/plan-link"). Both times it printed "cannot fingerprint the source working state (an unhashable path such as a nested repository?)". Before the fix these paths were skipped, and `git add` stores such links normally as mode `120000`.
- **Brackets in the sandbox path.** `refuse_nested_git` (`:131`) passes the sandbox path to `find -path` as a pattern. With a sandbox parent `…/par[1]`, the workspace's own root `.git` is not pruned, and `create` refuses itself with "workspace contains nested Git metadata: …/par[1]/isolated-review.eB4Uha/.git". Parents containing `*` and `?` still detected a planted nested `.git` (no fail-open).

**Recommendation:** Hash untracked symlinks by link text as git add does, and run the nested-metadata find relative to the workspace so path characters are not patterns.

**Resolution:** Untracked symlinks are fingerprinted by link text, and the
    nested-metadata scan runs from inside the workspace.
    TestInTreeSymlinksAreAccepted and
    TestPatternCharactersInTheSandboxParentPath pin both.
