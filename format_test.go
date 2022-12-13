package ximage_test

import (
	"testing"

	"deedles.dev/ximage"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	var data [4]byte
	ximage.ARGB8888.Write(data[:], 0x1111, 0x2222, 0x3333, 0xFFFF)
	require.Equal(t, [...]byte{0x33, 0x22, 0x11, 0xFF}, data)

	r, g, b, a := ximage.ARGB8888.Read(data[:])
	require.Equal(t, uint32(0x1111), r)
	require.Equal(t, uint32(0x2222), g)
	require.Equal(t, uint32(0x3333), b)
	require.Equal(t, uint32(0xFFFF), a)
}
