---
schema: 1
id: 6g96h83t3w26
bucket: closed
area: readable-thread-graph-nodes-implementation-antigravity
date: "2026-09-11"
updated_at: "2026-09-12"
---
# Audit: Readable Thread graph nodes implementation — antigravity — 2026-09-11

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Adversarially review the implementation of task
`planning/tasks/6g95fc3aj4ye-make-thread-graph-nodes-human-readable-at-review-scale.md`.
Assume the feature is plausible but unfinished until the evidence proves otherwise. Look for
systemic contract mistakes, not only local formatting bugs. The unrelated untracked
`planning/tasks/6g95m4eyvf4k-mine-the-web-mockup-directions-for-the-tui-data-organization-and-a-more-80s.md`
belongs to parallel work and must not be changed.

## Review target

Review the complete working delta from `main` on branch `feat/readable-thread-graph-nodes`, including:

- body-derived `domain.Task.Title` and the fence-aware `domain.FirstH1` parser;
- filesystem task reads, `core.ThreadGraphNode.Title`, and slug fallback for bodyless adapters;
- optional wire `title`, schema version 1.64, reflected schema, goldens, and historical compatibility;
- compact Mermaid/DOT node labels, deterministic Unicode/word-boundary truncation, role wording,
  visible legend guidance, and `RenderOptions`;
- the CLI `thread graph --details` behavior and its exclusion from `--json`;
- README, architecture, ADR, generated CLI documentation, tests, and planning evidence.

Do not limit the review to files named by the diff. Trace every producer and consumer of the changed
types and commands. Build a consumer inventory before reaching any verdict.

## Intended contract to challenge

The authored Markdown H1 is optional adapter-neutral presentation data, not stable identity and not
duplicated into frontmatter. Local reads derive it during the existing file parse; remote/bodyless
adapters may omit it. `ThreadGraphNode.Label` and wire `label` retain the stable slug-era meaning,
while optional wire `title` is additive. Renderers prefer title, fall back to a humanized label, keep
status/role and stable ID visible, omit descriptions by default, and add bounded descriptions only
under `--details`. Render options do not change topology, health, roles, or JSON. Completed non-member
context is visibly called an external prerequisite, not an active gate. All output remains bounded,
deterministic, injection-safe, and accessible without relying on color.

Challenge whether this split actually holds across filesystem, portable, core, wire, CLI, TUI, and
future served consumers. Treat schema 1.64 as a public compatibility claim, not bookkeeping.

## Mandatory evidence floor

1. Inventory every constructor, clone, mapper, serializer, renderer, fixture, and consumer touching
   `domain.Task`, `ThreadGraphNode`, `Label`, `Title`, `Description`, `Mermaid`, or `DOT`. State which
   consumers receive a real title and which exercise fallback.
2. Inspect real compact and `--details` Mermaid and DOT output from the repository Thread. Compare
   visible semantics, graph edge counts, health metadata, stable IDs, and output size. Do not infer
   usability from unit helpers alone.
3. Verify the 1.63-to-1.64 bump classification, optional-field schema, all JSON goldens, old Thread
   compatibility fixtures, and preservation of the old `label` and `description` meanings.
4. Exercise the read path with an exact H1, missing H1, multiple H1s, CRLF, fenced fake H1, long and
   Unicode titles, duplicate titles, empty labels, hostile controls, and malformed task sources.
5. Inspect mutation freshness: task creation, rename, body edit, lifecycle/frontmatter writes, and
   graph rereads must not persist or stale-cache the derived title. Execute at least one relevant
   mutation probe and prove its outcome from a fresh projection.
6. Run the focused suites and at least full tests, full race tests, vet, lint, tidy, planning/audit
   lint, generated docs/schema checks, and diff checks. Explain any environment-only exception.

## Required hostile angles

- Try to make an H1 inside nested or unterminated fences become the title; check leading indentation,
  inline Markdown, closing hashes, bidi/zero-width controls, combining characters, and invalid UTF-8.
- Look for byte/rune/grapheme and off-by-one mistakes, mid-word truncation, unbounded legend text,
  unstable whitespace normalization, or a compact label that silently hides ambiguity.
- Determine whether a user can confuse Thread membership, external prerequisite context, current
  blocking state, lifecycle state, or stable identity after the wording changes.
- Prove duplicate human titles remain unambiguous in rendered output and machine output.
- Attack Mermaid and DOT syntax with title and description payloads. Confirm detailed mode does not
  bypass the existing control-character or directive defenses.
- Check whether defaulting `Mermaid` and `DOT` to compact output silently breaks any caller that
  depended on descriptions, and whether the options API is reusable without CLI coupling.
- Check portable adapters and zero-value tasks for accidental filesystem/body requirements.
- Check that `--details --json`, explicit `--format --json`, and unsupported formats fail before any
  mutation or partial output. Although the command is read-only, trace the mutation boundary and
  verify no new presentation field leaks into task writes.
- Look for schema/version drift, required-versus-optional mismatch, stale generated evidence, or a
  compatibility test that passes only because it ignores the new branch.
- On a second pass, play devil's advocate about the architecture: is H1-derived title a useful shared
  semantic, or a filesystem convenience smuggled into core? Demonstrate the failure mode behind any
  objection and propose the narrowest stronger variant.

## Validation and restoration

Run probes only in the mandatory isolated workspace. You may add temporary tests or fixtures there,
but restore them before delivery. For each proposed regression test, first make it fail against the
challenged implementation or otherwise show the exact defect it detects. Do not update generated
files merely to make checks green; verify their source contract first. The only transferred change
must be your assigned audit document.

## Deliverable

Update only the assigned audit. Preserve this brief. Record findings using exactly
`#### H1. Title · **Status:** open`, `#### M1...`, or `#### L1...`; number independently by severity.
Each finding must include concrete evidence, impact, and the smallest sound remediation. If an issue
belongs outside this task, still record it and explain the appropriate follow-up boundary. If no
finding survives hostile verification, explicitly state what was tested and why each suspected
failure was disproved. Leave all findings open for implementation-owner triage.

## Reviewer report

### Isolation attestation and reviewed baseline

The audit was executed entirely within an independent, isolated review workspace created via `./scripts/isolated-review-workspace.sh create`. No inspection, mutation, or test commands were executed in the shared checkout root.

- **Sandbox workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.eX6i7r`
- **Resolved Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.eX6i7r/.git`
- **Baseline commit:** `074afa498ce557e8da30c14fe5fdc4a5be0e2938` (`chore: capture isolated review baseline`)
- **Review target branch / base:** `feat/readable-thread-graph-nodes` evaluated against base commit `f9047f9` (`Merge pull request #228 from andy-esch/feat/thread-graph-legends`)
- **Captured source blob (`$AUDIT_REL`):** `71aaacf8cbd7b1f5f01940bb94c91565fe85a80f`
- **Captured source fingerprint:** `9f8173bf0ade1dff09059498d81b0b26fd7c380c`
- **Assigned deliverable:** `planning/audits/6g96h83t3w26-2026-09-11-readable-thread-graph-nodes-implementation-antigravity.md`
- **Pre-transfer verification status:** Verified independent clone, zero reviewer commits, zero staged changes, clean working directory, and identical source deliverable blob before transfer.

---

### Concise verdict and confidence

- **Verdict:** **Findings Open** (2 Medium, 2 Low findings open for implementation-owner triage).
- **Confidence:** **High (95%)**.

The implementation successfully establishes a scannable, human-readable hierarchy for Mermaid and DOT Thread graphs without losing stable identity or compromising deterministic layout. Body-derived `domain.Task.Title` via fence-aware `domain.FirstH1`, slug fallback in `ThreadGraphNode.Label`, the additive wire schema bump to `1.64`, and mutual exclusivity between renderer flags and `--json` are architecturally sound and thoroughly covered by existing tests.

However, adversarial testing and hostile mutation probes revealed four open findings:
1. **M1**: `compactText` compares a byte offset (`lastSpace`) against a rune threshold (`(limit-1)*2/3`), causing severe and premature word-boundary truncation for non-ASCII titles and descriptions.
2. **M2**: `MermaidWithOptions` and `DOTWithOptions` statically hardcode the compact mode legend banner (`Legend · compact labels · add --details for descriptions`), producing contradictory guidance when `--details` is explicitly enabled.
3. **L1**: `nodeLabel` unconditionally forces the first rune of human-authored titles to uppercase, mutating intentional author casing (e.g. brand names, CLI tools, identifiers like `iOS`, `kubectl`, `eBPF`).
4. **L2**: `newThreadGraphCmd` delays `--format` validation until after executing service disk reads and graph projection construction.

---

### Consumer inventory

Every constructor, mapper, serializer, renderer, fixture, and consumer touching `domain.Task`, `ThreadGraphNode`, `Label`, `Title`, `Description`, `Mermaid`, and `DOT` was inventoried across the codebase:

| Component / Symbol | File Location | Consumer / Producer Role | Display Title vs. Slug Fallback |
| :--- | :--- | :--- | :--- |
| `domain.FirstH1` | [`internal/domain/body.go:191-204`](file:///internal/domain/body.go#L191-L204) | Parser: scans Markdown body line-by-line, ignores code fences, extracts first level-1 ATX heading. | Produces raw authored `Title` string. |
| `fsstore.parseTask` | [`internal/store/fsstore.go:336-338`](file:///internal/store/fsstore.go#L336-L338) | Store reader: calls `domain.FirstH1(string(body))` on every task read; populates `domain.Task.Title`. | Sets `t.Title = title` when H1 exists; leaves empty otherwise. |
| `domain.Task.Title` | [`internal/domain/task.go:8-11`](file:///internal/domain/task.go#L8-L11) | Domain entity: carries optional presentation title (`yaml:"-"`). | Real title for filesystem tasks with H1; empty for bodyless/portable stores. |
| `fsstore.CreateTask` | [`internal/store/create.go:168-214`](file:///internal/store/create.go#L168-L214) | Store writer: creates new task file from frontmatter fields and body scaffold. | Does not persist `Title` to frontmatter (`yaml:"-"`); re-read derives it from body. |
| `fsstore.RenameTask` | [`internal/store/rename.go:19-70`](file:///internal/store/rename.go#L19-L70) | Store mutator: renames task file, rewrites body H1 to new title, cascades links, parses fresh task. | Updates body H1 and reloads fresh `t.Title`. |
| `core.ThreadGraphNode` | [`internal/core/thread_graph.go:9-19`](file:///internal/core/thread_graph.go#L9-L19) | Core entity: carries raw unescaped `TaskID`, `Label`, `Title`, `Description`, `Status`, `Role`, `State`. | Carries both `Title` (optional H1) and `Label` (slug fallback). |
| `core.threadGraphNode` | [`internal/core/thread_graph.go:163-172`](file:///internal/core/thread_graph.go#L163-L172) | Mapper: constructs `ThreadGraphNode` from `ThreadTaskView`. | Maps `Title: item.Task.Title`, `Label: slug` (or `TaskID`). |
| `core.ProjectThreadGraph` | [`internal/core/thread_graph.go:48-91`](file:///internal/core/thread_graph.go#L48-L91) | Core projection: builds adapter-neutral `ThreadGraphProjection` containing nodes and edges. | Transports nodes with populated `Title` and `Label`. |
| `wire.ThreadGraphNodeJSON` | [`internal/wire/thread.go:188-197`](file:///internal/wire/thread.go#L188-L197) | Wire DTO: machine representation with optional `Title` (`json:"title,omitempty"`). | Serializes `title` when present; omits when empty. |
| `wire.ToThreadGraphProjectionJSON` | [`internal/wire/thread.go:218-237`](file:///internal/wire/thread.go#L218-L237) | Serializer: maps core projection to wire JSON envelope. | Passes `node.Title` into `ThreadGraphNodeJSON.Title`. |
| `graphfmt.RenderOptions` | [`internal/graphfmt/graphfmt.go:21-23`](file:///internal/graphfmt/graphfmt.go#L21-L23) | Presentation options: controls `IncludeDescriptions bool`. | Pure presentation parameter; does not alter projection. |
| `graphfmt.Mermaid` / `MermaidWithOptions` | [`internal/graphfmt/graphfmt.go:32-69`](file:///internal/graphfmt/graphfmt.go#L32-L69) | Text renderer: formats flowchart TD with synthetic node IDs, escaped labels, edges, and legend. | Prefers `Title`; falls back to humanized `Label`. |
| `graphfmt.DOT` / `DOTWithOptions` | [`internal/graphfmt/graphfmt.go:75-109`](file:///internal/graphfmt/graphfmt.go#L75-L109) | Text renderer: formats Graphviz digraph with custom attributes, escaped labels, edges, and legend. | Prefers `Title`; falls back to humanized `Label`. |
| `graphfmt.nodeLabel` | [`internal/graphfmt/graphfmt.go:151-177`](file:///internal/graphfmt/graphfmt.go#L151-L177) | Label composer: builds Title, Status/Role, optional Description, and ID lines. | Prefers `Title`, falls back to slug, then `"Task"`. |
| `graphfmt.compactText` | [`internal/graphfmt/graphfmt.go:189-200`](file:///internal/graphfmt/graphfmt.go#L189-L200) | Truncator: normalizes whitespace, truncates on rune / word boundaries with `"…"`. | Bounded to 64 runes (label) and 120 runes (description). |
| `cli.newThreadGraphCmd` | [`internal/cli/thread.go:333-377`](file:///internal/cli/thread.go#L333-L377) | CLI command: exposes `thread graph <thread>` with `--format`, `--details`, and `--json`. | Exposes rendered diagrams or neutral JSON. |
| `clirender.ThreadPlanHuman` | [`internal/cli/render/thread.go:244-288`](file:///internal/cli/render/thread.go#L244-L288) | CLI renderer: renders human-facing wave plan. | Intentionally uses `node.Label` (slug) for compact tabular rows. |
| `tui.threadGraphNodeRef` | [`internal/tui/detail.go:1392-1407`](file:///internal/tui/detail.go#L1392-L1407) | TUI renderer: formats task references in Thread detail and spatial canvas. | Uses `node.Label` (slug) for tight terminal column constraints. |

---

### Validation, hostile fixture, performance, and mutation-test evidence

#### Mandatory verification suite outcomes

All repo verification commands were executed within the isolated review workspace:
- `go test ./...`: **PASS** (All unit and integration tests green).
- `go test -race ./...`: **PASS** (Zero race conditions detected across all packages).
- `golangci-lint run ./...`: **PASS** (0 issues reported).
- `go vet ./...`: **PASS** (Clean).
- `go run ./cmd/tskflwctl lint`: **PASS** (`✔ all planning entities and dependency links pass lint`).
- `go run ./cmd/tskflwctl audit lint`: **PASS** (`✔ all audit findings pass lint`).
- `just docs-check`: **PASS** (Committed CLI reference in `docs/cli/` matches command tree).
- `just tidy-check`: **PASS** (`go mod tidy -diff` reports no module drift).
- `git diff --check`: **PASS** (No conflict markers or whitespace errors).

#### Hostile angle and boundary probe results

1. **Code Fence Isolation and Unterminated Fences:**
   - Probed `domain.FirstH1` with H1 headings inside standard backtick fences (` ``` `), tilde fences (`~~~`), 4-backtick outer fences containing 3-backtick inner fences, and unterminated code fences left open at EOF.
   - Verified that `fenceScanner` tracks fence character and length correctly. Code inside fences is never parsed as structure, and an unterminated fence skips all subsequent content without crashing.
2. **Line Ending and Whitespace Normalization:**
   - Probed CRLF (`\r\n`), standalone CR (`\r`), and mixed newlines. `normalizeNewlines` strips carriage returns cleanly.
   - Probed non-ATX headings (setext headings `===`), empty H1s (`#   `), and lower-level headings (`## H2`). `bodyHeadingRe` strictly requires `^(#{1,6})[ \t]+(.*\S)[ \t]*$`, so empty and non-level-1 headings are safely ignored.
   - Heading lines with leading indentation (e.g. `   # Indented`) are not matched by `bodyHeadingRe` and fall back cleanly to slug presentation.
3. **Inline Markdown and Injection Attacks:**
   - Probed titles containing inline Markdown (backticks, asterisks, bold, links, HTML tags `<script>`, and Mermaid delimiters `"` `[` `]` `%%`).
   - In Mermaid: `escapeMermaid` encodes every non-alphanumeric character outside of `" -_./:()"` into numeric HTML entities (`&#<codepoint>;`), and replaces newlines with `<br/>`. Quoted labels cannot be broken or injected with directives.
   - In DOT: `quoteDOT` escapes `\`, `"`, `\n`, `\r`, `\t`, preventing Graphviz injection.
4. **Bidi and Zero-Width Format Controls:**
   - Probed titles and descriptions containing visual-spoofing controls: U+202E (RLO), U+200B (Zero-Width Space), U+2066 (LRI), U+FEFF (BOM).
   - `unsafeFormatRune` checks `unicode.IsControl(r) || unicode.Is(unicode.Cf, r)`.
   - In Mermaid: replaced with `&#65533;`. In DOT: replaced with `\uFFFD`.
5. **Duplicate Titles and Empty Labels:**
   - Probed multiple nodes sharing the exact same human title (e.g. two tasks titled `"Same Title"`).
   - Verified that both Mermaid and DOT append `ID <taskID>` to the visible label, and DOT includes custom attribute `task_id="..."`. In `--json`, each node retains unique `task_id`. Output remains unambiguous.
   - Probed nodes with empty `Title` and empty `Label`: `nodeLabel` falls back to `"Task"`.
6. **CLI Flag Mutual Exclusivity and JSON Invariant:**
   - Verified `tskflwctl thread graph <thread> --details --json` returns `validation: renderer flags --format and --details cannot be combined with --json`.
   - Verified `tskflwctl thread graph <thread> --format dot --json` and `--format mermaid --json` return validation errors.
   - Verified `tskflwctl thread graph <thread> --json` outputs the raw, neutral JSON envelope with schema version `1.64`.
7. **Real Repository Thread Inspections:**
   - Inspected compact and detailed Mermaid and DOT outputs for `refine-thread-and-tui-navigation` (15 nodes, 16 edges) and `complete-production-threads` (33 nodes).
   - Edge counts, node counts, and health comments (`graph_health=healthy projection_health=healthy topology_complete=true`) remained identical between compact and detailed modes. Detailed mode added bounded descriptions without altering graph topology.

#### Mutation freshness probe

Executed an end-to-end mutation probe in the sandbox filesystem store verifying freshness across the entity lifecycle:
1. Created task `6fjangd7kvh0` with body `# Initial Authored Title\n\nTask body details.`.
2. Verified `fs.ListTasks()` immediately derived `Title: "Initial Authored Title"`.
3. Inspected task file on disk: verified YAML frontmatter contained zero `title:` keys (`yaml:"-"`).
4. Renamed task to `"Renamed Authored Title"` via `fs.RenameTask`: verified body H1 was rewritten, frontmatter remained unpolluted, and re-read returned `Title: "Renamed Authored Title"`.
5. Directly edited task body on disk to `# Third Mutated Title`: verified subsequent read immediately reflected `Title: "Third Mutated Title"` without cache invalidation or process restart.

#### Mutation test matrix

The following coordinated mutations were tested against new regression claims in `$SANDBOX` and restored:

| # | Mutation Target | Mutation Applied | Target Regression Test | Test Outcome | Verdict |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Fence check in `FirstH1` | Removed `if fence.inCode(line) { continue }` in `domain.FirstH1` | `TestFirstH1SkipsFencesAndNormalizesCRLF` | **FAILED:** `FirstH1 = "Example only", true` | **KILLED** |
| 2 | Title preference in `nodeLabel` | Changed `compactText(node.Title, ...)` to `compactText(node.Label, ...)` | `TestNodeLabelsPrioritizeReadableIdentityAndBoundOptionalDetails` | **FAILED:** body-derived title was not preferred | **KILLED** |
| 3 | Label truncation | Removed length truncation in `compactText`, returning raw string | `TestNodeLabelsPrioritizeReadableIdentityAndBoundOptionalDetails` | **FAILED:** long label was not visibly bounded | **KILLED** |
| 4 | Flag conflict guard | Removed `|| details` from `--json` validation in `newThreadGraphCmd` | `TestThreadGraphRejectsRendererSelectionInJSONAndUnknownFormats` | **FAILED:** `--details --json` succeeded instead of failing | **KILLED** |
| 5 | Wire schema bump | Reverted `wire.SchemaVersion` back to `"1.63"` | Golden snapshot suite (`TestCLIJSONOutputMatchesGolden`) | **FAILED:** 30+ golden snapshot tests failed with schema mismatch | **KILLED** |
| 6 | Unicode with spaces truncation | Passed Unicode title with early space at rune 20 (byte 60) | `TestNodeLabelsPrioritizeReadableIdentityAndBoundOptionalDetails` | **DEFECT DETECTED:** String truncated to 20 runes instead of 63 runes | **SURVIVED (Bug M1)** |

---

### Second-pass systemic assessment

#### Architectural Devil's Advocate: Body-Derived Presentation Data vs. Frontmatter Identity

A central question of this review is whether deriving `Title` from the first Markdown H1 is an appropriate shared semantic or an adapter convenience improperly surfaced into core domain types.

1. **Frontmatter Duplication vs. Body Derivation:**
   If `title` were stored in frontmatter, every task document would declare its title twice: once in YAML frontmatter (`title: Foo`) and once in the Markdown body heading (`# Foo`). In practice, human authors and automated tools frequently update headings without updating frontmatter (and vice versa), leading to dual-source-of-truth divergence. Keeping frontmatter lean (identities, lifecycle, graph edges) and deriving presentation headings from the authored document body maintains single-source truth.
2. **Adapter Neutrality and the Role of `Slug`:**
   By marking `domain.Task.Title` with `yaml:"-"` and treating it as optional on `core.ThreadGraphNode` and wire `ThreadGraphNodeJSON` (`omitempty`), non-filesystem adapters (such as remote APIs, in-memory graph stubs, or lightweight metadata stores) are not forced to manufacture or parse Markdown bodies. `Label` retains its stable, backwards-compatible role carrying the slug, ensuring zero breakage for existing consumers.
3. **I/O and Parsing Scaling Considerations:**
   In `fsstore.go`, `parseTask` previously parsed only frontmatter; the body byte slice was ignored. Now, `parseTask` converts `body []byte` to `string` and calls `domain.FirstH1` on every task during `ListTasks()`. While negligible on typical repositories (tens to hundreds of tasks), scanning every body string during whole-graph reads could become a hotspot if repositories scale to tens of thousands of tasks with multi-megabyte bodies. Should scale require it, a streaming header reader or lazy body accessor would preserve this contract without loading full task bodies into strings during graph operations.
4. **Presentation Consistency Across CLI and TUI:**
   `thread graph` foregrounds the human `Title` because visual diagrams benefit enormously from descriptive multi-word titles. Conversely, `thread plan` (human CLI) and TUI spatial views retain `Label` (slug) because tabular terminal outputs and fixed-width canvas cards require short, uniform tokens without mid-word line-wrapping. This divergence is intentional, principled, and documented in `docs/ARCHITECTURE.md`.

---

### Findings

#### M1. Byte-versus-rune index comparison in compactText causes premature truncation for non-ASCII text · **Status:** fixed 2026-09-12

- **Severity:** Medium
- **Location:** [`internal/graphfmt/graphfmt.go:189-200`](file:///internal/graphfmt/graphfmt.go#L189-L200)
- **Evidence:**
  In `compactText`:
  ```go
  func compactText(value string, limit int) string {
      value = strings.Join(strings.Fields(value), " ")
      runes := []rune(value)
      if len(runes) <= limit {
          return value
      }
      cut := strings.TrimSpace(string(runes[:limit-1]))
      if lastSpace := strings.LastIndex(cut, " "); lastSpace >= (limit-1)*2/3 {
          cut = strings.TrimSpace(cut[:lastSpace])
      }
      return cut + "…"
  }
  ```
  `strings.LastIndex(cut, " ")` returns the **byte index** of the last space in UTF-8 string `cut`. However, `(limit-1)*2/3` is a **rune count threshold** (e.g. 42 runes for `limit=64`, 79 runes for `limit=120`).

  For non-ASCII text (such as CJK, Cyrillic, Greek, or accented European characters), each rune occupies 2 to 4 bytes. If a title or description contains an early space—for example at rune 20 (byte 60 for 3-byte CJK runes)—the byte offset `60` is compared directly against the rune threshold `42`. Because `60 >= 42`, the condition evaluates to `true`, and `cut` is truncated at `cut[:lastSpace]` (rune 20).

  Hostile probe execution:
  ```go
  longNode.Title = strings.Repeat("界", 20) + " " + strings.Repeat("界", 45) // 66 runes
  firstLine := strings.Split(nodeLabel(longNode, RenderOptions{}), "\n")[0]
  // Result: runes=21 line="界界界界界界界界界界界界界界界界界界界界…"
  ```
  Out of an allowed budget of 64 runes, 43 runes (over 67% of the budget) were discarded because a space at rune 20 had a byte index of 60.
- **Impact:**
  Non-ASCII task titles and descriptions with spaces are aggressively and prematurely truncated at 20–30% of the visible budget rather than preserving text up to the 2/3 rune threshold or the 64/120 rune limit. The existing unit test (`strings.Repeat("界", maxNodeLabelRunes+10)`) masked this issue because it tested a Unicode string without any spaces.
- **Remediation:**
  Operate purely in rune space when locating the last word boundary, or count the runes up to `lastSpace`:
  ```go
  func compactText(value string, limit int) string {
      value = strings.Join(strings.Fields(value), " ")
      runes := []rune(value)
      if len(runes) <= limit {
          return value
      }
      cutRunes := runes[:limit-1]
      lastSpace := -1
      for i := len(cutRunes) - 1; i >= 0; i-- {
          if cutRunes[i] == ' ' {
              lastSpace = i
              break
          }
      }
      if lastSpace >= (limit-1)*2/3 {
          cutRunes = cutRunes[:lastSpace]
      }
      cut := strings.TrimSpace(string(cutRunes))
      return cut + "…"
  }
  ```

**Resolution:** Reworked compactText to find word boundaries in rune space; a
spaced multi-byte Unicode regression proves it consumes the intended 64-rune
budget.

#### M2. Graph formatters emit compact label guidance in legend even when --details is enabled · **Status:** fixed 2026-09-12

- **Severity:** Medium
- **Location:** [`internal/graphfmt/graphfmt.go:59`](file:///internal/graphfmt/graphfmt.go#L59), [`internal/graphfmt/graphfmt.go:103`](file:///internal/graphfmt/graphfmt.go#L103)
- **Evidence:**
  In `MermaidWithOptions`:
  ```go
  out.WriteString("  subgraph legend[\"Legend &#183; compact labels &#183; add --details for descriptions\"]\n")
  ```
  In `DOTWithOptions`:
  ```go
  out.WriteString("    label=\"Legend\\nCompact labels; add --details for descriptions\";\n")
  ```
  These legend titles are statically hardcoded regardless of `options.IncludeDescriptions`. When a user runs `tskflwctl thread graph <thread> --details`, the nodes render full bounded descriptions, but the legend continues to declare `Legend · compact labels · add --details for descriptions`.
- **Impact:**
  The legend contradicts the actual rendered graph state, falsely claiming that labels are compact and directing the user to pass a flag that is already enabled.
- **Remediation:**
  Condition the legend title on `options.IncludeDescriptions`:
  ```go
  legendTitle := "Legend &#183; compact labels &#183; add --details for descriptions"
  if options.IncludeDescriptions {
      legendTitle = "Legend &#183; detailed descriptions"
  }
  fmt.Fprintf(&out, "  subgraph legend[\"%s\"]\n", legendTitle)
  ```
  Apply the matching pattern to Graphviz DOT output.

**Resolution:** Made Mermaid and DOT legend guidance conditional on
RenderOptions; detailed output now says descriptions are included and is covered
at formatter and CLI levels.

#### L1. Formatters mutate casing of human-authored titles by unconditionally uppercasing initial rune · **Status:** fixed 2026-09-12

- **Severity:** Low
- **Location:** [`internal/graphfmt/graphfmt.go:159-161`](file:///internal/graphfmt/graphfmt.go#L159-L161)
- **Evidence:**
  In `nodeLabel`:
  ```go
  label := compactText(node.Title, maxNodeLabelRunes)
  if label == "" {
      label = compactText(strings.ReplaceAll(node.Label, "-", " "), maxNodeLabelRunes)
  }
  if label == "" {
      label = "Task"
  }
  labelRunes := []rune(label)
  labelRunes[0] = unicode.ToUpper(labelRunes[0])
  parts := []string{string(labelRunes)}
  ```
  When `node.Title` is present, it is human-authored Markdown presentation text. For titles starting with lowercase terms (e.g. `iOS client support`, `kubectl plugin`, `eBPF network monitor`, `eBay API integration`), `labelRunes[0] = unicode.ToUpper(...)` mutates the first letter to uppercase (`IOS client support`, `Kubectl plugin`, `EBPF network monitor`).
- **Impact:**
  Cosmetic mutation of intentional author casing. Sentence-casing the first letter is necessary when humanizing machine slugs (the fallback branch: `strings.ReplaceAll(node.Label, "-", " ")`), but should not override authored title casing.
- **Remediation:**
  Apply `unicode.ToUpper` only when humanizing the slug fallback:
  ```go
  label := compactText(node.Title, maxNodeLabelRunes)
  if label == "" {
      label = compactText(strings.ReplaceAll(node.Label, "-", " "), maxNodeLabelRunes)
      if label == "" {
          label = "Task"
      }
      labelRunes := []rune(label)
      labelRunes[0] = unicode.ToUpper(labelRunes[0])
      label = string(labelRunes)
  }
  parts := []string{label}
  ```

**Resolution:** Limited sentence-casing to humanized slug fallbacks so
body-authored titles such as iOS and eBPF retain their exact casing.

#### L2. thread graph command executes service reads before validating format flag · **Status:** fixed 2026-09-12

- **Severity:** Low
- **Location:** [`internal/cli/thread.go:343-363`](file:///internal/cli/thread.go#L343-L363)
- **Evidence:**
  In `newThreadGraphCmd`:
  ```go
  RunE: func(cmd *cobra.Command, args []string) error {
      if app.JSON && (cmd.Flags().Changed("format") || details) {
          return fmt.Errorf("%w: renderer flags --format and --details cannot be combined with --json...", domain.ErrValidation)
      }
      projection, err := app.Svc.ShowThreadGraph(args[0])
      if err != nil {
          return err
      }
      if app.JSON {
          return render.ThreadGraphJSON(app.Out, projection)
      }
      var output string
      options := graphfmt.RenderOptions{IncludeDescriptions: details}
      switch format {
      case "mermaid":
          output, err = graphfmt.MermaidWithOptions(projection, options)
      case "dot":
          output, err = graphfmt.DOTWithOptions(projection, options)
      default:
          return fmt.Errorf("%w: unsupported Thread graph format %q (want mermaid or dot)", domain.ErrValidation, format)
      }
  ```
  If an invalid `--format` (e.g. `--format ascii`) is supplied without `--json`, the command calls `app.Svc.ShowThreadGraph(args[0])` before validating `format`.
  If the thread ID is invalid or missing, `ShowThreadGraph` returns a thread-not-found error, masking the invalid format argument. If the thread is valid, unnecessary filesystem I/O and graph computation occur before rejecting the unsupported format.
- **Impact:**
  Inconsistent error precedence and unnecessary execution of read/projection pipelines on invalid command-line inputs.
- **Remediation:**
  Validate `format` upfront alongside the `--json` flag checks:
  ```go
  if format != "mermaid" && format != "dot" {
      return fmt.Errorf("%w: unsupported Thread graph format %q (want mermaid or dot)", domain.ErrValidation, format)
  }
  ```

---

**Resolution:** Moved renderer-format validation ahead of ShowThreadGraph; an
invalid format now wins even when the requested Thread is missing.

### Residual risks and follow-ups

1. **ATX Closing Hash Sequence:**
   CommonMark §4.2 specifies that trailing `#` characters separated by whitespace at the end of an ATX heading line constitute an optional closing sequence and should be ignored. `bodyHeadingRe` currently retains trailing hashes as part of the title (e.g. `# Title ###` yields `"Title ###"`). In taskflow templates this pattern is not used, but a future enhancement could strip trailing `#` sequences.
2. **Whole-Repository Task Body Parsing:**
   `fsstore.go:336` now converts every task body to string and scans it for H1 during `parseTask`. For large enterprise repositories with thousands of tasks, task listing performance should be profiled, and consider moving to a streaming scanner if body allocation overhead becomes measurable.
3. **Future TUI Adoption of Task Titles:**
   Currently, TUI spatial layout and detail navigation continue to use `node.Label` (slug) to maintain compact card geometries and prevent table misalignment. When spatial card rendering is extended with an expanded inspector or title tooltip, `node.Title` is already present on `core.ThreadGraphNode` and ready for presentation use without core schema changes.
