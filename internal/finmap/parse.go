package finmap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Parse reads a level written by the importer. Unknown fields, another
// format or version, and terrain that does not match the size are errors.
func Parse(raw []byte) (*Level, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var l Level
	if err := dec.Decode(&l); err != nil {
		return nil, fmt.Errorf("finmap: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("finmap: data after the level object")
	}
	if l.Format != Format || l.Version != Version {
		return nil, fmt.Errorf("finmap: format %q version %d, want %q version %d", l.Format, l.Version, Format, Version)
	}
	if l.Width < 1 || l.Height < 1 || len(l.Terrain) != l.Height {
		return nil, fmt.Errorf("finmap: %d terrain rows for a %d×%d map", len(l.Terrain), l.Width, l.Height)
	}
	for r, row := range l.Terrain {
		if len(row) != l.Width {
			return nil, fmt.Errorf("finmap: terrain row %d has %d cells, want %d", r, len(row), l.Width)
		}
	}
	return &l, nil
}
