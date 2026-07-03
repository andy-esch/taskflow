---
status: reference
created: "2026-06-24"
tags: [config, storage, sync, remote, concurrency, epic-23]
---

# Remote planning repos — backends & sync (epic-23 phase 2)

Epic 23 shipped **local** decoupled planning: an impl repo points at a sibling
planning repo via `planning_repo = "../desirelines-planning"`, both sides stay
honest (`tracked_repos` + `doctor`). Phase 2 is the question the design parked as
"remote later": point `planning_repo` at something that **isn't a local
directory** — a git remote, an object store, or a service.

This weighs the backend options against the genuinely hard part — **sync and
concurrency** — and against a philosophy this project holds today: *the tool
writes files; the human drives git*. No code here; greenlit items get filed as
their own tasks (the established pattern).

## TL;DR

| Question | Short answer |
| :-- | :-- |
| Is the storage layer ready for a second backend? | **Mostly.** `core.Store` is a clean 17-method port; Service + domain don't care where bytes live. |
| What's hard-coded to local FS? | **Discovery** (root is a physical path *everywhere*), **write atomicity** (POSIX rename / `O_EXCL`), **live-reload** (fsnotify). |
| Your "hash check before write" instinct? | **Correct, and it's a real gap today.** Current concurrency is *path*-CAS only — it catches a concurrent move, not a concurrent content edit. Remote multi-writer needs *version*-CAS (optimistic concurrency). |
| Best git-native fit? | **Git as a sync layer over the existing FS store** (clone → local cache → existing code → push/pull). Reuses ~everything and gets 3-way merge for free — but it makes the tool *drive git*, reversing today's stance. |
| Cheapest thing that works *today*? | **A remote filesystem mount** (NFS/SSHFS/rclone/Dropbox). Zero code — the root is just a path. Last-writer-wins, no real concurrency. |

**Recommendation:** build one backend-agnostic foundation — a **version-aware
`Store` port** (reads return a version token; writes take an `ifVersion`
precondition → `ErrConflict` on mismatch) — then make **git-sync-over-cache**
(Path A) the first real remote, with the git-touching kept **opt-in** so the
file-writer default survives. Treat object stores (Path B) and a `serve` daemon
(Path C) as later forks, not the opening move.

---

## 1. The seam: what's clean, what's welded to the disk

`core.Store` (internal/core/store.go) is the port the Service depends on —
`TaskStore` + `EpicStore` + `AuditStore`, 17 methods, **zero filesystem leakage**.
Core unit tests already run against an in-memory fake, so the interface is real,
not aspirational. A remote backend is "implement this interface." The use-case
layer and the domain don't change.

What *is* welded to the local filesystem:

| Coupling | Where | Severity | Notes |
| :-- | :-- | :-- | :-- |
| **Root is a physical path** | config discovery: `EvalSymlinks`, `filepath.Join(root, …)`, `store.NewFS(root)`, `filepath.Rel` for display, linkback `Abs+EvalSymlinks` | **High** | The biggest surface. A `git@…`/`s3://…` value breaks containment checks, relative-path joining, symlink eval, and the whole linkback machinery. |
| **Write atomicity** | `store/atomic.go`: `writeFileAtomic` (temp + `os.Rename`), `createFileAtomic` (`O_EXCL`) | **Med** | Rename-atomicity and `O_EXCL` don't exist on object stores; need conditional-put equivalents. Move = write-new + remove-old (already a non-atomic window we accept). |
| **No cache — re-scan every call** | Service re-walks the tree on every `ListTasks`/`Summary`/`resolve` | **Med** | Fine locally; a *direct* remote backend would re-fetch the world on every keystroke. Argues hard for a local cache (Path A). |
| **Live-reload via fsnotify** | `tui/watch.go` over `Layout.WatchPaths()` | **Low** | Already **optional**: if the watcher won't start, the TUI sets `watchOff` and falls back to manual `r`. A remote backend can poll, or just degrade. |
| **`.Path` on every entity** | store stamps a file path onto each Task/Epic/Audit | **Low** | Callers don't dereference it (display + internal routing only); a remote backend returns a URL-ish handle. |

Conclusion: **the I/O seam is good; the *addressing* seam is the work.** "Where is
the planning root, and what does it mean to resolve a path inside it" is assumed
to be local everywhere config touches it.

## 2. The hard part — sync & concurrency

Today's mutation flow is *read-modify-write in one process against one disk*. The
only concurrency guard is a **path-based compare-and-swap**: every `SetFields`/
`Move`/`EditBody` re-resolves the slug immediately before writing and fails with
`domain.ErrConflict` (exit 14) if the file moved underneath it
(store/fsstore.go). There is **no mtime / hash / etag / content check** — a
concurrent edit to the *same* path is last-writer-wins. Locally that's fine
(you're the only writer; the window is microseconds).

Remote breaks that assumption. The flow becomes **fetch → edit → push**, and
between fetch and push *another machine or agent can push*. That's the classic
**lost-update** problem, and it's exactly what your "hash check before writing"
question is reaching for. The name for it is **optimistic concurrency control
(OCC)**:

> On read, capture a **version token**. On write, send it as a **precondition**:
> "apply only if the remote version still matches what I read." Mismatch →
> reject → caller re-fetches, re-applies, retries (or merges).

The clean way to land this is to **generalize the existing path-CAS into a
version-CAS in the `Store` port** — a change that helps *every* backend, local
included:

- Reads return an opaque `version` alongside the entity.
- Writes accept an `ifVersion` (or `""` for "create, must not exist" — the
  `O_EXCL` case).
- On mismatch the store returns `ErrConflict` — the *same* sentinel and exit code
  the path-CAS already produces, so the CLI/agent contract is unchanged.

Each backend supplies the token natively — no hand-rolled hashing where the
backend already has one:

| Backend | Version token | Precondition primitive |
| :-- | :-- | :-- |
| Local FS | mtime+size, or a content hash | re-stat before rename (today's CAS, extended to content) |
| Git | blob/commit SHA | non-fast-forward push is *rejected by default* |
| S3 | ETag | `PutObject` `If-Match` / `If-None-Match` (conditional writes) |
| GCS | object generation | `ifGenerationMatch` / `ifGenerationNotMatch` |
| `serve`/DB | row version | a transaction |

> **Superseded (decided 2026-07-01 — see epic
> [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]).** The
> per-backend native tokens above were the early sketch. The decided design is **one
> backend-agnostic whole-file content hash (SHA-256)** used for *every* backend. A git
> **blob/commit SHA was explicitly rejected**: it fingerprints *committed* bytes and
> misses uncommitted working-tree edits (the serve write window), so it would hand a
> reader a token already stale against the file on disk. `mtime+size` was also rejected
> (unreliable across machines/containers). The mechanism is unchanged from the row above
> — reads return the hash as `version`, writes take `ifVersion` → `ErrConflict` — only
> the token's *definition* is now fixed to a content hash. Tracked in
> [[version-aware-occ-content-hash-token-and-plain-retry]].

**Granularity is where git pulls ahead.** OCC can be *per-resource* (each task is
its own file/object — a conflict only when two writers touch the *same* task) or
*repo-level* (one version for the whole tree). Per-resource is the natural fit and
maps onto the existing one-file-per-task model. Git gives you something better
than either: **3-way merge**. Two agents completing two *different* tasks are two
non-overlapping file changes → git merges them automatically; only a genuine
same-file clash needs resolution. And the data model is merge-robust: the
worst-case merge artifact (a task that ends up in two status dirs) degrades to a
*duplicate slug*, which the tool already treats as a recoverable `ErrAmbiguous`
(exit 13), not corruption. The status==directory invariant never breaks.

So: object stores give you **per-object OCC but no merge** (conflicts are
reject-and-retry); git gives you **repo-level OCC plus free merge** that
effectively recovers per-hunk granularity. That's a strong, concrete argument for
the git path on *this* data shape.

## 3. Backend options

### Path 0 — Remote filesystem mount (works today, zero code)
NFS / SSHFS / rclone-mount / Dropbox / iCloud Drive. The root is just a path, so
`planning_repo = "/Users/me/Dropbox/desirelines-planning"` **already works**.

- ➕ Nothing to build. Good enough for *one* person across *their* machines.
- ➖ Last-writer-wins (no OCC), fsnotify may not fire over the mount, latency, and
  it's outside the tool's awareness. Unsafe for concurrent agents.
- **Action:** document it as the supported escape hatch; it costs a README
  paragraph, not an epic.

### Path A — Git as a sync layer over the existing FS store
`planning_repo = "git@github.com:me/desirelines-planning.git"`. The tool clones to
a local cache dir, the **existing `FS` store runs against the cache unchanged**,
and a thin sync layer does `pull` before reads / `commit`+`push` after writes (or
on an explicit `sync`).

- ➕ Maximal reuse — FS store, atomic writes, fsnotify (watch the *cache*),
  surgical frontmatter edits all keep working. Git-native: real history, blame,
  3-way merge, any host (GitHub/GitLab/self-hosted/bare SSH). Offline-capable
  (the cache is a full clone).
- ➖ **Reverses the "tool doesn't touch git" stance** (§4). Auth (SSH keys /
  tokens / `gh`). Decide *when* to pull/push (every command? lazy? a `sync` verb?
  a daemon?). Merge-conflict UX when two people edit the same task's same lines.
- **Fit:** the strongest git-native answer, and the cheapest *real* remote because
  it borrows the whole local stack.

### Path B — Pluggable remote `Store` (S3 / GCS / R2 / Azure) direct
A second `core.Store` implementation talking to the object API; no git.

- ➕ No git semantics to manage; per-object OCC via etags/generations is clean;
  simple ops infra (a bucket + creds).
- ➖ Most code (a real backend per provider, or one via a library like `gocloud`).
  No merge (reject-and-retry only). No fsnotify → poll or accept no live-reload.
  The no-cache/re-scan design means `list` = list-objects + conditional GETs every
  call unless you add a manifest + cache. Loses history/authorship entirely
  (you'd reinvent it).
- **Fit:** pick only if a **git-free** infra is a hard requirement (e.g. a
  locked-down cloud env). Otherwise it trades git's free wins for more code.

### Path C — Client/server (`tskflwctl serve`)
A daemon owns the store + concurrency; CLI/TUI/web are clients over HTTP/gRPC.
This is **epic 19's** territory (web companion + shared core).

- ➕ Strongest concurrency story (server-side transactions/locks), real-time
  live-reload (SSE/websocket push), one auth surface, web UI falls out.
- ➖ Most infra — something has to *run*. Overkill for "my two laptops"; the right
  shape for "a team + a web app."
- **Fit:** the long-horizon convergence point, not the opening move. Worth keeping
  the version-CAS port (§2) compatible with it.

### Comparison

| | 0 mount | A git-sync | B object store | C serve |
| :-- | :-: | :-: | :-: | :-: |
| Git-native / history | — | ✅ | — | opt |
| Free 3-way merge | — | ✅ | — | — |
| Concurrency (OCC) | none (LWW) | repo OCC + merge | per-object etag | server txn |
| Live-reload | maybe | ✅ (watch cache) | poll only | ✅ push |
| Offline | — | ✅ (clone) | — | — |
| Auth pain | low | medium (ssh/token) | medium (cloud creds) | low-ish |
| Code cost | ~0 | medium | high | high |
| Fits "tool ≠ git" | ✅ | ❌ (the big one) | ✅ | depends |

## 4. The git-native tension (commits / authorship / auth)

A thing worth naming, because it's a *values* decision, not just engineering:
**today `tskflwctl` deliberately does not touch git** — CLAUDE.md is explicit
("it writes files; the user stages/commits"), and the planning is "git-native"
only in the sense that *the files are plain text git happens to version well*. A
git-**sync** backend (Path A) changes that: the tool would `commit` and `push`
on your behalf. That raises the things you flagged:

- **Authorship / commit identity.** Whose name is on a tool-made commit? Options:
  (a) the user's `git config` identity (honest, but every `task move` becomes a
  commit by *you*); (b) a bot identity (`tskflwctl <noreply>`) — clean history,
  but loses who actually did it; (c) batch/squash (one commit per `sync`, not per
  edit) to keep history readable.
- **Commit messages.** Auto-generated (`move add-retry-backoff → in-progress`) are
  fine and greppable; the risk is noise. Squash-on-sync mitigates.
- **Auth.** SSH keys, a PAT, or shell out to `gh` (already a baked-in dependency).
  GitHub-*API* writes (no local git) are a different, heavier model — and would
  lose the reuse that makes Path A attractive — so prefer plain git protocol.
- **The honest reframe.** Keep the *default* as "tool writes files, you commit."
  Make remote-sync an **opt-in mode** (`planning_repo` is a git URL ⇒ the user has
  chosen to let the tool drive git). The philosophy stays the default; the power
  user opts into the trade. That's the clean way to have both.

## 5. Recommended path forward

1. **Foundation (backend-agnostic, do first): version-aware `Store` port.**
   Generalize the path-CAS into read-returns-version / write-takes-`ifVersion` →
   `ErrConflict`. Implement it for the *existing* FS backend (mtime+hash). This
   makes the current `ErrConflict` honest for content (not just moves), is useful
   on its own, and is the precondition for *every* remote backend. Small, safe,
   high-leverage.
2. **Decouple addressing from the local FS.** Introduce a backend/locator
   abstraction in discovery so "the root" can be a handle, not only a path — and
   so `planning_repo`'s scheme (`/abs`, `../rel`, `git@…`, `https://…`) selects
   the backend. Quarantine the linkback/`EvalSymlinks`/`filepath.Rel` assumptions
   behind it.
3. **Ship Path 0 as docs now.** A README recipe for a mounted remote dir — the
   zero-code answer for single-user multi-machine, with the LWW caveat stated.
4. **Build Path A (git-sync-over-cache) as the first real remote**, git-touching
   **opt-in**, default sync model TBD (lean: explicit `sync` verb + lazy
   pull-on-read, *not* push-on-every-write). Merge conflicts surface as a state
   the tool can show, leaning on the existing ambiguous-slug handling.
5. **Defer B and C.** Object stores only if git-free is mandated; `serve` when
   epic 19 (web) is actually on the table — both ride the same version-CAS port.

## 6. Open questions for you

- **Git stance:** OK to make the tool `commit`/`push` in an opt-in remote mode, or
  is "tool never touches git" a hard line (which would push you toward B/C or
  mount-only)?
- **Who writes concurrently?** Just you across machines (mount or git is plenty),
  or multiple agents/people at once (you want real OCC, and probably git's merge)?
- **Sync trigger preference:** explicit `tskflwctl sync` (predictable, git-like)
  vs. automatic pull-on-read / push-on-write (seamless, surprising)?
- **Does this converge with epic 19?** If a web companion is likely, Path C's
  server shape may be worth biasing toward sooner.

## Update 2026-06-30 — the real enemy is *branching* the planning, not git

A reframe that sharpens §4's git-stance tension and the epic-19 convergence.

**The fork in the road.** You cannot have both "the branchy git tree is canonical"
and "no reconciliation of mutable state." One has to give:

- **Stop branching the planning** — keep git, but run planning on a single
  timeline (a decoupled, trunk-only planning repo; epic 23's local phase already
  enables it). Branch-divergence pain is largely an *artifact of co-location*:
  planning diverges because it rides code's feature branches. Pull it into its own
  repo, run it trunk-only, and that pressure evaporates — what's left is two writers
  on one trunk, the OCC problem (version-CAS, easy), not the branch-merge problem
  (hard). The cost is losing "this PR atomically flips its own task to done" — but
  that property *is* the divergence source, so losing it is the point; code and
  planning then join by **reference (task id), not co-location**.
- **Stop making git canonical** — a service/DB is the source of truth, git a
  derived read-only mirror (Path B/C, Dolt). The most direct
  single-source-of-truth + web-UI story, at the cost of the "edit a markdown file
  in a PR and it round-trips" identity. The exit ramp if the center of gravity
  shifts to a team/web app — not the opening move.

Both project goals — less merge friction *and* an out-of-terminal web UI — point at
the *first* resolution: a single canonical trunk, not abandoning git.

**Serve-owns-git dissolves the "the UI would have to provide git operations"
worry.** The browser need not drive git, so the web UI need not expose it. In the
Path C / epic 19 shape the `serve` daemon is the **single writer** to the
trunk-only checkout *and* the concurrency authority: it reads/writes through
`core.Service` with OCC and batches commits/pushes underneath (bot identity, or one
squashed commit per sync). The web user edits a task and sees it save; git is an
invisible server-side detail. The model then splits cleanly:

- **Local CLI/TUI** keeps today's "tool writes files, you commit" stance —
  git-native, offline, PR-able, unchanged.
- **Hosted web** opts into the server driving git against the same trunk — exactly
  §4's opt-in git-drive, with the server *also* owning concurrency, which is the
  piece that lets git stay underneath without the UI ever speaking it.

That makes §6's "is 'the tool never touches git' a hard line?" answerable: keep it
hard *for the local default*; relax it *only* inside an opt-in server. The residual
merge pain this addresses (concurrent edits to a file's mutable fields) and the
content-vs-workflow-state split behind it are in the 2026-06-30 update of
[[2026-06-24-task-storage-model-files-logs-or-versioned-db]].

## Update 2026-06-30 — two writable authorities is the trap; "central as a git client" is the way out

A deeper pass on §2 / §4 / §6: *how do you keep planning git-native AND update a
central source live, without conflict?* Two candidate mechanisms came up; both fail
the same way, and the failure points at the fix.

**Root cause: dual authority.** Any design where git *and* a central store can each
accept writes independently is a **sync problem** — drift, failed-apply log
archaeology, and "which side wins?" are its permanent signature, not bugs you can
tool away. The only escape is **one authority, with the other a pure derived
function of it.** Judge every proposal by one test: *how many things can be written
independently?* The answer has to be one.

**Mechanism 1 — GitOps (CI applies changes to central).** Separate two things. The
*messy* part — "apply a diff, watch it fail, read the GHA log, iterate" — is an
artifact of framing it **imperatively**. Real GitOps **reconciles to desired
state**: read git HEAD, make central *match* it, idempotently; re-running converges,
so there's no failed-patch archaeology. But the *real* limit survives that fix:
CI-applies-to-central only flows **git → central**. It buys a fast read projection,
**not a web write path** — a web mutation still has to become a git commit by some
other means. And it doesn't remove conflict; it delegates it to git's merge and lets
central mirror the outcome. Fine — but only once git is already canonical.

**Mechanism 2 — `tskflwctl apply` (write central, then commit on success).** A
textbook **dual-write**, with the classic failure: the process dies (or the push
fails) *between* the two writes and the stores disagree with no record of intent. A
drift-detector GHA *detects* that; it can't *prevent* it, and remediation still needs
a hand-picked winner. Making dual-write safe needs an **outbox / transaction-log**
(durably record intent; a worker drives both sides to convergence) — heavy machinery
for a planning tool. And "central-first, then commit" quietly makes **central the
authority and git a lagging mirror**, which guts the git-native motivation (why is
git hand-editable if it only trails?).

**The fix is hiding inside Mechanism 2.** The drift exists only because "central" is
assumed to be a *different kind of store* than git. Collapse that — make **central
itself a git repo** — and:

- `apply` becomes **pull → merge → commit → push**.
- There is **one authority** (git); the "central store" is just *the canonical clone*.
- Drift is **impossible by construction** — nothing of a different shape exists to
  drift from; a "reconcile" is just `git pull`.
- Conflicts route through the one mechanism built for them: git's **3-way merge**.

So `apply` is already drift-free *the moment "central" is git, not a foreign DB.* The
thing being reached for is just **the tool driving git sync**.

**What it is, concretely: serve-owns-git (confirmed from a second angle).**

- One **server owns the canonical clone** and is the single live writer: a read/write
  API (web talks to it, no git in the browser), writes serialized through
  `core.Service` with OCC, commits/pushes batched.
- The **local CLI stays git-native** — edit files in your own clone, commit/push as
  today. The server is just *another git client*, the privileged high-frequency one.
- **Conflicts are made rare by data shape, not by avoiding git:** per-entity files +
  stable paths + ULID identity ⇒ writers on *different* tasks never collide; the
  high-churn field (`status`) is the candidate for an append-log so even same-task
  flips concatenate. Live writes funnel through the one server, so human-vs-server
  clashes are occasional and handled by ordinary `pull`.
- **Residual cost (honest):** a human editing their own clone offline and pushing can
  still conflict with the server — ordinary git, rare when the server is where live
  activity flows. You don't escape git's conflict model; you shrink its surface until
  it almost never fires.

**CI / GHA — right role vs wrong role.** Good at **validation gates** on PRs (`lint`,
`schema`, the docs-check gate) and rebuilding an **idempotent derived read cache**
from HEAD (declarative, convergent, stateless). Bad at being the **transactional
write path** (Mechanism 1) or the **drift remediator** (Mechanism 2) — those need a
long-running serializer holding a working tree (a server), not a batch job firing
after the fact and leaving logs to read.

**Net (sharpens §6).** No mechanism makes two *writable* authorities cheap; that's the
shape of the problem, not a tooling gap. "Git-native AND a central writable source,
low-conflict" is reachable **only** by making the central thing a **git client, not a
git peer** — one hosted git authority, a server as its privileged live writer + API,
the CLI still git-native offline. This is the same serve-owns-git / version-CAS
foundation §2–§5 land on, reached here from the dual-write angle. So §6's "sync
trigger" and "git stance" resolve together: the *tool* drives git **only inside the
opt-in server**; the local default stays "you commit."

## Update 2026-06-30 — how web writes land: direct-to-trunk default, opt-in PR/approval escalation

Given serve-owns-git, the open question is *how* the server's writes land in history —
straight to trunk, or via PRs. The reframe that settles the default: **a PR is a
branch.** PR-per-change reintroduces the divergence trunk-only was built to kill —
canonical state forks into trunk vs the open PR(s), the UI must choose which it shows,
two web edits to one task in two PRs conflict, PRs go stale and need rebasing. So the
default leans *away* from PRs; the burden is on PRs to justify the cost.

What people reach for PRs to get, decomposed — only one actually needs a PR:

| Want | Need a PR? |
| :-- | :-- |
| **Validation** | **No** — the server validates inline (`core.Service` + OCC + lint), rejecting synchronously. The CI-validation gate belongs on the *human-pushes-files* local path (a hand edit can be invalid), not the server path (which can't emit invalid state). |
| **Audit trail** | **No** — every commit already is the who/when/what; `git log` is the trail. Needs attribution, not PRs. |
| **Review / approval** | The **only** PR-exclusive benefit — and only for *substantive* changes with an actual reviewer. |

And the unlock: **review ≠ PR.** A gate is cleanest as an **in-server approval queue** —
pending changes sit in the server's state; on approval it commits *straight to trunk*.
You get the human gate with no branches, no divergence, no host-API coupling. A PR is
one (costly) implementation of review; use a real PR only if you specifically want
GitHub's review UI/discussion as the surface.

Stakes are lower than the code instinct suggests: **"protect main" doesn't transfer to
planning data.** A bad code commit breaks a build; a bad planning commit (wrong status
flip, typo) breaks nothing and `git revert`s in a line. The thing that makes
direct-to-main scary for code mostly evaporates here.

### The two tiers

- **Default — direct to trunk, no PR.** The high-frequency, low-controversy 95% (status
  flips, field edits, snoozes). The server validates inline and commits,
  **batched/squashed per session** for readable history, authored by a
  **`tskflwctl-bot` identity with a `Co-authored-by:` trailer** naming the web user
  (honest attribution, clean log). This is what preserves SSOT and the out-of-terminal
  ease the web UI exists for.
- **Escalation — explicit "propose for review," rare, opt-in.** Substantive changes
  (re-scoping an epic, bulk restructuring). Default implementation: the **approval
  queue** (commit-to-trunk-on-approve). Reach for a real **PR only** to get GitHub's
  review UI — accepting that this is the one path where branching returns, so keep those
  branches short-lived.

### PR-tier merge policy (when a real PR is used)

Require the feature branch be **rebased / up-to-date with main before merge**, so any
conflict is resolved **on the branch, by the author, before it touches main** — main
never sees a conflicted merge and stays linear. Land it as a **squash-merge** (one tidy
commit per proposal, bot + `Co-authored-by` trailer). The principle this encodes:

> **The cost of branching is borne by whoever opted into the branch.** The direct tier
> stays frictionless; the brancher pays the reconciliation — acceptable *because* the
> PR tier is the rare, deliberate path.

Two honest consequences:

- **Non-git web author** ⇒ the server must surface conflict resolution in the web UI (or
  bounce: "this went stale against current state — re-apply?"). An **agent** author
  handles it more gracefully — it just re-runs its change against fresh main. This
  friction is fine for the opt-in tier, and is exactly *why the default must stay
  direct-to-trunk*.
- **Portability:** "require up-to-date before merge" is a GitHub branch-protection
  feature; a bare / self-hosted remote enforces the same rule in server logic, not host
  config.

**Net:** web writes default to the single trunk timeline (SSOT intact); branching is
opt-in, short-lived, and self-reconciling by the brancher. This resolves the "how do
web writes land?" question the serve-owns-git model left open.

## Related

- [[2026-06-24-task-storage-model-files-logs-or-versioned-db]] — the on-disk data
  model (status-as-directory vs frontmatter/log/DB). Same root cause as the OCC
  work here: mutable state encoded in the path. Decide it *before* locking OCC's
  shape.
- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]] — the
  storage / read-model / OCC foundation this remote work rides on.
- Epic [[23-point-an-impl-repo-at-an-external-planning-repo]] (local phase, done).
- Epic 19 — web companion / shared core (Path C convergence).
- Storage seam: `core.Store` (internal/core/store.go), `store/atomic.go`,
  config discovery (internal/config/config.go).
