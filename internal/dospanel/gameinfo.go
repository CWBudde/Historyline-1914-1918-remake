package dospanel

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Holdings are what one owner has: units, depots and factories.
type Holdings struct{ Units, Depots, Factories int }

// GameInfo is what the game info shows: the round, the level's name, the
// viewing side's mode (named), the best and the current score, the
// holdings of side one, side two and the whole map, each side's energy
// account and income, the orders given this phase and the move limit (0:
// none).
type GameInfo struct {
	Round       int
	Level, Mode string
	Best, Score int
	Sides       [2]Holdings
	Map         Holdings
	Energy      [2]int
	Income      [2]int
	Moves       int
	Limit       int
}

// GameInfoRects are the text block of the game info inside Panel in device
// pixels, with the line pitch and the scale it uses.
type GameInfoRects struct {
	Scale, Pitch float64
	Panel, Block image.Rectangle
}

// The game info in reference pixels: 23 lines of 6 pixels, 130 wide.
const (
	gameInfoLines = 23
	gameInfoPitch = 6
	gameInfoW     = 130
	gameInfoH     = gameInfoLines * gameInfoPitch
)

// GameInfoLayout lays out the game info in panel at scale.
func GameInfoLayout(panel image.Rectangle, scale float64) GameInfoRects {
	g := fit(panel, scale, gameInfoW, gameInfoH)
	return GameInfoRects{Scale: g.k, Pitch: gameInfoPitch * g.k, Panel: panel, Block: g.rect(0, 0, gameInfoW, gameInfoH)}
}

// Labels of the game info.
const (
	labelRound    = "ROUND"
	labelLevel    = "LEVEL"
	labelMode     = "MODE"
	labelBest     = "BEST"
	labelCurrent  = "CURRENT"
	labelOne      = "ONE"
	labelTwo      = "TWO"
	labelMap      = "MAP"
	labelUnits    = "UNITS"
	labelDepots   = "DEPOTS"
	labelFactory  = "FACTORY UNITS"
	labelResTotal = "RESOURCES TOTAL"
	labelResInc   = "RESOURCES INCOME"
	labelMoves    = "MOVES"
	labelLimit    = "LIMIT"
	labelNoLimit  = "NO LIMIT"
)

// shade darkens the map under the game info.
var shade = color.RGBA{0, 0, 0, 0x99}

// DrawGameInfo draws the game info of gi over panel at scale: the map
// darkens and the text stands on it.
func DrawGameInfo(dst *ebiten.Image, panel image.Rectangle, scale float64, gi GameInfo, t Text) {
	c := clip(dst, panel)
	vector.FillRect(c, float32(panel.Min.X), float32(panel.Min.Y), float32(panel.Dx()), float32(panel.Dy()), shade, false)

	g := fit(panel, scale, gameInfoW, gameInfoH)
	w := writer{c, t}
	line := func(n int) int { return g.y(gameInfoPitch/2 + gameInfoPitch*float64(n)) }
	num := strconv.Itoa

	for i, r := range [][2]string{
		{labelRound, num(gi.Round)},
		{labelLevel, gi.Level},
		{labelMode, gi.Mode},
		{labelBest, num(gi.Best)},
		{labelCurrent, num(gi.Score)},
	} {
		w.left(r[0], g.x(19), line(i), pale)
		w.left(r[1], g.x(73), line(i), pale)
	}

	cols := [3]float64{91, 109, 127} // right edges of the columns
	for i, h := range []string{labelOne, labelTwo, labelMap} {
		w.centre(h, g.x(cols[i]-6), line(9), pale)
	}
	for _, r := range []struct {
		n     int
		label string
		v     []int
	}{
		{11, labelUnits, []int{gi.Sides[0].Units, gi.Sides[1].Units, gi.Map.Units}},
		{13, labelDepots, []int{gi.Sides[0].Depots, gi.Sides[1].Depots, gi.Map.Depots}},
		{14, labelFactory, []int{gi.Sides[0].Factories, gi.Sides[1].Factories, gi.Map.Factories}},
		{16, labelResTotal, []int{gi.Energy[0], gi.Energy[1]}},
		{17, labelResInc, []int{gi.Income[0], gi.Income[1]}},
	} {
		w.left(r.label, g.x(2), line(r.n), pale)
		for i, v := range r.v {
			w.right(num(v), g.x(cols[i]), line(r.n), amber)
		}
	}

	limit := labelNoLimit
	if gi.Limit > 0 {
		limit = num(gi.Limit)
	}
	w.left(labelMoves, g.x(2), line(20), pale)
	w.left(num(gi.Moves), g.x(50), line(20), pale)
	w.left(labelLimit, g.x(2), line(22), pale)
	w.left(limit, g.x(50), line(22), pale)
}
