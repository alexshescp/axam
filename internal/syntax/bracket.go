// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package syntax

import (
	"unicode/utf8"

	"github.com/codersclub/axam/internal/text"
)

var pairs = map[rune]rune{
	'(': ')',
	'[': ']',
	'{': '}',
	')': '(',
	']': '[',
	'}': '{',
}

// MatchBracket finds the partner of the bracket under or just before p.
// Strings and line comments are skipped with a simple scan, not a real parser.
func MatchBracket(lines []string, p text.Pos) (text.Pos, bool) {
	if p.Line < 0 || p.Line >= len(lines) {
		return text.Pos{}, false
	}
	line := lines[p.Line]
	col, r, ok := bracketAt(line, p.Col)
	if !ok && p.Col > 0 {
		col, r, ok = bracketAt(line, text.Prev(line, p.Col))
	}
	if !ok {
		return text.Pos{}, false
	}
	open := r == '(' || r == '[' || r == '{'
	want := pairs[r]
	depth := 0
	if open {
		return scanForward(lines, text.Pos{Line: p.Line, Col: col}, r, want)
	}
	return scanBackward(lines, text.Pos{Line: p.Line, Col: col}, r, want, depth)
}

func bracketAt(line string, col int) (int, rune, bool) {
	if col < 0 || col >= len(line) {
		return 0, 0, false
	}
	r, _ := utf8.DecodeRuneInString(line[col:])
	_, ok := pairs[r]
	if !ok {
		return 0, 0, false
	}
	return col, r, true
}

func scanForward(lines []string, from text.Pos, open, close rune) (text.Pos, bool) {
	depth := 0
	for line := from.Line; line < len(lines); line++ {
		s := lines[line]
		i := 0
		if line == from.Line {
			i = from.Col
		}
		for i < len(s) {
			if isLineComment(s[i:]) {
				break
			}
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == '"' || r == '\'' || r == '`' {
				i = skipString(s, i, r)
				continue
			}
			if r == open {
				depth++
			} else if r == close {
				depth--
				if depth == 0 {
					return text.Pos{Line: line, Col: i}, true
				}
			}
			i += size
		}
	}
	return text.Pos{}, false
}

func scanBackward(lines []string, from text.Pos, close, open rune, depth int) (text.Pos, bool) {
	_ = close
	for line := from.Line; line >= 0; line-- {
		s := lines[line]
		end := len(s)
		if line == from.Line {
			end = from.Col
		}
		// Walk runes and remember positions, then scan backward.
		type rk struct {
			i int
			r rune
		}
		var rs []rk
		for i := 0; i < end; {
			r, size := utf8.DecodeRuneInString(s[i:])
			rs = append(rs, rk{i, r})
			i += size
		}
		for n := len(rs) - 1; n >= 0; n-- {
			r := rs[n].r
			if r == close {
				depth++
			} else if r == open {
				depth--
				if depth == 0 {
					return text.Pos{Line: line, Col: rs[n].i}, true
				}
			}
		}
	}
	return text.Pos{}, false
}

func isLineComment(s string) bool {
	return stringsHasPrefix(s, "//")
}

func stringsHasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }

func skipString(s string, i int, quote rune) int {
	_, size := utf8.DecodeRuneInString(s[i:])
	i += size
	for i < len(s) {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if r == '\\' && quote != '`' {
			i += sz
			if i < len(s) {
				_, sz2 := utf8.DecodeRuneInString(s[i:])
				i += sz2
			}
			continue
		}
		i += sz
		if r == quote {
			break
		}
	}
	return i
}
