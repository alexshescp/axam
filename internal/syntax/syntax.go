// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package syntax highlights viewport lines with Chroma.
// Tree-sitter was not used because the Go bindings require cgo, which would
// break the static single-binary builds shipped for apt and direct download.
// Highlighting is lexical and per line, so a multiline string can be wrong
// until the cursor is on a line the lexer can see on its own.
package syntax

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Span is a byte range on one line and a coarse token class.
type Span struct {
	Start int
	End   int
	Kind  string
}

// Language returns a stable language id for a path, or "text".
func Language(path string) string {
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "dockerfile":
		return "dockerfile"
	case "nginx.conf":
		return "nginx"
	}
	if strings.Contains(strings.ToLower(filepath.ToSlash(path)), "/nginx/") && strings.HasSuffix(base, ".conf") {
		return "nginx"
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".json":
		return "json"
	case ".yml", ".yaml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".sql":
		return "sql"
	case ".md", ".markdown":
		return "markdown"
	case ".conf":
		if strings.Contains(base, "nginx") {
			return "nginx"
		}
	}
	return "text"
}

// CommentPrefix is the line comment marker, or "" when the language has none.
func CommentPrefix(lang string) string {
	switch lang {
	case "go", "javascript", "typescript", "json":
		if lang == "json" {
			return ""
		}
		return "//"
	case "python", "yaml", "toml", "bash", "nginx", "dockerfile":
		return "#"
	case "sql":
		return "--"
	case "html", "xml", "markdown":
		return "<!--"
	case "css":
		return "/*"
	default:
		return ""
	}
}

// HighlightLine tokenizes one line. Unknown languages return no spans.
func HighlightLine(lang, line string) []Span {
	lexer := lexerFor(lang)
	if lexer == nil || line == "" {
		return nil
	}
	it, err := lexer.Tokenise(nil, line)
	if err != nil || it == nil {
		return nil
	}
	var spans []Span
	off := 0
	for tok := it(); tok != chroma.EOF; tok = it() {
		if tok.Value == "" {
			continue
		}
		kind := kindOf(tok.Type)
		end := off + len(tok.Value)
		if kind != "" {
			spans = append(spans, Span{Start: off, End: end, Kind: kind})
		}
		off = end
	}
	return spans
}

func lexerFor(lang string) chroma.Lexer {
	name := lang
	switch lang {
	case "go":
		name = "Go"
	case "python":
		name = "Python"
	case "javascript":
		name = "JavaScript"
	case "typescript":
		name = "TypeScript"
	case "json":
		name = "JSON"
	case "yaml":
		name = "YAML"
	case "toml":
		name = "TOML"
	case "xml":
		name = "XML"
	case "html":
		name = "HTML"
	case "css":
		name = "CSS"
	case "bash":
		name = "Bash"
	case "sql":
		name = "SQL"
	case "markdown":
		name = "Markdown"
	case "nginx":
		name = "Nginx configuration file"
	case "dockerfile":
		name = "Dockerfile"
	default:
		return nil
	}
	l := lexers.Get(name)
	if l == nil {
		return nil
	}
	return chroma.Coalesce(l)
}

func kindOf(t chroma.TokenType) string {
	switch {
	case t.InCategory(chroma.Comment):
		return "comment"
	case t.InCategory(chroma.Keyword):
		return "keyword"
	case t.InCategory(chroma.String):
		return "string"
	case t.InCategory(chroma.Number):
		return "number"
	case t.InCategory(chroma.NameFunction) || t == chroma.NameFunction:
		return "func"
	case t == chroma.NameClass || t == chroma.KeywordType || t == chroma.NameBuiltin:
		return "type"
	default:
		return ""
	}
}

// Supported lists language ids the highlighter knows.
func Supported() []string {
	return []string{
		"go", "python", "javascript", "typescript", "json", "yaml", "toml",
		"xml", "html", "css", "bash", "sql", "markdown", "nginx", "dockerfile",
	}
}
