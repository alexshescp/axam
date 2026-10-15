// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package text

import "testing"

func TestCyrillicBoundaries(t *testing.T) {
	s := "Привет"
	if Next(s, 0) != 2 {
		t.Fatalf("next = %d", Next(s, 0))
	}
	if Prev(s, 2) != 0 {
		t.Fatalf("prev = %d", Prev(s, 2))
	}
	if Snap(s, 1) != 0 {
		t.Fatalf("snap middle = %d", Snap(s, 1))
	}
	if VisualCol(s, 4, 4) != 2 {
		t.Fatalf("visual = %d", VisualCol(s, 4, 4))
	}
	if got := s[:Next(s, 0)]; got != "П" {
		t.Fatalf("cluster %q", got)
	}
}

func TestEmojiAndTabWidths(t *testing.T) {
	s := "a👍b"
	// 'a' is 1 byte, thumbs-up is 4 bytes, width 2.
	if w := VisualCol(s, len("a"), 4); w != 1 {
		t.Fatalf("before emoji visual %d", w)
	}
	after := Next(s, len("a"))
	if VisualCol(s, after, 4) != 3 {
		t.Fatalf("after emoji visual %d", VisualCol(s, after, 4))
	}
	if ColAtVisual(s, 1, 4) != len("a") {
		t.Fatalf("col at visual 1 = %d", ColAtVisual(s, 1, 4))
	}
	// A target inside the emoji lands on the emoji, not the following byte.
	if ColAtVisual(s, 2, 4) != len("a") {
		t.Fatalf("col inside emoji = %d", ColAtVisual(s, 2, 4))
	}

	tab := "\tX"
	if VisualCol(tab, 1, 4) != 4 {
		t.Fatalf("tab visual %d", VisualCol(tab, 1, 4))
	}
	if ColAtVisual(tab, 2, 4) != 0 {
		t.Fatalf("inside tab -> %d", ColAtVisual(tab, 2, 4))
	}
	if ColAtVisual(tab, 4, 4) != 1 {
		t.Fatalf("after tab -> %d", ColAtVisual(tab, 4, 4))
	}
}

func TestInvalidBytePreserved(t *testing.T) {
	s := "a\xffb"
	if Next(s, 1) != 2 {
		t.Fatalf("invalid byte width %d", Next(s, 1))
	}
	if VisualCol(s, 2, 4) != 2 {
		t.Fatalf("visual %d", VisualCol(s, 2, 4))
	}
}

func TestWrapRows(t *testing.T) {
	rows := Rows("abcdefgh", 4, 4)
	if len(rows) != 2 || rows[0].End != 4 || rows[1].Start != 4 {
		t.Fatalf("rows %+v", rows)
	}
	if RowIndex(rows, 4) != 1 {
		t.Fatalf("index %d", RowIndex(rows, 4))
	}
	if ColAtRow("abcdefgh", rows[1], 1, 4) != 5 {
		t.Fatalf("col %d", ColAtRow("abcdefgh", rows[1], 1, 4))
	}
}

func TestWordClasses(t *testing.T) {
	s := "foo_1 += бар"
	if ClassAt(s, 0) != ClassWord {
		t.Fatal("foo")
	}
	if ClassAt(s, 5) != ClassSpace {
		t.Fatal("space")
	}
	if ClassAt(s, 6) != ClassPunct {
		t.Fatal("punct")
	}
	if ClassAt(s, len("foo_1 += ")) != ClassWord {
		t.Fatal("cyrillic word")
	}
}
