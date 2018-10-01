package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"time"

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

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Print raw (cp437-encoded) bytes to the printer
func PrintBytes(buf []byte, feed_past_cutter bool) {
	f, err := os.Create("/dev/usb/lp0")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// write in chunks in order not to saturate the usblp buffer
	for offset := 0; offset < len(buf); offset += 1024 {
		end := min(offset+1024, len(buf))
		f.Write(buf[offset:end])
		time.Sleep(100 * time.Millisecond)
	}

	f.Write([]byte("\n"))
	if feed_past_cutter {
		f.Write([]byte("\n\n\n\n"))
	}
}

// Print a Unicode string to the printer, using EncodeForPrinter to convert it
// to raw bytes.
func PrintString(s string, feed_past_cutter bool) {
	PrintBytes(EncodeForPrinter(s), feed_past_cutter)
}

func PrintImage(pngimg []byte, feed_past_cutter bool) {
	var DOTS_PER_LINE = 384

	imgobj, err := png.Decode(bytes.NewReader(pngimg))
	if err != nil {
		fmt.Println(err)
	}
	b := imgobj.Bounds()
	imgWidth := b.Max.X
	imgHeight := b.Max.Y
	fmt.Printf("decoded image: %vx%v\n", imgWidth, imgHeight)

	if imgWidth > DOTS_PER_LINE || imgWidth%8 > 0 {
		// TODO: resize image if it's too large or its
		// width is not a multiple of 8
		fmt.Println("image too large; skipping")
		return
	}

	/*
	   // save image to a file for debugging purposes
	   f, err := os.Create("img.png")
	   if err != nil {
	       panic(err)
	   }
	   defer f.Close()
	   png.Encode(f, imgobj)
	*/

	// turn the decoded PNG image into an array of black/white dots
	var bits = make([]bool, imgWidth*imgHeight)
	bounds := imgobj.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// black = 1, white = 0
			bits[y*imgWidth+x] = color.GrayModel.Convert(imgobj.At(x, y)).(color.Gray).Y < 128
		}
	}

	var buf bytes.Buffer
	buf.WriteString("\x1b!\x03") // set font mode 3

	// prepare the command used for printing a single line
	cmd := make([]byte, 5)
	copy(cmd[0:], "\x1b*\x08") // select SDL graphics
	binary.LittleEndian.PutUint16(cmd[3:], uint16(DOTS_PER_LINE/8))

	// iterate over lines
	for i := 0; i < imgHeight; i++ {
		buf.Write(cmd)

		// calculate the bits range for this line
		start := imgWidth * i
		end := imgWidth * (i + 1)

		// iterate over the bits to be printed
		for k := start; k < end; k += 8 {
			// since we can't append individual bits to buf,
			// make a byte
			var b byte
			for j := 0; j < 8; j++ {
				// reversed endianness!
				if bits[k+(7-j)] {
					b |= 1 << uint16(j)
				} else {
					b |= 0 << uint16(j)
				}
			}

			buf.WriteByte(b)
		}

		// if the image to be printed is smaller than our printer width,
		// pad it with white dots
		buf.Write(make([]byte, (DOTS_PER_LINE-imgWidth)/8))
		buf.WriteByte('\n')
	}

	PrintBytes(buf.Bytes(), feed_past_cutter)
}
