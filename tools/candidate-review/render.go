package candidatereview

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/maprender"
	artreview "github.com/CWBudde/Historyline-1914-1918-remake/tools/art-review"
)

var slate = color.NRGBA{0x16, 0x26, 0x30, 255}

// unitAnchor is the logical anchor of every unit sprite (art/README.md, Geometry).
var unitAnchor = [2]float64{48, 49}

var naval = map[string]bool{"battleship": true, "destroyer": true, "patrol-boat": true, "submarine": true, "torpedo-boat": true, "transport-ship": true}

// stage is a composed map at 2x with a target hex whose ring is free of units.
type stage struct {
	img    *image.NRGBA
	target hexgrid.Coord
	window image.Rectangle // 2x pixels
}

type renderer struct {
	o      options
	work   string
	maps   *maprender.Renderer
	land   stage
	water  stage
	panel  *image.NRGBA
	assets []assetEntry
}

type assetEntry struct {
	ID      string `json:"id"`
	Density int    `json:"density"`
	Path    string `json:"path"`
	Art     string `json:"art"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

func newRenderer(o options) (*renderer, error) {
	r := &renderer{o: o, work: filepath.Join(filepath.Dir(o.out), "work")}
	var err error
	if r.maps, err = maprender.New(os.DirFS(filepath.Join(o.root, "assets"))); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(o.root, "assets", "index.json"))
	if err != nil {
		return nil, err
	}
	var index struct {
		Assets []assetEntry `json:"assets"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return nil, err
	}
	r.assets = index.Assets
	return r, nil
}

// prepare builds the shared backgrounds lazily: the PULSE map, a water field and the panel.
func (r *renderer) prepare() error {
	if r.land.img != nil {
		return nil
	}
	lv, err := level.Load(filepath.Join(r.o.root, "art", "reference", "pulse", "level.json"))
	if err != nil {
		return err
	}
	m, err := r.maps.Render(lv, 2, true)
	if err != nil {
		return err
	}
	r.land = stage{img: m.Image, target: pickTarget(lv)}
	r.land.window = window(r.land.target, m.Image.Bounds())
	water := &level.Level{MapName: "candidate-water", Lines: make([][]level.Hex, 9)}
	for l := range water.Lines {
		water.Lines[l] = make([]level.Hex, 7-l%2)
		for c := range water.Lines[l] {
			water.Lines[l][c].Terrain = "water"
		}
	}
	wm, err := r.maps.Render(water, 2, false)
	if err != nil {
		return err
	}
	r.water = stage{img: wm.Image, target: hexgrid.Coord{Line: 4, Col: 3}}
	r.water.window = window(r.water.target, wm.Image.Bounds())
	return nil
}

// pickTarget finds an empty grass hex whose ring holds no unit, with as many units
// as possible within three hexes, so a candidate is seen next to installed units.
func pickTarget(lv *level.Level) hexgrid.Coord {
	g := lv.Grid()
	w, h := maprender.Size(g)
	best, bestScore := hexgrid.Coord{}, -1
	for line, cells := range lv.Lines {
		for col, hx := range cells {
			at := hexgrid.Coord{Line: line, Col: col}
			if hx.Terrain != "grass" || hx.Unit != "" {
				continue
			}
			x, y := hexgrid.Center(at)
			if x < 300 || x > float64(w)-300 || y < 170 || y > float64(h)-170 {
				continue
			}
			ok := true
			for _, n := range hexgrid.Around(at) {
				if !g.Contains(n) || lv.Lines[n.Line][n.Col].Unit != "" {
					ok = false
				}
			}
			if !ok {
				continue
			}
			score := 0
			for l2, c2s := range lv.Lines {
				for c2, h2 := range c2s {
					if h2.Unit != "" && hexgrid.Distance(at, hexgrid.Coord{Line: l2, Col: c2}) <= 3 {
						score++
					}
				}
			}
			if score > bestScore {
				best, bestScore = at, score
			}
		}
	}
	return best
}

// window is a 568×300 logical view (four hexes wide) centred on the target, in 2x pixels.
func window(at hexgrid.Coord, bounds image.Rectangle) image.Rectangle {
	x, y := hexgrid.Center(at)
	r := image.Rect(int(x-284)*2, int(y-150)*2, int(x+284)*2, int(y+150)*2)
	return r.Intersect(bounds)
}

func (r *renderer) group(g *group) error {
	var err error
	switch g.Kind {
	case "unit":
		err = r.units(g)
	case "portrait":
		err = r.portraits(g)
	case "interface", "screen":
		err = r.ui(g)
	default:
		err = r.plain(g)
	}
	if err != nil {
		return err
	}
	// A context view makes the raw thumbnail redundant; drop it to keep the page small.
	for _, c := range g.Candidates {
		if len(c.Views) < 2 {
			continue
		}
		kept := c.Views[:0]
		for _, v := range c.Views {
			if strings.HasPrefix(v.Label, "raw ") {
				_ = os.Remove(filepath.Join(r.o.out, filepath.FromSlash(v.Media["img"])))
				continue
			}
			kept = append(kept, v)
		}
		c.Views = kept
	}
	return nil
}

// media saves img under media/<kind>/<group>/<name> and returns the site-relative path.
func (r *renderer) media(g *group, name string, img image.Image) (string, error) {
	rel := path.Join("media", g.Kind, slug(g.Family), name)
	return rel, save(filepath.Join(r.o.out, filepath.FromSlash(rel)), img)
}

var unsafe = regexp.MustCompile(`[^a-z0-9.-]+`)

func slug(s string) string {
	return strings.Trim(unsafe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// rawThumb adds a 512 px thumbnail of the raw candidate on a neutral background.
func (r *renderer) rawThumb(g *group, c *candidate) (image.Image, error) {
	src, err := loadPNG(c.raw)
	if err != nil {
		return nil, err
	}
	rel, err := r.media(g, c.Exec+"-raw.jpg", thumb(src, 512))
	if err != nil {
		return nil, err
	}
	c.Views = append(c.Views, &view{Label: "raw " + c.Size, Media: map[string]string{"img": rel}})
	return src, nil
}

func (r *renderer) plain(g *group) error {
	for _, c := range g.Candidates {
		if _, err := r.rawThumb(g, c); err != nil {
			return err
		}
	}
	return nil
}

// ---- units ----

type sprites map[string]image.Image // facing → cp normal 2x sprite

func loadSprites(dir string) (sprites, error) {
	s := sprites{}
	for _, f := range []string{"n", "nw", "sw", "s", "se", "ne"} {
		img, err := loadPNG(filepath.Join(dir, f+"-normal-2x.png"))
		if err != nil {
			return nil, err
		}
		s[f] = img
	}
	return s, nil
}

// unitViews composes the sprites on the stage: one at the target facing south-east
// next to the installed units, and the six facings on the ring, each facing outwards.
func (r *renderer) unitViews(g *group, name string, sp sprites) (*view, error) {
	st := r.land
	if naval[g.Family] {
		st = r.water
	}
	one := image.NewNRGBA(st.img.Bounds())
	draw.Draw(one, one.Bounds(), st.img, image.Point{}, draw.Src)
	maprender.Place(one, sp["se"], st.target, 2, unitAnchor)
	ring := image.NewNRGBA(st.img.Bounds())
	draw.Draw(ring, ring.Bounds(), st.img, image.Point{}, draw.Src)
	for i, n := range hexgrid.Around(st.target) {
		maprender.Place(ring, sp[[]string{"ne", "se", "s", "sw", "nw", "n"}[i]], n, 2, unitAnchor)
	}
	v := &view{Label: "on the map", Media: map[string]string{}}
	for _, item := range []struct {
		key string
		img *image.NRGBA
	}{{"map", one}, {"ring", ring}} {
		c2 := crop(item.img, st.window)
		rel2, err := r.media(g, name+"-"+item.key+"-2x.jpg", c2)
		if err != nil {
			return nil, err
		}
		v.Media[item.key] = rel2
	}
	return v, nil
}

func (r *renderer) units(g *group) error {
	if err := r.prepare(); err != nil {
		return err
	}
	installed, err := loadSprites(filepath.Join(r.o.root, "assets", "units", g.Family, "cp"))
	if err != nil {
		return err
	}
	if g.Installed, err = r.unitViews(g, "installed", installed); err != nil {
		return err
	}
	for _, c := range g.Candidates {
		if _, err := r.rawThumb(g, c); err != nil {
			return err
		}
		dir, verdict := r.exportUnit(g.Family, c)
		if verdict != "" {
			c.Verdict = verdict
			if c.Default == "" {
				c.Default, c.Reason = "delete", "the unit exporter rejects it: "+verdict
			}
			continue
		}
		sp, err := loadSprites(filepath.Join(dir, "cp"))
		if err != nil {
			return err
		}
		v, err := r.unitViews(g, c.Exec, sp)
		if err != nil {
			return err
		}
		c.Views = append([]*view{v}, c.Views...)
	}
	return nil
}

// exportUnit calls the art exporter on a copy of the family's manifest whose
// atlas points at the candidate. It returns the sprite directory
// <bundle>/sprites/<manifest family> without the palette, or a rejection.
func (r *renderer) exportUnit(family string, c *candidate) (string, string) {
	manifest := filepath.Join(r.o.root, "art", "pipeline", "unit-"+family+"-v1.json")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return "", "no unit manifest for " + family
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", err.Error()
	}
	sum, err := fileSHA(c.raw)
	if err != nil {
		return "", err.Error()
	}
	rel, err := filepath.Rel(r.o.root, c.raw)
	if err != nil {
		return "", err.Error()
	}
	m["atlas"] = map[string]any{"source": filepath.ToSlash(rel), "source_sha256": sum}
	dir := filepath.Join(r.work, c.Exec)
	out := filepath.Join(dir, "bundle")
	sprites := filepath.Join(out, "sprites", fmt.Sprint(m["family"]))
	if _, err := os.Stat(filepath.Join(sprites, "cp", "se-normal-2x.png")); err == nil {
		return sprites, ""
	}
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err.Error()
	}
	data, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o644); err != nil {
		return "", err.Error()
	}
	if err := artreview.Run([]string{"-unit", filepath.Join(dir, "manifest.json"), "-out", out, "-root", r.o.root}, io.Discard, io.Discard); err != nil {
		return "", err.Error()
	}
	return sprites, ""
}

func fileSHA(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
