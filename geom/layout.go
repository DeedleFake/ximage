package geom

import (
	"iter"

	"deedles.dev/xiter"
)

// hsplit splits a rectangle into two rectangles arranged
// horizontally.
func hsplit[T Scalar](r Rect[T], w T) (left, right Rect[T]) {
	left = r.Resize(Pt(w, r.Dy()))
	right = r.Resize(Pt(r.Dx()-w, r.Dy())).Add(Pt(w, 0))
	return left, right
}

func hsplitHalf[T Scalar](r Rect[T]) (left, right Rect[T]) {
	return hsplit(r, r.Dx()/2)
}

// vsplit splits a rectangle into two rectangles arranged vertically.
func vsplit[T Scalar](r Rect[T], h T) (top, bottom Rect[T]) {
	top = r.Resize(Pt(r.Dx(), h))
	bottom = r.Resize(Pt(r.Dx(), r.Dy()-h)).Add(Pt(0, h))
	return top, bottom
}

func vsplitHalf[T Scalar](r Rect[T]) (top, bottom Rect[T]) {
	return vsplit(r, r.Dy()/2)
}

// TileRightThenDown arranges and resizes the elements of tiles in
// order to split r into a series of rectangles that recursively split
// each section halfway to the right and then downwards. In other
// words,
//
//	tiles := make([]geom.Rect[float64], 4)
//	TileRightThenDown(tiles, r)
//
// will produce
//
//	------------
//	|    |     |
//	|    -------
//	|    |  |  |
//	------------
func TileRightThenDown[T Scalar](tiles []Rect[T], r Rect[T]) {
	insertTilesFromSeq(tiles, TiledRightThenDown(len(tiles), r))
}

// TiledRightThenDown is the same as [TileRightThenDown] but yields
// the successive tiles from an interator instead of inserting them
// into a slice.
func TiledRightThenDown[T Scalar](numtiles int, r Rect[T]) iter.Seq[Rect[T]] {
	return func(yield func(Rect[T]) bool) {
		split, next := hsplitHalf[T], vsplitHalf[T]

		c, n := split(r)
		for range numtiles - 1 {
			if !yield(c) {
				return
			}

			c, n = split(n)
			split, next = next, split
		}

		yield(n)
	}
}

// TileTwoThirdsSidebar arranges and resizes the elements of tiles so
// that the result are a series of rectangles where the first is
// two-thirds the width of r and the rest are arranged vertically in
// an even split in the remaining space.
func TileTwoThirdsSidebar[T Scalar](tiles []Rect[T], r Rect[T]) {
	insertTilesFromSeq(tiles, TiledTwoThirdsSidebar(len(tiles), r))
}

// TiledTwoThirdsSidebar is the same as [TileTwoThirdsSidebar] except
// that it yields the successive rectangles from an iterator instead
// of inserting them into a slice.
func TiledTwoThirdsSidebar[T Scalar](numtiles int, r Rect[T]) iter.Seq[Rect[T]] {
	return func(yield func(Rect[T]) bool) {
		first, rem := hsplit(r, 2*r.Dx()/3)
		if !yield(first) {
			return
		}

		for t := range TiledEvenVertically(numtiles-1, rem) {
			if !yield(t) {
				return
			}
		}
	}
}

// TileEvenVertically arranges and resizes the elements of tiles so
// that the result are a series of rectangles that comprise an even,
// vertical splitting of r. In other words,
//
//	tiles := make([]geom.Rect[float64], 3)
//	TileEvenVertically(tiles, r)
//
// will produce
//
//	----------
//	|        |
//	----------
//	|        |
//	----------
//	|        |
//	----------
func TileEvenVertically[T Scalar](tiles []Rect[T], r Rect[T]) {
	insertTilesFromSeq(tiles, TiledEvenVertically(len(tiles), r))
}

// TiledEvenVertically is the same as [TileEvenVertically] except that
// it yields the tiles from an iterator.
func TiledEvenVertically[T Scalar](numtiles int, r Rect[T]) iter.Seq[Rect[T]] {
	return func(yield func(Rect[T]) bool) {
		size := Pt(0, r.Dy()/T(numtiles))
		c, _ := vsplit(r, size.Y)
		for range numtiles {
			if !yield(c) {
				return
			}
			c = c.Add(size)
		}
	}
}

// TileEvenHorizontally arranges and resizes the elements of tiles so
// that the result are a series of rectangles that comprise an even,
// horizontal splitting of r. In other words,
//
//	tiles := make([]geom.Rect[float64], 3)
//	TileEvenHorizontally(tiles, r)
//
// will produce
//
// ----------
// |  |  |  |
// ----------
func TileEvenHorizontally[T Scalar](tiles []Rect[T], r Rect[T]) {
	insertTilesFromSeq(tiles, TiledEvenHorizontally(len(tiles), r))
}

func TiledEvenHorizontally[T Scalar](numtiles int, r Rect[T]) iter.Seq[Rect[T]] {
	return func(yield func(Rect[T]) bool) {
		size := Pt(r.Dx()/T(numtiles), 0)
		c, _ := hsplit(r, size.X)
		for range numtiles {
			if !yield(c) {
				return
			}
			c = c.Add(size)
		}
	}
}

// TileRows arranges and resizes the elements of tiles to produce a
// series of rows and columns the union of which reproduces r. The
// final row of the table is split evenly into at most cols columns.
// When that number is exceeded, a new row is added below it instead.
func TileRows[T Scalar](tiles []Rect[T], r Rect[T], cols int) {
	insertTilesFromSeq(tiles, TiledRows(len(tiles), r, cols))
}

// TiledRows is the same as [TileRows] except that it yields the tiles
// from an iterator.
func TiledRows[T Scalar](numtiles int, r Rect[T], cols int) iter.Seq[Rect[T]] {
	return func(yield func(Rect[T]) bool) {
		numrows := numtiles / cols
		if numtiles%cols != 0 {
			numrows++
		}
		rows := TiledEvenVertically(numrows, r)

		for row := range rows {
			if numtiles <= 0 {
				break
			}

			numcols := min(numtiles, cols)
			for t := range TiledEvenHorizontally(numcols, row) {
				if !yield(t) {
					return
				}
			}
			numtiles -= numcols
		}
	}
}

// VerticalStack returns an iterator that yields the rectangle
// provided and then identical copies shifted downwards by its height
// repeatedly, thus producing an infinite vertical stack of rectangles
// below the first.
func VerticalStack[T Scalar](first Rect[T]) iter.Seq[Rect[T]] {
	return func(yield func(Rect[T]) bool) {
		r := first.Canon()
		shift := Pt(0, r.Dy())
		for {
			if !yield(first) {
				return
			}
			first = first.Add(shift)
		}
	}
}

// ArrangeVerticalStack arranges the subsequent rectangles of rects
// underneath the first vertically, expanding all for which it is
// necessary so that they are all the same width including the first.
func ArrangeVerticalStack[T Scalar](rects []Rect[T]) {
	if len(rects) <= 1 {
		return
	}

	prev := rects[0].Canon()
	for _, rect := range rects {
		if rect.Dx() > prev.Dx() {
			prev.Max.X = prev.Min.X + rect.Dx()
		}
	}
	rects[0] = prev

	for i := 1; i < len(rects); i++ {
		rects[i] = Rt(
			prev.Min.X,
			prev.Max.Y,
			prev.Max.X,
			prev.Max.Y+rects[i].Dy(),
		)
		prev = rects[i]
	}
}

// Align shifts the specified edges of inner to align with the
// corresponding edges of outer, stretching the rectangle as
// necessary if opposite edges are specified.
func Align[T Scalar](outer, inner Rect[T], edges Edges) Rect[T] {
	inner = inner.CenterAt(outer.Center())
	switch {
	case edges&EdgeTop != 0:
		inner.Min.Y, inner.Max.Y = outer.Min.Y, outer.Min.Y+inner.Dy()
		if edges&EdgeBottom != 0 {
			inner.Max.Y = outer.Max.Y
		}
	case edges&EdgeBottom != 0:
		inner.Min.Y, inner.Max.Y = outer.Max.Y-inner.Dy(), outer.Max.Y
	}
	switch {
	case edges&EdgeLeft != 0:
		inner.Min.X, inner.Max.X = outer.Min.X, outer.Min.X+inner.Dx()
		if edges&EdgeRight != 0 {
			inner.Max.X = outer.Max.X
		}
	case edges&EdgeRight != 0:
		inner.Min.X, inner.Max.X = outer.Max.X-inner.Dx(), outer.Max.X
	}

	return inner
}

func insertTilesFromSeq[T Scalar](tiles []Rect[T], s iter.Seq[Rect[T]]) {
	for i, t := range xiter.Enumerate(s) {
		tiles[i] = t
	}
}
