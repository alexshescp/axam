// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package input

import (
	"testing"

	"github.com/alexshescp/axam/internal/keymap"
	"github.com/gdamore/tcell/v2"
)

func TestNormalizeDoesNotTreatLettersAsShortcuts(t *testing.T) {
	ev := tcell.NewEventKey(tcell.KeyRune, 'S', tcell.ModNone)
	r := Normalize(ev)
	if r.Command || r.Insert != "S" {
		t.Fatalf("%+v", r)
	}
	ev = tcell.NewEventKey(tcell.KeyRune, 'ы', tcell.ModNone)
	r = Normalize(ev)
	if r.Command || r.Insert != "ы" {
		t.Fatalf("%+v", r)
	}
}

func TestNormalizeModifiers(t *testing.T) {
	cases := []struct {
		key  tcell.Key
		ch   rune
		mod  tcell.ModMask
		want string
	}{
		{tcell.KeyLeft, 0, tcell.ModShift, "shift+left"},
		{tcell.KeyRight, 0, tcell.ModShift, "shift+right"},
		{tcell.KeyUp, 0, tcell.ModShift, "shift+up"},
		{tcell.KeyDown, 0, tcell.ModShift, "shift+down"},
		{tcell.KeyLeft, 0, tcell.ModCtrl | tcell.ModShift, "ctrl+shift+left"},
		{tcell.KeyHome, 0, tcell.ModCtrl | tcell.ModShift, "ctrl+shift+home"},
		{tcell.KeyEnd, 0, tcell.ModCtrl, "ctrl+end"},
		{tcell.KeyBacktab, 0, tcell.ModNone, "shift+tab"},
		{tcell.KeyCtrlS, 's', tcell.ModCtrl, "ctrl+s"},
		{tcell.KeyCtrlS, 'S', tcell.ModCtrl | tcell.ModShift, "ctrl+shift+s"},
	}
	for _, tc := range cases {
		r := Normalize(tcell.NewEventKey(tc.key, tc.ch, tc.mod))
		if !r.Command || r.Key.String() != tc.want {
			t.Fatalf("%s got %+v", tc.want, r.Key)
		}
	}
}

func TestAltGrStaysAvailable(t *testing.T) {
	ev := tcell.NewEventKey(tcell.KeyRune, '€', tcell.ModCtrl|tcell.ModAlt)
	r := Normalize(ev)
	if !r.Command || r.Insert != "€" {
		t.Fatalf("%+v", r)
	}
}

func TestSequenceHarness(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"\x1b[A", "up"},
		{"\x1b[1;2A", "shift+up"},
		{"\x1b[1;2B", "shift+down"},
		{"\x1b[1;2D", "shift+left"},
		{"\x1b[1;2C", "shift+right"},
		{"\x1b[1;5D", "ctrl+left"},
		{"\x1b[1;6D", "ctrl+shift+left"},
		{"\x1b[1;6C", "ctrl+shift+right"},
		{"\x1b[1;2H", "shift+home"},
		{"\x1b[1;2F", "shift+end"},
		{"\x1b[1;5H", "ctrl+home"},
		{"\x1b[1;5F", "ctrl+end"},
		{"\x1b[1;6H", "ctrl+shift+home"},
		{"\x1b[1;6F", "ctrl+shift+end"},
		{"\x1b[Z", "shift+tab"},
		{"\x1b[3;5~", "ctrl+delete"},
		{"\x1bOA", "up"},
		{"\x13", "ctrl+s"},
		{"\x1b[115;5u", "ctrl+s"},
		{"\x1b[115;6u", "ctrl+shift+s"},
		{"\x1b[27;6;115~", "ctrl+shift+s"},
		{"\x1b[57352;2u", "shift+up"},
		{"Привет", "П"},
	}
	for _, tc := range cases {
		got := DecodeSequences([]byte(tc.in))
		if len(got) == 0 {
			t.Fatalf("%q decoded nothing", tc.in)
		}
		if tc.want == "П" {
			if got[0].Name != "П" {
				t.Fatalf("cyrillic %+v", got[0])
			}
			continue
		}
		if got[0].String() != tc.want {
			t.Fatalf("%q -> %s, want %s", tc.in, got[0].String(), tc.want)
		}
	}
}

func TestUnrecognizedSequence(t *testing.T) {
	got := DecodeSequences([]byte("\x1b[99;9~"))
	if len(got) != 0 {
		t.Fatalf("expected drop, got %+v", got)
	}
	_ = keymap.Ctrl
}
