// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package keymap

import (
	"testing"

	"github.com/alexshescp/axam/internal/commands"
)

func TestProfilesResolveBaseline(t *testing.T) {
	// Profile | chord | command | environment where a terminal can deliver it.
	// The test locks the resolver. It does not claim every terminal sends the bytes.
	table := []struct {
		profile string
		chord   string
		command string
	}{
		{"axam", "ctrl+n", commands.FileNew},
		{"axam", "ctrl+o", commands.FileOpen},
		{"axam", "ctrl+s", commands.FileSave},
		{"axam", "ctrl+shift+s", commands.FileSaveAs},
		{"axam", "ctrl+k ctrl+s", commands.FileSaveAs},
		{"axam", "ctrl+w", commands.FileClose},
		{"axam", "ctrl+shift+t", commands.FileReopen},
		{"axam", "ctrl+tab", commands.NavNextTab},
		{"axam", "ctrl+shift+tab", commands.NavPrevTab},
		{"axam", "ctrl+z", commands.EditUndo},
		{"axam", "ctrl+y", commands.EditRedo},
		{"axam", "ctrl+x", commands.EditCut},
		{"axam", "ctrl+c", commands.EditCopy},
		{"axam", "ctrl+v", commands.EditPaste},
		{"axam", "ctrl+a", commands.EditSelectAll},
		{"axam", "ctrl+d", commands.EditNextOccurrence},
		{"axam", "ctrl+shift+k", commands.EditDeleteLine},
		{"axam", "ctrl+/", commands.EditComment},
		{"axam", "ctrl+_", commands.EditComment},
		{"axam", "tab", commands.EditIndent},
		{"axam", "shift+tab", commands.EditOutdent},
		{"axam", "ctrl+f", commands.FindOpen},
		{"axam", "ctrl+h", commands.FindReplace},
		{"axam", "ctrl+g", commands.FindGoto},
		{"axam", "ctrl+p", commands.NavQuickOpen},
		{"axam", "ctrl+shift+p", commands.ViewPalette},
		{"axam", "ctrl+k ctrl+p", commands.ViewPalette},
		{"axam", "ctrl+home", commands.MoveDocStart},
		{"axam", "ctrl+end", commands.MoveDocEnd},
		{"axam", "shift+up", commands.MoveSelectUp},
		{"axam", "shift+down", commands.MoveSelectDown},
		{"axam", "shift+left", commands.MoveSelectLeft},
		{"axam", "shift+right", commands.MoveSelectRight},
		{"axam", "ctrl+shift+left", commands.MoveSelectWordLeft},
		{"axam", "ctrl+shift+right", commands.MoveSelectWordRight},
		{"axam", "ctrl+shift+home", commands.MoveSelectDocStart},
		{"axam", "ctrl+shift+end", commands.MoveSelectDocEnd},
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
	m, _, err := Load("axam")
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
	m, _, err := Load("axam")
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

func TestLegacyTermStudioProfileAlias(t *testing.T) {
	m, info, err := Load("termstudio")
	if err != nil {
		t.Fatal(err)
	}
	if info.ID != "axam" || info.Title != "Axam Studio" {
		t.Fatalf("alias resolved to %+v", info)
	}
	cmd, _ := m.Resolve(MustChord("ctrl+d"))
	if cmd != commands.EditNextOccurrence {
		t.Fatalf("alias ctrl+d -> %q", cmd)
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
