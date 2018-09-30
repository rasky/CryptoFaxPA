package common

import (
    "bufio"
    "os"
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
	//ioutil.WriteFile("/dev/usb/lp0", buf, 0644)
	
	f, err := os.Create("/dev/usb/lp0")
	if err != nil {
		panic(err)
	}
	defer f.Close()
    
	w := bufio.NewWriter(f)
	w.Write(buf)
	w.Flush()
	
	f.Write([]byte("\n"))
	if (feed_past_cutter) {
	    f.Write([]byte("\n\n\n\n"))
	}
	w.Flush()
	w.Flush()
	
	//fmt.Println(string(buf))
}

// Print a Unicode string to the printer, using EncodeForPrinter to convert it
// to raw bytes.
func PrintString(s string, feed_past_cutter bool) {
	PrintBytes(EncodeForPrinter(s), feed_past_cutter)
}
