package main

import (
	"fmt"
	"image"
	"io/fs"
	"slices"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/assets"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dospanel"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/dosscreen"
	"github.com/hajimehoshi/ebiten/v2"
)

// screenDensities are the densities of the screen catalog (hlasset
// build-screens): the paintings have no 4x.
var screenDensities = []int{1, 2}

// screenArt is the screen catalog: frames, controls and paintings, one pack per
// family and density, laid out as <d>x/<family>/manifest.json.
type screenArt struct {
	packs  *artPacks
	family map[string]string // asset ID -> family pack
}

// openScreenArt opens every family the 1x directory lists, at every density.
func openScreenArt(fsys fs.FS, opt assets.Options) (*screenArt, error) {
	dirs, err := fs.ReadDir(fsys, "1x")
	if err != nil {
		return nil, fmt.Errorf("screen catalog: %w", err)
	}
	var families []string
	for _, d := range dirs {
		if d.IsDir() {
			families = append(families, d.Name())
		}
	}
	p, err := openPacks(fsys, opt, families, screenDensities)
	if err != nil {
		return nil, fmt.Errorf("screen %w", err)
	}
	s := &screenArt{packs: p, family: map[string]string{}}
	for _, f := range families {
		for _, e := range p.packs[packKey{f, 1}].Manifest().Entries {
			s.family[e.ID] = f
		}
	}
	return s, nil
}

// loadScreenArt opens the screen catalog in dir, counting its fetches in m.
func loadScreenArt(dir string, allowCandidates bool, m *meter) (*screenArt, error) {
	fsys, done, err := openArtDir(dir, m)
	if err != nil {
		return nil, err
	}
	defer done()
	return openScreenArt(fsys, assets.Options{AllowCandidates: allowCandidates})
}

// image returns the texture of id at density d, nil when the catalog is
// missing or has no such image.
func (s *screenArt) image(id string, d int) *ebiten.Image {
	if s == nil {
		return nil
	}
	f, ok := s.family[id]
	if !ok {
		return nil
	}
	img, _, err := s.packs.image(f, id, d)
	if err != nil {
		return nil
	}
	return img
}

// surfaces are the panels' painted surfaces at the highest density the
// catalog has; none without a catalog.
func (s *screenArt) surfaces() dospanel.Surfaces {
	if s == nil {
		return dospanel.Surfaces{}
	}
	return dospanel.Surfaces{Steel: s.decoded("surface/steel"), Ground: s.decoded("surface/ground"), Rivets: s.decoded("surface/rivets")}
}

// frameArt is the split screen frame's paintings: the hex plates surface/plate,
// surface/plate-2 and on, and the chrome's steel; none without a catalog.
func (s *screenArt) frameArt() dosscreen.Art {
	if s == nil {
		return dosscreen.Art{}
	}
	a := dosscreen.Art{Steel: s.decoded("surface/steel")}
	for i := 1; ; i++ {
		id := "surface/plate"
		if i > 1 {
			id = fmt.Sprintf("surface/plate-%d", i)
		}
		img := s.decoded(id)
		if img == nil {
			return a
		}
		a.Plates = append(a.Plates, img)
	}
}

// decoded is the decoded image of id at the highest density the catalog has,
// nil when it has none.
func (s *screenArt) decoded(id string) image.Image {
	if s == nil {
		return nil
	}
	f, ok := s.family[id]
	if !ok {
		return nil
	}
	for _, d := range slices.Backward(screenDensities) {
		if p, ok := s.packs.packs[packKey{f, d}]; ok {
			if img, err := p.Image(id, d); err == nil {
				return img
			}
		}
	}
	return nil
}
