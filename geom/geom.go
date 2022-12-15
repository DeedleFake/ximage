// Package geom provides utilities for manipulating rectangular geometry.
//
// It is patterned heavily after image.Rectangle and image.Point, but
// vastly extends their capabilities.
package geom

import "golang.org/x/exp/constraints"

// Scalar is a constraint for the types that geom types and functions
// can handle.
type Scalar interface {
	constraints.Integer | constraints.Float
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
