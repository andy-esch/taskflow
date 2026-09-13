---
schema: 1
id: 6g9cz5sv6kck
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Promote copyable GoDoc snippets and contract claims into compiled examples or focused tests so API evolution makes documentation drift fail loudly.
effort: 4-6 hours
tier: 3
priority: medium
autonomy_level: 4
tags: [documentation, go, testing, godoc]
created: "2026-09-12"
depends_on: [6g9cz5sk3b0d]
updated_at: "2026-09-13"
---
# Make Go package documentation examples executable

## Objective

Keep package documentation useful after APIs and conventions evolve by compiling the examples and pinning important claims close to the behavior they describe.

The Desirelines review found the failure mode to avoid: static `doc.go` examples call `http.Error` while its agent guide forbids that response path, and `dispatcher/ports/doc.go` says there are two ports while the package now declares six interfaces. Desirelines also has useful `Example...` tests in its ports package; Taskflow should favor that executable shape for copyable flows.

## Acceptance criteria

- [ ] Copyable package usage flows are implemented as `Example...` tests where practical so normal `go test ./...` compiles and runs them.
- [ ] Examples exercise public composition and failure behavior without importing concrete adapters into inward packages or duplicating large integration fixtures.
- [ ] Important prose claims that can be expressed as focused tests or existing architecture fitness rules link to those checks instead of restating mutable inventories.
- [ ] Static snippets retained in `doc.go` are short, symbol-linked, and audited against the executable examples; copied interface declarations and mutable enum or command lists are removed.
- [ ] Example names and output assertions are chosen deliberately: compile-only examples are used when output is unstable, and output examples only when byte behavior is part of the contract.
- [ ] The ordinary test command runs all documentation examples, and the package-documentation standard explains this maintenance expectation.

## Out of scope

- Turning every comment into a test.
- Snapshot-testing prose formatting or making harmless wording changes fail CI.
- Adding integration dependencies solely to make a documentation example realistic.
- Replacing existing domain, store, CLI, or TUI tests that already pin the same invariant.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
