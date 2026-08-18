---
schema: 1
id: 6g1c120fm99d
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: A leading UTF-8 BOM defeats splitFrontmatter's ---\n prefix check, so the file reports 'missing frontmatter' for every entity kind; pre-existing, not research-specific.
effort: Unknown
tier: 4
priority: low
autonomy_level: 3
tags: [store, robustness]
created: "2026-08-18"
---

# A UTF-8 BOM makes entity frontmatter unparseable

## Objective

A file whose first bytes are a UTF-8 BOM (`EF BB BF`) before its `---` fence is reported
as having **no frontmatter at all**, so the entity is unreadable:

    ! .../6dr29v000b0m-bom.md
        validation failed: malformed frontmatter: missing frontmatter — a research doc
        must open with a `---` YAML block (created; see `tskflwctl schema research`)

## Scope — pre-existing and entity-agnostic

Found while stress-testing research, but it is **not** a research bug and not a
regression: `splitFrontmatter` gates on `bytes.HasPrefix(content, []byte("---\n"))`, so a
BOM defeats it for every kind. Verified with a BOM-prefixed *task* file, which fails
identically. Filed separately from the research follow-ups for that reason.

Low priority: nothing in this repo writes a BOM (the atomic writers emit LF and no BOM),
so it only bites a file touched by an editor that adds one — common on Windows, and the
diagnostic points at the wrong cause, which is the real cost. The message says "missing
frontmatter" when the frontmatter is right there.

## Acceptance criteria

- [ ] A leading UTF-8 BOM is tolerated on read for every entity kind: the file parses and
      lists normally.
- [ ] A surgical edit of a BOM-prefixed file does not duplicate or strip the BOM
      unexpectedly — whatever the chosen behaviour, it is deliberate and tested (the
      frontmatter editor already preserves CRLF; the BOM deserves the same explicit call).
- [ ] Failing that, the diagnostic names the BOM instead of claiming the frontmatter is
      missing — a correct error is an acceptable outcome if tolerating it is judged wrong.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Found alongside the research stress-test findings on epic
  [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
