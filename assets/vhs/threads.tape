# Demo: Thread graph comprehension in the TUI. A touring-bike release converges
# parallel workshop work into one loaded shakedown, with a missing rack mount
# retained as an external gate. Regenerate with `just gifs`.
Output assets/threads.gif
Set Shell bash
Set FontSize 16
Set Width 1200
Set Height 760
Set Theme "Dracula"
Set Padding 16

Hide
Type "cd assets/demo-threads && clear"
Enter
Show

Type "tskflwctl ui"
Enter
Sleep 2s

# Jump straight to the focused entity without wandering through unrelated tabs.
Type ":threads"
Enter
Sleep 2s

# Read the Thread's release intent, then switch to its graph-derived waves.
Type "l"
Sleep 1.8s
Type "v"
Sleep 3s

# Start at the external gate, then walk into Wave 2's eligible wheel work.
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 500ms
Type "k"
Sleep 400ms
Type "j"
Sleep 1.2s

# Stable identity navigation opens the selected task and returns to the graph.
Enter
Sleep 2.2s
Ctrl+O
Sleep 1.2s
Type "l"
Sleep 500ms
Type "v"
Sleep 2.2s

Type "q"
Sleep 600ms
