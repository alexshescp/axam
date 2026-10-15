// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexshescp/axam/internal/buffer"
	"github.com/alexshescp/axam/internal/clipboard"
	"github.com/alexshescp/axam/internal/commands"
	"github.com/alexshescp/axam/internal/config"
	"github.com/alexshescp/axam/internal/doctor"
	"github.com/alexshescp/axam/internal/editor"
	"github.com/alexshescp/axam/internal/keymap"
	"github.com/alexshescp/axam/internal/project"
	"github.com/alexshescp/axam/internal/search"
	"github.com/alexshescp/axam/internal/syntax"
	"github.com/alexshescp/axam/internal/text"
	"github.com/alexshescp/axam/internal/ui"
)

func (a *App) commandTable() map[string]func() {
	return map[string]func(){
		commands.FileNew:             a.cmdNew,
		commands.FileOpen:            a.cmdOpen,
		commands.FileSave:            func() { a.saveActive(false) },
		commands.FileSaveAs:          a.cmdSaveAs,
		commands.FileClose:           a.cmdClose,
		commands.FileReopen:          a.cmdReopen,
		commands.FileQuit:            a.cmdQuit,
		commands.FileReload:          func() { a.reloadDoc(a.doc()) },
		commands.EditUndo:            func() { a.edit(func(e *editor.Editor) { e.Undo() }) },
		commands.EditRedo:            func() { a.edit(func(e *editor.Editor) { e.Redo() }) },
		commands.EditCut:             a.cmdCut,
		commands.EditCopy:            a.cmdCopy,
		commands.EditPaste:           a.cmdPaste,
		commands.EditSelectAll:       func() { a.edit(func(e *editor.Editor) { e.SelectAll() }) },
		commands.EditSelectWord:      func() { a.edit(func(e *editor.Editor) { e.SelectWord() }) },
		commands.EditSelectLine:      func() { a.edit(func(e *editor.Editor) { e.SelectLine() }) },
		commands.EditDeleteLine:      func() { a.edit(func(e *editor.Editor) { e.DeleteLine() }) },
		commands.EditDuplicateLine:   func() { a.edit(func(e *editor.Editor) { e.DuplicateLine() }) },
		commands.EditMoveLineUp:      func() { a.edit(func(e *editor.Editor) { e.MoveLines(-1) }) },
		commands.EditMoveLineDown:    func() { a.edit(func(e *editor.Editor) { e.MoveLines(1) }) },
		commands.EditIndent:          func() { a.edit(func(e *editor.Editor) { e.Indent() }) },
		commands.EditOutdent:         func() { a.edit(func(e *editor.Editor) { e.Outdent() }) },
		commands.EditComment:         a.cmdComment,
		commands.EditNewline:         func() { a.edit(func(e *editor.Editor) { e.Newline() }) },
		commands.EditBackspace:       func() { a.edit(func(e *editor.Editor) { e.Backspace() }) },
		commands.EditDelete:          func() { a.edit(func(e *editor.Editor) { e.DeleteForward() }) },
		commands.EditDeleteWordLeft:  func() { a.edit(func(e *editor.Editor) { e.DeleteWordLeft() }) },
		commands.EditDeleteWordRight: func() { a.edit(func(e *editor.Editor) { e.DeleteWordRight() }) },
		commands.EditNextOccurrence:  func() { a.edit(func(e *editor.Editor) { e.AddNextOccurrence() }) },
		commands.EditAllOccurrences:  func() { a.edit(func(e *editor.Editor) { e.SelectAllOccurrences() }) },
		commands.EditCursorAbove:     func() { a.edit(func(e *editor.Editor) { e.AddCursorVertical(-1) }) },
		commands.EditCursorBelow:     func() { a.edit(func(e *editor.Editor) { e.AddCursorVertical(1) }) },
		commands.EditInsertLineBelow: a.cmdInsertLine(1),
		commands.EditInsertLineAbove: a.cmdInsertLine(-1),
		commands.MoveLeft:            func() { a.move(func(e *editor.Editor, w int) { e.Left(false) }) },
		commands.MoveRight:           func() { a.move(func(e *editor.Editor, w int) { e.Right(false) }) },
		commands.MoveUp:              func() { a.move(func(e *editor.Editor, w int) { e.Up(false, w) }) },
		commands.MoveDown:            func() { a.move(func(e *editor.Editor, w int) { e.Down(false, w) }) },
		commands.MoveSelectLeft:      func() { a.move(func(e *editor.Editor, w int) { e.Left(true) }) },
		commands.MoveSelectRight:     func() { a.move(func(e *editor.Editor, w int) { e.Right(true) }) },
		commands.MoveSelectUp:        func() { a.move(func(e *editor.Editor, w int) { e.Up(true, w) }) },
		commands.MoveSelectDown:      func() { a.move(func(e *editor.Editor, w int) { e.Down(true, w) }) },
		commands.MoveHome:            func() { a.move(func(e *editor.Editor, w int) { e.Home(false) }) },
		commands.MoveEnd:             func() { a.move(func(e *editor.Editor, w int) { e.End(false) }) },
		commands.MoveSelectHome:      func() { a.move(func(e *editor.Editor, w int) { e.Home(true) }) },
		commands.MoveSelectEnd:       func() { a.move(func(e *editor.Editor, w int) { e.End(true) }) },
		commands.MoveWordLeft:        func() { a.move(func(e *editor.Editor, w int) { e.WordLeft(false) }) },
		commands.MoveWordRight:       func() { a.move(func(e *editor.Editor, w int) { e.WordRight(false) }) },
		commands.MoveSelectWordLeft:  func() { a.move(func(e *editor.Editor, w int) { e.WordLeft(true) }) },
		commands.MoveSelectWordRight: func() { a.move(func(e *editor.Editor, w int) { e.WordRight(true) }) },
		commands.MoveDocStart:        func() { a.move(func(e *editor.Editor, w int) { e.DocStart(false) }) },
		commands.MoveDocEnd:          func() { a.move(func(e *editor.Editor, w int) { e.DocEnd(false) }) },
		commands.MoveSelectDocStart:  func() { a.move(func(e *editor.Editor, w int) { e.DocStart(true) }) },
		commands.MoveSelectDocEnd:    func() { a.move(func(e *editor.Editor, w int) { e.DocEnd(true) }) },
		commands.MovePageUp:          func() { a.page(-1, false) },
		commands.MovePageDown:        func() { a.page(1, false) },
		commands.MoveSelectPageUp:    func() { a.page(-1, true) },
		commands.MoveSelectPageDown:  func() { a.page(1, true) },
		commands.FindOpen:            a.cmdFind,
		commands.FindReplace:         a.cmdReplace,
		commands.FindNext:            func() { a.findGo(1) },
		commands.FindPrev:            func() { a.findGo(-1) },
		commands.FindReplaceOne:      a.replaceOne,
		commands.FindReplaceAll:      a.replaceAll,
		commands.FindGoto:            a.cmdGoto,
		commands.FindProject:         a.cmdProjectSearch,
		commands.NavQuickOpen:        a.cmdQuick,
		commands.NavNextTab:          func() { a.cycleTab(1) },
		commands.NavPrevTab:          func() { a.cycleTab(-1) },
		commands.NavExplorer:         a.cmdExplorer,
		commands.NavSplit:            a.cmdSplit,
		commands.NavFocusExpl:        func() { a.focus = focusExplorer },
		commands.NavFocusEditor:      func() { a.focus = focusEditor },
		commands.ViewWrap:            a.cmdWrap,
		commands.ViewLineNumbers:     func() { a.cfg.LineNumbers = !a.cfg.LineNumbers },
		commands.ViewFold:            a.cmdFold,
		commands.ViewShortcuts:       a.cmdShortcuts,
		commands.ViewPalette:         a.cmdPalette,
		commands.ViewKeymap:          a.cmdShortcuts,
		commands.ViewDoctor:          a.cmdDoctor,
		commands.ViewCancel:          a.cmdCancel,
		commands.ConfigOpen:          a.cmdConfig,
		commands.ConfigReload:        a.cmdReloadConfig,
	}
}

func (a *App) exec(id string) {
	fn := a.cmds[id]
	if fn == nil {
		a.setStatus("Unknown command " + id)
		return
	}
	fn()
}

func (a *App) edit(fn func(*editor.Editor)) {
	d := a.doc()
	if d == nil {
		return
	}
	if d.ReadOnly {
		a.setStatus("Read-only")
		return
	}
	fn(d.Ed)
}

func (a *App) move(fn func(*editor.Editor, int)) {
	d := a.doc()
	if d == nil {
		return
	}
	fn(d.Ed, a.wrapWidth())
	a.revealCursor()
}

func (a *App) page(dir int, extend bool) {
	d := a.doc()
	if d == nil {
		return
	}
	h := a.editorHeight()
	if h < 1 {
		h = 1
	}
	d.Ed.Page(extend, dir, h-1, a.wrapWidth())
}

func (a *App) cmdNew() {
	n := 1
	for _, d := range a.docs {
		if strings.HasPrefix(d.Name, "Untitled") {
			n++
		}
	}
	a.OpenString(fmt.Sprintf("Untitled-%d", n), "")
	a.setStatus("New file")
}

func (a *App) cmdOpen() {
	a.promptPath("Open", func(path string) {
		if err := a.OpenPath(strings.TrimSpace(path)); err != nil {
			a.setStatus(err.Error())
		}
	})
}

func (a *App) cmdSaveAs() {
	d := a.doc()
	if d == nil {
		return
	}
	a.promptPath("Save as", func(path string) {
		d.Path = strings.TrimSpace(path)
		d.Name = filepath.Base(d.Path)
		d.Lang = syntax.Language(d.Path)
		a.writeDoc(d, true)
	})
}

func (a *App) cmdClose() {
	d := a.doc()
	if d == nil {
		a.quit = true
		return
	}
	a.closeIndex(a.panes[a.active].Doc, false)
}

func (a *App) cmdReopen() {
	if len(a.closed) == 0 {
		a.setStatus("Nothing to reopen")
		return
	}
	c := a.closed[len(a.closed)-1]
	a.closed = a.closed[:len(a.closed)-1]
	if c.path != "" {
		if err := a.OpenPath(c.path); err == nil {
			return
		}
	}
	d := a.OpenString(c.name, c.text)
	if c.path != "" {
		d.Path = c.path
		_, err := os.Stat(c.path)
		d.Exists = err == nil
	}
	d.Ed.Buf.SetEOL(c.eol)
}

func (a *App) cmdQuit() {
	for _, d := range a.docs {
		if d.Ed.Dirty() {
			a.confirm("Quit without saving?", func() { a.quit = true }, func() {})
			return
		}
	}
	a.quit = true
}

func (a *App) cmdCopy() {
	d := a.doc()
	if d == nil {
		return
	}
	s, line := d.Ed.Copy(a.info.CopyLine)
	if s == "" {
		a.setStatus("Nothing to copy")
		return
	}
	a.clip = clipboard.Clip{Text: s, Line: line}
	a.publishClip(s)
	a.setStatus(fmt.Sprintf("Copied %d characters to the editor clipboard", len(s)))
}

func (a *App) cmdCut() {
	d := a.doc()
	if d == nil {
		return
	}
	if d.ReadOnly {
		a.setStatus("Read-only")
		return
	}
	s, line := d.Ed.Cut(a.info.CutLine)
	if s == "" {
		a.setStatus("Nothing to cut")
		return
	}
	a.clip = clipboard.Clip{Text: s, Line: line}
	a.publishClip(s)
	a.setStatus(fmt.Sprintf("Cut %d characters to the editor clipboard", len(s)))
}

func (a *App) cmdPaste() {
	if a.clip.Text == "" {
		a.setStatus("Editor clipboard is empty")
		return
	}
	d := a.doc()
	if d == nil {
		return
	}
	if d.ReadOnly {
		a.setStatus("Read-only")
		return
	}
	d.Ed.Paste(a.clip.Text)
	a.setStatus("Pasted")
}

func (a *App) publishClip(s string) {
	if a.screen != nil && a.cfg.Clipboard != "internal" {
		a.screen.SetClipboard([]byte(s))
	}
	if err := clipboard.WriteExternal(a.cfg.Clipboard, s); err != nil {
		a.setStatus("Editor clipboard updated; local clipboard failed: " + err.Error())
	}
}

func (a *App) cmdComment() {
	d := a.doc()
	if d == nil {
		return
	}
	p := syntax.CommentPrefix(d.Lang)
	if p == "" {
		a.setStatus("No line comment for " + d.Lang)
		return
	}
	if !d.Ed.ToggleComment(p) {
		a.setStatus("Read-only")
	}
}

func (a *App) cmdInsertLine(dir int) func() {
	return func() {
		d := a.doc()
		if d == nil || d.ReadOnly {
			return
		}
		if dir > 0 {
			d.Ed.End(false)
			d.Ed.Newline()
		} else {
			d.Ed.Home(false)
			// Move to column 0 unconditionally, then insert a newline before.
			c := d.Ed.Primary()
			d.Ed.SetCaret(text.Pos{Line: c.Head.Line, Col: 0}, false)
			d.Ed.Insert("\n", false)
			d.Ed.Up(false, 0)
		}
	}
}

func (a *App) cmdCancel() {
	if len(a.chord) > 0 {
		a.chord = nil
		a.setStatus("")
		return
	}
	d := a.doc()
	if d == nil {
		return
	}
	if len(d.Ed.Cursors) > 1 {
		c := d.Ed.Primary()
		d.Ed.SetCaret(c.Head, false)
		return
	}
	d.Ed.ClearSelection()
}

func (a *App) cmdFind()    { a.openFind(false) }
func (a *App) cmdReplace() { a.openFind(true) }

func (a *App) openFind(repl bool) {
	d := a.doc()
	if d == nil {
		return
	}
	if sel := d.Ed.SelectedText(); sel != "" && !strings.Contains(sel, "\n") {
		a.find.query = sel
	}
	a.find.replace = repl
	a.find.field = 0
	a.find.pos = len(a.find.query)
	a.mode = modeFind
	a.runFind()
}

func (a *App) runFind() {
	d := a.doc()
	a.find.matches = nil
	a.find.err = ""
	if d == nil || a.find.query == "" {
		return
	}
	ms, err := search.Find(d.Ed.Buf.Lines(), search.Options{
		Pattern: a.find.query, CaseSensitive: a.find.sens, WholeWord: a.find.word, Regex: a.find.regex,
	})
	if err != nil {
		a.find.err = err.Error()
		return
	}
	head := d.Ed.Primary().Head
	a.find.at = 0
	for i, m := range ms {
		a.find.matches = append(a.find.matches, findMatch{m.Start.Line, m.Start.Col, m.End.Line, m.End.Col})
		if !m.Start.Less(head) && a.find.at == 0 && i > 0 {
			a.find.at = i
		}
	}
}

func (a *App) findGo(dir int) {
	if a.mode != modeFind {
		a.openFind(false)
	}
	if len(a.find.matches) == 0 {
		a.runFind()
	}
	if len(a.find.matches) == 0 {
		a.setStatus("No matches")
		return
	}
	a.find.at = (a.find.at + dir + len(a.find.matches)) % len(a.find.matches)
	m := a.find.matches[a.find.at]
	d := a.doc()
	if d == nil {
		return
	}
	d.Ed.SetCaret(text.Pos{Line: m.Line, Col: m.Col}, false)
	d.Ed.SetCaret(text.Pos{Line: m.EndLine, Col: m.EndCol}, true)
	// SetCaret with extend keeps anchor only if we didn't reset. Do it properly:
	d.Ed.SetCaret(text.Pos{Line: m.Line, Col: m.Col}, false)
	end := text.Pos{Line: m.EndLine, Col: m.EndCol}
	cur := d.Ed.Primary()
	cur.Anchor = text.Pos{Line: m.Line, Col: m.Col}
	cur.Head = end
	d.Ed.SetPrimary(cur)
	a.revealCursor()
}

func (a *App) replaceOne() {
	d := a.doc()
	if d == nil || d.ReadOnly || len(a.find.matches) == 0 {
		return
	}
	if a.find.at < 0 || a.find.at >= len(a.find.matches) {
		a.find.at = 0
	}
	m := a.find.matches[a.find.at]
	d.Ed.ReplaceRange(text.Pos{Line: m.Line, Col: m.Col}, text.Pos{Line: m.EndLine, Col: m.EndCol}, a.find.repl)
	a.runFind()
	a.setStatus("Replaced 1")
}

func (a *App) replaceAll() {
	d := a.doc()
	if d == nil || d.ReadOnly {
		return
	}
	a.runFind()
	ranges := make([][2]text.Pos, len(a.find.matches))
	for i, m := range a.find.matches {
		ranges[i] = [2]text.Pos{{Line: m.Line, Col: m.Col}, {Line: m.EndLine, Col: m.EndCol}}
	}
	n := d.Ed.ReplaceRanges(ranges, a.find.repl)
	a.runFind()
	a.setStatus(fmt.Sprintf("Replaced %d", n))
}

func (a *App) cmdGoto() {
	a.mode = modePrompt
	a.prompt = promptBox{title: "Go to line", kind: "line", onOK: func(s string) {
		n := 0
		fmt.Sscan(s, &n)
		if d := a.doc(); d != nil && n > 0 {
			d.Ed.GoToLine(n)
			a.revealCursor()
		}
	}}
}

func (a *App) cmdPalette() {
	var items []listItem
	for _, spec := range commands.Catalog() {
		ch := ""
		if cs := a.keys.ChordsFor(spec.ID); len(cs) > 0 {
			ch = cs[0]
		}
		items = append(items, listItem{Title: spec.Title, Detail: spec.Category, Shortcut: ch, ID: spec.ID})
	}
	a.openList(listPalette, "Command palette", items)
}

func (a *App) cmdQuick() {
	var items []listItem
	for _, f := range a.files {
		rel := f
		if a.root != "" {
			if r, err := filepath.Rel(a.root, f); err == nil {
				rel = r
			}
		}
		items = append(items, listItem{Title: rel, Payload: f})
	}
	if len(items) == 0 {
		for _, f := range a.recent.RecentFiles {
			items = append(items, listItem{Title: f, Payload: f, Detail: "recent"})
		}
	}
	a.openList(listQuick, "Quick open", items)
}

func (a *App) cmdProjectSearch() {
	a.openList(listSearch, "Search in project", nil)
	a.list.waiting = true
}

func (a *App) openList(kind int, title string, items []listItem) {
	a.mode = modeList
	a.list = listBox{kind: kind, title: title, items: items, all: items}
}

func (a *App) pickList() {
	if a.list.sel < 0 || a.list.sel >= len(a.list.items) {
		return
	}
	it := a.list.items[a.list.sel]
	kind := a.list.kind
	a.mode = modeEdit
	switch kind {
	case listPalette:
		a.exec(it.ID)
	case listQuick, listRecent:
		if err := a.OpenPath(it.Payload); err != nil {
			a.setStatus(err.Error())
		}
	case listSearch:
		if it.Payload == "" {
			return
		}
		if err := a.OpenPath(it.Payload); err != nil {
			a.setStatus(err.Error())
			return
		}
		if d := a.doc(); d != nil {
			d.Ed.GoToLine(it.Line)
			d.Ed.SetCaret(text.Pos{Line: it.Line - 1, Col: it.Col - 1}, false)
		}
	}
}

func (a *App) runProjectSearch() {
	if a.root == "" || a.list.query == "" {
		a.setStatus("Open a folder before searching a project")
		return
	}
	opt := search.Options{Pattern: a.list.query, Regex: a.find.regex, CaseSensitive: a.find.sens, WholeWord: a.find.word}
	hits, err := project.Search(context.Background(), a.files, opt, 400)
	if err != nil {
		a.list.waiting = true
		a.setStatus(err.Error())
		return
	}
	var items []listItem
	for _, h := range hits {
		rel := h.Path
		if r, err := filepath.Rel(a.root, h.Path); err == nil {
			rel = r
		}
		items = append(items, listItem{
			Title: fmt.Sprintf("%s:%d", rel, h.Line), Detail: h.Text, Payload: h.Path, Line: h.Line, Col: h.Col,
		})
	}
	a.list.items = items
	a.list.all = items
	a.list.sel = 0
	a.list.waiting = false
	a.list.searched = a.list.query
	a.setStatus(fmt.Sprintf("%d matches", len(items)))
}

func (a *App) cycleTab(dir int) {
	if len(a.docs) == 0 {
		return
	}
	i := a.panes[a.active].Doc
	i = (i + dir + len(a.docs)) % len(a.docs)
	a.panes[a.active].Doc = i
	a.panes[a.active].Top = 0
}

func (a *App) cmdExplorer() {
	a.expl.on = !a.expl.on
	if a.expl.on && len(a.expl.entries) == 0 {
		a.loadExplorer()
	}
	if a.expl.on {
		a.focus = focusExplorer
	} else {
		a.focus = focusEditor
	}
}

func (a *App) cmdSplit() {
	if len(a.panes) < 2 {
		src := pane{Doc: -1}
		if len(a.panes) == 1 {
			src = a.panes[0]
		}
		a.panes = append(a.panes, src)
		a.active = 1
	} else {
		keep := a.panes[a.active]
		a.panes = []pane{keep}
		a.active = 0
	}
}

func (a *App) cmdWrap() {
	a.cfg.WordWrap = !a.cfg.WordWrap
	if d := a.doc(); d != nil {
		d.Ed.ClearDesired()
	}
}

func (a *App) cmdFold() {
	d := a.doc()
	if d == nil {
		return
	}
	line := d.Ed.Primary().Head.Line
	if _, ok := d.Folds[line]; ok {
		delete(d.Folds, line)
		return
	}
	end := foldEnd(d.Ed.Buf, line)
	if end > line {
		d.Folds[line] = end
	} else {
		a.setStatus("Nothing to fold")
	}
}

func foldEnd(buf *buffer.Buffer, line int) int {
	base := indentOf(buf.Line(line))
	end := line
	for i := line + 1; i < buf.LineCount(); i++ {
		s := buf.Line(i)
		if strings.TrimSpace(s) == "" {
			end = i
			continue
		}
		if indentOf(s) > base {
			end = i
			continue
		}
		break
	}
	return end
}

func indentOf(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' {
			n++
		} else if r == '\t' {
			n += 4
		} else {
			break
		}
	}
	return n
}

func (a *App) cmdShortcuts() {
	var b strings.Builder
	fmt.Fprintf(&b, "axam keymap — %s\n\n", a.info.Title)
	b.WriteString("Profile | Shortcut | Command | Notes\n")
	for _, binding := range a.keys.Bindings() {
		env := "legacy and enhanced"
		s := binding.String()
		if strings.Contains(s, "shift+") && strings.Contains(s, "ctrl+") && !strings.Contains(s, "left") && !strings.Contains(s, "right") && !strings.Contains(s, "up") && !strings.Contains(s, "down") && !strings.Contains(s, "home") && !strings.Contains(s, "end") && !strings.Contains(s, "tab") && !strings.Contains(s, "enter") {
			env = "enhanced keyboard only; Ctrl+K chord is the fallback"
		}
		fmt.Fprintf(&b, "%s | %s | %s | %s\n", a.info.ID, binding.Display(), binding.Command, env)
	}
	b.WriteString("\nCommand, Super, and macOS Command shortcuts are not bound.\n")
	b.WriteString("Ctrl+K starts a chord. Esc cancels it.\n")
	d := a.OpenString("Keymap", b.String())
	d.Ed.ReadOnly = true
	d.ReadOnly = true
}

func (a *App) cmdDoctor() {
	w, h := 0, 0
	colors := 0
	if a.screen != nil {
		w, h = a.screen.Size()
		colors = a.screen.Colors()
	}
	text := doctor.Collect(w, h, colors).String()
	d := a.OpenString("Doctor", text)
	d.Ed.ReadOnly = true
	d.ReadOnly = true
}

func (a *App) cmdConfig() {
	path, err := config.Ensure()
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	if err := a.OpenPath(path); err != nil {
		a.setStatus(err.Error())
	}
}

func (a *App) cmdReloadConfig() {
	cfg, _, err := config.Load(config.Path())
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	km, info, err := keymap.Load(cfg.KeymapProfile)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	for chord, cmd := range cfg.Bindings {
		keys, err := keymap.ParseChord(chord)
		if err != nil {
			a.setStatus(err.Error())
			return
		}
		km.Set(keys, cmd)
	}
	a.cfg = cfg
	a.keys = km
	a.info = info
	colors := 256
	if a.screen != nil {
		colors = a.screen.Colors()
	}
	a.theme = ui.For(cfg.Theme, colors)
	a.setStatus("Configuration reloaded")
}

func (a *App) revealCursor() { a.reveal() }
