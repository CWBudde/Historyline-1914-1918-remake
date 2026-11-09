package llmplay

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

// querySession is RHEIN (1918-02:cp) at its first phase, CP to move.
func querySession(t *testing.T) *dosplay.Session {
	t.Helper()
	lvl, err := ShipLevel("1918-02", scenario.CP)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(lvl, 0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustQuery(t *testing.T, s *dosplay.Session, q string) string {
	t.Helper()
	out, err := Query(s, 0, q)
	if err != nil {
		t.Fatalf("Query(%q): %v", q, err)
	}
	return out
}

// queryUnits are side 0's units and the enemies on its map, by hex order.
func queryUnits(s *dosplay.Session) (own, enemy []dosplay.UnitView) {
	v := s.View(0)
	for _, at := range sortedCoords(v.Units) {
		switch u := v.Units[at]; {
		case u.Head != nil:
		case u.Side == 0:
			own = append(own, u)
		case u.Side == 1:
			enemy = append(enemy, u)
		}
	}
	return own, enemy
}

func TestQueryDist(t *testing.T) {
	s := querySession(t)
	own, enemy := queryUnits(s)
	if len(own) == 0 || len(enemy) == 0 {
		t.Fatal("no units")
	}
	for i, u := range own {
		e := enemy[i%len(enemy)]
		want := fmt.Sprintf("%d hexes ", hexgrid.Distance(u.At, e.At))
		for _, q := range []string{
			fmt.Sprintf("dist u%d e%d", u.Index, e.Index),
			fmt.Sprintf("dist e%d u%d", e.Index, u.Index),
			fmt.Sprintf("dist %s %s", show(u.At), show(e.At)),
			fmt.Sprintf("dist %s u%d", show(e.At), u.Index),
		} {
			if got := mustQuery(t, s, q); !strings.HasPrefix(got, want) {
				t.Errorf("%s = %q, want %q...", q, got, want)
			}
		}
	}
}

func TestQueryErrors(t *testing.T) {
	s := querySession(t)
	for q, want := range map[string]string{
		"":                  "forms:",
		"fly u1":            "forms:",
		"dist u1":           "forms:",
		"dist [999,1] u1":   "off the map",
		"dist u9999 [1,1]":  "no own unit u9999",
		"reach e9999":       "no enemy unit e9999",
		"reach [1,1]":       "no enemy unit",
		"eta e9 [1,1]":      "no own unit",
		"near [1,1] x":      "radius",
		"type nosuchthing":  "no unit type",
		"threat e9999":      "no enemy unit e9999",
		"dist [1,1] banana": "no hex",
	} {
		_, err := Query(s, 0, q)
		if err == nil || !strings.Contains(err.Error(), want) || strings.Contains(err.Error(), "\n") {
			t.Errorf("Query(%q) error = %v, want one line with %q", q, err, want)
		}
	}
}

var (
	queryMoves = regexp.MustCompile(`: (\d+) moves`)
	queryStops = regexp.MustCompile(`stops ((?:\[\d+,\d+\] ?)+)`)
)

// A unit is at its own hex in 0 moves and at each legal destination in 1;
// a far goal's first stop is a legal destination.
func TestQueryEta(t *testing.T) {
	s := querySession(t)
	own, _ := queryUnits(s)
	v := s.View(0)
	var far hexgrid.Coord
	for at, bv := range v.Buildings {
		if bv.HQ && s.SideOf(bv.Owner) == 1 {
			far = at
		}
	}
	routed := 0
	for _, u := range own {
		dests, err := s.Destinations(0, u.Index)
		if err != nil || len(dests) == 0 {
			continue
		}
		if got := mustQuery(t, s, fmt.Sprintf("eta u%d %s", u.Index, show(u.At))); !strings.Contains(got, ": 0 moves") {
			t.Errorf("u%d to its own hex: %q", u.Index, got)
		}
		cs := dosplay.Coords(dests)
		for _, c := range []hexgrid.Coord{cs[0], cs[len(cs)-1]} {
			got := mustQuery(t, s, fmt.Sprintf("eta u%d %s", u.Index, show(c)))
			if !strings.Contains(got, ": 1 moves (resolution 1, now=1), stops "+show(c)+"\n") {
				t.Errorf("u%d to its destination %s: %q", u.Index, show(c), got)
			}
		}
		got := mustQuery(t, s, fmt.Sprintf("eta u%d %s", u.Index, show(far)))
		m, st := queryMoves.FindStringSubmatch(got), queryStops.FindStringSubmatch(got)
		if m == nil || st == nil {
			continue // it cannot get there
		}
		n := queryInt(m[1])
		stops := hexRef.FindAllStringSubmatch(st[1], -1)
		if len(stops) != n {
			t.Errorf("u%d: %d moves but %d stops: %q", u.Index, n, len(stops), got)
		}
		if first := parseCoord([2]int{queryInt(stops[0][1]), queryInt(stops[0][2])}); !dests[first] && n > 0 {
			t.Errorf("u%d: first stop %s is no legal destination: %q", u.Index, show(first), got)
		}
		routed++
	}
	if routed == 0 {
		t.Error("no unit found a way to the enemy HQ")
	}
	if got := mustQuery(t, s, fmt.Sprintf("eta u%d buildings", own[0].Index)); strings.Count(got, "\n") < 2 || strings.Count(got, "\n") > queryBuildings+2 {
		t.Errorf("eta buildings: %q", got)
	}
}

// reach and threat answer for every unit and name only units of the
// side's view at the hexes the view shows them on.
func TestQueryReachThreatFair(t *testing.T) {
	s := querySession(t)
	own, enemy := queryUnits(s)
	v := s.View(0)
	known := map[string]bool{}
	for at, u := range v.Units {
		known[fmt.Sprintf("%d%s", u.Index, show(at))] = true
	}
	unitAt := regexp.MustCompile(`\b[ue](\d+) [^\[\n]*(\[\d+,\d+\])`)
	check := func(q string) {
		t.Helper()
		out := mustQuery(t, s, q)
		for _, m := range unitAt.FindAllStringSubmatch(out, -1) {
			if !known[m[1]+m[2]] {
				t.Errorf("%s names unit %s at %s, which the view does not show there:\n%s", q, m[1], m[2], out)
			}
		}
	}
	for _, e := range enemy {
		check(fmt.Sprintf("reach e%d", e.Index))
	}
	for _, u := range own {
		check(fmt.Sprintf("threat u%d", u.Index))
		check("threat " + show(u.At))
	}
	check("threat [1,1]")
	check(fmt.Sprintf("near u%d 6", own[0].Index))
}

func TestQueryType(t *testing.T) {
	s := querySession(t)
	own, enemy := queryUnits(s)
	seen := map[string]bool{}
	for _, u := range append(own, enemy...) {
		name := "u"
		if u.Side != 0 {
			name = "e"
		}
		byUnit := mustQuery(t, s, "type "+name+strconv.Itoa(u.Index))
		if seen[u.ID] {
			continue
		}
		seen[u.ID] = true
		head := fmt.Sprintf("type%d %s ", u.Type, u.ID)
		for _, q := range []string{"type " + u.ID, fmt.Sprintf("type type%d", u.Type)} {
			got := mustQuery(t, s, q)
			if q == "type "+u.ID && got != byUnit {
				t.Errorf("%s = %q, by unit %q", q, got, byUnit)
			}
			if !strings.Contains(got, "weapons:") || !strings.Contains(got, "flags:") || !strings.HasPrefix(got, "type"+strconv.Itoa(u.Type)+" ") {
				t.Errorf("%s = %q, want %q...", q, got, head)
			}
		}
	}
}

// The income lines add up to the header's incomes.
func TestQueryIncome(t *testing.T) {
	s := querySession(t)
	out := mustQuery(t, s, "income")
	st := s.Stats()
	plus := regexp.MustCompile(`\+(\d+)`)
	lines := strings.Split(out, "\n")
	for k, who := range []string{"you CP", "enemy EC"} {
		if !strings.HasPrefix(lines[k], fmt.Sprintf("%s: energy %d, income %d ", who, s.Energy(k), st.Income[k])) {
			t.Errorf("line %d: %q", k, lines[k])
		}
		sum := 0
		for _, m := range plus.FindAllStringSubmatch(lines[k], -1) {
			sum += queryInt(m[1])
		}
		if sum != st.Income[k] {
			t.Errorf("%s buildings pay %d, income is %d: %q", who, sum, st.Income[k], lines[k])
		}
	}
}

// The reach estimate used for enemies, run for the side's own units, is
// the engine's destination set but for loading (a carrier or a building on
// the hex): it never invents a hex, and seldom misses one.
func TestQueryReachEstimate(t *testing.T) {
	for _, month := range []string{"1918-02", "1915-12", "1916-05"} {
		lvl, err := ShipLevel(month, scenario.CP)
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewSession(lvl, 0)
		if err != nil {
			t.Fatal(err)
		}
		x := &querier{s: s, side: 0, v: s.View(0)}
		units, same := 0, 0
		for _, u := range x.ownOnMap() {
			dests, err := s.Destinations(0, u.Index)
			if err != nil || u.Head != nil {
				continue
			}
			costs, err := s.Costs(0, u.Index)
			if err != nil {
				t.Fatal(err)
			}
			got := map[hexgrid.Coord]bool{}
			for _, c := range x.viewReach(u, costs) {
				got[c] = true
				if !dests[c] {
					t.Errorf("%s u%d %s: %s is no destination", month, u.Index, u.ID, show(c))
				}
			}
			missed := 0
			for c := range dests {
				if _, taken := x.v.Units[c]; !got[c] && !taken {
					missed++
				}
			}
			units++
			if missed == 0 {
				same++
			} else {
				t.Logf("%s u%d %s: %d of %d destinations missed", month, u.Index, u.ID, missed, len(dests))
			}
		}
		if units == 0 || same*10 < units*9 {
			t.Errorf("%s: the estimate is the engine's for %d of %d units", month, same, units)
		}
	}
}
