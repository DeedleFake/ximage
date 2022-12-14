package xcursor

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var defaultLibraryPaths = []string{
	"~/.icons",
	"/usr/share/icons",
	"/usr/share/pixmaps",
	"~/.cursors",
	"/usr/share/cursors/xorg-x11",
	"/usr/X11R6/lib/X11/icons",
}

func libraryPaths() []string {
	if v, ok := os.LookupEnv("XCURSOR_PATH"); ok {
		return filepath.SplitList(v)
	}

	v, ok := os.LookupEnv("XDG_DATA_HOME")
	if !ok || !filepath.IsAbs(v) {
		v = "~/.local/share"
	}
	return append([]string{filepath.Join(v, "icons")}, defaultLibraryPaths...)
}

// Theme is an Xcursor theme.
type Theme struct {
	Name    string
	Cursors map[string]*Cursor
}

// LoadTheme loads the named theme from the system search paths. It
// resepects the $XURSOR_PATH and $XDG_DATA_HOME environment variables
// when looking. If the theme has an index.theme file and that file
// lists other themes to inherit from, those themes are also loaded
// and their cursors are added to the returned theme.
func LoadTheme(name string) (*Theme, error) {
	if name == "" {
		name = "default"
	}

	c := Theme{
		Name:    name,
		Cursors: make(map[string]*Cursor),
	}
	return &c, c.load(name)
}

// LoadThemeFromDir loads a theme from the directory at path, ignoring
// the system search path completely. The returned theme's name is the
// basename of the given path.
func LoadThemeFromDir(path string) (*Theme, error) {
	c := Theme{
		Name:    filepath.Base(path),
		Cursors: make(map[string]*Cursor),
	}
	return &c, c.loadDir(path)
}

func (t *Theme) load(theme string) error {
	for _, path := range libraryPaths() {
		dir := filepath.Join(path, theme, "cursors")
		err := t.loadDir(dir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("load dir %q: %w", dir, err)
		}

		inherits, err := loadInherits(filepath.Join(path, theme, "index.theme"))
		if (err != nil) && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("load inherited themes: %w", err)
		}
		for _, theme := range inherits {
			err := t.load(theme)
			if err != nil {
				return fmt.Errorf("load inherited theme %q: %w", theme, err)
			}
		}

		break
	}

	return nil
}

func (t *Theme) loadDir(path string) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, ent := range dir {
		if _, ok := t.Cursors[ent.Name()]; ok {
			continue
		}
		if t := ent.Type().Type(); !t.IsRegular() && (t != fs.ModeSymlink) {
			continue
		}

		entpath := filepath.Join(path, ent.Name())
		cur, err := DecodeFile(entpath)
		if err != nil {
			if errors.Is(err, ErrBadMagic) {
				continue
			}
			return fmt.Errorf("load %q: %w", entpath, err)
		}

		t.Cursors[ent.Name()] = cur
	}

	return nil
}

func loadInherits(index string) (inherits []string, err error) {
	file, err := os.Open(index)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, "Inherits") {
			continue
		}

		_, after, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		inherits = strings.FieldsFunc(after, func(c rune) bool {
			return (c == ':') || (c == ',')
		})
		for i, v := range inherits {
			inherits[i] = strings.TrimSpace(v)
		}

		break
	}
	if err := s.Err(); err != nil {
		return inherits, fmt.Errorf("scan: %w", err)
	}

	return inherits, nil
}
