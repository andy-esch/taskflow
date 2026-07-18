# Demo: `audit show` — the segmented finding bar (done / in-progress / dropped /
# open, banded by distinct glyphs) above the status-grouped finding tree.
# Regenerate with `just gifs`. Runs against the curated assets/demo-planning
# fixture (cd is hidden so the recorded command stays clean). Uses zsh so the
# slug can be Tab-completed: completion needs a modern shell (macOS ships bash
# 3.2, too old for cobra's bash completion); VHS gives every shell the same `> `
# prompt.
Output assets/audit-show.gif
Set Shell zsh
Set FontSize 16
Set Width 1100
Set Height 760
Set Theme "Dracula"
Set Padding 16

# cd into the fixture and load completion, off-camera.
Hide
Type "cd assets/demo-planning && autoload -Uz compinit && compinit -u && source <(tskflwctl completion zsh) && clear"
Enter
Sleep 1s
Show

# Tab through command, subcommand, and the audit slug from its date prefix.
Type "tskflwctl au"
Sleep 250ms
Tab
Sleep 350ms
Type "sh"
Sleep 250ms
Tab
Sleep 350ms
Type "2026-06-2"
Sleep 300ms
Tab
Sleep 800ms
Enter
Sleep 4.5s
