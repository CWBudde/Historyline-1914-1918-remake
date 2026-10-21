package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// record is one unit type of testdata/unitdat.json, the original unit table
// (numbers only). Ranges are stored encoded, one more than shown.
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

// strike is an attack against one target class: power and maximum range.
type strike struct{ power, rangeMax int }

// TestUnitsFollowUnitDat checks every unit type against its record in the unit table:
// the stats, and that the attack slots reproduce the per-class powers and
// ranges with nothing extra. Bit 0x40 of the target flags (medium, heavy and
// train artillery) is a minimum range of 2.
func TestUnitsFollowUnitDat(t *testing.T) {
	recs, s := loadUnitDat(t), loadSet(t)
	if len(recs) != 51 {
		t.Fatalf("fixture has %d records, want %d", len(recs), 51)
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
