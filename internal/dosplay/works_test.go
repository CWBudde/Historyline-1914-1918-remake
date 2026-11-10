package dosplay

import (
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// resolveAll confirms both sides and resolves once, returning the events.
func resolveAll(t *testing.T, s *Session) []Event {
	t.Helper()
	s.Confirm(0)
	s.Confirm(1)
	evs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

// ofKind are the events of kind k.
func ofKind(evs []Event, k EventKind) []Event {
	var out []Event
	for _, ev := range evs {
		if ev.Kind == k {
			out = append(out, ev)
		}
	}
	return out
}

// A wiring party wires up to three hexes of open ground; the map draws
// them as grass with barbed wire.
func TestWiringParty(t *testing.T) {
	from := hexgrid.Coord{Line: 4, Col: 3}
	to := hexgrid.Around(from)[3]
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "CP_wiring", {Line: 9, Col: 3}: "EC_infantry"}), Options{})
	party := s.View(0).Units[from].Index
	if !s.Wiring(party) || s.Charges(party) != dosrules.WireCharges {
		t.Fatalf("wiring %v, charges %d", s.Wiring(party), s.Charges(party))
	}
	passUntilAction(t, s, 0)
	if err := s.CanWire(0, party); err != nil {
		t.Fatal(err)
	}
	if !s.Targets(0, party)[to] {
		t.Fatalf("targets %v", s.Targets(0, party))
	}
	hexes := s.WireHexes(0, party, to)
	if len(hexes) != 3 || !slices.Contains(hexes, to) {
		t.Fatalf("wire hexes %v", hexes)
	}
	evs, err := s.Attack(0, party, to)
	if err != nil || len(evs) != 1 || evs[0].Kind != Tasked || evs[0].Text != "Wire laying ordered." {
		t.Fatalf("order %+v, %v", evs, err)
	}
	evs = resolveAll(t, s)
	w := ofKind(evs, Wired)
	if len(w) != 1 || w[0].From != from || w[0].To != to || !slices.Equal(w[0].Hexes, hexes) {
		t.Fatalf("wired %+v", w)
	}
	re := s.Reworked()
	for _, c := range hexes {
		if h := re[c]; h.Terrain != "grass" || !slices.Equal(h.Infra.IDs, []string{"barbedWire"}) {
			t.Errorf("hex %v drawn as %+v", c, h)
		}
	}
	if s.Charges(party) != dosrules.WireCharges-1 {
		t.Errorf("charges %d", s.Charges(party))
	}
}

// Tunnellers' mine blows two turns after it opens: the unit on the hex
// loses most of its group and the hex becomes a crater field.
func TestTunnellers(t *testing.T) {
	from := hexgrid.Coord{Line: 4, Col: 3}
	to := hexgrid.Around(from)[3]
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "CP_tunnellers", to: "EC_infantry"}), Options{})
	tun, enemy := s.View(0).Units[from].Index, s.View(0).Units[to].Index
	passUntilAction(t, s, 0)
	if !s.Tunneller(tun) || s.CanTunnel(0, tun) != nil || !s.Targets(0, tun)[to] {
		t.Fatalf("tunneller %v, %v, targets %v", s.Tunneller(tun), s.CanTunnel(0, tun), s.Targets(0, tun))
	}
	if evs, err := s.Attack(0, tun, to); err != nil || evs[0].Text != "Tunnelling ordered." {
		t.Fatalf("order %+v, %v", evs, err)
	}
	var blown, mined []Event
	for range 6 {
		evs := resolveAll(t, s)
		if ofKind(evs, Tunnelled) != nil {
			if tn := s.Tunnels(0); len(tn) != 1 || tn[0].Mine != to || tn[0].From != from || s.Tunnels(1) != nil {
				t.Fatalf("tunnels %+v / %+v", tn, s.Tunnels(1))
			}
		}
		blown = append(blown, ofKind(evs, Blown)...)
		mined = append(mined, ofKind(evs, Mined)...)
		if blown != nil {
			break
		}
	}
	if len(blown) != 1 || blown[0].To != to || !slices.Equal(blown[0].Hexes, []hexgrid.Coord{to}) {
		t.Fatalf("blown %+v", blown)
	}
	if len(mined) != 1 || mined[0].Unit != enemy || mined[0].Loss == 0 || !strings.HasSuffix(mined[0].Text, "to a mine.") {
		t.Fatalf("mined %+v", mined)
	}
	if h := s.Reworked()[to]; h.Terrain != "mud" || len(h.Infra.IDs) != 0 {
		t.Errorf("the crater drawn as %+v", h)
	}
	if _, ok := s.TunnelOf(tun); ok || s.Tunnels(0) != nil {
		t.Error("the tunnel outlived its mine")
	}
}

// A field ambulance heals the foot units of its side next to it once per
// turn.
func TestFieldAmbulance(t *testing.T) {
	from := hexgrid.Coord{Line: 4, Col: 3}
	next := hexgrid.Around(from)[3]
	s := session(t, grassLevel(map[hexgrid.Coord]string{
		from: "CP_ambulance", next: "CP_infantry", {Line: 9, Col: 3}: "EC_infantry",
	}), Options{})
	amb, inf := s.View(0).Units[from].Index, s.View(0).Units[next].Index
	u := s.Game.Units[inf]
	u.Size = 3
	s.Game.Units[inf] = u
	if !s.Ambulance(amb) || s.Ambulance(inf) {
		t.Fatal("Ambulance")
	}
	var healed []Event
	for range 2 {
		healed = append(healed, ofKind(resolveAll(t, s), Healed)...)
	}
	if len(healed) != 1 || healed[0].Unit != inf || healed[0].Gain != dosrules.AmbulanceHeal || healed[0].To != next ||
		!strings.HasSuffix(healed[0].Text, "regain 1.") {
		t.Fatalf("healed %+v", healed)
	}
	if s.Game.Units[inf].Size != 3+dosrules.AmbulanceHeal {
		t.Errorf("size %d", s.Game.Units[inf].Size)
	}
}

// The option brings only the extension types of the level's month into
// the factories.
func TestExtensionBuildsByMonth(t *testing.T) {
	for _, c := range []struct {
		month    string
		has, not []int
	}{
		{"1918-02", []int{51, 52, 53, 57, 59, 60, 61, 62}, []int{54, 55, 56, 58}},
		{"1915-05", []int{51, 52, 54, 59}, []int{53, 55, 56, 57, 58}},
	} {
		lv, err := scenario.ShippedLevel(c.month, scenario.CP)
		if err != nil {
			t.Fatal(err)
		}
		s := session(t, lv, Options{ExtendedUnits: true})
		for i, bd := range s.Game.Buildings {
			if dosrules.BuildingKind(i) != dosrules.KindFactory {
				continue
			}
			for _, typ := range c.has {
				if !slices.Contains(bd.Builds, typ) {
					t.Errorf("%s: factory %d lacks type %d", c.month, i, typ)
				}
			}
			for _, typ := range c.not {
				if slices.Contains(bd.Builds, typ) {
					t.Errorf("%s: factory %d builds type %d", c.month, i, typ)
				}
			}
		}
	}
}
