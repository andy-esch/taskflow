package domain

import "testing"

// Valid-body regressions must explicitly require a successful read now that
// structural diagnostics cannot be confused with empty criterion lists.
func countCriteria(t testing.TB, body string) ACCount {
	t.Helper()
	count, err := CountAcceptanceCriteria(body)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func listCriteria(t testing.TB, body string) []Criterion {
	t.Helper()
	criteria, err := ListAcceptanceCriteria(body)
	if err != nil {
		t.Fatal(err)
	}
	return criteria
}

func tallyCriteria(t testing.TB, body string) []CriterionCount {
	t.Helper()
	counts, err := TallyCriteria(body)
	if err != nil {
		t.Fatal(err)
	}
	return counts
}
