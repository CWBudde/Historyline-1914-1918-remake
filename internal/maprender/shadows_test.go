package maprender

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// A southern shadow reaches a northern crown. Drawing shadows per cell, or
// before infrastructure, would cover the crown or hide the shadow respectively.
func TestAllShadowsDrawAfterInfrastructureBeforeAnyCrown(t *testing.T) {
	green, blue := color.NRGBA{G: 255, A: 255}, color.NRGBA{B: 255, A: 255}
	r := &Renderer{files: map[string]Entry{}, has: map[string]bool{}, cache: map[string]image.Image{}}
	for _, a := range []struct {
		id    string
		w, h  int
		color color.NRGBA
	}{
		{"tile/forest", 96, 80, color.NRGBA{A: 255}},
		{"tile/road25", 192, 160, blue},
		{"standing/forest", 2, 2, red},
		{"standing/shadow/forest", 192, 160, green},
	} {
		r.has[a.id] = true
		im := image.NewNRGBA(image.Rect(0, 0, a.w*2, a.h*2))
		draw.Draw(im, im.Bounds(), image.NewUniform(a.color), image.Point{}, draw.Src)
		r.cache[a.id] = im
		r.files[fmt.Sprintf("2/%s", a.id)] = Entry{ID: a.id, Path: a.id, Anchor: [2]float64{float64(a.w) / 2, float64(a.h) / 2}}
	}
	lvl := &level.Level{Lines: [][]level.Hex{
		{{Terrain: "forest"}, {Terrain: "forest"}},
		{{Terrain: "forest", Infra: level.Infra{IDs: []string{"road25"}}}},
	}}
	m, err := r.Render(lvl, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Image.NRGBAAt(48*2, 40*2); got != red {
		t.Errorf("northern crown %v want %v", got, red)
	}
	if got := m.Image.NRGBAAt(70*2, 40*2); got != green {
		t.Errorf("shadow over road %v want %v", got, green)
	}
}
