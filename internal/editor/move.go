// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package editor

import (
	"github.com/alexshescp/axam/internal/text"
)

// width is the wrapped content width in cells. Zero disables wrapping.

func (e *Editor) Left(extend bool)  { e.horizontal(extend, true) }
func (e *Editor) Right(extend bool) { e.horizontal(extend, false) }

func (e *Editor) horizontal(extend, left bool) {
	e.Seal()
	e.fixPrimary()
	for i := range e.Cursors {
		c := &e.Cursors[i]
		if !extend && !c.Empty() {
			start, end := c.Span()
			dest := end
			if left {
				dest = start
			}
			c.Anchor, c.Head = dest, dest
			c.Desired = -1
			continue
		}
		var p text.Pos
		if left {
			p = e.graphemeLeft(c.Head)
		} else {
			p = e.graphemeRight(c.Head)
		}
		c.Head = p
		if !extend {
			c.Anchor = p
		}
		c.Desired = -1
	}
	e.merge()
}

func (e *Editor) graphemeLeft(p text.Pos) text.Pos {
	p = e.Buf.Clamp(p)
	if p.Col > 0 {
		p.Col = text.Prev(e.Buf.Line(p.Line), p.Col)
		return p
	}
	if p.Line == 0 {
		return p
	}
	p.Line--
	p.Col = len(e.Buf.Line(p.Line))
	return p
}

func (e *Editor) graphemeRight(p text.Pos) text.Pos {
	p = e.Buf.Clamp(p)
	line := e.Buf.Line(p.Line)
	if p.Col < len(line) {
		p.Col = text.Next(line, p.Col)
		return p
	}
	if p.Line+1 >= e.Buf.LineCount() {
		return p
	}
	return text.Pos{Line: p.Line + 1, Col: 0}
}

// Up moves by one visual row when width > 0, otherwise by one logical line.
func (e *Editor) Up(extend bool, width int)   { e.vertical(extend, -1, width) }
func (e *Editor) Down(extend bool, width int) { e.vertical(extend, 1, width) }

func (e *Editor) vertical(extend bool, dir, width int) {
	e.Seal()
	for i := range e.Cursors {
		c := &e.Cursors[i]
		if width > 0 {
			e.visualVertical(c, dir, width)
		} else {
			e.logicalVertical(c, dir)
		}
		if !extend {
			c.Anchor = c.Head
		}
	}
	e.merge()
}

func (e *Editor) logicalVertical(c *Cursor, dir int) {
	line := e.Buf.Line(c.Head.Line)
	if c.Desired < 0 {
		c.Desired = text.VisualCol(line, c.Head.Col, e.tab())
	}
	nl := c.Head.Line + dir
	if nl < 0 {
		nl = 0
	}
	if nl >= e.Buf.LineCount() {
		nl = e.Buf.LineCount() - 1
	}
	c.Head.Line = nl
	c.Head.Col = text.ColAtVisual(e.Buf.Line(nl), c.Desired, e.tab())
}

func (e *Editor) visualVertical(c *Cursor, dir, width int) {
	line := e.Buf.Line(c.Head.Line)
	rows := text.Rows(line, width, e.tab())
	ri := text.RowIndex(rows, c.Head.Col)
	if c.Desired < 0 {
		c.Desired = text.RowScreenCol(line, rows[ri], c.Head.Col, e.tab())
	}
	if dir < 0 {
		if ri > 0 {
			c.Head.Col = text.ColAtRow(line, rows[ri-1], c.Desired, e.tab())
			return
		}
		if c.Head.Line == 0 {
			c.Head.Col = 0
			return
		}
		c.Head.Line--
		pl := e.Buf.Line(c.Head.Line)
		prows := text.Rows(pl, width, e.tab())
		c.Head.Col = text.ColAtRow(pl, prows[len(prows)-1], c.Desired, e.tab())
		return
	}
	if ri+1 < len(rows) {
		c.Head.Col = text.ColAtRow(line, rows[ri+1], c.Desired, e.tab())
		return
	}
	if c.Head.Line+1 >= e.Buf.LineCount() {
		c.Head.Col = len(line)
		return
	}
	c.Head.Line++
	nl := e.Buf.Line(c.Head.Line)
	nrows := text.Rows(nl, width, e.tab())
	c.Head.Col = text.ColAtRow(nl, nrows[0], c.Desired, e.tab())
}

// Page moves by rows logical lines (the viewport height).
func (e *Editor) Page(extend bool, dir, rows, width int) {
	if rows < 1 {
		rows = 1
	}
	for i := 0; i < rows; i++ {
		e.vertical(extend, dir, width)
	}
}

// Home moves to the first non-space, or to column 0 if already there.
func (e *Editor) Home(extend bool) {
	e.Seal()
	for i := range e.Cursors {
		c := &e.Cursors[i]
		line := e.Buf.Line(c.Head.Line)
		ns := text.FirstNonSpace(line)
		target := ns
		if c.Head.Col == ns {
			target = 0
		}
		c.Head.Col = target
		if !extend {
			c.Anchor = c.Head
		}
		c.Desired = -1
	}
	e.merge()
}

// End moves to the end of the logical line.
func (e *Editor) End(extend bool) {
	e.Seal()
	for i := range e.Cursors {
		c := &e.Cursors[i]
		c.Head.Col = len(e.Buf.Line(c.Head.Line))
		if !extend {
			c.Anchor = c.Head
		}
		c.Desired = -1
	}
	e.merge()
}

// WordLeft moves to the start of the previous word group.
func (e *Editor) WordLeft(extend bool) { e.word(extend, true) }

// WordRight moves to the end of the next word group.
func (e *Editor) WordRight(extend bool) { e.word(extend, false) }

func (e *Editor) word(extend, left bool) {
	e.Seal()
	for i := range e.Cursors {
		c := &e.Cursors[i]
		var p text.Pos
		if left {
			p = e.wordLeftPos(c.Head)
		} else {
			p = e.wordRightPos(c.Head)
		}
		c.Head = p
		if !extend {
			c.Anchor = p
		}
		c.Desired = -1
	}
	e.merge()
}

func (e *Editor) wordLeftPos(p text.Pos) text.Pos {
	p = e.Buf.Clamp(p)
	if p.Line == 0 && p.Col == 0 {
		return p
	}
	p = e.graphemeLeft(p)
	for !(p.Line == 0 && p.Col == 0) && e.classAt(p) == text.ClassSpace {
		n := e.graphemeLeft(p)
		if n.Eq(p) {
			break
		}
		p = n
	}
	cls := e.classAt(p)
	for {
		if p.Line == 0 && p.Col == 0 {
			break
		}
		prev := e.graphemeLeft(p)
		if prev.Eq(p) || e.classAt(prev) != cls {
			break
		}
		p = prev
	}
	return p
}

func (e *Editor) wordRightPos(p text.Pos) text.Pos {
	p = e.Buf.Clamp(p)
	last := e.docEnd()
	if p.Eq(last) {
		return p
	}
	cls := e.classAt(p)
	if cls == text.ClassSpace {
		p = e.skipClass(p, text.ClassSpace)
		if p.Eq(last) {
			return p
		}
		cls = e.classAt(p)
	}
	return e.skipClass(p, cls)
}

func (e *Editor) skipClass(p text.Pos, cls text.Class) text.Pos {
	last := e.docEnd()
	for !p.Eq(last) && e.classAt(p) == cls {
		n := e.graphemeRight(p)
		if n.Eq(p) {
			break
		}
		p = n
	}
	return p
}

func (e *Editor) classAt(p text.Pos) text.Class {
	if p.Line < 0 || p.Line >= e.Buf.LineCount() {
		return text.ClassSpace
	}
	return text.ClassAt(e.Buf.Line(p.Line), p.Col)
}

// DocStart moves to the beginning of the buffer.
func (e *Editor) DocStart(extend bool) {
	e.Seal()
	for i := range e.Cursors {
		e.Cursors[i].Head = text.Pos{}
		if !extend {
			e.Cursors[i].Anchor = e.Cursors[i].Head
		}
		e.Cursors[i].Desired = -1
	}
	e.merge()
}

// DocEnd moves to the end of the buffer.
func (e *Editor) DocEnd(extend bool) {
	e.Seal()
	end := e.docEnd()
	for i := range e.Cursors {
		e.Cursors[i].Head = end
		if !extend {
			e.Cursors[i].Anchor = end
		}
		e.Cursors[i].Desired = -1
	}
	e.merge()
}

func (e *Editor) docEnd() text.Pos {
	n := e.Buf.LineCount()
	return text.Pos{Line: n - 1, Col: len(e.Buf.Line(n - 1))}
}

// SelectAll selects the whole buffer with one cursor.
func (e *Editor) SelectAll() {
	e.Seal()
	end := e.docEnd()
	e.Cursors = []Cursor{{Anchor: text.Pos{}, Head: end, Desired: -1}}
	e.primary = 0
}

// SelectWord selects the word or punctuation run under the primary caret.
func (e *Editor) SelectWord() {
	e.Seal()
	e.fixPrimary()
	c := e.Cursors[e.primary]
	p := c.Head
	if p.Col > 0 && (p.Col >= len(e.Buf.Line(p.Line)) || text.ClassAt(e.Buf.Line(p.Line), p.Col) == text.ClassSpace) {
		left := e.graphemeLeft(p)
		if e.classAt(left) != text.ClassSpace {
			p = left
		}
	}
	cls := e.classAt(p)
	start := p
	for {
		if start.Line == 0 && start.Col == 0 {
			break
		}
		prev := e.graphemeLeft(start)
		if prev.Eq(start) || e.classAt(prev) != cls {
			break
		}
		start = prev
	}
	end := e.skipClass(p, cls)
	e.Cursors = []Cursor{{Anchor: start, Head: end, Desired: -1}}
	e.primary = 0
}

// SelectLine selects the primary line. Repeating it extends to the next line.
func (e *Editor) SelectLine() {
	e.Seal()
	e.fixPrimary()
	c := &e.Cursors[e.primary]
	startLine := c.Anchor.Line
	endLine := c.Head.Line
	if c.Head.Less(c.Anchor) {
		startLine, endLine = endLine, startLine
	}
	whole := c.Anchor.Col == 0 && (c.Head.Col == 0 && c.Head.Line == endLine && endLine != startLine || e.coversThrough(c, endLine))
	if c.Empty() || !whole {
		startLine = c.Head.Line
		endLine = c.Head.Line
	}
	next := endLine + 1
	if next >= e.Buf.LineCount() {
		e.Cursors[e.primary] = Cursor{
			Anchor:  text.Pos{Line: startLine, Col: 0},
			Head:    e.docEnd(),
			Desired: -1,
		}
		return
	}
	e.Cursors[e.primary] = Cursor{
		Anchor:  text.Pos{Line: startLine, Col: 0},
		Head:    text.Pos{Line: next, Col: 0},
		Desired: -1,
	}
}

func (e *Editor) coversThrough(c *Cursor, endLine int) bool {
	_, b := c.Span()
	return b.Line > endLine || (b.Line == endLine && b.Col == len(e.Buf.Line(endLine)))
}

// ClearSelection collapses every cursor to its head.
func (e *Editor) ClearSelection() {
	for i := range e.Cursors {
		e.Cursors[i].Anchor = e.Cursors[i].Head
	}
}

// SetCaret moves the primary caret. Extra cursors are dropped unless extend is set.
func (e *Editor) SetCaret(p text.Pos, extend bool) {
	e.Seal()
	p = e.Buf.Clamp(p)
	e.fixPrimary()
	if !extend {
		e.Cursors = []Cursor{{Head: p, Anchor: p, Desired: -1}}
		e.primary = 0
		return
	}
	e.Cursors[e.primary].Head = p
	e.Cursors[e.primary].Desired = -1
}

// AddCaret appends a caret without clearing the others.
func (e *Editor) AddCaret(p text.Pos) {
	e.Seal()
	p = e.Buf.Clamp(p)
	e.Cursors = append(e.Cursors, Cursor{Head: p, Anchor: p, Desired: -1})
	e.primary = len(e.Cursors) - 1
	e.merge()
}

// SetColumn builds one caret per line between anchor and head, using visual columns.
func (e *Editor) SetColumn(anchor, head text.Pos) {
	e.Seal()
	anchor = e.Buf.Clamp(anchor)
	head = e.Buf.Clamp(head)
	top, bot := anchor.Line, head.Line
	if bot < top {
		top, bot = bot, top
	}
	vAnchor := text.VisualCol(e.Buf.Line(anchor.Line), anchor.Col, e.tab())
	vHead := text.VisualCol(e.Buf.Line(head.Line), head.Col, e.tab())
	lo, hi := vAnchor, vHead
	if hi < lo {
		lo, hi = hi, lo
	}
	cs := make([]Cursor, 0, bot-top+1)
	for line := top; line <= bot; line++ {
		a := text.ColAtVisual(e.Buf.Line(line), lo, e.tab())
		b := text.ColAtVisual(e.Buf.Line(line), hi, e.tab())
		headCol, anchorCol := b, a
		if vHead < vAnchor {
			headCol, anchorCol = a, b
		}
		cs = append(cs, Cursor{
			Anchor:  text.Pos{Line: line, Col: anchorCol},
			Head:    text.Pos{Line: line, Col: headCol},
			Desired: vHead,
		})
	}
	if len(cs) == 0 {
		return
	}
	e.Cursors = cs
	e.primary = len(cs) - 1
}

// AddCursorVertical adds a caret on the next or previous line at the same visual column.
func (e *Editor) AddCursorVertical(dir int) {
	e.Seal()
	e.fixPrimary()
	c := e.Cursors[e.primary]
	vis := text.VisualCol(e.Buf.Line(c.Head.Line), c.Head.Col, e.tab())
	nl := c.Head.Line + dir
	if nl < 0 || nl >= e.Buf.LineCount() {
		return
	}
	col := text.ColAtVisual(e.Buf.Line(nl), vis, e.tab())
	e.Cursors = append(e.Cursors, Cursor{
		Anchor:  text.Pos{Line: nl, Col: col},
		Head:    text.Pos{Line: nl, Col: col},
		Desired: vis,
	})
	e.primary = len(e.Cursors) - 1
	e.merge()
}

// GoToLine places the caret at a 1-based line.
func (e *Editor) GoToLine(line int) {
	e.Seal()
	if line < 1 {
		line = 1
	}
	if line > e.Buf.LineCount() {
		line = e.Buf.LineCount()
	}
	p := text.Pos{Line: line - 1, Col: 0}
	e.Cursors = []Cursor{{Head: p, Anchor: p, Desired: -1}}
	e.primary = 0
}

// ClearDesired drops remembered vertical columns, used when wrap mode changes.
func (e *Editor) ClearDesired() {
	for i := range e.Cursors {
		e.Cursors[i].Desired = -1
	}
}

func (e *Editor) tab() int {
	if e.TabSize < 1 {
		return 4
	}
	return e.TabSize
}
