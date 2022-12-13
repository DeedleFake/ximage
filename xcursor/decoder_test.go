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

	xc, err := xcursor.DecodeFile("testdata/left_ptr")
	require.Nil(t, err)
	xcimg := xc.Images[0].Image
	bounds := xcimg.Bounds()
	require.Equal(t, bounds, png.Bounds())
}
