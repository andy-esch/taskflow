---
schema: 1
id: 6g9czp7g9pt3
status: in-progress
description: Restructure project guidance around clear source ownership, scoped architecture and Go package contracts, and executable drift checks.
goal: Agents and contributors load the minimum authoritative guidance for a change, while code and CI expose documentation drift.
created: "2026-09-12"
tags: [documentation, architecture, agents, go, dx]
tasks: [6g63hjm7cp6w, 6g6x7e2ef37r, 6g9cz5saenme, 6g9cz5sk3b0d, 6g9cz5sv6kck, 6g9cz5t37dc0, 6gg7e59mm68g, 6ggxymjbf86r, 6gjthp9frrhq]
updated_at: "2026-10-11"
started_at: "2026-10-11"
---
# Thread: Make documentation layered, executable, and agent-navigable

**Goal.** Agents and contributors load the minimum authoritative guidance for a change, while code and CI expose documentation drift.

## Context

At Thread creation, `docs/ARCHITECTURE.md` was estimated at roughly 8,000 words, mixing stable boundaries, current inventory, subsystem manuals, historical review outcomes, and dated status. Mutable facts also appeared in `README.md`, `CLAUDE.md`, package comments, ADRs, generated CLI docs, schema output, and planning records. The 2026-10-11 research measured the architecture baseline at 11,466 words; the ownership slice replaces the old agent entry point, while the architecture split remains queued.

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

Start with the bounded [ownership/agent-loader research](../tasks/6gjthp9frrhq-research-documentation-ownership-and-portable-agent-routing.md),
then documentation ownership and agent routing so every later move has an explicit canonical
home. The [research snapshot](../research/6gjggjw01raw-documentation-ownership-and-portable-agent-routing.md)
informs shared `AGENTS.md` ownership and native Claude/Antigravity discovery without wrappers or
a new docs platform.
The architecture split and selective package contracts can then proceed in parallel. The
executable import-graph check follows the split so it targets the durable focused guide rather than
hard-coding the file layout being retired; executable Go examples follow the package-contract
standard. Reconciliation and drift enforcement close the Thread only after both documentation
branches have settled.

The [hexagonal dependency-policy ADR](../tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md)
is a separate decision-record slice after the concrete composition boundary settles. It should
inform the focused architecture guide, not duplicate the executable import inventory or silently
ratify a new policy without user acceptance.

The adapter-neutral refactor is completed. Its short
[planning-data checklist](../../docs/ARCHITECTURE.md#planning-data-change-checklist) is a routing baseline for this Thread,
not a replacement for documentation ownership, architecture restructuring, or package contracts.
The [direct-I/O/framework import guard followup](../tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md)
follows the policy ADR and precedes final reconciliation, so guidance distinguishes executable
import checks from snapshot/recovery guarantees proven by tests. Current core/domain imports comply;
the new task prevents future bypasses rather than repairing a present adapter violation.

The checklist is a section of the architecture reference, not a new peer document. A future
focused-guide extraction should follow this Thread's ownership and layout decisions.

That extraction now places the canonical checklist in the
[planning-data guide](../../docs/architecture/planning-data.md#change-checklist); the old index
anchor remains a forwarding route. Use the focused owner for future changes rather than copying
the checklist into a package overview.

## Ownership checkpoint (2026-10-12)

Research and ownership delivery are complete. The
[external review](../audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md)
found no lost safeguards or hidden wire changes; its two low findings are fixed with compiler-valid
mutation evidence and clearer limits on what prose/declaration tests prove. The current frontier
contains the architecture split, selective package contracts, and the dependency-policy ADR.

The review-sandbox push/attestation work belongs to
[a separate hardening Thread](6gk06gktr1q5-harden-isolated-review-workspace-safety-and-delivery-evidence.md),
not this Thread's completion criteria. Architecture restructuring, package docs, executable
examples/inventories, and final reconciliation remain unfinished here.

## Architecture extraction checkpoint (2026-10-12)

The architecture split is complete and ready for PR handoff. A short index
routes to six guides with code/test links and package-comment return paths. The dated feature
snapshot is preserved as task history; named exceptions, CAS/receipt limitations, preview rules
and source-authority constraints remain current guide content. A bounded navigation test runs in
the ordinary suite, but import-map enforcement, selective package overviews, executable examples
and broader reconciliation still belong to the sequenced tasks. Claude's four Low findings were
verified and fixed; Antigravity's useful probes do not support its exhaustive all-falsified claim.
The guard now rejects non-navigable decoys and checks retained anchors referenced by tasks/Threads.
Both audits are settled with explicit evidence limits, not treated as completion of downstream slices.

## Import-map checkpoint (2026-10-13)

The import-map task is in progress with a scoped checker implemented for exact default-build
production internal package/edge facts. The guide's presentation shorthand is replaced with
literal paths and explicit leaf rows; support/test sources are excluded without hiding production
imports into them. Native probes detect real edge drift, while formatting and metadata fixtures
pin the coverage limits. This is descriptive verification, not acceptance of the pending policy
ADR or completion of the pure-layer guard, package contracts, examples or reconciliation tasks.

Claude's import-map review found three real gaps, now fixed with warm-cache, persisted
GOFLAGS and compiler-valid uncovered-target regressions. Antigravity's follow-up adds
source-token rejection, CRLF equivalence and real-module root/support/near-collision
coverage. Both audits are closed and the import-map task is completed. Wildcard limits
are explicit rather than overstated as universal production coverage; import policy
remains separate. The healthy frontier has selective package contracts and the policy
ADR eligible next, while reconciliation still waits on its other prerequisites.
