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
