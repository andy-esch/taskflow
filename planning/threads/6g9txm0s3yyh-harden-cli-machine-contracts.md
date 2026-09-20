---
schema: 1
id: 6g9txm0s3yyh
status: completed
description: Turn compatibility audits into durable projected-output and command-contract guarantees.
goal: Machine-facing CLI views stay backward compatible and reject registry drift as they evolve.
created: "2026-09-13"
tags: [cli, json, contract, dogfood]
tasks: [6g9s401jtqb7, 6g9txf8x9m70, 6ga3z57hqa70, 6ga968z46e9y, 6gbmzgsnzhhp]
updated_at: "2026-09-19"
started_at: "2026-09-13"
ended_at: "2026-09-19"
---

# Thread: Harden CLI machine contracts

**Goal.** Machine-facing CLI views stay backward compatible and reject registry drift as they evolve.

## Context

The contract-and-compatibility audit exposed how one display-oriented column registry also serves
table, CSV, projected JSON, help, and completion. The first repair restores canonical selectors and
raw values without breaking existing projected keys; its independent reviews then demonstrated that
future registry declarations could recreate the same mismatch or silently shadow an alias.

Sequence the concrete compatibility repair before the registry-wide invariant pass. Preserve the
curated list surfaces: a column registry need not mirror every full DTO field, but any canonical
field it does expose must have one unambiguous selector identity and truthful value semantics.

The v0.22.0 release task is the final checkpoint for this bounded phase. It qualifies the four
completed contract changes together without pulling the separate exit-code taxonomy, executable
command discovery, Thread-list projection, or bounded-query work into the candidate. Once the
published artifacts are verified, complete this Thread and sequence those findings as a new phase.
