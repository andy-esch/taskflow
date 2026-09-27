---
schema: 1
id: 6g1dp77y0g5n
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: The migration tool emits a duplicate tags key on block-style YAML (unparseable file), mishandles CRLF frontmatter, and can mint an id already on disk.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tools, store]
created: "2026-08-18"
updated_at: "2026-09-27"
audited: "2026-09-27"
audit_sources: [2026-09-27-weekly-task-sweep]
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

## Sweep verification (2026-09-27)

Automated weekly sweep. **All four defects are still present** and no acceptance criterion
was ticked. `internal/tools/researchmigrate/main.go` has had no functional change since
this task was filed — the only commit touching the directory since 2026-08-18 is the merge
`7ed39c8`, so the churn that made this task eligible was a false positive. Line numbers
have drifted; refs re-anchored (2026-09-27):

- **§1 duplicate `tags:`** — `ensureFrontmatter:274-276` still derives `wantTags` from
  `frontmatterField(content, "tags") == ""`, and `:280` still calls
  `dropFrontmatterKeys(rest, "schema", "id", "created")` — `tags` and `description` are
  still absent from that drop list. Intact.
- **§2 CRLF** — `bytes.HasPrefix(content, []byte("---\n"))` at `:262` (`ensureFrontmatter`)
  and `:325` (`frontmatterField`). Intact, and still divergent from the store's
  CRLF-tolerant `splitFrontmatter`.
- **§3 cross-run id collisions** — `seenID := map[string]bool{}` at `:87` (seeded empty),
  `mintUnique(millis, seenID)` at `:121`, `checkCollisions(renames)` at `:138`. Intact.
- **§4 `applyPlan` not atomic** — `:459-476`, still plain `os.WriteFile` then `os.Remove`.
  Intact.
- The repair-path claim also still holds: `fixFrontmatterText`
  (`internal/store/fix.go:327`) and `backfillMissingID` (`:298`) both still exist and
  neither dedupes keys.

### Two corrections

- `## Scope` says the tool *"has already run cleanly against this repo's 28 docs"*. The
  corpus is now **32 docs**, all of them id-led, so the tool remains fully spent here and
  the LATENT framing is unchanged — only the count is stale.
- Acceptance criterion 5 (*"Either make `applyPlan` atomic or document why it needn't
  be"*) has a concrete obstacle worth knowing before starting: the obvious helper,
  `writeFileAtomic` in `internal/store/atomic.go:48`, is **unexported**, so
  `internal/tools/researchmigrate` cannot reuse it without either duplicating the pattern
  or waiting on a shared foundations package. That is exactly what
  [unify-the-three-divergent-writefileatomic-implementations](6g63jj1dh0sb-unify-the-three-divergent-writefileatomic-implementations.md)
  proposes. Taking the "document why it needn't be" branch keeps this task independent;
  taking the "make it atomic" branch creates a real ordering against that task. Flagged,
  not decided — and `depends_on` is deliberately untouched.

## Progress Log

- 2026-09-27: automated weekly sweep — all four defects re-verified present with refs re-anchored to current line numbers; corrected the corpus count (28 → 32, all id-led) and noted that AC 5's atomic-write helper is unexported, making that criterion a fork with a real ordering against the writeFileAtomic unification task.
