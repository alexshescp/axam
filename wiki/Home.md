# Axam Studio

**A real text editor for the terminal — Linux, macOS, and PowerShell sessions.**

We got tired of hunting for something that felt like a normal editor once you were already inside a shell: over SSH, inside tmux, on a headless box, or in a PowerShell window that still has to edit configs, scripts, and notes. Modal editors are powerful. Tiny editors are honest. Neither felt like *opening a file and working*.

So Coders Club (USA), Codepolder (Netherlands), and Aleksandr Shaman built **Axam Studio**.

The command you install is **`axam`**. The product is a modeless, Unicode-aware terminal editor with its own text engine and its own screen — not a skin on Vim, Nano, or Micro.

- **Source:** [github.com/alexshescp/axam](https://github.com/alexshescp/axam)
- **License:** GPL-3.0-or-later
- **Support:** [axam@codcl.com](mailto:axam@codcl.com)
- **Latest release:** [v1.1.0](https://github.com/alexshescp/axam/releases/tag/v1.1.0)

---

## Why we built it

We wanted an editor that:

- starts with `axam file.txt` and behaves like an editor, not a puzzle;
- works the same on Ubuntu servers, Mac terminals, and remote shells;
- does not invent key presses the terminal never delivered;
- saves without gambling with the original file;
- can sit next to Midnight Commander as the F4 editor.

Axam Studio is that tool for our own teams — and it is open source so you can use it the same way.

---

## What you get in 30 seconds

Homebrew (macOS or Linux):

```bash
brew install alexshescp/axam/axam
axam
```

Debian / Ubuntu (amd64):

```bash
curl -fL -o axam_1.1.0_amd64.deb \
  https://github.com/alexshescp/axam/releases/download/v1.1.0/axam_1.1.0_amd64.deb
sudo apt install ./axam_1.1.0_amd64.deb
axam
```

Open a missing name safely:

```bash
axam notes1.txt   # pending file — nothing on disk until Save
```

Press **F10** for a Sublime-style bottom menu: File, Edit, Selection, Find, View, Goto, Help.

---

## Highlights

| | |
| --- | --- |
| **Modeless editing** | Type, select, save. An anchored selection that does not jump when you reverse direction. |
| **Bottom command menu** | F10 opens a real menu above the status line — keyboard and mouse. |
| **Honest terminal limits** | `axam doctor` reports what this session can actually see. |
| **Keymap profiles** | Axam Studio, Sublime-inspired, VS Code-inspired, Notepad++-inspired. |
| **Safe saves** | Atomic replace on Unix; backup option; external-change detection. |
| **Pending files** | Close an untouched new path — no empty file left behind. |
| **Static binaries** | Linux and macOS, amd64 and arm64. Homebrew tap or direct download — no Go toolchain required. |
| **Midnight Commander** | Set `EDITOR=axam` and turn off the internal editor — F4 opens Axam. |

More detail: [Features](Features) · [Install](Install) · [Midnight Commander](Midnight-Commander) · [Terminal honesty](Terminal-Honesty)

---

## Who it is for

- Developers and admins who live in SSH and still want a desktop-shaped editing habit.
- Teams that refuse to pretend Command+C works over a remote tty.
- Anyone who wants one binary, one config, and a clear license.

## Who it is not for

- People looking for an IDE, LSP, or plugin marketplace.
- Workflows that need macOS Command chords as first-class input over SSH.

We are deliberate about that scope. A smaller, correct editor beats a large, wishful one.

---

## Start here

1. [Install](Install) for your OS
2. Run `axam doctor`
3. Press **F1** for the command palette, **F10** for the menu
4. Read [Features](Features) when you want the full picture

**Welcome to Axam Studio.** Edit like you mean it — even when the only window you have is a terminal.
