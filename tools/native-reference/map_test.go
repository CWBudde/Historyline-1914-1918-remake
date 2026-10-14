package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Reversing library FAT order must not reverse FIN indices: PARTS.DAT owns them.
func mapInputFixture(t *testing.T, root string) string {
	t.Helper()
	data := filepath.Join(root, "original", "input")
	for _, d := range []string{"LIB", "MAP"} {
		if err := os.MkdirAll(filepath.Join(data, d), 0700); err != nil {
			t.Fatal(err)
		}
	}
	m, inv := maskFixture()
	var names []byte
	lib := make([]byte, 4)
	var fat []byte
	for i := 174; i >= 0; i-- {
		name := fmt.Sprintf("SYNTH%03d", i)
		entry := append([]byte(name), 0, 0, 0, 0)
		binary.BigEndian.PutUint32(entry[8:], uint32(len(lib)))
		fat = append(fat, entry...)
		plane := make([]byte, 96)
		if i == 1 {
			for j := range plane {
				plane[j] = 255
			}
		}
		lib = append(lib, graphicFixture(24, 24, 1, nil, plane)...)
	}
	binary.BigEndian.PutUint32(lib, uint32(len(lib)))
	lib = append(lib, fat...)
	for i := 0; i < 175; i++ {
		names = append(names, fmt.Sprintf("SYNTH%03d", i)...)
	}
	files := map[string][]byte{"00.PAL": paletteFixture(), "HEXMSK00.DAT": m, "HEXMSK20.DAT": inv, "LIB/PARTS.LIB": lib, "LIB/PARTS.DAT": names, "MAP/00.FIN": {0, 3, 0, 2, 0, 255, 1, 42, 0, 255, 1, 7, 0, 255, 1, 255}}
	for n, b := range files {
		if err := os.WriteFile(filepath.Join(data, n), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return data
}

func TestMapCLIUsesIndexTableRowOrderAndColumnStagger(t *testing.T) {
	root := testRoot(t)
	t.Chdir(root)
	data := mapInputFixture(t, root)
	args := []string{"-data", data, "-libraries", "PARTS", "-map", "00"}
	if err := run(args, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "build", "native-reference")
	f, err := os.Open(filepath.Join(out, "maps", "00-PARTS.png"))
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 56 || im.Bounds().Dy() != 60 {
		t.Fatalf("map bounds=%v want56x60", im.Bounds())
	}
	pal, _ := parsePalette(paletteFixture())
	// Hand-picked tile centres for all six cells: row1 starts at source pair3.
	for _, p := range []struct{ x, y, index int }{{12, 12, 0}, {28, 24, 1}, {44, 12, 0}, {12, 36, 1}, {28, 48, 0}, {44, 36, 1}} {
		if got := color.NRGBAModel.Convert(im.At(p.x, p.y)); got != pal[p.index] {
			t.Errorf("(%d,%d)=%v want%v", p.x, p.y, got, pal[p.index])
		}
	}
	_, _, _, a := im.At(0, 0).RGBA()
	if a != 0 {
		t.Fatal("rectangular tile pasted over coverage")
	}
	covered := 0
	for y := 0; y < 60; y++ {
		for x := 0; x < 56; x++ {
			_, _, _, alpha := im.At(x, y).RGBA()
			if alpha != 0 {
				covered++
			}
		}
	}
	if covered != 2304 {
		t.Fatalf("six disjoint native masks cover %d pixels, want 2304", covered)
	}
	b, err := os.ReadFile(filepath.Join(out, "maps", "00-PARTS-cells.csv"))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 7 || strings.Join(rows[2], ",") != "1,0,6,1,SYNTH001,42,16,12" || strings.Join(rows[4], ",") != "0,1,10,1,SYNTH001,7,0,24" {
		t.Fatalf("map cell provenance %v", rows)
	}
	gallery, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(gallery, []byte("maps/00-PARTS-cells.csv")) || !bytes.Contains(gallery, []byte("ground reconstruction")) {
		t.Fatal("gallery lacks reconstruction label/cells link")
	}
	if err := run(args, &bytes.Buffer{}); err == nil {
		t.Fatal("existing output overwritten")
	}
}

func TestMapCLIRejectsInvalidDataBeforeOutput(t *testing.T) {
	// The original loader consumes only the low byte of each dimension word.
	highDimension := make([]byte, 516)
	highDimension[0], highDimension[3] = 1, 1 // BE256 columns, one row.
	for _, tt := range []struct {
		name, file string
		data       []byte
		flags      []string
	}{
		{"truncated", "MAP/00.FIN", []byte{0, 3, 0, 2, 0, 255}, nil},
		{"zero dimension", "MAP/00.FIN", []byte{0, 0, 0, 0}, nil},
		{"out of range terrain", "MAP/00.FIN", []byte{0, 1, 0, 1, 175, 255}, nil},
		{"excess pixels", "MAP/00.FIN", []byte{255, 255, 255, 255}, nil},
		{"dimension exceeds native byte", "MAP/00.FIN", highDimension, nil},
		{"short index", "LIB/PARTS.DAT", []byte("SYNTH000"), nil},
		{"missing named resource", "LIB/PARTS.DAT", bytes.Repeat([]byte("MISSING0"), 175), nil},
		{"path escape", "", nil, []string{"-map", "../00"}},
		{"unsupported season", "", nil, []string{"-map-library", "UNIT"}},
		{"unselected library", "", nil, []string{"-map-library", "PARTW"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := testRoot(t)
			t.Chdir(root)
			data := mapInputFixture(t, root)
			if tt.file != "" {
				if err := os.WriteFile(filepath.Join(data, tt.file), tt.data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := append([]string{"-data", data, "-libraries", "PARTS", "-map", "00"}, tt.flags...)
			if err := run(args, &bytes.Buffer{}); err == nil {
				t.Fatal("invalid map accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "build", "native-reference")); !os.IsNotExist(err) {
				t.Fatal("invalid input created output")
			}
		})
	}
}
