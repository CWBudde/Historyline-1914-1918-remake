package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
	"github.com/CWBudde/Historyline-1914-1918-remake/internal/level"
)

// Catch odd-line shifts and anchors drifting between authored densities.
func TestPulsePlacement(t *testing.T) {
	for _, density := range []int{1, 2, 4} {
		src := image.NewNRGBA(image.Rect(0, 0, 96*density, 80*density))
		src.SetNRGBA(48*density, 40*density, color.NRGBA{R: 255, A: 255})
		dst := image.NewNRGBA(image.Rect(0, 0, 300*density, 200*density))
		placePulse(dst, src, hexgrid.Coord{Line: 1, Col: 0}, density, [2]float64{48, 40})
		if got := dst.NRGBAAt(119*density, 80*density); got.R != 255 {
			t.Errorf("%dx odd-line anchor: %+v", density, got)
		}
		clear(dst.Pix)
		placePulse(dst, src, hexgrid.Coord{Line: 1, Col: 0}, density, [2]float64{48, 49})
		if got := dst.NRGBAAt(119*density, 71*density); got.R != 255 {
			t.Errorf("%dx unit anchor: %+v", density, got)
		}
	}
}

func TestPulseDOSRequiresPinnedDOSMap(t *testing.T) {
	for _, tc := range []struct {
		name               string
		platform           string
		badCells, badImage bool
		wantErr            bool
	}{
		{"valid", "dos", false, false, false}, {"wrong platform", "amiga", false, false, true}, {"changed landmarks", "dos", true, false, true}, {"changed image", "dos", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = r.Close() }()
			for _, dir := range []string{"media", "build/dos/maps"} {
				if err := r.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
			}
			var cells bytes.Buffer
			rows := [][]string{{"column", "row", "offset", "index", "resource", "second", "x", "y"}, {"1", "2", "0", "1", "SYNTH000", "255", "0", "0"}, {"0", "3", "0", "5", "SYNTH000", "255", "0", "0"}, {"2", "3", "0", "7", "SYNTH000", "255", "0", "0"}, {"1", "3", "0", "3", "SYNTH000", "255", "0", "0"}, {"0", "4", "0", "6", "SYNTH000", "255", "0", "0"}, {"2", "4", "0", "8", "SYNTH000", "255", "0", "0"}, {"1", "4", "0", "4", "SYNTH000", "255", "0", "0"}}
			if err := csv.NewWriter(&cells).WriteAll(rows); err != nil {
				t.Fatal(err)
			}
			if err := r.WriteFile("build/dos/maps/00-PARTS-cells.csv", cells.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			pictures := []map[string]any{}
			for i, size := range [][2]int{{264, 588}, {1090, 1920}} {
				var buf bytes.Buffer
				if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, size[0], size[1]))); err != nil {
					t.Fatal(err)
				}
				name := fmt.Sprintf("maps/picture-%d.png", i)
				if err := r.WriteFile("build/dos/"+name, buf.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				pin := fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
				if tc.badImage && i == 0 {
					pin = strings.Repeat("f", 64)
				}
				pictures = append(pictures, map[string]any{"path": name, "sha256": pin, "width": size[0], "height": size[1]})
			}
			cellPin := fmt.Sprintf("%x", sha256.Sum256(cells.Bytes()))
			if tc.badCells {
				cellPin = strings.Repeat("f", 64)
			}
			sources := map[string]string{}
			for _, name := range []string{"00.PAL", "LIB/PARTS.LIB", "LIB/PARTS.DAT", "MAP/00.FIN"} {
				sources[name] = strings.Repeat("0", 64)
			}
			data, err := json.Marshal(map[string]any{"schema": 1, "platform": tc.platform, "map": "00", "library": "PARTS", "palette": "00", "columns": 16, "rows": 24, "native": pictures[0], "aligned": pictures[1], "sources": sources, "cells_sha256": cellPin})
			if err != nil {
				t.Fatal(err)
			}
			if err := r.WriteFile("build/dos/reference.json", data, 0644); err != nil {
				t.Fatal(err)
			}
			c := collector{root: r, out: r, seen: map[string]picture{}}
			study := pulseStudy{Width: 1090, Height: 1920}
			err = c.pulseDOS(&study, "build/dos")
			if (err != nil) != tc.wantErr {
				t.Fatalf("DOS ingestion error=%v wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && (study.DOS == nil || study.DOSNative == nil || study.DOS.Width != 1090 || study.DOSNative.Width != 264) {
				t.Fatal("missing verified DOS snapshots")
			}
		})
	}
}

func TestPulseRejectsInvalidCaptureCollection(t *testing.T) {
	lvl, err := level.Load("../../levels/level1_cp.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := lvl.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, hash string }{{"stale-level", "wrong"}, {"duplicate-views", fmt.Sprintf("%x", sha256.Sum256(b))}} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = r.Close() }()
			if err := r.MkdirAll("build/captures/media", 0755); err != nil {
				t.Fatal(err)
			}
			frames := make([]pulseCapture, 9)
			for i := range frames {
				frames[i] = pulseCapture{File: "frame.png", View: "overview", Width: 1280, Height: 720, Zoom: 1, UI: 1, Density: 2, Reveal: true}
			}
			manifest := map[string]any{"level": "Pulse", "canonical_level_sha256": tc.hash, "frames": frames}
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := r.WriteFile("build/captures/captures.json", data, 0644); err != nil {
				t.Fatal(err)
			}
			var pngBytes bytes.Buffer
			if err := png.Encode(&pngBytes, image.NewNRGBA(image.Rect(0, 0, 1280, 720))); err != nil {
				t.Fatal(err)
			}
			if err := r.WriteFile("build/captures/frame.png", pngBytes.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
			out, err := r.OpenRoot("build/captures")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = out.Close() }()
			c := collector{root: r, out: out, seen: map[string]picture{}}
			if err := c.pulseCaptures(&pulseStudy{}, "build/captures", lvl); err == nil {
				t.Fatal("invalid capture collection was accepted")
			}
		})
	}
}

const pulseSmoke = `(async()=>{
const assert=(v,m)=>{if(!v)throw Error(m)};
const change=(id,value)=>{document.getElementById(id).value=value;document.getElementById(id).dispatchEvent(new Event('change'))};
try{
change('native','hide');assert(document.querySelectorAll('#map-stage figure').length===2,'side by side');
change('region','north');assert(document.querySelector('#map-stage .crop').style.height==='320px','matched northern crop');
assert(document.querySelector('#mode [value="wipe"]').disabled,'no invented DOS reference');
change('density','4');change('mode','current');assert(document.querySelectorAll('#map-stage figure').length===1,'current plane');
assert(document.querySelector('#map-stage img').style.top==='-40px','crop alignment');
change('layers','units');assert(document.querySelector('#map-stage figcaption').textContent.includes('with current units'),'unit layers');
change('density','1');change('mode','current');
await Promise.all(Array.from(document.querySelectorAll('#map-stage img')).map(im=>im.decode()));
assert(document.querySelector('#map-stage img').naturalWidth===1090,'reduced map');
assert(document.querySelector('#frame .empty'),'missing captures explicit');
assert(!document.getElementById('error').textContent,'image errors');
document.body.dataset.smoke='pass';
}catch(e){document.body.dataset.smoke='fail: '+e.message;}
})();`
