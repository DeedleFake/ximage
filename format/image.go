package format

import (
	"encoding/binary"
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
		// Return a zero value for out-of-bounds. Using color.RGBA avoids
		// allocating a *Color for the common built-in formats.
		return color.RGBA{}
	}

	switch img.Format {
	case ARGB8888:
		i := img.pixOffset(x, y, img.stride(4), 4)
		n := binary.LittleEndian.Uint32(img.Pix[i:])
		r, g, b, a := argb8888Read(n)
		return color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
	case XRGB8888:
		i := img.pixOffset(x, y, img.stride(4), 4)
		n := binary.LittleEndian.Uint32(img.Pix[i:])
		r, g, b, _ := xrgb8888Read(n)
		return color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: 0xFF,
		}
	default:
		// Generic path for custom Format implementations (preserves existing behavior)
		size := img.Format.Size()
		c := Color{Format: img.Format}
		i := img.pixOffset(x, y, img.stride(size), size)
		s := img.Pix[i : i+size : i+size]
		copy(c.slice(size), s)
		return &c
	}
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

	r, g, b, a := c.RGBA()

	switch img.Format {
	case ARGB8888:
		i := img.pixOffset(x, y, img.stride(4), 4)
		n := argb8888Write(r, g, b, a)
		binary.LittleEndian.PutUint32(img.Pix[i:], n)
	case XRGB8888:
		i := img.pixOffset(x, y, img.stride(4), 4)
		n := xrgb8888Write(r, g, b, a)
		binary.LittleEndian.PutUint32(img.Pix[i:], n)
	default:
		// Generic path for custom Formats (preserves existing allocation behavior)
		size := img.Format.Size()
		i := img.pixOffset(x, y, img.stride(size), size)
		c1 := img.ColorModel().Convert(c).(*Color)
		s := img.Pix[i : i+size : i+size]
		copy(s, c1.slice(size))
	}
}
