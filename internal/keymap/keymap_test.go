// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package keymap

import (
	"testing"

	"github.com/codersclub/axam/internal/commands"
)

func TestProfilesResolveBaseline(t *testing.T) {
	// Profile | chord | command | environment where a terminal can deliver it.
	// The test locks the resolver. It does not claim every terminal sends the bytes.
	table := []struct {
		profile string
		chord   string
		command string
	}{
		{"termstudio", "ctrl+n", commands.FileNew},
		{"termstudio", "ctrl+o", commands.FileOpen},
		{"termstudio", "ctrl+s", commands.FileSave},
		{"termstudio", "ctrl+shift+s", commands.FileSaveAs},
		{"termstudio", "ctrl+k ctrl+s", commands.FileSaveAs},
		{"termstudio", "ctrl+w", commands.FileClose},
		{"termstudio", "ctrl+shift+t", commands.FileReopen},
		{"termstudio", "ctrl+tab", commands.NavNextTab},
		{"termstudio", "ctrl+shift+tab", commands.NavPrevTab},
		{"termstudio", "ctrl+z", commands.EditUndo},
		{"termstudio", "ctrl+y", commands.EditRedo},
		{"termstudio", "ctrl+x", commands.EditCut},
		{"termstudio", "ctrl+c", commands.EditCopy},
		{"termstudio", "ctrl+v", commands.EditPaste},
		{"termstudio", "ctrl+a", commands.EditSelectAll},
		{"termstudio", "ctrl+d", commands.EditNextOccurrence},
		{"termstudio", "ctrl+shift+k", commands.EditDeleteLine},
		{"termstudio", "ctrl+/", commands.EditComment},
		{"termstudio", "ctrl+_", commands.EditComment},
		{"termstudio", "tab", commands.EditIndent},
		{"termstudio", "shift+tab", commands.EditOutdent},
		{"termstudio", "ctrl+f", commands.FindOpen},
		{"termstudio", "ctrl+h", commands.FindReplace},
		{"termstudio", "ctrl+g", commands.FindGoto},
		{"termstudio", "ctrl+p", commands.NavQuickOpen},
		{"termstudio", "ctrl+shift+p", commands.ViewPalette},
		{"termstudio", "ctrl+k ctrl+p", commands.ViewPalette},
		{"termstudio", "ctrl+home", commands.MoveDocStart},
		{"termstudio", "ctrl+end", commands.MoveDocEnd},
		{"termstudio", "shift+up", commands.MoveSelectUp},
		{"termstudio", "shift+down", commands.MoveSelectDown},
		{"termstudio", "shift+left", commands.MoveSelectLeft},
		{"termstudio", "shift+right", commands.MoveSelectRight},
		{"termstudio", "ctrl+shift+left", commands.MoveSelectWordLeft},
		{"termstudio", "ctrl+shift+right", commands.MoveSelectWordRight},
		{"termstudio", "ctrl+shift+home", commands.MoveSelectDocStart},
		{"termstudio", "ctrl+shift+end", commands.MoveSelectDocEnd},
		{"vscode", "alt+up", commands.EditMoveLineUp},
		{"vscode", "ctrl+shift+k", commands.EditDeleteLine},
		{"vscode", "ctrl+d", commands.EditNextOccurrence},
		{"sublime", "ctrl+shift+d", commands.EditDuplicateLine},
		{"sublime", "ctrl+l", commands.EditSelectLine},
		{"sublime", "ctrl+d", commands.EditNextOccurrence},
		{"notepadpp", "ctrl+d", commands.EditDuplicateLine},
		{"notepadpp", "ctrl+l", commands.EditDeleteLine},
		{"notepadpp", "ctrl+q", commands.EditComment},
	}
	cache := map[string]*Map{}
	for _, row := range table {
		m := cache[row.profile]
		if m == nil {
			var err error
			m, _, err = Load(row.profile)
			if err != nil {
				t.Fatal(err)
			}
			cache[row.profile] = m
		}
		chord, err := ParseChord(row.chord)
		if err != nil {
			t.Fatal(err)
		}
		cmd, prefix := m.Resolve(chord)
		if prefix && len(chord) == 1 {
			// A single key that is also a prefix should still expose its command
			// only when it is a complete binding. Our table entries are complete.
		}
		if cmd != row.command {
			t.Fatalf("%s %s -> %q, want %s (prefix %v)", row.profile, row.chord, cmd, row.command, prefix)
		}
		if !commands.Known(cmd) {
			t.Fatalf("unknown command %s", cmd)
		}
	}
}

func TestSequencePrefix(t *testing.T) {
	m, _, err := Load("termstudio")
	if err != nil {
		t.Fatal(err)
	}
	cmd, prefix := m.Resolve(MustChord("ctrl+k"))
	if cmd != "" || !prefix {
		t.Fatalf("ctrl+k cmd %q prefix %v", cmd, prefix)
	}
	cmd, prefix = m.Resolve(MustChord("ctrl+k ctrl+s"))
	if cmd != commands.FileSaveAs || prefix {
		t.Fatalf("save as %q %v", cmd, prefix)
	}
	cmd, prefix = m.Resolve(MustChord("ctrl+k ctrl+z"))
	if cmd != "" || prefix {
		t.Fatalf("unknown sequence recognized as %q", cmd)
	}
}

func TestDuplicateRejected(t *testing.T) {
	m := New("t")
	if err := m.Bind(MustChord("ctrl+s"), commands.FileSave); err != nil {
		t.Fatal(err)
	}
	if err := m.Bind(MustChord("ctrl+s"), commands.FileQuit); err == nil {
		t.Fatal("expected conflict")
	}
	m.Set(MustChord("ctrl+s"), commands.FileSaveAs)
	cmd, _ := m.Resolve(MustChord("ctrl+s"))
	if cmd != commands.FileSaveAs {
		t.Fatalf("override %s", cmd)
	}
}

func TestUserOverride(t *testing.T) {
	m, _, err := Load("termstudio")
	if err != nil {
		t.Fatal(err)
	}
	m.Set(MustChord("ctrl+s"), commands.FileQuit)
	cmd, _ := m.Resolve(MustChord("ctrl+s"))
	if cmd != commands.FileQuit {
		t.Fatal(cmd)
	}
}

func TestNotepadDoesNotStealQuickOpen(t *testing.T) {
	m, info, err := Load("notepadpp")
	if err != nil {
		t.Fatal(err)
	}
	if info.CopyLine {
		t.Fatal("notepad++ copies lines on empty selection in this profile; it should not")
	}
	cmd, prefix := m.Resolve(MustChord("ctrl+p"))
	if cmd != "" || prefix {
		t.Fatalf("ctrl+p should be free in notepad++, got %q prefix %v", cmd, prefix)
	}
	cmd, _ = m.Resolve(MustChord("ctrl+k ctrl+o"))
	if cmd != commands.NavQuickOpen {
		t.Fatal(cmd)
	}
}

func TestEveryBindingKnown(t *testing.T) {
	for _, name := range Profiles() {
		m, _, err := Load(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range m.Bindings() {
			if !commands.Known(b.Command) {
				t.Fatalf("%s binds unknown %s", name, b.Command)
			}
		}
	}
}
