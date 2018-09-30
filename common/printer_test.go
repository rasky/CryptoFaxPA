package common

import (
	"bytes"
	"testing"
)

func TestEncodeForPrinter(t *testing.T) {
	var tests = []struct {
		in  string
		out []byte
	}{
		{"prova", []byte("prova")},
		{"prova 🍷", []byte("prova \xfe")},
		{"àèìòù", []byte("\x85\x8a\x8d\x95\x97")},
	}

	for _, tc := range tests {
		got := EncodeForPrinter(tc.in)
		if !bytes.Equal(got, tc.out) {
			t.Errorf("invalid encoding: src=%q got=%q(%x) exp=%q(%x)", tc.in, string(got), got, string(tc.out), tc.out)
		}
	}
}
