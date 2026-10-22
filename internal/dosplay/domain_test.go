package dosplay

import (
	"maps"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// The battle domain of every UNIT.DAT type of the full build, and of the
// remake's units by id.
func TestTypeDomain(t *testing.T) {
	A, N, G, R, I := Air, Naval, Artillery, Armour, Infantry
	want := []Domain{
		R, R, // armoured car, A7V
		A, A, A, A, A, A, A, A, // the Central Powers' aircraft
		G, G, G, G, // light, medium and heavy artillery, bunker
		G, G, // mobile and static AA guns
		R,             // construction unit
		I, I, I, I, I, // elite infantry, infantry, anti-tank gun, sappers, cavalry
		R,       // supply car
		A,       // observation balloon
		G, G, R, // train artillery, armoured train, supply train
		N, N, N, N, N, N, N, N, N, // the boats and ships (two-cell ones with their rear)
		A, A, A, A, A, A, A, A, A, A, // the Entente's aircraft
		R, R, R, R, R, // Charron, FT-17, Mark I, Saint-Chamond, Mark IV
	}
	tables := dosrules.Full.Tables
	if len(want) != len(tables.Types) {
		t.Fatalf("%d domains for %d types", len(want), len(tables.Types))
	}
	for typ, d := range want {
		if got := TypeDomain(tables, typ); got != d {
			t.Errorf("type %d: %v, want %v", typ, got, d)
		}
	}
	if TypeDomain(tables, -1) != Infantry || TypeDomain(tables, len(tables.Types)) != Infantry {
		t.Error("a type the tables lack is not Infantry")
	}
	const balloon = 23
	set := dataSet(t)
	for id, d := range map[string]Domain{
		"CP_lArtillery": Artillery, "EC_hArtillery": Artillery, "CP_fokkerDrI": Air, "EC_camel": Air,
		"CP_cavalry": Infantry, "EC_infantry": Infantry, "CP_a7v": Armour, "EC_armCar": Armour, "CP_markIV": Armour,
		"CP_battleship": Naval, "EC_submarine": Naval, "CP_transShip": Naval,
	} {
		if got := TypeDomain(tables, set.Units[id].Dos); got != d {
			t.Errorf("%s: %v, want %v", id, got, d)
		}
	}
	// The domain follows the movement type of the remake's data; the
	// balloon, an aircraft to the rules, moves on the ground there.
	for _, id := range slices.Sorted(maps.Keys(set.Units)) {
		u := set.Units[id]
		if u.Dos == balloon {
			continue
		}
		d := TypeDomain(tables, u.Dos)
		if (u.MovementType == 3) != (d == Air) || (u.MovementType == 2) != (d == Naval) {
			t.Errorf("%s: movement type %d, domain %v", id, u.MovementType, d)
		}
	}
	for d := Infantry; d <= Air; d++ {
		if d.String() == "" || d.String() == Domain(99).String() {
			t.Errorf("domain %d: %q", int(d), d.String())
		}
	}
}

// The unit and type views carry the domain.
func TestDomainViews(t *testing.T) {
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{})
		for i, u := range s.Game.Units {
			d := TypeDomain(s.Game.Profile.Tables, u.Type)
			tv, _ := s.TypeInfo(0, u.Type)
			if s.DomainOf(i) != d || s.Unit(i).Domain != d || tv.Domain != d {
				t.Errorf("%s: unit %d of type %d: %v, %v, %v, want %v", lv.MapName, i, u.Type, s.DomainOf(i), s.Unit(i).Domain, tv.Domain, d)
			}
		}
		for _, v := range s.TrueView().Units {
			if v.Domain != s.DomainOf(v.Index) {
				t.Errorf("%s: unit %d's view has domain %v", lv.MapName, v.Index, v.Domain)
			}
		}
		if s.DomainOf(999) != Infantry {
			t.Errorf("%s: DomainOf(999) = %v", lv.MapName, s.DomainOf(999))
		}
	}
}
