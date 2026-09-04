---
schema: 1
id: 6g63jj1dh0sb
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: store/config/userconfig each define writeFileAtomic with different durability and symlink guarantees; unify behind a foundations package
effort: 4-6 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [store, config, architecture, robustness]
created: "2026-09-02"
updated_at: "2026-09-04"
audit_sources: [planning/audits/6g6qvrj15x97-2026-09-04-concurrency-and-atomicity.md]
---
# Unify the three divergent writeFileAtomic implementations

## Objective

Three packages define an unexported `writeFileAtomic` with the same name and
**different guarantees**:

| | temp + rename | file fsync | directory fsync | symlink-preserving |
| --- | --- | --- | --- | --- |
| `store/atomic.go` | yes | yes | **yes** (`syncDir`) | **no** (`os.Stat`) |
| `config/config.go` | yes | yes | **no** | **yes** (`Lstat` + `EvalSymlinks`) |
| `userconfig/paths.go` | yes | yes | **no** | **yes** |

`config/atomicWriteDestination` and `userconfig/atomicWriteDestination` are
additionally byte-identical apart from error wrapping: `config` wraps with the
offending path, `userconfig` returns the bare error and loses it.

The divergence has a legitimate cause — `.golangci.yml` forbids `config` from
importing `store`, and `config` from importing `userconfig`, so the code was
inlined rather than shared, and `config.go:1145` says so explicitly. The fix is
therefore not to relax a dependency rule but to add a **Foundations** package
(the role occupied by `internal/id`, `internal/tomledit`, `internal/editor`,
`internal/listfilter`), which every layer may import.

### What the research changed

Anchored 2026-09-02. The first framing of this task assumed the strictest
implementation was correct and the others were under-syncing. That is only half
right, and the correction matters for the design:

- **Directory fsync IS required for durability.** fsync on a file does not make
  the directory entry durable; a crash after rename can lose the name even though
  the data is on disk. `store/atomic.go`'s `syncDir` is correct on this point and
  `config`/`userconfig` genuinely are not durable across power loss.
- **But atomicity and durability are different guarantees, and most callers want
  atomicity.** `google/renameio` — the reference Go implementation — deliberately
  omits the directory fsync, stating it "concerns itself *only* with atomicity,
  i.e. making sure applications never see unexpected file content". For a tool
  whose files are also tracked in git, losing the last write to a power cut is
  recoverable; showing a half-written task file is not.
- **On macOS the strict path is expensive.** Go's `File.Sync()` on darwin issues
  `fcntl(F_FULLFSYNC)` (falling back to `fsync` only on `ENOTSUP`), which asks the
  drive to flush its cache — a full barrier that degrades disk I/O for the whole
  machine, not just this process. `store` performs TWO of these per file (the temp
  file, then the directory), and its bulk paths write in a loop:
  `rename.go:119`, `graphmutation.go:100`, `fix.go:92`/`117`, `threadapply.go:134`.
  A `thread apply` over N tasks is therefore 2N full drive barriers on the
  maintainer's own platform.

So the job is **not** to unify on the strictest behavior. It is to name the
guarantee each call site actually needs, implement those as distinct, documented
options on one helper, and stop the three copies from drifting further.

## Acceptance criteria

- [ ] One atomic-file-write implementation lives in a Foundations-role package and is used by `store`, `config`, and `userconfig`
- [ ] The helper exposes atomicity and durability as *distinct, documented* guarantees rather than one blended behavior, so a caller opts into the directory fsync
- [ ] Each call site's required guarantee is decided and recorded, with a stated reason — in particular whether planning-tree writes need durability or only atomicity given the files are git-tracked
- [ ] Symlink-preserving destination resolution is available to every caller, including `store`, which lacks it today
- [ ] The bulk write paths do not regress: measure the F_FULLFSYNC cost of a multi-file `thread apply` / `lint --fix` on macOS before and after, and if per-file directory fsync is retained there, justify it against that measurement
- [ ] The inaccurate "Mirrors store/atomic.go's contract" comment at `userconfig/paths.go:70` is gone, and the unified doc comment states the guarantee precisely
- [ ] Error context is consistent: a failure names the path it was operating on
- [ ] `store`'s exclusive-create path (`createFileAtomic`, which carries the version-CAS `ifVersion == ""` precondition) keeps its current semantics exactly — it is a CAS primitive, not just a writer
- [ ] Tests cover writing through a symlink and the directory-fsync-after-rename behavior

## Out of scope

- Relaxing any `.golangci.yml` import rule — the Foundations role exists precisely so this does not require one
- `internal/testutil`'s plain `os.WriteFile` (test support) and `config.go:555`'s zero-byte placeholder write, where atomicity is meaningless
- The wider `config` / `userconfig` overlap beyond atomic writing
- Changing when any caller decides to write
- Adopting `google/renameio` as a dependency — it is cited as reference, and it does not cover the symlink-preserving destination behavior two callers rely on

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)

### Sources

- Directory fsync required after rename: <https://calvin.loncaric.us/articles/CreateFile.html>, <https://blog.httrack.com/blog/2013/11/15/everything-you-always-wanted-to-know-about-fsync/>
- Postgres treating rename-without-fsync as a data-loss risk: <https://www.postgresql.org/message-id/E1adrE0-0001Os-CD%40gemulon.postgresql.org>
- renameio scoping itself to atomicity, not durability: <https://pkg.go.dev/github.com/google/renameio/v2>
- Go's darwin `File.Sync` using `F_FULLFSYNC` with an `ENOTSUP` fallback: <https://github.com/golang/go/issues/26650>, <https://github.com/golang/go/blob/master/src/internal/poll/fd_fsync_darwin.go>
- `F_FULLFSYNC` cost on macOS: <https://mjtsai.com/blog/2022/02/17/apple-ssd-benchmarks-and-f_fullsync/>, <https://bonsaidb.io/blog/acid-on-apple/>

Reinforced by audit 2026-09-04-concurrency-and-atomicity: L1. That audit confirms the store/config divergence is currently latent rather than live — `markdownDoc` (internal/store/resolve.go:23) gates every entity scan on `e.Type().IsRegular()`, so a symlinked task/audit/epic/research/thread file is never listed or resolved and `store.writeFileAtomic` is never handed one. The unification is still worth doing: the safety depends entirely on that gate, which atomic.go never mentions, and config.go:1146 currently claims the store 'has the same idea' when it does not.
