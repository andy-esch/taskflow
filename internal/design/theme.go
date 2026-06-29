package design

import (
	"sort"

	"github.com/andy-esch/taskflow/internal/theme"
)

// The built-in theme registry. Themes lean on ESTABLISHED schemes rather than
// hand-rolled values: the dark default is the base16 "Synth Midnight Terminal
// Dark" scheme (Michaël Ball) with one substitution — the danger-red is taken
// from base16 "Outrun Dark" (#FF4242) because Synth's own red fails WCAG contrast
// for small text. The light fallback is Catppuccin Latte. ANSI slots follow the
// base16 → terminal convention (base0E magenta -> bright magenta 13, etc.), so the
// neon identity survives the CLI's 16-color path instead of being guessed at.
//
// Slot indices (0..15): 0 black · 1 red · 2 green · 3 yellow · 4 blue · 5 magenta
// · 6 cyan · 7 white · 8..15 the bright variants.

// neonDark — "neon-night". base16 Synth Midnight (+ Outrun red), near-black bg so
// the accents glow.
var neonDark = Palette{
	Semantic: map[theme.Color]Hue{
		theme.ColorNone:   {Hex: "", ANSI: NoANSI},
		theme.ColorRed:    {Hex: "#FF4242", ANSI: 1}, // Outrun base08 (legible swap)
		theme.ColorGreen:  {Hex: "#06ea61", ANSI: 2}, // base0B
		theme.ColorYellow: {Hex: "#c9d364", ANSI: 3}, // base0A
		theme.ColorBlue:   {Hex: "#03aeff", ANSI: 4}, // base0D
		theme.ColorCyan:   {Hex: "#42fff9", ANSI: 6}, // base0C
		theme.ColorGray:   {Hex: "#a3a5a6", ANSI: 8}, // base04 (bright black)
	},
	Accent:       Hue{"#ea5ce2", 13}, // base0E magenta -> bright magenta
	BorderActive: Hue{"#ea5ce2", 13},
	BorderIdle:   Hue{"#474849", 8}, // base03
	Danger:       Hue{"#FF4242", 1},
	Heading:      Hue{"#ea5ce2", 13},
	Match:        Hue{"#c9d364", 3},  // yellow bg
	MatchCurrent: Hue{"#ea5ce2", 13}, // accent bg for the current hit
	MatchFg:      Hue{"#050608", 0},  // near-black text over a highlight
	Track:        Hue{"#474849", 8},  // base03 empty track
	Gradient: []Hue{
		{"#b026ff", 5},  // neon purple
		{"#00e5ff", 14}, // neon cyan
		{"#ff2ec4", 13}, // neon pink
	},
	Markdown: theme.MarkdownStyleDark,
}

// latteAA is the shared LIGHT palette: Catppuccin Latte with the semantic accents
// DARKENED from Latte's defaults to clear WCAG AA (>=4.5:1) on the light bg —
// statusText/priorityText color the LABEL text (not just the glyph), so the text
// itself must be legible, and Latte's own green/yellow/teal fail AA at small size.
// Both the neon and catppuccin themes use it as their light variant (most terminals
// run dark; the themes diverge on the dark side + the accent).
var latteAA = Palette{
	Semantic: map[theme.Color]Hue{
		theme.ColorNone:   {Hex: "", ANSI: NoANSI},
		theme.ColorRed:    {Hex: "#d20f39", ANSI: 1}, // Latte red (4.8:1)
		theme.ColorGreen:  {Hex: "#2e7d1f", ANSI: 2}, // darkened from Latte green for AA (4.6:1)
		theme.ColorYellow: {Hex: "#8a6000", ANSI: 3}, // dark amber: AA-legible "yellow" on light (5.0:1)
		theme.ColorBlue:   {Hex: "#2258cc", ANSI: 4}, // darkened from Latte blue for AA (5.6:1)
		theme.ColorCyan:   {Hex: "#0e6e74", ANSI: 6}, // darkened Latte teal for AA (5.3:1)
		theme.ColorGray:   {Hex: "#6c6f85", ANSI: 8}, // subtext0
	},
	Accent:       Hue{"#8839ef", 5}, // mauve (4.8:1)
	BorderActive: Hue{"#8839ef", 5},
	BorderIdle:   Hue{"#9ca0b0", 8}, // overlay0
	Danger:       Hue{"#d20f39", 1},
	Heading:      Hue{"#8839ef", 5},
	Match:        Hue{"#df8e1d", 3},
	MatchCurrent: Hue{"#8839ef", 5},
	MatchFg:      Hue{"#eff1f5", 15}, // base
	Track:        Hue{"#bcc0cc", 7},  // surface1
	Gradient: []Hue{
		{"#8839ef", 5},  // mauve
		{"#209fb5", 6},  // sapphire-ish
		{"#ea76cb", 13}, // pink
	},
	Markdown: theme.MarkdownStyleLight,
}

// mochaDark — Catppuccin Mocha: the soft pastel dark variant, a deliberate contrast
// to neon's synthwave. Canonical Catppuccin Mocha hexes; accent is mauve. ANSI slots
// follow the same terminal convention as neon (mauve -> bright magenta 13).
var mochaDark = Palette{
	Semantic: map[theme.Color]Hue{
		theme.ColorNone:   {Hex: "", ANSI: NoANSI},
		theme.ColorRed:    {Hex: "#f38ba8", ANSI: 1}, // red
		theme.ColorGreen:  {Hex: "#a6e3a1", ANSI: 2}, // green
		theme.ColorYellow: {Hex: "#f9e2af", ANSI: 3}, // yellow
		theme.ColorBlue:   {Hex: "#89b4fa", ANSI: 4}, // blue
		theme.ColorCyan:   {Hex: "#89dceb", ANSI: 6}, // sky
		theme.ColorGray:   {Hex: "#9399b2", ANSI: 8}, // overlay2
	},
	Accent:       Hue{"#cba6f7", 13}, // mauve
	BorderActive: Hue{"#cba6f7", 13},
	BorderIdle:   Hue{"#45475a", 8}, // surface1
	Danger:       Hue{"#f38ba8", 1}, // red
	Heading:      Hue{"#cba6f7", 13},
	Match:        Hue{"#f9e2af", 3},  // yellow bg
	MatchCurrent: Hue{"#cba6f7", 13}, // mauve bg for the current hit
	MatchFg:      Hue{"#1e1e2e", 0},  // base — dark text over a highlight
	Track:        Hue{"#313244", 8},  // surface0 empty track
	Gradient: []Hue{
		{"#cba6f7", 13}, // mauve
		{"#89b4fa", 4},  // blue
		{"#f5c2e7", 13}, // pink
	},
	Markdown: theme.MarkdownStyleDark,
}

var (
	// neon is the default theme: neon-night (dark) + the AA Latte light.
	neon = Theme{Name: "neon", Dark: neonDark, Light: latteAA}
	// catppuccin is the established-library alternative: Mocha (dark) + the AA Latte
	// light (Latte is Catppuccin's own light flavor).
	catppuccin = Theme{Name: "catppuccin", Dark: mochaDark, Light: latteAA}
)

// registry holds the built-in themes by name.
var registry = map[string]Theme{
	neon.Name:       neon,
	catppuccin.Name: catppuccin,
}

// Default is the project's default theme (neon / 80s).
func Default() Theme { return neon }

// Names returns the registered theme names, sorted — so `theme list` and its --json
// output are byte-stable.
func Names() []string {
	out := make([]string, 0, len(registry))
	for name := range registry {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Lookup returns the named theme, or the default with ok=false when name is empty
// or unknown — callers degrade to Default rather than erroring on a bad config.
func Lookup(name string) (Theme, bool) {
	if t, ok := registry[name]; ok {
		return t, true
	}
	return Default(), false
}
