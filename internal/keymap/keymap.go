// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package keymap resolves chords to command ids.
// Profiles and user config are data. They do not implement editing.
package keymap

import (
	"fmt"
	"strings"
)

// Mod is a set of modifier keys.
type Mod int

const (
	Shift Mod = 1 << iota
	Ctrl
	Alt
	Meta
)

// Key is one normalized keypress. Name is lowercase ("left", "s", "/").
type Key struct {
	Name string
	Mod  Mod
}

func (k Key) String() string {
	var b strings.Builder
	writeMod := func(on bool, name string) {
		if !on {
			return
		}
		if b.Len() > 0 {
			b.WriteByte('+')
		}
		b.WriteString(name)
	}
	writeMod(k.Mod&Ctrl != 0, "ctrl")
	writeMod(k.Mod&Alt != 0, "alt")
	writeMod(k.Mod&Shift != 0, "shift")
	writeMod(k.Mod&Meta != 0, "meta")
	if b.Len() > 0 {
		b.WriteByte('+')
	}
	b.WriteString(k.Name)
	return b.String()
}

// Display is the form shown in the palette.
func (k Key) Display() string {
	var parts []string
	if k.Mod&Ctrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if k.Mod&Alt != 0 {
		parts = append(parts, "Alt")
	}
	if k.Mod&Shift != 0 {
		parts = append(parts, "Shift")
	}
	if k.Mod&Meta != 0 {
		parts = append(parts, "Meta")
	}
	parts = append(parts, prettyName(k.Name))
	return strings.Join(parts, "+")
}

func prettyName(name string) string {
	switch name {
	case "left", "right", "up", "down", "home", "end", "tab", "enter", "escape", "space", "backspace", "delete":
		return strings.ToUpper(name[:1]) + name[1:]
	case "pgup":
		return "PgUp"
	case "pgdn":
		return "PgDn"
	default:
		if len(name) > 1 && name[0] == 'f' {
			return strings.ToUpper(name)
		}
		return name
	}
}

// Binding maps a chord to a command id.
type Binding struct {
	Chord   []Key
	Command string
}

func (b Binding) String() string {
	parts := make([]string, len(b.Chord))
	for i, k := range b.Chord {
		parts[i] = k.String()
	}
	return strings.Join(parts, " ")
}

func (b Binding) Display() string {
	parts := make([]string, len(b.Chord))
	for i, k := range b.Chord {
		parts[i] = k.Display()
	}
	return strings.Join(parts, " ")
}

type node struct {
	cmd  string
	kids map[Key]*node
}

// Map is a chord trie.
type Map struct {
	name string
	root node
	list []Binding
}

// New returns an empty map.
func New(name string) *Map { return &Map{name: name} }

// Name is the profile id.
func (m *Map) Name() string { return m.name }

// Bind adds a chord. A second command for the same chord is an error.
func (m *Map) Bind(chord []Key, command string) error {
	return m.insert(chord, command, false)
}

// Set replaces any command already bound to the chord.
// An empty command removes the binding.
func (m *Map) Set(chord []Key, command string) {
	_ = m.insert(chord, command, true)
}

func (m *Map) insert(chord []Key, command string, override bool) error {
	if len(chord) == 0 {
		return fmt.Errorf("empty chord")
	}
	if command == "" && !override {
		return fmt.Errorf("empty command")
	}
	n := &m.root
	for _, k := range chord {
		if n.kids == nil {
			n.kids = map[Key]*node{}
		}
		if n.kids[k] == nil {
			n.kids[k] = &node{}
		}
		n = n.kids[k]
	}
	if n.cmd != "" && n.cmd != command && !override {
		return fmt.Errorf("duplicate chord %s (%s vs %s)", formatChord(chord), n.cmd, command)
	}
	n.cmd = command
	filtered := m.list[:0]
	for _, b := range m.list {
		if !sameChord(b.Chord, chord) {
			filtered = append(filtered, b)
		}
	}
	m.list = filtered
	if command != "" {
		m.list = append(m.list, Binding{Chord: append([]Key(nil), chord...), Command: command})
	}
	return nil
}

func sameChord(a, b []Key) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func formatChord(chord []Key) string {
	parts := make([]string, len(chord))
	for i, k := range chord {
		parts[i] = k.String()
	}
	return strings.Join(parts, " ")
}

// Resolve walks seq. prefix is true when longer chords still exist.
// cmd may be set on a prefix. An unknown sequence returns "", false.
func (m *Map) Resolve(seq []Key) (cmd string, prefix bool) {
	n := &m.root
	for _, k := range seq {
		if n.kids == nil || n.kids[k] == nil {
			return "", false
		}
		n = n.kids[k]
	}
	return n.cmd, len(n.kids) > 0
}

// Bindings returns a copy of the bound chords.
func (m *Map) Bindings() []Binding {
	out := make([]Binding, len(m.list))
	copy(out, m.list)
	return out
}

// ChordsFor returns display strings for a command, shortest first.
func (m *Map) ChordsFor(command string) []string {
	var out []string
	best := 99
	for _, b := range m.list {
		if b.Command != command {
			continue
		}
		if len(b.Chord) < best {
			best = len(b.Chord)
			out = []string{b.Display()}
			continue
		}
		if len(b.Chord) == best {
			out = append(out, b.Display())
		}
	}
	return out
}

// ParseChord parses "ctrl+k ctrl+s".
func ParseChord(s string) ([]Key, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty chord")
	}
	fields := strings.Fields(s)
	out := make([]Key, 0, len(fields))
	for _, f := range fields {
		k, err := ParseKey(f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s, err)
		}
		out = append(out, k)
	}
	return out, nil
}

// ParseKey parses one key such as "ctrl+shift+left".
func ParseKey(s string) (Key, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	parts := strings.Split(s, "+")
	var k Key
	if len(parts) == 1 && parts[0] == "" {
		return Key{}, fmt.Errorf("empty key")
	}
	name := parts[len(parts)-1]
	if name == "backtab" {
		k.Mod |= Shift
		name = "tab"
	}
	for _, p := range parts[:len(parts)-1] {
		switch p {
		case "ctrl", "control":
			k.Mod |= Ctrl
		case "alt", "opt", "option", "meta-alt":
			k.Mod |= Alt
		case "shift":
			k.Mod |= Shift
		case "meta", "cmd", "command", "super", "win":
			k.Mod |= Meta
		default:
			return Key{}, fmt.Errorf("unknown modifier %q", p)
		}
	}
	name = canonicalName(name)
	if name == "" {
		return Key{}, fmt.Errorf("missing key name in %q", s)
	}
	if len([]rune(name)) != 1 && !knownName(name) {
		return Key{}, fmt.Errorf("unknown key %q", name)
	}
	k.Name = name
	return k, nil
}

func canonicalName(name string) string {
	switch name {
	case "return":
		return "enter"
	case "esc":
		return "escape"
	case "bs":
		return "backspace"
	case "del":
		return "delete"
	case "pageup", "pgup":
		return "pgup"
	case "pagedown", "pgdn", "pgdown":
		return "pgdn"
	default:
		return name
	}
}

func knownName(name string) bool {
	switch name {
	case "left", "right", "up", "down", "home", "end", "pgup", "pgdn",
		"tab", "enter", "escape", "space", "backspace", "delete", "insert":
		return true
	}
	if len(name) >= 2 && name[0] == 'f' {
		for _, r := range name[1:] {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// MustChord parses or panics. Profiles are compiled data, so a bad chord is a bug.
func MustChord(s string) []Key {
	c, err := ParseChord(s)
	if err != nil {
		panic(err)
	}
	return c
}
