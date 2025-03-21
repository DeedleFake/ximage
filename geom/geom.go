// Package geom provides utilities for manipulating rectangular geometry.
//
// It is patterned heavily after image.Rectangle and image.Point, but
// vastly extends their capabilities.
package geom

// Scalar is a constraint for the types that geom types and functions
// can handle.
type Scalar interface {
	~float32 | ~float64 | Integer
}

// Integer is a constraint for any integer type.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Edges is a bitmask representing zero or more edges of a rectangle.
type Edges uint32

const (
	EdgeNone Edges = 0
	EdgeTop  Edges = 1 << (iota - 1)
	EdgeBottom
	EdgeLeft
	EdgeRight
)
