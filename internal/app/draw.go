// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package app

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/codersclub/axam/internal/syntax"
	"github.com/codersclub/axam/internal/text"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

type cellBox struct{ x, y, w, h int }

type view struct {
	w, h   int
	expl   cellBox
	edits  []cellBox
	list   cellBox
	find   cellBox
	prompt cellBox
	menu   int
	status int
}

func (a *App) size() (int, int) {
	if a.screen != nil {
		w, h := a.screen.Size()
		if w > 0 && h > 0 {
			return w, h
		}
	}
	if a.w < 1 {
		return 80, 24
	}
	return a.w, a.h
}

func (a *App) view() view {
	w, h := a.size()
	if h < 4 {
		h = 4
	}
	v := view{w: w, h: h, menu: h - 2, status: h - 1}
	findH := 0
	if a.mode == modeFind {
		findH = 2
		if a.find.replace {
			findH = 3
		}
	}
	promptH := 0
	if a.mode == modePrompt {
		promptH = 1
	}
	listH := 0
	if a.mode == modeList {
		listH = h / 2
		if listH > 14 {
			listH = 14
		}
		if listH < 5 {
			listH = 5
		}
		if listH > h-4 {
			listH = h - 4
		}
	}
	bodyY := 1 + listH
	bodyH := h - 3 - listH - findH - promptH
	if bodyH < 1 {
		bodyH = 1
	}
	explW := 0
	if a.expl.on && w >= 40 {
		explW = w / 5
		if explW < 18 {
			explW = 18
		}
		if explW > 32 {
			explW = 32
		}
		if explW > w/2 {
			explW = w / 3
		}
	}
	v.expl = cellBox{0, bodyY, explW, bodyH}
	v.list = cellBox{0, 1, w, listH}
	v.find = cellBox{0, v.menu - promptH - findH, w, findH}
	v.prompt = cellBox{0, v.menu - promptH, w, promptH}
	n := len(a.panes)
	if n < 1 {
		n = 1
	}
	editW := w - explW
	if editW < 1 {
		editW = 1
	}
	base := editW / n
	extra := editW - base*n
	x := explW
	for i := 0; i < n; i++ {
		ww := base
		if i == n-1 {
			ww += extra
		}
		if ww < 1 {
			ww = 1
		}
		v.edits = append(v.edits, cellBox{x, bodyY, ww, bodyH})
		x += ww
	}
	return v
}

func (a *App) editorHeight() int {
	v := a.view()
	if a.active >= 0 && a.active < len(v.edits) {
		return v.edits[a.active].h
	}
	return 1
}

func (a *App) contentWidth() int {
	v := a.view()
	if a.active < 0 || a.active >= len(v.edits) || a.doc() == nil {
		return 80
	}
	w := v.edits[a.active].w - a.gutterWidth(a.doc())
	if w < 1 {
		return 1
	}
	return w
}

func (a *App) wrapWidth() int {
	if !a.cfg.WordWrap {
		return 0
	}
	return a.contentWidth()
}

func (a *App) gutterWidth(d *Document) int {
	if d == nil || !a.cfg.LineNumbers {
		return 1
	}
	n := d.Ed.Buf.LineCount()
	digits := 1
	for x := 10; n >= x && digits < 6; x *= 10 {
		digits++
	}
	if digits < 3 {
		digits = 3
	}
	return digits + 1
}

func (a *App) hidden(d *Document, line int) bool {
	for start, end := range d.Folds {
		if line > start && line <= end {
			return true
		}
	}
	return false
}

func (a *App) expandFold(d *Document, line int) {
	for start, end := range d.Folds {
		if line > start && line <= end {
			delete(d.Folds, start)
		}
	}
}

func (a *App) reveal() {
	d := a.doc()
	if d == nil || a.active < 0 || a.active >= len(a.panes) {
		return
	}
	head := d.Ed.Primary().Head
	a.expandFold(d, head.Line)
	v := a.view()
	if a.active >= len(v.edits) {
		return
	}
	box := v.edits[a.active]
	cw := box.w - a.gutterWidth(d)
	if cw < 1 {
		cw = 1
	}
	p := &a.panes[a.active]
	if head.Line < p.Top {
		p.Top = head.Line
		p.Row = 0
	}
	width := 0
	if a.cfg.WordWrap {
		width = cw
	}
	tab := d.Ed.TabSize
	rowsBefore := 0
	for line := p.Top; line < head.Line; line++ {
		if a.hidden(d, line) {
			continue
		}
		rowsBefore += len(text.Rows(d.Ed.Buf.Line(line), width, tab))
	}
	lineRows := text.Rows(d.Ed.Buf.Line(head.Line), width, tab)
	caretRow := rowsBefore + text.RowIndex(lineRows, head.Col) - p.Row
	for caretRow >= box.h && p.Top < head.Line {
		n := 1
		if !a.hidden(d, p.Top) {
			n = len(text.Rows(d.Ed.Buf.Line(p.Top), width, tab))
		}
		if n < 1 {
			n = 1
		}
		p.Top++
		p.Row = 0
		caretRow -= n
		for p.Top < head.Line && a.hidden(d, p.Top) {
			p.Top++
		}
	}
	for caretRow >= box.h && p.Row < len(lineRows)-1 {
		p.Row++
		caretRow--
	}
	if p.Top < 0 {
		p.Top = 0
	}
	if a.cfg.WordWrap {
		p.Left = 0
		return
	}
	vis := text.VisualCol(d.Ed.Buf.Line(head.Line), head.Col, tab)
	if vis < p.Left {
		p.Left = vis
	}
	if vis >= p.Left+cw {
		p.Left = vis - cw + 1
	}
	if p.Left < 0 {
		p.Left = 0
	}
}

// Draw paints the session. A nil screen is ignored.
func (a *App) Draw() {
	if a.screen == nil {
		return
	}
	if d := a.doc(); d != nil {
		a.reveal()
	}
	a.screen.Clear()
	a.screen.HideCursor()
	v := a.view()
	bg := style(a.theme.Fg, a.theme.Bg)
	a.fill(0, 0, v.w, v.h, bg)
	a.drawTabs(v)
	if v.expl.w > 0 {
		a.drawExplorer(v.expl)
	}
	for i := range v.edits {
		a.drawPane(i, v.edits[i])
	}
	if v.list.h > 0 {
		a.drawList(v.list)
	}
	if v.find.h > 0 {
		a.drawFind(v.find)
	}
	if v.prompt.h > 0 {
		a.drawPrompt(v.prompt)
	}
	a.drawBottomMenu(v)
	a.drawStatus(v)
	if a.screen != nil {
		a.screen.SetTitle(a.title())
	}
}

func style(fg, bg tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(fg).Background(bg)
}

func (a *App) fill(x, y, w, h int, st tcell.Style) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			a.screen.SetContent(xx, yy, ' ', nil, st)
		}
	}
}

func (a *App) drawText(x, y, maxX int, s string, st tcell.Style) int {
	state := -1
	for s != "" && x < maxX {
		cluster, rest, width, next := uniseg.FirstGraphemeClusterInString(s, state)
		if width < 1 {
			width = 1
		}
		if x+width > maxX {
			break
		}
		r, _ := utf8.DecodeRuneInString(cluster)
		if r == utf8.RuneError && cluster != "\uFFFD" {
			r = '·'
		}
		a.screen.SetContent(x, y, r, nil, st)
		for i := 1; i < width; i++ {
			a.screen.SetContent(x+i, y, ' ', nil, st)
		}
		x += width
		s = rest
		state = next
	}
	return x
}

func (a *App) drawTabs(v view) {
	st := style(a.theme.TabFg, a.theme.TabBg)
	a.fill(0, 0, v.w, 1, st)
	if len(a.docs) == 0 {
		a.drawText(1, 0, v.w, "axam", style(a.theme.Accent, a.theme.TabBg))
		return
	}
	x := 0
	active := -1
	if d := a.doc(); d != nil {
		active = a.panes[a.active].Doc
	}
	for i, d := range a.docs {
		label := d.Name
		if d.Ed.Dirty() {
			label = "● " + label
		} else {
			label = "  " + label
		}
		label = " " + label + " "
		bg := a.theme.TabBg
		fg := a.theme.TabFg
		if i == active {
			bg = a.theme.TabActive
			fg = a.theme.Fg
		}
		width := uniseg.StringWidth(label)
		if x+width > v.w {
			break
		}
		a.drawText(x, 0, v.w, label, style(fg, bg))
		if i == active {
			a.screen.SetContent(x, 0, ' ', nil, style(a.theme.Accent, bg))
		}
		x += width
	}
}

func (a *App) drawExplorer(b cellBox) {
	st := style(a.theme.Explorer, a.theme.Bg)
	a.fill(b.x, b.y, b.w, b.h, st)
	top := 0
	if a.expl.sel >= b.h {
		top = a.expl.sel - b.h + 1
	}
	for i := 0; i < b.h; i++ {
		idx := top + i
		if idx >= len(a.expl.entries) {
			break
		}
		e := a.expl.entries[idx]
		bg := a.theme.Bg
		fg := a.theme.Explorer
		if idx == a.expl.sel && a.focus == focusExplorer {
			bg = a.theme.Sel
			fg = a.theme.Fg
		}
		mark := "  "
		if e.Dir {
			if e.Open {
				mark = "▾ "
			} else {
				mark = "▸ "
			}
		}
		indent := strings.Repeat(" ", e.Depth)
		if e.Depth > 4 {
			indent = strings.Repeat(" ", 4)
		}
		a.fill(b.x, b.y+i, b.w, 1, style(fg, bg))
		a.drawText(b.x, b.y+i, b.x+b.w, indent+mark+e.Name, style(fg, bg))
	}
}

func (a *App) drawPane(i int, b cellBox) {
	if i >= len(a.panes) || a.panes[i].Doc < 0 || a.panes[i].Doc >= len(a.docs) {
		a.drawWelcome(b)
		return
	}
	d := a.docs[a.panes[i].Doc]
	p := a.panes[i]
	gw := a.gutterWidth(d)
	cw := b.w - gw
	if cw < 1 {
		cw = 1
	}
	width := 0
	if a.cfg.WordWrap {
		width = cw
	}
	tab := d.Ed.TabSize
	active := i == a.active
	var bracket text.Pos
	var hasBracket bool
	if active {
		bracket, hasBracket = syntax.MatchBracket(d.Ed.Buf.Lines(), d.Ed.Primary().Head)
	}
	y := b.y
	line := p.Top
	rowSkip := p.Row
	cursorX, cursorY := -1, -1
	for line < d.Ed.Buf.LineCount() && y < b.y+b.h {
		if a.hidden(d, line) {
			line++
			continue
		}
		src := d.Ed.Buf.Line(line)
		rows := text.Rows(src, width, tab)
		spans := a.lineSpans(d, line, src)
		for ri := rowSkip; ri < len(rows) && y < b.y+b.h; ri++ {
			a.paintRow(d, b, gw, cw, line, rows[ri], src, spans, p.Left, y, active, ri == 0, hasBracket, bracket, &cursorX, &cursorY)
			y++
		}
		rowSkip = 0
		line++
	}
	st := style(a.theme.Fg, a.theme.Bg)
	for ; y < b.y+b.h; y++ {
		a.fill(b.x, y, b.w, 1, st)
		if a.cfg.LineNumbers {
			a.fill(b.x, y, gw, 1, style(a.theme.LineNum, a.theme.Gutter))
		}
	}
	if active && a.mode == modeEdit && cursorX >= 0 {
		a.screen.ShowCursor(cursorX, cursorY)
	}
}

func (a *App) lineSpans(d *Document, line int, src string) []syntax.Span {
	if d.Large {
		return nil
	}
	gen := d.Ed.Buf.Gen()
	if h, ok := d.hl[line]; ok && h.gen == gen {
		return h.spans
	}
	spans := syntax.HighlightLine(d.Lang, src)
	if d.hl == nil {
		d.hl = map[int]hlLine{}
	}
	if len(d.hl) > 4000 {
		d.hl = map[int]hlLine{}
	}
	d.hl[line] = hlLine{gen: gen, spans: spans}
	return spans
}

func (a *App) paintRow(d *Document, b cellBox, gw, cw, line int, row text.Row, src string, spans []syntax.Span, left, y int, active, gutter bool, hasBracket bool, bracket text.Pos, cursorX, cursorY *int) {
	cur := false
	if active {
		for _, c := range d.Ed.Cursors {
			if c.Head.Line == line {
				cur = true
				break
			}
		}
	}
	bg := a.theme.Bg
	if cur {
		bg = a.theme.Current
	}
	a.fill(b.x, y, b.w, 1, style(a.theme.Fg, bg))
	if gutter {
		gst := style(a.theme.LineNum, a.theme.Gutter)
		if cur {
			gst = style(a.theme.LineNumCur, a.theme.Gutter)
		}
		a.fill(b.x, y, gw, 1, gst)
		if a.cfg.LineNumbers {
			label := fmt.Sprintf("%*d", gw-1, line+1)
			a.drawText(b.x, y, b.x+gw-1, label, gst)
		}
		if end, ok := d.Folds[line]; ok && end > line {
			a.screen.SetContent(b.x+gw-1, y, '▾', nil, gst)
		}
	} else if gw > 0 {
		a.fill(b.x, y, gw, 1, style(a.theme.LineNum, a.theme.Gutter))
	}
	x := b.x + gw
	maxX := b.x + b.w
	vis := text.VisualCol(src, row.Start, d.Ed.TabSize)
	abs := vis
	state := -1
	frag := src[row.Start:row.End]
	byteAt := row.Start
	// Advance uniseg state to the row start so combining marks stay intact.
	if row.Start > 0 {
		state = -1
		pre := src[:row.Start]
		for pre != "" {
			_, rest, _, next := uniseg.FirstGraphemeClusterInString(pre, state)
			pre = rest
			state = next
		}
	}
	for frag != "" && x < maxX {
		cluster, rest, width, next := uniseg.FirstGraphemeClusterInString(frag, state)
		cw := width
		if cluster == "\t" {
			stop := d.Ed.TabSize
			if stop < 1 {
				stop = 4
			}
			cw = stop - (abs % stop)
			if cw < 1 {
				cw = stop
			}
		}
		if cw < 1 {
			cw = 1
		}
		if abs+cw <= left {
			abs += cw
			vis = abs
			byteAt += len(cluster)
			frag = rest
			state = next
			continue
		}
		fg := a.theme.Fg
		cellBg := bg
		if kind := spanKind(spans, byteAt); kind != "" {
			fg = a.kindColor(kind)
		}
		if a.inSelection(d, line, byteAt) {
			cellBg = a.theme.Sel
		}
		if a.inMatch(line, byteAt) {
			cellBg = a.theme.Match
		}
		if hasBracket && bracket.Line == line && bracket.Col == byteAt {
			fg = a.theme.Bracket
		}
		if a.cfg.ShowTrailingSpace && isTrailing(src, byteAt) && (cluster == " " || cluster == "\t") {
			fg = a.theme.Whitespace
		}
		st := style(fg, cellBg)
		if a.isPrimaryHead(d, line, byteAt, len(cluster)) && active && a.mode == modeEdit {
			*cursorX = x
			*cursorY = y
		} else if a.isAnyHead(d, line, byteAt, len(cluster)) && active {
			st = st.Reverse(true)
		}
		if abs < left {
			// A wide cluster straddles the scroll origin. Skip the clipped cells.
			skip := left - abs
			if cluster == "\t" {
				spaces := cw - skip
				for i := 0; i < spaces && x < maxX; i++ {
					a.screen.SetContent(x, y, ' ', nil, st)
					x++
				}
			}
			abs += cw
			byteAt += len(cluster)
			frag = rest
			state = next
			continue
		}
		if cluster == "\t" {
			for i := 0; i < cw && x < maxX; i++ {
				ch := ' '
				if a.cfg.ShowTrailingSpace && isTrailing(src, byteAt) && i == 0 {
					ch = '›'
				}
				a.screen.SetContent(x, y, ch, nil, st)
				x++
			}
		} else {
			r, _ := utf8.DecodeRuneInString(cluster)
			if r == utf8.RuneError && cluster != "\uFFFD" {
				r = '·'
				fg = a.theme.Error
				st = style(fg, cellBg)
			}
			if x < maxX {
				a.screen.SetContent(x, y, r, nil, st)
				x++
			}
			for i := 1; i < cw && x < maxX; i++ {
				a.screen.SetContent(x, y, ' ', nil, st)
				x++
			}
		}
		abs += cw
		byteAt += len(cluster)
		frag = rest
		state = next
	}
	if a.isPrimaryHead(d, line, row.End, 0) && active && a.mode == modeEdit && x <= maxX {
		if *cursorX < 0 {
			*cursorX = x
			*cursorY = y
			if x >= maxX && maxX > b.x {
				*cursorX = maxX - 1
			}
		}
	}
}

func spanKind(spans []syntax.Span, col int) string {
	for _, s := range spans {
		if col >= s.Start && col < s.End {
			return s.Kind
		}
	}
	return ""
}

func (a *App) kindColor(kind string) tcell.Color {
	switch kind {
	case "keyword":
		return a.theme.Keyword
	case "string":
		return a.theme.String
	case "comment":
		return a.theme.Comment
	case "number":
		return a.theme.Number
	case "func":
		return a.theme.Func
	case "type":
		return a.theme.TypeName
	default:
		return a.theme.Fg
	}
}

func (a *App) inSelection(d *Document, line, col int) bool {
	for _, c := range d.Ed.Cursors {
		if c.Empty() {
			continue
		}
		s, e := c.Span()
		if line < s.Line || line > e.Line {
			continue
		}
		if s.Line == e.Line {
			if col >= s.Col && col < e.Col {
				return true
			}
			continue
		}
		if line == s.Line && col >= s.Col {
			return true
		}
		if line == e.Line && col < e.Col {
			return true
		}
		if line > s.Line && line < e.Line {
			return true
		}
	}
	return false
}

func (a *App) inMatch(line, col int) bool {
	if a.find.query == "" {
		return false
	}
	for _, m := range a.find.matches {
		if line < m.Line || line > m.EndLine {
			continue
		}
		if m.Line == m.EndLine {
			if col >= m.Col && col < m.EndCol {
				return true
			}
			continue
		}
		if line == m.Line && col >= m.Col || line == m.EndLine && col < m.EndCol || (line > m.Line && line < m.EndLine) {
			return true
		}
	}
	return false
}

func (a *App) isPrimaryHead(d *Document, line, col, n int) bool {
	h := d.Ed.Primary().Head
	if h.Line != line {
		return false
	}
	if n == 0 {
		return h.Col == col
	}
	return h.Col >= col && h.Col < col+n
}

func (a *App) isAnyHead(d *Document, line, col, n int) bool {
	p := d.Ed.Primary().Head
	for _, c := range d.Ed.Cursors {
		if c.Head == p {
			continue
		}
		if c.Head.Line == line && c.Head.Col >= col && c.Head.Col < col+n {
			return true
		}
	}
	return false
}

func isTrailing(src string, col int) bool {
	for i := col; i < len(src); i++ {
		if src[i] != ' ' && src[i] != '\t' {
			return false
		}
	}
	return col < len(src)
}

func (a *App) drawWelcome(b cellBox) {
	st := style(a.theme.Fg, a.theme.Bg)
	a.fill(b.x, b.y, b.w, b.h, st)
	lines := []string{
		"axam",
		"TermStudio",
		"",
		a.chordLabel("file.open") + "  open",
		a.chordLabel("nav.quickOpen") + "  quick open",
		a.chordLabel("view.palette") + "  commands",
		a.chordLabel("file.quit") + "  quit",
	}
	y := b.y + max(1, b.h/2-len(lines)/2)
	for _, line := range lines {
		if y >= b.y+b.h {
			break
		}
		fg := a.theme.Fg
		if line == "axam" {
			fg = a.theme.Accent
		}
		x := b.x + max(1, (b.w-uniseg.StringWidth(line))/2)
		a.drawText(x, y, b.x+b.w, line, style(fg, a.theme.Bg))
		y++
	}
}

func (a *App) chordLabel(id string) string {
	cs := a.keys.ChordsFor(id)
	if len(cs) == 0 {
		return "—"
	}
	return cs[0]
}

func (a *App) drawFind(b cellBox) {
	bg := a.theme.PromptBg
	a.fill(b.x, b.y, b.w, b.h, style(a.theme.Fg, bg))
	flags := ""
	if a.find.sens {
		flags += " case"
	} else {
		flags += " ignore-case"
	}
	if a.find.word {
		flags += " word"
	}
	if a.find.regex {
		flags += " regex"
	}
	count := fmt.Sprintf("%d", len(a.find.matches))
	if a.find.err != "" {
		count = a.find.err
	} else if len(a.find.matches) > 0 {
		count = fmt.Sprintf("%d/%d", a.find.at+1, len(a.find.matches))
	}
	q := a.find.query
	mark := " "
	if a.find.field == 0 {
		mark = "▎"
	}
	a.drawText(b.x, b.y, b.x+b.w, " Find "+mark+q+"   "+count+flags, style(a.theme.Fg, bg))
	if a.find.replace && b.h > 1 {
		mark = " "
		if a.find.field == 1 {
			mark = "▎"
		}
		a.drawText(b.x, b.y+1, b.x+b.w, " Replace "+mark+a.find.repl, style(a.theme.Fg, bg))
	}
	hintY := b.y + b.h - 1
	a.drawText(b.x, hintY, b.x+b.w, " Enter next   Alt+Enter replace   Alt+C/W/R toggles   Esc close", style(a.theme.LineNum, bg))
	if a.mode == modeFind {
		field := a.find.query
		label := " Find "
		if a.find.field == 1 {
			field = a.find.repl
			label = " Replace "
		}
		prefix := uniseg.StringWidth(label) + 1
		caret := prefix + uniseg.StringWidth(field[:min(a.find.pos, len(field))])
		row := b.y
		if a.find.field == 1 {
			row = b.y + 1
		}
		if caret >= b.w {
			caret = b.w - 1
		}
		a.screen.ShowCursor(b.x+caret, row)
	}
}

func (a *App) drawPrompt(b cellBox) {
	bg := a.theme.PromptBg
	fg := a.theme.Fg
	if a.prompt.kind == "confirm" {
		fg = a.theme.Dirty
	}
	a.fill(b.x, b.y, b.w, b.h, style(fg, bg))
	msg := a.prompt.title
	if a.prompt.kind != "confirm" {
		msg += " " + a.prompt.text
	} else {
		msg += "  y / n"
	}
	a.drawText(b.x+1, b.y, b.x+b.w, msg, style(fg, bg))
	if a.prompt.kind != "confirm" {
		prefix := uniseg.StringWidth(a.prompt.title + " ")
		caret := 1 + prefix + uniseg.StringWidth(a.prompt.text[:min(a.prompt.pos, len(a.prompt.text))])
		if caret >= b.w {
			caret = b.w - 1
		}
		a.screen.ShowCursor(b.x+caret, b.y)
	}
}

func (a *App) drawList(b cellBox) {
	bg := a.theme.PromptBg
	a.fill(b.x, b.y, b.w, b.h, style(a.theme.Fg, bg))
	title := a.list.title + "  " + a.list.query
	a.drawText(b.x+1, b.y, b.x+b.w, title, style(a.theme.Accent, bg))
	prefix := uniseg.StringWidth(a.list.title + "  ")
	caret := 1 + prefix + uniseg.StringWidth(a.list.query[:min(a.list.pos, len(a.list.query))])
	if caret < b.w {
		a.screen.ShowCursor(b.x+caret, b.y)
	}
	top := 0
	if a.list.sel >= b.h-1 {
		top = a.list.sel - (b.h - 2)
	}
	for i := 0; i < b.h-1; i++ {
		idx := top + i
		if idx >= len(a.list.items) {
			break
		}
		it := a.list.items[idx]
		fg := a.theme.Fg
		rowBg := bg
		if idx == a.list.sel {
			rowBg = a.theme.PromptSel
		}
		a.fill(b.x, b.y+1+i, b.w, 1, style(fg, rowBg))
		label := it.Title
		if it.Shortcut != "" {
			label += "    " + it.Shortcut
		}
		if it.Detail != "" && it.Shortcut == "" {
			label += "    " + it.Detail
		}
		a.drawText(b.x+1, b.y+1+i, b.x+b.w, label, style(fg, rowBg))
	}
}

func (a *App) drawStatus(v view) {
	bg := a.theme.StatusBg
	fg := a.theme.StatusFg
	a.fill(0, v.status, v.w, 1, style(fg, bg))
	d := a.doc()
	left := " axam"
	right := a.info.ID
	if d != nil {
		name := d.Name
		if d.Ed.Dirty() {
			name += " ●"
		}
		if d.ReadOnly {
			name += " [read-only]"
		}
		c := d.Ed.Primary().Head
		vis := text.VisualCol(d.Ed.Buf.Line(c.Line), c.Col, d.Ed.TabSize) + 1
		left = fmt.Sprintf(" %s   %d:%d", name, c.Line+1, vis)
		eol := d.Ed.Buf.EOL().String()
		right = fmt.Sprintf("%s  %s  %s  %s ", d.Lang, d.Encoding, eol, a.info.ID)
		if len(d.Ed.Cursors) > 1 {
			right = fmt.Sprintf("%d cursors  ", len(d.Ed.Cursors)) + right
		}
	}
	if a.status != "" {
		left += "   " + a.status
	}
	if len(a.chord) > 0 {
		left += "   " + inputChord(a)
	}
	a.drawText(0, v.status, v.w, left, style(fg, bg))
	rw := uniseg.StringWidth(right)
	if rw < v.w {
		a.drawText(v.w-rw, v.status, v.w, right, style(a.theme.Accent, bg))
	}
}

func inputChord(a *App) string {
	return a.chordText()
}

func (a *App) chordText() string {
	parts := make([]string, len(a.chord))
	for i, k := range a.chord {
		parts[i] = k.Display()
	}
	return strings.Join(parts, " ")
}

func (a *App) hitTest(x, y int) (hit, bool) {
	v := a.view()
	if h, ok := a.menuHit(x, y, v); ok {
		return h, true
	}
	if y == 0 {
		return a.tabHit(x)
	}
	if v.list.h > 0 && y >= v.list.y && y < v.list.y+v.list.h {
		return hit{}, false
	}
	if v.expl.w > 0 && x >= v.expl.x && x < v.expl.x+v.expl.w && y >= v.expl.y && y < v.expl.y+v.expl.h {
		top := 0
		if a.expl.sel >= v.expl.h {
			top = a.expl.sel - v.expl.h + 1
		}
		idx := top + (y - v.expl.y)
		if idx < 0 || idx >= len(a.expl.entries) {
			return hit{kind: hitExpl, expl: a.expl.sel}, true
		}
		return hit{kind: hitExpl, expl: idx}, true
	}
	for i, b := range v.edits {
		if x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h {
			if i >= len(a.panes) || a.panes[i].Doc < 0 {
				return hit{}, false
			}
			d := a.docs[a.panes[i].Doc]
			gw := a.gutterWidth(d)
			kind := hitText
			if x < b.x+gw {
				kind = hitGutter
			}
			return hit{kind: kind, pane: i}, true
		}
	}
	return hit{}, false
}

func (a *App) tabHit(x int) (hit, bool) {
	pos := 0
	for i, d := range a.docs {
		label := d.Name
		if d.Ed.Dirty() {
			label = "● " + label
		} else {
			label = "  " + label
		}
		label = " " + label + " "
		w := uniseg.StringWidth(label)
		if x >= pos && x < pos+w {
			return hit{kind: hitTab, tab: i}, true
		}
		pos += w
	}
	return hit{}, false
}

func (a *App) pointAt(pane, x, y int) text.Pos {
	v := a.view()
	if pane < 0 || pane >= len(v.edits) || pane >= len(a.panes) {
		return text.Pos{}
	}
	d := a.docs[a.panes[pane].Doc]
	b := v.edits[pane]
	gw := a.gutterWidth(d)
	cw := b.w - gw
	if cw < 1 {
		cw = 1
	}
	p := a.panes[pane]
	width := 0
	if a.cfg.WordWrap {
		width = cw
	}
	targetRow := y - b.y
	line := p.Top
	rowSkip := p.Row
	for line < d.Ed.Buf.LineCount() {
		if a.hidden(d, line) {
			line++
			continue
		}
		rows := text.Rows(d.Ed.Buf.Line(line), width, d.Ed.TabSize)
		for ri := rowSkip; ri < len(rows); ri++ {
			if targetRow == 0 {
				screenCol := x - (b.x + gw) + p.Left
				if screenCol < 0 {
					screenCol = 0
				}
				col := text.ColAtRow(d.Ed.Buf.Line(line), rows[ri], screenCol, d.Ed.TabSize)
				return text.Pos{Line: line, Col: col}
			}
			targetRow--
		}
		rowSkip = 0
		line++
	}
	last := d.Ed.Buf.LineCount() - 1
	if last < 0 {
		return text.Pos{}
	}
	return text.Pos{Line: last, Col: len(d.Ed.Buf.Line(last))}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
