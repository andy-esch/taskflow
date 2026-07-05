---
schema: 1
id: 6fk6nnca15wj
status: ready-to-start
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Shell completion offers bare slugs only; the flat layout also resolves by id-prefix/stem and allows dup slugs — id-prefix typing yields nothing and dup-slugs can't be disambiguated.
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, completion, flatten-followup]
created: "2026-07-05"
updated_at: "2026-07-05"
---

# Complete task/audit id-prefixes and full stems in shell completion

## Objective

<why / what — one short paragraph>

## Acceptance criteria

- [ ] <observable outcome>

## Out of scope

- <explicitly excluded>

## Related

- Epic [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]

## Finding (adversarial review, 2026-07-05)

`taskCompleter`/`auditCompleter` (`internal/cli/completion.go`) build candidates from
`flatSlug(<id>-<slug>)` — the bare human slug — and match `strings.HasPrefix(slug, toComplete)`.
Gaps under the flat layout (ADR-0003 §4):

- Typing an **id-prefix** (`6fj…`) or the **full `<id>-<slug>` stem** matches nothing, even
  though `resolveID` accepts both.
- **Dup slugs** are legal but indistinguishable (both offer the same bare slug).
- The already-typed `taken` map is keyed by the raw arg, so a stem-typed arg isn't excluded
  when the scan checks `taken[slug]`.

Fix: match completion against the full stem too (id-prefix + slug); key `taken` via `flatSlug`.
Adjacent to the scheme-2 id-prefix reference work.
