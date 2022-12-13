package format

import (
	"image"
	"image/color"
)

// Model implements color.Model using a Format.
type Model struct {
	Format Format
}

func (m Model) Convert(c color.Color) color.Color {
	fc := Color{Format: m.Format}
	r, g, b, a := c.RGBA()
	m.Format.Write(fc.Slice(), r, g, b, a)
	return &fc
}

// Color implements color.Color using a Format.
type Color struct {
	Format Format

	// Data contains the pixel data for the color. Only some bytes of
	// the array are used, dependant on the return value of Format.Size.
	Data [8]byte
}

// Slice returns a slice of Data correctly sized for the color's format.
func (c *Color) Slice() []byte {
	size := c.Format.Size()
	return c.slice(size)
}

func (c *Color) slice(size int) []byte {
	return c.Data[:size:size]
}

func (c *Color) RGBA() (r, g, b, a uint32) {
	return c.Format.Read(c.Slice())
}

// Image is an image with a color format defined by Format.
type Image struct {
	Format Format
	Rect   image.Rectangle
	Pix    []byte
}

func (img *Image) Bounds() image.Rectangle { return img.Rect }

func (img *Image) ColorModel() color.Model { return Model{Format: img.Format} }

func (img *Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(img.Rect)) {
		return &Color{Format: img.Format}
	}

	size := img.Format.Size()
	c := Color{Format: img.Format}

	i := img.pixOffset(x, y, img.stride(size), size)
	s := img.Pix[i : i+size : i+size]
	copy(c.slice(size), s)

	return &c
}

func (img *Image) Stride() int {
	return img.stride(img.Format.Size())
}

func (img *Image) stride(size int) int {
	return size * img.Rect.Dx()
}

func (img *Image) PixOffset(x, y int) int {
	return img.pixOffset(x, y, img.Stride(), img.Format.Size())
}

func (img *Image) pixOffset(x, y, stride, size int) int {
	x -= img.Rect.Min.X
	y -= img.Rect.Min.Y
	return (stride * y) + (x * size)
}

func (img *Image) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(img.Rect)) {
		return
	}

	size := img.Format.Size()
	i := img.pixOffset(x, y, img.stride(size), size)
	c1 := img.ColorModel().Convert(c).(*Color)
	s := img.Pix[i : i+size : i+size]
	copy(s, c1.slice(size))
}
