// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package search

import "testing"

func TestLiteralCaseAndWord(t *testing.T) {
	lines := []string{"Foo foo", "foobar", "Привет привет"}
	got, err := Find(lines, Options{Pattern: "foo"})
	if err != nil || len(got) != 3 {
		t.Fatalf("%d %v", len(got), err)
	}
	got, _ = Find(lines, Options{Pattern: "foo", CaseSensitive: true})
	if len(got) != 2 || got[0].Start.Col != 4 || got[1].Start.Line != 1 {
		t.Fatalf("%+v", got)
	}
	got, _ = Find(lines, Options{Pattern: "foo", WholeWord: true})
	if len(got) != 2 {
		t.Fatalf("whole %d", len(got))
	}
	got, _ = Find(lines, Options{Pattern: "привет"})
	if len(got) != 2 {
		t.Fatalf("cyr %d", len(got))
	}
}

func TestBadRegex(t *testing.T) {
	_, err := Find([]string{"a"}, Options{Pattern: "(", Regex: true})
	if err == nil || !stringsContains(err.Error(), "invalid regular expression") {
		t.Fatal(err)
	}
}

func TestRegex(t *testing.T) {
	got, err := Find([]string{"a1 b2"}, Options{Pattern: `[a-z]\d`, Regex: true})
	if err != nil || len(got) != 2 {
		t.Fatalf("%v %+v", err, got)
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && (s[:len(sub)] == sub || stringsContains(s[1:], sub))))
}
