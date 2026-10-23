package maprender

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/mapcompose"
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

// Ground props lie on the grass, under the infrastructure of every hex: a road
// sprite reaching over the whole map hides each of them.
func TestGroundPropsDrawUnderInfrastructure(t *testing.T) {
	red, blue := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 255}
	render := func(road bool) *image.NRGBA {
		r := &Renderer{files: map[string]Entry{}, has: map[string]bool{}, cache: map[string]image.Image{}}
		add := func(id string, w, h int, c color.NRGBA) {
			r.has[id] = true
			im := image.NewNRGBA(image.Rect(0, 0, w*2, h*2))
			draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
			r.cache[id] = im
			r.files[fmt.Sprintf("2/%s", id)] = Entry{ID: id, Path: id, Anchor: [2]float64{float64(w) / 2, float64(h) / 2}}
		}
		add("tile/grass", 96, 80, color.NRGBA{A: 255})
		add("tile/road25", 2000, 2000, blue)
		for _, class := range []struct {
			name string
			n    int
		}{{"tuft", 4}, {"flowers", 3}, {"soil", 3}, {"straw", 2}, {"pebbles", 3}} {
			for i := 1; i <= class.n; i++ {
				add(mapcompose.PropID(class.name, i), 2, 2, red)
			}
		}
		lvl := &level.Level{}
		for range 8 {
			lvl.Lines = append(lvl.Lines, []level.Hex{{Terrain: "grass"}, {Terrain: "grass"}, {Terrain: "grass"}, {Terrain: "grass"}})
		}
		if road {
			lvl.Lines[7][0].Infra = level.Infra{IDs: []string{"road25"}}
		}
		m, err := r.Render(lvl, 2, false)
		if err != nil {
			t.Fatal(err)
		}
		return m.Image
	}
	count := func(im *image.NRGBA) int {
		n := 0
		for i := 0; i < len(im.Pix); i += 4 {
			if im.Pix[i] == 255 && im.Pix[i+2] == 0 {
				n++
			}
		}
		return n
	}
	if n := count(render(false)); n == 0 {
		t.Fatal("no ground props drawn")
	}
	if n := count(render(true)); n != 0 {
		t.Fatalf("%d prop pixels over the road", n)
	}
}
