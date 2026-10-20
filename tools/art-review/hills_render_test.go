package main

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"
)

// The scalar is independent of terrain and periodic in world coordinates; a per-cell stamp breaks this.
func TestHillsGlobalScalarNormal(t *testing.T) {
	height := hillsTestHeight()
	for _, p := range []connectionPoint{{-13.125, 20.375}, {71.125, 40.125}, {221.75, 151.5}, {0, 0}} {
		v := hillsHeightAt(height, p)
		n := hillsGradientAt(height, p)
		for _, d := range []connectionPoint{{568, 0}, {0, 320}, {-568, -320}} {
			if math.Abs(hillsHeightAt(height, p.add(d))-v) > 1e-10 || n.sub(hillsGradientAt(height, p.add(d))).length() > 1e-10 {
				t.Fatal("global field period")
			}
		}
		// Independently sample a synthetic affine luminance image at cosine coordinates.
		u := (.5 - .5*math.Cos(2*math.Pi*(p[0]+113)/568)) * 511
		vv := (.5 - .5*math.Cos(2*math.Pi*(p[1]+79)/320)) * 511
		ix, iy := int(u), int(vv)
		fx, fy := u-float64(ix), vv-float64(iy)
		lum := 0.
		for dy := range 2 {
			for dx := range 2 {
				q := height.NRGBAAt(min(ix+dx, 511), min(iy+dy, 511))
				wx, wy := 1-fx, 1-fy
				if dx == 1 {
					wx = fx
				}
				if dy == 1 {
					wy = fy
				}
				lum += (.2126*float64(q.R) + .7152*float64(q.G) + .0722*float64(q.B)) * wx * wy
			}
		}
		if math.Abs(v-32*lum/255) > 1e-12 {
			t.Fatal("wrong Rec709/cosine height", v, lum)
		}
		want := connectionPoint{(hillsHeightAt(height, p.add(connectionPoint{2, 0})) - hillsHeightAt(height, p.add(connectionPoint{-2, 0}))) / 4, (hillsHeightAt(height, p.add(connectionPoint{0, 2})) - hillsHeightAt(height, p.add(connectionPoint{0, -2}))) / 4}
		if n != want {
			t.Fatal("wrong global central derivative")
		}
	}
}

// Every grass-facing final quiet band must survive whole-image filtering exactly.
func TestHillsAllMasksFinalQuietAndRelief(t *testing.T) {
	path := hillsTestFixture(t)
	in, err := loadHillsInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	r := newHillsRenderer(in.Sources["grass"].Image, in.Sources["height"].Image)
	base, _ := newHillsContext("grass", []string{"grass", "grass", "grass", "grass", "grass", "grass"}, connectionPoint{})
	grass, err := r.tiles(base)
	if err != nil {
		t.Fatal(err)
	}
	existing, err := exportGrassBaseTiles(in.Sources["grass"].Image)
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []int{2, 4} {
		if !bytes.Equal(grass[scale].Pix, existing[scale].Pix) {
			t.Fatal("base changed")
		}
	}
	for mask := range 64 {
		neighbors := []string{"grass", "grass", "grass", "grass", "grass", "grass"}
		for i := range 6 {
			if mask&(1<<i) != 0 {
				neighbors[i] = "hills"
			}
		}
		c, _ := newHillsContext("hills", neighbors, connectionPoint{})
		tiles, err := r.tiles(c)
		if err != nil {
			t.Fatal(err)
		}
		for scale, im := range tiles {
			changed, quiet := 0, 0
			for y := range 80 * scale {
				for x := range 96 * scale {
					pixel := im.NRGBAAt(x, y)
					if !groundTestOwned(x, y, scale) {
						if pixel.A != 0 || pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
							t.Fatal("unowned RGBA")
						}
						continue
					}
					if pixel.A != 255 {
						t.Fatal("owned alpha")
					}
					p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
					isQuiet := false
					for n, d := range groundTestDeltas {
						if neighbors[n] == "grass" && groundTestDistance(p[0]-float64(d[0]), p[1]-float64(d[1])) <= 4 {
							isQuiet = true
						}
					}
					if isQuiet {
						quiet++
						if pixel != grass[scale].NRGBAAt(x, y) {
							t.Fatalf("mask%d %dx quiet leak %v", mask, scale, p)
						}
					}
					if pixel != grass[scale].NRGBAAt(x, y) {
						changed++
					}
				}
			}
			if changed == 0 {
				t.Fatalf("mask%d %dx no relief", mask, scale)
			}
			if mask != 63 && quiet == 0 {
				t.Fatal("quiet not exercised")
			}
		}
	}
}

// The isolated cue must fade in strong global relief and never alter a mixed context.
func TestHillsIsolatedCuePhaseAndFade(t *testing.T) {
	flat := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	for y := range 512 {
		for x := range 512 {
			flat.SetNRGBA(x, y, color.NRGBA{128, 128, 128, 255})
		}
	}
	neighbors := []string{"grass", "grass", "grass", "grass", "grass", "grass"}
	c, _ := newHillsContext("hills", neighbors, connectionPoint{71, 40})
	cue := hillsCueFor(flat, c)
	if !cue.Enabled || cue.A != 6 || math.Abs(cue.S) > 1e-14 || math.Abs(cue.Theta-4*math.Pi*(3-math.Sqrt(5))) > 1e-14 {
		t.Fatalf("flat cue %#v", cue)
	}
	if math.Abs(hillsCueHeight(cue, connectionPoint{48, 40})-6) > 1e-12 || hillsCueHeight(cue, connectionPoint{80, 40}) != 0 {
		t.Fatal("cue compact support")
	}
	for _, id := range []string{"hills", "grass"} {
		n := append([]string(nil), neighbors...)
		if id == "hills" {
			n[2] = "hills"
		}
		mixed, _ := newHillsContext(id, n, connectionPoint{71, 40})
		q := hillsCueFor(flat, mixed)
		if q.Enabled || q.A != 0 || q.S != 0 || q.Theta != 0 {
			t.Fatal("cue enabled outside isolated hills")
		}
	}
	height := hillsTestHeight()
	low, high := math.Inf(1), 0.
	// A stronger independent synthetic field exercises the fade branch;
	// the smooth input fixture intentionally stays below the fade threshold.
	for y := range 512 {
		for x := range 512 {
			v := uint8(math.Round(127 + 100*math.Sin(2*math.Pi*float64(x)/128)))
			height.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	for line := range 8 {
		for col := range 4 {
			p := connectionPoint{float64(142*col + 71*(line%2)), float64(40 * line)}
			c, _ := newHillsContext("hills", neighbors, p)
			q := hillsCueFor(height, c)
			sum := 0.
			for _, d := range []connectionPoint{{0, 0}, {12, 0}, {-12, 0}, {0, 12}, {0, -12}, {8, 8}, {8, -8}, {-8, 8}, {-8, -8}} {
				g := hillsGradientAt(height, p.add(connectionPoint{48, 40}).add(d))
				z := hillsShade(g) - 1
				sum += z * z
			}
			s := math.Sqrt(sum / 9)
			v := max(0., min(1., (s-.08)/.12))
			a := 6 * (1 - v*v*v*(10-15*v+6*v*v))
			if math.Abs(q.S-s) > 1e-14 || math.Abs(q.A-a) > 1e-12 || math.Abs(q.Theta-float64(line*4+col)*math.Pi*(3-math.Sqrt(5))) > 1e-12 {
				t.Fatal("cue phase/amplitude formula")
			}
			low = min(low, q.A)
			high = max(high, q.A)
		}
	}
	if high <= low {
		t.Fatal("amplitude does not respond to relief")
	}
	if math.Abs(hillsShade(connectionPoint{.1, -.2})-.9672065612091721) > 1e-12 {
		t.Fatal("fixed world light")
	}
}

// Every isolated production phase has a compact visible cue and exact final grass boundaries.
func TestHillsAll32IsolatedPhaseWitness(t *testing.T) {
	path := hillsTestFixture(t)
	in, err := loadHillsInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	r := newHillsRenderer(in.Sources["grass"].Image, in.Sources["height"].Image)
	n := []string{"grass", "grass", "grass", "grass", "grass", "grass"}
	base, _ := newHillsContext("grass", n, connectionPoint{})
	grass, err := r.tiles(base)
	if err != nil {
		t.Fatal(err)
	}
	for line := range 8 {
		for col := range 4 {
			c, _ := newHillsContext("hills", n, connectionPoint{float64(142*col + 71*(line%2)), float64(40 * line)})
			tiles, err := r.tiles(c)
			if err != nil {
				t.Fatal(err)
			}
			for scale, im := range tiles {
				changed := 0
				maxDelta := 0
				for y := range 80 * scale {
					for x := range 96 * scale {
						if !groundTestOwned(x, y, scale) {
							continue
						}
						p := connectionPoint{(float64(x) + .5) / float64(scale), (float64(y) + .5) / float64(scale)}
						quiet := false
						for _, d := range groundTestDeltas {
							quiet = quiet || groundTestDistance(p[0]-float64(d[0]), p[1]-float64(d[1])) <= 4
						}
						a, b := im.NRGBAAt(x, y), grass[scale].NRGBAAt(x, y)
						if a.A != 255 {
							t.Fatal("isolated alpha")
						}
						if quiet && a != b {
							t.Fatalf("phase%v quiet boundary", c.Phase)
						}
						delta := max(hillsTestAbs(int(a.R)-int(b.R)), hillsTestAbs(int(a.G)-int(b.G)), hillsTestAbs(int(a.B)-int(b.B)))
						if delta > 0 {
							changed++
						}
						maxDelta = max(maxDelta, delta)
					}
				}
				if changed < 100*scale*scale || maxDelta < 3 {
					t.Fatalf("isolated phase%v %dx too flat (%d,%d)", c.Phase, scale, changed, maxDelta)
				}
			}
		}
	}
}

func hillsTestAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Multiplying gates after differentiation must flatten slopes without introducing a gate-gradient cliff.
func TestHillsGrassSlopeGateProduct(t *testing.T) {
	path := hillsTestFixture(t)
	in, err := loadHillsInputs(path)
	if err != nil {
		t.Fatal(err)
	}
	r := newHillsRenderer(in.Sources["grass"].Image, in.Sources["height"].Image)
	n := []string{"grass", "grass", "hills", "hills", "hills", "hills"}
	c, _ := newHillsContext("hills", n, connectionPoint{})
	raw := r.raw(c)
	for _, at := range []image.Point{{336, 176}, {320, 176}, {352, 176}, {336, 192}} {
		p := connectionPoint{(float64(at.X)+.5)/4 - 4, (float64(at.Y)+.5)/4 - 4}
		gate := 1.
		for i := range 2 {
			d := groundTestDeltas[i]
			distance := groundTestDistance(p[0]-float64(d[0]), p[1]-float64(d[1]))
			v := max(0., min(1., (distance-7)/12))
			gate *= v * v * v * (10 - 15*v + 6*v*v)
		}
		g := hillsGradientAt(in.Sources["height"].Image, p).mul(gate)
		lightLength := math.Sqrt(.55*.55 + .4*.4 + .75*.75)
		z := .75 / lightLength
		dot := (g[0]*.55/lightLength + g[1]*.4/lightLength + z) / math.Sqrt(1+g[0]*g[0]+g[1]*g[1])
		shade := max(.55, min(1.25, (.35+.65*max(0, dot))/(.35+.65*z)))
		grass := sampleGrassMaterial(in.Sources["grass"].Image, p)
		round := func(v uint8) uint8 { return uint8(math.Floor(max(0, min(255, float64(v)*shade)) + .5)) }
		want := color.NRGBA{round(grass.R), round(grass.G), round(grass.B), 255}
		if got := raw.NRGBAAt(at.X, at.Y); got != want {
			t.Fatalf("slope product at %v gate%g: %v want%v", p, gate, got, want)
		}
	}
}
