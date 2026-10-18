package dosrules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// Fixtures record executed results as numbers only. A casualty fixture
// follows docs/research/battle-fixture.schema.json, an engagement fixture
// docs/research/engagement-fixture.schema.json; internal/dosoracle writes
// both from the emulated routines.

// Source says how a fixture's expectation was obtained.
type Source struct {
	Kind  string `json:"kind"`  // emulated-routine, controlled-game, transcription, manual
	Build string `json:"build"` // SHA-256 prefix of the executable
	Note  string `json:"note,omitempty"`
}

// FixtureSide is one side of a casualty fixture: a type of the build's
// tables or a synthetic record, and the live values.
type FixtureSide struct {
	Type          *int  `json:"type,omitempty"`
	Record        *Type `json:"record,omitempty"`
	Size          int   `json:"size"`
	AlternateSize int   `json:"alternateSize"`
	Experience    int   `json:"experience"`
	Flags         int   `json:"flags"`
}

// FixtureContext are a casualty fixture's battle inputs.
type FixtureContext struct {
	Distance             int  `json:"distance"`
	AttackerTerrainClass int  `json:"attackerTerrainClass"`
	DefenderTerrainClass int  `json:"defenderTerrainClass"`
	Pincer               int  `json:"pincer"`
	Blocking             int  `json:"blocking"`
	ReturnFire           bool `json:"returnFire"`
}

// CasualtyExpected is the routine's result, or the error it raised.
type CasualtyExpected struct {
	Error                  string `json:"error,omitempty"`
	AttackerSizeAfter      int    `json:"attackerSizeAfter"`
	AttackerAlternateAfter int    `json:"attackerAlternateAfter"`
	DefenderSizeAfter      int    `json:"defenderSizeAfter"`
	DefenderAlternateAfter int    `json:"defenderAlternateAfter"`
}

// CasualtyFixture is one recorded run of the casualty routine.
type CasualtyFixture struct {
	Schema   int              `json:"schema"`
	Profile  string           `json:"profile"`
	ID       string           `json:"id"`
	Source   Source           `json:"source"`
	Attacker FixtureSide      `json:"attacker"`
	Defender FixtureSide      `json:"defender"`
	Context  FixtureContext   `json:"context"`
	Expected CasualtyExpected `json:"expected"`
}

func (s FixtureSide) fighter(t *Tables) (Fighter, error) {
	var typ Type
	switch {
	case s.Record != nil:
		typ = *s.Record
	case s.Type != nil && *s.Type >= 0 && *s.Type < len(t.Types):
		typ = t.Types[*s.Type]
	default:
		return Fighter{}, errors.New("fixture side has neither a known type nor a record")
	}
	return Fighter{typ, s.Size, s.AlternateSize, s.Experience, s.Flags}, nil
}

// Inputs returns the fixture's fighters and context.
func (f *CasualtyFixture) Inputs(t *Tables) (a, d Fighter, c Context, err error) {
	if a, err = f.Attacker.fighter(t); err != nil {
		return
	}
	if d, err = f.Defender.fighter(t); err != nil {
		return
	}
	x := f.Context
	return a, d, Context{x.Distance, x.AttackerTerrainClass, x.DefenderTerrainClass, x.Pincer, x.Blocking, x.ReturnFire}, nil
}

// errorName is the fixture spelling of an error.
func errorName(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrDivideError):
		return "divide-error"
	case errors.Is(err, ErrUndefined):
		return "undefined"
	}
	return err.Error()
}

// CasualtyResult is the fixture form of a Casualties result.
func CasualtyResult(r Result, err error) CasualtyExpected {
	if err != nil {
		return CasualtyExpected{Error: errorName(err)}
	}
	return CasualtyExpected{"", r.AttackerSize, r.AttackerAlt, r.DefenderSize, r.DefenderAlt}
}

// FixtureUnit is a live unit of an engagement fixture.
type FixtureUnit struct {
	Index         int `json:"index"`
	Type          int `json:"type"`
	Cell          int `json:"cell"`
	Size          int `json:"size"`
	AlternateSize int `json:"alternateSize"`
	Experience    int `json:"experience"`
	Flags4        int `json:"flags4"`
	Flags6        int `json:"flags6"`
}

// FixtureTerrain is a cell's terrain class.
type FixtureTerrain struct {
	Cell  int `json:"cell"`
	Class int `json:"class"`
}

// EngagementExpected is the state the order handler left, or its error.
type EngagementExpected struct {
	Error                 string `json:"error,omitempty"`
	Battle                bool   `json:"battle"`
	Defender              int    `json:"defender"`
	ReturnFire            bool   `json:"returnFire"`
	Pincer                int    `json:"pincer"`
	Blocking              int    `json:"blocking"`
	AttackerSize          int    `json:"attackerSize"`
	AttackerAlternateSize int    `json:"attackerAlternateSize"`
	DefenderSize          int    `json:"defenderSize"`
	DefenderAlternateSize int    `json:"defenderAlternateSize"`
	AttackerExperience    int    `json:"attackerExperience"`
	DefenderExperience    int    `json:"defenderExperience"`
	Destroyed             []int  `json:"destroyed"`
}

// EngagementFixture is one recorded attack order through the order handler.
type EngagementFixture struct {
	Schema   int                `json:"schema"`
	Profile  string             `json:"profile"`
	ID       string             `json:"id"`
	Source   Source             `json:"source"`
	Width    int                `json:"width"`
	Height   int                `json:"height"`
	Side     int                `json:"side"` // the player whose orders run (no effect on the outcome)
	Attacker int                `json:"attacker"`
	Target   int                `json:"target"`
	MapByte0 int                `json:"mapByte0"`
	Units    []FixtureUnit      `json:"units"`
	Terrain  []FixtureTerrain   `json:"terrain"`
	Expected EngagementExpected `json:"expected"`
}

// Battlefield returns the fixture's map.
func (f *EngagementFixture) Battlefield() *Battlefield {
	b := &Battlefield{Width: f.Width, Height: f.Height, Units: map[int]Unit{}, Terrain: map[int]int{}, MapByte0: f.MapByte0}
	for _, u := range f.Units {
		b.Units[u.Index] = Unit{u.Type, u.Cell, u.Size, u.AlternateSize, u.Experience, u.Flags4, u.Flags6}
	}
	for _, c := range f.Terrain {
		b.Terrain[c.Cell] = c.Class
	}
	return b
}

// Engagement returns a fixture of a battlefield and an attack order, with
// nothing expected yet.
func Engagement(b *Battlefield, attacker, target, side int) EngagementFixture {
	f := EngagementFixture{Schema: 1, Width: b.Width, Height: b.Height, Side: side,
		Attacker: attacker, Target: target, MapByte0: b.MapByte0, Units: []FixtureUnit{}, Terrain: []FixtureTerrain{}}
	for idx, u := range b.Units {
		f.Units = append(f.Units, FixtureUnit{idx, u.Type, u.Cell, u.Size, u.Alt, u.Exp, u.Flags4, u.Flags6})
	}
	slices.SortFunc(f.Units, func(a, b FixtureUnit) int { return a.Index - b.Index })
	for cell, class := range b.Terrain {
		f.Terrain = append(f.Terrain, FixtureTerrain{cell, class})
	}
	slices.SortFunc(f.Terrain, func(a, b FixtureTerrain) int { return a.Cell - b.Cell })
	return f
}

// EngagementResult is the fixture form of an Engage outcome.
func EngagementResult(o Outcome, err error) EngagementExpected {
	if err != nil {
		return EngagementExpected{Error: errorName(err), Destroyed: []int{}}
	}
	d := append([]int{}, o.Destroyed...)
	return EngagementExpected{"", o.Battle, o.Defender, o.ReturnFire, o.Pincer, o.Blocking,
		o.AttackerSize, o.AttackerAlt, o.DefenderSize, o.DefenderAlt, o.AttackerExp, o.DefenderExp, d}
}

// MarshalFixtures writes fixtures as a JSON array, one fixture per line.
func MarshalFixtures[T any](fs []T) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, f := range fs {
		b, err := json.Marshal(f)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
		if i < len(fs)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("]\n")
	return buf.Bytes(), nil
}

// UnmarshalFixtures reads a fixture array and refuses unknown fields.
func UnmarshalFixtures[T any](b []byte) ([]T, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var fs []T
	if err := dec.Decode(&fs); err != nil {
		return nil, fmt.Errorf("fixtures: %w", err)
	}
	return fs, nil
}
