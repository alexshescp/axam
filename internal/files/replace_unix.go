// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !windows

package files

import "os"

func replaceFile(tmp, dst string) error {
	return os.Rename(tmp, dst)
}
