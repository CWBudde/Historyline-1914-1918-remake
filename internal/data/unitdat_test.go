package data

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosexe"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

var update = flag.Bool("update", false, "rewrite testdata/unitdat.json from the DOS demo's UNIT.DAT")

// unitDatSource is UNIT.DAT of the freely available DOS demo (hl-dem.zip, see
// docs/research/original-dos-rules.md), unpacked into the ignored original/.
var unitDatSource = filepath.Join("..", "..", "original", "demo", "HL", "UNIT.DAT")

// record is the numeric part of a UNIT.DAT record; names and raw bytes stay
// out of the repository. Ranges are stored encoded, as the original shows one
// less.
type record struct {
	Index        int `json:"index"`
	Budget       int `json:"budget"`
	Speed        int `json:"speed"`
	Armour       int `json:"armour"`
	Group        int `json:"group"`
	MovementMask int `json:"movementMask"`
	TargetFlags  int `json:"targetFlags"`
	AirRange     int `json:"airRange"`
	SurfaceRange int `json:"surfaceRange"`
	AirPower     int `json:"airPower"`
	GroundPower  int `json:"groundPower"`
	WaterPower   int `json:"waterPower"`
	Cost         int `json:"cost"`
	Capacity     int `json:"capacity"`
	Weight       int `json:"weight"`
}

func decodeUnitDat(t *testing.T, packed []byte) []record {
	t.Helper()
	raw, err := tpwm.Unpack(packed, binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	us, err := dosexe.Units(raw)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]record, len(us))
	for i, u := range us {
		out[i] = record{u.Index, u.MovementBudgetRaw, u.DisplayedSpeed, u.Armour, u.MaxGroup, u.MovementMaskRaw,
			u.TargetFlagsRaw, u.AirRangeEncoded, u.SurfaceRangeEncoded, u.AirPower, u.GroundPower, u.WaterPower,
			u.CostRaw, u.CapacityRaw, u.WeightRaw}
	}
	return out
}

func loadUnitDat(t *testing.T) []record {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "unitdat.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rs []record
	if err := json.Unmarshal(b, &rs); err != nil {
		t.Fatal(err)
	}
	return rs
}

// TestUnitDatFixtureMatchesDemo checks testdata/unitdat.json against the demo's
// UNIT.DAT when it is present, and rewrites it with -update.
func TestUnitDatFixtureMatchesDemo(t *testing.T) {
	packed, err := os.ReadFile(unitDatSource)
	if os.IsNotExist(err) {
		t.Skipf("%s not present", unitDatSource)
	}
	if err != nil {
		t.Fatal(err)
	}
	got := decodeUnitDat(t, packed)
	if *update {
		b, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join("testdata", "unitdat.json"), append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if want := loadUnitDat(t); !reflect.DeepEqual(got, want) {
		t.Error("testdata/unitdat.json differs from UNIT.DAT; rerun with -update")
	}
}

// strike is an attack against one target class: power and maximum range.
type strike struct{ power, rangeMax int }

// TestUnitsFollowUnitDat checks every unit type against its UNIT.DAT record:
// the stats, and that the attack slots reproduce the per-class powers and
// ranges with nothing extra. Bit 0x40 of the target flags (medium, heavy and
// train artillery) is a minimum range of 2.
func TestUnitsFollowUnitDat(t *testing.T) {
	recs, s := loadUnitDat(t), loadSet(t)
	if len(recs) != dosexe.UnitCount {
		t.Fatalf("fixture has %d records, want %d", len(recs), dosexe.UnitCount)
	}
	for id, u := range s.Units {
		r := recs[u.Dos]
		var cost *int
		if r.Cost != 255 {
			cost = &r.Cost
		}
		if u.Armour != r.Armour || u.Speed != r.Speed || u.Weight != r.Weight || u.StorageMax != r.Capacity ||
			u.Group != r.Group || !eqInt(u.Cost, cost) {
			t.Errorf("%s: armour %d speed %d weight %d storage %d group %d cost %v, record %d has %+v",
				id, u.Armour, u.Speed, u.Weight, u.StorageMax, u.Group, show(u.Cost), u.Dos, r)
		}

		want := map[string]strike{}
		for class, st := range map[string]strike{
			"air": {r.AirPower, r.AirRange - 1}, "ground": {r.GroundPower, r.SurfaceRange - 1},
			"water": {r.WaterPower, r.SurfaceRange - 1},
		} {
			if st.power > 0 && st.rangeMax > 0 {
				want[class] = st
			}
		}
		rangeMin := 1
		if r.TargetFlags&0x40 != 0 {
			rangeMin = 2
		}
		got := map[string]strike{}
		for i, sl := range u.Weapons {
			if sl == nil || sl.Attack == nil {
				continue
			}
			a := sl.Attack
			if a.RangeMin != rangeMin {
				t.Errorf("%s slot %d: minimum range %d, want %d", id, i, a.RangeMin, rangeMin)
			}
			for class, on := range map[string]bool{"air": a.Air, "ground": a.Ground, "water": a.Water} {
				if !on {
					continue
				}
				if _, dup := got[class]; dup {
					t.Errorf("%s slot %d: second attack against %s", id, i, class)
				}
				got[class] = strike{a.Power, a.RangeMax}
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: attacks %v, record %d has %v", id, got, u.Dos, want)
		}
	}
}
