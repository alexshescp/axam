// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package clipboard moves text out of the editor only when that can be done
// honestly. The editor clipboard always works. OSC 52 is attempted by the
// terminal layer and is not reported as confirmed. External tools run only
// when the user sets clipboard to "xclip" or "wl-copy".
package clipboard

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
)

// Clip is the in-editor clipboard.
type Clip struct {
	Text string
	Line bool
}

// EncodeOSC52 builds an OSC 52 copy sequence. Terminals may ignore it.
func EncodeOSC52(text string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(text))
	return "\033]52;c;" + enc + "\033\\"
}

// WriteExternal copies text with an explicit local tool.
// mode "internal", "auto", and "osc52" do not spawn a process.
func WriteExternal(mode, text string) error {
	switch mode {
	case "", "auto", "internal", "osc52":
		return nil
	case "xclip":
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	case "wl-copy":
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	default:
		return fmt.Errorf("unknown clipboard mode %q", mode)
	}
}

// ReadExternal reads a local tool clipboard. It returns an error for modes
// that do not have a synchronous local clipboard.
func ReadExternal(mode string) (string, error) {
	var cmd *exec.Cmd
	switch mode {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	case "wl-copy":
		cmd = exec.Command("wl-paste", "-n")
	default:
		return "", fmt.Errorf("no local clipboard for mode %q", mode)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
