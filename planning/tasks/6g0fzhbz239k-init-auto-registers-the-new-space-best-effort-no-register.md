---
schema: 1
id: 6g0fzhbz239k
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: 'After a successful scaffold or pointer init, append a [[space]] entry. Best-effort like LinkBack: warns to stderr, never fails the init. Opt out via --no-register/TSKFLW_NO_REGISTER.'
effort: M
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, config, multi-repo]
created: "2026-08-15"
started_at: "2026-08-22"
updated_at: "2026-08-22"
completed_at: "2026-08-22"
---
# `init` auto-registers the new space (best-effort, `--no-register`)

## Objective

Close the loop so the registry populates itself: after `init` creates a fresh scaffold
**or** pointer configuration, register that checkout through `core.SpaceRegistryService`
and report the receipt alongside the topology result. Without this, every new repo needs a
second command nobody will remember to run.

Preserve the simplified lifecycle that shipped after this task was drafted: a bare `init`
against an existing configuration is a topology read, not a hidden migration or home-state
mutation. `config migrate` remains the configuration-upgrade path, and explicit
`space add` remains how an existing unregistered checkout opts in.

## Notes

- **Best-effort, exactly like `LinkBack`**: once the topology has been created (or would
  be created under dry-run), registry failure is a warning rather than an init failure.
  Warnings go to **stderr** and never corrupt `--json` stdout.
- Opt out with `--no-register` and `TSKFLW_NO_REGISTER=1`, mirroring
  `--no-link-back`. The env form matters for CI.
- Both fresh modes register. The candidate is the checkout/config directory, not the
  resolved planning root, so a pointer registers as itself while carrying the target's
  durable planning id.
- Registration flows through the application service. `init` must not import or call
  `userconfig`, `spacehealth`, or TOML registry functions directly.
- `--dry-run` previews the same registration and receipt without writing either the
  topology or home registry. A prospective scaffold id may differ from a later real run,
  just like any minted dry-run value; it must still be internally consistent within the
  preview.
- The `--json` init envelope grows a typed registration receipt (golden-tested and schema
  documented). Best-effort warnings remain stderr-only.
- Deliberately **not** in scope: auto-registering any repo you merely run a command
  in. A read-only command writing to `$HOME` is surprising, and throwaway clones and
  worktrees would accumulate.

## Design issue before implementation (2026-08-22)

The current config bootstrap functions return only `[]created`. A real init can be
rediscovered after writing, but a dry-run deliberately leaves no config to discover; in
scaffold mode the prospective durable id is also minted only inside the real config write.
Calling `SpaceRegistryService.Add(path)` therefore cannot produce an honest dry-run
receipt, while skipping registration in dry-run would weaken the global preview contract.

Recommended resolution:

1. Make scaffold and pointer bootstrap return a typed result containing `Created`, the
   checkout/config directory, and the durable planning id written or planned. Preserve
   the existing `Init`/`InitPointer` validation, exclusivity, and lifecycle behavior.
2. Add a narrow `SpaceRegistryService` entry point for an already-validated initialized
   checkout and share label validation and mutation logic with ordinary `Add`.
3. Represent the prepared candidate by its physical checkout path plus verification id;
   let `spacestore` normalize the persisted path. This also keeps storage spelling out of
   the core value.
4. Invoke it only when the config file was or would be freshly created. Bare existing
   `init`, scaffold repair, pointer no-op, and `config migrate` do not mutate home state.

The init receipt should intentionally expose only facts that exist in both preview and
apply: local id, persisted path, verification id, changed, and dry-run. It should not reuse
the diagnosed `SpaceEntry` wire shape: before a dry-run topology exists there is no honest
role/state/planning-root diagnosis to serialize. A real `space add` can continue returning
the richer diagnosed mutation receipt.

This is an application/config result-shape change, not a new workspace abstraction. The
broader `Resolve() -> Workspace` seam remains deferred until an atlas or served consumer
needs it.

## Acceptance criteria

- [x] Fresh scaffold and pointer `init` append one entry through
      `SpaceRegistryService`, using the checkout path and planning identity returned by a
      typed bootstrap result.
- [x] Bare `init` on an existing config remains a read with no home-registry write;
      explicit topology no-ops/repairs and `config migrate` do not auto-register either.
- [x] An unwritable registry or label conflict warns actionably on stderr; topology init
      still succeeds with exit 0 and valid JSON stdout.
- [x] `--no-register` / `TSKFLW_NO_REGISTER=1` suppress registration; the flag does not
      turn a bare existing init into a topology mutation.
- [x] Re-running init cannot duplicate an entry, and ordinary physical-path dedup remains
      owned by the registry adapter.
- [x] `--dry-run` reports the would-be registration without writing the config, tree, or
      registry.
- [x] Human output carries one registration line and `--json` carries a typed receipt;
      schema comments, goldens, CLI docs, and relevant user docs are updated.
- [x] No test writes to a real `$HOME`
- [x] `just test` + `just lint` green

## Implementation (2026-08-22)

- `config.Init` and `InitPointer` now return one `InitResult` carrying created paths,
  fresh-config ownership, the physical checkout/config directory, and the durable planning
  id written or previewed. Missing dry-run leaves resolve through their nearest existing
  ancestor, so symlinked-parent previews use the same physical path as apply.
- `core.SpaceRegistryService.RegisterInitialized` accepts that already-validated result and
  shares the ordinary add use case's label/default/mutation policy. The secondary adapter
  still owns persisted path spelling and physical-path dedup; repo-scoped `config` remains
  unable to import home-scoped `userconfig`. A deduped row whose stored verification id
  differs from the freshly initialized identity is treated as a stale-path conflict with
  an explicit forget/add repair, never as a misleading "already registered" receipt.
- Fresh scaffold and pointer modes invoke registration only after their hard topology work.
  Failure is an actionable stderr warning with an explicit `space add` retry; human output
  gets one receipt line, while schema `1.43` adds the preview/apply-stable typed JSON
  receipt without pretending a dry-run checkout has diagnosable topology.
- Focused coverage pins both modes and their shared planning identity, checkout-vs-planning
  root placement, physical-path dedup, dry-run non-mutation, flag/env opt-outs, existing
  init/repair non-registration, human receipt cardinality, and valid JSON under a label
  conflict. Existing CLI init tests opt out or isolate their config home so the suite never
  leaks registry state between cases.

Validation: full `go test -race ./...`; `golangci-lint run ./...` (0 issues); module tidy
diff, generated CLI docs/schema comments and machine goldens, planning lint, and
`git diff --check` are clean.

## Out of scope

- Auto-registration on ordinary command runs
- Registering pre-existing checkouts from bare `init` or `config migrate`; use `space add`
- The board's "unregistered current repo" prompt (TUI, later)
- Activating the reusable workspace resolver or changing the `spaces.toml` schema

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
- [Reusable space-registry application boundary](6g28rv8jm1g7-establish-one-reusable-space-registry-application-boundary.md)
