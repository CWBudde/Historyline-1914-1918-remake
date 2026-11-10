package dosai

import (
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func TestTunnellersMineAnEnemy(t *testing.T) {
	from := hexgrid.Coord{Line: 6, Col: 3}
	enemy := hexgrid.Around(from)[3]
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_tunnellers", enemy: "CP_infantry"}))
	tasked := digs(t, s)
	if len(tasked) != 1 || tasked[0].From != from || tasked[0].To != enemy || !s.Tunneller(tasked[0].Unit) {
		t.Fatalf("orders %+v", tasked)
	}

	// No enemy next to them: no tunnel.
	s = session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_tunnellers", {Line: 18, Col: 3}: "CP_infantry"}))
	if tasked := digs(t, s); len(tasked) != 0 {
		t.Errorf("a tunnel without a target: %+v", tasked)
	}
}

func TestWiringPartyWiresTheFront(t *testing.T) {
	from := hexgrid.Coord{Line: 6, Col: 3}
	s := session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_wiring", {Line: 10, Col: 3}: "CP_infantry"}))
	tasked := digs(t, s)
	if len(tasked) != 1 || tasked[0].From != from || !s.Wiring(tasked[0].Unit) {
		t.Fatalf("orders %+v", tasked)
	}
	if hexgrid.Distance(tasked[0].To, hexgrid.Coord{Line: 10, Col: 3}) >= hexgrid.Distance(from, hexgrid.Coord{Line: 10, Col: 3}) {
		t.Errorf("the wire at %v is not toward the enemy", tasked[0].To)
	}

	// The enemy far away: no wire.
	s = session(t, grassLevel(map[hexgrid.Coord]string{from: "EC_wiring", {Line: 18, Col: 3}: "CP_infantry"}))
	if tasked := digs(t, s); len(tasked) != 0 {
		t.Errorf("wire far from the enemy: %+v", tasked)
	}
}

// In a played-out battle the computer side wires, mines and heals with the
// new work units in place of some infantry.
func TestWorksPlayOut(t *testing.T) {
	lv, err := scenario.ShippedLevel("1915-05", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	swap := []string{"_wiring", "_tunnellers", "_ambulance", "_flamethrower"}
	n := 0
	for l, line := range lv.Lines {
		for c, h := range line {
			if strings.HasSuffix(h.Unit, "_infantry") {
				if n%2 == 0 {
					lv.Lines[l][c].Unit = strings.TrimSuffix(h.Unit, "_infantry") + swap[n/2%len(swap)]
				}
				n++
			}
		}
	}
	s := session(t, lv)
	_, evs := play(t, s, [2]*Planner{{}, {}})
	count := map[dosplay.EventKind]int{}
	for _, ev := range evs {
		count[ev.Kind]++
	}
	t.Logf("wired %d, tunnels %d (lost %d), mines %d hitting %d units, healed %d",
		count[dosplay.Wired], count[dosplay.Tunnelled], count[dosplay.TunnelLost], count[dosplay.Blown], count[dosplay.Mined], count[dosplay.Healed])
	if count[dosplay.Wired] == 0 || count[dosplay.Healed] == 0 || count[dosplay.Tunnelled] == 0 {
		t.Errorf("the work units did nothing: %v", count)
	}
}
