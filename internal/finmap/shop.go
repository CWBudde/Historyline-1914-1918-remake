package finmap

import (
	"errors"
	"fmt"
)

// Kind is the container table an SHP record fills.
type Kind int

// Record kinds.
const (
	HQ          Kind = 0 // containers 70 + index, one per side
	Factory     Kind = 1 // containers 82 + index (terrain flag 0x400)
	Depot       Kind = 2 // containers 72 + index (terrain flag 0x100)
	Transporter Kind = 3 // container index, a transporter's hold
)

// Container table layout: 70 transporter holds, then the two HQs, ten
// depots and ten factories, 0x1c bytes each in the game.
const (
	Transporters  = 70
	HQBase        = 70
	DepotBase     = 72
	FactoryBase   = 82
	Containers    = 92
	TableBuilding = 10 // depots and factories each
)

func (k Kind) String() string {
	switch k {
	case HQ:
		return "hq"
	case Factory:
		return "factory"
	case Depot:
		return "depot"
	case Transporter:
		return "transporter"
	}
	return fmt.Sprintf("kind%d", int(k))
}

// Record is one 12-byte SHP record.
type Record struct {
	Side   int  // byte 0: the side of the units the record creates (2 neutral)
	Kind   Kind // byte 1
	Index  int  // byte 2: the entry in the kind's table
	Energy int  // byte 3: energy the owner is credited with each command cycle
	// ComputerValue is byte 4: only the computer player reads it, as the
	// building's weight as a goal (ten times it is the goal's value).
	ComputerValue int
	// Slots are bytes 5-11: a unit type each, above 0xf0 none. Each unit
	// is created inside the container.
	Slots [7]byte
}

// SlotEmpty is the lowest slot value that creates no unit.
const SlotEmpty = 0xf1

// Units are the unit types the record creates, in slot order.
func (r Record) Units() []int {
	out := []int{}
	for _, s := range r.Slots {
		if s < SlotEmpty {
			out = append(out, int(s))
		}
	}
	return out
}

// Container is the container the record fills. An HQ index above 1 runs on
// into the depot table, as the game addresses it.
func (r Record) Container() int {
	switch r.Kind {
	case HQ:
		return HQBase + r.Index
	case Factory:
		return FactoryBase + r.Index
	case Depot:
		return DepotBase + r.Index
	}
	return r.Index
}

// Shop is a decoded SHP file: a byte per unit type (1: factories may not
// produce the type on this map, 0: they may), the record count, then the
// records.
type Shop struct {
	Locked  [TypeCount]bool
	Records []Record
}

// DecodeShop decodes an unpacked SHP file.
func DecodeShop(b []byte) (*Shop, error) {
	if len(b) < TypeCount+1 {
		return nil, errors.New("finmap: SHP truncated")
	}
	n := int(b[TypeCount])
	if len(b) != TypeCount+1+12*n {
		return nil, fmt.Errorf("finmap: SHP has %d bytes for %d records", len(b), n)
	}
	s := &Shop{Records: make([]Record, n)}
	for t := range TypeCount {
		if b[t] > 1 {
			return nil, fmt.Errorf("finmap: SHP type %d flag %d", t, b[t])
		}
		s.Locked[t] = b[t] == 1
	}
	for i := range n {
		p := b[TypeCount+1+12*i:]
		r := Record{Side: int(p[0]), Kind: Kind(p[1]), Index: int(p[2]), Energy: int(p[3]), ComputerValue: int(p[4])}
		copy(r.Slots[:], p[5:12])
		if r.Side > Neutral {
			return nil, fmt.Errorf("finmap: SHP record %d side %d", i, r.Side)
		}
		if err := r.checkIndex(); err != nil {
			return nil, fmt.Errorf("finmap: SHP record %d: %w", i, err)
		}
		for _, t := range r.Units() {
			if t >= TypeCount {
				return nil, fmt.Errorf("finmap: SHP record %d holds type %d", i, t)
			}
		}
		s.Records[i] = r
	}
	return s, nil
}

func (r Record) checkIndex() error {
	limit := map[Kind]int{HQ: Containers - HQBase, Factory: TableBuilding, Depot: TableBuilding, Transporter: Transporters}
	l, ok := limit[r.Kind]
	switch {
	case !ok:
		return fmt.Errorf("kind %d", int(r.Kind))
	case r.Index >= l:
		return fmt.Errorf("%v index %d", r.Kind, r.Index)
	}
	return nil
}

// Computer is a decoded COM file: six bytes per unit type that only the
// computer player reads. Their meaning is not decoded.
type Computer [TypeCount][6]byte

// DecodeComputer decodes an unpacked COM file.
func DecodeComputer(b []byte) (*Computer, error) {
	if len(b) != 6*TypeCount {
		return nil, fmt.Errorf("finmap: COM has %d bytes, want %d", len(b), 6*TypeCount)
	}
	var c Computer
	for t := range TypeCount {
		copy(c[t][:], b[6*t:])
	}
	return &c, nil
}

// Code is a map's 10-byte record of CODES.DAT, indexed by map number.
type Code struct {
	Password string // bytes 0-4, a letter each: 0x11 is 'A'
	Raw5     int    // byte 5: not read for the level; 2 in every record
	Players  int    // byte 6: 1 (against the computer) or 2
	// Step is byte 7: a win moves on to map number + Step; 0 ends the
	// series.
	Step int
	// Byte 8: bit 1 winter (the winter terrain table and graphics), bit 0
	// the last map of a campaign (a win ends the campaign).
	Winter, CampaignEnd bool
	Raw9                int // byte 9: not read for the level; 0 in every record
}

// DecodeCodes decodes an unpacked CODES.DAT.
func DecodeCodes(b []byte) ([]Code, error) {
	if len(b)%10 != 0 {
		return nil, fmt.Errorf("finmap: CODES.DAT has %d bytes, not whole records", len(b))
	}
	out := make([]Code, len(b)/10)
	for i := range out {
		p := b[10*i : 10*i+10]
		pw := make([]byte, 5)
		for k, c := range p[:5] {
			if c < 0x11 || c > 0x11+25 {
				return nil, fmt.Errorf("finmap: CODES.DAT record %d letter %#02x", i, c)
			}
			pw[k] = 'A' + c - 0x11
		}
		if p[6] != 1 && p[6] != 2 {
			return nil, fmt.Errorf("finmap: CODES.DAT record %d players %d", i, p[6])
		}
		if p[8]&^3 != 0 {
			return nil, fmt.Errorf("finmap: CODES.DAT record %d flags %#02x", i, p[8])
		}
		out[i] = Code{
			Password: string(pw), Raw5: int(p[5]), Players: int(p[6]), Step: int(p[7]),
			Winter: p[8]&2 != 0, CampaignEnd: p[8]&1 != 0, Raw9: int(p[9]),
		}
	}
	return out, nil
}
