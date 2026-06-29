// Package design is the single source of truth for the project's CONCRETE color
// decisions: the named themes and the palette tokens every surface draws from.
//
// It sits one layer below internal/theme. theme decides which *semantic slot* a
// domain value maps to (a status -> Yellow) and stays domain-only. design decides
// what that slot LOOKS LIKE under a given theme (Yellow -> #c9d364 truecolor /
// ANSI slot 3) and owns the non-semantic chrome (accent, borders, selection, the
// bar gradient) the 7-slot semantic enum can't express. The CLI render layer, the
// TUI, the picker, and the progress bars all resolve their colors through a
// Palette so there is one place a hue is decided — and never a stray literal.
//
// Dependency direction: design imports theme (to key its semantic slots by
// theme.Color); theme must NEVER import design.
package design

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/theme"
)

// Hue is one palette color carried in BOTH techs from a single definition: the
// truecolor Hex (the TUI / huh / glamour use it; lipgloss downsamples it for
// low-color terminals) and an explicit ANSI slot for the CLI's hand-written SGR
// path. The ANSI slot is a 0..15 color index (0..7 standard, 8..15 bright); the
// CLI maps it to an SGR code, the TUI ignores it. Carrying the slot explicitly is
// what keeps a neon hue degrading to a CHOSEN ANSI color rather than whatever a
// runtime nearest-color search picks (a base16 scheme already names this mapping).
// ANSI == NoANSI marks "no color" (the theme.ColorNone slot).
type Hue struct {
	Hex  string // "#rrggbb"; "" when the slot carries no truecolor
	ANSI int    // 0..15 color index, or NoANSI
}

// NoANSI is the sentinel ANSI slot for "emit no color" (theme.ColorNone).
const NoANSI = -1

// Color returns the truecolor value for lipgloss-based surfaces (TUI/huh/glamour).
// An empty Hex yields lipgloss's no-color, so an unstyled slot renders plain.
func (h Hue) Color() color.Color { return lipgloss.Color(h.Hex) }

// Palette is every color decision for one theme on one background.
type Palette struct {
	// Semantic — one Hue per theme.Color slot, so the existing Status/Bucket/
	// Priority/... tokens resolve to a concrete color through Of. A map (not a fixed
	// array) so a newly-added theme.Color can never silently index out of bounds.
	Semantic map[theme.Color]Hue

	// Chrome — structural tokens with no semantic meaning. The ANSI slot on these is
	// currently UNUSED: only the Semantic slots' ANSI feeds the CLI's SGR path, while
	// the TUI renders chrome in truecolor and lets lipgloss downsample.
	Accent       Hue // focus/selection accent (the neon signature)
	BorderActive Hue // focused pane / overlay border
	BorderIdle   Hue // unfocused border
	Danger       Hue // destructive borders + labels
	Heading      Hue // section headings (dashboard/help/action)
	Match        Hue // find match background
	MatchCurrent Hue // current find match background
	MatchFg      Hue // text drawn over a match highlight
	Track        Hue // progress empty-track tone

	// Gradient — the rollup bar fill stops (the deliberate truecolor exception;
	// purple -> cyan -> pink). Degrades per-cell on low-color terminals.
	Gradient []Hue

	// Markdown is the glamour standard-style name for this background.
	Markdown string
}

// Of resolves a semantic theme.Color to its palette Hue. An unmapped slot (a new
// theme.Color a palette hasn't filled) degrades to "no color" rather than panicking.
func (p Palette) Of(c theme.Color) Hue {
	if h, ok := p.Semantic[c]; ok {
		return h
	}
	return Hue{ANSI: NoANSI}
}

// Theme is a named palette pair: the dark- and light-background variants. A theme
// is selected by name; the active background (detected per surface) picks which
// Palette of the pair renders, via For.
type Theme struct {
	Name  string
	Dark  Palette
	Light Palette
}

// For picks the background-appropriate palette.
func (t Theme) For(darkBG bool) Palette {
	if darkBG {
		return t.Dark
	}
	return t.Light
}
