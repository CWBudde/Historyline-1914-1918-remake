package dosoracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// The fixtures of internal/dosrules are recorded here from the executed
// routines, numbers only: the build's tables (dosrules/<build>.json), every
// controlled battle of the engagement tests, every casualty coverage case,
// the movement cases, the move orders and the container cases
// (dosrules/testdata/<build>/).
// Without -update the test checks that the committed files are what the
// emulator gives now.

var update = flag.Bool("update", false, "rewrite the internal/dosrules tables and fixtures from the local exports")

// typeOf reads the fields of a type record that dosrules uses.
func typeOf(r []byte) dosrules.Type {
	return dosrules.Type{
		Armour: int(r[2]), Group: int(r[3]), TargetFlags: word(r, 6),
		AirRange: int(r[8]), SurfaceRange: int(r[9]),
		AirPower: int(r[10]), GroundPower: int(r[11]), WaterPower: int(r[12]),
		AttackMask: word(r, 0x0D), StrengthFlags: word(r, 0x0F),
	}
}

func tablesOf(e *Export) *dosrules.Tables {
	t := &dosrules.Tables{Build: e.B.SHA256[:8]}
	for i := range len(e.Units) / recordLen {
		t.Types = append(t.Types, typeOf(e.TypeRecord(i)))
	}
	for i, tt := range e.TerrainTerms() {
		t.Terms[i] = dosrules.Terms{Attack: tt[0], Defence: tt[1]}
	}
	return t
}

var profileCache = map[*Build]dosrules.Profile{}

// profileOf is profile dos-0.1 on the export's own tables.
func profileOf(e *Export) dosrules.Profile {
	p, ok := profileCache[e.B]
	if !ok {
		p = dosrules.Profile{Name: "dos-0.1", Tables: tablesOf(e), TransposedDistance: true}
		profileCache[e.B] = p
	}
	return p
}

func source(e *Export, note string) dosrules.Source {
	return dosrules.Source{Kind: "emulated-routine", Build: e.B.SHA256[:8], Note: note}
}

// rulesError maps an emulator or transcription error onto dosrules'.
func rulesError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, dos8086.ErrDivideError), errors.Is(err, ErrDivideError):
		return dosrules.ErrDivideError
	case errors.Is(err, ErrUndefined):
		return dosrules.ErrUndefined
	}
	return err
}

func battlefieldOf(g *Engagement) *dosrules.Battlefield {
	b := &dosrules.Battlefield{Width: g.Width, Height: g.Height, Units: map[int]dosrules.Unit{},
		Terrain: map[int]int{}, MapByte0: g.class(0)}
	for i, u := range g.Units {
		b.Units[i] = dosrules.Unit{Type: u.Type, Cell: u.Cell, Size: u.Size, Alt: u.Alt, Exp: u.Exp,
			Flags4: u.Flags4, Flags6: u.Flags6}
	}
	for c, class := range g.Terrain {
		b.Terrain[c] = class
	}
	return b
}

func expectedOf(o Outcome, err error) dosrules.EngagementExpected {
	return dosrules.EngagementResult(dosrules.Outcome{
		Battle: o.Battle, Defender: o.Defender, ReturnFire: o.ReturnFire, Pincer: o.Pincer, Blocking: o.Blocking,
		Result:      dosrules.Result{AttackerSize: o.AttackerSize, AttackerAlt: o.AttackerAlt, DefenderSize: o.DefenderSize, DefenderAlt: o.DefenderAlt},
		AttackerExp: o.AttackerExp, DefenderExp: o.DefenderExp, Destroyed: o.Destroyed,
	}, rulesError(err))
}

// resolved is dosrules' outcome of a scene.
func resolved(e *Export, g *Engagement) dosrules.EngagementExpected {
	return dosrules.EngagementResult(profileOf(e).Engage(battlefieldOf(g), g.Attacker, g.Target))
}

// fixtureSide names a build record by its index and gives a synthetic one in full.
func fixtureSide(e *Export, s Side) dosrules.FixtureSide {
	f := dosrules.FixtureSide{Size: s.Size, AlternateSize: s.Alt, Experience: s.Exp, Flags: s.Flg}
	for i := range len(e.Units) / recordLen {
		if bytes.Equal(s.Rec, e.TypeRecord(i)) {
			f.Type = &i
			return f
		}
	}
	t := typeOf(s.Rec)
	f.Record = &t
	return f
}

func casualtyFixture(e *Export, id string, c battleCase) dosrules.CasualtyFixture {
	return dosrules.CasualtyFixture{
		Schema: 1, Profile: "dos-0.1", ID: id, Source: source(e, "casualty routine"),
		Attacker: fixtureSide(e, c.a), Defender: fixtureSide(e, c.d),
		Context: dosrules.FixtureContext{Distance: c.p.Distance, AttackerTerrainClass: c.p.ATerrain,
			DefenderTerrainClass: c.p.DTerrain, Pincer: c.p.Pincer, Blocking: c.p.Blocking, ReturnFire: c.p.ReturnFire},
	}
}

// fixtureStride records every k-th case of a coverage class.
var fixtureStride = map[string]int{"type records": 6, "casualty arithmetic": 4}

// sceneSink, when set, receives every scene the engagement tests run.
var sceneSink func(what string, g *Engagement)

func TestFixtures(t *testing.T) { eachExport(t, fixtures) }

func fixtures(t *testing.T, e *Export) {
	needHandler(t, e)
	var engagements []dosrules.EngagementFixture
	for _, set := range []struct {
		name string
		run  func(*testing.T, *Export)
	}{
		{"duel", engagementDuels}, {"blocking", engagementBlocking}, {"pincer", engagementPincer},
		{"ranged", engagementRanged}, {"experience", engagementExperience},
	} {
		sceneSink = func(what string, g *Engagement) {
			f := dosrules.Engagement(battlefieldOf(g), g.Attacker, g.Target, g.Side)
			f.Profile, f.ID, f.Source = "dos-0.1", set.name+"/"+what, source(e, "order handler")
			f.Expected = expectedOf(e.RunEngagement(g))
			engagements = append(engagements, f)
		}
		set.run(t, e)
	}
	sceneSink = nil

	// The two large classes are thinned to keep the corpus small; the
	// coverage test compares every case against dosrules locally.
	var casualties []dosrules.CasualtyFixture
	perClass := map[string]int{}
	for i, c := range battleCases(e) {
		perClass[c.class]++
		if k := fixtureStride[c.class]; k > 1 && (perClass[c.class]-1)%k != 0 {
			continue
		}
		f := casualtyFixture(e, fmt.Sprintf("%s/%d", c.class, i), c)
		got, err := e.Battle(c.a, c.d, c.p)
		f.Expected = dosrules.CasualtyResult(dosrules.Result{AttackerSize: got.AttackerSize, AttackerAlt: got.AttackerAlt,
			DefenderSize: got.DefenderSize, DefenderAlt: got.DefenderAlt}, rulesError(err))
		casualties = append(casualties, f)
	}

	tables, err := json.MarshalIndent(tablesOf(e), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	eng, err := dosrules.MarshalFixtures(engagements)
	if err != nil {
		t.Fatal(err)
	}
	cas, err := dosrules.MarshalFixtures(casualties)
	if err != nil {
		t.Fatal(err)
	}
	rules := filepath.Join("..", "dosrules")
	files := map[string][]byte{
		filepath.Join(rules, e.B.Name+".json"):                         append(tables, '\n'),
		filepath.Join(rules, "testdata", e.B.Name, "engagements.json"): eng,
		filepath.Join(rules, "testdata", e.B.Name, "casualties.json"):  cas,
	}
	if e.B.Move != nil {
		tables, cases := movementFixtures(t, e)
		files[filepath.Join(rules, "testdata", e.B.Name, "movement-tables.json")] = tables
		files[filepath.Join(rules, "testdata", e.B.Name, "movements.json")] = cases
		files[filepath.Join(rules, "testdata", e.B.Name, "steps.json")] = stepFixtures(t, e)
		files[filepath.Join(rules, "testdata", e.B.Name, "containers.json")] = holdFixtures(t, e)
	}
	for path, b := range files {
		if *update {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		old, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(old, b) {
			t.Errorf("%s is not what the emulator gives; rerun with -update (%v)", path, err)
		}
	}
	t.Logf("%d engagements, %d casualty runs", len(engagements), len(casualties))
}
