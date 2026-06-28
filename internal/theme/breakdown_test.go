package theme

import (
	"fmt"
	"testing"
)

func TestBreakdown(t *testing.T) {
	seg := func(s string) string { return s }
	more := func(n int) string { return fmt.Sprintf("+%d more", n) }

	// No cap (max <= 0): every segment, joined with sep; more is never called.
	if got := Breakdown([]string{"a", "b", "c"}, " · ", 0, seg, nil); got != "a · b · c" {
		t.Errorf("uncapped = %q, want %q", got, "a · b · c")
	}
	// At/under the cap: all shown, no tail.
	if got := Breakdown([]string{"a", "b"}, ", ", 3, seg, more); got != "a, b" {
		t.Errorf("under cap = %q, want %q", got, "a, b")
	}
	// Over the cap: max segments + a "+N more" tail with the dropped count.
	if got := Breakdown([]string{"a", "b", "c", "d"}, " ", 2, seg, more); got != "a b +2 more" {
		t.Errorf("over cap = %q, want %q", got, "a b +2 more")
	}
	// Empty input is the empty string.
	if got := Breakdown(nil, " · ", 0, seg, nil); got != "" {
		t.Errorf("empty = %q, want empty", got)
	}
}
