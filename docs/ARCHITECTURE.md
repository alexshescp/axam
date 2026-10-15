# Axam Studio architecture

Axam Studio is the editor. The installable command is `axam`. The Go module
path is `github.com/alexshescp/axam`.

This page is the concise technical baseline. Related viewpoints:

- [C4 model](C4.md)
- [runtime and safety flows](FLOWS.md)
- [TOGAF-aligned architecture description](TOGAF.md)
- [dependencies](DEPENDENCIES.md)
- [architecture decision records](adr/README.md)

## Quality attributes

The priorities are, in order:

1. file integrity and terminal restoration;
2. correct Unicode cursor and anchored-selection behavior;
3. predictable, discoverable commands;
4. responsive viewport operations and bounded project work;
5. compact terminal presentation;
6. static, reproducible distribution.

The app is deliberately local-first. Basic editing does not need a network,
daemon, browser, graphical desktop, or external clipboard program.

## Stack

| Piece | Choice | Why |
| --- | --- | --- |
| Language | Go | Single static binary, no GUI runtime, straightforward cross-compilation for Linux and macOS. |
| Terminal | `github.com/gdamore/tcell/v2` v2.13 | Cell-level drawing, a real cursor, mouse, bracketed paste, focus events, and OSC 52 clipboard. v2.13 asks xterm-like terminals for modifyOtherKeys, Kitty CSI-u, and Windows win32-input-mode. |
| Unicode | `github.com/rivo/uniseg` | Grapheme clusters and terminal cell widths. Byte offsets are never treated as columns. |
| Highlighting | `github.com/alecthomas/chroma/v2` | Pure Go lexers for the required languages. Tree-sitter was not used: the maintained Go bindings need cgo, which breaks the single static binary. |
| Config | TOML via `github.com/pelletier/go-toml/v2` | Human-editable, with decode errors that include line and column. |

Bubble Tea was rejected for the core loop. A text editor needs stable key identities, exact selection geometry, and direct control of every cell. An Elm-style view tree would sit between the terminal and those details without making keyboard behavior more correct.

## Text model

Documents are a slice of lines without line-ending bytes. Positions are `{Line, Col}` where `Col` is a byte offset snapped to a grapheme boundary. Visual columns are computed separately and account for tabs, Cyrillic (width 1, often 2 bytes), and emoji (width 2).

A piece table or rope would snapshot large files more cheaply. It would also obscure the line, column, and Unicode invariants this editor has to test. Line slices keep those operations obvious. Inserting a line into a 100k-line file moves a slice of pointers, which is fast enough. Files larger than 256 MiB are refused unless `--force` is set. Syntax highlighting turns off above 2 MiB.

Line endings are detected (LF, CRLF, CR). Mixed endings are normalized to the majority style on save and reported in the status bar. Invalid UTF-8 bytes are preserved as individual width-1 units and written back unchanged. UTF-8 and UTF-16 BOMs are detected and preserved.

## Selection

Each cursor has an anchor and a head. Shift-movement moves the head and leaves the anchor where the selection started. Reversing direction shrinks the range through the anchor; the anchor does not jump. Ranges are half-open.

Without Shift, Left and Right collapse a selection to its start or end. Up and Down move from the head. Vertical movement keeps a desired visual column so short lines do not permanently change the column.

Multiple cursors are the same model repeated. Edits apply from the bottom of the buffer upward, and cursors already processed are shifted by the line and column delta. One keystroke is one undo transaction. Consecutive single-cursor typing without newlines coalesces into one undo step. A paste is one step.

## Keyboard

The path is:

1. tcell decodes terminal bytes (legacy sequences, modifyOtherKeys, Kitty CSI-u).
2. `internal/input` normalizes an event into a `keymap.Key` or a plain insertion.
3. `internal/keymap` resolves a chord, including multi-key sequences, with deterministic conflict detection.
4. `internal/app` runs one command id from a single registry.

Printable characters with no Ctrl or Alt are inserted. Shift is part of the character, not a shortcut. Ctrl+Alt plus a printable rune is treated as AltGr and inserts unless that exact chord is bound. Unbound modified keys do not insert a character.

Command, Super, and macOS Command shortcuts are not claimed. Terminals do not deliver them reliably over SSH.

`internal/input` also contains a sequence harness that maps the xterm and Kitty sequences used by the manual and automated keyboard tests. Runtime decoding stays in tcell; the harness locks the sequences we depend on and the normalized keys they must become.

## Commands and profiles

Profiles are data: Axam Studio (`axam`), Sublime, VS Code, and Notepad++. They are compatibility maps, not clones of those products. The pre-rename id `termstudio` still loads the `axam` profile. User bindings in the config file replace profile bindings after the profile is built. The same command ids serve the keymap, the command palette, and tests.

## UI

The screen is drawn by Axam Studio, not by a widget toolkit. One tab row, an optional file list, the editor, an interactive bottom command menu, and one status row make up the permanent layout. Find, prompts, menu popups, and the command palette are overlays. The default theme is a dark palette with an amber active tab; it is not a copy of a proprietary theme. Colors degrade from true color to 256 and 16.

The bottom menu contains command IDs, never alternate editing callbacks. It
opens upward and uses the same command table as keybindings and the palette.
F10, arrows, Enter, Escape, and mouse clicks are tested on SimulationScreen.
The extra row is included in viewport, overlay, mouse, and resize geometry.

## Files

Saves write a temporary file in the destination directory, sync it, then rename it over the target. On Windows the rename cannot replace an existing file atomically; the previous file is removed immediately before the rename. Permissions are copied when the filesystem allows it. The original is never truncated in place. A content hash detects external changes; overwrite requires confirmation.

Missing command-line filenames are pending documents. Their absolute target is
stored in memory, but opening and unchanged close perform no filesystem write.
The parent must already exist. Explicit Save—including Save on an unchanged
pending buffer—creates the file through the normal atomic-save path. This
distinguishes “look at a named empty buffer” from “create this file.”

## Package boundaries

- `cmd/axam`: process entry point, flags, diagnostic subcommands, and session
  startup.
- `internal/app`: session state, documents and panes, command handlers, event
  loop, terminal layout, overlays, menu, and renderer.
- `internal/input`, `keymap`, `commands`: raw-event normalization, binding
  resolution, profiles, and stable command identity.
- `internal/editor`, `buffer`, `text`: terminal-independent editing domain.
- `internal/files`: disk decoding, identity, backups, and safe replacement.
- `internal/search`, `syntax`, `project`: bounded developer features.
- `internal/config`, `clipboard`, `doctor`, `ui`: host adapters and theme data.

Dependencies point toward the editing domain. The editor does not import
tcell, Chroma, TOML, filesystem services, or the app package.

## State and concurrency

The application event loop owns mutable session and editor state. Project
operations currently run synchronously and accept contexts for bounded future
work. The only routine created by the session waits for SIGINT/SIGTERM and
posts an interrupt event back to the screen; it does not mutate documents.
There are no background renderers and no shared buffer locks.

Persisted state is intentionally small:

- `config.toml` is user-authored and strictly validated;
- `state.json` stores recent file and project paths;
- optional backups store previous target bytes;
- documents remain the user’s files, not an internal database.

## Testing architecture

Pure packages test Unicode geometry, editing, undo, keymaps, search, syntax,
configuration, project traversal, and file safety without a terminal.
`internal/app` tests use tcell SimulationScreen for 80x24 drawing, keyboard,
mouse, paste, menu, resize, and command-routing behavior. The input sequence
harness locks xterm, modifyOtherKeys, and Kitty CSI-u sequences independently
of runtime tcell parsing.

Simulation cannot prove what Terminal.app, iTerm2, tmux, or GNU Screen sends.
Those checks remain in [MANUAL_TEST.md](MANUAL_TEST.md), and support claims
must distinguish automated editor behavior from real terminal delivery.

## Distribution architecture

Release builds use `CGO_ENABLED=0`, `-trimpath`, and a linker-supplied version.
Four native targets are produced: Linux and macOS on amd64 and arm64. A
Windows binary is built for development validation. The Debian package is
constructed directly from `ar` and `tar.gz` members and installs `axam`,
its manual page, and copyright metadata. User configuration is not part of the package,
so upgrades retain it.

## What is intentionally out of scope

Language servers, autocomplete, and semantic diagnostics are not implemented. Folding collapses a following indent block; it is not syntax-aware. Syntax highlighting is lexical and per line, so a multiline string can be colored wrong. Clipboard success means the editor clipboard was updated; OSC 52 is attempted and is not reported as confirmed.
