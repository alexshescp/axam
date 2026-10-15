// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package commands is the single catalog of editor command ids.
// Key bindings, the command palette, and tests all use these ids.
package commands

// Command ids. Profiles bind keys to these strings; they do not embed behavior.
const (
	FileNew    = "file.new"
	FileOpen   = "file.open"
	FileSave   = "file.save"
	FileSaveAs = "file.saveAs"
	FileClose  = "file.close"
	FileReopen = "file.reopen"
	FileQuit   = "file.quit"
	FileReload = "file.reload"

	EditUndo            = "edit.undo"
	EditRedo            = "edit.redo"
	EditCut             = "edit.cut"
	EditCopy            = "edit.copy"
	EditPaste           = "edit.paste"
	EditSelectAll       = "edit.selectAll"
	EditSelectWord      = "edit.selectWord"
	EditSelectLine      = "edit.selectLine"
	EditDeleteLine      = "edit.deleteLine"
	EditDuplicateLine   = "edit.duplicateLine"
	EditMoveLineUp      = "edit.moveLineUp"
	EditMoveLineDown    = "edit.moveLineDown"
	EditIndent          = "edit.indent"
	EditOutdent         = "edit.outdent"
	EditComment         = "edit.comment"
	EditNewline         = "edit.newline"
	EditBackspace       = "edit.backspace"
	EditDelete          = "edit.delete"
	EditDeleteWordLeft  = "edit.deleteWordLeft"
	EditDeleteWordRight = "edit.deleteWordRight"
	EditNextOccurrence  = "edit.nextOccurrence"
	EditAllOccurrences  = "edit.allOccurrences"
	EditCursorAbove     = "edit.cursorAbove"
	EditCursorBelow     = "edit.cursorBelow"
	EditInsertLineBelow = "edit.insertLineBelow"
	EditInsertLineAbove = "edit.insertLineAbove"

	MoveLeft            = "move.left"
	MoveRight           = "move.right"
	MoveUp              = "move.up"
	MoveDown            = "move.down"
	MoveHome            = "move.home"
	MoveEnd             = "move.end"
	MoveWordLeft        = "move.wordLeft"
	MoveWordRight       = "move.wordRight"
	MoveDocStart        = "move.docStart"
	MoveDocEnd          = "move.docEnd"
	MovePageUp          = "move.pageUp"
	MovePageDown        = "move.pageDown"
	MoveSelectLeft      = "move.selectLeft"
	MoveSelectRight     = "move.selectRight"
	MoveSelectUp        = "move.selectUp"
	MoveSelectDown      = "move.selectDown"
	MoveSelectHome      = "move.selectHome"
	MoveSelectEnd       = "move.selectEnd"
	MoveSelectWordLeft  = "move.selectWordLeft"
	MoveSelectWordRight = "move.selectWordRight"
	MoveSelectDocStart  = "move.selectDocStart"
	MoveSelectDocEnd    = "move.selectDocEnd"
	MoveSelectPageUp    = "move.selectPageUp"
	MoveSelectPageDown  = "move.selectPageDown"

	FindOpen       = "find.open"
	FindReplace    = "find.replace"
	FindNext       = "find.next"
	FindPrev       = "find.prev"
	FindReplaceOne = "find.replaceOne"
	FindReplaceAll = "find.replaceAll"
	FindGoto       = "find.goto"
	FindProject    = "find.project"

	NavQuickOpen   = "nav.quickOpen"
	NavNextTab     = "nav.nextTab"
	NavPrevTab     = "nav.prevTab"
	NavExplorer    = "nav.explorer"
	NavSplit       = "nav.split"
	NavFocusExpl   = "nav.focusExplorer"
	NavFocusEditor = "nav.focusEditor"

	ViewWrap        = "view.wrap"
	ViewLineNumbers = "view.lineNumbers"
	ViewShortcuts   = "view.shortcuts"
	ViewPalette     = "view.palette"
	ViewKeymap      = "view.keymap"
	ViewDoctor      = "view.doctor"
	ViewCancel      = "view.cancel"
	ViewFold        = "view.fold"

	ConfigOpen   = "config.open"
	ConfigReload = "config.reload"
)

// Spec is one palette entry.
type Spec struct {
	ID       string
	Title    string
	Category string
}

// Catalog is the palette order.
func Catalog() []Spec {
	return []Spec{
		{FileNew, "New File", "File"},
		{FileOpen, "Open File", "File"},
		{FileSave, "Save", "File"},
		{FileSaveAs, "Save As", "File"},
		{FileClose, "Close Tab", "File"},
		{FileReopen, "Reopen Closed Tab", "File"},
		{FileReload, "Reload from Disk", "File"},
		{FileQuit, "Quit", "File"},
		{EditUndo, "Undo", "Edit"},
		{EditRedo, "Redo", "Edit"},
		{EditCut, "Cut", "Edit"},
		{EditCopy, "Copy", "Edit"},
		{EditPaste, "Paste", "Edit"},
		{EditSelectAll, "Select All", "Edit"},
		{EditSelectWord, "Select Word", "Edit"},
		{EditSelectLine, "Select Line", "Edit"},
		{EditDeleteLine, "Delete Line", "Edit"},
		{EditDuplicateLine, "Duplicate Line", "Edit"},
		{EditMoveLineUp, "Move Line Up", "Edit"},
		{EditMoveLineDown, "Move Line Down", "Edit"},
		{EditIndent, "Indent", "Edit"},
		{EditOutdent, "Outdent", "Edit"},
		{EditComment, "Toggle Comment", "Edit"},
		{EditNextOccurrence, "Select Next Occurrence", "Edit"},
		{EditAllOccurrences, "Select All Occurrences", "Edit"},
		{EditCursorAbove, "Add Cursor Above", "Edit"},
		{EditCursorBelow, "Add Cursor Below", "Edit"},
		{EditInsertLineAbove, "Insert Line Above", "Edit"},
		{EditInsertLineBelow, "Insert Line Below", "Edit"},
		{FindOpen, "Find", "Find"},
		{FindReplace, "Replace", "Find"},
		{FindNext, "Find Next", "Find"},
		{FindPrev, "Find Previous", "Find"},
		{FindReplaceOne, "Replace Current Match", "Find"},
		{FindReplaceAll, "Replace All", "Find"},
		{FindGoto, "Go to Line", "Find"},
		{FindProject, "Search in Project", "Find"},
		{NavQuickOpen, "Quick Open", "Navigation"},
		{NavNextTab, "Next Tab", "Navigation"},
		{NavPrevTab, "Previous Tab", "Navigation"},
		{NavExplorer, "Toggle Explorer", "Navigation"},
		{NavSplit, "Toggle Split", "Navigation"},
		{NavFocusExpl, "Focus Explorer", "Navigation"},
		{NavFocusEditor, "Focus Editor", "Navigation"},
		{ViewWrap, "Toggle Word Wrap", "View"},
		{ViewLineNumbers, "Toggle Line Numbers", "View"},
		{ViewFold, "Toggle Fold", "View"},
		{ViewShortcuts, "Keyboard Shortcuts", "View"},
		{ViewPalette, "Command Palette", "View"},
		{ViewKeymap, "Open Keymap Reference", "View"},
		{ViewDoctor, "Terminal Compatibility Report", "View"},
		{ConfigOpen, "Open Configuration", "Settings"},
		{ConfigReload, "Reload Configuration", "Settings"},
	}
}

// Known reports whether id is in the catalog.
func Known(id string) bool {
	for _, s := range Catalog() {
		if s.ID == id {
			return true
		}
	}
	// Movement commands are catalogued separately so the palette stays short,
	// but they are still real commands.
	switch id {
	case MoveLeft, MoveRight, MoveUp, MoveDown, MoveHome, MoveEnd,
		MoveWordLeft, MoveWordRight, MoveDocStart, MoveDocEnd, MovePageUp, MovePageDown,
		MoveSelectLeft, MoveSelectRight, MoveSelectUp, MoveSelectDown,
		MoveSelectHome, MoveSelectEnd, MoveSelectWordLeft, MoveSelectWordRight,
		MoveSelectDocStart, MoveSelectDocEnd, MoveSelectPageUp, MoveSelectPageDown,
		EditNewline, EditBackspace, EditDelete, EditDeleteWordLeft, EditDeleteWordRight,
		ViewCancel:
		return true
	default:
		return false
	}
}
