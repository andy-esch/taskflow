package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

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
	st := NewStyle(false) // plain: the bubbles progress bar, ANSI-stripped by the gate
	// The bubbles progress component rounds the filled cells (0.77×10 → 8), where
	// the old hand-rolled bar floored (→ 7); otherwise identical glyphs.
	cases := map[[2]int]string{
		{50, 10}: "█████░░░░░",
		{0, 4}:   "░░░░",
		{100, 4}: "████",
		{77, 10}: "████████░░",
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
	plain := ansi.Strip(b.String())
	for _, ln := range strings.Split(strings.TrimRight(plain, "\n"), "\n") {
		if !strings.Contains(ln, "  ") { // a 2-space gutter must survive
			t.Errorf("row not padded: %q", ln)
		}
	}
}

// TestVisibleWidth_DisplayCells pins the wide-rune fix: CJK/emoji occupy two
// display cells, so a rune count under-measured them and shifted columns.
func TestVisibleWidth_DisplayCells(t *testing.T) {
	if w := visibleWidth("日本語"); w != 6 {
		t.Errorf("CJK width should count cells, got %d want 6", w)
	}
	if w := visibleWidth("\x1b[31mab\x1b[0m"); w != 2 {
		t.Errorf("ANSI escapes must not count, got %d want 2", w)
	}
	if got := truncate("日本語のタイトル", 7); ansiStripWidth(got) > 7 {
		t.Errorf("truncate must respect display cells, got %q (%d cells)", got, ansiStripWidth(got))
	}
}

func ansiStripWidth(s string) int { return visibleWidth(s) }

// TestWriteTable_ClampsWideNonFinalColumn pins M7: a wide NON-final cell (a long
// slug/component) must not push a human-table row past maxWidth — the last-column
// shrink alone doesn't cover it, so the whole line is clamped.
func TestWriteTable_ClampsWideNonFinalColumn(t *testing.T) {
	const maxWidth = 40
	header := []string{"SLUG", "DESC"}
	rows := [][]string{{strings.Repeat("x", 100), "short"}} // wide first (non-final) column
	var buf bytes.Buffer
	writeTable(&buf, maxWidth, header, rows)
	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		if w := ansi.StringWidth(line); w > maxWidth {
			t.Errorf("line exceeds maxWidth %d (got %d): %q", maxWidth, w, line)
		}
	}
}

// TestWriteTable_ClampsColoredCell: the clamp is ANSI-aware, so a colored wide cell
// is clipped by display width (escapes don't count) without bleeding past maxWidth.
func TestWriteTable_ClampsColoredCell(t *testing.T) {
	const maxWidth = 30
	st := NewStyle(true) // color on
	header := []string{"SLUG", "DESC"}
	rows := [][]string{{st.Bold(strings.Repeat("x", 100)), st.Dim("d")}}
	var buf bytes.Buffer
	writeTable(&buf, maxWidth, header, rows)
	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		if w := ansi.StringWidth(line); w > maxWidth {
			t.Errorf("colored line exceeds maxWidth %d (got %d): %q", maxWidth, w, line)
		}
	}
}

// TestWriteTable_NoClampWhenPiped: maxWidth <= 0 (piped) keeps rows full-width.
func TestWriteTable_NoClampWhenPiped(t *testing.T) {
	wide := strings.Repeat("x", 100)
	var buf bytes.Buffer
	writeTable(&buf, 0, []string{"SLUG", "DESC"}, [][]string{{wide, "d"}})
	if !strings.Contains(buf.String(), wide) {
		t.Errorf("piped output (maxWidth=0) must not clamp the wide cell:\n%s", buf.String())
	}
}
