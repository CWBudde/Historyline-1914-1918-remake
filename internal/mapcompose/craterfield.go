package mapcompose

import (
	"image"
	"image/color"
	"math"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// A crater field's context tiles all show the same lattice-periodic material,
// so their hollows repeat once per hex at the contrast of a two-coset average.
// Draw time reworks them: inside the field the material's hollows are laid down
// again in world space, each patch shifted by its own offset and at more
// contrast, while the tile keeps its fade to the neighbouring ground.

const (
	// FieldPeriodX and FieldPeriodY are the world period of the reworked
	// ground, a whole number of lattice cells: tiles repeat only at this period.
	FieldPeriodX, FieldPeriodY = 1136, 800

	fieldSiteX, fieldSiteY = 71.0, 66.666666 // grid of patch sites, world px
	fieldPatchRadius       = 62.0            // world px a patch reaches
	fieldHollowBlur        = 3               // world px between grit and hollows
	fieldToneBlur          = 28              // world px between hollows and tone
	fieldEdgeNear          = 8.0             // world px from other ground the tile stays as is
	fieldEdgeFar           = 24.0            // ...and from where the field is all reworked
	fieldHollowGain        = 1.7
	fieldGritGain          = 1.25
	fieldSalt              = 0x51c7f3a29e0b64d3
)

// FieldGround says how the background of a crater field's hex is reworked:
// Exemplar is the material's tile with the material all around, Edges are the
// sides (hexgrid.Around order) that face other ground.
type FieldGround struct {
	Exemplar string
	Edges    [6]bool
}

// hexVertices are a hex's corners about its centre, world px.
var hexVertices = [6][2]float64{{-23, -40}, {23, -40}, {48, 0}, {23, 40}, {-23, 40}, {-48, 0}}

// edgeDistance is the distance of point (x, y), about the hex centre, from the
// side of the hex in direction port.
func edgeDistance(port int, x, y float64) float64 {
	e := (port + 1) % 6
	a, b := hexVertices[e], hexVertices[(e+1)%6]
	dx, dy := b[0]-a[0], b[1]-a[1]
	t := math.Max(0, math.Min(1, ((x-a[0])*dx+(y-a[1])*dy)/(dx*dx+dy*dy)))
	return math.Hypot(x-a[0]-t*dx, y-a[1]-t*dy)
}

func smoothstep(v float64) float64 {
	v = math.Max(0, math.Min(1, v))
	return v * v * (3 - 2*v)
}

// FieldMaterial is a crater material split into tone, hollows and grit at one
// density, ready to be laid down in world space. It is immutable.
type FieldMaterial struct {
	density  int
	w, h     int // one rectangular lattice period, device px
	mean     [3]float64
	hollows  []float32 // three per pixel
	grit     []float32
	sites    []fieldSite
	siteCols int
	siteRows int
}

type fieldSite struct {
	x, y   float64 // world px on the field's period
	dx, dy int     // the material's shift, device px
}

// NewFieldMaterial reads the material from its exemplar, a tile of the material
// on all sides with its anchor on the hex centre.
func NewFieldMaterial(exemplar image.Image, anchor [2]float64, density int) *FieldMaterial {
	m := &FieldMaterial{density: density, w: hexgrid.ColPitch * density, h: 2 * hexgrid.LinePitch * density}
	cell := m.latticeCell(exemplar, anchor)
	tone := wrapBlur(cell, m.w, m.h, fieldToneBlur*density)
	fine := wrapBlur(cell, m.w, m.h, fieldHollowBlur*density)
	m.hollows = make([]float32, len(cell))
	m.grit = make([]float32, len(cell))
	for i := range cell {
		m.hollows[i] = fine[i] - tone[i]
		m.grit[i] = cell[i] - fine[i]
		m.mean[i%3] += float64(cell[i])
	}
	for i := range m.mean {
		m.mean[i] /= float64(m.w * m.h)
	}
	m.siteCols = int(math.Round(FieldPeriodX / fieldSiteX))
	m.siteRows = int(math.Round(FieldPeriodY / fieldSiteY))
	m.sites = make([]fieldSite, m.siteCols*m.siteRows)
	for j := range m.siteRows {
		for i := range m.siteCols {
			v := mix(coordinateHash(j, i) ^ fieldSalt)
			x := (float64(i) + .15 + .7*unit(v)) * FieldPeriodX / float64(m.siteCols)
			v = mix(v)
			y := (float64(j) + .15 + .7*unit(v)) * FieldPeriodY / float64(m.siteRows)
			v = mix(v)
			dx := int(unit(v)*float64(hexgrid.ColPitch)) * density
			v = mix(v)
			dy := int(unit(v)*float64(2*hexgrid.LinePitch)) * density
			m.sites[j*m.siteCols+i] = fieldSite{x, y, dx, dy}
		}
	}
	return m
}

// latticeCell unrolls a hex tile of a lattice-periodic material to one
// rectangular period, three floats per pixel.
func (m *FieldMaterial) latticeCell(tile image.Image, anchor [2]float64) []float32 {
	b := tile.Bounds()
	d := m.density
	cx, cy := int(math.Round(anchor[0]*float64(d))), int(math.Round(anchor[1]*float64(d)))
	out := make([]float32, 3*m.w*m.h)
	var shifts [][2]int
	for a := -2; a <= 2; a++ {
		for b := -2; b <= 2; b++ {
			shifts = append(shifts, [2]int{a*hexgrid.ColPitch + b*hexgrid.HalfPitch, b * hexgrid.LinePitch})
		}
	}
	for y := range m.h {
		for x := range m.w {
			var best color.NRGBA
			for _, s := range shifts {
				px, py := b.Min.X+cx+x-s[0]*d, b.Min.Y+cy+y-s[1]*d
				if !(image.Pt(px, py).In(b)) {
					continue
				}
				c := color.NRGBAModel.Convert(tile.At(px, py)).(color.NRGBA)
				if c.A > best.A {
					best = c
				}
			}
			i := 3 * (y*m.w + x)
			out[i], out[i+1], out[i+2] = float32(best.R), float32(best.G), float32(best.B)
		}
	}
	return out
}

// wrapBlur is a box blur of radius r, run twice, of three-channel v wrapped
// over w×h.
func wrapBlur(v []float32, w, h, r int) []float32 {
	out := v
	for range 2 {
		out = wrapBox(wrapBox(out, w, h, r, true), w, h, r, false)
	}
	return out
}

func wrapBox(v []float32, w, h, r int, horizontal bool) []float32 {
	out := make([]float32, len(v))
	lines, n, step, next := h, w, 3, 3*w
	if !horizontal {
		lines, n, step, next = w, h, 3*w, 3
	}
	norm := float32(2*r + 1)
	for l := range lines {
		for c := range 3 {
			base := l*next + c
			var sum float32
			for i := -r; i <= r; i++ {
				sum += v[base+((i%n+n)%n)*step]
			}
			for i := range n {
				out[base+i*step] = sum / norm
				sum += v[base+((i+r+1)%n)*step] - v[base+((i-r+n)%n)*step]
			}
		}
	}
	return out
}

// Rework returns background tile, drawn with its top-left corner at origin (map
// pixels at the material's density) on hex at, with the field laid over it.
// Its alpha stays the tile's.
func (m *FieldMaterial) Rework(tile image.Image, origin image.Point, at hexgrid.Coord, edges [6]bool) *image.NRGBA {
	b := tile.Bounds()
	d := float64(m.density)
	cx, cy := hexgrid.Center(at)
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	near := m.nearSites(float64(origin.X)/d, float64(origin.Y)/d, float64(b.Dx())/d, float64(b.Dy())/d)
	var exposed []int
	for port, on := range edges {
		if on {
			exposed = append(exposed, port)
		}
	}
	for y := range b.Dy() {
		wy := (float64(origin.Y) + float64(y) + .5) / d
		for x := range b.Dx() {
			c := color.NRGBAModel.Convert(tile.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			if c.A == 0 {
				continue
			}
			wx := (float64(origin.X) + float64(x) + .5) / d
			weight := 1.
			for _, port := range exposed {
				weight = math.Min(weight, smoothstep((edgeDistance(port, wx-cx, wy-cy)-fieldEdgeNear)/(fieldEdgeFar-fieldEdgeNear)))
			}
			if weight > 0 {
				s := m.at(wx, wy, near)
				c.R = lerpByte(c.R, s[0], weight)
				c.G = lerpByte(c.G, s[1], weight)
				c.B = lerpByte(c.B, s[2], weight)
			}
			out.SetNRGBA(x, y, c)
		}
	}
	return out
}

func lerpByte(a uint8, b, t float64) uint8 {
	return uint8(math.Max(0, math.Min(255, math.Round(float64(a)*(1-t)+b*t))))
}

// nearSites lists the patches reaching the world rectangle, unrolled from the
// field's period to its position.
func (m *FieldMaterial) nearSites(x, y, w, h float64) []fieldSite {
	cw, ch := FieldPeriodX/float64(m.siteCols), FieldPeriodY/float64(m.siteRows)
	var out []fieldSite
	for j := int(math.Floor((y-fieldPatchRadius)/ch)) - 1; float64(j)*ch <= y+h+fieldPatchRadius; j++ {
		for i := int(math.Floor((x-fieldPatchRadius)/cw)) - 1; float64(i)*cw <= x+w+fieldPatchRadius; i++ {
			mi, mj := ((i%m.siteCols)+m.siteCols)%m.siteCols, ((j%m.siteRows)+m.siteRows)%m.siteRows
			s := m.sites[mj*m.siteCols+mi]
			s.x += float64(i-mi) * cw
			s.y += float64(j-mj) * ch
			if s.x < x-fieldPatchRadius || s.x > x+w+fieldPatchRadius || s.y < y-fieldPatchRadius || s.y > y+h+fieldPatchRadius {
				continue
			}
			out = append(out, s)
		}
	}
	return out
}

// at is the field's colour at a world point: the weighted patches, their sum
// scaled so that overlapping keeps the spread of one.
func (m *FieldMaterial) at(wx, wy float64, near []fieldSite) [3]float64 {
	px, py := int(math.Floor(wx*float64(m.density))), int(math.Floor(wy*float64(m.density)))
	var hollow, grit [3]float64
	var sum, sum2 float64
	for _, s := range near {
		dx, dy := (wx-s.x)/fieldPatchRadius, (wy-s.y)/fieldPatchRadius
		r := dx*dx + dy*dy
		if r >= 1 {
			continue
		}
		g := (1 - r) * (1 - r)
		sum += g
		sum2 += g * g
		i := 3 * (((py+s.dy)%m.h+m.h)%m.h*m.w + ((px+s.dx)%m.w+m.w)%m.w)
		for c := range 3 {
			hollow[c] += g * float64(m.hollows[i+c])
			grit[c] += g * float64(m.grit[i+c])
		}
	}
	tone := fieldTone(wx, wy)
	var out [3]float64
	norm := 0.
	if sum2 > 0 {
		norm = 1 / math.Sqrt(sum2)
	}
	for c := range 3 {
		out[c] = m.mean[c]*tone + norm*(fieldHollowGain*hollow[c]+fieldGritGain*grit[c])
	}
	return out
}

// fieldTone is the slow change of the ground's tone, on the field's period.
func fieldTone(x, y float64) float64 {
	u, v := 2*math.Pi*x/FieldPeriodX, 2*math.Pi*y/FieldPeriodY
	return 1 + .035*math.Sin(2*u+1.3)*math.Cos(v+.4) + .03*math.Sin(3*v+2.1+u) + .025*math.Cos(5*u-2*v+.7)
}

// fieldExemplar is the tile of terrain with terrain all around.
func fieldExemplar(terrain string) string {
	d, ok := Descriptor(Ground, terrain, [6]string{terrain, terrain, terrain, terrain, terrain, terrain}, [2]int{})
	if !ok {
		return ""
	}
	return ContextID(Ground, Key(d))
}

// registerField marks the hexes of a crater cluster for their field rework.
func (s *Scene) registerField(cl *Cluster, terrain string, has func(string) bool) {
	exemplar := fieldExemplar(terrain)
	if exemplar == "" || !has(exemplar) {
		return
	}
	in := map[hexgrid.Coord]bool{}
	for _, c := range cl.Cells {
		in[c] = true
	}
	for _, c := range cl.Cells {
		f := &FieldGround{Exemplar: exemplar}
		for port, n := range hexgrid.Around(c) {
			f.Edges[port] = s.Grid.Contains(n) && !in[n]
		}
		s.Cells[c.Line][c.Col].Field = f
	}
}

// FieldAt is the rework of background bg of hex c, nil when it has none: a
// stamp's under-tile is not the field.
func (s *Scene) FieldAt(c hexgrid.Coord, bg string) *FieldGround {
	cell := &s.Cells[c.Line][c.Col]
	if cell.Field == nil || bg != cell.Background {
		return nil
	}
	return cell.Field
}

// A winter crater stamp's under-tile is plain snow, whiter than the rough ground
// beside it, whose context tiles run up to the shared side and show the shelled
// ground they expect there. Draw time lays pure rough ground along both sides of
// that border, reaching into the snow, so the hex no longer shows as a notch.

const (
	fringeReach      = 26.0 // world px rough ground reaches into the under-tile
	fringeWander     = 8.0  // world px its limit wanders
	fringeWavelength = 37.0 // world px of that wander
	fringeChannel    = 23
)

// StampFringe says how a background beside the border between a winter crater
// stamp and rough ground is reworked: Exemplar is rough ground's tile with rough
// ground all around, Border the sides (hexgrid.Around order) on that border and
// Other those that face ground that is neither.
type StampFringe struct {
	Exemplar string
	Border   [6]bool
	Other    [6]bool
}

// registerFringes marks the stamped winter craters that touch rough ground and
// the rough hexes that touch them.
func (s *Scene) registerFringes(at func(hexgrid.Coord) Hex, has func(string) bool) {
	exemplar := fieldExemplar("rough_w")
	if exemplar == "" || !has(exemplar) {
		return
	}
	stamped := map[hexgrid.Coord]bool{}
	for _, cl := range s.Clusters {
		if cl.Kind == "crater_w" && cl.Stamp != "" {
			for _, c := range cl.Cells {
				stamped[c] = true
			}
		}
	}
	rough := func(c hexgrid.Coord) bool { return s.Grid.Contains(c) && at(c).Terrain == "rough_w" }
	mark := func(c hexgrid.Coord, border func(hexgrid.Coord) bool, same func(hexgrid.Coord) bool) {
		f := StampFringe{Exemplar: exemplar}
		touches := false
		for port, n := range hexgrid.Around(c) {
			switch {
			case border(n):
				f.Border[port], touches = true, true
			case s.Grid.Contains(n) && !same(n):
				f.Other[port] = true
			}
		}
		if touches {
			s.Cells[c.Line][c.Col].Fringe = &f
		}
	}
	inStamp := func(c hexgrid.Coord) bool { return stamped[c] }
	for c := range stamped {
		mark(c, rough, inStamp)
	}
	for l := range s.Cells {
		for col := range s.Cells[l] {
			if c := (hexgrid.Coord{Line: l, Col: col}); rough(c) {
				mark(c, inStamp, rough)
			}
		}
	}
}

// FringeAt is the rework of background bg of hex c, nil when it has none: on a
// stamped hex only the under-tile has one, not the per-hex art under fog.
func (s *Scene) FringeAt(c hexgrid.Coord, bg string) *StampFringe {
	cell := &s.Cells[c.Line][c.Col]
	if cell.Fringe == nil || (cell.Cluster != 0 && bg != cell.Under) {
		return nil
	}
	return cell.Fringe
}

// Rework returns background under, drawn at origin (map pixels at density) on hex
// at, with exemplar, a tile of the same size, reaching in from the border sides.
func (f *StampFringe) Rework(under, exemplar image.Image, origin image.Point, at hexgrid.Coord, density int) *image.NRGBA {
	b, r := under.Bounds(), exemplar.Bounds()
	d := float64(density)
	cx, cy := hexgrid.Center(at)
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := range b.Dy() {
		wy := (float64(origin.Y) + float64(y) + .5) / d
		for x := range b.Dx() {
			c := color.NRGBAModel.Convert(under.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			if c.A == 0 {
				continue
			}
			wx := (float64(origin.X) + float64(x) + .5) / d
			if t := f.weight(wx-cx, wy-cy, wx, wy); t > 0 {
				e := color.NRGBAModel.Convert(exemplar.At(r.Min.X+x, r.Min.Y+y)).(color.NRGBA)
				if e.A == 255 {
					c.R = lerpByte(c.R, float64(e.R), t)
					c.G = lerpByte(c.G, float64(e.G), t)
					c.B = lerpByte(c.B, float64(e.B), t)
				}
			}
			out.SetNRGBA(x, y, c)
		}
	}
	return out
}

// weight is how much pure rough ground shows at a point, (x, y) about the hex
// centre: all of it on a border side, none past a wandering reach nor on a side
// that faces other ground.
func (f *StampFringe) weight(x, y, wx, wy float64) float64 {
	reach := fringeReach + fringeWander*(2*valueNoise(wx/fringeWavelength, wy/fringeWavelength, fringeChannel)-1)
	t := 0.
	for port, on := range f.Border {
		if on {
			t = math.Max(t, 1-smoothstep(edgeDistance(port, x, y)/reach))
		}
	}
	for port, on := range f.Other {
		if on && t > 0 {
			t *= smoothstep(edgeDistance(port, x, y) / reach)
		}
	}
	return t
}

// FieldPhase is a tile origin (map pixels at density) on the field's period:
// tiles of one background with the same phase and edges rework alike.
func FieldPhase(origin image.Point, density int) image.Point {
	w, h := FieldPeriodX*density, FieldPeriodY*density
	return image.Pt(((origin.X%w)+w)%w, ((origin.Y%h)+h)%h)
}
