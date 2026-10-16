package game

import (
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// snapshot is the part of a unit an attack may change.
type snapshot struct {
	Size, Experience int
	Ammo             [4]int
	Attacked         bool
}

func snap(u *Unit) snapshot {
	s := snapshot{Size: u.Size, Experience: u.Experience, Attacked: u.Attacked}
	for i, a := range u.Ammo {
		s.Ammo[i] = -1
		if a != nil {
			s.Ammo[i] = *a
		}
	}
	return s
}

func TestPredictMatchesAttack(t *testing.T) {
	cases := []struct {
		name      string
		r         Rules
		att, def  string
		from, to  hexgrid.Coord
		slot      int
		aSize     int
		dSize     int
		wantKill  bool
		wantDeath bool
	}{
		{name: "rifle Standard", r: Standard, att: "CP_infantry", def: "EC_infantry", from: at(4, 2), to: at(6, 2), slot: 2},
		{name: "rifle Classic", r: Classic, att: "CP_infantry", def: "EC_infantry", from: at(4, 2), to: at(6, 2), slot: 2},
		{name: "artillery at range", r: Standard, att: "CP_lArtillery", def: "EC_infantry", from: at(4, 2), to: at(8, 2), slot: 0},
		{name: "kills the defender", r: Standard, att: "CP_lArtillery", def: "EC_infantry", from: at(4, 2), to: at(6, 2), slot: 0, dSize: 1, wantKill: true},
		{name: "loses the attacker", r: Classic, att: "CP_infantry", def: "EC_lArtillery", from: at(4, 2), to: at(6, 2), slot: 0, aSize: 1, wantDeath: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := boardWith(t, c.r, map[hexgrid.Coord]level.Hex{c.from: unit(c.att), c.to: unit(c.def)})
			a, d := s.At(c.from).Unit, s.At(c.to).Unit
			if c.aSize > 0 {
				a.Size = c.aSize
			}
			if c.dSize > 0 {
				d.Size = c.dSize
			}
			wa, wd := snap(a), snap(d)
			p, err := s.Predict(c.from, c.to, c.slot)
			if err != nil {
				t.Fatal(err)
			}
			if snap(a) != wa || snap(d) != wd || s.At(c.from).Unit != a || s.At(c.to).Unit != d {
				t.Fatalf("Predict changed the units: %+v %+v", snap(a), snap(d))
			}
			if p.DefenderDestroyed != c.wantKill || p.AttackerDestroyed != c.wantDeath {
				t.Errorf("Predict %+v, want kill %v death %v", p, c.wantKill, c.wantDeath)
			}
			b, err := s.Attack(c.from, c.to, c.slot, zero)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(p, b) {
				t.Errorf("Predict %+v, Attack %+v", p, b)
			}
		})
	}
}

func TestPredictErrorsLikeAttack(t *testing.T) {
	setups := []struct {
		name  string
		from  hexgrid.Coord
		to    hexgrid.Coord
		slot  int
		setup func(s *State)
	}{
		{name: "empty slot", from: at(4, 2), to: at(6, 2), slot: 3},
		{name: "slot out of bounds", from: at(4, 2), to: at(6, 2), slot: 7},
		{name: "no target", from: at(4, 2), to: at(8, 2), slot: 2},
		{name: "no unit", from: at(0, 0), to: at(6, 2), slot: 2},
		{name: "enemy unit", from: at(6, 2), to: at(4, 2), slot: 2},
		{name: "moved", from: at(4, 2), to: at(6, 2), slot: 2, setup: func(s *State) { s.At(at(4, 2)).Unit.Moved = true }},
		{name: "no ammo", from: at(4, 2), to: at(6, 2), slot: 2, setup: func(s *State) { *s.At(at(4, 2)).Unit.Ammo[2] = 0 }},
		{name: "game over", from: at(4, 2), to: at(6, 2), slot: 2, setup: func(s *State) { s.Winner = Entente }},
	}
	for _, c := range setups {
		t.Run(c.name, func(t *testing.T) {
			s, _, _ := duel(t, Standard, "CP_infantry", "EC_infantry")
			if c.setup != nil {
				c.setup(s)
			}
			_, perr := s.Predict(c.from, c.to, c.slot)
			_, aerr := s.Attack(c.from, c.to, c.slot, zero)
			if perr == nil || aerr == nil || perr.Error() != aerr.Error() {
				t.Errorf("Predict error %v, Attack error %v", perr, aerr)
			}
		})
	}
}
