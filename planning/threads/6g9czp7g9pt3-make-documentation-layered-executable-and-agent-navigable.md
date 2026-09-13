---
schema: 1
id: 6g9czp7g9pt3
status: unstarted
description: Restructure project guidance around clear source ownership, scoped architecture and Go package contracts, and executable drift checks.
goal: Agents and contributors load the minimum authoritative guidance for a change, while code and CI expose documentation drift.
created: "2026-09-12"
tags: [documentation, architecture, agents, go, dx]
tasks: [6g63hjm7cp6w, 6g6x7e2ef37r, 6g9cz5saenme, 6g9cz5sk3b0d, 6g9cz5sv6kck, 6g9cz5t37dc0]
updated_at: "2026-09-13"
---
# Thread: Make documentation layered, executable, and agent-navigable

**Goal.** Agents and contributors load the minimum authoritative guidance for a change, while code and CI expose documentation drift.

## Context

`docs/ARCHITECTURE.md` is valuable because it keeps implementation work aligned, but at roughly 8,000 words it now mixes stable boundaries, current inventory, subsystem manuals, historical review outcomes, and dated status. The same mutable facts are repeated in `README.md`, `CLAUDE.md`, package comments, ADRs, generated CLI docs, schema output, and planning records.

A reference review of the Desirelines `shared`, `dispatcher`, and `apigateway` Go projects confirmed the value of progressive disclosure:

- service READMEs own system maps and operating workflows;
- selective `doc.go` files own package-local contracts, usage, concurrency, failures, and composition;
- Go headings and `[Symbol]` links make those contracts navigable beside the code;
- OpenAPI, protobuf, and generated artifacts state their canonical source explicitly.

The review also changed two parts of the initial direction. `doc.go` should be risk-based rather than compulsory for every substantial package, and copyable examples or mutable inventories should be executable or generated. Desirelines itself shows the failure mode: conflicting `just` versus raw Go workflow guidance, `http.Error` examples beside a rule that forbids that response path, and a two-port package overview beside six current interfaces.

This Thread therefore treats documentation as a set of owned projections:

- a short root agent router and durable architecture index;
- scoped guides for deeper cross-package design;
- package contracts near Go code where locality matters;
- ADRs and compatibility documents for decisions and promises;
- generated CLI, schema, and import inventories for derivable facts;
- tests and lint rules for enforceable invariants;
- planning records for current status and history.

## Sequencing

Start with documentation ownership and agent routing so every later move has an explicit canonical
home. The architecture split and selective package contracts can then proceed in parallel. The
executable import-graph check follows the split so it targets the durable focused guide rather than
hard-coding the file layout being retired; executable Go examples follow the package-contract
standard. Reconciliation and drift enforcement close the Thread only after both documentation
branches have settled.
