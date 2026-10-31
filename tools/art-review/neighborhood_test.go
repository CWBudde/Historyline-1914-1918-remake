package artreview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The literal placements catch wrong staggering, cropping, density selection,
// and replacement of just the center rather than all matching resources.
func TestNeighborhoodRendersBothParities(t *testing.T) {
	for _, odd := range []bool{false, true} {
		t.Run(fmt.Sprint(odd), func(t *testing.T) {
			dir, config, positions := neighborhoodFixture(t, odd)
			out := filepath.Join(dir, "review")
			if err := run([]string{"-neighborhood", config, "-out", out}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			for _, density := range []int{1, 2, 4} {
				for _, mode := range []string{"original", "mixed"} {
					got, err := loadPNG(filepath.Join(out, "scene", fmt.Sprintf("%s-%dx.png", mode, density)))
					if err != nil {
						t.Fatal(err)
					}
					if got.Bounds() != image.Rect(0, 0, 56*density, 72*density) {
						t.Fatalf("bounds %v", got.Bounds())
					}
					want := image.NewNRGBA(got.Bounds())
					for i, p := range positions {
						c := color.NRGBA{R: 40, A: 255}
						if i%2 == 0 {
							c.R = 100
							if mode == "mixed" {
								c.R = 220
							}
						}
						for y := 0; y < 24*density; y++ {
							for x := 0; x < 24*density; x++ {
								if fixtureHex(x/density, y/density) {
									want.SetNRGBA(p.X*density+x, p.Y*density+y, c)
								}
							}
						}
					}
					if !bytes.Equal(got.Pix, want.Pix) {
						t.Fatalf("%s density %d differs from literal seven-cell placement", mode, density)
					}
				}
			}
			raw, err := os.ReadFile(filepath.Join(out, "diagnostics.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report struct {
				Inputs map[string]string `json:"input_sha256"`
				Scenes []struct {
					Cells     []any          `json:"cells"`
					Replaced  map[string]int `json:"replaced_resources"`
					Densities []struct {
						Coverage coverage `json:"coverage"`
						Contacts []struct {
							Pairs           int `json:"pair_count"`
							Original, Mixed *float64
						} `json:"contacts"`
					} `json:"densities"`
				} `json:"scenes"`
			}
			if err := json.Unmarshal(raw, &report); err != nil {
				t.Fatal(err)
			}
			if len(report.Inputs) != 6 || len(report.Scenes) != 1 || len(report.Scenes[0].Cells) != 7 || report.Scenes[0].Replaced["road0005"] != 4 || len(report.Scenes[0].Densities) != 3 {
				t.Fatalf("incomplete report: %s", raw)
			}
			for _, d := range report.Scenes[0].Densities {
				if d.Coverage.MissingPixels != 0 || d.Coverage.ExtraPixels != 0 || len(d.Contacts) != 6 {
					t.Fatalf("coverage/contacts: %s", raw)
				}
				for i, c := range d.Contacts {
					if c.Pairs == 0 || c.Original == nil || c.Mixed == nil {
						t.Fatalf("missing boundary measurement: %s", raw)
					}
					wantOriginal, wantMixed := 0.0, 0.0
					if i%2 == 0 {
						wantOriginal, wantMixed = 60.0/(3*255), 180.0/(3*255)
					}
					if *c.Original != wantOriginal || *c.Mixed != wantMixed {
						t.Fatalf("contact %d original=%v mixed=%v; want %v/%v", i, *c.Original, *c.Mixed, wantOriginal, wantMixed)
					}
				}
			}
			for p, h := range report.Inputs {
				if !filepath.IsAbs(p) || len(h) != 64 {
					t.Fatalf("bad input digest %q %q", p, h)
				}
			}
			html, err := os.ReadFile(filepath.Join(out, "index.html"))
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range []string{"Original Amiga", "Mixed original/redraw", "diagnostics.json", "mixed-4x.png"} {
				if !strings.Contains(string(html), s) {
					t.Fatalf("gallery missing %s", s)
				}
			}
			if err := run([]string{"-neighborhood", config, "-out", out}, &bytes.Buffer{}); err == nil {
				t.Fatal("overwrote prior review")
			}
		})
	}
}

func TestNeighborhoodNativeUsesBoxDownsampling(t *testing.T) {
	dir, config, _ := neighborhoodFixture(t, false)
	for _, scale := range []int{2, 4} {
		name := fmt.Sprintf("candidate-%dx.png", scale)
		img, err := loadPNG(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		for i, r := range []uint8{0, 100, 200, 252} {
			for y := 0; y < scale/2; y++ {
				for x := 0; x < scale/2; x++ {
					img.SetNRGBA(12*scale+(i%2)*(scale/2)+x, 12*scale+(i/2)*(scale/2)+y, color.NRGBA{R: r, A: 255})
				}
			}
		}
		writeFixture(t, dir, name, img)
	}
	out := filepath.Join(dir, "box-review")
	if err := run([]string{"-neighborhood", config, "-out", out}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	img, err := loadPNG(filepath.Join(out, "scene", "mixed-1x.png"))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.NRGBAAt(28, 36); got != (color.NRGBA{R: 138, A: 255}) {
		t.Fatalf("native candidate uses wrong density reduction: %+v", got)
	}
}

func TestNeighborhoodResolvesRelativeConfigPath(t *testing.T) {
	dir, config, _ := neighborhoodFixture(t, false)
	// A nested relative path is required: ../ paths to /tmp can accidentally
	// normalize the duplicated prefix away. Chdir is scoped and not parallel.
	t.Chdir(filepath.Dir(dir))
	relative := filepath.Join(filepath.Base(dir), filepath.Base(config))
	if err := run([]string{"-neighborhood", relative, "-out", filepath.Join(dir, "relative-review")}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
}

func TestNeighborhoodRejectsInputsBeforeWriting(t *testing.T) {
	for _, kind := range []string{"missing neighbor", "duplicate cell", "wrong position", "invalid integer", "unsafe resource", "alpha hole", "alpha outside", "density mismatch", "hidden density mismatch", "wrong dimensions", "empty scenes", "mixed flags", "bad json", "nonempty output"} {
		t.Run(kind, func(t *testing.T) {
			dir, config, _ := neighborhoodFixture(t, false)
			csvPath := filepath.Join(dir, "cells.csv")
			raw, _ := os.ReadFile(csvPath)
			lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
			switch kind {
			case "missing neighbor":
				lines = lines[:len(lines)-1]
			case "duplicate cell":
				lines = append(lines, lines[1])
			case "wrong position":
				fields := strings.Split(lines[1], ",")
				fields[6] = "33"
				lines[1] = strings.Join(fields, ",")
			case "invalid integer":
				fields := strings.Split(lines[1], ",")
				fields[0] = "bad"
				lines[1] = strings.Join(fields, ",")
			case "unsafe resource":
				lines[1] = strings.Replace(lines[1], "road0005", "../road5", 1)
			case "alpha hole", "alpha outside", "density mismatch", "hidden density mismatch", "wrong dimensions":
				p := filepath.Join(dir, "candidate-2x.png")
				img, err := loadPNG(p)
				if err != nil {
					t.Fatal(err)
				}
				switch kind {
				case "alpha hole":
					img.SetNRGBA(24, 24, color.NRGBA{})
				case "alpha outside":
					img.SetNRGBA(0, 0, color.NRGBA{A: 255})
				case "density mismatch":
					img.SetNRGBA(24, 24, color.NRGBA{R: 219, A: 255})
				case "hidden density mismatch":
					img.SetNRGBA(0, 0, color.NRGBA{R: 100})
				case "wrong dimensions":
					img = image.NewNRGBA(image.Rect(0, 0, 24, 24))
				}
				writeFixture(t, dir, "candidate-2x.png", img)
			case "empty scenes":
				data, _ := os.ReadFile(config)
				var c map[string]any
				if err := json.Unmarshal(data, &c); err != nil {
					t.Fatal(err)
				}
				c["scenes"] = []any{}
				data, _ = json.Marshal(c)
				if err := os.WriteFile(config, data, 0o600); err != nil {
					t.Fatal(err)
				}
			case "bad json":
				if err := os.WriteFile(config, []byte(`{"bogus":`), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(csvPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, "review")
			args := []string{"-neighborhood", config, "-out", out}
			if kind == "mixed flags" {
				args = append(args, "-reference", "unused.png")
			}
			if kind == "nonempty output" {
				if err := os.Mkdir(out, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(out, "keep"), []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			err := run(args, &bytes.Buffer{})
			if err == nil {
				t.Fatal("invalid inputs accepted")
			}
			if kind == "nonempty output" {
				entries, _ := os.ReadDir(out)
				if len(entries) != 1 {
					t.Fatal("wrote into nonempty output")
				}
			} else if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("created output before validation")
			}
		})
	}
}

func fixtureHex(x, y int) bool {
	insets := [24]int{8, 7, 6, 6, 5, 4, 4, 3, 2, 2, 1, 0, 0, 1, 2, 2, 3, 4, 4, 5, 6, 6, 7, 8}
	return x >= insets[y] && x < 24-insets[y]
}

func neighborhoodFixture(t *testing.T, odd bool) (string, string, []image.Point) {
	t.Helper()
	dir := t.TempDir()
	cells := [][2]int{{2, 2}, {2, 1}, {3, 1}, {3, 2}, {2, 3}, {1, 2}, {1, 1}}
	positions := []image.Point{{16, 24}, {16, 0}, {32, 12}, {32, 36}, {16, 48}, {0, 36}, {0, 12}}
	if odd {
		cells = [][2]int{{3, 2}, {3, 1}, {4, 2}, {4, 3}, {3, 3}, {2, 3}, {2, 2}}
		positions = []image.Point{{16, 24}, {16, 0}, {32, 12}, {32, 36}, {16, 48}, {0, 36}, {0, 12}}
	}
	var csv strings.Builder
	csv.WriteString("column,row,fin_offset,terrain_index,resource,second_byte_raw,pixel_x,pixel_y\n")
	for i, p := range cells {
		resource := "plain000"
		if i%2 == 0 {
			resource = "road0005"
		}
		fmt.Fprintf(&csv, "%d,%d,%d,0,%s,0,%d,%d\n", p[0], p[1], 2*i, resource, 16*p[0], 24*p[1]+12*(p[0]%2))
	}
	if err := os.WriteFile(filepath.Join(dir, "cells.csv"), []byte(csv.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"road0005", "plain000", "candidate-2x", "candidate-4x"} {
		n, r := 1, uint8(40)
		if name == "road0005" {
			r = 100
		}
		if name == "candidate-2x" {
			n, r = 2, 220
		}
		if name == "candidate-4x" {
			n, r = 4, 220
		}
		img := image.NewNRGBA(image.Rect(0, 0, 24*n, 24*n))
		for y := 0; y < 24*n; y++ {
			for x := 0; x < 24*n; x++ {
				if fixtureHex(x/n, y/n) {
					img.SetNRGBA(x, y, color.NRGBA{R: r, A: 255})
				}
			}
		}
		writeFixture(t, dir, name+".png", img)
	}
	col := 2
	if odd {
		col = 3
	}
	config := fmt.Sprintf(`{"tiles_directory":".","maps":[{"id":"map00","cells_csv":"cells.csv"}],"replacements":[{"resource":"road0005","image_2x":"candidate-2x.png","image_4x":"candidate-4x.png"}],"scenes":[{"id":"scene","map":"map00","column":%d,"row":2}]}`, col)
	path := filepath.Join(dir, "recipe.json")
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, path, positions
}
