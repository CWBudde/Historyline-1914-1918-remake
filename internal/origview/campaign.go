package origview

import (
	"fmt"
	"strconv"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosbind"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/finmap"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Campaign is the remake level of an imported map, the file levels/campaign
// holds for it: the view (Level) with the map's units by remake id
// (dosbind.RemakeIDs) and the units the container records put into the
// buildings as the entrances' content, in record order. The level format
// has no neutral units: a unit of a neutral record takes the first human
// side's remake unit, or the other side's where that side has none.
//
// The level format cannot hold everything the original rules bind
// (dosbind.FromOriginal stays the exact binding): a two-cell ship is its head
// on one hex, transporters start empty (their cargo is left out), terrain
// records that share a look become the same terrain, and credits, factory
// builds and the mission's next map are not part of the file. A unit type
// without a remake unit for its side is an error.
func Campaign(l *finmap.Level, set *data.Set) (*level.Level, error) {
	lv, err := Level(l)
	if err != nil {
		return nil, err
	}
	ids := dosbind.RemakeIDs(set)
	human := finmap.Central
	if m := l.Mission; m != nil && len(m.HumanSides) > 0 {
		human = m.HumanSides[0]
	}
	id := func(typ, side int) (string, error) {
		sides := []int{side}
		if side == finmap.Neutral {
			sides = []int{human, 1 - human}
		}
		for _, s := range sides {
			if id, ok := ids[[2]int{typ, s}]; ok {
				return id, nil
			}
		}
		return "", fmt.Errorf("origview: map %s: no remake unit for type %d of side %d", l.Source.Map, typ, side)
	}
	g := lv.Grid()
	for _, u := range l.Units {
		c := finmap.Coord(u.Cell[0], u.Cell[1])
		if !g.Contains(c) {
			return nil, fmt.Errorf("origview: map %s: unit at %v off the level", l.Source.Map, u.Cell)
		}
		h := &lv.Lines[c.Line][c.Col]
		if h.Unit != "" || h.Content != nil {
			return nil, fmt.Errorf("origview: map %s: unit at %v on an occupied hex or a building entrance", l.Source.Map, u.Cell)
		}
		if h.Unit, err = id(u.Type, u.Side); err != nil {
			return nil, err
		}
	}
	entrances := map[int]hexgrid.Coord{}
	for _, b := range l.Buildings {
		entrances[b.Container] = finmap.Coord(b.Cell[0], b.Cell[1])
	}
	for _, r := range l.Containers {
		if r.Holder != "building" {
			continue // a transporter's cargo, or a record for nothing
		}
		c, ok := entrances[r.Container]
		if !ok {
			return nil, fmt.Errorf("origview: map %s: record for building %d, which the map lacks", l.Source.Map, r.Container)
		}
		content := lv.Lines[c.Line][c.Col].Content
		for _, u := range r.Units {
			s, err := id(u.Type, r.Side)
			if err != nil {
				return nil, err
			}
			content.Units = append(content.Units, s)
		}
	}
	return lv, nil
}

// campaigns are the file suffixes of the three runs of 24 maps: the Central
// Powers campaign (maps 00-23), the Entente's (24-47) and the two-player
// maps (48-71).
var campaigns = [3]string{"cp", "ec", "2p"}

// FileName is the file stem of an imported map in levels/campaign: the
// map's campaign and its mission number there, "level4_cp" for map 03 and
// "level1_ec" for map 24, named like the shipped levels.
func FileName(l *finmap.Level) (string, error) {
	n, err := strconv.Atoi(l.Source.Map)
	if err != nil || n < 0 || n >= 24*len(campaigns) || l.Mission == nil {
		return "", fmt.Errorf("origview: map %q is no campaign mission", l.Source.Map)
	}
	k := n / 24
	sides := l.Mission.HumanSides
	if k < 2 && (len(sides) != 1 || sides[0] != k) || k == 2 && len(sides) != 2 {
		return "", fmt.Errorf("origview: map %s with human sides %v", l.Source.Map, sides)
	}
	return fmt.Sprintf("level%d_%s", n%24+1, campaigns[k]), nil
}
