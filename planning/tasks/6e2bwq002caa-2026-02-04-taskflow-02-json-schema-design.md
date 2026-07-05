---
status: completed
epic: 00-taskflow-v1-core
effort: 3-4 hours
tier: 1
priority: high
project: taskflow-bootstrap
tags: [architecture, json, performance, schema]
completed_at: 2026-02-04
id: 6e2bwq002caa
---

# TaskFlow 02: JSON Schema & Scalability Design

**Goal**: Design the schema for `planning-index.json` that balances read speed, write ease, and scalability for 10,000+ tasks.

## Context
The Go CLI reads this file directly for the "Fast Path" (listing/filtering).
- **Scale**: Must handle 10k tasks without CLI lag (< 100ms parse time).
- **Bloat**: Avoid duplicating large content bodies if possible.

## Key Questions to Answer
1.  **Normalization**: Should we store `tags` and `epics` as normalized lists with IDs, or denormalized strings? (Trade-off: File size vs Parse complexity).
2.  **Content**: Do we need the full markdown body in the index?
    - *Decision*: Probably NOT. Just frontmatter + Title + Objective summary.
    - *Why*: `grep` is fast enough for body search if needed, or we use the Vector DB for deep search.
3.  **Sharding**: Should `planning-index.json` be split (e.g., `active.json` vs `archive.json`)?

## Front Matter Standards (To Be Enforced)
The JSON schema must validate these fields (currently enforced by `pm validate`):

```yaml
status: ready-to-start | next-up | in-progress | completed | deprecated
epic: <epic-id> (must match `epics/*.md`)
effort: <string> (e.g. "2-4 hours")
tier: 1 | 2 | 3 | 4 (Priority Tier)
priority: high | medium | low
tags: [tag1, tag2]
project: <string> (Optional: Cross-cutting initiative)
```

## Deliverables
- [ ] Schema Definition (TypeScript/Go struct definition or JSON Schema).
- [ ] Benchmarks: Generate a dummy 10k task JSON and time the Go unmarshal.
- [ ] Strategy for handling "body content" (Index vs File Read).
