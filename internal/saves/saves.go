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
// who plays next) and the resolution a hot-seat human has not seen yet.
type Envelope struct {
	Map     string   `json:"map"`
	Side    int      `json:"side"`
	Data    []byte   `json:"data"`
	Mode    string   `json:"mode,omitempty"`
	Pending *Pending `json:"pending,omitempty"`
}

// Pending is a resolution the hot-seat human of Side has not seen yet, played
// on that human's next turn: the dosrules save bytes of the game just before
// the resolution (base64 in JSON) and its session events, an array of
// internal/dosplay events as encoding/json writes them (the web edition
// writes the same keys).
type Pending struct {
	Side   int             `json:"side"`
	Before []byte          `json:"before"`
	Events json.RawMessage `json:"events"`
}

// valid reports whether p names a side, a game and an array of events.
func (p *Pending) valid() bool {
	var evs []json.RawMessage
	return p != nil && (p.Side == 0 || p.Side == 1) && len(p.Before) > 0 && json.Unmarshal(p.Events, &evs) == nil && evs != nil
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
	e, err := Decode(b)
	if err != nil || e.Map == "" || len(e.Data) == 0 || e.Side < 0 || e.Side > 1 {
		return Envelope{}, false
	}
	return e, true
}

// Decode reads an envelope without checking its battle. The pending playback
// is read on its own: a malformed one is dropped, so the battle still loads.
func Decode(b []byte) (Envelope, error) {
	var raw struct {
		Envelope
		Pending json.RawMessage `json:"pending"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return Envelope{}, err
	}
	e := raw.Envelope
	var p Pending
	if json.Unmarshal(raw.Pending, &p) == nil && p.valid() {
		e.Pending = &p
	}
	return e, nil
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
