// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package clipboard

import "testing"

func TestOSC52(t *testing.T) {
	got := EncodeOSC52("hi")
	if got != "\033]52;c;aGk=\033\\" {
		t.Fatalf("%q", got)
	}
}

func TestAutoDoesNotSpawn(t *testing.T) {
	if err := WriteExternal("auto", "hi"); err != nil {
		t.Fatal(err)
	}
}
