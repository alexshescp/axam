// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package buffer stores document text as lines.
//
// The editing engine uses LF between lines. Serialization restores the
// file's line-ending style. A piece table was considered and rejected for
// the first release: line-oriented editing, undo logs, and viewport drawing
// stay easier to test on a slice of strings, and the cost of moving line
// headers is small for source files.
package buffer

import (
	"crypto/sha256"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/codersclub/axam/internal/text"
)

// EOL is the on-disk line ending.
type EOL int

const (
	LF EOL = iota
	CRLF
	CR
)

func (e EOL) String() string {
	switch e {
	case CRLF:
		return "CRLF"
	case CR:
		return "CR"
	default:
		return "LF"
	}
}

func (e EOL) Bytes() string {
	switch e {
	case CRLF:
		return "\r\n"
	case CR:
		return "\r"
	default:
		return "\n"
	}
}

// Encoding names stored with a document.
const (
	EncUTF8    = "UTF-8"
	EncUTF16LE = "UTF-16LE"
	EncUTF16BE = "UTF-16BE"
)

// Buffer is a sequence of lines without line-ending bytes.
// It is not safe for concurrent use.
type Buffer struct {
	lines []string
	eol   EOL
	mixed bool
	gen   uint64
}

// New returns an empty buffer with a single empty line.
func New() *Buffer {
	return &Buffer{lines: []string{""}, eol: LF}
}

// FromString parses s as UTF-8 text.
func FromString(s string) *Buffer {
	b, _ := FromBytes([]byte(s))
	return b.Buffer
}

// Loaded is the metadata needed to round-trip a file.
type Loaded struct {
	Buffer   *Buffer
	Encoding string
	BOM      bool
	Hash     [32]byte
}

// FromBytes decodes a file body. UTF-16 is recognized by BOM.
// Invalid UTF-8 is kept byte-for-byte inside Go strings.
func FromBytes(data []byte) (*Loaded, error) {
	sum := sha256.Sum256(data)
	enc := EncUTF8
	bom := false
	var body []byte
	switch {
	case len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF:
		bom = true
		body = data[3:]
	case len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE:
		bom = true
		enc = EncUTF16LE
		body = []byte(decodeUTF16(data[2:], true))
	case len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF:
		bom = true
		enc = EncUTF16BE
		body = []byte(decodeUTF16(data[2:], false))
	default:
		body = data
	}
	s := string(body)
	eol, mixed := detectEOL(s)
	norm := text.NormalizeNL(s)
	lines := strings.Split(norm, "\n")
	if lines == nil {
		lines = []string{""}
	}
	return &Loaded{
		Buffer:   &Buffer{lines: lines, eol: eol, mixed: mixed, gen: 1},
		Encoding: enc,
		BOM:      bom,
		Hash:     sum,
	}, nil
}

func decodeUTF16(b []byte, le bool) string {
	if len(b)%2 == 1 {
		b = b[:len(b)-1]
	}
	u := make([]uint16, len(b)/2)
	for i := 0; i < len(u); i++ {
		if le {
			u[i] = uint16(b[i*2]) | uint16(b[i*2+1])<<8
		} else {
			u[i] = uint16(b[i*2+1]) | uint16(b[i*2])<<8
		}
	}
	return string(utf16.Decode(u))
}

// Encode serializes UTF-8 file bytes into the original encoding.
func Encode(content []byte, encoding string, bom bool) []byte {
	switch encoding {
	case EncUTF16LE, EncUTF16BE:
		s := string(content)
		if strings.HasPrefix(s, "\uFEFF") {
			s = strings.TrimPrefix(s, "\uFEFF")
		}
		u := utf16.Encode([]rune(s))
		out := make([]byte, 0, 2+len(u)*2)
		if bom || true {
			if encoding == EncUTF16LE {
				out = append(out, 0xFF, 0xFE)
			} else {
				out = append(out, 0xFE, 0xFF)
			}
		}
		for _, w := range u {
			if encoding == EncUTF16LE {
				out = append(out, byte(w), byte(w>>8))
			} else {
				out = append(out, byte(w>>8), byte(w))
			}
		}
		return out
	default:
		if bom {
			return append([]byte{0xEF, 0xBB, 0xBF}, content...)
		}
		return content
	}
}

func detectEOL(s string) (EOL, bool) {
	crlf, lf, cr := 0, 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				crlf++
				i++
			} else {
				cr++
			}
		case '\n':
			lf++
		}
	}
	kinds := 0
	if crlf > 0 {
		kinds++
	}
	if lf > 0 {
		kinds++
	}
	if cr > 0 {
		kinds++
	}
	mixed := kinds > 1
	switch {
	case crlf >= lf && crlf >= cr && crlf > 0:
		return CRLF, mixed
	case cr > lf && cr > 0:
		return CR, mixed
	default:
		return LF, mixed
	}
}

// Gen changes whenever the text changes.
func (b *Buffer) Gen() uint64 { return b.gen }

// EOL reports the serialized line ending.
func (b *Buffer) EOL() EOL { return b.eol }

// SetEOL changes the serialized line ending.
func (b *Buffer) SetEOL(e EOL) { b.eol = e }

// Mixed reports whether the original bytes used more than one ending.
func (b *Buffer) Mixed() bool { return b.mixed }

// LineCount is the number of logical lines. An empty buffer has one.
func (b *Buffer) LineCount() int {
	if len(b.lines) == 0 {
		return 1
	}
	return len(b.lines)
}

// Line returns the line without its ending. Out-of-range returns "".
func (b *Buffer) Line(i int) string {
	if i < 0 || i >= len(b.lines) {
		return ""
	}
	return b.lines[i]
}

// SetLineForTest replaces one line. It exists for tests that build geometry.
func (b *Buffer) SetLineForTest(i int, s string) {
	if i >= 0 && i < len(b.lines) {
		b.lines[i] = s
		b.gen++
	}
}

// Lines returns the line slice. Callers must not mutate it.
func (b *Buffer) Lines() []string { return b.lines }

// String joins lines with LF.
func (b *Buffer) String() string { return strings.Join(b.lines, "\n") }

// Bytes serializes using the buffer line ending.
func (b *Buffer) Bytes() []byte {
	if len(b.lines) == 0 {
		return nil
	}
	eol := b.eol.Bytes()
	var sb strings.Builder
	for i, line := range b.lines {
		if i > 0 {
			sb.WriteString(eol)
		}
		sb.WriteString(line)
	}
	return []byte(sb.String())
}

// Clamp snaps p into the buffer and onto a grapheme boundary.
func (b *Buffer) Clamp(p text.Pos) text.Pos {
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	if p.Line < 0 {
		p.Line = 0
	}
	if p.Line >= len(b.lines) {
		p.Line = len(b.lines) - 1
		p.Col = len(b.lines[p.Line])
		return p
	}
	if p.Col < 0 {
		p.Col = 0
	}
	if p.Col > len(b.lines[p.Line]) {
		p.Col = len(b.lines[p.Line])
	}
	p.Col = text.Snap(b.lines[p.Line], p.Col)
	return p
}

// Text returns the half-open range [start, end) using LF between lines.
func (b *Buffer) Text(start, end text.Pos) string {
	start, end = text.Order(start, end)
	start = b.Clamp(start)
	end = b.Clamp(end)
	if !start.Less(end) {
		return ""
	}
	if start.Line == end.Line {
		line := b.lines[start.Line]
		return line[start.Col:end.Col]
	}
	var sb strings.Builder
	sb.WriteString(b.lines[start.Line][start.Col:])
	for i := start.Line + 1; i < end.Line; i++ {
		sb.WriteByte('\n')
		sb.WriteString(b.lines[i])
	}
	sb.WriteByte('\n')
	sb.WriteString(b.lines[end.Line][:end.Col])
	return sb.String()
}

// Insert inserts s at p and returns the caret after the insertion.
func (b *Buffer) Insert(p text.Pos, s string) text.Pos {
	p = b.Clamp(p)
	s = text.NormalizeNL(s)
	if !utf8.ValidString(s) {
		// Keep raw bytes; do not replace them.
	}
	parts := strings.Split(s, "\n")
	line := b.lines[p.Line]
	left, right := line[:p.Col], line[p.Col:]
	if len(parts) == 1 {
		b.lines[p.Line] = left + parts[0] + right
		b.gen++
		return text.Pos{Line: p.Line, Col: p.Col + len(parts[0])}
	}
	repl := make([]string, 0, len(parts))
	repl = append(repl, left+parts[0])
	if len(parts) > 2 {
		repl = append(repl, parts[1:len(parts)-1]...)
	}
	repl = append(repl, parts[len(parts)-1]+right)
	b.replace(p.Line, p.Line, repl)
	b.gen++
	return text.Pos{Line: p.Line + len(parts) - 1, Col: len(parts[len(parts)-1])}
}

// Delete removes [start, end) and returns the deleted LF-joined text.
func (b *Buffer) Delete(start, end text.Pos) string {
	start, end = text.Order(start, end)
	start = b.Clamp(start)
	end = b.Clamp(end)
	if !start.Less(end) {
		return ""
	}
	deleted := b.Text(start, end)
	left := b.lines[start.Line][:start.Col]
	right := b.lines[end.Line][end.Col:]
	b.replace(start.Line, end.Line, []string{left + right})
	b.gen++
	return deleted
}

func (b *Buffer) replace(first, last int, repl []string) {
	if len(repl) == 0 {
		repl = []string{""}
	}
	out := make([]string, 0, len(b.lines)-(last-first+1)+len(repl))
	out = append(out, b.lines[:first]...)
	out = append(out, repl...)
	if last+1 < len(b.lines) {
		out = append(out, b.lines[last+1:]...)
	}
	if len(out) == 0 {
		out = []string{""}
	}
	b.lines = out
}

// IsBinary reports whether a prefix looks like non-text.
func IsBinary(prefix []byte) bool {
	n := len(prefix)
	if n > 8192 {
		n = 8192
	}
	for i := 0; i < n; i++ {
		if prefix[i] == 0 {
			return true
		}
	}
	return false
}
