// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Command axam is Axam Studio, a terminal-native text editor.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alexshescp/axam/internal/app"
	"github.com/alexshescp/axam/internal/config"
	"github.com/alexshescp/axam/internal/doctor"
	"github.com/alexshescp/axam/internal/keymap"
	"github.com/gdamore/tcell/v2"
)

var version = "1.1.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("axam", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	profile := fs.String("profile", "", "keymap profile: axam, sublime, vscode, notepadpp, custom")
	configPath := fs.String("config", "", "path to config.toml")
	force := fs.Bool("force", false, "open files larger than 256 MiB (binary files are still refused)")
	noConfig := fs.Bool("no-config", false, "use defaults and do not read config.toml")
	fs.Usage = func() { fmt.Fprint(os.Stdout, usageText()) }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) > 0 {
		switch rest[0] {
		case "help", "--help", "-h":
			fmt.Print(usageText())
			return 0
		case "version", "--version", "-version":
			fmt.Printf("axam %s\nAxam Studio\nhttps://github.com/alexshescp/axam\nSupport: axam@codcl.com\nCopyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman\nLicense: GPL-3.0-or-later\n", version)
			return 0
		case "doctor":
			fmt.Print(doctor.Collect(0, 0, 0).String())
			return 0
		case "keymap":
			name := *profile
			if name == "" {
				name = "axam"
			}
			if len(rest) > 1 {
				name = rest[1]
			}
			return printKeymap(name)
		}
	}
	cfg := config.Defaults()
	if !*noConfig {
		path := *configPath
		if path == "" {
			path = config.Path()
		}
		loaded, _, err := config.Load(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "axam: %s\n", err)
			return 1
		}
		cfg = loaded
	}
	if *profile != "" {
		cfg.KeymapProfile = *profile
	}
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "axam: terminal: %s\n", err)
		return 1
	}
	session, err := app.New(screen, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axam: %s\n", err)
		return 1
	}
	if *force {
		session.Force()
	}
	for _, path := range rest {
		if err := session.OpenPath(path); err != nil {
			fmt.Fprintf(os.Stderr, "axam: %s: %s\n", path, err)
		}
	}
	if err := session.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "axam: %s\n", err)
		return 1
	}
	return 0
}

func printKeymap(name string) int {
	km, info, err := keymap.Load(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axam: %s\n", err)
		return 1
	}
	fmt.Printf("profile: %s (%s)\n", info.ID, info.Title)
	fmt.Printf("copy line when nothing is selected: %v\n", info.CopyLine)
	fmt.Printf("cut line when nothing is selected: %v\n", info.CutLine)
	fmt.Println("shortcut | command")
	for _, b := range km.Bindings() {
		env := "tested in the editor; delivery depends on the terminal"
		fmt.Printf("%s | %s | %s\n", b.Display(), b.Command, env)
	}
	fmt.Println("Command/Super is not bound. Ctrl+Shift+letter needs an enhanced keyboard protocol. Ctrl+K chords are the fallback.")
	return 0
}

func usageText() string {
	return strings.TrimSpace(`
axam — Axam Studio, a terminal-native text editor

Usage:
  axam [flags] [file|directory ...]
  axam doctor
  axam keymap [profile]
  axam version
  axam help

Flags:
  --profile name   axam, sublime, vscode, notepadpp, or custom
  --config path    configuration file (default: the user config path)
  --force          allow files larger than 256 MiB
  --no-config      ignore the user configuration

Install:
  sudo apt install ./axam_<version>_amd64.deb
  or copy the axam binary onto PATH

Source: https://github.com/alexshescp/axam
Support: axam@codcl.com
`) + "\n"
}
