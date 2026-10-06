---
schema: 1
id: 6ggxymjbf86r
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Make pure core/domain boundaries executable beyond repository-internal imports without freezing harmless utility dependencies.
effort: 4-6 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, lint, tests]
created: "2026-10-05"
depends_on: [6gg7e59mm68g]
updated_at: "2026-10-05"
---
# Guard core and domain against direct I/O and framework imports

## Objective

Make the implemented pure application/domain boundary harder to bypass accidentally. The
adapter-neutral closeout found that the core/domain depguard rules restrict repository-internal
imports, but do not restrict standard-library effect packages or third-party frameworks. Today's
production imports comply; this is an enforcement gap, not a discovered runtime bypass.

## Evidence and scope

- `.golangci.yml` uses `list-mode: lax` with a deny on the repository's `internal` prefix.
  Thus a future `os`, `net/http`, Cobra, or Bubble Tea import is outside those deny patterns.
- Current `go list` output for core/domain contains only standard-library utilities and allowed
  internal packages. Pure `path/filepath` formatting in core is not filesystem I/O and must not
  be confused with opening persistence.
- After the dependency-policy ADR is agreed, choose focused production import checks that forbid
  direct filesystem/network/process effects and adapter/framework coupling. Do not freeze today's
  entire import inventory or prohibit harmless pure primitives without a rationale.
- Cover direct files and future subpackages; keep legitimate test composition explicit.
- Use compiler-valid negative probes and restored controls. Report the named rule violation,
  not an unrelated compile failure. Keep normal `just lint` as the contributor entry point.
- Update the contributor checklist and focused architecture guidance to state exactly what is
  enforced and what still requires behavioral tests/review.

## Related

- [Adapter-neutral Thread closeout](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Dependency-policy ADR task](6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md)
- [Documentation Thread](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- [Descriptive import-graph verification](6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md)

## Out of scope

- Reopening the completed planning-port migration or claiming an existing I/O violation.
- A new backend, architectural framework, universal adapter conformance suite, or exact-import snapshot.
- Silently accepting the ADR or widening exceptions to make lint pass.

## Acceptance criteria

- [ ] An agreed focused policy rejects production direct I/O and
  adapter/framework imports in core/domain while retaining legitimate pure
  utilities and explicit test composition.
- [ ] Compiler-valid direct and nested forbidden-import probes fail standard
  lint with the intended diagnostics; restored controls pass.
- [ ] The standard lint path enforces the policy for current files and future
  subpackages without an exact snapshot of current imports.
- [ ] Contributor guidance identifies the implemented enforcement, its limits,
  and the governing ADR without claiming behavioral or security guarantees from
  import checks.
