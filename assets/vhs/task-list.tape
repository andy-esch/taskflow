# Demo: task list with column projection — built almost entirely from short
# prefixes + Tab (command, subcommand, -o value, and each -c column). Regenerate
# with `just gifs`. Uses zsh: completion needs a modern shell (macOS ships bash
# 3.2, too old for cobra's bash completion); VHS gives every shell the same `> `
# prompt, so it matches the other demos.
Output assets/task-list.gif
Set Shell zsh
Set FontSize 16
Set Width 1100
Set Height 620
Set Theme "Dracula"
Set Padding 16

# Load completion off-camera.
Hide
Type "autoload -Uz compinit && compinit -u && source <(tskflwctl completion zsh) && clear"
Enter
Sleep 1s
Show

# Build `task list -o table -c slug,status,epic` — Tab does the typing.
Type "tskflwctl ta"
Sleep 250ms
Tab
Sleep 350ms
Type "li"
Sleep 250ms
Tab
Sleep 350ms
Type "-o ta"
Sleep 250ms
Tab
Sleep 350ms
Type "-c sl"
Sleep 250ms
Tab
Sleep 250ms
Type ",st"
Sleep 250ms
Tab
Sleep 250ms
Type ",ep"
Sleep 250ms
Tab
Sleep 500ms
Enter
Sleep 3s
