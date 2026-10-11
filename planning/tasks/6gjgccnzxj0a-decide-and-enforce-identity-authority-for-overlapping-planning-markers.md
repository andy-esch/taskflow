---
schema: 1
id: 6gjgccnzxj0a
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Prevent overlapping configuration markers from silently assigning incompatible durable identities to one physical planning tree.
effort: M
tier: 2
priority: high
autonomy_level: 3
tags: [config, identity, topology]
created: "2026-10-10"
depends_on: [6gjdgpq17sr6]
updated_at: "2026-10-10"
---
# Decide and enforce identity authority for overlapping planning markers

## Objective

Ensure a single physical planning tree cannot silently acquire incompatible durable
identities through overlapping configuration markers. Decide the rejection/recovery
policy before changing existing topology semantics; do not auto-delete markers,
choose a winning ID, or silently rebind pointers.

## Evidence

An actual CLI probe on 2026-10-10:

```sh
tskflwctl init --path <tmp>/plan --taskflow-root planning --no-register
tskflwctl init --path <tmp>/plan/planning --taskflow-root . --no-register
tskflwctl init --path <tmp>/root-pointer --planning-repo <tmp>/plan --no-register --no-link-back
tskflwctl init --path <tmp>/data-pointer --planning-repo <tmp>/plan/planning --no-register --no-link-back
tskflwctl -C <tmp>/root-pointer config show --json
tskflwctl -C <tmp>/data-pointer config show --json
```

Both views report the same `planning_root` but different repository IDs and pointer
expectations. This is distinct from the single-governing-marker alias bug fixed by
`6gjdgpq17sr6`. Conflicting marker topology predates that fix; retaining one captured
identity does not itself resolve which marker should have authority.

## Scope and design decisions

- Inspect initialization, direct and pointer discovery, configuration migration,
  identity-bound Thread plan replay, and atlas grouping for overlapping roots.
- Decide whether bootstrap rejects a second governing marker and how existing
  overlapping markers are diagnosed across all entry-point spellings.
- Distinguish ordinary independent nested planning spaces from two markers claiming
  the exact same physical tree. Include symlinks and nested Git boundaries.
- Give explicit recovery advice preserving original IDs and documents. Never silently
  delete configuration, adopt a replacement identity, or rebind a pointer.
- Verify whether a machine-contract revision is required for changed error/value
  semantics and document the consumer action if so.

## Acceptance criteria

- [ ] Marker-authority and existing-topology recovery policy is decided and documented.
- [ ] Real bootstrap/discovery/replay tests prevent two successful entry points from silently claiming incompatible identities for one physical tree.
- [ ] Independent nested spaces, physical containment, Git boundaries and legacy identity-less behavior are pinned without accidental cross-binding.
- [ ] Refusals/recovery preserve documents, marker contents and pointer expectations; full validation passes.

## Sequencing

Follow the single-marker alias fix (`6gjdgpq17sr6`). This configuration-topology
design/hardening follow-up is deliberately outside the CLI-contract Thread's seven
delivery members; it is not another command-discovery or bounded-query gap. The
pointer-parity reviewers should challenge that boundary if the remaining topology
produces an in-scope failure for a single-governing-marker alias.

## Independent corroboration

The [Codex pointer-parity review](../audits/6gjgcy198g1m-2026-10-10-planning-pointer-alias-identity-implementation-codex.md)
independently reproduced two successful entry points reporting one physical root
with different IDs. A Thread plan composed through the outer identity was refused
through the inner identity with conflict exit 14 and no document write. That adds
an actual replay consequence to the config-view evidence, not a claim of corruption
or successful repair. The reviewer found no analogous single-marker defect and
considered separate authority/recovery-policy work defensible. This task remains
unresolved; choose the policy before implementation, with no automatic winning ID,
configuration deletion or pointer rebinding.
