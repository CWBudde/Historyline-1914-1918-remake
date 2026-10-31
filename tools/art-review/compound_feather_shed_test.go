package artreview

import (
	"image"
	"image/color"
	"math"
	"path/filepath"
	"testing"
)

// Shed move: the depot's south-west shed and the factory's north-west hall
// crossed the outer hex boundary in their masters, so a keep polygon held them
// opaque where the feather cuts everything else. compoundShedEdits moves them
// inward in new master versions by whole master pixels and refills what they
// leave with the master's own ground; the generator
// (compound_shed_generate_test.go, -tags campaigngfx) writes the masters, the
// tests below pin them.

// compoundShedAxis maps a destination coordinate to its source coordinate
// along one axis: a shift by Off, a mirror about the line between Mirror-1
// and Mirror (then shifted by Off), or a triangle wave over [Lo,Hi) that
// starts at Hi-1 at the region's first row or column.
type compoundShedAxis struct{ Off, Mirror, Lo, Hi int }

func (a compoundShedAxis) at(v, start int) int {
	switch {
	case a.Hi > a.Lo:
		n := a.Hi - a.Lo
		t := ((v-start)%(2*n) + 2*n) % (2 * n)
		if t < n {
			return a.Hi - 1 - t
		}
		return a.Lo + t - n
	case a.Mirror != 0:
		return 2*a.Mirror - 1 - v + a.Off
	default:
		return v + a.Off
	}
}

// compoundShedFill refills Dst (master px) with the base master's own pixels
// mapped by X and Y; Soft (left, top, right, bottom) blends it over that many
// px outside Dst.
type compoundShedFill struct {
	Dst  image.Rectangle
	X, Y compoundShedAxis
	Soft [4]int
}

// compoundShedPiece moves Src (master px: an object with a ground margin, or
// the ellipse inscribed in it) by D; the outer Feather px blend into what lies
// below, the core inside is copied exactly.
type compoundShedPiece struct {
	Name    string
	Src     image.Rectangle
	D       image.Point
	Feather int
	Ellipse bool
}

// compoundShedWarp rebuilds Rect (master px) from the patched base column by
// column: Map gives the source row of a destination pixel (a fractional row
// blends the two rows around it), ok false leaves the pixel to the fills.
type compoundShedWarp struct {
	Name string
	Rect image.Rectangle
	Map  func(x, y int) (float64, bool)
}

// compoundShedEdit makes a new master from Base: Patches repaint the base
// (an object in front of a moved wall) for the pieces and the fills, Fills
// refill the ground the pieces leave in order (each reading the image as the
// fills before left it), then Pieces are pasted from the patched base and
// Warps rebuilt from it.
type compoundShedEdit struct {
	Base, Out string
	Patches   []compoundShedFill
	Fills     []compoundShedFill
	Pieces    []compoundShedPiece
	Warps     []compoundShedWarp
}

// compoundShedCut shortens a building standing in columns [x0, x1): the rows
// cuts (half-open, top to bottom, inside [top, bottom)) drop out and the rows
// north of them slide south, so the rows from bottom on stay where they are;
// the strip the building's top leaves is the fills'.
func compoundShedCut(name string, x0, x1, top, bottom int, cuts ...[2]int) compoundShedWarp {
	var kept []int
	for y := top; y < bottom; y++ {
		cut := false
		for _, c := range cuts {
			cut = cut || (y >= c[0] && y < c[1])
		}
		if !cut {
			kept = append(kept, y)
		}
	}
	newTop := bottom - len(kept)
	return compoundShedWarp{Name: name, Rect: image.Rect(x0, newTop, x1, bottom), Map: func(x, y int) (float64, bool) {
		if y < newTop || y >= bottom {
			return 0, false
		}
		return float64(kept[y-newTop]), true
	}}
}

// compoundShedSample reads column x of m at the fractional row y.
func compoundShedSample(m *image.NRGBA, x int, y float64) color.NRGBA {
	y0 := math.Floor(y)
	at := func(r int) color.NRGBA {
		return m.NRGBAAt(x, min(max(r, m.Rect.Min.Y), m.Rect.Max.Y-1))
	}
	if f := y - y0; f > 1e-9 {
		return compoundShedMix(at(int(y0)), at(int(y0)+1), f)
	}
	return at(int(y0))
}

func (f compoundShedFill) weight(x, y int) float64 {
	side := func(d, soft int) float64 {
		if d <= 0 {
			return 1
		}
		if soft == 0 {
			return 0
		}
		return 1 - connectionSmoothStep(float64(d)/float64(soft+1))
	}
	return side(f.Dst.Min.X-x, f.Soft[0]) * side(y-f.Dst.Max.Y+1, f.Soft[3]) * side(f.Dst.Min.Y-y, f.Soft[1]) * side(x-f.Dst.Max.X+1, f.Soft[2])
}

func (f compoundShedFill) outer() image.Rectangle {
	return image.Rect(f.Dst.Min.X-f.Soft[0], f.Dst.Min.Y-f.Soft[1], f.Dst.Max.X+f.Soft[2], f.Dst.Max.Y+f.Soft[3])
}

// weight of the destination pixel (x, y): 1 in the core, 0 outside the piece.
func (p compoundShedPiece) weight(x, y int) float64 {
	r := p.Src.Add(p.D)
	if !image.Pt(x, y).In(r) {
		return 0
	}
	// d: whole px between the pixel and the piece's outline.
	d := float64(min(x-r.Min.X, y-r.Min.Y, r.Max.X-1-x, r.Max.Y-1-y))
	if p.Ellipse {
		ax, ay := float64(r.Dx())/2, float64(r.Dy())/2
		u, v := (float64(x)+.5-float64(r.Min.X))/ax-1, (float64(y)+.5-float64(r.Min.Y))/ay-1
		d = (1-math.Hypot(u, v))*min(ax, ay) - .5
	}
	if d >= float64(p.Feather) {
		return 1
	}
	return connectionSmoothStep((d + 1) / float64(p.Feather+1))
}

// compoundShedMix blends b over a by w in premultiplied alpha.
func compoundShedMix(a, b color.NRGBA, w float64) color.NRGBA {
	if w >= 1 {
		return b
	}
	if w <= 0 {
		return a
	}
	aa, ba := float64(a.A)/255, float64(b.A)/255
	alpha := aa*(1-w) + ba*w
	if alpha <= 0 {
		return color.NRGBA{}
	}
	ch := func(ac, bc uint8) uint8 {
		return uint8(math.Floor(min(255, (float64(ac)*aa*(1-w)+float64(bc)*ba*w)/alpha) + .5))
	}
	return color.NRGBA{ch(a.R, b.R), ch(a.G, b.G), ch(a.B, b.B), uint8(math.Floor(alpha*255 + .5))}
}

func compoundShedApplyFill(dst, from *image.NRGBA, f compoundShedFill) {
	r := f.outer().Intersect(dst.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			w := f.weight(x, y)
			if w <= 0 {
				continue
			}
			s := image.Pt(f.X.at(x, f.Dst.Min.X), f.Y.at(y, f.Dst.Min.Y))
			if !s.In(from.Bounds()) {
				continue
			}
			dst.SetNRGBA(x, y, compoundShedMix(dst.NRGBAAt(x, y), from.NRGBAAt(s.X, s.Y), w))
		}
	}
}

func compoundShedApply(base *image.NRGBA, e compoundShedEdit) *image.NRGBA {
	clone := func(m *image.NRGBA) *image.NRGBA {
		c := image.NewNRGBA(m.Bounds())
		copy(c.Pix, m.Pix)
		return c
	}
	src := clone(base)
	for _, f := range e.Patches {
		compoundShedApplyFill(src, base, f)
	}
	out := clone(src)
	for _, f := range e.Fills {
		compoundShedApplyFill(out, clone(out), f)
	}
	for _, p := range e.Pieces {
		r := p.Src.Add(p.D).Intersect(out.Bounds())
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				out.SetNRGBA(x, y, compoundShedMix(out.NRGBAAt(x, y), src.NRGBAAt(x-p.D.X, y-p.D.Y), p.weight(x, y)))
			}
		}
	}
	for _, w := range e.Warps {
		r := w.Rect.Intersect(out.Bounds())
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				if sy, ok := w.Map(x, y); ok {
					out.SetNRGBA(x, y, compoundShedSample(src, x, sy))
				}
			}
		}
	}
	return out
}

// The depot (cmp) masters: the shed moves up 100 master px (15.6 world px) to
// stand on the south-west fence line, hiding a fence post, and 8 px east; the
// crates beside it and the crates with the barrels follow by 60 px up and 8 px
// east. The crate in front of its door would still reach the boundary: it is
// dropped and the wall behind it repainted from the wall beside it. The ground
// they leave is the ground strip below the fence (rows 674-697, mirrored over
// and over) in the west and the ground east of the barrels (mirrored) in the
// east; further south, outside the hexes, the master's own apron edge below.
var compoundShedDepot = compoundShedEdit{
	Patches: []compoundShedFill{{Dst: image.Rect(440, 819, 469, 864), X: compoundShedAxis{Off: -30}}},
	Fills: []compoundShedFill{
		{Dst: image.Rect(334, 716, 476, 870), X: compoundShedAxis{Off: -8}, Y: compoundShedAxis{Off: 100}, Soft: [4]int{6, 4, 0, 6}},
		{Dst: image.Rect(476, 716, 546, 870), X: compoundShedAxis{Mirror: 546}, Soft: [4]int{12, 4, 0, 6}},
		{Dst: image.Rect(334, 719, 524, 790), Y: compoundShedAxis{Lo: 674, Hi: 698}, Soft: [4]int{6, 3, 16, 12}},
	},
	Pieces: []compoundShedPiece{
		{Name: "shed", Src: image.Rect(360, 698, 488, 846), D: image.Pt(8, -100), Feather: 3},
		{Name: "west crates", Src: image.Rect(334, 728, 365, 787), D: image.Pt(8, -60), Feather: 3},
		{Name: "east crates", Src: image.Rect(488, 728, 535, 793), D: image.Pt(8, -60), Feather: 3},
		{Name: "rear barrel", Src: image.Rect(485, 784, 506, 806), D: image.Pt(8, -60), Feather: 2, Ellipse: true},
		{Name: "front barrel", Src: image.Rect(494, 795, 523, 825), D: image.Pt(8, -60), Feather: 2, Ellipse: true},
	},
}

// The factory masters: the north-west hall (with the crate on its ridge and the
// flange of the pipe that feeds it) moves down 40 master px (6.5 world px)
// over the yard between it and the hall south of it, whose props it covers;
// a strip of the hedge along that yard stays visible below it. The trees above
// it are mirrored down into the strip it leaves, the pipe runs on.
var compoundShedFactory = compoundShedEdit{
	Fills: []compoundShedFill{
		{Dst: image.Rect(517, 222, 550, 240), X: compoundShedAxis{Off: 40}, Soft: [4]int{3, 3, 3, 0}},
		{Dst: image.Rect(330, 240, 603, 284), Y: compoundShedAxis{Lo: 227, Hi: 240}, Soft: [4]int{6, 0, 0, 0}},
		{Dst: image.Rect(456, 250, 560, 284), X: compoundShedAxis{Off: 40}, Y: compoundShedAxis{Off: -40}, Soft: [4]int{12, 8, 0, 0}},
		{Dst: image.Rect(560, 250, 603, 284), Y: compoundShedAxis{Off: -40}, Soft: [4]int{8, 8, 0, 0}},
		{Dst: image.Rect(603, 226, 634, 284), Y: compoundShedAxis{Off: -50}},
		{Dst: image.Rect(634, 240, 638, 284), Y: compoundShedAxis{Mirror: 240}},
		{Dst: image.Rect(339, 405, 382, 419), X: compoundShedAxis{Off: 44}, Soft: [4]int{4, 0, 4, 0}},
		{Dst: image.Rect(430, 405, 474, 419), X: compoundShedAxis{Off: 44}, Soft: [4]int{4, 0, 4, 0}},
	},
	Pieces: []compoundShedPiece{
		{Name: "hall", Src: image.Rect(335, 240, 638, 373), D: image.Pt(0, 40), Feather: 3},
		{Name: "ridge crate", Src: image.Rect(519, 223, 548, 248), D: image.Pt(0, 40), Feather: 3},
		{Name: "pipe flange", Src: image.Rect(601, 230, 636, 248), D: image.Pt(0, 40), Feather: 3},
	},
}

// The depot's north-west shed (on cmp-summer-v3 / cmp-winter-v4) is shortened
// so its north eave moves 19 master px (3 world px) south, off the W cell's
// top edge. Over the gable the 19 corrugation rows below the eave drop out
// (the corrugation runs north-south, so the cut leaves no seam); over the hip
// the corrugation north of the hip line slides with the hip line, and the hip's
// east face is pulled affinely towards its fixed ridge end and south-east
// corner, so the hip line still runs straight from the ridge end to the new
// north-east corner. The ground strip north of the shed (rows 245-256,
// mirrored) fills what the eave leaves.
const (
	depotShedTop, depotShedSlide = 257, 19
	depotShedTrim                = 4          // eave rows sliding whole
	depotShedRidgeX              = 452        // ridge end: the hip starts
	depotShedRidgeY, depotShedX  = 295., 510  // ridge row, east eave column
	depotShedNE, depotShedSE     = 257., 360. // hip corners on the east eave
	depotShedLeft, depotShedEnd  = 372, 514   // the shed's columns with outline
	depotShedBottom              = 362        // rows from here on stay
)

func compoundShedDepotNorthMap(x, y int) (float64, bool) {
	top := depotShedTop + depotShedSlide
	if y < top || y >= depotShedBottom || x < depotShedLeft || x >= depotShedEnd {
		return 0, false
	}
	trim := y < top+depotShedTrim
	if x < depotShedRidgeX {
		if trim {
			return float64(y - depotShedSlide), true
		}
		return float64(y), true
	}
	// Hip lines (master rows) at column xc: north before and after, south.
	xc := float64(min(x, depotShedX))
	run := xc - depotShedRidgeX
	span := float64(depotShedX - depotShedRidgeX)
	north := depotShedRidgeY - run*(depotShedRidgeY-depotShedNE)/span
	newNorth := depotShedRidgeY - run*(depotShedRidgeY-depotShedNE-depotShedSlide)/span
	south := depotShedRidgeY + run*(depotShedSE-depotShedRidgeY)/span
	fy := float64(y)
	switch {
	case fy < newNorth && trim:
		return fy - depotShedSlide, true
	case fy < newNorth:
		return fy + north - newNorth, true
	case fy < south:
		// The east face: an affine map fixing the ridge end and the
		// south-east corner, taking the new north-east corner to the old.
		b := (depotShedSE - depotShedNE) / (depotShedSE - depotShedNE - depotShedSlide)
		a := (depotShedSE - depotShedRidgeY - b*(depotShedSE-depotShedRidgeY)) / span
		return depotShedRidgeY + a*run + b*(fy-depotShedRidgeY), true
	}
	return fy, true
}

var compoundShedDepotNorth = compoundShedEdit{
	Fills: []compoundShedFill{
		{Dst: image.Rect(depotShedLeft, depotShedTop, depotShedEnd, depotShedTop+depotShedSlide), Y: compoundShedAxis{Lo: 245, Hi: depotShedTop}, Soft: [4]int{5, 0, 5, 0}},
	},
	Warps: []compoundShedWarp{
		{Name: "north-west shed", Rect: image.Rect(depotShedLeft, depotShedTop+depotShedSlide, depotShedEnd, depotShedBottom), Map: compoundShedDepotNorthMap},
	},
}

// The factory's two east sheds in the top row (on factory-summer-v3 /
// factory-winter-v5) are shortened by 36 master px (5.8 world px), so their
// north eaves lie 3 world px inside the E cell's top edge: 10 rows of the
// north slope and the band of the south slope with the skylight and its
// shadow drop out, the north part slides south, the south eave and wall stay.
// The strip they leave and the gaps beside them (taking the upper pipe and the
// beam there; the lower pipe stays) fill with the dense trees west of the
// north-west hall (columns 220-329, rows 230-273, mirrored across), so the
// hedge north of the row reaches down to the new eaves and the hex edge fades
// it out like the rest.
var compoundShedFactoryEast = compoundShedEdit{
	Fills: []compoundShedFill{
		{Dst: image.Rect(1002, 238, 1272, 282), X: compoundShedAxis{Lo: 220, Hi: 330}, Y: compoundShedAxis{Off: 230 - 238}, Soft: [4]int{0, 4, 0, 3}},
	},
	Warps: []compoundShedWarp{
		compoundShedCut("middle shed", 1033, 1117, 243, 305, [2]int{251, 261}, [2]int{279, 305}),
		compoundShedCut("east shed", 1190, 1272, 243, 306, [2]int{252, 262}, [2]int{280, 306}),
	},
}

// compoundShedEdits lists the new masters, in the order they build on each
// other.
func compoundShedEdits() []compoundShedEdit {
	with := func(e compoundShedEdit, base, out string) compoundShedEdit {
		e.Base, e.Out = base, out
		return e
	}
	return []compoundShedEdit{
		with(compoundShedDepot, "art/terrain/cmp/cmp-summer-v2.png", "art/terrain/cmp/cmp-summer-v3.png"),
		with(compoundShedDepot, "art/terrain/cmp/cmp-winter-v3.png", "art/terrain/cmp/cmp-winter-v4.png"),
		with(compoundShedFactory, "art/terrain/factory/factory-summer-v2.png", "art/terrain/factory/factory-summer-v3.png"),
		with(compoundShedFactory, "art/terrain/factory/factory-winter-v4.png", "art/terrain/factory/factory-winter-v5.png"),
		with(compoundShedDepotNorth, "art/terrain/cmp/cmp-summer-v3.png", "art/terrain/cmp/cmp-summer-v4.png"),
		with(compoundShedDepotNorth, "art/terrain/cmp/cmp-winter-v4.png", "art/terrain/cmp/cmp-winter-v5.png"),
		with(compoundShedFactoryEast, "art/terrain/factory/factory-summer-v3.png", "art/terrain/factory/factory-summer-v4.png"),
		with(compoundShedFactoryEast, "art/terrain/factory/factory-winter-v5.png", "art/terrain/factory/factory-winter-v6.png"),
	}
}

// compoundShedRegion is a keep polygon (world px) that held a building cut by
// the outer boundary before an edit. In Unkept regions (the shortened sheds)
// no keep polygon of the region's own reaches the 2 world px band at the
// boundary any more: a shrunk one may hold the sheds from 4 world px inside.
type compoundShedRegion struct {
	Name    string
	Polygon []connectionPoint
	Unkept  bool
}

// compoundShedRegions: the depot's south-west shed with its crates and
// barrels (moved) and its north-west shed (shortened), the factory's
// north-west hall (moved) and the two east sheds of its top row (shortened).
var compoundShedRegions = map[string][]compoundShedRegion{
	"depot": {
		{Name: "south-west shed", Polygon: []connectionPoint{{51, 106}, {83, 106}, {83, 131}, {51, 131}}},
		{Name: "north-west shed", Polygon: []connectionPoint{{55, 38}, {81, 38}, {81, 62}, {55, 62}}, Unkept: true},
	},
	"factory": {
		{Name: "north-west hall", Polygon: []connectionPoint{{50, 38}, {102, 38}, {102, 60}, {50, 60}}},
		{Name: "east sheds", Polygon: []connectionPoint{{138, 36}, {206, 36}, {206, 52}, {138, 52}}, Unkept: true},
	},
}

// compoundShedManifests are the manifests cutting the edited masters.
var compoundShedManifests = []string{"compound-depot-v6.json", "compound-factory-v6.json", "compound-factory-summer-v4.json"}

// TestCompoundShedMasters: in each new master the moved pieces equal the base
// master under their translation (except a repainted patch), and every pixel
// outside the pieces' destinations and the refilled ground equals the base.
func TestCompoundShedMasters(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, e := range compoundShedEdits() {
		t.Run(filepath.Base(e.Out), func(t *testing.T) {
			base, err := loadPNG(filepath.Join(root, e.Base))
			if err != nil {
				t.Fatal(err)
			}
			got, err := loadPNG(filepath.Join(root, e.Out))
			if err != nil {
				t.Fatal(err)
			}
			if got.Bounds() != base.Bounds() {
				t.Fatalf("size %v, want %v", got.Bounds(), base.Bounds())
			}
			if want := compoundShedApply(base, e); string(want.Pix) != string(got.Pix) {
				t.Errorf("master differs from the edit of %s", e.Base)
			}
			mad := func(n int, sum float64) float64 { return sum / float64(max(n, 1)) / 4 / 255 }
			diff := func(a, b color.NRGBA) float64 {
				return math.Abs(float64(a.R)-float64(b.R)) + math.Abs(float64(a.G)-float64(b.G)) + math.Abs(float64(a.B)-float64(b.B)) + math.Abs(float64(a.A)-float64(b.A))
			}
			for i, p := range e.Pieces {
				n, sum := 0, 0.
				r := p.Src.Add(p.D)
				for y := r.Min.Y; y < r.Max.Y; y++ {
				pixel:
					for x := r.Min.X; x < r.Max.X; x++ {
						// The core of the piece, unless a later piece covers it
						// or the source was repainted.
						if p.weight(x, y) < 1 {
							continue
						}
						for _, q := range e.Pieces[i+1:] {
							if q.weight(x, y) > 0 {
								continue pixel
							}
						}
						for _, f := range e.Patches {
							if image.Pt(x-p.D.X, y-p.D.Y).In(f.outer()) {
								continue pixel
							}
						}
						n++
						sum += diff(got.NRGBAAt(x, y), base.NRGBAAt(x-p.D.X, y-p.D.Y))
					}
				}
				t.Logf("%s: %d px moved by %v, MAD %.5f", p.Name, n, p.D, mad(n, sum))
				if n < 100 || mad(n, sum) >= 1./255 {
					t.Errorf("%s: moved pixels differ from the base (MAD %.5f over %d px)", p.Name, mad(n, sum), n)
				}
			}
			// A warp keeps the pixels it maps to whole source rows: they equal
			// the base under the slide; it resamples the others (a hip's east
			// face), which only count as rebuilt.
			for _, w := range e.Warps {
				n, slid, rebuilt, sum := 0, 0, 0, 0.
				for y := w.Rect.Min.Y; y < w.Rect.Max.Y; y++ {
					for x := w.Rect.Min.X; x < w.Rect.Max.X; x++ {
						sy, ok := w.Map(x, y)
						switch {
						case !ok:
							continue
						case sy != math.Floor(sy):
							rebuilt++
							continue
						case int(sy) != y:
							slid++
						}
						n++
						sum += diff(got.NRGBAAt(x, y), base.NRGBAAt(x, int(sy)))
					}
				}
				t.Logf("%s: %d px kept (%d slid south), %d px resampled, MAD %.5f", w.Name, n, slid, rebuilt, mad(n, sum))
				if slid < 100 || mad(n, sum) >= 1./255 {
					t.Errorf("%s: kept pixels differ from the base under the slide (MAD %.5f over %d px, %d slid)", w.Name, mad(n, sum), n, slid)
				}
			}
			edited := func(pt image.Point) bool {
				for _, f := range append(e.Patches[:len(e.Patches):len(e.Patches)], e.Fills...) {
					if pt.In(f.outer()) {
						return true
					}
				}
				for _, p := range e.Pieces {
					if pt.In(p.Src.Add(p.D)) {
						return true
					}
				}
				for _, w := range e.Warps {
					if pt.In(w.Rect) {
						return true
					}
				}
				return false
			}
			n, sum, changed := 0, 0., 0
			for y := range base.Rect.Dy() {
				for x := range base.Rect.Dx() {
					if edited(image.Pt(x, y)) {
						continue
					}
					n++
					d := diff(got.NRGBAAt(x, y), base.NRGBAAt(x, y))
					sum += d
					if d > 0 {
						changed++
					}
				}
			}
			t.Logf("outside the edit: %d px, MAD %.6f, %d changed", n, mad(n, sum), changed)
			if mad(n, sum) >= 1./255 || changed > 0 {
				t.Errorf("pixels outside the edit changed: %d, MAD %.6f", changed, mad(n, sum))
			}
		})
	}
}

// TestCompoundShedsInsideHexEdge: in the 2x scene artwork no building of the
// edited regions reaches the outer boundary any more. Inside each region every
// pixel within 2 world px of the boundary keeps alpha 64 at most, and the 50 %
// contour of the feather runs along the boundary there like elsewhere (2.5 to
// 8 world px deep). Pixels held by a keep polygon of another feature (a fence,
// a road reaching into the region) do not count. No keep polygon holds the
// shortened sheds at the boundary any more.
func TestCompoundShedsInsideHexEdge(t *testing.T) {
	for _, name := range compoundShedManifests {
		in, err := loadCompoundInputs(filepath.Join("..", "..", "art", "pipeline", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range in.Manifest.Scenes {
			l, err := compoundLayoutFor(s.ID)
			if err != nil {
				t.Fatal(err)
			}
			framed, art4, _, _, err := compoundSceneArtwork(in, s, l)
			if err != nil {
				t.Fatal(err)
			}
			art2, err := resizeLanczos(art4, l.Width*2, l.Height*2)
			if err != nil {
				t.Fatal(err)
			}
			before2, err := resizeLanczos(framed, l.Width*2, l.Height*2)
			if err != nil {
				t.Fatal(err)
			}
			f, _, err := compoundEntryRoad(in.EntryRoads[s.ID], l)
			if err != nil {
				t.Fatal(err)
			}
			corridor := f.Width/2 + f.Border + 2
			edges := compoundBoundary(l)
			for _, r := range compoundShedRegions[l.Layout] {
				t.Run(s.ID+"/"+r.Name, func(t *testing.T) {
					region := [][]connectionPoint{r.Polygon}
					// Keep polygons of other features: their bounding box
					// centre lies outside the region.
					var foreign, own [][]connectionPoint
					for _, k := range s.EdgeKeep {
						lo, hi := k[0], k[0]
						for _, p := range k {
							lo = connectionPoint{min(lo[0], p[0]), min(lo[1], p[1])}
							hi = connectionPoint{max(hi[0], p[0]), max(hi[1], p[1])}
						}
						if compoundKeepDistance(region, lo.add(hi).mul(.5)) > 0 {
							foreign = append(foreign, k)
						} else {
							own = append(own, k)
						}
					}
					held := func(p connectionPoint) bool {
						return len(foreign) > 0 && compoundKeepDistance(foreign, p) < compoundKeepRamp
					}
					band, opaque, worst, kept := 0, 0, uint8(0), 0
					for y := range l.Height * 2 {
						for x := range l.Width * 2 {
							p := connectionPoint{(float64(x) + .5) / 2, (float64(y) + .5) / 2}
							if d := compoundSignedDistance(l, edges, p); d < 0 || d > 2 || compoundKeepDistance(region, p) > 0 || held(p) {
								continue
							}
							band++
							if len(own) > 0 && compoundKeepDistance(own, p) < compoundKeepRamp {
								kept++
							}
							a := art2.NRGBAAt(x, y).A
							if a > 64 {
								opaque++
							}
							worst = max(worst, a)
						}
					}
					// Along the region's boundary the march may run into a
					// building's keep polygon: wherever the artwork was opaque at
					// the boundary, the 50 % contour lies inside the feather band
					// (the faded ground or the building beyond it), not on the
					// boundary as a cut.
					open := func(p connectionPoint) bool {
						lat, depth := compoundEntryLateral(t, l, p)
						return !(lat <= corridor+1.5 && depth < s.EdgeFeather+2)
					}
					samples := compoundFeatherContourSamples(l, s.EdgeFeather, before2, art2, open)
					var depths []float64
					for _, c := range samples {
						if compoundKeepDistance(region, c.at) == 0 && !held(c.at) {
							depths = append(depths, c.depth)
						}
					}
					mean, sd, lo, hi := compoundContourStats(t, depths)
					t.Logf("%s: %d 2x px of the region within 2 world px of the boundary, %d above alpha 64, max alpha %d; 50 %% contour over %d samples there: mean %.2f, sd %.2f, range %.2f..%.2f world px", name, band, opaque, worst, len(depths), mean, sd, lo, hi)
					if band < 40 {
						t.Fatalf("too few boundary pixels in the region: %d", band)
					}
					if worst > 64 {
						t.Errorf("%d px of the region within 2 world px of the boundary keep alpha above 64 (max %d)", opaque, worst)
					}
					if r.Unkept && kept > 0 {
						t.Errorf("a keep polygon of the region still holds %d px within 2 world px of the boundary", kept)
					}
					if len(depths) < 8 {
						t.Errorf("the 50 %% contour covers only %d samples of the region's boundary", len(depths))
					} else if lo < 2.5-1e-9 || hi > 8+1e-9 {
						t.Errorf("50 %% contour outside 2.5..8 world px in the region: %.2f..%.2f", lo, hi)
					}
				})
			}
		}
	}
}
