// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build ignore

// Command demo prints one frame of the editor from the simulation screen.
package main

import (
	"fmt"
	"os"

	"github.com/alexshescp/axam/internal/app"
	"github.com/alexshescp/axam/internal/config"
	"github.com/gdamore/tcell/v2"
)

func main() {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer s.Fini()
	s.SetSize(80, 24)
	session, err := app.New(s, config.Defaults())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	session.OpenString("main.go", "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"axam\")\n}\n")
	session.Draw()
	s.Show()
	cells, w, h := s.GetContents()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b := cells[y*w+x].Bytes
			if len(b) == 0 {
				fmt.Print(" ")
				continue
			}
			fmt.Print(string(b))
		}
		fmt.Println()
	}
}
