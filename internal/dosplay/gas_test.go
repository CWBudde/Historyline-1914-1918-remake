package dosplay

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// A level that places gas pioneers plays on the extended rules: their order
// releases a cloud both sides see, and the units in it lose strength.
func TestGasPioneers(t *testing.T) {
	from := hexgrid.Coord{Line: 4, Col: 3}
	to := hexgrid.Around(from)[3]
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "CP_gasPioneers", to: "EC_infantry"}), Options{})
	if s.Game.Profile.Tables != dosrules.ExtendedTables {
		t.Fatalf("tables %q", s.Game.Profile.Tables.Build)
	}
	pioneer, enemy := s.View(0).Units[from].Index, s.View(0).Units[to].Index
	if !s.GasPioneer(pioneer) || s.GasPioneer(enemy) || s.Charges(pioneer) != dosrules.GasCharges || s.Charges(enemy) != -1 {
		t.Fatalf("pioneer %v, charges %d", s.GasPioneer(pioneer), s.Charges(pioneer))
	}
	passUntilAction(t, s, 0)
	if err := s.CanRelease(0, pioneer); err != nil {
		t.Fatal(err)
	}
	targets := s.Targets(0, pioneer)
	for _, c := range hexgrid.Around(from) {
		if !targets[c] {
			t.Errorf("neighbour %v is no target", c)
		}
	}
	cloud := s.Cloud(0, pioneer, to)
	if len(cloud) != 3 || !slices.Contains(cloud, to) || s.Cloud(0, enemy, from) != nil {
		t.Fatalf("cloud %v", cloud)
	}
	evs, err := s.Attack(0, pioneer, to)
	if err != nil || len(evs) != 1 || evs[0].Kind != Tasked || evs[0].Text != "Gas release ordered." {
		t.Fatalf("order %+v, %v", evs, err)
	}
	if u := s.View(0).Units[from]; u.Target == nil || *u.Target != to {
		t.Errorf("the pioneers' view: %+v", u)
	}
	s.Confirm(0)
	s.Confirm(1)
	evs, err = s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	var released, gassed []Event
	for _, ev := range evs {
		switch ev.Kind {
		case Released:
			released = append(released, ev)
		case Gassed:
			gassed = append(gassed, ev)
		case Battle:
			t.Errorf("a battle: %+v", ev)
		}
	}
	if len(released) != 1 || released[0].From != from || released[0].To != to || !slices.Equal(released[0].Hexes, cloud) {
		t.Fatalf("released %+v", released)
	}
	// Side 0's resolution counts the turn: the infantry loses at once.
	if len(gassed) != 1 || gassed[0].Unit != enemy || gassed[0].To != to || gassed[0].Loss != 2 ||
		!strings.HasSuffix(gassed[0].Text, "to gas.") {
		t.Fatalf("gassed %+v", gassed)
	}
	gas := s.Gas()
	if len(gas) != 3 || gas[to] != dosrules.GasTurns-1 || s.GasAt(to) != gas[to] || s.Charges(pioneer) != 1 {
		t.Errorf("gas %v, charges %d", gas, s.Charges(pioneer))
	}

	saved, err := s.Save()
	if err != nil {
		t.Fatal(err)
	}
	r, err := LoadBoard(s.Board, s.Level, dataSet(t), saved)
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(r.Gas(), gas) {
		t.Errorf("after Load: %v", r.Gas())
	}
}

// A level of the original's units plays on the full rules; the option
// brings the extension units into every factory, and the month sets the gas.
func TestExtendedUnitsOption(t *testing.T) {
	lv, err := scenario.ShippedLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s := session(t, lv, Options{})
	if s.Game.Profile.Tables != dosrules.FullTables {
		t.Fatalf("tables %q", s.Game.Profile.Tables.Build)
	}
	s = session(t, lv, Options{ExtendedUnits: true})
	if s.Game.Profile.Tables != dosrules.ExtendedTables || s.Game.GasTurns != 4 || s.Game.GasLoss != 15 {
		t.Fatalf("tables %q, gas %d turns %d %%", s.Game.Profile.Tables.Build, s.Game.GasTurns, s.Game.GasLoss)
	}
	factories := 0
	for i, bd := range s.Game.Buildings {
		if dosrules.BuildingKind(i) == dosrules.KindFactory {
			factories++
			if !slices.Contains(bd.Builds, 51) {
				t.Errorf("factory %d builds %v", i, bd.Builds)
			}
		}
	}
	if factories == 0 {
		t.Error("the level has no factory")
	}
	for month, want := range map[string][2]int{"": {2, 30}, "1915-04": {2, 30}, "1916-01": {2, 15}, "1917-07": {4, 15}} {
		if turns, loss := GasRates(month); turns != want[0] || loss != want[1] {
			t.Errorf("GasRates(%q) = %d, %d", month, turns, loss)
		}
	}
}
