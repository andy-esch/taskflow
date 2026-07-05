---
status: completed
epic: 17-pm-go-cli
description: "Idiomatic-Go foundation before command porting: Pattern-C layout, CLI/TUI over a shared core, DI/no-globals, golden-file + lint harness."
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [pm-tooling, go, cli, architecture]
created: 2026-06-06
updated_at: 2026-06-07
started_at: 2026-06-07
completed_at: 2026-06-07
id: 6f9menr02b25
---

# Go CLI foundation: layout, core/store/cli boundary, DI, test+lint harness

## Objective

Stand up a **rock-solid, idiomatic-Go foundation** for the pm CLI in the
reused `../taskflow` repo, *before* porting any commands. Reuse taskflow's
good parts (stack, Pattern-C bones) but get the architecture right — the
thing taskflow skipped. Design spec:
`research/2026-06-06-go-cli-foundation-architecture.md`.

## Design (from the research doc)

Align with the org's pragmatic hexagonal (see
`research/hexagonal-architecture-go-best-practices.md` + the
`../desirelines` Go services). **CLI and the later TUI are primary adapters
over one shared core; the markdown filesystem is a secondary adapter.**

```
cmd/tskflwctl/main.go        # thin composition root, single os.Exit
internal/domain/      # pure entities + invariants (typed Status, "status==dir")
internal/core/        # use cases (Service) + store interfaces (near consumer)
internal/store/       # fsstore: markdown+frontmatter (yaml.v3), round-trip
internal/cli/         # cobra tree, NewRootCmd(deps) — NO globals; render/ (human|json)
internal/config/      # viper, split-repo paths
testdata/             # golden files + fixture trees
```

## Scope (foundation only — no command parity yet)

- [ ] Module/layout: reuse taskflow repo; create the domain/core/store/cli
      package skeleton; drop `services/`, `contracts/`, the `ui` brain
      wiring (see asset table in the research doc). Resolve module rename
      (Q1) and `domain`-vs-`core` split (Q2).
- [ ] Core boundary: `core.Service` with a couple of real use cases (e.g.
      `ListTasks`, `Show`) calling a `core.TaskStore` interface; `fsstore`
      implements it with `var _ core.TaskStore = (*FS)(nil)`.
- [ ] DI/no-globals: `cli.NewRootCmd(svc, cfg, out, errOut)`; `main.go` is
      the only wiring site; all output via injected `io.Writer`.
- [ ] Errors/exit: `RunE` returns wrapped errors + domain sentinels;
      `main` maps to conventional non-zero exit codes.
- [ ] Test harness: domain unit tests, store tests against `t.TempDir()`,
      and an in-process golden-file command test (`-update`).
- [ ] Build/lint: `.golangci.yml` mirroring the impl repo (gosec/wrapcheck/
      nolintlint); Justfile `build`/`test`/`lint`; version ldflags.

## Acceptance criteria

- [x] One vertical slice works end-to-end on the new architecture:
      `tskflwctl task list` (explicit noun-verb; no flat alias) reads via `fsstore` →
      `core.Service` → renders human + `--json` (+ `--status/--epic/--tag`),
      with **no cobra/global state** and **no fs access in core**.
- [x] `go test ./...` green incl. in-process CLI tests; `golangci-lint run` clean.
- [x] `cmd/tskflwctl/main.go` is the sole composition root; commands take injected
      deps (`*cli.App`, lazy-resolved in `PersistentPreRunE`); output via `io.Writer`.
- [x] Architecture note documents the layout + primary/secondary-adapter rule
      (`docs/ARCHITECTURE.md`).

## Progress Log

**2026-06-07 — foundation vertical slice landed (in taskflow).** Scaffolded the
DI architecture and proved it end-to-end on taskflow's own planning data.
- Packages: `cmd/tskflwctl`, `internal/{domain,core,store,cli,cli/render,config}`.
  Removed the spike's global-`rootCmd` entry (`cmd/taskflow`, old `internal/cli`).
  `var _ core.TaskStore = (*store.FS)(nil)`; core is fs/cobra-free.
- DI: `cli.NewRootCmd(out, errOut)` → empty `*App` populated in
  `PersistentPreRunE` (lazy shell; deps depend on `--chdir`). No globals, no
  context-as-DI.
- Frontmatter: zero-dep byte-scanner split + `go.yaml.in/yaml/v3` parse
  (dropped goldmark/`go.abhg.dev`). Surgical *writes* deferred to the port.
- Tests: domain via store, store round-trips on `t.TempDir()`, in-process CLI
  tests (json shape + filter + not-a-repo error). `go test ./...` green;
  `golangci-lint run` 0 issues (standard set + gofmt/goimports — full
  gosec/wrapcheck parity deferred, noted in `.golangci.yml`).
- **Real-data finding:** the relocated foundation task's `description` had an
  unquoted `:` → invalid YAML the no-PyYAML Python `pm` wrote and never
  validated. tskflwctl's strict YAML correctly rejected it. Fixed the file;
  logged [strict-yaml-frontmatter-fix-pm-serialize-quoting-backfill-invalid-values](6f9yr8m03xc4-strict-yaml-frontmatter-fix-pm-serialize-quoting-backfill-invalid-values.md).

**Deferred to the port task (not this foundation):** transition verbs + `move`,
`set` with flags, `audit`/`epic`/`project` groups, `init`, atomic writes +
surgical frontmatter, semantic exit codes, default-list-to-active, full
gosec/wrapcheck lint parity.

## Out of scope

- Porting the full command set (that's
  [port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec](6f9menr01nsd-port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md)).
- The TUI (`internal/tui` is seeded but not built).
- Anything in the intelligence layer (MCP/RAG/brain — out by a long shot).

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md).
- `research/2026-06-06-go-cli-foundation-architecture.md` — the design.
- [port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec](6f9menr01nsd-port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md) —
  the command port that builds on this foundation.
