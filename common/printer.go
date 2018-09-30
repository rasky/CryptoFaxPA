package common

import (
	"io/ioutil"

	"golang.org/x/text/encoding/charmap"
)

// Convert a Unicode string into the encoding understood by the printer.
// TODO: we could handle Markdown here as well
func EncodeForPrinter(s string) (out []byte) {
	cmap := charmap.CodePage437
	for _, r := range s {
		b, ok := cmap.EncodeRune(r)
		if !ok {
			b = 254 // replacement character: ■
		}
		out = append(out, b)
	}
	return
}

// Print raw (cp437-encoded) bytes to the printer
func PrintBytes(buf []byte, feed_past_cutter bool) {
	if feed_past_cutter {
		buf = append(buf, "\n\n\n\n\n"...)
	}
	ioutil.WriteFile("/dev/usb/lp0", buf, 0644)
}

// Print a Unicode string to the printer, using EncodeForPrinter to convert it
// to raw bytes.
func PrintString(s string, feed_past_cutter bool) {
	PrintBytes(EncodeForPrinter(s), feed_past_cutter)
}
