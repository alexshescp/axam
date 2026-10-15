// Copyright (C) 2026 Coders Club (USA), Codepolder (Netherlands), Aleksandr Shaman.
// SPDX-License-Identifier: GPL-3.0-or-later

package buffer

import (
	"bytes"
	"testing"

	"github.com/alexshescp/axam/internal/text"
)

func TestRoundTripEOL(t *testing.T) {
	in := []byte("a\r\nb\r\n")
	ld, err := FromBytes(in)
	if err != nil {
		t.Fatal(err)
	}
	if ld.Buffer.EOL() != CRLF || ld.Buffer.String() != "a\nb\n" {
		t.Fatalf("eol %s text %q", ld.Buffer.EOL(), ld.Buffer.String())
	}
	if !bytes.Equal(ld.Buffer.Bytes(), in) {
		t.Fatalf("bytes %q", ld.Buffer.Bytes())
	}
}

func TestMixedEOL(t *testing.T) {
	ld, err := FromBytes([]byte("a\r\nb\nc\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !ld.Buffer.Mixed() || ld.Buffer.EOL() != CRLF {
		t.Fatalf("mixed %v eol %s", ld.Buffer.Mixed(), ld.Buffer.EOL())
	}
}

func TestUTF16RoundTrip(t *testing.T) {
	raw := Encode([]byte("Привет\n"), EncUTF16LE, true)
	ld, err := FromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ld.Encoding != EncUTF16LE || !ld.BOM || ld.Buffer.String() != "Привет\n" {
		t.Fatalf("%s %v %q", ld.Encoding, ld.BOM, ld.Buffer.String())
	}
	if !bytes.Equal(Encode(ld.Buffer.Bytes(), EncUTF16LE, true), raw) {
		t.Fatal("encode mismatch")
	}
}

func TestInvalidUTF8Kept(t *testing.T) {
	in := []byte{'a', 0xff, 'b'}
	ld, _ := FromBytes(in)
	if ld.Buffer.String() != string(in) {
		t.Fatalf("%q", ld.Buffer.String())
	}
	if !bytes.Equal(ld.Buffer.Bytes(), in) {
		t.Fatal("bytes changed")
	}
}

func TestInsertDelete(t *testing.T) {
	b := FromString("hello")
	end := b.Insert(text.Pos{Col: 5}, "\nмир")
	if b.String() != "hello\nмир" || end.Line != 1 || end.Col != len("мир") {
		t.Fatalf("%q end %+v", b.String(), end)
	}
	b.Delete(text.Pos{Line: 0, Col: 1}, text.Pos{Line: 1, Col: 2})
	if b.String() != "hир" && b.String() != "h"+"ир" {
		// deleted "ello\nм" (м is 2 bytes), left "h" + "ир"
		t.Fatalf("%q", b.String())
	}
}

func TestBinaryDetect(t *testing.T) {
	if !IsBinary([]byte{'a', 0, 'b'}) {
		t.Fatal("expected binary")
	}
	if IsBinary([]byte("plain")) {
		t.Fatal("plain")
	}
}
