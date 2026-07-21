# Demo (hero): the interactive TUI — tabs, status glyphs, rollup bars, and the
# audit finding tree. Regenerate with `just gifs`. Runs against the curated
# assets/demo-planning fixture (cd is hidden so the recorded commands stay clean).
Output assets/tui.gif
Set Shell bash
Set FontSize 16
Set Width 1200
Set Height 760
Set Theme "Dracula"
Set Padding 16

Hide
Type "cd assets/demo-planning && clear"
Enter
Show

Type "tskflwctl ui"
Enter
Sleep 1.5s

# Tasks tab — status glyphs across the roster
Type "]"
Sleep 1s

# Epics tab — neon rollup bars per epic
Type "]"
Sleep 1.5s

# Audits tab — bucket glyph + the segmented finding bar
Type "]"
Sleep 1s

# open the selected audit — segmented bar + the status-grouped finding tree
Enter
Sleep 1.5s

# scroll down
Down 10
Sleep 0.8s

Type "q"
Sleep 750ms
