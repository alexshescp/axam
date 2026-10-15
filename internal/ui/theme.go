// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package ui holds the Axam Studio color themes. The layout itself lives in
// the app package so drawing stays next to editor state.
package ui

import "github.com/gdamore/tcell/v2"

// Theme is a cell palette. Hex colors are used when the terminal has a
// 256-color or true-color palette. Fewer colors use the 16-color names.
type Theme struct {
	Bg, Fg, Current, Sel          tcell.Color
	LineNum, LineNumCur, Gutter   tcell.Color
	TabBg, TabActive, TabFg       tcell.Color
	Accent, Dirty                 tcell.Color
	StatusBg, StatusFg            tcell.Color
	Keyword, String, Comment      tcell.Color
	Number, Func, TypeName        tcell.Color
	Match, MatchCur               tcell.Color
	PromptBg, PromptSel, Error    tcell.Color
	Whitespace, Bracket, Explorer tcell.Color
}

// For picks a palette. colors is screen.Colors().
func For(name string, colors int) Theme {
	if colors > 0 && colors < 256 {
		return basic(name)
	}
	switch name {
	case "light":
		return light()
	case "mono":
		return basic("mono")
	default:
		return dark()
	}
}

func dark() Theme {
	return Theme{
		Bg:         tcell.NewHexColor(0x1E2127),
		Fg:         tcell.NewHexColor(0xD7D7D7),
		Current:    tcell.NewHexColor(0x2A2E36),
		Sel:        tcell.NewHexColor(0x3D4F66),
		LineNum:    tcell.NewHexColor(0x6B7280),
		LineNumCur: tcell.NewHexColor(0xE6E6E6),
		Gutter:     tcell.NewHexColor(0x1E2127),
		TabBg:      tcell.NewHexColor(0x181A1F),
		TabActive:  tcell.NewHexColor(0x1E2127),
		TabFg:      tcell.NewHexColor(0xC8C8C8),
		Accent:     tcell.NewHexColor(0xD7A15F),
		Dirty:      tcell.NewHexColor(0xE5C07B),
		StatusBg:   tcell.NewHexColor(0x181A1F),
		StatusFg:   tcell.NewHexColor(0xC8C8C8),
		Keyword:    tcell.NewHexColor(0xC792EA),
		String:     tcell.NewHexColor(0xC3E88D),
		Comment:    tcell.NewHexColor(0x6B7280),
		Number:     tcell.NewHexColor(0xF78C6C),
		Func:       tcell.NewHexColor(0x82AAFF),
		TypeName:   tcell.NewHexColor(0xFFCB6B),
		Match:      tcell.NewHexColor(0x5C4A1F),
		MatchCur:   tcell.NewHexColor(0x8A5A12),
		PromptBg:   tcell.NewHexColor(0x252830),
		PromptSel:  tcell.NewHexColor(0x3D4F66),
		Error:      tcell.NewHexColor(0xF07178),
		Whitespace: tcell.NewHexColor(0x3A3F4B),
		Bracket:    tcell.NewHexColor(0xFFD580),
		Explorer:   tcell.NewHexColor(0xA8B0BD),
	}
}

func light() Theme {
	return Theme{
		Bg:         tcell.NewHexColor(0xF7F5F0),
		Fg:         tcell.NewHexColor(0x1F2328),
		Current:    tcell.NewHexColor(0xEFEAE0),
		Sel:        tcell.NewHexColor(0xD3E3F5),
		LineNum:    tcell.NewHexColor(0x8A8175),
		LineNumCur: tcell.NewHexColor(0x1F2328),
		Gutter:     tcell.NewHexColor(0xF7F5F0),
		TabBg:      tcell.NewHexColor(0xE7E2D8),
		TabActive:  tcell.NewHexColor(0xF7F5F0),
		TabFg:      tcell.NewHexColor(0x1F2328),
		Accent:     tcell.NewHexColor(0xB45309),
		Dirty:      tcell.NewHexColor(0xA15C07),
		StatusBg:   tcell.NewHexColor(0xE7E2D8),
		StatusFg:   tcell.NewHexColor(0x1F2328),
		Keyword:    tcell.NewHexColor(0x7C3AED),
		String:     tcell.NewHexColor(0x067A4B),
		Comment:    tcell.NewHexColor(0x8A8175),
		Number:     tcell.NewHexColor(0xC2410C),
		Func:       tcell.NewHexColor(0x1D4ED8),
		TypeName:   tcell.NewHexColor(0xA16207),
		Match:      tcell.NewHexColor(0xFDE68A),
		MatchCur:   tcell.NewHexColor(0xF59E0B),
		PromptBg:   tcell.NewHexColor(0xEFEAE0),
		PromptSel:  tcell.NewHexColor(0xD3E3F5),
		Error:      tcell.NewHexColor(0xB91C1C),
		Whitespace: tcell.NewHexColor(0xD6D3D1),
		Bracket:    tcell.NewHexColor(0xB45309),
		Explorer:   tcell.NewHexColor(0x44403C),
	}
}

func basic(name string) Theme {
	bg, fg := tcell.ColorBlack, tcell.ColorWhite
	if name == "light" {
		bg, fg = tcell.ColorWhite, tcell.ColorBlack
	}
	return Theme{
		Bg: bg, Fg: fg, Current: tcell.ColorNavy, Sel: tcell.ColorBlue,
		LineNum: tcell.ColorGray, LineNumCur: tcell.ColorWhite, Gutter: bg,
		TabBg: tcell.ColorBlack, TabActive: bg, TabFg: fg,
		Accent: tcell.ColorYellow, Dirty: tcell.ColorYellow,
		StatusBg: tcell.ColorBlack, StatusFg: tcell.ColorSilver,
		Keyword: tcell.ColorFuchsia, String: tcell.ColorGreen, Comment: tcell.ColorGray,
		Number: tcell.ColorRed, Func: tcell.ColorAqua, TypeName: tcell.ColorYellow,
		Match: tcell.ColorOlive, MatchCur: tcell.ColorYellow,
		PromptBg: tcell.ColorBlack, PromptSel: tcell.ColorBlue, Error: tcell.ColorRed,
		Whitespace: tcell.ColorGray, Bracket: tcell.ColorYellow, Explorer: tcell.ColorSilver,
	}
}
