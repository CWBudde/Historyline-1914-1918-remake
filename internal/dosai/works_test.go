package dosai

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
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
