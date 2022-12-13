package ximage

import (
	"encoding/binary"
	"image"
	"image/color"
)

type Format interface {
	// Size returns the number of bytes per pixel.
	Size() int

	// Read reads raw pixel data and converts it to alpha-premultiplied
	// RGBA values, similar to color.Color's RGBA method.
	Read([]byte) (r, g, b, a uint32)

	// Write writes alpha-premultiplied RGBA values into buf.
	Write(buf []byte, r, g, b, a uint32)
}

var (
	ARGB8888 formatARGB8888
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

type FormatModel struct {
	Format Format
}

func (m FormatModel) Convert(c color.Color) color.Color {
	fc := FormatColor{Format: m.Format}
	r, g, b, a := c.RGBA()
	m.Format.Write(fc.Slice(), r, g, b, a)
	return &fc
}

type FormatColor struct {
	Format Format
	Data   [8]byte
}

func (c *FormatColor) Slice() []byte {
	return c.Data[:c.Format.Size()]
}

func (c *FormatColor) RGBA() (r, g, b, a uint32) {
	return c.Format.Read(c.Slice())
}

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

	i := img.PixOffset(x, y)
	s := img.Pix[i : i+size : i+size]
	copy(c.Slice(), s)

	return &c
}

func (img *FormatImage) Stride() int {
	return img.Format.Size() * img.Rect.Dx()
}

func (img *FormatImage) PixOffset(x, y int) int {
	x -= img.Rect.Min.X
	y -= img.Rect.Min.Y
	return (img.Stride() * y) + x
}

func (img *FormatImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(img.Rect)) {
		return
	}

	size := img.Format.Size()
	i := img.PixOffset(x, y)
	c1 := img.ColorModel().Convert(c).(*FormatColor)
	s := img.Pix[i : i+size : i+size]
	copy(s, c1.Slice())
}
