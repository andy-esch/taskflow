package theme

import "fmt"

// The progress composite — "<bar>  <percent>  <done>/<total>" — is assembled per
// surface (a CLI table cell, a labeled `progress:` field, a TUI list row), so its
// LAYOUT is genuinely context-specific and not worth forcing into one renderer. Its
// bar and percent COLOR are already shared (progressbar.Render, theme.Percent); the
// remaining shared pieces are the percent and done/total NUMBER formats, kept here
// in one place so they can't drift (the "%d%% here, %3d%% there" inconsistency).
// Each surface still applies its own color (ANSI Style vs lipgloss) and bar width.

// PercentLabel renders a percent compactly ("7%") — for inline / prose contexts.
func PercentLabel(pct int) string { return fmt.Sprintf("%d%%", pct) }

// PercentLabelPadded right-justifies the percent to 3 digits ("  7%", " 70%",
// "100%") so it aligns in hand-laid-out columns / rows.
func PercentLabelPadded(pct int) string { return fmt.Sprintf("%3d%%", pct) }

// Counts renders a done/total rollup ("7/12"). Width-justification for aligned
// columns is the caller's concern (CLI tables pad cells; the TUI measures + pads).
func Counts(done, total int) string { return fmt.Sprintf("%d/%d", done, total) }
