package dosplay

import (
	"slices"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosrules"
)

// A side that owns a factory at the start can build in it within a few of
// its action phases of a real game, without energy set by hand: energy
// starts at 0, the buildings' credits pay for the cheapest type the factory
// builds, and builds are given in the action phase only (the manual's action
// cycle). On Civil each side owns one factory; Pulse has none and Mouse's are
// all neutral.
func TestFactoryProductionInPlay(t *testing.T) {
	const resolutions = 4
	for _, lv := range shipped(t) {
		s := session(t, lv, Options{})
		var owns, built [2]bool
		for n := 0; n < resolutions && s.Ended() == nil; n++ {
			for side := range 2 {
				for _, b := range s.View(side).Buildings {
					opts := s.Buildable(side, b.Hold)
					if b.Kind != Factory || int(b.Owner) != side+1 || len(opts) == 0 {
						continue
					}
					owns[side] = true
					for _, o := range opts {
						err := s.CanBuild(side, b.Hold, o.Type)
						built[side] = built[side] || err == nil
						if err == nil && (n == 0 || s.Phase(side) != dosrules.PhaseAction) {
							t.Errorf("%s side %d builds %s after %d resolutions in phase %d with energy %d",
								lv.MapName, side, o.Name, n, s.Phase(side), s.Energy(side))
						}
					}
				}
			}
			s.Confirm(0)
			s.Confirm(1)
			if _, err := s.Resolve(); err != nil {
				t.Fatal(err)
			}
		}
		for side := range 2 {
			if owns[side] != (lv.MapName == "Civil") || owns[side] && !built[side] {
				t.Errorf("%s side %d: owns a factory %v, builds within %d resolutions %v (energy %d)",
					lv.MapName, side, owns[side], resolutions, built[side], s.Energy(side))
			}
		}
	}
}

// TestFactoryBuildsLeave builds each type a shipped factory builds and takes
// it out in the side's next movement phase (a built supply car once had no
// container record of its own and no hex to leave for).
func TestFactoryBuildsLeave(t *testing.T) {
	for _, lv := range shipped(t) {
		for side := range 2 {
			for _, b := range session(t, lv, Options{}).View(side).Buildings {
				if b.Kind != Factory || int(b.Owner) != side+1 {
					continue
				}
				for _, o := range session(t, lv, Options{}).Buildable(side, b.Hold) {
					s := session(t, lv, Options{})
					resolve := func() {
						s.Confirm(0)
						s.Confirm(1)
						if _, err := s.Resolve(); err != nil {
							t.Fatal(err)
						}
					}
					for s.Phase(side) != dosrules.PhaseAction {
						resolve()
					}
					s.Game.Energy[side] = o.Cost
					evs, err := s.Build(side, b.Hold, o.Type)
					if err != nil || len(evs) != 1 || evs[0].Kind != Built {
						t.Fatalf("%s side %d builds %s: %v %+v", lv.MapName, side, o.Name, err, evs)
					}
					unit := evs[0].Unit
					for resolve(); s.Phase(side) != dosrules.PhaseMovement; resolve() {
					}
					h := s.Game.Holds[b.Hold]
					slot := slices.Index(h.Slots[7*side:7*side+7], unit)
					to, err := s.ExitDestinations(side, b.Hold, slot)
					if err != nil || len(to) == 0 {
						t.Errorf("%s side %d: the %s built has no hex to leave for (%v)", lv.MapName, side, o.Name, err)
						continue
					}
					at := Coords(to)[0]
					if _, err := s.Exit(side, b.Hold, slot, at); err != nil || s.View(side).Units[at].Index != unit {
						t.Errorf("%s side %d: the %s built does not leave for %v: %v", lv.MapName, side, o.Name, at, err)
					}
					checkBound(t, s)
				}
			}
		}
	}
}
