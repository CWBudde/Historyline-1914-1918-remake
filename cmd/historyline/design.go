package main

import "strings"

// The map designs F8 cycles through: the level's painted backdrop where it
// has one, the remake's terrain tiles, and — under -rules dos, where the
// original art can be loaded — the original (DOS) tiles and units. A design
// that is unavailable on this level or build is skipped, and the message
// says so; the backdrop and the tiles always remain reachable.

type mapDesign int

const (
	designBackdrop mapDesign = iota
	designTiles
	designOrig
)

// designOrder is the cycle's order.
var designOrder = []mapDesign{designBackdrop, designTiles, designOrig}

// mapDesign is the design the map is drawn in now.
func (g *game) mapDesign() mapDesign {
	if g.origOn() {
		return designOrig
	}
	if g.backdrop != nil {
		return designBackdrop
	}
	return designTiles
}

// hasBackdrop reports whether this level has a painted backdrop, shown or
// hidden behind another design.
func (g *game) hasBackdrop() bool { return g.backdrop != nil || g.backdropHidden != nil }

// cycleDesign advances F8 to the next available map design and tells the
// player, naming any design it had to skip.
func (g *game) cycleDesign() string {
	cur := g.mapDesign()
	start := 0
	for i, d := range designOrder {
		if d == cur {
			start = i
		}
	}
	var skipped []string
	for step := 1; step <= len(designOrder); step++ {
		next := designOrder[(start+step)%len(designOrder)]
		msg, reason, ok := g.showDesign(next)
		if !ok {
			skipped = append(skipped, reason)
			continue
		}
		if len(skipped) > 0 {
			msg += " (" + strings.Join(skipped, "; ") + ")"
		}
		return msg
	}
	return "map design: terrain tiles" // unreachable: the tiles always show
}

// showDesign switches the map to design d. It reports false, with the reason,
// when d is unavailable here, leaving the current design untouched in that
// case except where noted by the caller cycling on to the next one.
func (g *game) showDesign(d mapDesign) (msg, reason string, ok bool) {
	switch d {
	case designBackdrop:
		if !g.hasBackdrop() {
			return "", "no painted backdrop for this level", false
		}
		g.setOrig(false)
		if g.backdrop == nil {
			g.backdrop, g.backdropHidden = g.backdropHidden, nil
		}
		g.refreshDesign()
		return "map design: painted backdrop", "", true
	case designTiles:
		g.setOrig(false)
		g.hideBackdrop()
		g.refreshDesign()
		return "map design: terrain tiles", "", true
	case designOrig:
		if err := g.origAvailable(); err != "" {
			return "", "original (DOS) tiles unavailable: " + err, false
		}
		g.hideBackdrop()
		g.setOrig(true)
		g.refreshDesign()
		return "map design: original (DOS) tiles and units", "", true
	}
	return "", "unknown design", false
}

// hideBackdrop puts the painted backdrop away until the backdrop design is
// shown again.
func (g *game) hideBackdrop() {
	if g.backdrop != nil {
		g.backdropHidden, g.backdrop = g.backdrop, nil
	}
}

// setOrig turns the original art's drawing on or off, without loading it.
func (g *game) setOrig(on bool) {
	if g.orig != nil {
		g.orig.on = on
	}
}

// origAvailable loads the original art on first use and reports why it
// cannot be shown, or "" when it can.
func (g *game) origAvailable() string {
	if g.dos == nil {
		return "only under -rules dos"
	}
	if g.orig == nil {
		g.orig = &origArt{}
		if a, err := loadOrigArt(); err != nil {
			g.orig.err = err.Error()
		} else {
			g.orig.tiles, g.orig.units = a.tiles, a.units
		}
	}
	return g.orig.err
}

// refreshDesign re-lays the split screen out for the design now shown: a
// painted map fills the slim frame, the other designs use the plate frame.
func (g *game) refreshDesign() {
	if g.split != nil {
		g.layoutSplit()
	}
}
