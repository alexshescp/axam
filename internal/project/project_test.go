// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexshescp/axam/internal/search"
)

func TestSkipsGitAndLimit(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".gitdir"), 0o755); err == nil {
		_ = os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: .gitdir\n"), 0o644)
	}
	files, _, err := List(context.Background(), root, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if filepath.Base(f) == "config" {
			t.Fatal("walked .git")
		}
	}
	if len(files) != 1 {
		t.Fatalf("%v", files)
	}
	hits, err := Search(context.Background(), files, search.Options{Pattern: "func"}, 10)
	if err != nil || len(hits) != 1 || hits[0].Line != 2 {
		t.Fatalf("%v %v", hits, err)
	}
}

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(sub, "x.go")
	if err := os.WriteFile(file, []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindRoot(file); got != root {
		t.Fatalf("root %s want %s", got, root)
	}
}
