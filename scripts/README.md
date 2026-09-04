# Repository scripts

## Adversarial implementation reviews

`prepare-adversarial-review-audits.sh` turns one tailored review brief into independent audit
entities for Claude and Antigravity (or an explicit reviewer list). It uses `tskflwctl audit new`,
preflights the full pair before writing, injects the exact finding grammar, lints each audit, and
prints the short handoff prompt for each external agent. It also requires a distinct second pass
for systemic failure modes: shared abstractions or test helpers that can mask a class of defects,
state changing between projection and action, and boundaries that only appear to fail closed. This
is deliberately evidence-gated so “play devil's advocate” does not become permission to manufacture
speculative architecture findings.

Every generated review is also isolated from the handoff checkout. The injected protocol treats
that checkout as read-only, creates an independent local clone with its own `.git`, overlays the
current staged, unstaged, untracked, and deletion state, and records a sandbox-only baseline commit.
All tests, generators, formatting, and mutation probes happen against that baseline. The reviewer
may transfer only their assigned audit back, and only when its source hash still matches the value
recorded before the copy. The sandbox is retained until the implementation owner confirms receipt.
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

The report must attest to isolation with the sandbox path, resolved Git directory, sandbox baseline
commit, captured source-audit blob, and transfer result. This is a deliverable rather than an
honor-system preamble: a technically useful report that omits it is still protocol-incomplete.

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

- mutate the exact invariant each newly added regression test claims to protect and require that
  test—not merely some unrelated test—to fail;
- populate new optional wire/schema branches with non-default values in semantic validators;
- execute every suggested repair command against each diagnostic class that recommends it; and
- use coordinated mutations across an interface and its immediate caller so compilation by accident
  is not mistaken for a pinned architectural contract.

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

The generated handoff prompt includes the absolute audit path. Each reviewer must translate that to
the repository-relative `AUDIT_REL` used by the injected commands. If the source-audit hash changes
while review is underway, the reviewer must leave the sandbox intact and report the conflict rather
than overwrite or merge in the shared checkout.
