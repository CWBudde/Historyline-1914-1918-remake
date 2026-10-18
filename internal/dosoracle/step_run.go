package dosoracle

import (
	"errors"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

// Offsets in the side records (Move.Sides) and the order record.
const (
	sideLen        = 0x17
	sideOrders     = 4    // far pointer to the side's order list
	sideOrderCount = 8    // its count byte
	orderDest      = 0    // word: 2*cell of the destination
	orderKind      = 0x19 // byte; 3 takes the first saved occupant from +0x24
	orderUnit      = 0x1e // word: the unit index
	containerLen   = 0x1c
)

// MoveOrder is a move order of the moving unit: the destination cell, the
// cell the order is changed to between the order code and the start (-1:
// none), and cells whose reach bit for the side is left from an earlier
// query (the reach search ORs its bits into the array; its callers clear it).
type MoveOrder struct {
	Goal     int   `json:"goal"`
	Retarget int   `json:"retarget"`
	Stale    []int `json:"stale,omitempty"`
}

// StepRun is the executed outcome of one move order.
type StepRun struct {
	Order, Start int            // AX of the order code and of the move start (1, 0x12 refused, |8 two cells)
	Messages     [][2]int       // status messages shown: code, side
	Cells        []int          // the mover's cell as its side sees it, after each step
	Live         map[int][]byte // the live record of every placed unit afterwards
	Maps         [2][]byte      // the side's and the other side's map afterwards
	Bits         []byte         // the bit array afterwards
	Orders       []int          // the side's order list afterwards
	Count        int            // the route count afterwards
	Holds        map[int][]byte // the container records afterwards
	Container    []byte         // container record 0 afterwards
}

// stepMaxCalls bounds the steps of one move (a route has at most 790 cells).
const stepMaxCalls = 0x400

// RunMove executes a move order the way the player gives it: the reach search
// 0b06:0331 with the player's arguments (0ca0:0a6f), the order code 10cc:02dd
// to goal (0ca0:07d0), the move start 10cc:05c3 (0ca0:0912) and the step
// 10cc:071e as the main loop calls it (0661:43a8), once per call until the
// step that arrives (10cc:0ae2). Hooked: the stack check, the cell and route
// drawing, the redraw, the sound and the status message, whose codes are
// recorded. The container records are Movement.Holds.
func (e *Export) RunMove(g *Movement, o MoveOrder) (StepRun, error) {
	v, r, err := e.stepEnv(g)
	if err != nil {
		return StepRun{}, err
	}
	for _, c := range o.Stale {
		v.m.WB(v.data, v.mv.Bits+c/g.Width*64+c%g.Width, 1+g.Side)
	}
	if _, _, _, err = v.reach(g); err != nil {
		return *r, err
	}
	rec := make([]byte, 0x30)
	putWord(rec, orderUnit, g.Unit)
	return v.drive(g, o, r, v.ar.alloc(rec))
}

// stepEnv is moveEnv with the hooks of the move routines and the side
// record: the buffer the reach search fills and a guarded order list.
func (e *Export) stepEnv(g *Movement) (*moveEnv, *StepRun, error) {
	v, err := e.moveEnv(g)
	if err != nil {
		return nil, nil, err
	}
	m, data, mv := v.m, v.data, v.mv
	r := &StepRun{}
	m.Hooks[e.lin(mv.Redraw)] = func(mm *dos8086.Machine) error {
		// 0e40 calls it with "push cs; call near": drop the pushed CS the
		// hook's early return leaves (a far call arrives at the entry).
		if int(mm.IP) != mv.Redraw[1] {
			mm.R[dos8086.SP] += 2
		}
		return nil
	}
	for _, at := range [][2]int{mv.RouteDraw, mv.Refresh, mv.Sound, mv.View[0], mv.View[1]} {
		m.Hooks[e.lin(at)] = noHook
	}
	m.Hooks[e.lin(mv.Message)] = func(mm *dos8086.Machine) error {
		ss, sp := int(mm.S[dos8086.SS]), int(mm.R[dos8086.SP])
		r.Messages = append(r.Messages, [2]int{mm.RW(ss, sp+4), mm.RW(ss, sp+6)})
		return nil
	}
	side := mv.Sides + sideLen*g.Side
	m.WW(data, side, 0)
	m.WW(data, side+2, v.cost)
	v.orders = v.guarded(16)
	m.WW(data, side+sideOrders, 0)
	m.WW(data, side+sideOrders+2, v.orders)
	m.WB(data, side+sideOrderCount, 0)
	return v, r, nil
}

// drive runs the order code, the start and the steps with the order record
// at segment order, whose destination it sets, and reads the outcome.
func (v *moveEnv) drive(g *Movement, o MoveOrder, r *StepRun, order int) (StepRun, error) {
	m, data, mv := v.m, v.data, v.mv
	var err error
	m.WW(data, mv.Count, 0)
	m.WW(order, orderDest, 2*o.Goal)
	if r.Order, err = v.call("order", mv.Order, 0, order, 0, v.sideMap, g.Side); err != nil {
		return *r, err
	}
	if o.Retarget >= 0 {
		m.WW(order, orderDest, 2*o.Retarget)
	}
	if r.Order&1 != 0 {
		if r.Start, err = v.call("start", mv.Start, 0, order, g.Side); err != nil {
			return *r, err
		}
	}
	unit := v.units + liveLen*g.Unit
	for r.Start&1 != 0 {
		if len(r.Cells) == stepMaxCalls {
			return *r, errors.New("step: no arrival")
		}
		entry := m.RW(data, mv.Count)
		if _, err := v.call("step", mv.Step); err != nil {
			return *r, err
		}
		r.Cells = append(r.Cells, (m.RW(data, unit+0x0B+2*g.Side)-1)/2)
		if entry == 0 {
			break
		}
	}

	r.Live = map[int][]byte{}
	for idx := range liveRecords(g) {
		b := make([]byte, liveLen)
		copy(b, m.Mem[dos8086.Lin(data, v.units+liveLen*idx):])
		r.Live[idx] = b
	}
	for k, s := range [2]int{v.sideMap, v.otherMap} {
		r.Maps[k] = make([]byte, 2*g.Width*g.Height)
		copy(r.Maps[k], m.Mem[s*16:])
	}
	r.Bits = v.bits()
	for i := range m.RB(data, mv.Sides+sideLen*g.Side+sideOrderCount) {
		r.Orders = append(r.Orders, m.RB(v.orders, i))
	}
	for i := 16; i < 32; i++ {
		if m.RB(v.orders, i) != guard {
			return *r, errors.New("start: wrote past the order list")
		}
	}
	r.Count = int(int16(m.RW(data, mv.Count)))
	r.Holds = map[int][]byte{}
	for i := range holdRecords {
		b := make([]byte, containerLen)
		copy(b, m.Mem[dos8086.Lin(data, mv.Containers+containerLen*i):])
		r.Holds[i] = b
	}
	r.Container = r.Holds[0]
	return *r, nil
}
