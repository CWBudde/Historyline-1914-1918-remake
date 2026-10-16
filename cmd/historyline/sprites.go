package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	state "github.com/CWBudde/Historyline-1914-1918-remake/internal/game"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/unitsprite"
)

// spriteDensities are the unit pack densities hlasset build-units writes.
var spriteDensities = []int{2, 4}

// spriteDensity is the unit pack drawn at effective zoom eff: the asset density,
// but at least 2x, since no 1x unit sprites exist. Drawing the 2x sprite smaller is
// the renderer's choice, not a loader fallback.
func spriteDensity(eff float64) int {
	return max(2, assetDensity(eff))
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
	family  string
	density int
}

type texKey struct {
	id      string
	density int
}

// unitArt is the runtime unit catalog: one pack per family and density, laid out
// as <d>x/<family>/manifest.json, plus the GPU textures made from it so far.
type unitArt struct {
	packs     map[packKey]*assets.Pack
	tex       map[texKey]*ebiten.Image
	candidate bool // some pack is a review candidate, not approved art
}

// openUnitArt opens the whole catalog and checks that every family has every
// sprite at every density, so drawing never meets a missing variant.
func openUnitArt(fsys fs.FS, opt assets.Options) (*unitArt, error) {
	a := &unitArt{packs: map[packKey]*assets.Pack{}, tex: map[texKey]*ebiten.Image{}}
	for _, d := range spriteDensities {
		for _, fam := range unitsprite.Families() {
			dir := path.Join(fmt.Sprintf("%dx", d), fam)
			sub, err := fs.Sub(fsys, dir)
			if err != nil {
				return nil, err
			}
			p, err := assets.Open(sub, "manifest.json", opt)
			if err != nil {
				return nil, fmt.Errorf("unit pack %s: %w", dir, err)
			}
			a.packs[packKey{fam, d}] = p
			if p.Manifest().Status != "approved" {
				a.candidate = true
			}
			for _, pal := range unitsprite.Palettes {
				for f := range unitsprite.Facings {
					for _, s := range unitsprite.States {
						if _, err := p.Entry(unitsprite.ID(fam, pal, f, s), d); err != nil {
							return nil, fmt.Errorf("unit pack %s: %w", dir, err)
						}
					}
				}
			}
		}
	}
	return a, nil
}

// entry is the catalog record of sprite id at density d.
func (a *unitArt) entry(family, id string, d int) (assets.Entry, error) {
	p, ok := a.packs[packKey{family, d}]
	if !ok {
		return assets.Entry{}, fmt.Errorf("no %dx unit pack for %s", d, family)
	}
	return p.Entry(id, d)
}

// image returns the texture of sprite id at density d, decoding and uploading it
// on first use.
func (a *unitArt) image(family, id string, d int) (*ebiten.Image, assets.Entry, error) {
	e, err := a.entry(family, id, d)
	if err != nil {
		return nil, e, err
	}
	k := texKey{id, d}
	if img, ok := a.tex[k]; ok {
		return img, e, nil
	}
	src, err := a.packs[packKey{family, d}].Image(id, d)
	if err != nil {
		return nil, e, err
	}
	img := ebiten.NewImageFromImage(src)
	a.tex[k] = img
	return img, e, nil
}

// drawUnitSprite draws a unit sprite with its logical anchor on screen point
// (sx, sy), the hex centre, and reports whether it could. Errors go to the status
// line once and the caller falls back to the placeholder dot.
func (g *game) drawUnitSprite(dst *ebiten.Image, sx, sy float64, typ string, owner state.Player, spent, remembered bool, face unitsprite.Facing) bool {
	if g.art == nil {
		return false
	}
	family, id, ok := unitSprite(typ, owner, spent, remembered, face)
	if !ok {
		g.artErr = fmt.Sprintf("no sprite family for unit type %s", typ)
		return false
	}
	d := spriteDensity(g.cam.Effective())
	img, e, err := g.art.image(family, id, d)
	if err != nil {
		g.artErr = err.Error()
		return false
	}
	scale := g.cam.Effective() / float64(d)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-e.Anchor[0]*float64(d), -e.Anchor[1]*float64(d))
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(sx, sy)
	op.Filter = ebiten.FilterLinear
	if remembered {
		op.ColorScale.Scale(0.6, 0.6, 0.65, 1)
	}
	dst.DrawImage(img, &op)
	return true
}

// loadUnitArt opens the unit catalog in dir. The packs snapshot their bytes, so
// the directory is closed again at once.
func loadUnitArt(dir string, allowCandidates bool) (*unitArt, error) {
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return openUnitArt(r.FS(), assets.Options{AllowCandidates: allowCandidates})
}
