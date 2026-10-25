package ai

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/data"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/scenario"
)

func loadSet(t *testing.T) *data.Set {
	t.Helper()
	s, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// board builds a 6-wide grass map of the given lines played by the Central
// Powers, with the given hexes placed on it.
func board(t *testing.T, lines int, place map[hexgrid.Coord]level.Hex) *game.State {
	t.Helper()
	return wide(t, 6, lines, place)
}

// wide is board with the given width.
func wide(t *testing.T, width, lines int, place map[hexgrid.Coord]level.Hex) *game.State {
	t.Helper()
	lv := &level.Level{MapName: "test", MapNo: 1, Player: "Central Powers"}
	for l := range lines {
		line := make([]level.Hex, width-l%2)
		for c := range line {
			line[c] = level.Hex{Terrain: "grass"}
			if h, ok := place[hexgrid.Coord{Line: l, Col: c}]; ok {
				line[c] = h
			}
		}
		lv.Lines = append(lv.Lines, line)
	}
	s, err := game.New(lv, loadSet(t))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func at(l, c int) hexgrid.Coord { return hexgrid.Coord{Line: l, Col: c} }

func unit(id string) level.Hex { return level.Hex{Terrain: "grass", Unit: id} }

func building(terrain string, owner game.Player, units ...string) level.Hex {
	return level.Hex{Terrain: terrain, Content: &level.Content{Owner: int(owner), Units: units}}
}

func zero() int { return 0 }

// execute carries out o on s.
func execute(s *game.State, o Order, roll func() int) error {
	if o.Attack {
		_, err := s.Attack(o.From, o.To, o.Slot, roll)
		return err
	}
	_, err := s.Move(o.From, o.Unit, o.To)
	return err
}

// turn plays the side to move with p until it has nothing left to do and
// returns the orders it gave. It fails on a failing order or on more than
// 2 × units + 1 calls.
func turn(t *testing.T, s *game.State, p *Planner, roll func() int) []Order {
	t.Helper()
	limit := 2*len(s.Units(s.Turn)) + 1
	var out []Order
	for calls := 1; ; calls++ {
		if calls > limit {
			t.Fatalf("turn did not end within %d calls: %v", limit, out)
		}
		o, ok := p.Next(s)
		if !ok {
			return out
		}
		if err := execute(s, o, roll); err != nil {
			t.Fatalf("order %+v: %v", o, err)
		}
		out = append(out, o)
	}
}

func TestAttacksWeakEnemy(t *testing.T) {
	s := board(t, 12, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_infantry"), at(6, 2): unit("EC_infantry")})
	d := s.At(at(6, 2)).Unit
	d.Size = 2
	var p Planner
	o, ok := p.Next(s)
	if !ok || !o.Attack || o.From != at(4, 2) || o.To != at(6, 2) {
		t.Fatalf("order %+v %v, want an attack from (4,2) on (6,2)", o, ok)
	}
	a := s.At(o.From).Unit
	w := a.Weapons[o.Slot]
	if w == nil || !s.CanFire(a, o.Slot, d) || hexgrid.Distance(o.From, o.To) > w.RangeMax {
		t.Fatalf("slot %d is not usable", o.Slot)
	}
	b, err := s.Attack(o.From, o.To, o.Slot, zero)
	if err != nil || !b.DefenderDestroyed {
		t.Errorf("battle %+v, err %v, want the defender destroyed", b, err)
	}
	if o, ok := p.Next(s); ok {
		t.Errorf("after the attack: %+v, want nothing (one action per turn)", o)
	}
}

func TestTakesEnemyHQ(t *testing.T) {
	s := board(t, 12, map[hexgrid.Coord]level.Hex{
		at(4, 2): unit("CP_infantry"), at(6, 2): building("hqN", game.Entente), at(8, 4): building("cmpN", game.Entente),
	})
	var p Planner
	turn(t, s, &p, zero)
	if s.Winner != game.CentralPowers {
		t.Errorf("Winner = %v, want the AI side", s.Winner)
	}
}

func TestCapturesDepot(t *testing.T) {
	s := board(t, 12, map[hexgrid.Coord]level.Hex{at(4, 2): unit("CP_cavalry"), at(8, 2): building("factoryE_w", game.Entente)})
	var p Planner
	turn(t, s, &p, zero)
	if s.At(at(8, 2)).Owner != game.CentralPowers {
		t.Errorf("depot owner %v, want captured", s.At(at(8, 2)).Owner)
	}
}

// key is an order without the unit pointer, comparable across states.
type key struct {
	Attack   bool
	From, To hexgrid.Coord
	Type     string
	Slot     int
}

func keys(os []Order) []key {
	out := []key{}
	for _, o := range os {
		out = append(out, key{o.Attack, o.From, o.To, o.Unit.Type, o.Slot})
	}
	return out
}

func TestIgnoresUnseenEnemy(t *testing.T) {
	cases := []struct {
		name         string
		width, lines int
		place        map[hexgrid.Coord]level.Hex
		hidden       hexgrid.Coord
	}{
		{name: "known goals", width: 6, lines: 30, place: map[hexgrid.Coord]level.Hex{
			at(0, 0): unit("CP_infantry"), at(0, 3): unit("CP_cavalry"), at(2, 2): unit("CP_lArtillery"),
			at(4, 4): unit("EC_infantry"), at(12, 0): building("cmpN", game.Entente),
		}, hidden: at(29, 2)},
		{name: "exploring", width: 16, lines: 30, place: map[hexgrid.Coord]level.Hex{
			at(14, 0): unit("CP_infantry"), at(16, 0): unit("CP_cavalry"),
		}, hidden: at(15, 14)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			play := func(withHidden bool) []key {
				place := maps.Clone(c.place)
				if withHidden {
					place[c.hidden] = unit("EC_hArtillery")
				}
				s := wide(t, c.width, c.lines, place)
				var p Planner
				os := turn(t, s, &p, zero)
				if s.Fog(game.CentralPowers, c.hidden) != game.Unseen {
					t.Fatal("the hidden hex became visible; move it further away")
				}
				return keys(os)
			}
			without, with := play(false), play(true)
			if len(without) == 0 {
				t.Fatal("no orders at all")
			}
			if !reflect.DeepEqual(without, with) {
				t.Errorf("orders changed by an unseen enemy:\n without %v\n with    %v", without, with)
			}
		})
	}
}

func TestOneActionPerTurn(t *testing.T) {
	s := board(t, 16, map[hexgrid.Coord]level.Hex{
		at(0, 0): unit("CP_infantry"), at(0, 2): unit("CP_cavalry"), at(2, 4): unit("CP_lArtillery"),
		at(4, 2): unit("CP_infantry"), at(6, 2): unit("EC_infantry"), at(10, 3): unit("EC_cavalry"),
		at(14, 1): building("cmpN", game.Entente),
	})
	var p Planner
	os := turn(t, s, &p, zero)
	moved, attacked := map[*game.Unit]bool{}, map[*game.Unit]bool{}
	for _, o := range os {
		u := o.Unit
		if moved[u] || attacked[u] {
			t.Errorf("%s at %v acts twice: %+v", u.Type, o.From, o)
		}
		if o.Attack {
			attacked[u] = true
		} else {
			moved[u] = true
		}
	}
	if len(attacked) == 0 || len(moved) == 0 {
		t.Errorf("orders %v: want both moves and attacks", keys(os))
	}
}

func TestPlannerGivesOneActionPerUnit(t *testing.T) {
	// One action per unit per turn: the cavalry out of range moves, and is
	// not also given an attack in the same turn.
	s := board(t, 16, map[hexgrid.Coord]level.Hex{at(0, 0): unit("CP_cavalry"), at(6, 0): unit("EC_infantry")})
	s.At(at(6, 0)).Unit.Size = 4
	var p Planner
	os := turn(t, s, &p, zero)
	if len(os) != 1 || os[0].Attack {
		t.Errorf("orders %v: want a single move, no attack", keys(os))
	}
}

func TestPlaysShippedLevels(t *testing.T) {
	set := loadSet(t)
	for _, name := range []string{"monthly/1914-08-pulse", "monthly/1914-10-civil", "monthly/1914-12-mouse"} {
		for _, side := range []game.Player{game.CentralPowers, game.Entente} {
			t.Run(fmt.Sprintf("%s/%v", name, side), func(t *testing.T) {
				lv, err := scenario.ShippedLevel(name, scenario.CP)
				if err != nil {
					t.Fatal(err)
				}
				s, err := game.New(lv, set)
				if err != nil {
					t.Fatal(err)
				}
				n := 0
				roll := func() int { n++; return n%11 - 5 }
				var p Planner
				attacks, captures := 0, 0
				for s.Round <= 20 && s.Winner == game.None {
					if s.Turn == side {
						before := owned(s, side)
						for _, o := range turn(t, s, &p, roll) {
							if o.Attack {
								attacks++
							}
						}
						captures += owned(s, side) - before
						if s.Winner != game.None {
							break
						}
					}
					if err := s.EndTurn(); err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("rounds %d, attacks %d, captures %d, winner %v, own units left %d, enemy units left %d",
					s.Round, attacks, captures, s.Winner, len(s.Units(side)), len(s.Units(side.Opponent())))
				if attacks+captures == 0 {
					t.Error("no attack and no capture in the whole run")
				}
			})
		}
	}
}

// owned counts the buildings p owns.
func owned(s *game.State, p game.Player) int {
	n := 0
	for l := range s.Hexes {
		for c := range s.Hexes[l] {
			if h := &s.Hexes[l][c]; h.Container != nil && h.Owner == p {
				n++
			}
		}
	}
	return n
}

// The planner plays both sides of every shipped scenario setup (the
// monthly levels' three setups and the skirmish maps' two-player one: the
// original maps with aircraft, ships and the whole roster) for some rounds
// under both rule sets without an illegal order; with -short every sixth.
func TestPlaysCampaignLevels(t *testing.T) {
	set := loadSet(t)
	all, err := scenario.Shipped()
	if err != nil {
		t.Fatal(err)
	}
	var levels []scenario.Named
	for _, n := range all {
		if !strings.HasPrefix(n.Name, "skirmish/") || strings.HasSuffix(n.Name, ":2p") {
			levels = append(levels, n)
		}
	}
	if len(levels) < 99 {
		t.Fatalf("%d scenario setups", len(levels))
	}
	for i, n := range levels {
		if testing.Short() && i%6 != 0 {
			continue
		}
		s, err := game.New(n.Level, set)
		if err != nil {
			t.Fatalf("%s: %v", n.Name, err)
		}
		k := 0
		roll := func() int { k++; return k%11 - 5 }
		var p Planner
		for s.Round <= 6 && s.Winner == game.None {
			turn(t, s, &p, roll)
			if s.Winner != game.None {
				break
			}
			if err := s.EndTurn(); err != nil {
				t.Fatalf("%s: %v", n.Name, err)
			}
		}
	}
}
