package main

import (
	"fmt"
	"image"
	"image/draw"
	"io/fs"
	"path"
	"runtime"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// spriteDensities are the unit and terrain pack densities the game loads:
// what hlasset build-units and build-terrain write, but only 2x in the web
// build, which downloads its packs and cannot afford the 4x ones.
var spriteDensities = func() []int {
	if runtime.GOOS == "js" {
		return []int{2}
	}
	return []int{2, 4}
}()

// spriteDensity is the unit pack drawn at effective zoom eff: the asset density,
// but at least 2x, since no 1x unit sprites exist, and at most the largest
// loaded. Drawing the 2x sprite smaller is the renderer's choice, not a loader
// fallback.
func spriteDensity(eff float64) int {
	return min(max(2, assetDensity(eff)), spriteDensities[len(spriteDensities)-1])
}

// unitSpent reports whether u has acted this turn and is drawn greyscale.
func unitSpent(u *state.Unit) bool { return u.Moved || u.Attacked }

// unitSprite picks the sprite of a unit of type typ owned by owner. The palette
// follows the owner, so captured units change colour; remembered units use the
// greyscale spent sprite.
func unitSprite(typ string, owner state.Player, spent, remembered bool, face unitsprite.Facing) (family, id string, ok bool) {
	family, ok = unitsprite.Family(typ)
	if !ok {
		return "", "", false
	}
	st := unitsprite.Normal
	if spent || remembered {
		st = unitsprite.Spent
	}
	return family, unitsprite.ID(family, unitsprite.PaletteOf(owner), face, st), true
}

type packKey struct {
	group   string // unit family or terrain group
	density int
}

type texKey struct {
	id      string
	density int
	dressed bool           // with the map rim and contact shadow
	rim     unitsprite.Rim // of a dressed sprite
}

// artPacks is a runtime catalog of packs laid out as <d>x/<group>/manifest.json,
// plus the GPU textures made from them so far. Units have one group per family,
// terrain one per tile group.
type artPacks struct {
	packs     map[packKey]*assets.Pack
	tex       map[texKey]*ebiten.Image
	candidate bool // some pack is a review candidate, not approved art
	pad       int  // texels added around each texture
	// bleed lists the groups whose opaque edges are extended into the added
	// border; the others get a transparent border.
	bleed map[string]bool
}

// openPacks opens every group at every density.
func openPacks(fsys fs.FS, opt assets.Options, groups []string, densities []int) (*artPacks, error) {
	a := &artPacks{packs: map[packKey]*assets.Pack{}, tex: map[texKey]*ebiten.Image{}}
	for _, d := range densities {
		for _, group := range groups {
			dir := path.Join(fmt.Sprintf("%dx", d), group)
			sub, err := fs.Sub(fsys, dir)
			if err != nil {
				return nil, err
			}
			p, err := assets.Open(sub, "manifest.json", opt)
			if err != nil {
				return nil, fmt.Errorf("pack %s: %w", dir, err)
			}
			a.packs[packKey{group, d}] = p
			if p.Manifest().Status != "approved" {
				a.candidate = true
			}
		}
	}
	return a, nil
}

// openUnitArt opens the whole unit catalog and checks that every family has
// every sprite at every density, so drawing never meets a missing variant.
func openUnitArt(fsys fs.FS, opt assets.Options) (*artPacks, error) {
	a, err := openPacks(fsys, opt, unitsprite.Families(), spriteDensities)
	if err != nil {
		return nil, fmt.Errorf("unit %w", err)
	}
	for _, d := range spriteDensities {
		for _, fam := range unitsprite.Families() {
			for _, pal := range unitsprite.Palettes {
				for f := range unitsprite.Facings {
					for _, s := range unitsprite.States {
						if _, err := a.entry(fam, unitsprite.ID(fam, pal, f, s), d); err != nil {
							return nil, fmt.Errorf("unit pack %dx/%s: %w", d, fam, err)
						}
					}
				}
			}
		}
	}
	return a, nil
}

// entry is the catalog record of sprite id at density d.
func (a *artPacks) entry(group, id string, d int) (assets.Entry, error) {
	p, ok := a.packs[packKey{group, d}]
	if !ok {
		return assets.Entry{}, fmt.Errorf("no %dx pack %s", d, group)
	}
	return p.Entry(id, d)
}

// image returns the texture of sprite id at density d, decoding and uploading it
// on first use.
func (a *artPacks) image(group, id string, d int) (*ebiten.Image, assets.Entry, error) {
	return a.texture(group, id, d, false, unitsprite.Dark)
}

// mapImage returns the texture of unit sprite id at density d as the map draws
// it, dressed with rim r and its contact shadow (unitsprite.Dress), and its
// border in texels around the entry's canvas.
func (a *artPacks) mapImage(group, id string, d int, r unitsprite.Rim) (*ebiten.Image, assets.Entry, int, error) {
	img, e, err := a.texture(group, id, d, true, r)
	return img, e, a.pad + unitsprite.DressMargin(d), err
}

func (a *artPacks) texture(group, id string, d int, dressed bool, r unitsprite.Rim) (*ebiten.Image, assets.Entry, error) {
	e, err := a.entry(group, id, d)
	if err != nil {
		return nil, e, err
	}
	k := texKey{id, d, dressed, r}
	if img, ok := a.tex[k]; ok {
		return img, e, nil
	}
	src, err := a.packs[packKey{group, d}].Image(id, d)
	if err != nil {
		return nil, e, err
	}
	if dressed {
		src = unitsprite.Dress(src, d, unitsprite.Flies(group), r)
	}
	var img *ebiten.Image
	if a.pad == 0 {
		img = ebiten.NewImageFromImage(src)
	} else {
		b := src.Bounds()
		img = ebiten.NewImage(b.Dx()+2*a.pad, b.Dy()+2*a.pad)
		img.WritePixels(padded(src, a.pad, a.bleed[group]))
	}
	a.tex[k] = img
	return img, e, nil
}

// padded returns the RGBA pixels of src inside a border of pad texels,
// premultiplied as ebiten.Image.WritePixels expects. The border is transparent,
// or with bleed the opaque texels are grown into it one texel per step, copying
// an opaque neighbour's colour.
func padded(src image.Image, pad int, bleed bool) []byte {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()+2*pad, b.Dy()+2*pad))
	draw.Draw(dst, b.Sub(b.Min).Add(image.Pt(pad, pad)), src, b.Min, draw.Src)
	if !bleed {
		return dst.Pix
	}
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	for range pad {
		old := slices.Clone(dst.Pix)
		for y := range h {
			for x := range w {
				i := (y*w + x) * 4
				if old[i+3] == 255 {
					continue
				}
				for _, o := range [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}} {
					nx, ny := x+o[0], y+o[1]
					if j := (ny*w + nx) * 4; nx >= 0 && ny >= 0 && nx < w && ny < h && old[j+3] == 255 {
						copy(dst.Pix[i:i+4], old[j:j+4])
						break
					}
				}
			}
		}
	}
	return dst.Pix
}

// anchorGeoM places an image of entry e at density d, inside a border of pad
// texels, with its logical anchor on screen point (sx, sy), at effective zoom eff.
func anchorGeoM(e assets.Entry, d, pad int, eff, sx, sy float64) ebiten.GeoM {
	var m ebiten.GeoM
	m.Translate(-e.Anchor[0]*float64(d)-float64(pad), -e.Anchor[1]*float64(d)-float64(pad))
	m.Scale(eff/float64(d), eff/float64(d))
	m.Translate(sx, sy)
	return m
}

// unitRim is the rim of a unit on hex c (unitsprite.RimFor).
func (g *game) unitRim(c hexgrid.Coord) unitsprite.Rim {
	if g.lvl == nil || !g.grid.Contains(c) {
		return unitsprite.Dark
	}
	h := g.lvl.Lines[c.Line][c.Col]
	return unitsprite.RimFor(h.Terrain, len(h.Infra.IDs) > 0)
}

// drawUnitSprite draws a unit sprite on hex c, dressed with the hex's rim and
// its contact shadow, with its logical anchor on screen point (sx, sy), the hex
// centre, and reports whether it could. Errors go to the status
// line once and the caller falls back to the placeholder dot.
func (g *game) drawUnitSprite(dst *ebiten.Image, c hexgrid.Coord, sx, sy float64, typ string, owner state.Player, spent, remembered bool, face unitsprite.Facing) bool {
	if g.art == nil || typ == "" {
		// A ship's rear has no separate remake unit type.
		return false
	}
	family, id, ok := unitSprite(typ, owner, spent, remembered, face)
	if !ok {
		g.artErr = fmt.Sprintf("no sprite family for unit type %s", typ)
		return false
	}
	d := spriteDensity(g.cam.Effective())
	img, e, border, err := g.art.mapImage(family, id, d, g.unitRim(c))
	if err != nil {
		g.artErr = err.Error()
		return false
	}
	var op ebiten.DrawImageOptions
	op.GeoM = anchorGeoM(e, d, border, g.cam.Effective(), sx, sy)
	op.Filter = ebiten.FilterLinear
	if remembered {
		op.ColorScale.Scale(0.6, 0.6, 0.65, 1)
	}
	dst.DrawImage(img, &op)
	return true
}

// loadUnitArt opens the unit catalog in dir, counting its fetches in m. The
// packs snapshot their bytes, so the directory is closed again at once.
func loadUnitArt(dir string, allowCandidates bool, m *meter) (*artPacks, error) {
	fsys, done, err := openArtDir(dir, m)
	if err != nil {
		return nil, err
	}
	defer done()
	return openUnitArt(fsys, assets.Options{AllowCandidates: allowCandidates})
}
