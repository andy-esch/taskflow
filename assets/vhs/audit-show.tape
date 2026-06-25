# Demo: `audit show` — the segmented finding bar (done / in-progress / dropped /
# open, banded by distinct glyphs) above the status-grouped finding tree.
# Regenerate with `just gifs`. Runs against the curated assets/demo-planning
# fixture (cd is hidden so the recorded command stays clean).
Output assets/audit-show.gif
Set Shell bash
Set FontSize 16
Set Width 1100
Set Height 760
Set Theme "Dracula"
Set Padding 16

Hide
Type "cd assets/demo-planning && clear"
Enter
Show

Type "tskflwctl audit show 2026-06-20-api-gateway"
Sleep 700ms
Enter
Sleep 4.5s
