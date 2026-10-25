package game

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

func zero() int { return 0 }

func duel(t *testing.T, att, def string) (*State, *Unit, *Unit) {
	t.Helper()
	s := board(t, map[hexgrid.Coord]level.Hex{at(4, 2): unit(att), at(6, 2): unit(def)})
	return s, s.At(at(4, 2)).Unit, s.At(at(6, 2)).Unit
}

func TestOneActionPerTurn(t *testing.T) {
	// Attacking after a move, and moving after an attack, are refused.
	s, a, _ := duel(t, "CP_infantry", "EC_infantry")
	a.Moved = true
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err == nil {
		t.Error("attack after a move succeeded")
	}
	a.Moved = false
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err == nil {
		t.Error("second attack succeeded")
	}
	if _, err := s.Move(at(4, 2), a, at(2, 2)); err == nil {
		t.Error("move after an attack succeeded")
	}
}

func TestAmmoPerShot(t *testing.T) {
	s, a, d := duel(t, "CP_infantry", "EC_infantry")
	if *a.Ammo[2] != 3 || *d.Ammo[2] != 3 {
		t.Fatalf("start ammo %d/%d, want 3", *a.Ammo[2], *d.Ammo[2])
	}
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err != nil {
		t.Fatal(err)
	}
	if *a.Ammo[2] != 2 || *d.Ammo[2] != 2 {
		t.Errorf("ammo after a shot %d/%d, want 2/2 (shot and counter)", *a.Ammo[2], *d.Ammo[2])
	}
	a.Attacked = false
	if _, err := s.Attack(at(4, 2), at(6, 2), 0, zero); err != nil || a.Ammo[0] != nil {
		t.Errorf("bayonet: %v, ammo %v (unlimited stays nil)", err, a.Ammo[0])
	}

	// Empty weapon: refused, and it no longer gives range.
	a.Attacked = false
	*a.Ammo[2] = 0
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err == nil {
		t.Error("shot with an empty rifle succeeded")
	}
	s2, art, _ := duel(t, "CP_lArtillery", "EC_infantry")
	if len(s2.Targets(at(4, 2))) != 1 {
		t.Fatalf("artillery Targets = %v", s2.Targets(at(4, 2)))
	}
	*art.Ammo[0] = 0
	if got := s2.Targets(at(4, 2)); len(got) != 0 {
		t.Errorf("empty artillery Targets = %v", got)
	}
}

func TestTargetFlags(t *testing.T) {
	// A synthetic weapon without the ground flag hits nothing on land.
	s, a, _ := duel(t, "CP_infantry", "EC_infantry")
	w := *a.Weapons[2]
	w.Ground = false
	a.Weapons[2] = &w
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err == nil {
		t.Error("non-ground weapon hit a land unit")
	}
	if got := s.Targets(at(4, 2)); len(got) != 1 {
		t.Errorf("Targets = %v, the other weapons still reach", got)
	}
	// ...and a land unit's non-ground weapon does not counter.
	s, a, d := duel(t, "CP_infantry", "EC_infantry")
	for i, w := range d.Weapons {
		if w != nil {
			c := *w
			c.Ground = false
			d.Weapons[i] = &c
		}
	}
	if b, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err != nil || b.Counter != -1 {
		t.Errorf("counter %d, err %v, want no counter", b.Counter, err)
	}
	_ = a

	// The infantry rifle has power against water and hits a boat; the bayonet,
	// a ground-only weapon, does not.
	s, _, _ = duel(t, "CP_infantry", "EC_patrolBoat")
	s.At(at(6, 2)).Terrain = "water"
	if _, err := s.Attack(at(4, 2), at(6, 2), 2, zero); err != nil {
		t.Errorf("rifle on a boat: %v", err)
	}
	if _, err := s.Attack(at(4, 2), at(6, 2), 0, zero); err == nil {
		t.Error("a ground-only weapon hit a boat")
	}
	// A water-only weapon does not hit land but hits boats.
	ww := *mustUnit(t, "CP_infantry").Weapons[2]
	ww.Ground, ww.Water = false, true
	c := &ww
	if canHit(c, mustUnit(t, "EC_infantry")) || !canHit(c, mustUnit(t, "EC_patrolBoat")) {
		t.Error("water-only weapon: wrong target classes")
	}
}

func TestNoCounterNoDamage(t *testing.T) {
	// Artillery at range 3 gets no answer from infantry (range 1), and a
	// defender without a counter weapon deals no counter damage.
	s := board(t, map[hexgrid.Coord]level.Hex{at(2, 2): unit("CP_lArtillery"), at(8, 2): unit("EC_infantry"), at(7, 2): unit("CP_infantry")})
	b, err := s.Attack(at(2, 2), at(8, 2), 0, func() int { return 5 })
	if err != nil {
		t.Fatal(err)
	}
	if a := s.At(at(2, 2)).Unit; b.Counter != -1 || b.CounterDamage != 0 || a.Size != a.MaxSize {
		t.Errorf("%+v, attacker size %d", b, a.Size)
	}
	a, d := mustUnit(t, "CP_lArtillery"), mustUnit(t, "EC_infantry")
	a.Size, a.Experience = 10, 0
	d.Size, d.Experience = 10, 10
	if _, _, cd := resolve(a, d, 0, 4, 0, 100, func() int { return 5 }); cd != 0 {
		t.Errorf("counter damage %d, want 0 without a counter weapon", cd)
	}
}

func mustUnit(t *testing.T, id string) *Unit {
	t.Helper()
	u, err := NewUnit(loadSet(t), id)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
