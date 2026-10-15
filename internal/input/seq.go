// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package input

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alexshescp/axam/internal/keymap"
)

// DecodeSequences parses a byte stream the way a modern xterm-like terminal
// emits keys: legacy CSI, SS3, modifyOtherKeys, and Kitty CSI-u.
//
// The running editor does not use this parser. tcell decodes the tty.
// This function is the reproducible harness for the sequences Axam Studio
// depends on, and its results are the keymap keys Normalize must produce
// for the matching tcell events.
func DecodeSequences(in []byte) []keymap.Key {
	var out []keymap.Key
	i := 0
	for i < len(in) {
		if in[i] == 0x1b {
			k, n, ok := decodeEscape(in[i:])
			if n > 0 {
				if ok {
					out = append(out, k)
				}
				i += n
				continue
			}
		}
		if in[i] < 0x20 {
			switch in[i] {
			case 0x09:
				out = append(out, keymap.Key{Name: "tab"})
			case 0x0d:
				out = append(out, keymap.Key{Name: "enter"})
			case 0x7f, 0x08:
				out = append(out, keymap.Key{Name: "backspace"})
			case 0x1b:
				out = append(out, keymap.Key{Name: "escape"})
			default:
				if in[i] >= 1 && in[i] <= 26 {
					out = append(out, keymap.Key{Name: string(rune('a' + in[i] - 1)), Mod: keymap.Ctrl})
				}
			}
			i++
			continue
		}
		r, size := utf8.DecodeRune(in[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		out = append(out, keymap.Key{Name: string(r)})
		i += size
	}
	return out
}

func decodeEscape(in []byte) (keymap.Key, int, bool) {
	if len(in) < 2 {
		return keymap.Key{}, 0, false
	}
	if in[1] == 'O' && len(in) >= 3 {
		if k, ok := ss3(in[2]); ok {
			return k, 3, true
		}
	}
	if in[1] != '[' {
		if len(in) >= 2 && in[1] >= 0x20 && in[1] < 0x7f {
			r := rune(in[1])
			return keymap.Key{Name: string(unicodeLower(r)), Mod: keymap.Alt}, 2, true
		}
		return keymap.Key{}, 0, false
	}
	// CSI ... final
	i := 2
	for i < len(in) {
		if in[i] >= 0x40 && in[i] <= 0x7e {
			body := string(in[2:i])
			final := rune(in[i])
			k, ok := decodeCSI(body, final)
			return k, i + 1, ok
		}
		i++
		if i > len(in) || i-2 > 64 {
			break
		}
	}
	return keymap.Key{}, 0, false
}

func decodeCSI(body string, final rune) (keymap.Key, bool) {
	if final == 'u' {
		return decodeKitty(body)
	}
	nums := splitNums(body)
	if final == 'Z' {
		return keymap.Key{Name: "tab", Mod: keymap.Shift}, true
	}
	if final == '~' && len(nums) >= 1 && nums[0] == 27 && len(nums) >= 3 {
		mod := xtermMod(nums[1])
		code := nums[2]
		if code >= 1 && code < 32 {
			return keymap.Key{}, false
		}
		name := string(unicodeLower(rune(code)))
		if code == ' ' {
			name = "space"
		}
		if rune(code) >= 'A' && rune(code) <= 'Z' {
			mod |= keymap.Shift
			name = string(unicodeLower(rune(code)))
		}
		return keymap.Key{Name: name, Mod: mod}, true
	}
	if name, ok := arrowFinal(final); ok {
		mod := keymap.Mod(0)
		if len(nums) >= 2 {
			mod = xtermMod(nums[1])
		}
		return keymap.Key{Name: name, Mod: mod}, true
	}
	if final == '~' && len(nums) >= 1 {
		name := ""
		switch nums[0] {
		case 1, 7:
			name = "home"
		case 4, 8:
			name = "end"
		case 3:
			name = "delete"
		case 5:
			name = "pgup"
		case 6:
			name = "pgdn"
		case 2:
			name = "insert"
		default:
			return keymap.Key{}, false
		}
		mod := keymap.Mod(0)
		if len(nums) >= 2 {
			mod = xtermMod(nums[1])
		}
		return keymap.Key{Name: name, Mod: mod}, true
	}
	return keymap.Key{}, false
}

func decodeKitty(body string) (keymap.Key, bool) {
	nums := splitNums(body)
	if len(nums) == 0 {
		return keymap.Key{}, false
	}
	mod := keymap.Mod(0)
	if len(nums) >= 2 {
		mod = xtermMod(nums[1])
	}
	code := nums[0]
	if name, ok := kittySpecial(code); ok {
		return keymap.Key{Name: name, Mod: mod}, true
	}
	if code < 32 || code == 127 {
		return keymap.Key{}, false
	}
	r := rune(code)
	name := string(unicodeLower(r))
	if r == ' ' {
		name = "space"
	}
	if r >= 'A' && r <= 'Z' {
		mod |= keymap.Shift
	}
	return keymap.Key{Name: name, Mod: mod}, true
}

func kittySpecial(code int) (string, bool) {
	switch code {
	case 57350:
		return "left", true
	case 57351:
		return "right", true
	case 57352:
		return "up", true
	case 57353:
		return "down", true
	case 57354:
		return "pgup", true
	case 57355:
		return "pgdn", true
	case 57356:
		return "home", true
	case 57357:
		return "end", true
	case 57348:
		return "insert", true
	case 57349:
		return "delete", true
	case 13:
		return "enter", true
	case 9:
		return "tab", true
	case 27:
		return "escape", true
	case 127:
		return "backspace", true
	default:
		if code >= 57376 && code <= 57387 {
			return "f" + strconv.Itoa(code-57376+13), true
		}
		return "", false
	}
}

func arrowFinal(final rune) (string, bool) {
	switch final {
	case 'A':
		return "up", true
	case 'B':
		return "down", true
	case 'C':
		return "right", true
	case 'D':
		return "left", true
	case 'H':
		return "home", true
	case 'F':
		return "end", true
	default:
		return "", false
	}
}

func ss3(b byte) (keymap.Key, bool) {
	name, ok := arrowFinal(rune(b))
	if !ok {
		return keymap.Key{}, false
	}
	return keymap.Key{Name: name}, true
}

func xtermMod(n int) keymap.Mod {
	if n <= 1 {
		return 0
	}
	n--
	var m keymap.Mod
	if n&1 != 0 {
		m |= keymap.Shift
	}
	if n&2 != 0 {
		m |= keymap.Alt
	}
	if n&4 != 0 {
		m |= keymap.Ctrl
	}
	if n&8 != 0 {
		m |= keymap.Meta
	}
	return m
}

func splitNums(body string) []int {
	if body == "" {
		return nil
	}
	parts := strings.Split(body, ";")
	var out []int
	for _, p := range parts {
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		out = append(out, n)
	}
	return out
}

func unicodeLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}
