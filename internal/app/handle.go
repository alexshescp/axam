// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package app

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexshescp/axam/internal/editor"
	"github.com/alexshescp/axam/internal/input"
	"github.com/alexshescp/axam/internal/keymap"
	"github.com/alexshescp/axam/internal/text"
	"github.com/gdamore/tcell/v2"
)

// Handle applies one terminal event. Tests call this directly.
func (a *App) Handle(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		w, h := ev.Size()
		if w > 0 && h > 0 {
			a.w, a.h = w, h
		}
		if a.screen != nil {
			a.screen.Sync()
		}
	case *tcell.EventInterrupt:
		a.quit = true
	case *tcell.EventPaste:
		if ev.Start() {
			a.pasting = true
			a.pasteBuf = a.pasteBuf[:0]
		}
		if ev.End() {
			a.pasting = false
			if len(a.pasteBuf) > 0 {
				a.insertText(string(a.pasteBuf), false)
			}
			a.pasteBuf = nil
		}
	case *tcell.EventKey:
		if a.pasting {
			if s := input.PasteText(ev); s != "" {
				a.pasteBuf = append(a.pasteBuf, s...)
			}
			return
		}
		a.onKey(ev)
	case *tcell.EventMouse:
		a.onMouse(ev)
	}
}

func (a *App) onKey(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyF10 && a.mode != modePrompt {
		if a.mode == modeMenu {
			a.closeMenu()
		} else {
			a.openMenu(0)
		}
		return
	}
	switch a.mode {
	case modeFind:
		a.onFindKey(ev)
	case modePrompt:
		a.onPromptKey(ev)
	case modeList:
		a.onListKey(ev)
	case modeMenu:
		a.onMenuKey(ev)
	default:
		if a.focus == focusExplorer && a.expl.on && a.onExplorerKey(ev) {
			return
		}
		a.onEditKey(input.Normalize(ev))
	}
}

func (a *App) onEditKey(res input.Result) {
	if !res.Command {
		if res.Insert != "" {
			a.insertText(res.Insert, true)
		}
		return
	}
	if res.Key.Name == "escape" && res.Key.Mod == 0 && len(a.chord) == 0 {
		a.cmdCancel()
		return
	}
	seq := append(append([]keymap.Key{}, a.chord...), res.Key)
	cmd, prefix := a.keys.Resolve(seq)
	if prefix {
		a.chord = seq
		a.setStatus(input.FormatChord(seq) + " …")
		return
	}
	a.chord = nil
	if cmd != "" {
		a.exec(cmd)
		return
	}
	if len(seq) > 1 {
		a.setStatus("Unknown chord")
		return
	}
	if res.Insert != "" && res.Key.Mod&(keymap.Ctrl|keymap.Meta) == 0 {
		a.insertText(res.Insert, true)
	}
}

func (a *App) insertText(s string, coalesce bool) {
	d := a.doc()
	if d == nil {
		return
	}
	if d.ReadOnly || d.Ed.ReadOnly {
		a.setStatus("Read-only")
		return
	}
	if coalesce {
		coalesce = len(s) == 1 && s != "\n" && utf8.RuneCountInString(s) == 1
	}
	d.Ed.Insert(s, coalesce)
	a.reveal()
}

func (a *App) onFindKey(ev *tcell.EventKey) {
	res := input.Normalize(ev)
	if res.Key.Name == "escape" && res.Key.Mod == 0 {
		a.mode = modeEdit
		a.setStatus("")
		return
	}
	if res.Key.Name == "enter" && res.Key.Mod == 0 {
		a.findGo(1)
		return
	}
	if res.Key.Name == "enter" && res.Key.Mod == keymap.Alt {
		a.replaceOne()
		return
	}
	if res.Key.Name == "tab" && res.Key.Mod == 0 {
		if a.find.replace {
			a.find.field = 1 - a.find.field
			a.placeFindCaret()
		}
		return
	}
	if res.Key.Name == "up" && res.Key.Mod == 0 {
		a.findGo(-1)
		return
	}
	if res.Key.Name == "down" && res.Key.Mod == 0 {
		a.findGo(1)
		return
	}
	if res.Key.Mod == keymap.Alt && len(res.Key.Name) == 1 {
		switch res.Key.Name {
		case "c":
			a.find.sens = !a.find.sens
			a.runFind()
			return
		case "w":
			a.find.word = !a.find.word
			a.runFind()
			return
		case "r":
			a.find.regex = !a.find.regex
			a.runFind()
			return
		}
	}
	if res.Command && res.Key.Mod&(keymap.Ctrl|keymap.Meta) != 0 && res.Key.Name != "backspace" {
		a.onEditKey(res)
		return
	}
	if a.editQuery(res) {
		a.runFind()
		if len(a.find.matches) > 0 {
			a.find.at = 0
			a.selectFind(a.find.matches[0])
		}
	}
}

func (a *App) placeFindCaret() {
	if a.find.field == 1 {
		a.find.pos = len(a.find.repl)
	} else {
		a.find.pos = len(a.find.query)
	}
}

func (a *App) queryPtr() *string {
	if a.find.field == 1 {
		return &a.find.repl
	}
	return &a.find.query
}

func (a *App) editQuery(res input.Result) bool {
	s := a.queryPtr()
	switch res.Key.Name {
	case "backspace":
		if a.find.pos > 0 {
			_, n := utf8.DecodeLastRuneInString((*s)[:a.find.pos])
			if n > 0 {
				*s = (*s)[:a.find.pos-n] + (*s)[a.find.pos:]
				a.find.pos -= n
			}
		}
		return a.find.field == 0
	case "delete":
		if a.find.pos < len(*s) {
			_, n := utf8.DecodeRuneInString((*s)[a.find.pos:])
			*s = (*s)[:a.find.pos] + (*s)[a.find.pos+n:]
		}
		return a.find.field == 0
	case "left":
		if a.find.pos > 0 {
			_, n := utf8.DecodeLastRuneInString((*s)[:a.find.pos])
			a.find.pos -= n
		}
		return false
	case "right":
		if a.find.pos < len(*s) {
			_, n := utf8.DecodeRuneInString((*s)[a.find.pos:])
			a.find.pos += n
		}
		return false
	case "home":
		a.find.pos = 0
		return false
	case "end":
		a.find.pos = len(*s)
		return false
	}
	if res.Insert != "" && !res.Command {
		*s = (*s)[:a.find.pos] + res.Insert + (*s)[a.find.pos:]
		a.find.pos += len(res.Insert)
		return a.find.field == 0
	}
	return false
}

func (a *App) selectFind(m findMatch) {
	d := a.doc()
	if d == nil {
		return
	}
	start := text.Pos{Line: m.Line, Col: m.Col}
	end := text.Pos{Line: m.EndLine, Col: m.EndCol}
	d.Ed.SetPrimary(editor.Cursor{Anchor: start, Head: end, Desired: -1})
	a.reveal()
}

func (a *App) onPromptKey(ev *tcell.EventKey) {
	res := input.Normalize(ev)
	if a.prompt.kind == "confirm" {
		switch {
		case res.Key.Name == "escape" || res.Insert == "n" || res.Insert == "N":
			no := a.prompt.onNo
			a.mode = modeEdit
			a.prompt = promptBox{}
			if no != nil {
				no()
			}
		case res.Key.Name == "enter" || res.Insert == "y" || res.Insert == "Y":
			yes := a.prompt.onYes
			a.mode = modeEdit
			a.prompt = promptBox{}
			if yes != nil {
				yes()
			}
		}
		return
	}
	if res.Key.Name == "escape" && res.Key.Mod == 0 {
		a.mode = modeEdit
		a.prompt = promptBox{}
		return
	}
	if res.Key.Name == "enter" && res.Key.Mod == 0 {
		ok := a.prompt.onOK
		text := a.prompt.text
		a.mode = modeEdit
		a.prompt = promptBox{}
		if ok != nil {
			ok(text)
		}
		return
	}
	if res.Key.Name == "backspace" {
		if a.prompt.pos > 0 {
			_, n := utf8.DecodeLastRuneInString(a.prompt.text[:a.prompt.pos])
			a.prompt.text = a.prompt.text[:a.prompt.pos-n] + a.prompt.text[a.prompt.pos:]
			a.prompt.pos -= n
		}
		return
	}
	if res.Insert != "" && !res.Command {
		a.prompt.text = a.prompt.text[:a.prompt.pos] + res.Insert + a.prompt.text[a.prompt.pos:]
		a.prompt.pos += len(res.Insert)
	}
}

func (a *App) onListKey(ev *tcell.EventKey) {
	res := input.Normalize(ev)
	if res.Key.Name == "escape" && res.Key.Mod == 0 {
		a.mode = modeEdit
		return
	}
	if res.Key.Name == "enter" && res.Key.Mod == 0 {
		if a.list.kind == listSearch && (a.list.waiting || a.list.searched != a.list.query) {
			a.runProjectSearch()
			return
		}
		a.pickList()
		return
	}
	if res.Key.Name == "up" && res.Key.Mod == 0 {
		if a.list.sel > 0 {
			a.list.sel--
		}
		return
	}
	if res.Key.Name == "down" && res.Key.Mod == 0 {
		if a.list.sel+1 < len(a.list.items) {
			a.list.sel++
		}
		return
	}
	if res.Key.Name == "backspace" {
		if a.list.pos > 0 {
			_, n := utf8.DecodeLastRuneInString(a.list.query[:a.list.pos])
			a.list.query = a.list.query[:a.list.pos-n] + a.list.query[a.list.pos:]
			a.list.pos -= n
			a.afterListEdit()
		}
		return
	}
	if res.Insert != "" && !res.Command {
		a.list.query = a.list.query[:a.list.pos] + res.Insert + a.list.query[a.list.pos:]
		a.list.pos += len(res.Insert)
		a.afterListEdit()
	}
}

func (a *App) afterListEdit() {
	if a.list.kind == listSearch {
		a.list.waiting = true
		return
	}
	a.filterList()
}

func (a *App) filterList() {
	q := strings.ToLower(a.list.query)
	if q == "" {
		a.list.items = append([]listItem(nil), a.list.all...)
		a.list.sel = 0
		return
	}
	var out []listItem
	for _, it := range a.list.all {
		if fuzzy(q, strings.ToLower(it.Title+" "+it.Detail+" "+it.Shortcut)) {
			out = append(out, it)
		}
	}
	a.list.items = out
	if a.list.sel >= len(out) {
		a.list.sel = 0
	}
}

func fuzzy(q, s string) bool {
	i := 0
	qb := []byte(q)
	for si := 0; si < len(s) && i < len(qb); si++ {
		if s[si] == qb[i] {
			i++
		}
	}
	return i == len(qb)
}

func (a *App) onExplorerKey(ev *tcell.EventKey) bool {
	res := input.Normalize(ev)
	if res.Key.Mod != 0 || res.Command && res.Insert == "" && res.Key.Mod != 0 {
		return false
	}
	if res.Command && res.Key.Mod != 0 {
		return false
	}
	switch res.Key.Name {
	case "up":
		if a.expl.sel > 0 {
			a.expl.sel--
		}
		return true
	case "down":
		if a.expl.sel+1 < len(a.expl.entries) {
			a.expl.sel++
		}
		return true
	case "right", "enter":
		a.openExplSel(res.Key.Name == "right")
		return true
	case "left":
		a.collapseExplSel()
		return true
	case "escape":
		a.focus = focusEditor
		return true
	default:
		return res.Insert != ""
	}
}

func (a *App) openExplSel(expandOnly bool) {
	if a.expl.sel < 0 || a.expl.sel >= len(a.expl.entries) {
		return
	}
	e := a.expl.entries[a.expl.sel]
	if e.Dir {
		if e.Open {
			if !expandOnly {
				a.collapseExpl(a.expl.sel)
			}
			return
		}
		a.expandExpl(a.expl.sel)
		return
	}
	if expandOnly {
		return
	}
	if err := a.OpenPath(e.Path); err != nil {
		a.setStatus(err.Error())
	}
	a.focus = focusEditor
}

func (a *App) collapseExplSel() {
	if a.expl.sel < 0 || a.expl.sel >= len(a.expl.entries) {
		return
	}
	e := a.expl.entries[a.expl.sel]
	if e.Dir && e.Open {
		a.collapseExpl(a.expl.sel)
		return
	}
	if e.Depth > 0 && a.expl.sel > 0 {
		a.expl.sel--
	}
}

func (a *App) onMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	btn := ev.Buttons()
	if btn&tcell.WheelUp != 0 {
		a.scrollActive(-3)
		return
	}
	if btn&tcell.WheelDown != 0 {
		a.scrollActive(3)
		return
	}
	if btn&tcell.Button1 == 0 {
		a.click.drag = false
		return
	}
	if !a.click.drag {
		now := time.Now()
		if now.Sub(a.click.when) < 400*time.Millisecond && x == a.click.x && y == a.click.y {
			a.click.count++
		} else {
			a.click.count = 1
		}
		a.click.when = now
		a.click.x, a.click.y = x, y
		a.click.drag = true
		a.click.alt = ev.Modifiers()&tcell.ModAlt != 0
		a.mouseDown(x, y)
		return
	}
	a.mouseDrag(x, y)
}

func (a *App) mouseDown(x, y int) {
	if h, ok := a.hitTest(x, y); ok {
		switch h.kind {
		case hitMenuGroup:
			a.openMenu(h.menuGroup)
		case hitMenuItem:
			a.runMenuItem(h.menuGroup, h.menuItem)
		case hitTab:
			if h.tab >= 0 && h.tab < len(a.docs) {
				a.panes[a.active].Doc = h.tab
				a.focus = focusEditor
				a.mode = modeEdit
			}
		case hitExpl:
			a.focus = focusExplorer
			a.expl.sel = h.expl
			if a.click.count >= 2 {
				a.openExplSel(false)
			}
		case hitText, hitGutter:
			if a.mode == modeMenu {
				a.closeMenu()
			}
			a.focus = focusEditor
			a.active = h.pane
			p := a.pointAt(h.pane, x, y)
			a.click.pane = h.pane
			a.click.anchorLine, a.click.anchorCol = p.Line, p.Col
			d := a.docs[a.panes[h.pane].Doc]
			switch a.click.count {
			case 2:
				d.Ed.SetCaret(p, false)
				d.Ed.SelectWord()
			case 3:
				d.Ed.SetCaret(p, false)
				d.Ed.SelectLine()
			default:
				if a.click.alt {
					d.Ed.SetColumn(p, p)
				} else {
					d.Ed.SetCaret(p, false)
				}
			}
			a.reveal()
		}
	}
}

func (a *App) mouseDrag(x, y int) {
	h, ok := a.hitTest(x, y)
	if !ok || (h.kind != hitText && h.kind != hitGutter) {
		return
	}
	if h.pane != a.click.pane {
		return
	}
	p := a.pointAt(h.pane, x, y)
	d := a.doc()
	if d == nil {
		return
	}
	anchor := text.Pos{Line: a.click.anchorLine, Col: a.click.anchorCol}
	if a.click.alt {
		d.Ed.SetColumn(anchor, p)
	} else {
		d.Ed.SetCaret(anchor, false)
		d.Ed.SetCaret(p, true)
	}
}

func (a *App) scrollActive(delta int) {
	if a.active < 0 || a.active >= len(a.panes) {
		return
	}
	p := &a.panes[a.active]
	if p.Doc < 0 {
		return
	}
	p.Top += delta
	if p.Top < 0 {
		p.Top = 0
	}
	max := a.docs[p.Doc].Ed.Buf.LineCount() - 1
	if max < 0 {
		max = 0
	}
	if p.Top > max {
		p.Top = max
	}
}
