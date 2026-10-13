# taskflow contributor and agent guide

`tskflwctl` is a local-first Go planning CLI/TUI over Markdown and frontmatter.
The project tracks its own work in `planning/`. This is the shared working guide
for humans and coding agents, not a tool-specific policy copy.

## Working rules

- Preserve user-owned and unrelated changes. Stop and ask when a requirement, policy,
  or design choice is ambiguous; don't silently expand scope.
- Git inspection is fine. State-changing Git operations (including branching,
  committing, and pushing) require a user request. The planning CLI writes files,
  not Git history.
- CLI/TUI are primary adapters over core use cases and consumer-owned ports;
  persistence belongs in secondary adapters, selected at the binary composition
  boundary. Core does not perform filesystem or Cobra work. Follow the architecture
  guide's named exceptions, not a blanket exemption for local code.
- TUI I/O runs through application services in `tea.Cmd`, never in `Update`/`View`.
  Keep one `cli.App` per command tree, explicit dependencies, lazy planning opening
  after target flags are parsed, and output through injected writers.
- Use adapter-established stable record identities; an untrusted declaration or opaque
  `Location` is not identity or permission to open a path. Preserve coherent snapshots,
  explicit capability/policy binding, guarded persistence, and committed/partial receipts.
  Use the [planning-data checklist](docs/architecture/planning-data.md#change-checklist)
  for the detailed contracts and required regressions.
- Use existing guarded/atomic persistence helpers; preserve unknown frontmatter,
  comments, and key order. Wrap domain error sentinels rather than inventing
  adapter-specific error policy. Preserve versioned machine output through the neutral
  wire contract; observable changes follow
  [ADR-0008](planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md).
- Use the freshly built CLI for managed planning fields, lifecycle, dependencies,
  acceptance criteria, and audit findings. Ordinary body prose may be edited directly;
  do not hand-edit tool-owned sections or frontmatter to bypass guards.

## Build and validate

Recipes in [justfile](justfile) own the commands; don't copy version pins here.
The executable entry point is `cmd/tskflwctl`, not a root Go package.

```sh
just build                 # fresh bin/tskflwctl for planning dogfood
just test                  # full Go suite with the race detector
just lint
```

Get all three green for code changes. For documentation-only work, check the changed
links, examples, and routed contracts; report any checks not run. When command metadata
changes, run `just docs` and `just docs-check`; edit the command source, not generated pages.
For planning changes, run `./bin/tskflwctl -C . lint`. Release and toolchain gates belong
to [the release guide](docs/RELEASING.md), not this quick-start.

## Where truth lives

| Subject | Owner / where to change it |
| --- | --- |
| Build, test, generation, tool versions | [justfile](justfile) recipes; [go.mod](go.mod), CI/build configuration and [release policy](docs/RELEASING.md) for versions. |
| Architectural boundaries and exceptions | [Architecture reference](docs/ARCHITECTURE.md); `.golangci.yml` and regression tests show what is actually enforced. |
| Package/type contracts | Comments beside owning Go packages, ports, and types; start with [core ports](internal/core/store.go). Tests pin behavioral guarantees. |
| CLI behavior and machine output | Command/use-case implementation and tests; `--help`, [generated CLI reference](docs/cli/tskflwctl.md), `schema`, and [wire contracts](internal/wire/wire.go) expose it. Generate derived reference, don't hand-edit it. |
| Entity authoring and managed sections | `./bin/tskflwctl schema <kind>` and `template show <kind>`; the [entity registry](internal/domain/entity.go), templates and mutation tests own changes. Hand-authored guidance needs review against behavior, not copied inventories. |
| Compatibility promises | [Threads compatibility](docs/THREADS_COMPATIBILITY.md) and [ADR-0008](planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md); wire tests check revision declarations and covered behavior. Review must judge whether compatibility classifications are truthful. |
| Decisions and rationale | [ADRs](planning/adrs/): check status and amendments. Research is a dated recommendation, not an accepted decision. |
| Work status and evidence | [Planning](planning/): task/Thread state, ACs, and audit dispositions via the CLI. README feature summaries are not a backlog. |

Code and executable checks provide evidence of implemented behavior. Generated contracts expose
registries and output shapes, but hand-authored descriptions and conventions can still drift;
generation does not prove every prose claim. Accepted ADRs and compatibility documents describe
intended decisions/promises; planning records describe work and history. These are different
authorities, not a blanket "code wins" ranking. When they disagree, identify the mismatch,
reproduce or verify it, and fix or track the drift. Ask before
changing a promise or choosing between incompatible decisions. Accepted ADRs are amended rather
than silently rewritten ([ADR-0001](planning/adrs/0001-adopt-adrs.md)).

## Read for the change, not the whole repository

| Change | Start here, then follow relevant links |
| --- | --- |
| Planning reads/writes, identity, diagnostics, recovery | [Planning-data checklist](docs/architecture/planning-data.md#change-checklist), owning ports/type comments and their regressions. |
| Dependency direction or composition | [Boundaries guide](docs/architecture/boundaries.md), named exceptions and `.golangci.yml`. |
| CLI flags/help, JSON or tabular contracts | Command source, [CLI reference](docs/cli/tskflwctl.md), [wire](internal/wire/), ADR-0008 and related tests. |
| TUI interactions or watching | [TUI architecture](docs/architecture/tui.md), local model/command contracts and navigation/watch tests. |
| Release or compiler/linter changes | [Releasing](docs/RELEASING.md), its configuration links and toolchain checks. |
| Demos or review tooling | [Assets guide](assets/README.md) or [scripts guide](scripts/README.md). External reviews require isolated sandboxes, not the shared worktree. |
| Documentation | The ownership table above; update the owner alongside the change and link from summaries. Keep useful examples, remove confirmed dead/duplicate guidance. |

When a contract changes, update its owning code comment and relevant focused guide together.
Guides link to owning source and regression evidence; package comments point back with a
repository-relative guide path, and Go API references use symbol links. Navigation tests catch
broken routes, not semantic drift; don't copy method lists or add comment-only behavior promises.

Nested agent instructions are only for genuinely local constraints and should link to owning
contracts, not duplicate this guide or weaken its safety rules. Do not rely on identical nested
discovery across tools; read applicable local guidance when working in that directory.

## Planning dogfood

Use `./bin/tskflwctl -C .` to target this checkout explicitly. Start with compact reads:
`task list -o table -c id,slug,status,description`, `task info <ref> --json`, and
`thread frontier <thread>`. Full JSON is typed machine data; `--json -c ...` is a
string-valued column projection, not the same envelope. Ask `schema` / `schema <kind>`
and the relevant command's `--help` for current fields, verbs, flags, and process exits.

Use `task ac` for criteria, `audit finding` for findings and their managed candidate rows,
and lifecycle verbs rather than `set`/`edit` for status. A tracked finding must name its
destination task. Follow-up work belongs in a scoped task, linked into the appropriate
Thread; add dependencies only for real prerequisites. Completion needs met criteria or
an explicit reasoned disposition; report remaining work honestly.
