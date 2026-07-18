# Demo: interactive pickers — on a TTY, a missing required input PROMPTS instead
# of erroring (off a TTY it fails with the flag to pass). `task new` with no
# --epic opens an epic picker (type to filter, `> ` caret), then a tags text
# prompt. Prompts render to stderr, so stdout stays byte-identical to the
# flag-driven run. Regenerate with `just gifs`.
#
# Runs against a throwaway COPY of the fixture (in /tmp) so the demo can actually
# create the task without dirtying the committed assets/demo-planning tree.
Output assets/picker.gif
Set Shell bash
Set FontSize 16
Set Width 1100
Set Height 520
Set Theme "Dracula"
Set Padding 16

Hide
Type "rm -rf /tmp/tskflw-demo && cp -R assets/demo-planning /tmp/tskflw-demo && cd /tmp/tskflw-demo && clear"
Enter
Sleep 500ms
Show

# No --epic on a TTY → the epic picker opens (all epics first)…
Type 'tskflwctl task new "Swap brake pads"'
Sleep 500ms
Enter
Sleep 2s
# …type to filter it down, then Enter to choose.
Type "tour"
Sleep 1s
Enter
Sleep 1.2s
# Then the required tags — a free-text prompt (its own flag twin, --tags).
Type "brakes,safety"
Sleep 800ms
Enter
Sleep 2.5s
