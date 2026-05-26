package format_test

import (
	"image"
	"image/color"
	"testing"

	"deedles.dev/ximage/format"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	var data [4]byte
	format.ARGB8888.Write(data[:], 0x1111, 0x2222, 0x3333, 0xFFFF)
	require.Equal(t, [...]byte{0x33, 0x22, 0x11, 0xFF}, data)

	r, g, b, a := format.ARGB8888.Read(data[:])
	require.Equal(t, uint32(0x1111), r)
	require.Equal(t, uint32(0x2222), g)
	require.Equal(t, uint32(0x3333), b)
	require.Equal(t, uint32(0xFFFF), a)
}

func TestFormatAlphaRoundtrip(t *testing.T) {
	// Only values that survive the 16-bit → 8-bit → 16-bit conversion exactly.
	cases := []struct {
		name     string
		r, g, b, a uint32
	}{
		{"full alpha", 0x1111, 0x2222, 0x3333, 0xFFFF},
		{"zero alpha", 0, 0, 0, 0},
		{"full white", 0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf [4]byte
			format.ARGB8888.Write(buf[:], tc.r, tc.g, tc.b, tc.a)

			rr, gg, bb, aa := format.ARGB8888.Read(buf[:])
			require.Equal(t, tc.r, rr, "red")
			require.Equal(t, tc.g, gg, "green")
			require.Equal(t, tc.b, bb, "blue")
			require.Equal(t, tc.a, aa, "alpha")
		})
	}
}

// TestFormatAlphaScaling directly exercises the alpha scaling expression
// for every possible 8-bit alpha value. The original buggy expression
// (n >> 24 * 0xFFFF / 0xFF) would have returned 0 for every non-zero byte.
func TestFormatAlphaScaling(t *testing.T) {
	for alphaByte := 0; alphaByte <= 0xff; alphaByte++ {
		data := []byte{0x00, 0x00, 0x00, byte(alphaByte)}
		_, _, _, a := format.ARGB8888.Read(data)
		expected := uint32(alphaByte) * 0xFFFF / 0xFF
		require.Equal(t, expected, a, "alpha byte 0x%02x", alphaByte)
	}
}

// BenchmarkImageAt measures full-image scan performance for the two
// built-in formats. The fast paths should show 0 or 1 allocs/op
// (the unavoidable interface boxing of the returned color.Color).
func BenchmarkImageAt_ARGB8888(b *testing.B) {
	const w, h = 1024, 1024
	img := &format.Image{
		Format: format.ARGB8888,
		Rect:   image.Rect(0, 0, w, h),
		Pix:    make([]byte, w*h*4),
	}
	for i := range img.Pix {
		img.Pix[i] = byte(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum uint64
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				c := img.At(x, y)
				r, g, b, a := c.RGBA()
				sum += uint64(r + g + b + a)
			}
		}
		_ = sum
	}
}

func BenchmarkImageAt_XRGB8888(b *testing.B) {
	const w, h = 1024, 1024
	img := &format.Image{
		Format: format.XRGB8888,
		Rect:   image.Rect(0, 0, w, h),
		Pix:    make([]byte, w*h*4),
	}
	for i := range img.Pix {
		img.Pix[i] = byte(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sum uint64
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				c := img.At(x, y)
				r, g, b, a := c.RGBA()
				sum += uint64(r + g + b + a)
			}
		}
		_ = sum
	}
}

// BenchmarkImageSet measures Set performance for ARGB8888.
// A smaller size is used because Set is more expensive than At.
func BenchmarkImageSet_ARGB8888(b *testing.B) {
	const w, h = 256, 256
	img := &format.Image{
		Format: format.ARGB8888,
		Rect:   image.Rect(0, 0, w, h),
		Pix:    make([]byte, w*h*4),
	}
	c := image.NewUniform(color.RGBA{R: 0x80, G: 0x40, B: 0x20, A: 0xFF}).At(0, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.Set(x, y, c)
			}
		}
	}
}
