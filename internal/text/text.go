// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package text is the geometry layer for Axam Studio.
//
// Positions use byte offsets that are always on a grapheme boundary.
// Visual columns are a separate measurement: tabs, East-Asian width,
// and emoji do not share an axis with bytes or runes.
package text

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// Pos is a caret location. Col is a byte offset into the line.
type Pos struct {
	Line int
	Col  int
}

// Less reports whether p is strictly before q.
func (p Pos) Less(q Pos) bool {
	return p.Line < q.Line || (p.Line == q.Line && p.Col < q.Col)
}

// Eq reports whether both coordinates match.
func (p Pos) Eq(q Pos) bool { return p.Line == q.Line && p.Col == q.Col }

// Order returns the earlier position first.
func Order(a, b Pos) (Pos, Pos) {
	if b.Less(a) {
		return b, a
	}
	return a, b
}

// Class is the word-movement class of a grapheme.
type Class int

const (
	ClassSpace Class = iota
	ClassWord
	ClassPunct
)

// NormalizeNL converts CR, LF, and CRLF to LF.
func NormalizeNL(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

// Advance returns the position reached by inserting s at p.
// s must already use LF, or it is normalized.
func Advance(p Pos, s string) Pos {
	s = NormalizeNL(s)
	if s == "" {
		return p
	}
	parts := strings.Split(s, "\n")
	if len(parts) == 1 {
		return Pos{Line: p.Line, Col: p.Col + len(parts[0])}
	}
	return Pos{Line: p.Line + len(parts) - 1, Col: len(parts[len(parts)-1])}
}

// Snap returns the grapheme start at or before col. len(s) is preserved.
func Snap(s string, col int) int {
	if col <= 0 {
		return 0
	}
	if col >= len(s) {
		return len(s)
	}
	last := 0
	stopped := false
	walk(s, func(start, end, _ int) bool {
		if start == col || end == col {
			last = col
			stopped = true
			return false
		}
		if start > col {
			stopped = true
			return false
		}
		last = start
		return true
	})
	if !stopped && col == len(s) {
		return len(s)
	}
	return last
}

// Next is the next grapheme boundary after col, or len(s).
func Next(s string, col int) int {
	col = Snap(s, col)
	if col >= len(s) {
		return len(s)
	}
	next := len(s)
	walk(s, func(start, end, _ int) bool {
		if start == col {
			next = end
			return false
		}
		return start < col
	})
	return next
}

// Prev is the grapheme boundary before col, or 0.
func Prev(s string, col int) int {
	if col <= 0 {
		return 0
	}
	if col > len(s) {
		col = len(s)
	}
	last := 0
	walk(s, func(start, end, _ int) bool {
		if end >= col {
			return false
		}
		last = end
		if start < col {
			last = start
		}
		return true
	})
	// last was set to each start that is < col; the final one is the boundary before col
	// when end >= col we return false without updating if start == col? Handle explicitly.
	prev := 0
	walk(s, func(start, _, _ int) bool {
		if start >= col {
			return false
		}
		prev = start
		return true
	})
	_ = last
	return prev
}

// ClassAt classifies the grapheme starting at col. Past the end is space.
func ClassAt(s string, col int) Class {
	if col < 0 {
		col = 0
	}
	if col >= len(s) {
		return ClassSpace
	}
	r, size := utf8.DecodeRuneInString(s[col:])
	if r == utf8.RuneError && size == 1 {
		return ClassPunct
	}
	if unicode.IsSpace(r) {
		return ClassSpace
	}
	if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
		return ClassWord
	}
	return ClassPunct
}

// FirstNonSpace returns the byte offset of the first non-space grapheme.
// An all-whitespace line returns len(s).
func FirstNonSpace(s string) int {
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return i
		}
		if !unicode.IsSpace(r) {
			return i
		}
		if size < 1 {
			break
		}
		i += size
	}
	return len(s)
}

// VisualCol is the terminal column of the grapheme at the byte offset.
func VisualCol(s string, col, tabSize int) int {
	if tabSize < 1 {
		tabSize = 1
	}
	col = Snap(s, col)
	vis := 0
	walk(s, func(start, end, width int) bool {
		if start >= col {
			return false
		}
		vis += cellWidth(s[start:end], vis, width, tabSize)
		return true
	})
	return vis
}

// ColAtVisual returns the grapheme boundary covering the target visual column.
// The position after the last cell is len(s).
func ColAtVisual(s string, target, tabSize int) int {
	if tabSize < 1 {
		tabSize = 1
	}
	if target <= 0 {
		return 0
	}
	vis := 0
	found := len(s)
	walk(s, func(start, end, width int) bool {
		w := cellWidth(s[start:end], vis, width, tabSize)
		if target < vis+w {
			found = start
			return false
		}
		vis += w
		found = end
		return true
	})
	return found
}

// Row is one visual line produced by wrapping.
type Row struct {
	Start int
	End   int
}

// Rows splits a logical line into visual rows of at most width cells.
// Width <= 0 yields a single row. Graphemes are never split. A cluster
// wider than width occupies a row by itself.
func Rows(s string, width, tabSize int) []Row {
	if width <= 0 {
		return []Row{{Start: 0, End: len(s)}}
	}
	if tabSize < 1 {
		tabSize = 1
	}
	var rows []Row
	rowStart := 0
	rowVis := 0
	abs := 0
	walk(s, func(start, end, gw int) bool {
		w := cellWidth(s[start:end], abs, gw, tabSize)
		if rowVis > 0 && rowVis+w > width {
			rows = append(rows, Row{Start: rowStart, End: start})
			rowStart = start
			rowVis = 0
		}
		rowVis += w
		abs += w
		if rowVis >= width {
			rows = append(rows, Row{Start: rowStart, End: end})
			rowStart = end
			rowVis = 0
		}
		return true
	})
	if rowStart < len(s) || len(rows) == 0 || (len(s) == 0) {
		rows = append(rows, Row{Start: rowStart, End: len(s)})
	} else if rowStart == len(s) && rowVis == 0 && len(s) > 0 {
		// The line ended exactly on a wrap boundary. The caret after the
		// last cell belongs to a following empty visual row only when the
		// caller asks for the position past the end. Keep the caret on the
		// last emitted row by not adding an extra empty row.
	}
	if len(rows) == 0 {
		return []Row{{Start: 0, End: len(s)}}
	}
	return rows
}

// RowIndex returns the visual row that contains col.
// A column equal to a row's End belongs to the next row when one exists,
// which is the caret sitting on the wrap boundary.
func RowIndex(rows []Row, col int) int {
	if len(rows) == 0 {
		return 0
	}
	for i, r := range rows {
		if i == len(rows)-1 || col < r.End {
			return i
		}
	}
	return len(rows) - 1
}

// ColAtRow places a caret on row at the given in-row visual column.
func ColAtRow(s string, row Row, screenCol, tabSize int) int {
	if screenCol <= 0 {
		return row.Start
	}
	base := VisualCol(s, row.Start, tabSize)
	target := base + screenCol
	endVis := VisualCol(s, row.End, tabSize)
	if target >= endVis {
		return row.End
	}
	col := ColAtVisual(s, target, tabSize)
	if col < row.Start {
		return row.Start
	}
	if col > row.End {
		return row.End
	}
	return col
}

// RowScreenCol is the in-row visual column of col.
func RowScreenCol(s string, row Row, col, tabSize int) int {
	if col < row.Start {
		col = row.Start
	}
	if col > row.End {
		col = row.End
	}
	return VisualCol(s, col, tabSize) - VisualCol(s, row.Start, tabSize)
}

func cellWidth(cluster string, vis, graphemeWidth, tabSize int) int {
	if cluster == "\t" {
		w := tabSize - (vis % tabSize)
		if w <= 0 {
			return tabSize
		}
		return w
	}
	if graphemeWidth < 1 {
		return 1
	}
	return graphemeWidth
}

func walk(s string, fn func(start, end, width int) bool) {
	i := 0
	state := -1
	for i < len(s) {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && sz == 1 {
			if !fn(i, i+1, 1) {
				return
			}
			i++
			state = -1
			continue
		}
		cluster, _, width, ns := uniseg.FirstGraphemeClusterInString(s[i:], state)
		if cluster == "" {
			if !fn(i, i+1, 1) {
				return
			}
			i++
			state = -1
			continue
		}
		if width < 1 {
			width = 1
		}
		end := i + len(cluster)
		if !fn(i, end, width) {
			return
		}
		i = end
		state = ns
	}
}
