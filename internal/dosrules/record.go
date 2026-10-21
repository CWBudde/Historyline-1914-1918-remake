package dosrules

// The movement, route, container and visibility rules work on compact
// records: 26-byte live unit records, maps of two bytes per cell, the
// 0x1104-byte cost and reach-bit arrays and 0x1c-byte container records.
// board.go reads them into typed values.

// ContainerLen is the size of a container or building record.
const ContainerLen = 0x1c

func word(rec []byte, off int) int { return int(rec[off]) | int(rec[off+1])<<8 }

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

// MaxSteps bounds the steps of one move (a route has at most 790 cells).
const MaxSteps = 0x400
