// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package search

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/codersclub/axam/internal/text"
)

// Options controls one search.
type Options struct {
	Pattern       string
	CaseSensitive bool
	WholeWord     bool
	Regex         bool
}

// Match is a half-open range in line/byte coordinates.
type Match struct {
	Start text.Pos
	End   text.Pos
}

// Find searches lines. An invalid regular expression returns an error.
func Find(lines []string, opt Options) ([]Match, error) {
	if opt.Pattern == "" {
		return nil, nil
	}
	if opt.Regex {
		return findRegex(lines, opt)
	}
	return findLiteral(lines, opt), nil
}

func findLiteral(lines []string, opt Options) []Match {
	needle := opt.Pattern
	if !opt.CaseSensitive {
		needle = strings.ToLower(needle)
	}
	var out []Match
	for i, line := range lines {
		hay := line
		if !opt.CaseSensitive {
			hay = strings.ToLower(line)
		}
		from := 0
		for from <= len(hay) {
			idx := strings.Index(hay[from:], needle)
			if idx < 0 {
				break
			}
			start := from + idx
			end := start + len(needle)
			if !opt.WholeWord || isWord(line, start, end) {
				out = append(out, Match{
					Start: text.Pos{Line: i, Col: start},
					End:   text.Pos{Line: i, Col: end},
				})
			}
			from = start + 1
			if from > len(hay) {
				break
			}
		}
	}
	return out
}

func findRegex(lines []string, opt Options) ([]Match, error) {
	pat := opt.Pattern
	if !opt.CaseSensitive && !strings.HasPrefix(pat, "(?i)") {
		pat = "(?i)" + pat
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression: %w", err)
	}
	var out []Match
	for i, line := range lines {
		idxs := re.FindAllStringIndex(line, -1)
		for _, pair := range idxs {
			if opt.WholeWord && !isWord(line, pair[0], pair[1]) {
				continue
			}
			out = append(out, Match{
				Start: text.Pos{Line: i, Col: pair[0]},
				End:   text.Pos{Line: i, Col: pair[1]},
			})
		}
	}
	return out, nil
}

func isWord(line string, start, end int) bool {
	if start > 0 {
		r, _ := lastRune(line[:start])
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	if end < len(line) {
		r, _ := firstRune(line[end:])
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func firstRune(s string) (rune, int) {
	for i, r := range s {
		return r, i
	}
	return 0, 0
}

func lastRune(s string) (rune, int) {
	var r rune
	for _, rr := range s {
		r = rr
	}
	return r, 0
}
