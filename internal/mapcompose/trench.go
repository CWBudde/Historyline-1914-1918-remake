package mapcompose

import (
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// Trench marks a Gap of a neighbour-derived trench tile. It is not a context
// family: the tile ID is the descriptor.
const Trench Family = "trench"

// Trench hexes draw a course derived from their neighbours, so trench lines
// join across hexes. The levels keep the original's trench pieces: a set of
// ports (trench14, trench124, …) or the round dug-in position trench0. A
// hex's edge is a link when both hexes hold a trench and either piece has a
// port across it; a port beyond the map stays a link. A port toward a hex
// without a trench is a blind spur that ends inside the hex. The position
// takes its links as courses that end at its ring.

// TrenchShape is a derived trench: the ring of a dug-in position, the linked
// ports and the blind spurs, as masks of bit port-1.
type TrenchShape struct {
	Ring         bool
	Links, Spurs uint8
}

// ID is the shape's tile name: trench<links>[-s<spurs>], or trench0[-<links>]
// for a position. Ports are digits in ascending order.
func (s TrenchShape) ID() string {
	if s.Ring {
		if s.Links == 0 {
			return "trench0"
		}
		return "trench0-" + portDigits(s.Links)
	}
	id := "trench" + portDigits(s.Links)
	if s.Spurs != 0 {
		id += "-s" + portDigits(s.Spurs)
	}
	return id
}

// ParseTrench reads a derived trench ID.
func ParseTrench(id string) (TrenchShape, bool) {
	rest, ok := strings.CutPrefix(id, "trench")
	if !ok {
		return TrenchShape{}, false
	}
	if ring, ok := strings.CutPrefix(rest, "0"); ok {
		links, ok := parsePorts(strings.TrimPrefix(ring, "-"))
		return TrenchShape{Ring: true, Links: links}, ok && (ring == "") == (links == 0)
	}
	head, tail, spurred := strings.Cut(rest, "-s")
	links, ok := parsePorts(head)
	if !ok {
		return TrenchShape{}, false
	}
	spurs, ok := parsePorts(tail)
	if !ok || spurred != (spurs != 0) || links&spurs != 0 || links|spurs == 0 {
		return TrenchShape{}, false
	}
	return TrenchShape{Links: links, Spurs: spurs}, true
}

func portDigits(m uint8) string {
	var b strings.Builder
	for p := 1; p <= 6; p++ {
		if m&(1<<(p-1)) != 0 {
			b.WriteByte(byte('0' + p))
		}
	}
	return b.String()
}

// parsePorts reads ascending port digits.
func parsePorts(s string) (uint8, bool) {
	var m uint8
	last := 0
	for _, r := range s {
		p := int(r - '0')
		if p <= last || p > 6 {
			return 0, false
		}
		m |= 1 << (p - 1)
		last = p
	}
	return m, true
}

// trenchPiece reads a hex's authored trench: its infrastructure ID and ports,
// with ring set for the dug-in position. Variant letters (trench5a) keep the
// ports of their digits.
func trenchPiece(infra []string) (id string, ports uint8, ring, ok bool) {
	for _, id := range infra {
		rest, trench := strings.CutPrefix(id, "trench")
		if !trench || rest == "" {
			continue
		}
		if rest == "0" {
			return id, 0, true, true
		}
		for _, r := range rest {
			if r < '1' || r > '6' {
				break
			}
			ports |= 1 << (r - '1')
		}
		return id, ports, false, ports != 0
	}
	return "", 0, false, false
}

func trenchShape(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord) (TrenchShape, bool) {
	_, own, ring, ok := trenchPiece(at(c).Infra)
	if !ok {
		return TrenchShape{}, false
	}
	s := TrenchShape{Ring: ring}
	for i, n := range hexgrid.Around(c) {
		bit := uint8(1) << i
		if !g.Contains(n) {
			// Continue the course beyond the visible map boundary.
			s.Links |= own & bit
			continue
		}
		_, theirs, _, trench := trenchPiece(at(n).Infra)
		switch {
		case trench && (own&bit != 0 || theirs&(1<<((i+3)%6)) != 0):
			s.Links |= bit
		case own&bit != 0:
			s.Spurs |= bit
		}
	}
	return s, true
}

// trenches returns the infrastructure drawn on hex c: its trench piece as the
// derived shape where the catalog has it. A missing tile keeps the piece and
// records a Gap. The rules still see the piece.
func trenches(g hexgrid.Grid, at func(hexgrid.Coord) Hex, c hexgrid.Coord, infra []string, has func(string) bool) ([]string, *Gap) {
	piece, ports, ring, ok := trenchPiece(infra)
	if !ok {
		return infra, nil
	}
	s, _ := trenchShape(g, at, c)
	if s.Ring == ring && s.Links == ports && s.Spurs == 0 {
		return infra, nil
	}
	id := s.ID()
	if !has(TileID(id)) {
		return infra, &Gap{At: c, Terrain: at(c).Terrain, Family: Trench, Reason: Missing, Descriptor: id}
	}
	out := make([]string, len(infra))
	for i, x := range infra {
		if x == piece {
			x = id
		}
		out[i] = x
	}
	return out, nil
}
