# Shortcuts inside axam

The live list is the program, not this page. Profiles differ. See [KEYMAPS.md](KEYMAPS.md).

- F1 opens the command palette. Every palette row shows its shortcut when one is bound.
- F10 opens the bottom command menu. Left/Right chooses File, Edit,
  Selection, Find, View, Goto, or Help; Up/Down chooses an item; Enter runs
  it; Escape or F10 closes it. The menu is also clickable.
- The Keymap command opens a buffer of the active bindings.
- `axam keymap` prints the same table in the shell.
- Find: Ctrl+F. Replace: the palette, or Ctrl+H when the terminal does not use that byte for Backspace. Alt+C, Alt+W, and Alt+R toggle case, whole word, and regex. Enter jumps to the next match. Alt+Enter replaces the current match. Replace All is a palette command and is one undo step.
- Escape closes an overlay, cancels a Ctrl+K chord, or clears extra cursors and the selection.

A shortcut in the table is implemented and covered by the keymap tests. Whether a given terminal delivers the keys is a separate column: legacy Ctrl+letter and Shift+arrow are the tested baseline; Ctrl+Shift+letter is enhanced-keyboard only.
