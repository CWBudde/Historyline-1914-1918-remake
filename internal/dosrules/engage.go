package dosrules

// Unit is a live unit on the battlefield.
type Unit struct {
	Type           int // index into the profile's types
	Cell           int
	Size, Alt, Exp int
	Flags4, Flags6 int // live flag words: +4 side bits 1/2 and target classes, +6 bit 0x20 no award
}

// Battlefield is a DOS map: Width × Height cells numbered row by row, odd
// columns half a cell lower. Terrain gives a cell's combat class (0-8;
// missing cells have DefaultClass). MapByte0 is the low byte of cell 0 (its
// terrain index), which the support routine reads for off-map neighbours.
type Battlefield struct {
	Width, Height int
	Units         map[int]Unit
	Terrain       map[int]int
	MapByte0      int
}

// DefaultClass is the terrain class of cells without an entry; its terms are 0/0.
const DefaultClass = 7

func (b *Battlefield) class(cell int) int {
	if c, ok := b.Terrain[cell]; ok {
		return c
	}
	return DefaultClass
}

// at returns the unit in a cell.
func (b *Battlefield) at(cell int) (int, bool) {
	for idx, u := range b.Units {
		if u.Cell == cell {
			return idx, true
		}
	}
	return 0, false
}

// Outcome is the state after one attack order.
type Outcome struct {
	Battle           bool // false: no unit in the target cell, the order is skipped
	Defender         int
	ReturnFire       bool // the defender stood next to the attacker
	Pincer, Blocking int
	Result
	AttackerExp, DefenderExp int
	Destroyed                []int // units destroyed, attacker first
}

const (
	expCap      = 6
	noAwardFlag = 0x20
)

// award is the experience window after a battle (02dc:0908): a side that
// lost strength gives its opponent 1 unless its live +6 word has bit 0x20, a
// destroyed side gives 1 more; experience stops at 6.
func award(expA, expD int, lostA, lostD, deadA, deadD bool, flagsA, flagsD int) (int, int) {
	gainA, gainD := 0, 0
	if deadA {
		gainD++
	}
	if lostA && flagsA&noAwardFlag == 0 {
		gainD++
	}
	if deadD {
		gainA++
	}
	if lostD && flagsD&noAwardFlag == 0 {
		gainA++
	}
	return min(expCap, expA+gainA), min(expCap, expD+gainD)
}

// strength is the strength the experience window compares.
func strength(t Type, size, alt int) int {
	if t.alternate() {
		return alt
	}
	return size
}

// Engage resolves the attack order of unit attacker on the target cell as the
// order handler (02dc:0004 in the demo) does: the defender is the unit in
// the cell, it may fire back when it stands next to the attacker, Support
// rates the neighbours, Distance and the two cells' terrain classes feed
// Casualties, and the experience window compares the strengths before and
// after.
func (p Profile) Engage(b *Battlefield, attacker, target int) (Outcome, error) {
	def, ok := b.at(target)
	if !ok {
		return Outcome{}, nil
	}
	a, d := b.Units[attacker], b.Units[def]
	ta, td := p.Tables.Types[a.Type], p.Tables.Types[d.Type]
	o := Outcome{Battle: true, Defender: def}
	for _, c := range neighbours(a.Cell, b.Width, b.Height) {
		o.ReturnFire = o.ReturnFire || c == d.Cell
	}
	var err error
	if o.Pincer, o.Blocking, err = p.Support(b, attacker, def); err != nil {
		return o, err
	}
	ctx := Context{
		Distance: p.Distance(b.Width, a.Cell, d.Cell),
		ATerrain: b.class(a.Cell), DTerrain: b.class(d.Cell),
		Pincer: o.Pincer, Blocking: o.Blocking, ReturnFire: o.ReturnFire,
	}
	if o.Result, err = p.Casualties(
		Fighter{ta, a.Size, a.Alt, a.Exp, a.Flags4},
		Fighter{td, d.Size, d.Alt, d.Exp, d.Flags4}, ctx); err != nil {
		return o, err
	}

	beforeA, afterA := strength(ta, a.Size, a.Alt), strength(ta, o.AttackerSize, o.AttackerAlt)
	beforeD, afterD := strength(td, d.Size, d.Alt), strength(td, o.DefenderSize, o.DefenderAlt)
	o.AttackerExp, o.DefenderExp = award(a.Exp, d.Exp,
		(beforeA-afterA)&0xFF != 0, (beforeD-afterD)&0xFF != 0, afterA == 0, afterD == 0, a.Flags6, d.Flags6)
	if afterA == 0 {
		o.Destroyed = append(o.Destroyed, attacker)
	}
	if afterD == 0 {
		o.Destroyed = append(o.Destroyed, def)
	}
	return o, nil
}
