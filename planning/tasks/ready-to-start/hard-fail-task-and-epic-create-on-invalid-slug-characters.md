---
status: ready-to-start
epic: 17-pm-go-cli
description: task/epic new should reject titles with filename-hostile chars (colon, em-dash) with a clear error and a suggested clean title
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, validation]
created: "2026-06-13"
---

# Hard fail task and epic create on invalid slug characters

## Objective

`task new` / `epic new` run the title through `Slugify` and only fail on an
**empty** slug. As of the 2026-06-13 allowlist rewrite, `Slugify` *silently
normalizes* anything else — `"Foo: bar — baz"` → `foo-bar-baz`. That now prevents
bad chars leaking into filenames (good), but the user wants the create path to
**hard-fail with a clear error** instead of silently producing a different slug,
so the surprise surfaces at creation rather than as an unexpected filename later.
(Motivating bugs: an em-dash leaked into a filename pre-rewrite; colons have crept
in too.)

## Design tension to resolve

`Slugify` should **keep normalizing** — it's the safety net for every internal /
legacy path (lint, moves, imports). The new behavior belongs on the **create path
only**: validate the title before slugifying and reject if it contains characters
that would be normalized away.

- **Recommended:** a `domain` helper (e.g. `ValidateTitle`) that flags
  filename-hostile runes (`: — / \ * ? < > | "` and other unicode punctuation /
  symbols), returning `ErrValidation`. `NewTask`/`NewEpic` call it before
  `Slugify`. The error names the offending chars and **suggests the cleaned form**
  (`Slugify` output) so the fix is copy-paste: `did you mean "Foo bar baz"?`.
- Spaces, apostrophes, dots (versions) and letters/digits of any script stay
  valid — only punctuation/symbols that get dropped are rejected.

## Acceptance criteria

- [ ] `task new "Fix: the thing — now"` exits `ErrValidation` (code 11) naming the
      offending chars and suggesting a clean title; nothing is written.
- [ ] Plain titles (incl. apostrophes, version dots, non-ASCII letters) still
      create successfully.
- [ ] `epic new` gets the same guard. Tests for accept + reject cases.

## Out of scope

- Changing `Slugify` itself (stays a normalizing allowlist — the safety net).
- Renaming/validating *existing* misnamed files (a separate `lint`-style sweep if
  wanted).

## Related

- Epic [[17-pm-go-cli]]
- Builds on the 2026-06-13 `Slugify` allowlist rewrite (`internal/domain/slug.go`).
- Coordinate with the in-flight global `--dry-run` work touching the same create path.
