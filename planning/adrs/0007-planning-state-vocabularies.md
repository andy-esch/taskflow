---
status: accepted
date: "2026-08-25"
deciders: [andy-esch]
tags: [adr, domain, planning-model, cli, vocabulary]
supersedes: []
superseded_by: null
---

# ADR-0007: Planning state vocabularies — shared words, written by the tool

> Follows the ADR format established in [0001-adopt-adrs](0001-adopt-adrs.md). Sits beside
> [0003-stable-key-id-addressed-storage](0003-stable-key-id-addressed-storage.md), which
> governs where entities live and what is authoritative in frontmatter; this one governs the
> closed vocabularies inside a document's BODY and how they are written.

## Context and Problem Statement

Two body-level fields carry a closed vocabulary: an audit finding's `**Status:**` and,
until recently, an acceptance criterion's checkbox. Both were maintained by hand-editing
markdown, and both drifted.

The finding vocabulary drifted first, and visibly. `landed` was legal in code, absent from
the consumer-facing table, and contradicted by an assertion in the very comment that claimed
to hold the two in sync (finding M3 of `2026-08-17-finding-status-surface`). No audit in the
corpus had ever used it. Meanwhile the corpus was improvising a word the vocabulary lacked:
7 of 13 `deferred` findings were not deferrals but handoffs, written in prose as
`deferred → tracked in task X` by two authors months apart.

The criterion "vocabulary" was a single bit. An unchecked box meant *not yet*, *won't do*,
and *no longer applies* all at once, so a task could sit at 3/7 forever with no way to say
why the other four were not coming.

Both problems have the same root: **a value the tool cannot write is a value nobody can be
held to.** Linting a field the tool only reads catches typos after they land; it cannot stop
the vocabulary and its documentation from parting ways.

## Decision

**1. Criterion state is a checkbox plus an optional suffix.**

```
- [x] Criterion that is done
- [ ] Criterion still to do
- [ ] Criterion parked · **deferred:** waiting on the schema ADR
- [ ] Criterion abandoned · **wontfix:** superseded by the table layout
- [ ] Criterion handed off · **tracked:** carried by 6g3ag8py12y9
- [ ] Criterion that stopped applying · **n/a:** the tile grid was dropped
```

The bracket keeps its existing binary meaning and the suffix refines the NOT-MET case.
Chosen over a replacement marker specifically because it is **additive**: every criterion
ever authored still parses, no migration is needed, and a body written before the vocabulary
existed serialises unchanged. Each non-binary state REQUIRES a reason — a deferral with no
why is indistinguishable from an oversight, which is the defect the vocabulary exists to
remove.

**2. Words that mean the same thing in two places are spelled once.**

`domain/resolution.go` holds the shared pool — `deferred`, `wontfix`, `tracked` — and each
entity declares its own full set from that pool plus its own additions. It is modelled as an
**overlap, not a subset**: `met` is not a finding status and `superseded` is not a criterion
state, so claiming either set contains the other would be a lie, and the lie is precisely
what lets them diverge unnoticed. A test fails if a shared word is spelled differently in
either set, and `theme.CriterionState` delegates the shared words to `theme.FindingStatus`
so one word renders as one mark.

**3. `tracked` replaces `landed`.** `landed` had zero corpus uses; `tracked` had seven
improvised ones. It means "transferred, not abandoned" and REQUIRES a destination
(`tracked by <id>`) — a handoff with nowhere to follow is the improvisation it replaces. For
an audit it counts toward **done**, because the audit's interest concludes when a finding is
transferred; for a task it does NOT count as met, because routing a criterion elsewhere
splits the work rather than finishing it. That asymmetry is deliberate: an audit is a report,
a task is work.

**4. Every closed vocabulary gets a validated, atomic write verb.** `task ac` owns criterion
state; `audit finding` owns finding status and its `**Resolution:**` paragraph. Reads stay
tolerant so `lint` can REPORT malformed data already on disk; writes refuse to create it.

**5. Invariants are enforced at write time where a linter would only report them.**
`task complete` refuses a task with an unmet, unexplained criterion — the counterpart of
`audit close` refusing while findings are open. The gate is only tolerable BECAUSE of
decision 1: before it, "refuse on unmet criteria" meant "tick every box or never finish".
Now the gate blocks silence, not disagreement.

## Consequences

- Schema 1.45–1.48 carry the wire half. 1.47 is **not additive** — `landed` is no longer
  accepted — though no audit ever used it.
- An audit's headline percent is the **settled** share, so 100% is exactly `ready to close`.
  It previously counted only fixed findings, which let a closed, fully-resolved audit
  display "77% fixed" on the same line as its own ready-to-close marker.
- The vocabularies are now derived into `schema`/`schema audit` rather than transcribed, so
  the M3 failure cannot recur in the same shape.
- Hand-editing these fields still works and still lints. This ADR does not forbid it; it
  says the tool owns the well-formed path and the docs point at it.

## Alternatives considered

- **A separate criterion enum overlapping the finding one.** Rejected: two tables free to
  drift is exactly what M3 recorded.
- **A subset relationship.** Rejected as untrue in both directions.
- **A new marker instead of checkbox+suffix** (e.g. `- [~]`). Rejected: it would have needed
  a migration and broken every existing renderer, for no gain over a suffix.
- **Requiring an id-shaped destination for `tracked`.** Rejected: a destination is
  legitimately an epic, an ADR, or an external issue, and a Crockford regex would reject
  `tracked by ADR-0003`. The check worth having is *resolution* — lint flagging a named id
  that does not exist — which is not built.

## Provenance

Decided across `6g31g9f8x4cv`
(*let-an-acceptance-criterion-say-more-than-done-or-not-done*), whose body holds the
long-form reasoning and the rejected alternatives in more detail, and audits
`2026-08-17-finding-status-surface`, `2026-08-24-planning-state-vocabulary` (an adversarial
external review), and `2026-08-24-finding-note-and-vocabulary-selfreview`.
