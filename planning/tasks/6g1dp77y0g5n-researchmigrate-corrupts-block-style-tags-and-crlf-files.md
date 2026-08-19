---
schema: 1
id: 6g1dp77y0g5n
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: The migration tool emits a duplicate tags key on block-style YAML (making the file unparseable), mishandles CRLF frontmatter, and can mint an id that already exists on disk.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tools, store]
created: "2026-08-18"
---

# researchmigrate corrupts block-style tags and CRLF files

## Scope

`internal/tools/researchmigrate` is throwaway-by-design and has already run cleanly against
this repo's 28 docs (no duplicate ids, lint clean). These are LATENT bugs: the tool is
committed, so another repo adopting research-as-an-entity would hit them.

## 1. A block-style `tags:` produces a duplicate key, making the file permanently unparseable

`frontmatterField` reads a block list (`tags:\n  - tui`) as `""`, so `wantTags` is true and
`writeContractFields` emits `tags: []`; `dropFrontmatterKeys` only drops
`schema`/`id`/`created`, so the ORIGINAL `tags:` block survives too. Confirmed:

    PROBLEM …-block-tags.md: validation failed: malformed frontmatter:
      line 7: mapping key "tags" already defined at line 5

The doc is dropped from every listing and is unresolvable. `lint --fix` cannot repair it:
`fixFrontmatterText` doesn't dedupe keys and `backfillMissingID` bails on unparseable YAML.
The same hazard applies to a block-style `description:`.

## 2. A CRLF file is treated as having no frontmatter

`bytes.HasPrefix(content, []byte("---\n"))` (and the same check in `frontmatterField`) does
not tolerate `---\r\n`, unlike the store's own `splitFrontmatter`, which explicitly does.
Confirmed: a CRLF doc with real `description`/`tags`/`status` comes out with a FRESH block
(`description: ""`, `tags: []`) followed by the entire original file — old fence and all —
as body prose. The declared values are silently lost as fields.

## 3. Cross-run id collisions

`seenID` is seeded only from ids minted in the current run, and `checkCollisions` compares
`renames` against each other. Ids already on disk (the already-id-led docs the tool skips)
are invisible, so a second run over a newly-added same-day doc can mint an id that already
exists — the same bricking failure as the core mint path.

## 4. (note) `applyPlan` is not atomic

Plain `os.WriteFile` + `os.Remove`, so a crash between them leaves both filenames on disk.
Acceptable for a throwaway tool gated on a clean git tree, but worth a comment saying so
rather than leaving it implicit.

## Acceptance criteria

- [ ] Detect an existing `tags:`/`description:` in ANY YAML form (block or flow) so no
      duplicate key is ever emitted.
- [ ] Tolerate `---\r\n`, matching the store's `splitFrontmatter`; a CRLF doc keeps its
      declared fields.
- [ ] Seed `seenID` from the ids already on disk, and include them in the collision check.
- [ ] Fixtures for all three, since the live corpus doesn't exercise any of them.
- [ ] Either make `applyPlan` atomic or document why it needn't be.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Found by an independent adversarial correctness review, 2026-08-18
