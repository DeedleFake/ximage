package format

import (
	"encoding/binary"
)

// Format is a pixel format for a FormatImage and related types. This
// package contains several predefined formats, such as [ARGB8888].
type Format interface {
	// Size returns the number of bytes per pixel.
	Size() int

	// Read reads raw pixel data and converts it to alpha-premultiplied
	// RGBA values, similar to color.Color's RGBA method.
	Read([]byte) (r, g, b, a uint32)

	// Write writes alpha-premultiplied RGBA values into buf.
	Write(buf []byte, r, g, b, a uint32)
}

// Various predefined Formats.
var (
	ARGB8888 formatARGB8888
	XRGB8888 formatXRGB8888
)

type formatARGB8888 struct{}

func (formatARGB8888) String() string { return "ARGB8888" }

func (formatARGB8888) Size() int { return 4 }

func (formatARGB8888) Read(data []byte) (r, g, b, a uint32) {
	n := binary.LittleEndian.Uint32(data)
	a = (n >> 24 * 0xFFFF / 0xFF)
	r = (n >> 16 & 0xFF) * a / 0xFF
	g = (n >> 8 & 0xFF) * a / 0xFF
	b = (n & 0xFF) * a / 0xFF
	return
}

func (formatARGB8888) Write(buf []byte, r, g, b, a uint32) {
	r = (r * 0xFF / a) << 16
	g = (g * 0xFF / a) << 8
	b = b * 0xFF / a
	a = (a * 0xFF / 0xFFFF) << 24
	binary.LittleEndian.PutUint32(buf, r|g|b|a)
}

type formatXRGB8888 struct{}

func (formatXRGB8888) String() string { return "XRGB8888" }

func (formatXRGB8888) Size() int { return 4 }

func (formatXRGB8888) Read(data []byte) (r, g, b, a uint32) {
	n := binary.LittleEndian.Uint32(data)
	a = 0xFFFF
	r = (n >> 16 & 0xFF) * 0xFFFF / 0xFF
	g = (n >> 8 & 0xFF) * 0xFFFF / 0xFF
	b = (n & 0xFF) * 0xFFFF / 0xFF
	return
}

func (formatXRGB8888) Write(buf []byte, r, g, b, a uint32) {
	r = (r * 0xFF / 0xFFFF) << 16
	g = (g * 0xFF / 0xFFFF) << 8
	b = b * 0xFF / 0xFFFF
	a = 0xFF << 24
	binary.LittleEndian.PutUint32(buf, r|g|b|a)
}
