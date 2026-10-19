package dosoracle

// Distance transcribes the hex distance 0e40:2421 (0e47:2414 in the full
// build). The casualty routine passes the attacker's and the defender's
// (row, column) with row = cell / width.
func Distance(r1, c1, r2, c2 int) int {
	if c1 > c2 {
		r1, r2, c1, c2 = r2, r1, c2, c1
	}
	if r2 >= r1 {
		c1 -= r1 / 2
		c2 -= r2 / 2
	} else {
		r1, r2 = r2, r1
		c1 -= r1/2 + r1%2
		c2 -= r2/2 + r2%2
	}
	return r2 - r1 + max(0, c2-c1)
}

// Fighter is a live unit placed for an engagement. Flags4 and Flags6 are the
// live flag words at +4 (side bits 1/2, class bits 0x3c) and +6.
type Fighter struct{ Type, Cell, Size, Alt, Exp, Flags4, Flags6 int }

// Engagement is one controlled battle: units on a map and a single attack
// order of the live unit Attacker against the cell Target. Terrain gives a
// cell's combat class (0-8, default 7); Side is the player whose orders run.
// Units stand at Cell as Side sees them; Moved gives a unit's cell as the
// other side sees it where the two differ (a move of the other side during
// the phase), else both sides' maps and position words agree.
type Engagement struct {
	Width, Height    int
	Units            map[int]Fighter
	Attacker, Target int
	Side             int
	Terrain          map[int]int
	Moved            map[int]int
}

// Outcome is the state the order handler leaves after the battle.
type Outcome struct {
	Battle           bool // false: no unit at the target, the order was skipped
	Defender         int
	ReturnFire       bool
	Pincer, Blocking int
	Result
	AttackerExp, DefenderExp int
	Destroyed                []int // live units handed to the destroy helper, in call order
}

const defaultClass = 7 // terrain terms 0/0

func (g *Engagement) class(cell int) int {
	if c, ok := g.Terrain[cell]; ok {
		return c
	}
	return defaultClass
}

// cells is the unit byte of every map cell (0xff empty) as Side sees it.
func (g *Engagement) cells() []int {
	c := make([]int, g.Width*g.Height)
	for i := range c {
		c[i] = 0xFF
	}
	for idx, u := range g.Units {
		c[u.Cell] = idx
	}
	return c
}

// seen is a unit's cell as the other side sees it.
func (g *Engagement) seen(idx int) int {
	if c, ok := g.Moved[idx]; ok {
		return c
	}
	return g.Units[idx].Cell
}

// strength is the group size the experience window compares: the alternate
// byte when the type record's word at +0x0f has bit 4.
func strength(rec []byte, size, alt int) int {
	if word(rec, 0x0F)&4 != 0 {
		return alt
	}
	return size
}

// Engage composes the transcriptions in the order of the handler 02dc:0004:
// the defender is the unit in the target cell; return fire is allowed when
// that cell is one of the attacker's six neighbours; Support rates the
// neighbours with the type records' target-flag word (+6); Distance and the
// cells' terrain classes feed Resolve; Award compares the strengths saved
// before the battle with those after it. rec returns a type's 69-byte record.
func Engage(g *Engagement, rec func(int) []byte, terms [9][2]int) (Outcome, error) {
	cells := g.cells()
	idx := cells[g.Target]
	if idx > 0xF0 {
		return Outcome{}, nil
	}
	a, d := g.Units[g.Attacker], g.Units[idx]
	ra, rd := rec(a.Type), rec(d.Type)
	o := Outcome{Battle: true, Defender: idx}
	for _, c := range Ring(a.Cell, g.Width, g.Height) {
		if c == d.Cell {
			o.ReturnFire = true
		}
	}

	m := &Map{Width: g.Width, Height: g.Height, Cells: cells, Terrain0: g.class(0),
		Units: map[int]Unit{}, Types: map[int]int{}}
	for i, u := range g.Units {
		m.Units[i] = Unit{Flags4: u.Flags4, Flags6: u.Flags6, Type: u.Type, Cell: u.Cell}
		m.Types[u.Type] = word(rec(u.Type), 6)
	}
	var err error
	if o.Pincer, o.Blocking, err = Support(m, g.Attacker, idx); err != nil {
		return o, err
	}

	w := g.Width
	p := Params{
		Distance: Distance(a.Cell/w, a.Cell%w, d.Cell/w, d.Cell%w),
		ATerrain: g.class(a.Cell), DTerrain: g.class(d.Cell),
		Pincer: o.Pincer, Blocking: o.Blocking, ReturnFire: o.ReturnFire, TerrainTerms: terms,
	}
	if o.Result, err = Resolve(Side{ra, a.Size, a.Alt, a.Exp, a.Flags4}, Side{rd, d.Size, d.Alt, d.Exp, d.Flags4}, p); err != nil {
		return o, err
	}

	var kills []int
	o.AttackerExp, o.DefenderExp, kills = Award(
		ExpSide{a.Exp, strength(ra, a.Size, a.Alt), strength(ra, o.AttackerSize, o.AttackerAlt), a.Flags6},
		ExpSide{d.Exp, strength(rd, d.Size, d.Alt), strength(rd, o.DefenderSize, o.DefenderAlt), d.Flags6})
	for _, k := range kills {
		o.Destroyed = append(o.Destroyed, [2]int{g.Attacker, idx}[k])
	}
	return o, nil
}
