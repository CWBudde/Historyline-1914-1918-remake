package main

import (
	"fmt"
	"math/rand/v2"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// control turns hex clicks into rules commands for one player. It holds only the
// selection and its overlays; all game state stays in state.State, and every
// action goes through the command API so rejected actions come back as messages.
type control struct {
	st     *state.State
	player state.Player

	selected bool
	from     hexgrid.Coord
	unit     *state.Unit
	moves    map[hexgrid.Coord]state.Route
	targets  map[hexgrid.Coord]bool

	Message string // result of the last action, shown to the player
}

// roll is random.randint(-5, 5) of the Python reference.
func roll() int { return rand.IntN(11) - 5 }

func (c *control) clear() {
	c.selected, c.unit, c.moves, c.targets = false, nil, nil, nil
}

// mine reports whether it is the controlled player's turn.
func (c *control) mine() bool { return c.st.Turn == c.player && c.st.Winner == state.None }

// Select picks the own map unit at at and computes its overlays.
func (c *control) Select(at hexgrid.Coord) {
	c.clear()
	h := c.st.At(at)
	if h == nil || h.Unit == nil || h.Unit.Owner != c.player || c.st.Fog(c.player, at) != state.Visible {
		return
	}
	c.selected, c.from, c.unit = true, at, h.Unit
	c.moves = c.st.Movable(at, h.Unit)
	c.targets = map[hexgrid.Coord]bool{}
	if !(c.st.Rules.OneActionPerTurn && h.Unit.Moved) {
		for _, t := range c.st.Targets(at) {
			c.targets[t] = true
		}
	}
}

// Click handles a left click on a map hex: select, move, attack or deselect.
func (c *control) Click(at hexgrid.Coord) {
	c.Message = ""
	if !c.mine() {
		if c.st.Winner != state.None {
			c.Message = fmt.Sprintf("%v has won", c.st.Winner)
		}
		return
	}
	switch {
	case c.selected && c.targets[at]:
		c.attack(at)
	case c.selected && c.moves[at].Path != nil:
		c.move(at)
	default:
		c.Select(at)
	}
}

func (c *control) move(to hexgrid.Coord) {
	if _, err := c.st.Move(c.from, c.unit, to); err != nil {
		c.Message = err.Error()
		return
	}
	c.Message = fmt.Sprintf("%s moved", c.unit.Type)
	if c.st.Winner != state.None {
		c.Message = fmt.Sprintf("%v captured the HQ and wins", c.st.Winner)
	}
	c.Select(to)
}

// attack fires the first weapon the rules accept; the weapon menu comes later
// (PLAN.md 2.6).
func (c *control) attack(to hexgrid.Coord) {
	var err error
	for slot := range c.unit.Weapons {
		if c.unit.Weapons[slot] == nil {
			continue
		}
		var b state.Battle
		if b, err = c.st.Attack(c.from, to, slot, roll); err == nil {
			c.Message = fmt.Sprintf("hit for %d, took %d back", b.Damage, b.CounterDamage)
			switch {
			case b.DefenderDestroyed:
				c.Message += "; target destroyed"
			case b.AttackerDestroyed:
				c.Message += "; attacker destroyed"
			}
			c.clear()
			if !b.AttackerDestroyed {
				c.Select(c.from)
			}
			return
		}
	}
	c.Message = err.Error()
}

// EndTurn ends the player's turn. With no AI connected yet the opponent passes at
// once, so the game stays playable for a single human.
func (c *control) EndTurn() {
	c.clear()
	if !c.mine() {
		return
	}
	if err := c.st.EndTurn(); err != nil {
		c.Message = err.Error()
		return
	}
	for c.st.Turn != c.player && c.st.Winner == state.None {
		if err := c.st.EndTurn(); err != nil {
			c.Message = err.Error()
			return
		}
	}
	c.Message = fmt.Sprintf("round %d (the opponent passed: no AI yet)", c.st.Round)
}
