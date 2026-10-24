package main

import (
	"errors"
	"fmt"
	"image/color"
	"math"
)

// Winter streams and trenches: a connected family with season "winter" keeps
// its summer geometry and joins and paints the winter ground ("grass", the
// snow meadow of the winter manifests) onto its banks. A stream keeps open
// water in its centre, calmed of its sun glints, which the pack grade turns to
// winter water (tile suffix _w), and freezes a pale ice rim along both banks; a
// trench carries snow on
// its parapet and parados and dark slush on its floor, its timber unchanged.

// Ice rim and frost of a winter stream.
const (
	winterIceMin, winterIceSpan = 1.5, 2.5 // ice rim width 1.5..4 world px from the bank
	winterIceTint               = .45      // snow towards the cold ice grey
	winterThinIce               = .35      // the inner edge, where thin ice shows the water
	winterRockGrey              = .55      // dry stones turn cold grey (as the winter shore)
	winterRockSnow              = .85      // and carry snow on their light tops
	winterCalmRadius            = 3.0      // world px: the ring that gives the local water level
	winterCalmKnee              = 6.0      // L* a highlight keeps above the local water level, at most
)

// winterCalmRing is the number of water samples on the ring around a pixel.
const winterCalmRing = 12

var (
	winterIceColor  = color.NRGBA{176, 192, 200, 255}
	winterThinColor = color.NRGBA{142, 162, 170, 255}
	winterSlush     = color.NRGBA{54, 53, 55, 255}
)

// Snow, clods and slush of a winter trench.
const (
	winterTrenchClods   = .3  // dark spoil clods showing through the snow
	winterTrenchTrodden = .45 // the trodden, muddy lip along the revetment
	winterTrenchSlush   = .45 // the floor towards dark slush
	winterTrenchRelief  = .6  // share of the summer bank shading kept on the snow
)

func validateConnectionSeason(f connectionFamily) error {
	switch {
	case f.Season == "":
		return nil
	case f.Season != "winter":
		return errors.New("family season must be winter when set")
	case f.ID == "trench", f.ID == "stream" && f.BankProfile == authoredBankProfile:
		return nil
	}
	return errors.New("season winter requires a trench or an authored-rocks-v1 stream")
}

func connectionSeasonDescription(f connectionFamily) string {
	if f.Season != "winter" {
		return ""
	}
	if f.ID == "stream" {
		return fmt.Sprintf(" Stream season=winter: the grass material is the winter ground. Dry bank stones blend toward their own grey at %.2f*S((d-0.3)/1.8) and toward the snow at %.2f times that where their light tops are (as the winter shore). An ice rim covers the outer %.1f..%.1f world px of the channel: width=%.1f+%.1f*S((0.65*N14+0.35*N37-0.22)/0.56), Nn lattice value noise (lattice coordinates a=x/71, b=(y-40a)/80 cut into n cells, hashed modulo n, smoothstep bilinear), so it repeats under every runtime translation. Ice is the snow blended toward RGB(%d,%d,%d) at %.2f and toward RGB(%d,%d,%d) at up to %.2f over the inner 1.2 world px; coverage ramps in over 0.5 world px at the inner edge and hands over to the summer wet lip and stones over the first 0.6 world px of the bank. Open water is calm, without sun glints: a pixel lighter than the local water level (the mean of %d water samples on a %.1f world px ring) blends toward that mean colour so its L* excess e over the level shrinks to %.1f*tanh(e/%.1f), weighted by the open-water share 1-S((d+0.45)/0.7); the ice rim is painted over it unchanged. The pack grade of a _w tile makes it winter water.",
			winterRockGrey, winterRockSnow, winterIceMin, winterIceMin+winterIceSpan, winterIceMin, winterIceSpan,
			winterIceColor.R, winterIceColor.G, winterIceColor.B, winterIceTint, winterThinColor.R, winterThinColor.G, winterThinColor.B, winterThinIce,
			winterCalmRing, winterCalmRadius, winterCalmKnee, winterCalmKnee)
	}
	return fmt.Sprintf(" Trench season=winter: the grass material is the winter ground. Parapet and parados take the snow with %.1f of the summer bank shading, dark spoil clods show through at up to %.2f where the earth is dark, and the lip within 1.2 world px of the revetment is trodden toward the earth at up to %.2f. The floor blends toward RGB(%d,%d,%d) slush at %.2f. Timber is unchanged.",
		winterTrenchRelief, winterTrenchClods, winterTrenchTrodden, winterSlush.R, winterSlush.G, winterSlush.B, winterTrenchSlush)
}

// winterLatticeNoise is smooth value noise in [0,1] at world point p that
// repeats under every runtime lattice translation: p = a(71,40) + b(0,80), and
// a and b are each cut into n cells whose corner values are hashed modulo n.
func winterLatticeNoise(p connectionPoint, n int, seed uint32) float64 {
	a := p[0] / 71
	b := (p[1] - 40*a) / 80
	u := math.Round(a*float64(n)*1e9) / 1e9
	v := math.Round(b*float64(n)*1e9) / 1e9
	i, j := math.Floor(u), math.Floor(v)
	fu, fv := connectionSmoothStep(u-i), connectionSmoothStep(v-j)
	corner := func(di, dj int) float64 {
		x := uint32(((int(i)+di)%n + n) % n)
		y := uint32(((int(j)+dj)%n + n) % n)
		h := x*0x9e3779b1 ^ y*0x85ebca77 ^ seed*0xc2b2ae3d
		h ^= h >> 15
		h *= 0x2c1b3c6d
		h ^= h >> 12
		h *= 0x297a2d39
		h ^= h >> 15
		return float64(h&0xffff) / 0xffff
	}
	top := corner(0, 0) + (corner(1, 0)-corner(0, 0))*fu
	bottom := corner(0, 1) + (corner(1, 1)-corner(0, 1))*fu
	return top + (bottom-top)*fv
}

// winterIceWidth is the width of the ice rim at p, 1.5..4 world px.
func winterIceWidth(p connectionPoint) float64 {
	// Value noise gathers around 0.5; the smoothstep spreads it so the rim
	// visibly swells into floes and narrows to a thin lip.
	n := .65*winterLatticeNoise(p, 14, 1) + .35*winterLatticeNoise(p, 37, 2)
	return winterIceMin + winterIceSpan*connectionSmoothStep((n-.22)/.56)
}

// winterLightness is the CIE L* of an sRGB colour.
func winterLightness(c color.NRGBA) float64 {
	lin := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= .04045 {
			return s / 12.92
		}
		return math.Pow((s+.055)/1.055, 2.4)
	}
	y := .2126729*lin(c.R) + .7151522*lin(c.G) + .0721750*lin(c.B)
	if y > 216.0/24389 {
		return 116*math.Cbrt(y) - 16
	}
	return 24389.0 / 27 * y
}

// winterWaterLevel is the mean colour of the water material on a ring of
// winterCalmRadius world px around p, and its L*: the local water level.
func winterWaterLevel(in connectionInputs, f connectionFamily, p connectionPoint) (color.NRGBA, float64) {
	var r, g, b float64
	for i := range winterCalmRing {
		a := 2 * math.Pi * float64(i) / winterCalmRing
		c := sampleAuthoredWater(in.materials[f.Material], p.add(connectionPoint{winterCalmRadius * math.Cos(a), winterCalmRadius * math.Sin(a)}))
		r, g, b = r+float64(c.R), g+float64(c.G), b+float64(c.B)
	}
	mean := color.NRGBA{uint8(math.Round(r / winterCalmRing)), uint8(math.Round(g / winterCalmRing)), uint8(math.Round(b / winterCalmRing)), 255}
	return mean, winterLightness(mean)
}

// winterCalmWater calms the open water of c at p, signed distance d from the
// water's edge: winter water carries no sun glints. A pixel lighter than the
// local water level blends toward that level's colour so that it keeps at most
// winterCalmKnee L* above it (a soft knee, so the ripples keep their order);
// the share of open water in c, 1-S((d+0.45)/0.7), takes the calm.
func winterCalmWater(in connectionInputs, f connectionFamily, p connectionPoint, d float64, c color.NRGBA) color.NRGBA {
	open := 1 - connectionSmoothStep((d+.45)/.7)
	if open == 0 {
		return c
	}
	mean, level := winterWaterLevel(in, f, p)
	excess := winterLightness(c) - level
	if excess <= 0 {
		return c
	}
	keep := winterCalmKnee * math.Tanh(excess/winterCalmKnee) / excess
	return connectionBlend(c, mean, open*(1-keep))
}

// winterStreamColor turns the summer stream colour c at p, signed distance d
// from the water's edge (negative in the water), into winter: calm open water
// without glints, frosted stones and an ice rim along the bank.
func winterStreamColor(in connectionInputs, f connectionFamily, p connectionPoint, d float64, c color.NRGBA) color.NRGBA {
	c = winterCalmWater(in, f, p, d, c)
	snow := sampleGrassMaterial(in.materials["grass"], p)
	rock := sampleAuthoredBank(in.materials[f.BorderMaterial], p)
	lum := float64(rock.R) + float64(rock.G) + float64(rock.B)
	grey := uint8(lum / 3)
	dry := connectionSmoothStep((d - .3) / 1.8)
	c = connectionBlend(c, color.NRGBA{grey, grey, min(255, grey+4), 255}, winterRockGrey*dry)
	c = connectionBlend(c, snow, winterRockSnow*dry*connectionSmoothStep((lum/765-.38)/.22))
	width := winterIceWidth(p)
	ice := connectionBlend(snow, winterIceColor, winterIceTint)
	ice = connectionBlend(ice, winterThinColor, winterThinIce*(1-connectionSmoothStep((d+width)/1.2)))
	cover := connectionRamp(d+width, .5) * (1 - connectionSmoothStep(d/.6))
	return connectionBlend(c, ice, cover)
}

// winterTrenchPaint turns a trench's summer bank and floor at p into winter.
// earth is the sampled spoil, shading the summer bank brightness factor and
// coreDistance the signed distance from the floor's edge.
func winterTrenchPaint(in connectionInputs, p connectionPoint, earth color.NRGBA, shading, coreDistance float64, floor color.NRGBA) (color.NRGBA, color.NRGBA) {
	snow := connectionTrenchShade(sampleGrassMaterial(in.materials["grass"], p), 1+winterTrenchRelief*(shading-1))
	luminance := .2126*float64(earth.R) + .7152*float64(earth.G) + .0722*float64(earth.B)
	spoil := connectionTrenchShade(earth, shading)
	bank := connectionBlend(snow, spoil, winterTrenchClods*connectionSmoothStep((100-luminance)/40))
	bank = connectionBlend(bank, spoil, winterTrenchTrodden*(1-connectionSmoothStep(coreDistance/1.2)))
	return bank, connectionBlend(floor, winterSlush, winterTrenchSlush)
}
