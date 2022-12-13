package ximage

import (
	"encoding/binary"
	"image"
	"image/color"
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

// FormatModel implements color.Model using a Format.
type FormatModel struct {
	Format Format
}

func (m FormatModel) Convert(c color.Color) color.Color {
	fc := FormatColor{Format: m.Format}
	r, g, b, a := c.RGBA()
	m.Format.Write(fc.Slice(), r, g, b, a)
	return &fc
}

// FormatColor implements color.Color using a Format.
type FormatColor struct {
	Format Format

	// Data contains the pixel data for the color. Only some bytes of
	// the array are used, dependant on the return value of Format.Size.
	Data [8]byte
}

// Slice returns a slice of Data correctly sized for the color's format.
func (c *FormatColor) Slice() []byte {
	size := c.Format.Size()
	return c.slice(size)
}

func (c *FormatColor) slice(size int) []byte {
	return c.Data[:size:size]
}

func (c *FormatColor) RGBA() (r, g, b, a uint32) {
	return c.Format.Read(c.Slice())
}

// FormatImage is an image with a color format defined by Format.
type FormatImage struct {
	Format Format
	Rect   image.Rectangle
	Pix    []byte
}

func (img *FormatImage) Bounds() image.Rectangle { return img.Rect }

func (img *FormatImage) ColorModel() color.Model { return FormatModel{Format: img.Format} }

func (img *FormatImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(img.Rect)) {
		return &FormatColor{Format: img.Format}
	}

	size := img.Format.Size()
	c := FormatColor{Format: img.Format}

	i := img.pixOffset(x, y, img.stride(size), size)
	s := img.Pix[i : i+size : i+size]
	copy(c.slice(size), s)

	return &c
}

func (img *FormatImage) Stride() int {
	return img.stride(img.Format.Size())
}

func (img *FormatImage) stride(size int) int {
	return size * img.Rect.Dx()
}

func (img *FormatImage) PixOffset(x, y int) int {
	return img.pixOffset(x, y, img.Stride(), img.Format.Size())
}

func (img *FormatImage) pixOffset(x, y, stride, size int) int {
	x -= img.Rect.Min.X
	y -= img.Rect.Min.Y
	return (stride * y) + (x * size)
}

func (img *FormatImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(img.Rect)) {
		return
	}

	size := img.Format.Size()
	i := img.pixOffset(x, y, img.stride(size), size)
	c1 := img.ColorModel().Convert(c).(*FormatColor)
	s := img.Pix[i : i+size : i+size]
	copy(s, c1.slice(size))
}
