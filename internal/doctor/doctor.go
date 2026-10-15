// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package doctor reports what the terminal can and cannot deliver.
// It does not claim that a probe succeeded when the program only saw an
// environment variable.
package doctor

import (
	"fmt"
	"os"
	"strings"
)

// Report is a point-in-time description of the session.
type Report struct {
	Width     int
	Height    int
	Colors    int
	Term      string
	ColorTerm string
	SSH       bool
	Tmux      bool
	Screen    bool
	Locale    string
	Notes     []string
}

// Collect reads the environment and the dimensions supplied by the caller.
// colors is the palette size reported by the terminal library, or 0 when unknown.
func Collect(width, height, colors int) Report {
	r := Report{
		Width:     width,
		Height:    height,
		Colors:    colors,
		Term:      os.Getenv("TERM"),
		ColorTerm: os.Getenv("COLORTERM"),
		SSH:       os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CLIENT") != "",
		Tmux:      os.Getenv("TMUX") != "",
		Screen:    os.Getenv("STY") != "",
		Locale:    firstEnv("LC_ALL", "LC_CTYPE", "LANG"),
	}
	if r.Term == "" {
		r.Notes = append(r.Notes, "TERM is empty. Color and keyboard behavior may fall back to the minimum.")
	}
	if r.SSH {
		r.Notes = append(r.Notes, "SSH session. The keyboard is the remote terminal's byte stream, not the Mac's Command key. Command+C, Command+X, and Command+V are not available here.")
	} else {
		r.Notes = append(r.Notes, "Command/Super shortcuts are not claimed. Terminals do not deliver them as distinct events to a remote or local TUI in a way Axam Studio can rely on.")
	}
	r.Notes = append(r.Notes, keyboardNote(r))
	r.Notes = append(r.Notes, colorNote(r))
	r.Notes = append(r.Notes, "Mouse reporting is requested (SGR). A terminal can ignore it. Selection inside axam uses the editor selection, not the terminal's mouse selection.")
	r.Notes = append(r.Notes, "Clipboard: axam always keeps an internal clipboard. Copy also sends OSC 52 when the terminal library can. That is not confirmation the system clipboard changed. Over SSH, OSC 52 reaches the local terminal only if that terminal allows it. Set clipboard = \"xclip\" or \"wl-copy\" only on a local Linux desktop where those tools exist.")
	r.Notes = append(r.Notes, "Ctrl+Shift+letter is distinguishable only with modifyOtherKeys or the Kitty keyboard protocol. Otherwise use the Ctrl+K chords (Ctrl+K Ctrl+S saves as, Ctrl+K Ctrl+P opens the palette).")
	r.Notes = append(r.Notes, "Ctrl+H is Backspace on terminals that send byte 0x08 for Backspace. Replace is also bound to Ctrl+K Ctrl+F's neighbor: use the palette or Ctrl+H only where it arrives as Ctrl+H. Find is Ctrl+F. A reliable replace entry is the command palette.")
	if r.Tmux {
		r.Notes = append(r.Notes, "tmux is wrapping the session. Enable extended keys if Shift+arrows do nothing: set -g extended-keys on, and set -as terminal-features ',xterm*:extkeys'.")
	}
	if r.Screen {
		r.Notes = append(r.Notes, "GNU Screen filters many modified keys. Prefer tmux, or bind the Ctrl+K alternatives.")
	}
	if !strings.Contains(strings.ToUpper(r.Locale), "UTF-8") && r.Locale != "" {
		r.Notes = append(r.Notes, "Locale does not mention UTF-8. Cyrillic and emoji need a UTF-8 locale.")
	}
	if r.Width > 0 && (r.Width < 80 || r.Height < 24) {
		r.Notes = append(r.Notes, fmt.Sprintf("Size is %dx%d. The supported minimum layout is 80x24.", r.Width, r.Height))
	}
	return r
}

func keyboardNote(r Report) string {
	term := strings.ToLower(r.Term)
	switch {
	case strings.Contains(term, "kitty") || strings.Contains(term, "wezterm"):
		return "Keyboard: TERM looks like Kitty or WezTerm. axam asks for the Kitty protocol and modifyOtherKeys. Ctrl+Shift+letter can work when the emulator is configured to send it."
	case strings.Contains(term, "xterm") || strings.Contains(term, "tmux") || strings.Contains(term, "screen"):
		return "Keyboard: tcell requests modifyOtherKeys and Kitty CSI-u on xterm-like terminals. Legacy terminals still collapse Ctrl+Shift+letter into Ctrl+letter. Shift+arrow is widely available as CSI 1;2A."
	default:
		return "Keyboard: protocol support was not identified from TERM. Assume legacy sequences. Shift+arrow, Ctrl+letter, and the Ctrl+K chords are the dependable set."
	}
}

func colorNote(r Report) string {
	switch {
	case strings.Contains(strings.ToLower(r.ColorTerm), "truecolor") || strings.Contains(strings.ToLower(r.ColorTerm), "24bit") || r.Colors > 256:
		return "Color: true color (or better than 256) is advertised. The dark theme uses it."
	case r.Colors >= 256 || strings.Contains(r.Term, "256color"):
		return "Color: 256-color palette. The theme is quantized."
	case r.Colors > 0:
		return fmt.Sprintf("Color: %d colors. The theme falls back to the 16-color set.", r.Colors)
	default:
		return "Color: palette size is unknown outside a live screen. axam degrades from true color to 256 to 16."
	}
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// String renders the report for a terminal or a document.
func (r Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "axam doctor — Axam Studio\n\n")
	if r.Width > 0 {
		fmt.Fprintf(&b, "Size:        %d×%d\n", r.Width, r.Height)
	} else {
		fmt.Fprintf(&b, "Size:        unknown (not a live screen)\n")
	}
	fmt.Fprintf(&b, "Colors:      %d\n", r.Colors)
	fmt.Fprintf(&b, "TERM:        %s\n", empty(r.Term))
	fmt.Fprintf(&b, "COLORTERM:   %s\n", empty(r.ColorTerm))
	fmt.Fprintf(&b, "Locale:      %s\n", empty(r.Locale))
	fmt.Fprintf(&b, "SSH:         %v\n", r.SSH)
	fmt.Fprintf(&b, "tmux:        %v\n", r.Tmux)
	fmt.Fprintf(&b, "screen:      %v\n", r.Screen)
	fmt.Fprintf(&b, "Mouse:       requested, not probed\n")
	fmt.Fprintf(&b, "Clipboard:   internal always; OSC 52 attempted on copy; not confirmed\n")
	b.WriteString("\n")
	for _, n := range r.Notes {
		b.WriteString("• ")
		b.WriteString(n)
		b.WriteByte('\n')
	}
	return b.String()
}

func empty(s string) string {
	if s == "" {
		return "(unset)"
	}
	return s
}
