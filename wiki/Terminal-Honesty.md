# Terminal honesty

Terminals are not desktop GUIs. Axam Studio was written to respect that instead of marketing around it.

## What we claim

- Shift+arrow selection when the terminal sends the CSI sequence
- Ctrl+letter chords that almost every tty delivers
- Ctrl+K sequences as a documented fallback
- OSC 52 clipboard **attempts** (always with an internal clipboard update)
- Mouse selection when SGR mouse events arrive

## What we do not claim

- macOS **Command** shortcuts over SSH (or as a dependable local TUI event)
- Ctrl+Shift+letter on stock Terminal.app / legacy SSH without modifyOtherKeys or Kitty CSI-u
- Confirmed system clipboard success just because OSC 52 was sent
- That every profile shortcut exists on every terminal

## Your toolkit

```bash
axam doctor
```

Inside the editor: **Terminal Compatibility Report** and the Keymap buffer. Manual matrix: the repository’s `docs/MANUAL_TEST.md`.

Honesty is a feature. If a key never arrives, Axam Studio will not pretend it did.
