package dosbind

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// TestExtensionUnitsFollowTables checks the remake entries of the extension
// types (docs/new-units.md) against dosrules' extension records, as
// TestUnitsFollowUnitDat checks the original types against UNIT.DAT: the
// stats and the ground and water attack (the range byte is one more than
// the reach).
func TestExtensionUnitsFollowTables(t *testing.T) {
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	x := dosrules.ExtendedTables
	seen := map[int]int{}
	for id, u := range set.Units {
		if u.Dos < len(dosrules.FullTables.Types) {
			continue
		}
		seen[u.Dos]++
		if u.Dos >= len(x.Types) {
			t.Errorf("%s: no extension type %d", id, u.Dos)
			continue
		}
		r, m := x.Types[u.Dos], x.Move.Types[u.Dos]
		if u.Armour != r.Armour || u.Group != r.Group || u.Cost == nil || *u.Cost != r.Cost || u.Speed != m.Speed {
			t.Errorf("%s: armour %d group %d cost %v speed %d, type %d has %+v, speed %d",
				id, u.Armour, u.Group, u.Cost, u.Speed, u.Dos, r, m.Speed)
		}
		var ground, water, reach int
		for _, s := range u.Weapons {
			if s == nil {
				continue
			}
			if s.Attack.Ground {
				ground, reach = s.Attack.Power, s.Attack.RangeMax
			}
			if s.Attack.Water {
				water = s.Attack.Power
			}
		}
		if ground != r.GroundPower || water != r.WaterPower || reach+1 != r.SurfaceRange {
			t.Errorf("%s: ground %d water %d reach %d, type %d has %d %d %d",
				id, ground, water, reach, u.Dos, r.GroundPower, r.WaterPower, r.SurfaceRange-1)
		}
	}
	for dos := len(dosrules.FullTables.Types); dos < len(x.Types); dos++ {
		if seen[dos] != 2 {
			t.Errorf("extension type %d has %d remake ids, want one per side", dos, seen[dos])
		}
	}
}
