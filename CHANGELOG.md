# Changelog

All notable changes are documented here. The project uses semantic versioning
for release artifacts.

## 1.1.0 — 2026-10-15

### Added

- Interactive bottom command menu with File, Edit, Selection, Find, View,
  Goto, and Help categories.
- F10, arrow, Enter, Escape, and mouse interaction for the menu.
- Pending-file lifecycle for `axam new-name.txt`: the path is remembered but
  no file is created before explicit Save.
- C4, TOGAF-aligned, runtime-flow, dependency, and ADR documentation.
- Simulation tests for menu registry integrity, keyboard/mouse behavior,
  resizing, and pending-file creation.

### Changed

- The editor viewport reserves one row above the status bar for the menu.
- Missing command-line paths now require an existing parent directory and
  report a clear error otherwise.
- Save As and reopen no longer mark a path as existing before disk state
  confirms it.
- README, terminal guidance, shortcuts, manual matrix, demo, packaging, and
  version output now describe 1.1 behavior.
- The product name is Axam Studio. The Go module path is
  `github.com/alexshescp/axam`.
- The default keymap profile id is `axam`. Configs that still say
  `termstudio` load that same profile.
- The Debian package installs `/usr/bin/axam` without a second command name.
- Support mail is axam@codcl.com.

### Preserved limitations

- Command/Super shortcuts are not available as dependable terminal events.
- Ctrl+Shift+letter requires modifyOtherKeys or Kitty CSI-u.
- OSC 52 is an attempt, not confirmation of a system clipboard change.
- Highlighting is lexical and per line; there is no LSP or autocomplete.
- Files are fully loaded; files above 256 MiB require `--force`.
- Unix replacement is atomic; the Windows replacement fallback is not.
- Real Terminal.app, iTerm2, tmux, and GNU Screen checks remain manual.

## 1.0.0 — 2026-10-15

Initial working release:

- Unicode-aware editor, anchor/head selections, desired vertical columns,
  multiple cursors, undo/redo, and line operations.
- tcell terminal UI with tabs, panes, explorer, status, find/replace, command
  palette, quick open, project search, mouse, and bracketed paste.
- Axam Studio, Sublime, VS Code, and Notepad++ compatibility profiles.
- Chroma lexical highlighting for the documented language set.
- atomic-save preparation, external-change detection, backup option, encoding
  and EOL preservation, binary and large-file guards.
- static Linux/macOS builds and local-install Debian packages.
