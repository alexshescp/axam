# Midnight Commander

Axam Studio is a strong match for **mc**: you stay in the file manager, press F4, and get a modeless editor instead of mcedit.

## One-shot setup

Quit a running `mc` first. On exit it can rewrite `~/.config/mc/ini` and undo the change.

```bash
rc="$HOME/.bashrc"
case "$SHELL" in */zsh) rc="$HOME/.zshrc" ;; esac
touch "$rc"
grep -qx 'export EDITOR=axam' "$rc" || printf '\nexport EDITOR=axam\n' >> "$rc"
export EDITOR=axam

ini="${XDG_CONFIG_HOME:-$HOME/.config}/mc/ini"
mkdir -p "$(dirname "$ini")"
touch "$ini"
awk '
  BEGIN { found = 0; section = 0; inserted = 0 }
  /^\[Midnight-Commander\][[:space:]]*$/ { section = 1; print; next }
  /^\[/ {
    if (section && !found) { print "use_internal_edit=0"; inserted = 1 }
    section = 0
    print
    next
  }
  section && /^use_internal_edit=/ { print "use_internal_edit=0"; found = 1; next }
  { print }
  END {
    if (!found && !inserted) {
      if (section) print "use_internal_edit=0"
      else printf "\n[Midnight-Commander]\nuse_internal_edit=0\n"
    }
  }
' "$ini" > "$ini.tmp" && mv "$ini.tmp" "$ini"
```

Then start `mc` and press **F4**.

## Manual UI path

1. Install `axam` on `PATH`
2. `export EDITOR=axam` (put it in your shell profile)
3. In mc: **Options → Configuration** → uncheck **Use internal edit**
4. Save setup
5. F4 opens Axam Studio

Older Midnight Commander may use `~/.mc/ini` instead of `~/.config/mc/ini`. Put `use_internal_edit=0` in the file your installation actually reads.

## Why this pairing works

- Same terminal, same session, no GUI hop
- Pending files and safe saves stay available from mc workflows
- F10 menu and F1 palette remain available inside the editor
