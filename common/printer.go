package common

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "golang.org/x/text/encoding/charmap"
)

func EncodeForPrinter(s string) (string) {
    s, err := charmap.CodePage437.NewEncoder().String(s)
    if err != nil {
        // the supplied message contains unsupported characters
        fmt.Println(err)
    }
    return s
}

func PrintBytes(buf []byte, feed_past_cutter bool) {
    if feed_past_cutter {
        buf = append(buf, "\n\n\n\n\n"...)
    }
    ioutil.WriteFile("/dev/usb/lp0", buf, 0644)
}

func PrintBuffer(buf bytes.Buffer, feed_past_cutter bool) {
    PrintBytes(buf.Bytes(), feed_past_cutter)
}

func PrintString(s string, feed_past_cutter bool) {
    PrintBytes([]byte(s), feed_past_cutter)
}
