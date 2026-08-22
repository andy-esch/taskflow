# Demo: the ATLAS — cross-space navigation. Launched from a directory with no
# planning repo above it, `ui` lands on the atlas: one card per durable planning
# identity, every registered entry point beneath it, and reversible navigation
# into a space and back without restarting the TUI.
#
# Setup is a script rather than an inline Hide line because this tape needs
# something no other tape does: several registered spaces, two of which share a
# planning identity, in a throwaway home registry. See atlas-setup.sh — it never
# touches the recorder's real spaces.toml. Regenerate with `just gifs`.
Output assets/atlas.gif
Set Shell bash
Set FontSize 16
Set Width 1200
Set Height 760
Set Theme "Dracula"
Set Padding 16

Hide
Type "source assets/vhs/atlas-setup.sh && clear"
Enter
Sleep 1s
Show

# No planning repo here — so `ui` opens the atlas rather than failing discovery.
Type "tskflwctl ui"
Enter
Sleep 2.5s

# bike-workshop is one logical space reached two ways: the direct planning
# checkout and an impl repo pointing at it. h/l selects between them.
Type "l"
Sleep 1.5s
Type "h"
Sleep 1.2s

# o cycles the card order (name → activity → registered), O reverses it.
Type "o"
Sleep 1.2s
Type "O"
Sleep 1.2s
Type "O"
Sleep 800ms
Type "o"
Sleep 800ms
Type "o"
Sleep 1s

# Enter the kitchen space — a different planning identity entirely.
Type "j"
Sleep 800ms
Enter
Sleep 2.5s

# Its own dashboard, tabs, and rollups — a full session, not a preview.
Type "]"
Sleep 2s

# `a` returns to the atlas with the space still open behind it…
Type "a"
Sleep 2s

# …and entering the other space restores ITS own tabs and cursor.
Type "k"
Sleep 800ms
Enter
Sleep 2.5s
Type "]"
Sleep 2s
Type "]"
Sleep 2s

Type "q"
Sleep 1s
