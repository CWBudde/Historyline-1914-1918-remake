package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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

// TestRosterCoversUnitDat checks that every unit type of the unit table has a
// remake id for each side that fields it, except the rear records of the
// two-cell ships, and pins the side, country and movement type of each id.
func TestRosterCoversUnitDat(t *testing.T) {
	s := loadSet(t)
	type want struct {
		dos      int
		country  string
		movement int
	}
	roster := map[string]want{
		"CP_armCar": {0, "Germany", MoveNormal}, "EC_armCar": {0, "France", MoveNormal},
		"CP_a7v":     {1, "Germany", MoveNormal},
		"CP_gothaGV": {2, "Germany", MoveAir}, "CP_staaken": {3, "Germany", MoveAir},
		"CP_junkersJI": {4, "Germany", MoveAir}, "CP_fokkerEI": {5, "Germany", MoveAir},
		"CP_fokkerEIII": {6, "Germany", MoveAir}, "CP_albatrosDIII": {7, "Germany", MoveAir},
		"CP_fokkerDrI": {8, "Germany", MoveAir}, "CP_fokkerDVII": {9, "Germany", MoveAir},
		"CP_lArtillery": {10, "Germany", MoveNormal}, "EC_lArtillery": {10, "France", MoveNormal},
		"CP_mArtillery": {11, "Germany", MoveNormal}, "EC_mArtillery": {11, "France", MoveNormal},
		"CP_hArtillery": {12, "Germany", MoveNormal}, "EC_hArtillery": {12, "France", MoveNormal},
		"CP_bunker": {13, "Germany", MoveNormal}, "EC_bunker": {13, "France", MoveNormal},
		"CP_mobileAA": {14, "Germany", MoveNormal}, "EC_mobileAA": {14, "France", MoveNormal},
		"CP_staticAA": {15, "Germany", MoveNormal}, "EC_staticAA": {15, "France", MoveNormal},
		"CP_construction": {16, "Germany", MoveNormal}, "EC_construction": {16, "France", MoveNormal},
		"CP_eliteInf": {17, "Germany", MoveNormal}, "EC_eliteInf": {17, "France", MoveNormal},
		"CP_infantry": {18, "Germany", MoveNormal}, "EC_infantry": {18, "France", MoveNormal},
		"CP_antiTank": {19, "Germany", MoveNormal}, "EC_antiTank": {19, "France", MoveNormal},
		"CP_sappers": {20, "Germany", MoveNormal}, "EC_sappers": {20, "France", MoveNormal},
		"CP_cavalry": {21, "Germany", MoveNormal}, "EC_cavalry": {21, "France", MoveNormal},
		"CP_supplyCar": {22, "Germany", MoveNormal}, "EC_supplyCar": {22, "France", MoveNormal},
		"CP_balloon": {23, "Germany", MoveNormal}, "EC_balloon": {23, "France", MoveNormal},
		"CP_trainArt": {24, "Germany", MoveTracks}, "EC_trainArt": {24, "France", MoveTracks},
		"CP_armTrain": {25, "Germany", MoveTracks}, "EC_armTrain": {25, "France", MoveTracks},
		"CP_transTrain": {26, "Germany", MoveTracks}, "EC_transTrain": {26, "France", MoveTracks},
		"CP_patrolBoat": {27, "Germany", MoveWater}, "EC_patrolBoat": {27, "France", MoveWater},
		"CP_torpedoBoat": {28, "Germany", MoveWater}, "EC_torpedoBoat": {28, "France", MoveWater},
		"CP_submarine": {29, "Germany", MoveWater}, "EC_submarine": {29, "France", MoveWater},
		"CP_transShip": {31, "Germany", MoveWater}, "EC_transShip": {31, "France", MoveWater},
		"CP_destroyer": {32, "Germany", MoveWater}, "EC_destroyer": {32, "France", MoveWater},
		"CP_battleship": {34, "Germany", MoveWater}, "EC_battleship": {34, "France", MoveWater},
		"EC_voisin": {36, "France", MoveAir}, "EC_handleyPage": {37, "Britain", MoveAir},
		"EC_dh4": {38, "Britain", MoveAir}, "EC_moraneN": {39, "France", MoveAir},
		"EC_dh2": {40, "Britain", MoveAir}, "EC_nieuport17": {41, "France", MoveAir},
		"EC_spad7": {42, "France", MoveAir}, "EC_se5a": {43, "Britain", MoveAir},
		"EC_spad13": {44, "France", MoveAir}, "EC_camel": {45, "Britain", MoveAir},
		"EC_charron": {46, "France", MoveNormal},
		"EC_ft17":    {47, "France", MoveNormal}, "CP_ft17": {47, "Germany", MoveNormal},
		"EC_markI": {48, "Britain", MoveNormal}, "CP_markI": {48, "Germany", MoveNormal},
		"EC_stChamond": {49, "France", MoveNormal},
		"EC_markIV":    {50, "Britain", MoveNormal}, "CP_markIV": {50, "Germany", MoveNormal},
	}
	if len(s.Units) != len(roster) {
		t.Errorf("%d unit types, want %d", len(s.Units), len(roster))
	}
	covered := map[int]bool{}
	for id, w := range roster {
		u, ok := s.Units[id]
		if !ok {
			t.Errorf("unit %s missing", id)
			continue
		}
		covered[u.Dos] = true
		if u.Dos != w.dos || u.Country != w.country || u.MovementType != w.movement {
			t.Errorf("%s: record %d country %s movement %d, want %d %s %d",
				id, u.Dos, u.Country, u.MovementType, w.dos, w.country, w.movement)
		}
		if u.Exp != 0 || u.Fuel <= 0 || u.Sight <= 0 {
			t.Errorf("%s: exp %d fuel %d sight %d", id, u.Exp, u.Fuel, u.Sight)
		}
		if u.MovementType == MoveAir && slices.Contains(u.Skills, 1) {
			t.Errorf("%s: an aircraft may not capture", id)
		}
	}
	for dos := range 51 {
		rear := dos == 30 || dos == 33 || dos == 35 // second cell of submarine, destroyer, battleship
		if covered[dos] == rear {
			t.Errorf("record %d: covered %v, rear cell of a two-cell ship %v", dos, covered[dos], rear)
		}
	}
}
