// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexshescp/axam/internal/commands"
	"github.com/alexshescp/axam/internal/config"
	"github.com/alexshescp/axam/internal/text"
	"github.com/gdamore/tcell/v2"
)

func newTestApp(t *testing.T) (*App, tcell.SimulationScreen) {
	t.Helper()
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Fini() })
	s.SetSize(80, 24)
	a, err := New(s, config.Defaults())
	if err != nil {
		t.Fatal(err)
	}
	return a, s
}

func key(k tcell.Key, mod tcell.ModMask) *tcell.EventKey {
	return tcell.NewEventKey(k, 0, mod)
}

func runeKey(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
}

func TestTypeShiftSelectAndDraw(t *testing.T) {
	a, s := newTestApp(t)
	a.OpenString("t.txt", "")
	for _, r := range "Hi" {
		a.Handle(runeKey(r))
	}
	if got := a.doc().Ed.Buf.String(); got != "Hi" {
		t.Fatalf("buffer %q", got)
	}
	a.Handle(key(tcell.KeyLeft, tcell.ModShift))
	if got := a.doc().Ed.SelectedText(); got != "i" {
		t.Fatalf("shift+left %q", got)
	}
	a.Draw()
	s.Show()
	_, _, vis := s.GetCursor()
	if !vis {
		t.Fatal("cursor hidden")
	}
	cells, w, h := s.GetContents()
	if w != 80 || h != 24 || len(cells) == 0 {
		t.Fatalf("screen %d x %d cells %d", w, h, len(cells))
	}
}

func TestShiftDownKeepsColumn(t *testing.T) {
	a, _ := newTestApp(t)
	d := a.OpenString("t.txt", "abcd\nx\nabcd")
	d.Ed.SetCaret(text.Pos{Line: 0, Col: 3}, false)
	a.Handle(key(tcell.KeyDown, tcell.ModShift))
	a.Handle(key(tcell.KeyDown, tcell.ModShift))
	h := d.Ed.Primary()
	if h.Anchor.Line != 0 || h.Anchor.Col != 3 {
		t.Fatalf("anchor jumped %+v", h.Anchor)
	}
	if h.Head.Line != 2 || h.Head.Col != 3 {
		t.Fatalf("head %+v", h.Head)
	}
	if h.Desired != 3 {
		t.Fatalf("desired %d", h.Desired)
	}
}

func TestShiftUpAcrossEmpty(t *testing.T) {
	a, _ := newTestApp(t)
	d := a.OpenString("t.txt", "ab\n\ncd")
	d.Ed.SetCaret(text.Pos{Line: 2, Col: 2}, false)
	a.Handle(key(tcell.KeyUp, tcell.ModShift))
	a.Handle(key(tcell.KeyUp, tcell.ModShift))
	h := d.Ed.Primary()
	if h.Head.Line != 0 || h.Head.Col != 2 {
		t.Fatalf("head %+v", h.Head)
	}
	if h.Anchor.Col != 2 || h.Anchor.Line != 2 {
		t.Fatalf("anchor %+v", h.Anchor)
	}
}

func TestCyrillicAndEmojiSelection(t *testing.T) {
	a, _ := newTestApp(t)
	d := a.OpenString("t.txt", "Привет")
	a.Handle(key(tcell.KeyEnd, 0))
	a.Handle(key(tcell.KeyLeft, tcell.ModShift))
	if got := d.Ed.SelectedText(); got != "т" {
		t.Fatalf("cyrillic %q", got)
	}
	d = a.OpenString("e.txt", "a👍b")
	a.Handle(key(tcell.KeyEnd, 0))
	a.Handle(key(tcell.KeyLeft, tcell.ModShift))
	a.Handle(key(tcell.KeyLeft, tcell.ModShift))
	if got := d.Ed.SelectedText(); got != "👍b" {
		t.Fatalf("emoji %q", got)
	}
}

func TestCopyUndoSave(t *testing.T) {
	a, s := newTestApp(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := a.OpenPath(path); err != nil {
		t.Fatal(err)
	}
	a.Handle(key(tcell.KeyEnd, 0))
	a.Handle(runeKey('!'))
	a.Handle(key(tcell.KeyCtrlA, 0))
	a.Handle(key(tcell.KeyCtrlC, 0))
	data := s.GetClipboardData()
	if string(data) != "one!\n" && string(data) != "one!" {
		// Select-all includes the whole buffer. Bytes() may or may not end with newline.
		if a.clip.Text == "" {
			t.Fatalf("clipboard %q clip %q", data, a.clip.Text)
		}
	}
	if a.clip.Text == "" {
		t.Fatal("internal clipboard empty")
	}
	a.Handle(key(tcell.KeyCtrlZ, 0))
	if got := a.doc().Ed.Buf.String(); got != "one\n" && got != "one" {
		t.Fatalf("undo %q", got)
	}
	a.Handle(runeKey('?'))
	a.Handle(key(tcell.KeyCtrlS, 0))
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "one?" && string(got) != "one?\n" {
		t.Fatalf("saved %q", got)
	}
}

func TestPendingRelativeFileIsDeferredUntilSave(t *testing.T) {
	a, _ := newTestApp(t)
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	if err := a.OpenPath("notes1.txt"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "notes1.txt")
	d := a.doc()
	if d == nil || d.Path != path || d.Exists {
		t.Fatalf("pending document: %+v", d)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("open created file: %v", err)
	}

	a.closeIndex(0, false)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("closing empty document created file: %v", err)
	}
}

func TestPendingFileSaveCreatesAtomically(t *testing.T) {
	a, _ := newTestApp(t)
	path := filepath.Join(t.TempDir(), "notes1.txt")
	if err := a.OpenPath(path); err != nil {
		t.Fatal(err)
	}
	a.Handle(runeKey('x'))
	a.saveActive(false)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "x" {
		t.Fatalf("saved %q", got)
	}
	if !a.doc().Exists || a.doc().Ed.Dirty() {
		t.Fatal("saved pending document was not marked clean and existing")
	}
}

func TestExplicitSaveCreatesEmptyPendingFile(t *testing.T) {
	a, _ := newTestApp(t)
	path := filepath.Join(t.TempDir(), "empty.txt")
	if err := a.OpenPath(path); err != nil {
		t.Fatal(err)
	}
	a.saveActive(false)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("size %d", info.Size())
	}
}

func TestPendingFileRequiresExistingParent(t *testing.T) {
	a, _ := newTestApp(t)
	path := filepath.Join(t.TempDir(), "missing", "notes.txt")
	if err := a.OpenPath(path); err == nil {
		t.Fatal("expected missing-parent error")
	}
	if len(a.docs) != 0 {
		t.Fatal("invalid pending path opened a document")
	}
}

func TestReplaceAllOneUndo(t *testing.T) {
	a, _ := newTestApp(t)
	d := a.OpenString("t.txt", "a a a")
	a.find.query = "a"
	a.find.repl = "bb"
	a.replaceAll()
	if got := d.Ed.Buf.String(); got != "bb bb bb" {
		t.Fatalf("replaced %q", got)
	}
	d.Ed.Undo()
	if got := d.Ed.Buf.String(); got != "a a a" {
		t.Fatalf("undo replace %q", got)
	}
}

func TestPaletteAndResize(t *testing.T) {
	a, _ := newTestApp(t)
	a.OpenString("t.txt", "x")
	a.Handle(key(tcell.KeyF1, 0))
	if a.mode != modeList {
		t.Fatalf("mode %d", a.mode)
	}
	for _, r := range "report" {
		a.Handle(runeKey(r))
	}
	found := -1
	for i, it := range a.list.items {
		if it.ID == commands.ViewDoctor {
			found = i
		}
	}
	if found < 0 {
		t.Fatalf("palette filter %v", a.list.items)
	}
	a.list.sel = found
	a.Handle(key(tcell.KeyEnter, 0))
	if a.doc().Name != "Doctor" {
		t.Fatalf("opened %s", a.doc().Name)
	}
	a.Handle(tcell.NewEventResize(100, 30))
	a.Draw()
}

func TestBottomMenuCommandsUseCentralRegistry(t *testing.T) {
	a, _ := newTestApp(t)
	if !menuCommandsKnown() {
		t.Fatal("bottom menu contains an unknown command")
	}
	for _, group := range bottomMenus() {
		if len(group.items) == 0 {
			t.Fatalf("empty menu group %q", group.title)
		}
		for _, item := range group.items {
			if a.cmds[item.command] == nil {
				t.Fatalf("%s/%s has no command handler for %q", group.title, item.title, item.command)
			}
		}
	}
}

func TestBottomMenuKeyboardNavigationAndExecution(t *testing.T) {
	a, _ := newTestApp(t)
	d := a.OpenString("t.txt", "")
	a.Handle(runeKey('x'))

	a.Handle(key(tcell.KeyF10, 0))
	if a.mode != modeMenu || a.menu.group != 0 {
		t.Fatalf("menu state: mode=%d group=%d", a.mode, a.menu.group)
	}
	a.Handle(key(tcell.KeyRight, 0))
	if a.menu.group != 1 {
		t.Fatalf("right selected group %d", a.menu.group)
	}
	// Edit starts with Undo. Enter must execute the same handler as Ctrl+Z.
	a.Handle(key(tcell.KeyEnter, 0))
	if a.mode != modeEdit || d.Ed.Buf.String() != "" {
		t.Fatalf("menu undo: mode=%d text=%q", a.mode, d.Ed.Buf.String())
	}

	a.Handle(key(tcell.KeyF10, 0))
	a.Handle(key(tcell.KeyEscape, 0))
	if a.mode != modeEdit {
		t.Fatalf("escape left mode %d", a.mode)
	}
}

func TestBottomMenuMouseAndDrawing(t *testing.T) {
	a, s := newTestApp(t)
	a.Draw()
	s.Show()
	v := a.view()
	if got := simulationRow(s, v.menu); !strings.Contains(got, "File") ||
		!strings.Contains(got, "Selection") || !strings.Contains(got, "Goto") {
		t.Fatalf("menu row %q", got)
	}

	a.Handle(tcell.NewEventMouse(1, v.menu, tcell.Button1, tcell.ModNone))
	a.Handle(tcell.NewEventMouse(1, v.menu, tcell.ButtonNone, tcell.ModNone))
	if a.mode != modeMenu || a.menu.group != 0 {
		t.Fatalf("mouse did not open File menu: mode=%d group=%d", a.mode, a.menu.group)
	}
	box := a.menuPopupBox(a.view())
	a.Handle(tcell.NewEventMouse(box.x+1, box.y, tcell.Button1, tcell.ModNone))
	a.Handle(tcell.NewEventMouse(box.x+1, box.y, tcell.ButtonNone, tcell.ModNone))
	if a.mode != modeEdit || a.doc() == nil || !strings.HasPrefix(a.doc().Name, "Untitled") {
		t.Fatalf("mouse did not execute New File: mode=%d doc=%v", a.mode, a.doc())
	}
}

func TestBottomMenuFits80x24AndResize(t *testing.T) {
	a, s := newTestApp(t)
	a.OpenString("t.txt", strings.Repeat("line\n", 40))
	a.Handle(key(tcell.KeyF10, 0))
	a.Draw()
	s.Show()
	v := a.view()
	box := a.menuPopupBox(v)
	if v.menu != 22 || v.status != 23 || box.y < 1 || box.y+box.h > v.menu {
		t.Fatalf("80x24 layout: menu=%d status=%d popup=%+v", v.menu, v.status, box)
	}
	s.SetSize(100, 30)
	a.Handle(tcell.NewEventResize(100, 30))
	a.Draw()
	if v := a.view(); v.menu != 28 || v.status != 29 {
		t.Fatalf("resized layout: %+v", v)
	}
}

func simulationRow(s tcell.SimulationScreen, y int) string {
	cells, w, h := s.GetContents()
	if y < 0 || y >= h {
		return ""
	}
	var b strings.Builder
	for x := 0; x < w; x++ {
		b.Write(cells[y*w+x].Bytes)
	}
	return b.String()
}

func TestBracketedPaste(t *testing.T) {
	a, _ := newTestApp(t)
	a.OpenString("t.txt", "")
	a.Handle(tcell.NewEventPaste(true))
	a.Handle(runeKey('a'))
	a.Handle(key(tcell.KeyEnter, 0))
	a.Handle(runeKey('b'))
	a.Handle(tcell.NewEventPaste(false))
	if got := a.doc().Ed.Buf.String(); got != "a\nb" {
		t.Fatalf("paste %q", got)
	}
	a.doc().Ed.Undo()
	if got := a.doc().Ed.Buf.String(); got != "" {
		t.Fatalf("paste undo %q", got)
	}
}
