# Terminal compatibility

axam reads the terminal through tcell. tcell 2.13 requests modifyOtherKeys and the Kitty keyboard protocol on xterm-like terminals. Older terminals stay on legacy sequences. The editor does not pretend a missing sequence arrived.

`axam doctor` prints size, color, TERM, SSH, tmux, screen, and the limits below. Inside the editor the same report is the Terminal Compatibility Report command. Mouse support is requested and not probed: a terminal can ignore it.

## What is dependable

- Typing, including Cyrillic, when the locale is UTF-8.
- Arrows, Home, End, Page Up, Page Down.
- Shift+arrow on xterm-style terminals (`CSI 1;2A` for Shift+Up, `CSI 1;2B` for Shift+Down, `CSI 1;2D` / `1;2C` for left and right).
- Ctrl+letter for the baseline shortcuts (save, copy, find, undo).
- Ctrl+K chords when a modified letter cannot be distinguished.
- F10 for the bottom command menu on terminals that deliver function keys.
- Bracketed paste, when the terminal supports it. The pasted text is one undo step.
- 16-color, 256-color, and true-color themes. The theme drops to named colors below 256.

## What is not dependable

- Command+C, Command+X, Command+V, and any other Command or Super chord. A Mac keyboard over SSH does not send those as distinct events to the remote program. They are not bound and not claimed.
- Ctrl+Shift+letter on a legacy terminal. It usually arrives as Ctrl+letter. Use the Ctrl+K chord.
- Ctrl+H as Replace. Many terminals send Backspace as byte 0x08, which is Ctrl+H. Backspace deletes. Open Replace from the palette (F1) when Ctrl+H does not arrive as a modified H.
- Ctrl+/ . Some terminals send Ctrl+_. Both are bound to toggle comment.
- Mouse reporting inside tmux unless tmux mouse mode is on.
- OSC 52 clipboard. Copy always updates the editor clipboard. The terminal may also receive an OSC 52 sequence. axam does not report that as a confirmed system copy. Over SSH, the local clipboard changes only if the local terminal allows OSC 52. Set `clipboard = "xclip"` or `"wl-copy"` only on a local Linux desktop where that tool exists. `auto`, `internal`, and `osc52` do not spawn a process.

## Clients

### macOS Terminal.app to Ubuntu over SSH

Terminal.app can be set to report a Unicode keyboard and to use Option as Meta. That still does not produce Command chords. Shift+arrow works in current Terminal.app when the remote TERM is xterm-256color. If Shift+Up moves the cursor without selecting, the terminal is not sending the shift modifier; use `axam doctor` and the checklist in `docs/MANUAL_TEST.md`.

### iTerm2 to Ubuntu over SSH

Profiles → Keys → Report modifiers using CSI u, if you want Ctrl+Shift+letter. Leave this off if you only need the baseline and the Ctrl+K chords. Do not enable a preset that swallows Ctrl+C before it reaches the remote tty if you want copy to reach axam.

### tmux

```
set -g extended-keys on
set -as terminal-features ',xterm*:extkeys'
set -g mouse on
```

Without extended keys, Shift+arrow may be consumed or downgraded.

### GNU Screen

Screen drops many modified keys. Prefer tmux, or use the Ctrl+K chords.

## Exit

axam restores the terminal on a normal quit. A panic is caught only far enough to leave raw mode: the panic is still reported after the terminal is finalized. SIGINT and SIGTERM also request a clean exit.

## Terminal layout

Tabs remain on the top row. The discoverable File/Edit/Selection/Find/View/
Goto/Help menu is immediately above the status row and opens upward. It is an
axam-rendered terminal control, not the graphical menu bar of Terminal.app,
iTerm2, Windows Terminal, or the Cursor terminal host. Clicking a host menu
does not send an axam menu command.
