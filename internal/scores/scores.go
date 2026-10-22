// Package scores keeps the best score of each level on this machine: a JSON
// file in the user's config directory, or the browser's local storage in the
// web build. A store that cannot be read counts as empty and a failed write is
// reported, but neither stops the game.
package scores

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Store is the best score per level, keyed by the caller's level name.
type Store struct {
	load func() ([]byte, error)
	save func([]byte) error

	best   map[string]int
	loaded bool
}

// OpenFile is a store kept in the JSON file at path; the file and its
// directory are made on the first record.
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

// Best is the best score recorded for level, 0 for none.
func (s *Store) Best(level string) int {
	if s == nil {
		return 0
	}
	s.read()
	return s.best[level]
}

// Record keeps score as level's best if it beats the one recorded and reports
// whether it did.
func (s *Store) Record(level string, score int) (bool, error) {
	if s == nil {
		return false, nil
	}
	s.read()
	if old, ok := s.best[level]; ok && score <= old {
		return false, nil
	}
	s.best[level] = score
	b, err := json.MarshalIndent(s.best, "", "  ")
	if err != nil {
		return true, err
	}
	return true, s.save(b)
}

// read loads the store once; anything unreadable counts as empty.
func (s *Store) read() {
	if s.loaded {
		return
	}
	s.loaded, s.best = true, map[string]int{}
	b, err := s.load()
	if err != nil {
		return
	}
	var m map[string]int
	if json.Unmarshal(b, &m) == nil && m != nil {
		s.best = m
	}
}

// errNoStore is the error of a store with nowhere to keep its scores.
var errNoStore = errors.New("scores: no place to keep them")

// none is a store that keeps scores for the session only.
func none() *Store {
	return &Store{
		load: func() ([]byte, error) { return nil, fs.ErrNotExist },
		save: func([]byte) error { return errNoStore },
	}
}
