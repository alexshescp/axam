// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package input turns tcell events into keymap keys or plain text.
// Printable characters are not shortcuts. Ctrl+Alt+rune is AltGr unless bound.
package input

import (
	"strings"
	"unicode"

	"github.com/codersclub/axam/internal/keymap"
	"github.com/gdamore/tcell/v2"
)

// Result is the normalized meaning of one key event.
type Result struct {
	Key    keymap.Key
	Insert string
	// Command is true when the event should be offered to the keymap.
	// Insert may still be set for AltGr / unbound Alt+rune fallback.
	Command bool
}

// Normalize classifies a key event.
func Normalize(ev *tcell.EventKey) Result {
	if ev == nil {
		return Result{}
	}
	mod := mods(ev.Modifiers())
	switch ev.Key() {
	case tcell.KeyRune:
		r := ev.Rune()
		if r == 0 {
			return Result{}
		}
		if mod&(keymap.Ctrl|keymap.Alt|keymap.Meta) == 0 {
			return Result{Insert: string(r)}
		}
		name := string(unicode.ToLower(r))
		if r == ' ' {
			name = "space"
		}
		k := keymap.Key{Name: name, Mod: mod}
		// Uppercase with Ctrl/Alt and no explicit shift still means shift
		// when the terminal reported the shifted rune.
		if unicode.IsLetter(r) && unicode.IsUpper(r) && mod&keymap.Shift == 0 && mod&keymap.Ctrl != 0 {
			k.Mod |= keymap.Shift
		}
		return Result{Key: k, Insert: string(r), Command: true}
	case tcell.KeyBacktab:
		mod |= keymap.Shift
		mod &^= keymap.Shift
		mod |= keymap.Shift
		return Result{Key: keymap.Key{Name: "tab", Mod: mod}, Command: true}
	case tcell.KeyTab:
		if ev.Modifiers()&tcell.ModShift != 0 {
			mod |= keymap.Shift
		}
		return Result{Key: keymap.Key{Name: "tab", Mod: mod}, Command: true}
	case tcell.KeyEnter:
		return Result{Key: keymap.Key{Name: "enter", Mod: mod}, Command: true}
	case tcell.KeyBackspace:
		return Result{Key: keymap.Key{Name: "backspace", Mod: mod}, Command: true}
	case tcell.KeyDelete:
		return Result{Key: keymap.Key{Name: "delete", Mod: mod}, Command: true}
	case tcell.KeyEsc:
		return Result{Key: keymap.Key{Name: "escape", Mod: mod}, Command: true}
	case tcell.KeyUp:
		return named("up", mod)
	case tcell.KeyDown:
		return named("down", mod)
	case tcell.KeyLeft:
		return named("left", mod)
	case tcell.KeyRight:
		return named("right", mod)
	case tcell.KeyHome:
		return named("home", mod)
	case tcell.KeyEnd:
		return named("end", mod)
	case tcell.KeyPgUp:
		return named("pgup", mod)
	case tcell.KeyPgDn:
		return named("pgdn", mod)
	case tcell.KeyInsert:
		return named("insert", mod)
	case tcell.KeyCtrlSpace:
		return named("space", mod|keymap.Ctrl)
	}

	if ev.Key() >= tcell.KeyF1 && ev.Key() <= tcell.KeyF64 {
		n := int(ev.Key()-tcell.KeyF1) + 1
		return named("f"+itoa(n), mod)
	}
	if ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ {
		ch := ev.Rune()
		if ch == 0 {
			ch = 'a' + rune(ev.Key()-tcell.KeyCtrlA)
		}
		kmod := keymap.Ctrl | (mod &^ keymap.Ctrl)
		if unicode.IsUpper(ch) {
			kmod |= keymap.Shift
		}
		return Result{Key: keymap.Key{Name: string(unicode.ToLower(ch)), Mod: kmod}, Command: true}
	}
	switch ev.Key() {
	case tcell.KeyCtrlLeftSq:
		return named("[", mod|keymap.Ctrl)
	case tcell.KeyCtrlRightSq:
		return named("]", mod|keymap.Ctrl)
	case tcell.KeyCtrlBackslash:
		return named("\\", mod|keymap.Ctrl)
	case tcell.KeyCtrlUnderscore:
		return named("_", mod|keymap.Ctrl)
	case tcell.KeyCtrlCarat:
		return named("^", mod|keymap.Ctrl)
	}
	return Result{}
}

func named(name string, mod keymap.Mod) Result {
	return Result{Key: keymap.Key{Name: name, Mod: mod}, Command: true}
}

func mods(m tcell.ModMask) keymap.Mod {
	var out keymap.Mod
	if m&tcell.ModShift != 0 {
		out |= keymap.Shift
	}
	if m&tcell.ModCtrl != 0 {
		out |= keymap.Ctrl
	}
	if m&tcell.ModAlt != 0 {
		out |= keymap.Alt
	}
	if m&tcell.ModMeta != 0 {
		out |= keymap.Meta
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// PasteText converts a key delivered inside a bracketed paste into text.
func PasteText(ev *tcell.EventKey) string {
	if ev == nil {
		return ""
	}
	switch ev.Key() {
	case tcell.KeyRune:
		return string(ev.Rune())
	case tcell.KeyEnter:
		return "\n"
	case tcell.KeyTab:
		return "\t"
	default:
		r := Normalize(ev)
		if r.Insert != "" && !r.Command {
			return r.Insert
		}
		return ""
	}
}

// FormatChord renders a chord in progress.
func FormatChord(keys []keymap.Key) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k.Display()
	}
	return strings.Join(parts, " ")
}
