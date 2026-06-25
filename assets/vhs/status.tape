# Demo: the at-a-glance status dashboard — status counts, the in-progress set,
# epic rollup bars, and the Open-audits section with its segmented finding bar.
# Regenerate with `just gifs`. Runs against the curated assets/demo-planning
# fixture (cd is hidden so the recorded command stays clean).
Output assets/status.gif
Set Shell bash
Set FontSize 16
Set Width 1100
Set Height 720
Set Theme "Dracula"
Set Padding 16

Hide
Type "cd assets/demo-planning && clear"
Enter
Show

Type "tskflwctl status"
Sleep 700ms
Enter
Sleep 4s
