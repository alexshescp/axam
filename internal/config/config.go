// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

// Config is the user-editable settings file.
type Config struct {
	KeymapProfile     string            `toml:"keymap_profile"`
	Theme             string            `toml:"theme"`
	TabSize           int               `toml:"tab_size"`
	InsertSpaces      bool              `toml:"insert_spaces"`
	LineNumbers       bool              `toml:"line_numbers"`
	WordWrap          bool              `toml:"word_wrap"`
	ShowTrailingSpace bool              `toml:"show_trailing_whitespace"`
	Explorer          bool              `toml:"explorer"`
	Clipboard         string            `toml:"clipboard"`
	DetectIndent      bool              `toml:"detect_indent"`
	Backup            bool              `toml:"backup"`
	RecentLimit       int               `toml:"recent_limit"`
	Ignore            []string          `toml:"ignore"`
	Bindings          map[string]string `toml:"bindings"`
}

// State is recent-file data, kept out of the config the user edits.
type State struct {
	RecentFiles    []string `json:"recent_files"`
	RecentProjects []string `json:"recent_projects"`
}

// Defaults is the configuration used when no file exists.
func Defaults() Config {
	return Config{
		KeymapProfile:     "axam",
		Theme:             "dark",
		TabSize:           4,
		InsertSpaces:      true,
		LineNumbers:       true,
		WordWrap:          false,
		ShowTrailingSpace: true,
		Explorer:          false,
		Clipboard:         "auto",
		DetectIndent:      true,
		Backup:            true,
		RecentLimit:       50,
	}
}

// Template is the file written the first time axam runs.
func Template() string {
	return `# axam — Axam Studio
# Edit this file like source code. Invalid values are reported with their location.

keymap_profile = "axam"        # axam | sublime | vscode | notepadpp | custom
theme = "dark"                 # dark | light | mono
tab_size = 4
insert_spaces = true
line_numbers = true
word_wrap = false
show_trailing_whitespace = true
explorer = false
clipboard = "auto"             # auto | internal | osc52 | xclip | wl-copy
detect_indent = true
backup = true
recent_limit = 50

# Extra directories and globs skipped by quick open and project search.
# Built-in skips (.git, node_modules, vendor, dist, build, ...) always apply.
ignore = []

# Replace profile shortcuts. The command id is the same one shown in the palette.
# [bindings]
# "ctrl+s" = "file.save"
`
}

// Dir is the configuration directory.
func Dir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "axam")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".axam"
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "axam")
	}
	return filepath.Join(home, ".config", "axam")
}

// Path is the config file path.
func Path() string { return filepath.Join(Dir(), "config.toml") }

// Load reads path. A missing file returns defaults and okMissing=true.
func Load(path string) (Config, bool, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, true, nil
		}
		return cfg, false, err
	}
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		var de *toml.DecodeError
		if errors.As(err, &de) {
			row, col := de.Position()
			return Defaults(), false, fmt.Errorf("%s:%d:%d: %s", path, row, col, de.Error())
		}
		return Defaults(), false, fmt.Errorf("%s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Defaults(), false, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, false, nil
}

// Validate checks fields that TOML types cannot express.
func (c Config) Validate() error {
	switch c.KeymapProfile {
	case "axam", "termstudio", "sublime", "vscode", "notepadpp", "custom", "default":
	default:
		return fmt.Errorf("keymap_profile %q is not one of axam, sublime, vscode, notepadpp, custom", c.KeymapProfile)
	}
	switch c.Theme {
	case "dark", "light", "mono":
	default:
		return fmt.Errorf("theme %q is not one of dark, light, mono", c.Theme)
	}
	if c.TabSize < 1 || c.TabSize > 16 {
		return fmt.Errorf("tab_size %d is outside 1..16", c.TabSize)
	}
	switch c.Clipboard {
	case "auto", "internal", "osc52", "xclip", "wl-copy":
	default:
		return fmt.Errorf("clipboard %q is not one of auto, internal, osc52, xclip, wl-copy", c.Clipboard)
	}
	if c.RecentLimit < 0 || c.RecentLimit > 500 {
		return fmt.Errorf("recent_limit %d is outside 0..500", c.RecentLimit)
	}
	return nil
}

// Ensure writes the template when the file is missing.
func Ensure() (string, error) {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return path, err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err := os.WriteFile(path, []byte(Template()), 0o644); err != nil {
		return path, err
	}
	return path, nil
}

func statePath() string { return filepath.Join(Dir(), "state.json") }

// LoadState reads recent files. Missing state is empty.
func LoadState() State {
	data, err := os.ReadFile(statePath())
	if err != nil {
		return State{}
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return State{}
	}
	return st
}

// SaveState writes recent files.
func SaveState(st State) error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(), data, 0o644)
}

// Remember adds a path to the front of a list and trims it.
func Remember(list []string, path string, limit int) []string {
	if path == "" {
		return list
	}
	out := []string{path}
	for _, p := range list {
		if p != path {
			out = append(out, p)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
