# axam

TermStudio is a modeless terminal text editor. The command you install and run is **axam**. A package also links that binary as `termstudio`.

It is a Go program with its own text engine and its own screen. It is not a skin on Vim, Nano, or Micro.

Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.  
License: GPL-3.0-or-later. See [LICENSE](LICENSE).

## Contents

- [Install and update](#install-and-update)
- [Start editing](#start-editing)
- [Interface and commands](#interface-and-commands)
- [Keyboard and terminal behavior](#keyboard-and-terminal-behavior)
- [Configuration](#configuration)
- [File safety](#file-safety)
- [Build, test, and package](#build-test-and-package)
- [Architecture and project layout](#architecture-and-project-layout)
- [Current limits](#current-limits)

## Install and update

A local Debian package, then a copy of the binary, are the two install paths. You do not need a Go toolchain for either.

```bash
sudo apt install ./axam_1.1.0_amd64.deb
axam
termstudio
```

`apt` accepts a path to a `.deb`. The package installs `/usr/bin/axam` and `/usr/bin/termstudio` (a symlink). arm64 builds use `axam_1.1.0_arm64.deb`.

Or download the static binary for your OS and put it on `PATH`:

| File | System |
| --- | --- |
| `axam_linux_amd64` | Ubuntu and other Linux, 64-bit Intel/AMD |
| `axam_linux_arm64` | Linux, 64-bit ARM |
| `axam_darwin_amd64` | macOS, Intel |
| `axam_darwin_arm64` | macOS, Apple silicon |

```bash
install -m 0755 axam_linux_amd64 /usr/local/bin/axam
ln -sf /usr/local/bin/axam /usr/local/bin/termstudio
```

Build from source:

```bash
go test ./...
go vet ./...
go build -o axam ./cmd/axam
```

`make dist` and `build.ps1` cross-compile the four binaries and the two `.deb` packages. Version 1.1.0 is the default; `-X main.version=` overrides `axam version`.

Update by installing a newer `.deb` or by replacing the binary. `config.toml`
and `state.json` are user files and are not removed by package upgrades.
Release changes are in [CHANGELOG.md](CHANGELOG.md).

## Start editing

```bash
axam
axam app.go
axam /etc/nginx/nginx.conf
axam ./project/
axam doctor
axam keymap
axam version
```

If a filename does not exist, axam opens a pending document at that absolute
path:

```bash
cd ~/notes
axam notes1.txt
```

Opening and closing the empty document does **not** create `notes1.txt`.
Typing and then saving creates it through the same safe replacement path as
other saves. Pressing Save explicitly on the unchanged pending document
creates an empty file. The parent directory must already exist; axam does not
create directory trees from a possibly mistyped path.

Editing a root-owned file: see [docs/PRIVILEGED.md](docs/PRIVILEGED.md). `sudo axam` runs the whole session as root. `sudoedit` with `EDITOR=axam` is the better habit. Ordinary editing does not need root.

## Interface and commands

The permanent layout is:

1. document tabs at the top;
2. optional explorer and one or two editor panes;
3. the bottom command menu;
4. the status row.

The bottom row menu above the status bar contains **File, Edit, Selection,
Find, View, Goto, and Help**. Press F10, click a category, or use the mouse.
While it is open:

- Left/Right changes category;
- Up/Down changes command;
- Enter runs the highlighted command;
- Escape or F10 closes it;
- typing the first letter selects a category.

The popup opens upward, preserving terminal bounds. Shortcut labels come from
the active profile. Menu items, keybindings, the F1 command palette, and tests
all resolve the same central command IDs.

The status row shows the active file, dirty/read-only state, one-based logical
line and visual terminal column, language, encoding, line endings, cursor
count, and keymap profile. Selection highlighting, current-line highlighting,
bracket matching, syntax coloring, search matches, and trailing whitespace
share the editor viewport.

Tabs, find and replace, quick open, bounded project search, a file explorer,
one optional split, folding, multiple cursors, column selection, and mouse
selection are implemented. A multi-line paste and Replace All are each one
undo transaction.

## Keyboard and terminal behavior

axam is modeless. Shift movement uses an anchored selection: the anchor does
not jump when the selection reverses direction. Vertical movement remembers a
terminal-cell column through short lines. Tabs, Cyrillic, combining
characters, invalid UTF-8 bytes, emoji, bytes, and terminal cells are not
treated as one coordinate system.

Profiles are TermStudio Default, Sublime Text-inspired, VS Code-inspired, and
Notepad++-inspired. They are compatibility maps, not claims of desktop
equivalence. Print one with:

```bash
axam keymap
axam keymap vscode
```

Command+C and the other Command shortcuts are not supported over SSH, and they
are not claimed locally. Ctrl+Shift+letter needs modifyOtherKeys or the Kitty
protocol; otherwise use Ctrl+K chords. For example, Ctrl+K then Ctrl+S saves
as, Ctrl+K then Ctrl+P opens the palette, and Ctrl+K then Ctrl+O opens Quick
Open.

Run `axam doctor` for a report based on this terminal. It reports facts and
advertised capabilities, not unverified success. See
[docs/TERMINAL.md](docs/TERMINAL.md),
[docs/KEYMAPS.md](docs/KEYMAPS.md), and
[docs/SHORTCUTS.md](docs/SHORTCUTS.md).

## Configuration

The config path is `~/.config/axam/config.toml` on Linux unless
`XDG_CONFIG_HOME` is set. macOS uses Application Support. Open Configuration
from the Help menu or palette to create the template. A missing file uses
defaults. Invalid or unknown settings name the file and location and prevent
startup.

```toml
keymap_profile = "termstudio"
theme = "dark"
tab_size = 4
insert_spaces = true
line_numbers = true
word_wrap = false
show_trailing_whitespace = true
explorer = false
clipboard = "auto"
```

`clipboard` is `auto`, `internal`, `osc52`, `xclip`, or `wl-copy`. Only `xclip` and `wl-copy` run a local program. Copy always fills the editor clipboard. OSC 52 is sent when the mode is not `internal`, and that send is not treated as proof the system clipboard changed.

Custom bindings override a selected profile:

```toml
[bindings]
"ctrl+q" = "file.quit"
"ctrl+k ctrl+x" = ""
```

An empty command unbinds the exact chord. Unknown commands and invalid chords
are rejected.

## File safety

Existing files are decoded as UTF-8 or BOM-marked UTF-16 LE/BE. LF, CRLF,
and CR are detected. Mixed endings are normalized to the majority style and
disclosed. Invalid UTF-8 bytes are preserved and rendered safely.

Save compares disk identity and content hash before overwriting an externally
changed file. With backups enabled, old bytes are copied under the axam
configuration directory. A sibling temporary file is written, synchronized,
permissioned, and renamed over the target. Unix rename-over is atomic.
Windows must remove the destination before its fallback rename and cannot make
the same atomicity guarantee.

Symlinks are resolved and the target is replaced so the link remains. File
contents are never executed. Binary files with NUL in the initial sample are
refused. Files over 256 MiB require `--force`; syntax coloring turns off above
2 MiB.

## Build, test, and package

Prerequisite: Go 1.26.3 or a compatible newer toolchain.

```bash
go mod verify
go test ./...
go vet ./...
go test -race ./...
go build -trimpath -o axam ./cmd/axam
```

The race detector needs cgo and a C compiler on the host. Linux CI runs it.
SimulationScreen tests cover an 80x24 frame, resize, bottom-menu keyboard and
mouse interaction, paste, Unicode selection, command routing, and pending-file
lifecycle. Real terminal applications remain a manual matrix.

```bash
make dist
make deb
# Windows PowerShell:
./build.ps1
```

Release targets are Linux amd64/arm64 and macOS amd64/arm64 with
`CGO_ENABLED=0`. The repository also builds `axam.exe` for local validation.
The pure-Go Debian builder writes the three required `ar` members without a
dpkg dependency. CI inspects the result.

## Architecture and project layout

The code is one executable with explicit package boundaries:

```text
cmd/axam/             flags, subcommands, startup
internal/app/         session, documents, commands, TUI, bottom menu
internal/input/       tcell event normalization and sequence harness
internal/keymap/      profiles, chords, conflicts
internal/commands/    stable command IDs and palette catalog
internal/editor/      selections, operations, cursors, undo/redo
internal/buffer/      lines, encodings, EOL, insert/delete
internal/text/        graphemes, wrapping, visual cells
internal/files/       read guards, identity, backup, atomic save
internal/search/      literal and regular-expression search
internal/syntax/      Chroma mapping and bracket matching
internal/project/     bounded traversal and project search
internal/config/      TOML validation and recent-path state
internal/clipboard/   internal, OSC 52, xclip, wl-copy paths
internal/doctor/      honest capability report
internal/ui/          terminal themes
packaging/            Debian builder and simulator demo
```

Read these views:

- [Architecture baseline](docs/ARCHITECTURE.md)
- [C4 system, container, component, and dynamic views](docs/C4.md)
- [TOGAF-aligned enterprise viewpoints](docs/TOGAF.md)
- [Runtime and file-safety block diagrams](docs/FLOWS.md)
- [Libraries and dependency policy](docs/DEPENDENCIES.md)
- [Architecture decision records](docs/adr/README.md)

The major design choices are tcell rather than a widget framework, a custom
line-slice editing engine, anchor/head cursors, Chroma rather than cgo
Tree-sitter bindings, and one command registry across all interaction
surfaces.

## Layout

```
 main.go  ● config.toml
  1  package main
  2
  3  func main() {
> 4      fmt.Println("axam")
  5  }
 File  Edit  Selection  Find  View  Goto  Help              F10 Menu
 main.go ●   4:5          go  UTF-8  LF  termstudio
```

A captured 80×24 frame from the simulator is in [docs/demo.txt](docs/demo.txt). It is the real draw path, not a picture of another editor.

## Current limits

- There is no language server, autocomplete, semantic diagnostics, plugin
  host, or remote service.
- Highlighting is lexical and per line; multiline tokens can be colored
  incorrectly.
- Folding collapses a following indentation block and is not syntax-aware.
- Files are fully loaded in memory.
- Project traversal stops at 20,000 files; project search limits result count
  and skips binary or oversized inputs.
- OSC 52 and SGR mouse can be ignored by a terminal.
- More than two editor panes are not implemented.
- Signed artifacts and automatic update checks are not implemented.

## Checks that are not automated

[docs/MANUAL_TEST.md](docs/MANUAL_TEST.md) is the checklist for Terminal.app, iTerm2, a local Linux terminal, tmux, and GNU Screen. Shift+Up and Shift+Down are called out on purpose. Design notes are in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Contributing

Keep input parsing out of rendering and keep editing commands out of keymap
data. Add commands to `internal/commands`, implement them once in the app,
then expose the ID through the menu, palette, or profile. Editing invariants
belong in pure editor/text tests; terminal layout belongs in SimulationScreen
tests. Run formatting, tests, vet, and the relevant manual terminal checks
before proposing a release change.
