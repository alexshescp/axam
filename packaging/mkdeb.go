// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build ignore

// Command mkdeb writes a .deb that apt can install from a local file.
// It does not need dpkg. The payload is a static axam binary plus a termstudio symlink.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func main() {
	bin := flag.String("bin", "", "path to the axam binary")
	arch := flag.String("arch", "amd64", "Debian architecture (amd64 or arm64)")
	version := flag.String("version", "1.1.0", "package version")
	out := flag.String("out", "", "output .deb path")
	flag.Parse()
	if *bin == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "mkdeb: -bin and -out are required")
		os.Exit(2)
	}
	blob, err := os.ReadFile(*bin)
	if err != nil {
		fatal(err)
	}
	man := manPage(*version)
	copyright := copyrightText()
	control := controlText(*version, *arch)
	sums := md5sums(map[string][]byte{
		"usr/bin/axam":                 blob,
		"usr/share/doc/axam/copyright": []byte(copyright),
		"usr/share/man/man1/axam.1":    []byte(man),
	})
	data, err := tarGz(map[string]entry{
		"usr/bin/axam":                 {body: blob, mode: 0o755},
		"usr/bin/termstudio":           {link: "axam", mode: 0o777},
		"usr/share/doc/axam/copyright": {body: []byte(copyright), mode: 0o644},
		"usr/share/man/man1/axam.1":    {body: []byte(man), mode: 0o644},
	})
	if err != nil {
		fatal(err)
	}
	ctrl, err := tarGz(map[string]entry{
		"control": {body: []byte(control), mode: 0o644},
		"md5sums": {body: []byte(sums), mode: 0o644},
	})
	if err != nil {
		fatal(err)
	}
	var buf bytes.Buffer
	buf.WriteString("!<arch>\n")
	writeAr(&buf, "debian-binary", []byte("2.0\n"))
	writeAr(&buf, "control.tar.gz", ctrl)
	writeAr(&buf, "data.tar.gz", data)
	if err := os.WriteFile(*out, buf.Bytes(), 0o644); err != nil {
		fatal(err)
	}
}

type entry struct {
	body []byte
	mode int64
	link string
}

func tarGz(files map[string]entry) ([]byte, error) {
	var raw bytes.Buffer
	gz := gzip.NewWriter(&raw)
	tw := tar.NewWriter(gz)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sortStrings(names)
	seen := map[string]bool{}
	var dirs []string
	for _, name := range names {
		dir := path.Dir(name)
		for dir != "." && dir != "/" && !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
			dir = path.Dir(dir)
		}
	}
	sortStrings(dirs)
	for _, dir := range dirs {
		hdr := &tar.Header{
			Name:     dir + "/",
			Mode:     0o755,
			Typeflag: tar.TypeDir,
			ModTime:  time.Unix(0, 0),
			Format:   tar.FormatUSTAR,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
	}
	for _, name := range names {
		e := files[name]
		hdr := &tar.Header{
			Name:    name,
			Mode:    e.mode,
			ModTime: time.Unix(0, 0),
			Format:  tar.FormatUSTAR,
			Uname:   "root",
			Gname:   "root",
		}
		if e.link != "" {
			hdr.Typeflag = tar.TypeSymlink
			hdr.Linkname = e.link
			hdr.Size = 0
		} else {
			hdr.Typeflag = tar.TypeReg
			hdr.Size = int64(len(e.body))
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if e.link == "" {
			if _, err := tw.Write(e.body); err != nil {
				return nil, err
			}
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return raw.Bytes(), nil
}

func writeAr(w io.Writer, name string, body []byte) {
	if len(name) > 16 {
		name = name[:16]
	}
	hdr := make([]byte, 60)
	copy(hdr[0:16], pad(name, 16))
	copy(hdr[16:28], pad("0", 12))
	copy(hdr[28:34], pad("0", 6))
	copy(hdr[34:40], pad("0", 6))
	copy(hdr[40:48], pad("100644", 8))
	copy(hdr[48:58], pad(strconv.Itoa(len(body)), 10))
	hdr[58] = 0x60
	hdr[59] = 0x0a
	_, _ = w.Write(hdr)
	_, _ = w.Write(body)
	if len(body)%2 == 1 {
		_, _ = w.Write([]byte{'\n'})
	}
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func controlText(version, arch string) string {
	return fmt.Sprintf(`Package: axam
Version: %s
Section: editors
Priority: optional
Architecture: %s
Maintainer: Coders Club, Codepolder, Aleksandr Shaman <packaging@codersclub.invalid>
Description: TermStudio, a terminal-native text editor
 axam edits text in a Linux or macOS terminal, including SSH sessions.
 The same binary is also installed as termstudio.
 The packaging address is not a monitored mailbox.
`, version, arch)
}

func copyrightText() string {
	return `Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: axam
Upstream-Contact: Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman
Source: github.com/codersclub/axam

Files: *
Copyright: 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman
License: GPL-3.0-or-later
 The GNU General Public License, version 3 or any later version.
 On Debian systems the full text is in /usr/share/common-licenses/GPL-3.
`
}

func manPage(version string) string {
	return fmt.Sprintf(`.TH AXAM 1 "2026" "axam %s" "TermStudio"
.SH NAME
axam, termstudio \- terminal-native text editor
.SH SYNOPSIS
.B axam
[\fIflags\fR] [\fIfile\fR|\fIdirectory\fR ...]
.br
.B axam doctor
.br
.B axam keymap
[\fIprofile\fR]
.br
.B axam version
.SH DESCRIPTION
axam is TermStudio, a modeless terminal editor. It runs in a local Linux or
macOS terminal and over SSH. Command-key shortcuts from a Mac keyboard are
not available through the terminal byte stream.
.PP
A missing filename is opened as a pending document. It is not created until
an explicit Save. Closing it unchanged leaves no file.
.PP
F10 opens the bottom File/Edit/Selection/Find/View/Goto/Help command menu.
.SH OPTIONS
.TP
.B \-\-profile name
Keymap profile: termstudio, sublime, vscode, notepadpp, or custom.
.TP
.B \-\-config path
Configuration file. The default is the user config path under
~/.config/axam/config.toml.
.TP
.B \-\-force
Open files larger than 256 MiB. Binary files are still refused.
.TP
.B \-\-no-config
Ignore the user configuration.
.SH FILES
.TP
~/.config/axam/config.toml
User settings and key bindings.
.SH SEE ALSO
The installed documentation and
.B axam doctor
for terminal limits.
.SH LICENSE
GPL-3.0-or-later
.SH AUTHORS
Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
`, version)
}

func md5sums(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sortStrings(names)
	var b strings.Builder
	for _, name := range names {
		sum := md5.Sum(files[name])
		fmt.Fprintf(&b, "%x  %s\n", sum, name)
	}
	return b.String()
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mkdeb:", err)
	os.Exit(1)
}
