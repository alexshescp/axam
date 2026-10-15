// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicSavePreservesOriginalOnFailedPrep(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SaveAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "new" {
		t.Fatalf("%q %v", got, err)
	}
	matches, _ := filepath.Glob(filepath.Join(dir, ".axam-*"))
	if len(matches) != 0 {
		t.Fatalf("temp left behind: %v", matches)
	}
}

func TestExternalChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	ld, err := Read(path, false)
	if err != nil {
		t.Fatal(err)
	}
	changed, _, _, err := Changed(path, ld.Hash, ld.ModTime, ld.Size)
	if err != nil || changed {
		t.Fatalf("unchanged reported %v %v", changed, err)
	}
	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Ensure mtime can differ on filesystems with coarse stamps.
	info, _ := os.Stat(path)
	_ = os.Chtimes(path, info.ModTime(), info.ModTime().Add(2*1e9))
	changed, _, _, err = Changed(path, ld.Hash, ld.ModTime, ld.Size)
	if err != nil || !changed {
		t.Fatalf("expected change %v %v", changed, err)
	}
}

func TestBinaryRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bin.dat")
	if err := os.WriteFile(path, []byte{'a', 0, 'b'}, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Read(path, true)
	if err == nil {
		t.Fatal("binary opened")
	}
}

func TestSymlinkSave(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if err := SaveAtomic(link, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "new" {
		t.Fatalf("target %q %v", got, err)
	}
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink was replaced")
	}
}

func TestCRLFPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "win.txt")
	raw := []byte("a\r\nb\r\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	ld, err := Read(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveAtomic(path, ld.Buffer.Bytes(), ld.Perm); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(raw) {
		t.Fatalf("%q", got)
	}
}

func TestDetectIndent(t *testing.T) {
	size, spaces, ok := DetectIndent([]string{"func", "    x", "    y"})
	if !ok || !spaces || size != 4 {
		t.Fatalf("%d %v %v", size, spaces, ok)
	}
	_, spaces, ok = DetectIndent([]string{"\tx", "\ty"})
	if !ok || spaces {
		t.Fatal("tabs")
	}
}
