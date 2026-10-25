package minimap

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/camera"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// StrategicCell is the strategic map's base block edge, the original's 4×4 cell;
// FrameWidth is the recessed frame around the map.
const (
	StrategicCell = 4
	FrameWidth    = 4
)

// Strategic-map colours: the olive plate, the frame from outside in, the view
// rectangle and the depot and factory plates.
var (
	plateColor  = color.RGBA{48, 48, 36, 255}
	frameOuter  = color.RGBA{81, 81, 65, 255}
	frameMiddle = color.RGBA{36, 36, 28, 255}
	frameInner  = color.RGBA{28, 28, 24, 255}
	viewLine    = color.RGBA{154, 154, 138, 255}
	viewShade   = frameMiddle
	viewDrop    = frameOuter
	depotLit    = color.RGBA{240, 236, 214, 255}
	depotBody   = color.RGBA{214, 210, 188, 255}
	depotEdge   = color.RGBA{92, 90, 76, 255}
	depotMark   = frameMiddle
)

// Kind is the kind of a building plate.
type Kind int

const (
	HQ      Kind = iota // a shield in the side's shade with "HQ"
	Depot               // a light disc with a dark "D", edged in the side's dark tone
	Factory             // a light square with a dark "F", edged like the depot
)

// Building is an HQ, depot or factory plate on the strategic map; a zero Shade is unowned.
type Building struct {
	At    hexgrid.Coord
	Kind  Kind
	Shade Shade
}

// Scene is what the strategic map shows: a terrain colour per hex (transparent
// leaves the hex dark), unit dots, building plates and the panel's camera (nil for
// no view rectangle). The caller decides what the side may see.
type Scene struct {
	Terrain   func(hexgrid.Coord) color.RGBA
	Units     []Unit
	Buildings []Building
	View      *camera.Camera
}

// Strategic is the strategic map laid out in a panel of Size pixels: the panel is
// an olive plate, the map sits centred in it at Scale times the StrategicCell block,
// the largest whole factor that fits inside the frame, at Map in panel pixels.
type Strategic struct {
	Size   image.Point
	Raster Raster
	Map    image.Rectangle
	Scale  int
}

// NewStrategic lays out g in a panel of the given size. A panel too small for
// Scale 1 still gets it, centred and clipped.
func NewStrategic(g hexgrid.Grid, panel image.Point) Strategic {
	base := Cells(g, StrategicCell, StrategicCell).Size()
	k := 1
	if base.X > 0 && base.Y > 0 {
		k = max(1, min((panel.X-2*FrameWidth)/base.X, (panel.Y-2*FrameWidth)/base.Y))
	}
	r := Cells(g, k*StrategicCell, k*StrategicCell)
	s := r.Size()
	at := panel.Sub(s).Div(2)
	return Strategic{Size: panel, Raster: r, Map: image.Rectangle{Min: at, Max: at.Add(s)}, Scale: k}
}

// Render returns the panel as an image of Size.
func (s Strategic) Render(sc Scene) *image.RGBA {
	img := image.NewRGBA(image.Rectangle{Max: s.Size})
	s.Draw(img, image.Point{}, sc)
	return img
}

// Draw paints the panel into dst with its top-left corner at at.
func (s Strategic) Draw(dst *image.RGBA, at image.Point, sc Scene) {
	sub, ok := dst.SubImage(image.Rectangle{Min: at, Max: at.Add(s.Size)}).(*image.RGBA)
	if !ok || sub.Rect.Empty() {
		return
	}
	m := s.Map.Add(at)
	fillRect(sub, sub.Rect, plateColor)
	strokeRect(sub, m.Inset(-4), frameOuter)
	strokeRect(sub, m.Inset(-3), frameMiddle)
	strokeRect(sub, m.Inset(-2), frameMiddle)
	strokeRect(sub, m.Inset(-1), frameInner)
	fillRect(sub, m, frameInner)
	if sc.Terrain != nil {
		s.Raster.DrawCells(sub, m.Min, sc.Terrain)
	}
	for _, u := range sc.Units {
		s.Raster.DrawUnit(sub, m.Min, u)
	}
	for _, b := range sc.Buildings {
		blk := s.Raster.Block(b.At).Add(m.Min)
		centre := blk.Min.Add(image.Pt(blk.Dx()/2, blk.Dy()/2))
		s.drawPlate(sub, centre, b)
	}
	if sc.View != nil {
		drawViewRect(sub, s.ViewRect(sc.View).Add(at))
	}
}

// ToWorld converts a panel position to the world position under it; ok is false
// off the map (plate and frame). A click centres the panel's camera there.
func (s Strategic) ToWorld(px, py float64) (wx, wy float64, ok bool) {
	m := s.Map
	if px < float64(m.Min.X) || py < float64(m.Min.Y) || px >= float64(m.Max.X) || py >= float64(m.Max.Y) {
		return 0, 0, false
	}
	wx, wy = s.Raster.ToWorld(px-float64(m.Min.X), py-float64(m.Min.Y))
	return wx, wy, true
}

// ViewRect returns the camera's visible area in panel pixels, clipped to the map.
func (s Strategic) ViewRect(c *camera.Camera) image.Rectangle {
	w, h := s.Raster.Grid.Bounds()
	vx, vy, vw, vh := c.Visible()
	x0, y0 := s.Raster.FromWorld(max(vx, rasterInset), max(vy, 0))
	x1, y1 := s.Raster.FromWorld(min(vx+vw, w-rasterInset), min(vy+vh, h))
	r := image.Rect(int(math.Round(x0)), int(math.Round(y0)), int(math.Round(x1)), int(math.Round(y1)))
	return r.Add(s.Map.Min).Intersect(s.Map)
}

// drawViewRect draws the view rectangle: a light outline, a dark inner shadow
// along the top and left, a drop shadow below and to the right.
func drawViewRect(dst *image.RGBA, r image.Rectangle) {
	if r.Empty() {
		return
	}
	fillRect(dst, image.Rect(r.Max.X, r.Min.Y+1, r.Max.X+1, r.Max.Y+1), viewDrop)
	fillRect(dst, image.Rect(r.Min.X+1, r.Max.Y, r.Max.X+1, r.Max.Y+1), viewDrop)
	strokeRect(dst, r, viewLine)
	fillRect(dst, image.Rect(r.Min.X+1, r.Min.Y+1, r.Min.X+2, r.Max.Y-1), viewShade)
	fillRect(dst, image.Rect(r.Min.X+1, r.Min.Y+1, r.Max.X-1, r.Min.Y+2), viewShade)
}

// strokeRect draws the one-pixel border inside r.
func strokeRect(dst *image.RGBA, r image.Rectangle, c color.RGBA) {
	fillRect(dst, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+1), c)
	fillRect(dst, image.Rect(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y), c)
	fillRect(dst, image.Rect(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), c)
	fillRect(dst, image.Rect(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y), c)
}

// Plate shapes at scale 1, drawn at Scale × Scale pixels per shape pixel: an HQ
// shield 11×12 (straight sides, a pointed foot), a depot disc 9×9 and a factory
// square 9×9, each with a 3×5 mark.
const (
	shieldW, shieldH = 11, 12
	discD            = 9
)

var (
	glyphH = [5]string{"#.#", "#.#", "###", "#.#", "#.#"}
	glyphQ = [5]string{".#.", "#.#", "#.#", "##.", ".##"}
	glyphD = [5]string{"##.", "#.#", "#.#", "#.#", "##."}
	glyphF = [5]string{"###", "#..", "##.", "#..", "#.."}
)

func inShield(x, y int) bool {
	if x < 0 || y < 0 || y >= shieldH {
		return false
	}
	d := max(0, y-6)
	return x >= d && x <= shieldW-1-d
}

func inDisc(x, y int) bool {
	dx, dy := x-discD/2, y-discD/2
	return dx*dx+dy*dy <= 20
}

func inSquare(x, y int) bool {
	return x >= 0 && y >= 0 && x < discD && y < discD
}

// Plate pixel tones.
const (
	toneNone = iota
	toneEdge
	toneLit
	toneMid
	toneMark
)

// plateTones returns the tone of every pixel of a w×h shape: the edge where a
// neighbour is outside, lit along the inner top and left, the marks at their
// positions, mid elsewhere.
func plateTones(w, h int, inside func(x, y int) bool, marks map[image.Point][5]string) [][]uint8 {
	interior := func(x, y int) bool {
		return inside(x, y) && inside(x-1, y) && inside(x+1, y) && inside(x, y-1) && inside(x, y+1)
	}
	t := make([][]uint8, h)
	for y := range t {
		t[y] = make([]uint8, w)
		for x := range t[y] {
			switch {
			case !inside(x, y):
			case !interior(x, y):
				t[y][x] = toneEdge
			case !interior(x-1, y) || !interior(x, y-1):
				t[y][x] = toneLit
			default:
				t[y][x] = toneMid
			}
		}
	}
	for at, g := range marks {
		for gy, row := range g {
			for gx, ch := range row {
				if ch == '#' {
					t[at.Y+gy][at.X+gx] = toneMark
				}
			}
		}
	}
	return t
}

var (
	shieldTones = plateTones(shieldW, shieldH, inShield, map[image.Point][5]string{{2, 2}: glyphH, {6, 2}: glyphQ})
	discTones   = plateTones(discD, discD, inDisc, map[image.Point][5]string{{3, 2}: glyphD})
	squareTones = plateTones(discD, discD, inSquare, map[image.Point][5]string{{3, 2}: glyphF})
)

// drawPlate draws building b centred on centre.
func (s Strategic) drawPlate(dst *image.RGBA, centre image.Point, b Building) {
	tones, pal := shieldTones, [5]color.RGBA{{}, b.Shade.Dark, b.Shade.Lit, b.Shade.Mid, b.Shade.Dark}
	if b.Shade == (Shade{}) {
		pal = [5]color.RGBA{{}, Neutral.Dark, Neutral.Lit, Neutral.Mid, Neutral.Dark}
	}
	if b.Kind != HQ {
		edge := depotEdge
		if b.Shade != (Shade{}) {
			edge = b.Shade.Dark
		}
		tones, pal = discTones, [5]color.RGBA{{}, edge, depotLit, depotBody, depotMark}
		if b.Kind == Factory {
			tones = squareTones
		}
	}
	k := s.Scale
	h, w := len(tones), len(tones[0])
	origin := centre.Sub(image.Pt(w*k/2, h*k/2))
	for y, row := range tones {
		for x, tone := range row {
			if tone != toneNone {
				p := origin.Add(image.Pt(x*k, y*k))
				fillRect(dst, image.Rect(p.X, p.Y, p.X+k, p.Y+k), pal[tone])
			}
		}
	}
}
