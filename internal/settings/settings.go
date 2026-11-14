// Package settings keeps the menus' game values on this machine between runs:
// a JSON file in the user's config directory, beside the saved battle and the
// scores. It mirrors internal/saves. A store that cannot be read counts as the
// defaults, and a value it does not know falls back to its default.
package settings

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Values are the game values, named as the web edition names its settings.
type Values struct {
	PlayMode string `json:"playMode"` // single, simultaneous (two humans) or hotseat
	Reduced  bool   `json:"reduced"`  // short battles: moves and fire are instant
	Opponent string `json:"opponent"` // player B: computer or passive
	Limit    int    `json:"limit"`    // orders each side may give a phase, 0 for no limit
	Scroll   string `json:"scroll"`   // how fast the keys scroll the map: normal, fast or slow
}

// The values each field takes; the first is its default.
var (
	PlayModes = []string{"single", "simultaneous", "hotseat"}
	Opponents = []string{"computer", "passive"}
	Limits    = []int{0, 5, 10, 15, 20}
	Scrolls   = []string{"normal", "fast", "slow"}
)

// Defaults are the values of a first run.
func Defaults() Values {
	return Values{PlayMode: PlayModes[0], Opponent: Opponents[0], Limit: Limits[0], Scroll: Scrolls[0]}
}

// Store is the kept game values.
type Store struct {
	load func() ([]byte, error)
	save func([]byte) error
}

// OpenFile is a store kept in the JSON file at path; the file and its
// directory are made on the first save.
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

// Load returns the kept values: the defaults where there are none, each
// value the store does not know replaced by its default.
func (s *Store) Load() Values {
	v := Defaults()
	if s == nil {
		return v
	}
	b, err := s.load()
	if err != nil {
		return v
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(b, &fields) != nil {
		return v
	}
	read(fields["playMode"], &v.PlayMode, PlayModes)
	read(fields["reduced"], &v.Reduced, []bool{false, true})
	read(fields["opponent"], &v.Opponent, Opponents)
	read(fields["limit"], &v.Limit, Limits)
	read(fields["scroll"], &v.Scroll, Scrolls)
	return v
}

// read sets *dst to the value in raw when it is one of known.
func read[T comparable](raw json.RawMessage, dst *T, known []T) {
	var v T
	if raw != nil && json.Unmarshal(raw, &v) == nil && slices.Contains(known, v) {
		*dst = v
	}
}

// Save keeps v.
func (s *Store) Save(v Values) error {
	if s == nil {
		return errNoStore
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return s.save(b)
}

var errNoStore = errors.New("settings: no place to keep them")

func none() *Store {
	return &Store{
		load: func() ([]byte, error) { return nil, fs.ErrNotExist },
		save: func([]byte) error { return errNoStore },
	}
}
