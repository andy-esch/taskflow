---
schema: 1
id: 6g1397jfke23
bucket: closed
area: finding-status-surface
date: "2026-08-17"
updated_at: "2026-08-24"
---

# Audit: finding-status-surface — 2026-08-17

> Edit findings in place and flip each `**Status:**` as you work it.

**Source:** field report from `andy-esch/desirelines-planning`, the largest
external consumer of the audit surface. `tskflwctl audit lint` had been failing
there continuously — **11 audits, ~60 findings** — and was repaired by hand on
2026-08-17. Not one of the ~60 was a genuine structural problem. They were six
different spellings of six legal statuses, plus two audits written before the
convention existed.

The interesting part is not that a consumer wrote bad statuses. It is that the
failure was *invisible for months, uniform in kind, and mechanically fixable* —
which points at this surface rather than at that consumer. This audit is
propose-only, in the usual shape.

## Context: how it stayed red

Three audit-producing routines in that repo each ended with a validation step
that ran `tskflwctl lint` and carried the same note: *"Audit files are not
tskflwctl-managed, but lint sanity-checks the rest of the tree."* True before
`audit new`/`audit lint` shipped; never revisited after. So the only command
that validates a finding's `**Status:**` was never run by anything, and a
permanently-red lint stopped being a signal — a genuinely malformed new audit
was indistinguishable from historical noise. Two of that repo's own tasks had
already been filed about the redness, and both were still open.

The consumer-side fixes are theirs and are done. What follows is the part this
repo owns.

## Findings

#### H1. There is no validated write path for a finding's status  · **Status:** fixed 2026-08-24

**File:** `internal/cli/audit.go:28-37` | **Component:** cli / audit
**Effort:** M · **Urgency:** soon

`task` has a validated mutation surface: `task set` refuses an illegal value,
writes atomically, and stamps `updated_at`. The audit surface registers
`new/list/show/info/path/edit/append/findings/lint` plus the bucket-move verbs —
**nothing that writes a finding's status**. Every status transition since the
convention began, in every consuming repo, has been a hand-edit of markdown
validated only by a command nothing ran (see Context).

This is the root cause the rest of this audit is downstream of. `ValidFindingStatus`
(`internal/domain/finding.go:99`) already exists and is exactly the check such a
command would call; there is simply no command calling it on the write path. The
valid path is not merely harder than the free-text path — it does not exist.

**Recommendation:** `tskflwctl audit finding <slug> <code> --status <value>
[--note <text>]`, mirroring `task set`: validated against `FindingStatuses()`,
atomic single write, `--json`. Agent-facing, like `audit append`.

**Tightening (adjacent):** the same command is the natural place to enforce the
"on `fixed`, add a resolution block" convention that currently lives only in
consumer docs.

#### H2. `lint --fix` advertises repairing audits but never looks at finding status  · **Status:** superseded 2026-08-24

**File:** `internal/cli/lint.go:16,28` | **Component:** cli / lint
**Effort:** S · **Urgency:** soon

The two lint surfaces are split in a way that reads backwards from the outside:

```
$ tskflwctl lint --help
  Validate active task and epic frontmatter (--fix repairs tasks/audits …)
  --fix   auto-repair frontmatter: … backfill missing task/audit ids

$ tskflwctl audit lint --help
  (no --fix flag)
```

So the command whose help text says it repairs **audits** only touches
frontmatter and ids, while the command that actually finds audit *content*
problems can repair nothing. A reader who runs `lint --fix` and sees it succeed
has reasonable grounds to believe their audits are clean.

Roughly 55 of the ~60 field-report findings were mechanical substitutions from a
closed set (see M1's table). A conservative `--fix` would have resolved them in
one command instead of a scripted hand-repair.

**Recommendation:** add `--fix` to `audit lint`, over a static alias table, honoring
the existing global `--dry-run`. Separately, narrow the `lint` help text so
"repairs tasks/audits" cannot be read as covering finding status.

**Resolution:** The task carrying this was deprecated: audit lint now passes on
the whole corpus, so there is no legacy debt for a --fix to repair. The emoji
half went obsolete when M2 made the parser decoration-tolerant, and the legacy
words it wanted to normalise are gone (landed) or promoted (tracked). See
6fq9zy13wkdc for the measurement.

#### M1. The status error names the offending value but not the legal set  · **Status:** fixed

**File:** `internal/domain/finding.go:114,116` | **Component:** domain / lint
**Effort:** XS · **Urgency:** soon

```go
issues = append(issues, Issue{Field: f.Code, Message: "missing **Status:**"})
issues = append(issues, Issue{Field: f.Code,
    Message: fmt.Sprintf("unknown status %q", f.Status)})
```

`unknown status "tracked"` says you are wrong without saying what would be right,
so every fix costs a documentation round-trip — and in the field report the
documentation was itself incomplete (M3), so that round-trip could *fail*.
`FindingStatuses()` is eleven lines above at `:89`, already sorted, already used
for help/schema.

These are the substitutions that actually showed up, all six mapping onto the
legal set without ambiguity:

| Written | Legal value | Audits |
|---|---|---|
| `✅ fixed <date>` | `fixed <date>` | 4 |
| `✅ done <date>` | `fixed <date>` | 1 |
| `✅ implemented … (pending commit)` | `fixed <date>` once merged | 2 |
| `tracked`, `tracked → <task>` | `superseded by <link>` | 2 |
| `✅ moot` | `superseded by <link>` | 1 |
| `⚠️ PARTIALLY FIXED` | `superseded by <link>` | 1 |
| `⛔ declined` | `wontfix (reason)` | 1 |
| `⏳ deferred — <reason>` | `deferred (<reason>)` | 1 |

Note these are *semantic* near-misses, not typos — edit-distance suggestion would
catch none of them. A static alias table is the right shape, and doubles as the
`--fix` map for H2.

**Recommendation:** append the legal set to both messages, e.g.
`unknown status "tracked" (legal: deferred, fixed, in-progress, landed, open,
superseded, wontfix)`. Optionally add `did you mean superseded?` from the alias
table.

**Resolution:** Named the legal set in the error message, derived from
FindingStatuses() so the diagnostic cannot fall behind the vocabulary it is
checking against.

#### M2. A leading emoji is captured *as* the status  · **Status:** fixed 2026-08-24

**File:** `internal/domain/finding.go:43` | **Component:** domain / parsing
**Effort:** S · **Urgency:** eventually

```
statusRe = `(?mi)(?:^[ \t]*|·[ \t]*)\*\*Status:\*\*[ \t]*([^\s·|*]+)`

input      **Status:** ✅ fixed 2026-06-04
captured   "✅"                → unknown status "✅"
```

Seven of the eleven red audits were exactly this, making it the single highest-frequency
cause. The capture rule is working as documented — first run with no
whitespace/`·`/`|` — and the `*`-exclusion comment shows the edge cases were
thought about; an emoji prefix just was not among them.

Worth knowing *why* consumers write it: that repo's execute-side HOWTO teaches
✅/⚠️/⏳/⛔ for the candidate-tasks list roughly 130 lines above the status table,
and instructs authors to keep the two "synced". A decorative prefix reads as
agreement between the two lists, not as a violation. Any consumer following the
same two-list pattern this tool's own `audit new` scaffold suggests — findings
above, `<!-- ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->` below — is exposed
to the same collision.

Today's behaviour is the least useful of the options: it neither accepts the
value nor explains the confusion.

**Recommendation:** decide between the two and commit. Either skip a leading
emoji/punctuation run before capturing (`✅ fixed` becomes valid), or detect a
candidate-list symbol and say so: *"that's a candidate-list symbol; a finding
`**Status:**` takes a word."* The pointed error is probably better — silent
tolerance lets the two namespaces keep merging, and the scaffold at
`audit new` is what puts them adjacent in the first place.

**Resolution:** The glyph is decoration, not the status: `statusRe` now swallows
a leading run of non-alphanumerics and the token is read through
`stripLeadingDecoration`, the same helper criterion suffixes use. The span
deliberately still covers the glyph, so re-stamping replaces it — leaving `✅
deferred` would restate the same two-lists-disagree bug inside the file.

#### M3. `landed` is legal in code, absent from the docs, and the code asserts otherwise  · **Status:** fixed 2026-08-24

**File:** `internal/domain/finding.go:80-86` | **Component:** domain / vocabulary
**Effort:** XS · **Urgency:** soon

```go
// findingStatuses is the legal finding-status vocabulary (the audit HOWTO + the
// `audit new` scaffold). A free-text Status edit can write a typo; `audit lint`
// catches it against this set.
var findingStatuses = map[string]bool{
	"open": true, "in-progress": true, "fixed": true, "landed": true,
	...
```

The comment names two sources of truth and holds itself in sync with both.
Neither held. The consumer's HOWTO cheat sheet listed **six** values with no
`landed`; the `audit new` scaffold only ever emits `open`. No audit in that
corpus has ever used `landed`, and nothing tests the claim.

The concrete harm is not that `landed` is accepted — it is that a reader of a
six-row table believes the vocabulary is *closed at six*, so the table becomes
the thing they reason from and the seventh value stays invisible.

**Recommendation:** drop `landed`, or document how it differs from `fixed`.

**Follow-up:** this is a small instance of exactly what
[`26-frontmatter-schema-declared-validation-contract`](../epics/26-frontmatter-schema-declared-validation-contract.md)
targets — a hand-maintained vocabulary with docs asserted, not derived. Emitting
the status table from `FindingStatuses()` (it is already sorted for `schema`)
would make the comment's claim structurally true. Worth folding in rather than
tracking separately.

**Resolution:** Dropped. `landed` had zero uses in the corpus, so nothing was
lost — but removing it exposed a word that WAS missing: 7 of 13 `deferred`
findings were handoffs, improvised in prose as `deferred → tracked in task X` by
two authors months apart. Those seven now read `tracked by <id>`, a status that
counts toward the audit's done band and refuses to be written without a
destination. The follow-up was taken too: the `schema audit` conventions line is
built from FindingStatuses(), so the guidance cannot fall behind the vocabulary
the way this finding's table did.

#### L1. `audit edit`'s lint-on-save cannot reach the writers that matter  · **Status:** superseded 2026-08-24

**File:** `internal/cli/audit.go:33` (`newAuditEditCmd`) | **Component:** cli / audit
**Effort:** XS · **Urgency:** eventually

`audit edit` re-parses on save and surfaces bad statuses — the right instinct,
and the only existing safety net. But it only fires for a human in `$EDITOR`.
The routines that produce audits are agents writing files directly, so in
practice the net has never caught anything. It also surfaces issues as a
*warning*, which is the right call while there is no strict path and worth
revisiting once H1 provides one.

**Recommendation:** low priority standalone; mostly an argument for H1. If
`audit append` grows a `--strict`, the agent-facing half closes.


**Fixed 2026-08-24** while building the criterion vocabulary, which is the same lesson in a
new place. `LintFindings` now names the legal set on BOTH the unknown-status and
missing-status paths, and the new `SetFindingStatus` write path rejects with the same list.
The criterion vocabulary was built with this rule from the start: every rejection there
names its legal set too.

**Resolution:** This finding's own recommendation called it mostly an argument
for H1, and H1 shipped: agents resolve findings through audit finding
--status/--note, which validates, so the net it said had never caught anything
is no longer the only one. What remains — lint-on-save warning rather than
refusing in $EDITOR — is now a deliberate choice for the human whole-file path
rather than a fallback while no strict path existed. A parse break still reopens
the editor; a vocabulary typo warns and audit lint catches it.
## What audited clean

- **The vocabulary itself.** Six of the seven statuses are well-chosen, and every
  invented value in the field report mapped onto one of them without inventing a
  new concept — including the two that looked like they needed one
  (`partially fixed`, `moot` both resolve to `superseded by`). The set is not
  missing a member; `partial` in particular would be a mistake to add.
- **`audit lint`'s scoping.** Taking an optional slug is what makes the routine-side
  fix a one-liner (`audit lint <slug>` right after writing the file). Without it
  the consumer fix would have needed whole-tree lint in a hot path.
- **The bucket↔state invariant** (`LintFindings`, `finding.go:118-123`) — a
  non-open audit with open findings is a real and cheap check, and it caught
  genuine drift in the field-report corpus.
- **`statusRe`'s authority rule** — restricting the marker to line-start or
  post-`·` means a `**Status:**` mentioned in prose or a title cannot be mistaken
  for the real one. The fenced-block strip (`fenceRe`) means the scaffold's own
  example doesn't parse as a finding. Both are the kind of thing that only shows
  up as a bug report if you get them wrong.

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

- ⏳ `tskflwctl task new "Add audit finding --status: a validated write path for finding status" --epic 20-cli-ux-and-ergonomics --tags audit,cli,agent-surface` — H1; mirrors `task set`, calls the existing `ValidFindingStatus`.
- ⏳ `tskflwctl task new "Add --fix to audit lint and narrow lint's repairs-audits claim" --epic 20-cli-ux-and-ergonomics --tags audit,lint` — H2; static alias table, honors `--dry-run`.
- ⏳ `tskflwctl task new "Name the legal status set in audit lint's error messages" --epic 20-cli-ux-and-ergonomics --tags lint,diagnostics` — M1; XS, `FindingStatuses()` is already there.
- ⏳ `tskflwctl task new "Decide how audit lint handles an emoji-prefixed status" --epic 20-cli-ux-and-ergonomics --tags audit,parsing` — M2; needs the tolerate-vs-explain call before it can be written.
- ⏳ `tskflwctl task new "Resolve the landed status and derive the documented vocabulary from FindingStatuses" --epic 26-frontmatter-schema-declared-validation-contract --tags schema,docs` — M3; fold into the declared-contract work rather than tracking alone.
- ⛔ L1 — no standalone task; falls out of H1.

Epic ids were read from `epic list` on 2026-08-17; re-check before running these.
