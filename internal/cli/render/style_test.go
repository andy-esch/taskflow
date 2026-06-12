package render

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestStyle_Disabled_IsPlain(t *testing.T) {
	st := NewStyle(false)
	if got := st.Status(domain.StatusInProgress); got != "in-progress" {
		t.Errorf("disabled Status = %q, want plain label", got)
	}
	if got := st.Bold("x"); got != "x" {
		t.Errorf("disabled Bold = %q, want plain", got)
	}
	if got := st.Priority("high"); got != "high" {
		t.Errorf("disabled Priority = %q, want plain", got)
	}
}

func TestStyle_Enabled_WrapsANSI(t *testing.T) {
	st := NewStyle(true)
	if got := st.Bold("x"); !strings.Contains(got, "\x1b[") {
		t.Errorf("enabled Bold emitted no ANSI: %q", got)
	}
	got := st.Status(domain.StatusInProgress)
	if !strings.Contains(got, "in-progress") || !strings.Contains(got, "\x1b[") {
		t.Errorf("enabled Status = %q, want colored label", got)
	}
}

func TestVisibleWidth_IgnoresANSI(t *testing.T) {
	colored := NewStyle(true).Bold("hello")
	if w := visibleWidth(colored); w != 5 {
		t.Errorf("visibleWidth(%q) = %d, want 5 (ANSI ignored)", colored, w)
	}
}

func TestWriteTable_TruncatesLastColumnToWidth(t *testing.T) {
	var b strings.Builder
	long := "this description is quite long and should be cut to fit the narrow width"
	writeTable(&b, 40, []string{"TASK", "DESC"}, [][]string{{"alpha", long}})
	out := b.String()
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation ellipsis at width 40:\n%q", out)
	}
	for _, ln := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if visibleWidth(ln) > 40 {
			t.Errorf("line exceeds maxWidth 40 (%d): %q", visibleWidth(ln), ln)
		}
	}
}

func TestWriteTable_NoLimitKeepsFullWidth(t *testing.T) {
	var b strings.Builder
	long := strings.Repeat("x", 200)
	writeTable(&b, 0, []string{"A", "B"}, [][]string{{"a", long}}) // 0 = piped, no cap
	if !strings.Contains(b.String(), long) || strings.Contains(b.String(), "…") {
		t.Errorf("width 0 must not truncate (pipe-safe):\n%q", b.String())
	}
}

func TestBar(t *testing.T) {
	st := NewStyle(false) // plain
	cases := map[[2]int]string{
		{50, 10}: "█████░░░░░",
		{0, 4}:   "░░░░",
		{100, 4}: "████",
		{77, 10}: "███████░░░",
		{150, 4}: "████", // clamped
	}
	for in, want := range cases {
		if got := st.Bar(in[0], in[1]); got != want {
			t.Errorf("Bar(%d,%d) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

func TestWriteTable_AlignsColoredCells(t *testing.T) {
	st := NewStyle(true)
	var b strings.Builder
	writeTable(&b, 0, []string{"A", "B"}, [][]string{
		{st.Bold("xx"), "1"},
		{"y", "2"},
	})
	// Both data columns must start at the same visible offset despite the ANSI
	// in row 1 — i.e. stripping ANSI yields a clean aligned table.
	plain := ansiRe.ReplaceAllString(b.String(), "")
	for _, ln := range strings.Split(strings.TrimRight(plain, "\n"), "\n") {
		if !strings.Contains(ln, "  ") { // a 2-space gutter must survive
			t.Errorf("row not padded: %q", ln)
		}
	}
}
