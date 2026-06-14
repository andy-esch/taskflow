---
status: completed
epic: 17-pm-go-cli
description: task/epic new should reject titles with filename-hostile chars (colon, em-dash) with a clear error and a suggested clean title
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, validation]
created: "2026-06-13"
updated_at: "2026-06-13"
started_at: "2026-06-13"
completed_at: "2026-06-13"
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

- [x] `task new "Fix: the thing — now"` exits `ErrValidation` (code 11) naming the
      offending chars and suggesting a clean title; nothing is written.
- [x] Plain titles (incl. apostrophes, version dots, non-ASCII letters) still
      create successfully.
- [x] `epic new` gets the same guard. Tests for accept + reject cases.

## Out of scope

- Changing `Slugify` itself (stays a normalizing allowlist — the safety net).
- Renaming/validating *existing* misnamed files (a separate `lint`-style sweep if
  wanted).

## Progress Log

### 2026-06-13 — implemented (suite + lint green)

- **`domain.ValidateTitle`** (`slug.go`) — rejects filename-hostile runes:
  filesystem-reserved ASCII (`: / \ * ? < > | "`), control chars, and **non-ASCII
  punctuation/symbols** (em/en dashes, curly quotes, bullets, arrows — the class
  that motivated this). Benign ASCII punctuation (parens, commas, hyphens) and
  apostrophes (incl. curly) are allowed — they slugify predictably. The error names
  the offending chars and **suggests a clean title** via `suggestTitle` (hostile
  runes → spaces, whitespace collapsed): `… (: —); try "Fix the thing now"`.
- **Wired into `NewTask` + `NewEpic`** before `Slugify` (the empty-slug check
  stays). `Slugify` is unchanged — still the normalizing safety net for every
  internal/legacy path; this guard is create-path-only, per the design.
- Verified end-to-end: `task new "Fix: the thing — now"` and `epic new
  "Plan: phase — two"` both exit 11 with the suggestion; clean titles still create.
- Tests: `domain.TestValidateTitle` (reject set incl. colon/em-dash/slash/curly
  quotes/bullet/arrow; accept parens/apostrophes/dots/hyphens/non-ASCII letters/CJK;
  suggestion text), `core.TestService_Create_RejectsHostileTitle` (task + epic
  reject, clean title creates).

## Related

- Epic [[17-pm-go-cli]]
- Builds on the 2026-06-13 `Slugify` allowlist rewrite (`internal/domain/slug.go`).
