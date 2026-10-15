// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package editor

import "github.com/codersclub/axam/internal/text"

// Undo lives next to the editor because a transaction has to restore carets,
// not only bytes. A separate undo engine would re-encode the same snapshots.

const historyLimit = 1000

type edit struct {
	start    text.Pos
	deleted  string
	inserted string
}

type tx struct {
	edits  []edit
	before []Cursor
	after  []Cursor
}

type stack struct {
	undo  []tx
	redo  []tx
	clean int
}

func (s *stack) markClean() { s.clean = len(s.undo) }

func (s *stack) dirty() bool { return len(s.undo) != s.clean }

func (s *stack) push(t tx, coalesce bool) {
	if coalesce && len(s.undo) > 0 {
		last := &s.undo[len(s.undo)-1]
		if canCoalesce(*last, t) {
			last.edits[0].inserted += t.edits[0].inserted
			last.after = t.after
			s.redo = nil
			return
		}
	}
	s.undo = append(s.undo, t)
	if len(s.undo) > historyLimit {
		drop := len(s.undo) - historyLimit
		s.undo = s.undo[drop:]
		s.clean -= drop
		if s.clean < 0 {
			s.clean = -1
		}
	}
	s.redo = nil
}

func canCoalesce(prev, next tx) bool {
	if len(prev.edits) != 1 || len(next.edits) != 1 {
		return false
	}
	a, b := prev.edits[0], next.edits[0]
	if a.deleted != "" || b.deleted != "" {
		return false
	}
	if stringsHasNL(a.inserted) || stringsHasNL(b.inserted) {
		return false
	}
	at := text.Advance(a.start, a.inserted)
	return at.Eq(b.start)
}

func stringsHasNL(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}

func (s *stack) popUndo() (tx, bool) {
	if len(s.undo) == 0 {
		return tx{}, false
	}
	t := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	s.redo = append(s.redo, t)
	return t, true
}

func (s *stack) popRedo() (tx, bool) {
	if len(s.redo) == 0 {
		return tx{}, false
	}
	t := s.redo[len(s.redo)-1]
	s.redo = s.redo[:len(s.redo)-1]
	s.undo = append(s.undo, t)
	return t, true
}
