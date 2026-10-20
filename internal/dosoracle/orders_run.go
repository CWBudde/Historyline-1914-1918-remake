package dosoracle

import (
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dos8086"
)

// RunOrders executes the order handler over a whole list of attack orders,
// on the frame RunEngagement uses, until its loop ends (Handler.Done). Each
// order's outcome is read when its battle ends (Handler.Stop); the runner
// then goes on at the loop's test (Handler.Skip), past the destruction
// animation, which only draws and restores the map bytes it changed. The
// destroy helper runs as in the game; the units handed to it are recorded
// per order.
func (e *Export) RunOrders(g *OrderList) (Resolution, error) {
	m, err := e.handlerFrame(&g.Engagement, g.Orders, g.Idle, nil)
	if err != nil {
		return Resolution{}, err
	}
	b, h := e.B, e.B.Handler
	data, code := baseSeg+b.Data, baseSeg+b.Code
	mark := func(idx int) bool { return m.RB(data, b.Units+26*idx+0x14)&4 != 0 }
	before := map[int]bool{}
	for i := range g.Units {
		before[i] = mark(i)
	}

	var r Resolution
	limit := engagementMaxSteps * (len(g.Orders) + 1)
	for int(m.S[dos8086.CS]) != code || int(m.IP) != h.Done {
		if int(m.S[dos8086.CS]) == code {
			switch int(m.IP) {
			case h.Next:
				r.Orders = append(r.Orders, Outcome{})
			case b.ExpDestroy:
				// "push cs; call near": the unit index lies above both
				o := &r.Orders[len(r.Orders)-1]
				o.Destroyed = append(o.Destroyed, m.RB(int(m.S[dos8086.SS]), int(m.R[dos8086.SP])+4))
			case h.Stop:
				o := &r.Orders[len(r.Orders)-1]
				out := e.outcome(m)
				out.Destroyed = o.Destroyed
				*o = out
				m.IP = uint16(h.Skip)
				continue
			}
		}
		if m.Steps++; m.Steps > limit {
			return r, wrapRun("order handler", dos8086.ErrStepLimit)
		}
		if err := m.Step(); err != nil {
			return r, wrapRun("order handler", err)
		}
	}

	r.Units = map[int]Fighter{}
	for i, u := range g.Units {
		base := b.Units + 26*i
		u.Exp, u.Size, u.Alt = m.RB(data, base+1), m.RB(data, base+2), m.RB(data, base+3)
		r.Units[i] = u
		if mark(i) && !before[i] {
			r.Marked = append(r.Marked, i)
		}
	}
	slices.Sort(r.Marked)
	ptr := h.Maps + 4*(1-g.Side&1) // Maps holds side 1's map, then side 0's
	seg, off := m.RW(data, ptr+2), m.RW(data, ptr)
	r.Cells = make([]int, g.Width*g.Height)
	for c := range r.Cells {
		r.Cells[c] = m.RB(seg, off+2*c+1)
	}
	return r, nil
}
