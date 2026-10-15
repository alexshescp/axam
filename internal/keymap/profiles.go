// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package keymap

import (
	"fmt"

	"github.com/alexshescp/axam/internal/commands"
)

// Info describes a compatibility profile. These maps are not claims that a
// terminal can deliver every desktop shortcut.
type Info struct {
	ID       string
	Title    string
	CopyLine bool
	CutLine  bool
}

// Profiles returns the built-in profile ids.
func Profiles() []string {
	return []string{"axam", "sublime", "vscode", "notepadpp"}
}

// Load builds a profile. Unknown names return an error.
// "custom" inherits the Axam Studio profile; user bindings are applied by the
// caller. "termstudio" is the pre-rename id and stays accepted so existing
// config files keep loading.
func Load(name string) (*Map, Info, error) {
	switch name {
	case "", "axam", "termstudio", "custom", "default":
		return build("axam", Info{ID: "axam", Title: "Axam Studio", CopyLine: true, CutLine: true}, axamExtra())
	case "sublime":
		return build("sublime", Info{ID: "sublime", Title: "Sublime Text", CopyLine: true, CutLine: true}, sublimeExtra())
	case "vscode":
		return build("vscode", Info{ID: "vscode", Title: "VS Code", CopyLine: true, CutLine: true}, vscodeExtra())
	case "notepadpp", "notepad++", "notepad":
		return build("notepadpp", Info{ID: "notepadpp", Title: "Notepad++", CopyLine: false, CutLine: false}, notepadExtra())
	default:
		return nil, Info{}, fmt.Errorf("unknown keymap profile %q", name)
	}
}

func build(id string, info Info, extra [][2]string) (*Map, Info, error) {
	m := New(id)
	for _, b := range shared() {
		if err := m.Bind(MustChord(b[0]), b[1]); err != nil {
			return nil, Info{}, fmt.Errorf("%s: %w", id, err)
		}
	}
	for _, b := range extra {
		if err := m.Bind(MustChord(b[0]), b[1]); err != nil {
			return nil, Info{}, fmt.Errorf("%s: %w", id, err)
		}
	}
	return m, info, nil
}

func shared() [][2]string {
	return [][2]string{
		{"left", commands.MoveLeft},
		{"right", commands.MoveRight},
		{"up", commands.MoveUp},
		{"down", commands.MoveDown},
		{"shift+left", commands.MoveSelectLeft},
		{"shift+right", commands.MoveSelectRight},
		{"shift+up", commands.MoveSelectUp},
		{"shift+down", commands.MoveSelectDown},
		{"ctrl+left", commands.MoveWordLeft},
		{"ctrl+right", commands.MoveWordRight},
		{"ctrl+shift+left", commands.MoveSelectWordLeft},
		{"ctrl+shift+right", commands.MoveSelectWordRight},
		{"home", commands.MoveHome},
		{"end", commands.MoveEnd},
		{"shift+home", commands.MoveSelectHome},
		{"shift+end", commands.MoveSelectEnd},
		{"ctrl+home", commands.MoveDocStart},
		{"ctrl+end", commands.MoveDocEnd},
		{"ctrl+shift+home", commands.MoveSelectDocStart},
		{"ctrl+shift+end", commands.MoveSelectDocEnd},
		{"pgup", commands.MovePageUp},
		{"pgdn", commands.MovePageDown},
		{"shift+pgup", commands.MoveSelectPageUp},
		{"shift+pgdn", commands.MoveSelectPageDown},
		{"backspace", commands.EditBackspace},
		{"ctrl+backspace", commands.EditDeleteWordLeft},
		{"delete", commands.EditDelete},
		{"ctrl+delete", commands.EditDeleteWordRight},
		{"enter", commands.EditNewline},
		{"tab", commands.EditIndent},
		{"shift+tab", commands.EditOutdent},
		{"escape", commands.ViewCancel},
		{"ctrl+a", commands.EditSelectAll},
		{"ctrl+z", commands.EditUndo},
		{"ctrl+y", commands.EditRedo},
		{"ctrl+shift+z", commands.EditRedo},
		{"ctrl+c", commands.EditCopy},
		{"ctrl+x", commands.EditCut},
		{"ctrl+v", commands.EditPaste},
		{"ctrl+s", commands.FileSave},
		{"ctrl+o", commands.FileOpen},
		{"ctrl+n", commands.FileNew},
		{"ctrl+w", commands.FileClose},
		{"ctrl+tab", commands.NavNextTab},
		{"ctrl+shift+tab", commands.NavPrevTab},
		{"ctrl+f", commands.FindOpen},
		{"ctrl+h", commands.FindReplace},
		{"ctrl+g", commands.FindGoto},
		{"ctrl+b", commands.NavExplorer},
		{"f1", commands.ViewPalette},
		{"alt+x", commands.ViewPalette},
		{"f3", commands.FindNext},
		{"shift+f3", commands.FindPrev},
		{"ctrl+\\", commands.NavSplit},
		{"alt+up", commands.EditMoveLineUp},
		{"alt+down", commands.EditMoveLineDown},
		// Chords that work when the terminal cannot encode Ctrl+Shift+letter.
		{"ctrl+k ctrl+s", commands.FileSaveAs},
		{"ctrl+k ctrl+p", commands.ViewPalette},
		{"ctrl+k ctrl+o", commands.NavQuickOpen},
		{"ctrl+k ctrl+c", commands.EditComment},
		{"ctrl+k ctrl+l", commands.EditDeleteLine},
		{"ctrl+k ctrl+d", commands.EditDuplicateLine},
		{"ctrl+k ctrl+t", commands.FileReopen},
		{"ctrl+k ctrl+q", commands.FileQuit},
		{"ctrl+k ctrl+f", commands.FindProject},
		{"ctrl+k ctrl+g", commands.ViewKeymap},
		{"ctrl+/", commands.EditComment},
		{"ctrl+_", commands.EditComment},
		{"ctrl+enter", commands.EditInsertLineBelow},
		{"ctrl+shift+enter", commands.EditInsertLineAbove},
	}
}

func axamExtra() [][2]string {
	return [][2]string{
		{"ctrl+d", commands.EditNextOccurrence},
		{"ctrl+shift+d", commands.EditDuplicateLine},
		{"ctrl+shift+k", commands.EditDeleteLine},
		{"ctrl+l", commands.EditSelectLine},
		{"ctrl+shift+l", commands.EditAllOccurrences},
		{"ctrl+p", commands.NavQuickOpen},
		{"ctrl+shift+p", commands.ViewPalette},
		{"ctrl+shift+s", commands.FileSaveAs},
		{"ctrl+shift+t", commands.FileReopen},
		{"ctrl+q", commands.FileQuit},
		{"ctrl+shift+f", commands.FindProject},
		{"ctrl+shift+up", commands.EditMoveLineUp},
		{"ctrl+shift+down", commands.EditMoveLineDown},
		{"alt+shift+up", commands.EditCursorAbove},
		{"alt+shift+down", commands.EditCursorBelow},
		{"ctrl+]", commands.EditIndent},
	}
}

func sublimeExtra() [][2]string {
	return [][2]string{
		{"ctrl+d", commands.EditNextOccurrence},
		{"ctrl+shift+d", commands.EditDuplicateLine},
		{"ctrl+shift+k", commands.EditDeleteLine},
		{"ctrl+l", commands.EditSelectLine},
		{"ctrl+p", commands.NavQuickOpen},
		{"ctrl+shift+p", commands.ViewPalette},
		{"ctrl+shift+s", commands.FileSaveAs},
		{"ctrl+shift+t", commands.FileReopen},
		{"ctrl+q", commands.FileQuit},
		{"ctrl+shift+up", commands.EditMoveLineUp},
		{"ctrl+shift+down", commands.EditMoveLineDown},
		{"alt+shift+down", commands.EditCursorBelow},
		{"alt+shift+up", commands.EditCursorAbove},
		{"ctrl+]", commands.EditIndent},
		{"ctrl+shift+l", commands.EditAllOccurrences},
		{"ctrl+shift+f", commands.FindProject},
	}
}

func vscodeExtra() [][2]string {
	return [][2]string{
		{"ctrl+d", commands.EditNextOccurrence},
		{"ctrl+shift+l", commands.EditAllOccurrences},
		{"ctrl+shift+k", commands.EditDeleteLine},
		{"ctrl+l", commands.EditSelectLine},
		{"ctrl+p", commands.NavQuickOpen},
		{"ctrl+shift+p", commands.ViewPalette},
		{"ctrl+shift+s", commands.FileSaveAs},
		{"ctrl+shift+t", commands.FileReopen},
		{"ctrl+q", commands.FileQuit},
		{"ctrl+shift+f", commands.FindProject},
		{"alt+shift+down", commands.EditDuplicateLine},
		{"alt+shift+up", commands.EditDuplicateLine},
		{"ctrl+alt+up", commands.EditCursorAbove},
		{"ctrl+alt+down", commands.EditCursorBelow},
	}
}

func notepadExtra() [][2]string {
	return [][2]string{
		{"ctrl+d", commands.EditDuplicateLine},
		{"ctrl+l", commands.EditDeleteLine},
		{"ctrl+q", commands.EditComment},
		{"ctrl+shift+up", commands.EditMoveLineUp},
		{"ctrl+shift+down", commands.EditMoveLineDown},
		{"ctrl+shift+s", commands.FileSaveAs},
		{"ctrl+shift+t", commands.FileReopen},
		{"f1", commands.ViewPalette},
		// Ctrl+P is Print in Notepad++. Quick open stays on the chord and palette.
		{"ctrl+shift+f", commands.FindProject},
		{"ctrl+f3", commands.EditNextOccurrence},
	}
}
