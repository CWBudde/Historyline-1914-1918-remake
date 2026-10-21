// Command reference-catalogue exports RefDesign2 as actual game-grid hex tiles.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const referenceHash = "b3182d5d39243717283d8684787281fff13f4f9f331c4ded3180c4e4cc7625dd"

//go:embed groups.json
var groupsJSON []byte

//go:embed labels.json
var labelsJSON []byte

//go:embed index.html
var page string

type picture struct {
	Image         string `json:"image"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	SHA256        string `json:"sha256"`
	MissingPixels int    `json:"missing_pixels,omitempty"`
}
type tile struct {
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	Family        string             `json:"family"`
	Notes         string             `json:"notes"`
	Row           int                `json:"row"`
	Column        int                `json:"column"`
	Terrain       string             `json:"terrain"`
	SourcePolygon [][2]float64       `json:"source_polygon"`
	Outputs       map[string]picture `json:"outputs"`
}
type group struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Family string   `json:"family"`
	Notes  string   `json:"notes"`
	Tiles  []string `json:"tiles"`
}
type catalogue struct {
	Source       string             `json:"source"`
	SHA256       string             `json:"sha256"`
	Width        int                `json:"width"`
	Height       int                `json:"height"`
	Method       string             `json:"method"`
	GridMethod   string             `json:"grid_method"`
	Tiles        []tile             `json:"tiles"`
	Curated      []group            `json:"curated"`
	MapOutputs   map[string]picture `json:"map_outputs"`
	AtlasOutputs map[string]picture `json:"atlas_outputs"`
	Archive      string             `json:"archive"`
}

func main() {
	source := flag.String("source", "RefDesign2.png", "user-supplied reference image")
	out := flag.String("out", "build/refdesign2-catalogue", "hex tile catalogue output")
	check := flag.Bool("check", false, "verify hex masks, source samples and decoded PNG assembly")
	flag.Parse()
	if err := run(*source, *out, *check); err != nil {
		log.Fatal(err)
	}
}

func run(source, out string, check bool) error {
	raw, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if hash(raw) != referenceHash {
		return fmt.Errorf("reference differs from reviewed RefDesign2: %s", hash(raw))
	}
	im, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if im.Bounds() != image.Rect(0, 0, 945, 1664) {
		return fmt.Errorf("unexpected source bounds %v", im.Bounds())
	}
	im = prepareSource(im)
	c := catalogue{Source: "reference.png", SHA256: referenceHash, Width: 945, Height: 1664, Archive: "hex-tiles.zip", MapOutputs: map[string]picture{}, AtlasOutputs: map[string]picture{},
		Method:     "Actual alpha-masked game hex tiles: 96x80, 192x160 and 384x320 PNGs. Only exact black source pixels connected to the canvas perimeter are removed as exterior backdrop; enclosed black details remain. One global source registration uses premultiplied bilinear sampling followed by complementary hard polygon ownership. 2x/4x enlarge reference pixels; they add no new painted detail. No per-feature crop fitting, rotation or inpainting. Assembled maps use decoded exported tile PNGs.",
		GridMethod: "Canonical game origins x=142*column+71*(row%2), y=40*row (zero-based). Polygon=(25,0),(71,0),(96,40),(71,80),(25,80),(0,40). Inferred source registration: source_x=world_x*123/142+42-48*123/142; source_y=world_y*35/40. Catalogue IDs are one-based. Painted features can drift from inferred cells. Missing bottom source pixels are transparent. The final eight red boundary cells are separate from the 345 landscape tiles.",
	}
	if err := json.Unmarshal(groupsJSON, &c.Curated); err != nil {
		return err
	}
	var labels []string
	if err := json.Unmarshal(labelsJSON, &labels); err != nil {
		return err
	}
	if err := populateTiles(&c, labels); err != nil {
		return err
	}
	if !check {
		if err := os.MkdirAll(out, 0755); err != nil {
			return err
		}
	}
	for _, d := range []int{1, 2, 4} {
		if err := exportDensity(im, out, &c, d, check); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	for name, contents := range map[string][]byte{"catalogue.json": append(b, '\n'), "index.html": []byte(strings.Replace(page, "/*CATALOGUE*/null", string(b), 1)), "reference.png": raw} {
		if err := writeOrCheck(filepath.Join(out, name), contents, check); err != nil {
			return err
		}
	}
	if err := archiveTiles(out, c, check); err != nil {
		return err
	}
	verb := "Exported"
	if check {
		verb = "Verified"
	}
	fmt.Printf("%s %d hex tiles at 1x/2x/4x, three atlases and three decoded-PNG map assemblies: %s\n", verb, len(c.Tiles), filepath.Join(out, "index.html"))
	return nil
}

func populateTiles(c *catalogue, labels []string) error {
	families := map[byte]string{'G': "Meadow", 'F': "Woodland", 'H': "Hills & rock", 'M': "Earthworks", 'S': "Structures", 'R': "Roads", 'P': "Dirt tracks", 'W': "Water & shore", 'B': "Bridges & crossings", 'U': "Boundary symbols"}
	if len(labels) != 47 {
		return fmt.Errorf("need 47 classification rows, got %d", len(labels))
	}
	known := map[string]bool{}
	for row, line := range labels {
		line = strings.ReplaceAll(line, " ", "")
		if len(line) != 8-row%2 {
			return fmt.Errorf("classification row %d has %d entries", row+1, len(line))
		}
		for col := range len(line) {
			family, ok := families[line[col]]
			if !ok {
				return fmt.Errorf("unknown classification %c", line[col])
			}
			t := tile{ID: fmt.Sprintf("r%02dc%02d", row+1, col+1), Title: fmt.Sprintf("%s · %02d/%02d", family, row+1, col+1), Family: family, Terrain: family, Row: row + 1, Column: col + 1, Outputs: map[string]picture{}, Notes: "Reference tile at its inferred grid position. Context and neighbouring feature fragments are retained. Family labels describe visible content rather than verified original-game terrain IDs."}
			ox, oy := tileOrigin(row, col)
			for _, p := range hexPolygon {
				sx, sy := sourcePoint(float64(ox)+p[0], float64(oy)+p[1])
				t.SourcePolygon = append(t.SourcePolygon, [2]float64{sx, sy})
			}
			if row == 46 {
				t.Notes = "Partial bottom boundary tile with red interface artwork. Unavailable source pixels below the image are transparent."
			}
			known[t.ID] = true
			c.Tiles = append(c.Tiles, t)
		}
	}
	for _, g := range c.Curated {
		for _, id := range g.Tiles {
			if !known[id] {
				return fmt.Errorf("group %s references unknown tile %s", g.ID, id)
			}
		}
	}
	return nil
}

func exportDensity(src image.Image, out string, c *catalogue, d int, check bool) error {
	key := strconv.Itoa(d)
	if !check {
		if err := os.MkdirAll(filepath.Join(out, "tiles", key+"x"), 0755); err != nil {
			return err
		}
	}
	assembled := image.NewNRGBA(image.Rect(0, 0, 1090*d, 1920*d))
	atlas := image.NewNRGBA(image.Rect(0, 0, 16*100*d, 23*100*d))
	for i := range c.Tiles {
		t := &c.Tiles[i]
		ox, oy := tileOrigin(t.Row-1, t.Column-1)
		im, missing := renderTile(src, ox, oy, d)
		p, decoded, err := savePNG(out, "tiles/"+key+"x/"+t.ID+".png", im, check)
		if err != nil {
			return err
		}
		p.MissingPixels = missing
		t.Outputs[key] = p
		// These are decoded PNG bytes placed at canonical integer grid offsets.
		for y := range 80 * d {
			for x := range 96 * d {
				if insideTile(x, y, d) {
					// Copy owned straight-RGBA pixels, including source-edge alpha,
					// without a premultiply/unpremultiply rounding round trip.
					assembled.Set(ox*d+x, oy*d+y, decoded.At(x, y))
				}
			}
		}
		ax, ay := ((i%16)*100+2)*d, ((i/16)*100+2)*d
		draw.Draw(atlas, image.Rect(ax, ay, ax+96*d, ay+80*d), decoded, image.Point{}, draw.Over)
		label := image.NewNRGBA(image.Rect(0, 0, 96, 16))
		fd := font.Drawer{Dst: label, Src: image.White, Face: basicfont.Face7x13, Dot: fixed.P(1, 12)}
		fd.DrawString(t.ID)
		for y := range 16 * d {
			for x := range 96 * d {
				atlas.SetNRGBA(ax+x, ay+81*d+y, label.NRGBAAt(x/d, y/d))
			}
		}
	}
	if check {
		if err := verifyAssembly(src, assembled, d); err != nil {
			return err
		}
	}
	var err error
	c.MapOutputs[key], _, err = savePNG(out, "assembled-"+key+"x.png", assembled, check)
	if err != nil {
		return err
	}
	c.AtlasOutputs[key], _, err = savePNG(out, "atlas-"+key+"x.png", atlas, check)
	return err
}

func savePNG(out, name string, im *image.NRGBA, check bool) (picture, image.Image, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		return picture{}, nil, err
	}
	if err := writeOrCheck(filepath.Join(out, filepath.FromSlash(name)), b.Bytes(), check); err != nil {
		return picture{}, nil, err
	}
	raw, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
	if err != nil {
		return picture{}, nil, err
	}
	decoded, err := png.Decode(bytes.NewReader(raw))
	return picture{Image: name, Width: im.Bounds().Dx(), Height: im.Bounds().Dy(), SHA256: hash(raw)}, decoded, err
}
func hash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func writeOrCheck(path string, contents []byte, check bool) error {
	if check {
		got, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Equal(got, contents) {
			return fmt.Errorf("%s differs; regenerate hex catalogue", path)
		}
		return nil
	}
	return os.WriteFile(path, contents, 0644)
}
func archiveTiles(out string, c catalogue, check bool) error {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	paths := []string{"catalogue.json"}
	for _, t := range c.Tiles {
		for _, d := range []string{"1", "2", "4"} {
			paths = append(paths, t.Outputs[d].Image)
		}
	}
	for _, path := range paths {
		f, err := w.CreateHeader(&zip.FileHeader{Name: path, Method: zip.Store})
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(path)))
		if err != nil {
			return err
		}
		if _, err := f.Write(raw); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return writeOrCheck(filepath.Join(out, c.Archive), b.Bytes(), check)
}
