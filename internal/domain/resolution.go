package domain

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// The shared resolution vocabulary.
//
// Findings and acceptance criteria both need to say "parked" and "not doing this", and they
// mean the same thing by it. Spelling those words once, here, is what stops the two from
// drifting apart — the failure recorded as finding M3 of the 2026-08-17
// finding-status-surface audit, where `landed` was legal in code, missing from the docs,
// and contradicted by an assertion.
//
// This is deliberately modelled as an OVERLAP, not a subset. `met` is not a finding status
// and `superseded` is not a criterion state, so claiming either set contains the other would be
// a lie — and the lie is precisely what lets them diverge unnoticed. Each entity declares
// its own full set; the words below are the ones that must be spelled identically in both,
// which TestSharedResolutionWordsAgree enforces.
const (
	ResolutionDeferred = "deferred"
	ResolutionWontFix  = "wontfix"
	// ResolutionTracked is "transferred, not abandoned": the work has been handed to a
	// tracked item and is no longer this document's business.
	//
	// It exists because the corpus improvised it. Seven of thirteen `deferred` findings
	// were not deferrals at all but handoffs, written as `deferred → tracked in task X` by
	// two authors months apart — the vocabulary was missing a word people needed often
	// enough to route around it in prose. `deferred` means "not now" and stays owned here;
	// `tracked` means "not here" and concludes this document's interest.
	ResolutionTracked = "tracked"
)

// SharedResolutionWords are the vocabulary members that carry the same meaning wherever
// they appear. Any entity vocabulary that uses one of these concepts must use this exact
// spelling.
func SharedResolutionWords() []string {
	return []string{ResolutionDeferred, ResolutionWontFix, ResolutionTracked}
}

// CriterionState is an acceptance criterion's disposition. The bracket carries the binary
// met/not-met that every existing checkbox already means; the non-binary states refine the
// NOT-MET case, so no existing file changes meaning and there is nothing to migrate.
type CriterionState string

const (
	CriterionMet      CriterionState = "met"
	CriterionUnmet    CriterionState = "not met"
	CriterionDeferred CriterionState = CriterionState(ResolutionDeferred)
	CriterionWontFix  CriterionState = CriterionState(ResolutionWontFix)
	CriterionTracked  CriterionState = CriterionState(ResolutionTracked)
	// CriterionNA is "turned out not to apply" — distinct from won't-do, which is a choice.
	// Findings have no word for this; criteria need one, because scope genuinely evaporates.
	CriterionNA CriterionState = "n/a"
)

// criterionStates maps the WRITTEN suffix word to its state. Only the non-binary states are
// spellable as a suffix: met and not-met are what the bracket already says, and allowing
// them twice would create two ways to say one thing and a contradiction to lint for.
var criterionStates = map[string]CriterionState{
	string(CriterionDeferred): CriterionDeferred,
	string(CriterionWontFix):  CriterionWontFix,
	string(CriterionTracked):  CriterionTracked,
	string(CriterionNA):       CriterionNA,
}

// CriterionSuffixStates returns the writable suffix words, sorted, for help and diagnostics.
// Every error that rejects a state names this set — the omission finding M1 flagged.
func CriterionSuffixStates() []string {
	out := make([]string, 0, len(criterionStates))
	for s := range criterionStates {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// CriterionStates returns every state a criterion can be in, including the two the bracket
// expresses. Used by schema/docs generation so the documented vocabulary derives from this
// definition rather than being retyped beside it.
func CriterionStates() []CriterionState {
	return []CriterionState{CriterionMet, CriterionUnmet, CriterionDeferred, CriterionWontFix, CriterionTracked, CriterionNA}
}

// ParseCriterionState maps a written suffix word to its state, tolerating case, surrounding
// space, and a leading decoration. Decoration is handled deliberately rather than
// discovered later: this repo's own candidate lists use ✅ ⏳ ⛔, and finding M2 records
// what happens when an emoji is silently captured AS the value.
func ParseCriterionState(word string) (CriterionState, bool) {
	s, ok := criterionStates[strings.ToLower(strings.TrimSpace(stripLeadingDecoration(word)))]
	return s, ok
}

// NeedsReason reports whether a state must be accompanied by an explanation. A deferral
// with no why is indistinguishable from an oversight, which is the defect this vocabulary
// exists to remove.
func (c CriterionState) NeedsReason() bool {
	return c == CriterionDeferred || c == CriterionWontFix || c == CriterionNA || c == CriterionTracked
}

// Met reports whether the criterion counts as satisfied. Only CriterionMet does: a deferred
// or abandoned criterion is explicitly NOT done, which is the whole point of saying so.
func (c CriterionState) Met() bool { return c == CriterionMet }

// stripLeadingDecoration removes a leading emoji/symbol run so "⏳ deferred" matches
// "deferred". It stops at the first letter or digit, so a word is never truncated.
//
// It uses the SAME notion of decoration as criterionMarkerRe's `[^*\p{L}\p{N}]*`: anything
// that is not a letter or a digit. An earlier version tested `r < 0x2000`, which treated
// accented Latin, Greek, Cyrillic, Hebrew, and Arabic letters as decoration to skip while
// treating some symbols as token characters — two different answers to one question, in
// two places that must agree.
func stripLeadingDecoration(s string) string {
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return s[i:]
		}
	}
	return ""
}

// InvalidCriterionStateError is the shared rejection, naming the legal set so a reader can
// self-correct without opening the source.
func InvalidCriterionStateError(word string) error {
	return fmt.Errorf("%w: unknown criterion state %q — expected one of: %s",
		ErrValidation, word, strings.Join(CriterionSuffixStates(), ", "))
}
