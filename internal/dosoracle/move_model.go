package dosoracle

import "errors"

// The movement queries of segment 0b06 (demo addresses) work on a byte per
// cell in arrays of 64 cells per row (index row*64+column, 0x1104 bytes);
// the map itself is two bytes per cell in rows of Width cells (terrain
// record, unit index, 0xf1 and up empty). Column and row are cell % Width and
// cell / Width; odd columns sit half a cell lower.

// bufLen is the size of the cost buffer and of the reach-bit array.
const bufLen = 0x1104

// Cost buffer markers (signed bytes).
const (
	costBlocked = -80  // 0xb0: terrain, a neutral or enemy unit, or another unit for a ship
	zocSide0    = -102 // 0x9a: next to an enemy of side 0
	zocSide1    = -101 // 0x9b: next to an enemy of side 1
)

// ErrContainer marks a destination check that reaches the transporter or
// building branch of the end-cell check, which is not transcribed (R2 item 3).
var ErrContainer = errors.New("container branch of the end-cell check")

// MoveType holds the fields of a type record the movement queries read.
type MoveType struct {
	Budget int `json:"budget"` // byte 0, copied into live byte 0 by the constructor
	Speed  int `json:"speed"`  // byte 1, the displayed speed (not read here)
	Mask   int `json:"mask"`   // byte 5, ANDed with the terrain mask
	Class  int `json:"class"`  // word +0x0d, copied into live +4 (0x40: a ship two cells long)
	Flags  int `json:"flags"`  // word +0x11, copied into live +6 (0x10: a flat cost of 2)

	// The path score weighs the distance to the goal by +0x15 and the entry
	// cost by +0x16 (signed bytes).
	DistanceWeight int `json:"distanceWeight"`
	CostWeight     int `json:"costWeight"`
}

// MoveTerrain is a 6-byte terrain record as the movement queries read it.
type MoveTerrain struct {
	Flags   int `json:"flags"`   // word 0: 0x4000 closed to air units; 0x400/0x100/0x40 buildings
	Mask    int `json:"mask"`    // byte 2
	Cost    int `json:"cost"`    // byte 3
	AirCost int `json:"airCost"` // byte 4, the cost for units with live +4 bit 0x10
}

// MoveTables are a build's type and terrain records.
type MoveTables struct {
	Types    []MoveType    `json:"types"`
	Terrains []MoveTerrain `json:"terrains"`
}

// Piece is a live unit on a movement map: flags at +4 (side bits 1/2,
// class bits from the type's +0x0d), flags at +6, type and cell.
type Piece struct {
	Type   int `json:"type"`
	Cell   int `json:"cell"`
	Flags4 int `json:"flags4"`
	Flags6 int `json:"flags6"`
}

// Movement is one movement query of the live unit Unit for player Side.
// Terrain holds a terrain record per cell. Units are the units on the side's
// own map; Unseen ones are only on the other side's map, which none of the
// queries read. Budget is the argument the callers take from live byte 0;
// Block is the terrain flag mask that closes cells (the player's order code
// passes 0xffff for live +6 bit 1, else 6 for a side-1 unit and 3 for
// others). Goals are the path targets.
type Movement struct {
	Width, Height int
	Terrain       []int
	Units         map[int]Piece
	Unseen        map[int]Piece
	Unit, Side    int
	Budget, Block int
	Goals         []int
}

// occupants maps the cells of the side's map to unit indices.
func (g *Movement) occupants() map[int]int {
	o := map[int]int{}
	for i, u := range g.Units {
		o[u.Cell] = i
	}
	return o
}

// liveLen is the size of a live unit record (Build.Units).
const liveLen = 26

// liveRecords are the live records of Units and Unseen as the queries read
// them: flags +4 and +6, the type at +8, the position 2*cell+1 at +0x0b and
// +0x0d (the cell as each side sees it) and, for the moving unit, the budget
// in byte 0.
func liveRecords(g *Movement) map[int][]byte {
	recs := map[int][]byte{}
	place := func(idx int, u Piece) {
		r := make([]byte, liveLen)
		putWord(r, 4, u.Flags4)
		putWord(r, 6, u.Flags6)
		r[8] = byte(u.Type)
		for k := range 2 {
			putWord(r, 0x0B+2*k, 2*u.Cell+1)
		}
		recs[idx] = r
	}
	for idx, u := range g.Units {
		place(idx, u)
	}
	for idx, u := range g.Unseen {
		place(idx, u)
	}
	recs[g.Unit][0] = byte(g.Budget)
	return recs
}

// movementMaps are the side's map, holding Units, and the other side's map,
// holding Units and Unseen: the terrain record and the unit index per cell
// (0xff empty).
func movementMaps(g *Movement) (own, other []byte) {
	own, other = make([]byte, 2*g.Width*g.Height), make([]byte, 2*g.Width*g.Height)
	for c, t := range g.Terrain {
		own[2*c], own[2*c+1] = byte(t), 0xFF
		other[2*c], other[2*c+1] = byte(t), 0xFF
	}
	for idx, u := range g.Units {
		own[2*u.Cell+1], other[2*u.Cell+1] = byte(idx), byte(idx)
	}
	for idx, u := range g.Unseen {
		other[2*u.Cell+1] = byte(idx)
	}
	return own, other
}

func putWord(b []byte, off, v int) { b[off], b[off+1] = byte(v), byte(v>>8) }

// neighbours64 is 0b06:005f with its edge clip 0b06:021f: the six neighbours
// of (col, row) as indices of 64 per row, -1 off the map.
func neighbours64(col, row, width, height int) [6]int {
	i := row*64 + col
	var n [6]int
	n[0], n[3] = i-64, i+64
	if col&1 != 0 {
		n[1], n[5], n[4], n[2] = i+1, i-1, i+63, i+65
		if row <= 0 {
			n[0] = -1
		} else if height-1 <= row {
			n[3], n[4], n[2] = -1, -1, -1
		}
	} else {
		n[2], n[4], n[5], n[1] = i+1, i-1, i-65, i-63
		if row <= 0 {
			n[0], n[5], n[1] = -1, -1, -1
		} else if height-1 <= row {
			n[3] = -1
		}
	}
	if col <= 0 {
		n[4], n[5] = -1, -1
	} else if width-1 <= col {
		n[1], n[2] = -1, -1
	}
	return n
}

// costs is 0b06:09c1: the negated entry cost of every cell, or a marker.
func (t *MoveTables) costs(g *Movement) []int8 {
	buf := make([]int8, bufLen)
	for i := range buf {
		buf[i] = -1
	}
	u := g.Units[g.Unit]
	typ := t.Types[u.Type]
	mode := 2 // ground: byte 3
	if u.Flags4&0x10 != 0 {
		mode = 0 // air: byte 4
	}
	if u.Flags6&0x10 != 0 {
		mode = 1 // a flat cost of 2
	}
	for row := range g.Height {
		for col := range g.Width {
			tr := t.Terrains[g.Terrain[row*g.Width+col]]
			v := int8(costBlocked)
			if tr.Mask&typ.Mask != 0 {
				switch mode {
				case 0:
					v = int8(-tr.AirCost)
				case 1:
					v = -2
				default:
					v = int8(-tr.Cost)
				}
			}
			if g.Block != 0xFFFF && tr.Flags&g.Block != 0 {
				v = costBlocked
			}
			buf[row*64+col] = v
		}
	}
	occ := g.occupants()
	for row := range g.Height {
		for col := range g.Width {
			idx, ok := occ[row*g.Width+col]
			if !ok {
				continue
			}
			i := row*64 + col
			f := g.Units[idx].Flags4
			zoc, own := 0, false
			switch {
			case f&2 != 0:
				buf[i] = costBlocked
			case f&1 != 0:
				own, zoc = g.Side == 1, zocSide0
			default:
				own, zoc = g.Side == 0, zocSide1
			}
			if zoc != 0 {
				if own {
					if f&0x3000 != 0 {
						buf[i] = -2
					}
				} else {
					buf[i] = costBlocked
					for _, n := range neighbours64(col, row, g.Width, g.Height) {
						if n >= 0 && buf[n] != costBlocked {
							buf[n] = int8(zoc)
						}
					}
				}
			}
			if typ.Class&0x40 != 0 && idx != g.Unit {
				buf[i] = costBlocked
			}
		}
	}
	return buf
}

// Reach is 0b06:0331: the cost buffer after the search (a value >= 0 is the
// budget left on reaching that cell), the side's reach bits it ORs into
// 0x3b50 and the number of reached cells. The search expands, from the
// budget down to 1, every cell holding that value inside a window of Budget
// cells around the start; a neighbour still negative takes the value minus
// its cost if that is not negative, and a zone-of-control cell takes 0.
func (t *MoveTables) Reach(g *Movement) (buf []int8, bits []byte, count int) {
	buf = t.costs(g)
	start := g.Units[g.Unit].Cell
	col, row := start%g.Width, start/g.Width
	half := g.Budget >> 1
	colLo, rowLo := max(col-half, 0), max(row-half, 0)
	colHi, rowHi := min(colLo+g.Budget, g.Width), min(rowLo+g.Budget, g.Height)
	zoc := int8(zocSide0)
	if g.Side != 0 {
		zoc = zocSide1
	}
	buf[row*64+col] = int8(g.Budget)
	for level := int(buf[row*64+col]); level > 0; level-- {
		for r := rowLo; r < rowHi; r++ {
			for c := colLo; c < colHi; c++ {
				if int(buf[r*64+c]) != level {
					continue
				}
				for _, n := range neighbours64(c, r, g.Width, g.Height) {
					switch {
					case n < 0 || buf[n] >= 0:
					case buf[n] == zoc:
						buf[n] = 0
					case int(buf[n])+level >= 0:
						buf[n] += int8(level)
					}
				}
			}
		}
	}
	bit := byte(1)
	if g.Side == 1 {
		bit = 2
	}
	bits = make([]byte, bufLen)
	for i, v := range buf {
		if v >= 0 {
			count++
			bits[i] |= bit
		}
	}
	return buf, bits, count
}

// own is 0ca0:020c: whether live flags +4 belong to player side.
func own(flags, side int) bool {
	switch {
	case flags&1 != 0:
		return side&1 != 0
	case flags&2 != 0:
		return side&2 != 0
	}
	return side&1 == 0
}

// EndCode is 10cc:0f74 for the moving unit: 0 if it may end on the cell,
// 0x37 for a unit of another side there, 0x39 for an own unit that carries
// nothing (the moving unit's own cell included) and 0x40 for terrain its
// mask does not enter or, for an air unit, terrain flag 0x4000. An own
// transporter or a building cell gives ErrContainer.
func (t *MoveTables) EndCode(g *Movement, cell int) (int, error) {
	mover := g.Units[g.Unit]
	if idx, ok := g.occupants()[cell]; ok {
		o := g.Units[idx]
		if o.Flags4&0x40 != 0 && o.Flags4&0x80 == 0 {
			o = g.Units[idx-1] // the rear of a two-cell ship
		}
		switch {
		case !own(o.Flags4, g.Side):
			return 0x37, nil
		case o.Flags4&0x1000 == 0:
			return 0x39, nil
		}
		return 0, ErrContainer
	}
	tr := t.Terrains[g.Terrain[cell]]
	if tr.Flags&(0x400|0x100|0x40) != 0 {
		return 0, ErrContainer
	}
	if t.Types[mover.Type].Mask&tr.Mask == 0 || mover.Flags4&0x10 != 0 && tr.Flags&0x4000 != 0 {
		return 0x40, nil
	}
	return 0, nil
}

// Destinations is 0b06:1383 after Reach: the cells with the side's reach bit,
// column by column, whose end code is 0 (at most 790).
func (t *MoveTables) Destinations(g *Movement, bits []byte) ([]int, error) {
	bit := byte(1)
	if g.Side != 0 {
		bit = 2
	}
	var out []int
	for col := range g.Width {
		for row := range g.Height {
			if bits[row*64+col]&bit == 0 {
				continue
			}
			code, err := t.EndCode(g, row*g.Width+col)
			if err != nil {
				return nil, err
			}
			if code == 0 {
				if out = append(out, row*g.Width+col); len(out) >= 0x316 {
					return out, nil
				}
			}
		}
	}
	return out, nil
}

// pathNode is a 10-byte node of the path search: column, row, parent, score
// and the sorted list links.
type pathNode struct{ col, row, parent, score, prev, next int }

// Path is 0b06:0e12 after Reach: a best-first search over the cells with the
// side's reach bit, in a list sorted by score = DistanceWeight x distance to
// the goal + CostWeight x entry cost (0 at the goal), each cell entered once.
// It returns the cells from the goal back to the start, or nil when the
// search runs dry or out of its 790 nodes. The score is not summed along the
// route, so the route need not be the cheapest. The node pool is the buffer
// the reach search left (buf); the sentinel's column and row stay buffer
// bytes 20 and 21, so a goal they match, a score past the sentinel's 30000,
// or expanding the sentinel past the bit array (a search that runs dry)
// gives ErrUndefined. The order code only asks for destinations.
func (t *MoveTables) Path(g *Movement, buf []int8, bits []byte, goal int) ([]int, error) {
	w := g.Width
	u := g.Units[g.Unit]
	typ := t.Types[u.Type]
	reach, visit := byte(1), byte(0x40)
	if g.Side != 0 {
		reach, visit = 2, 0x80
	}
	bits = append([]byte(nil), bits...)
	gcol, grow := goal%w, goal/w
	start := u.Cell
	nodes := make([]pathNode, 0x317+1)
	nodes[0].next = 1
	nodes[1] = pathNode{col: start % w, row: start / w, parent: 1, next: 2}
	nodes[2] = pathNode{col: int(byte(buf[20])), row: int(byte(buf[21])), score: 30000, prev: 1}
	free := 3
	for cur := nodes[0].next; cur != 0; cur = nodes[cur].next {
		n := nodes[cur]
		if n.col == gcol && n.row == grow {
			if cur == 2 {
				return nil, ErrUndefined // the sentinel's column and row are buffer bytes
			}
			route := []int{grow*w + gcol}
			for cur > 1 {
				cur = nodes[cur].parent
				route = append(route, nodes[cur].row*w+nodes[cur].col)
				if len(route) > 0x316 {
					return nil, nil
				}
			}
			return route, nil
		}
		for _, k := range neighbours64(n.col, n.row, g.Width, g.Height) {
			if k >= len(bits) {
				// The sentinel's column and row (buffer bytes 20 and 21) lie
				// past the array: the search reads the data behind it.
				return nil, ErrUndefined
			}
			if k < 0 || bits[k]&reach == 0 || bits[k]&visit != 0 {
				continue
			}
			ncol, nrow := k%64, k/64
			d := int(int8(Distance(ncol, nrow, gcol, grow)))
			// A cell past the right edge (see neighbours64) reads the map
			// bytes of the next row's first cell.
			cell := nrow*w + ncol
			if cell >= len(g.Terrain) {
				return nil, ErrUndefined
			}
			tr := t.Terrains[g.Terrain[cell]]
			cost := tr.Cost
			if u.Flags4&0x10 != 0 {
				cost = tr.AirCost
			}
			score := 0
			if d != 0 {
				score = int(int16(d*typ.DistanceWeight + int(int8(cost))*typ.CostWeight))
			}
			at := nodes[cur].next
			for nodes[at].score <= score {
				if at = nodes[at].next; at == 0 {
					return nil, ErrUndefined // past the sentinel into the list head
				}
			}
			nodes[free] = pathNode{col: ncol, row: nrow, parent: cur, score: score, prev: nodes[at].prev, next: at}
			nodes[nodes[at].prev].next = free
			nodes[at].prev = free
			bits[k] |= visit
			if free++; free > 0x316 {
				return nil, nil
			}
		}
	}
	return nil, nil
}
