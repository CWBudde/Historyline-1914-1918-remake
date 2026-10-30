package data

import "strings"

// Hex is the result of applying infrastructure to a terrain.
type Hex struct {
	Movement, Battle, Sight *int
	MovementAllowed         []int // movement types that may enter, in Python order
	// LightBridge is a bridge only light units may cross (see LightBridge);
	// it restricts MoveNormal units, water and air movement is unchanged.
	LightBridge bool
}

// LightBridge reports whether infra crosses water on a light bridge: a
// bridge (bridge* or rBridge*) over a stream or lake whose deck is a path
// or the plain bridge, not a road or a railway. Such a bridge carries foot
// units and light towed guns (Unit.CrossesLightBridge), not vehicles or
// heavier guns; a road or rail bridge carries everything its deck does.
func LightBridge(infra []string) bool {
	var water, bridge, deck, full bool
	for _, id := range infra {
		switch {
		case strings.HasPrefix(id, "stream"), strings.HasPrefix(id, "lake"):
			water = true
		case strings.HasPrefix(id, "road"), strings.HasPrefix(id, "rail"):
			full = true
		case strings.HasPrefix(id, "path"), strings.HasPrefix(id, "bridge"):
			deck = true
		}
		if strings.HasPrefix(id, "rBridge") || strings.HasPrefix(id, "bridge") {
			bridge = true
		}
	}
	return water && bridge && deck && !full
}

// Modifiers ports the infrastructure loop of HexSquare.__init__ in hlrData.py.
// Infrastructure ids are matched by prefix and applied in order, so a later id
// overrides an earlier one. Ids no rule matches (such as rBridge*) change
// nothing.
func Modifiers(t Terrain, infra []string) Hex {
	return modifiers(t, infra, false)
}

// FixedModifiers is Modifiers with road bridges (rBridge*) treated like roads,
// so a bridge keeps its hex passable wherever it comes in the infrastructure list.
func FixedModifiers(t Terrain, infra []string) Hex {
	return modifiers(t, infra, true)
}

func modifiers(t Terrain, infra []string, roadBridges bool) Hex {
	h := Hex{
		Movement:        clone(t.Movement),
		Battle:          clone(t.Battle),
		Sight:           clone(t.Sight),
		MovementAllowed: []int{MoveNormal},
		LightBridge:     LightBridge(infra),
	}
	for _, i := range infra {
		switch {
		case strings.HasPrefix(i, "road"), roadBridges && strings.HasPrefix(i, "rBridge"):
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
