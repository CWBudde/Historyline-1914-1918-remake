package dosai

import (
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
)

// The computer plays out maps 06 and 62 (two-cell ships on both sides): a
// ship is one unit throughout, its halves of one strength and never one
// without the other.
func TestShipsStayWhole(t *testing.T) {
	for _, i := range []int{6, 62} {
		s := original(t, i)
		hits, sunk := 0, 0
		check := func(evs []dosplay.Event) {
			for _, ev := range evs {
				if ev.Kind != dosplay.Battle {
					continue
				}
				for _, f := range []dosplay.Fighter{ev.Battle.Attacker, ev.Battle.Defender} {
					if s.Game.Units[f.Unit.Index].Flags4&0x40 != 0 {
						hits++
						if f.Destroyed {
							sunk++
						}
					}
				}
			}
			v := s.TrueView()
			for c, u := range v.Units {
				other := u.Rear
				if other == nil {
					other = u.Head
				}
				if other == nil {
					continue
				}
				if p, ok := v.Units[*other]; !ok || p.Size != u.Size || p.Exp != u.Exp {
					t.Fatalf("map %02d turn %d: ship half %d at %v (size %d) has its other half %+v at %v",
						i, s.Turn(), u.Index, c, u.Size, p, *other)
				}
			}
		}
		if _, err := playOut(s, [2]*Planner{{}, {}}, originalTurns, check); err != nil {
			t.Fatalf("map %02d: %v", i, err)
		}
		t.Logf("map %02d: %d ship fighters, %d sunk", i, hits, sunk)
	}
}
