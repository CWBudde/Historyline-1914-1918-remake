package dosoracle

import (
	"errors"
	"slices"
	"testing"
)

func record(armour, power byte) []byte {
	r := make([]byte, recordLen)
	r[2], r[8], r[9], r[10], r[11], r[12] = armour, 1, 1, 0, power, 0
	r[6] = 0x04
	r[0x0D] = 0x04
	return r
}

func side(rec []byte, size int) Side { return Side{Rec: rec, Size: size} }

func TestDivTruncatesTowardZero(t *testing.T) {
	if div(-7, 2) != -3 || div(7, -2) != -3 || div(7, 2) != 3 {
		t.Fatal("division does not truncate toward zero")
	}
	if _, err := Resolve(side(record(20, 30), 6), side(record(20, 30), 6), Params{}); err != nil {
		t.Fatal(err)
	}
}

func TestAdjust(t *testing.T) {
	for _, c := range [][3]int64{{20, 0, 20}, {100, 50, 150}, {20, 150, 49}} { // 13107*150 falls just short of 30.0
		if got := adjust(c[0], c[1]); got != c[2] {
			t.Errorf("adjust(%d,%d)=%d want %d", c[0], c[1], got, c[2])
		}
	}
}

func resolve(t *testing.T, a, d Side, p Params) Result {
	t.Helper()
	r, err := Resolve(a, d, p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestEqualInfantry(t *testing.T) {
	// power 30 vs armour 20, six men, experience 0: 80 margin / 70 denominator = 1 loss.
	r := resolve(t, side(record(20, 30), 6), side(record(20, 30), 6), Params{ReturnFire: true})
	if r.AttackerSize != 5 || r.DefenderSize != 5 {
		t.Fatalf("%+v", r)
	}
}

func TestNoReturnFireLeavesAttackerUntouched(t *testing.T) {
	r := resolve(t, side(record(20, 30), 6), side(record(20, 30), 6), Params{})
	if r.AttackerSize != 6 || r.DefenderSize != 5 {
		t.Fatalf("%+v", r)
	}
}

func TestLossesNeverExceedGroup(t *testing.T) {
	r := resolve(t, side(record(20, 250), 6), side(record(1, 30), 3), Params{})
	if r.DefenderSize != 0 {
		t.Fatalf("%+v", r)
	}
}

func TestBlockingAndPincerChangeTheResult(t *testing.T) {
	a, d := side(record(20, 45), 6), side(record(20, 30), 6)
	base := resolve(t, a, d, Params{}).DefenderSize
	pincer := resolve(t, a, d, Params{Pincer: 100}).DefenderSize
	block := resolve(t, a, d, Params{Blocking: 150}).DefenderSize
	if pincer >= base || block <= base {
		t.Fatalf("base=%d pincer=%d block=%d", base, pincer, block)
	}
}

func TestDivideErrorIsReturned(t *testing.T) {
	// ranged weapon with a zero armour record: 8-armour * 0 = 0 divisor
	a, d := side(record(0, 30), 6), side(record(0, 30), 6)
	if _, err := Resolve(a, d, Params{ReturnFire: true}); !errors.Is(err, ErrDivideError) {
		t.Fatalf("got %v", err)
	}
}

func exp(e, before, after, f6 int) ExpSide { return ExpSide{e, before, after, f6} }

func TestAward(t *testing.T) {
	cases := []struct {
		name   string
		a, d   ExpSide
		wa, wd int
		kills  []int
	}{
		{"no losses", exp(0, 10, 10, 0), exp(0, 10, 10, 0), 0, 0, nil},
		{"damage to defender", exp(0, 10, 10, 0), exp(0, 10, 7, 0), 1, 0, nil},
		{"damage to attacker", exp(0, 10, 7, 0), exp(0, 10, 10, 0), 0, 1, nil},
		{"kill awards two", exp(0, 10, 10, 0), exp(0, 10, 0, 0), 2, 0, []int{1}},
		{"flag withholds damage award", exp(0, 10, 10, 0), exp(0, 10, 7, 0x20), 0, 0, nil},
		{"flag keeps kill award", exp(0, 10, 10, 0), exp(0, 10, 0, 0x20), 1, 0, []int{1}},
		{"cap is six", exp(5, 10, 10, 0), exp(0, 10, 0, 0), 6, 0, []int{1}},
	}
	for _, c := range cases {
		wa, wd, k := Award(c.a, c.d)
		if wa != c.wa || wd != c.wd || !slices.Equal(k, c.kills) {
			t.Errorf("%s: got %d %d %v", c.name, wa, wd, k)
		}
	}
}

func TestRingEdges(t *testing.T) {
	if got := Ring(0, 4, 4); got != [6]int{-1, -1, 1, 4, -1, -1} {
		t.Fatalf("%v", got)
	}
	n := 0
	for _, c := range Ring(5, 4, 4) {
		if c >= 0 {
			n++
		}
	}
	if n != 6 {
		t.Fatalf("%d neighbours", n)
	}
}

func emptyMap(w, h int) *Map {
	cells := make([]int, w*h)
	for i := range cells {
		cells[i] = 0xFF
	}
	return &Map{Width: w, Height: h, Cells: cells, Types: map[int]int{0: 4}}
}

func TestSingleSupporterAndBlockers(t *testing.T) {
	m := emptyMap(7, 7)
	c := 3*7 + 3
	// attacker left of the defender, a friend beside the defender, an enemy beside the attacker
	m.Cells[c], m.Cells[c+1], m.Cells[c+2], m.Cells[c-1] = 0, 1, 2, 3
	m.Units = map[int]Unit{
		0: {Flags4: 1 | 4, Cell: c}, 1: {Flags4: 2 | 4, Cell: c + 1},
		2: {Flags4: 1 | 4, Cell: c + 2}, 3: {Flags4: 2 | 4, Cell: c - 1},
	}
	p, b, err := Support(m, 0, 1)
	if err != nil || p != 50 || b != 0 {
		t.Fatalf("%d %d %v", p, b, err)
	}
}

func TestFarApartUnitsGetNothing(t *testing.T) {
	m := emptyMap(7, 7)
	m.Cells[0], m.Cells[24] = 0, 1
	m.Units = map[int]Unit{0: {Flags4: 5, Cell: 0}, 1: {Flags4: 6, Cell: 24}}
	p, b, err := Support(m, 0, 1)
	if err != nil || p != 0 || b != 0 {
		t.Fatalf("%d %d %v", p, b, err)
	}
}
