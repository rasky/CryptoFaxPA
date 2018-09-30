package common

import (
    "bytes"
    "io/ioutil"
)

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
