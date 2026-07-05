---
date: 2026-06-06
topic: pm CLI architecture — command hierarchy, Go framework, Python→Go port
purpose: Decide the noun-verb hierarchy + Go CLI framework + port strategy before building the Go pm (epic 17). Design input for the port.
status: in-progress
related_tasks:
  - rethink-pm-command-hierarchy-pm-noun-verb-research-cli-best-practices.md
  - port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md
  - bucket-audits-into-openclosed-and-ship-pm-audit-cli.md
---

# pm CLI architecture: command hierarchy, Go framework, and port strategy

**Goal.** The Python `bin/pm` is a prototype; the real tool is a Go CLI
(epic [17-pm-go-cli](../epics/17-pm-go-cli.md)). This doc decides the command hierarchy, the Go
framework, the data model, and the port strategy. It is design input —
implementation lives in epic 17.

> Status: **in-progress** (2026-06-06). Updated after reviewing the prior
> **taskflow** spike (`../taskflow`) — a previous Go attempt at this tool.
>
> **Naming (decided):** the Go binary is **`tskflwctl`** (kubectl-style),
> built in the reused taskflow repo. Below, **`pm`** = the Python prototype
> it ports from; **`tskflwctl`** = the Go tool being designed.

## 0. Scope fence (set by the user, 2026-06-06)

The taskflow spike over-reached (semantic engine, pgvector, MCP, LiteLLM)
and stalled at scaffolding. Hard fence for this effort:

- **In scope NOW:** a self-contained Go cobra CLI at parity with the
  Python `pm`, reading the markdown files directly. No services, no DB,
  no network.
- **In scope EVENTUALLY:** a **TUI preview experience** (Bubble Tea) —
  split-pane list + preview, vim keys. Plan for it; don't build it first.
- **OUT of scope by a long shot:** MCP server, RAG/semantic search,
  pgvector, the Python "brain" service, LiteLLM/AI generation, embeddings,
  doc auto-generation. Capture taskflow's research on these for later;
  build none of it now.

**The lesson from taskflow:** it led with the intelligence layer and never
shipped a solid CLI. Invert that — ship the boring fast CLI first.

## 1. Prior art: the `../taskflow` spike

A prior Go reimagining of `pm` ("taskflow") exists at `../taskflow`. The
**Go code is thin scaffolding** (cobra root + a `ui` command + a stub
Bubble Tea model + a JSON-index struct), but the **planning/research is
substantial** and several decisions are directly reusable.

**Stack it chose** (`go.mod`): `spf13/cobra` 1.10, `spf13/viper` 1.21,
`charmbracelet/bubbletea` 1.3 + `lipgloss`, `pelletier/go-toml`,
`go.yaml.in/yaml/v3`, `protobuf`/buf. Layout = "Pattern C" (Standard Go
`cmd/`+`internal/`, `contracts/` proto, `services/` for the Python brain).

**Decisions worth keeping (within our fence):**
- **Go + cobra + Bubble Tea** for CLI + TUI (`planning/research/cli-language-options.md`):
  single static binary, sub-10ms start, gh/kubectl-class ergonomics.
  Validates our framework direction — it's settled prior art, not a fresh
  bet.
- **JSON fast-index** (`hybrid-search-architecture.md`): read one
  `planning-index.json` for `list`/`filter` instead of scanning N files.
  **Secondary / Optional Path:** Since we are dealing with local data,
  Go's concurrency should be fast enough to avoid caching for now. Keep the
  index as an optional performance path, and only prioritize it once we
  move to remote sources (e.g., GitHub Issues) where network latency
  makes caching mandatory.
- **Pattern C / no Go workspaces** (`monorepo-structure-proposal.md`,
  `workspaces-vs-root-modules.md`): one root `go.mod`, no `go.work` for a
  single binary.
- **Config + onboarding** (`configuration-and-onboarding-flow.md`): borrow
  the `init`-wizard idea. A **planning repo** holds all planning data
  (single-repo) and records `tracked_repos` = the 1..N code repos it plans
  *for* (metadata/anchoring; e.g. desirelines-planning → desirelines +
  desirelines-deploy). `tskflwctl init` scaffolds the tree + config. One
  planning repo per product — **no cross-product registry** (decided).

**Decisions to DEFER (out of fence, captured for later):** the thin-client
+ Python brain split, pgvector hybrid search, MCP server vs context-broker
(`ai-interaction-interfaces.md`), LiteLLM provider abstraction
(`ai-provider-abstraction.md`), living-docs generation
(`documentation-integration-strategy.md`), AI block-diff file editing
(`ai-file-editing-strategy.md`). Good thinking — none of it now.

**Bootstrap, don't greenfield:** epic 17 should start from taskflow's
cobra skeleton + Pattern C layout rather than from scratch — but strip it
back to the CLI-only fence (drop `services/`, the `ui` brain wiring, the
proto/buf toolchain unless we adopt proto per §3).

## 2. Command hierarchy — `pm <noun> <verb>`

Current `pm` is a flat verb namespace (`list`, `show`, `set`, `lint`,
`start`, `promote`, …) *plus* nested `epic`/`adr`/`project` groups. Adding
`audit` makes the inconsistency stark. Target (gh/kubectl model):

```
tskflwctl task   list|show|new|set|lint|move|touch|rename
tskflwctl audit  list|show|new|close|reopen|status|fixed|landed|followup|defer|sync|lint|noop|findings|stats
tskflwctl epic   list|show|new|lint|orphans
tskflwctl adr    list|new
tskflwctl schema|index|recommend|stats   # cross-cutting → stay top-level
```

- **Fully explicit, no sugar** (decided 2026-06-06): no flat top-level
  aliases (`tskflwctl list` is not provided — use `tskflwctl task list`),
  and lifecycle is one explicit `task move <t> <status>` (no
  `start`/`promote`/`complete`/…). `tskflwctl` is a new binary, so there's
  no muscle memory to preserve; explicit over convenient. *(Full per-command
  purpose = Phase 0.5 spec.)*
- `taskflow`'s root was just `taskflow <verb>` + a `ui` command — it never
  got to noun groups, so this hierarchy is our call to make, greenfield.

## 3. Data model — proto vs plain Go structs (decision needed)

taskflow defined `Task` in **protobuf** (`contracts/`, buf codegen) so Go
and Python could share one schema. **With the Python brain out of scope,
that cross-language justification disappears.** So:

- **Option A — plain Go structs** (recommended for the fence): a `Task`
  struct with yaml/json tags, parsed via `yaml.v3`. Simplest; no buf
  toolchain; YAGNI on codegen for a single-language CLI.
- **Option B — proto schema now**: future-proof (a second consumer later),
  and desirelines already uses buf/proto heavily (org familiarity), and
  taskflow already wrote the `.proto`. Cost: buf/codegen toolchain weight
  for a one-language tool today.

**Lean: Option A now, revisit proto if/when a second consumer (TUI sharing,
or an eventual service) appears.** Either way the **file formats don't
change** — same markdown + frontmatter the Python tool reads.

## 4. Go framework — cobra + viper (settled by prior art); Bubble Tea for the eventual TUI

- **cobra + viper** — chosen by taskflow, matches the gh/kubectl model,
  free completion + help, viper for the `.pm.toml`/split-repo config.
  Treat as **decided** (prior art + ecosystem); don't re-litigate with
  kong.
- **Bubble Tea + Lip Gloss** for the eventual **TUI preview** (in scope
  *eventually*): split-pane list+preview, vim `j/k`, `y/n`. taskflow has a
  stub `ui` command + model to grow from. Build the plain CLI first; add
  `pm ui` as a later phase.
- **fang** (cobra wrapper, styled help/errors) — optional polish; gate on
  "does it leak ANSI into `--json`/pipes?"

## 5. Cross-cutting conventions

- **Exit codes — fix the prototype quirk.** Python `pm` uses
  exit-0-with-error (`test_set_validates_tier` asserts rc 0). No
  back-compat reason to keep it — adopt conventional **non-zero on error**
  in Go (CI/agents branch on it). Intentional behavior change.
- **`--json` everywhere as a global persistent flag**; stable schemas
  matching the Python `--json` shapes; JSON error object + non-zero exit in
  JSON mode; no ANSI on stdout (important once Lip Gloss/fang are in).
- **No back-compat aliases** (decided 2026-06-06): `tskflwctl` is a fresh
  binary with no existing callers, so everything is explicit
  `tskflwctl <noun> <verb>` — no flat-verb shims to maintain.
- **One parser, real YAML** (`yaml.v3`): removes the prototype's biggest
  gotcha — the live Python runtime has **no PyYAML**, so it silently runs
  the regex fallback parser (see [[pm-runtime-no-pyyaml-fallback-parser]]).

## 6. Port strategy

> **Code architecture / layout / patterns live in the companion doc:**
> `research/2026-06-06-go-cli-foundation-architecture.md` (primary/secondary
> adapter split over a shared core, DI/no-globals, testing harness). Build
> the foundation first ([go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness](../tasks/6f9menr02b25-go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness.md)),
> then port commands onto it.

- **Bootstrap from taskflow** (cobra skeleton + Pattern C), stripped to the
  CLI-only fence.
- **Tests are a starting point, not the whole spec.** `tests/test_pm.py`
  encodes the contract, but acknowledged to be incomplete. Phase 0.5 
  (hierarchy tree) is the primary spec for the port. Port group-by-group, 
  translating logic into Go table tests (red→green). Markdown fixtures 
  port verbatim.
- **Phase it** (avoid taskflow's fate):
  1. `init` + core CLI parity (task + epic + adr + project + schema/lint/set),
     markdown direct, single-repo.
  2. `tskflwctl audit` group (after the Python `pm audit` prototype soaks).
  3. Optional JSON fast-index for speed.
  4. **TUI preview** (`tskflwctl ui`, Bubble Tea).
  5. *(out of fence, later)* remote/hybrid sources, then maybe the
     intelligence layer.
- **Distribution:** `just build` → `bin/tskflwctl` Go binary. Long-term,
  distribute via **GitHub Releases** (e.g., `go install` or a curl-to-bin
  installer); include `cobra` shell completion generation in the build
  process. Update AI_README/README/routers.
  *(External reviewer comment: Using GoReleaser can automate Homebrew Tap updates and GitHub Releases simultaneously, providing the most professional distribution experience for macOS users.)*

## 7. Open questions

1. **Data model:** plain structs (lean) vs proto (§3) — confirm.
2. ✅ **Repo home (resolved 2026-06-06):** reuse the **taskflow** repo;
   binary **`tskflwctl`** at `cmd/tskflwctl/`. Epic 17 *is* taskflow
   revived under the scope fence (bootstrap from its skeleton).
3. ✅ **Config (resolved 2026-06-06):** one **planning repo** per product
   holds all data + records `tracked_repos`; `tskflwctl init` writes
   `.tskflwctl.toml`, discovered by walking up (git-style). **No**
   cross-product registry, **no** `source` noun, **no** per-task `repo:`
   field (all decided). Only the config filename is cosmetic.
4. **Concurrency:** While Go handles parallel scans easily, the current
   task count (N < 100) means a simple sequential scan is likely
   sufficient for v1. Design for concurrency, but don't let it block
   the MVP.
5. **Exit-code break:** confirm no scripted caller depends on exit-0.
6. fang vs `--json` cleanliness.

## Recommendation summary

- **Fence:** CLI now · TUI (Bubble Tea) eventually · MCP/RAG/brain out.
- **Hierarchy:** `pm <noun> <verb>`, task verbs also top-level as aliases.
- **Framework:** cobra + viper (settled by the taskflow spike); Bubble Tea
  for the later TUI; fang optional.
- **Data model:** plain Go structs + `yaml.v3` now; proto only if a second
  consumer appears. File formats unchanged.
- **Port:** bootstrap from taskflow's skeleton, strip to CLI-only,
  test-suite-as-spec, phase it, fix the exit-code quirk; sequence after
  `pm audit` soaks.

## Sources

- `../taskflow/` — prior Go spike: code (`cmd/`, `internal/cli`,
  `internal/tui`, `internal/index`) + `planning/research/*.md`
  (cli-language-options, hybrid-search-architecture,
  monorepo-structure-proposal, workspaces-vs-root-modules,
  ai-interaction-interfaces, ai-provider-abstraction,
  configuration-and-onboarding-flow, documentation-integration-strategy,
  ai-file-editing-strategy) + `planning/epics/00-taskflow-v1-core.md`.
- [go-cli-comparison](https://github.com/gschauer/go-cli-comparison) ·
  [Choosing a Go CLI Library](https://mt165.co.uk/blog/golang-cli-library/) ·
  [charmbracelet/fang](https://github.com/charmbracelet/fang) ·
  [Cobra guide](https://www.bytesizego.com/blog/cobra-cli-golang)

  [charmbracelet/fang](https://github.com/charmbracelet/fang) ·
  [Cobra guide](https://www.bytesizego.com/blog/cobra-cli-golang)
