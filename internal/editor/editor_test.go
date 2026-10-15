// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package editor

import (
	"strings"
	"testing"

	"github.com/codersclub/axam/internal/buffer"
	"github.com/codersclub/axam/internal/text"
)

func doc(s string) *Editor {
	e := New(buffer.FromString(s))
	e.TabSize = 4
	e.UseSpaces = true
	return e
}

func TestShiftArrowsAnchor(t *testing.T) {
	e := doc("abcdef\nghijkl\nmnopqr")
	e.SetCaret(text.Pos{Col: 3}, false)
	e.Right(true)
	e.Right(true)
	if e.SelectedText() != "de" {
		t.Fatalf("select %q", e.SelectedText())
	}
	a := e.Primary().Anchor
	e.Left(true)
	e.Left(true)
	e.Left(true)
	e.Left(true)
	if !e.Primary().Anchor.Eq(a) {
		t.Fatalf("anchor jumped to %+v", e.Primary().Anchor)
	}
	if e.SelectedText() != "bc" {
		t.Fatalf("reversed %q head %+v", e.SelectedText(), e.Primary().Head)
	}
	// Direction is now backward: head is before the anchor.
	if !e.Primary().Head.Less(e.Primary().Anchor) {
		t.Fatal("expected backward selection")
	}
}

func TestShiftVerticalUnevenAndEmpty(t *testing.T) {
	e := doc("abcdef\n\nxy\nabcdefghij")
	// Place caret at visual column 4 of the first line.
	e.SetCaret(text.Pos{Col: 4}, false)
	e.Down(true, 0)
	e.Down(true, 0)
	// Empty line clamps the caret but must not destroy the desired column.
	e.Down(true, 0)
	c := e.Primary()
	if c.Head.Line != 3 || c.Head.Col != 4 {
		t.Fatalf("head %+v desired %d", c.Head, c.Desired)
	}
	if !c.Anchor.Eq(text.Pos{Col: 4}) {
		t.Fatalf("anchor %+v", c.Anchor)
	}
	got := e.SelectedText()
	if got != "ef\n\nxy\nabcd" {
		t.Fatalf("selected %q", got)
	}
}

func TestShiftUpDownWrapped(t *testing.T) {
	e := doc("abcdefghij")
	e.Down(true, 4)
	if e.Primary().Head.Col != 4 {
		t.Fatalf("wrapped head col %d", e.Primary().Head.Col)
	}
	if e.SelectedText() != "abcd" {
		t.Fatalf("wrapped sel %q", e.SelectedText())
	}
	e.Up(true, 4)
	if !e.Primary().Head.Eq(text.Pos{}) || e.SelectedText() != "" {
		t.Fatalf("back to start %+v %q", e.Primary().Head, e.SelectedText())
	}
}

func TestTabsCyrillicEmojiSelection(t *testing.T) {
	e := doc("\tП👍")
	e.Right(true)
	if e.SelectedText() != "\t" {
		t.Fatalf("tab %q", e.SelectedText())
	}
	e.Right(true)
	if !strings.HasSuffix(e.SelectedText(), "П") {
		t.Fatalf("cyr %q", e.SelectedText())
	}
	e.Right(true)
	if !strings.HasSuffix(e.SelectedText(), "👍") {
		t.Fatalf("emoji %q", e.SelectedText())
	}
	// Visual column after tab (4) + П (1) + emoji (2).
	if text.VisualCol(e.Buf.Line(0), e.Primary().Head.Col, 4) != 7 {
		t.Fatalf("visual %d col %d", text.VisualCol(e.Buf.Line(0), e.Primary().Head.Col, 4), e.Primary().Head.Col)
	}
}

func TestWordAndLineAndDocumentSelection(t *testing.T) {
	e := doc("foo bar\nbaz")
	e.WordRight(true)
	if e.SelectedText() != "foo" {
		t.Fatalf("word %q", e.SelectedText())
	}
	e.SetCaret(text.Pos{Col: 4}, false)
	e.WordLeft(true)
	if e.SelectedText() != "foo " && e.SelectedText() != " " {
		// From the space, word-left lands at start of foo, selecting "foo ".
		if e.SelectedText() != "foo " {
			t.Fatalf("word left %q", e.SelectedText())
		}
	}
	e.SetCaret(text.Pos{Line: 0, Col: 2}, false)
	e.SelectWord()
	if e.SelectedText() != "foo" {
		t.Fatalf("select word %q", e.SelectedText())
	}
	e.SelectLine()
	if e.SelectedText() != "foo bar\n" {
		t.Fatalf("line %q", e.SelectedText())
	}
	e.DocEnd(true)
	if !strings.HasSuffix(e.SelectedText(), "baz") {
		t.Fatalf("doc %q", e.SelectedText())
	}
	e.SetCaret(text.Pos{Line: 1, Col: 1}, false)
	e.DocStart(true)
	if e.SelectedText() != "foo bar\nb" {
		t.Fatalf("to start %q", e.SelectedText())
	}
}

func TestCollapseWithoutShift(t *testing.T) {
	e := doc("abcdef")
	e.Right(true)
	e.Right(true)
	e.Left(false)
	if e.Primary().Head.Col != 0 || !e.Primary().Empty() {
		t.Fatalf("collapse left %+v", e.Primary())
	}
	e.Right(true)
	e.Right(true)
	e.Right(false)
	if e.Primary().Head.Col != 2 || !e.Primary().Empty() {
		t.Fatalf("collapse right %+v", e.Primary())
	}
}

func TestTypingUndoPasteMultiCursor(t *testing.T) {
	e := doc("")
	for _, ch := range "hello" {
		e.Insert(string(ch), true)
	}
	if e.Buf.String() != "hello" {
		t.Fatal(e.Buf.String())
	}
	if !e.Undo() || e.Buf.String() != "" {
		t.Fatalf("coalesced undo %q", e.Buf.String())
	}
	e.Redo()
	e.Paste("\nworld\n")
	if e.Buf.String() != "hello\nworld\n" {
		t.Fatalf("paste %q", e.Buf.String())
	}
	if !e.Undo() || e.Buf.String() != "hello" {
		t.Fatalf("paste undo %q", e.Buf.String())
	}

	e = doc("abcde")
	e.SetCaret(text.Pos{Col: 1}, false)
	e.AddCaret(text.Pos{Col: 3})
	e.Insert("X", false)
	if e.Buf.String() != "aXbcXde" {
		t.Fatalf("multi %q", e.Buf.String())
	}
	e.Undo()
	if e.Buf.String() != "abcde" {
		t.Fatalf("multi undo %q", e.Buf.String())
	}
}

func TestLineOpsAndComment(t *testing.T) {
	e := doc("a\nb\nc")
	e.SetCaret(text.Pos{Line: 1, Col: 0}, false)
	e.DuplicateLine()
	if e.Buf.String() != "a\nb\nb\nc" {
		t.Fatalf("dup %q", e.Buf.String())
	}
	e.Undo()
	e.SetCaret(text.Pos{Line: 1, Col: 0}, false)
	e.MoveLines(-1)
	if e.Buf.String() != "b\na\nc" {
		t.Fatalf("move %q", e.Buf.String())
	}
	e.Undo()
	e.SetCaret(text.Pos{Line: 1, Col: 0}, false)
	e.DeleteLine()
	if e.Buf.String() != "a\nc" {
		t.Fatalf("del %q", e.Buf.String())
	}
	e = doc("\tkeep\n\talso")
	e.SelectAll()
	e.UseSpaces = false
	e.ToggleComment("#")
	if e.Buf.String() != "\t# keep\n\t# also" {
		t.Fatalf("comment %q", e.Buf.String())
	}
	e.SelectAll()
	e.ToggleComment("#")
	if e.Buf.String() != "\tkeep\n\talso" {
		t.Fatalf("uncomment %q", e.Buf.String())
	}
}

func TestIndentOutdentDesiredColumn(t *testing.T) {
	e := doc("ab\ncd")
	e.SetCaret(text.Pos{Col: 0}, false)
	e.SelectAll()
	e.Indent()
	if e.Buf.String() != "    ab\n    cd" {
		t.Fatalf("indent %q", e.Buf.String())
	}
	e.SelectAll()
	e.Outdent()
	if e.Buf.String() != "ab\ncd" {
		t.Fatalf("outdent %q", e.Buf.String())
	}
}

func TestReadOnlyAndLineEndings(t *testing.T) {
	ld, err := buffer.FromBytes([]byte("a\r\nb"))
	if err != nil {
		t.Fatal(err)
	}
	e := New(ld.Buffer)
	e.Insert("Z", true)
	if string(e.Buf.Bytes()) != "Za\r\nb" {
		t.Fatalf("%q", e.Buf.Bytes())
	}
	e.ReadOnly = true
	e.Insert("Q", true)
	if strings.Contains(e.Buf.String(), "Q") {
		t.Fatal("read-only accepted input")
	}
}

func TestLargeFileEdit(t *testing.T) {
	var b strings.Builder
	const n = 20000
	for i := 0; i < n; i++ {
		b.WriteString("line\n")
	}
	e := doc(b.String())
	e.GoToLine(10000)
	e.Insert("X", false)
	if !strings.HasPrefix(e.Buf.Line(9999), "X") {
		t.Fatalf("line %q", e.Buf.Line(9999))
	}
	e.Undo()
	if e.Buf.Line(9999) != "line" {
		t.Fatalf("undo %q", e.Buf.Line(9999))
	}
}

func TestNextOccurrence(t *testing.T) {
	e := doc("one two one")
	e.SelectWord()
	if !e.AddNextOccurrence() {
		t.Fatal("no next")
	}
	if len(e.Cursors) != 2 {
		t.Fatalf("cursors %d", len(e.Cursors))
	}
	e.Insert("1", false)
	if e.Buf.String() != "1 two 1" {
		t.Fatalf("%q", e.Buf.String())
	}
}

func TestReplaceRangesOneUndo(t *testing.T) {
	e := doc("a a a")
	n := e.ReplaceRanges([][2]text.Pos{
		{{Line: 0, Col: 0}, {Line: 0, Col: 1}},
		{{Line: 0, Col: 2}, {Line: 0, Col: 3}},
		{{Line: 0, Col: 4}, {Line: 0, Col: 5}},
	}, "bb")
	if n != 3 || e.Buf.String() != "bb bb bb" {
		t.Fatalf("n=%d %q", n, e.Buf.String())
	}
	e.Undo()
	if e.Buf.String() != "a a a" {
		t.Fatalf("undo %q", e.Buf.String())
	}
}

func TestCopyLineAndCut(t *testing.T) {
	e := doc("alpha\nbeta")
	s, line := e.Copy(true)
	if !line || s != "alpha\n" {
		t.Fatalf("copy %q %v", s, line)
	}
	e.Cut(true)
	if e.Buf.String() != "beta" {
		t.Fatalf("cut %q", e.Buf.String())
	}
}
