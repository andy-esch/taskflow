package domain

import (
	"errors"
	"strings"
	"testing"
)

const multiBody = "# T\n\n## Acceptance criteria\n\n- [ ] one\n- [ ] two\n- [ ] three\n- [ ] four\n"

func checkedCount(body string) int {
	return strings.Count(body, "- [x]")
}

func TestSetCriteriaState_FlipsEveryIndexInOnePass(t *testing.T) {
	out, err := SetCriteriaState(multiBody, []int{1, 2, 4}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := checkedCount(out); got != 3 {
		t.Errorf("checked %d boxes, want 3:\n%s", got, out)
	}
	if !strings.Contains(out, "- [ ] three") {
		t.Errorf("criterion 3 was not requested and must be untouched:\n%s", out)
	}
}

func TestSetCriteriaState_DeduplicatesAndIgnoresOrder(t *testing.T) {
	forward, err := SetCriteriaState(multiBody, []int{1, 2}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := SetCriteriaState(multiBody, []int{2, 1, 2, 1}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if forward != reverse {
		t.Errorf("order and duplicates must not matter:\n%q\n%q", forward, reverse)
	}
}

// An out-of-range index anywhere rejects the whole request. Checked by comparing
// against the input: a half-applied body is the failure mode this guards.
func TestSetCriteriaState_OutOfRangeLeavesBodyUntouched(t *testing.T) {
	for _, ns := range [][]int{{0}, {9}, {1, 9}, {9, 1}, {-1, 2}} {
		out, err := SetCriteriaState(multiBody, ns, CriterionMet, "")
		if !errors.Is(err, ErrValidation) {
			t.Errorf("%v should be a validation error, got %v", ns, err)
		}
		if out != "" {
			t.Errorf("%v returned a body alongside its error: %q", ns, out)
		}
	}
	// And nothing was written even for the valid prefix: re-running the valid part
	// alone is the only thing that changes anything.
	out, err := SetCriteriaState(multiBody, []int{1}, CriterionMet, "")
	if err != nil || checkedCount(out) != 1 {
		t.Errorf("the valid single flip should still work: %v\n%s", err, out)
	}
}

func TestSetCriteriaState_EmptyIndexListIsValidationError(t *testing.T) {
	if _, err := SetCriteriaState(multiBody, nil, CriterionMet, ""); !errors.Is(err, ErrValidation) {
		t.Errorf("an empty index list should be a validation error, got %v", err)
	}
}

// All-already-in-state returns the input body byte-identically, which is how the
// caller knows to skip the write entirely.
func TestSetCriteriaState_NoOpReturnsInputUnchanged(t *testing.T) {
	once, err := SetCriteriaState(multiBody, []int{1, 2}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	twice, err := SetCriteriaState(once, []int{1, 2}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if twice != once {
		t.Errorf("a repeated set must be byte-identical:\n%q\n%q", once, twice)
	}
}

// The single-index entry point is now a wrapper; it must behave exactly as before.
func TestSetCriterionState_StillWrapsTheSingleCase(t *testing.T) {
	viaOne, err := SetCriterionState(multiBody, 2, CriterionDeferred, "waiting on the ADR")
	if err != nil {
		t.Fatal(err)
	}
	viaMany, err := SetCriteriaState(multiBody, []int{2}, CriterionDeferred, "waiting on the ADR")
	if err != nil {
		t.Fatal(err)
	}
	if viaOne != viaMany {
		t.Errorf("the wrapper diverged from the multi path:\n%q\n%q", viaOne, viaMany)
	}
	if !strings.Contains(viaOne, "**deferred:** waiting on the ADR") {
		t.Errorf("the reason suffix was lost:\n%s", viaOne)
	}
}

// A state needing a reason still demands one, and a met state still clears the
// suffix — across a list, not just a single index.
func TestSetCriteriaState_ReasonRulesHoldAcrossAList(t *testing.T) {
	if _, err := SetCriteriaState(multiBody, []int{1, 2}, CriterionDeferred, ""); !errors.Is(err, ErrValidation) {
		t.Errorf("a reason-needing state must still demand one, got %v", err)
	}
	deferred, err := SetCriteriaState(multiBody, []int{1, 2}, CriterionDeferred, "blocked")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(deferred, "**deferred:** blocked") != 2 {
		t.Errorf("both criteria should carry the suffix:\n%s", deferred)
	}
	met, err := SetCriteriaState(deferred, []int{1, 2}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(met, "deferred:") {
		t.Errorf("met must clear every suffix it supersedes:\n%s", met)
	}
	if checkedCount(met) != 2 {
		t.Errorf("both boxes should be checked:\n%s", met)
	}
}

// Checkboxes inside fenced code blocks are not criteria, so indices must skip them.
func TestSetCriteriaState_IgnoresFencedCheckboxes(t *testing.T) {
	body := "# T\n\n## Acceptance criteria\n\n- [ ] real one\n\n```\n- [ ] not a criterion\n```\n\n- [ ] real two\n"
	out, err := SetCriteriaState(body, []int{1, 2}, CriterionMet, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- [ ] not a criterion") {
		t.Errorf("a fenced checkbox must not be touched:\n%s", out)
	}
	if checkedCount(out) != 2 {
		t.Errorf("both real criteria should be checked:\n%s", out)
	}
}

func TestSetCriteriaState_NoAcceptanceSection(t *testing.T) {
	if _, err := SetCriteriaState("# T\n\nno criteria\n", []int{1}, CriterionMet, ""); !errors.Is(err, ErrValidation) {
		t.Errorf("a body with no acceptance section should be a validation error, got %v", err)
	}
}
