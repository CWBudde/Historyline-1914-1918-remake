package dosrules

import (
	"maps"
	"slices"
)

// Capture: the action pass of the order handler (02dc:0e91-19a4) after the
// action side's attacks and before the phase change. It walks the live
// records in order (work, which also runs the sappers' work in sapper.go)
// and, for a unit whose arrival this cycle was onto another
// side's or a neutral building (action 4, 02dc:11a6), hands the building and
// every unit in it to the unit's side; the unit goes in. Only the movement
// side can have such an arrival, and only a type of flag +6 bit 1 (infantry
// and cavalry) reaches a foreign building's entrance. Taking an HQ wins the
// battle.
//
// Assumed, not verified: a captured unit's new serial number follows the
// side's highest one, as a built unit's does (the original counts per type
// and side).

// The kind bits of a building record's flag word, as the record creators
// set them (159e:040e HQ, 159e:024b factory, 159e:032d depot).
const (
	KindHQ      = 0x04
	KindFactory = 0x08
	KindDepot   = 0x10
	kindBits    = KindHQ | KindFactory | KindDepot
)

// BuildingKind is the kind bit of building record r: the HQ table
// (HoldFlag40), the depot table (HoldFlag100) or the factory table
// (HoldFlag400); 0 for a transporter's record.
func BuildingKind(r int) int {
	switch {
	case r >= HoldFlag400:
		return KindFactory
	case r >= HoldFlag100:
		return KindDepot
	case r >= HoldFlag40:
		return KindHQ
	}
	return 0
}

// markKinds gives every building record its kind bit.
func (g *Game) markKinds() {
	for r, h := range g.Holds {
		if k := BuildingKind(r); k != 0 && h.Flags&0x8000 == 0 {
			h.Flags = h.Flags&^kindBits | k
			g.Holds[r] = h
		}
	}
}

// capture is the capture of unit i in side a's action pass (work): its
// event, and whether it took an HQ; ok is false where its building record
// is missing.
func (g *Game) capture(a, i int) (Event, bool, bool) {
	b := 1 - a
	u := g.Units[i]
	// The record by the flag word the arrival copied (+0x17) and its
	// index (+0x15), 02dc:11a6-1208.
	r := HoldFlag100
	switch f := u.Action[4] | u.Action[5]<<8; {
	case f&KindHQ != 0:
		r = HoldFlag40
	case f&KindFactory != 0:
		r = HoldFlag400
	}
	r += u.Action[2]
	h, found := g.Holds[r]
	if !found {
		return Event{}, false, false
	}
	// B's slots take A's (137d), then the capturer's slot is B's first
	// free one; in a full building the unit in slot 0 is destroyed and
	// the capturer takes its slot (13b4).
	copy(h.Slots[7*b:7*b+7], h.Slots[7*a:7*a+7])
	k := slices.IndexFunc(h.Slots[7*b:7*b+7], func(s int) bool { return s > SlotFree })
	if k < 0 {
		k = 0
		g.Holds[r] = h
		g.remove(h.Slots[7*b])
		h = g.Holds[r]
	}
	u.Action[6] = k
	// Every unit at the building's first cell as B sees it goes to B:
	// the units held there and their cargo (1412).
	for _, j := range slices.Sorted(maps.Keys(g.Units)) {
		v := g.Units[j]
		if v.Flags4&0x8000 != 0 || v.Pos[b] != h.Cells[0] {
			continue
		}
		v.Flags4 = v.Flags4&^3 | b
		v.Serial = g.serial(b)
		g.Units[j] = v
		if v.Flags4&0x1000 != 0 {
			if c, ok := g.Holds[v.Hold]; ok {
				c.Flags = c.Flags&^3 | b
				g.Holds[v.Hold] = c
			}
		}
	}
	// The entrance shows B's barrier on A's map (15d6); the phase
	// change copies it to B's.
	if t, ok := g.ownedTerrain(g.Maps[a][h.Cells[0]], b); ok {
		g.Maps[a][h.Cells[0]] = t
	}
	hq := h.Flags&(KindFactory|KindDepot) == 0 && h.Flags&KindHQ != 0
	h.Flags = h.Flags&^3 | b
	h.Slots[7*b+k] = i
	g.Holds[r] = h
	// The capturer goes in (1749-1787): held, moved, at the building's
	// first cell as B sees it; a transporter's cargo follows.
	if u.Pos[b] < len(g.Maps[a]) {
		g.Maps[a][u.Pos[b]] = InBuilding
	}
	u.Pos[b] = h.Cells[0]
	u.Flags4 = u.Flags4&^0x100 | 0x4200
	g.Units[i] = u
	if u.Flags4&0x1000 != 0 {
		g.follow(i, b)
	}
	return Event{Kind: EvCaptured, Side: b, Unit: i, Cell: h.Cells[0] / 2, Hold: r}, hq, true
}

// ownedTerrain is the terrain record of building terrain tr that shows side
// as its owner: the same flags but the owner bits (4 side 0, 1 side 1, 2
// neutral).
func (g *Game) ownedTerrain(tr, side int) (int, bool) {
	t, err := g.Profile.moveTables()
	if err != nil || tr < 0 || tr >= len(t.Terrains) {
		return 0, false
	}
	f := t.Terrains[tr].Flags
	if f&0x540 == 0 {
		return 0, false
	}
	want := f&^7 | [2]int{4, 1}[side]
	for i, x := range t.Terrains {
		if x.Flags == want {
			return i, true
		}
	}
	return 0, false
}

// serial is the serial number of a unit side gains: one above its highest.
func (g *Game) serial(side int) int {
	s := -1
	for _, u := range g.Units {
		if u.Flags4&2 == 0 && Owns(u.Flags4, side) {
			s = max(s, u.Serial)
		}
	}
	return s + 1
}

// remove is the destroy helper 159e:0595 on unit i: its record holds no
// unit (+4 0x8000), nor do a ship's other half and a transporter's cargo,
// whose own record is cleared (+0x19 0x8000).
func (g *Game) remove(i int) {
	u, ok := g.Units[i]
	if !ok {
		return
	}
	if u.Flags4&0x40 != 0 {
		p := i - 1
		if u.Flags4&0x80 != 0 {
			p = i + 1
		}
		if v, ok := g.Units[p]; ok {
			v.Flags4 = 0x8000
			g.Units[p] = v
		}
	}
	if u.Flags4&0x1000 != 0 {
		if c, ok := g.Holds[u.Hold]; ok {
			for _, s := range c.Slots {
				if v, ok := g.Units[s]; ok && s <= SlotFree {
					v.Flags4 = 0x8000
					g.Units[s] = v
				}
			}
			c.Flags = 0x8000
			g.Holds[u.Hold] = c
		}
	}
	u.Flags4 = 0x8000
	g.Units[i] = u
}

// follow is 159e:0805 on the game: the units in side's slots of
// transporter i's record and the record's first cell take i's position as
// side sees it.
func (g *Game) follow(i, side int) {
	u := g.Units[i]
	c, ok := g.Holds[u.Hold]
	if !ok {
		return
	}
	for _, s := range c.Slots[7*side : 7*side+7] {
		if v, ok := g.Units[s]; ok && s <= SlotFree {
			v.Pos[side] = u.Pos[side]
			g.Units[s] = v
		}
	}
	c.Cells[0] = u.Pos[side]
	g.Holds[u.Hold] = c
}
