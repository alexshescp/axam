// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build windows

package files

import "os"

// Windows cannot rename a file over an existing name atomically.
// The previous file is removed immediately before the rename. A crash in
// that window can lose the previous name; the temporary file still holds
// the new bytes in the same directory.
func replaceFile(tmp, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			_ = os.Remove(tmp)
			return err
		}
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
