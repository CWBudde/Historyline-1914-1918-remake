package dosrules

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/tpwm"
)

// The shipped types use the primary strength. Alternate strength belongs to
// the casualty routine's synthetic cases, not to a missing unit mechanic.
func TestShippedStrengthData(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			for i, typ := range p.Tables.Types {
				if typ.StrengthFlags&4 != 0 {
					t.Errorf("type %d enables alternate strength", i)
				}
			}
		})
	}
}

// Check the embedded strength data against the local original when present.
// Only numeric fields enter the assertions; original files stay local.
func TestStrengthDataMatchesOriginal(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			dir := "../../original/extracted/dos-full"
			if build == "demo" {
				dir = "../../original/demo/HL"
			}
			b, err := os.ReadFile(filepath.Join(dir, "UNIT.DAT"))
			if errors.Is(err, os.ErrNotExist) {
				t.Skip("no local original unit table")
			}
			if err != nil {
				t.Fatal(err)
			}
			b, err = tpwm.Unpack(b, binary.LittleEndian)
			if err != nil {
				t.Fatal(err)
			}
			const recordLen = 69
			if len(b) != recordLen*len(p.Tables.Types) {
				t.Fatalf("unit table has %d bytes, want %d", len(b), recordLen*len(p.Tables.Types))
			}
			for i, typ := range p.Tables.Types {
				r := b[recordLen*i : recordLen*(i+1)]
				if group, flags := int(r[3]), int(binary.LittleEndian.Uint16(r[15:17])); typ.Group != group || typ.StrengthFlags != flags {
					t.Errorf("type %d: group/flags %d/%d, original %d/%d", i, typ.Group, typ.StrengthFlags, group, flags)
				}
				if r[4] != 0 {
					t.Errorf("type %d: alternate group %d, want 0", i, r[4])
				}
			}
		})
	}
}

// These gates inspect recorded expectations, independently of the resolver,
// so the corpus must keep exercising alternate strength on both sides.
func TestAlternateStrengthEvidence(t *testing.T) {
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			seen := map[string]int{
				"alternate attacker": 0, "alternate defender": 0, "both alternate": 0,
				"attacker alternate loss": 0, "defender alternate loss": 0,
				"attacker primary cleared": 0, "defender primary cleared": 0,
				"attacker primary preserved": 0, "defender primary preserved": 0,
			}
			var sizes [2][7]bool
			for _, f := range load[CasualtyFixture](t, build, "casualties.json") {
				if f.Expected.Error != "" {
					continue
				}
				a, d, _, err := f.Inputs(p.Tables)
				if err != nil {
					t.Fatal(err)
				}
				if a.Type.StrengthFlags&4 != 0 && d.Type.StrengthFlags&4 != 0 {
					seen["both alternate"]++
				}
				for side, c := range []struct {
					name      string
					unit      Fighter
					size, alt int
				}{
					{"attacker", a, f.Expected.AttackerSizeAfter, f.Expected.AttackerAlternateAfter},
					{"defender", d, f.Expected.DefenderSizeAfter, f.Expected.DefenderAlternateAfter},
				} {
					if c.unit.Type.StrengthFlags&4 == 0 {
						if c.alt != c.unit.Alt {
							t.Errorf("%s: ordinary %s changed alternate strength", f.ID, c.name)
						}
						continue
					}
					seen["alternate "+c.name]++
					if c.unit.Alt >= 0 && c.unit.Alt <= 6 {
						sizes[side][c.unit.Alt] = true
					}
					if c.alt < c.unit.Alt {
						seen[c.name+" alternate loss"]++
					}
					if c.alt > 0 {
						if c.size != c.unit.Size {
							t.Errorf("%s: surviving %s changed primary strength", f.ID, c.name)
						}
						seen[c.name+" primary preserved"]++
					} else if c.size == 0 && c.unit.Size > 0 {
						seen[c.name+" primary cleared"]++
					}
				}
			}
			for side, values := range sizes {
				for size, present := range values {
					if !present {
						t.Errorf("no alternate strength %d on side %d", size, side)
					}
				}
			}
			require(t, seen)
		})
	}
}
