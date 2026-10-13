# Repository scripts

## General isolated review workspaces

`isolated-review-workspace.sh` is a standalone shell tool for any review that needs an independent
copy of a live Git working tree and exactly one deliverable copied back. It has no dependency on
taskflow's Go packages, CLI, planning schema, or audit layout:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
DELIVERABLE="path/to/review-report.md"
SANDBOX="$(scripts/isolated-review-workspace.sh create \
  --source "$SOURCE_ROOT" --deliverable "$DELIVERABLE" --print-path)"

"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
"$SANDBOX/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"
```

`create` makes an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted state, rejects a source that changes during capture, and makes one local
baseline commit. The overlay never copies Git metadata: every `.git` file or directory is excluded
at any depth, so a linked-worktree source, or an ignored nested worktree such as
`.claude/worktrees/*`, cannot bind part of the workspace to the shared repository. `create`, `verify`,
and `transfer` refuse nested `.git` entries, and `create` re-checks independence before its first
Git write. A non-ignored nested repository cannot be fingerprinted, so `create` refuses it. Reads of
the source disable index auto-refresh and optional locks, so capturing a stat-dirty checkout does
not rewrite its index. `verify` fails if the reviewer committed, staged anything, changed another file,
or if the source deliverable changed. `transfer` repeats those checks, requires a real deliverable
delta, and atomically replaces only that source file. State and attestation evidence live under the
sandbox's private `.git`; the sandbox remains available until its owner confirms receipt.

A report must be final before `transfer` runs, so it cannot record its own delivery. `verify`
therefore attests verification only (`verify=passed`, `transfer=pending`). After the transferred
bytes are confirmed in the source, `transfer` writes a receipt under the sandbox's private state and
prints it: the verification fields plus `transfer=succeeded`, the transferred blob, the source path,
and a UTC timestamp. `receipt` fails until a transfer has succeeded, and afterwards whenever the
source or sandbox no longer holds the transferred bytes; on success it reprints the receipt with
`delivery=confirmed`. Once a receipt exists, `verify` and `transfer` refuse and point to `receipt`.
A refused or failed transfer records no receipt.

A clone's `origin` would otherwise push to the live source checkout, so `create` sets every sandbox
push URL to `isolated-review-push-disabled://source-is-read-only`. Git has no transport for that
scheme, so `git push`, `git push origin …`, and branch, `--all`, or tag pushes fail before contacting
the source; fetching from `origin` still reads it. `verify` and `transfer` refuse a removed `origin`,
any remote whose effective push URL (after `insteadOf` expansion) is anything else, and a workspace
`remote.pushDefault` or `branch.<name>.pushRemote`/`remote` setting that names a path or URL instead
of a guarded remote. Each refusal prints the `git -C <sandbox> …` command that restores the guard.

The guard prevents accidents; it is not a security boundary. A push that names the source path or
URL explicitly, a `git -c`/environment override, or user-wide or system Git configuration bypasses
a workspace-local setting, and the helper never edits source or user-wide configuration to compensate.
The source checkout remains read-only by protocol.

The source checkout still needs a handoff freeze. A filesystem copy cannot make an arbitrary live
working tree transactionally atomic, and the source-hash guard cannot coordinate with an editor that
ignores it. The helper detects ordinary drift and fails closed; it is not a repository lock.

## Adversarial implementation reviews

`prepare-adversarial-review-audits.sh` turns one tailored review brief into independent audit
entities for Claude and Antigravity (or an explicit reviewer list). It uses `tskflwctl audit new`,
preflights the full pair before writing, injects the exact finding grammar, lints each audit, and
prints the short handoff prompt for each external agent. It also requires a distinct second pass
for systemic failure modes: shared abstractions or test helpers that can mask a class of defects,
state changing between projection and action, and boundaries that only appear to fail closed. This
is deliberately evidence-gated so “play devil's advocate” does not become permission to manufacture
speculative architecture findings.

Every generated review uses the general isolated-workspace helper above. The injected protocol
treats the handoff checkout as read-only, and the reviewer may transfer only their assigned audit
back after the helper's verification passes. The sandbox is retained until the implementation
owner confirms receipt.
This is required even for nominally read-only reviews: two reviewers may run simultaneously, and
review tooling or restoration commands must never share an index, branch, generated files, or
working tree with each other or the implementation owner. A Git worktree is intentionally not used
because it still shares repository administration and refs.

The sandbox baseline is the only reviewer-created commit permitted. A tailored brief's validation
rules must not forbid that local checkpoint, though they should continue to forbid every commit in
the source checkout and any push. If the reviewer cannot create or verify the independent sandbox,
the correct outcome is a reported blocker—never a fallback to the shared checkout.

Freeze the handoff before launching reviewers. Both audit briefs and the implementation snapshot
must be final for that review round, and the implementation owner must not edit an assigned audit
while its reviewer is active. If either target changes, cancel and relaunch or explicitly accept a
report scoped to the already captured baseline. This avoids forcing the final source-hash guard into
a conflict and makes each verdict's reviewed state unambiguous.

Isolation evidence has one contract, split by when each part can exist. The report carries the final
`verify` output: sandbox path, resolved Git directory, baseline commit, captured source-audit blob and
fingerprint, deliverable, push guard, `verify=passed`, and `transfer=pending`. Its `transfer=pending`
is expected and is never delivery evidence. The reviewer's handoff message carries the receipt that
`transfer` prints, and the implementation owner runs `receipt` against the retained sandbox before
triage or cleanup. This is a deliverable rather than an honor-system preamble: a report without the
verification attestation, or a handoff without a confirmable receipt, is protocol-incomplete. The
receipt lives in reviewer-writable sandbox state, so it detects accidents and stale delivery claims,
not a deliberately forged handoff.

The script deliberately does not generate the substantive brief. Review quality has come from
deriving the implementation contract and hostile cases for the change at hand, not from a generic
checklist. Start from a previous strong brief and preserve these required sections:

- review target and repository-wide consumer inventory;
- intended contract, including explicit non-goals;
- a mandatory evidence floor with real reproductions, repeated race tests, mutation probes,
  resource evidence, and exact validation commands;
- change-specific hostile angles and platform/support-boundary discipline;
- sandbox-local contamination/restoration rules and a precise deliverable; and
- a `## Reviewer report` placeholder for the assigned agent to replace.

The most productive repeatable probes should also be tailored into the brief:

- mutate the critical invariants selected in the brief and require the intended regression
  test—not merely some unrelated test—to fail; count compile errors as invalid probes, not kills;
- populate new optional wire/schema branches with non-default values in semantic validators;
- execute every suggested repair command against each diagnostic class that recommends it; and
- use coordinated mutations across an interface and its immediate caller so compilation by accident
  is not mistaken for a pinned architectural contract.

Set the evidence floor by the review's risk and primary lenses rather than requiring exhaustive
mutation of every new test. Give paired reviewers complementary primary responsibilities with a
bounded cross-check; keep one shared intended contract and explicit ownership of seam failures.

Every named symbol, field, lock path, test, fixture, and shipped capability must also be verified in
the sandbox with a path/line or command result. Keep implemented evidence separate from requirements
found only in planning documents. A clean verdict built on invented names or future fixtures is not
corroborating evidence and should be rejected during owner triage.

These checks came from review-loop evidence: two independent reviewers converged on a surviving
diagnostic de-duplication mutation, while running the advertised remedy and exercising non-default
schema paths exposed separate gaps that ordinary green tests did not.

Example:

```sh
scripts/prepare-adversarial-review-audits.sh \
  --area watcher-reconciliation-implementation \
  --title "Watcher reconciliation implementation" \
  --brief-file /tmp/watcher-review-brief.md
```

Use `--dry-run` to validate the brief, names, and audit collisions without creating files. This is
an interim standalone-audit workflow; it does not decide the task-attached review storage and
verdict questions tracked by epic 27.

The generated handoff prompt includes the absolute audit path. Each reviewer translates that to the
repository-relative deliverable used by the helper. If its source hash changes while review is
underway, the reviewer must leave the sandbox intact and report the conflict rather than overwrite
or merge in the shared checkout.
