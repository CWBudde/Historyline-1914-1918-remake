// Command reference-catalogue dissects the user-supplied RefDesign2 study.
// It copies pixels; it does not generate or replace production artwork.
package main

import (
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
	"math"
	"os"
	"path/filepath"
	"strings"
)

const referenceHash = "b3182d5d39243717283d8684787281fff13f4f9f331c4ded3180c4e4cc7625dd"

//go:embed selections.json
var selections []byte

//go:embed index.html
var page string

type specimen struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Family string `json:"family"`
	Box    [4]int `json:"box"` // x, y, width, height in reference pixels
	Notes  string `json:"notes"`
	Image  string `json:"image,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
	Row    int    `json:"row,omitempty"` // one-based inferred reference grid
	Column int    `json:"column,omitempty"`
}

type catalogue struct {
	Source     string     `json:"source"`
	SHA256     string     `json:"sha256"`
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	Method     string     `json:"method"`
	GridMethod string     `json:"grid_method"`
	Curated    []specimen `json:"curated"`
	Cells      []specimen `json:"cells"`
}

func main() {
	source := flag.String("source", "RefDesign2.png", "user-supplied reference image (not committed)")
	out := flag.String("out", "build/refdesign2-catalogue", "local study output directory")
	check := flag.Bool("check", false, "verify the existing catalogue and every extracted pixel")
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
		return fmt.Errorf("reference SHA-256 differs from the reviewed RefDesign2: %s", hash(raw))
	}
	im, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if im.Bounds() != image.Rect(0, 0, 945, 1664) {
		return fmt.Errorf("unexpected reference bounds %v", im.Bounds())
	}
	c := catalogue{
		Source: "reference.png", SHA256: referenceHash, Width: 945, Height: 1664,
		Method:     "Unaltered rectangular source-pixel crops. No inpainting, reconstruction, upscaling or deduplication. Context and neighbouring features remain in each crop.",
		GridMethod: "Inferred reference grid, not recovered source tiles or verified PULSE coordinates: centre=(42+123*column+61.5*(row%2),35+35*row), nominal footprint=84x70, 47 alternating rows of 8/7. Indices in the catalogue are one-based. Last row contains red boundary symbols. The optional hex outline is a viewing guide only; PNGs remain rectangular and unmasked.",
	}
	if err := json.Unmarshal(selections, &c.Curated); err != nil {
		return err
	}
	for row := range 47 {
		for col := range 8 - row%2 {
			cx := 42 + 123*float64(col) + 61.5*float64(row%2)
			cy := 35 + 35*row
			x, y := int(math.Floor(cx-42)), cy-35
			right, bottom := min(945, int(math.Ceil(cx+42))), min(1664, cy+35)
			family := "Reference cells"
			if row == 46 {
				family = "Boundary symbols"
			}
			c.Cells = append(c.Cells, specimen{
				ID:     fmt.Sprintf("r%02dc%02d", row+1, col+1),
				Title:  fmt.Sprintf("Reference row %02d / cell %02d", row+1, col+1),
				Family: family, Box: [4]int{x, y, right - x, bottom - y}, Row: row + 1, Column: col + 1,
				Notes: "Approximate grid window from the supplied painting. Contains neighbouring context; labels do not identify original game terrain or recover a reusable tile.",
			})
		}
	}
	if !check {
		if err := os.MkdirAll(filepath.Join(out, "crops"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, c.Source), raw, 0644); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, group := range [][]specimen{c.Curated, c.Cells} {
		for i := range group {
			s := &group[i]
			if seen[s.ID] || s.ID == "" || strings.ContainsAny(s.ID, "/\\.") {
				return fmt.Errorf("invalid or duplicate specimen ID %q", s.ID)
			}
			seen[s.ID] = true
			if err := extract(im, out, s, check); err != nil {
				return err
			}
		}
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	html := []byte(strings.Replace(page, "/*CATALOGUE*/null", string(b), 1))
	for name, contents := range map[string][]byte{"catalogue.json": append(b, '\n'), "index.html": html, c.Source: raw} {
		path := filepath.Join(out, name)
		if check {
			got, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !bytes.Equal(got, contents) {
				return fmt.Errorf("%s differs; regenerate the catalogue", path)
			}
		} else if err := os.WriteFile(path, contents, 0644); err != nil {
			return err
		}
	}
	verb := "Wrote"
	if check {
		verb = "Verified every source pixel in"
	}
	fmt.Printf("%s %d curated crops and %d reference cells: %s\n", verb, len(c.Curated), len(c.Cells), filepath.Join(out, "index.html"))
	return nil
}

func hash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func extract(src image.Image, out string, s *specimen, check bool) error {
	b := s.Box
	r := image.Rect(b[0], b[1], b[0]+b[2], b[1]+b[3])
	if r.Empty() || !r.In(src.Bounds()) {
		return fmt.Errorf("%s: crop %v outside reference", s.ID, r)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, b[2], b[3]))
	draw.Draw(dst, dst.Bounds(), src, r.Min, draw.Src)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, dst); err != nil {
		return err
	}
	s.Image = "crops/" + s.ID + ".png"
	s.SHA256 = hash(encoded.Bytes())
	path := filepath.Join(out, filepath.FromSlash(s.Image))
	if !check {
		return os.WriteFile(path, encoded.Bytes(), 0644)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if hash(raw) != s.SHA256 {
		return fmt.Errorf("%s: crop hash differs", s.ID)
	}
	decoded, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if decoded.Bounds() != dst.Bounds() {
		return fmt.Errorf("%s: crop dimensions differ", s.ID)
	}
	for y := range b[3] {
		for x := range b[2] {
			a, bb, cc, d := src.At(x+b[0], y+b[1]).RGBA()
			e, f, g, h := decoded.At(x, y).RGBA()
			if a != e || bb != f || cc != g || d != h {
				return fmt.Errorf("%s: pixel %d,%d changed", s.ID, x, y)
			}
		}
	}
	return nil
}
