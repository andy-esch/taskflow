---
status: in-progress
description: Port the prototype Python pm tool to a Go CLI in the decided noun-verb hierarchy; Python pm + its tests are the executable spec.
created: 2026-06-06
updated_at: 2026-06-08
tags: [pm-tooling, go, cli]
priority: medium
---

# Epic 17: PM Go CLI (`tskflwctl`)

## Progress (2026-06-08)

The Go CLI is **substantially functional** and dogfooding on taskflow's own
planning. Phase 0.5 (spec), the foundation, and a large slice of the port have
landed; built in `cmd/tskflwctl` + `internal/{domain,core,store,cli,config}`
with the primary/secondary-adapter architecture holding up cleanly.

**Working today** (`go test ./...` + `golangci-lint` green):
- `init`, `completion` (command/flag/**slug** completion, status-aware), `lint`
  (+ `--fix`/`--dry-run` auto-repair of pm-written frontmatter)
- `task new|list|show|set|move|start|promote|demote|complete|defer|deprecate`
- `epic new|list|show` (auto-numbered create; cross-task rollup)
- `audit list|show|close|reopen|defer` (finding-count rollup, bucket lifecycle)
- Cross-cutting: explicit noun-verb, semantic exit codes (10–13), atomic +
  surgical-`yaml.v3` writes, `--json` everywhere with `schema_version`,
  resilient reads with **actionable** frontmatter errors, agent safety tags.

With `task new`/`epic new` in, the **full daily loop (create→update→move→lint)
runs without Python `pm`** — the bare-bones-release bar.

**Remaining** (see the port task's progress log): `adr`/`project` groups, the
audit finding-level commands (`status`/`fixed`/`followup`/`sync`/`new`/`noop`/
`findings`/`stats`), reporting views (`stats`/`index`/`tags`), `track`,
`schema --type cli`, a global `--dry-run`, advisory `flock`, structured JSON
error envelope, interactive `init` wizard.

The Python `pm` (epic 16) remains the prototype/spec and still manages the
desirelines-planning repo; `tskflwctl` reads/writes the same markdown and the
two interoperate during the transition.

**Goal.** Port the prototype Python `pm` tool to a Go CLI — binary
**`tskflwctl`** (kubectl-style), built in the reused **taskflow** repo
(`github.com/andy-esch/taskflow`) — in the decided noun-verb hierarchy.
The Python `pm` + its tests are a starting-point spec (incomplete; see
Phase 0.5).

## Why this is its own epic

The Python `bin/pm` (epic 16) is explicitly a **prototype** — it exists to
settle the task/epic/audit conventions and the CLI command surface cheaply,
under fast iteration. The long-term tool is a **Go CLI**. That's a distinct,
long-lived effort (new language, framework choice, distribution, the
noun-verb hierarchy built once natively) and shouldn't be folded into the
pm-tooling prototype epic — 16 is "make the prototype good enough to spec
from," 17 is "build the real thing."

## Current state

- Prototype lives at `bin/pm` (single ~2.4k-line stdlib Python script) with
  `tests/test_pm.py`. Commands: task lifecycle, `set`/`lint`/`schema`,
  `epic`/`adr`/`project` groups, soon `audit` (epic 16).
- **Runtime caveat:** the live runtime has no PyYAML → the regex fallback
  parser is the real path; tests run *with* PyYAML. The Go port removes this
  split (one parser, real YAML).
- Command structure today is a flat verb namespace + a few noun-groups —
  the redesign target is decided in the research task, not here.

## Decided design (2026-06-06)

Settled across the three research docs (`research/2026-06-06-{pm-cli-architecture-and-go-port,go-cli-foundation-architecture,tskflwctl-command-spec}.md`):

- **Binary `tskflwctl`** (kubectl-style); reuse the **taskflow** repo
  (module `github.com/andy-esch/taskflow`), entrypoint `cmd/tskflwctl/`.
- **Stack:** cobra + viper; **Bubble Tea** for the later TUI; AST frontmatter
  (`goccy/go-yaml` + `go.abhg.dev/frontmatter`) for lossless round-trips.
- **Architecture:** CLI (and later TUI) are **primary adapters** over a
  shared **core**; the markdown filesystem is a **secondary adapter**.
  Layout `cmd/ internal/{domain,core,store,cli,tui,config}/`; DI, no globals,
  `io.Writer` output, real YAML.
- **CLI shape:** fully explicit `tskflwctl <noun> <verb>` — **no flat
  aliases, no lifecycle sugar**; lifecycle = one `task move <t> <status>`.
- **Conventions:** non-zero exit on error; `--json` global with a semver
  `schema_version`; commands tagged read-only/mutating (cobra Annotations).
- **One planning repo per product** (no cross-product registry), tracking
  1..N code repos (`tracked_repos`); `init` scaffolds; anchor by walking up
  for `.tskflwctl.toml`.
- **Projects** = optional, cross-cutting initiatives (own `projects/` dir +
  `projects:` list on tasks), orthogonal to domain epics.
- **Out (long shot):** MCP / RAG / semantic engine / pgvector / AI gen.

## Build sequence

1. **Phase 0.5 — command spec**
   [[phase-0.5-formal-tskflwctl-command-hierarchy-purpose-spec]]: finalize
   the build-to spec. *(Don't rely on the Python tests alone — incomplete.)*
2. **Foundation**
   [[go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness]]:
   layout, core/store/cli boundary, DI, golden-file + lint harness; prove
   one vertical slice (`task list`).
3. **Port**
   [[port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec]]:
   `init` + core commands (task/epic/adr/project) + the `audit` group, to spec.
4. **TUI preview** (Bubble Tea) — later phase, over the same core.

Gate: start *after* (a) Phase 0.5 lands and (b) the Python `pm audit`
prototype has soaked a few weeks of real audits.

## Prior art & scope fence

A previous Go spike of this tool — **taskflow** (`../taskflow`) — exists:
cobra + viper + Bubble Tea, Pattern-C layout, a JSON fast-index, and
substantial `planning/research/`. Its Go is thin scaffolding (it stalled by
leading with an intelligence layer). **Bootstrap epic 17 from its skeleton**
rather than greenfield, stripped to the fence below.

**Scope fence (user, 2026-06-06):**
- **NOW:** self-contained Go cobra CLI at parity with Python `pm`, reading
  markdown directly. No services/DB/network.
- **EVENTUALLY:** a **TUI preview** (Bubble Tea — split-pane list+preview,
  vim keys). Plan for it; don't build it first.
- **OUT (long shot):** MCP, RAG/semantic search, pgvector, Python "brain",
  LiteLLM/AI generation, doc auto-gen. taskflow's research on these is
  captured for later; build none of it now.

## Out of scope / Non-goals

- Changing the task/epic/audit **file formats** — the Go CLI reads the same
  markdown+frontmatter the Python tool does.
- Re-litigating the command hierarchy — that's the research task's output;
  this epic implements it.
- The whole intelligence layer (see fence) — deferred indefinitely.
- Keeping the Python tool alive long-term — it's retired once the Go CLI
  reaches parity.

## References

- [[rethink-pm-command-hierarchy-pm-noun-verb-research-cli-best-practices]]
  — the design input (hierarchy + Go framework + port strategy).
- `research/2026-06-06-pm-cli-architecture-and-go-port.md` — research doc
  (incorporates the taskflow spike's findings under the scope fence).
- `research/2026-06-06-go-cli-foundation-architecture.md` — the code
  foundation (layout, primary/secondary-adapter split, DI, test harness).
  **Build [[go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness]]
  first**, then port commands onto it.
- `research/2026-06-06-tskflwctl-command-spec.md` — the Phase 0.5 formal
  command spec ([[phase-0.5-formal-tskflwctl-command-hierarchy-purpose-spec]]).
- `../taskflow/` — the prior Go spike to bootstrap from (cobra/viper/Bubble
  Tea skeleton, Pattern-C layout, `planning/research/`).
- `research/sqlite-vs-markdown-for-pm-system.md` — prior pm-architecture
  thinking (storage layer).
- Epic 16 (pm-tooling) — the prototype this ports from.
