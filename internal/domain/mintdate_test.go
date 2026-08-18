package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/id"
)

// The boundaries are the whole point of ValidateMintableDate, so pin them exactly: the
// first and last representable days are accepted and the days either side are not.
// Derived from id.MaxMillis rather than hardcoded, so widening the id layout moves the
// test with the code instead of failing it.
func TestValidateMintableDate_Boundaries(t *testing.T) {
	day := 24 * time.Hour
	maxDay := time.UnixMilli(id.MaxMillis).UTC().Truncate(day)
	cases := []struct {
		name  string
		date  string
		valid bool
	}{
		{"epoch day", "1970-01-01", true},
		{"day before epoch", "1969-12-31", false},
		{"last representable day", maxDay.Format(time.DateOnly), true},
		{"day after last representable", maxDay.Add(day).Format(time.DateOnly), false},
		{"a normal date", "2026-08-18", true},
		{"plausible year typo (1026)", "1026-06-15", false},
		{"plausible year typo (9026)", "9026-06-15", false},
		{"malformed still rejected", "2026-6-3", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMintableDate(tc.date)
			if tc.valid && err != nil {
				t.Errorf("ValidateMintableDate(%q) = %v, want nil", tc.date, err)
			}
			if !tc.valid {
				if err == nil {
					t.Fatalf("ValidateMintableDate(%q) = nil, want an error", tc.date)
				}
				if !errors.Is(err, ErrValidation) {
					t.Errorf("error must wrap ErrValidation for exit 11, got %v", err)
				}
			}
		})
	}
}

// The reason the guard exists: for every accepted date, sorting by minted id must equal
// sorting by date. This is the invariant an out-of-range date silently destroys.
func TestMintableDates_IDOrderMatchesDateOrder(t *testing.T) {
	dates := []string{
		"1970-01-01", "1970-01-02", "1999-12-31", "2000-01-01", "2024-02-29",
		"2026-01-03", "2026-06-23", "2026-08-18", "2100-05-05", "2247-01-01",
	}
	type pair struct{ date, eid string }
	pairs := make([]pair, 0, len(dates))
	for _, d := range dates {
		if err := ValidateMintableDate(d); err != nil {
			t.Fatalf("fixture date %q should be mintable: %v", d, err)
		}
		ts, _ := time.Parse(time.DateOnly, d)
		pairs = append(pairs, pair{d, id.NewAt(ts.UnixMilli())})
	}
	// dates is already ascending, so ids must be ascending too.
	for i := 1; i < len(pairs); i++ {
		if pairs[i-1].eid >= pairs[i].eid {
			t.Errorf("id order breaks chronology: %s->%s but %s >= %s",
				pairs[i-1].date, pairs[i].date, pairs[i-1].eid, pairs[i].eid)
		}
	}
	// And each id decodes back to its own date (UTC — a date-only string is UTC midnight).
	for _, p := range pairs {
		if got := id.Time(p.eid).UTC().Format(time.DateOnly); got != p.date {
			t.Errorf("id %s decodes to %s, want %s", p.eid, got, p.date)
		}
	}
}

// Representable is the id-layer predicate the domain guard delegates to; check the raw
// millisecond edges directly so a layout change is caught here too.
func TestRepresentable_MillisEdges(t *testing.T) {
	cases := []struct {
		ms    int64
		valid bool
	}{
		{0, true},
		{-1, false},
		{id.MaxMillis, true},
		{id.MaxMillis + 1, false},
	}
	for _, tc := range cases {
		if got := id.Representable(tc.ms); got != tc.valid {
			t.Errorf("Representable(%d) = %v, want %v", tc.ms, got, tc.valid)
		}
	}
}
