// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package editor implements editing operations independent of the terminal.
//
// A cursor is an anchor plus a head. Shift-movement changes the head only.
// The anchor stays put until a non-selecting command collapses it. Edits
// record inverse operations so undo restores both text and carets.
package editor

import (
	"unicode"
	"unicode/utf8"

	"github.com/alexshescp/axam/internal/buffer"
	"github.com/alexshescp/axam/internal/text"
)

// Cursor is one selection. An empty selection has Anchor == Head.
// Desired is the visual column vertical movement tries to keep.
// Negative means "recompute from the caret".
type Cursor struct {
	Anchor  text.Pos
	Head    text.Pos
	Desired int
}

// Empty reports whether the selection contains no characters.
func (c Cursor) Empty() bool { return c.Anchor.Eq(c.Head) }

// Span returns the half-open selected range.
func (c Cursor) Span() (text.Pos, text.Pos) { return text.Order(c.Anchor, c.Head) }

// Editor owns a buffer, its cursors, and its undo history.
type Editor struct {
	Buf        *buffer.Buffer
	Cursors    []Cursor
	primary    int
	TabSize    int
	UseSpaces  bool
	ReadOnly   bool
	hist       stack
	seal       bool
	lastInsert bool
}

// New returns an editor on buf. A nil buffer becomes an empty document.
func New(buf *buffer.Buffer) *Editor {
	if buf == nil {
		buf = buffer.New()
	}
	return &Editor{
		Buf:       buf,
		Cursors:   []Cursor{{Desired: -1}},
		TabSize:   4,
		UseSpaces: true,
	}
}

// Primary is the cursor that the hardware caret follows.
func (e *Editor) Primary() Cursor {
	e.fixPrimary()
	return e.Cursors[e.primary]
}

// SetPrimary replaces the sole cursor.
func (e *Editor) SetPrimary(c Cursor) {
	c.Anchor = e.Buf.Clamp(c.Anchor)
	c.Head = e.Buf.Clamp(c.Head)
	if c.Desired == 0 && c.Head.Col == 0 && c.Head.Line == 0 {
		c.Desired = -1
	}
	e.Cursors = []Cursor{c}
	e.primary = 0
	e.seal = true
}

// PrimaryIndex is the index of the hardware caret.
func (e *Editor) PrimaryIndex() int {
	e.fixPrimary()
	return e.primary
}

func (e *Editor) fixPrimary() {
	if len(e.Cursors) == 0 {
		e.Cursors = []Cursor{{Desired: -1}}
	}
	if e.primary < 0 || e.primary >= len(e.Cursors) {
		e.primary = len(e.Cursors) - 1
	}
}

// SelectedText is the primary selection.
func (e *Editor) SelectedText() string {
	c := e.Primary()
	a, b := c.Span()
	return e.Buf.Text(a, b)
}

// CanUndo reports whether an undo step exists.
func (e *Editor) CanUndo() bool { return len(e.hist.undo) > 0 }

// CanRedo reports whether a redo step exists.
func (e *Editor) CanRedo() bool { return len(e.hist.redo) > 0 }

// Dirty reports edits since the last MarkClean.
func (e *Editor) Dirty() bool { return e.hist.dirty() }

// MarkClean remembers the current history position as saved.
func (e *Editor) MarkClean() { e.hist.markClean() }

func (e *Editor) snapshot() []Cursor {
	out := make([]Cursor, len(e.Cursors))
	copy(out, e.Cursors)
	return out
}

func (e *Editor) restore(cs []Cursor) {
	if len(cs) == 0 {
		cs = []Cursor{{Desired: -1}}
	}
	e.Cursors = make([]Cursor, len(cs))
	copy(e.Cursors, cs)
	e.primary = len(e.Cursors) - 1
}

// begin starts a transaction. Mutations call commit.
func (e *Editor) begin() *tx {
	e.fixPrimary()
	return &tx{before: e.snapshot()}
}

func (e *Editor) commit(t *tx, coalesce bool) {
	if len(t.edits) == 0 {
		return
	}
	t.after = e.snapshot()
	e.hist.push(*t, coalesce && !e.seal)
	e.seal = false
	e.lastInsert = coalesce
}

// Seal stops the next edit from merging into the previous typing undo.
func (e *Editor) Seal() { e.seal = true }

// Insert replaces each selection with s.
// coalesce merges consecutive single-cursor plain typing into one undo step.
func (e *Editor) Insert(s string, coalesce bool) bool {
	if e.ReadOnly {
		return false
	}
	if s == "" && !e.hasSelection() {
		return false
	}
	s = text.NormalizeNL(s)
	e.merge()
	t := e.begin()
	order := e.orderDesc()
	done := map[int]bool{}
	for _, i := range order {
		c := e.Cursors[i]
		start, end := c.Span()
		deleted := ""
		if start.Less(end) {
			deleted = e.Buf.Delete(start, end)
		}
		endAfter := start
		if s != "" {
			endAfter = e.Buf.Insert(start, s)
		}
		t.edits = append(t.edits, edit{start: start, deleted: deleted, inserted: s})
		e.Cursors[i].Anchor = endAfter
		e.Cursors[i].Head = endAfter
		e.Cursors[i].Desired = -1
		for j := range e.Cursors {
			if j == i || !done[j] {
				continue
			}
			e.Cursors[j].Anchor = shiftPoint(e.Cursors[j].Anchor, start, end, endAfter)
			e.Cursors[j].Head = shiftPoint(e.Cursors[j].Head, start, end, endAfter)
		}
		done[i] = true
	}
	plain := coalesce && len(t.edits) == 1 && !stringsContainsNL(s) && t.edits[0].deleted == ""
	e.commit(t, plain)
	e.merge()
	return true
}

func stringsContainsNL(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}

// Newline inserts a line break and copies indentation.
// A '{' immediately before the caret adds one indent level.
func (e *Editor) Newline() bool {
	if e.ReadOnly {
		return false
	}
	e.merge()
	t := e.begin()
	order := e.orderDesc()
	done := map[int]bool{}
	for _, i := range order {
		c := e.Cursors[i]
		start, end := c.Span()
		deleted := ""
		if start.Less(end) {
			deleted = e.Buf.Delete(start, end)
		}
		line := e.Buf.Line(start.Line)
		indent := leadingWS(line)
		extra := ""
		before := line[:start.Col]
		trim := stringsTrimRightSpace(before)
		if stringsHasSuffix(trim, "{") {
			extra = e.indentUnit()
		}
		ins := "\n" + indent + extra
		endAfter := e.Buf.Insert(start, ins)
		t.edits = append(t.edits, edit{start: start, deleted: deleted, inserted: ins})
		e.Cursors[i].Anchor = endAfter
		e.Cursors[i].Head = endAfter
		e.Cursors[i].Desired = -1
		for j := range e.Cursors {
			if j == i || !done[j] {
				continue
			}
			e.Cursors[j].Anchor = shiftPoint(e.Cursors[j].Anchor, start, end, endAfter)
			e.Cursors[j].Head = shiftPoint(e.Cursors[j].Head, start, end, endAfter)
		}
		done[i] = true
	}
	e.commit(t, false)
	e.merge()
	return true
}

func leadingWS(s string) string {
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !unicode.IsSpace(r) || r == '\n' {
			break
		}
		i += size
	}
	return s[:i]
}

func stringsTrimRightSpace(s string) string {
	i := len(s)
	for i > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		if !unicode.IsSpace(r) {
			break
		}
		i -= size
	}
	return s[:i]
}

func stringsHasSuffix(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}

func (e *Editor) indentUnit() string {
	if !e.UseSpaces {
		return "\t"
	}
	n := e.TabSize
	if n < 1 {
		n = 4
	}
	return stringsRepeat(" ", n)
}

func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

// Backspace deletes the selection or the previous grapheme.
func (e *Editor) Backspace() bool { return e.deleteBackward(false) }

// DeleteForward deletes the selection or the next grapheme.
func (e *Editor) DeleteForward() bool { return e.deleteForward(false) }

// DeleteWordLeft deletes back to the previous word start.
func (e *Editor) DeleteWordLeft() bool { return e.deleteBackward(true) }

// DeleteWordRight deletes through the next word end.
func (e *Editor) DeleteWordRight() bool { return e.deleteForward(true) }

func (e *Editor) deleteBackward(word bool) bool {
	if e.ReadOnly {
		return false
	}
	e.prepareDeleteRanges(func(c Cursor) (text.Pos, text.Pos) {
		if !c.Empty() {
			return c.Span()
		}
		end := c.Head
		var start text.Pos
		if word {
			start = e.wordLeftPos(end)
		} else {
			start = e.graphemeLeft(end)
		}
		return start, end
	})
	return true
}

func (e *Editor) deleteForward(word bool) bool {
	if e.ReadOnly {
		return false
	}
	e.prepareDeleteRanges(func(c Cursor) (text.Pos, text.Pos) {
		if !c.Empty() {
			return c.Span()
		}
		start := c.Head
		var end text.Pos
		if word {
			end = e.wordRightPos(start)
		} else {
			end = e.graphemeRight(start)
		}
		return start, end
	})
	return true
}

func (e *Editor) prepareDeleteRanges(fn func(Cursor) (text.Pos, text.Pos)) {
	e.merge()
	t := e.begin()
	order := e.orderDesc()
	done := map[int]bool{}
	for _, i := range order {
		start, end := fn(e.Cursors[i])
		start, end = text.Order(start, end)
		start = e.Buf.Clamp(start)
		end = e.Buf.Clamp(end)
		deleted := ""
		if start.Less(end) {
			deleted = e.Buf.Delete(start, end)
		}
		t.edits = append(t.edits, edit{start: start, deleted: deleted, inserted: ""})
		e.Cursors[i].Anchor = start
		e.Cursors[i].Head = start
		e.Cursors[i].Desired = -1
		for j := range e.Cursors {
			if j == i || !done[j] {
				continue
			}
			e.Cursors[j].Anchor = shiftPoint(e.Cursors[j].Anchor, start, end, start)
			e.Cursors[j].Head = shiftPoint(e.Cursors[j].Head, start, end, start)
		}
		done[i] = true
	}
	e.commit(t, false)
	e.merge()
}

// ReplaceRange replaces one range and collapses cursors to the end.
// It is the primitive used by search-and-replace.
func (e *Editor) ReplaceRange(start, end text.Pos, s string) bool {
	if e.ReadOnly {
		return false
	}
	start, end = text.Order(e.Buf.Clamp(start), e.Buf.Clamp(end))
	s = text.NormalizeNL(s)
	t := e.begin()
	deleted := ""
	if start.Less(end) {
		deleted = e.Buf.Delete(start, end)
	}
	endAfter := e.Buf.Insert(start, s)
	t.edits = append(t.edits, edit{start: start, deleted: deleted, inserted: s})
	e.Cursors = []Cursor{{Anchor: endAfter, Head: endAfter, Desired: -1}}
	e.primary = 0
	e.commit(t, false)
	return true
}

// ReplaceRanges applies every range in one undo step.
// Ranges may overlap the replacement text; they are applied from the end
// of the document backward so earlier coordinates stay valid.
func (e *Editor) ReplaceRanges(ranges [][2]text.Pos, s string) int {
	if e.ReadOnly || len(ranges) == 0 {
		return 0
	}
	s = text.NormalizeNL(s)
	order := make([]int, len(ranges))
	for i := range order {
		order[i] = i
	}
	for i := 1; i < len(order); i++ {
		for j := i; j > 0; j-- {
			cur, _ := text.Order(ranges[order[j]][0], ranges[order[j]][1])
			prev, _ := text.Order(ranges[order[j-1]][0], ranges[order[j-1]][1])
			if !prev.Less(cur) {
				break
			}
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	t := e.begin()
	var endAfter text.Pos
	n := 0
	for _, idx := range order {
		start, end := text.Order(e.Buf.Clamp(ranges[idx][0]), e.Buf.Clamp(ranges[idx][1]))
		deleted := ""
		if start.Less(end) {
			deleted = e.Buf.Delete(start, end)
		}
		endAfter = e.Buf.Insert(start, s)
		t.edits = append(t.edits, edit{start: start, deleted: deleted, inserted: s})
		n++
	}
	e.Cursors = []Cursor{{Anchor: endAfter, Head: endAfter, Desired: -1}}
	e.primary = 0
	e.commit(t, false)
	return n
}

// Undo reverts one transaction.
func (e *Editor) Undo() bool {
	t, ok := e.hist.popUndo()
	if !ok {
		return false
	}
	for i := len(t.edits) - 1; i >= 0; i-- {
		ed := t.edits[i]
		end := text.Advance(ed.start, ed.inserted)
		if ed.inserted != "" {
			e.Buf.Delete(ed.start, e.Buf.Clamp(end))
		}
		if ed.deleted != "" {
			e.Buf.Insert(ed.start, ed.deleted)
		}
	}
	e.restore(t.before)
	e.seal = true
	return true
}

// Redo reapplies one transaction.
func (e *Editor) Redo() bool {
	t, ok := e.hist.popRedo()
	if !ok {
		return false
	}
	for _, ed := range t.edits {
		end := text.Advance(ed.start, ed.deleted)
		if ed.deleted != "" {
			e.Buf.Delete(ed.start, e.Buf.Clamp(end))
		}
		if ed.inserted != "" {
			e.Buf.Insert(ed.start, ed.inserted)
		}
	}
	e.restore(t.after)
	e.seal = true
	return true
}

func (e *Editor) hasSelection() bool {
	for _, c := range e.Cursors {
		if !c.Empty() {
			return true
		}
	}
	return false
}

func (e *Editor) orderDesc() []int {
	idx := make([]int, len(e.Cursors))
	for i := range idx {
		idx[i] = i
	}
	for i := 1; i < len(idx); i++ {
		j := i
		for j > 0 && startOf(e.Cursors[idx[j]]).Less(startOf(e.Cursors[idx[j-1]])) == false && !startOf(e.Cursors[idx[j]]).Eq(startOf(e.Cursors[idx[j-1]])) {
			idx[j], idx[j-1] = idx[j-1], idx[j]
			j--
		}
		// The condition above is awkward. Use a direct sort below if equal handling fails.
	}
	// Reliable sort.
	for i := 0; i < len(idx); i++ {
		for j := i + 1; j < len(idx); j++ {
			si := startOf(e.Cursors[idx[i]])
			sj := startOf(e.Cursors[idx[j]])
			if si.Less(sj) || (si.Eq(sj) && idx[i] < idx[j]) {
				idx[i], idx[j] = idx[j], idx[i]
			}
		}
	}
	return idx
}

func startOf(c Cursor) text.Pos {
	a, _ := c.Span()
	return a
}

func shiftPoint(p, start, end, endAfter text.Pos) text.Pos {
	if p.Less(start) {
		return p
	}
	if p.Less(end) && !p.Eq(end) {
		return endAfter
	}
	// p >= end
	if p.Line == end.Line {
		delta := p.Col - end.Col
		return text.Pos{Line: endAfter.Line, Col: endAfter.Col + delta}
	}
	return text.Pos{Line: p.Line + (endAfter.Line - end.Line), Col: p.Col}
}

func (e *Editor) merge() {
	e.fixPrimary()
	if len(e.Cursors) <= 1 {
		e.primary = 0
		return
	}
	oldPrimary := e.primary
	type item struct {
		c     Cursor
		index int
	}
	items := make([]item, len(e.Cursors))
	for i, c := range e.Cursors {
		items[i] = item{c: c, index: i}
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			si, _ := items[i].c.Span()
			sj, _ := items[j].c.Span()
			if sj.Less(si) || (sj.Eq(si) && items[j].index < items[i].index) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	out := []item{items[0]}
	for _, it := range items[1:] {
		last := &out[len(out)-1]
		if rangesOverlap(last.c, it.c) {
			a1, b1 := last.c.Span()
			a2, b2 := it.c.Span()
			start := a1
			if a2.Less(start) {
				start = a2
			}
			end := b1
			if end.Less(b2) {
				end = b2
			}
			keep := last.c
			idx := last.index
			if it.index == oldPrimary || (last.index != oldPrimary && it.index > last.index) {
				keep = it.c
				idx = it.index
			}
			keep.Anchor, keep.Head = start, end
			last.c = keep
			last.index = idx
			continue
		}
		out = append(out, it)
	}
	primary := len(out) - 1
	for i, it := range out {
		if it.index == oldPrimary {
			primary = i
			break
		}
	}
	e.Cursors = make([]Cursor, len(out))
	for i, it := range out {
		e.Cursors[i] = it.c
	}
	if primary != len(e.Cursors)-1 {
		e.Cursors[primary], e.Cursors[len(e.Cursors)-1] = e.Cursors[len(e.Cursors)-1], e.Cursors[primary]
	}
	e.primary = len(e.Cursors) - 1
}

func rangesOverlap(a, b Cursor) bool {
	a0, a1 := a.Span()
	b0, b1 := b.Span()
	if a.Empty() && b.Empty() {
		return a0.Eq(b0)
	}
	if a.Empty() {
		return !a0.Less(b0) && a0.Less(b1) || a0.Eq(b0)
	}
	if b.Empty() {
		return !b0.Less(a0) && b0.Less(a1) || b0.Eq(a0)
	}
	return a0.Less(b1) && b0.Less(a1)
}
