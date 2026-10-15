# Manual verification

Automated tests cover the editor, the keymap, escape-sequence decoding, and a simulated 80×24 screen. They do not drive Terminal.app, iTerm2, or a real Linux tty. Those checks are manual. Judge them by the buffer and the selection, not by whether a key event was logged.

On each client, open a UTF-8 file that contains:

```
abcd
x
абвг
a👍b

```

## Checklist

| Check | Pass when |
| --- | --- |
| Launch with no arguments | Empty state, status shows the profile, terminal is restored on Ctrl+Q |
| `axam notes1.txt`, then close unchanged | `notes1.txt` never appears on disk |
| `axam notes1.txt`, type text, save | file appears with exactly that text |
| F10 bottom menu | menu opens upward; arrows and Enter run commands; Escape closes |
| Mouse on File → New File | the File popup opens and New File creates an untitled tab |
| Shift+Right twice on `abcd` | Selection is `ab` or the two characters from the caret, and the anchor stays put if you then Shift+Left |
| Shift+Down from column 3 of `abcd` through `x` onto `абвг` | The caret returns to column 3 on the Cyrillic line. It does not stick at column 1 |
| Shift+Up from the end of the file across the empty-looking short line | Selection extends; the anchor does not jump |
| Shift+Left on `т` and on `👍` | One grapheme is selected, not a single byte, and `👍` occupies two cells |
| Ctrl+C, Ctrl+V, Ctrl+Z | Copy, paste, and undo change the buffer. The status does not claim the system clipboard changed |
| Bracketed paste of two lines | One undo removes the whole paste |
| Mouse drag | Selection follows the drag when the terminal sends SGR mouse. If it does not, say so; do not call mouse supported |
| Resize below and above 80×24 | The status and the current line stay visible. Below 80×24 is usable but cramped |
| Resize at 80×24 | tabs, editor, bottom menu, and status do not overlap |
| `axam doctor` | Size matches the window. SSH is true on a remote session. Command keys are described as unavailable |

## Paths

1. Ubuntu local terminal → `axam`.
2. macOS Terminal.app → SSH → Ubuntu → `axam`. Confirm Shift+Up and Shift+Down select. Confirm Command+C does not reach axam.
3. iTerm2 → SSH → Ubuntu → `axam`. Same Shift+Up / Shift+Down check. If CSI u is enabled, try Ctrl+Shift+P and record whether the palette opened.
4. tmux on Ubuntu, with and without `extended-keys`.
5. GNU Screen, expecting missing modified keys and working Ctrl+K chords.

Record the TERM value and the doctor report with any failure.
