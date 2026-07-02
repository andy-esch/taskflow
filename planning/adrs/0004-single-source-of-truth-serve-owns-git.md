---
status: proposed
date: "2026-06-30"
deciders: [andy-esch]
tags: [adr, storage, sync, concurrency, web]
supersedes: []
superseded_by: null
---

# ADR-0004: Single source of truth — git canonical, one live writer (serve-owns-git)

> Follows the ADR format established in [[0001-adopt-adrs]]. Builds on
> [[0003-stable-key-id-addressed-storage]], which fixed the *on-disk* model and
> deliberately left *sync / single-source-of-truth* to this ADR (ADR-0003's identity is
> coordination-free precisely so it doesn't presuppose this decision). Full rationale:
> [[2026-06-24-remote-planning-repos-backends-and-sync]].

## Context and Problem Statement

The project is **git-native**: planning is plain markdown a human edits and commits.
Two forces strain that:

- **Branch divergence.** By definition the same task file can exist in several branches,
  so it must be reconciled on merge/rebase — the friction that motivated this thread.
  (ADR-0003 removes the *rename*-induced duplicates; it does **not** remove concurrent
  edits to the same file's mutable fields across branches — that residue is inherent to
  "the branchy git tree is canonical.")
- **An out-of-terminal web UI** (epic 19) needs a **non-local, writable** source — and
  it can't sanely ask a browser to drive git.

The instinct is to keep git **and** add a central writable store. The trap: **any design
where git and a central store can each be written independently is a sync problem** —
drift, failed-apply log archaeology, "which side wins?" are its permanent signature, not
bugs to tool away. The only escape is **one authority, with the other a pure derived
function of it.** (This rules out dual-write `apply`-then-commit and CI-applies-a-diff
GitOps — both create two authorities; see the spike.)

So the real question is *which* one authority — and how a web app writes through it
without re-introducing the branch divergence we're trying to escape.

## Considered Options

- **A — Trunk-only git, no central writer (baseline).** Decouple planning into its own
  repo (epic 23, local phase shipped) and run it **trunk-only**; the tool writes files,
  you commit. Removes most divergence (which was largely an artifact of planning riding
  *code's* feature branches). Simple and fully git-native — but offers **no web write
  path**: a browser would have to drive git itself.
- **B — Serve-owns-git: git is the one authority, a server is its privileged live
  writer (chosen).** The canonical planning repo is git; a `serve` daemon owns the
  canonical clone and is the **single live writer** + API + concurrency authority,
  committing/pushing on behalf of web users. The local CLI stays git-native and offline.
  No second store ⇒ no drift; conflicts are git's, made rare by data shape. This is A
  **plus** a web write path, without a second authority.
- **C — Make a service/DB canonical, git a derived read-only mirror.** The most direct
  SSOT + web story, but it abandons "edit a markdown file in a PR and it round-trips."
  The **exit ramp** if the center of gravity ever shifts from documents to a team/data
  app; rejected for now, to keep git canonical.

## Decision

Adopt **serve-owns-git**: one git authority, a single live writer for the central case,
the local default unchanged.

### 1. Git is the single canonical authority; planning runs on one timeline

Planning lives in its **own repo** (epic 23), operated **trunk-only**. There is no
second writable store, so there is no drift by construction; "reconcile" is just
`git pull`.

### 2. The local default is unchanged — "tool writes files, you commit"

The CLI/TUI stay git-native and offline; the tool does **not** touch git locally. This
posture is preserved as the default; everything below is the *opt-in* central case.

### 3. A central / web writer is serve-owns-git

A `tskflwctl serve` daemon owns the canonical clone and is the **single live writer**:
a read/write API (the web UI talks to it — **no git in the browser**), writes serialized
through `core.Service` with **version-aware OCC** (whole-file content hash + plain
retry — epic 24), commits/pushes **batched** underneath.
It is *another git client* — the privileged, high-frequency one — **never a second
authority**.

### 4. Conflicts are git's, minimized by data shape

Per-entity files + the stable paths/ids of ADR-0003 mean writers on **different**
entities never collide. The high-churn field (`status`) is a candidate for an append-log
later if same-entity flips ever clash. A human editing their own clone offline and
pushing can still conflict with the server — ordinary `git pull`/merge, rare when live
activity flows through the one server.

### 5. Web writes land in two tiers

- **Default — direct to trunk, no PR.** The trivial 95% (status flips, field edits). The
  server validates inline and commits, **batched/squashed per session**, authored by a
  **`tskflwctl-bot` identity + a `Co-authored-by:` trailer** naming the web user.
- **Escalation — explicit "propose for review," rare, opt-in.** An **in-server approval
  queue** (commit-on-approve to trunk) by default — *review ≠ PR*. A real PR only to get
  GitHub's review UI, governed by **rebase / up-to-date-before-merge**, so conflicts
  resolve **on the branch, by the author** (whoever opted into the branch pays the
  reconciliation), landing as a squash-merge. Short-lived branches only.

### 6. CI / GHA's role

Validation gates (`lint` / `schema` / docs-check) and **idempotent** derived-cache
rebuilds — declarative, stateless. **Not** the transactional write path or a drift
remediator; those need a long-running serializer (the server), not a batch job firing
after the fact and leaving logs to read.

## Consequences

**Positive.**

- Single source of truth **without abandoning git** — plain markdown stays canonical,
  diffable, PR-able, offline.
- The web UI **never exposes git**; the server hides it. Answers "out-of-terminal use"
  without asking a browser to merge.
- Drift is impossible by construction (one authority); conflicts are rare by data shape.
- Rides existing seams: ADR-0003's coordination-free id, the version-aware `Store` port,
  `core.Service` as the one write path, and the read-model projection shared with the
  board.

**Negative / cost.**

- **New infra.** Something must *run* (the daemon), plus auth (SSH key / token / `gh`).
  Overkill for "my two laptops" (where a trunk-only clone is plenty); the right shape for
  "a team + a web app."
- **The tool drives git inside the server** — reversing the "tool never touches git"
  stance, but **only there**, opt-in; the local default survives.
- **Web conflict UX** for the PR tier — a non-git web author needs a resolution surface
  (or a stale-bounce); an agent author just re-runs against fresh main.
- **A coordination point.** The single live writer is a bottleneck / SPOF for the
  *served* path (the local CLI path has none).

## Out of scope (deferred — NOT decided here)

- The actual **`serve` daemon + web app** implementation (epic 19) — this ADR decides the
  *model*, not the surface.
- **Object-store / non-git backends** (epic 23 phase 2, Path B) — only if a git-free
  infra is ever mandated.
- **DB-canonical** (option C) — the exit ramp; revisit only if the center of gravity
  shifts to a team/data app.

## Amendments

<!-- Append-only, dated entries added AFTER this ADR is accepted. Format:
     ### 2026-07-01 — <what changed and why> -->

_None yet (still `proposed`)._

## Related

- Sync / concurrency rationale, the dual-authority trap, and the web-write thread:
  [[2026-06-24-remote-planning-repos-backends-and-sync]].
- The on-disk model this builds on: [[0003-stable-key-id-addressed-storage]].
- Decoupled-planning-repo foundation (local phase shipped): epic
  [[23-point-an-impl-repo-at-an-external-planning-repo]].
- Web companion (the `serve` surface this enables): epic
  [[19-web-companion-apps-over-a-shared-core]].
- Storage / identity / OCC foundation:
  [[24-data-model-evolution-stable-key-storage-read-model-content-occ]].
- ADR format this follows: [[0001-adopt-adrs]].
