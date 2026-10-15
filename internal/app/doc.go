// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codersclub/axam/internal/buffer"
	"github.com/codersclub/axam/internal/config"
	"github.com/codersclub/axam/internal/editor"
	"github.com/codersclub/axam/internal/files"
	"github.com/codersclub/axam/internal/project"
	"github.com/codersclub/axam/internal/syntax"
)

func (a *App) newDoc(name string, ed *editor.Editor) *Document {
	ed.TabSize = a.cfg.TabSize
	ed.UseSpaces = a.cfg.InsertSpaces
	if a.cfg.DetectIndent {
		if size, spaces, ok := files.DetectIndent(ed.Buf.Lines()); ok {
			ed.TabSize = size
			ed.UseSpaces = spaces
		}
	}
	d := &Document{
		Name:     name,
		Ed:       ed,
		Encoding: buffer.EncUTF8,
		Lang:     "text",
		Folds:    map[int]int{},
		hl:       map[int]hlLine{},
	}
	return d
}

// OpenPath opens a file or, for a directory, sets the project root.
func (a *App) OpenPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	info, err := os.Stat(abs)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// A command-line path is a pending file. Opening it must not create
		// anything on disk; the first explicit Save performs the atomic write.
		parent := filepath.Dir(abs)
		parentInfo, parentErr := os.Stat(parent)
		if parentErr != nil {
			return fmt.Errorf("cannot create %s: parent directory: %w", abs, parentErr)
		}
		if !parentInfo.IsDir() {
			return fmt.Errorf("cannot create %s: parent %s is not a directory", abs, parent)
		}
		ed := editor.New(buffer.New())
		d := a.newDoc(filepath.Base(abs), ed)
		d.Path = abs
		d.Exists = false
		d.Lang = syntax.Language(abs)
		a.addDoc(d)
		a.setStatus("New file (not created until save): " + abs)
		return nil
	}
	if info.IsDir() {
		a.SetRoot(abs)
		a.setStatus("Project " + abs)
		readme := ""
		for _, name := range []string{"README.md", "README", "readme.md"} {
			p := filepath.Join(abs, name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				readme = p
				break
			}
		}
		if readme != "" {
			return a.OpenPath(readme)
		}
		return nil
	}
	ld, err := files.Read(abs, a.force)
	if err != nil {
		return err
	}
	ed := editor.New(ld.Buffer)
	d := a.newDoc(filepath.Base(abs), ed)
	d.Path = abs
	d.Encoding = ld.Encoding
	d.BOM = ld.BOM
	d.Perm = ld.Perm
	d.ModTime = ld.ModTime
	d.Hash = ld.Hash
	d.DiskSize = ld.Size
	d.ReadOnly = ld.ReadOnly
	d.Symlink = ld.Symlink
	d.Exists = true
	d.Large = ld.TooBig
	d.Lang = syntax.Language(abs)
	ed.ReadOnly = ld.ReadOnly
	a.addDoc(d)
	a.recent.RecentFiles = config.Remember(a.recent.RecentFiles, abs, a.cfg.RecentLimit)
	if a.root == "" {
		a.root = filesRoot(abs)
		a.refreshFiles()
	}
	msg := abs
	if ld.ReadOnly {
		msg += " (read-only)"
	}
	if ld.Buffer.Mixed() {
		msg += " (mixed line endings normalized to " + ld.Buffer.EOL().String() + ")"
	}
	if ld.TooBig {
		msg += " (large file, syntax highlighting off)"
	}
	a.setStatus(msg)
	return nil
}

// OpenString is used by tests and by reopening an untitled buffer.
func (a *App) OpenString(name, content string) *Document {
	ed := editor.New(buffer.FromString(content))
	d := a.newDoc(name, ed)
	d.Lang = syntax.Language(name)
	a.addDoc(d)
	return d
}

func (a *App) addDoc(d *Document) {
	a.docs = append(a.docs, d)
	if len(a.panes) == 0 {
		a.panes = []pane{{Doc: 0}}
	}
	a.panes[a.active].Doc = len(a.docs) - 1
	a.panes[a.active].Top = 0
	a.panes[a.active].Left = 0
	if a.screen != nil {
		a.screen.SetTitle(a.title())
	}
}

func filesRoot(path string) string {
	dir := filepath.Dir(path)
	return dir
}

func (a *App) saveActive(force bool) {
	d := a.doc()
	if d == nil {
		return
	}
	if d.Path == "" {
		a.promptPath("Save as", func(path string) {
			d.Path = path
			d.Name = filepath.Base(path)
			d.Lang = syntax.Language(path)
			a.writeDoc(d, true)
		})
		return
	}
	if !force && d.Exists {
		changed, _, _, err := files.Changed(d.Path, d.Hash, d.ModTime, d.DiskSize)
		if err != nil {
			a.setStatus(err.Error())
			return
		}
		if changed {
			a.confirm("Changed on disk. Overwrite "+d.Name+"?", func() { a.writeDoc(d, true) }, func() {
				a.setStatus("Save cancelled")
			})
			return
		}
	}
	a.writeDoc(d, false)
}

func listProject(root string, ignore []string) ([]string, bool, error) {
	return project.List(context.Background(), root, ignore, 20000)
}

func (a *App) writeDoc(d *Document, ignoreDisk bool) {
	_ = ignoreDisk
	data := buffer.Encode(d.Ed.Buf.Bytes(), d.Encoding, d.BOM)
	if d.Exists && a.cfg.Backup {
		if old, err := os.ReadFile(d.Path); err == nil {
			if _, err := files.Backup(config.Dir(), d.Path, old); err != nil {
				a.setStatus("Backup failed: " + err.Error())
			}
		}
	}
	perm := d.Perm
	if perm == 0 {
		perm = 0o644
	}
	if err := files.SaveAtomic(d.Path, data, perm); err != nil {
		a.setStatus("Save failed: " + err.Error())
		return
	}
	info, err := os.Stat(d.Path)
	if err == nil {
		d.ModTime = info.ModTime()
		d.Perm = info.Mode().Perm()
	}
	sum, mod, size, ok, err := files.DiskHash(d.Path)
	if err == nil && ok {
		d.Hash = sum
		d.ModTime = mod
		d.DiskSize = size
	}
	d.Exists = true
	d.Ed.MarkClean()
	a.recent.RecentFiles = config.Remember(a.recent.RecentFiles, d.Path, a.cfg.RecentLimit)
	a.setStatus("Saved " + d.Path)
	if a.screen != nil {
		a.screen.SetTitle(a.title())
	}
}

func (a *App) reloadDoc(d *Document) {
	if d == nil || d.Path == "" {
		return
	}
	if d.Ed.Dirty() {
		a.confirm("Discard unsaved changes in "+d.Name+"?", func() { a.reloadNow(d) }, func() {})
		return
	}
	a.reloadNow(d)
}

func (a *App) reloadNow(d *Document) {
	ld, err := files.Read(d.Path, true)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	d.Ed = editor.New(ld.Buffer)
	d.Ed.TabSize = a.cfg.TabSize
	d.Ed.UseSpaces = a.cfg.InsertSpaces
	d.Ed.ReadOnly = ld.ReadOnly
	d.Hash = ld.Hash
	d.ModTime = ld.ModTime
	d.ReadOnly = ld.ReadOnly
	d.Encoding = ld.Encoding
	d.BOM = ld.BOM
	d.hl = map[int]hlLine{}
	a.setStatus("Reloaded " + d.Path)
}

func (a *App) closeIndex(i int, force bool) {
	if i < 0 || i >= len(a.docs) {
		return
	}
	d := a.docs[i]
	if d.Ed.Dirty() && !force {
		a.confirm(fmt.Sprintf("Save changes to %s?", d.Name), func() {
			if d.Path == "" {
				a.promptPath("Save as", func(path string) {
					d.Path = path
					d.Name = filepath.Base(path)
					a.writeDoc(d, true)
					if !d.Ed.Dirty() {
						a.closeIndex(i, true)
					}
				})
				return
			}
			a.writeDoc(d, false)
			if !d.Ed.Dirty() {
				a.closeIndex(indexOf(a.docs, d), true)
			}
		}, func() { a.closeIndex(i, true) })
		return
	}
	if d.Path != "" || d.Ed.Buf.String() != "" {
		a.closed = append(a.closed, closedDoc{
			path: d.Path, name: d.Name, text: d.Ed.Buf.String(), eol: d.Ed.Buf.EOL(),
		})
		if len(a.closed) > 20 {
			a.closed = a.closed[len(a.closed)-20:]
		}
	}
	a.docs = append(a.docs[:i], a.docs[i+1:]...)
	for pi := range a.panes {
		switch {
		case a.panes[pi].Doc == i:
			if len(a.docs) == 0 {
				a.panes[pi].Doc = -1
			} else if i >= len(a.docs) {
				a.panes[pi].Doc = len(a.docs) - 1
			}
		case a.panes[pi].Doc > i:
			a.panes[pi].Doc--
		}
	}
	a.setStatus("Closed")
}

func indexOf(docs []*Document, d *Document) int {
	for i, x := range docs {
		if x == d {
			return i
		}
	}
	return -1
}

func (a *App) promptPath(title string, onOK func(string)) {
	a.mode = modePrompt
	a.prompt = promptBox{title: title, kind: "path", onOK: onOK}
}

func (a *App) confirm(title string, yes, no func()) {
	a.mode = modePrompt
	a.prompt = promptBox{title: title + "  [y]es [n]o [esc]", kind: "confirm", onYes: yes, onNo: no}
}

func (a *App) refreshFiles() {
	if a.root == "" {
		a.files = nil
		return
	}
	files, truncated, err := listProject(a.root, a.cfg.Ignore)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	a.files = files
	if truncated {
		a.setStatus(fmt.Sprintf("File list stopped at %d files", len(files)))
	}
}

func (a *App) loadExplorer() {
	a.expl.entries = nil
	a.expl.sel = 0
	if a.root == "" {
		return
	}
	a.expl.entries = []explEntry{{
		Path: a.root, Name: filepath.Base(a.root), Dir: true, Open: true, Depth: 0,
	}}
	a.expandExpl(0)
}

func (a *App) expandExpl(i int) {
	if i < 0 || i >= len(a.expl.entries) {
		return
	}
	e := a.expl.entries[i]
	if !e.Dir {
		return
	}
	ents, err := os.ReadDir(e.Path)
	if err != nil {
		a.setStatus(err.Error())
		return
	}
	var kids []explEntry
	for _, ent := range ents {
		name := ent.Name()
		if name == ".git" || name == "node_modules" || name == "vendor" {
			continue
		}
		kids = append(kids, explEntry{
			Path:  filepath.Join(e.Path, name),
			Name:  name,
			Depth: e.Depth + 1,
			Dir:   ent.IsDir(),
		})
	}
	// directories first
	dirs, files := splitExpl(kids)
	kids = append(dirs, files...)
	a.expl.entries[i].Open = true
	rest := append([]explEntry{}, a.expl.entries[i+1:]...)
	a.expl.entries = append(a.expl.entries[:i+1], kids...)
	a.expl.entries = append(a.expl.entries, rest...)
}

func splitExpl(in []explEntry) (dirs, files []explEntry) {
	for _, e := range in {
		if e.Dir {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}
	return dirs, files
}

func (a *App) collapseExpl(i int) {
	if i < 0 || i >= len(a.expl.entries) || !a.expl.entries[i].Open {
		return
	}
	depth := a.expl.entries[i].Depth
	j := i + 1
	for j < len(a.expl.entries) && a.expl.entries[j].Depth > depth {
		j++
	}
	a.expl.entries[i].Open = false
	a.expl.entries = append(a.expl.entries[:i+1], a.expl.entries[j:]...)
}

func quoteStatus(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}
