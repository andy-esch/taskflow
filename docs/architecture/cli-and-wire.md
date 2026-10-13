# CLI and machine contracts

[Architecture index](../ARCHITECTURE.md) · [Composition](boundaries.md#composition-and-local-exceptions)

The [CLI package](../../internal/cli/root.go) owns user interaction and output
routing, not persistence construction or application policy.

## Invocation and output

One `cli.App` per command tree receives explicit dependencies. Planning opens only
after target flags; metadata-only help/schema/generation does not require a planning
tree. All output uses injected writers, not process-global printing.

Controllers call application services, then map/render results.
[cli/render](../../internal/cli/render/render.go) owns human formatters and writer
wrappers; [wire](../../internal/wire/wire.go) owns neutral machine values and
constructors that another adapter can reuse. JSON never emits ANSI.

Every runnable leaf declares read-only/mutating capability.
[Command safety](../../internal/cli/command_safety.go) binds that capability before
discovery; guarded secondary adapters reject a mutation reached by a read-only leaf.
Hidden, deprecated and framework-generated routes are covered. Capability describes
what the command can do, not whether this invocation writes; preview support is
command-specific. See [safety tests](../../internal/cli/command_safety_test.go).

Optional list queries live in [core](../../internal/core/list_query.go):
filters, stable ordering, metadata search, dates, limits and continuation precede
rendering. A bound limits returned rows, not reader I/O. Cursor digests detect matching
view/diagnostic changes, not a transactional snapshot or persisted source-set identity.
Typed and string-valued projected JSON share page metadata.
[README query examples](../../README.md#bounded-queries) own the user walkthrough.

## Compatibility and vocabulary

[ADR-0008](../../planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
owns the single monotonic machine revision—not SemVer and not the advisory on-disk
`schema:` marker. It covers typed envelopes and dynamic projections. Generated JSON
Schema covers typed envelopes only, with a revision-qualified ID and exact revision
constraints. Classified changelog tests require declarations; human review still
judges whether additive/non-additive promises are truthful.

[The exit taxonomy](../../internal/cli/exit.go) owns stable codes, names, meanings
and active/reserved state. Domain failures remain neutral; success, prompt abort,
generic failures and reservations are not invented domain classes. `schema --json`
publishes the table. Process tests pin emitted code/envelope pairs.

Closed semantic vocabularies come from domain owners. Shared finding/criterion words
are overlap, not interchangeable full sets. Schema/authoring guidance derives metadata
from [the descriptor](../../internal/domain/entity.go), but hand-written convention
prose can still drift. Generation is not proof of meaning.

The [2026-10-12 machine-contract audit](../../planning/audits/6gjzzvdy5jnw-2026-10-12-arch-machine-contract.md)
identifies unpublished projection selectors, vocabulary/schema tolerance policy and
command-to-envelope discovery gaps. This extraction neither adopts its proposals nor
claims those registries/mappings already ship.

## Validation

- `just docs` regenerates [CLI reference](../cli/tskflwctl.md) from command metadata;
  `just docs-check` fails on derived-page drift. Edit the owner, not generated pages.
- Pure domain/core tests use portable fixtures; adapter round trips use temporary
  trees; in-process CLI tests inject services and capture writers.
- [Golden tests](../../internal/cli/golden_test.go) pin byte contracts.
  Update with `go test ./internal/cli -update`, not `./...`: first advance/classify
  the revision. The updater refuses changed/new machine snapshots at an unchanged
  revision; only a complete successful update advances the marker.
- The projection-contract golden covers selector/alias behavior that the static
  envelope schema does not express.
- [Wire registry tests](../../internal/wire/registry_test.go),
  [revision tests](../../internal/wire/wire_changelog_test.go) and
  [envelope tests](../../internal/wire/envelopes_test.go) cover registration,
  revision declarations and payload validation.
- [Binary smoke tests](../../cmd/tskflwctl/main_test.go) exercise process exits and
  lifecycle. Run `just test` (race-enabled), `just lint`, and `just build`.
