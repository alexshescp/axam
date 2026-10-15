# Features

Axam Studio is built as one product: one command registry, one editing engine, one screen. Keybindings, the F1 palette, the F10 menu, and the tests all call the same command IDs.

## Editing that feels normal

- **Modeless** — no insert/command modes to remember mid-edit.
- **Anchored selection** — Shift+arrows select like a desktop editor; the anchor stays put when the selection reverses.
- **Desired column** — vertical movement remembers the terminal-cell column through short lines.
- **Multiple cursors**, column selection, line duplicate/move/delete, comment toggle.
- **Unicode geometry** — bytes, graphemes, logical lines, wrapped rows, and terminal cells are separate coordinate systems. Cyrillic, combining marks, emoji, and invalid UTF-8 are handled without pretending they are ASCII.
- **Undo/redo** — multi-line paste and Replace All are each one transaction.

## Interface

Permanent layout:

1. Document **tabs** at the top
2. Optional **explorer** and one or two editor panes
3. Interactive **bottom command menu**
4. **Status** row

The bottom menu (File / Edit / Selection / Find / View / Goto / Help) opens with **F10**, arrows, letter shortcuts, Enter, Escape, and mouse. The popup grows upward so it stays inside the terminal.

Status shows file, dirty/read-only, line and column, language, encoding, EOL, cursor count, and keymap profile.

## Find, project, and navigation

- Find / replace, go to line
- Quick open
- Bounded project search (skips binary and oversized inputs; traversal capped)
- Optional split pane
- Folding of following indentation blocks (not syntax-aware)
- Bracket matching and lexical syntax highlighting (Chroma)

## Keymaps without mythology

Profiles:

- **axam** — Axam Studio default
- **sublime** — Sublime Text-inspired
- **vscode** — VS Code-inspired
- **notepadpp** — Notepad++-inspired

```bash
axam keymap
axam keymap vscode
```

`termstudio` still loads as an alias of `axam` for older configs.

Ctrl+K chords are the reliable fallback when Ctrl+Shift+letter never arrives. Command/Super shortcuts are not claimed over SSH.

## Files you can trust

- Pending paths: `axam new-name.txt` remembers the path; **Save** creates the file; close unchanged → **no file**.
- Parent directory must already exist.
- UTF-8 and BOM-marked UTF-16; LF / CRLF / CR detection.
- Atomic replacement on Unix; backup option; content-hash check before overwriting an externally changed file.
- Privileged editing: prefer `sudoedit` with `EDITOR=axam` over a long-running `sudo axam`.

## Packaging

- Static `CGO_ENABLED=0` binaries for Linux/macOS amd64/arm64
- Local-install `.deb` packages with man page and copyright
- GitHub Releases attach assets automatically on `v*` tags
- Config and recent-file state survive package upgrades

## Deliberate non-features

No language server, autocomplete, semantic diagnostics, plugin host, or remote editing service. Highlighting is lexical and per line. Files load fully into memory (large files need `--force`). That is product focus, not a missing checkbox we forgot to tick.
