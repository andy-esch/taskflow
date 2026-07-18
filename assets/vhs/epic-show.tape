# Demo: epic show — Tab-complete the epic id, then the status-grouped lipgloss/v2
# tree. Regenerate with `just gifs`. Uses zsh: the id completion needs a modern
# shell, and macOS ships bash 3.2 (too old for cobra's bash completion); VHS
# renders every shell with the same `> ` prompt, so it matches the other demos.
Output assets/epic-show.gif
Set Shell zsh
Set FontSize 15
Set Width 1200
Set Height 800
Set Theme "Dracula"
Set Padding 16

# Load shell completion off-camera so the Tab below expands the epic id.
Hide
Type "autoload -Uz compinit && compinit -u && source <(tskflwctl completion zsh) && clear"
Enter
Sleep 1s
Show

# Type a short prefix, Tab to complete the full id, then run it.
Type "tskflwctl epic show 18-tui"
Sleep 500ms
Tab
Sleep 1.2s
Enter
Sleep 4s
