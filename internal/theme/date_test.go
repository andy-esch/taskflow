package theme

import (
	"testing"
	"time"
)

func TestRelativeDate(t *testing.T) {
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	cases := map[string]string{
		"2026-06-09": "today",
		"2026-06-08": "yesterday",
		"2026-06-06": "3d ago",
		"2026-05-26": "2w ago",
		"2026-03-01": "3mo ago",
		"2024-06-09": "2y ago",
		"":           "", // empty
		"not-a-date": "", // unparseable
	}
	for in, want := range cases {
		if got := relativeDateFrom(in, now); got != want {
			t.Errorf("relativeDateFrom(%q) = %q, want %q", in, got, want)
		}
	}
}
