---
schema: 1
id: 6gppr2js1bqq
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Make every optional envelope field's omission encoder-independent with omitzero, so the later json/v2 swap cannot add or drop fields, keeping existing JSON bytes unless a revision is classified.
effort: half day
tier: 4
priority: low
autonomy_level: 3
tags: [wire, modernization, json]
created: "2026-10-23"
audit_sources: [planning/audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md]
updated_at: "2026-10-23"
---
# Make wire envelope field omission explicit with omitzero

## Objective

About fifteen optional envelope fields are heap pointers only so `omitempty` drops them.
`omitzero` states that intent directly. Convert only the fields where nil meant "absent", and
prove the machine contract is unchanged.

This also prepares the [Go 1.27 json/v2 swap](6gppr2jwqnpm-raise-the-go-minimum-to-1.27-and-adopt-its-language-and-library-features.md).
`omitempty` means different things in the two encoders: v1 drops `false`, `0` and nil pointers but
keeps a non-nil pointer to an empty struct, while v2 keeps `false` and `0` and drops any value that
encodes as empty JSON. Verified on go1.27.2: the same struct encodes as `{"p":{}}` under v1 and
`{"b":false,"n":0}` under v2. `omitzero` means the same in both, so making omission explicit here
lets the encoder swap land without fields silently appearing or disappearing.

## Acceptance criteria

- [ ] Optional single-struct fields in `internal/wire` whose nil meant omission
  use value types tagged `omitzero`.
- [ ] Slice fields keep their `omitempty` semantics, and `dry_run` stays always-present on mutation envelopes.
- [ ] Machine goldens are byte-identical, or any output difference is recorded as a classified ADR-0008 revision.
- [ ] The generated JSON Schema still validates real output for every converted envelope.
- [ ] Every bool and numeric `omitempty` field in `internal/wire` either moves
  to `omitzero` or is documented as intentionally present, so v1 and v2 encoders
  emit the same fields; a test encodes representative envelopes both ways and
  compares them.

## Out of scope

- `encoding/json/v2` adoption (tracked by the Go 1.27 minimum task).

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-14 Go 1.27 modernization](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md), finding M1
- [ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
