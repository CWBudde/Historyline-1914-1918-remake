package dosoracle

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

// The field census looks for ammo, fuel and resupply. Every executed runner
// family is repeated with the machine's write watch on, and each write an
// instruction makes into the live unit records or the type records is
// counted by field, writing instruction and direction. Python's ammo and
// fuel would need a field that firing or moving lowers, and resupply one
// that something raises again; the census finds neither.

// liveFields names the live record bytes the executed runners may write.
var liveFields = map[int]string{
	0x01: "experience", 0x02: "group size", 0x03: "alternate strength",
	0x04: "flags", 0x05: "flags",
	0x0b: "side 0 position", 0x0c: "side 0 position", 0x0d: "side 1 position", 0x0e: "side 1 position",
	0x0f: "side 0 facing", 0x10: "side 1 facing", 0x11: "target or start", 0x12: "target or start",
	0x13: "action", 0x14: "action flags", 0x15: "action", 0x16: "action", 0x17: "action",
	0x18: "action", 0x19: "action",
}

// lowered are the fields a write may lower: the casualties' group size, flag
// bits cleared, and positions, facings and targets, which take any value.
var lowered = map[int]bool{0x02: true, 0x04: true, 0x05: true, 0x0b: true, 0x0c: true, 0x0d: true, 0x0e: true,
	0x0f: true, 0x10: true, 0x11: true, 0x12: true}

// strengthFields would be raised by a resupply: group size, alternate
// strength and the construction charges (+0x0a).
var strengthFields = []int{0x02, 0x03, 0x0a}

// site is a writing instruction: image segment and offset.
type site [2]int

func (s site) String() string { return fmt.Sprintf("%04x:%04x", s[0], s[1]) }

// census counts the writes of one runner family.
type census struct {
	live      [26]map[site][3]int // by field and site: writes lowering, keeping, raising the byte
	typeSites map[site]int        // writes into the type records
	run       map[[2]int]int      // writes by unit and field in the current run
}

func newCensus() *census {
	c := &census{typeSites: map[site]int{}}
	for k := range c.live {
		c.live[k] = map[site][3]int{}
	}
	return c
}

// watching is a copy of e whose machines report their writes into c.
func watching(e *Export, c *census) *Export {
	w := *e
	data := baseSeg + e.B.Data
	units := dos8086.Lin(data, e.B.Units)
	types := dos8086.Lin(data, e.B.Handler.Types)
	w.watch = func(m *dos8086.Machine) {
		m.OnWrite = func(mm *dos8086.Machine, lin, wd, v int) {
			at := site{int(mm.S[dos8086.CS]) - baseSeg, int(mm.At)}
			for i := range wd + 1 {
				a, b := lin+i, v>>(8*i)&0xFF
				switch {
				case a >= types && a < types+len(e.Units):
					c.typeSites[at]++
				case a >= units && a < units+26*liveUnits:
					k := (a - units) % 26
					dir := 1
					switch old := int(mm.Mem[a]); {
					case b < old:
						dir = 0
					case b > old:
						dir = 2
					}
					n := c.live[k][at]
					n[dir]++
					c.live[k][at] = n
					c.run[[2]int{(a - units) / 26, k}]++
				}
			}
		}
	}
	return &w
}

// begin starts counting the writes of one run.
func (c *census) begin() { c.run = map[[2]int]int{} }

func bySite(a, b site) int {
	if a[0] != b[0] {
		return a[0] - b[0]
	}
	return a[1] - b[1]
}

// check applies the census rules to a family and counts them in seen.
func (c *census) check(t *testing.T, family string, seen map[string]int) {
	t.Helper()
	for _, s := range slices.SortedFunc(maps.Keys(c.typeSites), bySite) {
		t.Errorf("%s: %v wrote into the type records %d times", family, s, c.typeSites[s])
	}
	if len(c.typeSites) == 0 {
		seen[family+": type records untouched"]++
	}
	for k, ws := range c.live {
		for _, s := range slices.SortedFunc(maps.Keys(ws), bySite) {
			n := ws[s]
			t.Logf("%s: live +%#02x by %v: lowered %d, kept %d, raised %d", family, k, s, n[0], n[1], n[2])
			if _, ok := liveFields[k]; !ok {
				t.Errorf("%s: %v wrote unnamed live field +%#02x", family, s, k)
			}
			if n[0] > 0 {
				if !lowered[k] {
					t.Errorf("%s: %v lowered live +%#02x %d times: a consumed field", family, s, k, n[0])
				}
				seen[fmt.Sprintf("%s: lowered %s", family, liveFields[k])]++
			}
			if slices.Contains(strengthFields, k) && n[2] > 0 {
				t.Errorf("%s: %v raised live +%#02x %d times: a resupply", family, s, k, n[2])
			}
		}
	}
	if len(c.live[0]) == 0 {
		seen[family+": movement budget never written"]++
	}
}

func TestConsumables(t *testing.T) { eachExport(t, consumables) }

func consumables(t *testing.T, e *Export) {
	needHandler(t, e)
	needMove(t, e)
	if e.B.Handler.Phase[1] == 0 {
		t.Skip("the phase change is not located in this build")
	}
	tab := moveTablesOf(e)
	seen := map[string]int{}
	ran := map[string]int{}

	eng := newCensus()
	we := watching(e, eng)
	sceneSink = func(_ string, g *Engagement) {
		eng.begin()
		o, err := we.RunEngagement(g)
		if err != nil {
			t.Fatal(err)
		}
		ran["engagement"]++
		if o.Battle && o.AttackerSize < g.Units[g.Attacker].Size {
			seen["engagement: the attacker lost members"]++
		}
	}
	for _, run := range []func(*testing.T, *Export){engagementDuels, engagementBlocking, engagementPincer,
		engagementRanged, engagementExperience, engagementHidden} {
		run(t, e)
	}
	sceneSink = nil
	eng.check(t, "engagement", seen)

	// Fuel would be spent per cell walked: no unit's field but positions and
	// facings is written more than perRun times in a move, while routes run
	// longer.
	const perRun = 4
	mv := newCensus()
	wm := watching(e, mv)
	longest := 0
	for _, c := range stepCases(tab) {
		mv.begin()
		r, err := wm.RunMove(c.g, c.o)
		if err != nil {
			continue
		}
		ran["move"]++
		longest = max(longest, len(r.Cells))
		for uk, n := range mv.run {
			if k := uk[1]; (k < 0x0b || k > 0x10) && n > perRun {
				t.Errorf("move %s/%s: unit %d's live +%#02x written %d times over %d cells", c.class, c.name, uk[0], k, n, len(r.Cells))
			}
		}
	}
	if longest > perRun {
		seen["move: per-cell writes only to positions and facings"]++
	}
	mv.check(t, "move", seen)

	ho := newCensus()
	wh := watching(e, ho)
	for _, c := range holdCases(tab) {
		ho.begin()
		var err error
		if c.o.Hold < 0 {
			_, err = wh.RunMove(c.g, c.o.MoveOrder)
		} else {
			_, err = wh.RunExit(c.g, c.o)
		}
		if err == nil {
			ran["container"]++
		}
	}
	ho.check(t, "container", seen)

	ph := newCensus()
	wp := watching(e, ph)
	for _, c := range phaseCases() {
		ph.begin()
		if _, err := wp.RunPhaseChange(c.g); err == nil {
			ran["phase"]++
		}
	}
	ph.check(t, "phase", seen)

	t.Logf("runs %v, longest route %d cells", ran, longest)
	keys := slices.Sorted(maps.Keys(seen))
	for _, k := range keys {
		t.Logf("%s: %d", k, seen[k])
	}
	for _, f := range []string{"engagement", "move", "container", "phase"} {
		if ran[f] == 0 {
			t.Errorf("no %s ran", f)
		}
		for _, k := range []string{": type records untouched", ": movement budget never written"} {
			if seen[f+k] == 0 {
				t.Errorf("no %s%s", f, k)
			}
		}
	}
	for _, k := range []string{"engagement: lowered group size", "engagement: the attacker lost members",
		"move: per-cell writes only to positions and facings", "move: lowered side 0 position",
		"container: lowered flags", "phase: lowered side 0 facing"} {
		if seen[k] == 0 {
			t.Errorf("no case showed %q", k)
		}
	}
}
