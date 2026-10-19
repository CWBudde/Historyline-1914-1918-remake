package dosoracle

import "fmt"

// The demo has no sight: no routine reads a sight range from the type or
// terrain records, and nothing hides a unit because it is out of view.
// Each side has its own map and its own position word in every live record
// (+0x0b + 2*side). During a phase a move changes only the moving side's
// (10cc:071e), and the order handler 02dc:0004 resolves the attacking
// side's orders on the attacking side's map. At its end the phase change
// brings both views into line; ChangePhase transcribes it.

const (
	liveUnits = 0xF1 // the phase change walks live records 0-0xf0
	playerLen = 0x31 // the per-side records (Handler.Players)
	animStart = 0x0101
)

// Sighted is a live unit as the phase change reads it.
type Sighted struct {
	Flags4 int    `json:"flags4"`
	Size   int    `json:"size"`   // +2: 0 marks the unit destroyed
	Pos    [2]int `json:"pos"`    // +0x0b + 2*side: 2*cell+1 as side 0 and side 1 see it
	Facing [2]int `json:"facing"` // +0x0f + side: the direction each side sees it face
	Hold   int    `json:"hold,omitempty"`
}

func (u Sighted) bytes() []byte {
	r := make([]byte, liveLen)
	r[2] = byte(u.Size)
	putWord(r, 4, u.Flags4)
	r[0x0A] = byte(u.Hold)
	for k := range 2 {
		putWord(r, 0x0B+2*k, u.Pos[k])
		r[0x0F+k] = byte(u.Facing[k])
	}
	return r
}

// PhaseChange is what the phase change reads: both sides' maps (two bytes
// per cell, the terrain record and the occupant, rows of Width cells), the
// live units (every other record of 0-0xf0 carries 0x8000, no unit), the
// container records (EmptyHold where none is given), the side whose attack
// orders ran, the phase bytes (+0x16 of the per-side records; side 0's is
// Handler.Turn) and the turn counter.
type PhaseChange struct {
	Width, Height int
	Maps          [2][]int
	Units         map[int]Sighted
	Holds         map[int]Hold
	Attacker      int
	Phase         [2]int
	Count         int
}

// PhaseRun is the state the phase change leaves.
type PhaseRun struct {
	Maps     [2][]byte      // side 0's and side 1's map
	Live     map[int][]byte // the records of PhaseChange.Units
	Holds    map[int][]byte // all container records
	Players  [2][]byte      // the per-side records
	Orders   [2]int         // the count bytes of the two attack order lists
	Count    int            // the turn counter
	Anim     int            // the animation flags (animStart before)
	Messages [][2]int       // status messages shown: code, side
}

// live are the records 0-0xf0 the phase change walks.
func (g *PhaseChange) live() map[int][]byte {
	out := make(map[int][]byte, liveUnits)
	for idx := range liveUnits {
		u, ok := g.Units[idx]
		if !ok {
			u = Sighted{Flags4: 0x8000}
		}
		out[idx] = u.bytes()
	}
	return out
}

// playerBytes is a side's per-side record before the phase change: filler
// that tells the bytes apart, and the phase byte.
func playerBytes(side, phase int) []byte {
	b := make([]byte, playerLen)
	for i := range b {
		b[i] = byte(0x40 + playerLen*side + i)
	}
	b[0x16] = byte(phase)
	return b
}

// ChangePhase is the phase change in the tail of the order handler,
// 02dc:1a09-1cc6. A is the side whose orders ran, B the other.
//   - Container sync (02dc:23e5 on the transporter records and the three
//     building tables): records with +0x19 & 0x8002 are skipped. Of a record
//     B owns (0ca0:020c with the side passed as the flags: B's flags & 1, or
//     for side 0 none), B's slots are copied to A's, a slot holding a unit
//     already marked 0x8000 cleared first; of any other, A's slots to B's.
//   - Every occupant byte of A's map becomes 0xff.
//   - Each unit of B (flags & 1, or for side 0 flags & 3 == 0) without 0x8000
//     gets B's position word and facing in A's.
//   - Units with +2 == 0 are marked 0x8000 (02dc:2b1d).
//   - Every unit without 0x8000 or 0x4000 (in a container) is written into
//     A's map at its +0x0b, side 0's position word, whichever side A is.
//   - A's whole map, terrain too, is copied over B's (0e40:1d9d).
//   - The per-side records get +0x17..+0x19 = 0 and +0x1b = 5, their phase
//     bytes swap between 1 and 2 (side 0 becomes 2 if it was 1, else 1), both
//     attack order counts become 0, the animation flags gain 0x18, messages
//     0x35/0x36 go to side 0 and side 1 (0x35 to the side whose phase becomes
//     1), and the turn counter counts up when side 0's becomes 1.
//
// Undefined: a slot of B's other than a unit or 0xff in a record B owns, and
// a position word +0x0b beyond the map.
func ChangePhase(g *PhaseChange) (PhaseRun, error) {
	a, b := g.Attacker&1, 1-g.Attacker&1
	size := 2 * g.Width * g.Height
	r := PhaseRun{Live: map[int][]byte{}, Holds: holdsOf(g.Holds), Anim: animStart | 0x18}
	for k := range 2 {
		if len(g.Maps[k]) != size {
			return r, fmt.Errorf("map %d holds %d bytes, not %d", k, len(g.Maps[k]), size)
		}
		r.Maps[k] = make([]byte, size)
		for i, v := range g.Maps[k] {
			r.Maps[k][i] = byte(v)
		}
	}
	live := g.live()

	for _, t := range [][2]int{{0, 0x46}, {holdFlag100, holdTable}, {holdFlag400, holdTable}, {holdFlag40, 2}} {
		for i := t[0]; i < t[0]+t[1]; i++ {
			h := r.Holds[i]
			f := word(h, 0x19)
			if f&0x8002 != 0 {
				continue
			}
			for k := range 7 {
				if !own(b, f) {
					h[7*b+k] = h[7*a+k]
					continue
				}
				u := h[7*b+k]
				switch {
				case u <= slotFree && word(live[int(u)], 4)&0x8000 != 0:
					h[7*b+k], u = 0xFF, 0xFF
				case u > slotFree && u != 0xFF:
					return r, ErrUndefined
				}
				h[7*a+k] = u
			}
		}
	}

	am := r.Maps[a]
	for i := 1; i < size; i += 2 {
		am[i] = 0xFF
	}
	for idx := range liveUnits {
		l := live[idx]
		f := word(l, 4)
		if f&0x8000 != 0 || b == 1 && f&1 == 0 || b == 0 && f&3 != 0 {
			continue
		}
		putWord(l, 0x0B+2*a, word(l, 0x0B+2*b))
		l[0x0F+a] = l[0x0F+b]
	}
	for idx := range liveUnits {
		if l := live[idx]; word(l, 4)&0x8000 == 0 && l[2] == 0 {
			putWord(l, 4, word(l, 4)|0x8000)
		}
	}
	for idx := range liveUnits {
		l := live[idx]
		if word(l, 4)&0xC000 != 0 {
			continue
		}
		p := word(l, 0x0B)
		if p >= size {
			return r, ErrUndefined
		}
		am[p] = byte(idx)
	}
	copy(r.Maps[b], am)

	for k := range 2 {
		r.Players[k] = playerBytes(k, g.Phase[k])
		p := r.Players[k]
		p[0x17], p[0x18], p[0x19], p[0x1B] = 0, 0, 0, 5
	}
	codes := [2]int{0x36, 0x35}
	r.Count = g.Count
	if r.Players[0][0x16] == 1 {
		r.Players[0][0x16], r.Players[1][0x16] = 2, 1
	} else {
		r.Players[0][0x16], r.Players[1][0x16] = 1, 2
		codes = [2]int{0x35, 0x36}
		r.Count = (g.Count + 1) & 0xFFFF
	}
	r.Messages = [][2]int{{codes[0], 0}, {codes[1], 1}}
	for idx := range g.Units {
		r.Live[idx] = live[idx]
	}
	return r, nil
}
