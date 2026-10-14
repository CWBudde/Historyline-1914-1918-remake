package data

import "strings"

// Hex is the result of applying infrastructure to a terrain.
type Hex struct {
	Movement, Battle, Sight *int
	MovementAllowed         []int // movement types that may enter, in Python order
}

// Modifiers ports the infrastructure loop of HexSquare.__init__ in hlrData.py.
// Infrastructure ids are matched by prefix and applied in order, so a later id
// overrides an earlier one. Ids no rule matches (such as rBridge*) change
// nothing.
func Modifiers(t Terrain, infra []string) Hex {
	h := Hex{
		Movement:        clone(t.Movement),
		Battle:          clone(t.Battle),
		Sight:           clone(t.Sight),
		MovementAllowed: []int{MoveNormal},
	}
	for _, i := range infra {
		switch {
		case strings.HasPrefix(i, "road"):
			h.Movement, h.Battle = ptr(0), ptr(0)
		case strings.HasPrefix(i, "path"), strings.HasPrefix(i, "bridge"):
			h.Movement, h.Battle = ptr(1), ptr(0)
		case strings.HasPrefix(i, "rail"), strings.HasPrefix(i, "cross"):
			h.Movement = ptr(0)
			h.MovementAllowed = append(h.MovementAllowed, MoveTracks)
		case strings.HasPrefix(i, "barbed"):
			h.Movement = ptr(10)
		case strings.HasPrefix(i, "trench"):
			h.Movement, h.Battle = ptr(10), ptr(10)
		case strings.HasPrefix(i, "stream"), strings.HasPrefix(i, "lake"):
			h.Movement = nil
			h.MovementAllowed = append(h.MovementAllowed, MoveWater)
			// Python's list.remove raises when 0 is already gone (two water
			// ids on one hex); no level does that, so it is simply skipped here.
			for k, m := range h.MovementAllowed {
				if m == MoveNormal {
					h.MovementAllowed = append(h.MovementAllowed[:k], h.MovementAllowed[k+1:]...)
					break
				}
			}
		}
	}
	return h
}

func ptr(v int) *int { return &v }

func clone(p *int) *int {
	if p == nil {
		return nil
	}
	return ptr(*p)
}
