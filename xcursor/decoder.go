package xcursor

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"strings"
	"time"

	"deedles.dev/ximage/format"
)

func init() {
	largest := func(c *Cursor) (largest int) {
		for s := range c.Images {
			if s > largest {
				largest = s
			}
		}
		return
	}

	image.RegisterFormat(
		"xcursor",
		"Xcur",
		func(r io.Reader) (image.Image, error) {
			cur, err := Decode(r)
			if err != nil {
				return nil, err
			}
			if len(cur.Images) == 0 {
				return nil, errors.New("no images in cursor")
			}

			return cur.Images[largest(cur)][0].Image, nil
		},
		func(r io.Reader) (image.Config, error) {
			cur, err := Decode(r)
			if err != nil {
				return image.Config{}, err
			}
			if len(cur.Images) == 0 {
				return image.Config{}, errors.New("no images in cursor")
			}

			largest := cur.Images[largest(cur)]
			bounds := largest[0].Image.Bounds()
			return image.Config{
				ColorModel: largest[0].Image.ColorModel(),
				Width:      bounds.Dx(),
				Height:     bounds.Dy(),
			}, nil
		},
	)
}

// ErrBadMagic indicates an unrecognized magic number when attempting
// to load a cursor.
var ErrBadMagic = errors.New("bad magic")

const (
	fileMagic = 0x72756358 // ASCII "Xcur"
)

// Cursor contains information decoded from a Xcursor file.
type Cursor struct {
	Comments []*Comment
	Images   map[int][]*Image
}

// Comment is a comment section of an Xcursor file.
type Comment struct {
	Subtype CommentSubtype
	Comment string
}

type CommentSubtype uint32

const (
	CommentSubtypeCopyright CommentSubtype = 1 + iota
	CommentSubtypeLicense
	CommentSubtypeOther
)

const (
	tocTypeComment = 0xfffe0001
	tocTypeImage   = 0xfffd0002
)

// Image is an image section of an Xcursor file.
type Image struct {
	NominalSize int
	Delay       time.Duration
	Hot         image.Point
	Image       *format.Image
}

// BestSize searches the available sizes for the cursor and returns
// the one that is closest to the target size. If two are equidistant
// to size, the larger of the two is returned.
func (c *Cursor) BestSize(size int) (best int) {
	for s := range c.Images {
		best = betterSize(size, best, s)
	}
	return best
}

func betterSize(target, a, b int) int {
	da := dist(target, a)
	db := dist(target, b)
	switch {
	case da < db:
		return a
	case db < da:
		return b
	default:
		if a > b {
			return a
		}
		return b
	}
}

func dist(a, b int) int {
	if a < b {
		return b - a
	}
	return a - b
}

type decoder struct {
	r   io.Reader
	br  *bufio.Reader
	n   int
	err error
}

// DecodeFile decodes the Xcursor file at path.
func DecodeFile(path string) (*Cursor, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer file.Close()

	return Decode(file)
}

// Decode decodes an Xcursor file from r.
func Decode(r io.Reader) (*Cursor, error) {
	d := decoder{
		r:  r,
		br: bufio.NewReader(r),
	}
	return d.Decode()
}

func (d *decoder) Decode() (c *Cursor, err error) {
	if d.err != nil {
		return nil, d.err
	}

	defer d.catch(&err)

	cursor := Cursor{
		Images: make(map[int][]*Image),
	}

	tocs := d.header()
	for _, toc := range tocs {
		d.SeekTo(int(toc.Position))
		d.tocHeader(toc)
		switch toc.Type {
		case tocTypeComment:
			cursor.Comments = append(cursor.Comments, d.comment(toc))
		case tocTypeImage:
			img := d.image(toc)
			cursor.Images[img.NominalSize] = append(cursor.Images[img.NominalSize], img)
		default:
			d.throw(fmt.Errorf("unknown TOC type: %x", toc.Type))
		}
	}

	return &cursor, nil
}

func (d *decoder) header() []fileToc {
	magic := d.uint32()
	if magic != fileMagic {
		d.throw(ErrBadMagic)
	}
	d.uint32() // Header size.
	d.uint32() // Version.
	ntoc := int(d.uint32())

	tocs := make([]fileToc, 0, ntoc)
	for range ntoc {
		tocs = append(tocs, fileToc{
			Type:     d.uint32(),
			Subtype:  d.uint32(),
			Position: d.uint32(),
		})
	}

	return tocs
}

func (d *decoder) tocHeader(toc fileToc) {
	d.uint32() // Header size.

	tocType := d.uint32()
	if tocType != toc.Type {
		d.throw(fmt.Errorf("TOC type mismatch: expected: %v, got: %v", toc.Type, tocType))
	}

	tocSubtype := d.uint32()
	if tocSubtype != toc.Subtype {
		d.throw(fmt.Errorf("TOC subtype mismatch: expected: %v, got: %v", toc.Subtype, tocSubtype))
	}

	d.uint32() // Version.
}

func (d *decoder) comment(toc fileToc) *Comment {
	length := d.uint32()

	var buf strings.Builder
	buf.Grow(int(length))
	_, err := io.CopyN(&buf, d, int64(length))
	d.throw(err)

	return &Comment{
		Subtype: CommentSubtype(toc.Subtype),
		Comment: buf.String(),
	}
}

func (d *decoder) image(toc fileToc) *Image {
	w := d.uint32()
	h := d.uint32()
	xhot := d.uint32()
	yhot := d.uint32()
	delay := d.uint32()

	pixels := make([]byte, w*h*4)
	_, err := io.ReadFull(d, pixels)
	d.throw(err)

	return &Image{
		NominalSize: int(toc.Subtype),
		Delay:       time.Duration(delay) * time.Millisecond,
		Hot:         image.Pt(int(xhot), int(yhot)),
		Image: &format.Image{
			Format: format.ARGB8888,
			Rect:   image.Rect(0, 0, int(w), int(h)),
			Pix:    pixels,
		},
	}
}

func (d *decoder) uint32() (v uint32) {
	d.throw(binary.Read(d, binary.LittleEndian, &v))
	return v
}

func (d *decoder) Read(buf []byte) (int, error) {
	n, err := d.br.Read(buf)
	d.throw(err)
	d.n += n
	return n, err
}

func (d *decoder) Discard(n int) (int, error) {
	disc, err := d.br.Discard(n)
	d.throw(err)
	d.n += disc
	return disc, err
}

func (d *decoder) SeekTo(n int) error {
	diff := n - d.n
	if diff < 0 {
		panic("tried to seek backwards")
	}
	if diff == 0 {
		return nil
	}

	s, ok := d.r.(io.Seeker)
	if !ok || (diff <= d.br.Buffered()) {
		_, err := d.Discard(diff)
		d.throw(err)
		return nil
	}

	_, err := s.Seek(int64(n), io.SeekStart)
	d.throw(err)
	d.br.Reset(d.r)
	d.n = n
	return nil
}

type fileToc struct {
	Type     uint32
	Subtype  uint32
	Position uint32
}

type decoderError struct {
	err error
}

func (d *decoder) throw(err error) {
	if err != nil {
		panic(decoderError{err: err})
	}
}

func (d *decoder) catch(err *error) {
	switch r := recover().(type) {
	case decoderError:
		*err = r.err
		d.err = r.err
	case nil:
		*err = d.err
	default:
		panic(r)
	}
}
