// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultsAndOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	body := "keymap_profile = \"vscode\"\nword_wrap = true\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, missing, err := Load(path)
	if err != nil || missing {
		t.Fatal(err, missing)
	}
	if cfg.KeymapProfile != "vscode" || !cfg.WordWrap || !cfg.LineNumbers || cfg.TabSize != 4 {
		t.Fatalf("%+v", cfg)
	}
}

func TestInvalidLocation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("tab_size = \"wide\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatal(err)
	}
}

func TestUnknownField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("not_a_setting = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBadProfile(t *testing.T) {
	cfg := Defaults()
	cfg.KeymapProfile = "vim"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid profile")
	}
}
