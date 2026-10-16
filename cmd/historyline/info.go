package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
)

// The info box is a text placeholder for the terrain and unit panels (2.5): it
// shows what the player knows of the selected unit and the hovered hex, and
// nothing of hexes the player has not seen.

// value prints a terrain modifier; nil is Python's None.
func value(v *int) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprint(*v)
}

// terrainInfo describes hex h as the player knows it: nothing when Unseen,
// otherwise its terrain, infrastructure, modifiers and the known owner.
func terrainInfo(h *state.Hex, f state.Fog, owner state.Player) []string {
	if f == state.Unseen {
		return []string{"unexplored"}
	}
	out := []string{h.Terrain}
	if len(h.Infra) > 0 {
		out = append(out, strings.Join(h.Infra, " "))
	}
	out = append(out, fmt.Sprintf("move %s  battle %s  sight %s", value(h.Mods.Movement), value(h.Mods.Battle), value(h.Mods.Sight)))
	if h.Container != nil {
		out = append(out, fmt.Sprintf("owner %v", owner))
	}
	return out
}

// unitInfo describes a unit the player can see.
func unitInfo(u *state.Unit) []string {
	out := []string{
		fmt.Sprintf("%s (%v)", u.Def.Name, u.Owner),
		fmt.Sprintf("size %d/%d  rank %d  fuel %d", u.Size, u.MaxSize, u.Experience, u.Fuel),
	}
	switch {
	case u.Moved && u.Attacked:
		out = append(out, "moved, attacked")
	case u.Moved:
		out = append(out, "moved")
	case u.Attacked:
		out = append(out, "attacked")
	}
	for i, w := range u.Weapons {
		if w == nil {
			continue
		}
		ammo := "-"
		if a := u.Ammo[i]; a != nil {
			ammo = fmt.Sprint(*a)
		}
		out = append(out, fmt.Sprintf("%d %s r%d-%d p%d ammo %s", i+1, w.Name, w.RangeMin, w.RangeMax, w.Power, ammo))
	}
	return out
}

// memoryInfo describes the player's last sight of a unit on a Seen hex.
func memoryInfo(m *state.UnitMemory) []string {
	return []string{fmt.Sprintf("%s (%v), last seen", m.Type, m.Owner), fmt.Sprintf("size %d", m.Size)}
}

// infoLines is the info box text: the selected unit, then the hovered hex and
// the unit the player sees on it.
func (g *game) infoLines() []string {
	var out []string
	if g.ctl.selected && g.ctl.unit != nil {
		out = append(out, "selected:")
		out = append(out, unitInfo(g.ctl.unit)...)
	}
	if !g.hoverOnMap {
		return out
	}
	c := g.hover
	if len(out) > 0 {
		out = append(out, "")
	}
	out = append(out, fmt.Sprintf("hex %d,%d:", c.Line, c.Col))
	out = append(out, terrainInfo(g.state.At(c), g.fogOf(c), g.ownerShown(c))...)
	shown, remembered := g.unitShown(c)
	switch {
	case !shown:
	case remembered:
		if m, _ := g.state.Remembered(g.player, c); m.Unit != nil {
			out = append(out, memoryInfo(m.Unit)...)
		}
	case g.state.At(c).Unit != g.ctl.unit || !g.ctl.selected:
		out = append(out, unitInfo(g.state.At(c).Unit)...)
	}
	return out
}

// drawInfo draws the info box at the top right, below the dev indicators.
func (g *game) drawInfo(dst *ebiten.Image) {
	lines := g.infoLines()
	if len(lines) == 0 {
		return
	}
	w := 0
	for _, l := range lines {
		w = max(w, textWidth(l))
	}
	r := g.hud.infoRect(w, len(lines))
	vector.FillRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), color.RGBA{20, 20, 40, 220}, false)
	for i, l := range lines {
		ebitenutil.DebugPrintAt(dst, l, r.X+infoPad, r.Y+infoPad/2+i*infoLineHeight)
	}
}
