// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package files loads and saves documents without truncating the original
// before a replacement is ready.
package files

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode"

	"github.com/codersclub/axam/internal/buffer"
)

const (
	// SoftHighlight is the size above which syntax highlighting is disabled.
	SoftHighlight = 2 << 20
	// HardLimit refuses a file unless force is set.
	HardLimit = 256 << 20
)

var (
	ErrBinary    = errors.New("binary file")
	ErrDirectory = errors.New("directory")
	ErrTooLarge  = errors.New("file is larger than 256 MiB")
)

// Loaded is a file mapped into a buffer plus the metadata needed to save it.
type Loaded struct {
	Path     string
	Buffer   *buffer.Buffer
	Encoding string
	BOM      bool
	Perm     os.FileMode
	ModTime  time.Time
	Size     int64
	Hash     [32]byte
	ReadOnly bool
	Symlink  string
	TooBig   bool
}

// Read opens path. Directories return ErrDirectory. Binary files return
// ErrBinary unless force is set, in which case they still return ErrBinary
// because writing them back would corrupt them. Huge files return ErrTooLarge
// unless force is set.
func Read(path string, force bool) (*Loaded, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	link := ""
	if info.Mode()&os.ModeSymlink != 0 {
		if dest, err := os.Readlink(path); err == nil {
			link = dest
		}
		info, err = os.Stat(path)
		if err != nil {
			return nil, err
		}
	}
	if info.IsDir() {
		return nil, ErrDirectory
	}
	if info.Size() > HardLimit && !force {
		return nil, fmt.Errorf("%w (%d bytes)", ErrTooLarge, info.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if buffer.IsBinary(data) {
		return nil, fmt.Errorf("%w: %s", ErrBinary, path)
	}
	ld, err := buffer.FromBytes(data)
	if err != nil {
		return nil, err
	}
	ro := !writable(path)
	return &Loaded{
		Path:     path,
		Buffer:   ld.Buffer,
		Encoding: ld.Encoding,
		BOM:      ld.BOM,
		Perm:     info.Mode().Perm(),
		ModTime:  info.ModTime(),
		Size:     info.Size(),
		Hash:     ld.Hash,
		ReadOnly: ro,
		Symlink:  link,
		TooBig:   info.Size() > SoftHighlight,
	}, nil
}

// DiskHash reads the current content hash. ok is false when the file is gone.
func DiskHash(path string) (sum [32]byte, mod time.Time, size int64, ok bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return sum, time.Time{}, 0, false, nil
		}
		return sum, time.Time{}, 0, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return sum, time.Time{}, 0, false, err
	}
	return sha256.Sum256(data), info.ModTime(), info.Size(), true, nil
}

// Changed reports whether the file on disk differs from the stored hash.
func Changed(path string, prev [32]byte, prevMod time.Time, prevSize int64) (bool, [32]byte, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, [32]byte{}, time.Time{}, nil
		}
		return false, [32]byte{}, time.Time{}, err
	}
	if info.ModTime().Equal(prevMod) && info.Size() == prevSize {
		return false, prev, prevMod, nil
	}
	sum, mod, _, ok, err := DiskHash(path)
	if err != nil || !ok {
		return false, sum, mod, err
	}
	return sum != prev, sum, mod, nil
}

// SaveAtomic replaces path by writing a temporary sibling, syncing it, and
// renaming it into place. The original is not truncated first. Symlinks are
// followed so the link itself is kept and the target is replaced.
func SaveAtomic(path string, data []byte, perm os.FileMode) error {
	target := path
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("symlink: %w", err)
		}
		target = resolved
		if st, err := os.Stat(target); err == nil {
			perm = st.Mode().Perm()
		}
	}
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".axam-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if perm != 0 {
		if err := os.Chmod(tmpName, perm); err != nil {
			// Some filesystems refuse chmod. Keep going; the bytes are what matter.
			_ = err
		}
	}
	if err := replaceFile(tmpName, target); err != nil {
		return err
	}
	ok = true
	return nil
}

// Backup copies data into dir/backups. A failure here must not block a save
// by itself; the caller decides.
func Backup(dir, srcName string, data []byte) (string, error) {
	destDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s-%s", filepath.Base(srcName), time.Now().Format("20060102-150405"))
	path := filepath.Join(destDir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// DetectIndent guesses tab width and whether the file uses spaces.
func DetectIndent(lines []string) (size int, spaces bool, ok bool) {
	tabs, spacesN := 0, 0
	var indents []int
	limit := len(lines)
	if limit > 400 {
		limit = 400
	}
	for i := 0; i < limit; i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		if line[0] == '\t' {
			tabs++
			continue
		}
		n := 0
		for n < len(line) && line[n] == ' ' {
			n++
		}
		if n > 0 && n < len(line) && !unicode.IsSpace(rune(line[n])) {
			spacesN++
			indents = append(indents, n)
		}
	}
	if tabs == 0 && spacesN == 0 {
		return 0, false, false
	}
	if tabs > spacesN {
		return 4, false, true
	}
	width := gcdIndent(indents)
	if width < 1 {
		width = 4
	}
	if width > 8 {
		width = 4
	}
	return width, true, true
}

func gcdIndent(ns []int) int {
	g := 0
	for _, n := range ns {
		g = gcd(g, n)
	}
	return g
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func writable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
