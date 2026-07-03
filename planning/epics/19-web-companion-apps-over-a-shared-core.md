---
status: active
description: Explore a web UI sister to the TUI (tskflwctl serve first), module separation into core + bundled apps, and GitHub discoverability
priority: low
tags: [web, architecture]
created: "2026-06-13"
---

# Web companion apps over a shared core

**Goal.** Explore a web UI sister to the TUI (tskflwctl serve first), module separation into core + bundled apps, and GitHub discoverability

## Why this is its own epic

The web companion is a **third primary adapter** over the same `core.Service` the
CLI and TUI use — not a fork of the data. It earns its own epic because it adds a
runtime shape the terminal adapters don't have (a long-running `serve` process,
HTTP/JSON transport, auth) plus a module-separation question (core + bundled
apps). Its read side has a concrete first slice already in view — the materialized
projection / board (below).

## Read model / projection (convergence)

The generated **board** scoped in
`planning/research/2026-06-24-task-storage-model-files-logs-or-versioned-db.md`
is this epic's read side in embryo: **one `core` projection** of planning state
(`core.Summary()` is today's seed), rendered by every adapter — a committed
`BOARD.md`, the TUI board, a CLI `board`, and here as a JSON/HTTP read endpoint.
So `serve`'s read path is "expose the projection," not new logic; writes funnel
through `core.Service` with the version-aware OCC (see
[[2026-06-24-remote-planning-repos-backends-and-sync]]).
Building the board (epic 23 / storage spike) is a down payment on this epic.

## Current aim (2026-07)

Sharpened from the earlier "explore everything" framing:

- **Build the local server now** — `tskflwctl serve` as a third primary adapter,
  peer of `cli`/`tui`, over the *existing* local FS store. No auth, no GitHub API.
- **Single-user throughout.** The hosted step is a *later* deployment of the same
  server for **one user across machines**, not multi-tenant. This keeps auth to a
  session, not an org model, and lets serve-owns-git (below) stay a single-writer
  story.
- **Freshness is the headline requirement**, not a nice-to-have (its own section
  below). A board that lags behind `task move` / cron-agent writes is a stale
  snapshot, and staleness is the specific thing this UI must not have.

The tier ladder (serve → hosted single-user → WASM) from
[[research-doc-web-companion-directions-serve-shared-core-github-backed-store]]
still holds; we are committing to tier 1 with tier 2 (hosted, still single-user)
as the deliberate next rung. WASM stays a curiosity.

## Freshness is the headline requirement (live board)

The read side is "expose the projection" — but a *live* projection. Where a task
sits on the board must reflect reality within a second of any writer touching the
tree, whether that writer is the web UI itself, a local `task move`, or a
**concurrent cron AI-agent** (this repo already has those). The design:

- **Event source: reuse the watcher.** The TUI already live-reloads via
  `Layout.WatchPaths()` (fsnotify) — `serve` subscribes to the same signal instead
  of polling. When a file changes, recompute the affected column(s) of the
  projection.
- **Transport: Server-Sent Events (SSE), not polling.** One-way server→browser is
  exactly SSE's shape; the browser auto-reconnects; it's plain HTTP (no websocket
  upgrade). Stdlib is sufficient — a **broker** goroutine (register / unregister /
  broadcast channels), a per-connection buffered channel, **non-blocking send with
  `default:` to skip a slow client**, `r.Context().Done()` cleanup,
  `http.NewResponseController(w).Flush()`, and a ~30 s heartbeat comment to keep
  proxies from timing the stream out. `http.ResponseWriter` is not concurrency-safe,
  so heartbeat + event writes share one per-connection `select` loop.
- **The write→broadcast loop.** A UI mutation goes through `core.Service`, lands as
  a file change, the watcher fires, the broker broadcasts, every open board patches
  the moved card. The mutating client's own optimistic move is confirmed by the same
  event — no special-casing.
- **Optimistic UI + OCC reconciliation.** The browser moves the card immediately;
  if the write comes back `409` (stale version, below), it snaps back and refetches.
  This is what keeps drag-drop feeling instant without lying about state.

**Note on the hosted rung:** fsnotify watches a *local* tree. In the serve-owns-git
model (below) the server owns the canonical clone and is the single live writer, so
its own commits are the event source — the watcher still works because the server
writes files locally before pushing. A push from *another* clone (the user's laptop)
is caught on the next pull, not instantly; acceptable for single-user.

## Content OCC — reuse epic 24, don't reinvent

"Tasks content matching" — making a write apply only if the task hasn't changed
underneath the writer — is **already scoped as
[[version-aware-occ-content-hash-token-and-plain-retry]] under epic
[[24-data-model-evolution-stable-key-storage-read-model-content-occ]]**. This epic
does **not** own that work; it *consumes* it and adds the HTTP surface:

- **The version token is a content-hash, canonically — everywhere.** Epic 24
  generalizes the current path-CAS into a version-CAS: reads return an opaque
  `version`, writes take `ifVersion` → `ErrConflict`. That `version` is a **hash of
  the exact bytes the read returned**, computed the same way on every backend; over
  HTTP it rides as the **`ETag`**. One definition flows through all layers, so
  conflict handling is a single code path local and hosted.
- **Do *not* use the git blob SHA as the token.** It is tempting (git's blob SHA
  *is* a content hash) but it hashes the **committed** bytes, not the live ones. In
  the serve-owns-git model the server writes the working tree and commits *after*, so
  between write and commit the file on disk diverges from HEAD's blob — a token from
  the blob SHA would be stale against the very file a reader is looking at, silently
  defeating OCC. Hash the working-tree/returned content instead. A backend's native
  content-addressed id may be used only when it *provably equals* the hash of the
  bytes returned (e.g. an S3 ETag, or a GitHub API store that only ever serves
  committed content with no working-tree layer) — never in a store that has an
  uncommitted write window.
- **The HTTP mapping this epic adds.** Reads emit `ETag: "<version>"`. Mutating
  requests (`POST`/`PATCH`) must carry `If-Match: "<version>"`; the handler compares,
  and on mismatch returns **412 Precondition Failed** (include the current ETag so
  the client resyncs). `domain.ErrConflict` — which the CLI maps to exit 14 — maps to
  412/409 here. Missing `If-Match` on a mutation ⇒ `400` (require it; don't silently
  last-writer-win from a browser).
- **Design serve around the token from day one**, even in the read-only slice, so
  the write path is a small addition rather than a retrofit. If epic 24's version-CAS
  isn't landed when serve starts, serve can compute a content hash itself as a stopgap
  and swap to the port's token later — same wire contract.

## Operational readiness (server ≠ CLI)

The CLI is one-shot; a long-running server inherits a runtime the CLI never had.
These are net-new for `serve` and easy to forget because the adapter reuse hides
them:

- **Lifecycle.** `signal.NotifyContext(ctx, SIGINT, SIGTERM)`, run
  `ListenAndServe` in a goroutine, treat `http.ErrServerClosed` as normal, drain via
  `server.Shutdown(ctx)` with a timeout. Broadcast a `shutdown` SSE event first so
  browsers don't thrash-reconnect.
- **Timeouts (security, not just hygiene).** Set `ReadHeaderTimeout` (Slowloris),
  `ReadTimeout`, `WriteTimeout`, `IdleTimeout` explicitly on `http.Server` — gosec
  flags a missing `ReadHeaderTimeout`, so CI would too. **Caveat:** a short
  `WriteTimeout` kills SSE streams; the SSE route needs its own handler without a
  write deadline (or `NewResponseController.SetWriteDeadline` per-write).
- **Operational logging.** The CLI's injected `io.Writer` is *user* output; a server
  also needs *request* logging via `log/slog` (JSON handler). Two different axes —
  keep them separate.
- **Browser trust boundary.** Mutations from a browser need CSRF defense — for a
  single-user local server, `SameSite=Strict` cookies + a same-origin (`Origin`/`Sec-
  Fetch-Site`) check is enough; add a token only if that proves insufficient. Cap
  request bodies with `http.MaxBytesReader`. The domain still validates content, but
  request *decoding* limits are the server's job.
- **Config.** Serve needs bind address / port and which planning repo to serve
  (today's `-C/--chdir` anchor plus a `--addr`). Keep it flag/env, no config-file
  sprawl for the local rung.
- **Testing.** The whole adapter is `httptest`-able because it's stdlib `http.Handler`
  over the same `core.Service` the CLI uses — assert status codes, ETag/412 behavior,
  and SSE frames without a real socket.

## Recommended stack & dependency budget

Bias — matching this repo's anti-bloat, stdlib-first stance — is that **tier 1 adds
essentially zero dependencies**; the heavier libraries appear only in the tier-2
GitHub store, quarantined to that one package.

| Concern | Tier 1 (local `serve`) | New deps |
| :-- | :-- | :-: |
| Routing | stdlib `net/http.ServeMux` (Go 1.22+ method+path patterns; we're on 1.25) | 0 |
| HTML | `html/template` + `//go:embed` (defer `templ` until the template surface hurts — it adds a codegen build step) | 0 |
| Interactivity | **htmx**, embedded as one ~14 KB static asset; degrades without JS | 0 Go deps |
| Live updates | hand-rolled SSE broker (stdlib) — or `tmaxmax/go-sse` if we want spec-compliance | 0 |
| Lifecycle / logging | `signal.NotifyContext`, `server.Shutdown`, `log/slog` | 0 |
| OCC / CSRF / limits | `ETag`/`If-Match`/412, `SameSite` cookies, `MaxBytesReader` — all stdlib | 0 |

**Interactivity note:** htmx is the gentle default and its markup-centric model fits
the team. **Datastar** is a genuine alternative worth a spike *because* our headline
requirement is streaming freshness — it's SSE-first with built-in signals and no
out-of-band-swap gymnastics, at a similar size — but it's a steeper, handler-centric
mental model. Start htmx; revisit Datastar only if the live board becomes the whole
product.

**Tier 2 (hosted single-user, GitHub-backed store) — the deps that earn their keep,**
all confined to the GitHub `core.Store` implementation:

- `google/go-github` — trees/contents for reads, commits for writes.
- A **caching transport** via `github.WithHTTPClient(...)` — `bartventer/httpcache`
  (RFC 9111) for a PAT, or `bored-engineer/github-conditional-http-transport` for a
  GitHub App token. This is what turns "ETag caching" into 304s that cost zero rate
  limit. Sharp edges: ETags are **per-page**, and **token rotation invalidates them**.
- `golang.org/x/sync/singleflight` — coalesce the board projection so N concurrent
  loads collapse to one GitHub tree fetch. Pair with `gofri/go-github-ratelimit` for
  automatic primary/secondary backoff.

## Open questions

- **Intra-column ordering.** A real kanban wants drag-to-reorder *within* a lane to
  persist; status-as-directory has no field for it. Decide: add a fractional-index
  `rank` frontmatter field, or accept "sorted by priority then age, no manual order."
  This is a data-model change → coordinate with epic 24 if we want it.
- **Serve-owns-git for the hosted rung.** The sync research
  ([[2026-06-24-remote-planning-repos-backends-and-sync]]) lands on the server as a
  single privileged git client (writes serialized through `core.Service` + OCC,
  commits batched under a `tskflwctl-bot` identity with a `Co-authored-by` trailer).
  For single-user this is nearly free; confirm we want the hosted server to drive git
  rather than talk to the GitHub API directly.
- **BOARD.md baseline.** A generated committed board file solves GitHub
  discoverability for ~1% of the effort. The web UI is justified by what it can't do —
  live freshness, drag interactivity, mutation from a browser. Keep both honest; the
  board file may ship first as the cheap win.

## Proposed follow-up tasks

To be filed under this epic once the shape is agreed:

1. **`serve` skeleton** — stdlib router, DI over `core.Service`, `ErrConflict`→412
   and sentinel→status mapping, lifecycle (signals/timeouts/shutdown), `slog`,
   `httptest` coverage. Read-only.
2. **Board read endpoint** — `GET /board` reusing `BoardEnvelope`; `html/template` +
   htmx render.
3. **Live board (SSE)** — broker + watcher subscription + browser patch. *This is the
   headline slice.*
4. **Write path + OCC surface** — `POST /tasks/{slug}/move` (and field edits) through
   `core.Service`, `If-Match`/412, optimistic UI. Depends on / stopgaps
   [[version-aware-occ-content-hash-token-and-plain-retry]].
5. **CSRF + body limits + config** — the browser trust-boundary hardening.
6. *(tier 2, later)* **GitHub-backed `core.Store`** — go-github + caching transport +
   singleflight + ratelimit, in its own package.

## Research / sources (2026-07-03)

Backing for the stack choices above:

- Router (stdlib-first): Alex Edwards, *Which Go Router Should I Use?*; Ben Hoyt,
  *Different approaches to HTTP routing in Go* (updated for 1.22 ServeMux).
- templ + htmx / GoTH stack: templ.guide htmx docs; Three Dots Labs, *Live website
  updates with Go, SSE, and htmx*.
- Datastar vs htmx: htmx.org/essays/alternatives; data-star.dev.
- SSE broker pattern: tmaxmax/go-sse; oneuptime *Real-time Applications with Go and
  SSE*.
- Server lifecycle: VictoriaMetrics *Graceful Shutdown in Go*; oneuptime
  *Production-Ready HTTP Server in Go*.
- ETag/If-Match OCC: oneuptime *API ETag Headers*; fideloper *ETags and Optimistic
  Concurrency Control*.
- GitHub store: google/go-github README; Lunar *Managing Rate Limits for the GitHub
  API*; gofri/go-github-ratelimit; golang.org/x/sync/singleflight.

## Out of scope

- **Multi-user / multi-tenant** — single-user throughout; no org/permission model.
- **OCC/version-CAS mechanism itself** — owned by epic
  [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]; this epic
  only adds the HTTP `ETag`/`If-Match` surface over it.
- **WASM client** — kept as a noted tier, not pursued.
- **The module split** — stay one module with `internal/` seams until a real external
  consumer exists; the `core.Store`/`core.Service` ports make a later split mechanical.
- **PR/approval-queue write policy** — a hosted-rung concern already explored in the
  sync research; not part of the local server.

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]] — OCC /
  version-CAS this epic's write path consumes.
- [[version-aware-occ-content-hash-token-and-plain-retry]] — the OCC task itself.
- [[2026-06-24-remote-planning-repos-backends-and-sync]] — serve-owns-git, freshness,
  single-writer sync (Path C convergence).
- [[2026-06-24-task-storage-model-files-logs-or-versioned-db]] — content-vs-workflow-
  state split behind the OCC shape.
- `docs/ARCHITECTURE.md` — ports-and-adapters layout `serve` becomes a third adapter of.
- `internal/core/store.go` — the `Store` port a GitHub adapter would implement.
