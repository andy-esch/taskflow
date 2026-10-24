---
schema: 1
id: 6gppr2jwqnpm
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Move go.mod's minimum to 1.27 per the release policy, then adopt encoding/json/v2, strings.CutLast, and generic methods where they simplify, classifying JSON byte changes.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [toolchain, modernization, go]
created: "2026-10-23"
audit_sources: [planning/audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md]
updated_at: "2026-10-23"
depends_on: [6gppr2js1bqq]
---
# Raise the Go minimum to 1.27 and adopt its language and library features

## Objective

During 2026-10-23 audit triage the owner chose to raise the supported minimum from Go 1.26 to
1.27 so the remaining modernization findings can land. Bump it the way the release guide
prescribes, then adopt the three 1.27 features where they simplify real code. Checked locally on
go1.27.2: `encoding/json/v2`, `jsontext`, `strings.CutLast`, and generic methods all compile
without a GOEXPERIMENT.

## Acceptance criteria

- [ ] `go.mod`, CI, the release container, the linter build, and Renovate rules agree on the 1.27 minimum per the release guide, and `just toolchain-check` passes.
- [ ] `wire.EncodeJSON` and ordered column projections use `encoding/json/v2`/`jsontext`; the HTML-escaping change in emitted bytes is recorded as a classified ADR-0008 revision, with goldens regenerated through the gate.
- [ ] Manual `LastIndex` delimiter arithmetic at the named sites uses `strings.CutLast` with unchanged behavior.
- [ ] `retryOnConflict`, `editFile`/`writeBody`, and `runMoves` become receiver methods only where that removes closure plumbing; any skipped conversion records why.
- [ ] `just build`, `just test`, `just lint`, and `just docs-check` pass.
- [ ] Swapping to json/v2 adds or drops no envelope field relative to v1 output:
  `omitempty` semantics are checked field by field (v2 keeps `false` and `0` but
  drops empty-JSON values), and goldens show only the classified escaping
  change.

## Out of scope

- Go 1.24-1.26 idioms, `os.Root`, and `omitzero` (separate tasks).
- Replacing taskflow's Crockford base32 IDs with the standard `uuid` package.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-14 Go 1.27 modernization](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md), findings H2, M3, M4
- [Release guide](../../docs/RELEASING.md); completed [toolchain coordination](6gd68x5gqkk0-reconcile-go-toolchain-versioning-across-go.mod-ci-linters-and-renovate.md)
