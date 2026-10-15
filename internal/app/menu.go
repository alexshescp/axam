// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package app

import (
	"strings"

	"github.com/codersclub/axam/internal/commands"
	"github.com/codersclub/axam/internal/input"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

type menuItem struct {
	title   string
	command string
}

type menuGroup struct {
	title string
	items []menuItem
}

// bottomMenus is deliberately command-id data. Keyboard bindings, the
// command palette, and this menu all execute the same handlers.
func bottomMenus() []menuGroup {
	return []menuGroup{
		{title: "File", items: []menuItem{
			{"New File", commands.FileNew},
			{"Open File", commands.FileOpen},
			{"Save", commands.FileSave},
			{"Save As", commands.FileSaveAs},
			{"Close Tab", commands.FileClose},
			{"Reopen Closed Tab", commands.FileReopen},
			{"Reload from Disk", commands.FileReload},
			{"Quit", commands.FileQuit},
		}},
		{title: "Edit", items: []menuItem{
			{"Undo", commands.EditUndo},
			{"Redo", commands.EditRedo},
			{"Cut", commands.EditCut},
			{"Copy", commands.EditCopy},
			{"Paste", commands.EditPaste},
			{"Delete Line", commands.EditDeleteLine},
			{"Duplicate Line", commands.EditDuplicateLine},
			{"Move Line Up", commands.EditMoveLineUp},
			{"Move Line Down", commands.EditMoveLineDown},
			{"Indent", commands.EditIndent},
			{"Outdent", commands.EditOutdent},
			{"Toggle Comment", commands.EditComment},
		}},
		{title: "Selection", items: []menuItem{
			{"Select All", commands.EditSelectAll},
			{"Select Word", commands.EditSelectWord},
			{"Select Line", commands.EditSelectLine},
			{"Select Next Occurrence", commands.EditNextOccurrence},
			{"Select All Occurrences", commands.EditAllOccurrences},
			{"Add Cursor Above", commands.EditCursorAbove},
			{"Add Cursor Below", commands.EditCursorBelow},
		}},
		{title: "Find", items: []menuItem{
			{"Find", commands.FindOpen},
			{"Replace", commands.FindReplace},
			{"Find Next", commands.FindNext},
			{"Find Previous", commands.FindPrev},
			{"Replace Current Match", commands.FindReplaceOne},
			{"Replace All", commands.FindReplaceAll},
			{"Search in Project", commands.FindProject},
		}},
		{title: "View", items: []menuItem{
			{"Toggle Explorer", commands.NavExplorer},
			{"Toggle Split", commands.NavSplit},
			{"Toggle Word Wrap", commands.ViewWrap},
			{"Toggle Line Numbers", commands.ViewLineNumbers},
			{"Toggle Fold", commands.ViewFold},
			{"Command Palette", commands.ViewPalette},
		}},
		{title: "Goto", items: []menuItem{
			{"Go to Line", commands.FindGoto},
			{"Quick Open", commands.NavQuickOpen},
			{"Next Tab", commands.NavNextTab},
			{"Previous Tab", commands.NavPrevTab},
			{"Focus Explorer", commands.NavFocusExpl},
			{"Focus Editor", commands.NavFocusEditor},
		}},
		{title: "Help", items: []menuItem{
			{"Keyboard Shortcuts", commands.ViewShortcuts},
			{"Open Keymap Reference", commands.ViewKeymap},
			{"Terminal Compatibility Report", commands.ViewDoctor},
			{"Open Configuration", commands.ConfigOpen},
			{"Reload Configuration", commands.ConfigReload},
		}},
	}
}

func (a *App) openMenu(group int) {
	groups := bottomMenus()
	if len(groups) == 0 {
		return
	}
	if group < 0 {
		group = 0
	}
	if group >= len(groups) {
		group = len(groups) - 1
	}
	a.mode = modeMenu
	a.menu.group = group
	a.menu.item = 0
	a.chord = nil
	a.setStatus("Menu: arrows navigate, Enter runs, Esc closes")
}

func (a *App) closeMenu() {
	a.mode = modeEdit
	a.setStatus("")
}

func (a *App) onMenuKey(ev *tcell.EventKey) {
	res := input.Normalize(ev)
	groups := bottomMenus()
	if len(groups) == 0 {
		a.closeMenu()
		return
	}
	switch {
	case res.Key.Name == "escape" || res.Key.Name == "f10":
		a.closeMenu()
	case res.Key.Name == "left":
		a.menu.group = (a.menu.group - 1 + len(groups)) % len(groups)
		a.menu.item = 0
	case res.Key.Name == "right" || res.Key.Name == "tab":
		a.menu.group = (a.menu.group + 1) % len(groups)
		a.menu.item = 0
	case res.Key.Name == "up":
		items := groups[a.menu.group].items
		a.menu.item = (a.menu.item - 1 + len(items)) % len(items)
	case res.Key.Name == "down":
		items := groups[a.menu.group].items
		a.menu.item = (a.menu.item + 1) % len(items)
	case res.Key.Name == "home":
		a.menu.item = 0
	case res.Key.Name == "end":
		a.menu.item = len(groups[a.menu.group].items) - 1
	case res.Key.Name == "enter":
		a.runMenuItem(a.menu.group, a.menu.item)
	default:
		if res.Insert != "" {
			needle := strings.ToLower(res.Insert)
			for i, group := range groups {
				if strings.HasPrefix(strings.ToLower(group.title), needle) {
					a.menu.group = i
					a.menu.item = 0
					break
				}
			}
		}
	}
}

func (a *App) runMenuItem(group, item int) {
	groups := bottomMenus()
	if group < 0 || group >= len(groups) || item < 0 || item >= len(groups[group].items) {
		return
	}
	command := groups[group].items[item].command
	a.mode = modeEdit
	a.setStatus("")
	a.exec(command)
}

func (a *App) menuGroupAt(x int) int {
	pos := 0
	for i, group := range bottomMenus() {
		width := uniseg.StringWidth(" " + group.title + " ")
		if x >= pos && x < pos+width {
			return i
		}
		pos += width
	}
	return -1
}

func (a *App) menuGroupX(group int) int {
	x := 0
	groups := bottomMenus()
	for i := 0; i < group && i < len(groups); i++ {
		x += uniseg.StringWidth(" " + groups[i].title + " ")
	}
	return x
}

func (a *App) menuShortcut(command string) string {
	chords := a.keys.ChordsFor(command)
	if len(chords) == 0 {
		return ""
	}
	return chords[0]
}

func (a *App) menuPopupBox(v view) cellBox {
	groups := bottomMenus()
	if a.mode != modeMenu || a.menu.group < 0 || a.menu.group >= len(groups) {
		return cellBox{}
	}
	group := groups[a.menu.group]
	width := 0
	for _, item := range group.items {
		label := item.title
		if chord := a.menuShortcut(item.command); chord != "" {
			label += "  " + chord
		}
		if n := uniseg.StringWidth(label); n > width {
			width = n
		}
	}
	width += 4
	if width > v.w {
		width = v.w
	}
	height := len(group.items)
	if height > v.menu {
		height = v.menu
	}
	x := a.menuGroupX(a.menu.group)
	if x+width > v.w {
		x = v.w - width
	}
	if x < 0 {
		x = 0
	}
	y := v.menu - height
	if y < 1 {
		y = 1
	}
	return cellBox{x: x, y: y, w: width, h: height}
}

func (a *App) drawBottomMenu(v view) {
	base := style(a.theme.TabFg, a.theme.TabBg)
	a.fill(0, v.menu, v.w, 1, base)
	x := 0
	for i, group := range bottomMenus() {
		label := " " + group.title + " "
		fg, bg := a.theme.TabFg, a.theme.TabBg
		if a.mode == modeMenu && i == a.menu.group {
			fg, bg = a.theme.Fg, a.theme.PromptSel
		}
		x = a.drawText(x, v.menu, v.w, label, style(fg, bg))
	}
	if x < v.w {
		hint := "F10 Menu"
		hw := uniseg.StringWidth(hint)
		if x+1 < v.w-hw {
			a.drawText(v.w-hw-1, v.menu, v.w, hint, style(a.theme.LineNum, a.theme.TabBg))
		}
	}
	if a.mode == modeMenu {
		a.drawMenuPopup(v)
	}
}

func (a *App) drawMenuPopup(v view) {
	groups := bottomMenus()
	if a.menu.group < 0 || a.menu.group >= len(groups) {
		return
	}
	box := a.menuPopupBox(v)
	group := groups[a.menu.group]
	start := len(group.items) - box.h
	if start < 0 {
		start = 0
	}
	for row := 0; row < box.h; row++ {
		idx := start + row
		item := group.items[idx]
		fg, bg := a.theme.Fg, a.theme.PromptBg
		if idx == a.menu.item {
			bg = a.theme.PromptSel
		}
		st := style(fg, bg)
		a.fill(box.x, box.y+row, box.w, 1, st)
		a.drawText(box.x+1, box.y+row, box.x+box.w, item.title, st)
		if chord := a.menuShortcut(item.command); chord != "" {
			cw := uniseg.StringWidth(chord)
			if cw+2 < box.w {
				a.drawText(box.x+box.w-cw-1, box.y+row, box.x+box.w, chord, style(a.theme.LineNum, bg))
			}
		}
	}
}

func (a *App) menuHit(x, y int, v view) (hit, bool) {
	if y == v.menu {
		group := a.menuGroupAt(x)
		if group >= 0 {
			return hit{kind: hitMenuGroup, menuGroup: group}, true
		}
		return hit{}, false
	}
	if a.mode != modeMenu {
		return hit{}, false
	}
	box := a.menuPopupBox(v)
	if x < box.x || x >= box.x+box.w || y < box.y || y >= box.y+box.h {
		return hit{}, false
	}
	group := bottomMenus()[a.menu.group]
	start := len(group.items) - box.h
	if start < 0 {
		start = 0
	}
	return hit{kind: hitMenuItem, menuGroup: a.menu.group, menuItem: start + y - box.y}, true
}

func menuCommandsKnown() bool {
	for _, group := range bottomMenus() {
		for _, item := range group.items {
			if !commands.Known(item.command) {
				return false
			}
		}
	}
	return true
}
