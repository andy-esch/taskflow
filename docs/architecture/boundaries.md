# Boundaries and composition

[Architecture index](../ARCHITECTURE.md) · [Planning data](planning-data.md)

This guide owns cross-package direction and the named local exceptions. Type comments
own API details; [.golangci.yml](../../.golangci.yml) owns enforceable import policy.

## Package map

| Role | Packages | Responsibility |
| --- | --- | --- |
| Foundations | `id`, `tomledit`, `editor`, `listfilter` | Focused primitives, not use cases or persistence composition |
| Domain | `domain` | Entities, closed vocabularies and pure validation |
| Application | `core` | Use cases, immutable projections, consumer-owned ports and recovery policy |
| Neutral contracts | `wire` | Machine DTOs and compatibility; depends inward on core/domain |
| Secondary adapters | `store`, `config`, `userconfig`, `spacehealth`, `configstore`, `spacestore`, `workspacestore` | Local decoding, discovery and persistence |
| Primary adapters | `cli`, `tui`, `configui` | User interaction over application services, not persistence construction |
| Invocation contracts | `cli/ports` | Framework-neutral composition and launch requests |
| Binary composition | `appwiring`, selected by `cmd/tskflwctl` | Select concrete adapters and UI launchers |
| Presentation | `theme`, `design`, `progressbar`, `themepreview`, `graphfmt`, `cli/render`, `cli/prompt` | Semantic tokens, formatting and interaction, not planning persistence |
| Test/tool support | `testutil`, `tools/*` | Fixture or one-off executable dependencies, not runtime layers |

Names above are under `internal/` unless explicitly prefixed with `cmd/`.
The descriptive map below intentionally summarizes presentation edges; it is not
the lint allowlist or a complete machine-generated import inventory.

```text
domain      -> id
core        -> domain, id
wire        -> core, domain
store       -> core, domain, id
config      -> domain, id, tomledit
userconfig  -> tomledit
spacehealth -> config, domain, userconfig
configstore -> core, config, userconfig
spacestore  -> config, core, domain, spacehealth, store, userconfig
workspacestore -> config, core, domain, store
configui    -> core + presentation utilities
tui         -> core, domain, configui + presentation/process utilities
cli/render  -> core, domain, wire + presentation utilities
cli/ports   -> core, design
cli         -> cli/ports, core, domain, wire + presentation/process utilities
               config only in init.go and workspace.go (named local exceptions)
graphfmt    -> core
appwiring   -> cli/ports, core, design, config, configstore, spacestore,
               userconfig, workspacestore, tui, configui
```

The [import-map task](../../planning/tasks/6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md)
will verify this descriptive map after choosing its parser/representation. Until
then it remains a manual projection; do not replace its policy distinctions with
a date saying someone reviewed it.

## Enforced fitness rules

Production depguard rules cover direct and recursive package files:

- Domain permits only `id` from this repository; core permits only domain/id.
  Ports stay with their application consumer.
- Wire permits only core/domain from this repository.
- Primary adapters and CLI presentation packages deny internal imports by default,
  with exact application/presentation allowances. An allowed package's future child
  is not automatically permitted.
- CLI controllers cannot construct persistence or import Bubble Tea. Only
  `init.go` and `workspace.go` may import config for the exceptions below;
  neither is exempt from the remaining controller rule.
- `cli/ports` permits only the standard library, core and design, preventing
  external framework types in invocation contracts.
- Local startup readers are selected in `appwiring/local_sources.go`; other
  production composition files cannot directly import config/userconfig.
- Config cannot import userconfig: home preferences and registry must not influence
  repo-scoped planning discovery.
- Only the full TUI embeds configui; the focused configuration UI and CLI
  render/prompt packages cannot launch it.

Tests may compose real adapters under `t.TempDir()`; their exemption does not
weaken production policy. These internal-import checks do not enforce every
standard-library/external-framework ban in pure packages. That stronger guard is
[tracked separately](../../planning/tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md).

## Composition and local exceptions

[Invocation ports](../../internal/cli/ports/runtime.go) are supplied by the
[binary-selected wiring](../../internal/appwiring/wiring.go). Each command tree gets
one `cli.App`, fresh repo-independent services and a live mutation authorizer.
No package-global dependency defaults, `cmd.Context()` service locator or fallback
local store exists in controllers.

Planning opening is lazy, after target flags are parsed. Completion defers it
until Cobra has parsed the command being completed. A failed composition publishes
no partial operational bundle; help, version, schema and script generation remain
usable with metadata-only bindings. Entity completion fails quietly rather than
falling back to local discovery. UI startup requires configuration, registry,
workspace and overview services on either landing route.

Ordinary and workspace opening share
[NewPlanningStore](../../internal/workspacestore/planning.go), constructed from the
already-observed configuration without a second startup scan. It retains the initial
corpus and watcher layout, checked source set, invocation policy, and fresh identity
reader. Apply re-discovers from the original marker directory, including pointer
checkouts—not cwd or the resolved target. Root/identity changes fail closed.
[Opening parity tests](../../internal/appwiring/thread_apply_test.go) and
[the constructor fitness test](../../internal/workspacestore/planning_architecture_test.go)
exercise these distinct guarantees.

Every persistence constructor requires an explicit
[mutation policy](../../internal/core/mutation_policy.go). Guarded policy checks
each mutation, including previews, before callbacks/effects; read-only refuses all
writes and previews. Workspace opening propagates it, and overview stores are
read-only. Unrestricted is a deliberate trusted-caller choice, not omission, and
does not bypass locks, CAS or domain rules. Zero-value adapters fail closed.
This is accidental-composition protection, not a hostile-code security boundary.
See [constructor/policy tests](../../internal/appwiring/persistence_policy_test.go).

| Edge | Why it remains |
| --- | --- |
| `appwiring -> configstore/spacestore/workspacestore` | Binary-selected concrete composition; all persistence families receive policy |
| `cli -> core.Service` maintenance/completion | Application orchestration through source-set-checked optional ports |
| CLI/TUI `Layout` watcher use | Explicit local process capability, not semantic entity data |
| `cli/init.go -> config` | Bootstrap topology before a planning service exists; explicit mutation guard |
| `cli/workspace.go -> config` | Checkout provenance via DescribeCheckout/ConfigFile, not planning reads |
| `appwiring -> config/userconfig` | Startup discovery/presentation readers; registry/config use cases remain application-owned |
| `appwiring -> tui/configui` | Framework launch requests selected by the binary |
| `tui -> configui` | Full TUI embeds the same focused configuration editor |
| `tui -> editor/os/exec` | Narrow terminal/process actions; planning data still goes through services |

These are named exceptions, not general licenses for adapter imports.
The [policy ADR task](../../planning/tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md)
still requires acceptance; this extraction does not enact a new decision.

## Why these boundaries

CLI and TUI both ship over the same core, so the application seam has two real
primary consumers. Portable test adapters also exercise use cases without the
filesystem. Sibling-visible exported types in `internal/` are layer contracts,
not automatically public-library leakage.

Keep semantic validation in domain and YAML parsing/surgical encoding in store.
Keep local watching separate from portable reads. `cli/render` is a replaceable
CLI presentation seam; its use of core result values does not justify deleting the
shared core. Collapse or extract packages only when demonstrated friction warrants
it, not because an invented future HTTP adapter might want them.

Typed entities intentionally have different shapes. The
[descriptor registry](../../internal/domain/entity.go) centralizes authoring metadata,
scaffolds and conventions; generic scanners/columns and registry-driven TUI lifecycle
remove mechanics, not semantic variance. Further persistence/render framework work
needs evidence. The [zero-variance verb task](../../planning/tasks/6g63hjmb3d8s-collapse-the-zero-variance-cross-entity-cli-verbs-behind-the-existing-factory.md)
proposes bounded mechanical reuse, not rejection of typed entities; that factory
work is still queued, not delivered by this extraction.

Earlier review provenance remains in
[planning](../../planning/tasks/6fakbec02ctt-triage-architecture-collapse-review-deferred-relitigates-settled-multi-adapter.md).
