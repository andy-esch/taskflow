---
schema: 1
id: 6g3ahpetw89g
bucket: open
area: finding-note-and-vocabulary-selfreview
date: "2026-08-24"
updated_at: "2026-08-24"
---

# Audit: finding-note-and-vocabulary-selfreview — 2026-08-24

> Self-review of this session's work: the `landed`→`tracked` vocabulary swap, the M2
> decoration fix, `audit finding --note` / `--pr`, and the lint rules added alongside them.
> Every finding below was reproduced before it was written down.

## Findings

#### H1. A `**Resolution:**` label followed by a blank line orphans its own paragraph  · **Status:** fixed 2026-08-24

**File:** `internal/domain/finding.go` (`findingNote`) | **Component:** domain / parsing
**Effort:** XS · **Urgency:** acute

`findingNote` ends a note's paragraph at the first `\n\n`. When a hand edit puts the blank
line immediately AFTER the label, that terminator is at offset 0, so the note parses as
empty while its span covers the label alone:

```
input   **Resolution:**\n\nThe text lives here.
parsed  Note="" NoteSpan={34 49}   ← non-empty span, empty text
```

`SetFindingNote` then takes the replace branch (the span is non-empty), rewrites only the
label, and STRANDS the original paragraph as loose prose:

```
**Resolution:** replacement

The text lives here.          ← orphaned, now belongs to nothing
```

The document silently loses the association between a resolution and its finding, and the
command reports success. This is the same failure shape as H1 of
`2026-08-24-planning-state-vocabulary` — a span that does not cover what the writer assumes
it covers — reached by a different route.

**Recommendation:** when the label is followed by a blank line, treat the note as absent
AND the span as empty, so the write appends a fresh block instead of overwriting a label
whose text it never read. Add the case to `TestSetFindingNote`.

**Resolution:** findingNote now returns no note AND no span when the label has
no text on its line, so the write appends a fresh block and leaves the orphaned
paragraph intact. LintFindings names the stray label.

#### M1. `--status ""` is a silent no-op reported as success  · **Status:** fixed 2026-08-24

**File:** `internal/core/finding.go` (`FindingEdit.apply`) | **Component:** core / cli
**Effort:** XS · **Urgency:** soon

`apply` skips the status edit when `e.Status == ""`, and the CLI only checks that the FLAG
was passed, not that it carries a value:

```
$ tskflwctl audit finding <audit> H2 --status "" --dry-run
• H2 in <audit> already reads that way
```

Nothing is written and the receipt is shaped like success. The realistic way to hit this is
the scripted one an agent would write — `--status "$STATUS"` with the variable unset — which
is exactly the caller this write path exists to serve. `domain.SetFindingStatus("")` would
have rejected it properly; the guard in `apply` is what swallows it.

**Recommendation:** reject an empty `--status` at the CLI with the vocabulary listed, the
same way a missing flag is rejected. Keep `apply`'s zero-means-untouched contract, which is
correct for the struct.

**Resolution:** The CLI rejects an empty --status with the vocabulary listed.
FindingEdit keeps its zero-means-untouched contract, which is right for the
struct — the check belongs where the flag is read.

#### M2. The criterion-state wire descriptions never learned `tracked`  · **Status:** fixed 2026-08-24

**File:** `internal/wire/dto.go:57,95` | **Component:** wire / contract
**Effort:** XS · **Urgency:** soon

Adding `tracked` to the criterion vocabulary updated the changelog in `wire.go` but not the
two `jsonschema` descriptions a consumer actually reads:

```go
// Explained is how many UNMET criteria state why (deferred / wontfix / n/a). Zero for
State string `json:"state,omitempty" jsonschema:"description=disposition beyond the checkbox: deferred | wontfix | n/a — absent for a plain met/not-met criterion"`
```

Both are published through `schema --json-schema`. A consumer reading a three-value
enumeration will treat a `tracked` criterion as malformed. This is M3 of
`2026-08-17-finding-status-surface` recurring inside the very change that closed it: a
hand-kept description falling a word behind the code.

**Recommendation:** derive both strings from `domain.CriterionStates()` rather than
transcribing them, matching what the `schema audit` conventions line now does.

**Resolution:** The schema contract now publishes criterion_states, the way it
already publishes finding_statuses, and the two wire descriptions point at it
instead of transcribing a list that had already fallen a word behind.

#### M3. A `tracked` criterion needs no destination, while a `tracked` finding does  · **Status:** open

**File:** `internal/domain/resolution.go` · `internal/cli/task.go` | **Component:** domain / vocabulary
**Effort:** S · **Urgency:** soon

`SetFindingStatus` refuses a destination-less `tracked` and `LintFindings` flags one. The
criterion side requires only *a* reason, so this is accepted:

```
$ tskflwctl task ac <slug> --tracked 1 --reason "just because"
✔ would set criterion 1 tracked
```

The word means "handed off" in both places — that is the entire premise of
`SharedResolutionWords` — but only one of the two makes the handoff followable. A shared
vocabulary whose guarantees differ per entity is a vocabulary that has already started to
drift.

**Recommendation:** decide deliberately and record it. Either require an id-shaped token in
a `tracked` criterion's reason, or document why a criterion's destination is softer than a
finding's. The asymmetry is defensible — an audit concludes its interest on handoff while a
task's work merely moves — but it is currently accidental, not stated.

#### M4. A note containing the literal label can wrap into a false duplicate  · **Status:** fixed 2026-08-24

**File:** `internal/domain/finding.go` (`wrapNote` / `DuplicateNotes`) | **Component:** domain / lint
**Effort:** XS · **Urgency:** soon

`wrapNote` breaks lines without regard to content, and `DuplicateNotes` counts label matches
at line start. A note whose text quotes `**Resolution:**` — likely, since audits about this
tool discuss its own grammar — can have that text land at a line start:

```
**Resolution:** xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
**Resolution:** trailing words here        ← wrap put it here; lint calls it a duplicate
```

Reproduced at padding widths 49 and up. The note still round-trips correctly, so this is a
false lint positive rather than corruption — but it lights up on exactly the audits most
likely to discuss the feature.

**Recommendation:** never break a line immediately before a `**` token that would start a
line, or count duplicates from the parse (which already knows where the real note ends)
rather than by re-scanning the section.

**Resolution:** wrapNote never breaks before a token starting with **, so no
continuation line can read as a second label. Verified across every wrap
position from 30 to 90.

#### L1. Neither the scaffold nor the authoring guidance mentions `**Resolution:**`  · **Status:** fixed 2026-08-24

**File:** `internal/domain/entity.go` (audit `Conventions`, `auditBodyTemplate`) | **Component:** domain / docs
**Effort:** XS · **Urgency:** eventually

`schema audit` now derives the status vocabulary, but says nothing about the resolution
block: zero matches for "resolution" in its whole output, and the body template shows a
finding without one. An agent authoring an audit has no way to discover that `--note`
exists, so it will keep hand-typing the paragraph — the practice `--note` was built to end.

**Recommendation:** add the block to the finding shape in `auditBodyTemplate` and name
`audit finding --note` in the conventions, beside the status line that is already there.

**Resolution:** The audit body template carries a **Resolution:** line and the
conventions name audit finding --status/--note as the way to write both.

#### L2. README documents the finding READ path and not the write path  · **Status:** fixed 2026-08-24

**File:** `README.md:114,129` | **Component:** docs
**Effort:** XS · **Urgency:** eventually

The cheat sheet lists `audit findings --status open …` (query) but never `audit finding`
(write), and describes `task ac` as "--check/--uncheck <n> to flip one" — the binary
vocabulary that `--defer/--wontfix/--tracked/--na` replaced. A reader of the README would
conclude the non-binary states do not exist.

**Recommendation:** one line for `audit finding`, and widen the `task ac` line past the two
binary verbs.

**Resolution:** README gained the audit finding write path and a task ac line
showing the non-binary states.

#### L3. The no-op receipt lost the value it used to name  · **Status:** fixed 2026-08-24

**File:** `internal/cli/audit.go` (`newAuditFindingCmd`) | **Component:** cli / ux
**Effort:** XS · **Urgency:** eventually

Before `--note`, an unchanged re-stamp said `H2 in <audit> is already fixed`. It now says
`H2 in <audit> already reads that way`, because the built phrase reads badly after "is
already". The information a reader wants — WHICH value it already carries — is what got
dropped, and the phrasing problem only affects the multi-flag cases.

**Recommendation:** keep the old sentence for the status-only case and use the general one
only when a note is involved.

**Resolution:** The status-only no-op says 'is already <status>' again; the
general phrasing is used only when a note is involved.

#### L4. `DuplicateNotes` re-scans a section `findingNote` just scanned  · **Status:** fixed 2026-08-24

**File:** `internal/domain/finding.go` (`ParseFindings`) | **Component:** domain / performance
**Effort:** XS · **Urgency:** eventually

Per finding, `findingNote` runs `FindStringIndex` and then `ParseFindings` runs
`FindAllStringIndex` over the same section, allocating a full match slice only to compare
its length against 1. `ParseFindings` costs ~89µs / 12KB / 85 allocs for a 20-finding audit,
so this is not a problem today — the point is that the second scan buys nothing a
count-returning `findingNote` would not.

**Recommendation:** have `findingNote` return the match count alongside the note.

**Resolution:** findingNote returns the label count from the scan it was already
doing; ParseFindings no longer re-scans the section.

#### L5. The `--pr` duplicate guard misses `(PR#9)`  · **Status:** fixed 2026-08-24

**File:** `internal/cli/audit.go` (`decorateWithPR`) | **Component:** cli / validation
**Effort:** XS · **Urgency:** eventually

The guard looks for `PR #` with a space, so a status already carrying the un-spaced form
gets a second reference appended:

```
$ … --status "fixed (PR#9)" --pr 12
✔ would set H2 fixed (PR#9) (PR #12) in <audit>
```

Minor, and the flag exists precisely to stop hand-spelled references — but the guard should
recognise the shapes it is displacing.

**Recommendation:** match `PR` followed by optional space and `#`, case-insensitively.

**Resolution:** The guard is a case-insensitive regex over PR, optional space,
optional #, digit — the hand-spelled shapes --pr exists to displace.

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

- ⏳ H1, M1, M2, M4, L4, L5 — mechanical fixes to this session's own work; no task needed if
  taken in the same branch.
- ⏳ M3 — a vocabulary decision, and the one finding here worth thinking about rather than
  patching. Candidate for folding into
  `let-an-acceptance-criterion-say-more-than-done-or-not-done`.
- ⏳ L1, L2, L3 — documentation and receipt polish; fold into the same branch.
