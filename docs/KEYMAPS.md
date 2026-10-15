# Keymaps

Profiles are compatibility maps. They are not claims that axam matches Sublime Text, VS Code, or Notepad++. A terminal cannot deliver every desktop shortcut. Command, Super, and macOS Command are not bound.

Run `axam keymap` or `axam keymap vscode` to print the active table. Inside the editor, F1 opens the command palette and the Terminal Compatibility Report lists what this session can see. The in-app keymap buffer is the Keymap command.

F10 opens the bottom command menu independently of the selected profile. Menu
items display the shortest binding from that profile and execute the same
command IDs. F10 is handled as an interaction key rather than a remappable
editing command.

User bindings in `config.toml` override a profile. An empty command unbinds a chord.

```toml
keymap_profile = "termstudio"

[bindings]
"ctrl+q" = "file.quit"
"ctrl+k ctrl+x" = ""
```

Duplicate chords that name two different commands are rejected when the profile is built. A user binding replaces the previous command for that exact chord.

## Shared bindings

These exist in every profile. Ctrl+K starts a chord. Escape cancels it.

| Shortcut | Command | Delivery |
| --- | --- | --- |
| Ctrl+N | New file | Ctrl+letter, widely delivered |
| Ctrl+O | Open | widely delivered |
| Ctrl+S | Save | widely delivered |
| Ctrl+W | Close tab | widely delivered |
| Ctrl+Tab / Ctrl+Shift+Tab | Next / previous tab | depends on the terminal |
| Ctrl+Z / Ctrl+Y / Ctrl+Shift+Z | Undo / redo | Ctrl+Y is not Ctrl+Shift+Y |
| Ctrl+X / Ctrl+C / Ctrl+V | Cut / copy / paste | editor clipboard always; OSC 52 is not confirmed |
| Ctrl+A | Select all | widely delivered |
| Ctrl+F | Find | widely delivered |
| Ctrl+H | Replace | often stolen by Backspace (byte 0x08). Use the palette |
| Ctrl+G | Go to line | widely delivered |
| Ctrl+/ and Ctrl+_ | Toggle comment | Ctrl+/ may arrive as Ctrl+_ |
| Tab / Shift+Tab | Indent / outdent | widely delivered |
| Arrows, Shift+arrows | Move / select | Shift+arrow is CSI `1;2A` style and is the tested path |
| Ctrl+Left/Right, Ctrl+Shift+Left/Right | Word move / select | common, not universal |
| Ctrl+Home/End, Ctrl+Shift+Home/End | Document edges | common |
| Alt+Up/Down | Move line | Alt may be Meta on some Mac terminals |
| F1, Alt+X | Command palette | F1 is the dependable palette key |
| Ctrl+B | Toggle explorer | |
| Ctrl+\ | Split pane | |
| Ctrl+K then Ctrl+S | Save as | fallback when Ctrl+Shift+S is not delivered |
| Ctrl+K then Ctrl+P | Command palette | |
| Ctrl+K then Ctrl+O | Quick open | |
| Ctrl+K then Ctrl+C | Toggle comment | |
| Ctrl+K then Ctrl+L | Delete line | |
| Ctrl+K then Ctrl+D | Duplicate line | |
| Ctrl+K then Ctrl+T | Reopen closed | |
| Ctrl+K then Ctrl+Q | Quit | |
| Ctrl+K then Ctrl+F | Project search | |
| Ctrl+K then Ctrl+G | Keymap | |

Ctrl+Shift+letter is listed below only as an enhanced-keyboard binding. Legacy terminals collapse it into Ctrl+letter. Do not treat it as supported on Terminal.app or a stock SSH session unless `axam doctor` and a manual key check say the sequence arrived.

## TermStudio and Sublime

Copy and cut with an empty selection take the whole line.

| Shortcut | Command |
| --- | --- |
| Ctrl+D | Select next occurrence |
| Ctrl+L | Select line |
| Ctrl+P | Quick open |
| Ctrl+Q | Quit |
| Ctrl+Shift+P | Command palette (enhanced keyboard) |
| Ctrl+Shift+S | Save as (enhanced keyboard) |
| Ctrl+Shift+T | Reopen (enhanced keyboard) |
| Ctrl+Shift+K | Delete line (enhanced keyboard) |
| Ctrl+Shift+D | Duplicate line (enhanced keyboard) |
| Ctrl+Shift+L | Select all occurrences (enhanced keyboard) |
| Ctrl+Shift+F | Project search (enhanced keyboard) |
| Ctrl+Shift+Up/Down | Move line |
| Alt+Shift+Up/Down | Add cursor above / below |

## VS Code

Same copy-line behavior. Differences that matter:

| Shortcut | Command |
| --- | --- |
| Alt+Shift+Up/Down | Duplicate line |
| Ctrl+Alt+Up/Down | Add cursor above / below |
| Ctrl+Shift+Up/Down | not bound to move-line; use Alt+Up/Down |

Ctrl+P is quick open, as in VS Code. Ctrl+Shift+P is the palette only when the terminal sends Ctrl+Shift+P.

## Notepad++

Empty selection does not copy or cut the whole line.

| Shortcut | Command |
| --- | --- |
| Ctrl+D | Duplicate line |
| Ctrl+L | Delete line |
| Ctrl+Q | Toggle comment |
| Ctrl+P | not bound (Notepad++ uses it for print) |
| Ctrl+F3 | Select next occurrence |

Quick open is Ctrl+K Ctrl+O or the command palette.
