package dosrules

import "testing"

// Return fire: a defender next to its attacker fires back only when its type
// can; medium, heavy and train artillery (long range), the construction unit
// and the static balloon (never return fire) take the losses without a reply.
func TestEngageReturned(t *testing.T) {
	const cMedium, cHeavy, cConstruction, cBalloon = 11, 12, 16, 23
	for build, p := range profiles {
		t.Run(build, func(t *testing.T) {
			n := neighbours(40, 9, 9)
			for _, c := range []struct {
				defender int
				returned bool
			}{
				{cInfantry, true}, {cPanzer, true}, {cMedium, false}, {cHeavy, false},
				{cConstruction, false}, {cBalloon, false},
			} {
				td := p.Tables.Types[c.defender]
				b := &Battlefield{Width: 9, Height: 9, Units: map[int]Unit{
					0: {Type: cInfantry, Cell: 40, Size: 6, Flags4: 1 | p.Tables.Types[cInfantry].TargetFlags&classBits},
					1: {Type: c.defender, Cell: n[0], Size: td.Group, Flags4: 2 | td.TargetFlags&classBits},
				}, Terrain: map[int]int{}}
				o, err := p.Engage(b, 0, n[0])
				if err != nil {
					t.Fatal(err)
				}
				if !o.ReturnFire || o.Returned != c.returned {
					t.Errorf("infantry on an adjacent type %d: next to it %v, fired back %v, want %v",
						c.defender, o.ReturnFire, o.Returned, c.returned)
				}
				if !c.returned && o.AttackerSize != 6 {
					t.Errorf("type %d did not fire back, yet the attacker is down to %d", c.defender, o.AttackerSize)
				}
			}
		})
	}
}
