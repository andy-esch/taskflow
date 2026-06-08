---
date: 2026-06-06
topic: tskflwctl — formal command hierarchy + purpose spec (Phase 0.5)
purpose: The build-to command spec for the Go CLI. Full noun-verb tree with per-command purpose, flags, and output. Python pm is a starting point, not the whole contract.
status: in-progress
related_tasks:
  - phase-0.5-formal-tskflwctl-command-hierarchy-purpose-spec.md
  - go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness.md
---

# tskflwctl command spec (Phase 0.5)

The formal `tskflwctl {noun} {verb}` surface to build the Go CLI to. Derived
from the Python `pm` command inventory but **re-specified** for the
noun-verb design — the Python tool/tests are a starting point, not the
whole contract (epic 17 Phase 0.5). First draft 2026-06-06.

> Naming: binary **`tskflwctl`** (taskflow repo). **Fully explicit
> `tskflwctl <noun> <verb>` — no flat top-level aliases and no lifecycle
> "sugar" verbs** (decided 2026-06-06: explicit over convenient; shortcuts
> can be added later if missed).

## Global flags (persistent, root)

| Flag | Purpose |
|---|---|
| `--json` | Machine output on stdout; no ANSI. Errors → structured envelope on stderr + semantic exit code. |
| `-C, --chdir <dir>` | Anchor to the planning repo regardless of cwd. |
| `--config <file>` | Override config discovery (default `.tskflwctl.toml`). |
| `--dry-run` | (mutating cmds) Don't write; in `--json`, emit the would-be file changes. |
| `--yes` / `--force` | Skip interactive confirmation (required for interactive cmds in non-TTY). |
| `--no-color` | Disable ANSI (also auto-off when piped / non-TTY). |
| `-v, --verbose` | Diagnostic logging to stderr. |

Convention: **semantic non-zero exit codes on error** (see Agent-interaction
contract — fixes the Python exit-0 quirk); all human output via injected
writer; `--json` shapes are stable (carry `schema_version`).

**`--json` schema versioning (semver, internal):** every JSON payload
carries a top-level `"schema_version": "MAJOR.MINOR"`. **Adding** a field →
minor bump; **renaming/removing** a field → major bump. Lets a consumer (or
the future agent layer) assert compatibility. Internal dev concern, but
cheap to set from day one.

## Planning repo, tracked repos, and projects (terminology — settled 2026-06-06)

Three distinct things; "project" was overloaded:
- **planning repo** = the **one** repo holding *all* planning data for a
  product/purpose (`tasks/ epics/ projects/ adrs/ audits/`). Either
  standalone (`desirelines-planning`) **or** a subdir inside a single code
  repo (FuturisticCodeThing plans in-repo). **One planning repo per product
  — no cross-product / multi-repo registry** (decided: "adding multiple
  feels like chaos"). The tool anchors by walking up for `.tskflwctl.toml`
  (git-style); no `source`/`context` noun needed.
- **tracked repos** = the **1..N code repos a planning repo plans *for***
  (`tracked_repos` config) — this is "track multiple repos." Metadata only:
  planning data stays in the one repo; **tasks cite code by prose path —
  no per-task `repo:` field** (decided).
  - *Desirelines:* `desirelines-planning` tracks `desirelines` +
    `desirelines-deploy`.
  - *FuturisticCodeThing:* plans in-repo, tracks just itself.
- **project** = a cross-cutting *initiative within the planning repo* (the
  `projects/` group) — **the only meaning of "project."**

Config (`.tskflwctl.toml`, written by `init`):
```toml
taskflow_root = "."                 # or "planning/"
tracked_repos = ["../desirelines", "../desirelines-deploy"]
```

| Command | Purpose | Safety |
|---|---|---|
| `init` | Make the current repo/subdir a planning repo: scaffold the tree + write config (incl. tracked repos). Interactive wizard / `--yes` | mutating |
| `track <path>` / `untrack <path>` | Add/remove a tracked code repo | mutating |

## `tskflwctl task`

The day-to-day noun. Lifecycle = directory; `status` frontmatter must match.
**Explicit transition verbs** (`start`/`promote`/`demote`/`complete`/`defer`/
`deprecate`) are the first-class surface, over one internal transition
engine; `move <t> <status>` remains as the generic escape hatch.
`<status>` ∈ {next-up, ready-to-start, in-progress, completed, deprecated,
deferred}.

| Command | Purpose | Key flags / args |
|---|---|---|
| `list` | Filtered task list for triage | `--status --tag --epic --project --package --due --waiting --oneline --limit --json` |
| `show <t>` | One task's metadata + body | `--json` |
| `new "<title>"` | Scaffold a task (frontmatter + handoff body skeleton) | `--epic(req) --tier --priority --autonomy --tags --description --status --body --edit` |
| `set <t> [--<field> val ...] [--set k=v ...]` | Set one or more frontmatter fields in a **single atomic write** (validated). Typed flags for known fields; repeatable `--set k=v` for custom/extension fields | (batched — no per-field re-parse) |
| `lint [<t>...]` | Validate frontmatter (+handoff advisory) | `--fix --errors-only --include-archived --no-handoff-check` |
| `start <t>...` | → in-progress | |
| `promote <t>...` | → next-up | |
| `demote <t>...` | → ready-to-start | |
| `complete <t>...` | → completed | `--recommend` |
| `defer <t>...` | → deferred | `--reason` (req) `--revisit` |
| `deprecate <t>...` | → deprecated | `--reason` |
| `move <t>... <status>` | Generic transition (escape hatch, e.g. reopen completed→ready-to-start); validates legal + idempotent | status-appropriate flags |
| `touch <t>...` | bump `updated_at` | |
| `rename` | normalize filenames | `--all --dry-run` |

**Lifecycle verbs — REVERSED 2026-06-06 (agent-operator lens).** Bring back
**explicit transition verbs** (`start`/`promote`/`demote`/`complete`/`defer`/
`deprecate`) as first-class subcommands over one internal `move` engine. For
the *primary user (LLM agents)*, a single `move <status>` invites **enum
hallucination** (`done`/`finished` vs `completed`) and **conditional-flag
validation failures** (a tool schema can't say "`--reason` required only for
deferred"); explicit verbs remove both (the verb *is* the intent; per-verb
required flags). The verbs aren't "sugar" — they're the explicit, safe
surface; `move`'s enum is the real hazard. `move` stays as the generic
escape hatch. All transitions are idempotent (re-applying the current status
exits 0). `task new --status next-up` = "create directly in next-up".
*(Token-efficiency was not a reason — more subcommands = more tool tokens, not
fewer; the verbs win on correctness.)*

## `tskflwctl audit`

**Two state levels — don't conflate** (decided 2026-06-06):
- **Audit-level = bucket/dir:** `open` / `closed` / `deferred` (mirrors
  tasks — the directory is the source of truth). A whole audit moves
  between these. `deferred` keeps an audit we don't want to act on now but
  won't discard.
- **Finding-level = the `**Status:**` line** in the body:
  `open · in-progress · fixed · landed · deferred · superseded · wontfix`.
  An audit has many findings, so it's never globally "fixed"/"landed" —
  `closed`/`deferred` cover the audit; per-finding fates live in the Status
  lines + the Closeout block (inferred from the text, not one frontmatter
  flag).

Markdown is the source of truth; the CLI parses findings + Status lines +
candidate-list.

| Command | Purpose |
|---|---|
| `list [--open\|--closed\|--deferred\|--all]` | Audits + finding counts by severity + open count |
| `show <slug>` | Print an audit body |
| `new --routine <r> --area <slug>` | Emit the canonical skeleton into `open/` (the generator routines call) |
| `close <slug> [--message]` | Append Closeout, move → `closed/` |
| `reopen <slug>` | Move → `open/` |
| `defer <slug> [--reason]` | Move the whole audit → `deferred/` |
| `status <slug> <code> <value> [--pr N] [--note]` | Set one finding's Status; `--pr` stamps `fixed (PR #N)`; syncs its candidate mark |
| `followup <slug> <code> [--new "title" --epic …]` | Create+link a follow-up task; mark candidate ⏳/⚠️ (does more than `status`) |
| `sync <slug>` | Re-derive candidate-list symbols from finding Status lines |
| `lint [<slug>]` | Findings well-formed; candidate-list ↔ Status consistent; body not corrupt |
| `noop --routine <r> --area <slug>` | Append a clean-day entry to `no-op-log.md` (create if missing) |
| `findings [--status --severity]` | Cross-audit finding search |
| `stats` | Counts by severity/status (feeds the dashboard line) |

Resolved: **explicit, no sugar** here too — finding `fixed`/`landed`/`defer`
fold into `status <code> <value>` (`--pr` stamps `fixed`); `followup` stays
(it creates a task); audit-level `defer` adds the third bucket. **Open
gap — audit frontmatter:** today it carries `status: completed` meaning
"audit *written*", which now collides with open/closed/deferred. Recommend
the **directory is the audit-level state** (like tasks) and drop/rename the
production marker. Touches the bucket-audits task + HOWTO + routine generator.

## `tskflwctl epic`

| Command | Purpose |
|---|---|
| `list` | Epics + task rollup + descriptions (`--json`) |
| `show <id>` | Epic details + its tasks (`--json`) |
| `new "<title>"` | Scaffold an epic (auto-numbered) `--description(req) --tags --status --priority --owner` |
| `lint` | Validate epic frontmatter `--fix` |
| `orphans` | Tasks referencing unknown epics `--include-archived --json` |

## `tskflwctl adr`

| Command | Purpose |
|---|---|
| `list` | List ADRs |
| `new "<title>"` | Create an ADR (writes to the impl repo's ADR dir) |

## Cross-cutting (top-level, not a noun group)

| Command | Purpose |
|---|---|
| *(bare)* / `index` | Dashboard: ready-to-revisit, hot, waiting, quick wins, **🔬 audits** |
| `schema [--type task\|epic\|project\|cli\|all] [--json]` | Print the frontmatter contract; `--type cli` emits the command tree for agent introspection |
| `recommend` | Generate RECOMMENDATIONS.md |
| `tags [--json]` | Tag vocabulary |
| `version` | Build version (ldflags) |
| `completion <shell>` | cobra-generated shell completion |

(`init`/`register`/`source …` are in **Sources & multi-product config**
above. A source's planning data is single-repo; the product's *code* may
span repos. `schema --type project` is new per the project design.)

## `tskflwctl project`

Projects = **optional, cross-cutting** initiatives spanning domains (and
thus multiple epics); a task belongs to **zero or more** projects.
First-class like epics: `projects/<slug>.md` (slug-only) + a validated
`projects:` list on tasks. Frontmatter: `status` ∈ {unstarted, in-progress,
complete, abandoned}, nullable `created`/`ended`, `description`, `tags` (no
owner, no related_epics). Design:
`research/2026-06-06-project-concept-cross-cutting-initiatives.md`.

| Command | Purpose |
|---|---|
| `list` | Projects + cross-epic rollup + progress |
| `show <slug>` | A project's tasks **grouped by epic** + how many remain open |
| `new "<title>"` | Scaffold `projects/<slug>.md` |
| `set <slug> <field> <val>` | Edit project frontmatter (`status complete` stamps `ended`) |
| `lint` | Validate project frontmatter; enforce slug uniqueness vs epics |
| `add <slug> <task>...` / `rm <slug> <task>...` | Edit a task's `projects:` membership |

No auto-close: `project show` / `task list --project <slug> --status …`
surface what's left; closing is an explicit `set status complete`.
(Supersedes pm's tag-based `project` — 0 tasks use it, so no migration.)

## Agent-interaction contract (adopted 2026-06-06 from the agent-operator audit)

Make `tskflwctl` machine-operable, not just human-runnable:

- **Command safety tags:** every command carries cobra
  `Annotations{"safety": "read-only"|"mutating"}`.
  - *Read-only:* all `list`/`show`/`findings`/`stats`, plus `schema`,
    `index`, `tags`, `version`, `completion`, `lint` *without* `--fix`.
  - *Mutating:* `new`, `set`, the transition verbs + `move`, `touch`,
    `rename`, `recommend`, `lint --fix`, every `audit` write, `project
    new/add/rm`, `init`, `track`/`untrack`.
- **Semantic exit codes** (route without parsing text):
  `0` success / idempotent no-op · `10` ERR_NOT_FOUND · `11` ERR_VALIDATION ·
  `12` ERR_INVALID_TRANSITION · `13` ERR_AMBIGUOUS_TARGET ·
  `14` ERR_LOCK_CONFLICT. (Reserve a block for more.)
- **Structured error envelope** — in `--json` mode, errors go to **stderr**
  as `{ "error_code", "message", "candidates"? }` (e.g. the match list on
  ERR_AMBIGUOUS_TARGET).
- **Fail-fast on ambiguity:** a fuzzy `<t>` matching >1 file never silently
  first-matches — exit `13` + candidate list. (pm already detects this;
  make it structured.)
- **Batch ops** (`complete a b c…`): emit a structured per-item
  success/failure report; non-zero exit if any item failed (no silent
  partials). True multi-file atomicity isn't possible, so report honestly.
- **Headless safety:** interactive commands (`init` wizard, `rename --all`)
  detect non-TTY (`isatty`) and **fail-fast unless `--yes`/`--force`** —
  never hang an agent loop on a prompt.
- **`--dry-run`** (global, mutating commands): in `--json` mode, output the
  exact would-be file create/modify payloads — preview before write.
- **`schema --type cli --json`:** emit the command tree (commands, flags,
  enums, safety tags) so an agent introspects syntax instead of scraping
  `--help`. Feeds a future MCP layer.

## Deliberate departures from the Python `pm`

1. **Non-zero exit on error** (not exit-0-with-message).
2. **`--json` is a global persistent flag** with stable, documented schemas
   (not per-command ad hoc).
3. **Validation at write** for every constrained field (pm only added this
   recently in `tighten-pm-cli-ergonomics`).
4. **Audit finding states incl. `landed`** (implemented-pending-commit),
   set via `audit status` (no Python equivalent); audit-level `deferred`
   bucket alongside open/closed.
5. **Explicit noun-verb, no flat aliases; explicit transition verbs** over
   an internal `move` engine (resurrected 2026-06-06 for agent-correctness —
   no enum to hallucinate, per-verb required flags). Python is flat verbs.
9. **Built for agents:** semantic exit codes, structured JSON error
   envelopes, `--dry-run`, `schema --type cli`, TTY-aware fail-fast.
6. **Surgical, preserving frontmatter writes** (`yaml.v3` Node — keeps
   unknown fields + comments + order; normalize only on `lint --fix`).
7. **Atomic writes** (temp + fsync + rename) — pm's `write_text` isn't atomic.
8. **`set` takes flags** (`--field`/`--set k=v`) → batched multi-field edits
   in one write (pm is one positional field per invocation).

## Resolved 2026-06-06

- ✅ Audit verbs: fold finding `fixed`/`landed`/`defer` into `status`;
  audit-level `open/closed/deferred` buckets; `followup` stays.
- ✅ `task new --status <s>` replaces a `--next` flag.
- ✅ `--json` carries a semver `schema_version` (minor=add, major=rename/remove).
- ✅ Command safety tagged via cobra `Annotations` (read-only vs mutating).
- ✅ `project` is a first-class group (own dir + `projects:` list on tasks).

## Still open / needs scoping

- **Audit frontmatter alignment** — `status: completed` (production marker)
  collides with the open/closed/deferred bucket; make the directory the
  audit-level state. (Bucket-audits task + HOWTO + routine generator.)
- **`project` design gaps** — see
  `research/2026-06-06-project-concept-cross-cutting-initiatives.md`.
- **Where each new concept is built first** — prototype in Python `pm`
  (cheap iteration) vs straight into `tskflwctl`? (Per-concept call.)

## References

- `research/2026-06-06-pm-cli-architecture-and-go-port.md` (§2 hierarchy).
- `research/2026-06-06-go-cli-foundation-architecture.md` (foundation).
- `bin/pm` + `tests/test_pm.py` (the prototype inventory/spec).
- `tasks/ready-to-start/bucket-audits-into-openclosed-and-ship-pm-audit-cli.md`
  (the `audit` surface).
