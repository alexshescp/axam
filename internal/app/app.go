// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package app is the Axam Studio session: documents, commands, and the screen.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexshescp/axam/internal/buffer"
	"github.com/alexshescp/axam/internal/clipboard"
	"github.com/alexshescp/axam/internal/commands"
	"github.com/alexshescp/axam/internal/config"
	"github.com/alexshescp/axam/internal/editor"
	"github.com/alexshescp/axam/internal/keymap"
	"github.com/alexshescp/axam/internal/syntax"
	"github.com/alexshescp/axam/internal/ui"
	"github.com/gdamore/tcell/v2"
)

const (
	modeEdit = iota
	modeFind
	modePrompt
	modeList
	modeMenu
)

const (
	focusEditor = iota
	focusExplorer
)

const (
	listPalette = iota
	listQuick
	listSearch
	listRecent
)

// App is one editing session.
type App struct {
	screen   tcell.Screen
	cfg      config.Config
	theme    ui.Theme
	keys     *keymap.Map
	info     keymap.Info
	cmds     map[string]func()
	docs     []*Document
	panes    []pane
	active   int
	focus    int
	mode     int
	chord    []keymap.Key
	clip     clipboard.Clip
	status   string
	find     findBox
	prompt   promptBox
	list     listBox
	menu     menuState
	expl     explorer
	root     string
	files    []string
	hits     []hit
	w, h     int
	quit     bool
	force    bool
	closed   []closedDoc
	recent   config.State
	pasting  bool
	pasteBuf []byte
	click    clickState
	sigs     chan os.Signal
}

type pane struct {
	Doc  int
	Top  int
	Left int
	Row  int
}

// Document is one open buffer.
type Document struct {
	Path     string
	Name     string
	Ed       *editor.Editor
	Encoding string
	BOM      bool
	Perm     os.FileMode
	ModTime  time.Time
	Hash     [32]byte
	DiskSize int64
	ReadOnly bool
	Lang     string
	Symlink  string
	Exists   bool
	Large    bool
	Folds    map[int]int
	hl       map[int]hlLine
}

type hlLine struct {
	gen   uint64
	spans []syntax.Span
}

type closedDoc struct {
	path string
	name string
	text string
	eol  buffer.EOL
}

type findBox struct {
	query, repl       string
	pos               int
	field             int
	replace           bool
	sens, word, regex bool
	err               string
	matches           []findMatch
	at                int
}

type findMatch struct {
	Line, Col, EndLine, EndCol int
}

type promptBox struct {
	title string
	text  string
	pos   int
	kind  string
	onOK  func(string)
	onYes func()
	onNo  func()
}

type listBox struct {
	kind     int
	title    string
	query    string
	pos      int
	sel      int
	items    []listItem
	all      []listItem
	waiting  bool
	searched string
}

type listItem struct {
	Title, Detail, Shortcut, ID, Payload string
	Line, Col                            int
}

type menuState struct {
	group int
	item  int
}

type explorer struct {
	on      bool
	entries []explEntry
	sel     int
}

type explEntry struct {
	Path  string
	Name  string
	Depth int
	Dir   bool
	Open  bool
}

type hit struct {
	kind      int
	line      int
	col       int
	tab       int
	expl      int
	pane      int
	menuGroup int
	menuItem  int
}

const (
	hitNone = iota
	hitText
	hitGutter
	hitTab
	hitExpl
	hitMenuGroup
	hitMenuItem
)

type clickState struct {
	when                  time.Time
	x, y                  int
	count                 int
	drag                  bool
	alt                   bool
	pane                  int
	anchorLine, anchorCol int
}

// New builds a session on an already initialized screen.
func New(screen tcell.Screen, cfg config.Config) (*App, error) {
	km, info, err := keymap.Load(cfg.KeymapProfile)
	if err != nil {
		return nil, err
	}
	for chord, cmd := range cfg.Bindings {
		keys, err := keymap.ParseChord(chord)
		if err != nil {
			return nil, fmt.Errorf("binding %q: %w", chord, err)
		}
		if cmd != "" && !commands.Known(cmd) {
			return nil, fmt.Errorf("binding %q: unknown command %q", chord, cmd)
		}
		km.Set(keys, cmd)
	}
	a := &App{
		screen: screen,
		cfg:    cfg,
		theme:  ui.For(cfg.Theme, 256),
		keys:   km,
		info:   info,
		panes:  []pane{{Doc: -1}},
		recent: config.LoadState(),
		expl:   explorer{on: cfg.Explorer},
	}
	a.cmds = a.commandTable()
	return a, nil
}

// Run owns the terminal until quit. The screen is finalized on return.
func (a *App) Run() error {
	if err := a.screen.Init(); err != nil {
		return err
	}
	defer a.screen.Fini()
	defer func() {
		if rec := recover(); rec != nil {
			panic(rec)
		}
	}()
	a.screen.EnableMouse(tcell.MouseDragEvents)
	a.screen.EnablePaste()
	a.screen.EnableFocus()
	a.screen.SetCursorStyle(tcell.CursorStyleSteadyBar)
	a.theme = ui.For(a.cfg.Theme, a.screen.Colors())
	a.sigs = make(chan os.Signal, 1)
	signal.Notify(a.sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-a.sigs
		_ = a.screen.PostEvent(tcell.NewEventInterrupt(nil))
	}()
	a.Draw()
	a.screen.Show()
	for !a.quit {
		ev := a.screen.PollEvent()
		if ev == nil {
			break
		}
		a.Handle(ev)
		if a.quit {
			break
		}
		a.Draw()
		a.screen.Show()
	}
	_ = config.SaveState(a.recent)
	return nil
}

// SetRoot sets the project root and refreshes the file list.
func (a *App) SetRoot(path string) {
	a.root = path
	a.refreshFiles()
	if path != "" {
		a.expl.on = true
		a.loadExplorer()
		a.recent.RecentProjects = config.Remember(a.recent.RecentProjects, path, a.cfg.RecentLimit)
	}
}

// Force allows opening files past the size guard. Binary files stay refused.
func (a *App) Force() { a.force = true }

func (a *App) doc() *Document {
	if a.active < 0 || a.active >= len(a.panes) {
		return nil
	}
	i := a.panes[a.active].Doc
	if i < 0 || i >= len(a.docs) {
		return nil
	}
	return a.docs[i]
}

func (a *App) setStatus(s string) { a.status = s }

func (a *App) title() string {
	d := a.doc()
	if d == nil {
		return "axam"
	}
	name := d.Name
	if d.Ed.Dirty() {
		name = "● " + name
	}
	return "axam — " + name
}
