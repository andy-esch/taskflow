---
schema: 1
id: 6fpfcecdymca
status: completed
epic: 20-cli-ux-and-ergonomics
description: 'Structure-aware body mutation + metadata reads (task log/ac, task path/info, show --section): the read-only cluster shipped'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, agents, ux, dx]
created: "2026-07-15"
started_at: "2026-07-15"
updated_at: "2026-07-16"
completed_at: "2026-07-16"
---
> ⚠️ **Externally proposed — filed 2026-07-15** from a second agent dogfooding
> session — the sequel to
> [agent-facing-cli-ergonomics-batch](6fbj87000anh-agent-facing-cli-ergonomics-batch.md),
> which already shipped `--body-file`, `task append`, `task set --body`, and the
> create envelope. These are the friction points that bit *this* round, in the
> agent's own impact order. The unifying thread: **the CLI still treats the body
> as an opaque append-only blob.** The wins are in teaching it that the AC list
> and the Progress Log are *structural sections*, plus a cheap file-location /
> metadata read.

## Objective

1. **`task log <slug>` — dated Progress Log append (the headline).** Adding a
   progress entry is the agent's single most frequent write, and `task append`
   is structure-blind: it re-declares `## Progress Log` when one already exists,
   producing a duplicate header that then needs a hand-merge. `task log <slug>
   --body "…"` (and `--body-file -`) should (a) append a bullet under the
   *existing* Progress Log section, creating the section only if absent, and
   (b) auto-stamp today's date. Kills both the duplicate-header footgun and the
   manual date-typing.
2. **`task ac <slug>` — acceptance-criteria checkboxes via CLI.** Closing a task
   today means finding the file and hand-flipping `- [ ]` → `- [x]` in an editor.
   Add `task ac <slug> --check <n>` / `--uncheck <n>`, with `task ac <slug>
   --list` printing the criteria numbered. Index-based matching is the robust
   form (substring matching is fiddly): `--list`, then `--check 3`.
3. **`task path` / `task info --json` — file location + cheap metadata.** The
   agent fell back to `find` every time it needed body content the CLI can't
   reach (items 1–2), and globbing on a slug substring is fragile because files
   are `<id>-<slug>.md`. `task path <slug>` prints just the absolute path
   (pipe-friendly). `task info <slug> --json` returns `{path, status, epic,
   bucket, ac:{checked, total}}` — also the token-cheap metadata read for item 4.
   (This is the `task path` / `task info` idea floated earlier but never filed —
   confirmed absent from planning.)
4. **Section-scoped / metadata-only reads.** `task show` is token-heavy when the
   agent only wants the frontmatter, the ACs, or the Progress Log — it ends up
   piping to `grep`/`sed`/`head`. Add `task show <slug> --section <name>` (e.g.
   `acceptance`, `progress`) and `--frontmatter-only`. Complements the terse
   `task list -o table -c` steering already in the schema guidance.

## Already ships — advertise, don't build

The agent also asked for two things **that already exist** — a discoverability
gap, not a missing feature. The fix is steering (help/README/docs), not code:

- **Non-interactive whole-body replace** → `task set --body`/`--body-file -`
  already does exactly this (prior batch; its own atomic write, mutually
  exclusive with field flags).
- **`task new --body-file -`** → already exists, and already collapses
  new+append into one call. Only "create *and defer* in one call" / an initial
  `--status` beyond `--next`/`--start` is genuinely absent, and that's minor.

Make sure `--help`, the README "agent use" section, and the generated `docs/cli`
reference actually surface `task set --body-file -` and `task new --body-file -`
so the next agent finds them without dogfooding-by-discovery.

## Acceptance criteria

- [x] `task log` (item 1) **spun off** to its own task —
      [task-log-append-a-dated-progress-log-entry](6fpnn6zk157b-task-log-append-a-dated-progress-log-entry.md)
      — because it is blocked on the canonical progress-section decision, not just
      unstarted. Tracked and designed there rather than holding this batch open.
- [x] `task ac <slug> --list` numbers the criteria; `--check <n>`/`--uncheck <n>`
      flip the box atomically, preserving surrounding body; `--json` supported.
- [x] `task path <slug>` prints the absolute file path and nothing else.
- [x] `task info <slug> --json` returns `{path, status, epic,
      ac:{checked, total}}` (+ `id`/`slug`), validating against the published JSON
      schema. (No `bucket` — tasks have no bucket under the flat id-led layout;
      `status` is authoritative.)
- [x] `task show <slug> --section <name>` and `--frontmatter-only` project a
      single section without the rest of the body.
- [x] `--help`, README, and the generated `docs/cli` reference advertise the
      already-shipped `task set --body-file -` and `task new --body-file -`. (The
      `schema` command is the data-model contract — statuses/fields/exit codes —
      not a command index, so it is *not* one of these surfaces.)
- [x] Suite + lint green; docs regenerated (`docgen`).

## Design notes / risks

- **Structure-awareness is the shared engine.** Items 1, 2, and 4 all need a
  markdown-section model of the body (locate `## Progress Log`, enumerate the AC
  list, slice a named section). Build it once — a small body-structure helper —
  and route `log`/`ac`/`show --section` through it. This is the same
  markdown-structure-aware muscle the rename/link-lint work grew
  ([make-rename-cascade-dangler-lint-markdown-structure-aware](6fka8khkb3jv-make-rename-cascade-dangler-lint-markdown-structure-aware.md)).
- **Reuse `EditBody`.** `log`/`ac` are structural body rewrites — route them
  through the existing `FS.EditBody` / surgical `yaml.Node` path so unknown
  frontmatter, comments, key order, and the parse-before-write +
  compare-and-swap discipline all survive (the same guarantees `task append` /
  `task set --body` already carry).
- **`ac:{checked, total}`** in `task info` needs the same AC enumerator as
  `task ac --list` — one parser, two consumers.

## Related

- Sequel to [agent-facing-cli-ergonomics-batch](6fbj87000anh-agent-facing-cli-ergonomics-batch.md).
- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md).
- Terse-output steering: [steer-agents-to-the-terse-output-path-in-schema-and-guidance](6fes83r0137q-steer-agents-to-the-terse-output-path-in-schema-and-guidance.md).
- Touches `internal/cli/`, `internal/core/` (body-structure helper + `EditBody`),
  `internal/cli/render/`, `docs/cli`, `README.md`.

## Progress (2026-07-15)

Shipped the **read-only cluster** (items 3, 4 + the `task info` metadata read) —
no writes, no convention forks:

- **`task path <slug>`** — prints the absolute file path (`--json` wraps it as
  `{schema_version,path}`). Replaces `find`/glob on the id-led filename.
- **`task info <slug>`** — token-cheap metadata: `{path,status,epic,ac:{checked,
  total}}` (+ `id`/`slug`), no body. New `task_info` envelope (schema 1.26 → 1.27,
  alongside a small `path` envelope). Path is absolute (`filepath.Abs`) so it
  resolves from anywhere.
- **`task show --section <name>` / `--frontmatter-only`** — narrow the body to one
  named section (heading substring match) or drop it entirely; reuses the existing
  `task_show` envelope (body just shrinks). A missing section is a clean
  `ErrNotFound` (exit 10).

Shared engine: a pure, fence-aware body-structure helper in `internal/domain`
(`Section` + `CountAcceptanceCriteria`) — the same muscle `task ac`/`task log`
will reuse. README "agent use" now advertises the reads plus the already-shipped
`task set --body-file -` / `task new --body-file -` (the discoverability half).
Full suite + lint green; golden fixtures + CLI docs regenerated.

**Remaining:**
- `task ac` (item 2) — the acceptance-criteria checkbox writes. Independent,
  index-based per the spec, reuses the new AC enumerator. The natural fast-follow.
  (These very checkboxes were flipped by hand — the exact gap `task ac` closes.)
- `task log` (item 1) — parked pending a decision on the canonical Progress-section
  shape (single `## Progress Log` with dated bullets, as the agent's repo uses, vs
  this repo's de-facto per-entry `## Progress (date)` headings). Needs the shape
  settled before building, or `task log` would hardcode a format that fights the
  existing corpus.

## Progress (2026-07-15) — adversarial-review fixes

An independent adversarial review found four issues in the read-only cluster;
each was verified against the code and the real ones fixed:

- **CRLF (parser blind to `\r\n`)** — the ATX heading regex is anchored `[ \t]*$`,
  so a trailing `\r` made every heading vanish (→ empty sections, 0/0 AC tally) on
  an autocrlf checkout or a CRLF `--body-file`. Fixed by folding CRLF/CR→LF at
  parse entry (`normalizeNewlines`). Proven end-to-end: a piped CRLF body now
  reports the right tally.
- **Nested code fences** — the old `inFence = !inFence` toggle closed on a shorter
  inner fence or an info-stringed line (` ```go `), leaking fenced `##`/`- [ ]`
  into structure scanning. Replaced with a CommonMark-ish `fenceScanner` that
  tracks the opening fence's char + length and closes only on a same-or-longer
  bare fence of the same char.
- **Golden coverage** — `task info`/`task path` weren't in the byte-golden harness
  because their output carries an absolute path (the machine checkout root). Added
  a path-redactor (`<ROOT>` placeholder) so both are now pinned portably.
- **Trailing blank line** — `--frontmatter-only` left a stray blank line under the
  fields (`\n%s` with an empty body); guarded in `TaskShowHuman`.

Regression tests added (CRLF heading + AC tally, nested-fence non-leak,
frontmatter-only no-trailing-blank). Two review recommendations were **declined
with reason**: "Suite + lint green" stays checked (literally true — the suite
passes), and the `schema`-advertises claim was **reworded** rather than unchecked,
to name the surfaces that actually advertise the `--body-file` variants (`--help`,
README, `docs/cli`) — the `schema` command is the data contract, not a command
index. Full suite + lint green; golden fixtures + comment map + CLI docs
regenerated.

## Progress (2026-07-16) — epic/audit parity + Q1 cleanups

Extended the read-only cluster to epics/audits (high-value parity) and landed the
two Q1 follow-ups:

- **`epic path` + `audit path`** — the file locator, **parse-free**
  (`Svc.EpicPath`/`AuditPath` over the store's existing resolvers).
- **`audit info` (+`--json`)** — token-cheap audit metadata: `{path, bucket,
  findings:{total, open, in_progress, done, dropped}}` — the audit analogue of the
  AC tally, read straight off the tally `parseAudit` already populates. New
  `audit_info` envelope (schema 1.27 → 1.28). **Skipped `epic info`** (would
  duplicate `epic list -c`).
- **`--section` / `--frontmatter-only` on `epic show` + `audit show`** — the same
  body-scope flags as `task show`, now via a shared `narrowBody` +
  `addBodyScopeFlags` helper (three shows, one implementation). Highest value on
  audits (long bodies: `audit show --section findings`).

Q1 cleanups:
- **Shared field-writer** — `render.fieldPrinter` now backs the metadata blocks of
  `task show`, `task info`, and `audit info` (dropped the duplicated closures).
- **Parse-free `task path`** — `task path` (and epic/audit path) resolve WITHOUT
  parsing, so they work on a file with broken frontmatter — exactly when you need
  the path to open and fix it. Regression-tested (path resolves a broken file;
  `task show` still fails on it).

Tests: epic path / show --frontmatter-only / section-not-found; audit path / info
tally / show --section; parse-free broken-file path. Golden: `audit_info_json`
(path-redacted). docgen + comment map + README updated. Full suite + lint green.

## Progress (2026-07-16) — adversarial-review round 2 (epic/audit diff)

External review verdict was **GO with caveats**; verified each finding and fixed the
real ones:

- **[major → hygiene] golden portability** — the path redactor matched only the raw
  (single-backslash) root, so path-bearing goldens (`task/audit info`, `*/path`) would
  fail on Windows (where `json.Marshal` escapes separators as `\\`). Made the redactor
  fold `\\`→`/` and match the forward-slash root, so goldens are one form on every
  platform. (CI is Ubuntu-only, so this was latent, not a live break — fixed anyway.)
- **[minor] TTY path truncation** — the shared `fieldPrinter` had added terminal-width
  truncation to `task info`/`audit info`, cutting off the very path you run the command
  to copy. Gave `fieldPrinter` a `fit` flag: browse views (`* show`) still truncate long
  values; the `info` reads print full. Render-level regression tests at width 40.
- **[minor] epic/audit `--frontmatter-only` trailing blank line** — `EpicShowHuman`/
  `AuditShowHuman` printed `\n%s` unconditionally (the same nit fixed for tasks last
  round, not carried over). Guarded both, plus the CLI-side `RenderBody` call. Asserted.
- **[nit] golden symmetry** — added `epic_path_json` / `audit_path_json` cases (task
  path already had one).

Reviewer PASSED lenses 1–3, 5, 6, 8 (narrowing-under-`--json`, parse-free resolution,
tally trust, port churn, section semantics, completion claims) and un-checked no boxes.
Full suite + lint green; goldens + docs regenerated.

## Progress (2026-07-16) — `task ac` shipped (item 2)

The acceptance-criteria checkbox CLI — the fast-follow — is done, closing item 2
(checked off with the command itself):

- **`task ac <slug>`** (default / `--list`) — numbered acceptance checklist;
  `--json` returns the new `acceptance` envelope `[{index,checked,text}]` (schema
  1.28 → 1.29).
- **`task ac <slug> --check <n>` / `--uncheck <n>`** — flip one criterion by
  1-based index (substring matching deliberately not offered). The flip rewrites
  only that checkbox's `[ ]`/`[x]` char and routes through the atomic,
  frontmatter-preserving `EditBody` path; `--json` returns the `task_mutation`
  envelope. Idempotent: flipping to the current state is a no-op with no write (no
  spurious `updated_at` bump); `--dry-run` previews; out-of-range / no-AC-section →
  ErrValidation (exit 11).

Reuses the fence-aware AC enumerator (`scanAcceptanceCheckboxes`) that already
backs `task info`'s tally — one parser, three consumers (count, list, flip).
Verified surgical end-to-end: a `--check`→`--uncheck` round-trip restores the file
byte-for-byte. Tests: domain (list/flip/uncheck/idempotent/out-of-range/no-section)
+ CLI (list, list --json, flip-file, frontmatter-preserved, dry-run, no-op,
mutual-exclusion, mutation envelope). Golden `task_acceptance_json` added (the
fixture task gained an acceptance section, so `task_info`'s tally is now a real
1/2). docgen + comment map + README updated. Full suite + lint green.

**Remaining:** only `task log` (item 1) — still parked pending the canonical
Progress-section shape decision.

## Progress (2026-07-16) — misconfiguration guard for acceptance criteria

Raised in review: the `ac:` tally / `task ac` key off "a heading containing 'acceptance'
+ well-formed checkboxes", so a misconfigured task made them silently lie. Added a lint
guard so misconfiguration is loud, not a false positive:

- **`lint` now flags** (a) a botched checkbox in the acceptance section that the scanner
  silently drops (`[]`, `[ x]`, `[  ]` — blanks/x/X but not the canonical
  `[ ]`/`[x]`/`[X]`), and (b) more than one acceptance section (only the first is used).
- **Deliberately conservative** — `[1]` citations, `[-]` partial markers, and
  `[text](url)` links are NOT flagged (a false positive would break lint-clean on legit
  content). Verified the real corpus stays lint-clean.
- **Architecture:** `domain.LintAcceptanceCriteria(body)` (fence-aware, reuses the AC
  scanner's model) wired through a new `ListTasksWithBodies` store scan — the task twin
  of `ListAuditsWithFindings`, so lint reads every body ONCE (no O(N²) re-resolve). The
  write path already guarded (out-of-range / no-section → exit 11); this closes the read
  path.

Tests: domain (malformed variants · no-false-positives · multiple sections · fence-aware
· out-of-section), store (`ListTasksWithBodies` carries bodies), core (Lint flags it via
the body scan), CLI (`lint --json` surfaces it end-to-end). Full suite + lint green.

## Progress (2026-07-16) — batch complete; `task log` spun off

Closing this batch. Three of the four items shipped and are on branch
`feat/various-read-modalities` (PR #106): `task ac` (item 2), `task path`/`info`
(item 3), and section reads (item 4) — plus the epic/audit parity and the
acceptance-criteria lint guard that grew out of review.

Item 1 (`task log`) is **spun off** to
[task-log-append-a-dated-progress-log-entry](6fpnn6zk157b-task-log-append-a-dated-progress-log-entry.md):
it's blocked on the canonical progress-section-shape decision (single `## Progress
Log` + dated bullets vs this repo's per-entry `## Progress (date)` headings), so it
gets its own home with the decision + design captured, rather than holding this batch
open. This batch's responsibility — the structure-aware read/write surface — is
discharged.
