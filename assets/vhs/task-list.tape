# Demo: task list with column projection (-o table -c). Regenerate with `just gifs`.
Output assets/task-list.gif
Set Shell bash
Set FontSize 16
Set Width 1100
Set Height 620
Set Theme "Dracula"
Set Padding 16

Type "tskflwctl task list -o table -c slug,status,epic"
Sleep 700ms
Enter
Sleep 3s
