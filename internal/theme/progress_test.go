package theme

import "testing"

// TestProgressLabels pins the shared progress-composite number formats so the
// "%d%% vs %3d%%" and done/total drift the audit (H5) flagged can't reappear.
func TestProgressLabels(t *testing.T) {
	for _, c := range []struct{ name, got, want string }{
		{"PercentLabel", PercentLabel(7), "7%"},
		{"PercentLabel 100", PercentLabel(100), "100%"},
		{"PercentLabelPadded 7", PercentLabelPadded(7), "  7%"},
		{"PercentLabelPadded 70", PercentLabelPadded(70), " 70%"},
		{"PercentLabelPadded 100", PercentLabelPadded(100), "100%"},
		{"Counts", Counts(7, 12), "7/12"},
		{"Counts wide", Counts(115, 166), "115/166"},
	} {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}
