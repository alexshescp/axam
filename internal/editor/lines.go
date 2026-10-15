// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package editor

import (
	"strings"
	"unicode"

	"github.com/alexshescp/axam/internal/text"
)

// DeleteLine removes every line touched by a cursor.
func (e *Editor) DeleteLine() bool {
	return e.withLineRanges(func(first, last int) []string {
		return nil
	})
}

// DuplicateLine copies each touched line block directly below itself.
func (e *Editor) DuplicateLine() bool {
	return e.withLineRanges(func(first, last int) []string {
		block := append([]string{}, e.Buf.Lines()[first:last+1]...)
		return append(append([]string{}, block...), block...)
	})
}

// MoveLines swaps each touched block one line up or down.
func (e *Editor) MoveLines(dir int) bool {
	if dir == 0 || e.ReadOnly {
		return false
	}
	ranges := e.lineRanges()
	if dir < 0 {
		for _, r := range ranges {
			if r[0] == 0 {
				return false
			}
		}
	} else {
		last := e.Buf.LineCount() - 1
		for _, r := range ranges {
			if r[1] >= last {
				return false
			}
		}
	}
	t := e.begin()
	idxs := make([]int, len(ranges))
	for i := range idxs {
		idxs[i] = i
	}
	if dir > 0 {
		for i, j := 0, len(idxs)-1; i < j; i, j = i+1, j-1 {
			idxs[i], idxs[j] = idxs[j], idxs[i]
		}
	}
	for _, i := range idxs {
		first, last := ranges[i][0], ranges[i][1]
		var from, to int
		if dir < 0 {
			from, to = first-1, last
		} else {
			from, to = first, last+1
		}
		lines := append([]string{}, e.Buf.Lines()[from:to+1]...)
		var rearranged []string
		if dir < 0 {
			rearranged = append(append([]string{}, lines[1:]...), lines[0])
		} else {
			rearranged = append([]string{lines[len(lines)-1]}, lines[:len(lines)-1]...)
		}
		ed := e.replaceLines(from, to, rearranged)
		t.edits = append(t.edits, ed)
		for ci := range e.Cursors {
			h := e.Cursors[ci].Head.Line
			a := e.Cursors[ci].Anchor.Line
			if h >= first && h <= last {
				e.Cursors[ci].Head.Line += dir
			}
			if a >= first && a <= last {
				e.Cursors[ci].Anchor.Line += dir
			}
		}
		ranges[i][0] += dir
		ranges[i][1] += dir
	}
	e.clampCursors()
	e.commit(t, false)
	return true
}

// Indent inserts one level at the caret, or indents every touched line
// when a selection is present.
func (e *Editor) Indent() bool {
	if e.ReadOnly {
		return false
	}
	if !e.hasSelection() {
		return e.Insert(e.indentUnit(), false)
	}
	unit := e.indentUnit()
	return e.withLineRanges(func(first, last int) []string {
		repl := make([]string, 0, last-first+1)
		for i := first; i <= last; i++ {
			repl = append(repl, unit+e.Buf.Line(i))
		}
		return repl
	})
}

// Outdent removes one indent level from each touched line.
func (e *Editor) Outdent() bool {
	if e.ReadOnly {
		return false
	}
	return e.withLineRanges(func(first, last int) []string {
		repl := make([]string, 0, last-first+1)
		for i := first; i <= last; i++ {
			repl = append(repl, outdentLine(e.Buf.Line(i), e.tab()))
		}
		return repl
	})
}

func outdentLine(line string, tabSize int) string {
	if strings.HasPrefix(line, "\t") {
		return line[1:]
	}
	n := 0
	for n < len(line) && n < tabSize && line[n] == ' ' {
		n++
	}
	return line[n:]
}

// ToggleComment comments or uncomments touched lines.
// prefix is the line marker without a forced trailing space ("//", "#", "--").
func (e *Editor) ToggleComment(prefix string) bool {
	if e.ReadOnly || prefix == "" {
		return false
	}
	ranges := e.lineRanges()
	commented := true
	for _, r := range ranges {
		for i := r[0]; i <= r[1]; i++ {
			trim := strings.TrimLeftFunc(e.Buf.Line(i), unicode.IsSpace)
			if trim == "" {
				continue
			}
			if !hasComment(trim, prefix) {
				commented = false
			}
		}
	}
	mark := prefix
	if prefix != "--" && !strings.HasSuffix(prefix, " ") {
		mark = prefix + " "
	}
	return e.withLineRanges(func(first, last int) []string {
		repl := make([]string, 0, last-first+1)
		for i := first; i <= last; i++ {
			line := e.Buf.Line(i)
			if strings.TrimSpace(line) == "" {
				repl = append(repl, line)
				continue
			}
			if commented {
				repl = append(repl, stripComment(line, prefix))
			} else {
				ws := leadingWS(line)
				repl = append(repl, ws+mark+strings.TrimPrefix(line, ws))
			}
		}
		return repl
	})
}

func hasComment(trim, prefix string) bool {
	if strings.HasPrefix(trim, prefix+" ") || strings.HasPrefix(trim, prefix) {
		return true
	}
	return false
}

func stripComment(line, prefix string) string {
	ws := leadingWS(line)
	rest := strings.TrimPrefix(line, ws)
	if strings.HasPrefix(rest, prefix+" ") {
		return ws + strings.TrimPrefix(rest, prefix+" ")
	}
	if strings.HasPrefix(rest, prefix) {
		return ws + strings.TrimPrefix(rest, prefix)
	}
	return line
}

// Copy returns text for the clipboard. Empty selections copy the whole line
// when lineIfEmpty is set; the boolean reports that line mode was used.
func (e *Editor) Copy(lineIfEmpty bool) (string, bool) {
	var parts []string
	lineMode := lineIfEmpty
	for _, c := range e.Cursors {
		if c.Empty() {
			if !lineIfEmpty {
				lineMode = false
				continue
			}
			parts = append(parts, e.Buf.Line(c.Head.Line)+"\n")
			continue
		}
		lineMode = false
		a, b := c.Span()
		parts = append(parts, e.Buf.Text(a, b))
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, "\n"), lineMode
}

// Cut copies, then deletes the same region.
func (e *Editor) Cut(lineIfEmpty bool) (string, bool) {
	if e.ReadOnly {
		s, line := e.Copy(lineIfEmpty)
		return s, line
	}
	out, lineMode := e.Copy(lineIfEmpty)
	if out == "" {
		return "", false
	}
	if lineMode {
		e.DeleteLine()
		return out, true
	}
	e.Insert("", false)
	return out, false
}

// Paste inserts s, replacing selections.
func (e *Editor) Paste(s string) bool {
	if s == "" {
		return false
	}
	return e.Insert(s, false)
}

// AddNextOccurrence selects the next copy of the primary selection or word.
func (e *Editor) AddNextOccurrence() bool {
	e.fixPrimary()
	if e.Cursors[e.primary].Empty() {
		e.SelectWord()
	}
	a, b := e.Cursors[e.primary].Span()
	needle := e.Buf.Text(a, b)
	if needle == "" {
		return false
	}
	pos, ok := e.findRaw(needle, b)
	if !ok {
		pos, ok = e.findRaw(needle, text.Pos{})
	}
	if !ok {
		return false
	}
	end := text.Advance(pos, needle)
	for _, existing := range e.Cursors {
		es, ee := existing.Span()
		if es.Eq(pos) && ee.Eq(end) {
			return false
		}
	}
	e.Cursors = append(e.Cursors, Cursor{Anchor: pos, Head: end, Desired: -1})
	e.primary = len(e.Cursors) - 1
	return true
}

// SelectAllOccurrences selects every exact match of the primary needle.
func (e *Editor) SelectAllOccurrences() int {
	e.fixPrimary()
	if e.Cursors[e.primary].Empty() {
		e.SelectWord()
	}
	a, b := e.Cursors[e.primary].Span()
	needle := e.Buf.Text(a, b)
	if needle == "" {
		return 0
	}
	var cs []Cursor
	pos := text.Pos{}
	for {
		p, ok := e.findRaw(needle, pos)
		if !ok {
			break
		}
		end := text.Advance(p, needle)
		cs = append(cs, Cursor{Anchor: p, Head: end, Desired: -1})
		pos = end
		if !p.Less(end) {
			break
		}
	}
	if len(cs) == 0 {
		return 0
	}
	e.Cursors = cs
	e.primary = len(cs) - 1
	return len(cs)
}

func (e *Editor) findRaw(needle string, from text.Pos) (text.Pos, bool) {
	lines := e.Buf.Lines()
	for line := from.Line; line < len(lines); line++ {
		s := lines[line]
		col := 0
		if line == from.Line {
			col = from.Col
			if col > len(s) {
				continue
			}
		}
		idx := strings.Index(s[col:], needle)
		if idx >= 0 {
			return text.Pos{Line: line, Col: col + idx}, true
		}
	}
	return text.Pos{}, false
}

func (e *Editor) lineRanges() [][2]int {
	e.fixPrimary()
	ranges := make([][2]int, 0, len(e.Cursors))
	for _, c := range e.Cursors {
		a, b := c.Span()
		first, last := a.Line, b.Line
		if b.Col == 0 && last > first {
			last--
		}
		if last < first {
			last = first
		}
		if last >= e.Buf.LineCount() {
			last = e.Buf.LineCount() - 1
		}
		ranges = append(ranges, [2]int{first, last})
	}
	if len(ranges) == 0 {
		return [][2]int{{0, 0}}
	}
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[j][0] < ranges[i][0] {
				ranges[i], ranges[j] = ranges[j], ranges[i]
			}
		}
	}
	out := [][2]int{ranges[0]}
	for _, r := range ranges[1:] {
		last := &out[len(out)-1]
		if r[0] <= last[1]+1 {
			if r[1] > last[1] {
				last[1] = r[1]
			}
			continue
		}
		out = append(out, r)
	}
	return out
}

// withLineRanges replaces each touched line block. A nil replacement deletes it.
func (e *Editor) withLineRanges(fn func(first, last int) []string) bool {
	if e.ReadOnly {
		return false
	}
	ranges := e.lineRanges()
	before := e.snapshot()
	t := e.begin()
	for i := len(ranges) - 1; i >= 0; i-- {
		repl := fn(ranges[i][0], ranges[i][1])
		t.edits = append(t.edits, e.replaceLines(ranges[i][0], ranges[i][1], repl))
	}
	// Keep carets that still fall inside the document, preferring the
	// pre-edit primary column on its line.
	p := before[len(before)-1].Head
	p = e.Buf.Clamp(p)
	e.Cursors = []Cursor{{Head: p, Anchor: p, Desired: -1}}
	e.primary = 0
	e.commit(t, false)
	return true
}

// replaceLines replaces lines first..last inclusive.
// The newline that separates the block from the following line is preserved.
// Deleting the last line also removes the preceding newline.
func (e *Editor) replaceLines(first, last int, repl []string) edit {
	if first < 0 {
		first = 0
	}
	if last >= e.Buf.LineCount() {
		last = e.Buf.LineCount() - 1
	}
	deleteLast := repl == nil && last == e.Buf.LineCount()-1 && first > 0
	trailingNL := last+1 < e.Buf.LineCount() && !deleteLast
	var start, end text.Pos
	if deleteLast {
		start = text.Pos{Line: first - 1, Col: len(e.Buf.Line(first - 1))}
		end = text.Pos{Line: last, Col: len(e.Buf.Line(last))}
	} else {
		start = text.Pos{Line: first, Col: 0}
		if trailingNL {
			end = text.Pos{Line: last + 1, Col: 0}
		} else {
			end = text.Pos{Line: last, Col: len(e.Buf.Line(last))}
		}
	}
	old := e.Buf.Text(start, end)
	e.Buf.Delete(start, end)
	ins := ""
	if repl != nil {
		ins = strings.Join(repl, "\n")
		if trailingNL {
			ins += "\n"
		}
		if ins != "" {
			e.Buf.Insert(start, ins)
		}
	}
	return edit{start: start, deleted: old, inserted: ins}
}

func (e *Editor) clampCursors() {
	for i := range e.Cursors {
		e.Cursors[i].Anchor = e.Buf.Clamp(e.Cursors[i].Anchor)
		e.Cursors[i].Head = e.Buf.Clamp(e.Cursors[i].Head)
	}
	e.fixPrimary()
}
