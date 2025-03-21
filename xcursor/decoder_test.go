//go:build go1.24

package xcursor_test

import (
	"image"
	_ "image/png"
	"os"
	"testing"

	"deedles.dev/ximage/xcursor"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	pngFile, err := os.Open("testdata/left_ptr.png")
	require.Nil(t, err)
	defer pngFile.Close()

	png, _, err := image.Decode(pngFile)
	require.Nil(t, err)

	theme, err := xcursor.LoadThemeFromDir("testdata")
	require.Nil(t, err)
	xc, ok := theme.Cursors["left_ptr"]
	require.True(t, ok)
	xcimg := xc.Images[xc.BestSize(24)][0].Image
	bounds := xcimg.Bounds()
	require.Equal(t, bounds, png.Bounds())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			xr, xg, xb, xa := xcimg.At(x, y).RGBA()
			pr, pg, pb, pa := png.At(x, y).RGBA()

			require.Equal(t, xr, pr)
			require.Equal(t, xg, pg)
			require.Equal(t, xb, pb)
			require.Equal(t, xa, pa)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	for b.Loop() {
		theme, _ := xcursor.LoadThemeFromDir("testdata")
		xc, _ := theme.Cursors["left_ptr"]
		xcimg := xc.Images[xc.BestSize(24)][0].Image
		bounds := xcimg.Bounds()

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				xcimg.At(x, y).RGBA()
			}
		}
	}
}
