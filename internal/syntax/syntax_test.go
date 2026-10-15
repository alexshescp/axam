// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package syntax

import (
	"testing"

	"github.com/codersclub/axam/internal/text"
)

func TestLanguagesPresent(t *testing.T) {
	for _, lang := range Supported() {
		if lexerFor(lang) == nil {
			t.Fatalf("missing lexer %s", lang)
		}
	}
	spans := HighlightLine("go", "func main() {}")
	found := false
	for _, s := range spans {
		if s.Kind == "keyword" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no keyword in %+v", spans)
	}
	if HighlightLine("text", "func") != nil {
		t.Fatal("plain text should not highlight")
	}
}

func TestBracket(t *testing.T) {
	lines := []string{"func main() {", "\tx := (1)", "}"}
	pos, ok := MatchBracket(lines, text.Pos{Line: 0, Col: len(lines[0]) - 1})
	if !ok || pos.Line != 2 {
		t.Fatalf("%v %+v", ok, pos)
	}
}

func TestCommentPrefixes(t *testing.T) {
	if CommentPrefix("go") != "//" || CommentPrefix("python") != "#" || CommentPrefix("sql") != "--" {
		t.Fatal(CommentPrefix("go"), CommentPrefix("python"), CommentPrefix("sql"))
	}
	if CommentPrefix("json") != "" {
		t.Fatal("json")
	}
}
