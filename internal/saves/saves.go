// Package saves keeps one saved battle (single slot) on this machine: a JSON
// file in the user's config directory.
// It mirrors internal/scores. A store that cannot be read counts as empty.
package saves

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Envelope is one saved battle: the menu map name, the human side (dos side
// 0 or 1), the dosrules save bytes (base64 in JSON) and, for two humans on
// one screen, their mode ("simultaneous" or "hotseat"; Side is then the human
// who plays next).
type Envelope struct {
	Map  string `json:"map"`
	Side int    `json:"side"`
	Data []byte `json:"data"`
	Mode string `json:"mode,omitempty"`
}

// Store is the single saved battle.
type Store struct {
	load func() ([]byte, error)
	save func([]byte) error
}

// OpenFile is a store kept in the JSON file at path.
func OpenFile(path string) *Store {
	return &Store{
		load: func() ([]byte, error) { return os.ReadFile(path) },
		save: func(b []byte) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return os.WriteFile(path, b, 0o644)
		},
	}
}

// Put replaces the saved battle.
func (s *Store) Put(e Envelope) error {
	if s == nil {
		return errNoStore
	}
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return s.save(b)
}

// Get returns the saved battle, false when there is none readable.
func (s *Store) Get() (Envelope, bool) {
	if s == nil {
		return Envelope{}, false
	}
	b, err := s.load()
	if err != nil {
		return Envelope{}, false
	}
	var e Envelope
	if json.Unmarshal(b, &e) != nil || e.Map == "" || len(e.Data) == 0 || e.Side < 0 || e.Side > 1 {
		return Envelope{}, false
	}
	return e, true
}

// Has reports whether a saved battle is present.
func (s *Store) Has() bool { _, ok := s.Get(); return ok }

var errNoStore = errors.New("saves: no place to keep them")

func none() *Store {
	return &Store{
		load: func() ([]byte, error) { return nil, fs.ErrNotExist },
		save: func([]byte) error { return errNoStore },
	}
}
