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

The script deliberately does not generate the substantive brief. Review quality has come from
deriving the implementation contract and hostile cases for the change at hand, not from a generic
checklist. Start from a previous strong brief and preserve these required sections:

- review target and repository-wide consumer inventory;
- intended contract, including explicit non-goals;
- a mandatory evidence floor with real reproductions, repeated race tests, mutation probes,
  resource evidence, and exact validation commands;
- change-specific hostile angles and platform/support-boundary discipline;
- contamination/restoration rules and a precise deliverable; and
- a `## Reviewer report` placeholder for the assigned agent to replace.

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
