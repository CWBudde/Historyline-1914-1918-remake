package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosplay"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/gesture"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/pixfont"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
	"github.com/hajimehoshi/ebiten/v2"
)

// The info plates of -rules dos (internal/dospanel): the info gesture on an
// own unit opens the unit info of its type in the human side's panel, on a
// cell without a unit (the own HQ too) the game info; enemy units give no
// information. Fire, the cancel direction or Esc closes the plate; the other
// panel keeps running and the frame stays.

// refCell is a map cell's height in the reference's pixels.
const refCell = 24

// panelScale is the plates' scale, device pixels per reference pixel: the
// reference's proportion of a cell at the default zoom, so a plate fills the
// panel as in the original and keeps its size at every zoom.
func panelScale(sc dosscreen.Screen) float64 { return sc.FitScale * hexgrid.HexHeight / refCell }

// plateScale is the scale side's plate is drawn at: panelScale, shrunk by
// the layout fit until the plate stays inside the frame's teeth.
func (g *game) plateScale(side int, fit func(r image.Rectangle, scale float64) float64) float64 {
	r := g.strategicRect(side)
	return fit(image.Rect(r.X, r.Y, r.Right(), r.Bottom()), panelScale(g.split.screen))
}

// The fits of the three plates and the battle scene.
func unitInfoFit(r image.Rectangle, scale float64) float64 {
	return dospanel.UnitInfoLayout(r, scale).Scale
}

func gameInfoFit(r image.Rectangle, scale float64) float64 {
	return dospanel.GameInfoLayout(r, scale).Scale
}

func contentsFit(r image.Rectangle, scale float64) float64 {
	return dospanel.ContentsLayout(r, scale).Scale
}

// battleFit is the battle scene's: as large as fits, whatever the plates'
// scale.
func battleFit(r image.Rectangle, _ float64) float64 {
	return dospanel.BattleLayout(r, math.Inf(1)).Scale
}

// textScale is the pixel font's scale on a plate laid out at scale.
func textScale(scale float64) int { return max(1, int(math.Floor(scale+1e-9))) }

// panelText draws the plates' text in the pixel font at scale k: Small for
// the small size, Large for the large one.
type panelText struct{ k int }

var _ dospanel.Text = panelText{}

func (t panelText) face(size dospanel.Size) pixfont.Face {
	if size == dospanel.Large {
		return pixfont.Large
	}
	return pixfont.Small
}

func (t panelText) Measure(s string, size dospanel.Size) (w, h int) {
	return t.face(size).Measure(s, t.k)
}

func (t panelText) Draw(dst *ebiten.Image, s string, x, y int, size dospanel.Size, c color.Color) {
	t.face(size).Draw(dst, s, x, y, t.k, c)
}

// panelRect is side's panel as an image rectangle.
func (g *game) panelRect(side int) image.Rectangle {
	p := g.split.screen.Panels[side]
	return image.Rect(p.X, p.Y, p.Right(), p.Bottom())
}

// openInfo runs the info gesture on c: the unit info of an own unit, a
// refusal for any other unit, else the game info.
func (g *game) openInfo(c hexgrid.Coord) {
	s, d := g.split, g.dos
	if u, ok := d.units(g.reveal)[c]; ok {
		if u.Side != d.side {
			d.Message = "no information on enemy units"
			return
		}
		s.infoUnit = u
		g.openPanel(panelUnitInfo)
		return
	}
	g.openPanel(panelGameInfo)
}

// openPanel shows mode in the human side's panel; the well goes back to the
// cursor's line.
func (g *game) openPanel(mode panelMode) {
	g.split.mode[g.dos.side] = mode
	g.split.track.Reset()
	g.dos.Message = ""
}

// infoGesture handles the human side's input while an info plate is open:
// fire (a click or a run), the cancel direction, a right click or a lost
// drag closes it.
func (g *game) infoGesture(ev gesture.Event) {
	switch ev.Kind {
	case gesture.Click, gesture.Run, gesture.Cancelled, gesture.RingOpen:
		g.split.mode[g.dos.side] = panelMap
		g.split.track.Reset()
	default:
	}
}

// unitInfo is the unit info of u's type as its side sees it, with the
// portrait and the sprite when the art is loaded.
func (g *game) unitInfo(u dosplay.UnitView) dospanel.UnitInfo {
	tv, ok := g.dos.s.TypeInfo(u.Side&1, u.Type)
	if !ok {
		return dospanel.UnitInfo{Name: u.Name}
	}
	return dospanel.UnitInfo{
		Name: tv.Name, Air: dospanel.Weapon(tv.Air), Ground: dospanel.Weapon(tv.Ground), Water: dospanel.Weapon(tv.Water),
		Armour: tv.Armour, Speed: tv.Speed, Weight: tv.Weight, Group: tv.Group,
		Portrait: g.portrait(u.ID), Sprite: g.panelUnit(u, unitsprite.N),
	}
}

// portrait is the screen catalog's portrait of unit type id, nil without
// one.
func (g *game) portrait(id string) *ebiten.Image {
	family, ok := unitsprite.Family(id)
	if !ok {
		return nil
	}
	return g.screens.image("portrait/"+family, 2)
}

// panelUnit is the sprite a panel shows of u facing f: the original's while
// F8 draws the original art, else the remake's map sprite at 2x.
func (g *game) panelUnit(u dosplay.UnitView, f unitsprite.Facing) *ebiten.Image {
	if img := g.origPanelUnit(u, f); img != nil {
		return img
	}
	return g.unitFacing(u.ID, u.Owner, f)
}

// unitFacing is that sprite facing f.
func (g *game) unitFacing(id string, owner state.Player, f unitsprite.Facing) *ebiten.Image {
	if g.art == nil {
		return nil
	}
	family, sid, ok := unitSprite(id, 0, owner, false, false, f)
	if !ok {
		return nil
	}
	img, _, err := g.art.image(family, sid, 2)
	if err != nil {
		return nil
	}
	return img
}

// levelKey names a level in the game info and the best scores: its map name
// in capitals (the shipped levels' code words; an original map's number).
func levelKey(l *level.Level) string {
	if l.MapName == "" {
		return fmt.Sprintf("%02d", l.MapNo)
	}
	return strings.ToUpper(l.MapName)
}

// gameInfo is the game info: the session's numbers, the human side's mode
// and the level's best score.
func (g *game) gameInfo() dospanel.GameInfo {
	d := g.dos
	st := d.s.Stats()
	key := levelKey(d.s.Level)
	gi := dospanel.GameInfo{
		Round: st.Round, Level: key, Mode: modeWord(st.Mode[d.side]), Best: g.scores.Best(key), Score: st.Score,
		Map: dospanel.Holdings(st.Map), Energy: st.Energy, Income: st.Income, Moves: st.Orders[d.side], Limit: st.Limit,
	}
	for k := range 2 {
		gi.Sides[k] = dospanel.Holdings(st.Sides[k])
	}
	return gi
}

// drawOverlay draws side's plate in panel r: the unit info or the contents
// over the map, or the game info on the darkened map.
func (g *game) drawOverlay(dst *ebiten.Image, side int, r image.Rectangle) {
	switch g.split.mode[side] {
	case panelUnitInfo:
		k := g.plateScale(side, unitInfoFit)
		dospanel.DrawUnitInfo(dst, r, k, g.unitInfo(g.split.infoUnit), panelText{textScale(k)})
	case panelGameInfo:
		k := g.plateScale(side, gameInfoFit)
		dospanel.DrawGameInfo(dst, r, k, g.gameInfo(), panelText{textScale(k)})
	case panelContents:
		g.drawContents(dst, r)
	default:
	}
}
