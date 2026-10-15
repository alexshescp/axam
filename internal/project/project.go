// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package project

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexshescp/axam/internal/buffer"
	"github.com/alexshescp/axam/internal/search"
)

// DefaultIgnore is always applied, in addition to user patterns.
var DefaultIgnore = []string{
	".git", ".hg", ".svn", "node_modules", "vendor", "dist", "build",
	"target", "__pycache__", ".idea", ".vscode", "bin", "obj",
}

// Hit is one project-search result.
type Hit struct {
	Path string
	Line int
	Col  int
	Text string
}

// FindRoot walks up from start looking for a .git directory.
// A directory start is used as the root when no repository is found.
func FindRoot(start string) string {
	info, err := os.Stat(start)
	if err != nil {
		return start
	}
	dir := start
	if !info.IsDir() {
		dir = filepath.Dir(start)
	}
	cur := dir
	for {
		if st, err := os.Stat(filepath.Join(cur, ".git")); err == nil && st.IsDir() {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return dir
		}
		cur = parent
	}
}

// List walks root and returns file paths. Directory symlinks are not followed.
// The walk stops at limit. truncated is set when the limit was hit.
func List(ctx context.Context, root string, extra []string, limit int) (files []string, truncated bool, err error) {
	if limit <= 0 {
		limit = 20000
	}
	ignore := append(append([]string{}, DefaultIgnore...), extra...)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && ignored(name, path, root, ignore) {
				return fs.SkipDir
			}
			if d.Type()&os.ModeSymlink != 0 && path != root {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				return nil
			}
		}
		if ignored(name, path, root, ignore) || skippedExt(name) {
			return nil
		}
		files = append(files, path)
		if len(files) >= limit {
			truncated = true
			return fs.SkipAll
		}
		return nil
	})
	return files, truncated, err
}

// Search scans files for opt and stops after maxFiles or maxHits.
func Search(ctx context.Context, files []string, opt search.Options, maxHits int) ([]Hit, error) {
	if maxHits <= 0 {
		maxHits = 500
	}
	var hits []Hit
	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return hits, err
		}
		info, err := os.Stat(path)
		if err != nil || info.Size() > 1<<20 {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil || buffer.IsBinary(data) {
			continue
		}
		text := strings.ReplaceAll(string(data), "\r\n", "\n")
		lines := strings.Split(text, "\n")
		matches, err := search.Find(lines, opt)
		if err != nil {
			return hits, err
		}
		for _, m := range matches {
			preview := lines[m.Start.Line]
			if len(preview) > 200 {
				preview = preview[:200]
			}
			hits = append(hits, Hit{
				Path: path,
				Line: m.Start.Line + 1,
				Col:  m.Start.Col + 1,
				Text: preview,
			})
			if len(hits) >= maxHits {
				return hits, nil
			}
		}
	}
	return hits, nil
}

func ignored(name, path, root string, pats []string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = name
	}
	rel = filepath.ToSlash(rel)
	base := filepath.Base(path)
	for _, p := range pats {
		p = filepath.ToSlash(p)
		if p == "" {
			continue
		}
		if p == base || p == name {
			return true
		}
		if ok, _ := filepath.Match(p, base); ok {
			return true
		}
		if ok, _ := filepath.Match(p, rel); ok {
			return true
		}
		if strings.HasSuffix(p, "/") && strings.HasPrefix(rel, strings.TrimSuffix(p, "/")) {
			return true
		}
	}
	return false
}

func skippedExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".exe", ".dll", ".so", ".dylib", ".a", ".o", ".obj", ".pyc", ".class", ".jar", ".zip", ".gz", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".pdf", ".woff", ".woff2":
		return true
	default:
		return false
	}
}
